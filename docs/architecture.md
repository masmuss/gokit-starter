# Architecture

Gokit follows a **module-based layered architecture** inspired by hexagonal (ports & adapters) principles, simplified for beginner accessibility.

## Terminology

| Term           | Description                                                                                                                                                   |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Module**     | Self-contained feature grouping (e.g., `auth`). Each module has its own domain, app, handler, repository, and a `wire.go` entry point.                        |
| **Domain**     | Core business models and errors. No external dependencies.                                                                                                    |
| **App**        | Use cases / application service. Orchestrates domain logic, depends on interfaces (ports), not implementations.                                               |
| **Handler**    | HTTP adapter. Translates HTTP requests into app calls and formats responses. Lives in the module, not the transport layer.                                    |
| **Inbound**    | Shared transport layer: route registration, middleware, response envelopes. Modules register routes via the `RouteRegistrar` interface.                       |
| **Outbound**   | Shared infrastructure layer: database, cache, auth tokens, JWT. Modules depend on interfaces defined here.                                                    |
| **Repository** | Database adapter. Module-specific GORM implementation that satisfies the module's `Repository` interface.                                                     |
| **Wire**       | Factory function per module. Takes `Dependencies` (what the module needs) and returns a `Module` struct (what it provides). Called from `cmd/server/main.go`. |

## Layers

```
┌──────────────────────────────────────────┐
│  cmd/server/main.go    (composition root) │
├──────────────────────────────────────────┤
│  internal/inbound/     (shared transport) │
│  internal/outbound/    (shared infra)     │
├──────────────────────────────────────────┤
│  internal/modules/*/   (feature modules)  │
│  ├── domain/           (models, errors)   │
│  ├── app/              (use cases)        │
│  ├── handler/          (HTTP adapter)     │
│  ├── repository/       (DB adapter)       │
│  └── wire.go           (module factory)   │
├──────────────────────────────────────────┤
│  internal/pkg/         (shared utilities) │
│  internal/database/    (GORM models)      │
└──────────────────────────────────────────┘
```

## Dependency Flow

```
main.go
  ├── wire modules (auth, ...)
  ├── build router (chi)
  └── run server (graceful shutdown)

module.Wire(Dependencies) → Module
  ├── repository (implements app.Repository)
  ├── service (app.Service)
  ├── handler (handler.AuthHandler)
  ├── registrar (RouteRegistrar)
  └── doc registrar (OperationRegistrar)
```

## Adding a New Module

1. Create folder: `internal/modules/<name>/`
2. Define domain models in `domain/`
3. Define `Repository` interface + `Service` in `app/`
4. Implement `Repository` with GORM in `repository/`
5. Create HTTP handler in `handler/`
6. Create `wire.go` with `Dependencies` struct + `Wire()` function
7. Call `module.Wire(deps)` in `cmd/server/main.go`

## Design Decisions

- **No DI framework**: Manual wiring in ~30 lines of main.go. Readable, debuggable.
- **No ORM codegen**: GORM with struct tags. AutoMigrate runs at startup.
- **Module as deployment unit**: Each module owns its handler, repository, and wiring.
- **Interfaces at app layer**: The `app` package defines what it needs (Repository, Hasher, TokenIssuer). Implementations live in `repository/` and `outbound/`.
