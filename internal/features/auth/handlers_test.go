package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"pin/internal/domain"
)

type authDeps struct {
	hasUser     bool
	store       sessions.Store
	identity    domain.Identity
	identityErr error
	user        domain.User
	userErr     error
}

func (d authDeps) HasUser(ctx context.Context) (bool, error) { return d.hasUser, nil }
func (d authDeps) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return d.store.Get(r, name)
}
func (authDeps) EnsureCSRF(session *sessions.Session) string { return "token" }
func (authDeps) ValidateCSRF(session *sessions.Session, token string) bool {
	return true
}
func (d authDeps) GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error) {
	if d.identityErr != nil {
		return domain.Identity{}, d.identityErr
	}
	if d.identity.ID == 0 {
		return domain.Identity{}, errors.New("not found")
	}
	return d.identity, nil
}
func (d authDeps) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	if d.userErr != nil {
		return domain.User{}, d.userErr
	}
	if d.user.ID == 0 || d.user.ID != id {
		return domain.User{}, errors.New("not found")
	}
	return d.user, nil
}
func (authDeps) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

// settings.Store
func (authDeps) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (authDeps) GetSetting(ctx context.Context, key string) (string, bool, error) {
	return "", false, nil
}
func (authDeps) SetSetting(ctx context.Context, key, value string) error         { return nil }
func (authDeps) SetSettings(ctx context.Context, values map[string]string) error { return nil }
func (authDeps) DeleteSetting(ctx context.Context, key string) error             { return nil }
func (authDeps) UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	return nil
}
func (authDeps) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return domain.User{}, errors.New("no user")
}

func TestLoginRedirectsToSetup(t *testing.T) {
	deps := authDeps{hasUser: false, store: sessions.NewCookieStore([]byte("test-secret"))}
	handler := NewHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/setup" {
		t.Fatalf("expected /setup, got %q", loc)
	}
}

func TestLogoutRedirectsHome(t *testing.T) {
	deps := authDeps{hasUser: true, store: sessions.NewCookieStore([]byte("test-secret"))}
	handler := NewHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	handler.Logout(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Fatalf("expected /, got %q", loc)
	}
}

func TestLoginSetsSessionFromIdentityUser(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test-secret"))
	password := "super-secret"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	secret := "JBSWY3DPEHPK3PXP"
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("generate totp: %v", err)
	}
	deps := authDeps{
		hasUser: true,
		store:   store,
		identity: domain.Identity{
			ID:     7,
			UserID: 42,
			Handle: "alice",
		},
		user: domain.User{
			ID:           42,
			PasswordHash: string(passwordHash),
			TOTPSecret:   secret,
		},
	}
	handler := NewHandler(deps)

	form := url.Values{
		"csrf_token": {"token"},
		"handle":     {"alice"},
		"password":   {password},
		"totp":       {code},
	}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/settings" {
		t.Fatalf("expected /settings, got %q", loc)
	}

	resp := rec.Result()
	defer resp.Body.Close()
	reqSession := httptest.NewRequest(http.MethodGet, "/settings", nil)
	for _, cookie := range resp.Cookies() {
		reqSession.AddCookie(cookie)
	}
	session, err := store.Get(reqSession, "pin_session")
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if got, ok := session.Values["user_id"]; !ok || got != 42 {
		t.Fatalf("expected user_id 42, got %v", session.Values["user_id"])
	}
}
