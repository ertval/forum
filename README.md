# Forum (Go Modular Monolith)

A web forum built with Go using **Hexagonal Architecture (Ports & Adapters)** in a **modular monolith**.

- **Language**: Go 1.24+
- **Database**: SQLite (`mattn/go-sqlite3`, CGO enabled)
- **Entry point**: `cmd/forum/main.go`
- **Composition root**: `cmd/forum/wire/`

---

## What is implemented

### Core modules (complete)
- `auth` — register/login/logout, session validation, one active session per user
- `user` — user profile + user stats + account settings (username/email/password/avatar)
- `post` — CRUD, categories, filtering, image upload
- `comment` — CRUD + ownership validation
- `reaction` — like/dislike toggles on posts/comments
- `notification` — post owner notifications for comments/likes/dislikes + read/unread APIs

### Optional/scaffolded modules
- `moderation`
- OAuth-based authentication extensions

### Platform layer (complete)
- config, database connection, migrator, HTTP server + middleware, logger, validator, upload, health checks

For implementation status details: `docs/IMPLEMENTATION_ROADMAP.md`.

---

## Non-negotiable project rule

### Public IDs must be UUIDs
Internal DB IDs are integers, but **never expose sequential IDs** in URLs, templates, JSON, or request context.

- Use `PublicID` externally
- Keep integer IDs internal to persistence and domain relations

---

## Architecture in one view

Each module follows this strict structure:

```text
internal/modules/{module}/
├── domain/      # entities + business rules (stdlib only)
├── ports/       # input/output interfaces
├── application/ # use-case orchestration
└── adapters/    # HTTP handlers + SQLite repositories
```

Dependency direction:

- `adapters` -> can import `application/ports/domain`
- `application` -> can import `ports/domain`
- `ports` -> can import `domain`
- `domain` -> no project-layer imports

Detailed reference: `docs/ARCHITECTURE.md`.

---

## Quick start (local)

### Prerequisites
- Go 1.24+
- CGO toolchain (required by SQLite driver)
- `sqlite3` CLI (for seeding scripts)

### Run

```bash
make go
```

The app starts on:
- HTTP: `http://localhost:8080`
- HTTPS: `https://localhost:8443` (only if cert/key files exist)

### Seed test data (optional but useful)

```bash
make seed
```

This runs migrations first, then loads seed data.

---

## Docker

### Start

```bash
make up
```

### Stop

```bash
make down
```

Compose exposes:
- `8080:8080`
- `8443:8443`

It mounts:
- `./data -> /app/data`
- `./static/uploads -> /app/static/uploads`
- `./certs -> /app/certs:ro`

---

## Testing

```bash
make test         # full suite (go + integration + script tests)
make test-go      # go tests only
make test-script  # e2e script tests only
make test-coverage
```

Current script-audit expectation:
- `test_audit_advanced.sh` should pass
- `test_audit_moderation.sh` is pending (optional feature)
- `test_audit_authentication.sh` is pending (OAuth not implemented)

---

## Core commands

```bash
make go           # run with go run
make build        # build binary
make migrate      # run SQL migrations
make seed         # migrate + seed DB
make help         # full target list
```

---

## Key paths

- App entry: `cmd/forum/main.go`
- DI wiring: `cmd/forum/wire/`
- Modules: `internal/modules/`
- Shared platform: `internal/platform/`
- SQL migrations: `migrations/`
- Templates: `templates/`
- Static files/uploads: `static/`
- Tests: `tests/` and `scripts/tests/`

---

## API pattern

All JSON endpoints are under `/api`.

Examples:
- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/posts`
- `POST /api/comments/posts/{post_id}`
- `POST /api/reactions`

---

## Adding a new feature/module

1. Add `domain` entities + validation + errors
2. Define interfaces in `ports/service.go` and `ports/repository.go`
3. Implement use cases in `application/service.go`
4. Add adapters (`http_handler*.go`, `sqlite_repository.go`)
5. Register in `cmd/forum/wire/{repositories,services,handlers}.go` and routes in `app.go`
6. Add migration: `migrations/NNN_{module}_{description}.sql`

Use `internal/modules/auth/` as the reference implementation.
