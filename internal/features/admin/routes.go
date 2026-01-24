package admin

import (
	"net/http"

	"pin/internal/features/audit"
	"pin/internal/features/users"
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
	auditHandler := audit.NewHandler(deps)
	usersHandler := users.NewHandler(deps)

	register("/admin", http.HandlerFunc(requireLogin(handler.Root)))
	register("/admin/", http.HandlerFunc(requireLogin(handler.Root)))
	register("/settings", http.HandlerFunc(requireLogin(handler.Root)))
	register("/settings/", http.HandlerFunc(requireLogin(handler.Root)))
	register("/settings/identity", http.HandlerFunc(requireLogin(handler.Profile)))
	register("/settings/profile", http.HandlerFunc(requireLogin(handler.Profile)))
	register("/settings/security", http.HandlerFunc(requireLogin(handler.Security)))
	register("/settings/appearance", http.HandlerFunc(requireLogin(handler.Appearance)))
	register("/settings/admin/audit-log/download", http.HandlerFunc(requireLogin(auditHandler.Download)))
	register("/settings/admin/server", http.HandlerFunc(requireLogin(handler.Server)))
	register("/settings/admin/users/", http.HandlerFunc(requireLogin(usersHandler.User)))
	register("/settings/admin/users", http.HandlerFunc(requireLogin(usersHandler.Users)))
}
