package admin

import (
	"strings"

	"pin/internal/domain"
)

func isAdmin(user domain.User) bool {
	return strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
}
