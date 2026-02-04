package sqlitestore

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestListInvitesParsesUsedFields(t *testing.T) {
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

	_, err = db.Exec(`INSERT INTO invite (token, role, created_by, created_at, used_at, used_by, used_by_name) VALUES ('tok-1', 'user', 1, '2026-02-04T12:00:00Z', '2026-02-04T12:05:00Z', 42, 'alice')`)
	if err != nil {
		t.Fatalf("insert invite: %v", err)
	}

	invites, err := ListInvites(context.Background(), db)
	if err != nil {
		t.Fatalf("list invites: %v", err)
	}
	if len(invites) != 1 {
		t.Fatalf("expected 1 invite, got %d", len(invites))
	}
	if !invites[0].UsedAt.Valid {
		t.Fatalf("expected used_at to be set")
	}
	if !invites[0].UsedBy.Valid || invites[0].UsedBy.Int64 != 42 {
		t.Fatalf("expected used_by=42, got %+v", invites[0].UsedBy)
	}
	if invites[0].UsedByName != "alice" {
		t.Fatalf("expected used_by_name=alice, got %q", invites[0].UsedByName)
	}
}

func TestMarkInviteUsedSetsUsedByAndUsedAt(t *testing.T) {
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

	_, err = db.Exec(`INSERT INTO invite (id, token, role, created_by, created_at) VALUES (1, 'tok-1', 'user', 1, '2026-02-04T12:00:00Z')`)
	if err != nil {
		t.Fatalf("insert invite: %v", err)
	}
	_, err = db.Exec(`INSERT INTO user (id, role, password_hash, totp_secret) VALUES (42, 'user', 'h', 's')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.Exec(`INSERT INTO identity (user_id, handle) VALUES (42, 'alice')`)
	if err != nil {
		t.Fatalf("insert identity: %v", err)
	}

	if err := MarkInviteUsed(context.Background(), db, 1, 42); err != nil {
		t.Fatalf("mark invite used: %v", err)
	}

	row := db.QueryRow(`SELECT used_at, used_by, used_by_name FROM invite WHERE id = 1`)
	var usedAt sql.NullString
	var usedBy sql.NullInt64
	var usedByName sql.NullString
	if err := row.Scan(&usedAt, &usedBy, &usedByName); err != nil {
		t.Fatalf("scan invite: %v", err)
	}
	if !usedAt.Valid || usedAt.String == "" {
		t.Fatalf("expected used_at to be set")
	}
	if !usedBy.Valid || usedBy.Int64 != 42 {
		t.Fatalf("expected used_by=42, got %+v", usedBy)
	}
	if !usedByName.Valid || usedByName.String != "alice" {
		t.Fatalf("expected used_by_name=alice, got %+v", usedByName)
	}
}

func TestInviteKeepsUsedByNameAfterUserDeletion(t *testing.T) {
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

	_, err = db.Exec(`INSERT INTO invite (id, token, role, created_by, created_at) VALUES (1, 'tok-1', 'user', 1, '2026-02-04T12:00:00Z')`)
	if err != nil {
		t.Fatalf("insert invite: %v", err)
	}
	_, err = db.Exec(`INSERT INTO user (id, role, password_hash, totp_secret) VALUES (7, 'user', 'h', 's')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.Exec(`INSERT INTO identity (user_id, handle) VALUES (7, 'invited-user')`)
	if err != nil {
		t.Fatalf("insert identity: %v", err)
	}

	if err := MarkInviteUsed(context.Background(), db, 1, 7); err != nil {
		t.Fatalf("mark invite used: %v", err)
	}
	if err := DeleteUser(context.Background(), db, 7); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	invites, err := ListInvites(context.Background(), db)
	if err != nil {
		t.Fatalf("list invites: %v", err)
	}
	if len(invites) != 1 {
		t.Fatalf("expected 1 invite, got %d", len(invites))
	}
	if invites[0].UsedByName != "invited-user" {
		t.Fatalf("expected used_by_name snapshot to survive delete, got %q", invites[0].UsedByName)
	}
}
