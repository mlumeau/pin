package admin

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/crypto/bcrypt"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/core"
)

// Security handles the HTTP request.
func (h Handler) Security(w http.ResponseWriter, r *http.Request) {
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

	passkeys, _ := h.deps.ListPasskeys(r.Context(), current.ID)
	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), &current)
	isAdminUser := isAdmin(current)
	showAppearanceNav := isAdminUser || settingsSvc.ServerThemePolicy(r.Context()).AllowUserTheme
	message := ""
	data := map[string]interface{}{
		"User":               currentIdentity,
		"IsAdmin":            isAdminUser,
		"Passkeys":           passkeys,
		"Title":              "Settings - Privacy & security",
		"SectionTitle":       "Privacy & security",
		"Message":            message,
		"CSRFToken":          h.deps.EnsureCSRF(session),
		"PrivateIdentityURL": h.deps.BaseURL(r) + "/p/" + url.PathEscape(core.ShortHash(strings.ToLower(currentIdentity.Handle), 7)) + "/" + url.PathEscape(currentIdentity.PrivateToken),
		"Theme":              theme,
		"ShowAppearanceNav":  showAppearanceNav,
	}
	if toast := r.URL.Query().Get("toast"); toast != "" {
		message = toast
		data["Message"] = message
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if !h.deps.ValidateCSRF(session, r.FormValue("csrf_token")) {
			http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
			return
		}

		if newPassword := r.FormValue("new_password"); newPassword != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to update password", http.StatusInternalServerError)
				return
			}
			current.PasswordHash = string(hash)
			h.deps.AuditAttempt(r.Context(), current.ID, "password.update", currentIdentity.Handle, nil)
			if err := h.deps.UpdateUser(r.Context(), current); err != nil {
				h.deps.AuditOutcome(r.Context(), current.ID, "password.update", currentIdentity.Handle, err, nil)
				http.Error(w, "Failed to update password", http.StatusInternalServerError)
				return
			}
			h.deps.AuditOutcome(r.Context(), current.ID, "password.update", currentIdentity.Handle, nil, nil)
			data["Message"] = "Password updated successfully."
		} else {
			data["Message"] = "Enter a new password to update."
		}
	}

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.deps.RenderTemplate(w, "settings_security.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// PrivateIdentityRegenerate handles HTTP requests for identity regenerate.
func (h Handler) PrivateIdentityRegenerate(w http.ResponseWriter, r *http.Request) {
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
	currentIdentity, err := h.deps.CurrentIdentity(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "private_identity.regenerate", currentIdentity.Handle, nil)
	token := core.RandomTokenURL(32)
	if err := h.deps.UpdateIdentityPrivateToken(r.Context(), currentIdentity.ID, token); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "private_identity.regenerate", currentIdentity.Handle, err, nil)
		http.Error(w, "Failed to update private identity", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "private_identity.regenerate", currentIdentity.Handle, nil, nil)
	http.Redirect(w, r, "/settings/security?toast=Private%20link%20regenerated#section-private-identity", http.StatusFound)
}
