package wiring

import "context"

// Settings.
func (d Deps) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	return d.repos.Settings.GetSettings(ctx, keys...)
}

func (d Deps) GetSetting(ctx context.Context, key string) (string, bool, error) {
	return d.repos.Settings.GetSetting(ctx, key)
}

func (d Deps) SetSetting(ctx context.Context, key, value string) error {
	return d.repos.Settings.SetSetting(ctx, key, value)
}

func (d Deps) SetSettings(ctx context.Context, values map[string]string) error {
	return d.repos.Settings.SetSettings(ctx, values)
}

func (d Deps) DeleteSetting(ctx context.Context, key string) error {
	return d.repos.Settings.DeleteSetting(ctx, key)
}
