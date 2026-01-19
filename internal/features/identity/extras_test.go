package identity

import (
	"database/sql"
	"testing"
	"time"

	"pin/internal/domain"
)

func TestDomainsHelpers(t *testing.T) {
	rows := []domain.DomainVerification{
		{Domain: "example.com", VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}},
		{Domain: "unused.com", VerifiedAt: sql.NullTime{Valid: false}},
		{Domain: "test.com", VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}},
	}
	text := DomainsToText(rows)
	if text != "example.com, unused.com, test.com" {
		t.Fatalf("unexpected text: %q", text)
	}
	verified := VerifiedDomains(rows)
	if len(verified) != 2 {
		t.Fatalf("expected 2 verified domains, got %d", len(verified))
	}
	if !IsATProtoHandleVerified("user.example.com", verified) {
		t.Fatalf("expected handle to be verified")
	}
	if IsATProtoHandleVerified("user.other.com", verified) {
		t.Fatalf("expected handle to be unverified")
	}
}

func TestMergeSocialProfiles(t *testing.T) {
	existing := []domain.SocialProfile{
		{URL: "https://github.com/alice", Provider: "github", Verified: true},
	}
	updated := []domain.SocialProfile{
		{URL: "https://github.com/alice", Label: "GitHub"},
	}
	out := MergeSocialProfiles(updated, existing)
	if len(out) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(out))
	}
	if !out[0].Verified {
		t.Fatalf("expected verified flag to be preserved")
	}
	if out[0].Provider != "github" {
		t.Fatalf("expected provider to be preserved")
	}
}
