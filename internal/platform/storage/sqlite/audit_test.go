package sqlitestore

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestAuditKeepsActorNameAfterUserDeletion(t *testing.T) {
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

	_, err = db.Exec(`INSERT INTO user (id, role, password_hash, totp_secret) VALUES (11, 'user', 'h', 's')`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.Exec(`INSERT INTO identity (user_id, handle) VALUES (11, 'actor-handle')`)
	if err != nil {
		t.Fatalf("insert identity: %v", err)
	}

	if err := WriteAuditLog(context.Background(), db, 11, "user.login", "session", map[string]string{"status": "ok"}); err != nil {
		t.Fatalf("write audit log: %v", err)
	}
	if err := DeleteUser(context.Background(), db, 11); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	logs, err := ListAuditLogs(context.Background(), db, 10, 0)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}
	if logs[0].ActorName != "actor-handle" {
		t.Fatalf("expected actor name snapshot to survive delete, got %q", logs[0].ActorName)
	}
}
