# Repository Structure

```
.
├── cmd/server/main.go         # Composition root — wires modules and starts server
├── docs/                      # Architecture documentation
├── internal/
│   ├── database/model/        # GORM database models (shared across modules)
│   ├── inbound/               # Shared HTTP transport layer ("inbound adapters")
│   │   ├── handler/           # Health handler (cross-cutting)
│   │   ├── middleware/        # Auth middleware, role guard
│   │   ├── response/          # JSON response helpers
│   │   └── route_registrar.go # RouteRegistrar interface
│   ├── modules/               # Feature modules (one per business domain)
│   │   └── auth/              # Authentication module
│   │       ├── domain/        # Business models, errors, constants
│   │       ├── app/           # Use cases (Service), Repository interface
│   │       ├── handler/       # HTTP handlers, request/response DTOs
│   │       ├── repository/    # GORM implementation of app.Repository
│   │       └── wire.go        # Module factory (Dependencies → Module)
│   ├── outbound/              # Shared infrastructure ("outbound adapters")
│   │   ├── authtoken/         # JWT signing/verification, blacklist
│   │   ├── cache/             # Redis + NullCache
│   │   └── database/          # GORM connection + AutoMigrate
│   └── pkg/                   # Shared utilities (no business logic)
│       ├── apperr/            # Standardized error types
│       ├── audit/             # Structured audit logger
│       ├── doc/               # OpenAPI spec builder + Scalar UI
│       ├── eventbus/          # In-memory event bus
│       ├── logger/            # slog factory
│       ├── pagination/        # Pagination helpers
│       └── validate/          # Request validation (go-playground/validator)
├── scripts/generator/         # Module scaffolding CLI
└── test/mocks/                # Generated mocks (mockery)
```

## File Naming Conventions

| Type                  | Pattern               | Example                                |
| --------------------- | --------------------- | -------------------------------------- |
| Domain model          | `<entity>.go`         | `user.go`, `session.go`                |
| Domain errors         | `errors.go`           | `errors.go`                            |
| Service               | `service.go`          | `app/service.go`                       |
| Service test          | `service_test.go`     | `app/service_test.go`                  |
| HTTP handler          | `<entity>_handler.go` | `auth_handler.go`                      |
| Request/response DTOs | Same file as handler  | `RegisterRequest` in `auth_handler.go` |
| Repository            | `repository.go`       | `repository/repository.go`             |
| Module wiring         | `wire.go`             | `wire.go`                              |
| Doc registrar         | `<entity>_doc.go`     | `auth_doc.go`, `health_doc.go`         |

## Module Anatomy

Every module follows this structure:

```
modules/<name>/
├── domain/       # What the module is
│   ├── <entity>.go
│   └── errors.go
├── app/          # What the module does
│   ├── service.go
│   └── service_test.go
├── handler/      # HTTP interface
│   ├── <entity>_handler.go
│   ├── <entity>_handler_test.go
│   └── <entity>_doc.go (optional)
├── repository/   # Database interface
│   ├── repository.go
│   └── repository_test.go
└── wire.go       # Module entry point
```
