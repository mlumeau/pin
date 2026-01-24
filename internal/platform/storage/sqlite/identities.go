package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"pin/internal/domain"
)

// GetIdentityByID returns identity by ID.
func GetIdentityByID(ctx context.Context, db *sql.DB, id int) (domain.Identity, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT id, user_id, handle, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(updated_at,'') FROM identity WHERE id = ?`,
		id,
	)
	return scanIdentity(row)
}

// GetIdentityByHandle returns identity by handle.
func GetIdentityByHandle(ctx context.Context, db *sql.DB, handle string) (domain.Identity, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT id, user_id, handle, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(updated_at,'') FROM identity WHERE lower(handle) = lower(?)`,
		handle,
	)
	return scanIdentity(row)
}

// GetIdentityByPrivateToken returns identity by private token.
func GetIdentityByPrivateToken(ctx context.Context, db *sql.DB, token string) (domain.Identity, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT id, user_id, handle, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(updated_at,'') FROM identity WHERE private_token = ?`,
		token,
	)
	return scanIdentity(row)
}

// GetIdentityByUserID returns identity by user ID.
func GetIdentityByUserID(ctx context.Context, db *sql.DB, userID int) (domain.Identity, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT id, user_id, handle, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(updated_at,'') FROM identity WHERE user_id = ?`,
		userID,
	)
	return scanIdentity(row)
}

// GetOwnerIdentity returns the owner identity in the SQLite store.
func GetOwnerIdentity(ctx context.Context, db *sql.DB) (domain.Identity, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT identity.id, identity.user_id, identity.handle, COALESCE(identity.email,''), COALESCE(identity.display_name,''), COALESCE(identity.bio,''), COALESCE(identity.organization,''), COALESCE(identity.job_title,''), COALESCE(identity.birthdate,''), COALESCE(identity.languages,''), COALESCE(identity.phone,''), COALESCE(identity.address,''), COALESCE(identity.custom_fields,'{}'), COALESCE(identity.visibility,''), COALESCE(identity.private_token,''), COALESCE(identity.links,'[]'), COALESCE(identity.social_profiles,'[]'), COALESCE(identity.wallets,'{}'), COALESCE(identity.public_keys,'{}'), COALESCE(identity.location,''), COALESCE(identity.website,''), COALESCE(identity.pronouns,''), COALESCE(identity.verified_domains,'[]'), COALESCE(identity.atproto_handle,''), COALESCE(identity.atproto_did,''), COALESCE(identity.timezone,''), identity.profile_picture_id, COALESCE(identity.updated_at,'') FROM identity JOIN user ON identity.user_id = user.id WHERE user.role = 'owner' ORDER BY identity.id LIMIT 1`,
	)
	return scanIdentity(row)
}

// ListIdentities returns the identities list in the SQLite store.
func ListIdentities(ctx context.Context, db *sql.DB) ([]domain.Identity, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, handle, COALESCE(email,''), COALESCE(display_name,''), COALESCE(updated_at,'') FROM identity ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var identities []domain.Identity
	for rows.Next() {
		var identity domain.Identity
		var updatedAt string
		if err := rows.Scan(&identity.ID, &identity.UserID, &identity.Handle, &identity.Email, &identity.DisplayName, &updatedAt); err != nil {
			return nil, err
		}
		if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			identity.UpdatedAt = parsed
		}
		identities = append(identities, identity)
	}
	return identities, nil
}

// ListIdentitiesPaged returns a page of identities paged using limit/offset in the SQLite store.
func ListIdentitiesPaged(ctx context.Context, db *sql.DB, query, sort, dir string, limit, offset int) ([]domain.Identity, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	sortKey := "id"
	if strings.ToLower(sort) == "handle" {
		sortKey = "handle"
	} else if strings.ToLower(sort) == "display_name" {
		sortKey = "display_name"
	} else if strings.ToLower(sort) == "email" {
		sortKey = "email"
	} else if strings.ToLower(sort) == "updated" {
		sortKey = "updated_at"
	}
	sortDir := "ASC"
	if strings.ToLower(dir) == "desc" {
		sortDir = "DESC"
	}
	where := ""
	args := []interface{}{}
	if strings.TrimSpace(query) != "" {
		where = " WHERE handle LIKE ? OR email LIKE ? OR display_name LIKE ?"
		pattern := "%" + strings.TrimSpace(query) + "%"
		args = append(args, pattern, pattern, pattern)
	}

	countQuery := "SELECT COUNT(*) FROM identity" + where
	row := db.QueryRowContext(ctx, countQuery, args...)
	var total int
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}

	queryStr := fmt.Sprintf("SELECT id, user_id, handle, COALESCE(email,''), COALESCE(display_name,''), COALESCE(updated_at,'') FROM identity%s ORDER BY %s %s LIMIT ? OFFSET ?", where, sortKey, sortDir)
	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var identities []domain.Identity
	for rows.Next() {
		var identity domain.Identity
		var updatedAt string
		if err := rows.Scan(&identity.ID, &identity.UserID, &identity.Handle, &identity.Email, &identity.DisplayName, &updatedAt); err != nil {
			return nil, 0, err
		}
		if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			identity.UpdatedAt = parsed
		}
		identities = append(identities, identity)
	}
	return identities, total, nil
}

// CreateIdentity creates identity using the supplied input in the SQLite store.
func CreateIdentity(ctx context.Context, db *sql.DB, identity domain.Identity) (int64, error) {
	if strings.TrimSpace(identity.Handle) == "" {
		return 0, errors.New("handle is required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := db.ExecContext(
		ctx,
		`INSERT INTO identity (user_id, handle, email, display_name, bio, organization, job_title, birthdate, languages, phone, address, custom_fields, visibility, private_token, links, social_profiles, wallets, public_keys, location, website, pronouns, verified_domains, atproto_handle, atproto_did, timezone, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		identity.UserID, identity.Handle, identity.Email, identity.DisplayName, identity.Bio, identity.Organization, identity.JobTitle, identity.Birthdate, identity.Languages, identity.Phone, identity.Address, identity.CustomFieldsJSON, identity.VisibilityJSON, identity.PrivateToken, identity.LinksJSON, identity.SocialProfilesJSON, identity.WalletsJSON, identity.PublicKeysJSON, identity.Location, identity.Website, identity.Pronouns, identity.VerifiedDomainsJSON, identity.ATProtoHandle, identity.ATProtoDID, identity.Timezone, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateIdentity updates identity using the supplied data in the SQLite store.
func UpdateIdentity(ctx context.Context, db *sql.DB, identity domain.Identity) error {
	_, err := db.ExecContext(
		ctx,
		`UPDATE identity SET handle = ?, email = ?, display_name = ?, bio = ?, organization = ?, job_title = ?, birthdate = ?, languages = ?, phone = ?, address = ?, custom_fields = ?, visibility = ?, private_token = ?, links = ?, social_profiles = ?, wallets = ?, public_keys = ?, location = ?, website = ?, pronouns = ?, verified_domains = ?, atproto_handle = ?, atproto_did = ?, timezone = ?, profile_picture_id = ?, updated_at = ? WHERE id = ?`,
		identity.Handle, identity.Email, identity.DisplayName, identity.Bio, identity.Organization, identity.JobTitle, identity.Birthdate, identity.Languages, identity.Phone, identity.Address, identity.CustomFieldsJSON, identity.VisibilityJSON, identity.PrivateToken, identity.LinksJSON, identity.SocialProfilesJSON, identity.WalletsJSON, identity.PublicKeysJSON, identity.Location, identity.Website, identity.Pronouns, identity.VerifiedDomainsJSON, identity.ATProtoHandle, identity.ATProtoDID, identity.Timezone, nullInt(identity.ProfilePictureID), time.Now().UTC().Format(time.RFC3339), identity.ID,
	)
	return err
}

// UpdatePrivateToken updates private token using the supplied data in the SQLite store.
func UpdatePrivateToken(ctx context.Context, db *sql.DB, identityID int, token string) error {
	_, err := db.ExecContext(ctx, "UPDATE identity SET private_token = ?, updated_at = ? WHERE id = ?", token, time.Now().UTC().Format(time.RFC3339), identityID)
	return err
}

// CheckHandleCollision checks handle collision and reports whether it matches.
func CheckHandleCollision(ctx context.Context, db *sql.DB, handle string, excludeID int) error {
	handle = strings.TrimSpace(handle)
	if handle == "" {
		return nil
	}
	row := db.QueryRowContext(
		ctx,
		"SELECT id FROM identity WHERE lower(handle) = lower(?) AND id != ? LIMIT 1",
		handle,
		excludeID,
	)
	var id int
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return errors.New("handle already exists")
}

// DeleteIdentity deletes identity in the SQLite store.
func DeleteIdentity(ctx context.Context, db *sql.DB, identityID int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM identity WHERE id = ?", identityID)
	return err
}

// scanIdentity scans a single-row result into an identity model.
func scanIdentity(row *sql.Row) (domain.Identity, error) {
	var identity domain.Identity
	var updatedAt string
	if err := row.Scan(
		&identity.ID,
		&identity.UserID,
		&identity.Handle,
		&identity.Email,
		&identity.DisplayName,
		&identity.Bio,
		&identity.Organization,
		&identity.JobTitle,
		&identity.Birthdate,
		&identity.Languages,
		&identity.Phone,
		&identity.Address,
		&identity.CustomFieldsJSON,
		&identity.VisibilityJSON,
		&identity.PrivateToken,
		&identity.LinksJSON,
		&identity.SocialProfilesJSON,
		&identity.WalletsJSON,
		&identity.PublicKeysJSON,
		&identity.Location,
		&identity.Website,
		&identity.Pronouns,
		&identity.VerifiedDomainsJSON,
		&identity.ATProtoHandle,
		&identity.ATProtoDID,
		&identity.Timezone,
		&identity.ProfilePictureID,
		&updatedAt,
	); err != nil {
		return domain.Identity{}, err
	}
	if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		identity.UpdatedAt = parsed
	}
	return identity, nil
}
