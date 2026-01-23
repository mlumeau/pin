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

func (r repos) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return GetOwnerUser(ctx, r.db)
}

func (r repos) ListUsers(ctx context.Context) ([]domain.User, error) {
	return ListUsers(ctx, r.db)
}

func (r repos) ListUsersPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.User, int, error) {
	return ListUsersPaged(ctx, r.db, query, sort, dir, limit, offset)
}

func (r repos) HasUser(ctx context.Context) (bool, error) {
	return HasUser(ctx, r.db)
}

func (r repos) CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error) {
	return CreateUser(ctx, r.db, role, passwordHash, totpSecret, themeProfile)
}

func (r repos) DeleteUser(ctx context.Context, userID int) error {
	return DeleteUser(ctx, r.db, userID)
}

func (r repos) UpdateUser(ctx context.Context, u domain.User) error {
	return UpdateUser(ctx, r.db, u)
}

func (r repos) ResetAllUserThemes(ctx context.Context, themeValue string) error {
	return ResetAllUserThemes(ctx, r.db, themeValue)
}

func (r repos) UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	return UpdateUserTheme(ctx, r.db, userID, themeProfile, customCSSPath, customCSSInline)
}

// IdentitiesStore
func (r repos) GetIdentityByID(ctx context.Context, id int) (domain.Identity, error) {
	return GetIdentityByID(ctx, r.db, id)
}

func (r repos) GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error) {
	return GetIdentityByHandle(ctx, r.db, handle)
}

func (r repos) GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error) {
	return GetIdentityByPrivateToken(ctx, r.db, token)
}

func (r repos) GetIdentityByUserID(ctx context.Context, userID int) (domain.Identity, error) {
	return GetIdentityByUserID(ctx, r.db, userID)
}

func (r repos) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return GetOwnerIdentity(ctx, r.db)
}

func (r repos) ListIdentities(ctx context.Context) ([]domain.Identity, error) {
	return ListIdentities(ctx, r.db)
}

func (r repos) ListIdentitiesPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.Identity, int, error) {
	return ListIdentitiesPaged(ctx, r.db, query, sort, dir, limit, offset)
}

func (r repos) CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error) {
	return CreateIdentity(ctx, r.db, identity)
}

func (r repos) UpdateIdentity(ctx context.Context, identity domain.Identity) error {
	return UpdateIdentity(ctx, r.db, identity)
}

func (r repos) UpdatePrivateToken(ctx context.Context, identityID int, token string) error {
	return UpdatePrivateToken(ctx, r.db, identityID, token)
}

func (r repos) CheckHandleCollision(ctx context.Context, handle string, excludeID int) error {
	return CheckHandleCollision(ctx, r.db, handle, excludeID)
}

func (r repos) DeleteIdentity(ctx context.Context, identityID int) error {
	return DeleteIdentity(ctx, r.db, identityID)
}

// InvitesStore
func (r repos) CreateInvite(ctx context.Context, token, role string, createdBy int) error {
	return CreateInvite(ctx, r.db, token, role, createdBy)
}

func (r repos) ListInvites(ctx context.Context) ([]domain.Invite, error) {
	return ListInvites(ctx, r.db)
}

func (r repos) GetInviteByToken(ctx context.Context, token string) (domain.Invite, error) {
	return GetInviteByToken(ctx, r.db, token)
}

func (r repos) MarkInviteUsed(ctx context.Context, id int, usedBy int) error {
	return MarkInviteUsed(ctx, r.db, id, usedBy)
}

func (r repos) DeleteInvite(ctx context.Context, id int) error {
	return DeleteInvite(ctx, r.db, id)
}

// PasskeysStore
func (r repos) ListPasskeys(ctx context.Context, userID int) ([]domain.Passkey, error) {
	return ListPasskeys(ctx, r.db, userID)
}

func (r repos) LoadPasskeyCredentials(ctx context.Context, userID int) ([]webauthn.Credential, error) {
	return LoadPasskeyCredentials(ctx, r.db, userID)
}

func (r repos) InsertPasskey(ctx context.Context, userID int, name string, credential webauthn.Credential) error {
	return InsertPasskey(ctx, r.db, userID, name, credential)
}

func (r repos) UpdatePasskeyCredential(ctx context.Context, userID int, credentialID string, credential webauthn.Credential) error {
	return UpdatePasskeyCredential(ctx, r.db, userID, credentialID, credential)
}

func (r repos) DeletePasskey(ctx context.Context, userID, id int) error {
	return DeletePasskey(ctx, r.db, userID, id)
}

// AuditStore
func (r repos) WriteAuditLog(ctx context.Context, actorID int, action, target string, metadata map[string]string) error {
	return WriteAuditLog(ctx, r.db, actorID, action, target, metadata)
}

func (r repos) ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	return ListAuditLogs(ctx, r.db, limit, offset)
}

func (r repos) CountAuditLogs(ctx context.Context) (int, error) {
	return CountAuditLogs(ctx, r.db)
}

func (r repos) ListAllAuditLogs(ctx context.Context) ([]domain.AuditLog, error) {
	return ListAllAuditLogs(ctx, r.db)
}

// DomainsStore
func (r repos) ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error) {
	return ListDomainVerifications(ctx, r.db, identityID)
}

func (r repos) ListVerifiedDomains(ctx context.Context, identityID int) ([]string, error) {
	return ListVerifiedDomains(ctx, r.db, identityID)
}

func (r repos) UpsertDomainVerification(ctx context.Context, identityID int, domain, token string) error {
	return UpsertDomainVerification(ctx, r.db, identityID, domain, token)
}

func (r repos) DeleteDomainVerification(ctx context.Context, identityID int, domain string) error {
	return DeleteDomainVerification(ctx, r.db, identityID, domain)
}

func (r repos) MarkDomainVerified(ctx context.Context, identityID int, domain string) error {
	return MarkDomainVerified(ctx, r.db, identityID, domain)
}

func (r repos) UpdateIdentityVerifiedDomains(ctx context.Context, identityID int, domains []string) error {
	return UpdateIdentityVerifiedDomains(ctx, r.db, identityID, domains)
}

func (r repos) HasDomainVerification(ctx context.Context, identityID int, domain string) (bool, error) {
	return HasDomainVerification(ctx, r.db, identityID, domain)
}

func (r repos) ProtectedDomain(ctx context.Context) string {
	if value, ok, err := r.GetSetting(ctx, "protected_domain"); err == nil && ok {
		return value
	}
	return ""
}

func (r repos) SetProtectedDomain(ctx context.Context, domain string) error {
	return r.SetSetting(ctx, "protected_domain", domain)
}

// ProfilePicturesStore
func (r repos) ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error) {
	return ListProfilePictures(ctx, r.db, identityID)
}

func (r repos) CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error) {
	return CreateProfilePicture(ctx, r.db, identityID, filename, alt)
}

func (r repos) SetActiveProfilePicture(ctx context.Context, identityID int, pictureID int64) error {
	return SetActiveProfilePicture(ctx, r.db, identityID, pictureID)
}

func (r repos) DeleteProfilePicture(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return DeleteProfilePicture(ctx, r.db, identityID, pictureID)
}

func (r repos) UpdateProfilePictureAlt(ctx context.Context, identityID int, pictureID int64, alt string) error {
	return UpdateProfilePictureAlt(ctx, r.db, identityID, pictureID, alt)
}

func (r repos) ClearProfilePictureSelection(ctx context.Context, identityID int) error {
	return ClearProfilePictureSelection(ctx, r.db, identityID)
}

func (r repos) GetProfilePictureFilename(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return GetProfilePictureFilename(ctx, r.db, identityID, pictureID)
}

func (r repos) GetProfilePictureAlt(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return GetProfilePictureAlt(ctx, r.db, identityID, pictureID)
}

// SettingsStore
func (r repos) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	return GetSettings(ctx, r.db, keys...)
}

func (r repos) GetSetting(ctx context.Context, key string) (string, bool, error) {
	return GetSetting(ctx, r.db, key)
}

func (r repos) SetSetting(ctx context.Context, key, value string) error {
	return SetSetting(ctx, r.db, key, value)
}

func (r repos) SetSettings(ctx context.Context, values map[string]string) error {
	return SetSettings(ctx, r.db, values)
}

func (r repos) DeleteSetting(ctx context.Context, key string) error {
	return DeleteSetting(ctx, r.db, key)
}
