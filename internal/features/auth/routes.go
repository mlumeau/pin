package auth

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register registers routes and handlers.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}

	handler := NewHandler(deps)
	register("/login", http.HandlerFunc(handler.Login))
	register("/logout", http.HandlerFunc(handler.Logout))
}
