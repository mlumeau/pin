package profilepictures

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for profile pictures.
type Repository interface {
	ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error)
	CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error)
	SetActiveProfilePicture(ctx context.Context, identityID int, pictureID int64) error
	DeleteProfilePicture(ctx context.Context, identityID int, pictureID int64) (string, error)
	UpdateProfilePictureAlt(ctx context.Context, identityID int, pictureID int64, alt string) error
	ClearProfilePictureSelection(ctx context.Context, identityID int) error
	GetProfilePictureFilename(ctx context.Context, identityID int, pictureID int64) (string, error)
	GetProfilePictureAlt(ctx context.Context, identityID int, pictureID int64) (string, error)
}
