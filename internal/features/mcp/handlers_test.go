package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestServeHTTPMethodNotAllowed verifies serve HTTP method not allowed behavior.
func TestServeHTTPMethodNotAllowed(t *testing.T) {
	handler := NewHandler(Config{Enabled: true}, nil)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

// TestServeHTTPDisabled verifies serve HTTP disabled behavior.
func TestServeHTTPDisabled(t *testing.T) {
	handler := NewHandler(Config{Enabled: false}, nil)
	body := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// TestServeHTTPUnauthorized verifies serve HTTP unauthorized behavior.
func TestServeHTTPUnauthorized(t *testing.T) {
	handler := NewHandler(Config{Enabled: true, Token: "secret"}, nil)
	body := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != -32001 {
		t.Fatalf("expected unauthorized error code, got %+v", resp.Error)
	}
}
