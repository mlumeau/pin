package invites

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"pin/internal/domain"
	"pin/internal/features/identity"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/core"
)

type Dependencies interface {
	featuresettings.Store
	CurrentUser(r *http.Request) (domain.User, error)
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	EnsureCSRF(session *sessions.Session) string
	ValidateCSRF(session *sessions.Session, token string) bool
	CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error)
	CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error)
	DeleteUser(ctx context.Context, userID int) error
	CheckHandleCollision(ctx context.Context, handle string, excludeID int) error
	Reserved() map[string]struct{}
	GetInviteByToken(ctx context.Context, token string) (domain.Invite, error)
	MarkInviteUsed(ctx context.Context, id int, usedBy int) error
	CreateInvite(ctx context.Context, token, role string, createdBy int) error
	DeleteInvite(ctx context.Context, id int) error
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	deps Dependencies
}

// NewHandler constructs a new handler.
func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}

// Invite renders signup from an invite token.
func (h Handler) Invite(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/invite/")
	token = strings.TrimSpace(token)
	if token == "" {
		http.NotFound(w, r)
		return
	}
	invite, err := h.deps.GetInviteByToken(r.Context(), token)
	if err != nil || invite.UsedAt.Valid {
		http.NotFound(w, r)
		return
	}

	session, _ := h.deps.GetSession(r, "pin_session")
	settingsSvc := featuresettings.NewService(h.deps)
	theme := settingsSvc.ThemeSettings(r.Context(), nil)
	data := map[string]interface{}{
		"Error":            "",
		"Success":          false,
		"CSRFToken":        h.deps.EnsureCSRF(session),
		"Theme":            theme,
		"PageTitle":        "Pin - Accept Invite",
		"PageHeading":      "Join this Pin instance",
		"PageSubheading":   "Create your account to get started.",
		"FormAction":       r.URL.String(),
		"FormButtonLabel":  "Create account",
		"SuccessMessage":   "Account created. Set up your authenticator app to finish.",
		"TOTP":             "",
		"TOTPURL":          "",
		"IsAdmin":          false,
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

			h.deps.AuditAttempt(r.Context(), 0, "user.create", handle, map[string]string{"source": "invite"})
			privateToken := core.RandomToken(32)
			userID, err := h.deps.CreateUser(r.Context(), invite.Role, string(hash), secret, defaultTheme)
			if err != nil {
				h.deps.AuditOutcome(r.Context(), 0, "user.create", handle, err, map[string]string{"source": "invite"})
				http.Error(w, "Failed to create account", http.StatusInternalServerError)
				return
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
			if _, err := h.deps.CreateIdentity(r.Context(), identityRecord); err != nil {
				_ = h.deps.DeleteUser(r.Context(), int(userID))
				h.deps.AuditOutcome(r.Context(), 0, "user.create", handle, err, map[string]string{"source": "invite"})
				http.Error(w, "Failed to create account", http.StatusInternalServerError)
				return
			}
			_ = h.deps.MarkInviteUsed(r.Context(), invite.ID, int(userID))
			h.deps.AuditOutcome(r.Context(), int(userID), "user.create", handle, nil, map[string]string{"source": "invite"})
			data["Success"] = true
			data["TOTP"] = secret
			data["TOTPURL"] = otpURL
		}
	}

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.deps.RenderTemplate(w, "account-setup.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// Create issues a new invite token for the requested role.
func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	current, err := h.deps.CurrentUser(r)
	if err != nil || !isAdmin(current) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
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
	role := strings.TrimSpace(r.FormValue("role"))
	if role == "" {
		role = "user"
	}
	if role != "user" && role != "admin" {
		role = "user"
	}
	token := core.RandomToken(16)
	h.deps.AuditAttempt(r.Context(), current.ID, "invite.create", token, map[string]string{"role": role})
	if err := h.deps.CreateInvite(r.Context(), token, role, current.ID); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "invite.create", token, err, map[string]string{"role": role})
		http.Error(w, "Failed to create invite", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "invite.create", token, nil, map[string]string{"role": role})
	http.Redirect(w, r, "/settings/admin/server?toast=Invite%20created#section-invites", http.StatusFound)
}

// Delete revokes an invite by ID.
func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	current, err := h.deps.CurrentUser(r)
	if err != nil || !isAdmin(current) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
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
	id, err := strconv.Atoi(r.FormValue("invite_id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid invite", http.StatusBadRequest)
		return
	}
	h.deps.AuditAttempt(r.Context(), current.ID, "invite.delete", strconv.Itoa(id), nil)
	if err := h.deps.DeleteInvite(r.Context(), id); err != nil {
		h.deps.AuditOutcome(r.Context(), current.ID, "invite.delete", strconv.Itoa(id), err, nil)
		http.Error(w, "Failed to delete invite", http.StatusInternalServerError)
		return
	}
	h.deps.AuditOutcome(r.Context(), current.ID, "invite.delete", strconv.Itoa(id), nil, nil)
	http.Redirect(w, r, "/settings/admin/server?toast=Invite%20deleted#section-invites", http.StatusFound)
}

// isAdmin reports whether admin is true.
func isAdmin(user domain.User) bool {
	return user.Role == "owner" || user.Role == "admin"
}
