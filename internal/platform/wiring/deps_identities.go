package wiring

import (
	"context"

	"pin/internal/domain"
)

// Identities.
func (d Deps) GetIdentityByID(ctx context.Context, id int) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByID(ctx, id)
}

// GetIdentityByHandle returns identity by handle.
func (d Deps) GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByHandle(ctx, handle)
}

// GetIdentityByPrivateToken returns identity by private token.
func (d Deps) GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByPrivateToken(ctx, token)
}

// GetIdentityByUserID returns identity by user ID.
func (d Deps) GetIdentityByUserID(ctx context.Context, userID int) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByUserID(ctx, userID)
}

// GetOwnerIdentity returns the owner identity by delegating to configured services.
func (d Deps) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return d.repos.Identities.GetOwnerIdentity(ctx)
}

// ListIdentities returns the identities list by delegating to configured services.
func (d Deps) ListIdentities(ctx context.Context) ([]domain.Identity, error) {
	return d.repos.Identities.ListIdentities(ctx)
}

// ListIdentitiesPaged returns a page of identities paged using limit/offset by delegating to configured services.
func (d Deps) ListIdentitiesPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.Identity, int, error) {
	return d.repos.Identities.ListIdentitiesPaged(ctx, query, sort, dir, limit, offset)
}

// CreateIdentity creates identity using the supplied input by delegating to configured services.
func (d Deps) CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error) {
	return d.repos.Identities.CreateIdentity(ctx, identity)
}

// UpdateIdentity updates identity using the supplied data by delegating to configured services.
func (d Deps) UpdateIdentity(ctx context.Context, identity domain.Identity) error {
	return d.repos.Identities.UpdateIdentity(ctx, identity)
}

// UpdateIdentityPrivateToken updates identity private token using the supplied data by delegating to configured services.
func (d Deps) UpdateIdentityPrivateToken(ctx context.Context, identityID int, token string) error {
	return d.repos.Identities.UpdatePrivateToken(ctx, identityID, token)
}

// CheckHandleCollision checks handle collision and reports whether it matches.
func (d Deps) CheckHandleCollision(ctx context.Context, handle string, excludeID int) error {
	return d.repos.Identities.CheckHandleCollision(ctx, handle, excludeID)
}

// DeleteIdentity deletes identity by delegating to configured services.
func (d Deps) DeleteIdentity(ctx context.Context, identityID int) error {
	return d.repos.Identities.DeleteIdentity(ctx, identityID)
}
