package admin

import (
	"strings"

	"pin/internal/domain"
)

// isAdmin reports whether admin is true.
func isAdmin(user domain.User) bool {
	return strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
}
