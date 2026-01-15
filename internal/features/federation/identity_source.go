package federation

import (
	"context"
	"net/http"

	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/features/identity/export"
	"pin/internal/features/profilepicture"
	"pin/internal/platform/core"
)

type identitySource struct {
	deps Dependencies
	pics profilepicture.Store
}

func NewExportSource(deps Dependencies, pics profilepicture.Store) export.Source {
	return identitySource{deps: deps, pics: pics}
}

func (s identitySource) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return s.deps.GetOwnerUser(ctx)
}

func (s identitySource) VisibleIdentity(user domain.User, isPrivate bool) (domain.User, map[string]string) {
	return identity.VisibleIdentity(user, isPrivate)
}

func (s identitySource) ActiveProfilePictureAlt(ctx context.Context, user domain.User) string {
	return profilepicture.NewService(s.pics).ActiveAlt(ctx, user)
}

func (s identitySource) BaseURL(r *http.Request) string {
	return core.BaseURL(r)
}
