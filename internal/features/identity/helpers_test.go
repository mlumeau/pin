package identity

import (
	"testing"

	"pin/internal/domain"
)

func TestVisibleIdentityFiltersPrivateFields(t *testing.T) {
	links := []domain.Link{
		{Label: "Public", URL: "https://example.com"},
		{Label: "Private", URL: "https://private.example.com"},
	}
	social := []domain.SocialProfile{
		{Label: "Mastodon", URL: "https://social.example/@user"},
		{Label: "Private", URL: "https://private.example.com"},
	}

	user := domain.User{
		Username:           "alice",
		Email:              "alice@example.com",
		Phone:              "123",
		Location:           "Paris",
		LinksJSON:          EncodeLinks(links),
		SocialProfilesJSON: EncodeSocialProfiles(social),
		WalletsJSON:        EncodeStringMap(map[string]string{"btc": "1", "eth": "2"}),
		PublicKeysJSON:     EncodeStringMap(map[string]string{"pgp": "key"}),
		CustomFieldsJSON:   EncodeStringMap(map[string]string{"foo": "bar"}),
		VisibilityJSON: EncodeVisibilityMap(map[string]string{
			"email":      "private",
			"phone":      "private",
			"link:1":     "private",
			"social:1":   "private",
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
	publicLinks := DecodeLinks(publicUser.LinksJSON)
	if len(publicLinks) != 1 || publicLinks[0].Label != "Public" {
		t.Fatalf("expected only public link")
	}
	publicSocial := DecodeSocialProfiles(publicUser.SocialProfilesJSON)
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
	out, visibility := ParseSocialForm(
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
	if visibility[SocialVisibilityKey(0)] != "public" {
		t.Fatalf("expected public visibility, got %q", visibility[SocialVisibilityKey(0)])
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

func TestEncodeDecodeVisibilityMapRoundTrip(t *testing.T) {
	values := map[string]string{
		"email":      "PRIVATE",
		"wallet.btc": "public",
		"custom.foo": "private",
		"link:1":     "public",
	}
	encoded := EncodeVisibilityMap(values)
	if encoded == "" {
		t.Fatalf("expected encoded visibility map")
	}
	decoded := DecodeVisibilityMap(encoded)
	if decoded["email"] != "private" {
		t.Fatalf("expected normalized email visibility, got %q", decoded["email"])
	}
	if decoded["wallet.btc"] != "public" {
		t.Fatalf("expected wallet visibility, got %q", decoded["wallet.btc"])
	}
	if decoded["custom.foo"] != "private" {
		t.Fatalf("expected custom visibility, got %q", decoded["custom.foo"])
	}
	if decoded["link:1"] != "public" {
		t.Fatalf("expected link visibility, got %q", decoded["link:1"])
	}
}

func TestVisibleIdentityFiltersLinksAndSocialByIndex(t *testing.T) {
	user := domain.User{
		LinksJSON:          EncodeLinks([]domain.Link{{Label: "Keep", URL: "https://keep.example"}, {Label: "Drop", URL: "https://drop.example"}}),
		SocialProfilesJSON: EncodeSocialProfiles([]domain.SocialProfile{{Label: "Keep", URL: "https://social.keep"}, {Label: "Drop", URL: "https://social.drop"}}),
		VisibilityJSON: EncodeVisibilityMap(map[string]string{
			LinkVisibilityKey(1):   "private",
			SocialVisibilityKey(1): "private",
		}),
	}
	publicUser, _ := VisibleIdentity(user, false)
	links := DecodeLinks(publicUser.LinksJSON)
	if len(links) != 1 || links[0].Label != "Keep" {
		t.Fatalf("expected only public link, got %+v", links)
	}
	social := DecodeSocialProfiles(publicUser.SocialProfilesJSON)
	if len(social) != 1 || social[0].Label != "Keep" {
		t.Fatalf("expected only public social profile, got %+v", social)
	}
}

func TestVisibleIdentityFiltersWalletsByVisibilityKey(t *testing.T) {
	user := domain.User{
		WalletsJSON: EncodeStringMap(map[string]string{"BTC": "1", "eth": "2"}),
		VisibilityJSON: EncodeVisibilityMap(map[string]string{
			"wallet.btc": "private",
		}),
	}
	publicUser, _ := VisibleIdentity(user, false)
	wallets := DecodeStringMap(publicUser.WalletsJSON)
	if _, ok := wallets["BTC"]; ok {
		t.Fatalf("expected BTC wallet to be removed")
	}
	if wallets["eth"] != "2" {
		t.Fatalf("expected ETH wallet to remain")
	}
}

func TestEncodeDecodeLinks(t *testing.T) {
	links := []domain.Link{{Label: "Site", URL: "https://example.com"}}
	encoded := EncodeLinks(links)
	if encoded == "" {
		t.Fatalf("expected encoded links")
	}
	decoded := DecodeLinks(encoded)
	if len(decoded) != 1 || decoded[0].URL != "https://example.com" {
		t.Fatalf("unexpected decoded links: %+v", decoded)
	}
}

func TestEncodeDecodeStringSlice(t *testing.T) {
	values := []string{"one", "two"}
	encoded := EncodeStringSlice(values)
	if encoded == "" {
		t.Fatalf("expected encoded string slice")
	}
	decoded := DecodeStringSlice(encoded)
	if len(decoded) != 2 || decoded[1] != "two" {
		t.Fatalf("unexpected decoded slice: %+v", decoded)
	}
}

func TestWalletsMapToStructsSkipsEmpty(t *testing.T) {
	wallets := map[string]string{"": "1", "btc": " "}
	out := WalletsMapToStructs(wallets)
	if len(out) != 0 {
		t.Fatalf("expected empty wallets, got %+v", out)
	}
}

func TestPublicKeysMapToStructsFiltersEmpty(t *testing.T) {
	keys := map[string]string{"pgp": " ", "ssh": "key"}
	out := PublicKeysMapToStructs(keys)
	if len(out) != 1 || out[0].Algorithm != "ssh" {
		t.Fatalf("unexpected public keys: %+v", out)
	}
}

func TestVerifiedDomainsSliceToStructsSkipsEmpty(t *testing.T) {
	domains := []string{"", "example.com"}
	out := VerifiedDomainsSliceToStructs(domains)
	if len(out) != 1 || out[0].Domain != "example.com" {
		t.Fatalf("unexpected verified domains: %+v", out)
	}
}

func TestEncodeDecodeStringMap(t *testing.T) {
	values := map[string]string{"foo": "bar"}
	encoded := EncodeStringMap(values)
	if encoded == "" {
		t.Fatalf("expected encoded map")
	}
	decoded := DecodeStringMap(encoded)
	if decoded["foo"] != "bar" {
		t.Fatalf("unexpected decoded map: %+v", decoded)
	}
}

func TestEncodeDecodeSocialProfiles(t *testing.T) {
	social := []domain.SocialProfile{{Label: "Mastodon", URL: "https://social.example"}}
	encoded := EncodeSocialProfiles(social)
	if encoded == "" {
		t.Fatalf("expected encoded social profiles")
	}
	decoded := DecodeSocialProfiles(encoded)
	if len(decoded) != 1 || decoded[0].Label != "Mastodon" {
		t.Fatalf("unexpected decoded social profiles: %+v", decoded)
	}
}

func TestAliasesSliceToStructsSkipsEmpty(t *testing.T) {
	aliases := []string{"", "alt"}
	out := AliasesSliceToStructs(aliases)
	if len(out) != 1 || out[0].Name != "alt" {
		t.Fatalf("unexpected aliases: %+v", out)
	}
}

func TestEncodeDecodeCustomFields(t *testing.T) {
	fields := []domain.CustomField{{Key: "site", Value: "example"}}
	encoded := EncodeCustomFields(fields)
	if encoded == "" {
		t.Fatalf("expected encoded custom fields")
	}
	decoded := DecodeCustomFields(encoded)
	if len(decoded) != 1 || decoded[0].Key != "site" {
		t.Fatalf("unexpected decoded custom fields: %+v", decoded)
	}
}

func TestEncodeDecodeWallets(t *testing.T) {
	wallets := []domain.Wallet{{Label: "btc", Address: "1"}}
	encoded := EncodeWallets(wallets)
	if encoded == "" {
		t.Fatalf("expected encoded wallets")
	}
	decoded := DecodeWallets(encoded)
	if len(decoded) != 1 || decoded[0].Label != "btc" {
		t.Fatalf("unexpected decoded wallets: %+v", decoded)
	}
}

func TestEncodeDecodePublicKeys(t *testing.T) {
	keys := []domain.PublicKey{{Algorithm: "pgp", Key: "key"}}
	encoded := EncodePublicKeys(keys)
	if encoded == "" {
		t.Fatalf("expected encoded public keys")
	}
	decoded := DecodePublicKeys(encoded)
	if len(decoded) != 1 || decoded[0].Algorithm != "pgp" {
		t.Fatalf("unexpected decoded public keys: %+v", decoded)
	}
}

func TestEncodeDecodeVerifiedDomains(t *testing.T) {
	domains := []domain.VerifiedDomain{{Domain: "example.com"}}
	encoded := EncodeVerifiedDomains(domains)
	if encoded == "" {
		t.Fatalf("expected encoded verified domains")
	}
	decoded := DecodeVerifiedDomains(encoded)
	if len(decoded) != 1 || decoded[0].Domain != "example.com" {
		t.Fatalf("unexpected decoded verified domains: %+v", decoded)
	}
}

func TestEncodeDecodeAliases(t *testing.T) {
	aliases := []domain.Alias{{Name: "alt"}}
	encoded := EncodeAliases(aliases)
	if encoded == "" {
		t.Fatalf("expected encoded aliases")
	}
	decoded := DecodeAliases(encoded)
	if len(decoded) != 1 || decoded[0].Name != "alt" {
		t.Fatalf("unexpected decoded aliases: %+v", decoded)
	}
}
