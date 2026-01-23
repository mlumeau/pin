package identities

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for identities.
type Repository interface {
	GetIdentityByID(ctx context.Context, id int) (domain.Identity, error)
	GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error)
	GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error)
	GetIdentityByUserID(ctx context.Context, userID int) (domain.Identity, error)
	GetOwnerIdentity(ctx context.Context) (domain.Identity, error)
	ListIdentities(ctx context.Context) ([]domain.Identity, error)
	ListIdentitiesPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.Identity, int, error)
	CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error)
	UpdateIdentity(ctx context.Context, identity domain.Identity) error
	UpdatePrivateToken(ctx context.Context, identityID int, token string) error
	CheckHandleCollision(ctx context.Context, handle string, excludeID int) error
	DeleteIdentity(ctx context.Context, identityID int) error
}
