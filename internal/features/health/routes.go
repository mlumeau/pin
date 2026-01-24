package health

import (
	"net/http"

	"pin/internal/config"
	"pin/internal/platform/transport"
)

// Register registers routes and handlers.
func Register(mux *http.ServeMux, reg transport.Registrar, cfg config.Config) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}
	handler := NewHandler(cfg)
	register("/health/images", http.HandlerFunc(handler.ImageHealth))
}
