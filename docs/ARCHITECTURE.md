# Forum Application - Architecture Documentation

## Overview

This forum application is built as a **Modular Monolith** with each module following **Hexagonal Architecture** (Ports and Adapters) principles. The architecture is designed to be AI-agent-friendly, maintainable, testable, and follows Go's idiomatic patterns.

## Architectural Principles

### Go's Five Principles
1. **Simplicity**: Keep solutions simple and straightforward
2. **Readability**: Code should be easy to read and understand
3. **Explicitness**: Make behavior explicit, avoid hidden magic
4. **Minimalism**: Use minimal dependencies and abstractions
5. **Composition**: Build complex behavior through composition

### SOLID Principles
1. **Single Responsibility**: Each module/component has one reason to change
2. **Open/Closed**: Open for extension, closed for modification
3. **Liskov Substitution**: Subtypes must be substitutable for their base types
4. **Interface Segregation**: Many specific interfaces better than one general
5. **Dependency Inversion**: Depend on abstractions, not concretions

### KISS Principle
**Keep It Simple, Stupid**: Favor simple solutions over complex ones

## Project Structure

```
forum/
├── cmd/
│   └── forum/                 # Application entry point
│       └── main.go           # Bootstrap and dependency injection
├── internal/
│   ├── modules/              # Business modules (core + optional)
│   │   ├── auth/            # [CORE] Authentication & session management
│   │   ├── user/            # [CORE] User management & roles
│   │   ├── post/            # [CORE] Posts & categories
│   │   ├── comment/         # [CORE] Comment management
│   │   ├── reaction/        # [CORE] Likes & dislikes
│   │   ├── moderation/      # [OPTIONAL] Forum moderation system
│   │   └── notification/    # [OPTIONAL] User notifications
│   └── platform/            # Shared infrastructure
│       ├── database/        # Database connection & migrations
│       ├── config/          # Configuration management
│       ├── logger/          # Structured logging
│       ├── httpserver/      # HTTP server & middleware
│       ├── errors/          # Common error types
│       └── validator/       # Input validation
├── migrations/              # Database migrations (organized by module)
├── static/                  # Static assets (CSS, JS, images)
│   ├── css/
│   ├── js/
│   └── uploads/
├── templates/               # HTML templates
├── tests/
│   ├── integration/         # Integration tests
│   └── unit/               # Unit tests
├── docker-compose.yml       # Docker Compose configuration
├── Dockerfile              # Multi-stage Docker build
├── .gitignore
├── LICENSE
├── ARCHITECTURE.md         # This file
└── README.md              # Project overview & setup
```

## Module Structure (Hexagonal Architecture)

Each module follows this flattened structure with exactly 4 directories:

```
module/
├── domain/                  # Core business logic & entities
│   ├── entity.go           # Domain entities with business rules
│   ├── value_object.go     # Immutable value objects
│   ├── repository.go       # Repository interface (output port)
│   └── errors.go           # Domain-specific errors
├── ports/                   # Interface definitions
│   ├── service.go          # Service interface (input port) - defines use cases
│   └── repository.go       # Repository interface (output port) - data access contract
├── application/             # Application services (orchestration)
│   └── service.go          # Service implementation - business logic orchestration
└── adapters/               # Interface implementations
    ├── http_handler.go     # [INPUT] HTTP/REST API handlers
    ├── sqlite_repository.go # [OUTPUT] SQLite database implementation
    └── external_api.go     # [OUTPUT] External API clients (if needed)
```

### Port Types (Noted as Comments)

Each file in `ports/` and `adapters/` includes a comment at the top indicating its type:

- **INPUT PORT**: Defines use cases (service interfaces)
- **OUTPUT PORT**: Defines data access contracts (repository interfaces)
- **INPUT ADAPTER**: HTTP handlers, CLI, gRPC servers
- **OUTPUT ADAPTER**: Database implementations, external APIs

### Key Concepts

#### Domain Layer
- **Pure business logic** with no external dependencies
- Contains entities, value objects, and domain errors
- Defines repository interfaces (output ports)
- No knowledge of HTTP, databases, or frameworks

#### Ports
- **Input Ports**: Service interfaces defining use cases
- **Output Ports**: Repository interfaces for data access
- Abstract interfaces that the domain depends on

#### Application Layer
- **Orchestrates** domain logic to implement use cases
- Depends on domain entities and port interfaces
- Implements input ports (service interfaces)
- Uses output ports (repository interfaces) via dependency injection

#### Adapters
- **Input Adapters**: HTTP handlers, CLI commands, gRPC services
- **Output Adapters**: Database repositories, external API clients
- Implement port interfaces
- Handle technical details (HTTP, SQL, JSON)

## Module Descriptions

### Core Modules (Required)

#### 1. Auth Module
**Purpose**: Authentication and session management

**Responsibilities**:
- User registration (email, username, password)
- User login/logout
- Session management with UUID-based cookies
- Password encryption (bcrypt)
- Session expiration handling

**Key Entities**: Session

**Ports**:
- Input: AuthService (Register, Login, Logout, ValidateSession)
- Output: SessionRepository

---

#### 2. User Module
**Purpose**: User management and roles

**Responsibilities**:
- User profile management
- User role assignment (Guest, User, Moderator, Admin)
- User permissions
- User activity tracking

**Key Entities**: User, Role

**Ports**:
- Input: UserService (CreateUser, GetUser, UpdateUser, AssignRole)
- Output: UserRepository

---

#### 3. Post Module
**Purpose**: Posts and categories management

**Responsibilities**:
- Create, read, update, delete posts
- Category management
- Associate posts with categories
- Image upload (JPEG, PNG, GIF, max 20MB)
- Post filtering (by category, by user, by liked)

**Key Entities**: Post, Category

**Ports**:
- Input: PostService (CreatePost, GetPost, UpdatePost, DeletePost, FilterPosts)
- Output: PostRepository, CategoryRepository

---

#### 4. Comment Module
**Purpose**: Comment management

**Responsibilities**:
- Create, read, update, delete comments
- Associate comments with posts
- View comment threads

**Key Entities**: Comment

**Ports**:
- Input: CommentService (CreateComment, GetComment, UpdateComment, DeleteComment)
- Output: CommentRepository

---

#### 5. Reaction Module
**Purpose**: Likes and dislikes for posts and comments

**Responsibilities**:
- Like/unlike posts and comments
- Dislike/undislike posts and comments
- View reaction counts
- Track user reactions

**Key Entities**: Reaction

**Ports**:
- Input: ReactionService (React, GetReactions, CountReactions)
- Output: ReactionRepository

---

### Optional Modules (Extra Features)

#### 6. Moderation Module [OPTIONAL]
**Purpose**: Forum moderation system

**Responsibilities**:
- Report posts and comments
- Review and handle reports
- Delete inappropriate content
- Promote/demote moderators
- Admin actions

**Key Entities**: Report, ModerationAction

**Ports**:
- Input: ModerationService (CreateReport, ReviewReport, DeleteContent, PromoteUser)
- Output: ReportRepository

**Requirements**: forum-moderation feature

---

#### 7. Notification Module [OPTIONAL]
**Purpose**: User notifications

**Responsibilities**:
- Notify users on post likes/dislikes
- Notify users on new comments
- Mark notifications as read
- Real-time notification delivery

**Key Entities**: Notification

**Ports**:
- Input: NotificationService (CreateNotification, GetNotifications, MarkAsRead)
- Output: NotificationRepository

**Requirements**: forum-advanced-features

---

## Platform (Shared Infrastructure)

### Database
- SQLite connection management
- Migration execution
- Transaction handling
- Connection pooling

### Config
- Environment variable loading
- Configuration validation
- Feature flags

### Logger
- Structured logging (JSON format)
- Log levels (DEBUG, INFO, WARN, ERROR)
- Request/response logging

### HTTP Server
- HTTP server setup
- Middleware chain
- Route registration
- TLS/HTTPS support

### Errors
- Common error types
- Error wrapping
- HTTP status mapping
- Error responses

### Validator
- Input validation
- Sanitization
- Business rule validation

## Dependency Flow

```
HTTP Request
    ↓
Input Adapter (HTTP Handler)
    ↓
Input Port (Service Interface)
    ↓
Application Service
    ↓
Domain Logic
    ↓
Output Port (Repository Interface)
    ↓
Output Adapter (Database Repository)
    ↓
Database
```

## Module Communication

Modules communicate through their **exposed service interfaces** (input ports):

- **Direct**: Module A imports Module B's service interface
- **Event-Driven**: Modules publish/subscribe to domain events (future enhancement)
- **NO direct imports** of internal implementation details

### Example

```go
// post module needs to verify user exists
// BAD: import "forum/internal/modules/user/adapters"
// GOOD: import "forum/internal/modules/user/ports"

type PostService struct {
    userService ports.UserService // Depend on interface
}
```

## Dependency Injection

All modules are wired together at the application bootstrap level (`cmd/forum/main.go`):

1. Initialize platform services (database, logger, config)
2. Create repositories (output adapters)
3. Create services (application layer) with injected dependencies
4. Create HTTP handlers (input adapters) with injected services
5. Register routes and start server

## Testing Strategy

### Unit Tests
- Domain logic (pure functions, entities)
- Application services (mock repositories)
- Located in each module's directory

### Integration Tests
- HTTP handlers with real database
- End-to-end workflows
- Located in `tests/integration/`

### Test Structure
```
module/
├── domain/
│   └── entity_test.go
├── application/
│   └── service_test.go
└── adapters/
    └── http_handler_test.go
```

## Database Design

### Migration Strategy
- SQL migrations in `migrations/` directory
- Numbered sequentially (001_, 002_, etc.)
- Organized by module
- Run at application startup

### Key Tables
- **sessions**: User sessions (auth module)
- **users**: User accounts (user module)
- **roles**: User roles (user module)
- **posts**: Forum posts (post module)
- **categories**: Post categories (post module)
- **post_categories**: Many-to-many relationship (post module)
- **comments**: Post comments (comment module)
- **reactions**: Likes/dislikes (reaction module)
- **reports**: Moderation reports (moderation module - optional)
- **notifications**: User notifications (notification module - optional)

## Security Considerations

1. **Authentication**: bcrypt password hashing, secure session management
2. **Authorization**: Role-based access control (RBAC)
3. **Input Validation**: Sanitize and validate all user input
4. **SQL Injection**: Parameterized queries only
5. **XSS Prevention**: HTML template escaping
6. **CSRF Protection**: CSRF tokens for state-changing operations
7. **Rate Limiting**: Prevent abuse (platform/httpserver)
8. **HTTPS/TLS**: Encrypt data in transit

## AI Agent Optimization

This architecture is optimized for AI agent collaboration:

### Clear Module Boundaries
- Each module has a single, well-defined responsibility
- Easy to understand scope and context

### Consistent Patterns
- Every module follows the same structure
- Predictable file locations and naming

### Explicit Dependencies
- Ports clearly define contracts
- No hidden dependencies or magic

### Self-Documenting Code
- File names and structure explain purpose
- Comments explain "why", code explains "how"

### Small, Focused Files
- Single Responsibility Principle applied to files
- Easy to understand and modify

### Type-Safe Interfaces
- Leverage Go's type system
- Compile-time guarantees

## Future Enhancements

1. **Event-Driven Architecture**: Introduce domain events for loose coupling
2. **CQRS**: Separate read and write models for complex queries
3. **Caching**: Redis for session storage and query caching
4. **Message Queue**: Background job processing
5. **Observability**: Metrics, tracing, and monitoring
6. **API Versioning**: Support multiple API versions

## References

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

---

**Last Updated**: November 3, 2025
