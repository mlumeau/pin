package users

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for users.
type Repository interface {
	LoadUser(ctx context.Context) (domain.User, error)
	UpdateUser(ctx context.Context, u domain.User) error
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	GetUserByPrivateToken(ctx context.Context, token string) (domain.User, error)
	GetOwnerUser(ctx context.Context) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	ListUsersPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.User, int, error)
	UpdatePrivateToken(ctx context.Context, userID int, token string) error
	HasUser(ctx context.Context) (bool, error)
	CreateUser(ctx context.Context, username, email, role, passwordHash, totpSecret, themeProfile, privateToken string) (int64, error)
	DeleteUser(ctx context.Context, userID int) error
	ResetAllUserThemes(ctx context.Context, themeValue string) error
	UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error
}
