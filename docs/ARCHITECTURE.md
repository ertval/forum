# Forum Application - Architecture

## Overview

A modular monolith web forum built with Go, following **Hexagonal Architecture** (Ports and Adapters). Clean boundaries, testable components, idiomatic Go.

Current status: this repository contains an initial scaffolding of the application where module structure and many placeholder files are present but significant business logic is still to be implemented. The project is roughly 10% complete. Many files contain TODOs and reference the implementation roadmap — see `docs/IMPLEMENTATION_ROADMAP.md` for priorities and milestones.

Note on migrations: the project includes SQL migrations in the `migrations/` directory. The intended startup flow (see wiring in `cmd/forum/wire/`) runs migrations automatically when the database connection is established. Migration files use the `-- +migrate Up`/`-- +migrate Down` markers and follow the repository's conventions.

## Core Principles

### Go Philosophy

- **Simplicity**: Straightforward solutions over clever tricks
- **Readability**: Code clarity over brevity
- **Explicitness**: No hidden magic or implicit behavior
- **Minimalism**: Minimal dependencies and abstractions
- **Composition**: Build complexity through composition

### SOLID + KISS

- Single Responsibility: One reason to change per component
- Interface Segregation: Small, focused interfaces
- Dependency Inversion: Depend on abstractions
- **Keep It Simple**: Simplest solution that works

---

## Architecture Pattern: Hexagonal (Ports & Adapters)

### The Hexagon

```
                    ┌─────────────────────────┐
                    │   HTTP Handlers (IN)    │
                    │   CLI Commands (IN)     │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │    INPUT PORTS          │
                    │  (Service Interfaces)   │
                    └───────────┬─────────────┘
                                │
            ┌───────────────────▼───────────────────┐
            │          DOMAIN CORE                  │
            │   • Entities                          │
            │   • Business Rules                    │
            │   • Domain Logic                      │
            │   • NO external dependencies          │
            └───────────────────┬───────────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │   OUTPUT PORTS          │
                    │ (Repository Interfaces) │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │  SQLite Repos (OUT)     │
                    │  External APIs (OUT)    │
                    └─────────────────────────┘
```

### What This Means

**Domain Core** = Business logic with zero external dependencies
**Ports** = Interfaces (contracts) that define how to interact with the core
**Adapters** = Concrete implementations that plug into ports

**Data Flow**: HTTP Request → Input Adapter (Handler) → Input Port (Service Interface) → Application Service → Domain Logic → Output Port (Repository Interface) → Output Adapter (SQLite) → Database

---

## Module Structure

Every module follows this **exact 4-directory layout**:

```
module/
├── domain/          # Pure business logic (no imports except stdlib)
│   ├── entity.go    # Domain entities with validation
│   └── errors.go    # Domain-specific errors
│
├── ports/           # Interface definitions
│   ├── service.go   # INPUT PORT - Use case definitions
│   └── repository.go # OUTPUT PORT - Data access contract
│
├── application/     # Orchestration layer
│   └── service.go   # Implements ports/service.go
│                    # Uses ports/repository.go
│
└── adapters/        # Technical implementations (flat, no subdirs)
    ├── http_handler.go       # INPUT - HTTP endpoints
    └── sqlite_repository.go  # OUTPUT - Database access
```

### Port/Adapter Markers

Every file in `ports/` and `adapters/` has a header comment:

```go
// INPUT PORT - Service Interface
// OUTPUT PORT - Repository Interface
// INPUT ADAPTER - HTTP Handler
// OUTPUT ADAPTER - SQLite Repository
```

This makes navigation and understanding instant.

---

## Modules

### Core Modules (Required)
1. **auth** - Registration, login, sessions
2. **user** - User profiles, roles
3. **post** - Create/read/update/delete posts, categories
4. **comment** - Comments on posts
5. **reaction** - Like/dislike posts and comments

### Optional Modules
6. **moderation** - Reports, content moderation, role management
7. **notification** - User notifications for interactions

---

## Dependency Rules

### Layer Dependencies (Strict)

```
┌──────────────┐
│  adapters    │  ─┐
└──────────────┘   │
                   ├─► Can import: domain, ports
┌──────────────┐   │
│ application  │  ─┘
└──────────────┘

┌──────────────┐
│    ports     │  ───► Can import: domain only
└──────────────┘

┌──────────────┐
│   domain     │  ───► Can import: NOTHING (only stdlib)
└──────────────┘
```

### Module Communication

Modules talk via **service interfaces** only:

```go
// ✅ GOOD: Import service interface
import "forum/internal/modules/user/ports"

type PostService struct {
    userService ports.UserService
}

// ❌ BAD: Import internal implementation
import "forum/internal/modules/user/adapters"
```

---

## Dependency Injection

All wiring happens in `cmd/forum/wire/` package, called from `main.go`:

```go
func main() {
    // 1. Platform services
    cfg := config.Load()
    lgr := logger.New(cfg.Log)
    
    // 2. All dependency injection happens here
    app, err := wire.InitializeApp(cfg, lgr)
    if err != nil {
        lgr.Fatal("Failed to initialize app", logger.Error(err))
    }
    defer app.Cleanup()
    
    // 3. Start server
    app.Start()
}
```

The wire package organizes DI into focused files:
- `app.go` - Main orchestration and app lifecycle
- `repos.go` - Repository initialization
- `services.go` - Service initialization  
- `handlers.go` - HTTP handler initialization

No magic frameworks. Everything explicit and organized.

---

## Platform Services

Shared infrastructure in `internal/platform/`:

- **config** - Environment variable loading
- **database** - SQLite connection, migrations, transactions
- **logger** - Structured logging (JSON output)
- **httpserver** - HTTP server wrapper with middleware
- **errors** - Common error types with HTTP status mapping
- **validator** - Input validation and sanitization

---

## Database Design

### Technology: SQLite

Simple, embedded, single-file database. Perfect for this use case.

### Migrations

Sequential numbered SQL files in `migrations/`:

```
001_auth_create_sessions.sql
002_user_create_users.sql
003_post_create_tables.sql
004_comment_create_comments.sql
005_reaction_create_reactions.sql
```

Each migration has `-- +migrate Up` and `-- +migrate Down` sections.

### Key Patterns

- Foreign keys with `ON DELETE CASCADE`
- Indexes on frequently queried columns (user_id, post_id, session_token)
- Timestamps (created_at, updated_at) on all entities
- Soft deletes where appropriate

---

## Error Handling

### Two-Layer System

**Domain Errors** (simple):
```go
// internal/modules/auth/domain/errors.go
var ErrSessionExpired = errors.New("session has expired")
```

**Platform Errors** (structured):
```go
// internal/platform/errors/errors.go
return errors.Wrap(err, errors.ErrCodeInternal, "failed to create session")
```

HTTP handlers map errors to status codes:
- Domain `ErrNotFound` → 404
- Domain `ErrUnauthorized` → 401
- Platform `ErrCodeValidation` → 400
- Platform `ErrCodeInternal` → 500

---

## Testing Strategy

### Unit Tests
- Domain logic (pure functions, business rules)
- Application services with mocked repositories
- Location: Each module's directory (`*_test.go`)

### Integration Tests
- HTTP handlers with real database
- End-to-end workflows
- Location: `tests/integration/`

### Repository Tests
- Test against in-memory SQLite (`:memory:`)
- Verify SQL queries work correctly

---

## Configuration

All config in `internal/platform/config/config.go`. Loaded from environment variables with sane defaults.

```env
# Server
PORT=8080
ENVIRONMENT=development

# Database
DB_PATH=./forum.db

# Session
SESSION_DURATION=24h

# Security
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

No config files, no external config libraries. Simple env vars.

---

## HTTP Server & Middleware

Standard library `http.ServeMux` wrapped for convenience.

### Middleware Chain (Global)

1. **Recovery** - Panic handling
2. **Logger** - Request/response logging
3. **CORS** - Cross-origin headers
4. **RateLimit** - Request throttling

### HTTP Status Code Conventions

- `200 OK` - Successful GET
- `201 Created` - Successful POST (create resource)
- `204 No Content` - Successful DELETE
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing/invalid session
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource doesn't exist
- `409 Conflict` - Duplicate email/username
- `413 Payload Too Large` - File too large
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Unexpected error

---

## Security Features

- **HTTPS/TLS** - TLS 1.2+, strong ciphers
- **Password Hashing** - bcrypt with cost factor 12
- **Session Management** - UUID tokens, HttpOnly cookies, server-side storage
- **Rate Limiting** - Per-IP and per-user limits
- **Input Validation** - Email format, password strength, data sanitization
- **CSRF Protection** - CSRF tokens on state-changing operations
- **Security Headers** - CSP, X-Frame-Options, HSTS

---

## Build & Deployment

### Local Development
```bash
go run cmd/forum/main.go
```

### Docker
```bash
docker-compose up --build
```

Multi-stage Dockerfile:
1. Build stage (compile binary)
2. Runtime stage (minimal alpine image)

**Important**: Requires `CGO_ENABLED=1` for SQLite.

---

## Key Design Decisions

### Why SQLite?
Simple, embedded, zero-config. Perfect for learning and small-medium forums. Can migrate to Postgres later without changing much code (thanks to repository pattern).

### Why No ORM?
ORMs hide SQL and add complexity. Raw SQL with `database/sql` is explicit, performant, and idiomatic Go.

### Why Flat Adapters?
Prevents over-engineering. One handler file, one repository file per module. Easy to find, easy to understand.

### Why Manual DI?
No magic. Every dependency is visible in `main.go`. Easy to trace, easy to debug.

### Why Hexagonal Architecture?
Testability + flexibility. Swap SQLite for Postgres? Change only the repository adapter. Add gRPC API? Add a new input adapter. Core logic untouched.

---

## References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Effective Go](https://go.dev/doc/effective_go)
