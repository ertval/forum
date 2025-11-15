# Forum Application - Comprehensive Project Guide

## Project Overview

This is a modern web forum built as a **Modular Monolith** using **Go 1.24+**, following **Hexagonal Architecture** (Ports and Adapters) principles. The application implements a clean, testable, and idiomatic Go architecture with a focus on maintainability and scalability.

**Current Status**: MVP Core Features - 75% Complete. Authentication, Posts, and Categories are fully implemented with comprehensive tests. Comments, reactions, and optional features (moderation, notifications) are in development.

## Architecture

### Hexagonal Architecture Pattern

The project follows a strict **Hexagonal Architecture** (Ports and Adapters) pattern organized as a **Modular Monolith**:

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

### Module Structure

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

## Dependencies

- **Database**: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- **Password Hashing**: [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **UUID Generation**: [github.com/gofrs/uuid/v5](https://github.com/gofrs/uuid/v5)
- **Go Version**: 1.24

## Building and Running

### Prerequisites

- Go 1.24+
- Docker & Docker Compose (for containerized deployment)
- SQLite3

### Local Development

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd forum
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Configure environment** (optional - has defaults)

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the application**

   ```bash
   go run cmd/forum/main.go
   ```

5. **Access the forum**
   - HTTP: http://localhost:8080
   - HTTPS: https://localhost:8443 (if TLS configured)

### Using Makefile

The project includes a comprehensive Makefile with common development tasks:

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run with hot reload (requires air)
make dev

# Build Docker image
make docker-build

# Create a new migration
make migration NAME=create_posts

# Run database migrations
make migrate
```

### Docker Deployment

1. **Build and run with Docker Compose**

   ```bash
   docker-compose up --build
   ```

2. **Access the forum**
   - http://localhost:8080

## Key Features

### Core Features (Implemented ✅)

- **Authentication & Authorization** ✅
  - User registration with email, username, and password
  - Session-based authentication (UUID cookies, server-side storage)
  - Secure password hashing (bcrypt)
  - Session expiration and management (24h default)
  - Cookie security (HttpOnly, Secure, SameSite)
  - Only ONE active session per user
  - Middleware for protected routes (RequireAuth, OptionalAuth)

- **Posts & Categories** ✅
  - Create, read, update, delete posts
  - Multiple categories per post (minimum 1 required)
  - Title validation (max 300 characters)
  - Content validation (max 50,000 characters)
  - Category management (create, list, retrieve)
  - Public viewing (all users can view posts)
  - Post ownership (only owner can edit/delete)
  - Filter posts by category
  - Image upload support (planned - JPEG, PNG, GIF, max 20MB)

### Planned Features

- **Comments**
  - Comment on posts
  - Edit/delete own comments
  - Comments visible to all
  - Empty comments rejected

- **Reactions**
  - Like/dislike posts and comments
  - Toggle reactions (can't like and dislike simultaneously)
  - Reaction counts visible to all
  - Filter posts by user's liked posts

- **User Profiles**
  - View user profiles
  - Activity tracking (created posts, comments, reactions)
  - Filter by created posts
  - Filter by liked posts

### Optional Features

- **Moderation System** [OPTIONAL]
  - User roles (Guest, User, Moderator, Administrator)
  - Report posts and comments
  - Review and handle reports
  - Delete inappropriate content
  - Promote/demote users

- **Notifications** [OPTIONAL]
  - Notify on post reactions
  - Notify on new comments
  - Mark notifications as read
  - Real-time notification updates

## API Endpoints

### Authentication ✅

- `POST /auth/register` - Register new user
  - Body: `{"email": "user@example.com", "username": "user", "password": "password123"}`
  - Returns: 201 Created with session cookie
  - Errors: 400 (validation), 409 (duplicate email/username), 500
- `POST /auth/login` - Login
  - Body: `{"email": "user@example.com", "password": "password123"}`
  - Returns: 200 OK with session cookie
  - Errors: 401 (invalid credentials), 500
- `POST /auth/logout` - Logout (requires auth)
  - Returns: 204 No Content
- `GET /auth/session` - Get current session (requires auth)
  - Returns: 200 OK with user info
  - Errors: 401 (not authenticated)

### Posts ✅

- `POST /posts` - Create post (requires auth)
  - Body: `{"title": "Post Title", "content": "Post content...", "categories": ["general"]}`
  - Returns: 201 Created with post object
  - Errors: 400 (validation - empty title/content, no categories), 401 (not authenticated), 404 (category not found), 500
- `GET /posts` - List posts (public)
  - Query params: `?category=general` (filter by category)
  - Returns: 200 OK with array of posts (includes author, categories, reaction counts)
- `GET /posts/{id}` - Get post by ID (public)
  - Returns: 200 OK with post object (includes author, categories, reaction counts)
  - Errors: 404 (post not found), 500
- `PUT /posts/{id}` - Update post (requires auth + ownership)
  - Body: `{"title": "Updated Title", "content": "Updated content...", "categories": ["general"]}`
  - Returns: 200 OK with updated post object
  - Errors: 400 (validation), 401 (not authenticated), 403 (not owner), 404 (post not found), 500
- `DELETE /posts/{id}` - Delete post (requires auth + ownership)
  - Returns: 204 No Content
  - Errors: 401 (not authenticated), 403 (not owner), 404 (post not found), 500

### Categories ✅

- Categories are managed internally and associated with posts
- Available categories are retrieved via post listing
- Future: Dedicated category management endpoints

## Testing

### Run All Tests

```bash
go test ./...
```

### Run With Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Tests

```bash
# Unit tests
go test ./internal/modules/auth/...

# Integration tests
go test ./tests/integration/...
```

## Configuration

Environment variables with sane defaults:

```env
# Server
PORT=8080
TLS_PORT=8443
ENVIRONMENT=development

# Database
DB_PATH=./forum.db

# Session
SESSION_SECRET=your-secret-key-here
SESSION_DURATION=24h

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# File Upload
MAX_UPLOAD_SIZE=20971520  # 20MB
UPLOAD_DIR=./static/uploads

# OAuth (optional)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
```

## Development Conventions

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

### Dependency Rules

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

### Dependency Injection

All wiring happens in the `cmd/forum/wire/` package, called from `main.go`. The application uses a unified DI pattern where:

- A single concrete `ServiceContainer` holds all application service implementations
- Handlers declare a local interface with only the accessor methods they need
- All handlers use the same constructor signature with the ServiceContainer

### Database Design

- **Technology**: SQLite for simplicity and embedded storage
- **Migrations**: Sequential numbered SQL files in `migrations/` with `-- +migrate Up`/`-- +migrate Down` markers
- **Key Patterns**: Foreign keys with `ON DELETE CASCADE`, indexes on frequently queried columns, timestamps on all entities

### Error Handling

**Two-Layer System**:

- **Domain Errors** (simple): `var ErrSessionExpired = errors.New("session has expired")`
- **Platform Errors** (structured): `return errors.Wrap(err, errors.ErrCodeInternal, "failed to create session")`

## Security Features

- **HTTPS/TLS** - TLS 1.2+, strong ciphers
- **Password Hashing** - bcrypt with cost factor 12
- **Session Management** - UUID tokens, HttpOnly cookies, server-side storage
- **Rate Limiting** - Per-IP and per-user limits
- **Input Validation** - Email format, password strength, data sanitization
- **CSRF Protection** - CSRF tokens on state-changing operations
- **Security Headers** - CSP, X-Frame-Options, HSTS

## Key Directories and Files

- `cmd/forum/main.go` - Application entry point
- `cmd/forum/wire/` - Dependency injection package
- `internal/modules/` - Business modules (auth, user, post, comment, reaction, moderation, notification)
- `internal/platform/` - Shared infrastructure (config, database, logger, httpserver, errors, validator)
- `migrations/` - Database migration files
- `static/` - Static assets (CSS, JS, uploads)
- `templates/` - HTML templates
- `tests/` - Unit and integration tests
- `Dockerfile` - Multi-stage Docker build
- `docker-compose.yml` - Docker Compose configuration
- `Makefile` - Common development tasks
- `docs/` - Documentation
  - `docs/ARCHITECTURE.md` - Detailed architecture explanation
  - `docs/IMPLEMENTATION_ROADMAP.md` - Development progress and plan

## Project Status

The project is currently 75% complete with core features like authentication, posts, and categories fully implemented. Comments, reactions, and optional modules are in development. The architecture is well-established and ready for expansion.

For detailed implementation progress, see `docs/IMPLEMENTATION_ROADMAP.md`.