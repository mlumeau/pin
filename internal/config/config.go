package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime settings loaded from environment variables.
type Config struct {
	Env                string
	IsProd             bool
	SecretKey          []byte
	DBPath             string
	StaticDir          string
	UploadsDir         string
	ProfilePictureDir  string
	AllowedExts        map[string]bool
	Host               string
	Port               string
	CookieSecure       bool
	CookieSameSite     http.SameSite
	MaxUploadBytes     int64
	AdminUser          string
	AdminPassword      string
	AdminEmail         string
	TOTPSecret         string
	DisableCSRF        bool
	BaseURL            string
	GitHubClientID     string
	GitHubClientSecret string
	RedditClientID     string
	RedditClientSecret string
	RedditUserAgent    string
	BlueskyPDS         string
	CacheAltFormats    bool
	MCPEnabled         bool
	MCPToken           string
	MCPReadOnly        bool
}

// LoadConfig reads environment variables, applies defaults, and validates required settings.
func LoadConfig() (Config, error) {
	env := strings.ToLower(strings.TrimSpace(getEnv("PIN_ENV", "development")))
	isProd := env == "production"

	secret := os.Getenv("PIN_SECRET_KEY")
	if secret == "" && isProd {
		return Config{}, errors.New("PIN_SECRET_KEY is required in production")
	}
	if secret == "" {
		secret = randomSecret(32)
	}

	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(getEnv("PIN_COOKIE_SAMESITE", "lax")) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	maxUpload := int64(10 * 1024 * 1024)
	if v := os.Getenv("PIN_MAX_UPLOAD_BYTES"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxUpload = parsed
		}
	}

	uploadsDir := getEnv("PIN_UPLOADS_DIR", filepath.Join(getBaseDir(), "static", "uploads"))

	return Config{
		Env:                env,
		IsProd:             isProd,
		SecretKey:          []byte(secret),
		DBPath:             getEnv("PIN_DB_PATH", filepath.Join(getBaseDir(), "identity.db")),
		StaticDir:          filepath.Join(getBaseDir(), "static"),
		UploadsDir:         uploadsDir,
		ProfilePictureDir:  filepath.Join(uploadsDir, "profile-pictures"),
		AllowedExts:        map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true},
		Host:               getEnv("PIN_HOST", "127.0.0.1"),
		Port:               getEnv("PIN_PORT", "5000"),
		CookieSecure:       envBool("PIN_COOKIE_SECURE", isProd),
		CookieSameSite:     sameSite,
		MaxUploadBytes:     maxUpload,
		AdminUser:          getEnv("PIN_ADMIN_USERNAME", "admin"),
		AdminPassword:      os.Getenv("PIN_ADMIN_PASSWORD"),
		AdminEmail:         getEnv("PIN_ADMIN_EMAIL", "admin@example.com"),
		TOTPSecret:         os.Getenv("PIN_TOTP_SECRET"),
		DisableCSRF:        envBool("PIN_DISABLE_CSRF", !isProd),
		BaseURL:            strings.TrimRight(getEnv("PIN_BASE_URL", ""), "/"),
		GitHubClientID:     os.Getenv("PIN_OAUTH_GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("PIN_OAUTH_GITHUB_CLIENT_SECRET"),
		RedditClientID:     os.Getenv("PIN_OAUTH_REDDIT_CLIENT_ID"),
		RedditClientSecret: os.Getenv("PIN_OAUTH_REDDIT_CLIENT_SECRET"),
		RedditUserAgent:    getEnv("PIN_OAUTH_REDDIT_USER_AGENT", "pin/1.0"),
		BlueskyPDS:         strings.TrimRight(getEnv("PIN_BSKY_PDS", "https://bsky.social"), "/"),
		CacheAltFormats:    envBool("PIN_CACHE_ALT_FORMATS", false),
		MCPEnabled:         envBool("PIN_MCP_ENABLED", true),
		MCPToken:           os.Getenv("PIN_MCP_TOKEN"),
		MCPReadOnly:        envBool("PIN_MCP_READONLY", true),
	}, nil
}

// getBaseDir returns the working directory or executable directory as a fallback.
func getBaseDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// randomSecret returns a hex token, falling back to a timestamp on RNG failure.
func randomSecret(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// getEnv returns the environment value or fallback when empty.
func getEnv(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

// envBool parses common boolean env values and falls back when empty/invalid.
func envBool(name string, fallback bool) bool {
	v := os.Getenv(name)
	if v == "" {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
