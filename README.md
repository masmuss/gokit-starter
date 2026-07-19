# Gokit Starter

Boilerplate Go API — beginner-friendly, production-ready. Module-based architecture, manual DI, GORM auto-migration.

## Tech Stack

| Layer      | Library                                                                                       |
| ---------- | --------------------------------------------------------------------------------------------- |
| Router     | [Chi](https://github.com/go-chi/chi)                                                          |
| ORM        | [GORM](https://gorm.io/) + PostgreSQL                                                         |
| Cache      | [Redis](https://redis.io/) via [go-redis](https://github.com/redis/go-redis)                  |
| Auth       | [golang-jwt](https://github.com/golang-jwt/jwt) + bcrypt                                      |
| Validation | [go-playground/validator](https://github.com/go-playground/validator)                         |
| Logging    | [slog](https://pkg.go.dev/log/slog)                                                           |
| Docs       | [swaggest/openapi-go](https://github.com/swaggest/openapi-go) + [Scalar](https://scalar.com/) |
| Testing    | [Testify](https://github.com/stretchr/testify) + [Mockery](https://github.com/vektra/mockery) |
| Tasks      | [Taskfile](https://taskfile.dev/) + [Lefthook](https://github.com/evilmartians/lefthook)      |

## Quick Start

Requirements: Go 1.24+, Docker, [Task](https://taskfile.dev/installation/).

```bash
git clone git@github.com:masmuss/gokit-starter.git
cd gokit-starter
cp .env.example .env
docker compose up -d
task server
```

API docs at **`http://localhost:8080/docs`**.

## Folder Structure

```
.
├── cmd/server/main.go         # Entry point — wires modules, starts server
├── docs/                      # Architecture documentation
├── internal/
│   ├── database/model/        # GORM models (shared)
│   ├── inbound/               # HTTP transport (middleware, response)
│   ├── modules/               # Feature modules
│   │   └── auth/
│   │       ├── domain/        # Models, errors, constants
│   │       ├── app/           # Use cases, repository interface
│   │       ├── handler/       # HTTP handlers + request DTOs
│   │       ├── repository/    # GORM implementation
│   │       └── wire.go        # Module factory
│   ├── outbound/              # Shared infrastructure
│   │   ├── authtoken/         # JWT, password, blacklist
│   │   ├── cache/             # Redis + NullCache fallback
│   │   └── database/          # GORM connection + AutoMigrate
│   └── pkg/                   # Pure utilities
├── test/mocks/                # Generated mocks
└── taskfile.yml               # Task runner
```

## Architecture

Module-based layered architecture. Each module is self-contained with its own `wire.go` factory.

```
main.go
  ├── openDatabase()  → AutoMigrate
  ├── openCache()     → Redis / NullCache
  ├── auth.Wire(deps) → Module{Handler, Registrar, DocRegistrar, Middleware}
  ├── buildRouter()
  └── runServer()     → graceful shutdown
```

See [docs/architecture.md](docs/architecture.md) for details.

## Adding a New Module

```bash
task new:module name=product
```

Then:

1. Define models in `modules/product/domain/`
2. Define `Repository` interface + `Service` in `app/`
3. Implement repository with GORM in `repository/`
4. Create HTTP handler in `handler/`
5. Write `wire.go` factory
6. Call `product.Wire(deps)` in `cmd/server/main.go`

## Tasks

| Command                    | Description                 |
| -------------------------- | --------------------------- |
| `task server`              | Run with hot-reload (air)   |
| `task test`                | Run all unit tests          |
| `task test-integration`    | Run integration tests       |
| `task mocks`               | Regenerate mockery mocks    |
| `task lint`                | Run golangci-lint           |
| `task check`               | Format + lint + test + tidy |
| `task build`               | Build binary                |
| `task new:module name=...` | Scaffold new module         |
| `task clean`               | Remove build artifacts      |

## Default Endpoints

| Method | Path             | Auth | Description              |
| ------ | ---------------- | ---- | ------------------------ |
| GET    | `/health`        | No   | Health check             |
| GET    | `/version`       | No   | App version              |
| POST   | `/auth/register` | No   | Register new account     |
| POST   | `/auth/login`    | No   | Login                    |
| POST   | `/auth/refresh`  | No   | Refresh access token     |
| GET    | `/auth/profile`  | Yes  | Get current user profile |
| POST   | `/auth/logout`   | Yes  | Revoke tokens            |
| PUT    | `/auth/password` | Yes  | Change password          |

---

_Designed for developers who want clean, readable, production-ready Go APIs._

## Release

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

GitHub Actions + [GoReleaser](https://goreleaser.com/) builds binaries, generates changelog, and creates a GitHub release. Binary artifacts: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`.
