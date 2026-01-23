package domain

import (
	"database/sql"
	"time"
)

// User represents an authenticated account.
type User struct {
	ID                   int
	Role                 string
	PasswordHash         string
	TOTPSecret           string
	ThemeProfile         string
	ThemeCustomCSSPath   string
	ThemeCustomCSSInline string
	UpdatedAt            time.Time
}

// Identity represents a public/private profile owned by a user.
type Identity struct {
	ID                  int
	UserID              int
	Handle              string
	Email               string
	DisplayName         string
	Bio                 string
	Organization        string
	JobTitle            string
	Birthdate           string
	Languages           string
	Phone               string
	Address             string
	CustomFieldsJSON    string
	VisibilityJSON      string
	PrivateToken        string
	LinksJSON           string
	SocialProfilesJSON  string
	WalletsJSON         string
	PublicKeysJSON      string
	Location            string
	Website             string
	Pronouns            string
	VerifiedDomainsJSON string
	ATProtoHandle       string
	ATProtoDID          string
	Timezone            string
	ProfilePictureID    sql.NullInt64
	UpdatedAt           time.Time
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
	IdentityID int      `json:"identity_id"`
	Filename  string    `json:"filename"`
	AltText   string    `json:"alt_text"`
	CreatedAt time.Time `json:"created_at"`
}

type DomainVerification struct {
	ID         int
	IdentityID int
	Domain     string
	Token      string
	VerifiedAt sql.NullTime
	CreatedAt  time.Time
}

// Visibility represents a keyed visibility entry stored on the user.
type Visibility struct {
	Key        string `json:"key"`
	Visibility string `json:"visibility"`
}

// Link is a label + URL pair serialized into JSON for storage.
type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// SocialProfile represents a social link with optional verification metadata.
type SocialProfile struct {
	Label    string `json:"label"`
	URL      string `json:"url"`
	Provider string `json:"provider,omitempty"`
	Verified bool   `json:"verified"`
}

// CustomField represents a user-defined custom field in key-value format.
type CustomField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Wallet represents a cryptocurrency or payment address with visibility settings.
type Wallet struct {
	Label   string `json:"label"`
	Address string `json:"address"`
}

// PublicKey represents a cryptographic public key for a specific algorithm.
type PublicKey struct {
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
}

// VerifiedDomain represents a verified domain associated with a user.
type VerifiedDomain struct {
	Domain string `json:"domain"`
}
