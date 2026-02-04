package sqlitestore

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestInitDBEnforcesCaseInsensitiveHandleUniqueness(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := InitDB(db); err != nil {
		t.Fatalf("init db: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO user (id, role, password_hash, totp_secret) VALUES (1, 'user', 'h1', 's1'), (2, 'user', 'h2', 's2')`); err != nil {
		t.Fatalf("insert users: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO identity (user_id, handle) VALUES (1, 'Alice')`); err != nil {
		t.Fatalf("insert first identity: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO identity (user_id, handle) VALUES (2, 'alice')`); err == nil {
		t.Fatalf("expected case-insensitive uniqueness violation")
	}
}
