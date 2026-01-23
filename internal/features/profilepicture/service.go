package profilepicture

import (
	"context"
	"strings"

	"pin/internal/domain"
)

type Store interface {
	ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error)
	CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error)
	SetActiveProfilePicture(ctx context.Context, identityID int, pictureID int64) error
	DeleteProfilePicture(ctx context.Context, identityID int, pictureID int64) (string, error)
	UpdateProfilePictureAlt(ctx context.Context, identityID int, pictureID int64, alt string) error
	ClearProfilePictureSelection(ctx context.Context, identityID int) error
	GetProfilePictureFilename(ctx context.Context, identityID int, pictureID int64) (string, error)
	GetProfilePictureAlt(ctx context.Context, identityID int, pictureID int64) (string, error)
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s Service) List(ctx context.Context, identityID int) ([]domain.ProfilePicture, error) {
	return s.store.ListProfilePictures(ctx, identityID)
}

func (s Service) Create(ctx context.Context, identityID int, filename, alt string) (int64, error) {
	return s.store.CreateProfilePicture(ctx, identityID, filename, alt)
}

func (s Service) Select(ctx context.Context, identityID int, pictureID int64) error {
	return s.store.SetActiveProfilePicture(ctx, identityID, pictureID)
}

func (s Service) Delete(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return s.store.DeleteProfilePicture(ctx, identityID, pictureID)
}

func (s Service) UpdateAlt(ctx context.Context, identityID int, pictureID int64, alt string) error {
	return s.store.UpdateProfilePictureAlt(ctx, identityID, pictureID, alt)
}

func (s Service) ClearSelection(ctx context.Context, identityID int) error {
	return s.store.ClearProfilePictureSelection(ctx, identityID)
}

func (s Service) Filename(ctx context.Context, identityID int, pictureID int64) (string, error) {
	return s.store.GetProfilePictureFilename(ctx, identityID, pictureID)
}

func (s Service) ActiveAlt(ctx context.Context, identity domain.Identity) string {
	if identity.ProfilePictureID.Valid {
		if alt, err := s.store.GetProfilePictureAlt(ctx, identity.ID, identity.ProfilePictureID.Int64); err == nil && strings.TrimSpace(alt) != "" {
			return alt
		}
	}
	return ""
}
