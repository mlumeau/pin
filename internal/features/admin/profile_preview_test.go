package admin

import (
	"strings"
	"testing"
	"time"

	"pin/internal/domain"
	"pin/internal/features/identity"
	featuresettings "pin/internal/features/settings"
	"pin/internal/platform/core"
)

func TestBuildProfilePreviewDataBuildsPublicAndPrivateSnapshots(t *testing.T) {
	visibility := map[string]string{
		"display_name": "private",
		"bio":          "private",
		"email":        "private",
	}
	identityRecord := domain.Identity{
		Handle:              "alice",
		DisplayName:         "Alice",
		Bio:                 "hello",
		Email:               "alice@example.com",
		PrivateToken:        "tok_123",
		CustomFieldsJSON:    identity.EncodeStringMap(map[string]string{"note": "hello"}),
		LinksJSON:           identity.EncodeLinks([]domain.Link{{Label: "Site", URL: "https://example.com"}}),
		SocialProfilesJSON:  identity.EncodeSocialProfiles([]domain.SocialProfile{{Label: "Social", URL: "https://social.example"}}),
		WalletsJSON:         identity.EncodeStringMap(map[string]string{"btc": "addr"}),
		PublicKeysJSON:      identity.EncodeStringMap(map[string]string{"pgp": "PGPKEY"}),
		VerifiedDomainsJSON: identity.EncodeStringSlice([]string{"example.com"}),
		VisibilityJSON:      identity.EncodeVisibilityMap(visibility),
		UpdatedAt:           time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	preview := buildProfilePreviewData(identityRecord, featuresettings.ThemeSettings{})

	publicPreview := preview.Data.Public
	privatePreview := preview.Data.Private

	if publicPreview.User.Email != "" {
		t.Fatalf("expected public email redacted, got %q", publicPreview.User.Email)
	}
	if publicPreview.User.DisplayName != "" {
		t.Fatalf("expected public display name redacted, got %q", publicPreview.User.DisplayName)
	}
	if publicPreview.User.Bio != "" {
		t.Fatalf("expected public bio redacted, got %q", publicPreview.User.Bio)
	}
	if privatePreview.User.Email != "alice@example.com" {
		t.Fatalf("expected private email present, got %q", privatePreview.User.Email)
	}
	if privatePreview.User.DisplayName != "Alice" {
		t.Fatalf("expected private display name present, got %q", privatePreview.User.DisplayName)
	}
	if privatePreview.User.Bio != "hello" {
		t.Fatalf("expected private bio present, got %q", privatePreview.User.Bio)
	}
	if len(publicPreview.Links) != 1 || publicPreview.Links[0].Label != "Site" {
		t.Fatalf("expected public links present, got %+v", publicPreview.Links)
	}
	if publicPreview.CustomFields["note"] != "hello" {
		t.Fatalf("expected custom field in public preview, got %+v", publicPreview.CustomFields)
	}

	publicPath := "/" + identityRecord.Handle
	if publicPreview.ExportBase != publicPath {
		t.Fatalf("expected public export base %q, got %q", publicPath, publicPreview.ExportBase)
	}
	privateHash := core.ShortHash(strings.ToLower(strings.TrimSpace(identityRecord.Handle)), 7)
	privatePath := "/p/" + privateHash + "/" + identityRecord.PrivateToken
	if privatePreview.ExportBase != privatePath {
		t.Fatalf("expected private export base %q, got %q", privatePath, privatePreview.ExportBase)
	}
	if publicPreview.UpdatedAt.IsZero() || privatePreview.UpdatedAt.IsZero() {
		t.Fatalf("expected updated timestamps to be set")
	}
}

func TestBuildProfilePreviewDataFallsBackWithoutToken(t *testing.T) {
	identityRecord := domain.Identity{
		Handle:    "bob",
		UpdatedAt: time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC),
	}

	preview := buildProfilePreviewData(identityRecord, featuresettings.ThemeSettings{})

	publicPath := "/" + identityRecord.Handle
	if preview.Data.Public.ExportBase != publicPath {
		t.Fatalf("expected public export base %q, got %q", publicPath, preview.Data.Public.ExportBase)
	}
	if preview.Data.Private.ExportBase != publicPath {
		t.Fatalf("expected private export base fallback %q, got %q", publicPath, preview.Data.Private.ExportBase)
	}
}
