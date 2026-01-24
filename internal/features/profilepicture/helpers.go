package profilepicture

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"pin/internal/domain"
	"pin/internal/platform/media"
)

// parseProfilePictureSize reads s/size from query params and clamps it to safe bounds.
func parseProfilePictureSize(r *http.Request) int {
	raw := r.URL.Query().Get("s")
	if raw == "" {
		raw = r.URL.Query().Get("size")
	}
	if raw == "" {
		return media.DefaultSize
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return media.DefaultSize
	}
	if parsed < media.MinSize {
		return media.MinSize
	}
	if parsed > media.MaxSize {
		return media.MaxSize
	}
	return parsed
}

// wantsJSON reports whether the request's Accept header prefers JSON.
func wantsJSON(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/json") || strings.Contains(accept, "json")
}

// activePictureID returns the active profile picture ID or zero when unset.
func activePictureID(identity domain.Identity) int64 {
	if identity.ProfilePictureID.Valid {
		return identity.ProfilePictureID.Int64
	}
	return 0
}

// WriteProfilePictureStoreError writes profile picture store error to the response/output.
func WriteProfilePictureStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, media.ErrCWebPUnavailable):
		http.Error(w, "WebP encoder unavailable", http.StatusServiceUnavailable)
	case errors.Is(err, media.ErrImageTooSmall):
		http.Error(w, "Image too small", http.StatusBadRequest)
	default:
		http.Error(w, "Failed to process profile picture", http.StatusInternalServerError)
	}
}

// resolveProfilePictureFormat selects an output format based on request and Accept header.
func resolveProfilePictureFormat(r *http.Request, format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "png", "jpeg", "jpg", "webp":
		return format
	default:
		accept := strings.ToLower(r.Header.Get("Accept"))
		if strings.Contains(accept, "image/webp") {
			return "webp"
		}
		return "png"
	}
}

// serveProfilePictureWithFormat currently serves the on-disk file without transcoding.
// The original behavior cached resized WebP and optionally produced alt formats; this
// keeps the API stable while the refactor progresses.
func serveProfilePictureWithFormat(w http.ResponseWriter, r *http.Request, path string, format string) {
	switch format {
	case "jpeg", "jpg":
		w.Header().Set("Content-Type", "image/jpeg")
	case "png":
		w.Header().Set("Content-Type", "image/png")
	default:
		w.Header().Set("Content-Type", "image/webp")
	}
	http.ServeFile(w, r, path)
}
