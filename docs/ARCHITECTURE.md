# Forum Application - Architecture

## Overview

A modular monolith web forum built with Go, following **Hexagonal Architecture** (Ports and Adapters). Clean boundaries, testable components, idiomatic Go.

**Status**: Production-ready MVP with auth, posts, categories, comments, filtering, and image upload. All tests pass.

**Stack**: Go 1.24+ | SQLite (CGO required) | Minimal deps (uuid, bcrypt, sqlite3)

---

## Architecture: Hexagonal (Ports & Adapters)

```
                    ┌─────────────────────────┐
                    │   HTTP Handlers (IN)    │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │    INPUT PORTS          │
                    │  (Service Interfaces)   │
                    └───────────┬─────────────┘
                                │
            ┌───────────────────▼───────────────────┐
            │          DOMAIN CORE                  │
            │   • Entities + Business Rules         │
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
                    └─────────────────────────┘
```

**Data Flow**: HTTP Request → Handler → Service Interface → Application Service → Domain → Repository Interface → SQLite

---

## Module Structure

Every module follows this **exact 4-directory layout**:

```
internal/modules/{module}/
├── domain/           # Pure business logic (stdlib only)
│   ├── entity.go     # Domain entities with Validate()
│   └── errors.go     # Domain-specific errors
│
├── ports/            # Interface definitions
│   ├── service.go    # INPUT PORT - Use case interface
│   └── repository.go # OUTPUT PORT - Data access contract
│
├── application/      # Business logic implementation
│   └── service.go    # Implements ports/service.go
│
└── adapters/         # Technical implementations (FLAT)
    ├── http_handler.go       # Base handler, ServiceContainer, routes
    ├── http_handler_api.go   # JSON API handlers (/api/...)
    ├── http_handler_page.go  # HTML page handlers
    └── sqlite_repository.go  # Database access
```

### Handler File Organization

1. **http_handler.go** - Base handler struct, `ServiceContainer` interface, `NewHTTPHandler()`, `RegisterRoutes()`
2. **http_handler_api.go** - JSON API handlers with `API` suffix (e.g., `CreatePostAPI`)
3. **http_handler_page.go** - HTML page handlers with `Page` suffix (e.g., `LoginPage`)

### File Header Markers

Every ports/adapters file has a header comment:
```go
// INPUT PORT - Service Interface
// OUTPUT PORT - Repository Interface
// INPUT ADAPTER - HTTP Handler
// OUTPUT ADAPTER - SQLite Repository
```

---

## API URL Pattern

All JSON API endpoints: `/api/{module}/{action}`

```
Auth:     POST /api/auth/register, /login, /logout | GET /api/auth/session
Posts:    GET/POST /api/posts | GET/PUT/DELETE /api/posts/{id}
Comments: GET/POST /api/comments/posts/{post_id} | GET/PUT/DELETE /api/comments/{id}
Reactions: POST/DELETE /api/reactions | GET /api/reactions/{targetType}/{targetId}
```

---

## Modules

### Complete
| Module | Description |
|--------|-------------|
| **auth** | Registration, login, sessions (one per user), logout |
| **user** | User profiles, cached stats (post/comment counts) |
| **post** | Full CRUD, categories, filtering, image upload |
| **comment** | Full CRUD with ownership validation |

### Scaffolded (Partial Implementation)
| Module | Description |
|--------|-------------|
| **reaction** | Like/dislike - routes defined, handlers return 501 |
| **moderation** | Reports, roles - minimal implementation |
| **notification** | User notifications - minimal implementation |

---

## Dependency Rules

```
adapters    ─┐
             ├─► Can import: domain, ports
application ─┘

ports       ───► Can import: domain only

domain      ───► Can import: NOTHING (stdlib only)
```

**Module Communication**: Via service interfaces only

```go
// ✅ Import service interface
import "forum/internal/modules/user/ports"

// ❌ Never import internal implementation
import "forum/internal/modules/user/adapters"
```

---

## Dependency Injection

All wiring in `cmd/forum/wire/`:

```
wire/
├── app.go          # Main app lifecycle
├── repos.go        # Repository initialization
├── services.go     # ServiceContainer with all services
└── handlers.go     # HTTP handler initialization
```

**ServiceContainer Pattern**:
- Single concrete `ServiceContainer` holds all services
- Each handler declares a local interface with only needed accessors
- Constructors: `NewHTTPHandler(services ServiceContainer, templates *template.Template)`

---

## Platform Services

Shared infrastructure in `internal/platform/`:

| Package | Purpose |
|---------|---------|
| config | Environment variable loading |
| database | SQLite connection, migrations |
| logger | Structured logging (JSON) |
| httpserver | HTTP server, middleware |
| errors | Common errors, HTTP status mapping |
| validator | Input validation |
| upload | Image upload handling |

---

## Database

**SQLite** with migrations in `migrations/`:
```
001_auth_create_sessions.sql
002_user_create_users.sql
003_post_create_tables.sql
004_comment_create_comments.sql
005_reaction_create_reactions.sql
```

Migrations auto-apply on startup via `database.Migrator`.

---

## ID Security

**INT internally, UUID publicly** - Never expose sequential IDs.

```go
// ✅ Context stores UUID
ctx.Value(UserIDKey) // user.PublicID (UUID)

// ✅ Templates use PublicID
<a href="/board?user={{.User.PublicID}}">

// ✅ JSON uses UUID
{"id": "uuid-here", "user_id": "uuid-here"}
```

---

## HTTP Status Codes

| Code | Usage |
|------|-------|
| 200 | Successful GET |
| 201 | Successful POST (create) |
| 204 | Successful DELETE |
| 400 | Invalid input |
| 401 | Missing/invalid session |
| 403 | Insufficient permissions |
| 404 | Resource not found |
| 409 | Duplicate email/username |
| 413 | File too large |
| 500 | Internal error |

---

## Testing

| Type | Location | Coverage |
|------|----------|----------|
| Unit | `internal/modules/*/` | Domain, services, repos |
| Integration | `tests/integration/` | Full request/response |
| E2E | `scripts/tests/` | API, audit, pages, images |

Run all: `make test`

---

## Commands

```bash
make go           # Run locally
make test         # Full test suite
make test-go      # Go tests only
make test-script  # E2E scripts only
```

---

## Key Design Decisions

| Choice | Reason |
|--------|--------|
| SQLite | Simple, embedded, zero-config |
| No ORM | Raw SQL is explicit and performant |
| Flat adapters | Prevents over-engineering |
| Manual DI | No magic, every dependency visible |
| Hexagonal | Testability + flexibility |
