# Forum - Modular Monolith Web Application

## Project Overview

This is a modern web forum built with Go, following Hexagonal Architecture principles. It is a clean, testable, idiomatic Go codebase organized as a Modular Monolith. The project implements a comprehensive forum with authentication, posts, comments, reactions, user profiles, and optional moderation and notification systems.

### Key Technologies

- **Language**: Go 1.25
- **Database**: SQLite with migrations
- **Architecture**: Hexagonal Architecture (Ports and Adapters)
- **Dependencies**: 

  - `github.com/mattn/go-sqlite3` - SQLite driver
  - `golang.org/x/crypto` - Bcrypt for password hashing
  - `github.com/gofrs/uuid/v5` - UUID generation

### Architecture Pattern

The project follows **Hexagonal Architecture** (Ports and Adapters) organized as a **Modular Monolith** with these key principles:

- Go's Five Principles: Simplicity, Readability, Explicitness, Minimalism, Composition
- SOLID Principles: Single Responsibility, Interface Segregation, Dependency Inversion
- KISS: Keep It Simple

Each module follows a strict 4-directory layout:
```
module/
├── domain/          # Pure business logic (no external dependencies)
├── ports/           # Interface definitions (contracts)
├── application/     # Business logic orchestration
└── adapters/        # Technical implementations (HTTP, database)
```

## Module Structure

The application is organized into several modules:

- `auth` - Authentication & sessions
- `user` - User management & roles
- `post` - Posts & categories
- `comment` - Comments
- `reaction` - Likes & dislikes
- `moderation` - Forum moderation (optional)
- `notification` - Notifications (optional)

## Building and Running

### Prerequisites

- Go 1.25+
- Docker & Docker Compose (for containerized deployment)
- SQLite3

### Local Development

1. **Install dependencies**

   ```bash
   go mod download
   ```

2. **Run the application**

   ```bash
   go run cmd/forum/main.go
   ```

3. **Access the forum**
   - HTTP: http://localhost:8080
   - HTTPS: https://localhost:8443 (if TLS configured)

### Docker Deployment

1. **Build and run with Docker Compose**

   ```bash
   docker-compose up --build
   ```

2. **Access the forum**
   - http://localhost:8080

### Build Binary

```bash
# Local build
go build -o bin/forum cmd/forum/main.go

# With CGO (required for SQLite)
CGO_ENABLED=1 go build -o bin/forum cmd/forum/main.go
```

### Production Deployment

1. Build with optimizations:

   ```bash
   CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o forum cmd/forum/main.go
   ```

2. Set production environment variables

## Configuration

The application uses environment variables with defaults for configuration. Key configuration options include:

- Server: `PORT`, `TLS_PORT`, `ENVIRONMENT`
- Database: `DB_PATH`
- Session: `SESSION_SECRET`, `SESSION_DURATION`
- Rate limiting: `RATE_LIMIT_REQUESTS`, `RATE_LIMIT_WINDOW`
- File uploads: `MAX_UPLOAD_SIZE`, `UPLOAD_DIR`
- OAuth: Various provider-specific variables

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout
- `GET /api/auth/session` - Get current session

### Posts

- `POST /api/posts` - Create post
- `GET /api/posts` - List posts (with filters)
- `GET /api/posts/:id` - Get post by ID
- `PUT /api/posts/:id` - Update post
- `DELETE /api/posts/:id` - Delete post

### Comments

- `POST /api/posts/:id/comments` - Add comment
- `GET /api/posts/:id/comments` - List comments
- `PUT /api/comments/:id` - Update comment
- `DELETE /api/comments/:id` - Delete comment

### Reactions

- `POST /api/reactions` - Add/toggle reaction
- `DELETE /api/reactions/:id` - Remove reaction
- `GET /api/posts/:id/reactions` - Get reaction counts
- `GET /api/comments/:id/reactions` - Get reaction counts

### Users

- `GET /api/users/:id` - Get user profile
- `PUT /api/users/:id` - Update profile
- `GET /api/users/:id/posts` - User's posts
- `GET /api/users/:id/activity` - User's activity

### Moderation [OPTIONAL]

- `POST /api/reports` - Create report
- `GET /api/reports` - List reports (moderators)
- `PUT /api/reports/:id` - Review report (moderators)

### Notifications [OPTIONAL]

- `GET /api/notifications` - List notifications
- `PUT /api/notifications/:id/read` - Mark as read

## Features

### Core Features

- **Authentication & Authorization**: User registration with email, username, and password; Session-based authentication (UUID cookies, server-side storage); Secure password hashing (bcrypt); Session expiration and management
- **Posts & Categories**: Create, read, update, delete posts; Multiple categories per post; Image upload support (JPEG, PNG, GIF, max 20MB)
- **Comments**: Comment on posts; Edit/delete own comments; Nested comment threads
- **Reactions**: Like/dislike posts and comments; Toggle reactions; Reaction counts visible to all
- **User Profiles**: View user profiles; Activity tracking (created posts, comments, reactions); Profile editing

### Security Features

- **HTTPS/TLS**: TLS 1.2+ with strong cipher suites
- **Rate Limiting**: Per-IP and per-user request throttling
- **CSRF Protection**: Token-based CSRF prevention
- **Input Validation**: Email format, password strength, data sanitization
- **Secure Sessions**: UUID tokens, server-side storage, secure cookies
- **Security Headers**: CSP, X-Frame-Options, HSTS

## Development Conventions

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write clear commit messages
- Keep functions small and focused

### Architecture Patterns

- Domain entities with business logic in the `domain` package
- Interfaces in the `ports` package (input and output ports)
- Business logic orchestration in the `application` package
- Technical implementations in the `adapters` package
- Dependency injection via the wire package

### Testing

- Run all tests: `go test ./...`
- Run with coverage: `go test -cover ./...`
- Run specific tests: `go test ./internal/modules/auth/...`

## Project Structure

```text
forum/
├── bin/                     # Compiled binaries
├── cmd/
│   └── forum/              # Application entry point
│       ├── main.go         # Minimal entry point
│       └── wire/           # Dependency injection package
├── docs/                   # Documentation
├── internal/
│   ├── modules/            # Business modules (hexagonal architecture)
│   │   ├── auth/          # Authentication & sessions
│   │   ├── user/          # User management & roles
│   │   ├── post/          # Posts & categories
│   │   ├── comment/       # Comments
│   │   ├── reaction/      # Likes & dislikes
│   │   ├── moderation/    # Forum moderation
│   │   └── notification/  # Notifications
│   └── platform/          # Shared infrastructure
│       ├── config/        # Configuration management
│       ├── database/      # SQLite connection & migrations
│       ├── logger/        # Structured logging
│       ├── httpserver/    # HTTP server & middleware
│       ├── errors/        # Error handling
│       └── validator/     # Input validation
├── migrations/            # Database migrations
├── static/                # Static assets
├── templates/             # HTML templates
├── tests/                 # Unit and integration tests
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile             # Docker build configuration
├── go.mod                 # Go module definition
└── README.md              # Documentation
```
