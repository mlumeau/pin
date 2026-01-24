package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"pin/internal/config"
	"pin/internal/domain"
	"pin/internal/features/domains"
	"pin/internal/features/identity"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/media"
)

type Dependencies interface {
	featuresettings.Store
	Config() config.Config
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	EnsureCSRF(session *sessions.Session) string
	ValidateCSRF(session *sessions.Session, token string) bool
	CurrentUser(r *http.Request) (domain.User, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	GetIdentityByUserID(ctx context.Context, userID int) (domain.Identity, error)
	DeleteUser(ctx context.Context, userID int) error
	UpdateUser(ctx context.Context, user domain.User) error
	UpdateIdentity(ctx context.Context, identity domain.Identity) error
	CheckHandleCollision(ctx context.Context, handle string, excludeID int) error
	Reserved() map[string]struct{}
	ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error)
	UpsertDomainVerification(ctx context.Context, identityID int, domainName, token string) error
	DeleteDomainVerification(ctx context.Context, identityID int, domainName string) error
	MarkDomainVerified(ctx context.Context, identityID int, domainName string) error
	ProtectedDomain(ctx context.Context) string
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	deps Dependencies
}

// NewHandler constructs a new handler.
func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}

// Users handles the HTTP request.
func (h Handler) Users(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/settings/admin/server#section-users", http.StatusFound)
}

// User handles the HTTP request.
func (h Handler) User(w http.ResponseWriter, r *http.Request) {
	current, err := h.deps.CurrentUser(r)
	if err != nil || !isAdmin(current) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/settings/admin/users/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(path, "/delete") {
		idStr := strings.TrimSuffix(path, "/delete")
		id, _ := strconv.Atoi(strings.Trim(idStr, "/"))
		targetUser, err := h.deps.GetUserByID(r.Context(), id)
		if err != nil || targetUser.Role == "owner" || targetUser.ID == current.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		targetIdentity, err := h.deps.GetIdentityByUserID(r.Context(), targetUser.ID)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.deps.AuditAttempt(r.Context(), current.ID, "user.delete", targetIdentity.Handle, map[string]string{"role": targetUser.Role})
		if err := h.deps.DeleteUser(r.Context(), targetUser.ID); err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "user.delete", targetIdentity.Handle, err, map[string]string{"role": targetUser.Role})
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
		h.deps.AuditOutcome(r.Context(), current.ID, "user.delete", targetIdentity.Handle, nil, map[string]string{"role": targetUser.Role})
		http.Redirect(w, r, "/settings/admin/server#section-users", http.StatusFound)
		return
	}
	if strings.HasSuffix(path, "/edit") {
		idStr := strings.TrimSuffix(path, "/edit")
		id, _ := strconv.Atoi(strings.Trim(idStr, "/"))
		targetUser, err := h.deps.GetUserByID(r.Context(), id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		targetIdentity, err := h.deps.GetIdentityByUserID(r.Context(), targetUser.ID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		session, _ := h.deps.GetSession(r, "pin_session")
		var links []domain.Link
		if targetIdentity.LinksJSON != "" {
			links = identity.DecodeLinks(targetIdentity.LinksJSON)
		}
		var socialProfiles []domain.SocialProfile
		if targetIdentity.SocialProfilesJSON != "" {
			socialProfiles = identity.DecodeSocialProfiles(targetIdentity.SocialProfilesJSON)
		}
		visibility := identity.DecodeVisibilityMap(targetIdentity.VisibilityJSON)

		settingsSvc := featuresettings.NewService(h.deps)
		theme := settingsSvc.ThemeSettings(r.Context(), &current)
		showAppearanceNav := isAdmin(current)
		userView := struct {
			domain.Identity
			Role string
		}{
			Identity: targetIdentity,
			Role:     targetUser.Role,
		}
		data := map[string]interface{}{
			"User":                  userView,
			"Links":                 BuildLinkEntries(links, visibility),
			"SocialProfiles":        BuildSocialEntries(socialProfiles, visibility),
			"CustomFields":          identity.DecodeStringMap(targetIdentity.CustomFieldsJSON),
			"FieldVisibility":       visibility,
			"CustomFieldVisibility": VisibilityCustomMap(visibility),
			"Wallets":               identity.DecodeStringMap(targetIdentity.WalletsJSON),
			"WalletEntries":         BuildWalletEntries(identity.DecodeStringMap(targetIdentity.WalletsJSON), visibility),
			"PublicKeys":            identity.DecodeStringMap(targetIdentity.PublicKeysJSON),
			"VerifiedDomains":       VerifiedDomainsToText(targetIdentity.VerifiedDomainsJSON),
			"DomainVerifications":   []domain.DomainVerification{},
			"GitHubOAuthEnabled":    false,
			"RedditOAuthEnabled":    false,
			"BlueskyEnabled":        false,
			"IsAdmin":               true,
			"IsOwner":               targetUser.Role == "owner",
			"IsSelf":                false,
			"FormAction":            "/settings/admin/users/" + strconv.Itoa(targetUser.ID) + "/edit",
			"CanEditRole":           targetUser.Role != "owner",
			"Title":                 "Settings - Edit User",
			"Message":               "",
			"CSRFToken":             h.deps.EnsureCSRF(session),
			"Theme":                 theme,
			"ShowAppearanceNav":     showAppearanceNav,
			"ProtectedDomain":       h.deps.ProtectedDomain(r.Context()),
			"DomainVisibility":      DomainVisibilityMap(visibility),
		}
		if rows, err := h.deps.ListDomainVerifications(r.Context(), targetIdentity.ID); err == nil {
			if len(rows) == 0 {
				rows = domains.NewService(h.deps).SeedDomains(r.Context(), targetIdentity.ID, identity.DecodeStringSlice(targetIdentity.VerifiedDomainsJSON), func() string {
					return domains.RandomTokenURL(12)
				})
			}
			data["DomainVerifications"] = rows
			data["VerifiedDomains"] = identity.DomainsToText(rows)
			data["ATProtoHandleVerified"] = identity.IsATProtoHandleVerified(targetIdentity.ATProtoHandle, identity.VerifiedDomains(rows))
		}

		if r.Method == http.MethodPost {
			if err := r.ParseMultipartForm(h.deps.Config().MaxUploadBytes); err != nil {
				http.Error(w, "Upload too large", http.StatusBadRequest)
				return
			}
			if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
				http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
				return
			}

			displayName := strings.TrimSpace(r.FormValue("display_name"))
			email := strings.TrimSpace(r.FormValue("email"))
			bio := strings.TrimSpace(r.FormValue("bio"))
			links, linkVisibility := ParseLinksForm(r.Form["link_label"], r.Form["link_url"], r.Form["link_visibility"])
			customFields := ParseCustomFieldsForm(r.Form["custom_key"], r.Form["custom_value"])
			fieldVisibility := ParseVisibilityForm(r.Form, []string{
				"email",
				"organization",
				"job_title",
				"birthdate",
				"languages",
				"phone",
				"address",
				"location",
				"website",
				"pronouns",
				"timezone",
				"atproto_handle",
				"atproto_did",
				"key_pgp",
				"key_ssh",
				"key_age",
				"key_activitypub",
			})
			customVisibility := ParseCustomVisibilityForm(r.Form["custom_key"], r.Form["custom_value"], r.Form["custom_visibility"])
			social, socialVisibility := identity.ParseSocialForm(r.Form["social_label"], r.Form["social_url"], r.Form["social_visibility"])
			social = identity.MergeSocialProfiles(social, socialProfiles)
			handle := strings.TrimSpace(r.FormValue("handle"))
			if err := identity.ValidateHandle(r.Context(), handle, targetIdentity.ID, h.deps.Reserved(), h.deps.CheckHandleCollision); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			wallets, walletVisibility, err := ParseWalletForm(r.Form["wallet_label"], r.Form["wallet_address"], r.Form["wallet_visibility"])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			for key, value := range walletVisibility {
				fieldVisibility[key] = value
			}
			publicKeys := map[string]string{
				"pgp":         strings.TrimSpace(r.FormValue("key_pgp")),
				"ssh":         strings.TrimSpace(r.FormValue("key_ssh")),
				"age":         strings.TrimSpace(r.FormValue("key_age")),
				"activitypub": strings.TrimSpace(r.FormValue("key_activitypub")),
			}
			verifiedDomains := ParseVerifiedDomainsText(r.FormValue("verified_domains"))
			domainVisibility := ParseVerifiedDomainVisibilityForm(r.Form["verified_domain"], r.Form["verified_domain_visibility"])
			h.deps.AuditAttempt(r.Context(), current.ID, "domain.sync", targetIdentity.Handle, nil)
			_, verified, err := domains.NewService(h.deps).CreateDomains(r.Context(), targetIdentity.ID, verifiedDomains, func() string {
				return domains.RandomTokenURL(12)
			})
			if err != nil {
				h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", targetIdentity.Handle, err, nil)
				http.Error(w, "Failed to update verified domains", http.StatusInternalServerError)
				return
			}
			h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", targetIdentity.Handle, nil, nil)

			profilePictureFile, profilePictureHeader, err := r.FormFile("profile_picture")
			if err == nil && profilePictureHeader != nil && profilePictureHeader.Filename != "" {
				defer profilePictureFile.Close()
				ext := strings.ToLower(filepath.Ext(profilePictureHeader.Filename))
				if !h.deps.Config().AllowedExts[ext] {
					data["Message"] = "Profile picture must be an image (png/jpg/gif)"
				} else {
					if err := os.MkdirAll(h.deps.Config().ProfilePictureDir, 0755); err != nil {
						http.Error(w, "Failed to store profile picture", http.StatusInternalServerError)
						return
					}
					filename := fmt.Sprintf("profile_picture_%d.webp", time.Now().UTC().UnixNano())
					if err := media.WriteWebP(profilePictureFile, filepath.Join(h.deps.Config().ProfilePictureDir, filename)); err != nil {
						switch {
						case errors.Is(err, media.ErrCWebPUnavailable):
							http.Error(w, "WebP encoder unavailable", http.StatusServiceUnavailable)
						case errors.Is(err, media.ErrImageTooSmall):
							http.Error(w, "Image too small", http.StatusBadRequest)
						default:
							http.Error(w, "Failed to process profile picture", http.StatusInternalServerError)
						}
						return
					}
				}
			}

			if newPassword := r.FormValue("new_password"); newPassword != "" {
				hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
				if err != nil {
					http.Error(w, "Failed to update password", http.StatusInternalServerError)
					return
				}
				targetUser.PasswordHash = string(hash)
			}

			targetIdentity.Handle = handle
			targetIdentity.DisplayName = displayName
			targetIdentity.Email = email
			targetIdentity.Bio = bio
			targetIdentity.Organization = strings.TrimSpace(r.FormValue("organization"))
			targetIdentity.JobTitle = strings.TrimSpace(r.FormValue("job_title"))
			targetIdentity.Birthdate = strings.TrimSpace(r.FormValue("birthdate"))
			targetIdentity.Languages = strings.TrimSpace(r.FormValue("languages"))
			targetIdentity.Phone = strings.TrimSpace(r.FormValue("phone"))
			targetIdentity.Address = strings.TrimSpace(r.FormValue("address"))
			targetIdentity.Location = strings.TrimSpace(r.FormValue("location"))
			targetIdentity.Website = strings.TrimSpace(r.FormValue("website"))
			targetIdentity.Pronouns = strings.TrimSpace(r.FormValue("pronouns"))
			targetIdentity.Timezone = strings.TrimSpace(r.FormValue("timezone"))
			targetIdentity.ATProtoHandle = strings.TrimSpace(r.FormValue("atproto_handle"))
			targetIdentity.ATProtoDID = strings.TrimSpace(r.FormValue("atproto_did"))
			targetIdentity.LinksJSON = identity.EncodeLinks(links)
			if customJSON, err := json.Marshal(customFields); err == nil {
				targetIdentity.CustomFieldsJSON = string(customJSON)
			}
			visibility := BuildVisibilityMap(fieldVisibility, FilterCustomVisibility(customFields, customVisibility))
			for domain, vis := range domainVisibility {
				visibility["verified_domain:"+domain] = NormalizeVisibility(vis)
			}
			for key, value := range linkVisibility {
				visibility[key] = value
			}
			for key, value := range socialVisibility {
				visibility[key] = value
			}
			targetIdentity.VisibilityJSON = identity.EncodeVisibilityMap(visibility)
			targetIdentity.SocialProfilesJSON = identity.EncodeSocialProfiles(social)
			if walletsJSON, err := json.Marshal(identity.StripEmptyMap(wallets)); err == nil {
				targetIdentity.WalletsJSON = string(walletsJSON)
			}
			if keysJSON, err := json.Marshal(identity.StripEmptyMap(publicKeys)); err == nil {
				targetIdentity.PublicKeysJSON = string(keysJSON)
			}
			if domainsJSON, err := json.Marshal(verified); err == nil {
				targetIdentity.VerifiedDomainsJSON = string(domainsJSON)
			}
			if role := strings.TrimSpace(r.FormValue("role")); targetUser.Role != "owner" && (role == "admin" || role == "user") {
				targetUser.Role = role
			}

			h.deps.AuditAttempt(r.Context(), current.ID, "user.update", targetIdentity.Handle, nil)
			if err := h.deps.UpdateUser(r.Context(), targetUser); err != nil {
				h.deps.AuditOutcome(r.Context(), current.ID, "user.update", targetIdentity.Handle, err, nil)
				http.Error(w, "Failed to update profile", http.StatusInternalServerError)
				return
			}
			if err := h.deps.UpdateIdentity(r.Context(), targetIdentity); err != nil {
				h.deps.AuditOutcome(r.Context(), current.ID, "user.update", targetIdentity.Handle, err, nil)
				http.Error(w, "Failed to update profile", http.StatusInternalServerError)
				return
			}
			h.deps.AuditOutcome(r.Context(), current.ID, "user.update", targetIdentity.Handle, nil, nil)
			http.Redirect(w, r, "/settings/admin/server#section-users", http.StatusFound)
			return
		}

		if err := session.Save(r, w); err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		if err := h.deps.RenderTemplate(w, "settings_profile.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
		return
	}

	http.NotFound(w, r)
}

// isAdmin reports whether admin is true.
func isAdmin(user domain.User) bool {
	return strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
}
