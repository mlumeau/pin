package server

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"pin/internal/domain"
	"pin/internal/platform/core"
)

// EnsureCSRF ensures CSRF is initialized and available.
func (s *Server) EnsureCSRF(session *sessions.Session) string {
	return s.ensureCSRF(session)
}

// RenderTemplate renders a named template with the provided data.
func (s *Server) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return s.tmpl.ExecuteTemplate(w, name, data)
}

// Reserved returns the set of reserved path segments.
func (s *Server) Reserved() map[string]struct{} {
	return s.reserved
}

// GetSession returns the session.
func (s *Server) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return s.store.Get(r, name)
}

// ValidateCSRF validates CSRF and returns an error on failure.
func (s *Server) ValidateCSRF(session *sessions.Session, token string) bool {
	return s.validateCSRF(session, token)
}

// CurrentUser returns the authenticated user from the session.
func (s *Server) CurrentUser(r *http.Request) (domain.User, error) {
	session, _ := s.store.Get(r, "pin_session")
	id, ok := core.SessionUserID(session)
	if !ok {
		return domain.User{}, errNotLoggedIn
	}
	return s.repos.Users.GetUserByID(r.Context(), id)
}

// CurrentIdentity returns the authenticated identity from the session.
func (s *Server) CurrentIdentity(r *http.Request) (domain.Identity, error) {
	session, _ := s.store.Get(r, "pin_session")
	id, ok := core.SessionUserID(session)
	if !ok {
		return domain.Identity{}, errNotLoggedIn
	}
	return s.repos.Identities.GetIdentityByUserID(r.Context(), id)
}

// AuditAttempt records attempt as an audit event.
func (s *Server) AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string) {
	s.auditAttempt(ctx, actorID, action, target, meta)
}

// AuditOutcome records outcome as an audit event.
func (s *Server) AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string) {
	s.auditOutcome(ctx, actorID, action, target, err, meta)
}

// BaseURL returns scheme and host for the incoming request.
func (s *Server) BaseURL(r *http.Request) string {
	return core.BaseURL(r)
}
