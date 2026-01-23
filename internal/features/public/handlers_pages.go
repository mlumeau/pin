package public

import (
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
		if handle, ok := profilePictureHandleFromPath(r.URL.Path); ok {
			handler := h.profilePictureHandler()
			handler.ProfilePictureByHandle(w, r, handle)
			return
		}
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
	user, err := h.deps.GetOwnerIdentity(r.Context())
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
	profilePath := "/" + url.PathEscape(user.Handle)
	profilePictureAlt := profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user)

	data := map[string]interface{}{
		"User":              publicUser,
		"Theme":             theme,
		"ProfilePictureAlt": profilePictureAlt,
		"ShowLanding":       showLanding,
	}

	if !showLanding {
		links := identity.DecodeLinks(publicUser.LinksJSON)
		socialProfiles := identity.DecodeSocialProfiles(publicUser.SocialProfilesJSON)
		wallets := identity.WalletsMapToStructs(identity.DecodeStringMap(publicUser.WalletsJSON))
		publicKeys := identity.PublicKeysMapToStructs(identity.DecodeStringMap(publicUser.PublicKeysJSON))
		verifiedDomains := identity.VerifiedDomainsSliceToStructs(identity.DecodeStringSlice(publicUser.VerifiedDomainsJSON))
		data["Links"] = links
		data["SocialProfiles"] = socialProfiles
		data["CustomFields"] = customFields
		data["Wallets"] = wallets
		data["PublicKeys"] = publicKeys
		data["VerifiedDomains"] = verifiedDomains
		data["ProfileURL"] = h.deps.BaseURL(r)
		data["ExportBase"] = profilePath
		data["ProfilePictureURL"] = "/" + url.PathEscape(user.Handle) + "/profile-picture"
		updatedAt := user.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = time.Now().UTC()
		}
		data["UpdatedAt"] = updatedAt
		if authUser, err := h.deps.GetUserByID(r.Context(), user.UserID); err == nil {
			theme = settingsSvc.ThemeSettings(r.Context(), &authUser)
		}
	} else {
		data["ProfilePath"] = profilePath
		data["ProfileURL"] = h.deps.BaseURL(r) + profilePath
		data["ExportBase"] = profilePath
		data["ProfilePictureURL"] = "/" + url.PathEscape(user.Handle) + "/profile-picture"
		data["HasCustomLandingHTML"] = false
		if authUser, err := h.deps.GetUserByID(r.Context(), user.UserID); err == nil {
			theme = settingsSvc.ThemeSettings(r.Context(), &authUser)
		}
	}
	data["Theme"] = theme

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
	user, err := h.deps.GetOwnerIdentity(r.Context())
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
	profilePath := "/" + url.PathEscape(user.Handle)
	profilePictureAlt := profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user)
	data := map[string]interface{}{
		"User":                 publicUser,
		"ProfilePath":          profilePath,
		"ProfileURL":           h.deps.BaseURL(r) + profilePath,
		"ExportBase":           profilePath,
		"ProfilePictureAlt":    profilePictureAlt,
		"ProfilePictureURL":    "/" + url.PathEscape(user.Handle) + "/profile-picture",
		"HasCustomLandingHTML": false,
		"Theme":                theme,
		"ShowLanding":          true,
	}
	if authUser, err := h.deps.GetUserByID(r.Context(), user.UserID); err == nil {
		theme = settingsSvc.ThemeSettings(r.Context(), &authUser)
		data["Theme"] = theme
	}
	if err := h.deps.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// Profile renders a user's public profile page.
func (h Handler) Profile(w http.ResponseWriter, r *http.Request) {
	ident := strings.Trim(r.URL.Path, "/")
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
		user, err := h.deps.GetIdentityByHandle(r.Context(), name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		publicUser, customFields := identity.VisibleIdentity(user, false)
		profileURL := h.deps.BaseURL(r) + "/" + url.PathEscape(user.Handle)
		if ext == "json" {
			selfURL := h.deps.BaseURL(r) + r.URL.Path
			if r.URL.RawQuery != "" {
				selfURL += "?" + r.URL.RawQuery
			}
			if err := handler.ServePINCJSON(w, r, publicUser, customFields, "public", selfURL); err != nil {
				http.Error(w, "Failed to load identity", http.StatusInternalServerError)
			}
			return
		}
		if err := handler.ServeIdentity(w, r, publicUser, customFields, profileURL, ext, false); err != nil {
			http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		}
		return
	}
	user, err := h.deps.GetIdentityByHandle(r.Context(), ident)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.DefaultThemeSettings(r.Context())
	if authUser, err := h.deps.GetUserByID(r.Context(), user.UserID); err == nil {
		theme = settingsSvc.ThemeSettings(r.Context(), &authUser)
	}
	publicUser, customFields := identity.VisibleIdentity(user, false)
	links := identity.DecodeLinks(publicUser.LinksJSON)
	socialProfiles := identity.DecodeSocialProfiles(publicUser.SocialProfilesJSON)
	wallets := identity.WalletsMapToStructs(identity.DecodeStringMap(publicUser.WalletsJSON))
	publicKeys := identity.PublicKeysMapToStructs(identity.DecodeStringMap(publicUser.PublicKeysJSON))
	verifiedDomains := identity.VerifiedDomainsSliceToStructs(identity.DecodeStringSlice(publicUser.VerifiedDomainsJSON))
	updatedAt := user.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	profilePath := "/" + url.PathEscape(user.Handle)
	profileURL := h.deps.BaseURL(r) + profilePath
	profilePictureAlt := profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user)
	data := map[string]interface{}{
		"User":              publicUser,
		"Links":             links,
		"SocialProfiles":    socialProfiles,
		"CustomFields":      customFields,
		"Wallets":           wallets,
		"PublicKeys":        publicKeys,
		"VerifiedDomains":   verifiedDomains,
		"ProfileURL":        profileURL,
		"ExportBase":        profilePath,
		"ProfilePictureURL": "/" + url.PathEscape(user.Handle) + "/profile-picture",
		"ProfilePictureAlt": profilePictureAlt,
		"UpdatedAt":         updatedAt,
		"Theme":             theme,
	}

	if err := h.deps.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// PrivateIdentity serves a private profile view and private exports via /p/<handlehash>/<token>.
func (h Handler) PrivateIdentity(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/p/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	isProfilePicture := false
	if strings.HasSuffix(path, "/profile-picture") {
		isProfilePicture = true
		path = strings.TrimSuffix(path, "/profile-picture")
		path = strings.TrimSuffix(path, "/")
		if path == "" {
			http.NotFound(w, r)
			return
		}
	}
	name, ext := identity.FromIdent(path)
	if ext != "" {
		if isProfilePicture {
			http.NotFound(w, r)
			return
		}
		path = name
	}
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	handleHash := strings.TrimSpace(parts[0])
	token := strings.TrimSpace(parts[1])
	if handleHash == "" || token == "" {
		http.NotFound(w, r)
		return
	}
	user, err := h.deps.GetIdentityByPrivateToken(r.Context(), token)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	expectedHash := core.ShortHash(strings.ToLower(strings.TrimSpace(user.Handle)), 7)
	if !strings.EqualFold(handleHash, expectedHash) {
		http.NotFound(w, r)
		return
	}
	if isProfilePicture {
		handler := h.profilePictureHandler()
		handler.ProfilePictureForUser(w, r, user)
		return
	}
	privateUser, customFields := identity.VisibleIdentity(user, true)
	if ext != "" {
		handler := export.NewHandler(identitySource{deps: h.deps})
		profileURL := h.deps.BaseURL(r) + "/p/" + url.PathEscape(expectedHash) + "/" + url.PathEscape(user.PrivateToken)
		if ext == "json" {
			selfURL := h.deps.BaseURL(r) + r.URL.Path
			if r.URL.RawQuery != "" {
				selfURL += "?" + r.URL.RawQuery
			}
			if err := handler.ServePINCJSON(w, r, privateUser, customFields, "private", selfURL); err != nil {
				http.Error(w, "Failed to load identity", http.StatusInternalServerError)
			}
			return
		}
		if err := handler.ServeIdentity(w, r, privateUser, customFields, profileURL, ext, true); err != nil {
			http.Error(w, "Failed to load identity", http.StatusInternalServerError)
		}
		return
	}

	links := identity.DecodeLinks(privateUser.LinksJSON)
	socialProfiles := identity.DecodeSocialProfiles(privateUser.SocialProfilesJSON)
	theme := featuresettings.NewService(h.deps).DefaultThemeSettings(r.Context())
	if authUser, err := h.deps.GetUserByID(r.Context(), user.UserID); err == nil {
		theme = featuresettings.NewService(h.deps).ThemeSettings(r.Context(), &authUser)
	}
	updatedAt := user.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	profilePath := "/p/" + url.PathEscape(expectedHash) + "/" + url.PathEscape(user.PrivateToken)
	profileURL := h.deps.BaseURL(r) + profilePath
	data := map[string]interface{}{
		"User":              privateUser,
		"Links":             links,
		"SocialProfiles":    socialProfiles,
		"CustomFields":      customFields,
		"Wallets":           identity.WalletsMapToStructs(identity.DecodeStringMap(privateUser.WalletsJSON)),
		"PublicKeys":        identity.PublicKeysMapToStructs(identity.DecodeStringMap(privateUser.PublicKeysJSON)),
		"VerifiedDomains":   identity.VerifiedDomainsSliceToStructs(identity.DecodeStringSlice(privateUser.VerifiedDomainsJSON)),
		"ProfileURL":        profileURL,
		"ExportBase":        profilePath,
		"ProfilePictureURL": profilePath + "/profile-picture",
		"ProfilePictureAlt": profilepicture.NewService(h.deps).ActiveAlt(r.Context(), user),
		"IsPrivateIdentity": true,
		"UpdatedAt":         updatedAt,
		"Theme":             theme,
	}
	if err := h.deps.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (h Handler) profilePictureHandler() profilepicture.Handler {
	cfg := h.deps.Config()
	return profilepicture.NewHandler(profilepicture.Config{
		ProfilePictureDir: cfg.ProfilePictureDir,
		StaticDir:         cfg.StaticDir,
		AllowedExts:       cfg.AllowedExts,
		MaxUploadBytes:    cfg.MaxUploadBytes,
		CacheAltFormats:   cfg.CacheAltFormats,
	}, h.deps)
}

func profilePictureHandleFromPath(path string) (string, bool) {
	path = strings.Trim(path, "/")
	if path == "" {
		return "", false
	}
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", false
	}
	if parts[1] != "profile-picture" {
		return "", false
	}
	if strings.TrimSpace(parts[0]) == "" {
		return "", false
	}
	return parts[0], true
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
		"Error":           "",
		"Success":         false,
		"CSRFToken":       h.deps.EnsureCSRF(session),
		"Theme":           theme,
		"PageTitle":       "Pin - Setup Admin",
		"PageHeading":     "Set up your admin account",
		"PageSubheading":  "Create the first owner account for this Pin instance.",
		"FormAction":      "/setup",
		"FormButtonLabel": "Create admin",
		"FormIntro":       "You can change these details later in settings.",
		"SuccessMessage":  "Account created. Set up your authenticator app to finish.",
		"TOTP":            "",
		"TOTPURL":         "",
		"IsAdmin":         true,
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

		handle := strings.TrimSpace(r.FormValue("handle"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		if handle == "" || password == "" {
			data["Error"] = "Handle and password are required"
		} else if identity.IsReservedIdentifier(handle, h.deps.Reserved()) {
			data["Error"] = "Handle is reserved"
		} else if err := identity.ValidateHandle(r.Context(), handle, 0, h.deps.Reserved(), h.deps.CheckHandleCollision); err != nil {
			data["Error"] = err.Error()
		} else {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to create account", http.StatusInternalServerError)
				return
			}
			key, err := totp.Generate(totp.GenerateOpts{Issuer: "pin", AccountName: handle})
			if err != nil {
				http.Error(w, "Failed to create account", http.StatusInternalServerError)
				return
			}
			secret := key.Secret()
			otpURL := key.URL()

			defaultTheme := featuresettings.DefaultThemeName
			if themeValue, ok, _ := settingsSvc.ServerDefaultTheme(r.Context()); ok {
				defaultTheme = themeValue
			}

			h.deps.AuditAttempt(r.Context(), 0, "user.create", handle, map[string]string{"source": "setup"})
			privateToken := core.RandomToken(32)
			userID, err := h.deps.CreateUser(r.Context(), "owner", string(hash), secret, defaultTheme)
			if err != nil {
				h.deps.AuditOutcome(r.Context(), 0, "user.create", handle, err, map[string]string{"source": "setup"})
				data["Error"] = "Failed to create account"
				goto renderSetup
			}
			identityRecord := domain.Identity{
				UserID:              int(userID),
				Handle:              handle,
				Email:               email,
				DisplayName:         handle,
				CustomFieldsJSON:    "{}",
				VisibilityJSON:      "{}",
				PrivateToken:        privateToken,
				LinksJSON:           "[]",
				SocialProfilesJSON:  "[]",
				WalletsJSON:         "{}",
				PublicKeysJSON:      "{}",
				VerifiedDomainsJSON: "[]",
			}
			identityID, err := h.deps.CreateIdentity(r.Context(), identityRecord)
			if err != nil {
				_ = h.deps.DeleteUser(r.Context(), int(userID))
				h.deps.AuditOutcome(r.Context(), 0, "user.create", handle, err, map[string]string{"source": "setup"})
				data["Error"] = "Handle already exists"
				goto renderSetup
			}
			h.deps.AuditOutcome(r.Context(), 0, "user.create", handle, nil, map[string]string{"source": "setup"})
			domains.NewService(h.deps).EnsureServerDomainVerification(r.Context(), h.deps.Config().BaseURL, int(identityID), h.deps, func() string {
				return domains.RandomTokenURL(12)
			})

			data["Success"] = true
			data["TOTP"] = secret
			data["TOTPURL"] = otpURL
		}
	}

renderSetup:
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.deps.RenderTemplate(w, "account-setup.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
