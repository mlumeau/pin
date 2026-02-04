package identity

import (
	"context"
	"errors"
	"testing"
)

func TestValidateHandleRejectsInvalidChars(t *testing.T) {
	err := ValidateHandle(context.Background(), "alice/@example", 0, map[string]struct{}{}, func(context.Context, string, int) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected invalid-char error")
	}
	if got, want := err.Error(), "Handle can only contain letters, numbers, dot (.), underscore (_), and hyphen (-)"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateHandleRejectsNonASCII(t *testing.T) {
	err := ValidateHandle(context.Background(), "élise", 0, map[string]struct{}{}, func(context.Context, string, int) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected non-ascii error")
	}
	if got, want := err.Error(), "Handle can only contain letters, numbers, dot (.), underscore (_), and hyphen (-)"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateHandleAllowsURLSafeChars(t *testing.T) {
	err := ValidateHandle(context.Background(), "alice.dev_01-test", 0, map[string]struct{}{}, func(context.Context, string, int) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected URL-safe handle to pass, got %v", err)
	}
}

func TestValidateHandleRejectsCollision(t *testing.T) {
	err := ValidateHandle(context.Background(), "alice", 0, map[string]struct{}{}, func(context.Context, string, int) error {
		return errors.New("collision")
	})
	if err == nil {
		t.Fatalf("expected collision error")
	}
	if got, want := err.Error(), "Handle already exists"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}
