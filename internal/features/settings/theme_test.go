package settings

import "testing"

// TestNormalizeThemeChoice verifies normalize theme choice behavior.
func TestNormalizeThemeChoice(t *testing.T) {
	if got := NormalizeThemeChoice(""); got != DefaultThemeName {
		t.Fatalf("expected default theme, got %q", got)
	}
	if got := NormalizeThemeChoice("Noir"); got != "noir" {
		t.Fatalf("expected noir, got %q", got)
	}
	if got := NormalizeThemeChoice(DefaultCustomThemeName); got != DefaultCustomThemeName {
		t.Fatalf("expected default custom theme, got %q", got)
	}
	if got := NormalizeThemeChoice("unknown"); got != DefaultThemeName {
		t.Fatalf("expected default theme for unknown, got %q", got)
	}
}

// TestThemeCustomCSSURL verifies theme custom cssurl behavior.
func TestThemeCustomCSSURL(t *testing.T) {
	if got := ThemeCustomCSSURL("themes/custom.css"); got != "/static/uploads/themes/custom.css" {
		t.Fatalf("unexpected css url: %q", got)
	}
	if got := ThemeCustomCSSURL(" "); got != "" {
		t.Fatalf("expected empty css url, got %q", got)
	}
}
