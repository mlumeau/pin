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

// Config returns an empty config for test handlers.
func (d publicDeps) Config() config.Config                     { return config.Config{} }
// HasUser reports whether user exists.
func (d publicDeps) HasUser(ctx context.Context) (bool, error) { return d.hasUser, nil }
// GetUserByID returns user by ID.
func (d publicDeps) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	return domain.User{}, errors.New("not found")
}
// GetOwnerIdentity returns the owner identity.
func (d publicDeps) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return domain.Identity{}, errors.New("no identity")
}
// GetIdentityByHandle returns identity by handle.
func (publicDeps) GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error) {
	return domain.Identity{}, errors.New("not found")
}
// GetIdentityByPrivateToken returns identity by private token.
func (publicDeps) GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error) {
	return domain.Identity{}, errors.New("not found")
}
// Reserved returns an empty reserved set for tests.
func (publicDeps) Reserved() map[string]struct{} { return map[string]struct{}{} }
// RenderTemplate stubs template rendering for tests.
func (publicDeps) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
// BaseURL returns a fixed base URL for tests.
func (publicDeps) BaseURL(r *http.Request) string { return "http://example.test" }
// GetSession returns the session.
func (publicDeps) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	store := sessions.NewCookieStore([]byte("test-secret"))
	return store.Get(r, name)
}
// CurrentUser returns an error for unauthenticated test requests.
func (publicDeps) CurrentUser(r *http.Request) (domain.User, error) {
	return domain.User{}, errors.New("no user")
}
// CurrentIdentity returns an error for unauthenticated test requests.
func (publicDeps) CurrentIdentity(r *http.Request) (domain.Identity, error) {
	return domain.Identity{}, errors.New("no identity")
}
// EnsureCSRF ensures CSRF is initialized and available.
func (publicDeps) EnsureCSRF(session *sessions.Session) string               { return "token" }
// ValidateCSRF validates CSRF and returns an error on failure.
func (publicDeps) ValidateCSRF(session *sessions.Session, token string) bool { return true }
// CheckHandleCollision checks handle collision and reports whether it matches.
func (publicDeps) CheckHandleCollision(ctx context.Context, handle string, excludeID int) error {
	return nil
}
// CreateUser creates user using the supplied input.
func (publicDeps) CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error) {
	return 0, nil
}
// CreateIdentity creates identity using the supplied input.
func (publicDeps) CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error) {
	return 0, nil
}
// UpdateIdentityPrivateToken updates identity private token using the supplied data.
func (publicDeps) UpdateIdentityPrivateToken(ctx context.Context, identityID int, token string) error {
	return nil
}
// DeleteUser deletes user.
func (publicDeps) DeleteUser(ctx context.Context, userID int) error { return nil }
// AuditAttempt records attempt as an audit event.
func (publicDeps) AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string) {
}
// AuditOutcome records outcome as an audit event.
func (publicDeps) AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string) {
}

// settings.Store
func (publicDeps) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	return map[string]string{}, nil
}
// GetSetting returns the setting.
func (publicDeps) GetSetting(ctx context.Context, key string) (string, bool, error) {
	return "", false, nil
}
// SetSetting sets setting to the provided value.
func (publicDeps) SetSetting(ctx context.Context, key, value string) error         { return nil }
// SetSettings sets settings to the provided value.
func (publicDeps) SetSettings(ctx context.Context, values map[string]string) error { return nil }
// DeleteSetting deletes setting.
func (publicDeps) DeleteSetting(ctx context.Context, key string) error             { return nil }
// UpdateUserTheme updates user theme using the supplied data.
func (publicDeps) UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	return nil
}
// GetOwnerUser returns the owner user.
func (publicDeps) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return domain.User{}, errors.New("no user")
}

// domains.Store
func (publicDeps) ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error) {
	return nil, nil
}
// UpsertDomainVerification is a no-op for tests.
func (publicDeps) UpsertDomainVerification(ctx context.Context, identityID int, domain, token string) error {
	return nil
}
// DeleteDomainVerification deletes domain verification.
func (publicDeps) DeleteDomainVerification(ctx context.Context, identityID int, domain string) error {
	return nil
}
// MarkDomainVerified returns domain verified.
func (publicDeps) MarkDomainVerified(ctx context.Context, identityID int, domain string) error {
	return nil
}

// domains.Protector
func (publicDeps) HasDomainVerification(ctx context.Context, identityID int, domain string) (bool, error) {
	return false, nil
}
// ProtectedDomain returns domain.
func (publicDeps) ProtectedDomain(ctx context.Context) string                  { return "" }
// SetProtectedDomain sets protected domain to the provided value.
func (publicDeps) SetProtectedDomain(ctx context.Context, domain string) error { return nil }

// profilepicture.Store
func (publicDeps) ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error) {
	return nil, nil
}
// CreateProfilePicture creates profile picture using the supplied input.
func (publicDeps) CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error) {
	return 0, nil
}
// SetActiveProfilePicture sets active profile picture to the provided value.
func (publicDeps) SetActiveProfilePicture(ctx context.Context, identityID int, pictureID int64) error {
	return nil
}
// DeleteProfilePicture deletes profile picture.
func (publicDeps) DeleteProfilePicture(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return "", nil
}
// UpdateProfilePictureAlt updates profile picture alt using the supplied data.
func (publicDeps) UpdateProfilePictureAlt(ctx context.Context, identityID int, pictureID int64, alt string) error {
	return nil
}
// ClearProfilePictureSelection returns profile picture selection.
func (publicDeps) ClearProfilePictureSelection(ctx context.Context, identityID int) error { return nil }
// GetProfilePictureFilename returns the profile picture filename.
func (publicDeps) GetProfilePictureFilename(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return "", nil
}
// GetProfilePictureAlt returns the profile picture alt.
func (publicDeps) GetProfilePictureAlt(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return "", nil
}

// TestIndexRedirectsToSetup verifies index redirects to setup behavior.
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
