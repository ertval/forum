# Forum - Architecture Documentation

## Overview

This project implements a web forum using a **Modular Monolith** architecture combined with **Hexagonal Architecture** (Ports and Adapters) principles. The design follows Go's idiomatic practices, SOLID principles, and KISS philosophy.

## Architectural Principles

### Go's Five Principles
1. **Simplicity**: Keep solutions straightforward and avoid unnecessary complexity
2. **Readability**: Write clear, self-documenting code
3. **Explicitness**: Make dependencies and behavior obvious
4. **Minimalism**: Use only what you need
5. **Composition over Inheritance**: Favor interfaces and composition

### SOLID Principles
- **Single Responsibility**: Each module/package has one reason to change
- **Open/Closed**: Open for extension, closed for modification
- **Liskov Substitution**: Interfaces are properly abstracted
- **Interface Segregation**: Small, focused interfaces
- **Dependency Inversion**: Depend on abstractions, not concretions

### KISS (Keep It Simple, Stupid)
- Avoid over-engineering
- Choose straightforward solutions
- Clear naming and structure

## Architecture Overview

### Modular Monolith

The application is organized as a monolith but with clear module boundaries. Each module is independent and can potentially be extracted into a microservice if needed.

```
forum/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── modules/           # Business modules
│   └── platform/          # Shared infrastructure
├── pkg/                   # Public libraries (if needed)
└── migrations/            # Database migrations
```

### Hexagonal Architecture (Ports and Adapters)

Each module follows hexagonal architecture:

```
module/
├── domain/                # Business logic and entities (core)
├── ports/                 # Interfaces (primary and secondary)
│   ├── input/            # Inbound ports (use cases)
│   └── output/           # Outbound ports (repositories, external services)
├── application/           # Application services (orchestration)
└── adapters/             # Implementations of ports
    ├── input/            # HTTP handlers, CLI commands
    └── output/           # Database repositories, external API clients
```

**Flow**: `Adapters (Input) → Application Services → Domain Logic → Adapters (Output)`

## Module Structure

### Core Modules

#### 1. Authentication Module (`internal/modules/auth`)
Handles user authentication, session management, cookies, and security.

**Responsibilities:**
- User registration and login
- Session management with UUID
- Cookie handling
- Password hashing (bcrypt)
- OAuth integration (Google, GitHub)
- Rate limiting
- HTTPS/TLS certificate management

**Ports:**
- Input: AuthService (registration, login, logout, OAuth)
- Output: SessionRepository, UserRepository (read-only)

#### 2. User Module (`internal/modules/user`)
Manages user profiles, roles, and permissions.

**Responsibilities:**
- User CRUD operations
- Role management (Guest, User, Moderator, Admin)
- User promotion/demotion
- Profile management

**Ports:**
- Input: UserService
- Output: UserRepository

#### 3. Post Module (`internal/modules/post`)
Handles forum posts and categories.

**Responsibilities:**
- Post creation, editing, deletion
- Category management
- Post filtering (by category, created posts, liked posts)
- Image upload (JPEG, PNG, GIF, max 20MB)

**Ports:**
- Input: PostService, CategoryService
- Output: PostRepository, CategoryRepository, ImageStorage

#### 4. Comment Module (`internal/modules/comment`)
Manages comments on posts.

**Responsibilities:**
- Comment creation, editing, deletion
- Comment retrieval by post

**Ports:**
- Input: CommentService
- Output: CommentRepository

#### 5. Reaction Module (`internal/modules/reaction`)
Handles likes and dislikes for posts and comments.

**Responsibilities:**
- Like/dislike posts
- Like/dislike comments
- Count reactions
- Toggle reactions

**Ports:**
- Input: ReactionService
- Output: ReactionRepository

#### 6. Moderation Module (`internal/modules/moderation`)
Forum moderation functionality.

**Responsibilities:**
- Report posts/comments
- Review reports
- Delete content
- Manage moderator actions

**Ports:**
- Input: ModerationService
- Output: ReportRepository

#### 7. Notification Module (`internal/modules/notification`)
Real-time user notifications.

**Responsibilities:**
- Notify on post likes/dislikes
- Notify on comments
- Track user activity
- Notification preferences

**Ports:**
- Input: NotificationService
- Output: NotificationRepository

### Platform/Shared Infrastructure (`internal/platform`)

Shared code used across all modules:

- **database**: SQLite connection, migrations, transaction handling
- **config**: Configuration loading and management
- **logger**: Structured logging
- **httpserver**: HTTP server setup, middleware
- **errors**: Common error types and handling
- **validator**: Input validation utilities

## Dependency Rules

1. **Modules are independent**: No direct imports between modules
2. **Depend on interfaces**: Modules expose interfaces in `ports/input`
3. **Platform is shared**: All modules can import from `platform`
4. **Domain is pure**: No external dependencies in domain layer
5. **Dependency injection**: Wire modules at application bootstrap (`cmd/forum/main.go`)

## Module Communication

Modules communicate through:
1. **Defined interfaces** in `ports/input`
2. **Dependency injection** at startup
3. **Events** (for asynchronous communication, if needed)

Example:
```go
// In post module
type ReactionCounter interface {
    CountReactions(targetID string, targetType string) (likes, dislikes int, err error)
}

// Injected from reaction module
postService := post.NewService(postRepo, reactionCounter)
```

## Database Design

- **Single SQLite database** shared across all modules
- Each module owns its tables
- **Migrations** are versioned and organized by module
- Use **foreign keys** for referential integrity
- Implement **Entity Relationship Diagram** (ERD) before implementation

## Security Considerations

1. **HTTPS**: TLS 1.2+ with strong cipher suites
2. **Password encryption**: bcrypt hashing
3. **Session management**: UUID-based, server-side storage
4. **Rate limiting**: Per-endpoint and per-user
5. **Input validation**: All user inputs sanitized
6. **SQL injection prevention**: Parameterized queries
7. **XSS prevention**: Template escaping
8. **CSRF protection**: Token-based

## Testing Strategy

### Per Module
- **Unit tests**: `domain/` and `application/` layers
- **Integration tests**: `adapters/output/` (database)
- **HTTP tests**: `adapters/input/` (handlers)

### Project Level
- **Integration tests**: `tests/integration/` - Cross-module workflows
- **E2E tests**: Full user journeys

## AI Agent Optimization

The structure is optimized for AI coding agents:

1. **Clear module boundaries**: Easy to understand scope
2. **Consistent patterns**: Every module follows same structure
3. **Explicit dependencies**: Ports clearly define contracts
4. **Self-documenting**: Comments explain purpose
5. **Small files**: Focused, single-responsibility files
6. **Type-safe**: Leverage Go's type system

## Build and Deployment

### Docker
- Multi-stage build for optimization
- Separate development and production configs
- Health checks
- Volume management for database

### Running Locally
```bash
go run cmd/forum/main.go
```

### Running with Docker
```bash
docker-compose up --build
```

## Future Considerations

- **Event-driven architecture**: Introduce event bus for async module communication
- **CQRS**: Separate read and write models if needed
- **Microservices extraction**: Modules can be extracted independently
- **API versioning**: Plan for API evolution
- **Observability**: Metrics, tracing, and monitoring

## References

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [SOLID Principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design)
- [Effective Go](https://go.dev/doc/effective_go)
