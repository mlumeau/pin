package identity

import (
	"net/http"
	"strings"
)

// FromIdent splits an identifier into name and extension.
func FromIdent(ident string) (string, string) {
	ident = strings.TrimSpace(ident)
	if ident == "" {
		return "", ""
	}
	parts := strings.Split(ident, ".")
	if len(parts) < 2 {
		return ident, ""
	}
	ext := strings.ToLower(parts[len(parts)-1])
	name := strings.Join(parts[:len(parts)-1], ".")
	switch ext {
	case "json", "xml", "txt", "vcf":
		return name, ext
	default:
		return ident, ""
	}
}

// ExtensionFromPath extracts identity export extension from a path.
func ExtensionFromPath(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	switch strings.ToLower(path) {
	case "json", "xml", "txt", "vcf":
		return strings.ToLower(path)
	}
	name, ext := FromIdent(path)
	if name == "" && ext != "" {
		return ext
	}
	return ""
}

func WriteIdentityCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=300, stale-while-revalidate=300")
}

// FirstNonEmpty returns the first non-empty, trimmed string.
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}

func EscapeVCard(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, ";", "\\;")
	value = strings.ReplaceAll(value, ",", "\\,")
	return value
}

func SanitizeVCardKey(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, ":", "_")
	value = strings.ReplaceAll(value, ";", "_")
	value = strings.ReplaceAll(value, ",", "_")
	value = strings.ReplaceAll(value, ".", "_")
	value = strings.ToUpper(value)
	return value
}
