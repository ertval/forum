# Forum - Modular Monolith Web Application

A modern web forum built with Go, following Hexagonal Architecture principles. Clean, testable, idiomatic Go code organized as a Modular Monolith.

**Current status**: Production-ready MVP - 100% Core Features Complete. All core modules (Authentication, User Management, Posts, Categories, Comments, and Reactions) are fully implemented with unit and integration tests. Platform infrastructure (config, database migrations, logging, HTTP server, health checks, upload handling, caching, and validator) is complete. Optional modules (moderation, notification) remain scaffolded. See `docs/IMPLEMENTATION_ROADMAP.md` for detailed progress.

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

- **User Management** ✅

  - User CRUD operations (get, list, update)
  - Role management (Guest, User, Moderator, Administrator)
  - User activation/deactivation
  - View user profiles with activity stats
  - Filter by created posts or liked posts

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
  - Image upload support (JPEG, PNG, GIF, max 20MB)

- **Comments** ✅

  - Comment on posts
  - Edit/delete own comments
  - Comments visible to all
  - Empty comments rejected
  - Threaded comment display

- **Reactions** ✅
  - Like/dislike posts and comments
  - Toggle reactions (can't like and dislike simultaneously)
  - Reaction counts visible to all
  - Filter posts by user's liked posts

### Security ✅

- **HTTPS/TLS** - TLS 1.2+ with strong cipher suites (AEAD)
- **Rate Limiting** - Per-IP request throttling middleware
- **Input Validation** - Email format, password strength, data sanitization
- **Secure Sessions** - UUID tokens, server-side storage, secure cookies
- **Security Headers** - CSP, X-Frame-Options, HSTS, X-XSS-Protection, Referrer-Policy, Permissions-Policy
- **Certificate Generation** - Script for self-signed certificates (`scripts/seed/generate_certs.sh`)

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

```text
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
│   │   └── main.js
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

### Docker Deployment

> For HTTPS in Docker with self-signed certificates, generate local certs first:
>
> ```bash
> bash scripts/seed/generate_certs.sh ./certs
> ```

1. **Build and run with Docker Compose**

   ```bash
   docker-compose up --build
   ```

2. **Access the forum**
  - HTTP: http://localhost:8080
  - HTTPS: https://localhost:8443

---

## Configuration

Environment variables with sane defaults:

```env
# Server
SERVER_HOST=localhost
SERVER_PORT=8080
SERVER_TLS_PORT=8443
SERVER_ENVIRONMENT=development

# Database
DATABASE_PATH=./data/forum.db
DATABASE_MIGRATIONS_DIR=./migrations

# Session
SESSION_SECRET=your-secret-key-here
SESSION_DURATION=24h
SESSION_COOKIE_NAME=forum_session
SESSION_SECURE=false

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# File Upload
UPLOAD_MAX_SIZE_MB=20
UPLOAD_DIR=./static/uploads

# OAuth (optional)
GOOGLE_OAUTH_CLIENT_ID=your-google-client-id
GOOGLE_OAUTH_CLIENT_SECRET=your-google-client-secret
GOOGLE_OAUTH_REDIRECT_URL=https://localhost:8443/auth/google/callback
GITHUB_OAUTH_CLIENT_ID=your-github-client-id
GITHUB_OAUTH_CLIENT_SECRET=your-github-client-secret
GITHUB_OAUTH_REDIRECT_URL=https://localhost:8443/auth/github/callback
```

---

## API Endpoints

All JSON API endpoints are served under the `/api` prefix using resource-oriented routes.

### Authentication ✅

- `POST /api/auth/register` - Register new user
  - Body: `{"email": "user@example.com", "username": "user", "password": "password123"}`
  - Returns: 201 Created with session cookie
  - Errors: 400 (validation), 409 (duplicate email/username), 500
- `POST /api/auth/login` - Login
  - Body: `{"email": "user@example.com", "password": "password123"}`
  - Returns: 200 OK with session cookie
  - Errors: 401 (invalid credentials), 500
- `POST /api/auth/logout` - Logout (requires auth)
  - Returns: 204 No Content
- `GET /api/auth/session` - Get current session (requires auth)
  - Returns: 200 OK with user info
  - Errors: 401 (not authenticated)

### Posts ✅

- `POST /api/posts` - Create post (requires auth)
  - Body: `{"title": "Post Title", "content": "Post content...", "categories": ["Tests"]}`
  - Returns: 201 Created with post object
  - Errors: 400 (validation - empty title/content, no categories), 401 (not authenticated), 404 (category not found), 500
- `GET /api/posts` - List posts (public)
  - Query params: `?category=Tests` (filter by category)
  - Returns: 200 OK with array of posts (includes author, categories, reaction counts)
- `GET /api/posts/{id}` - Get post by ID (public)
  - Returns: 200 OK with post object (includes author, categories, reaction counts)
  - Errors: 404 (post not found), 500
- `PUT /api/posts/{id}` - Update post (requires auth + ownership)
  - Body: `{"title": "Updated Title", "content": "Updated content...", "categories": ["Tests"]}`
  - Returns: 200 OK with updated post object
  - Errors: 400 (validation), 401 (not authenticated), 403 (not owner), 404 (post not found), 500
- `DELETE /api/posts/{id}` - Delete post (requires auth + ownership)
  - Returns: 204 No Content
  - Errors: 401 (not authenticated), 403 (not owner), 404 (post not found), 500
- `GET /api/posts/load-more` - Load more posts with pagination (public)
  - Query params: `?offset=20&limit=20&category=Tests`
  - Returns: 200 OK with array of posts

### Categories ✅

- Categories are managed internally and associated with posts
- Available categories are retrieved via post listing
- Dedicated category management endpoints exist in the codebase but are minimal; further administrative APIs can be added if needed

### Comments ✅

- `POST /api/comments/posts/{post_id}` - Add comment (requires auth)
  - Body: `{"content": "Comment text"}`
  - Returns: 201 Created with comment object
- `GET /api/comments/{id}` - Get comment (public)
  - Returns: 200 OK with comment object
- `PUT /api/comments/{id}` - Update comment (requires auth + ownership)
  - Body: `{"content": "Updated comment"}`
  - Returns: 200 OK with updated comment
- `DELETE /api/comments/{id}` - Delete comment (requires auth + ownership)
  - Returns: 204 No Content
- `GET /api/comments/posts/{post_id}` - List comments for post (public)
  - Returns: 200 OK with array of comments

### Reactions ✅

- `POST /api/reactions` - Add/toggle reaction
  - Body: `{"target_type": "post|comment", "target_id": "<uuid>", "type": "like|dislike"}`
  - Returns: 200 OK
- `DELETE /api/reactions` - Remove reaction
  - Body: `{"target_type": "post|comment", "target_id": "<uuid>"}`
  - Returns: 204 No Content
- `GET /api/reactions/{target_type}/{target_id}` - Get reactions for target (public)
  - Returns: 200 OK with array of reactions
- `GET /api/reactions/{target_type}/{target_id}/count` - Count reactions (public)
  - Returns: 200 OK with reaction counts

### Users ✅

- `GET /api/users` - List all users (public)
  - Returns: 200 OK with array of users
- `GET /api/users/{id}` - Get user profile by public ID (public)
  - Returns: 200 OK with user object
- `PUT /api/users/{id}/role` - Update user role (requires admin)
  - Body: `{"role": "user|moderator|admin"}`
- `PUT /api/users/{id}/deactivate` - Deactivate user (requires auth)
- `PUT /api/users/{id}/activate` - Activate user (requires auth)

> [!NOTE]
> HTML profile pages are currently pending; use API endpoints for user data.

### Moderation [OPTIONAL]

- `POST /api/moderation/reports` - Create report
- `GET /api/moderation/reports` - List reports (moderators)
- `PUT /api/moderation/reports/{id}` - Review report (moderators)

### Notifications [OPTIONAL]

- `GET /api/notifications` - List notifications
- `PUT /api/notifications/{id}/read` - Mark as read

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

# Image upload tests
go test ./internal/platform/upload/... -v
go test ./internal/modules/post/application/... -run Image -v
```

### Page Rendering Tests

The `scripts/tests/test_pages.sh` script tests HTML page rendering, static assets, and JavaScript API URL correctness:

```bash
./scripts/tests/test_pages.sh
```

This tests:

- HTML page structure and accessibility
- Static assets (CSS, JS)
- JavaScript API URL verification (ensures all JS files use correct `/api/` prefix)
- Template file verification

---

## Utility Scripts

### Password Verification Tool

Verify bcrypt password hashes without running the server:

```bash
# Generate a hash
go run ./scripts/verify_password -generate "mypassword"

# Verify password against a hash
go run ./scripts/verify_password "mypassword" '$2a$10$...'

# Verify from database
go run ./scripts/verify_password -db "user@example.com" "mypassword"

# With custom database path
go run ./scripts/verify_password -db -dbpath "./data/forum.db" "user@example.com" "mypassword"
```

### Schema Verification

Verify database schema matches migrations:

```bash
go run ./scripts/check/check_schema.go
```

### Database Migrations

The forum uses SQL migrations tracked in `migrations/` directory. Migrations are automatically applied on application startup.

**Manual Migration:**

```bash
# Run all pending migrations
make migrate

# Or run directly
bash scripts/seed/run_migrations.sh
```

**Create New Migration:**

```bash
# Generate a new migration file
make migration NAME=add_user_bio

# Creates: migrations/<NNN>_add_user_bio.sql (e.g. 008_add_user_bio.sql)
```

**Migration Structure:**

Each migration file follows this format:

```sql
-- +migrate Up
-- SQL to apply changes
CREATE TABLE example (id INTEGER PRIMARY KEY);

-- +migrate Down
-- SQL to rollback (if feasible)
DROP TABLE example;
```

See `migrations/MIGRATIONS_GUIDE.md` for detailed authoring guidelines.

---

## Image Uploads

The forum supports image uploads for posts with the following specifications:

### Configuration

| Environment Variable | Default            | Description                   |
| -------------------- | ------------------ | ----------------------------- |
| `STATIC_UPLOADS_DIR` | `./static/uploads` | Directory for uploaded images |
| `UPLOAD_MAX_SIZE_MB` | `20`               | Maximum file size in megabytes |

### Supported Formats

- **JPEG** (`.jpg`, `.jpeg`) - `image/jpeg`
- **PNG** (`.png`) - `image/png`
- **GIF** (`.gif`) - `image/gif`

### Security Features

- **Magic bytes validation** - Files are validated by their actual content (magic bytes), not just MIME type or extension
- **File size limits** - Configurable maximum upload size (default 20MB)
- **UUID filenames** - Uploaded files are renamed to UUIDs to prevent path traversal and filename conflicts
- **Automatic cleanup** - Images are deleted when their associated post is deleted
- **Rollback on failure** - If post creation fails after image upload, the orphaned image is automatically deleted

### Usage

**Creating a post with image:**

```html
<form method="POST" action="/posts" enctype="multipart/form-data">
  <input type="file" name="image" accept="image/jpeg,image/png,image/gif" />
  <!-- other fields -->
</form>
```

**Updating/removing image:**

```html
<form method="POST" action="/posts/{id}" enctype="multipart/form-data">
  <input type="file" name="image" />
  <input type="checkbox" name="remove_image" value="true" /> Remove current
  image
</form>
```

### Testing Image Uploads

```bash
# Run all upload package tests
go test ./internal/platform/upload/... -v -cover

# Run service-level image tests
go test ./internal/modules/post/application/... -run Image -v

# Run HTTP handler image tests
go test ./internal/modules/post/adapters/... -run Image -v

# Test with coverage
go test ./internal/platform/upload/... -coverprofile=upload_coverage.out
go tool cover -func=upload_coverage.out

# Run E2E image upload tests (requires server)
chmod +x scripts/tests/test_image_upload.sh
./scripts/tests/test_image_upload.sh           # Starts server automatically
./scripts/tests/test_image_upload.sh --no-server  # Use existing server
./scripts/tests/test_image_upload.sh -v        # Verbose output
```

The E2E test script (`scripts/tests/test_image_upload.sh`) validates all audit requirements:

- PNG, JPEG, GIF image uploads
- Oversized image (>20MB) rejection
- Image persistence verification
- Unsupported format rejection (BMP, WebP, SVG, TIFF)

---

## Dependencies

- **Database**: [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- **Password Hashing**: [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **UUID**: [gofrs/uuid/v5](https://github.com/gofrs/uuid)

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
  SERVER_ENVIRONMENT=production
   SESSION_SECRET=<strong-secret-key>
   TLS_CERT_FILE=/path/to/cert.pem
   TLS_KEY_FILE=/path/to/key.pem
   ```

3. **Obtain proper TLS certificates** (not self-signed):

   - **Recommended**: Use [Let's Encrypt](https://letsencrypt.org/) with certbot for free, automated certificates
   - **Alternative**: Purchase from a commercial CA (DigiCert, Sectigo) or use a reverse proxy (Caddy, nginx) for TLS termination
   - See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md#production-tls-certificates) for detailed setup instructions

4. Run with TLS:

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
