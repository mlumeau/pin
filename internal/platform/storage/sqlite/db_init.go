package sqlitestore

import (
	"database/sql"
)

// InitDB returns db.
func InitDB(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS user (
            id INTEGER PRIMARY KEY,
            role TEXT,
            password_hash TEXT NOT NULL,
            totp_secret TEXT NOT NULL,
            theme_profile TEXT,
            theme_custom_css_path TEXT,
            theme_custom_css_inline TEXT,
            updated_at TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS identity (
            id INTEGER PRIMARY KEY,
            user_id INTEGER NOT NULL,
            handle TEXT UNIQUE NOT NULL,
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
            updated_at TEXT,
            UNIQUE(user_id)
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
            identity_id INTEGER NOT NULL,
            domain TEXT NOT NULL,
            token TEXT NOT NULL,
            verified_at TEXT,
            created_at TEXT NOT NULL,
            UNIQUE(identity_id, domain)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_domain_verification_identity ON domain_verification(identity_id)`,
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
            identity_id INTEGER NOT NULL,
            filename TEXT NOT NULL,
            alt_text TEXT,
            created_at TEXT NOT NULL
        )`,
		`CREATE INDEX IF NOT EXISTS idx_profile_picture_identity ON profile_picture(identity_id)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

// nullInt returns int.
func nullInt(value sql.NullInt64) interface{} {
	if value.Valid {
		return value.Int64
	}
	return nil
}
