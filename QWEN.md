# Forum - AI Coding Agent Instructions

## Quick Reference

- **Architecture**: Modular Monolith with Hexagonal Architecture (Ports & Adapters)
- **Language**: Go 1.24+ | **Database**: SQLite (CGO required) | **Dependencies**: Minimal (uuid, bcrypt, sqlite3)
- **Entry Point**: `cmd/forum/main.go` → `cmd/forum/wire/` for DI wiring
- **Reference Implementation**: `internal/modules/auth/` (fully implementation with tests)

## 🔒 ID Security (CRITICAL)

**INT internally, UUID publicly** - Never expose sequential IDs in URLs, templates, or JSON.

```go
// ✅ Context stores UUID    | ❌ Never store INT
ctx.Value(UserIDKey) // user.PublicID (UUID)

// ✅ Templates use PublicID | ❌ Never .ID for INT fields  
<a href="/board?user={{.User.PublicID}}">

// ✅ JSON uses "id" for UUID | ❌ Never expose internal IDs
{"id": "uuid-here", "user_id": "uuid-here"}
```

## Module Structure (Strict 4-Directory Layout)

```text
internal/modules/{module}/
├── domain/      # Entities + errors (stdlib ONLY, no project imports)
├── ports/       # Interfaces: service.go (INPUT), repository.go (OUTPUT)
├── application/ # Business logic (imports: domain, ports)
└── adapters/    # FLAT dir: http_handler*.go, sqlite_repository.go (imports: domain, ports)
    ├── http_handler.go       # Base handler struct, ServiceContainer, RegisterRoutes
    ├── http_handler_api.go   # JSON API handlers (/api/...), suffix: *API
    ├── http_handler_page.go  # HTML page handlers, suffix: *Page
    └── sqlite_repository.go  # Database access
```

**File headers required**: `// INPUT PORT - Service Interface`, `// OUTPUT ADAPTER - SQLite Repository`, etc.

## API URL Pattern

All JSON API endpoints use the `/api` prefix with resource-style routes.

```text
POST /api/auth/register   POST /api/auth/login     GET /api/auth/session
GET  /api/posts           POST /api/posts          GET /api/posts/{id}
POST /api/comments/posts/{post_id}                 GET /api/comments/{id}
POST /api/reactions                                DELETE /api/reactions
GET  /api/activity                                 GET /api/notifications
```

## Adding Features (6-Step Workflow)

1. **Domain**: Entity with `Validate()`, errors in `domain/errors.go`
2. **Ports**: Define interfaces in `ports/service.go` and `ports/repository.go`
3. **Application**: Implement service in `application/service.go`
4. **Adapters**: HTTP handler + SQLite repository (keep adapters/ flat)
5. **Wire**: Register in `cmd/forum/wire/{repos,services,handlers}.go`, add routes in `cmd/forum/wire/app.go`
6. **Migration**: Add SQL in `migrations/NNN_{module}_{desc}.sql` (auto-applies on startup)

## Dependency Injection Pattern

```go
// cmd/forum/wire/services.go - Central ServiceContainer
type ServiceContainer struct { auth authPorts.AuthService; ... }
func (sc *ServiceContainer) Auth() authPorts.AuthService { return sc.auth }

// Each handler: local interface declaring ONLY what it needs
type ServiceContainer interface { Auth() authPorts.AuthService }
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler
```

## Handler Conventions

- **API handlers** (JSON): suffix `API` → `RegisterAPI`, `CreatePostAPI`
- **Page handlers** (HTML): suffix `Page` → `LoginPage`, `PostDetailPage`
- **Error responses**: Always `{"error": "message"}` with appropriate HTTP status

## Commands

```bash
make go           # Run locally (go run)
make test         # Full test suite (Go tests + scripts/tests/run_all_tests.sh)
make test-go      # Go tests only
make test-script  # E2E bash tests only
go test ./tests/integration/... -v  # Integration tests
```

## Key Files / Guides

| Purpose | Location |
|---------|----------|
| DI Wiring Config | `cmd/forum/wire/README.md` |
| Bounded Contexts | `internal/modules/README.md` |
| Auth Reference | `internal/modules/auth/` (complete implementation) |
| Migrations | `migrations/*.sql` (auto-applied) |
| System Guides | `docs/guides/ONBOARDING_GUIDE.md` |
| Progress | `docs/IMPLEMENTATION_ROADMAP.md` |

## Status & TODOs

**✅ Complete**: auth, user settings/avatar, post, category, comment, reaction, notification, activity page, platform layer  
**⚠️ Pending Optional Features**: moderation, OAuth authentication extensions  
**Script-audit expectation**: advanced passes; moderation and authentication (OAuth) remain pending.
