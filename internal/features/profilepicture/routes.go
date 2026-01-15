package profilepicture

import (
	"net/http"

	"pin/internal/config"
	"pin/internal/platform/transport"
)

// Register wires profile picture routes.
func Register(mux *http.ServeMux, reg transport.Registrar, deps Dependencies, cfg config.Config) {
	register := func(pattern string, handler http.Handler) {
		reg.RegisterRoute(mux, pattern, handler)
	}

	handler := NewHandler(Config{
		ProfilePictureDir: cfg.ProfilePictureDir,
		StaticDir:         cfg.StaticDir,
		AllowedExts:       cfg.AllowedExts,
		MaxUploadBytes:    cfg.MaxUploadBytes,
		CacheAltFormats:   cfg.CacheAltFormats,
	}, deps)

	register("/profile-picture", http.HandlerFunc(handler.ProfilePictureRoot))
	register("/profile-picture/", http.HandlerFunc(handler.ProfilePicture))

	requireLogin := func(next http.HandlerFunc) http.HandlerFunc {
		return reg.RequireSession(next, "/login?next=/settings")
	}
	register("/settings/profile/profile-picture/select", http.HandlerFunc(requireLogin(handler.Select)))
	register("/settings/profile/profile-picture/delete", http.HandlerFunc(requireLogin(handler.Delete)))
	register("/settings/profile/profile-picture/alt", http.HandlerFunc(requireLogin(handler.UpdateAlt)))
	register("/settings/profile/profile-picture/upload", http.HandlerFunc(requireLogin(handler.Upload)))
}
