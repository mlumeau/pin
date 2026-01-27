# Configuration

PIN is configured via environment variables and read on startup.

## Core
- `PIN_ENV` (default: `development`) - controls defaults for security-related settings.
- `PIN_SECRET_KEY` (required in production) - session signing key; use a long random value.
- `PIN_HOST` (default: `127.0.0.1`) - bind address for the HTTP server.
- `PIN_PORT` (default: `5000`) - bind port for the HTTP server.
- `PIN_BASE_URL` (default: empty) - base URL for absolute links; set when behind a reverse proxy or using HTTPS.
- `PIN_COOKIE_SECURE` (default: `true` in production) - set to `true` when served over HTTPS.
- `PIN_COOKIE_SAMESITE` (default: `lax`) - `lax`, `strict`, or `none` (requires `PIN_COOKIE_SECURE=true`).
- `PIN_MAX_UPLOAD_BYTES` (default: `10485760` = 10MB) - maximum upload size for profile pictures and CSS.
- `PIN_DISABLE_CSRF` (default: `true` in development) - disable CSRF checks (development only).

## Storage
- `PIN_DB_PATH` (default: `./identity.db`) - SQLite file path; ensure the directory exists.
- `PIN_UPLOADS_DIR` (default: `./static/uploads`) - base directory for uploads (profile pictures, themes).
- `PIN_CACHE_ALT_FORMATS` (default: `false`) - cache PNG/JPEG variants of profile pictures.

## OAuth (optional)
Features are active only when their credentials are set.
- `PIN_OAUTH_GITHUB_CLIENT_ID`
- `PIN_OAUTH_GITHUB_CLIENT_SECRET`
- `PIN_OAUTH_REDDIT_CLIENT_ID`
- `PIN_OAUTH_REDDIT_CLIENT_SECRET`
- `PIN_OAUTH_REDDIT_USER_AGENT` (default: `pin/1.0`)
- `PIN_BSKY_PDS` (default: `https://bsky.social`)

## MCP
- `PIN_MCP_ENABLED` (default: `true`) - enable or disable the MCP endpoint.
- `PIN_MCP_TOKEN` (default: empty) - bearer or `X-MCP-Token` auth token.
- `PIN_MCP_READONLY` (default: `true`) - restrict to read-only methods.

## Example
```bash
PIN_ENV=production \
PIN_SECRET_KEY=change-me \
PIN_BASE_URL=https://example.com \
PIN_DB_PATH=/var/lib/pin/identity.db \
PIN_UPLOADS_DIR=/var/lib/pin/uploads \
PIN_COOKIE_SECURE=true \
./pin
```
