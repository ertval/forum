# Forum - Modular Monolith Web Application

A modern web forum built with Go, following Hexagonal Architecture principles. Clean, testable, idiomatic Go code organized as a Modular Monolith.

Current status: This repository contains an initial scaffolding of the application (most modules and files are present with placeholders and TODO notes). The project is approximately 10% complete — there are many unfinished implementations and TODOs. See `docs/IMPLEMENTATION_ROADMAP.md` for the prioritized plan and next steps.

## Features

### Core Features

- **Authentication & Authorization**
  - User registration with email, username, and password
  - Session-based authentication (UUID cookies, server-side storage)
  - Secure password hashing (bcrypt)
  - Session expiration and management
  - Cookie security (HttpOnly, Secure, SameSite)

- **Posts & Categories**
  - Create, read, update, delete posts
  - Multiple categories per post
  - Image upload support (JPEG, PNG, GIF, max 20MB)
  - Filter by category
  - Public viewing (all users)
  - Post ownership (edit/delete own posts)

- **Comments**
  - Comment on posts
  - Edit/delete own comments
  - Nested comment threads
  - Public viewing

- **Reactions**
  - Like/dislike posts and comments
  - Toggle reactions (can't like and dislike simultaneously)
  - Reaction counts visible to all
  - Filter posts by user's liked posts

- **User Profiles**
  - View user profiles
  - Activity tracking (created posts, comments, reactions)
  - Profile editing

### Security

- **HTTPS/TLS** - TLS 1.2+ with strong cipher suites
- **Rate Limiting** - Per-IP and per-user request throttling
- **CSRF Protection** - Token-based CSRF prevention
- **Input Validation** - Email format, password strength, data sanitization
- **Secure Sessions** - UUID tokens, server-side storage, secure cookies
- **Security Headers** - CSP, X-Frame-Options, HSTS

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

---

## Architecture

This project follows **Hexagonal Architecture** (Ports and Adapters) organized as a **Modular Monolith**.

### Key Principles

- **Go's Five Principles**: Simplicity, Readability, Explicitness, Minimalism, Composition
- **SOLID Principles**: Single Responsibility, Interface Segregation, Dependency Inversion
- **KISS**: Keep It Simple

### Module Structure

Every module follows this exact 4-directory layout:

```
module/
├── domain/          # Pure business logic (no external dependencies)
├── ports/           # Interface definitions (contracts)
├── application/     # Business logic orchestration
└── adapters/        # Technical implementations (HTTP, database)
```

**For detailed architecture documentation**, see [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

---

## Project Structure

``` text
forum/
├── bin/                     # Compiled binaries
│   └── forum               # Main application binary
├── cmd/
│   └── forum/              # Application entry point
│       ├── main.go         # Minimal entry point
│       └── wire/           # Dependency injection package
│           ├── app.go      # Application struct and lifecycle
│           ├── handlers.go # HTTP handler initialization
│           ├── repos.go    # Repository initialization
│           ├── services.go # Service initialization
│           └── README.md   # Wire package documentation
├── data/                   # Data directory (for databases, etc.)
├── docs/                   # Documentation
│   ├── ARCHITECTURE.md     # Architecture documentation
│   ├── IMPLEMENTATION_ROADMAP.md
│   ├── requirements.md
│   ├── morefeats.md
│   └── copilot-instructions.md
├── internal/
│   ├── modules/            # Business modules (hexagonal architecture)
│   │   ├── auth/          # Authentication & sessions
│   │   │   ├── adapters/  # HTTP handlers, SQLite repositories
│   │   │   ├── application/# Business logic orchestration
│   │   │   ├── domain/    # Pure business logic
│   │   │   └── ports/     # Interface definitions
│   │   ├── user/          # User management & roles
│   │   │   ├── adapters/
│   │   │   ├── application/
│   │   │   ├── domain/
│   │   └── ports/
│   │   ├── post/          # Posts & categories
│   │   │   ├── adapters/
│   │   │   ├── application/
│   │   │   ├── domain/
│   │   └── ports/
│   │   ├── comment/       # Comments
│   │   │   ├── adapters/
│   │   │   ├── application/
│   │   │   ├── domain/
│   │   └── ports/
│   │   ├── reaction/      # Likes & dislikes
│   │   │   ├── adapters/
│   │   │   ├── application/
│   │   │   ├── domain/
│   │   └── ports/
│   │   ├── moderation/    # [OPTIONAL] Forum moderation
│   │   │   ├── adapters/
│   │   │   ├── application/
│   │   │   ├── domain/
│   │   └── ports/
│   │   └── notification/  # [OPTIONAL] Notifications
│   │       ├── adapters/
│   │       ├── application/
│   │       ├── domain/
│   │       └── ports/
│   └── platform/          # Shared infrastructure
│       ├── config/        # Configuration management
│       │   ├── config.go
│       │   └── env_parser.go
│       ├── database/      # SQLite connection & migrations
│       ├── logger/        # Structured logging
│       ├── httpserver/    # HTTP server & middleware
│       ├── errors/        # Error handling
│       └── validator/     # Input validation
├── migrations/            # Database migrations
│   ├── 001_auth_create_sessions.sql
│   ├── 002_user_create_users.sql
│   ├── 003_post_create_tables.sql
│   ├── 004_comment_create_comments.sql
│   ├── 005_reaction_create_reactions.sql
│   ├── 006_moderation_create_reports.sql
│   ├── 007_notification_create_notifications.sql
│   └── README.md
├── static/                # Static assets
│   ├── css/
│   │   └── style.css
│   ├── js/
│   │   └── app.js
│   └── uploads/           # User uploaded files
├── templates/             # HTML templates
│   ├── base.html
│   └── home.html
├── tests/                 # Unit and integration tests
│   ├── integration/
│   │   ├── integration_test.go
│   │   └── config/
│   │       └── main.go
│   └── unit/
│       └── unit_test.go
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile             # Docker build configuration
├── go.mod                 # Go module definition
├── LICENSE                # License file
└── README.md              # This file
```


## Getting Started

### Prerequisites

- Go 1.25+
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

### Docker Deployment

1. **Build and run with Docker Compose**

   ```bash
   docker-compose up --build
   ```

2. **Access the forum**
   - http://localhost:8080

---

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

---

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
- `DELETE /api/posts/:id` - Delete post (moderators)
- `DELETE /api/comments/:id` - Delete comment (moderators)

### Notifications [OPTIONAL]

- `GET /api/notifications` - List notifications
- `PUT /api/notifications/:id/read` - Mark as read
- `PUT /api/notifications/read-all` - Mark all as read

---

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

---

## Dependencies

- **Database**: [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- **Password Hashing**: [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **UUID**: [gofrs/uuid](https://github.com/gofrs/uuid) or [google/uuid](https://github.com/google/uuid)

**Minimal dependencies** - Prefer standard library where reasonable.

---

## Development

### Implementation Roadmap

See [docs/IMPLEMENTATION_ROADMAP.md](./docs/IMPLEMENTATION_ROADMAP.md) for:
- Current implementation status
- Phase-by-phase development plan
- MVP path (6-8 days to functional forum)
- Post-MVP enhancements

### Architecture Documentation

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) for:

- Detailed architecture explanation
- Hexagonal Architecture patterns
- Module structure and dependencies
- Design decisions and rationale

---

## Build & Deployment

### Build Binary

```bash
# Local build
go build -o bin/forum cmd/forum/main.go

# With CGO (required for SQLite)
CGO_ENABLED=1 go build -o bin/forum cmd/forum/main.go
```

### Docker Build

```bash
docker build -t forum:latest .
docker run -p 8080:8080 forum:latest
```

### Production Deployment

1. Build with optimizations:

   ```bash
   CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o forum cmd/forum/main.go
   ```

2. Set production environment:

   ```env
   ENVIRONMENT=production
   SESSION_SECRET=<strong-secret-key>
   TLS_CERT_FILE=/path/to/cert.pem
   TLS_KEY_FILE=/path/to/key.pem
   ```

3. Run with TLS:

   ```bash
   ./forum
   ```

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow Go conventions and project architecture patterns
4. Write tests for new features
5. Commit changes (`git commit -m 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write clear commit messages
- Keep functions small and focused

---

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

---

## Acknowledgments

- Built following [Go Project Layout](https://github.com/golang-standards/project-layout)
- Inspired by [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- Follows [Effective Go](https://go.dev/doc/effective_go) guidelines

---

**A learning project demonstrating clean architecture, SOLID principles, and idiomatic Go best practices.**
