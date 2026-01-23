package wiring

import (
	"context"

	"pin/internal/domain"
)

// Identities.
func (d Deps) GetIdentityByID(ctx context.Context, id int) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByID(ctx, id)
}

func (d Deps) GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByHandle(ctx, handle)
}

func (d Deps) GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByPrivateToken(ctx, token)
}

func (d Deps) GetIdentityByUserID(ctx context.Context, userID int) (domain.Identity, error) {
	return d.repos.Identities.GetIdentityByUserID(ctx, userID)
}

func (d Deps) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return d.repos.Identities.GetOwnerIdentity(ctx)
}

func (d Deps) ListIdentities(ctx context.Context) ([]domain.Identity, error) {
	return d.repos.Identities.ListIdentities(ctx)
}

func (d Deps) ListIdentitiesPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.Identity, int, error) {
	return d.repos.Identities.ListIdentitiesPaged(ctx, query, sort, dir, limit, offset)
}

func (d Deps) CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error) {
	return d.repos.Identities.CreateIdentity(ctx, identity)
}

func (d Deps) UpdateIdentity(ctx context.Context, identity domain.Identity) error {
	return d.repos.Identities.UpdateIdentity(ctx, identity)
}

func (d Deps) UpdateIdentityPrivateToken(ctx context.Context, identityID int, token string) error {
	return d.repos.Identities.UpdatePrivateToken(ctx, identityID, token)
}

func (d Deps) CheckHandleCollision(ctx context.Context, handle string, excludeID int) error {
	return d.repos.Identities.CheckHandleCollision(ctx, handle, excludeID)
}

func (d Deps) DeleteIdentity(ctx context.Context, identityID int) error {
	return d.repos.Identities.DeleteIdentity(ctx, identityID)
}
