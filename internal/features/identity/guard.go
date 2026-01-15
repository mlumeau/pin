package identity

import (
	"context"
	"encoding/json"
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

// ValidateIdentifiers checks for reserved names and collisions.
func ValidateIdentifiers(ctx context.Context, username string, aliases []string, email string, excludeID int, reserved map[string]struct{}, checkCollisions func(context.Context, []string, int) error) error {
	name := strings.ToLower(strings.TrimSpace(username))
	if name == "" {
		return errors.New("Username is required")
	}
	if IsReservedIdentifier(name, reserved) {
		return errors.New("Username is reserved")
	}
	for _, alias := range aliases {
		if IsReservedIdentifier(alias, reserved) {
			return errors.New("Alias is reserved")
		}
	}
	identifiers := BuildIdentifiers(name, aliases, email)
	if err := checkCollisions(ctx, identifiers, excludeID); err != nil {
		return errors.New("Identifier already exists")
	}
	return nil
}

// MatchesIdentity checks an identifier against user identifiers (including hashed).
func MatchesIdentity(user domain.User, identifier string) bool {
	needle := strings.ToLower(strings.TrimSpace(identifier))
	if needle == "" {
		return false
	}

	if strings.EqualFold(user.Username, needle) {
		return true
	}
	if strings.EqualFold(user.Email, needle) {
		return true
	}

	if user.Email != "" {
		email := strings.ToLower(strings.TrimSpace(user.Email))
		if needle == Sha256Hex(email) {
			return true
		}
	}
	if user.Username != "" {
		name := strings.ToLower(strings.TrimSpace(user.Username))
		if needle == Sha256Hex(name) {
			return true
		}
	}

	var aliases []string
	if user.AliasesJSON != "" {
		_ = json.Unmarshal([]byte(user.AliasesJSON), &aliases)
	}
	for _, alias := range aliases {
		if strings.EqualFold(alias, needle) {
			return true
		}
		if needle == Sha256Hex(strings.ToLower(strings.TrimSpace(alias))) {
			return true
		}
	}
	return false
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
