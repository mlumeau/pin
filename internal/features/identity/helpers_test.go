package identity

import (
	"encoding/json"
	"testing"

	"pin/internal/domain"
)

func TestVisibleIdentityFiltersPrivateFields(t *testing.T) {
	links := []domain.Link{
		{Label: "Public", URL: "https://example.com"},
		{Label: "Private", URL: "https://private.example.com", Visibility: "private"},
	}
	linksJSON, _ := json.Marshal(links)
	social := []domain.SocialProfile{
		{Label: "Mastodon", URL: "https://social.example/@user"},
		{Label: "Private", URL: "https://private.example.com", Visibility: "private"},
	}
	socialJSON, _ := json.Marshal(social)

	user := domain.User{
		Username:           "alice",
		Email:              "alice@example.com",
		Phone:              "123",
		Location:           "Paris",
		LinksJSON:          string(linksJSON),
		SocialProfilesJSON: string(socialJSON),
		WalletsJSON:        EncodeStringMap(map[string]string{"btc": "1", "eth": "2"}),
		PublicKeysJSON:     EncodeStringMap(map[string]string{"pgp": "key"}),
		CustomFieldsJSON:   EncodeStringMap(map[string]string{"foo": "bar"}),
		VisibilityJSON: EncodeStringMap(map[string]string{
			"email":      "private",
			"phone":      "private",
			"wallet.btc": "private",
			"key.pgp":    "private",
			"custom.foo": "private",
		}),
	}

	publicUser, customFields := VisibleIdentity(user, false)
	if publicUser.Email != "" {
		t.Fatalf("expected email to be filtered")
	}
	if publicUser.Phone != "" {
		t.Fatalf("expected phone to be filtered")
	}
	var publicLinks []domain.Link
	if err := json.Unmarshal([]byte(publicUser.LinksJSON), &publicLinks); err != nil {
		t.Fatalf("failed to parse links: %v", err)
	}
	if len(publicLinks) != 1 || publicLinks[0].Label != "Public" {
		t.Fatalf("expected only public link")
	}
	var publicSocial []domain.SocialProfile
	if err := json.Unmarshal([]byte(publicUser.SocialProfilesJSON), &publicSocial); err != nil {
		t.Fatalf("failed to parse social profiles: %v", err)
	}
	if len(publicSocial) != 1 || publicSocial[0].Label != "Mastodon" {
		t.Fatalf("expected only public social profile")
	}
	wallets := DecodeStringMap(publicUser.WalletsJSON)
	if _, ok := wallets["btc"]; ok {
		t.Fatalf("expected private wallet to be removed")
	}
	keys := DecodeStringMap(publicUser.PublicKeysJSON)
	if _, ok := keys["pgp"]; ok {
		t.Fatalf("expected private key to be removed")
	}
	if publicUser.CustomFieldsJSON != "" {
		t.Fatalf("expected custom fields to be removed from user")
	}
	if customFields["foo"] != "bar" {
		t.Fatalf("expected custom fields to be returned separately")
	}

	privateUser, _ := VisibleIdentity(user, true)
	if privateUser.Email != "alice@example.com" {
		t.Fatalf("expected email to remain in private view")
	}
}

func TestParseSocialForm(t *testing.T) {
	out := ParseSocialForm(
		[]string{" GitHub ", "", "X"},
		[]string{" https://github.com/user ", "https://ignored", ""},
		[]string{"public", "private", "private"},
	)
	if len(out) != 1 {
		t.Fatalf("expected 1 social profile, got %d", len(out))
	}
	if out[0].Label != "GitHub" || out[0].URL != "https://github.com/user" {
		t.Fatalf("unexpected profile data: %+v", out[0])
	}
}

func TestNormalizeVisibility(t *testing.T) {
	if got := NormalizeVisibility("private"); got != "private" {
		t.Fatalf("expected private, got %q", got)
	}
	if got := NormalizeVisibility("PUBLIC"); got != "public" {
		t.Fatalf("expected public, got %q", got)
	}
	if got := NormalizeVisibility("unknown"); got != "public" {
		t.Fatalf("expected fallback to public, got %q", got)
	}
}
