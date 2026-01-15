package invites

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for invites.
type Repository interface {
	CreateInvite(ctx context.Context, token, role string, createdBy int) error
	ListInvites(ctx context.Context) ([]domain.Invite, error)
	GetInviteByToken(ctx context.Context, token string) (domain.Invite, error)
	MarkInviteUsed(ctx context.Context, id int, usedBy int) error
	DeleteInvite(ctx context.Context, id int) error
}
