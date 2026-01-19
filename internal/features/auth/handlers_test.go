package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"pin/internal/domain"
)

type authDeps struct {
	hasUser bool
	store   sessions.Store
}

func (d authDeps) HasUser(ctx context.Context) (bool, error) { return d.hasUser, nil }
func (d authDeps) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return d.store.Get(r, name)
}
func (authDeps) EnsureCSRF(session *sessions.Session) string { return "token" }
func (authDeps) ValidateCSRF(session *sessions.Session, token string) bool {
	return true
}
func (authDeps) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	return domain.User{}, errors.New("not found")
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
