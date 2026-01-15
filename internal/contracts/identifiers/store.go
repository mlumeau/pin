package identifiers

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for user identifiers.
type Repository interface {
	CheckIdentifierCollisions(ctx context.Context, identifiers []string, excludeID int) error
	UpsertUserIdentifiers(ctx context.Context, userID int, username string, aliases []string, email string) error
	FindUserByIdentifier(ctx context.Context, identifier string) (domain.User, error)
}
