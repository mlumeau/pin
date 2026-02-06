package settings

import (
	"context"
)

const (
	footerLinkAboutKey = "footer_link_about"
	footerLinkLoginKey = "footer_link_login"
)

// FooterLinksSettings controls visibility of subtle profile footer links.
type FooterLinksSettings struct {
	ShowAbout bool
	ShowLogin bool
}

// FooterLinksSettings returns the effective footer-link visibility settings.
func (s Service) FooterLinksSettings(ctx context.Context) FooterLinksSettings {
	settings := FooterLinksSettings{
		ShowAbout: true,
		ShowLogin: true,
	}
	values, err := s.store.GetSettings(ctx, footerLinkAboutKey, footerLinkLoginKey)
	if err != nil {
		return settings
	}
	settings.ShowAbout = parseSettingBool(values[footerLinkAboutKey])
	settings.ShowLogin = parseSettingBool(values[footerLinkLoginKey])
	return settings
}

// SaveFooterLinksSettings saves profile footer-link visibility settings.
func (s Service) SaveFooterLinksSettings(ctx context.Context, settings FooterLinksSettings) error {
	aboutValue := "0"
	if settings.ShowAbout {
		aboutValue = "1"
	}
	loginValue := "0"
	if settings.ShowLogin {
		loginValue = "1"
	}
	return s.store.SetSettings(ctx, map[string]string{
		footerLinkAboutKey: aboutValue,
		footerLinkLoginKey: loginValue,
	})
}
