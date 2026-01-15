package profilepicture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/platform/media"
)

type Config struct {
	ProfilePictureDir string
	StaticDir         string
	AllowedExts       map[string]bool
	MaxUploadBytes    int64
	CacheAltFormats   bool
}

type Dependencies interface {
	Store
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	ValidateCSRF(session *sessions.Session, token string) bool
	CurrentUser(r *http.Request) (domain.User, error)
	FindUserByIdentifier(ctx context.Context, identifier string) (domain.User, error)
	GetOwnerUser(ctx context.Context) (domain.User, error)
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	cfg  Config
	deps Dependencies
	svc  Service
}

func NewHandler(cfg Config, deps Dependencies) Handler {
	return Handler{cfg: cfg, deps: deps, svc: NewService(deps)}
}

// ProfilePicture serves the uploaded profile picture or falls back to the default image.
func (h Handler) ProfilePicture(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/profile-picture/" {
		h.ProfilePictureRoot(w, r)
		return
	}
	username := strings.TrimPrefix(r.URL.Path, "/profile-picture/")
	if username == "" {
		http.NotFound(w, r)
		return
	}

	user, err := h.deps.FindUserByIdentifier(r.Context(), username)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !identity.MatchesIdentity(user, username) {
		http.NotFound(w, r)
		return
	}

	size := parseProfilePictureSize(r)
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	desiredFormat := resolveProfilePictureFormat(r, format)

	activeFilename := sql.NullString{}
	if picID := strings.TrimSpace(r.URL.Query().Get("picture_id")); picID != "" {
		if id, err := strconv.ParseInt(picID, 10, 64); err == nil {
			if filename, err := h.svc.Filename(r.Context(), user.ID, id); err == nil && filename != "" {
				activeFilename = sql.NullString{String: filename, Valid: true}
			}
		}
	} else if user.ProfilePictureID.Valid {
		if filename, err := h.svc.Filename(r.Context(), user.ID, user.ProfilePictureID.Int64); err == nil && filename != "" {
			activeFilename = sql.NullString{String: filename, Valid: true}
		}
	}

	if activeFilename.Valid {
		profilePicturePath := filepath.Join(h.cfg.ProfilePictureDir, activeFilename.String)
		if _, err := os.Stat(profilePicturePath); err == nil {
			base := strings.TrimSuffix(activeFilename.String, filepath.Ext(activeFilename.String))
			cacheName := fmt.Sprintf("%s_%d.webp", base, size)
			cachePath := filepath.Join(h.cfg.ProfilePictureDir, "cache", cacheName)
			if _, err := os.Stat(cachePath); err == nil {
				serveProfilePictureWithFormat(w, r, cachePath, desiredFormat)
				return
			}
			if err := media.ResizeAndCache(profilePicturePath, cachePath, size); err == nil {
				serveProfilePictureWithFormat(w, r, cachePath, desiredFormat)
				return
			} else if errors.Is(err, media.ErrImageTooSmall) {
				serveProfilePictureWithFormat(w, r, profilePicturePath, desiredFormat)
				return
			}
			serveProfilePictureWithFormat(w, r, profilePicturePath, desiredFormat)
			return
		}
	}

	defaultPath := filepath.Join(h.cfg.StaticDir, "img", "default_profile_picture.png")
	if size == media.DefaultSize && desiredFormat == "webp" {
		webpPath := filepath.Join(h.cfg.StaticDir, "img", "default_profile_picture.webp")
		if _, err := os.Stat(webpPath); err == nil {
			serveProfilePictureWithFormat(w, r, webpPath, desiredFormat)
			return
		}
	}
	cachePath := filepath.Join(h.cfg.ProfilePictureDir, "cache", fmt.Sprintf("default_%d.webp", size))
	if _, err := os.Stat(cachePath); err == nil {
		serveProfilePictureWithFormat(w, r, cachePath, desiredFormat)
		return
	}
	if err := media.ResizeAndCache(defaultPath, cachePath, size); err == nil {
		serveProfilePictureWithFormat(w, r, cachePath, desiredFormat)
		return
	} else if errors.Is(err, media.ErrImageTooSmall) {
		serveProfilePictureWithFormat(w, r, defaultPath, desiredFormat)
		return
	}
	serveProfilePictureWithFormat(w, r, defaultPath, desiredFormat)
}

// ProfilePictureRoot redirects /profile-picture to the default user's profile picture.
func (h Handler) ProfilePictureRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile-picture" && r.URL.Path != "/profile-picture/" {
		http.NotFound(w, r)
		return
	}
	user, err := h.deps.GetOwnerUser(r.Context())
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}
	target := "/profile-picture/" + url.PathEscape(user.Username)
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (h Handler) Select(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	picID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("profile_picture_id")), 10, 64)
	if err != nil || picID <= 0 {
		http.Error(w, "Invalid profile picture", http.StatusBadRequest)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.select", strconv.FormatInt(picID, 10), nil)
	if err := h.svc.Select(r.Context(), current.ID, picID); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.select", strconv.FormatInt(picID, 10), err, nil)
		http.Error(w, "Failed to select profile picture", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.select", strconv.FormatInt(picID, 10), nil, nil)
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"ok":       true,
			"active":   picID,
			"pictures": h.mustListProfilePictures(r.Context(), current.ID),
		})
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	picID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("profile_picture_id")), 10, 64)
	if err != nil || picID <= 0 {
		http.Error(w, "Invalid profile picture", http.StatusBadRequest)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.delete", strconv.FormatInt(picID, 10), nil)
	filename, err := h.svc.Delete(r.Context(), current.ID, picID)
	if err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.delete", strconv.FormatInt(picID, 10), err, nil)
		http.Error(w, "Failed to delete profile picture", http.StatusInternalServerError)
		return
	}
	if filename != "" {
		_ = os.Remove(filepath.Join(h.cfg.ProfilePictureDir, filename))
	}
	if current.ProfilePictureID.Valid && current.ProfilePictureID.Int64 == picID {
		pics := h.mustListProfilePictures(r.Context(), current.ID)
		if len(pics) > 0 {
			if err := h.svc.Select(r.Context(), current.ID, pics[0].ID); err == nil {
				current.ProfilePictureID = sql.NullInt64{Int64: pics[0].ID, Valid: true}
			}
		} else {
			_ = h.svc.ClearSelection(r.Context(), current.ID)
			current.ProfilePictureID = sql.NullInt64{}
		}
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.delete", strconv.FormatInt(picID, 10), nil, map[string]string{"filename": filename})
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"ok":       true,
			"active":   activePictureID(current),
			"pictures": h.mustListProfilePictures(r.Context(), current.ID),
		})
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

func (h Handler) UpdateAlt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	picID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("profile_picture_id")), 10, 64)
	if err != nil || picID <= 0 {
		http.Error(w, "Invalid profile picture", http.StatusBadRequest)
		return
	}
	alt := strings.TrimSpace(r.FormValue("profile_picture_alt"))
	h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.alt", strconv.FormatInt(picID, 10), nil)
	if err := h.svc.UpdateAlt(r.Context(), current.ID, picID, alt); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.alt", strconv.FormatInt(picID, 10), err, nil)
		http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.alt", strconv.FormatInt(picID, 10), nil, nil)
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"ok":       true,
			"active":   activePictureID(current),
			"pictures": h.mustListProfilePictures(r.Context(), current.ID),
		})
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

func (h Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		http.Error(w, "Upload too large", http.StatusBadRequest)
		return
	}
	if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	profilePictureFile, profilePictureHeader, err := r.FormFile("profile_picture")
	if err != nil || profilePictureHeader == nil || profilePictureHeader.Filename == "" {
		http.Error(w, "Missing profile picture", http.StatusBadRequest)
		return
	}
	defer profilePictureFile.Close()
	ext := strings.ToLower(filepath.Ext(profilePictureHeader.Filename))
	if !h.cfg.AllowedExts[ext] {
		http.Error(w, "Profile picture must be an image (png/jpg/gif)", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(h.cfg.ProfilePictureDir, 0755); err != nil {
		http.Error(w, "Failed to store profile picture", http.StatusInternalServerError)
		return
	}
	filename := fmt.Sprintf("profile_picture_%d.webp", time.Now().UTC().UnixNano())
	if err := media.WriteWebP(profilePictureFile, filepath.Join(h.cfg.ProfilePictureDir, filename)); err != nil {
		WriteProfilePictureStoreError(w, err)
		return
	}
	altText := strings.TrimSpace(r.FormValue("profile_picture_alt"))
	meta := map[string]string{"filename": filename}
	h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(current.ID), 10), meta)
	picID, err := h.svc.Create(r.Context(), current.ID, filename, altText)
	if err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(current.ID), 10), err, meta)
		http.Error(w, "Failed to save profile picture", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(picID, 10), nil, meta)
	h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.select", strconv.FormatInt(picID, 10), nil)
	err = h.svc.Select(r.Context(), current.ID, picID)
	h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.select", strconv.FormatInt(picID, 10), err, nil)
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"ok":       err == nil,
			"active":   picID,
			"pictures": h.mustListProfilePictures(r.Context(), current.ID),
		})
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}

func (h Handler) ActiveAlt(ctx context.Context, user domain.User) string {
	return h.svc.ActiveAlt(ctx, user)
}

func (h Handler) mustListProfilePictures(ctx context.Context, userID int) []domain.ProfilePicture {
	pics, err := h.svc.List(ctx, userID)
	if err != nil {
		return []domain.ProfilePicture{}
	}
	return pics
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(data)
}
