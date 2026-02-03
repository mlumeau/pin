package public

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type setupDeps struct {
	hasUser bool
	err     error
}

// HasUser reports whether user exists.
func (d setupDeps) HasUser(ctx context.Context) (bool, error) {
	return d.hasUser, d.err
}

// TestWithSetupRedirect verifies with setup redirect behavior.
func TestWithSetupRedirect(t *testing.T) {
	handler := WithSetupRedirect(setupDeps{hasUser: false}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/setup" {
		t.Fatalf("expected redirect to /setup, got %q", loc)
	}
}

// TestWithSetupRedirectAllowsSetup verifies with setup redirect allows setup behavior.
func TestWithSetupRedirectAllowsSetup(t *testing.T) {
	handler := WithSetupRedirect(setupDeps{hasUser: false}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected passthrough, got %d", rec.Code)
	}
}

func TestWithPrivateRateLimit(t *testing.T) {
	originalLimiter := privateRateLimiter
	privateRateLimiter = newRateLimiter(1, time.Minute)
	defer func() {
		privateRateLimiter = originalLimiter
	}()

	handler := WithPrivateRateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/p/abc/def", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected first request to pass, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/p/abc/def", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rate limited response, got %d", rec2.Code)
	}
}
