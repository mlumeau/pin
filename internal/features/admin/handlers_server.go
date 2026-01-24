package admin

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	sortpkg "sort"
	"strconv"
	"strings"
	"time"

	"pin/internal/domain"
	featuresettings "pin/internal/features/settings"
)

// Paging defaults for the admin server view.
const (
	userPageSize  = 10
	auditPageSize = 20
)

// userSummary is a lightweight view model for the admin user list.
type userSummary struct {
	ID          int
	Handle      string
	DisplayName string
	Email       string
	Role        string
	UpdatedAt   time.Time
}

// Server handles the HTTP request.
func (h Handler) Server(w http.ResponseWriter, r *http.Request) {
	session, _ := h.deps.GetSession(r, "pin_session")

	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), &current)
	isAdminUser := isAdmin(current)
	landing := settingsSvc.LandingSettings(r.Context())
	message := r.URL.Query().Get("toast")
	var users []userSummary
	var invites []domain.Invite
	var auditLogs []domain.AuditLog
	usedInvitesCount := 0
	defaultTheme := featuresettings.DefaultThemeName
	defaultThemeForce := false
	themeValue, ok, _ := settingsSvc.ServerDefaultTheme(r.Context())
	if ok {
		defaultTheme = themeValue
	}
	defaultCustomCSSPath, hasDefaultCustomCSS := settingsSvc.ServerDefaultCustomCSS(r.Context())
	themePolicy := settingsSvc.ServerThemePolicy(r.Context())
	showAppearanceNav := isAdminUser || themePolicy.AllowUserTheme
	userPage := 1
	userPrevPage := 1
	userNextPage := 1
	userTotalPages := 1
	userQuery := ""
	userSort := "id"
	userDir := "desc"
	auditPage := 1
	auditPrevPage := 1
	auditNextPage := 1
	auditHasMore := false
	auditTotalPages := 1

	if r.Method == http.MethodPost {
		if !isAdminUser {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, h.deps.Config().MaxUploadBytes)
		if err := r.ParseMultipartForm(h.deps.Config().MaxUploadBytes); err != nil {
			if errors.Is(err, http.ErrNotMultipart) {
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}
			} else {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}
		if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
			http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
			return
		}
		action := serverAction(r)
		result, err := h.handleServerAction(w, r, settingsSvc, landing, defaultCustomCSSPath, defaultThemeForce, action)
		if err != nil {
			if reqErr, ok := err.(requestError); ok {
				http.Error(w, reqErr.Error(), reqErr.status)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if result.wroteResponse {
			return
		}
		if result.redirectURL != "" {
			http.Redirect(w, r, result.redirectURL, http.StatusSeeOther)
			return
		}
		if result.message != "" {
			message = result.message
		}
	}

	userQuery, userSort, userDir, userPage = parseUserListParams(r)
	users, total, err := loadUserSummaries(r.Context(), h.deps, userQuery, userSort, userDir, userPage, userPageSize)
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}
	userPrevPage, userNextPage, userTotalPages = pageBounds(userPage, userPageSize, total)

	invites, err = h.deps.ListInvites(r.Context())
	if err != nil {
		http.Error(w, "Failed to load invites", http.StatusInternalServerError)
		return
	}
	for _, invite := range invites {
		if invite.UsedAt.Valid {
			usedInvitesCount++
		}
	}

	auditPage = parsePageParam(r, "audit_page")
	auditLogs, auditTotal, err := loadAuditLogs(r.Context(), h.deps, auditPage, auditPageSize)
	if err != nil {
		http.Error(w, "Failed to load audit log", http.StatusInternalServerError)
		return
	}
	auditPrevPage, auditNextPage, auditTotalPages = pageBounds(auditPage, auditPageSize, auditTotal)
	auditHasMore = auditPage < auditTotalPages

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":                 current,
		"IsAdmin":              isAdmin(current),
		"Message":              message,
		"CSRFToken":            h.deps.EnsureCSRF(session),
		"Theme":                theme,
		"ShowAppearanceNav":    showAppearanceNav,
		"Landing":              landing,
		"LandingMode":          landing.Mode,
		"LandingCustomPath":    landing.CustomPath,
		"HasCustomLandingFile": landing.CustomPath != "",
		"LandingCustomURL":     featuresettings.LandingCustomURL(landing.CustomPath),
		"DefaultTheme":         defaultTheme,
		"DefaultThemeForce":    defaultThemeForce,
		"DefaultCustomCSS":     featuresettings.ThemeCustomCSSURL(defaultCustomCSSPath),
		"DefaultCustomCSSName": filepath.Base(defaultCustomCSSPath),
		"HasDefaultCustomCSS":  hasDefaultCustomCSS,
		"ThemeOptions":         featuresettings.ThemeOptions(),
		"ThemePolicy":          themePolicy,
		"AllowUserTheme":       themePolicy.AllowUserTheme,
		"AllowUserCustomCSS":   themePolicy.AllowUserCustomCSS,
		"Users":                users,
		"UsersQuery":           userQuery,
		"UsersSort":            userSort,
		"UsersDir":             userDir,
		"UsersPage":            userPage,
		"UsersPrevPage":        userPrevPage,
		"UsersNextPage":        userNextPage,
		"UsersTotal":           userTotalPages,
		"Invites":              invites,
		"UsedInvitesCount":     usedInvitesCount,
		"InviteBaseURL":        h.deps.BaseURL(r),
		"AuditLogs":            auditLogs,
		"AuditPage":            auditPage,
		"AuditPrevPage":        auditPrevPage,
		"AuditNextPage":        auditNextPage,
		"AuditHasMore":         auditHasMore,
		"AuditTotal":           auditTotal,
		"AuditTotalPages":      auditTotalPages,
	}

	if err := h.deps.RenderTemplate(w, "settings_admin.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

type serverActionResult struct {
	message       string
	redirectURL   string
	wroteResponse bool
}

// requestError preserves a caller-controlled HTTP status for expected validation errors.
type requestError struct {
	err    error
	status int
}

func (e requestError) Error() string {
	return e.err.Error()
}

// serverAction normalizes the form action name across legacy field names.
func serverAction(r *http.Request) string {
	action := strings.TrimSpace(r.FormValue("server_action"))
	if action == "" {
		action = strings.TrimSpace(r.FormValue("action"))
	}
	return action
}

// handleServerAction routes admin actions and returns any follow-up message/redirect.
func (h Handler) handleServerAction(w http.ResponseWriter, r *http.Request, settingsSvc featuresettings.Service, landing featuresettings.LandingSettings, defaultCustomCSSPath string, defaultThemeForce bool, action string) (serverActionResult, error) {
	switch action {
	case "export-users":
		if err := h.exportUsersCSV(w, r); err != nil {
			return serverActionResult{}, err
		}
		return serverActionResult{wroteResponse: true}, nil
	case "landing":
		message, err := h.saveLandingSettings(r, settingsSvc, landing)
		return serverActionResult{message: message}, err
	case "theme_default":
		result, err := h.saveDefaultTheme(r, settingsSvc, defaultCustomCSSPath, defaultThemeForce)
		return result, err
	case "theme_access":
		message, err := h.saveThemePolicy(r, settingsSvc)
		return serverActionResult{message: message}, err
	case "theme_default_css":
		message, err := h.saveDefaultCustomCSS(r, settingsSvc, defaultCustomCSSPath)
		return serverActionResult{message: message}, err
	default:
		return serverActionResult{}, nil
	}
}

// exportUsersCSV writes a CSV of identities with role metadata.
func (h Handler) exportUsersCSV(w http.ResponseWriter, r *http.Request) error {
	identities, err := h.deps.ListIdentities(r.Context())
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=users.csv")
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"id", "handle", "email", "role", "updated_at"})
	for _, identityRecord := range identities {
		role := ""
		if authUser, err := h.deps.GetUserByID(r.Context(), identityRecord.UserID); err == nil {
			role = authUser.Role
		}
		_ = writer.Write([]string{
			strconv.Itoa(identityRecord.ID),
			identityRecord.Handle,
			identityRecord.Email,
			role,
			identityRecord.UpdatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
	return nil
}

// saveLandingSettings updates the landing mode and optional custom HTML asset.
func (h Handler) saveLandingSettings(r *http.Request, settingsSvc featuresettings.Service, landing featuresettings.LandingSettings) (string, error) {
	landingMode := strings.TrimSpace(r.FormValue("landing_mode"))
	if landingMode != "generic" && landingMode != "profile" && landingMode != "custom" {
		landingMode = "generic"
	}
	if r.FormValue("landing_remove_custom") == "1" {
		customPath := strings.TrimSpace(landing.CustomPath)
		if customPath != "" {
			_ = os.Remove(filepath.Join(featuresettings.LandingDir(h.deps.Config()), customPath))
		}
		if err := settingsSvc.SaveLandingSettings(r.Context(), featuresettings.LandingSettings{Mode: landingMode, CustomPath: ""}); err != nil {
			return "", err
		}
		return "Custom landing page removed.", nil
	}
	if file, header, err := r.FormFile("landing_custom_file"); err == nil && header != nil && header.Filename != "" {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".html" && ext != ".htm" {
			return "", requestError{err: errors.New("Landing page must be an HTML file"), status: http.StatusBadRequest}
		}
		if err := os.MkdirAll(featuresettings.LandingDir(h.deps.Config()), 0755); err != nil {
			return "", err
		}
		filename := fmt.Sprintf("landing_custom_%d.html", time.Now().UTC().UnixNano())
		destPath := filepath.Join(featuresettings.LandingDir(h.deps.Config()), filename)
		dest, err := os.Create(destPath)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(dest, io.LimitReader(file, h.deps.Config().MaxUploadBytes)); err != nil {
			_ = dest.Close()
			_ = os.Remove(destPath)
			return "", err
		}
		_ = dest.Close()
		if prev := strings.TrimSpace(landing.CustomPath); prev != "" && prev != filename {
			_ = os.Remove(filepath.Join(featuresettings.LandingDir(h.deps.Config()), prev))
		}
		if err := settingsSvc.SaveLandingSettings(r.Context(), featuresettings.LandingSettings{Mode: landingMode, CustomPath: filename}); err != nil {
			return "", err
		}
		return "Landing page updated.", nil
	}
	if err := settingsSvc.SaveLandingSettings(r.Context(), featuresettings.LandingSettings{Mode: landingMode, CustomPath: landing.CustomPath}); err != nil {
		return "", err
	}
	return "Landing page saved.", nil
}

// saveDefaultTheme persists the server default theme and optional force behavior.
func (h Handler) saveDefaultTheme(r *http.Request, settingsSvc featuresettings.Service, defaultCustomCSSPath string, defaultThemeForce bool) (serverActionResult, error) {
	defaultThemeForce = r.FormValue("default_theme_force") == "1"
	themeValue := featuresettings.NormalizeThemeChoice(r.FormValue("default_theme"))
	if themeValue == featuresettings.DefaultCustomThemeName && defaultCustomCSSPath == "" {
		themeValue = featuresettings.DefaultThemeName
	}
	if err := settingsSvc.SaveServerDefaultTheme(r.Context(), themeValue, defaultThemeForce); err != nil {
		return serverActionResult{}, err
	}
	message := "Default theme saved."
	if defaultThemeForce && themeValue != "" {
		if err := h.deps.ResetAllUserThemes(r.Context(), themeValue); err != nil {
			return serverActionResult{}, err
		}
		message = "Default theme saved and user themes reset."
	}
	redirectURL := "/settings/admin/server?toast=" + url.QueryEscape(message) + "#section-theme"
	return serverActionResult{message: message, redirectURL: redirectURL}, nil
}

// saveThemePolicy updates whether users can choose themes and custom CSS.
func (h Handler) saveThemePolicy(r *http.Request, settingsSvc featuresettings.Service) (string, error) {
	policy := featuresettings.ThemePolicy{
		AllowUserTheme:     r.FormValue("allow_user_theme") == "1",
		AllowUserCustomCSS: r.FormValue("allow_user_custom_css") == "1",
	}
	if err := settingsSvc.SaveServerThemePolicy(r.Context(), policy); err != nil {
		return "", err
	}
	return "Theme policy saved.", nil
}

// saveDefaultCustomCSS uploads or removes the server-wide default CSS file.
func (h Handler) saveDefaultCustomCSS(r *http.Request, settingsSvc featuresettings.Service, defaultCustomCSSPath string) (string, error) {
	if r.FormValue("remove_default_custom_css") == "1" {
		if defaultCustomCSSPath != "" {
			_ = os.Remove(filepath.Join(featuresettings.ThemeDir(h.deps.Config()), defaultCustomCSSPath))
		}
		_ = settingsSvc.SaveServerDefaultCustomCSS(r.Context(), "")
		return "Default CSS removed.", nil
	}
	file, header, err := r.FormFile("default_custom_css_file")
	if err != nil || header == nil || header.Filename == "" {
		return "", requestError{err: errors.New("Missing CSS file"), status: http.StatusBadRequest}
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".css" {
		return "", requestError{err: errors.New("Default CSS must be a .css file"), status: http.StatusBadRequest}
	}
	if err := os.MkdirAll(featuresettings.ThemeDir(h.deps.Config()), 0755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("theme_default_%d.css", time.Now().UTC().UnixNano())
	destPath := filepath.Join(featuresettings.ThemeDir(h.deps.Config()), filename)
	dest, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(dest, io.LimitReader(file, h.deps.Config().MaxUploadBytes)); err != nil {
		_ = dest.Close()
		_ = os.Remove(destPath)
		return "", err
	}
	_ = dest.Close()
	if prev := strings.TrimSpace(defaultCustomCSSPath); prev != "" && prev != filename {
		_ = os.Remove(filepath.Join(featuresettings.ThemeDir(h.deps.Config()), prev))
	}
	if err := settingsSvc.SaveServerDefaultCustomCSS(r.Context(), filename); err != nil {
		return "", err
	}
	return "Default CSS uploaded.", nil
}

// parseUserListParams normalizes sorting, filtering, and pagination query params.
func parseUserListParams(r *http.Request) (query, sort, dir string, page int) {
	query = strings.TrimSpace(r.URL.Query().Get("q"))
	sort = "id"
	dir = "desc"
	dirOverride := false
	if dirParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dir"))); dirParam == "asc" || dirParam == "desc" {
		dir = dirParam
		dirOverride = true
	}
	if sortParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort"))); sortParam == "handle" || sortParam == "email" || sortParam == "role" || sortParam == "updated" {
		sort = sortParam
		if !dirOverride {
			dir = "asc"
		}
	}
	page = parsePageParam(r, "page")
	return query, sort, dir, page
}

// loadUserSummaries returns the requested user page and total count.
func loadUserSummaries(ctx context.Context, deps Dependencies, query, sort, dir string, page, pageSize int) ([]userSummary, int, error) {
	if strings.EqualFold(sort, "role") {
		identities, err := deps.ListIdentities(ctx)
		if err != nil {
			return nil, 0, err
		}
		users := make([]userSummary, 0, len(identities))
		for _, identityRecord := range identities {
			users = append(users, buildUserSummary(ctx, deps, identityRecord))
		}
		sortpkg.Slice(users, func(i, j int) bool {
			left := strings.ToLower(users[i].Role)
			right := strings.ToLower(users[j].Role)
			if left == right {
				if strings.EqualFold(dir, "desc") {
					return strings.ToLower(users[i].Handle) > strings.ToLower(users[j].Handle)
				}
				return strings.ToLower(users[i].Handle) < strings.ToLower(users[j].Handle)
			}
			if strings.EqualFold(dir, "desc") {
				return left > right
			}
			return left < right
		})
		total := len(users)
		start, end := pageSlice(page, pageSize, total)
		return users[start:end], total, nil
	}
	identities, totalCount, err := deps.ListIdentitiesPaged(ctx, query, sort, dir, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	users := make([]userSummary, 0, len(identities))
	for _, identityRecord := range identities {
		users = append(users, buildUserSummary(ctx, deps, identityRecord))
	}
	return users, totalCount, nil
}

// buildUserSummary attaches role metadata to an identity record.
func buildUserSummary(ctx context.Context, deps Dependencies, identityRecord domain.Identity) userSummary {
	role := ""
	if authUser, err := deps.GetUserByID(ctx, identityRecord.UserID); err == nil {
		role = authUser.Role
	}
	return userSummary{
		ID:          identityRecord.ID,
		Handle:      identityRecord.Handle,
		DisplayName: identityRecord.DisplayName,
		Email:       identityRecord.Email,
		Role:        role,
		UpdatedAt:   identityRecord.UpdatedAt,
	}
}

// loadAuditLogs returns a page of audit logs and the total count.
func loadAuditLogs(ctx context.Context, deps Dependencies, page, pageSize int) ([]domain.AuditLog, int, error) {
	logs, err := deps.ListAuditLogs(ctx, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, _ := deps.CountAuditLogs(ctx)
	return logs, total, nil
}

// parsePageParam reads a positive page number with a default of 1.
func parsePageParam(r *http.Request, name string) int {
	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get(name)); err == nil && p > 0 {
		page = p
	}
	return page
}

// pageBounds computes pagination helpers for templates.
func pageBounds(page, pageSize, total int) (prev, next, totalPages int) {
	if total == 0 {
		return 1, 1, 1
	}
	totalPages = (total + pageSize - 1) / pageSize
	prev = page - 1
	if prev < 1 {
		prev = 1
	}
	next = page + 1
	if next > totalPages {
		next = totalPages
	}
	return prev, next, totalPages
}

// pageSlice returns start/end indices clamped to the total size.
func pageSlice(page, pageSize, total int) (int, int) {
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	return start, end
}
