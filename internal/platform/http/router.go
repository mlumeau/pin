package http

import (
	"net/http"

	"pin/internal/features/admin"
	"pin/internal/features/auth"
	"pin/internal/features/domains"
	"pin/internal/features/federation"
	"pin/internal/features/health"
	"pin/internal/features/invites"
	"pin/internal/features/mcp"
	"pin/internal/features/oauth"
	"pin/internal/features/passkeys"
	"pin/internal/features/profilepicture"
	"pin/internal/features/public"
	pinserver "pin/internal/platform/server"
	"pin/internal/platform/wiring"
)

// Routes builds the HTTP mux using server handlers. As features move out of
// internal/app, routing will live here.
func Routes(s *pinserver.Server) http.Handler {
	mux := http.NewServeMux()
	register := func(pattern string, handler http.Handler) {
		s.RegisterRoute(mux, pattern, handler)
	}

	cfg := s.Config()
	deps := wiring.NewDeps(s)
	register("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))

	public.Register(mux, s, deps)
	federation.Register(mux, s, deps, deps, cfg)
	health.Register(mux, s, cfg)
	profilepicture.Register(mux, s, deps, cfg)
	auth.Register(mux, s, deps)
	admin.Register(mux, s, deps)
	domains.Register(mux, s, deps)
	passkeys.Register(mux, s, deps)
	invites.Register(mux, s, deps)
	oauth.Register(mux, s, deps, oauth.Config{
		BaseURL:            cfg.BaseURL,
		GitHubClientID:     cfg.GitHubClientID,
		GitHubClientSecret: cfg.GitHubClientSecret,
		RedditClientID:     cfg.RedditClientID,
		RedditClientSecret: cfg.RedditClientSecret,
		RedditUserAgent:    cfg.RedditUserAgent,
		BlueskyPDS:         cfg.BlueskyPDS,
	})
	mcp.Register(mux, s, deps, mcp.Config{
		Enabled:  cfg.MCPEnabled,
		Token:    cfg.MCPToken,
		ReadOnly: cfg.MCPReadOnly,
	})
	adminHandler := admin.NewHandler(deps)
	register("/settings/security/private-identity/regenerate", http.HandlerFunc(s.RequireSession(adminHandler.PrivateIdentityRegenerate, "/login?next=/settings")))

	return s.WithSecurityHeaders(public.WithPrivateRateLimit(public.WithSetupRedirect(deps, mux)))
}
