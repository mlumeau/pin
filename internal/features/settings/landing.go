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
	landingModeGeneric      = "generic"
	landingModeProfile      = "profile"
	landingModeCustom       = "custom"
	landingModeKey          = "landing_mode"
	landingCustomPathKey    = "landing_custom_path"
	legacyLandingCustomHTML = "landing_custom_html"
	landingUploadsSubdir    = "landing"
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

func (s Service) SaveLandingSettings(ctx context.Context, settings LandingSettings) error {
	settings.Mode = normalizeLandingMode(settings.Mode)
	settings.CustomPath = normalizeLandingCustomPath(settings.CustomPath)

	if err := s.store.SetSettings(ctx, map[string]string{
		landingModeKey:       settings.Mode,
		landingCustomPathKey: settings.CustomPath,
	}); err != nil {
		return err
	}
	_ = s.store.DeleteSetting(ctx, legacyLandingCustomHTML)
	return nil
}

func LandingDir(cfg config.Config) string {
	return filepath.Join(cfg.UploadsDir, landingUploadsSubdir)
}

func LandingCustomURL(path string) string {
	base := normalizeLandingCustomPath(path)
	if base == "" {
		return ""
	}
	return "/static/uploads/" + landingUploadsSubdir + "/" + url.PathEscape(base)
}

func LandingCustomPath(cfg config.Config, filename string) string {
	base := normalizeLandingCustomPath(filename)
	if base == "" {
		return ""
	}
	return filepath.Join(LandingDir(cfg), base)
}

func RemoveLandingFile(cfg config.Config, filename string) {
	if strings.TrimSpace(filename) == "" {
		return
	}
	_ = os.Remove(LandingCustomPath(cfg, filename))
}

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

func normalizeLandingCustomPath(path string) string {
	base := filepath.Base(strings.TrimSpace(path))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}
