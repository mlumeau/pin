package server_test

import (
	"net/http/httptest"
	"testing"

	featuresettings "pin/internal/features/settings"
	"pin/internal/testutil"
)

// TestRenderLoginTemplate verifies render login template behavior.
func TestRenderLoginTemplate(t *testing.T) {
	testutil.ChdirRepoRoot(t)
	srv := testutil.NewServer(t)

	rec := httptest.NewRecorder()
	data := map[string]interface{}{
		"Error":     "",
		"Next":      "/settings",
		"CSRFToken": "token",
		"Theme": featuresettings.ThemeSettings{
			ProfileTheme: "classic",
			AdminTheme:   "classic",
		},
	}
	if err := srv.RenderTemplate(rec, "login.html", data); err != nil {
		t.Fatalf("render template: %v", err)
	}
	if rec.Body.Len() == 0 {
		t.Fatalf("expected template output")
	}
}
