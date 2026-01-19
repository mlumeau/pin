package identity

import "testing"

func TestBuildIdentifiers(t *testing.T) {
	ids := BuildIdentifiers("Alice", []string{"Alias", "alias"}, "ALICE@example.com")
	if len(ids) != 6 {
		t.Fatalf("expected 6 identifiers, got %d", len(ids))
	}
	seen := map[string]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	expected := []string{
		"alice",
		"alice@example.com",
		"alias",
		Sha256Hex("alice"),
		Sha256Hex("alice@example.com"),
		Sha256Hex("alias"),
	}
	for _, value := range expected {
		if !seen[value] {
			t.Fatalf("missing identifier %q", value)
		}
	}
}
