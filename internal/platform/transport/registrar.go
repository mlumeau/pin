package transport

import "net/http"

// Registrar abstracts route registration and auth middleware.
type Registrar interface {
	RegisterRoute(mux *http.ServeMux, pattern string, handler http.Handler)
	RequireSession(next http.HandlerFunc, redirectTo string) http.HandlerFunc
}
