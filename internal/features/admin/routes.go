package admin

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register wires admin/settings routes that are not feature-specific elsewhere.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}
	requireLogin := func(next http.HandlerFunc) http.HandlerFunc {
		return reg.RequireSession(next, "/login?next=/settings")
	}
	handler := NewHandler(deps)

	register("/admin", http.HandlerFunc(requireLogin(handler.Root)))
	register("/admin/", http.HandlerFunc(requireLogin(handler.Root)))
	register("/settings", http.HandlerFunc(requireLogin(handler.Root)))
	register("/settings/", http.HandlerFunc(requireLogin(handler.Root)))
	register("/settings/identity", http.HandlerFunc(requireLogin(handler.Profile)))
	register("/settings/profile", http.HandlerFunc(requireLogin(handler.Profile)))
	register("/settings/security", http.HandlerFunc(requireLogin(handler.Security)))
	register("/settings/appearance", http.HandlerFunc(requireLogin(handler.Appearance)))
	register("/settings/admin/audit-log/download", http.HandlerFunc(requireLogin(handler.AuditLogDownload)))
	register("/settings/admin/server", http.HandlerFunc(requireLogin(handler.Server)))
	register("/settings/admin/users/", http.HandlerFunc(requireLogin(handler.User)))
	register("/settings/admin/users", http.HandlerFunc(requireLogin(handler.Users)))
}
