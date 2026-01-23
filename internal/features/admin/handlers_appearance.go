package admin

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	featuresettings "pin/internal/features/settings"
)

func (h Handler) Appearance(w http.ResponseWriter, r *http.Request) {
	session, _ := h.deps.GetSession(r, "pin_session")
	current, err := h.deps.CurrentUser(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	currentIdentity, err := h.deps.CurrentIdentity(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), &current)
	policy := settingsSvc.ServerThemePolicy(r.Context())
	isAdminUser := isAdmin(current)
	canSelectTheme := isAdminUser || policy.AllowUserTheme
	canCustomCSS := isAdminUser || (policy.AllowUserTheme && policy.AllowUserCustomCSS)
	showAppearanceNav := canSelectTheme
	defaultCustomCSSPath, hasDefaultCustomCSS := settingsSvc.ServerDefaultCustomCSS(r.Context())
	defaultCustomThemeOption := featuresettings.ThemeOption{
		Name:        featuresettings.DefaultCustomThemeName,
		Label:       "Default theme",
		Description: "Use the server default CSS file for the app.",
	}
	cfg := h.deps.Config()
	message := ""
	data := map[string]interface{}{
		"User":                 currentIdentity,
		"IsAdmin":              isAdminUser,
		"Themes":               featuresettings.ThemeOptions(),
		"Theme":                theme,
		"Title":                "Settings - Appearance",
		"DefaultCustomTheme":   defaultCustomThemeOption,
		"HasDefaultCustomCSS":  hasDefaultCustomCSS,
		"DefaultCustomCSSURL":  featuresettings.ThemeCustomCSSURL(defaultCustomCSSPath),
		"DefaultCustomCSSName": filepath.Base(defaultCustomCSSPath),
		"Message":              message,
		"CSRFToken":            h.deps.EnsureCSRF(session),
		"CanSelectTheme":       canSelectTheme,
		"CanCustomCSS":         canCustomCSS,
		"ShowAppearanceNav":    showAppearanceNav,
	}
	if toast := r.URL.Query().Get("toast"); toast != "" {
		message = toast
		data["Message"] = message
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
		if !canSelectTheme {
			data["Message"] = "Theme selection is disabled by the server administrator."
			goto renderAppearance
		}

		profileTheme := featuresettings.NormalizeThemeChoice(r.FormValue("profile_theme"))
		inlineCSS := strings.TrimSpace(r.FormValue("inline_css"))
		customCSSPath := strings.TrimSpace(current.ThemeCustomCSSPath)
		if !canCustomCSS {
			inlineCSS = strings.TrimSpace(current.ThemeCustomCSSInline)
		}

		updated := featuresettings.ThemeSettings{
			ProfileTheme:  profileTheme,
			InlineCSS:     inlineCSS,
			CustomCSSPath: customCSSPath,
		}

		if canCustomCSS && r.FormValue("action") == "delete-css" {
			if updated.CustomCSSPath != "" && updated.CustomCSSPath != defaultCustomCSSPath {
				_ = os.Remove(filepath.Join(featuresettings.ThemeDir(cfg), updated.CustomCSSPath))
			}
			updated.CustomCSSPath = ""
		} else if canCustomCSS {
			updated.CustomCSSPath = theme.CustomCSSPath
		}

		if canCustomCSS {
			if file, header, err := r.FormFile("custom_css_file"); err == nil && header != nil && header.Filename != "" {
				defer file.Close()
				ext := strings.ToLower(filepath.Ext(header.Filename))
				if ext != ".css" {
					updated.CustomCSSURL = featuresettings.ThemeCustomCSSURL(updated.CustomCSSPath)
					updated.InlineCSSTemplate = template.CSS(updated.InlineCSS)
					data["Theme"] = updated
					data["Message"] = "Custom CSS must be a .css file"
					goto renderAppearance
				}
				if err := os.MkdirAll(featuresettings.ThemeDir(cfg), 0755); err != nil {
					http.Error(w, "Failed to store custom CSS", http.StatusInternalServerError)
					return
				}
				filename := fmt.Sprintf("theme_u%d_%d.css", current.ID, time.Now().UTC().UnixNano())
				destPath := filepath.Join(featuresettings.ThemeDir(cfg), filename)
				dest, err := os.Create(destPath)
				if err != nil {
					http.Error(w, "Failed to store custom CSS", http.StatusInternalServerError)
					return
				}
				if _, err := io.Copy(dest, io.LimitReader(file, cfg.MaxUploadBytes)); err != nil {
					_ = dest.Close()
					_ = os.Remove(destPath)
					http.Error(w, "Failed to store custom CSS", http.StatusInternalServerError)
					return
				}
				_ = dest.Close()
				if updated.CustomCSSPath != "" && updated.CustomCSSPath != filename {
					_ = os.Remove(filepath.Join(featuresettings.ThemeDir(cfg), updated.CustomCSSPath))
				}
				updated.CustomCSSPath = filename
				updated.InlineCSS = ""
			}
		}

		meta := map[string]string{
			"profile_theme": profileTheme,
		}
		h.deps.AuditAttempt(r.Context(), current.ID, "appearance.update", currentIdentity.Handle, meta)
		if err := settingsSvc.SaveThemeSettings(r.Context(), current.ID, updated); err != nil {
			h.deps.AuditOutcome(r.Context(), current.ID, "appearance.update", currentIdentity.Handle, err, meta)
			http.Error(w, "Failed to save appearance settings", http.StatusInternalServerError)
			return
		}
		meta["has_custom_css"] = strconv.FormatBool(updated.CustomCSSPath != "")
		h.deps.AuditOutcome(r.Context(), current.ID, "appearance.update", currentIdentity.Handle, nil, meta)

		current.ThemeProfile = updated.ProfileTheme
		current.ThemeCustomCSSPath = updated.CustomCSSPath
		current.ThemeCustomCSSInline = updated.InlineCSS
		theme = settingsSvc.ThemeSettings(r.Context(), &current)
		data["Theme"] = theme
		data["Message"] = "Appearance updated."
	}

renderAppearance:
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.deps.RenderTemplate(w, "settings_appearance.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
