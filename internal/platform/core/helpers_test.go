package core

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
)

// TestBaseURL verifies base URL behavior.
func TestBaseURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)
	if got := BaseURL(req); got != "http://example.com" {
		t.Fatalf("expected http base URL, got %q", got)
	}

	secure := httptest.NewRequest(http.MethodGet, "https://example.com/path", nil)
	secure.TLS = &tls.ConnectionState{}
	if got := BaseURL(secure); got != "https://example.com" {
		t.Fatalf("expected https base URL, got %q", got)
	}
}

// TestIsSafeRedirect verifies is safe redirect behavior.
func TestIsSafeRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/base", nil)

	if !IsSafeRedirect(req, "/settings") {
		t.Fatalf("expected relative redirect to be safe")
	}
	if !IsSafeRedirect(req, "https://example.com/path") {
		t.Fatalf("expected same-host redirect to be safe")
	}
	if IsSafeRedirect(req, "https://evil.example.com/") {
		t.Fatalf("expected different host to be unsafe")
	}
	if IsSafeRedirect(req, "javascript:alert(1)") {
		t.Fatalf("expected javascript URL to be unsafe")
	}
	if IsSafeRedirect(req, "") {
		t.Fatalf("expected empty target to be unsafe")
	}
}

// TestNormalizeDomain verifies normalize domain behavior.
func TestNormalizeDomain(t *testing.T) {
	if got := NormalizeDomain(" https://Example.com/path "); got != "example.com" {
		t.Fatalf("expected normalized domain, got %q", got)
	}
	if got := NormalizeDomain("http://example.com/"); got != "example.com" {
		t.Fatalf("expected normalized domain, got %q", got)
	}
	if got := NormalizeDomain(""); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

// TestShortHash verifies short hash behavior.
func TestShortHash(t *testing.T) {
	hash := ShortHash("pin", 8)
	if len(hash) != 8 {
		t.Fatalf("expected length 8, got %d", len(hash))
	}
	full := ShortHash("pin", 0)
	if len(full) != 64 {
		t.Fatalf("expected full hash length, got %d", len(full))
	}
}

// TestSessionUserID verifies session user ID behavior.
func TestSessionUserID(t *testing.T) {
	session := &sessions.Session{Values: map[interface{}]interface{}{"user_id": 42}}
	if id, ok := SessionUserID(session); !ok || id != 42 {
		t.Fatalf("expected int user id")
	}
	session.Values["user_id"] = int64(7)
	if id, ok := SessionUserID(session); !ok || id != 7 {
		t.Fatalf("expected int64 user id")
	}
	session.Values["user_id"] = float64(9)
	if id, ok := SessionUserID(session); !ok || id != 9 {
		t.Fatalf("expected float64 user id")
	}
	delete(session.Values, "user_id")
	if _, ok := SessionUserID(session); ok {
		t.Fatalf("expected missing user id")
	}
}
