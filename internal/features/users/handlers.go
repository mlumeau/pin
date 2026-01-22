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
	DeleteUser(ctx context.Context, userID int) error
	UpdateUser(ctx context.Context, user domain.User) error
	UpsertUserIdentifiers(ctx context.Context, userID int, username string, aliases []string, email string) error
	CheckIdentifierCollisions(ctx context.Context, identifiers []string, excludeID int) error
	Reserved() map[string]struct{}
	ListDomainVerifications(ctx context.Context, userID int) ([]domain.DomainVerification, error)
	UpsertDomainVerification(ctx context.Context, userID int, domainName, token string) error
	DeleteDomainVerification(ctx context.Context, userID int, domainName string) error
	MarkDomainVerified(ctx context.Context, userID int, domainName string) error
	ProtectedDomain(ctx context.Context) string
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	deps Dependencies
}

func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}

func (h Handler) Users(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/settings/admin/server#section-users", http.StatusFound)
}

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
		target, err := h.deps.GetUserByID(r.Context(), id)
		if err != nil || target.Role == "owner" || target.ID == current.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.deps.AuditAttempt(r.Context(), current.ID, "user.delete", target.Username, map[string]string{"role": target.Role})
		if err := h.deps.DeleteUser(r.Context(), target.ID); err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "user.delete", target.Username, err, map[string]string{"role": target.Role})
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
		h.deps.AuditOutcome(r.Context(), current.ID, "user.delete", target.Username, nil, map[string]string{"role": target.Role})
		http.Redirect(w, r, "/settings/admin/server#section-users", http.StatusFound)
		return
	}
	if strings.HasSuffix(path, "/edit") {
		idStr := strings.TrimSuffix(path, "/edit")
		id, _ := strconv.Atoi(strings.Trim(idStr, "/"))
		target, err := h.deps.GetUserByID(r.Context(), id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		session, _ := h.deps.GetSession(r, "pin_session")
		var links []domain.Link
		if target.LinksJSON != "" {
			links = identity.DecodeLinks(target.LinksJSON)
		}
		var socialProfiles []domain.SocialProfile
		if target.SocialProfilesJSON != "" {
			socialProfiles = identity.DecodeSocialProfiles(target.SocialProfilesJSON)
		}
		visibility := identity.DecodeVisibilityMap(target.VisibilityJSON)

		settingsSvc := featuresettings.NewService(h.deps)
		theme := settingsSvc.ThemeSettings(r.Context(), &current)
		showAppearanceNav := isAdmin(current)
		data := map[string]interface{}{
			"User":                  target,
			"Links":                 BuildLinkEntries(links, visibility),
			"SocialProfiles":        BuildSocialEntries(socialProfiles, visibility),
			"CustomFields":          identity.DecodeStringMap(target.CustomFieldsJSON),
			"FieldVisibility":       visibility,
			"CustomFieldVisibility": VisibilityCustomMap(visibility),
			"Wallets":               identity.DecodeStringMap(target.WalletsJSON),
			"WalletEntries":         BuildWalletEntries(identity.DecodeStringMap(target.WalletsJSON), visibility),
			"PublicKeys":            identity.DecodeStringMap(target.PublicKeysJSON),
			"VerifiedDomains":       VerifiedDomainsToText(target.VerifiedDomainsJSON),
			"DomainVerifications":   []domain.DomainVerification{},
			"Aliases":               AliasesToText(target.AliasesJSON),
			"GitHubOAuthEnabled":    false,
			"RedditOAuthEnabled":    false,
			"BlueskyEnabled":        false,
			"IsAdmin":               true,
			"IsOwner":               target.Role == "owner",
			"IsSelf":                false,
			"FormAction":            "/settings/admin/users/" + strconv.Itoa(target.ID) + "/edit",
			"CanEditRole":           target.Role != "owner",
			"Title":                 "Settings - Edit User",
			"Message":               "",
			"CSRFToken":             h.deps.EnsureCSRF(session),
			"Theme":                 theme,
			"ShowAppearanceNav":     showAppearanceNav,
			"ProtectedDomain":       h.deps.ProtectedDomain(r.Context()),
			"DomainVisibility":      DomainVisibilityMap(visibility),
		}
		if rows, err := h.deps.ListDomainVerifications(r.Context(), target.ID); err == nil {
			if len(rows) == 0 {
				rows = domains.NewService(h.deps).SeedDomains(r.Context(), target.ID, identity.DecodeStringSlice(target.VerifiedDomainsJSON), func() string {
					return domains.RandomTokenURL(12)
				})
			}
			data["DomainVerifications"] = rows
			data["VerifiedDomains"] = identity.DomainsToText(rows)
			data["ATProtoHandleVerified"] = identity.IsATProtoHandleVerified(target.ATProtoHandle, identity.VerifiedDomains(rows))
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
			aliases := ParseAliasesText(r.FormValue("aliases"))
			if err := identity.ValidateIdentifiers(r.Context(), target.Username, aliases, target.Email, target.ID, h.deps.Reserved(), h.deps.CheckIdentifierCollisions); err != nil {
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
			h.deps.AuditAttempt(r.Context(), current.ID, "domain.sync", target.Username, nil)
			_, verified, err := domains.NewService(h.deps).CreateDomains(r.Context(), target.ID, verifiedDomains, func() string {
				return domains.RandomTokenURL(12)
			})
			if err != nil {
				h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", target.Username, err, nil)
				http.Error(w, "Failed to update verified domains", http.StatusInternalServerError)
				return
			}
			h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", target.Username, nil, nil)

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
				target.PasswordHash = string(hash)
			}

			target.DisplayName = displayName
			target.Email = email
			target.Bio = bio
			target.Organization = strings.TrimSpace(r.FormValue("organization"))
			target.JobTitle = strings.TrimSpace(r.FormValue("job_title"))
			target.Birthdate = strings.TrimSpace(r.FormValue("birthdate"))
			target.Languages = strings.TrimSpace(r.FormValue("languages"))
			target.Phone = strings.TrimSpace(r.FormValue("phone"))
			target.Address = strings.TrimSpace(r.FormValue("address"))
			target.Location = strings.TrimSpace(r.FormValue("location"))
			target.Website = strings.TrimSpace(r.FormValue("website"))
			target.Pronouns = strings.TrimSpace(r.FormValue("pronouns"))
			target.Timezone = strings.TrimSpace(r.FormValue("timezone"))
			target.ATProtoHandle = strings.TrimSpace(r.FormValue("atproto_handle"))
			target.ATProtoDID = strings.TrimSpace(r.FormValue("atproto_did"))
			target.LinksJSON = identity.EncodeLinks(links)
			if customJSON, err := json.Marshal(customFields); err == nil {
				target.CustomFieldsJSON = string(customJSON)
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
			target.VisibilityJSON = identity.EncodeVisibilityMap(visibility)
			target.SocialProfilesJSON = identity.EncodeSocialProfiles(social)
			if walletsJSON, err := json.Marshal(identity.StripEmptyMap(wallets)); err == nil {
				target.WalletsJSON = string(walletsJSON)
			}
			if keysJSON, err := json.Marshal(identity.StripEmptyMap(publicKeys)); err == nil {
				target.PublicKeysJSON = string(keysJSON)
			}
			if domainsJSON, err := json.Marshal(verified); err == nil {
				target.VerifiedDomainsJSON = string(domainsJSON)
			}
			if aliasesJSON, err := json.Marshal(aliases); err == nil {
				target.AliasesJSON = string(aliasesJSON)
			}
			if role := strings.TrimSpace(r.FormValue("role")); target.Role != "owner" && (role == "admin" || role == "user") {
				target.Role = role
			}

			if err := h.deps.UpsertUserIdentifiers(r.Context(), target.ID, target.Username, aliases, target.Email); err != nil {
				http.Error(w, "Identifier already exists", http.StatusBadRequest)
				return
			}
			h.deps.AuditAttempt(r.Context(), current.ID, "user.update", target.Username, nil)
			if err := h.deps.UpdateUser(r.Context(), target); err != nil {
				h.deps.AuditOutcome(r.Context(), current.ID, "user.update", target.Username, err, nil)
				http.Error(w, "Failed to update profile", http.StatusInternalServerError)
				return
			}
			h.deps.AuditOutcome(r.Context(), current.ID, "user.update", target.Username, nil, nil)
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

func isAdmin(user domain.User) bool {
	return strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
}
