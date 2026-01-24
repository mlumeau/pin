package users

import (
	"testing"

	"pin/internal/domain"
)

// TestBuildVisibilityMapPrefixesCustomKeys verifies build visibility map prefixes custom keys behavior.
func TestBuildVisibilityMapPrefixesCustomKeys(t *testing.T) {
	fieldVisibility := map[string]string{
		"email": "private",
	}
	customVisibility := map[string]string{
		"favorite": "public",
	}
	out := BuildVisibilityMap(fieldVisibility, customVisibility)
	if out["email"] != "private" {
		t.Fatalf("expected field visibility to be preserved")
	}
	if out["custom:favorite"] != "public" {
		t.Fatalf("expected custom key to be prefixed")
	}
}

// TestVisibilityCustomMapFiltersOnlyCustomKeys verifies visibility custom map filters only custom keys behavior.
func TestVisibilityCustomMapFiltersOnlyCustomKeys(t *testing.T) {
	visibility := map[string]string{
		"email":           "private",
		"custom:favorite": "public",
	}
	out := VisibilityCustomMap(visibility)
	if _, ok := out["email"]; ok {
		t.Fatalf("expected non-custom key to be excluded")
	}
	if out["favorite"] != "public" {
		t.Fatalf("expected custom key to be returned without prefix")
	}
}

// TestDomainVisibilityMapNormalizesDomains verifies domain visibility map normalizes domains behavior.
func TestDomainVisibilityMapNormalizesDomains(t *testing.T) {
	visibility := map[string]string{
		"verified_domain:Example.COM": "private",
		"verified_domain:":            "public",
	}
	out := DomainVisibilityMap(visibility)
	if out["example.com"] != "private" {
		t.Fatalf("expected normalized domain key")
	}
	if len(out) != 1 {
		t.Fatalf("expected empty domain to be dropped")
	}
}

// TestBuildLinkEntriesUsesVisibilityIndex verifies build link entries uses visibility index behavior.
func TestBuildLinkEntriesUsesVisibilityIndex(t *testing.T) {
	links := []domain.Link{
		{Label: "One", URL: "https://one.example"},
		{Label: "Two", URL: "https://two.example"},
	}
	visibility := map[string]string{
		"link:1": "private",
	}
	out := BuildLinkEntries(links, visibility)
	if len(out) != 2 {
		t.Fatalf("expected two link entries, got %d", len(out))
	}
	if out[1].Visibility != "private" {
		t.Fatalf("expected second link visibility to be private")
	}
}

// TestBuildSocialEntriesAllowsProviderLabel verifies build social entries allows provider label behavior.
func TestBuildSocialEntriesAllowsProviderLabel(t *testing.T) {
	social := []domain.SocialProfile{
		{Provider: "mastodon", URL: "https://social.example"},
	}
	visibility := map[string]string{
		"social:0": "public",
	}
	out := BuildSocialEntries(social, visibility)
	if len(out) != 1 {
		t.Fatalf("expected one social entry")
	}
	if out[0].Visibility != "public" {
		t.Fatalf("expected visibility to be public")
	}
}
