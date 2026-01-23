package export

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pin/internal/domain"
	"pin/internal/features/identity"
)

type pincSource struct {
	baseURL string
	alt     string
}

func (p pincSource) GetOwnerIdentity(ctx context.Context) (domain.Identity, error) {
	return domain.Identity{}, nil
}

func (p pincSource) VisibleIdentity(user domain.Identity, isPrivate bool) (domain.Identity, map[string]string) {
	return user, nil
}

func (p pincSource) ActiveProfilePictureAlt(ctx context.Context, user domain.Identity) string {
	return p.alt
}

func (p pincSource) BaseURL(r *http.Request) string {
	return p.baseURL
}

func TestBuildPINCUsesSelfProfileImageAndMeta(t *testing.T) {
	source := pincSource{baseURL: "https://pin.example", alt: "Portrait"}
	handler := NewHandler(source)
	selfURL := "https://pin.example/p/abc/def.json"

	env, err := handler.BuildPINC(context.Background(), httptest.NewRequest(http.MethodGet, "/alice", nil), domain.Identity{
		ID:            10,
		Handle:        "alice",
		DisplayName:   "",
		Email:         " alice@example.com ",
		ATProtoHandle: " @alice ",
	}, map[string]string{
		"note":  " hello ",
		"empty": " ",
	}, "private", selfURL)
	if err != nil {
		t.Fatalf("build pinc: %v", err)
	}

	if env.Meta.Version != identity.PincVersion {
		t.Fatalf("expected version %q, got %q", identity.PincVersion, env.Meta.Version)
	}
	if env.Meta.BaseURL != "https://pin.example" {
		t.Fatalf("expected base URL, got %q", env.Meta.BaseURL)
	}
	if env.Meta.View != "private" {
		t.Fatalf("expected private view, got %q", env.Meta.View)
	}
	if env.Meta.Subject != identity.SubjectForIdentity(domain.Identity{ID: 10, Handle: "alice"}) {
		t.Fatalf("expected subject for identity, got %q", env.Meta.Subject)
	}
	if env.Meta.Self != selfURL {
		t.Fatalf("expected self URL, got %q", env.Meta.Self)
	}

	if env.Identity.DisplayName != "alice" {
		t.Fatalf("expected fallback display name, got %q", env.Identity.DisplayName)
	}
	if env.Identity.Email != "alice@example.com" {
		t.Fatalf("expected trimmed email, got %q", env.Identity.Email)
	}
	if env.Identity.ATProtoHandle != "@alice" {
		t.Fatalf("expected trimmed atproto handle, got %q", env.Identity.ATProtoHandle)
	}
	if env.Identity.ProfileImage != "https://pin.example/p/abc/def/profile-picture" {
		t.Fatalf("expected private profile image, got %q", env.Identity.ProfileImage)
	}
	if env.Identity.ImageAltText != "Portrait" {
		t.Fatalf("expected profile image alt, got %q", env.Identity.ImageAltText)
	}
	if env.Identity.CustomFields["note"] != "hello" || len(env.Identity.CustomFields) != 1 {
		t.Fatalf("expected custom fields stripped, got %+v", env.Identity.CustomFields)
	}
	if _, err := time.Parse(time.RFC3339, env.Identity.UpdatedAt); err != nil {
		t.Fatalf("expected updated_at, got %q", env.Identity.UpdatedAt)
	}
}

func TestComputePINCRevStableForMaps(t *testing.T) {
	payload := pincIdentity{
		Handle:      "alice",
		DisplayName: "alice",
		URL:         "https://pin.example/alice",
		UpdatedAt:   "2025-01-01T00:00:00Z",
		CustomFields: map[string]string{
			"b": "2",
			"a": "1",
		},
		Wallets: map[string]string{
			"eth": "2",
			"btc": "1",
		},
		PublicKeys: map[string]string{
			"ssh": "key",
			"pgp": "key2",
		},
	}
	rev1 := computePINCRev(payload)

	payload.CustomFields = map[string]string{"a": "1", "b": "2"}
	payload.Wallets = map[string]string{"btc": "1", "eth": "2"}
	payload.PublicKeys = map[string]string{"pgp": "key2", "ssh": "key"}
	rev2 := computePINCRev(payload)

	if rev1 != rev2 {
		t.Fatalf("expected stable rev, got %q and %q", rev1, rev2)
	}
}
