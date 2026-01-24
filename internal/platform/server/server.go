package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gorilla/sessions"
	"pin/internal/config"
	"pin/internal/contracts"
	"pin/internal/platform/core"
	"pin/internal/platform/media"
	sqlitestore "pin/internal/platform/storage/sqlite"
)

// Server bundles dependencies for HTTP handlers.
type Server struct {
	cfg      config.Config
	db       *sql.DB
	store    *sessions.CookieStore
	tmpl     *template.Template
	reserved map[string]struct{}
	repos    contracts.Repos
}

// NewServer configures dependencies and templates for handlers using the default SQLite-backed repositories.
func NewServer(cfg config.Config, db *sql.DB, extraFuncs ...template.FuncMap) (*Server, error) {
	return NewServerWithRepos(cfg, db, sqlitestore.NewRepos(db), extraFuncs...)
}

// NewServerWithRepos constructs a new server with repos.
func NewServerWithRepos(cfg config.Config, db *sql.DB, repos contracts.Repos, extraFuncs ...template.FuncMap) (*Server, error) {
	store := sessions.NewCookieStore(cfg.SecretKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: cfg.CookieSameSite,
	}

	funcs := template.FuncMap{
		"toJSON":   toJSON,
		"toUpper":  strings.ToUpper,
		"urlquery": url.QueryEscape,
		"dict": func(values ...interface{}) map[string]interface{} {
			out := make(map[string]interface{}, len(values)/2)
			for i := 0; i+1 < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				out[key] = values[i+1]
			}
			return out
		},
		"firstNonEmpty": core.FirstNonEmpty,
	}
	applyTemplateFuncs(funcs, extraFuncs...)
	tmpl := template.New("").Funcs(funcs)

	// Load templates from feature-specific and shared locations
	templatePatterns := []string{
		filepath.Join("templates", "public", "*.html"),
		filepath.Join("templates", "auth", "*.html"),
		filepath.Join("templates", "settings", "*.html"),
	}

	for _, pattern := range templatePatterns {
		if _, err := tmpl.ParseGlob(pattern); err != nil {
			return nil, fmt.Errorf("failed to parse templates from %s: %w", pattern, err)
		}
	}

	if err := media.EnsureDefaultWebP(cfg.StaticDir); err != nil {
		return nil, err
	}

	return &Server{
		cfg:      cfg,
		db:       db,
		store:    store,
		tmpl:     tmpl,
		reserved: map[string]struct{}{},
		repos:    repos,
	}, nil
}

// Routes builds the HTTP route tree.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	return s.WithSecurityHeaders(mux)
}

// register registers routes and handlers.
func (s *Server) register(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, handler)
	segment := routeSegment(pattern)
	if segment != "" {
		s.reserved[segment] = struct{}{}
	}
}

// RegisterRoute registers routes and handlers for route.
func (s *Server) RegisterRoute(mux *http.ServeMux, pattern string, handler http.Handler) {
	s.register(mux, pattern, handler)
}

// WithSecurityHeaders wraps the handler with additional behavior.
func (s *Server) WithSecurityHeaders(next http.Handler) http.Handler {
	return s.withSecurityHeaders(next)
}

// Config returns a copy of the server configuration.
func (s *Server) Config() config.Config {
	return s.cfg
}

// Repos returns the repository bundle for storage access.
func (s *Server) Repos() contracts.Repos {
	return s.repos
}

// routeSegment returns the first URL path segment from a route pattern.
func routeSegment(pattern string) string {
	if pattern == "" || pattern == "/" {
		return ""
	}
	trimmed := strings.TrimPrefix(pattern, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}

// withSecurityHeaders wraps the handler with additional behavior.
func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}

// RequireSession wraps a handler with session enforcement and redirect behavior.
func (s *Server) RequireSession(next http.HandlerFunc, redirectTo string) http.HandlerFunc {
	return s.requireSession(next, redirectTo)
}

// requireSession checks for a session user_id and redirects when missing.
func (s *Server) requireSession(next http.HandlerFunc, redirectTo string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.store.Get(r, "pin_session")
		if _, ok := core.SessionUserID(session); !ok {
			if redirectTo == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			http.Redirect(w, r, redirectTo, http.StatusFound)
			return
		}
		next(w, r)
	}
}

// toJSON returns JSON-safe template content.
func toJSON(value interface{}) template.JS {
	b, err := json.Marshal(value)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}

// ensureCSRF ensures CSRF is initialized and available.
func (s *Server) ensureCSRF(session *sessions.Session) string {
	if token, ok := session.Values["csrf_token"].(string); ok && token != "" {
		return token
	}
	token := core.RandomToken(32)
	session.Values["csrf_token"] = token
	return token
}

// validateCSRF checks the submitted CSRF token unless disabled by config.
func (s *Server) validateCSRF(session *sessions.Session, token string) bool {
	if s.cfg.DisableCSRF {
		return true
	}
	stored, _ := session.Values["csrf_token"].(string)
	if stored == "" || token == "" {
		return false
	}
	return core.SubtleCompare(stored, token)
}
