package sqlitestore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"pin/internal/domain"
)

func CheckIdentifierCollisions(ctx context.Context, db *sql.DB, identifiers []string, excludeID int) error {
	if len(identifiers) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(identifiers)+1)
	placeholders := make([]string, 0, len(identifiers))
	for _, ident := range identifiers {
		placeholders = append(placeholders, "?")
		args = append(args, strings.TrimSpace(ident))
	}
	query := "SELECT COUNT(*) FROM user_identifier WHERE identifier IN (" + strings.Join(placeholders, ",") + ")"
	if excludeID > 0 {
		query += " AND user_id != ?"
		args = append(args, excludeID)
	}
	row := db.QueryRowContext(ctx, query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("collision")
	}
	return nil
}

func UpsertUserIdentifiers(ctx context.Context, db *sql.DB, userID int, username string, aliases []string, email string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM user_identifier WHERE user_id = ?", userID); err != nil {
		_ = tx.Rollback()
		return err
	}
	ids := buildIdentifiers(username, aliases, email)
	for _, ident := range ids {
		if _, err := tx.ExecContext(ctx, "INSERT INTO user_identifier (identifier, user_id) VALUES (?, ?)", strings.TrimSpace(ident), userID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func FindUserByIdentifier(ctx context.Context, db *sql.DB, identifier string) (domain.User, error) {
	ident := strings.ToLower(strings.TrimSpace(identifier))
	if ident == "" {
		return domain.User{}, sql.ErrNoRows
	}
	row := db.QueryRowContext(ctx, "SELECT u.id, u.username, COALESCE(u.email,''), COALESCE(u.display_name,''), COALESCE(u.bio,''), COALESCE(u.organization,''), COALESCE(u.job_title,''), COALESCE(u.birthdate,''), COALESCE(u.languages,''), COALESCE(u.phone,''), COALESCE(u.address,''), COALESCE(u.custom_fields,'{}'), COALESCE(u.visibility,''), COALESCE(u.private_token,''), COALESCE(u.links,'[]'), COALESCE(u.aliases,'[]'), COALESCE(u.social_profiles,'[]'), COALESCE(u.wallets,'{}'), COALESCE(u.public_keys,'{}'), COALESCE(u.location,''), COALESCE(u.website,''), COALESCE(u.pronouns,''), COALESCE(u.verified_domains,'[]'), COALESCE(u.atproto_handle,''), COALESCE(u.atproto_did,''), COALESCE(u.timezone,''), u.profile_picture_id, COALESCE(u.role,'user'), u.password_hash, u.totp_secret, COALESCE(u.theme_profile,''), COALESCE(u.theme_custom_css_path,''), COALESCE(u.theme_custom_css_inline,''), COALESCE(u.updated_at,'') FROM user u JOIN user_identifier ui ON ui.user_id = u.id WHERE ui.identifier = ?", ident)
	var u domain.User
	var updatedAt string
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Bio, &u.Organization, &u.JobTitle, &u.Birthdate, &u.Languages, &u.Phone, &u.Address, &u.CustomFieldsJSON, &u.VisibilityJSON, &u.PrivateToken, &u.LinksJSON, &u.AliasesJSON, &u.SocialProfilesJSON, &u.WalletsJSON, &u.PublicKeysJSON, &u.Location, &u.Website, &u.Pronouns, &u.VerifiedDomainsJSON, &u.ATProtoHandle, &u.ATProtoDID, &u.Timezone, &u.ProfilePictureID, &u.Role, &u.PasswordHash, &u.TOTPSecret, &u.ThemeProfile, &u.ThemeCustomCSSPath, &u.ThemeCustomCSSInline, &updatedAt); err != nil {
		return domain.User{}, err
	}
	if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		u.UpdatedAt = parsed
	}
	return u, nil
}

// buildIdentifiers mirrors the logic from app/utils.go to keep collision rules in storage.
func buildIdentifiers(username string, aliases []string, email string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
		hash := sha256Hex(value)
		if _, ok := seen[hash]; !ok {
			seen[hash] = struct{}{}
			out = append(out, hash)
		}
	}
	add(username)
	add(email)
	for _, alias := range aliases {
		add(alias)
	}
	return out
}

func sha256Hex(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])
}
