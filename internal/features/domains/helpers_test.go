package domains

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNormalizeDomain verifies normalize domain behavior.
func TestNormalizeDomain(t *testing.T) {
	if got := normalizeDomain(" https://Example.com/path "); got != "example.com" {
		t.Fatalf("expected normalized domain, got %q", got)
	}
	if got := normalizeDomain("http://example.com/"); got != "example.com" {
		t.Fatalf("expected normalized domain, got %q", got)
	}
	if got := normalizeDomain(""); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

// TestParseDomains verifies parse domains behavior.
func TestParseDomains(t *testing.T) {
	values := parseDomains("Example.com, test.com\nFoo.com")
	if len(values) != 3 {
		t.Fatalf("expected 3 domains, got %d", len(values))
	}
	if values[0] != "example.com" || values[1] != "test.com" || values[2] != "foo.com" {
		t.Fatalf("unexpected parsed domains: %v", values)
	}
}

// TestWantsJSON verifies wants JSON behavior.
func TestWantsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	if !wantsJSON(req) {
		t.Fatalf("expected json accept")
	}
	req.Header.Set("Accept", "text/html")
	if wantsJSON(req) {
		t.Fatalf("expected non-json accept")
	}
	req.Header.Set("Accept", "application/vnd.api+json")
	if !wantsJSON(req) {
		t.Fatalf("expected json accept for vendor type")
	}
}
