# AGENTS.md — Gokit Starter AI Guide

This file gives AI coding agents context about the project structure, conventions, and patterns.

## Architecture

Module-based layered architecture. Every feature is a self-contained module under `internal/modules/<name>/`.

```
internal/modules/<name>/
├── domain/       # Business models, errors, constants
├── app/          # Use cases (Service), Repository interface
├── handler/      # HTTP handlers, request DTOs
├── repository/   # GORM implementation of app.Repository
└── wire.go       # Module factory: Dependencies → Module
```

## Key Conventions

### Interfaces belong in the `app` package

The `app` package defines what it needs (`Repository`, `Hasher`, etc.). Implementations live elsewhere (`repository/`, `outbound/`).

```go
// app/service.go — INTERFACE DEFINITION (what the module needs)
type Repository interface {
    FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

// repository/repository.go — IMPLEMENTATION (how it's done)
type Repository struct { db *gorm.DB }
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) { ... }
```

### Wiring is manual, in `wire.go`

No DI framework. Each module exports a `Dependencies` struct (what it needs) and a `Wire()` function (what it provides).

```go
// wire.go
type Dependencies struct {
    DB     *database.DB
    Cache  cache.Cache
    Log    *slog.Logger
}

type Module struct {
    Handler   *handler.SomeHandler
    Registrar delivery.RouteRegistrar
}

func Wire(deps Dependencies) Module { ... }
```

### HTTP handlers are in the module, not a shared layer

Each module owns its HTTP handlers. Route registration uses the `RouteRegistrar` interface from `internal/inbound/`.

```go
// handler registers routes via the module's RegisterRoutes method
func (h *SomeHandler) RegisterRoutes(r chi.Router) { ... }
```

### GORM models are in `internal/database/model/`

Shared across modules. AutoMigrate runs at startup in `internal/outbound/database/`.

```go
type User struct {
    ID    uuid.UUID `gorm:"type:uuid;primaryKey"`
    Email string    `gorm:"size:128;not null;uniqueIndex"`
}
```

### Error handling

- Domain errors: `var ErrNotFound = errors.New(...)` in `domain/errors.go`
- HTTP errors: use `apperr.NotFound("code", "message")` from `pkg/apperr`
- Response: `response.WriteAppError(w, err)` handles both — sanitizes non-apperr errors

### Testing

- Mock interfaces with mockery: `task mocks`
- Handler tests: use mock service + `httptest`
- Service tests: use mock repository + mock dependencies
- Repository tests: use in-memory SQLite with build tag `//go:build integration`

## Adding a New Feature

1. Scaffold: `task new:module name=<feature>`
2. Add GORM model in `internal/database/model/`
3. Define domain types in `modules/<feature>/domain/`
4. Define `Repository` interface + `Service` in `app/`
5. Implement repository in `repository/`
6. Create handler in `handler/`
7. Write `wire.go` factory
8. Wire in `cmd/server/main.go`

## Shared Packages

| Package                        | Purpose                                                       |
| ------------------------------ | ------------------------------------------------------------- |
| `internal/inbound/`            | HTTP transport: middleware, response helpers, route registrar |
| `internal/outbound/authtoken/` | JWT signing/verification, bcrypt, token blacklist             |
| `internal/outbound/cache/`     | Redis wrapper + NullCache fallback                            |
| `internal/outbound/database/`  | GORM connection + AutoMigrate                                 |
| `internal/pkg/apperr/`         | Standardized error types with HTTP status mapping             |
| `internal/pkg/audit/`          | Structured audit logger                                       |
| `internal/pkg/validate/`       | Request validation (go-playground/validator)                  |
| `internal/pkg/doc/`            | OpenAPI spec builder + Scalar UI                              |
| `internal/pkg/eventbus/`       | In-memory event bus                                           |
| `internal/pkg/logger/`         | slog factory (text for dev, JSON for prod)                    |

## Task Commands

| Command                  | Purpose                                         |
| ------------------------ | ----------------------------------------------- |
| `task check`             | Format → lint → test → tidy (use before commit) |
| `task test`              | Run unit tests with race detector               |
| `task test-integration`  | Run integration tests                           |
| `task mocks`             | Regenerate mockery mocks                        |
| `task server`            | Run with hot-reload                             |
| `task new:module name=X` | Scaffold new module                             |
