package admin

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

	"pin/internal/config"
	"pin/internal/domain"
	"pin/internal/features/domains"
	"pin/internal/features/identity"
	featuresettings "pin/internal/features/settings"
	"pin/internal/features/users"
	"pin/internal/platform/core"
	"pin/internal/platform/media"
)

// profileFormData captures validated form inputs for profile updates.
type profileFormData struct {
	handle           string
	displayName      string
	email            string
	bio              string
	organization     string
	jobTitle         string
	birthdate        string
	languages        string
	phone            string
	address          string
	location         string
	website          string
	pronouns         string
	timezone         string
	atprotoHandle    string
	atprotoDID       string
	links            []domain.Link
	linkVisibility   map[string]string
	customFields     map[string]string
	fieldVisibility  map[string]string
	customVisibility map[string]string
	social           []domain.SocialProfile
	socialVisibility map[string]string
	wallets          map[string]string
	walletVisibility map[string]string
	publicKeys       map[string]string
	verifiedDomains  []string
	domainVisibility map[string]string
}

// uploadResult describes the outcome of a profile picture upload.
type uploadResult struct {
	message   string
	pictureID sql.NullInt64
}

// prioritizeActiveProfilePicture returns a copy with the active picture first.
func prioritizeActiveProfilePicture(pics []domain.ProfilePicture, activeID int64) []domain.ProfilePicture {
	if len(pics) < 2 || activeID <= 0 {
		return pics
	}
	activeIndex := -1
	for i := range pics {
		if pics[i].ID == activeID {
			activeIndex = i
			break
		}
	}
	if activeIndex <= 0 {
		return pics
	}
	ordered := make([]domain.ProfilePicture, 0, len(pics))
	ordered = append(ordered, pics[activeIndex])
	ordered = append(ordered, pics[:activeIndex]...)
	ordered = append(ordered, pics[activeIndex+1:]...)
	return ordered
}

// Profile handles the HTTP request.
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
		"SectionTitle":           "Profile",
		"Message":                message,
		"CSRFToken":              h.deps.EnsureCSRF(session),
		"PrivateIdentityURL":     h.deps.BaseURL(r) + "/p/" + url.PathEscape(core.ShortHash(strings.ToLower(currentIdentity.Handle), 7)) + "/" + url.PathEscape(currentIdentity.PrivateToken),
		"Theme":                  theme,
		"ShowAppearanceNav":      showAppearanceNav,
		"ProtectedDomain":        h.deps.ProtectedDomain(r.Context()),
		"DomainVisibility":       users.DomainVisibilityMap(visibility),
	}
	data["Preview"] = buildProfilePreviewData(currentIdentity, theme)
	if toast := r.URL.Query().Get("toast"); toast != "" {
		message = toast
		data["Message"] = message
	}
	if pics, err := h.deps.ListProfilePictures(r.Context(), currentIdentity.ID); err == nil {
		if currentIdentity.ProfilePictureID.Valid {
			pics = prioritizeActiveProfilePicture(pics, currentIdentity.ProfilePictureID.Int64)
		}
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

		form, err := h.parseProfileForm(r, currentIdentity, socialProfiles)
		if err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "profile.update", currentIdentity.Handle, err, nil)
			data["Message"] = err.Error()
			goto renderProfile
		}
		domainRows, verified, err := h.syncProfileDomains(r.Context(), current, currentIdentity, form.verifiedDomains)
		if err != nil {
			http.Error(w, "Failed to update verified domains", http.StatusInternalServerError)
			return
		}
		upload, err := h.uploadProfilePicture(r, cfg, current, currentIdentity.ID)
		if err != nil {
			if reqErr, ok := err.(requestError); ok {
				http.Error(w, reqErr.Error(), reqErr.status)
				return
			}
			http.Error(w, "Failed to process profile picture", http.StatusInternalServerError)
			return
		}
		if upload.message != "" {
			data["Message"] = upload.message
		}
		if upload.pictureID.Valid {
			currentIdentity.ProfilePictureID = upload.pictureID
		}

		currentIdentity = applyProfileForm(currentIdentity, form)
		visibility := buildProfileVisibility(form)
		currentIdentity.VisibilityJSON = identity.EncodeVisibilityMap(visibility)
		currentIdentity.SocialProfilesJSON = identity.EncodeSocialProfiles(form.social)
		if walletsJSON, err := json.Marshal(identity.StripEmptyMap(form.wallets)); err == nil {
			currentIdentity.WalletsJSON = string(walletsJSON)
		}
		if keysJSON, err := json.Marshal(identity.StripEmptyMap(form.publicKeys)); err == nil {
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
		data["Links"] = users.BuildLinkEntries(form.links, visibility)
		data["CustomFields"] = form.customFields
		data["FieldVisibility"] = form.fieldVisibility
		data["CustomFieldVisibility"] = users.FilterCustomVisibility(form.customFields, form.customVisibility)
		data["SocialProfiles"] = users.BuildSocialEntries(form.social, visibility)
		data["Wallets"] = identity.DecodeStringMap(currentIdentity.WalletsJSON)
		data["PublicKeys"] = identity.DecodeStringMap(currentIdentity.PublicKeysJSON)
		data["DomainVerifications"] = domainRows
		data["VerifiedDomains"] = identity.DomainsToText(domainRows)
		data["ATProtoHandleVerified"] = identity.IsATProtoHandleVerified(currentIdentity.ATProtoHandle, identity.VerifiedDomains(domainRows))
		data["DomainVisibility"] = users.DomainVisibilityMap(identity.DecodeVisibilityMap(currentIdentity.VisibilityJSON))
		data["Preview"] = buildProfilePreviewData(currentIdentity, theme)
	}

renderProfile:
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Template execution may have already written headers; avoid double WriteHeader.
	_ = h.deps.RenderTemplate(w, "settings_profile.html", data)
}

// parseProfileForm extracts form values and validates the handle.
func (h Handler) parseProfileForm(r *http.Request, currentIdentity domain.Identity, socialProfiles []domain.SocialProfile) (profileFormData, error) {
	form := profileFormData{
		handle:        strings.TrimSpace(r.FormValue("handle")),
		displayName:   strings.TrimSpace(r.FormValue("display_name")),
		email:         strings.TrimSpace(r.FormValue("email")),
		bio:           strings.TrimSpace(r.FormValue("bio")),
		organization:  strings.TrimSpace(r.FormValue("organization")),
		jobTitle:      strings.TrimSpace(r.FormValue("job_title")),
		birthdate:     strings.TrimSpace(r.FormValue("birthdate")),
		languages:     strings.TrimSpace(r.FormValue("languages")),
		phone:         strings.TrimSpace(r.FormValue("phone")),
		address:       strings.TrimSpace(r.FormValue("address")),
		location:      strings.TrimSpace(r.FormValue("location")),
		website:       strings.TrimSpace(r.FormValue("website")),
		pronouns:      strings.TrimSpace(r.FormValue("pronouns")),
		timezone:      strings.TrimSpace(r.FormValue("timezone")),
		atprotoHandle: strings.TrimSpace(r.FormValue("atproto_handle")),
		atprotoDID:    strings.TrimSpace(r.FormValue("atproto_did")),
	}
	form.links, form.linkVisibility = users.ParseLinksForm(r.Form["link_label"], r.Form["link_url"], r.Form["link_visibility"])
	form.customFields = users.ParseCustomFieldsForm(r.Form["custom_key"], r.Form["custom_value"])
	form.fieldVisibility = users.ParseVisibilityForm(r.Form, []string{
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
	form.customVisibility = users.ParseCustomVisibilityForm(r.Form["custom_key"], r.Form["custom_value"], r.Form["custom_visibility"])
	form.social, form.socialVisibility = identity.ParseSocialForm(r.Form["social_label"], r.Form["social_url"], r.Form["social_visibility"])
	form.social = identity.MergeSocialProfiles(form.social, socialProfiles)

	if err := identity.ValidateHandle(r.Context(), form.handle, currentIdentity.ID, h.deps.Reserved(), h.deps.CheckHandleCollision); err != nil {
		return profileFormData{}, err
	}
	var err error
	form.wallets, form.walletVisibility, err = users.ParseWalletForm(r.Form["wallet_label"], r.Form["wallet_address"], r.Form["wallet_visibility"])
	if err != nil {
		return profileFormData{}, err
	}
	for key, value := range form.walletVisibility {
		form.fieldVisibility[key] = value
	}
	form.publicKeys = map[string]string{
		"pgp":         strings.TrimSpace(r.FormValue("key_pgp")),
		"ssh":         strings.TrimSpace(r.FormValue("key_ssh")),
		"age":         strings.TrimSpace(r.FormValue("key_age")),
		"activitypub": strings.TrimSpace(r.FormValue("key_activitypub")),
	}
	form.verifiedDomains = users.ParseVerifiedDomainsText(r.FormValue("verified_domains"))
	form.domainVisibility = users.ParseVerifiedDomainVisibilityForm(r.Form["verified_domain"], r.Form["verified_domain_visibility"])
	return form, nil
}

// syncProfileDomains reconciles verified domains with user input.
func (h Handler) syncProfileDomains(ctx context.Context, current domain.User, currentIdentity domain.Identity, verifiedDomains []string) ([]domain.DomainVerification, []string, error) {
	h.deps.AuditAttempt(ctx, current.ID, "domain.sync", currentIdentity.Handle, nil)
	domainRows, verified, err := domains.NewService(h.deps).CreateDomains(ctx, currentIdentity.ID, verifiedDomains, func() string {
		return domains.RandomTokenURL(12)
	})
	if err != nil {
		h.deps.AuditOutcome(ctx, current.ID, "domain.sync", currentIdentity.Handle, err, nil)
		return nil, nil, err
	}
	h.deps.AuditOutcome(ctx, current.ID, "domain.sync", currentIdentity.Handle, nil, nil)
	return domainRows, verified, nil
}

// uploadProfilePicture stores a new profile picture and returns the new ID when present.
func (h Handler) uploadProfilePicture(r *http.Request, cfg config.Config, current domain.User, identityID int) (uploadResult, error) {
	file, header, err := r.FormFile("profile_picture")
	if err != nil || header == nil || header.Filename == "" {
		return uploadResult{}, nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !cfg.AllowedExts[ext] {
		return uploadResult{message: "Profile picture must be an image (png/jpg/gif)"}, nil
	}
	if err := os.MkdirAll(cfg.ProfilePictureDir, 0755); err != nil {
		return uploadResult{}, err
	}
	filename := fmt.Sprintf("profile_picture_%d.webp", time.Now().UTC().UnixNano())
	if err := media.WriteWebP(file, filepath.Join(cfg.ProfilePictureDir, filename)); err != nil {
		switch {
		case errors.Is(err, media.ErrCWebPUnavailable):
			return uploadResult{}, requestError{err: errors.New("WebP encoder unavailable"), status: http.StatusServiceUnavailable}
		case errors.Is(err, media.ErrImageTooSmall):
			return uploadResult{}, requestError{err: errors.New("Image too small"), status: http.StatusBadRequest}
		default:
			return uploadResult{}, err
		}
	}
	altText := strings.TrimSpace(r.FormValue("profile_picture_alt"))
	meta := map[string]string{"filename": filename}
	h.deps.AuditAttempt(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(identityID), 10), meta)
	picID, err := h.deps.CreateProfilePicture(r.Context(), identityID, filename, altText)
	if err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(int64(identityID), 10), err, meta)
		return uploadResult{}, err
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "profile_picture.upload", strconv.FormatInt(picID, 10), nil, meta)
	return uploadResult{pictureID: sql.NullInt64{Int64: picID, Valid: true}}, nil
}

// applyProfileForm copies form fields into the identity record.
func applyProfileForm(identityRecord domain.Identity, form profileFormData) domain.Identity {
	identityRecord.Handle = form.handle
	identityRecord.DisplayName = form.displayName
	identityRecord.Email = form.email
	identityRecord.Bio = form.bio
	identityRecord.Organization = form.organization
	identityRecord.JobTitle = form.jobTitle
	identityRecord.Birthdate = form.birthdate
	identityRecord.Languages = form.languages
	identityRecord.Phone = form.phone
	identityRecord.Address = form.address
	identityRecord.Location = form.location
	identityRecord.Website = form.website
	identityRecord.Pronouns = form.pronouns
	identityRecord.Timezone = form.timezone
	identityRecord.ATProtoHandle = form.atprotoHandle
	identityRecord.ATProtoDID = form.atprotoDID
	identityRecord.LinksJSON = identity.EncodeLinks(form.links)
	if customJSON, err := json.Marshal(form.customFields); err == nil {
		identityRecord.CustomFieldsJSON = string(customJSON)
	}
	return identityRecord
}

// buildProfileVisibility merges core, custom, link, social, and domain visibility values.
func buildProfileVisibility(form profileFormData) map[string]string {
	visibility := users.BuildVisibilityMap(form.fieldVisibility, users.FilterCustomVisibility(form.customFields, form.customVisibility))
	for domain, vis := range form.domainVisibility {
		visibility["verified_domain:"+domain] = users.NormalizeVisibility(vis)
	}
	for key, value := range form.linkVisibility {
		visibility[key] = value
	}
	for key, value := range form.socialVisibility {
		visibility[key] = value
	}
	return visibility
}
