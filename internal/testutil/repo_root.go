package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ChdirRepoRoot changes the working directory to the repository root.
func ChdirRepoRoot(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to locate testutil file")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
}
