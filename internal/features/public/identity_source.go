package public

import (
	"context"
	"net/http"

	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/features/profilepicture"
)

type identitySource struct {
	deps Dependencies
}

func (s identitySource) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return s.deps.GetOwnerUser(ctx)
}

func (s identitySource) VisibleIdentity(user domain.User, isPrivate bool) (domain.User, map[string]string) {
	return identity.VisibleIdentity(user, isPrivate)
}

func (s identitySource) ActiveProfilePictureAlt(ctx context.Context, user domain.User) string {
	return profilepicture.NewService(s.deps).ActiveAlt(ctx, user)
}

func (s identitySource) BaseURL(r *http.Request) string {
	return s.deps.BaseURL(r)
}
