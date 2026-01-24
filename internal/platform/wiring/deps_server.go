package wiring

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"pin/internal/config"
	"pin/internal/domain"
)

// Platform helpers.
func (d Deps) Config() config.Config {
	return d.srv.Config()
}

// GetSession returns the session by delegating to configured services.
func (d Deps) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return d.srv.GetSession(r, name)
}

// EnsureCSRF ensures CSRF is initialized and available by delegating to configured services.
func (d Deps) EnsureCSRF(session *sessions.Session) string {
	return d.srv.EnsureCSRF(session)
}

// ValidateCSRF validates CSRF and returns an error on failure.
func (d Deps) ValidateCSRF(session *sessions.Session, token string) bool {
	return d.srv.ValidateCSRF(session, token)
}

// RenderTemplate renders a named template with the provided data.
func (d Deps) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return d.srv.RenderTemplate(w, name, data)
}

// CurrentUser returns the authenticated user from the request.
func (d Deps) CurrentUser(r *http.Request) (domain.User, error) {
	return d.srv.CurrentUser(r)
}

// CurrentIdentity returns the authenticated identity from the request.
func (d Deps) CurrentIdentity(r *http.Request) (domain.Identity, error) {
	return d.srv.CurrentIdentity(r)
}

// AuditAttempt records attempt as an audit event.
func (d Deps) AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string) {
	d.srv.AuditAttempt(ctx, actorID, action, target, meta)
}

// AuditOutcome records outcome as an audit event.
func (d Deps) AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string) {
	d.srv.AuditOutcome(ctx, actorID, action, target, err, meta)
}

// BaseURL returns scheme and host for the request.
func (d Deps) BaseURL(r *http.Request) string {
	return d.srv.BaseURL(r)
}

// Reserved returns the set of reserved path segments.
func (d Deps) Reserved() map[string]struct{} {
	return d.srv.Reserved()
}
