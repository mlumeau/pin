package users

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for users.
type Repository interface {
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	GetOwnerUser(ctx context.Context) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	ListUsersPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.User, int, error)
	HasUser(ctx context.Context) (bool, error)
	CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error)
	UpdateUser(ctx context.Context, u domain.User) error
	DeleteUser(ctx context.Context, userID int) error
	ResetAllUserThemes(ctx context.Context, themeValue string) error
	UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error
}
