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

func (d Deps) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return d.srv.GetSession(r, name)
}

func (d Deps) EnsureCSRF(session *sessions.Session) string {
	return d.srv.EnsureCSRF(session)
}

func (d Deps) ValidateCSRF(session *sessions.Session, token string) bool {
	return d.srv.ValidateCSRF(session, token)
}

func (d Deps) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return d.srv.RenderTemplate(w, name, data)
}

func (d Deps) CurrentUser(r *http.Request) (domain.User, error) {
	return d.srv.CurrentUser(r)
}

func (d Deps) CurrentIdentity(r *http.Request) (domain.Identity, error) {
	return d.srv.CurrentIdentity(r)
}

func (d Deps) AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string) {
	d.srv.AuditAttempt(ctx, actorID, action, target, meta)
}

func (d Deps) AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string) {
	d.srv.AuditOutcome(ctx, actorID, action, target, err, meta)
}

func (d Deps) BaseURL(r *http.Request) string {
	return d.srv.BaseURL(r)
}

func (d Deps) Reserved() map[string]struct{} {
	return d.srv.Reserved()
}
