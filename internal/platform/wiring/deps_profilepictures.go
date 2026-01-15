package wiring

import (
	"context"

	"pin/internal/domain"
)

// ProfilePictures.
func (d Deps) ListProfilePictures(ctx context.Context, userID int) ([]domain.ProfilePicture, error) {
	return d.repos.ProfilePictures.ListProfilePictures(ctx, userID)
}

func (d Deps) CreateProfilePicture(ctx context.Context, userID int, filename, alt string) (int64, error) {
	return d.repos.ProfilePictures.CreateProfilePicture(ctx, userID, filename, alt)
}

func (d Deps) SetActiveProfilePicture(ctx context.Context, userID int, pictureID int64) error {
	return d.repos.ProfilePictures.SetActiveProfilePicture(ctx, userID, pictureID)
}

func (d Deps) DeleteProfilePicture(ctx context.Context, userID int, pictureID int64) (string, error) {
	return d.repos.ProfilePictures.DeleteProfilePicture(ctx, userID, pictureID)
}

func (d Deps) UpdateProfilePictureAlt(ctx context.Context, userID int, pictureID int64, alt string) error {
	return d.repos.ProfilePictures.UpdateProfilePictureAlt(ctx, userID, pictureID, alt)
}

func (d Deps) ClearProfilePictureSelection(ctx context.Context, userID int) error {
	return d.repos.ProfilePictures.ClearProfilePictureSelection(ctx, userID)
}

func (d Deps) GetProfilePictureFilename(ctx context.Context, userID int, pictureID int64) (string, error) {
	return d.repos.ProfilePictures.GetProfilePictureFilename(ctx, userID, pictureID)
}

func (d Deps) GetProfilePictureAlt(ctx context.Context, userID int, pictureID int64) (string, error) {
	return d.repos.ProfilePictures.GetProfilePictureAlt(ctx, userID, pictureID)
}
