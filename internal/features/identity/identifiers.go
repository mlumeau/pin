package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Sha256Hex returns the SHA-256 hex digest of the input string.
func Sha256Hex(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])
}

// BuildIdentifiers builds normalized identifiers and their hashed variants.
func BuildIdentifiers(username string, aliases []string, email string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
		hash := Sha256Hex(value)
		if _, ok := seen[hash]; !ok {
			seen[hash] = struct{}{}
			out = append(out, hash)
		}
	}
	add(username)
	add(email)
	for _, alias := range aliases {
		add(alias)
	}
	return out
}
