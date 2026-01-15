package testutil

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
	"pin/internal/config"
	"pin/internal/features/identity"
	platformserver "pin/internal/platform/server"
	sqlitestore "pin/internal/platform/storage/sqlite"
)

func TestConfig(t *testing.T) config.Config {
	t.Helper()
	tempDir := t.TempDir()
	uploadsDir := filepath.Join(tempDir, "uploads")
	return config.Config{
		SecretKey:         []byte("test-secret"),
		StaticDir:         tempDir,
		UploadsDir:        uploadsDir,
		ProfilePictureDir: filepath.Join(uploadsDir, "profile-pictures"),
		AllowedExts:       map[string]bool{".png": true},
		BaseURL:           "http://example.test",
		CookieSameSite:    http.SameSiteLaxMode,
	}
}

func NewServer(t *testing.T) *platformserver.Server {
	t.Helper()
	cfg := TestConfig(t)

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := sqlitestore.InitDB(db); err != nil {
		t.Fatalf("init db: %v", err)
	}

	srv, err := platformserver.NewServer(cfg, db, identity.TemplateFuncs())
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	return srv
}
