package mcp

import (
	"net/http"

	"pin/internal/platform/transport"
)

// Register registers routes and handlers.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies, cfg Config) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}
	handler := NewHandler(cfg, deps)
	register("/mcp", handler)
}
