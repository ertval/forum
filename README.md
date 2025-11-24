# Forum - Modular Monolith Web Application

A modern web forum built with Go, following Hexagonal Architecture principles. Clean, testable, idiomatic Go code organized as a Modular Monolith.

**Current status**: MVP Core Features - 90% Complete. Authentication, Posts, Categories, and platform infrastructure (config, database migrations, logging, HTTP server, and validator) are implemented with unit and integration tests. Remaining items: Comments and Reactions modules (mostly scaffolded), some HTTP handlers for `user` and optional modules (moderation, notification). See `docs/IMPLEMENTATION_ROADMAP.md` for detailed progress and next steps.

## Features

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
  - Filter posts by category, user, liked posts, and date range
  - Date filtering (Today, This Week, This Month, All Time)
  - Dedicated FilterService in application layer for centralized filtering logic
  - Image upload support (implemented as upload handling + validation; storage and CDN integration remain optional - JPEG, PNG, GIF, max 20MB)

### Core Features (Planned / Near Complete)

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

Every module follows this exact 4-directory layout (enforced across implemented modules):

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
  - Body: `{"title": "Post Title", "content": "Post content...", "categories": ["Tests"]}`
  - Returns: 201 Created with post object
  - Errors: 400 (validation - empty title/content, no categories), 401 (not authenticated), 404 (category not found), 500
- `GET /posts` - List posts (public)
  - Query params: `?category=Tests` (filter by category)
  - Returns: 200 OK with array of posts (includes author, categories, reaction counts)
- `GET /posts/{id}` - Get post by ID (public)
  - Returns: 200 OK with post object (includes author, categories, reaction counts)
  - Errors: 404 (post not found), 500
- `PUT /posts/{id}` - Update post (requires auth + ownership)
  - Body: `{"title": "Updated Title", "content": "Updated content...", "categories": ["Tests"]}`
  - Returns: 200 OK with updated post object
  - Errors: 400 (validation), 401 (not authenticated), 403 (not owner), 404 (post not found), 500
- `DELETE /posts/{id}` - Delete post (requires auth + ownership)
  - Returns: 204 No Content
  - Errors: 401 (not authenticated), 403 (not owner), 404 (post not found), 500

### Categories ✅

- Categories are managed internally and associated with posts
- Available categories are retrieved via post listing
- Dedicated category management endpoints exist in the codebase but are minimal; further administrative APIs can be added if needed

### Comments (Planned / Scaffolded)

- `POST /posts/:id/comments` - Add comment (adapter & repository scaffolding exists)
- `GET /posts/:id/comments` - List comments
- `PUT /comments/:id` - Update comment
- `DELETE /comments/:id` - Delete comment

### Reactions (Planned / Scaffolded)

- `POST /reactions` - Add/toggle reaction (data model and migrations present)
- `DELETE /reactions/:id` - Remove reaction
- `GET /posts/:id/reactions` - Get reaction counts
- `GET /comments/:id/reactions` - Get reaction counts

### Users (Planned / Partially Implemented)

- `GET /users/:id` - Get user profile
- `PUT /users/:id` - Update profile (handlers not fully implemented)
- `GET /users/:id/posts` - User's posts
- `GET /users/:id/activity` - User's activity

### Moderation [OPTIONAL]

- `POST /reports` - Create report
- `GET /reports` - List reports (moderators)
- `PUT /reports/:id` - Review report (moderators)
- `DELETE /posts/:id` - Delete post (moderators)
- `DELETE /comments/:id` - Delete comment (moderators)

### Notifications [OPTIONAL]

- `GET /notifications` - List notifications
- `PUT /notifications/:id/read` - Mark as read
- `PUT /notifications/read-all` - Mark all as read

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
