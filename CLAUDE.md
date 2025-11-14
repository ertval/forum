# Forum - AI Coding Agent Instructions

## Architecture Overview

This is a **Modular Monolith** Go 1.24+ application implementing a forum system. Each module follows **Hexagonal Architecture** (Ports & Adapters) with a strict 4-directory structure:

```
module/
├── domain/          # Entities, business rules, errors (no external dependencies)
├── ports/           # service.go (INPUT PORT), repository.go (OUTPUT PORT)
├── application/     # service.go - business logic orchestration
└── adapters/        # http_handler.go (INPUT), sqlite_repository.go (OUTPUT)
```

### Critical Pattern: Port/Adapter Annotations

Every file in `ports/` and `adapters/` MUST have a header comment declaring its type:
- `// INPUT PORT - Service Interface` - Defines use cases
- `// OUTPUT PORT - Repository Interface` - Data access contracts  
- `// INPUT ADAPTER - HTTP Handler` - HTTP request handlers
- `// OUTPUT ADAPTER - SQLite Repository` - Database implementations

**Example**: See `internal/modules/auth/ports/service.go` and `internal/modules/auth/adapters/http_handler.go`

### Dependency Rules (Strict)
```
domain      → Can import: stdlib ONLY
ports       → Can import: domain only
application → Can import: domain, ports
adapters    → Can import: domain, ports
```
**Cross-module communication**: Import `ports.XService` interface only, never internal implementations.

## Module Categories

**Core Modules** (required): auth, user, post, comment, reaction  
**Optional Modules** (marked with `[OPTIONAL FEATURE:]` in code): moderation, notification

When working on optional features, preserve all `[OPTIONAL FEATURE]` comments in code and documentation.

## Current Implementation Status (~10% Complete)

⚠️ **Initial scaffolding phase**:
- Module structure fully scaffolded with placeholder files
- Most implementations contain `// TODO:` comments marking unfinished work
- Database migrations defined in `migrations/` (auto-applied on startup via `database.Migrator`)
- When implementing features, **replace TODO placeholders** with actual logic
- See `docs/IMPLEMENTATION_ROADMAP.md` for detailed phase breakdown and priorities

## Critical Workflows

### Adding a New Feature (Follow This Exact Order)

```text
# 1. Domain layer
internal/modules/{module}/domain/{entity}.go      # Add entities
internal/modules/{module}/domain/errors.go        # Add domain errors

# 2. Ports (interfaces)
internal/modules/{module}/ports/service.go        # Define use cases (INPUT PORT)
internal/modules/{module}/ports/repository.go     # Define data access (OUTPUT PORT)

# 3. Application layer
internal/modules/{module}/application/service.go  # Implement business logic

# 4. Adapters
internal/modules/{module}/adapters/sqlite_repository.go  # Implement data access
internal/modules/{module}/adapters/http_handler.go       # Implement HTTP handlers

# 5. Wire in cmd/forum/wire/ (critical for DI)
cmd/forum/wire/repositories.go    # Add NewRepository() call
cmd/forum/wire/services.go        # Add NewService() call
cmd/forum/wire/handlers.go        # Add NewHTTPHandler() call
cmd/forum/wire/app.go             # Register routes: handler.RegisterRoutes(server.Router())

# 6. Database (if schema changes)
migrations/NNN_{module}_{description}.sql  # Add migration with Up/Down markers
```

### Dependency Injection Wiring Order (cmd/forum/wire/)

All DI happens in `cmd/forum/wire/` called from `main.go`:

1. **Config & Logger** → `config.Load()`, `logger.New()` (in main.go)
2. **Database** → Connect & run migrations (in `wire/app.go:initDatabase()`)
3. **Repositories** → All SQLite repos (in `wire/repositories.go`)
4. **Services** → Inject repos into services (in `wire/services.go`)
5. **Handlers** → Inject services into handlers (in `wire/handlers.go`)
6. **Routes** → Register via `handler.RegisterRoutes(router)` (in `wire/app.go`)

**Example from wire/app.go:**
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

**Never use dependency injection frameworks**. All wiring is explicit.

## Feature Implementation Requirements

**Authentication (auth module):**
- Email + username + password registration
- Session management with UUID tokens (via `gofrs/uuid` or `google/uuid`)
- Password encryption with bcrypt (cost factor: 10-12)
- Cookie-based sessions: HttpOnly, Secure, SameSite=Lax
- Session expiration (default: 24h, configurable)
- Only ONE active session per user (invalidate old sessions on new login)
- Validate email format, username uniqueness, password strength

**Posts (post module):**
- Title + content + optional image
- Associate 1+ categories per post
- Image validation: JPEG, PNG, GIF only, max 20MB
- Store images in `static/uploads/` with unique filenames
- Posts visible to all (guests + users)
- Only registered users can create/edit/delete own posts

**Comments (comment module):**
- Associate with posts, include user_id + timestamp
- Comments visible to all users
- Only registered users can create/edit/delete own comments
- Empty comments must be rejected

**Reactions (reaction module):**
- Like/dislike for posts AND comments
- Track user_id + target (post/comment) + type (like/dislike)
- User cannot like AND dislike same target (toggle behavior)
- Reaction counts visible to all users

**Filtering (post module):**
- By category (all users)
- By created posts (registered user's own posts)
- By liked posts (registered user's liked posts)

**User Roles (user module - OPTIONAL):**
- Guest: read-only access
- User: create, comment, react
- Moderator: delete content, create reports
- Administrator: promote/demote users, manage categories, review reports


## Database & Migrations

- **Database**: SQLite with `github.com/mattn/go-sqlite3` (requires CGO)
- **Migrations**: Sequential numbered SQL files in `migrations/`, named `NNN_module_description.sql`
- **Auto-apply**: Migrations run automatically on startup via `database.Migrator` in `main.go`
- **Pattern**: Each migration includes `-- +migrate Up` and `-- +migrate Down` markers
- **Foreign keys**: Always include `ON DELETE CASCADE` where appropriate
- **Indexes**: Create for frequently queried columns (tokens, user_id, expires_at)

**Example migration structure (`migrations/001_auth_create_sessions.sql`):**
```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_sessions_token ON sessions(token);

-- +migrate Down
DROP INDEX IF EXISTS idx_sessions_token;
DROP TABLE IF EXISTS sessions;
```

## Error Handling

Two-layer error system:
1. **Domain errors** (`domain/errors.go`): Simple `errors.New()` declarations per module
2. **Platform errors** (`internal/platform/errors/errors.go`): Structured errors with codes

```go
// Domain layer - simple errors
var ErrSessionExpired = errors.New("session has expired")

// Platform layer - structured errors with HTTP mapping
return errors.Wrap(err, errors.ErrCodeInternal, "failed to create session")
```

HTTP handlers should map domain errors to HTTP status codes using `errors.HTTPStatus()`.

## Configuration

All config in `internal/platform/config/config.go` uses structs, loaded from environment variables with defaults. No external config libraries (e.g., no viper). Configuration includes:
- Server (ports, timeouts, TLS)
- Database (path, connection pooling)
- Session (duration, cookie settings)
- Security (rate limiting, TLS certs)
- Upload (max size, allowed types)

## Logging

Structured logger in `internal/platform/logger/` with levels: Debug, Info, Warn, Error. Usage:
```go
lgr.Info("Starting service", logger.String("module", "auth"))
lgr.Error("Failed operation", logger.Error(err))
```

## HTTP Server & Middleware

Custom HTTP server wrapper (`internal/platform/httpserver/`) around standard library `http.ServeMux`. Global middleware registered before routes:
1. Recovery (panic handling)
2. Logger
3. CORS
4. RateLimit

Routes registered per-module via `handler.RegisterRoutes(router)` pattern.

**HTTP Status Codes:**
- 200 OK: Successful GET requests
- 201 Created: Successful POST (create user, post, comment)
- 204 No Content: Successful DELETE
- 400 Bad Request: Invalid input, empty required fields
- 401 Unauthorized: Missing/invalid session
- 403 Forbidden: Insufficient permissions
- 404 Not Found: Resource doesn't exist
- 409 Conflict: Duplicate email/username
- 413 Payload Too Large: Image > 20MB
- 429 Too Many Requests: Rate limit exceeded
- 500 Internal Server Error: Unexpected errors

**Error Handling:**
- Always return appropriate HTTP status codes
- Return JSON error responses with `{error: "message"}` format
- Log all 500 errors with full context
- Never expose internal errors to clients

## Testing Strategy

- **Unit tests**: `tests/unit/` - Test business logic in isolation
- **Integration tests**: `tests/integration/` - Test full request/response cycles
- **Repository tests**: Test against real SQLite database (or in-memory)

**TDD Workflow:**
```bash
# 1. Write failing test
go test ./internal/modules/auth/... -run TestRegister
# 2. Implement feature
# 3. Verify test passes
go test ./internal/modules/auth/... -run TestRegister
# 4. Refactor and ensure tests still pass
```

**Test Coverage Requirements:**
- Domain logic: Unit tests for all business rules
- Application services: Test all use cases with mocked repositories
- HTTP handlers: Integration tests simulating real requests
- Repositories: Test actual database operations
- Audit compliance: Integration tests covering every `.github/audit.md` scenario

Run tests: `go test ./...`  
Coverage report: `go test -cover ./...`

## Build & Run

**Local development:**
```bash
go run cmd/forum/main.go
```

**Docker build** (multi-stage):
```bash
docker build -t forum .
docker-compose up
```

**Important**: Build requires CGO for SQLite: `CGO_ENABLED=1`

## Key Files to Reference

- **Main entry point**: `cmd/forum/main.go` - Minimal lifecycle management
- **Wire package**: `cmd/forum/wire/` - Complete DI setup and component wiring
- **Architecture docs**: `docs/ARCHITECTURE.md` - Full design rationale with dependency rules
- **Implementation status**: `docs/IMPLEMENTATION_ROADMAP.md` - TODO tracking with phase breakdown
- **Module example**: `internal/modules/auth/` - Reference implementation with all 4 layers
- **Migration example**: `migrations/001_auth_create_sessions.sql` - Shows Up/Down markers
- **Audit spec**: `.github/audit.md` - Authoritative test scenarios (DO NOT modify)
- **Requirements**: `.github/requirements.md` - Core feature requirements
- **Additional features**: `.github/morefeats.md` - Optional feature specifications

## Go Module

- **Module path**: `forum` (import as `forum/internal/...`)
- **Go version**: 1.25
- **Dependencies**: Minimal - only uuid, sqlite3 driver, bcrypt

## Development Workflow

1. When adding new features, follow the module structure exactly
2. Start with domain entities and errors
3. Define ports (interfaces) before implementations
4. Implement application service, then adapters
5. Update migrations if database changes needed
6. Wire new components in `cmd/forum/wire/`
7. Update `IMPLEMENTATION_ROADMAP.md` with progress

**Example workflow for adding a new feature:**
```text
# 1. Add domain entity
# Edit: internal/modules/post/domain/post.go
# 2. Add domain errors
# Edit: internal/modules/post/domain/errors.go
# 3. Define service port (interface)
# Edit: internal/modules/post/ports/service.go
# 4. Define repository port (interface)
# Edit: internal/modules/post/ports/repository.go
# 5. Implement application service
# Edit: internal/modules/post/application/service.go
# 6. Implement SQLite repository
# Edit: internal/modules/post/adapters/sqlite_repository.go
# 7. Implement HTTP handler
# Edit: internal/modules/post/adapters/http_handler.go
# 8. Wire in wire package
# Edit: cmd/forum/wire/repos.go (add repository)
# Edit: cmd/forum/wire/services.go (add service)
# Edit: cmd/forum/wire/handlers.go (add handler)
# Edit: cmd/forum/wire/app.go (register routes)
```

## Common Pitfalls to Avoid

- ❌ Don't add subdirectories to `adapters/` - keep it flat
- ❌ Don't skip port/adapter type annotations in file headers
- ❌ Don't import domain from adapters - only from application layer
- ❌ Don't use external frameworks unnecessarily - prefer standard library
- ❌ Don't forget to mark optional features with `[OPTIONAL FEATURE]` comments
- ❌ Don't hard-code configuration - use the config system
- ❌ Don't modify `docs/requirements.md` or `docs/morefeats.md` - they're the authoritative specs

## Code Style

- Follow Go idioms: simplicity, readability, explicitness
- Apply SOLID + KISS principles to every component
- Use standard library over external dependencies where reasonable
- Prefer composition over inheritance
- Keep functions small and focused
- Use context.Context for cancellation and request-scoped values
- Use meaningful variable names, avoid abbreviations unless conventional

## Workflow Conventions

**Test-Driven Development (TDD):**

1. Write tests first, see them fail
2. Implement minimal code to pass tests
3. Refactor while keeping tests green
4. Commit after each complete feature

**Development Cycle:**

- Follow idiomatic Go patterns - consistency is critical
- Check `docs/IMPLEMENTATION_ROADMAP.md` for current progress
- Update roadmap checkboxes as you complete tasks
- Keep commits scoped to specific checklist items
- Update README.md if endpoints or Docker usage changes

**Audit Requirements:**

- **DO NOT modify** `.github/audit.md` - it's the authoritative test specification
- All audit questions must be covered by integration/e2e tests in `tests/integration/`
- Test every scenario described in `audit.md` including edge cases
- Project is only complete when all audit requirements pass

**Git Practices:**

- Commit frequently with descriptive messages
- Format: `[module] Brief description` (e.g., `[auth] Implement session validation`)
- One feature/fix per commit when possible
