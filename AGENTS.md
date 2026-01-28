# AGENTS

This file orients coding agents working on PIN. Follow repo conventions and keep changes scoped.

## Quick context
- Product: PIN (Personal Identity Node), a Go identity server that implements the PINC contract.
- Stack: Go 1.22+, net/http, SQLite (via modernc.org/sqlite), server-rendered templates.
- Run locally: `go run .` (binds to 127.0.0.1:5000 by default).

## Project structure
- Feature-first code lives in `internal/features/<feature>`.
- Platform layers live under `internal/platform/*` (http router, server helpers, storage, wiring).
- Core types and shared helpers are in `internal/domain` and `internal/config`.
- Repo interfaces live in `internal/contracts`.
- Templates: `templates/<feature>` with shared partials in `templates/partials`.
- Static assets: `static/` (subfolders by type or feature).

## Code conventions
- Keep files small and feature-local; split handlers by concern when they grow.
- Handlers are HTTP-only: parse/validate, call services, render/redirect/write JSON.
- Services own business logic and avoid `net/http` types.
- Repositories handle persistence and live in `internal/platform/storage/sqlite`.
- Avoid cross-feature imports; prefer small interfaces/adapters in `internal/contracts` or `internal/platform/wiring`.
- Wire dependencies explicitly in `main.go` or `internal/platform/http` (avoid globals).
- Use ASCII in text files unless the file already uses Unicode.

## Local setup and config
- Configuration is via environment variables only (no `.env` file by default).
- Config defaults and env var list live in `docs/user/configuration.md`; update it when adding or changing settings.

## Format/lint
- Format Go code with `gofmt -w .`.
- Run `go vet ./...` for basic static checks when touching core logic.

## Database schema
- SQLite schema is initialized in `internal/platform/storage/sqlite/db_init.go`.
- If you change schema, keep it backward-compatible or include a safe manual migration note in docs.

## Tests
- Tests live next to code (same package or `package_test`).
- Prefer `net/http/httptest` for handler coverage.
- Shared test helpers live in `internal/testutil`.
- Run all tests: `go test ./...` (PowerShell: `go test --% ./...` when needed).

## Docs update expectations
- If you add/modify endpoints or response formats, update `docs/user/endpoints.md`.
- If you add/modify config settings, update `docs/user/configuration.md`.

## Documents worth checking
- `docs/contributor/architecture.md` for deeper layout and dependency guidance.
- `docs/contributor/testing.md` for testing details.
- `RFC-PINC.md` for the protocol contract.

## Files to avoid touching
- `identity.db`, `identity.db-shm`, `identity.db-wal` (local dev DB files).
- `pin`, `pin.exe` (built artifacts).
