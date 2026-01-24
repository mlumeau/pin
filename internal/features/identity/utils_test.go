package identity

import (
	"testing"

	"pin/internal/domain"
)

// TestFromIdent verifies from ident behavior.
func TestFromIdent(t *testing.T) {
	name, ext := FromIdent("alice.json")
	if name != "alice" || ext != "json" {
		t.Fatalf("expected alice/json, got %q/%q", name, ext)
	}
	name, ext = FromIdent("alice.profile.txt")
	if name != "alice.profile" || ext != "txt" {
		t.Fatalf("expected alice.profile/txt, got %q/%q", name, ext)
	}
	name, ext = FromIdent("bob")
	if name != "bob" || ext != "" {
		t.Fatalf("expected bob/empty, got %q/%q", name, ext)
	}
	name, ext = FromIdent("alice.doc")
	if name != "alice.doc" || ext != "" {
		t.Fatalf("expected no extension, got %q/%q", name, ext)
	}
}

// TestExtensionFromPath verifies extension from path behavior.
func TestExtensionFromPath(t *testing.T) {
	if got := ExtensionFromPath("/json"); got != "json" {
		t.Fatalf("expected json, got %q", got)
	}
	if got := ExtensionFromPath("/alice.xml"); got != "" {
		t.Fatalf("expected empty for named export, got %q", got)
	}
	if got := ExtensionFromPath(""); got != "" {
		t.Fatalf("expected empty for blank path, got %q", got)
	}
}

// TestFirstNonEmpty verifies first non empty behavior.
func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "  ", "pin"); got != "pin" {
		t.Fatalf("expected pin, got %q", got)
	}
	if got := FirstNonEmpty("  ", ""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

// TestEscapeVCard verifies escape v card behavior.
func TestEscapeVCard(t *testing.T) {
	value := "a,b;c\\d\n"
	if got := EscapeVCard(value); got != "a\\,b\\;c\\\\d\\n" {
		t.Fatalf("unexpected vcard escape: %q", got)
	}
}

// TestSanitizeVCardKey verifies sanitize v card key behavior.
func TestSanitizeVCardKey(t *testing.T) {
	if got := SanitizeVCardKey(" Foo:Bar.Baz "); got != "FOO_BAR_BAZ" {
		t.Fatalf("unexpected key: %q", got)
	}
}

// TestSubjectForIdentityNormalizesHandle verifies subject for identity normalizes handle behavior.
func TestSubjectForIdentityNormalizesHandle(t *testing.T) {
	first := SubjectForIdentity(domain.Identity{ID: 12, Handle: "Alice"})
	second := SubjectForIdentity(domain.Identity{ID: 12, Handle: " alice "})
	if first != second {
		t.Fatalf("expected normalized subject, got %q and %q", first, second)
	}
	third := SubjectForIdentity(domain.Identity{ID: 13, Handle: "alice"})
	if first == third {
		t.Fatalf("expected different subject for different ID")
	}
}
