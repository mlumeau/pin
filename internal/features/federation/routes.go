package federation

import (
	"net/http"

	"pin/internal/config"
	"pin/internal/features/public"
	"pin/internal/platform/transport"
)

// Register wires well-known and federation discovery routes.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies, publicDeps public.Dependencies, cfg config.Config) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}

	handler := NewHandler(cfg, deps)
	publicHandler := public.NewHandler(publicDeps)
	register("/.well-known/webfinger", http.HandlerFunc(handler.Webfinger))
	register("/.well-known/atproto-did", http.HandlerFunc(handler.AtprotoDID))
	register("/.well-known/pin-verify", http.HandlerFunc(handler.WellKnownPinVerify))
	register("/users/", http.HandlerFunc(handler.Actor))
	register("/p/", http.HandlerFunc(publicHandler.PrivateIdentity))
	register("/.well-known/pinc", http.HandlerFunc(handler.PincCapability))
	register("/.well-known/pinc/identity", http.HandlerFunc(handler.PincIdentitySchema))
}
