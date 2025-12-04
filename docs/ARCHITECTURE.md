# Forum Application - Architecture

## Overview

A modular monolith web forum built with Go, following **Hexagonal Architecture** (Ports and Adapters). Clean boundaries, testable components, idiomatic Go.

**Current status**: Active development — core features implemented; additional modules scaffolded. The repository contains a complete `auth` module (registration, login, sessions) and a mature `post`/`category` implementation with unit and integration tests (see `tests/unit` and `tests/integration`). Several optional modules (`comment`, `reaction`, `moderation`, `notification`, and parts of `user`) are scaffolded and include domain/ports/application/adapters directories, but many contain `// TODO:` markers in `application` packages indicating remaining business logic to implement. See `docs/IMPLEMENTATION_ROADMAP.md` for priorities and remaining work.

**Migrations**: The project includes SQL migrations in the `migrations/` directory. Migrations run automatically on startup via `database.Migrator` in `cmd/forum/wire/app.go`. Migration files use `-- +migrate Up`/`-- +migrate Down` markers for forward/backward migrations.

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
│   └── filter_service.go # Specialized filtering logic (post module)
│                    # Uses ports/repository.go
│
└── adapters/        # Technical implementations (flat, no subdirs)
    ├── http_handler.go       # INPUT - Base handler struct & route registration
    ├── http_handler_api.go   # INPUT - JSON API endpoints (/api/...)
    ├── http_handler_page.go  # INPUT - HTML page endpoints (/, /posts/...)
    └── sqlite_repository.go  # OUTPUT - Database access
```

### Handler File Organization

Each module's HTTP handlers are split into 3 files:

1. **http_handler.go** - Base handler with:
   - `HTTPHandler` struct definition
   - `ServiceContainer` interface (declares needed dependencies)
   - `NewHTTPHandler()` constructor
   - `RegisterRoutes()` dispatcher that calls API and page route registration
   - Helper methods (e.g., `GetCurrentUser`, `parseJSON`)

2. **http_handler_api.go** - JSON API handlers:
   - `RegisterAPIRoutes()` - Registers all `/api/...` routes
   - Handler methods with suffix `API` (e.g., `RegisterAPI`, `CreatePostAPI`)
   - Returns JSON responses, uses proper HTTP status codes
   - Routes follow pattern: `/api/{module}/{action}`

3. **http_handler_page.go** - HTML page handlers:
   - `RegisterPageRoutes()` - Registers page routes (e.g., `/`, `/login`, `/posts/{id}`)
   - Handler methods with suffix `Page` (e.g., `LoginPage`, `CreatePostPage`)
   - Returns rendered HTML templates

### API URL Pattern

All JSON API endpoints follow the pattern: `/api/{module}/{action}`

```
Authentication:
  POST /api/auth/register    - Register new user
  POST /api/auth/login       - Login user
  POST /api/auth/logout      - Logout user
  GET  /api/auth/session     - Get current session

Posts:
  GET    /api/posts          - List posts (with filtering)
  POST   /api/posts          - Create post
  GET    /api/posts/{id}     - Get post
  PUT    /api/posts/{id}     - Update post
  DELETE /api/posts/{id}     - Delete post

Comments:
  GET    /api/comments/posts/{post_id}  - List comments for post
  POST   /api/comments/posts/{post_id}  - Create comment
  GET    /api/comments/{id}             - Get comment
  PUT    /api/comments/{id}             - Update comment
  DELETE /api/comments/{id}             - Delete comment

Reactions:
  POST   /api/reactions                             - Add reaction
  DELETE /api/reactions                             - Remove reaction
  GET    /api/reactions/{targetType}/{targetId}     - Get reactions
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
3. **post** - Create/read/update/delete posts, categories, filtering
   - **FilterService**: Dedicated application service for post filtering logic
   - Supports filtering by category, user, liked posts, and date range
   - Date filters: Today, This Week, This Month, All Time
   - Uses `FilterParams` for query parameter parsing and `PostFilter` for repository queries
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

All wiring happens in the `cmd/forum/wire/` package, called from `main.go`:

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
- `repositories.go` - Repository initialization
- `services.go` - Service initialization  
- `handlers.go` - HTTP handler initialization

No magic frameworks. Everything explicit and organized.

### Unified Dependency Injection Pattern (Implemented)

This project now uses a unified DI pattern for all HTTP handlers. The implementation lives in `cmd/forum/wire/` and is applied across module adapters.

Key points:
- A single concrete `ServiceContainer` (defined in `cmd/forum/wire/services.go`) holds all application service implementations as private fields.
- `ServiceContainer` exposes accessor methods for each service (for example: `Auth()`, `User()`, `Post()`, `Category()`).
- Every HTTP handler uses the same constructor signature:
  ```go
  func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler
  ```
- Each handler declares a small local `ServiceContainer` interface that lists only the accessor methods it needs (for example:
  ```go
  type ServiceContainer interface {
      Post() postPorts.PostService
      Auth() authPorts.AuthService
  }
  ```
  The concrete `wire.ServiceContainer` satisfies these local interfaces.
- Handlers call accessor methods to obtain the service interfaces they depend on and remain decoupled from concrete implementations.

Why this was added:
- **Consistency**: all handlers use identical constructor signatures, simplifying wiring and tests.
- **Explicit dependencies**: handlers declare precisely what they require via a minimal interface.
- **Testability**: unit tests can provide small mocks implementing the handler-local `ServiceContainer` without constructing the full application container.
- **Avoid circular imports**: handlers depend only on service port interfaces and the small accessor interface, not concrete implementations or the `wire` package internals.

Where to look:
- Implementation & examples: `docs/UNIFIED_DI_PATTERN.md`
- Wire code: `cmd/forum/wire/services.go`, `cmd/forum/wire/handlers.go`
- Handler examples: `internal/modules/*/adapters/http_handler.go` (each handler shows a local `ServiceContainer` interface)

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
