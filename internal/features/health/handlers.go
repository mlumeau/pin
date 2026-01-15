package health

import (
	"encoding/json"
	"image"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"pin/internal/config"
	"pin/internal/platform/media"
)

type Handler struct {
	cfg config.Config
}

// NewHandler builds a health handler with config dependencies.
func NewHandler(cfg config.Config) Handler {
	return Handler{cfg: cfg}
}

// ImageHealth reports availability of image processing utilities.
func (h Handler) ImageHealth(w http.ResponseWriter, _ *http.Request) {
	status := map[string]interface{}{
		"cwebp":           false,
		"profile_picture": map[string]interface{}{},
	}
	if _, err := exec.LookPath("cwebp"); err == nil {
		status["cwebp"] = true
	}
	test := map[string]interface{}{
		"decode_webp": false,
		"resize_ok":   false,
	}
	status["profile_picture"] = test
	tmpDir := filepath.Join(h.cfg.ProfilePictureDir, "cache")
	if err := os.MkdirAll(tmpDir, 0755); err == nil {
		tmpInput := filepath.Join(tmpDir, "health_input.png")
		tmpOutput := filepath.Join(tmpDir, "health_output.webp")
		_ = os.WriteFile(tmpInput, []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 31, 21, 196, 137, 0, 0, 0, 10, 73, 68, 65, 84, 120, 156, 99, 0, 1, 0, 0, 5, 0, 1, 13, 10, 44, 118, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}, 0644)
		if err := media.EncodeWebP(tmpInput, tmpOutput, 1, 1, 80); err == nil {
			if _, err := os.Stat(tmpOutput); err == nil {
				test["resize_ok"] = true
			}
		}
		if file, err := os.Open(tmpOutput); err == nil {
			defer file.Close()
			if _, _, err := image.Decode(file); err == nil {
				test["decode_webp"] = true
			}
		}
		_ = os.Remove(tmpInput)
		_ = os.Remove(tmpOutput)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(status)
}
