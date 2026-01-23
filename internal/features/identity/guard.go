package identity

import (
	"context"
	"errors"
	"strings"

	"pin/internal/domain"
)

// IsReservedPath returns true if the first path segment is reserved.
func IsReservedPath(path string, reserved map[string]struct{}) bool {
	if path == "" || path == "/" {
		return true
	}
	segment := RouteSegment(path)
	if segment == "" {
		return true
	}
	_, ok := reserved[segment]
	return ok
}

// IsReservedIdentifier returns true if the identifier is reserved.
func IsReservedIdentifier(identifier string, reserved map[string]struct{}) bool {
	name := strings.ToLower(strings.TrimSpace(identifier))
	if name == "" {
		return true
	}
	_, ok := reserved[name]
	return ok
}

// ValidateHandle checks for reserved names and collisions.
func ValidateHandle(ctx context.Context, handle string, excludeID int, reserved map[string]struct{}, checkCollision func(context.Context, string, int) error) error {
	normalized := strings.ToLower(strings.TrimSpace(handle))
	if normalized == "" {
		return errors.New("Handle is required")
	}
	if IsReservedIdentifier(normalized, reserved) {
		return errors.New("Handle is reserved")
	}
	if err := checkCollision(ctx, normalized, excludeID); err != nil {
		return errors.New("Handle already exists")
	}
	return nil
}

// MatchesIdentity checks whether the identifier matches the user's handle.
func MatchesIdentity(identity domain.Identity, identifier string) bool {
	needle := strings.ToLower(strings.TrimSpace(identifier))
	if needle == "" {
		return false
	}
	return strings.EqualFold(identity.Handle, needle)
}

// RouteSegment returns the first path segment (lowercased).
func RouteSegment(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}
