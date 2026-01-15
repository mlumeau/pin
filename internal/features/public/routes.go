package public

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register wires public-facing routes.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}

	handler := NewHandler(deps)

	register("/", http.HandlerFunc(handler.Index))
	register("/landing", http.HandlerFunc(handler.Landing))
	register("/qr", http.HandlerFunc(handler.QR))
	register("/setup", http.HandlerFunc(handler.Setup))
}
