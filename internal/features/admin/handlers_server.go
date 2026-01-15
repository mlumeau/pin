package admin

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"pin/internal/domain"
	"pin/internal/features/domains"
	"pin/internal/features/identity"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/media"
)

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
	var users []domain.User
	var invites []domain.Invite
	var auditLogs []domain.AuditLog
	usedInvitesCount := 0
	defaultTheme := featuresettings.DefaultThemeName
	defaultThemeForce := false
	themeValue, ok, force := settingsSvc.ServerDefaultTheme(r.Context())
	defaultThemeForce = force
	if ok {
		defaultTheme = themeValue
	}
	defaultCustomCSSPath, _ := settingsSvc.ServerDefaultCustomCSS(r.Context())
	themePolicy := settingsSvc.ServerThemePolicy(r.Context())
	showAppearanceNav := isAdminUser || themePolicy.AllowUserTheme
	userPage := 1
	userPrevPage := 1
	userNextPage := 1
	userTotalPages := 1
	userQuery := ""
	userSort := "id"
	userDir := "desc"
	userDirOverride := false
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
		action := strings.TrimSpace(r.FormValue("action"))
		switch action {
		case "export-users":
			users, err := h.deps.ListUsers(r.Context())
			if err != nil {
				http.Error(w, "Failed to export users", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
			w.Header().Set("Content-Disposition", "attachment; filename=users.csv")
			writer := csv.NewWriter(w)
			_ = writer.Write([]string{"id", "username", "email", "role", "updated_at"})
			for _, user := range users {
				_ = writer.Write([]string{
					strconv.Itoa(user.ID),
					user.Username,
					user.Email,
					user.Role,
					user.UpdatedAt.Format(time.RFC3339),
				})
			}
			writer.Flush()
			return
		case "set-default-theme":
			if defaultCustomCSSPath != "" && r.FormValue("theme_default_force") == "on" {
				defaultThemeForce = true
			} else {
				defaultThemeForce = false
			}
			themeValue := featuresettings.NormalizeThemeChoice(r.FormValue("theme_default"))
			if themeValue == featuresettings.DefaultCustomThemeName && defaultCustomCSSPath == "" {
				themeValue = featuresettings.DefaultThemeName
			}
			if err := settingsSvc.SaveServerDefaultTheme(r.Context(), themeValue, defaultThemeForce); err != nil {
				http.Error(w, "Failed to save default theme", http.StatusInternalServerError)
				return
			}
			if defaultThemeForce && themeValue != "" {
				if err := h.deps.ResetAllUserThemes(r.Context(), themeValue); err != nil {
					http.Error(w, "Failed to reset user themes", http.StatusInternalServerError)
					return
				}
				message = "Default theme saved and user themes reset."
			} else {
				message = "Default theme saved."
			}
		case "set-theme-policy":
			policy := featuresettings.ThemePolicy{
				AllowUserTheme:     r.FormValue("theme_user_select") == "on",
				AllowUserCustomCSS: r.FormValue("theme_user_custom_css") == "on",
			}
			if err := settingsSvc.SaveServerThemePolicy(r.Context(), policy); err != nil {
				http.Error(w, "Failed to save theme policy", http.StatusInternalServerError)
				return
			}
			message = "Theme policy saved."
		case "upload-default-css":
			file, header, err := r.FormFile("default_css_file")
			if err != nil || header == nil || header.Filename == "" {
				http.Error(w, "Missing CSS file", http.StatusBadRequest)
				return
			}
			defer file.Close()
			ext := strings.ToLower(filepath.Ext(header.Filename))
			if ext != ".css" {
				http.Error(w, "Default CSS must be a .css file", http.StatusBadRequest)
				return
			}
			if err := os.MkdirAll(featuresettings.ThemeDir(h.deps.Config()), 0755); err != nil {
				http.Error(w, "Failed to store CSS", http.StatusInternalServerError)
				return
			}
			filename := fmt.Sprintf("theme_default_%d.css", time.Now().UTC().UnixNano())
			destPath := filepath.Join(featuresettings.ThemeDir(h.deps.Config()), filename)
			dest, err := os.Create(destPath)
			if err != nil {
				http.Error(w, "Failed to store CSS", http.StatusInternalServerError)
				return
			}
			if _, err := io.Copy(dest, io.LimitReader(file, h.deps.Config().MaxUploadBytes)); err != nil {
				_ = dest.Close()
				_ = os.Remove(destPath)
				http.Error(w, "Failed to store CSS", http.StatusInternalServerError)
				return
			}
			_ = dest.Close()
			if prev := strings.TrimSpace(defaultCustomCSSPath); prev != "" && prev != filename {
				_ = os.Remove(filepath.Join(featuresettings.ThemeDir(h.deps.Config()), prev))
			}
			if err := settingsSvc.SaveServerDefaultCustomCSS(r.Context(), filename); err != nil {
				http.Error(w, "Failed to save CSS", http.StatusInternalServerError)
				return
			}
			message = "Default CSS uploaded."
		case "delete-default-css":
			if defaultCustomCSSPath != "" {
				_ = os.Remove(filepath.Join(featuresettings.ThemeDir(h.deps.Config()), defaultCustomCSSPath))
			}
			_ = settingsSvc.SaveServerDefaultCustomCSS(r.Context(), "")
			message = "Default CSS removed."
		}
	}

	userQuery = strings.TrimSpace(r.URL.Query().Get("q"))
	if dir := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dir"))); dir == "asc" || dir == "desc" {
		userDir = dir
		userDirOverride = true
	}
	if sort := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort"))); sort == "username" || sort == "email" || sort == "role" || sort == "updated" {
		userSort = sort
		if !userDirOverride {
			userDir = "asc"
		}
	}
	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		userPage = page
	}
	users, total, err := h.deps.ListUsersPaged(r.Context(), userQuery, userSort, userDir, 10, (userPage-1)*10)
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}
	if total > 10 {
		userTotalPages = (total + 9) / 10
		userNextPage = userPage + 1
		if userNextPage > userTotalPages {
			userNextPage = userTotalPages
		}
		userPrevPage = userPage - 1
		if userPrevPage < 1 {
			userPrevPage = 1
		}
	}

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

	auditPage = 1
	if p, err := strconv.Atoi(r.URL.Query().Get("audit_page")); err == nil && p > 0 {
		auditPage = p
	}
	logs, err := h.deps.ListAuditLogs(r.Context(), 20, (auditPage-1)*20)
	if err != nil {
		http.Error(w, "Failed to load audit log", http.StatusInternalServerError)
		return
	}
	auditLogs = logs
	auditTotal, _ := h.deps.CountAuditLogs(r.Context())
	if auditTotal > 20 {
		auditTotalPages = (auditTotal + 19) / 20
		auditNextPage = auditPage + 1
		if auditNextPage > auditTotalPages {
			auditNextPage = auditTotalPages
		}
		auditPrevPage = auditPage - 1
		if auditPrevPage < 1 {
			auditPrevPage = 1
		}
		auditHasMore = auditPage < auditTotalPages
	}

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":                 current,
		"IsAdmin":              isAdmin(current),
		"Users":                users,
		"UserQuery":            userQuery,
		"UserSort":             userSort,
		"UserDir":              userDir,
		"UserPage":             userPage,
		"UserPrevPage":         userPrevPage,
		"UserNextPage":         userNextPage,
		"UserTotalPages":       userTotalPages,
		"UserDirOverride":      userDirOverride,
		"Invites":              invites,
		"InvitesUsedCount":     usedInvitesCount,
		"AuditLogs":            auditLogs,
		"AuditPage":            auditPage,
		"AuditPrevPage":        auditPrevPage,
		"AuditNextPage":        auditNextPage,
		"AuditHasMore":         auditHasMore,
		"AuditTotalPages":      auditTotalPages,
		"Theme":                theme,
		"Landing":              landing,
		"CSRFToken":            h.deps.EnsureCSRF(session),
		"Message":              message,
		"DefaultTheme":         defaultTheme,
		"DefaultThemeForce":    defaultThemeForce,
		"DefaultCustomCSS":     featuresettings.ThemeCustomCSSURL(defaultCustomCSSPath),
		"DefaultCustomCSSName": filepath.Base(defaultCustomCSSPath),
		"ThemePolicy":          themePolicy,
		"ShowAppearanceNav":    showAppearanceNav,
	}

	if err := h.deps.RenderTemplate(w, "settings_admin_server.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (h Handler) AuditLogDownload(w http.ResponseWriter, r *http.Request) {
	current, err := h.deps.CurrentUser(r)
	if err != nil || !isAdmin(current) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if format == "" {
		format = "csv"
	}
	if scope == "" {
		scope = "page"
	}

	var logs []domain.AuditLog
	if scope == "all" {
		logs, err = h.deps.ListAllAuditLogs(r.Context())
	} else {
		page := 1
		if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
			page = p
		}
		const auditPageSize = 20
		offset := (page - 1) * auditPageSize
		logs, err = h.deps.ListAuditLogs(r.Context(), auditPageSize, offset)
	}
	if err != nil {
		http.Error(w, "Failed to load audit log", http.StatusInternalServerError)
		return
	}

	filename := "audit-log"
	if scope == "all" {
		filename += "-all"
	} else {
		filename += "-page"
	}
	filename += "." + format

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		_ = json.NewEncoder(w).Encode(logs)
	case "txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		for _, logEntry := range logs {
			actor := logEntry.ActorName
			if actor == "" {
				actor = "system"
			}
			target := logEntry.Target
			if target == "" {
				target = "n/a"
			}
			fmt.Fprintf(
				w,
				"%s | %s | by %s | object %s | log #%d\n",
				logEntry.CreatedAt.Format("2006-01-02 15:04:05"),
				logEntry.Action,
				actor,
				target,
				logEntry.ID,
			)
		}
	default:
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		writer := csv.NewWriter(w)
		_ = writer.Write([]string{"id", "timestamp", "action", "actor", "target", "metadata"})
		for _, logEntry := range logs {
			actor := logEntry.ActorName
			if actor == "" {
				actor = "system"
			}
			_ = writer.Write([]string{
				strconv.Itoa(logEntry.ID),
				logEntry.CreatedAt.Format(time.RFC3339),
				logEntry.Action,
				actor,
				logEntry.Target,
				logEntry.Metadata,
			})
		}
		writer.Flush()
	}
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
			_ = json.Unmarshal([]byte(target.LinksJSON), &links)
		}
		var socialProfiles []domain.SocialProfile
		if target.SocialProfilesJSON != "" {
			_ = json.Unmarshal([]byte(target.SocialProfilesJSON), &socialProfiles)
		}

		settingsSvc := featuresettings.NewService(h.deps)
		theme := settingsSvc.ThemeSettings(r.Context(), &current)
		showAppearanceNav := isAdmin(current)
		data := map[string]interface{}{
			"User":                  target,
			"Links":                 links,
			"SocialProfiles":        socialProfiles,
			"CustomFields":          identity.DecodeStringMap(target.CustomFieldsJSON),
			"FieldVisibility":       identity.DecodeStringMap(target.VisibilityJSON),
			"CustomFieldVisibility": visibilityCustomMap(identity.DecodeStringMap(target.VisibilityJSON)),
			"Wallets":               identity.DecodeStringMap(target.WalletsJSON),
			"WalletEntries":         buildWalletEntries(identity.DecodeStringMap(target.WalletsJSON), identity.DecodeStringMap(target.VisibilityJSON)),
			"PublicKeys":            identity.DecodeStringMap(target.PublicKeysJSON),
			"VerifiedDomains":       verifiedDomainsToText(target.VerifiedDomainsJSON),
			"DomainVerifications":   []domain.DomainVerification{},
			"Aliases":               aliasesToText(target.AliasesJSON),
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
			"DomainVisibility":      domainVisibilityMap(identity.DecodeStringMap(target.VisibilityJSON)),
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
			links := parseLinksForm(r.Form["link_label"], r.Form["link_url"], r.Form["link_visibility"])
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
			if err := identity.ValidateIdentifiers(r.Context(), target.Username, aliases, target.Email, target.ID, h.deps.Reserved(), h.deps.CheckIdentifierCollisions); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			wallets, walletVisibility, err := parseWalletForm(r.Form["wallet_label"], r.Form["wallet_address"], r.Form["wallet_visibility"])
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
			verifiedDomains := parseVerifiedDomainsText(r.FormValue("verified_domains"))
			domainVisibility := parseVerifiedDomainVisibilityForm(r.Form["verified_domain"], r.Form["verified_domain_visibility"])
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
			if linksJSON, err := json.Marshal(links); err == nil {
				target.LinksJSON = string(linksJSON)
			}
			if customJSON, err := json.Marshal(customFields); err == nil {
				target.CustomFieldsJSON = string(customJSON)
			}
			visibility := buildVisibilityMap(fieldVisibility, filterCustomVisibility(customFields, customVisibility))
			for domain, vis := range domainVisibility {
				visibility["verified_domain:"+domain] = normalizeVisibility(vis)
			}
			if visibilityJSON, err := json.Marshal(visibility); err == nil {
				target.VisibilityJSON = string(visibilityJSON)
			}
			if socialJSON, err := json.Marshal(social); err == nil {
				target.SocialProfilesJSON = string(socialJSON)
			}
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
