# PIN

PIN stands for Personal Identity Node. It is a lightweight, self-hosted identity server designed to be the authoritative source of one or multiple people and organizations on the web. It exposes clean profile pages, multi-profile management in settings, and machine-readable identity exports intended for humans, bots, and agents.

## Philosophy

### Identity as infrastructure

Modern web identity is fragmented across platforms, formats, and protocols. PIN treats identity as
foundational infrastructure, in the same class as DNS or email:

- One canonical source of truth
- Stable, long-lived URLs
- Predictable, explicit formats
- Compatibility over time rather than novelty

The goal is durability and reliability.

### Ownership and locality

PIN is designed to be:

- Self-hosted by default
- Operable on a single machine
- Useful without external services

All identity data lives locally. There is no dependency on third-party APIs for core functionality.
Optional integrations exist, but the system remains fully functional without them.

### Protocol-first, UI-second

PIN prioritizes interfaces and exports over presentation:

- Humans get a clean profile page
- Machines get structured, deterministic representations
- Developers get stable endpoints with minimal ambiguity

The identity data is the real product, the web page is only a view.

### Boring by design

PIN intentionally favors:

- Conservative technology choices
- Explicit configuration
- Few moving parts
- Minimal background processes

If the software disappears for ten years, the data should still be intelligible.

## Non-goals

PIN explicitly does not aim to:

- Replace personal websites or blogs
- Act as a social network or community platform
- Host long-form content or media libraries
- Provide analytics, tracking, or growth features
- Act as a general authentication provider
- Optimize for virality or social reach

If a feature does not strengthen identity clarity or portability, it likely does not belong here.

## Install

### Prerequisites

- Go 1.22+
- SQLite (bundled via Go driver)
- `cwebp` installed and available in PATH (for WebP encoding)

### Clone and run

```bash
git clone <repo>
cd pin
go run .
```

## Run

By default, PIN listens on `127.0.0.1:5000` and stores data in `identity.db` in the project folder.

```bash
go run .
```

### Backup/export

See `docs/backup.md`.

### Key endpoints

See `docs/endpoints.md`.

## Configuration

All settings are configured through environment variables. See `docs/configuration.md`.

## Architecture

The codebase uses a feature-first layout with small platform layers. Routes and handlers live with their feature (public, auth, admin, domains, invites, passkeys, OAuth, profile pictures, MCP, identity, federation, health, settings), while shared concerns sit under `internal/config`, `internal/domain`, `internal/platform/http`, `internal/platform/server`, `internal/platform/storage/sqlite`, `internal/platform/media`, and `internal/platform/core`. See `docs/architecture.md` for conventions.

## Docs

- `docs/getting-started.md`
- `docs/testing.md`
- `docs/configuration.md`
- `docs/deployment.md`
- `docs/backup.md`
- `docs/themes.md`
- `docs/endpoints.md`

## Contributing

1. Create a feature branch.
2. Keep changes small and focused.
3. Run `go run .` to validate behavior.
4. Use ASCII in text files unless a file already uses Unicode.

## Roadmap

## License

Apache-2.0. See `LICENSE`.
