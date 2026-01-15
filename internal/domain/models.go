package domain

import (
	"database/sql"
	"time"
)

// User represents the single profile stored in the database.
type User struct {
	ID                   int
	Username             string
	Email                string
	DisplayName          string
	Bio                  string
	Organization         string
	JobTitle             string
	Birthdate            string
	Languages            string
	Phone                string
	Address              string
	CustomFieldsJSON     string
	VisibilityJSON       string
	PrivateToken         string
	LinksJSON            string
	AliasesJSON          string
	SocialProfilesJSON   string
	WalletsJSON          string
	PublicKeysJSON       string
	Location             string
	Website              string
	Pronouns             string
	VerifiedDomainsJSON  string
	ATProtoHandle        string
	ATProtoDID           string
	Timezone             string
	ProfilePictureID     sql.NullInt64
	Role                 string
	PasswordHash         string
	TOTPSecret           string
	UpdatedAt            time.Time
	ThemeProfile         string
	ThemeCustomCSSPath   string
	ThemeCustomCSSInline string
}

type Invite struct {
	ID         int
	Token      string
	Role       string
	CreatedBy  int
	CreatedAt  time.Time
	UsedAt     sql.NullTime
	UsedBy     sql.NullInt64
	UsedByName string
}

type Passkey struct {
	ID             int
	UserID         int
	Name           string
	CredentialID   string
	CredentialJSON string
	CreatedAt      time.Time
	LastUsedAt     sql.NullTime
}

type AuditLog struct {
	ID        int
	ActorID   sql.NullInt64
	ActorName string
	Action    string
	Target    string
	Metadata  string
	CreatedAt time.Time
}

type ProfilePicture struct {
	ID        int64     `json:"id"`
	UserID    int       `json:"user_id"`
	Filename  string    `json:"filename"`
	AltText   string    `json:"alt_text"`
	CreatedAt time.Time `json:"created_at"`
}

type DomainVerification struct {
	ID         int
	UserID     int
	Domain     string
	Token      string
	VerifiedAt sql.NullTime
	CreatedAt  time.Time
}

// Link is a label + URL pair serialized into JSON for storage.
type Link struct {
	Label      string `json:"label"`
	URL        string `json:"url"`
	Visibility string `json:"visibility,omitempty"`
}

// SocialProfile represents a social link with optional verification metadata.
type SocialProfile struct {
	Label      string `json:"label"`
	URL        string `json:"url"`
	Provider   string `json:"provider,omitempty"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility,omitempty"`
}
