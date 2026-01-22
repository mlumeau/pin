package admin

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pin/internal/domain"
	featuresettings "pin/internal/features/settings"
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
		action := strings.TrimSpace(r.FormValue("server_action"))
		if action == "" {
			action = strings.TrimSpace(r.FormValue("action"))
		}
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
		case "landing":
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
					http.Error(w, "Failed to save landing page", http.StatusInternalServerError)
					return
				}
				message = "Custom landing page removed."
			} else if file, header, err := r.FormFile("landing_custom_file"); err == nil && header != nil && header.Filename != "" {
				defer file.Close()
				ext := strings.ToLower(filepath.Ext(header.Filename))
				if ext != ".html" && ext != ".htm" {
					http.Error(w, "Landing page must be an HTML file", http.StatusBadRequest)
					return
				}
				if err := os.MkdirAll(featuresettings.LandingDir(h.deps.Config()), 0755); err != nil {
					http.Error(w, "Failed to store landing page", http.StatusInternalServerError)
					return
				}
				filename := fmt.Sprintf("landing_custom_%d.html", time.Now().UTC().UnixNano())
				destPath := filepath.Join(featuresettings.LandingDir(h.deps.Config()), filename)
				dest, err := os.Create(destPath)
				if err != nil {
					http.Error(w, "Failed to store landing page", http.StatusInternalServerError)
					return
				}
				if _, err := io.Copy(dest, io.LimitReader(file, h.deps.Config().MaxUploadBytes)); err != nil {
					_ = dest.Close()
					_ = os.Remove(destPath)
					http.Error(w, "Failed to store landing page", http.StatusInternalServerError)
					return
				}
				_ = dest.Close()
				if prev := strings.TrimSpace(landing.CustomPath); prev != "" && prev != filename {
					_ = os.Remove(filepath.Join(featuresettings.LandingDir(h.deps.Config()), prev))
				}
				if err := settingsSvc.SaveLandingSettings(r.Context(), featuresettings.LandingSettings{Mode: landingMode, CustomPath: filename}); err != nil {
					http.Error(w, "Failed to save landing page", http.StatusInternalServerError)
					return
				}
				message = "Landing page updated."
			} else {
				if err := settingsSvc.SaveLandingSettings(r.Context(), featuresettings.LandingSettings{Mode: landingMode, CustomPath: landing.CustomPath}); err != nil {
					http.Error(w, "Failed to save landing page", http.StatusInternalServerError)
					return
				}
				message = "Landing page saved."
			}
		case "theme_default":
			defaultThemeForce = r.FormValue("default_theme_force") == "1"
			themeValue := featuresettings.NormalizeThemeChoice(r.FormValue("default_theme"))
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
			http.Redirect(w, r, "/settings/admin/server?toast="+url.QueryEscape(message)+"#section-theme", http.StatusSeeOther)
			return
		case "theme_access":
			policy := featuresettings.ThemePolicy{
				AllowUserTheme:     r.FormValue("allow_user_theme") == "1",
				AllowUserCustomCSS: r.FormValue("allow_user_custom_css") == "1",
			}
			if err := settingsSvc.SaveServerThemePolicy(r.Context(), policy); err != nil {
				http.Error(w, "Failed to save theme policy", http.StatusInternalServerError)
				return
			}
			message = "Theme policy saved."
		case "theme_default_css":
			if r.FormValue("remove_default_custom_css") == "1" {
				if defaultCustomCSSPath != "" {
					_ = os.Remove(filepath.Join(featuresettings.ThemeDir(h.deps.Config()), defaultCustomCSSPath))
				}
				_ = settingsSvc.SaveServerDefaultCustomCSS(r.Context(), "")
				message = "Default CSS removed."
			} else {
				file, header, err := r.FormFile("default_custom_css_file")
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
			}
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
		// Backwards-compatible invite keys expected by template
		"UsedInvitesCount": usedInvitesCount,
		"InviteBaseURL":    h.deps.BaseURL(r),
		"AuditLogs":        auditLogs,
		"AuditPage":        auditPage,
		"AuditPrevPage":    auditPrevPage,
		"AuditNextPage":    auditNextPage,
		"AuditHasMore":     auditHasMore,
		"AuditTotal":       auditTotal,
		"AuditTotalPages":  auditTotalPages,
	}

	if err := h.deps.RenderTemplate(w, "settings_admin.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
