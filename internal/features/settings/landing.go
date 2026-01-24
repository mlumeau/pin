package settings

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"pin/internal/config"
)

const (
	landingModeGeneric   = "generic"
	landingModeProfile   = "profile"
	landingModeCustom    = "custom"
	landingModeKey       = "landing_mode"
	landingCustomPathKey = "landing_custom_path"
	landingUploadsSubdir = "landing"
)

const (
	LandingModeGeneric = landingModeGeneric
	LandingModeProfile = landingModeProfile
	LandingModeCustom  = landingModeCustom
)

// LandingSettings controls what renders on the root path.
type LandingSettings struct {
	Mode       string
	CustomPath string
}

// LandingSettings returns the effective landing mode and custom path.
func (s Service) LandingSettings(ctx context.Context) LandingSettings {
	settings := LandingSettings{Mode: landingModeGeneric}
	values, err := s.store.GetSettings(ctx, landingModeKey, landingCustomPathKey)
	if err != nil {
		return settings
	}
	if value, ok := values[landingModeKey]; ok {
		settings.Mode = normalizeLandingMode(value)
	}
	if value, ok := values[landingCustomPathKey]; ok {
		settings.CustomPath = normalizeLandingCustomPath(value)
	}
	settings.Mode = normalizeLandingMode(settings.Mode)
	settings.CustomPath = normalizeLandingCustomPath(settings.CustomPath)
	return settings
}

// SaveLandingSettings normalizes and persists landing configuration.
func (s Service) SaveLandingSettings(ctx context.Context, settings LandingSettings) error {
	settings.Mode = normalizeLandingMode(settings.Mode)
	settings.CustomPath = normalizeLandingCustomPath(settings.CustomPath)

	if err := s.store.SetSettings(ctx, map[string]string{
		landingModeKey:       settings.Mode,
		landingCustomPathKey: settings.CustomPath,
	}); err != nil {
		return err
	}
	return nil
}

// LandingDir returns the upload directory for landing assets.
func LandingDir(cfg config.Config) string {
	return filepath.Join(cfg.UploadsDir, landingUploadsSubdir)
}

// LandingCustomURL returns a public URL for a landing asset.
func LandingCustomURL(path string) string {
	base := normalizeLandingCustomPath(path)
	if base == "" {
		return ""
	}
	return "/static/uploads/" + landingUploadsSubdir + "/" + url.PathEscape(base)
}

// LandingCustomPath returns the full filesystem path for a landing asset.
func LandingCustomPath(cfg config.Config, filename string) string {
	base := normalizeLandingCustomPath(filename)
	if base == "" {
		return ""
	}
	return filepath.Join(LandingDir(cfg), base)
}

// RemoveLandingFile deletes the configured landing file if present.
func RemoveLandingFile(cfg config.Config, filename string) {
	if strings.TrimSpace(filename) == "" {
		return
	}
	_ = os.Remove(LandingCustomPath(cfg, filename))
}

// normalizeLandingMode normalizes the landing mode to a supported value.
func normalizeLandingMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case landingModeProfile:
		return landingModeProfile
	case landingModeCustom:
		return landingModeCustom
	default:
		return landingModeGeneric
	}
}

// normalizeLandingCustomPath sanitizes the landing filename to a base name.
func normalizeLandingCustomPath(path string) string {
	base := filepath.Base(strings.TrimSpace(path))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}
