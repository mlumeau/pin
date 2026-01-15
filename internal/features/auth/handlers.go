package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"pin/internal/domain"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/core"
)

type Dependencies interface {
	featuresettings.Store
	HasUser(ctx context.Context) (bool, error)
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	EnsureCSRF(session *sessions.Session) string
	ValidateCSRF(session *sessions.Session, token string) bool
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
}

type Handler struct {
	deps Dependencies
}

func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}

// Login renders the login form and authenticates the admin user.
func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	if ok, err := h.deps.HasUser(r.Context()); err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	} else if !ok {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	session, _ := h.deps.GetSession(r, "pin_session")
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/settings"
	}
	if !core.IsSafeRedirect(r, next) {
		next = "/settings"
	}

	theme := featuresettings.NewService(h.deps).DefaultThemeSettings(r.Context())
	data := map[string]interface{}{
		"Error":     "",
		"Next":      next,
		"CSRFToken": h.deps.EnsureCSRF(session),
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
		user, err := h.deps.GetUserByUsername(r.Context(), username)
		if err != nil {
			data["Error"] = "Invalid username"
			goto render
		}
		password := r.FormValue("password")
		code := strings.TrimSpace(r.FormValue("totp"))

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			data["Error"] = "Invalid password"
		} else if !totp.Validate(code, user.TOTPSecret) {
			data["Error"] = "Invalid one-time code"
		} else {
			session.Values["user_id"] = user.ID
			if err := session.Save(r, w); err != nil {
				http.Error(w, "Session error", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, next, http.StatusFound)
			return
		}
	}

render:
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.deps.RenderTemplate(w, "login.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// Logout clears the session and redirects to the home page.
func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.deps.GetSession(r, "pin_session")
	session.Values = map[interface{}]interface{}{}
	session.Options.MaxAge = -1
	_ = session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}
