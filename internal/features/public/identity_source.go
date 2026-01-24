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

// GetOwnerIdentity returns the owner identity.
func (s identitySource) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return s.deps.GetOwnerIdentity(ctx)
}

// VisibleIdentity returns the visible identity fields for the requested view.
func (s identitySource) VisibleIdentity(user domain.Identity, isPrivate bool) (domain.Identity, map[string]string) {
	return identity.VisibleIdentity(user, isPrivate)
}

// ActiveProfilePictureAlt returns profile picture alt.
func (s identitySource) ActiveProfilePictureAlt(ctx context.Context, user domain.Identity) string {
	return profilepicture.NewService(s.deps).ActiveAlt(ctx, user)
}

// BaseURL returns the base URL.
func (s identitySource) BaseURL(r *http.Request) string {
	return s.deps.BaseURL(r)
}
