package wiring

import (
	"context"

	"pin/internal/domain"
)

// Invites.
func (d Deps) CreateInvite(ctx context.Context, token, role string, createdBy int) error {
	return d.repos.Invites.CreateInvite(ctx, token, role, createdBy)
}

// ListInvites returns the invites list by delegating to configured services.
func (d Deps) ListInvites(ctx context.Context) ([]domain.Invite, error) {
	return d.repos.Invites.ListInvites(ctx)
}

// GetInviteByToken returns invite by token.
func (d Deps) GetInviteByToken(ctx context.Context, token string) (domain.Invite, error) {
	return d.repos.Invites.GetInviteByToken(ctx, token)
}

// MarkInviteUsed returns invite used.
func (d Deps) MarkInviteUsed(ctx context.Context, id int, usedBy int) error {
	return d.repos.Invites.MarkInviteUsed(ctx, id, usedBy)
}

// DeleteInvite deletes invite by delegating to configured services.
func (d Deps) DeleteInvite(ctx context.Context, id int) error {
	return d.repos.Invites.DeleteInvite(ctx, id)
}
