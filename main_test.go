package main

import (
	"testing"

	pinhttp "pin/internal/platform/http"
	"pin/internal/testutil"
)

func TestMainWiring(t *testing.T) {
	testutil.ChdirRepoRoot(t)
	srv := testutil.NewServer(t)
	if handler := pinhttp.Routes(srv); handler == nil {
		t.Fatalf("expected router handler")
	}
}
