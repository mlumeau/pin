package settings

import (
	"context"
	"errors"
	"testing"

	"pin/internal/domain"
)

type footerLinksStore struct {
	values map[string]string
}

func (s *footerLinksStore) GetSettings(ctx context.Context, keys ...string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *footerLinksStore) GetSetting(ctx context.Context, key string) (string, bool, error) {
	value, ok := s.values[key]
	return value, ok, nil
}

func (s *footerLinksStore) SetSetting(ctx context.Context, key, value string) error {
	s.values[key] = value
	return nil
}

func (s *footerLinksStore) SetSettings(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		s.values[key] = value
	}
	return nil
}

func (s *footerLinksStore) DeleteSetting(ctx context.Context, key string) error {
	delete(s.values, key)
	return nil
}

func (s *footerLinksStore) UpdateUserTheme(ctx context.Context, userID int, themeProfile, customCSSPath, customCSSInline string) error {
	return nil
}

func (s *footerLinksStore) GetOwnerUser(ctx context.Context) (domain.User, error) {
	return domain.User{}, errors.New("not found")
}

func TestFooterLinksSettingsDefaultsToVisible(t *testing.T) {
	store := &footerLinksStore{values: map[string]string{
		footerLinkAboutKey: "1",
		footerLinkLoginKey: "1",
	}}
	svc := NewService(store)

	settings := svc.FooterLinksSettings(context.Background())
	if !settings.ShowAbout {
		t.Fatalf("expected about link visible by default")
	}
	if !settings.ShowLogin {
		t.Fatalf("expected login link visible by default")
	}
}

func TestFooterLinksSettingsReadsStoredValues(t *testing.T) {
	store := &footerLinksStore{values: map[string]string{
		footerLinkAboutKey: "0",
		footerLinkLoginKey: "1",
	}}
	svc := NewService(store)

	settings := svc.FooterLinksSettings(context.Background())
	if settings.ShowAbout {
		t.Fatalf("expected about link hidden from stored setting")
	}
	if !settings.ShowLogin {
		t.Fatalf("expected login link visible from stored setting")
	}
}

func TestSaveFooterLinksSettingsPersistsValues(t *testing.T) {
	store := &footerLinksStore{values: map[string]string{}}
	svc := NewService(store)

	err := svc.SaveFooterLinksSettings(context.Background(), FooterLinksSettings{
		ShowAbout: false,
		ShowLogin: true,
	})
	if err != nil {
		t.Fatalf("save footer links settings: %v", err)
	}
	if got := store.values[footerLinkAboutKey]; got != "0" {
		t.Fatalf("expected footer about = 0, got %q", got)
	}
	if got := store.values[footerLinkLoginKey]; got != "1" {
		t.Fatalf("expected footer login = 1, got %q", got)
	}
}
