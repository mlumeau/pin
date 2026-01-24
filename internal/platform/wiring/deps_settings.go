package wiring

import "context"

// Settings.
func (d Deps) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	return d.repos.Settings.GetSettings(ctx, keys...)
}

// GetSetting returns the setting by delegating to configured services.
func (d Deps) GetSetting(ctx context.Context, key string) (string, bool, error) {
	return d.repos.Settings.GetSetting(ctx, key)
}

// SetSetting sets setting to the provided value by delegating to configured services.
func (d Deps) SetSetting(ctx context.Context, key, value string) error {
	return d.repos.Settings.SetSetting(ctx, key, value)
}

// SetSettings sets settings to the provided value by delegating to configured services.
func (d Deps) SetSettings(ctx context.Context, values map[string]string) error {
	return d.repos.Settings.SetSettings(ctx, values)
}

// DeleteSetting deletes setting by delegating to configured services.
func (d Deps) DeleteSetting(ctx context.Context, key string) error {
	return d.repos.Settings.DeleteSetting(ctx, key)
}
