package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	platformhttp "pin/internal/platform/http"
	"pin/internal/testutil"
)

func TestRoutesSetupPage(t *testing.T) {
	testutil.ChdirRepoRoot(t)
	srv := testutil.NewServer(t)
	handler := platformhttp.Routes(srv)

	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
