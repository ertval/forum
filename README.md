# Forum - Modular Monolith Web Application

A modern web forum application built with Go, following Hexagonal Architecture principles and organized as a Modular Monolith. This project implements authentication, posts, comments, reactions, moderation, and notification systems with a focus on clean architecture, maintainability, and AI-agent-friendly structure.

## 🏗️ Architecture

This project follows a **Modular Monolith** architecture with **Hexagonal Architecture** (Ports and Adapters) for each module. For detailed architecture documentation, see [ARCHITECTURE.md](./ARCHITECTURE.md).

### Key Principles

- **Go's Five Principles**: Simplicity, Readability, Explicitness, Minimalism, Composition
- **SOLID Principles**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **KISS**: Keep It Simple, Stupid

## 📁 Project Structure

```
forum/
├── cmd/
│   └── forum/                 # Application entry point
│       └── main.go           # Bootstrap and dependency injection
├── internal/
│   ├── modules/              # Business modules (hexagonal architecture)
│   │   ├── auth/            # Authentication & session management
│   │   ├── user/            # User management & roles
│   │   ├── post/            # Posts & categories
│   │   ├── comment/         # Comment management
│   │   ├── reaction/        # Likes & dislikes
│   │   ├── moderation/      # Forum moderation
│   │   └── notification/    # User notifications
│   └── platform/            # Shared infrastructure
│       ├── database/        # Database connection & migrations
│       ├── config/          # Configuration management
│       ├── logger/          # Structured logging
│       ├── httpserver/      # HTTP server & middleware
│       ├── errors/          # Common error types
│       └── validator/       # Input validation
├── migrations/              # Database migrations (organized by module)
├── static/                  # Static assets (CSS, JS, images)
├── tests/
│   ├── integration/         # Integration tests
│   └── unit/               # Unit tests
├── docker-compose.yml       # Docker Compose configuration
├── Dockerfile              # Multi-stage Docker build
├── ARCHITECTURE.md         # Detailed architecture documentation
└── README.md              # This file
```

### Module Structure (Hexagonal Architecture)

Each module follows this **flattened** structure with exactly 4 directories:

```
module/
├── domain/                  # Core business logic & entities
│   ├── entity.go           # Domain entities with business rules
│   ├── value_object.go     # Immutable value objects (if needed)
│   └── errors.go           # Domain-specific errors
├── ports/                   # Interface definitions
│   ├── service.go          # INPUT PORT: Service interface defining use cases
│   └── repository.go       # OUTPUT PORT: Repository interface for data access
├── application/             # Application services (orchestration)
│   └── service.go          # Service implementation - business logic orchestration
└── adapters/               # Interface implementations (single files, no subfolders)
    ├── http_handler.go     # INPUT ADAPTER: HTTP/REST API handlers
    ├── sqlite_repository.go # OUTPUT ADAPTER: SQLite database implementation
    └── external_api.go     # OUTPUT ADAPTER: External API clients (if needed)
```

**Port/Adapter Type Annotations:**
- Each file in `ports/` and `adapters/` includes a comment at the top indicating its type:
  - `// INPUT PORT` - Service interfaces defining use cases
  - `// OUTPUT PORT` - Repository interfaces for data access
  - `// INPUT ADAPTER` - HTTP handlers, CLI, gRPC servers
  - `// OUTPUT ADAPTER` - Database implementations, external APIs

## ✨ Features

### Core Features

- ✅ **Authentication & Authorization**
  - User registration with email, username, and password
  - Login with session management (UUID-based cookies)
  - Password encryption (bcrypt)
  - Session expiration handling

- ✅ **Posts & Categories**
  - Create, read, update, delete posts
  - Associate multiple categories with posts
  - Image upload support (JPEG, PNG, GIF, max 20MB)
  - Public viewing for all users

- ✅ **Comments**
  - Comment on posts
  - View comments (public)
  - Edit/delete own comments

- ✅ **Reactions (Likes/Dislikes)**
  - Like/dislike posts and comments
  - View reaction counts (public)
  - Only registered users can react

- ✅ **Filtering**
  - Filter by categories
  - Filter by user's created posts
  - Filter by user's liked posts

### Security Features

- ✅ **HTTPS/TLS**
  - TLS 1.2+ support
  - Strong cipher suites
  - SSL certificate management

- ✅ **Rate Limiting**
  - Per-endpoint rate limiting
  - Per-user rate limiting

- ✅ **Secure Sessions**
  - UUID-based session identifiers
  - Server-side session storage
  - Secure cookie attributes

### Moderation Features

- ✅ **User Roles** [OPTIONAL FEATURE: forum-moderation]
  - Guest (view only)
  - User (create, comment, react)
  - Moderator (monitor, delete, report)
  - Administrator (manage users, categories, reports)

- ✅ **Moderation Actions** [OPTIONAL FEATURE: forum-moderation]
  - Report posts/comments
  - Delete inappropriate content
  - Promote/demote moderators
  - Review moderation reports

### Advanced Features

- ✅ **Notifications** [OPTIONAL FEATURE: forum-advanced-features]
  - Notify on post likes/dislikes
  - Notify on new comments
  - Real-time notification system

- ✅ **Activity Tracking** [OPTIONAL FEATURE: forum-advanced-features]
  - User's created posts
  - User's likes and dislikes
  - User's comments with context

- ✅ **Content Management**
  - Edit posts and comments
  - Delete own content
  - Moderator content removal

### Authentication Features

- ✅ **OAuth Integration**
  - Google OAuth
  - GitHub OAuth
  - Session management across providers

## 🚀 Getting Started

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

3. **Run the application**
   ```bash
   go run cmd/forum/main.go
   ```

4. **Access the forum**
   - HTTP: `http://localhost:8080`
   - HTTPS: `https://localhost:8443`

### Docker Deployment

1. **Build and run with Docker Compose**
   ```bash
   docker-compose up --build
   ```

2. **Access the forum**
   - `https://localhost:8443`

### Environment Configuration

Create a `.env` file in the root directory:

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

# OAuth (optional)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# File Upload
MAX_UPLOAD_SIZE=20971520  # 20MB in bytes
UPLOAD_DIR=./static/uploads
```

## 🧪 Testing

### Run all tests
```bash
go test ./...
```

### Run with coverage
```bash
go test -cover ./...
```

### Run integration tests
```bash
go test ./tests/integration/...
```

### Run specific module tests
```bash
go test ./internal/modules/auth/...
```

## 📦 Dependencies

- **Database**: [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- **Password Hashing**: [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **UUID**: [gofrs/uuid](https://github.com/gofrs/uuid)
- **TLS**: [golang.org/x/crypto/acme/autocert](https://pkg.go.dev/golang.org/x/crypto/acme/autocert)

## 🏛️ Design Patterns

- **Hexagonal Architecture**: Clear separation of concerns with ports and adapters
- **Dependency Injection**: Modules wired together at bootstrap
- **Repository Pattern**: Data access abstraction
- **Service Layer**: Business logic orchestration
- **Middleware Pattern**: Cross-cutting concerns (auth, logging, rate limiting)

## 🤖 AI Agent Optimization

This project is structured to be AI-agent-friendly:

- **Clear module boundaries**: Easy scope understanding
- **Consistent patterns**: Every module follows the same structure
- **Explicit dependencies**: Ports clearly define contracts
- **Self-documenting code**: Comments explain purpose
- **Small, focused files**: Single responsibility
- **Type-safe interfaces**: Leverage Go's type system

## 📝 Documentation

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detailed architecture documentation
- Module-specific documentation in each module directory

## 🔒 Security

- HTTPS/TLS encryption
- bcrypt password hashing
- Rate limiting protection
- SQL injection prevention (parameterized queries)
- XSS prevention (template escaping)
- CSRF protection
- Secure session management

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## 🙏 Acknowledgments

- Built following [Go Project Layout](https://github.com/golang-standards/project-layout)
- Inspired by [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- Follows [Effective Go](https://go.dev/doc/effective_go) guidelines

---

**Note**: This is a learning project following idiomatic Go best practices, SOLID principles, and clean architecture patterns. The structure is optimized for maintainability, testability, and AI-agent collaboration.
