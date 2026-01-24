package sqlitestore

import (
	"context"
	"database/sql"

	"github.com/go-webauthn/webauthn/webauthn"
	"pin/internal/contracts"
	"pin/internal/domain"
)

type repos struct {
	db *sql.DB
}

// NewRepos wires sqlite-backed repositories for the app layer.
func NewRepos(db *sql.DB) contracts.Repos {
	r := repos{db: db}
	return contracts.Repos{
		Users:           r,
		Identities:      r,
		Invites:         r,
		Passkeys:        r,
		Audit:           r,
		Domains:         r,
		ProfilePictures: r,
		Settings:        r,
	}
}

// UsersStore
func (r repos) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	return GetUserByID(ctx, r.db, id)
}

// GetOwnerUser returns the owner user in the SQLite store.
func (r repos) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return GetOwnerUser(ctx, r.db)
}

// ListUsers returns the users list in the SQLite store.
func (r repos) ListUsers(ctx context.Context) ([]domain.User, error) {
	return ListUsers(ctx, r.db)
}

// ListUsersPaged returns a page of users paged using limit/offset in the SQLite store.
func (r repos) ListUsersPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.User, int, error) {
	return ListUsersPaged(ctx, r.db, query, sort, dir, limit, offset)
}

// HasUser reports whether user exists in the SQLite store.
func (r repos) HasUser(ctx context.Context) (bool, error) {
	return HasUser(ctx, r.db)
}

// CreateUser creates user using the supplied input in the SQLite store.
func (r repos) CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error) {
	return CreateUser(ctx, r.db, role, passwordHash, totpSecret, themeProfile)
}

// DeleteUser deletes user in the SQLite store.
func (r repos) DeleteUser(ctx context.Context, userID int) error {
	return DeleteUser(ctx, r.db, userID)
}

// UpdateUser updates user using the supplied data in the SQLite store.
func (r repos) UpdateUser(ctx context.Context, u domain.User) error {
	return UpdateUser(ctx, r.db, u)
}

// ResetAllUserThemes resets all user themes to its default state.
func (r repos) ResetAllUserThemes(ctx context.Context, themeValue string) error {
	return ResetAllUserThemes(ctx, r.db, themeValue)
}

// UpdateUserTheme updates user theme using the supplied data in the SQLite store.
func (r repos) UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	return UpdateUserTheme(ctx, r.db, userID, themeProfile, customCSSPath, customCSSInline)
}

// IdentitiesStore
func (r repos) GetIdentityByID(ctx context.Context, id int) (domain.Identity, error) {
	return GetIdentityByID(ctx, r.db, id)
}

// GetIdentityByHandle returns identity by handle.
func (r repos) GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error) {
	return GetIdentityByHandle(ctx, r.db, handle)
}

// GetIdentityByPrivateToken returns identity by private token.
func (r repos) GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error) {
	return GetIdentityByPrivateToken(ctx, r.db, token)
}

// GetIdentityByUserID returns identity by user ID.
func (r repos) GetIdentityByUserID(ctx context.Context, userID int) (domain.Identity, error) {
	return GetIdentityByUserID(ctx, r.db, userID)
}

// GetOwnerIdentity returns the owner identity in the SQLite store.
func (r repos) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return GetOwnerIdentity(ctx, r.db)
}

// ListIdentities returns the identities list in the SQLite store.
func (r repos) ListIdentities(ctx context.Context) ([]domain.Identity, error) {
	return ListIdentities(ctx, r.db)
}

// ListIdentitiesPaged returns a page of identities paged using limit/offset in the SQLite store.
func (r repos) ListIdentitiesPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.Identity, int, error) {
	return ListIdentitiesPaged(ctx, r.db, query, sort, dir, limit, offset)
}

// CreateIdentity creates identity using the supplied input in the SQLite store.
func (r repos) CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error) {
	return CreateIdentity(ctx, r.db, identity)
}

// UpdateIdentity updates identity using the supplied data in the SQLite store.
func (r repos) UpdateIdentity(ctx context.Context, identity domain.Identity) error {
	return UpdateIdentity(ctx, r.db, identity)
}

// UpdatePrivateToken updates private token using the supplied data in the SQLite store.
func (r repos) UpdatePrivateToken(ctx context.Context, identityID int, token string) error {
	return UpdatePrivateToken(ctx, r.db, identityID, token)
}

// CheckHandleCollision checks handle collision and reports whether it matches.
func (r repos) CheckHandleCollision(ctx context.Context, handle string, excludeID int) error {
	return CheckHandleCollision(ctx, r.db, handle, excludeID)
}

// DeleteIdentity deletes identity in the SQLite store.
func (r repos) DeleteIdentity(ctx context.Context, identityID int) error {
	return DeleteIdentity(ctx, r.db, identityID)
}

// InvitesStore
func (r repos) CreateInvite(ctx context.Context, token, role string, createdBy int) error {
	return CreateInvite(ctx, r.db, token, role, createdBy)
}

// ListInvites returns the invites list in the SQLite store.
func (r repos) ListInvites(ctx context.Context) ([]domain.Invite, error) {
	return ListInvites(ctx, r.db)
}

// GetInviteByToken returns invite by token.
func (r repos) GetInviteByToken(ctx context.Context, token string) (domain.Invite, error) {
	return GetInviteByToken(ctx, r.db, token)
}

// MarkInviteUsed returns invite used.
func (r repos) MarkInviteUsed(ctx context.Context, id int, usedBy int) error {
	return MarkInviteUsed(ctx, r.db, id, usedBy)
}

// DeleteInvite deletes invite in the SQLite store.
func (r repos) DeleteInvite(ctx context.Context, id int) error {
	return DeleteInvite(ctx, r.db, id)
}

// PasskeysStore
func (r repos) ListPasskeys(ctx context.Context, userID int) ([]domain.Passkey, error) {
	return ListPasskeys(ctx, r.db, userID)
}

// LoadPasskeyCredentials loads passkey credentials from storage.
func (r repos) LoadPasskeyCredentials(ctx context.Context, userID int) ([]webauthn.Credential, error) {
	return LoadPasskeyCredentials(ctx, r.db, userID)
}

// InsertPasskey returns passkey.
func (r repos) InsertPasskey(ctx context.Context, userID int, name string, credential webauthn.Credential) error {
	return InsertPasskey(ctx, r.db, userID, name, credential)
}

// UpdatePasskeyCredential updates passkey credential using the supplied data in the SQLite store.
func (r repos) UpdatePasskeyCredential(ctx context.Context, userID int, credentialID string, credential webauthn.Credential) error {
	return UpdatePasskeyCredential(ctx, r.db, userID, credentialID, credential)
}

// DeletePasskey deletes passkey in the SQLite store.
func (r repos) DeletePasskey(ctx context.Context, userID, id int) error {
	return DeletePasskey(ctx, r.db, userID, id)
}

// AuditStore
func (r repos) WriteAuditLog(ctx context.Context, actorID int, action, target string, metadata map[string]string) error {
	return WriteAuditLog(ctx, r.db, actorID, action, target, metadata)
}

// ListAuditLogs returns a page of audit logs using limit/offset in the SQLite store.
func (r repos) ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	return ListAuditLogs(ctx, r.db, limit, offset)
}

// CountAuditLogs returns audit logs.
func (r repos) CountAuditLogs(ctx context.Context) (int, error) {
	return CountAuditLogs(ctx, r.db)
}

// ListAllAuditLogs returns the all audit logs list in the SQLite store.
func (r repos) ListAllAuditLogs(ctx context.Context) ([]domain.AuditLog, error) {
	return ListAllAuditLogs(ctx, r.db)
}

// DomainsStore
func (r repos) ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error) {
	return ListDomainVerifications(ctx, r.db, identityID)
}

// ListVerifiedDomains returns the verified domains list in the SQLite store.
func (r repos) ListVerifiedDomains(ctx context.Context, identityID int) ([]string, error) {
	return ListVerifiedDomains(ctx, r.db, identityID)
}

// UpsertDomainVerification returns domain verification.
func (r repos) UpsertDomainVerification(ctx context.Context, identityID int, domain, token string) error {
	return UpsertDomainVerification(ctx, r.db, identityID, domain, token)
}

// DeleteDomainVerification deletes domain verification in the SQLite store.
func (r repos) DeleteDomainVerification(ctx context.Context, identityID int, domain string) error {
	return DeleteDomainVerification(ctx, r.db, identityID, domain)
}

// MarkDomainVerified returns domain verified.
func (r repos) MarkDomainVerified(ctx context.Context, identityID int, domain string) error {
	return MarkDomainVerified(ctx, r.db, identityID, domain)
}

// UpdateIdentityVerifiedDomains updates identity verified domains using the supplied data in the SQLite store.
func (r repos) UpdateIdentityVerifiedDomains(ctx context.Context, identityID int, domains []string) error {
	return UpdateIdentityVerifiedDomains(ctx, r.db, identityID, domains)
}

// HasDomainVerification reports whether domain verification exists in the SQLite store.
func (r repos) HasDomainVerification(ctx context.Context, identityID int, domain string) (bool, error) {
	return HasDomainVerification(ctx, r.db, identityID, domain)
}

// ProtectedDomain returns domain.
func (r repos) ProtectedDomain(ctx context.Context) string {
	if value, ok, err := r.GetSetting(ctx, "protected_domain"); err == nil && ok {
		return value
	}
	return ""
}

// SetProtectedDomain sets protected domain to the provided value in the SQLite store.
func (r repos) SetProtectedDomain(ctx context.Context, domain string) error {
	return r.SetSetting(ctx, "protected_domain", domain)
}

// ProfilePicturesStore
func (r repos) ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error) {
	return ListProfilePictures(ctx, r.db, identityID)
}

// CreateProfilePicture creates profile picture using the supplied input in the SQLite store.
func (r repos) CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error) {
	return CreateProfilePicture(ctx, r.db, identityID, filename, alt)
}

// SetActiveProfilePicture sets active profile picture to the provided value in the SQLite store.
func (r repos) SetActiveProfilePicture(ctx context.Context, identityID int, pictureID int64) error {
	return SetActiveProfilePicture(ctx, r.db, identityID, pictureID)
}

// DeleteProfilePicture deletes profile picture in the SQLite store.
func (r repos) DeleteProfilePicture(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return DeleteProfilePicture(ctx, r.db, identityID, pictureID)
}

// UpdateProfilePictureAlt updates profile picture alt using the supplied data in the SQLite store.
func (r repos) UpdateProfilePictureAlt(ctx context.Context, identityID int, pictureID int64, alt string) error {
	return UpdateProfilePictureAlt(ctx, r.db, identityID, pictureID, alt)
}

// ClearProfilePictureSelection returns profile picture selection.
func (r repos) ClearProfilePictureSelection(ctx context.Context, identityID int) error {
	return ClearProfilePictureSelection(ctx, r.db, identityID)
}

// GetProfilePictureFilename returns the profile picture filename in the SQLite store.
func (r repos) GetProfilePictureFilename(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return GetProfilePictureFilename(ctx, r.db, identityID, pictureID)
}

// GetProfilePictureAlt returns the profile picture alt in the SQLite store.
func (r repos) GetProfilePictureAlt(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return GetProfilePictureAlt(ctx, r.db, identityID, pictureID)
}

// SettingsStore
func (r repos) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	return GetSettings(ctx, r.db, keys...)
}

// GetSetting returns the setting in the SQLite store.
func (r repos) GetSetting(ctx context.Context, key string) (string, bool, error) {
	return GetSetting(ctx, r.db, key)
}

// SetSetting sets setting to the provided value in the SQLite store.
func (r repos) SetSetting(ctx context.Context, key, value string) error {
	return SetSetting(ctx, r.db, key, value)
}

// SetSettings sets settings to the provided value in the SQLite store.
func (r repos) SetSettings(ctx context.Context, values map[string]string) error {
	return SetSettings(ctx, r.db, values)
}

// DeleteSetting deletes setting in the SQLite store.
func (r repos) DeleteSetting(ctx context.Context, key string) error {
	return DeleteSetting(ctx, r.db, key)
}
