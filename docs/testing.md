# Testing

## Scope
- Unit tests for helpers and services.
- Handler tests for HTTP routes using `httptest`.
- Smoke tests for router and template wiring.

## Where tests live
Tests live next to the code they cover (same package or `package_test`). Shared helpers live in `internal/testutil`.

## Run tests
```bash
go test ./...
```

Disable caching when needed:
```bash
go test ./... -count=1
```

Run a focused subset:
```bash
go test ./... -run TestName
```

## Coverage
Bash or zsh:
```bash
go test ./... -coverpkg=./... -coverprofile=coverage.out
go tool cover -func coverage.out
go tool cover -html=coverage.out
```

PowerShell (use `--%` to avoid argument parsing):
```powershell
go test --% ./... -coverpkg=./... -coverprofile=coverage.out
go tool cover -func coverage.out
go tool cover -html=coverage.out
```

## Tips
- Use `t.TempDir()` for filesystem work and cleanups.
- Avoid external services and network calls in unit tests.
- Prefer table-driven tests for input/output coverage.
- Use `internal/testutil` helpers when wiring servers or loading templates.
