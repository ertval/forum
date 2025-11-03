# Forum - Modular Monolith Web Application

A modern web forum application built with Go, following Hexagonal Architecture principles and organized as a Modular Monolith. This project implements authentication, posts, comments, reactions, moderation, and notification systems with a focus on clean architecture, maintainability, and AI-agent-friendly structure.

## ⚠️ Project Status: Early Development (10% Complete)

**Current State**: Project scaffolding and architecture are complete. Most implementations contain placeholder code with TODO markers.

**What's Complete**:
- ✅ Project structure and module organization
- ✅ Database migration files defined
- ✅ Dependency injection wiring in `main.go`
- ✅ Domain entities and interfaces defined
- ✅ Error types and basic validation structures

**What's In Progress**:
- ⏳ Platform services (config, database, logger, HTTP server) - skeletons exist with TODO markers
- ⏳ All module implementations - repositories, services, and handlers are placeholder stubs

**For detailed implementation status**, see [docs/IMPLEMENTATION_ROADMAP.md](./docs/IMPLEMENTATION_ROADMAP.md)

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

## ✨ Features (Planned & In Development)

### Core Features (Structure Complete, Implementation Pending)

- 🔄 **Authentication & Authorization** (NOT IMPLEMENTED)
  - User registration with email, username, and password (TODO)
  - Login with session management (UUID-based cookies) (TODO)
  - Password encryption (bcrypt) (TODO)
  - Session expiration handling (TODO)

- 🔄 **Posts & Categories** (PARTIAL - Structure exists)
  - Create, read, update, delete posts (TODO)
  - Associate multiple categories with posts (TODO)
  - Image upload support (JPEG, PNG, GIF, max 20MB) (TODO)
  - Public viewing for all users (TODO)

- 🔄 **Comments** (PARTIAL - Structure exists)
  - Comment on posts (TODO)
  - View comments (public) (TODO)
  - Edit/delete own comments (TODO)

- 🔄 **Reactions (Likes/Dislikes)** (PARTIAL - Structure exists)
  - Like/dislike posts and comments (TODO)
  - View reaction counts (public) (TODO)
  - Only registered users can react (TODO)

- ❌ **Filtering** (NOT STARTED)
  - Filter by categories
  - Filter by user's created posts
  - Filter by user's liked posts

### Security Features (Structure Defined, Not Implemented)

- ❌ **HTTPS/TLS** (NOT IMPLEMENTED)
  - TLS 1.2+ support
  - Strong cipher suites
  - SSL certificate management

- ❌ **Rate Limiting** (NOT IMPLEMENTED)
  - Per-endpoint rate limiting
  - Per-user rate limiting

- ❌ **Secure Sessions** (NOT IMPLEMENTED)
  - UUID-based session identifiers
  - Server-side session storage
  - Secure cookie attributes

### Moderation Features (Optional - Structure exists)

- 🔄 **User Roles** [OPTIONAL FEATURE: forum-moderation] (PARTIAL)
  - Guest (view only) (TODO)
  - User (create, comment, react) (TODO)
  - Moderator (monitor, delete, report) (TODO)
  - Administrator (manage users, categories, reports) (TODO)

- 🔄 **Moderation Actions** [OPTIONAL FEATURE: forum-moderation] (PARTIAL)
  - Report posts/comments (TODO)
  - Delete inappropriate content (TODO)
  - Promote/demote moderators (TODO)
  - Review moderation reports (TODO)

### Advanced Features (Optional - Structure exists)

- 🔄 **Notifications** [OPTIONAL FEATURE: forum-advanced-features] (PARTIAL)
  - Notify on post likes/dislikes (TODO)
  - Notify on new comments (TODO)
  - Real-time notification system (TODO)

- ❌ **Activity Tracking** [OPTIONAL FEATURE: forum-advanced-features] (NOT STARTED)
  - User's created posts
  - User's likes and dislikes
  - User's comments with context

- ❌ **Content Management** (NOT STARTED)
  - Edit posts and comments
  - Delete own content
  - Moderator content removal

### Authentication Features (Future)

- ❌ **OAuth Integration** (NOT STARTED)
  - Google OAuth
  - GitHub OAuth
  - Session management across providers

**Legend**: ✅ Complete | 🔄 In Progress/Partial | ❌ Not Started

## 🚀 Getting Started

⚠️ **Note**: The application is in early development. Most features are not yet implemented.

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

3. **Run the application** (Currently will fail - implementations pending)
   ```bash
   go run cmd/forum/main.go
   ```
   
   ⚠️ **Expected**: The application will fail to start as core platform services (config, database, logger) are not yet implemented.

4. **Check implementation status**
   - See [docs/IMPLEMENTATION_ROADMAP.md](./docs/IMPLEMENTATION_ROADMAP.md) for detailed TODO list
   - Most files contain `// TODO: Implement...` comments

### Docker Deployment

⚠️ **Not Functional Yet**: Docker deployment is configured but won't work until core implementations are complete.

1. **Build and run with Docker Compose**
   ```bash
   docker-compose up --build
   ```

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

⚠️ **Status**: Test infrastructure is in place, but actual tests are not yet written.

### Current Test Status

- Test files exist with stub implementations
- `tests/unit/unit_test.go`: Placeholder with message "Unit tests will be implemented following TDD principles"
- `tests/integration/integration_test.go`: Placeholder with message "Integration tests will be implemented after core functionality"

### When Implemented

Run all tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

Run integration tests:
```bash
go test ./tests/integration/...
```

Run specific module tests:
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

- [ARCHITECTURE.md](./docs/ARCHITECTURE.md) - Detailed architecture documentation
- [IMPLEMENTATION_ROADMAP.md](./docs/IMPLEMENTATION_ROADMAP.md) - **Current implementation status with detailed TODO tracking**
- [PROJECT_STRUCTURE.md](./docs/PROJECT_STRUCTURE.md) - Project structure overview
- Module-specific documentation in each module directory

## 🔨 Development Status

### Next Steps

1. **Implement Platform Services** (Phase 1 - Priority)
   - Configuration loading (`internal/platform/config/`)
   - Database connection and migrations (`internal/platform/database/`)
   - Logger implementation (`internal/platform/logger/`)
   - HTTP server setup (`internal/platform/httpserver/`)
   - Middleware implementation

2. **Implement Auth Module** (Phase 2)
   - Session management
   - User authentication
   - Password hashing with bcrypt
   - HTTP handlers for login/register/logout

3. **Continue with remaining modules** following the roadmap

For detailed checklist, see [docs/IMPLEMENTATION_ROADMAP.md](./docs/IMPLEMENTATION_ROADMAP.md)

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
