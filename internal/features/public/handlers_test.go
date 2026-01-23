package public

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"pin/internal/config"
	"pin/internal/domain"
)

type publicDeps struct {
	hasUser bool
}

func (d publicDeps) Config() config.Config                     { return config.Config{} }
func (d publicDeps) HasUser(ctx context.Context) (bool, error) { return d.hasUser, nil }
func (d publicDeps) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	return domain.User{}, errors.New("not found")
}
func (d publicDeps) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return domain.Identity{}, errors.New("no identity")
}
func (publicDeps) GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error) {
	return domain.Identity{}, errors.New("not found")
}
func (publicDeps) GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error) {
	return domain.Identity{}, errors.New("not found")
}
func (publicDeps) Reserved() map[string]struct{} { return map[string]struct{}{} }
func (publicDeps) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
func (publicDeps) BaseURL(r *http.Request) string { return "http://example.test" }
func (publicDeps) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	store := sessions.NewCookieStore([]byte("test-secret"))
	return store.Get(r, name)
}
func (publicDeps) CurrentUser(r *http.Request) (domain.User, error) {
	return domain.User{}, errors.New("no user")
}
func (publicDeps) CurrentIdentity(r *http.Request) (domain.Identity, error) {
	return domain.Identity{}, errors.New("no identity")
}
func (publicDeps) EnsureCSRF(session *sessions.Session) string               { return "token" }
func (publicDeps) ValidateCSRF(session *sessions.Session, token string) bool { return true }
func (publicDeps) CheckHandleCollision(ctx context.Context, handle string, excludeID int) error {
	return nil
}
func (publicDeps) CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error) {
	return 0, nil
}
func (publicDeps) CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error) {
	return 0, nil
}
func (publicDeps) UpdateIdentityPrivateToken(ctx context.Context, identityID int, token string) error {
	return nil
}
func (publicDeps) DeleteUser(ctx context.Context, userID int) error { return nil }
func (publicDeps) AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string) {
}
func (publicDeps) AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string) {
}

// settings.Store
func (publicDeps) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (publicDeps) GetSetting(ctx context.Context, key string) (string, bool, error) {
	return "", false, nil
}
func (publicDeps) SetSetting(ctx context.Context, key, value string) error         { return nil }
func (publicDeps) SetSettings(ctx context.Context, values map[string]string) error { return nil }
func (publicDeps) DeleteSetting(ctx context.Context, key string) error             { return nil }
func (publicDeps) UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	return nil
}
func (publicDeps) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return domain.User{}, errors.New("no user")
}

// domains.Store
func (publicDeps) ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error) {
	return nil, nil
}
func (publicDeps) UpsertDomainVerification(ctx context.Context, identityID int, domain, token string) error {
	return nil
}
func (publicDeps) DeleteDomainVerification(ctx context.Context, identityID int, domain string) error {
	return nil
}
func (publicDeps) MarkDomainVerified(ctx context.Context, identityID int, domain string) error {
	return nil
}

// domains.Protector
func (publicDeps) HasDomainVerification(ctx context.Context, identityID int, domain string) (bool, error) {
	return false, nil
}
func (publicDeps) ProtectedDomain(ctx context.Context) string                  { return "" }
func (publicDeps) SetProtectedDomain(ctx context.Context, domain string) error { return nil }

// profilepicture.Store
func (publicDeps) ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error) {
	return nil, nil
}
func (publicDeps) CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error) {
	return 0, nil
}
func (publicDeps) SetActiveProfilePicture(ctx context.Context, identityID int, pictureID int64) error {
	return nil
}
func (publicDeps) DeleteProfilePicture(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return "", nil
}
func (publicDeps) UpdateProfilePictureAlt(ctx context.Context, identityID int, pictureID int64, alt string) error {
	return nil
}
func (publicDeps) ClearProfilePictureSelection(ctx context.Context, identityID int) error { return nil }
func (publicDeps) GetProfilePictureFilename(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return "", nil
}
func (publicDeps) GetProfilePictureAlt(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return "", nil
}

func TestIndexRedirectsToSetup(t *testing.T) {
	handler := Handler{deps: publicDeps{hasUser: false}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.Index(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/setup" {
		t.Fatalf("expected /setup, got %q", loc)
	}
}
