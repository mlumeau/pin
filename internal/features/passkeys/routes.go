package passkeys

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register wires passkey routes.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}
	requireLogin := func(next http.HandlerFunc) http.HandlerFunc {
		return reg.RequireSession(next, "/login?next=/settings")
	}

	handler := NewHandler(deps)
	register("/passkeys/register/options", http.HandlerFunc(requireLogin(handler.RegisterOptions)))
	register("/passkeys/register/finish", http.HandlerFunc(requireLogin(handler.RegisterFinish)))
	register("/passkeys/delete", http.HandlerFunc(requireLogin(handler.Delete)))
	register("/passkeys/login/options", http.HandlerFunc(handler.LoginOptions))
	register("/passkeys/login/finish", http.HandlerFunc(handler.LoginFinish))
}
