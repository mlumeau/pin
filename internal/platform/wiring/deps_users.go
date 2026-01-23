package wiring

import (
	"context"

	"pin/internal/domain"
)

// Users.
func (d Deps) HasUser(ctx context.Context) (bool, error) {
	return d.repos.Users.HasUser(ctx)
}

func (d Deps) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	return d.repos.Users.GetUserByID(ctx, id)
}

func (d Deps) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return d.repos.Users.GetOwnerUser(ctx)
}

func (d Deps) ListUsers(ctx context.Context) ([]domain.User, error) {
	return d.repos.Users.ListUsers(ctx)
}

func (d Deps) ListUsersPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.User, int, error) {
	return d.repos.Users.ListUsersPaged(ctx, query, sort, dir, limit, offset)
}

func (d Deps) CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error) {
	return d.repos.Users.CreateUser(ctx, role, passwordHash, totpSecret, themeProfile)
}

func (d Deps) DeleteUser(ctx context.Context, userID int) error {
	return d.repos.Users.DeleteUser(ctx, userID)
}

func (d Deps) UpdateUser(ctx context.Context, user domain.User) error {
	return d.repos.Users.UpdateUser(ctx, user)
}

func (d Deps) ResetAllUserThemes(ctx context.Context, themeValue string) error {
	return d.repos.Users.ResetAllUserThemes(ctx, themeValue)
}

func (d Deps) UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	return d.repos.Users.UpdateUserTheme(ctx, userID, themeProfile, customCSSPath, customCSSInline)
}
