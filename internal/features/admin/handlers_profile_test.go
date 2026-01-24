package admin

import (
	"encoding/json"
	"testing"

	"pin/internal/domain"
	"pin/internal/features/identity"
	"pin/internal/features/users"
)

func TestApplyProfileFormCopiesFields(t *testing.T) {
	start := domain.Identity{
		Handle:      "old",
		DisplayName: "Old",
	}
	form := profileFormData{
		handle:        "alice",
		displayName:   "Alice",
		email:         "alice@example.com",
		bio:           "hello",
		organization:  "Pin Co",
		jobTitle:      "Builder",
		birthdate:     "2000-01-01",
		languages:     "en",
		phone:         "123",
		address:       "street",
		location:      "city",
		website:       "https://example.com",
		pronouns:      "they/them",
		timezone:      "UTC",
		atprotoHandle: "@alice",
		atprotoDID:    "did:example:alice",
		links: []domain.Link{
			{Label: "Site", URL: "https://example.com"},
		},
		customFields: map[string]string{"note": "hello"},
	}
	updated := applyProfileForm(start, form)

	if updated.Handle != "alice" || updated.DisplayName != "Alice" {
		t.Fatalf("expected handle/display updated, got %q/%q", updated.Handle, updated.DisplayName)
	}
	if updated.Organization != "Pin Co" || updated.JobTitle != "Builder" {
		t.Fatalf("expected org/title updated, got %q/%q", updated.Organization, updated.JobTitle)
	}
	if updated.ATProtoHandle != "@alice" || updated.ATProtoDID != "did:example:alice" {
		t.Fatalf("expected atproto fields updated, got %q/%q", updated.ATProtoHandle, updated.ATProtoDID)
	}
	links := identity.DecodeLinks(updated.LinksJSON)
	if len(links) != 1 || links[0].Label != "Site" {
		t.Fatalf("expected links encoded, got %+v", links)
	}
	var custom map[string]string
	if err := json.Unmarshal([]byte(updated.CustomFieldsJSON), &custom); err != nil {
		t.Fatalf("decode custom fields: %v", err)
	}
	if custom["note"] != "hello" {
		t.Fatalf("expected custom note, got %+v", custom)
	}
}

func TestBuildProfileVisibilityMergesSources(t *testing.T) {
	form := profileFormData{
		fieldVisibility:  map[string]string{"email": "private"},
		customFields:     map[string]string{"note": "hello"},
		customVisibility: map[string]string{"note": "private"},
		linkVisibility:   map[string]string{"link:0": "public"},
		socialVisibility: map[string]string{"social:0": "private"},
		domainVisibility: map[string]string{"example.com": "private"},
	}
	visibility := buildProfileVisibility(form)
	if visibility["email"] != "private" {
		t.Fatalf("expected email private, got %q", visibility["email"])
	}
	if visibility["custom:note"] != "private" {
		t.Fatalf("expected custom:note private, got %q", visibility["custom:note"])
	}
	if visibility["link:0"] != "public" {
		t.Fatalf("expected link:0 public, got %q", visibility["link:0"])
	}
	if visibility["social:0"] != "private" {
		t.Fatalf("expected social:0 private, got %q", visibility["social:0"])
	}
	if visibility["verified_domain:example.com"] != "private" {
		t.Fatalf("expected domain private, got %q", visibility["verified_domain:example.com"])
	}
	if visibility["custom:note"] != users.NormalizeVisibility("private") {
		t.Fatalf("expected normalized custom visibility, got %q", visibility["custom:note"])
	}
}
