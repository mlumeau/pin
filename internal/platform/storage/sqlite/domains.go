package sqlitestore

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"pin/internal/domain"
)

func ListDomainVerifications(ctx context.Context, db *sql.DB, identityID int) ([]domain.DomainVerification, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, identity_id, domain, token, verified_at, created_at FROM domain_verification WHERE identity_id = ? ORDER BY domain", identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DomainVerification
	for rows.Next() {
		var row domain.DomainVerification
		var verified sql.NullString
		var created string
		if err := rows.Scan(&row.ID, &row.IdentityID, &row.Domain, &row.Token, &verified, &created); err != nil {
			return nil, err
		}
		if verified.Valid {
			if parsed, err := time.Parse(time.RFC3339, verified.String); err == nil {
				row.VerifiedAt = sql.NullTime{Time: parsed, Valid: true}
			}
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, row)
	}
	return out, nil
}

func UpsertDomainVerification(ctx context.Context, db *sql.DB, identityID int, domain, token string) error {
	_, err := db.ExecContext(
		ctx,
		"INSERT INTO domain_verification (identity_id, domain, token, created_at) VALUES (?, ?, ?, ?) ON CONFLICT(identity_id, domain) DO UPDATE SET token = excluded.token",
		identityID, domain, token, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func DeleteDomainVerification(ctx context.Context, db *sql.DB, identityID int, domain string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM domain_verification WHERE identity_id = ? AND domain = ?", identityID, domain)
	return err
}

func MarkDomainVerified(ctx context.Context, db *sql.DB, identityID int, domain string) error {
	_, err := db.ExecContext(ctx, "UPDATE domain_verification SET verified_at = ? WHERE identity_id = ? AND domain = ?", time.Now().UTC().Format(time.RFC3339), identityID, domain)
	return err
}

func ListVerifiedDomains(ctx context.Context, db *sql.DB, identityID int) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT domain FROM domain_verification WHERE identity_id = ? AND verified_at IS NOT NULL", identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, err
		}
		if strings.TrimSpace(domain) == "" {
			continue
		}
		out = append(out, domain)
	}
	return out, nil
}

func HasDomainVerification(ctx context.Context, db *sql.DB, identityID int, domain string) (bool, error) {
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM domain_verification WHERE identity_id = ? AND domain = ?", identityID, domain)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateIdentityVerifiedDomains(ctx context.Context, db *sql.DB, identityID int, domains []string) error {
	raw, err := json.Marshal(domains)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "UPDATE identity SET verified_domains = ?, updated_at = ? WHERE id = ?", string(raw), time.Now().UTC().Format(time.RFC3339), identityID)
	return err
}
