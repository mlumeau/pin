package sqlitestore

import (
	"context"
	"database/sql"
	"strings"
)

// GetSettings fetches a set of keys from the settings table.
func GetSettings(ctx context.Context, db *sql.DB, keys ...string) (map[string]string, error) {
	results := map[string]string{}
	if len(keys) == 0 {
		return results, nil
	}
	placeholders := make([]string, 0, len(keys))
	args := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		placeholders = append(placeholders, "?")
		args = append(args, key)
	}
	query := "SELECT key, value FROM settings WHERE key IN (" + strings.Join(placeholders, ",") + ")"
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		results[key] = value
	}
	return results, nil
}

// GetSetting returns a single setting and whether it existed.
func GetSetting(ctx context.Context, db *sql.DB, key string) (string, bool, error) {
	values, err := GetSettings(ctx, db, key)
	if err != nil {
		return "", false, err
	}
	value, ok := values[key]
	return value, ok, nil
}

// SetSetting upserts a single setting.
func SetSetting(ctx context.Context, db *sql.DB, key, value string) error {
	_, err := db.ExecContext(ctx, "INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value)
	return err
}

// SetSettings upserts multiple settings in a transaction.
func SetSettings(ctx context.Context, db *sql.DB, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	for key, value := range values {
		if _, err := tx.ExecContext(ctx, "INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// DeleteSetting removes a key from the settings table.
func DeleteSetting(ctx context.Context, db *sql.DB, key string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM settings WHERE key = ?", key)
	return err
}
