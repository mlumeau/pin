# Getting Started

## Prerequisites
- Go 1.22+
- SQLite (bundled via Go driver)
- `cwebp` installed and available in PATH (for WebP encoding)

Install `cwebp`:
- macOS: `brew install webp`
- Linux (Debian/Ubuntu): `sudo apt-get install webp`
- Linux (Fedora): `sudo dnf install libwebp-tools`
- Windows: `choco install webp`

## Run locally
```bash
git clone <repo>
cd pin
go run .
```

Open http://127.0.0.1:5000 and complete the setup flow at `/setup` to create the owner account.

## First run checklist
- Visit `/setup` and create the owner account.
- Save the TOTP secret in your authenticator app.
- Log in at `/login` to access settings.

## Configuration
All settings are configured through environment variables. See [configuration.md](configuration.md).

## Build a binary
```bash
go build -o pin
./pin
```

## Where data lives
- Database: `identity.db` (default)
- Uploads: `static/uploads/` (default)

If you change `PIN_DB_PATH` or `PIN_UPLOADS_DIR`, ensure the parent directories exist and are writable.

## Base URL (recommended for deployment)
Set `PIN_BASE_URL` when running behind a reverse proxy or over HTTPS so links and OAuth callbacks are correct.

## Backup/export
See [backup.md](backup.md).

## Tests
```bash
go test ./...
```
