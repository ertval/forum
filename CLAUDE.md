# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Modular Monolith** forum application built with Go, following **Hexagonal Architecture** (Ports & Adapters). The project implements a clean, testable architecture with strict separation of concerns. Each business module follows an identical 4-layer structure: domain, ports, application, and adapters.

**Current Status**: ~10% complete - project scaffolding is done, most implementations contain TODO placeholders that need to be replaced with actual logic.

## Architecture

### Module Structure (Applied to ALL modules)

Every module follows this **exact 4-directory layout**:

```text
module/
├── domain/          # Pure business logic, NO external dependencies
│   ├── entity.go    # Domain entities with validation
│   └── errors.go    # Domain-specific errors
│
├── ports/           # Interface definitions (contracts)
│   ├── service.go   # INPUT PORT - Use case interfaces
│   └── repository.go # OUTPUT PORT - Data access contracts
│
├── application/     # Business logic orchestration
│   └── service.go   # Implements ports/service.go, uses ports/repository.go
│
└── adapters/        # Technical implementations (flat, no subdirs)
    ├── http_handler.go       # INPUT ADAPTER - HTTP endpoints
    └── sqlite_repository.go  # OUTPUT ADAPTER - Database access
```

**Critical**: Every file in `ports/` and `adapters/` MUST have a header comment declaring its type:

- `// INPUT PORT - Service Interface`
- `// OUTPUT PORT - Repository Interface`
- `// INPUT ADAPTER - HTTP Handler`
- `// OUTPUT ADAPTER - SQLite Repository`

### Modules

**Core Modules** (required):

- `auth` - Registration, login, sessions
- `user` - User profiles, roles
- `post` - Posts and categories
- `comment` - Comments on posts
- `reaction` - Likes/dislikes

**Optional Modules** (marked with `[OPTIONAL FEATURE]`):

- `moderation` - Reports, content moderation, role management
- `notification` - User notifications for interactions

### Dependency Rules

```text
adapters    ─┐
             ├─► Can import: domain, ports ONLY
application ─┘

ports       ──► Can import: domain ONLY

domain      ──► Can import: stdlib ONLY
```

Modules communicate via **service interfaces only** - never import internal implementations.

## Common Commands

### Run the Application

```bash
# Development
go run cmd/forum/main.go

# Local build (requires CGO for SQLite)
CGO_ENABLED=1 go build -o bin/forum cmd/forum/main.go
./bin/forum

# Production build
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o forum cmd/forum/main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test ./internal/modules/auth/... -run TestRegister

# Integration tests
go test ./tests/integration/...

# Unit tests
go test ./tests/unit/...
```

### Docker

```bash
# Build and run with Docker Compose
docker-compose up --build

# Build container
docker build -t forum .

# Run container
docker run -p 8080:8080 forum
```

## Development Workflow

### Dependency Injection Pattern

Dependencies are wired manually in `cmd/forum/wire/` package following this exact order:

1. Load config → Initialize logger (in `wire/app.go`)
2. Connect database → Run migrations (in `wire/app.go`)
3. Create repositories (output adapters) - all SQLite repos (in `wire/repos.go`)
4. Create services (inject repositories) - application layer (in `wire/services.go`)
5. Create HTTP handlers (inject services) - input adapters (in `wire/handlers.go`)
6. Register routes → Start server (in `wire/app.go`)

**Example pattern:**

```go
// 3. Repositories (Output Adapters)
sessionRepo := authAdapters.NewSQLiteSessionRepository(dbConn.DB())
// 4. Services (Application Layer)
authService := authApp.NewAuthService(sessionRepo, authUserRepo)
// 5. Handlers (Input Adapters)
authHandler := authAdapters.NewHTTPHandler(authService)
// 6. Register routes
authHandler.RegisterRoutes(server.Router())
```

### Database & Migrations

- **Database**: SQLite with `github.com/mattn/go-sqlite3` (requires CGO_ENABLED=1)
- **Migrations**: Numbered SQL files in `migrations/` directory
- **Pattern**: Each migration has `-- +migrate Up` and `-- +migrate Down` markers
- **Naming**: `NNN_module_description.sql` (e.g., `001_auth_create_sessions.sql`)

Migrations are applied automatically on startup via `database.Migrator`.

### Error Handling

**Two-layer system:**

1. **Domain errors** (simple, in `domain/errors.go`):

```go
var ErrSessionExpired = errors.New("session has expired")
```

2. **Platform errors** (structured, in `internal/platform/errors/errors.go`):

```go
return errors.Wrap(err, errors.ErrCodeInternal, "failed to create session")
```

HTTP handlers map domain errors to appropriate status codes using `errors.HTTPStatus()`.

### HTTP Status Codes

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

## Implementation Priority

The project follows a structured implementation roadmap defined in `docs/IMPLEMENTATION_ROADMAP.md`:

**Part 1 - MVP (9-12 days)**:

1. Platform basics (HTTP server, logger, errors, validation)
2. Authentication (register, login, logout with session cookies)
3. Posts (create, view, edit, delete)
4. Docker & testing

**Part 2 - Core Requirements (22-28 days total)**:
5. Categories
6. Comments
7. Reactions (like/dislike)
8. Filtering
9. Testing & error handling

**Part 3 - Bonus Features**:
10. Security (HTTPS, rate limiting)
11. Image upload
12. Moderation
13. Notifications
14. OAuth authentication

## Key Files to Reference

### Critical Entry Points

- **`cmd/forum/main.go`** - Application entry point, minimal lifecycle management
- **`cmd/forum/wire/`** - Complete dependency injection setup (app.go, repos.go, services.go, handlers.go)

### Documentation

- **`docs/ARCHITECTURE.md`** - Full architecture design with dependency rules
- **`docs/copilot-instructions.md`** - Detailed implementation guidance
- **`docs/IMPLEMENTATION_ROADMAP.md`** - Current phase breakdown and TODO tracking
- **`README.md`** - Project overview, API endpoints, configuration

### Module Examples

- **`internal/modules/auth/`** - Reference implementation showing all 4 layers properly structured
- **`migrations/001_auth_create_sessions.sql`** - Migration example with proper Up/Down markers

### Platform Services

- **`internal/platform/config/`** - Configuration loading from environment
- **`internal/platform/database/`** - SQLite connection and migrations
- **`internal/platform/logger/`** - Structured JSON logging
- **`internal/platform/httpserver/`** - HTTP server with middleware stack
- **`internal/platform/errors/`** - Error handling with HTTP mapping
- **`internal/platform/validator/`** - Input validation and sanitization

## Configuration

All configuration in `internal/platform/config/config.go` uses structs loaded from environment variables with sane defaults. Key settings:

```env
# Server
PORT=8080
ENVIRONMENT=development

# Database
DB_PATH=./forum.db

# Session
SESSION_SECRET=your-secret-key
SESSION_DURATION=24h

# Security
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# File Upload
MAX_UPLOAD_SIZE=20971520  # 20MB
UPLOAD_DIR=./static/uploads
```

## Testing Strategy

- **Unit tests**: `tests/unit/` - Business logic in isolation
- **Integration tests**: `tests/integration/` - Full request/response cycles
- **TDD workflow**: Write failing test → Implement → Verify → Refactor

Test coverage should include:

- Domain logic (unit tests)
- Application services (mocked repositories)
- HTTP handlers (integration tests)
- Repositories (real SQLite tests)

## Common Pitfalls to Avoid

- ❌ Don't add subdirectories to `adapters/` - keep it flat
- ❌ Don't skip port/adapter type annotations in file headers
- ❌ Don't import domain from adapters - only from application layer
- ❌ Don't use external DI frameworks - manual wiring only
- ❌ Don't forget to mark optional features with `[OPTIONAL FEATURE]` comments
- ❌ Don't hard-code configuration - use the config system
- ❌ Don't build without CGO_ENABLED=1 (required for SQLite)

## Go Module & Dependencies

- **Module path**: `forum`
- **Go version**: 1.25
- **Minimal dependencies**:
  - `github.com/gofrs/uuid/v5` - UUID generation
  - `github.com/mattn/go-sqlite3` - SQLite driver
  - `golang.org/x/crypto/bcrypt` - Password hashing

## Build Requirements

**Always use `CGO_ENABLED=1`** when building (SQLite requires CGO).

Production build:
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o forum cmd/forum/main.go
```
