package oauth

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register wires OAuth and social linking routes.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies, cfg Config) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}
	requireLogin := func(next http.HandlerFunc) http.HandlerFunc {
		return reg.RequireSession(next, "/login?next=/settings")
	}

	handler := NewHandler(cfg, deps)

	register("/oauth/github/start", http.HandlerFunc(requireLogin(handler.GitHubStart)))
	register("/oauth/github/callback", http.HandlerFunc(requireLogin(handler.GitHubCallback)))
	register("/oauth/reddit/start", http.HandlerFunc(requireLogin(handler.RedditStart)))
	register("/oauth/reddit/callback", http.HandlerFunc(requireLogin(handler.RedditCallback)))
	register("/settings/profile/social/bluesky", http.HandlerFunc(requireLogin(handler.BlueskyConnect)))
}
