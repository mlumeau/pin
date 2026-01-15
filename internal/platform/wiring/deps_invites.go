package wiring

import (
	"context"

	"pin/internal/domain"
)

// Invites.
func (d Deps) CreateInvite(ctx context.Context, token, role string, createdBy int) error {
	return d.repos.Invites.CreateInvite(ctx, token, role, createdBy)
}

func (d Deps) ListInvites(ctx context.Context) ([]domain.Invite, error) {
	return d.repos.Invites.ListInvites(ctx)
}

func (d Deps) GetInviteByToken(ctx context.Context, token string) (domain.Invite, error) {
	return d.repos.Invites.GetInviteByToken(ctx, token)
}

func (d Deps) MarkInviteUsed(ctx context.Context, id int, usedBy int) error {
	return d.repos.Invites.MarkInviteUsed(ctx, id, usedBy)
}

func (d Deps) DeleteInvite(ctx context.Context, id int) error {
	return d.repos.Invites.DeleteInvite(ctx, id)
}
