package admin

import (
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

	"pin/internal/domain"
	"pin/internal/features/domains"
	"pin/internal/features/identity"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/core"
	"pin/internal/platform/media"
)

func (h Handler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/settings/identity" {
		http.Redirect(w, r, "/settings/profile", http.StatusFound)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")

	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}
	if current.PrivateToken == "" {
		token := core.RandomTokenURL(32)
		if err := h.deps.UpdatePrivateToken(r.Context(), current.ID, token); err == nil {
			current.PrivateToken = token
		}
	}

	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), &current)
	showAppearanceNav := isAdmin(current) || settingsSvc.ServerThemePolicy(r.Context()).AllowUserTheme

	var links []domain.Link
	if current.LinksJSON != "" {
		_ = json.Unmarshal([]byte(current.LinksJSON), &links)
	}
	var socialProfiles []domain.SocialProfile
	if current.SocialProfilesJSON != "" {
		_ = json.Unmarshal([]byte(current.SocialProfilesJSON), &socialProfiles)
	}

	cfg := h.deps.Config()
	data := map[string]interface{}{
		"User":                   current,
		"Links":                  links,
		"SocialProfiles":         socialProfiles,
		"CustomFields":           identity.DecodeStringMap(current.CustomFieldsJSON),
		"FieldVisibility":        identity.DecodeStringMap(current.VisibilityJSON),
		"CustomFieldVisibility":  visibilityCustomMap(identity.DecodeStringMap(current.VisibilityJSON)),
		"Wallets":                identity.DecodeStringMap(current.WalletsJSON),
		"WalletEntries":          buildWalletEntries(identity.DecodeStringMap(current.WalletsJSON), identity.DecodeStringMap(current.VisibilityJSON)),
		"PublicKeys":             identity.DecodeStringMap(current.PublicKeysJSON),
		"VerifiedDomains":        verifiedDomainsToText(current.VerifiedDomainsJSON),
		"DomainVerifications":    []domain.DomainVerification{},
		"Aliases":                aliasesToText(current.AliasesJSON),
		"GitHubOAuthEnabled":     cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" && cfg.BaseURL != "",
		"RedditOAuthEnabled":     cfg.RedditClientID != "" && cfg.RedditClientSecret != "" && cfg.BaseURL != "",
		"BlueskyEnabled":         cfg.BlueskyPDS != "",
		"IsAdmin":                isAdmin(current),
		"IsSelf":                 true,
		"FormAction":             "/settings/profile",
		"ProfilePictures":        []domain.ProfilePicture{},
		"ActiveProfilePictureID": int64(0),
		"Title":                  "Settings - Profile",
		"Message":                "",
		"CSRFToken":              h.deps.EnsureCSRF(session),
		"PrivateIdentityURL":     h.deps.BaseURL(r) + "/p/" + url.PathEscape(core.ShortHash(strings.ToLower(current.Username), 7)) + "/" + url.PathEscape(current.PrivateToken),
		"Theme":                  theme,
		"ShowAppearanceNav":      showAppearanceNav,
		"ProtectedDomain":        h.deps.ProtectedDomain(r.Context()),
		"DomainVisibility":       domainVisibilityMap(identity.DecodeStringMap(current.VisibilityJSON)),
	}
	if toast := r.URL.Query().Get("toast"); toast != "" {
		data["Message"] = toast
	}
	if pics, err := h.deps.ListProfilePictures(r.Context(), current.ID); err == nil {
		data["ProfilePictures"] = pics
	}
	if current.ProfilePictureID.Valid {
		data["ActiveProfilePictureID"] = current.ProfilePictureID.Int64
	}
	rows, err := h.deps.ListDomainVerifications(r.Context(), current.ID)
	if err == nil && len(rows) == 0 {
		rows = domains.NewService(h.deps).SeedDomains(r.Context(), current.ID, identity.DecodeStringSlice(current.VerifiedDomainsJSON), func() string {
			return domains.RandomTokenURL(12)
		})
	}
	if err == nil {
		data["DomainVerifications"] = rows
		data["VerifiedDomains"] = identity.DomainsToText(rows)
		data["ATProtoHandleVerified"] = identity.IsATProtoHandleVerified(current.ATProtoHandle, identity.VerifiedDomains(rows))
	}

	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxUploadBytes)
		if err := r.ParseMultipartForm(cfg.MaxUploadBytes); err != nil {
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
		links = parseLinksForm(r.Form["link_label"], r.Form["link_url"], r.Form["link_visibility"])
		customFields := parseCustomFieldsForm(r.Form["custom_key"], r.Form["custom_value"])
		fieldVisibility := parseVisibilityForm(r.Form, []string{
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
		customVisibility := parseCustomVisibilityForm(r.Form["custom_key"], r.Form["custom_value"], r.Form["custom_visibility"])
		social := identity.MergeSocialProfiles(identity.ParseSocialForm(r.Form["social_label"], r.Form["social_url"], r.Form["social_visibility"]), socialProfiles)
		aliases := parseAliasesText(r.FormValue("aliases"))
		if err := identity.ValidateIdentifiers(r.Context(), current.Username, aliases, current.Email, current.ID, h.deps.Reserved(), h.deps.CheckIdentifierCollisions); err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "profile.update", current.Username, err, nil)
			data["Message"] = err.Error()
			goto renderProfile
		}
		wallets, walletVisibility, err := parseWalletForm(r.Form["wallet_label"], r.Form["wallet_address"], r.Form["wallet_visibility"])
		if err != nil {
			data["Message"] = err.Error()
			goto renderProfile
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
		verifiedDomains := parseVerifiedDomainsText(r.FormValue("verified_domains"))
		domainVisibility := parseVerifiedDomainVisibilityForm(r.Form["verified_domain"], r.Form["verified_domain_visibility"])
		h.deps.AuditAttempt(r.Context(), current.ID, "domain.sync", current.Username, nil)
		domainRows, verified, err := domains.NewService(h.deps).CreateDomains(r.Context(), current.ID, verifiedDomains, func() string {
			return domains.RandomTokenURL(12)
		})
		if err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", current.Username, err, nil)
			http.Error(w, "Failed to update verified domains", http.StatusInternalServerError)
			return
		}
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", current.Username, nil, nil)

		profilePictureFile, profilePictureHeader, err := r.FormFile("profile_picture")
		if err == nil && profilePictureHeader != nil && profilePictureHeader.Filename != "" {
			defer profilePictureFile.Close()
			ext := strings.ToLower(filepath.Ext(profilePictureHeader.Filename))
			if !cfg.AllowedExts[ext] {
				data["Message"] = "Profile picture must be an image (png/jpg/gif)"
			} else {
				if err := os.MkdirAll(cfg.ProfilePictureDir, 0755); err != nil {
					http.Error(w, "Failed to store profile picture", http.StatusInternalServerError)
					return
				}
				filename := fmt.Sprintf("profile_picture_%d.webp", time.Now().UTC().UnixNano())
				if err := media.WriteWebP(profilePictureFile, filepath.Join(cfg.ProfilePictureDir, filename)); err != nil {
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
				altText := strings.TrimSpace(r.FormValue("profile_picture_alt"))
				meta := map[string]string{"filename": filename}
				h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(current.ID), 10), meta)
				picID, err := h.deps.CreateProfilePicture(r.Context(), current.ID, filename, altText)
				if err != nil {
					h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(current.ID), 10), err, meta)
					http.Error(w, "Failed to save profile picture", http.StatusInternalServerError)
					return
				}
				h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(picID, 10), nil, meta)
				current.ProfilePictureID = sql.NullInt64{Int64: picID, Valid: true}
			}
		}

		current.DisplayName = displayName
		current.Email = email
		current.Bio = bio
		current.Organization = strings.TrimSpace(r.FormValue("organization"))
		current.JobTitle = strings.TrimSpace(r.FormValue("job_title"))
		current.Birthdate = strings.TrimSpace(r.FormValue("birthdate"))
		current.Languages = strings.TrimSpace(r.FormValue("languages"))
		current.Phone = strings.TrimSpace(r.FormValue("phone"))
		current.Address = strings.TrimSpace(r.FormValue("address"))
		current.Location = strings.TrimSpace(r.FormValue("location"))
		current.Website = strings.TrimSpace(r.FormValue("website"))
		current.Pronouns = strings.TrimSpace(r.FormValue("pronouns"))
		current.Timezone = strings.TrimSpace(r.FormValue("timezone"))
		current.ATProtoHandle = strings.TrimSpace(r.FormValue("atproto_handle"))
		current.ATProtoDID = strings.TrimSpace(r.FormValue("atproto_did"))
		if linksJSON, err := json.Marshal(links); err == nil {
			current.LinksJSON = string(linksJSON)
		}
		if customJSON, err := json.Marshal(customFields); err == nil {
			current.CustomFieldsJSON = string(customJSON)
		}
		visibility := buildVisibilityMap(fieldVisibility, filterCustomVisibility(customFields, customVisibility))
		for domain, vis := range domainVisibility {
			visibility["verified_domain:"+domain] = normalizeVisibility(vis)
		}
		if visibilityJSON, err := json.Marshal(visibility); err == nil {
			current.VisibilityJSON = string(visibilityJSON)
		}
		if socialJSON, err := json.Marshal(social); err == nil {
			current.SocialProfilesJSON = string(socialJSON)
		}
		if walletsJSON, err := json.Marshal(identity.StripEmptyMap(wallets)); err == nil {
			current.WalletsJSON = string(walletsJSON)
		}
		if keysJSON, err := json.Marshal(identity.StripEmptyMap(publicKeys)); err == nil {
			current.PublicKeysJSON = string(keysJSON)
		}
		if domainsJSON, err := json.Marshal(verified); err == nil {
			current.VerifiedDomainsJSON = string(domainsJSON)
		}
		if aliasesJSON, err := json.Marshal(aliases); err == nil {
			current.AliasesJSON = string(aliasesJSON)
		}

		if err := h.deps.UpsertUserIdentifiers(r.Context(), current.ID, current.Username, aliases, current.Email); err != nil {
			data["Message"] = "Identifier already exists"
			goto renderProfile
		}
		if err := h.deps.UpdateUser(r.Context(), current); err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "profile.update", current.Username, err, nil)
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}
		h.deps.AuditOutcome(r.Context(), current.ID, "profile.update", current.Username, nil, nil)

		data["Message"] = core.FirstNonEmpty(data["Message"].(string), "Profile updated successfully.")
		data["User"] = current
		data["Links"] = links
		data["CustomFields"] = customFields
		data["FieldVisibility"] = fieldVisibility
		data["CustomFieldVisibility"] = filterCustomVisibility(customFields, customVisibility)
		data["SocialProfiles"] = social
		data["Wallets"] = identity.DecodeStringMap(current.WalletsJSON)
		data["PublicKeys"] = identity.DecodeStringMap(current.PublicKeysJSON)
		data["DomainVerifications"] = domainRows
		data["VerifiedDomains"] = identity.DomainsToText(domainRows)
		data["ATProtoHandleVerified"] = identity.IsATProtoHandleVerified(current.ATProtoHandle, identity.VerifiedDomains(domainRows))
		data["DomainVisibility"] = domainVisibilityMap(identity.DecodeStringMap(current.VisibilityJSON))
		data["Aliases"] = aliasesToText(current.AliasesJSON)
	}

renderProfile:
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.deps.RenderTemplate(w, "settings_profile.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
