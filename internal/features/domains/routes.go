package domains

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register registers routes and handlers.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}
	requireLogin := func(next http.HandlerFunc) http.HandlerFunc {
		return reg.RequireSession(next, "/login?next=/settings")
	}

	handler := NewHandler(deps)
	register("/settings/profile/verified-domains/create", http.HandlerFunc(requireLogin(handler.Create)))
	register("/settings/profile/verified-domains/verify", http.HandlerFunc(requireLogin(handler.Verify)))
	register("/settings/profile/verified-domains/delete", http.HandlerFunc(requireLogin(handler.Delete)))
}
