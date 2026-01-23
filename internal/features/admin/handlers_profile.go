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
	"pin/internal/features/users"
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
	currentIdentity, err := h.deps.CurrentIdentity(r)
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}
	if currentIdentity.PrivateToken == "" {
		token := core.RandomTokenURL(32)
		if err := h.deps.UpdateIdentityPrivateToken(r.Context(), currentIdentity.ID, token); err == nil {
			currentIdentity.PrivateToken = token
		}
	}

	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), &current)
	isAdminUser := isAdmin(current)
	showAppearanceNav := isAdminUser || settingsSvc.ServerThemePolicy(r.Context()).AllowUserTheme

	var links []domain.Link
	if currentIdentity.LinksJSON != "" {
		links = identity.DecodeLinks(currentIdentity.LinksJSON)
	}
	var socialProfiles []domain.SocialProfile
	if currentIdentity.SocialProfilesJSON != "" {
		socialProfiles = identity.DecodeSocialProfiles(currentIdentity.SocialProfilesJSON)
	}

	cfg := h.deps.Config()
	visibility := identity.DecodeVisibilityMap(currentIdentity.VisibilityJSON)
	wallets := identity.DecodeStringMap(currentIdentity.WalletsJSON)
	publicKeys := identity.DecodeStringMap(currentIdentity.PublicKeysJSON)
	message := ""
	userView := struct {
		domain.Identity
		Role string
	}{
		Identity: currentIdentity,
		Role:     current.Role,
	}
	data := map[string]interface{}{
		"User":                   userView,
		"Links":                  users.BuildLinkEntries(links, visibility),
		"SocialProfiles":         users.BuildSocialEntries(socialProfiles, visibility),
		"CustomFields":           identity.DecodeStringMap(currentIdentity.CustomFieldsJSON),
		"FieldVisibility":        visibility,
		"CustomFieldVisibility":  users.VisibilityCustomMap(visibility),
		"Wallets":                wallets,
		"WalletEntries":          users.BuildWalletEntries(wallets, visibility),
		"PublicKeys":             publicKeys,
		"VerifiedDomains":        users.VerifiedDomainsToText(currentIdentity.VerifiedDomainsJSON),
		"DomainVerifications":    []domain.DomainVerification{},
		"GitHubOAuthEnabled":     cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" && cfg.BaseURL != "",
		"RedditOAuthEnabled":     cfg.RedditClientID != "" && cfg.RedditClientSecret != "" && cfg.BaseURL != "",
		"BlueskyEnabled":         cfg.BlueskyPDS != "",
		"IsAdmin":                isAdminUser,
		"IsSelf":                 true,
		"FormAction":             "/settings/profile",
		"ProfilePictures":        []domain.ProfilePicture{},
		"ActiveProfilePictureID": int64(0),
		"Title":                  "Settings - Profile",
		"Message":                message,
		"CSRFToken":              h.deps.EnsureCSRF(session),
		"PrivateIdentityURL":     h.deps.BaseURL(r) + "/p/" + url.PathEscape(core.ShortHash(strings.ToLower(currentIdentity.Handle), 7)) + "/" + url.PathEscape(currentIdentity.PrivateToken),
		"Theme":                  theme,
		"ShowAppearanceNav":      showAppearanceNav,
		"ProtectedDomain":        h.deps.ProtectedDomain(r.Context()),
		"DomainVisibility":       users.DomainVisibilityMap(visibility),
	}
	if toast := r.URL.Query().Get("toast"); toast != "" {
		message = toast
		data["Message"] = message
	}
	if pics, err := h.deps.ListProfilePictures(r.Context(), currentIdentity.ID); err == nil {
		data["ProfilePictures"] = pics
	}
	if currentIdentity.ProfilePictureID.Valid {
		data["ActiveProfilePictureID"] = currentIdentity.ProfilePictureID.Int64
	}
	rows, err := h.deps.ListDomainVerifications(r.Context(), currentIdentity.ID)
	if err == nil && len(rows) == 0 {
		rows = domains.NewService(h.deps).SeedDomains(r.Context(), currentIdentity.ID, identity.DecodeStringSlice(currentIdentity.VerifiedDomainsJSON), func() string {
			return domains.RandomTokenURL(12)
		})
	}
	if err == nil {
		data["DomainVerifications"] = rows
		data["VerifiedDomains"] = identity.DomainsToText(rows)
		data["ATProtoHandleVerified"] = identity.IsATProtoHandleVerified(currentIdentity.ATProtoHandle, identity.VerifiedDomains(rows))
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

		handle := strings.TrimSpace(r.FormValue("handle"))
		displayName := strings.TrimSpace(r.FormValue("display_name"))
		email := strings.TrimSpace(r.FormValue("email"))
		bio := strings.TrimSpace(r.FormValue("bio"))
		links, linkVisibility := users.ParseLinksForm(r.Form["link_label"], r.Form["link_url"], r.Form["link_visibility"])
		customFields := users.ParseCustomFieldsForm(r.Form["custom_key"], r.Form["custom_value"])
		fieldVisibility := users.ParseVisibilityForm(r.Form, []string{
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
		customVisibility := users.ParseCustomVisibilityForm(r.Form["custom_key"], r.Form["custom_value"], r.Form["custom_visibility"])
		social, socialVisibility := identity.ParseSocialForm(r.Form["social_label"], r.Form["social_url"], r.Form["social_visibility"])
		social = identity.MergeSocialProfiles(social, socialProfiles)
		if err := identity.ValidateHandle(r.Context(), handle, currentIdentity.ID, h.deps.Reserved(), h.deps.CheckHandleCollision); err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "profile.update", currentIdentity.Handle, err, nil)
			data["Message"] = err.Error()
			goto renderProfile
		}
		wallets, walletVisibility, err := users.ParseWalletForm(r.Form["wallet_label"], r.Form["wallet_address"], r.Form["wallet_visibility"])
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
		verifiedDomains := users.ParseVerifiedDomainsText(r.FormValue("verified_domains"))
		domainVisibility := users.ParseVerifiedDomainVisibilityForm(r.Form["verified_domain"], r.Form["verified_domain_visibility"])
		h.deps.AuditAttempt(r.Context(), current.ID, "domain.sync", currentIdentity.Handle, nil)
		domainRows, verified, err := domains.NewService(h.deps).CreateDomains(r.Context(), currentIdentity.ID, verifiedDomains, func() string {
			return domains.RandomTokenURL(12)
		})
		if err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", currentIdentity.Handle, err, nil)
			http.Error(w, "Failed to update verified domains", http.StatusInternalServerError)
			return
		}
		h.deps.AuditOutcome(r.Context(), current.ID, "domain.sync", currentIdentity.Handle, nil, nil)

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
				h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(currentIdentity.ID), 10), meta)
				picID, err := h.deps.CreateProfilePicture(r.Context(), currentIdentity.ID, filename, altText)
				if err != nil {
					h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(currentIdentity.ID), 10), err, meta)
					http.Error(w, "Failed to save profile picture", http.StatusInternalServerError)
					return
				}
				h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(picID, 10), nil, meta)
				currentIdentity.ProfilePictureID = sql.NullInt64{Int64: picID, Valid: true}
			}
		}

		currentIdentity.Handle = handle
		currentIdentity.DisplayName = displayName
		currentIdentity.Email = email
		currentIdentity.Bio = bio
		currentIdentity.Organization = strings.TrimSpace(r.FormValue("organization"))
		currentIdentity.JobTitle = strings.TrimSpace(r.FormValue("job_title"))
		currentIdentity.Birthdate = strings.TrimSpace(r.FormValue("birthdate"))
		currentIdentity.Languages = strings.TrimSpace(r.FormValue("languages"))
		currentIdentity.Phone = strings.TrimSpace(r.FormValue("phone"))
		currentIdentity.Address = strings.TrimSpace(r.FormValue("address"))
		currentIdentity.Location = strings.TrimSpace(r.FormValue("location"))
		currentIdentity.Website = strings.TrimSpace(r.FormValue("website"))
		currentIdentity.Pronouns = strings.TrimSpace(r.FormValue("pronouns"))
		currentIdentity.Timezone = strings.TrimSpace(r.FormValue("timezone"))
		currentIdentity.ATProtoHandle = strings.TrimSpace(r.FormValue("atproto_handle"))
		currentIdentity.ATProtoDID = strings.TrimSpace(r.FormValue("atproto_did"))
		currentIdentity.LinksJSON = identity.EncodeLinks(links)
		if customJSON, err := json.Marshal(customFields); err == nil {
			currentIdentity.CustomFieldsJSON = string(customJSON)
		}
		visibility := users.BuildVisibilityMap(fieldVisibility, users.FilterCustomVisibility(customFields, customVisibility))
		for domain, vis := range domainVisibility {
			visibility["verified_domain:"+domain] = users.NormalizeVisibility(vis)
		}
		for key, value := range linkVisibility {
			visibility[key] = value
		}
		for key, value := range socialVisibility {
			visibility[key] = value
		}
		currentIdentity.VisibilityJSON = identity.EncodeVisibilityMap(visibility)
		currentIdentity.SocialProfilesJSON = identity.EncodeSocialProfiles(social)
		if walletsJSON, err := json.Marshal(identity.StripEmptyMap(wallets)); err == nil {
			currentIdentity.WalletsJSON = string(walletsJSON)
		}
		if keysJSON, err := json.Marshal(identity.StripEmptyMap(publicKeys)); err == nil {
			currentIdentity.PublicKeysJSON = string(keysJSON)
		}
		if domainsJSON, err := json.Marshal(verified); err == nil {
			currentIdentity.VerifiedDomainsJSON = string(domainsJSON)
		}
		if err := h.deps.UpdateIdentity(r.Context(), currentIdentity); err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "profile.update", currentIdentity.Handle, err, nil)
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}
		h.deps.AuditOutcome(r.Context(), current.ID, "profile.update", currentIdentity.Handle, nil, nil)

		data["Message"] = core.FirstNonEmpty(data["Message"].(string), "Profile updated successfully.")
		data["User"] = current
		data["Links"] = users.BuildLinkEntries(links, visibility)
		data["CustomFields"] = customFields
		data["FieldVisibility"] = fieldVisibility
		data["CustomFieldVisibility"] = users.FilterCustomVisibility(customFields, customVisibility)
		data["SocialProfiles"] = users.BuildSocialEntries(social, visibility)
		data["Wallets"] = identity.DecodeStringMap(currentIdentity.WalletsJSON)
		data["PublicKeys"] = identity.DecodeStringMap(currentIdentity.PublicKeysJSON)
		data["DomainVerifications"] = domainRows
		data["VerifiedDomains"] = identity.DomainsToText(domainRows)
		data["ATProtoHandleVerified"] = identity.IsATProtoHandleVerified(currentIdentity.ATProtoHandle, identity.VerifiedDomains(domainRows))
		data["DomainVisibility"] = users.DomainVisibilityMap(identity.DecodeVisibilityMap(currentIdentity.VisibilityJSON))
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
