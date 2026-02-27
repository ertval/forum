# Project Overview

This document is the technical onboarding guide for the Forum project. It explains how the system is structured, how requests flow, and which files matter most when you start contributing.

## 1) What this project is

- **Type**: Modular Monolith
- **Architecture**: Hexagonal Architecture (Ports & Adapters)
- **Language**: Go (module is currently `go 1.24`)
- **Database**: SQLite (CGO required)
- **Entry point**: `cmd/forum/main.go`

The project keeps business logic isolated from HTTP and database details. Every module follows `domain`, `ports`, `application`, and `adapters`.

## 2) High-level architecture

### Modules

Core modules live under `internal/modules/`:

- `auth` (sessions, login/register)
- `user`
- `post`
- `comment`
- `reaction`
- `moderation` *(optional/scaffolded)*
- `notification` *(optional/scaffolded)*

Shared infrastructure lives under `internal/platform/`:

- `config` (env parsing + validation)
- `database` (connection + migrations)
- `httpserver` (server, middleware, TLS, health handlers)
- `logger`, `errors`, `validator`, `upload`, `health`

### Request path (critical)

A typical API request path is:

1. `http.ServeMux` route match
2. global middleware chain (recover, log, security headers, CORS, rate limit)
3. module handler (`adapters/http_handler*.go`)
4. service interface (`ports/service.go`)
5. application service (`application/service.go`)
6. repository interface (`ports/repository.go`)
7. SQLite adapter (`adapters/sqlite_repository.go`)
8. SQLite DB

## 3) Dependency injection and app startup

The composition root is `cmd/forum/wire/`:

- `app.go`: app lifecycle and startup wiring
- `repositories.go`: construct SQLite repositories
- `services.go`: construct application services and middleware
- `handlers.go`: parse templates and construct module handlers

Startup sequence in `cmd/forum/main.go` + `cmd/forum/wire/app.go`:

1. Load and validate config
2. Initialize logger
3. Open database
4. Apply migrations
5. Build repositories
6. Build services
7. Build handlers
8. Register routes and middleware
9. Start HTTP server and graceful shutdown handling

## 4) Security model you must keep intact

### Public IDs vs internal IDs

- Internal DB IDs are integers.
- Public-facing IDs are UUIDs.
- Never expose sequential internal IDs in JSON, templates, or URLs.

### Auth/session fundamentals

- Session cookies are server-side tokens.
- Auth middleware validates token and writes user public UUID to request context.
- Login keeps a one-active-session-per-user policy.

### Platform security features

- TLS support with configurable cert/key paths
- Secure headers middleware
- Rate limiting middleware
- Input validation at domain/service boundaries

## 5) Data and migrations

- Migration files are in `migrations/` using `NNN_description.sql` naming.
- Migrations auto-apply on app startup via the platform migrator.
- Manual migration command: `make migrate` (internally calls `scripts/seed/run_migrations.sh`).

## 6) Seed and local HTTPS cert workflow

The seed workflow is in `scripts/seed/seed.sh` and now does this in order:

1. Ensure TLS cert/key exist (generates them if missing via `scripts/seed/generate_certs.sh`)
2. Run migrations (`scripts/seed/run_migrations.sh`)
3. Validate required tables
4. Load `scripts/seed/seed_data.sql`
5. Print dataset counts and test credentials

Use:

```bash
make seed
```

## 7) How to run the app

### Local

```bash
make go
```

or:

```bash
go run ./cmd/forum/main.go
```

Health checks:

- `GET /health`
- `GET /health-api`

### Docker

Generate local certs first (for HTTPS):

```bash
bash scripts/seed/generate_certs.sh ./certs
```

```bash
docker compose up --build -d
```

The compose setup maps:

- `8080:8080`
- `8443:8443`
- `./data:/app/data`
- `./certs:/app/certs:ro`
- `DATABASE_PATH=/app/data/forum.db`

## 8) Testing strategy

Primary commands:

```bash
make test       # full suite
make test-go    # go tests
make test-script # shell-based E2E scripts
```

Additional targeted checks:

```bash
go test ./tests/integration/... -v
go test ./internal/platform/upload/... -v
```

## 9) Key conventions for contributors

1. Keep module boundaries strict (domain/ports/application/adapters).
2. In `domain`, use stdlib-only imports.
3. API handlers return consistent JSON error shape: `{"error":"message"}`.
4. Keep adapters flat (`http_handler.go`, `http_handler_api.go`, `http_handler_page.go`, `sqlite_repository.go`).
5. Add new SQL files; do not modify already-applied migration files.
6. Preserve public UUID usage in all user-visible surfaces.

## 10) Critical files map

- App entry: `cmd/forum/main.go`
- Composition root: `cmd/forum/wire/*.go`
- Config: `internal/platform/config/config.go`
- Migrator: `internal/platform/database/migrator.go`
- HTTP middleware stack: `internal/platform/httpserver/middleware.go`
- Seed scripts: `scripts/seed/*.sh`
- Primary project docs: `README.md`, `docs/ARCHITECTURE.md`, `docs/IMPLEMENTATION_ROADMAP.md`

## 11) First-day checklist for new devs

1. Install prerequisites: Go, SQLite, Docker (optional).
2. Run `make seed`.
3. Run `make go` (or Docker compose).
4. Verify `GET /health-api` returns healthy status.
5. Read one full module end-to-end (recommended: `internal/modules/auth/`).
6. Before coding, identify the change location across domain, ports, application, adapters, and wiring.
