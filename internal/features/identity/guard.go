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

// ValidateHandle validates handle and returns an error on failure.
func ValidateHandle(ctx context.Context, handle string, excludeID int, reserved map[string]struct{}, checkCollision func(context.Context, string, int) error) error {
	normalized := strings.ToLower(strings.TrimSpace(handle))
	if normalized == "" {
		return errors.New("Handle is required")
	}
	if !isURLSafeHandle(normalized) {
		return errors.New("Handle can only contain letters, numbers, dot (.), underscore (_), and hyphen (-)")
	}
	if IsReservedIdentifier(normalized, reserved) {
		return errors.New("Handle is reserved")
	}
	if err := checkCollision(ctx, normalized, excludeID); err != nil {
		return errors.New("Handle already exists")
	}
	return nil
}

// isURLSafeHandle reports whether a handle is URL-path safe for profile routes.
func isURLSafeHandle(handle string) bool {
	for i := 0; i < len(handle); i++ {
		c := handle[i]
		switch {
		case c >= 'a' && c <= 'z':
			continue
		case c >= '0' && c <= '9':
			continue
		case c == '.', c == '_', c == '-':
			continue
		default:
			return false
		}
	}
	return true
}

// MatchesIdentity reports whether the identifier matches the user's handle.
func MatchesIdentity(identity domain.Identity, identifier string) bool {
	needle := strings.ToLower(strings.TrimSpace(identifier))
	if needle == "" {
		return false
	}
	return strings.EqualFold(identity.Handle, needle)
}

// RouteSegment returns segment.
func RouteSegment(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}
