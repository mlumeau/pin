package sqlitestore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pin/internal/platform/core"
)

// InitDB ensures the SQLite schema exists and applies lightweight migrations.
func InitDB(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS user (
            id INTEGER PRIMARY KEY,
            username TEXT UNIQUE NOT NULL,
            email TEXT,
            display_name TEXT,
            bio TEXT,
            organization TEXT,
            job_title TEXT,
            birthdate TEXT,
            languages TEXT,
            phone TEXT,
            address TEXT,
            custom_fields TEXT,
            visibility TEXT,
            private_token TEXT,
            links TEXT,
            aliases TEXT,
            social_profiles TEXT,
            wallets TEXT,
            public_keys TEXT,
            location TEXT,
            website TEXT,
            pronouns TEXT,
            verified_domains TEXT,
            atproto_handle TEXT,
            atproto_did TEXT,
            timezone TEXT,
            profile_picture_id INTEGER,
            role TEXT,
            password_hash TEXT NOT NULL,
            totp_secret TEXT NOT NULL,
            theme_profile TEXT,
            theme_custom_css_path TEXT,
            theme_custom_css_inline TEXT,
            updated_at TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY,
            value TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS invite (
            id INTEGER PRIMARY KEY,
            token TEXT UNIQUE NOT NULL,
            role TEXT NOT NULL,
            created_by INTEGER NOT NULL,
            created_at TEXT NOT NULL,
            used_at TEXT,
            used_by INTEGER
        )`,
		`CREATE TABLE IF NOT EXISTS user_identifier (
            identifier TEXT PRIMARY KEY,
            user_id INTEGER NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS passkey (
            id INTEGER PRIMARY KEY,
            user_id INTEGER NOT NULL,
            name TEXT NOT NULL,
            credential_id TEXT NOT NULL UNIQUE,
            credential_json TEXT NOT NULL,
            created_at TEXT NOT NULL,
            last_used_at TEXT
        )`,
		`CREATE INDEX IF NOT EXISTS idx_passkey_user ON passkey(user_id)`,
		`CREATE TABLE IF NOT EXISTS domain_verification (
            id INTEGER PRIMARY KEY,
            user_id INTEGER NOT NULL,
            domain TEXT NOT NULL,
            token TEXT NOT NULL,
            verified_at TEXT,
            created_at TEXT NOT NULL,
            UNIQUE(user_id, domain)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_domain_verification_user ON domain_verification(user_id)`,
		`CREATE TABLE IF NOT EXISTS audit_log (
            id INTEGER PRIMARY KEY,
            actor_id INTEGER,
            action TEXT NOT NULL,
            target TEXT,
            metadata TEXT,
            created_at TEXT NOT NULL
        )`,
		`CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at)`,
		`CREATE TABLE IF NOT EXISTS profile_picture (
            id INTEGER PRIMARY KEY,
            user_id INTEGER NOT NULL,
            filename TEXT NOT NULL,
            alt_text TEXT,
            created_at TEXT NOT NULL
        )`,
		`CREATE INDEX IF NOT EXISTS idx_profile_picture_user ON profile_picture(user_id)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	columns := map[string]string{
		"email":                   "TEXT",
		"display_name":            "TEXT",
		"bio":                     "TEXT",
		"organization":            "TEXT",
		"job_title":               "TEXT",
		"birthdate":               "TEXT",
		"languages":               "TEXT",
		"phone":                   "TEXT",
		"address":                 "TEXT",
		"custom_fields":           "TEXT",
		"visibility":              "TEXT",
		"private_token":           "TEXT",
		"links":                   "TEXT",
		"aliases":                 "TEXT",
		"social_profiles":         "TEXT",
		"wallets":                 "TEXT",
		"public_keys":             "TEXT",
		"location":                "TEXT",
		"website":                 "TEXT",
		"pronouns":                "TEXT",
		"verified_domains":        "TEXT",
		"atproto_handle":          "TEXT",
		"atproto_did":             "TEXT",
		"timezone":                "TEXT",
		"profile_picture_id":      "INTEGER",
		"role":                    "TEXT",
		"password_hash":           "TEXT",
		"totp_secret":             "TEXT",
		"theme_profile":           "TEXT",
		"theme_custom_css_path":   "TEXT",
		"theme_custom_css_inline": "TEXT",
		"updated_at":              "TEXT",
	}
	for col, typ := range columns {
		if err := ensureColumn(db, "user", col, typ); err != nil {
			return err
		}
	}

	if err := backfillUserDefaults(db); err != nil {
		return err
	}
	if err := backfillUserIdentifiers(db); err != nil {
		return err
	}
	return nil
}

func backfillUserDefaults(db *sql.DB) error {
	rows, err := db.Query("SELECT id, COALESCE(private_token,''), COALESCE(role,''), COALESCE(updated_at,'') FROM user")
	if err != nil {
		return err
	}
	defer rows.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for rows.Next() {
		var id int
		var privateToken, role, updatedAt string
		if err := rows.Scan(&id, &privateToken, &role, &updatedAt); err != nil {
			return err
		}
		if strings.TrimSpace(privateToken) == "" {
			if _, err := db.Exec("UPDATE user SET private_token = ? WHERE id = ?", core.RandomToken(32), id); err != nil {
				return err
			}
		}
		if strings.TrimSpace(role) == "" {
			if _, err := db.Exec("UPDATE user SET role = 'user' WHERE id = ?", id); err != nil {
				return err
			}
		}
		if strings.TrimSpace(updatedAt) == "" {
			if _, err := db.Exec("UPDATE user SET updated_at = ? WHERE id = ?", now, id); err != nil {
				return err
			}
		}
	}
	return nil
}

func backfillUserIdentifiers(db *sql.DB) error {
	rows, err := db.Query("SELECT id, username, COALESCE(aliases,'[]'), COALESCE(email,'') FROM user")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username, aliasesJSON, email string
		if err := rows.Scan(&id, &username, &aliasesJSON, &email); err != nil {
			return err
		}
		var aliases []string
		_ = json.Unmarshal([]byte(aliasesJSON), &aliases)
		idents := buildIdentifiers(username, aliases, email)
		for _, ident := range idents {
			if _, err := db.Exec("INSERT OR IGNORE INTO user_identifier (identifier, user_id) VALUES (?, ?)", strings.ToLower(strings.TrimSpace(ident)), id); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureColumn(db *sql.DB, table, column, columnType string) error {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if strings.EqualFold(name, column) {
			return nil
		}
	}
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, columnType))
	return err
}

func nullInt(value sql.NullInt64) interface{} {
	if value.Valid {
		return value.Int64
	}
	return nil
}
