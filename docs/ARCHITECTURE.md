# Forum Application - Architecture

## Overview

A modular monolith web forum built with Go, following **Hexagonal Architecture** (Ports and Adapters). Clean boundaries, testable components, idiomatic Go.

**Status**: Production-ready core forum with auth, user, post, comment, reaction, notification, activity, filtering, and image upload. Moderation remains partial, and OAuth social login is deferred.

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

**Current exception (documented and intentional):** API-only modules may omit `http_handler_page.go`. In this repository, `moderation`, `notification`, and `reaction` are API-only handlers at this time.

### Handler File Organization

1. **http_handler.go** - Base handler struct, `ServiceContainer` interface, `NewHTTPHandler()`, `RegisterRoutes()`
2. **http_handler_api.go** - JSON API handlers with `API` suffix (e.g., `CreatePostAPI`)
3. **http_handler_page.go** - HTML page handlers with `Page` suffix (e.g., `LoginPage`)
4. **API-only exception** - Modules without HTML pages keep `http_handler.go` + `http_handler_api.go` only

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

All JSON API endpoints use the `/api` prefix with resource-style routes.

```
Auth:     POST /api/auth/register, /login, /logout | GET /api/auth/session
Posts:    GET/POST /api/posts | GET/PUT/DELETE /api/posts/{id}
Comments: GET/POST /api/comments/posts/{post_id} | GET/PUT/DELETE /api/comments/{id}
Reactions: POST/DELETE /api/reactions | GET /api/reactions/{targetType}/{targetId}
Notifications: GET /api/notifications | PUT /api/notifications/{id}/read | PUT /api/notifications/read-all
Activity: GET /api/activity
Moderation: POST/GET /api/moderation/reports | PUT /api/moderation/reports/{report_id}
```

---

## Modules

### Complete
| Module | Description |
|--------|-------------|
| **auth** | Registration, login, sessions (one per user), logout |
| **user** | User profiles, cached stats (post/comment counts) |
| **post** | Full CRUD, categories, filtering, image upload |
| **comment** | Full CRUD with ownership validation, my comments page |
| **reaction** | Like/dislike toggles for posts/comments with counts |
| **notification** | Notification creation/list/read flows and APIs |

### Partial / Deferred
| Module | Description |
|--------|-------------|
| **moderation** | Core report routes exist; full lifecycle/response workflow still partial |
| **oauth (social auth)** | Deferred (GitHub/Google integration not implemented in current release) |

---

## Domain-Driven Design (DDD) Module Division

This project follows **DDD tactical patterns** within a **modular monolith** architecture. Each module encapsulates a bounded context with clear responsibilities.

### Core Concepts

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         MODULAR MONOLITH                                │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │    AUTH     │  │    USER     │  │    POST     │  │   COMMENT   │    │
│  │  Bounded    │  │  Bounded    │  │  Bounded    │  │  Bounded    │    │
│  │  Context    │  │  Context    │  │  Context    │  │  Context    │    │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘    │
│         │                │                │                │           │
│         └────────────────┼────────────────┼────────────────┘           │
│                          │                │                             │
│                    Service Interface Communication                      │
└─────────────────────────────────────────────────────────────────────────┘
```

### Bounded Contexts & Responsibilities

| Bounded Context | Aggregate Root | Key Entities | Responsibilities |
|-----------------|----------------|--------------|------------------|
| **Auth** | Session | Session, Credentials | Authentication, session lifecycle, credential validation |
| **User** | User | User, UserStats | User identity, profile, statistics aggregation |
| **Post** | Post | Post, Category | Content creation, categorization, filtering, image handling |
| **Comment** | Comment | Comment | Threaded discussions, comment CRUD, author context |
| **Reaction** | Reaction | Reaction | Like/dislike tracking, reaction counts |
| **Moderation** | Report | Report, Action | Content reporting, moderation actions |
| **Notification** | Notification | Notification | User notifications, read/unread state |

### DDD Layers Within Each Module

```
┌─────────────────────────────────────────────────────────────────────┐
│                        MODULE STRUCTURE                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ DOMAIN LAYER (domain/)                                       │   │
│   │ ┌─────────────┐  ┌─────────────────┐  ┌──────────────────┐  │   │
│   │ │  Entities   │  │ Value Objects   │  │ Domain Errors    │  │   │
│   │ │ • User      │  │ • Email         │  │ • ErrNotFound    │  │   │
│   │ │ • Post      │  │ • Username      │  │ • ErrValidation  │  │   │
│   │ │ • Comment   │  │ • Role          │  │ • ErrUnauthorized│  │   │
│   │ └─────────────┘  └─────────────────┘  └──────────────────┘  │   │
│   │ • Pure Go - NO external dependencies                         │   │
│   │ • Business rules enforced via Validate() methods            │   │
│   │ • Invariants protected within entity boundaries             │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ PORTS LAYER (ports/)                                         │   │
│   │ ┌────────────────────────┐  ┌────────────────────────────┐  │   │
│   │ │ INPUT PORTS (service.go)│  │OUTPUT PORTS (repository.go)│  │   │
│   │ │ • Service Interface    │  │ • Repository Interface     │  │   │
│   │ │ • Use case definitions │  │ • Data access contracts    │  │   │
│   │ └────────────────────────┘  └────────────────────────────┘  │   │
│   │ • Defines contracts, not implementations                     │   │
│   │ • Can only import from domain layer                          │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ APPLICATION LAYER (application/)                             │   │
│   │ ┌─────────────────────────────────────────────────────────┐ │   │
│   │ │ Service Implementation (service.go)                      │ │   │
│   │ │ • Implements INPUT PORT interface                        │ │   │
│   │ │ • Orchestrates domain logic                              │ │   │
│   │ │ • Coordinates repositories (OUTPUT PORTS)                │ │   │
│   │ └─────────────────────────────────────────────────────────┘ │   │
│   │ • Business logic orchestration                               │   │
│   │ • Transaction coordination                                   │   │
│   │ • Cross-entity workflows                                     │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              ▼                                       │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ ADAPTERS LAYER (adapters/)                                   │   │
│   │ ┌────────────────────────┐  ┌────────────────────────────┐  │   │
│   │ │ INPUT ADAPTERS         │  │ OUTPUT ADAPTERS            │  │   │
│   │ │ • http_handler.go      │  │ • sqlite_repository.go     │  │   │
│   │ │ • http_handler_api.go  │  │ • (future: cache, etc.)    │  │   │
│   │ │ • http_handler_page.go │  │                            │  │   │
│   │ └────────────────────────┘  └────────────────────────────┘  │   │
│   │ • Technical implementations of ports                         │   │
│   │ • HTTP request/response translation                          │   │
│   │ • Database query execution                                   │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Module Communication Rules

**Strict Boundaries**: Modules communicate ONLY through service interfaces, never by importing internal implementations.

```go
// ✅ CORRECT: Use service interface from other module
type ServiceContainer interface {
    Auth() authPorts.AuthService
    User() userPorts.UserService
}

// ✅ CORRECT: Import port interface
import authPorts "forum/internal/modules/auth/ports"

// ❌ WRONG: Never import internal implementations
import "forum/internal/modules/auth/application"  // FORBIDDEN
import "forum/internal/modules/auth/adapters"     // FORBIDDEN
```

### Entity Design Principles

Each entity follows these DDD principles:

1. **Identity**: Each entity has a public UUID (`PublicID`) for external use and internal ID for database references
2. **Validation**: Business rules enforced via `Validate()` method
3. **Encapsulation**: State changes through defined methods, not direct field access
4. **Invariants**: Entity ensures its own consistency

```go
// Example: Post entity with DDD principles
type Post struct {
    ID         int       // Internal identity
    PublicID   string    // External identity (UUID)
    AuthorID   int       // Reference to User aggregate
    Title      string    // Value object candidate
    Content    string    
    ImagePath  string    // Optional
    Categories []Category // Related aggregate
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

func (p *Post) Validate() error {
    // Enforces business invariants
    if strings.TrimSpace(p.Title) == "" {
        return ErrInvalidTitle
    }
    if len(p.Categories) == 0 {
        return ErrCategoryRequired
    }
    return nil
}
```

### Aggregate Boundaries

Each module owns its aggregate root and is the sole authority over its data:

| Module | Aggregate Root | Owned Data | References To |
|--------|----------------|------------|---------------|
| **auth** | Session | sessions table | users (FK) |
| **user** | User | users table | - |
| **post** | Post | posts, post_categories, categories | users (author_id) |
| **comment** | Comment | comments table | posts (post_id), users (author_id) |
| **reaction** | Reaction | reactions table | posts/comments (target_id), users (user_id) |

### Cross-Cutting Concerns

Infrastructure concerns that span all modules are handled in `internal/platform/`:

| Platform Service | DDD Concept | Purpose |
|------------------|-------------|---------|
| **database** | Infrastructure | Persistence mechanism (SQLite) |
| **logger** | Infrastructure | Structured logging across all layers |
| **httpserver** | Infrastructure | HTTP transport, middleware |
| **validator** | Domain Support | Input validation utilities |
| **upload** | Infrastructure | File storage mechanism |
| **cache** | Infrastructure | Performance optimization |
| **errors** | Shared Kernel | Common error types |

---

## Dependency Rules

Current repository convention:

```
domain       ───► stdlib only

ports        ───► domain + stdlib

application  ───► own domain/ports + other modules' ports + selected internal/platform utilities

adapters     ───► ports + domain + internal/platform
```

**Module Communication Rule**: Cross-module access goes through exported `ports` interfaces; do not import another module's `application` or `adapters` packages.

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
├── repositories.go # Repository initialization
├── services.go     # ServiceContainer with all services
└── handlers.go     # HTTP handler initialization
```

**ServiceContainer Pattern**:
- Single concrete `ServiceContainer` holds all services
- Each handler declares a local interface with only needed accessors
- Constructors: `NewHTTPHandler(services ServiceContainer, templates *platformTemplates.Registry)`

---

## Platform Services

Shared infrastructure in `internal/platform/`:

| Package | Purpose |
|---------|---------|
| config | Environment variable loading |
| database | SQLite connection, migrations |
| logger | Structured logging (JSON) |
| httpserver | HTTP server, middleware, TLS, security headers, health handler |
| health | Health checker (database + route verification) |
| errors | Common errors, HTTP status mapping |
| validator | Input validation |
| upload | Image upload handling |

**Health handler** follows the same `RegisterRoutes(router)` pattern as module handlers, keeping route registration consistent across the codebase.

---

## Database

**SQLite** with migrations in `migrations/`:
```
001_auth.sql
002_user.sql
003_post.sql
004_comment.sql
005_reaction.sql
006_moderation.sql
007_notification.sql
```

**Running Migrations:**
- Automatic: Migrations auto-apply on application startup via `database.Migrator`
- Manual: Run `make migrate` or `bash scripts/seed/run_migrations.sh`
- The migrator creates a `schema_migrations` table to track applied migrations
- Already-applied migrations are automatically skipped
- See `migrations/README.md` for migration authoring guidelines and conventions

---

## Security

### TLS Configuration
- TLS 1.2 minimum version
- Strong cipher suites (AEAD only)
- Certificate configuration via environment variables
- Self-signed certificate generation script: `scripts/seed/generate_certs.sh`

### Production TLS Certificates

**For production deployments, use proper TLS certificates from a trusted Certificate Authority (CA), not self-signed certificates.**

#### Option 1: Let's Encrypt (Free, Automated)

**Recommended for most deployments.** Let's Encrypt provides free, automated TLS certificates with 90-day validity.

**Using Certbot:**
```bash
# Install certbot
sudo apt-get install certbot  # Debian/Ubuntu
sudo yum install certbot      # RHEL/CentOS

# Obtain certificate (standalone mode - requires port 80/443 temporarily)
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# Certificates will be saved to:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem  (certificate)
# /etc/letsencrypt/live/yourdomain.com/privkey.pem    (private key)

# Update .env with certificate paths
TLS_CERT_FILE=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
TLS_KEY_FILE=/etc/letsencrypt/live/yourdomain.com/privkey.pem

# Set up auto-renewal (certbot installs a systemd timer automatically)
sudo certbot renew --dry-run  # Test renewal
```

**Using ACME clients (alternative):**
- [acme.sh](https://github.com/acmesh-official/acme.sh) - Lightweight, shell-based
- [lego](https://github.com/go-acme/lego) - Go-based ACME client

#### Option 2: Commercial Certificate Authority

Purchase certificates from trusted CAs like DigiCert, Sectigo, or GlobalSign for extended validation or wildcard certificates.

**Steps:**
1. Generate a Certificate Signing Request (CSR):
   ```bash
   openssl req -new -newkey rsa:2048 -nodes \
     -keyout yourdomain.key \
     -out yourdomain.csr \
     -subj "/C=US/ST=State/L=City/O=Organization/CN=yourdomain.com"
   ```

2. Submit CSR to your chosen CA and complete their validation process

3. Download the issued certificate and intermediate certificates

4. Configure the forum:
   ```env
   TLS_CERT_FILE=/path/to/yourdomain.crt
   TLS_KEY_FILE=/path/to/yourdomain.key
   ```

#### Option 3: Reverse Proxy with TLS Termination

Use a reverse proxy (nginx, Caddy, Traefik) to handle TLS, allowing the forum to run on HTTP internally.

**Example with Caddy (automatic HTTPS):**
```caddy
yourdomain.com {
    reverse_proxy localhost:8080
}
```

**Example with nginx:**
```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

#### Certificate Permissions

Ensure the forum process has read access to certificate files:
```bash
# Option 1: Copy certificates to forum directory with proper permissions
sudo cp /etc/letsencrypt/live/yourdomain.com/*.pem /path/to/forum/certs/
sudo chown forum-user:forum-user /path/to/forum/certs/*.pem
sudo chmod 600 /path/to/forum/certs/*.pem

# Option 2: Add forum user to certificate group
sudo usermod -aG ssl-cert forum-user
sudo chgrp ssl-cert /etc/letsencrypt/live/yourdomain.com/*.pem
sudo chmod 640 /etc/letsencrypt/live/yourdomain.com/*.pem
```

### Security Headers
Applied via middleware to all responses:
- **Content-Security-Policy**: Restricts resource loading
- **X-Frame-Options**: DENY (prevents clickjacking)
- **X-Content-Type-Options**: nosniff
- **X-XSS-Protection**: 1; mode=block
- **Referrer-Policy**: strict-origin-when-cross-origin
- **Strict-Transport-Security**: max-age=31536000 (HSTS)
- **Permissions-Policy**: Restricts browser features

### Rate Limiting
Per-IP request throttling to prevent abuse.

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
