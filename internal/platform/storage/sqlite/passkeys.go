package sqlitestore

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"time"

	"pin/internal/domain"

	"github.com/go-webauthn/webauthn/webauthn"
)

func ListPasskeys(ctx context.Context, db *sql.DB, userID int) ([]domain.Passkey, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, name, credential_id, credential_json, created_at, last_used_at FROM passkey WHERE user_id = ? ORDER BY id", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passkeys []domain.Passkey
	for rows.Next() {
		var pk domain.Passkey
		var created string
		var last sql.NullString
		if err := rows.Scan(&pk.ID, &pk.UserID, &pk.Name, &pk.CredentialID, &pk.CredentialJSON, &created, &last); err != nil {
			return nil, err
		}
		pk.CreatedAt, _ = time.Parse(time.RFC3339, created)
		if last.Valid {
			if parsed, err := time.Parse(time.RFC3339, last.String); err == nil {
				pk.LastUsedAt = sql.NullTime{Time: parsed, Valid: true}
			}
		}
		passkeys = append(passkeys, pk)
	}
	return passkeys, nil
}

func LoadPasskeyCredentials(ctx context.Context, db *sql.DB, userID int) ([]webauthn.Credential, error) {
	rows, err := db.QueryContext(ctx, "SELECT credential_json FROM passkey WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []webauthn.Credential
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var cred webauthn.Credential
		if err := json.Unmarshal([]byte(raw), &cred); err != nil {
			return nil, err
		}
		out = append(out, cred)
	}
	return out, nil
}

func InsertPasskey(ctx context.Context, db *sql.DB, userID int, name string, credential webauthn.Credential) error {
	payload, err := json.Marshal(credential)
	if err != nil {
		return err
	}
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	_, err = db.ExecContext(
		ctx,
		"INSERT INTO passkey (user_id, name, credential_id, credential_json, created_at) VALUES (?, ?, ?, ?, ?)",
		userID,
		name,
		credentialID,
		string(payload),
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func UpdatePasskeyCredential(ctx context.Context, db *sql.DB, userID int, credentialID string, credential webauthn.Credential) error {
	payload, err := json.Marshal(credential)
	if err != nil {
		return err
	}
	res, err := db.ExecContext(
		ctx,
		"UPDATE passkey SET credential_json = ?, last_used_at = ? WHERE user_id = ? AND credential_id = ?",
		string(payload),
		time.Now().UTC().Format(time.RFC3339),
		userID,
		credentialID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func DeletePasskey(ctx context.Context, db *sql.DB, userID, id int) error {
	res, err := db.ExecContext(ctx, "DELETE FROM passkey WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
