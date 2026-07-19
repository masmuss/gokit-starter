# Changelog

## Unreleased

### Architecture

- Rename folders: `internal/delivery` → `internal/inbound`, `internal/infra` → `internal/outbound`
- Rename `modules/auth/infra` → `modules/auth/repository`
- Add `AGENTS.md` — AI agent guide for project conventions
- Add `docs/architecture.md` and `docs/repository-structure.md`

### Release

- Add GoReleaser config (`.goreleaser.yml`)
- Add release workflow (git tag → build + changelog + GitHub release)
- Add `VERSION` file

---

## 0.1.0

### Architecture

- Module-based layered architecture (`domain` / `app` / `handler` / `repository`)
- Manual dependency injection — `wire.go` per module, wired in `cmd/server/main.go`
- Remove `uber-fx`, `entgo.io/ent`, `ariga.io/atlas` dependencies

### Database

- Replace Ent ORM with GORM
- Auto-migrate on startup (`gorm.AutoMigrate`)
- In-memory SQLite for integration tests

### Auth

- JWT access + refresh token support
- Token blacklist via Redis (logout)
- Role-based access control (`admin` / `member`)
- `RequireRole` middleware
- Account status check on login (banned/inactive blocked)
- Password change endpoint
- Organization-scoped profile queries
- JWT `jti` claim for token revocation

### Security

- Error messages sanitized (non-apperr → generic "internal server error")
- CORS configuration fixed (removed invalid `AllowCredentials: true` with wildcard)
- Token blacklist fail-closed on Redis error
- Graceful shutdown via signal handling + error channel

### Logging

- Structured audit logger (`internal/pkg/audit/`)
- Audit events: register, login, logout, password change, token refresh, token validation
- Scoped loggers per module (`slog.With("module", "auth")`)

### API

- OpenAPI spec via `swaggest/openapi-go` + Scalar UI
- Health check and version endpoints
- Route registration via `RouteRegistrar` interface
- Doc registration via `OperationRegistrar` interface

### CI/CD

- GitHub Actions: lint, test, integration test
- Lefthook pre-commit hooks (format, test, lint)

### Code Quality

- `Service.Config` struct (8 params → 1)
- `TokenSubject` struct for JWT issuance
- `serverConfig` struct for `runServer`
- `authtoken` package (no more `auth`/`auth` name collision)
- `chi` import aliases removed
- Unused exports made private (`fail`, `tokenTypeAccess`, `tokenTypeRefresh`)
