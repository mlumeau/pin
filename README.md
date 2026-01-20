# PIN

PIN stands for Personal Identity Node. It is a lightweight, self-hosted identity server written in Go and designed to be the authoritative source of one or multiple people and organizations on the web. It exposes clean profile pages, multi-profile management in settings, and machine-readable identity exports intended for humans, bots, and agents.

## Quick start

### Prerequisites

- Go 1.22+
- SQLite (bundled via Go driver)
- `cwebp` installed and available in PATH (for WebP encoding). Install: macOS `brew install webp`; Linux `sudo apt-get install webp` (Debian/Ubuntu) or `sudo dnf install libwebp-tools`; Windows `choco install webp`.

### Clone and run

```bash
git clone <repo>
cd pin
go run .
```

By default, PIN listens on `127.0.0.1:5000` and stores data in `identity.db` in the project folder.

### Next steps

- Getting started guide: see [docs/getting-started.md](docs/getting-started.md).
- Configuration: all settings are configured through environment variables. See [docs/configuration.md](docs/configuration.md).
- Backup/export: see [docs/backup.md](docs/backup.md).
- Key endpoints: see [docs/endpoints.md](docs/endpoints.md).

## Documentation

- [docs/getting-started.md](docs/getting-started.md)
- [docs/testing.md](docs/testing.md)
- [docs/configuration.md](docs/configuration.md)
- [docs/deployment.md](docs/deployment.md)
- [docs/backup.md](docs/backup.md)
- [docs/themes.md](docs/themes.md)
- [docs/endpoints.md](docs/endpoints.md)

## Technical stack

- Go (net/http) with server-rendered HTML templates
- SQLite storage via `modernc.org/sqlite` (pure Go)
- Auth/session stack: passkeys (WebAuthn), TOTP (`pquerna/otp`), `gorilla/sessions`
- Media/encoding: WebP via `cwebp` + `golang.org/x/image`, QR codes via `go-qrcode`

## Architecture

The codebase uses a feature-first layout with small platform layers. Routes and handlers live with their feature (public, auth, admin, domains, invites, passkeys, OAuth, profile pictures, MCP, identity, federation, health, settings), while shared concerns sit under `internal/config`, `internal/domain`, `internal/platform/http`, `internal/platform/server`, `internal/platform/storage/sqlite`, `internal/platform/media`, and `internal/platform/core`. See [docs/architecture.md](docs/architecture.md) for conventions.

## Philosophy

### Identity as infrastructure

Identity on the web is scattered across platforms, formats, and protocols. They change often, accumulate metadata opportunistically, and optimize for visibility instead of stability.

PIN starts from a different point of view: identity should behave like infrastructure. Once configured, it should be safe to rely on without constant attention. This leads to a small number of strong assumptions: one authoritative source, stable URLs, explicit formats, and a preference for compatibility over experimentation.

### Ownership and locality

Identity data is not transient, and PIN does not treat it as such.

All data is stored locally and managed on a single machine. The system is fully usable without network dependencies or third-party services. External integrations exist only to extend reach, never to replace core functionality. Losing an integration must not result in data loss or semantic breakage.

### Protocol-first, UI-second

PIN is built around structured data and well-defined interfaces. The profile page is not the primary product, identity is. Identity information is exposed through deterministic, machine-readable formats with stable endpoints. These exports are intended to be consumed directly by programs, tools, and agents. The web UI is simply one consumer of that data, no different in principle from any other client.

### Boring by design

PIN avoids cleverness, background automation, and implicit behavior. Configuration is explicit. State is visible. The internal model is kept small enough to be understood without extensive context. A system that continues to make sense after years of neglect is preferred over one that requires constant maintenance.

If development were to stop entirely, the data should remain readable and the intent of the system should still be clear.

## Non-goals

PIN explicitly does not aim to:

- Replace personal websites or blogs
- Act as a social network or community platform
- Host long-form content or media libraries
- Provide analytics, tracking, or growth features
- Act as a general authentication provider
- Optimize for virality or social reach

If a feature does not strengthen identity clarity or portability, it likely does not belong here.

## Roadmap

## License

Apache-2.0. See [LICENSE](LICENSE).
