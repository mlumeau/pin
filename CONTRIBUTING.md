# Contributing

Thanks for helping improve PIN. We aim for small, focused changes that are easy to review.

## Quick start
- Fork and create a feature branch.
- Run `go test ./...` before opening a PR.
- Keep changes scoped to a single concern.

## Project structure
- Features live under `internal/features/<feature>`.
- Platform layers are under `internal/platform/*` (routing, server helpers, storage, wiring).
- Core types and shared helpers are in `internal/domain`.
- Repo interfaces live in `internal/contracts`.

See `docs/architecture.md` for conventions.

## Coding guidelines
- Prefer small files and feature-local helpers.
- Handlers should be HTTP-only (parse, validate, call services, render/redirect).
- Services own business logic; repositories handle persistence.
- Avoid global registration or hidden wiring; wire dependencies explicitly.
- Use ASCII in text files unless the file already uses Unicode.

## Testing
- Add `net/http/httptest` coverage for new routes/handlers.
- Keep tests fast; prefer unit tests over full integration runs.

## Submitting changes
- Open a PR with a clear description and the reason for the change.
- If your change alters behavior, include tests or a short manual test plan.

## Security
If you find a security issue, please open a private report instead of a public issue.
