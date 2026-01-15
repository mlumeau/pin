package sqlitestore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"pin/internal/domain"
)

func LoadUser(ctx context.Context, db *sql.DB) (domain.User, error) {
	row := db.QueryRowContext(
		ctx,
		"SELECT id, username, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(aliases,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(role,'user'), password_hash, totp_secret, COALESCE(theme_profile,''), COALESCE(theme_custom_css_path,''), COALESCE(theme_custom_css_inline,''), COALESCE(updated_at,'') FROM user LIMIT 1",
	)
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

func UpdateUser(ctx context.Context, db *sql.DB, u domain.User) error {
	_, err := db.ExecContext(
		ctx,
		`UPDATE user SET username = ?, email = ?, display_name = ?, bio = ?, organization = ?, job_title = ?, birthdate = ?, languages = ?, phone = ?, address = ?, custom_fields = ?, visibility = ?, private_token = ?, links = ?, aliases = ?, social_profiles = ?, wallets = ?, public_keys = ?, location = ?, website = ?, pronouns = ?, verified_domains = ?, atproto_handle = ?, atproto_did = ?, timezone = ?, profile_picture_id = ?, role = ?, password_hash = ?, totp_secret = ?, theme_profile = ?, theme_custom_css_path = ?, theme_custom_css_inline = ?, updated_at = ? WHERE id = ?`,
		u.Username, u.Email, u.DisplayName, u.Bio, u.Organization, u.JobTitle, u.Birthdate, u.Languages, u.Phone, u.Address, u.CustomFieldsJSON, u.VisibilityJSON, u.PrivateToken, u.LinksJSON, u.AliasesJSON, u.SocialProfilesJSON, u.WalletsJSON, u.PublicKeysJSON, u.Location, u.Website, u.Pronouns, u.VerifiedDomainsJSON, u.ATProtoHandle, u.ATProtoDID, u.Timezone, nullInt(u.ProfilePictureID), u.Role, u.PasswordHash, u.TOTPSecret, u.ThemeProfile, u.ThemeCustomCSSPath, u.ThemeCustomCSSInline, time.Now().UTC().Format(time.RFC3339), u.ID,
	)
	return err
}

func GetUserByID(ctx context.Context, db *sql.DB, id int) (domain.User, error) {
	row := db.QueryRowContext(
		ctx,
		"SELECT id, username, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(aliases,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(role,'user'), password_hash, totp_secret, COALESCE(theme_profile,''), COALESCE(theme_custom_css_path,''), COALESCE(theme_custom_css_inline,''), COALESCE(updated_at,'') FROM user WHERE id = ?",
		id,
	)
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

func GetUserByUsername(ctx context.Context, db *sql.DB, username string) (domain.User, error) {
	row := db.QueryRowContext(
		ctx,
		"SELECT id, username, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(aliases,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(role,'user'), password_hash, totp_secret, COALESCE(theme_profile,''), COALESCE(theme_custom_css_path,''), COALESCE(theme_custom_css_inline,''), COALESCE(updated_at,'') FROM user WHERE username = ?",
		username,
	)
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

func GetUserByPrivateToken(ctx context.Context, db *sql.DB, token string) (domain.User, error) {
	row := db.QueryRowContext(
		ctx,
		"SELECT id, username, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(aliases,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(role,'user'), password_hash, totp_secret, COALESCE(theme_profile,''), COALESCE(theme_custom_css_path,''), COALESCE(theme_custom_css_inline,''), COALESCE(updated_at,'') FROM user WHERE private_token = ?",
		token,
	)
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

func GetOwnerUser(ctx context.Context, db *sql.DB) (domain.User, error) {
	row := db.QueryRowContext(
		ctx,
		"SELECT id, username, COALESCE(email,''), COALESCE(display_name,''), COALESCE(bio,''), COALESCE(organization,''), COALESCE(job_title,''), COALESCE(birthdate,''), COALESCE(languages,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(custom_fields,'{}'), COALESCE(visibility,''), COALESCE(private_token,''), COALESCE(links,'[]'), COALESCE(aliases,'[]'), COALESCE(social_profiles,'[]'), COALESCE(wallets,'{}'), COALESCE(public_keys,'{}'), COALESCE(location,''), COALESCE(website,''), COALESCE(pronouns,''), COALESCE(verified_domains,'[]'), COALESCE(atproto_handle,''), COALESCE(atproto_did,''), COALESCE(timezone,''), profile_picture_id, COALESCE(role,'user'), password_hash, totp_secret, COALESCE(theme_profile,''), COALESCE(theme_custom_css_path,''), COALESCE(theme_custom_css_inline,''), COALESCE(updated_at,'') FROM user WHERE role = 'owner' ORDER BY id LIMIT 1",
	)
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

func ListUsers(ctx context.Context, db *sql.DB) ([]domain.User, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, username, COALESCE(email,''), COALESCE(display_name,''), COALESCE(role,'user') FROM user ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func ListUsersPaged(ctx context.Context, db *sql.DB, query, sort, dir string, limit, offset int) ([]domain.User, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	sortKey := "id"
	if strings.ToLower(sort) == "username" {
		sortKey = "username"
	} else if strings.ToLower(sort) == "email" {
		sortKey = "email"
	} else if strings.ToLower(sort) == "role" {
		sortKey = "role"
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
		where = " WHERE username LIKE ? OR email LIKE ? OR display_name LIKE ?"
		pattern := "%" + strings.TrimSpace(query) + "%"
		args = append(args, pattern, pattern, pattern)
	}

	countQuery := "SELECT COUNT(*) FROM user" + where
	row := db.QueryRowContext(ctx, countQuery, args...)
	var total int
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}

	queryStr := fmt.Sprintf("SELECT id, username, COALESCE(email,''), COALESCE(display_name,''), COALESCE(role,'user') FROM user%s ORDER BY %s %s LIMIT ? OFFSET ?", where, sortKey, sortDir)
	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func UpdatePrivateToken(ctx context.Context, db *sql.DB, userID int, token string) error {
	_, err := db.ExecContext(ctx, "UPDATE user SET private_token = ?, updated_at = ? WHERE id = ?", token, time.Now().UTC().Format(time.RFC3339), userID)
	return err
}

func HasUser(ctx context.Context, db *sql.DB) (bool, error) {
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user")
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func CreateUser(ctx context.Context, db *sql.DB, username, email, role, passwordHash, totpSecret, themeProfile, privateToken string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := db.ExecContext(
		ctx,
		`INSERT INTO user (username, email, display_name, bio, organization, job_title, birthdate, languages, phone, address, custom_fields, visibility, private_token, links, aliases, social_profiles, wallets, public_keys, location, website, pronouns, verified_domains, atproto_handle, atproto_did, timezone, role, password_hash, totp_secret, theme_profile, theme_custom_css_path, theme_custom_css_inline, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		username, email, username, "", "", "", "", "", "", "", "{}", "{}", privateToken, "[]", "[]", "[]", "{}", "{}", "", "", "", "[]", "", "", "", role, passwordHash, totpSecret, themeProfile, "", "", now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteUser(ctx context.Context, db *sql.DB, userID int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM user WHERE id = ?", userID)
	return err
}

func ResetAllUserThemes(ctx context.Context, db *sql.DB, themeValue string) error {
	_, err := db.ExecContext(ctx, "UPDATE user SET theme_profile = ?, theme_custom_css_path = '', theme_custom_css_inline = ''", themeValue)
	return err
}

func UpdateUserTheme(ctx context.Context, db *sql.DB, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	_, err := db.ExecContext(ctx, "UPDATE user SET theme_profile = ?, theme_custom_css_path = ?, theme_custom_css_inline = ?, updated_at = ? WHERE id = ?", themeProfile, customCSSPath, customCSSInline, time.Now().UTC().Format(time.RFC3339), userID)
	return err
}
