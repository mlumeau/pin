package profilepicture

import (
	"context"
	"strings"

	"pin/internal/domain"
)

type Store interface {
	ListProfilePictures(ctx context.Context, userID int) ([]domain.ProfilePicture, error)
	CreateProfilePicture(ctx context.Context, userID int, filename, alt string) (int64, error)
	SetActiveProfilePicture(ctx context.Context, userID int, pictureID int64) error
	DeleteProfilePicture(ctx context.Context, userID int, pictureID int64) (string, error)
	UpdateProfilePictureAlt(ctx context.Context, userID int, pictureID int64, alt string) error
	ClearProfilePictureSelection(ctx context.Context, userID int) error
	GetProfilePictureFilename(ctx context.Context, userID int, pictureID int64) (string, error)
	GetProfilePictureAlt(ctx context.Context, userID int, pictureID int64) (string, error)
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s Service) List(ctx context.Context, userID int) ([]domain.ProfilePicture, error) {
	return s.store.ListProfilePictures(ctx, userID)
}

func (s Service) Create(ctx context.Context, userID int, filename, alt string) (int64, error) {
	return s.store.CreateProfilePicture(ctx, userID, filename, alt)
}

func (s Service) Select(ctx context.Context, userID int, pictureID int64) error {
	return s.store.SetActiveProfilePicture(ctx, userID, pictureID)
}

func (s Service) Delete(ctx context.Context, userID int, pictureID int64) (string, error) {
	return s.store.DeleteProfilePicture(ctx, userID, pictureID)
}

func (s Service) UpdateAlt(ctx context.Context, userID int, pictureID int64, alt string) error {
	return s.store.UpdateProfilePictureAlt(ctx, userID, pictureID, alt)
}

func (s Service) ClearSelection(ctx context.Context, userID int) error {
	return s.store.ClearProfilePictureSelection(ctx, userID)
}

func (s Service) Filename(ctx context.Context, userID int, pictureID int64) (string, error) {
	return s.store.GetProfilePictureFilename(ctx, userID, pictureID)
}

func (s Service) ActiveAlt(ctx context.Context, user domain.User) string {
	if user.ProfilePictureID.Valid {
		if alt, err := s.store.GetProfilePictureAlt(ctx, user.ID, user.ProfilePictureID.Int64); err == nil && strings.TrimSpace(alt) != "" {
			return alt
		}
	}
	return ""
}
