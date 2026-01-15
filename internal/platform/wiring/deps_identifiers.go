package wiring

import (
	"context"

	"pin/internal/domain"
)

// Identifiers.
func (d Deps) CheckIdentifierCollisions(ctx context.Context, identifiers []string, excludeID int) error {
	return d.repos.Identifiers.CheckIdentifierCollisions(ctx, identifiers, excludeID)
}

func (d Deps) UpsertUserIdentifiers(ctx context.Context, userID int, username string, aliases []string, email string) error {
	return d.repos.Identifiers.UpsertUserIdentifiers(ctx, userID, username, aliases, email)
}

func (d Deps) FindUserByIdentifier(ctx context.Context, identifier string) (domain.User, error) {
	return d.repos.Identifiers.FindUserByIdentifier(ctx, identifier)
}
