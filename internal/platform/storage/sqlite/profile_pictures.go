package sqlitestore

import (
	"context"
	"database/sql"
	"time"

	"pin/internal/domain"
)

// ListProfilePictures returns the profile pictures list in the SQLite store.
func ListProfilePictures(ctx context.Context, db *sql.DB, identityID int) ([]domain.ProfilePicture, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, identity_id, filename, COALESCE(alt_text,''), created_at FROM profile_picture WHERE identity_id = ? ORDER BY created_at DESC", identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pics []domain.ProfilePicture
	for rows.Next() {
		var pic domain.ProfilePicture
		var created string
		if err := rows.Scan(&pic.ID, &pic.IdentityID, &pic.Filename, &pic.AltText, &created); err != nil {
			return nil, err
		}
		pic.CreatedAt, _ = time.Parse(time.RFC3339, created)
		pics = append(pics, pic)
	}
	return pics, nil
}

// CreateProfilePicture creates profile picture using the supplied input in the SQLite store.
func CreateProfilePicture(ctx context.Context, db *sql.DB, identityID int, filename, alt string) (int64, error) {
	res, err := db.ExecContext(ctx, "INSERT INTO profile_picture (identity_id, filename, alt_text, created_at) VALUES (?, ?, ?, ?)", identityID, filename, alt, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// SetActiveProfilePicture sets active profile picture to the provided value in the SQLite store.
func SetActiveProfilePicture(ctx context.Context, db *sql.DB, identityID int, pictureID int64) error {
	_, err := db.ExecContext(ctx, "UPDATE identity SET profile_picture_id = ?, updated_at = ? WHERE id = ?", pictureID, time.Now().UTC().Format(time.RFC3339), identityID)
	return err
}

// DeleteProfilePicture deletes profile picture in the SQLite store.
func DeleteProfilePicture(ctx context.Context, db *sql.DB, identityID int, pictureID int64) (string, error) {
	row := db.QueryRowContext(ctx, "SELECT filename FROM profile_picture WHERE id = ? AND identity_id = ?", pictureID, identityID)
	var filename string
	if err := row.Scan(&filename); err != nil {
		return "", err
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM profile_picture WHERE id = ? AND identity_id = ?", pictureID, identityID); err != nil {
		return "", err
	}
	return filename, nil
}

// UpdateProfilePictureAlt updates profile picture alt using the supplied data in the SQLite store.
func UpdateProfilePictureAlt(ctx context.Context, db *sql.DB, identityID int, pictureID int64, alt string) error {
	_, err := db.ExecContext(ctx, "UPDATE profile_picture SET alt_text = ? WHERE id = ? AND identity_id = ?", alt, pictureID, identityID)
	return err
}

// ClearProfilePictureSelection returns profile picture selection.
func ClearProfilePictureSelection(ctx context.Context, db *sql.DB, identityID int) error {
	_, err := db.ExecContext(ctx, "UPDATE identity SET profile_picture_id = NULL, updated_at = ? WHERE id = ?", time.Now().UTC().Format(time.RFC3339), identityID)
	return err
}

// GetProfilePictureFilename returns the profile picture filename in the SQLite store.
func GetProfilePictureFilename(ctx context.Context, db *sql.DB, identityID int, pictureID int64) (string, error) {
	row := db.QueryRowContext(ctx, "SELECT filename FROM profile_picture WHERE id = ? AND identity_id = ?", pictureID, identityID)
	var filename string
	if err := row.Scan(&filename); err != nil {
		return "", err
	}
	return filename, nil
}

// GetProfilePictureAlt returns the profile picture alt in the SQLite store.
func GetProfilePictureAlt(ctx context.Context, db *sql.DB, identityID int, pictureID int64) (string, error) {
	row := db.QueryRowContext(ctx, "SELECT COALESCE(alt_text,'') FROM profile_picture WHERE id = ? AND identity_id = ?", pictureID, identityID)
	var alt string
	if err := row.Scan(&alt); err != nil {
		return "", err
	}
	return alt, nil
}
