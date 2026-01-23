# Architecture

This project uses a feature-first layout with thin, reusable platform layers. The goal is small files, clear ownership, and easy navigation for contributors.
For the external contract and protocol details, see RFC-PINC.md.

## High-level layers
- `main.go` - explicit wiring for the binary (kept minimal).
- `internal/config` - configuration loading and environment defaults.
- `internal/domain` - core types and validation helpers shared across features.
- `internal/contracts` - repository interfaces (ports) grouped by feature, plus the repo bundle used for wiring.
- `internal/platform/core` - shared utilities (tokens, hashing, URL helpers).
- `internal/platform/media` - image processing helpers shared across features.
- `internal/platform/server` - server wiring, middleware (security headers, auth/CSRF), template helpers.
- `internal/platform/http` - router setup that registers feature routes.
- `internal/platform/transport` - small interfaces for HTTP wiring and middleware contracts.
- `internal/platform/wiring` - adapters that bind server + repos to feature dependency interfaces.
- `internal/platform/storage` - storage helpers (backup/export); concrete repos live in sqlite.
- `internal/platform/storage/sqlite` - DB init/migrations plus repositories grouped by feature (users/identifiers, invites, domains, profile pictures, passkeys, audit, settings).
- Feature packages (under `internal/features/`): `public`, `auth`, `admin`, `domains`, `invites`, `passkeys`, `oauth`, `profilepicture`, `mcp`, `identity`, `federation`, `health`, `settings`. Each owns its handlers + service logic; they depend on interfaces from platform layers.

## Handler/service/repo conventions
- Clear separation of concerns: handlers only speak HTTP, services own business rules, repositories handle persistence.
- Handlers: HTTP only. Parse/validate request, call a service, render/redirect/write JSON. No DB access or template plumbing here.
- Services/use-cases: feature logic, composed from repositories and helpers. Return typed results/errors; no `http` types.
- Repositories: concrete persistence in `internal/platform/storage/sqlite`. Expose interfaces consumed by services so tests can use fakes.

## Wiring and adapters
- Wiring is explicit in `main.go` (no init-based registration).
- `internal/platform/http` builds the mux and registers each feature.
- `internal/platform/wiring` adapts server + repo bundles to feature dependency interfaces.
- Template helpers are injected when building the server (ex: `identity.TemplateFuncs()`).

## Dependency direction
- `main` and `internal/platform/http` wire features together; they are the only places that import many features.
- Platform packages stay feature-agnostic (no feature-specific helpers inside `internal/platform/*`).
- Features depend on `internal/domain`, `internal/config`, and platform interfaces, not on platform internals.
- Cross-feature use should be explicit and narrow (ex: `federation` uses `identity/export` through a small source adapter).

## Routing
- Central router in `internal/platform/http` wires static files and feature subrouters. Features register their routes so routes live next to handlers.
- Security headers and `requireLogin` middleware live in `internal/platform/server`.
- Setup redirect is feature-specific and lives in `internal/features/public`.

## Templates and assets
- Group by feature: `templates/public`, `templates/settings`, `templates/auth`, `templates/admin`, `templates/invites`, `templates/passkeys`, etc.
- Shared fragments in `templates/partials`.
- Static assets remain in `static/` with subfolders by type or feature (`static/js/settings/appearance.js`, etc.).

## Testing
- Favor `net/http/httptest` to cover routes and handler behavior.
- For template safety, add lightweight render tests that load all templates and render with minimal data.

## Guidelines
- Keep feature files small; split handlers by concern and keep services focused.
- Prefer interface-driven dependencies at feature boundaries.
- Avoid global registries; wire dependencies explicitly in `main` or `platform/http`.
