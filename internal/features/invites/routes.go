package invites

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register wires invite routes.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}
	requireLogin := func(next http.HandlerFunc) http.HandlerFunc {
		return reg.RequireSession(next, "/login?next=/settings")
	}

	handler := NewHandler(deps)
	register("/settings/admin/invites/create", http.HandlerFunc(requireLogin(handler.Create)))
	register("/settings/admin/invites/delete", http.HandlerFunc(requireLogin(handler.Delete)))
	register("/invite/", http.HandlerFunc(handler.Invite))
}
