package wiring

import (
	"context"

	"pin/internal/domain"
)

// ProfilePictures.
func (d Deps) ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error) {
	return d.repos.ProfilePictures.ListProfilePictures(ctx, identityID)
}

func (d Deps) CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error) {
	return d.repos.ProfilePictures.CreateProfilePicture(ctx, identityID, filename, alt)
}

func (d Deps) SetActiveProfilePicture(ctx context.Context, identityID int, pictureID int64) error {
	return d.repos.ProfilePictures.SetActiveProfilePicture(ctx, identityID, pictureID)
}

func (d Deps) DeleteProfilePicture(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return d.repos.ProfilePictures.DeleteProfilePicture(ctx, identityID, pictureID)
}

func (d Deps) UpdateProfilePictureAlt(ctx context.Context, identityID int, pictureID int64, alt string) error {
	return d.repos.ProfilePictures.UpdateProfilePictureAlt(ctx, identityID, pictureID, alt)
}

func (d Deps) ClearProfilePictureSelection(ctx context.Context, identityID int) error {
	return d.repos.ProfilePictures.ClearProfilePictureSelection(ctx, identityID)
}

func (d Deps) GetProfilePictureFilename(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return d.repos.ProfilePictures.GetProfilePictureFilename(ctx, identityID, pictureID)
}

func (d Deps) GetProfilePictureAlt(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return d.repos.ProfilePictures.GetProfilePictureAlt(ctx, identityID, pictureID)
}
