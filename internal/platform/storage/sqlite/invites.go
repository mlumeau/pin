package sqlitestore

import (
	"context"
	"database/sql"
	"time"

	"pin/internal/domain"
)

// CreateInvite creates invite using the supplied input in the SQLite store.
func CreateInvite(ctx context.Context, db *sql.DB, token, role string, createdBy int) error {
	_, err := db.ExecContext(ctx, "INSERT INTO invite (token, role, created_by, created_at) VALUES (?, ?, ?, ?)", token, role, createdBy, time.Now().UTC().Format(time.RFC3339))
	return err
}

// ListInvites returns the invites list in the SQLite store.
func ListInvites(ctx context.Context, db *sql.DB) ([]domain.Invite, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, token, role, created_by, created_at, used_at, used_by FROM invite ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []domain.Invite
	for rows.Next() {
		var invite domain.Invite
		var created string
		if err := rows.Scan(&invite.ID, &invite.Token, &invite.Role, &invite.CreatedBy, &created, &invite.UsedAt, &invite.UsedBy); err != nil {
			return nil, err
		}
		invite.CreatedAt, _ = time.Parse(time.RFC3339, created)
		invites = append(invites, invite)
	}
	return invites, nil
}

// GetInviteByToken returns invite by token.
func GetInviteByToken(ctx context.Context, db *sql.DB, token string) (domain.Invite, error) {
	row := db.QueryRowContext(ctx, "SELECT id, token, role, created_by, created_at, used_at, used_by FROM invite WHERE token = ? LIMIT 1", token)
	var invite domain.Invite
	var created string
	if err := row.Scan(&invite.ID, &invite.Token, &invite.Role, &invite.CreatedBy, &created, &invite.UsedAt, &invite.UsedBy); err != nil {
		return domain.Invite{}, err
	}
	invite.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return invite, nil
}

// MarkInviteUsed returns invite used.
func MarkInviteUsed(ctx context.Context, db *sql.DB, id int, usedBy int) error {
	_, err := db.ExecContext(ctx, "UPDATE invite SET used_at = ?, used_by = ? WHERE id = ?", time.Now().UTC().Format(time.RFC3339), usedBy, id)
	return err
}

// DeleteInvite deletes invite in the SQLite store.
func DeleteInvite(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM invite WHERE id = ?", id)
	return err
}
