package domains

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// normalizeDomain normalizes domain into a canonical form.
func normalizeDomain(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return ""
	}
	raw = strings.SplitN(raw, "/", 2)[0]
	raw = strings.ToLower(raw)
	return raw
}

// parseDomains parses domains from the provided input.
func parseDomains(input string) []string {
	var out []string
	for _, part := range strings.FieldsFunc(input, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' }) {
		normalized := normalizeDomain(part)
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

// wantsJSON returns JSON.
func wantsJSON(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/json") || strings.Contains(accept, "json")
}

// RandomTokenURL generates token URL using available randomness.
func RandomTokenURL(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "pin:" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return "pin:" + strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
