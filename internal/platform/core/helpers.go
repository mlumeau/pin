package core

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

// WriteJSON writes a JSON response with the default content type.
func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(data)
}

// BaseURL returns scheme and host from the incoming request.
func BaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// IsSafeRedirect reports whether the redirect stays on the same host or is relative.
func IsSafeRedirect(r *http.Request, target string) bool {
	if target == "" {
		return false
	}
	u, err := url.Parse(target)
	if err != nil {
		return false
	}
	if u.IsAbs() {
		return u.Host == r.Host && (u.Scheme == "http" || u.Scheme == "https")
	}
	return strings.HasPrefix(target, "/")
}

// FirstNonEmpty returns the first non-empty, trimmed string.
func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// RandomToken returns a hex token, falling back to a timestamp on RNG failure.
func RandomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// RandomTokenURL returns a URL-safe token, falling back to a timestamp on RNG failure.
func RandomTokenURL(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "pin:" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return "pin:" + strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

// SubtleCompare performs a constant-time string comparison.
func SubtleCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SessionUserID extracts user_id from session values and normalizes numeric types.
func SessionUserID(session *sessions.Session) (int, bool) {
	switch v := session.Values["user_id"].(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// NormalizeDomain trims whitespace, strips scheme/path, and lowercases to a bare host.
func NormalizeDomain(raw string) string {
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

// ShortHash returns a truncated SHA-256 hex string.
func ShortHash(value string, length int) string {
	hash := Sha256Hex(value)
	if length <= 0 || length >= len(hash) {
		return hash
	}
	return hash[:length]
}

// Sha256Hex returns the hex-encoded SHA-256 digest of the input.
func Sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
