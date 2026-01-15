package sqlitestore

import (
	"context"
	"database/sql"
	"time"

	"pin/internal/domain"
)

func ListProfilePictures(ctx context.Context, db *sql.DB, userID int) ([]domain.ProfilePicture, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, filename, COALESCE(alt_text,''), created_at FROM profile_picture WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pics []domain.ProfilePicture
	for rows.Next() {
		var pic domain.ProfilePicture
		var created string
		if err := rows.Scan(&pic.ID, &pic.UserID, &pic.Filename, &pic.AltText, &created); err != nil {
			return nil, err
		}
		pic.CreatedAt, _ = time.Parse(time.RFC3339, created)
		pics = append(pics, pic)
	}
	return pics, nil
}

func CreateProfilePicture(ctx context.Context, db *sql.DB, userID int, filename, alt string) (int64, error) {
	res, err := db.ExecContext(ctx, "INSERT INTO profile_picture (user_id, filename, alt_text, created_at) VALUES (?, ?, ?, ?)", userID, filename, alt, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func SetActiveProfilePicture(ctx context.Context, db *sql.DB, userID int, pictureID int64) error {
	_, err := db.ExecContext(ctx, "UPDATE user SET profile_picture_id = ?, updated_at = ? WHERE id = ?", pictureID, time.Now().UTC().Format(time.RFC3339), userID)
	return err
}

func DeleteProfilePicture(ctx context.Context, db *sql.DB, userID int, pictureID int64) (string, error) {
	row := db.QueryRowContext(ctx, "SELECT filename FROM profile_picture WHERE id = ? AND user_id = ?", pictureID, userID)
	var filename string
	if err := row.Scan(&filename); err != nil {
		return "", err
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM profile_picture WHERE id = ? AND user_id = ?", pictureID, userID); err != nil {
		return "", err
	}
	return filename, nil
}

func UpdateProfilePictureAlt(ctx context.Context, db *sql.DB, userID int, pictureID int64, alt string) error {
	_, err := db.ExecContext(ctx, "UPDATE profile_picture SET alt_text = ? WHERE id = ? AND user_id = ?", alt, pictureID, userID)
	return err
}

func ClearProfilePictureSelection(ctx context.Context, db *sql.DB, userID int) error {
	_, err := db.ExecContext(ctx, "UPDATE user SET profile_picture_id = NULL, updated_at = ? WHERE id = ?", time.Now().UTC().Format(time.RFC3339), userID)
	return err
}

func GetProfilePictureFilename(ctx context.Context, db *sql.DB, userID int, pictureID int64) (string, error) {
	row := db.QueryRowContext(ctx, "SELECT filename FROM profile_picture WHERE id = ? AND user_id = ?", pictureID, userID)
	var filename string
	if err := row.Scan(&filename); err != nil {
		return "", err
	}
	return filename, nil
}

func GetProfilePictureAlt(ctx context.Context, db *sql.DB, userID int, pictureID int64) (string, error) {
	row := db.QueryRowContext(ctx, "SELECT COALESCE(alt_text,'') FROM profile_picture WHERE id = ? AND user_id = ?", pictureID, userID)
	var alt string
	if err := row.Scan(&alt); err != nil {
		return "", err
	}
	return alt, nil
}
