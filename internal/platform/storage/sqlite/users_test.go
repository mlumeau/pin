package sqlitestore

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestDeleteUserRemovesAllUserData(t *testing.T) {
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

	_, err = db.Exec(`INSERT INTO user (id, role, password_hash, totp_secret) VALUES (1, 'user', 'h', 's')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.Exec(`INSERT INTO identity (id, user_id, handle) VALUES (10, 1, 'to-delete')`)
	if err != nil {
		t.Fatalf("insert identity: %v", err)
	}
	_, err = db.Exec(`INSERT INTO passkey (id, user_id, name, credential_id, credential_json, created_at) VALUES (1, 1, 'pk', 'cred-1', '{}', '2026-02-04T12:00:00Z')`)
	if err != nil {
		t.Fatalf("insert passkey: %v", err)
	}
	_, err = db.Exec(`INSERT INTO domain_verification (id, identity_id, domain, token, created_at) VALUES (1, 10, 'example.com', 'tok', '2026-02-04T12:00:00Z')`)
	if err != nil {
		t.Fatalf("insert domain verification: %v", err)
	}
	_, err = db.Exec(`INSERT INTO profile_picture (id, identity_id, filename, alt_text, created_at) VALUES (1, 10, 'avatar.png', '', '2026-02-04T12:00:00Z')`)
	if err != nil {
		t.Fatalf("insert profile picture: %v", err)
	}

	if err := DeleteUser(context.Background(), db, 1); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	assertCount(t, db, "user", 0)
	assertCount(t, db, "identity", 0)
	assertCount(t, db, "passkey", 0)
	assertCount(t, db, "domain_verification", 0)
	assertCount(t, db, "profile_picture", 0)

	identities, err := ListIdentities(context.Background(), db)
	if err != nil {
		t.Fatalf("list identities: %v", err)
	}
	if len(identities) != 0 {
		t.Fatalf("expected no identities, got %d", len(identities))
	}
}

func assertCount(t *testing.T, db *sql.DB, table string, expected int) {
	t.Helper()
	row := db.QueryRow("SELECT COUNT(*) FROM " + table)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != expected {
		t.Fatalf("expected %s count=%d, got %d", table, expected, count)
	}
}
