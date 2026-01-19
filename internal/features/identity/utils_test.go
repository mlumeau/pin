package identity

import "testing"

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

func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "  ", "pin"); got != "pin" {
		t.Fatalf("expected pin, got %q", got)
	}
	if got := FirstNonEmpty("  ", ""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestEscapeVCard(t *testing.T) {
	value := "a,b;c\\d\n"
	if got := EscapeVCard(value); got != "a\\,b\\;c\\\\d\\n" {
		t.Fatalf("unexpected vcard escape: %q", got)
	}
}

func TestSanitizeVCardKey(t *testing.T) {
	if got := SanitizeVCardKey(" Foo:Bar.Baz "); got != "FOO_BAR_BAZ" {
		t.Fatalf("unexpected key: %q", got)
	}
}
