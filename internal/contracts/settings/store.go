package settings

import "context"

// Repository defines persistence operations for key/value settings.
type Repository interface {
	GetSettings(ctx context.Context, keys ...string) (map[string]string, error)
	GetSetting(ctx context.Context, key string) (string, bool, error)
	SetSetting(ctx context.Context, key, value string) error
	SetSettings(ctx context.Context, values map[string]string) error
	DeleteSetting(ctx context.Context, key string) error
}
