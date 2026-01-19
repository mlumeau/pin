package main

import (
	"testing"

	platformhttp "pin/internal/platform/http"
	"pin/internal/testutil"
)

func TestMainWiring(t *testing.T) {
	testutil.ChdirRepoRoot(t)
	srv := testutil.NewServer(t)
	if handler := platformhttp.Routes(srv); handler == nil {
		t.Fatalf("expected router handler")
	}
}
