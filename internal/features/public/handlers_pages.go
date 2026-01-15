package public

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"pin/internal/domain"
	"pin/internal/features/domains"
	"pin/internal/features/identity"
	"pin/internal/features/identity/export"
	"pin/internal/features/profilepicture"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/core"
)

// Index renders the root page (landing page or profile).
func (h Handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if ext := identity.ExtensionFromPath(r.URL.Path); ext != "" {
			handler := export.NewHandler(identitySource{deps: h.deps})
			switch ext {
			case "json":
				handler.IdentityJSON(w, r)
			case "xml":
				handler.IdentityXML(w, r)
			case "txt":
				handler.IdentityTXT(w, r)
			case "vcf":
				handler.IdentityVCF(w, r)
			default:
				http.NotFound(w, r)
			}
			return
		}
		if !identity.IsReservedPath(r.URL.Path, h.deps.Reserved()) {
			h.Profile(w, r)
			return
		}
		http.NotFound(w, r)
		return
	}
	if ok, err := h.deps.HasUser(r.Context()); err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	} else if !ok {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	user, err := h.deps.GetOwnerUser(r.Context())
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	settingsSvc := featuresettings.NewService(h.deps)
	landing := settingsSvc.LandingSettings(r.Context())
	theme := settingsSvc.DefaultThemeSettings(r.Context())

	if landing.Mode == featuresettings.LandingModeCustom && landing.CustomPath != "" {
		customPath := featuresettings.LandingCustomPath(h.deps.Config(), landing.CustomPath)
		if data, err := os.ReadFile(customPath); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(data)
			return
		}
	}

	publicUser, customFields := identity.VisibleIdentity(user, false)
	showLanding := landing.Mode != featuresettings.LandingModeProfile

	data := map[string]interface{}{
		"User":              publicUser,
		"ProfilePictureAlt": profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user),
		"Theme":             theme,
		"ShowLanding":       showLanding,
	}

	if !showLanding {
		var links []domain.Link
		if publicUser.LinksJSON != "" {
			_ = json.Unmarshal([]byte(publicUser.LinksJSON), &links)
		}
		var socialProfiles []domain.SocialProfile
		if publicUser.SocialProfilesJSON != "" {
			_ = json.Unmarshal([]byte(publicUser.SocialProfilesJSON), &socialProfiles)
		}
		wallets := identity.DecodeStringMap(publicUser.WalletsJSON)
		publicKeys := identity.DecodeStringMap(publicUser.PublicKeysJSON)
		verifiedDomains := identity.DecodeStringSlice(publicUser.VerifiedDomainsJSON)
		data["Links"] = links
		data["SocialProfiles"] = socialProfiles
		data["CustomFields"] = customFields
		data["Wallets"] = wallets
		data["PublicKeys"] = publicKeys
		data["VerifiedDomains"] = verifiedDomains
		data["ProfileURL"] = h.deps.BaseURL(r)
		data["ExportBase"] = "/" + url.PathEscape(user.Username)
		updatedAt := user.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = time.Now().UTC()
		}
		data["UpdatedAt"] = updatedAt
	} else {
		profilePath := "/" + url.PathEscape(user.Username)
		data["ProfilePath"] = profilePath
		data["ProfileURL"] = h.deps.BaseURL(r) + profilePath
		data["ExportBase"] = profilePath
		data["ProfilePictureURL"] = "/profile-picture/" + url.PathEscape(user.Username) + "?s=160"
		data["HasCustomLandingHTML"] = false
	}

	if err := h.deps.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// Landing renders the landing page regardless of the default mode.
func (h Handler) Landing(w http.ResponseWriter, r *http.Request) {
	if ok, err := h.deps.HasUser(r.Context()); err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	} else if !ok {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	user, err := h.deps.GetOwnerUser(r.Context())
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	settingsSvc := featuresettings.NewService(h.deps)
	landing := settingsSvc.LandingSettings(r.Context())
	theme := settingsSvc.DefaultThemeSettings(r.Context())

	if landing.Mode == featuresettings.LandingModeCustom && landing.CustomPath != "" {
		customPath := featuresettings.LandingCustomPath(h.deps.Config(), landing.CustomPath)
		if data, err := os.ReadFile(customPath); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(data)
			return
		}
	}

	publicUser, _ := identity.VisibleIdentity(user, false)
	profilePath := "/" + url.PathEscape(user.Username)
	data := map[string]interface{}{
		"User":                 publicUser,
		"ProfilePath":          profilePath,
		"ProfileURL":           h.deps.BaseURL(r) + profilePath,
		"ExportBase":           profilePath,
		"ProfilePictureAlt":    profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user),
		"ProfilePictureURL":    "/profile-picture/" + url.PathEscape(user.Username) + "?s=160",
		"HasCustomLandingHTML": false,
		"Theme":                theme,
		"ShowLanding":          true,
	}
	if err := h.deps.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// Profile renders a user's public profile page.
func (h Handler) Profile(w http.ResponseWriter, r *http.Request) {
	ident := strings.Trim(r.URL.Path, "/")
	if strings.HasPrefix(ident, "u/") {
		ident = strings.TrimPrefix(ident, "u/")
	}
	ident = strings.Trim(ident, "/")
	if ident == "" {
		http.NotFound(w, r)
		return
	}
	if name, ext := identity.FromIdent(ident); ext != "" {
		handler := export.NewHandler(identitySource{deps: h.deps})
		if name == "" {
			switch ext {
			case "json":
				handler.IdentityJSON(w, r)
			case "xml":
				handler.IdentityXML(w, r)
			case "txt":
				handler.IdentityTXT(w, r)
			case "vcf":
				handler.IdentityVCF(w, r)
			default:
				http.NotFound(w, r)
			}
			return
		}
		user, err := h.deps.FindUserByIdentifier(r.Context(), name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		publicUser, customFields := identity.VisibleIdentity(user, false)
		profileURL := h.deps.BaseURL(r) + "/" + url.PathEscape(user.Username)
		if err := handler.ServeIdentity(w, r, publicUser, customFields, profileURL, ext); err != nil {
			http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		}
		return
	}
	user, err := h.deps.FindUserByIdentifier(r.Context(), ident)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), &user)
	publicUser, customFields := identity.VisibleIdentity(user, false)
	var links []domain.Link
	if publicUser.LinksJSON != "" {
		_ = json.Unmarshal([]byte(publicUser.LinksJSON), &links)
	}
	var socialProfiles []domain.SocialProfile
	if publicUser.SocialProfilesJSON != "" {
		_ = json.Unmarshal([]byte(publicUser.SocialProfilesJSON), &socialProfiles)
	}
	wallets := identity.DecodeStringMap(publicUser.WalletsJSON)
	publicKeys := identity.DecodeStringMap(publicUser.PublicKeysJSON)
	verifiedDomains := identity.DecodeStringSlice(publicUser.VerifiedDomainsJSON)
	updatedAt := user.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	data := map[string]interface{}{
		"User":              publicUser,
		"Links":             links,
		"SocialProfiles":    socialProfiles,
		"CustomFields":      customFields,
		"Wallets":           wallets,
		"PublicKeys":        publicKeys,
		"VerifiedDomains":   verifiedDomains,
		"ProfileURL":        h.deps.BaseURL(r) + "/" + url.PathEscape(user.Username),
		"ExportBase":        "/" + url.PathEscape(user.Username),
		"ProfilePictureAlt": profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user),
		"UpdatedAt":         updatedAt,
		"Theme":             theme,
	}

	if err := h.deps.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// PrivateIdentity serves a private profile view and private exports via /p/<usernamehash>/<token>.
func (h Handler) PrivateIdentity(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/p/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	name, ext := identity.FromIdent(path)
	if ext != "" {
		path = name
	}
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	usernameHash := strings.TrimSpace(parts[0])
	token := strings.TrimSpace(parts[1])
	if usernameHash == "" || token == "" {
		http.NotFound(w, r)
		return
	}
	user, err := h.deps.GetUserByPrivateToken(r.Context(), token)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	expectedHash := core.ShortHash(strings.ToLower(strings.TrimSpace(user.Username)), 7)
	if !strings.EqualFold(usernameHash, expectedHash) {
		http.NotFound(w, r)
		return
	}
	privateUser, customFields := identity.VisibleIdentity(user, true)
	if ext != "" {
		handler := export.NewHandler(identitySource{deps: h.deps})
		profileURL := h.deps.BaseURL(r) + "/p/" + url.PathEscape(expectedHash) + "/" + url.PathEscape(user.PrivateToken)
		if err := handler.ServeIdentity(w, r, privateUser, customFields, profileURL, ext); err != nil {
			http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		}
		return
	}

	var links []domain.Link
	if privateUser.LinksJSON != "" {
		_ = json.Unmarshal([]byte(privateUser.LinksJSON), &links)
	}
	var socialProfiles []domain.SocialProfile
	if privateUser.SocialProfilesJSON != "" {
		_ = json.Unmarshal([]byte(privateUser.SocialProfilesJSON), &socialProfiles)
	}
	theme := featuresettings.NewService(h.deps).ThemeSettings(r.Context(), &user)
	updatedAt := user.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	data := map[string]interface{}{
		"User":              privateUser,
		"Links":             links,
		"SocialProfiles":    socialProfiles,
		"CustomFields":      customFields,
		"Wallets":           identity.DecodeStringMap(privateUser.WalletsJSON),
		"PublicKeys":        identity.DecodeStringMap(privateUser.PublicKeysJSON),
		"VerifiedDomains":   identity.DecodeStringSlice(privateUser.VerifiedDomainsJSON),
		"ProfileURL":        h.deps.BaseURL(r) + "/p/" + url.PathEscape(expectedHash) + "/" + url.PathEscape(user.PrivateToken),
		"ExportBase":        "/p/" + url.PathEscape(expectedHash) + "/" + url.PathEscape(user.PrivateToken),
		"ProfilePictureAlt": profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user),
		"IsPrivateIdentity": true,
		"UpdatedAt":         updatedAt,
		"Theme":             theme,
	}
	if err := h.deps.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// Setup creates the first admin user when none exists.
func (h Handler) Setup(w http.ResponseWriter, r *http.Request) {
	if ok, err := h.deps.HasUser(r.Context()); err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	} else if ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	session, _ := h.deps.GetSession(r, "pin_session")
	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), nil)
	data := map[string]interface{}{
		"Error":     "",
		"Success":   false,
		"CSRFToken": h.deps.EnsureCSRF(session),
		"TOTP":      "",
		"Theme":     theme,
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

		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		if username == "" || password == "" {
			data["Error"] = "Username and password are required"
		} else if identity.IsReservedIdentifier(username, h.deps.Reserved()) {
			data["Error"] = "Username is reserved"
		} else if err := identity.ValidateIdentifiers(r.Context(), username, nil, "", 0, h.deps.Reserved(), h.deps.CheckIdentifierCollisions); err != nil {
			data["Error"] = err.Error()
		} else {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to create account", http.StatusInternalServerError)
				return
			}
			key, err := totp.Generate(totp.GenerateOpts{Issuer: "pin", AccountName: username})
			if err != nil {
				http.Error(w, "Failed to create account", http.StatusInternalServerError)
				return
			}
			secret := key.Secret()

			defaultTheme := featuresettings.DefaultThemeName
			if themeValue, ok, _ := settingsSvc.ServerDefaultTheme(r.Context()); ok {
				defaultTheme = themeValue
			}

			h.deps.AuditAttempt(r.Context(), 0, "user.create", username, map[string]string{"source": "setup"})
			privateToken := core.RandomToken(32)
			userID, err := h.deps.CreateUser(r.Context(), username, email, "owner", string(hash), secret, defaultTheme, privateToken)
			if err != nil {
				h.deps.AuditOutcome(r.Context(), 0, "user.create", username, err, map[string]string{"source": "setup"})
				data["Error"] = "Username already exists"
				goto renderSetup
			}
			h.deps.AuditOutcome(r.Context(), 0, "user.create", username, nil, map[string]string{"source": "setup"})
			_ = h.deps.UpsertUserIdentifiers(r.Context(), int(userID), username, nil, email)
			domains.NewService(h.deps).EnsureServerDomainVerification(r.Context(), h.deps.Config().BaseURL, int(userID), h.deps, func() string {
				return domains.RandomTokenURL(12)
			})

			data["Success"] = true
			data["TOTP"] = secret
		}
	}

renderSetup:
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.deps.RenderTemplate(w, "setup.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
