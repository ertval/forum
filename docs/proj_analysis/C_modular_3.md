# Modular Monolith Architecture Proposal

## Overview
A modular monolith organized by business capabilities (modules), where each module is independently maintainable yet deployed as a single application. Follows Go idioms, Clean Architecture, and SOLID principles.

## Core Principles
- **Simplicity**: Each module solves one domain problem
- **Readability**: Clear package names reflecting business intent
- **Explicitness**: Dependencies declared via interfaces
- **Minimalism**: Build what's needed, nothing more
- **Composition**: Small, composable functions over complex hierarchies

---

## Module Structure

```
forum/
├── cmd/
│   └── forum/
│       └── main.go                    # Wire modules, start server
│
├── internal/
│   ├── shared/                        # Cross-cutting concerns (kernel)
│   │   ├── database/
│   │   │   ├── db.go                  # DB connection pool
│   │   │   ├── migrator.go            # Migration runner
│   │   │   └── tx.go                  # Transaction helper
│   │   ├── errors/
│   │   │   └── errors.go              # Domain error types
│   │   ├── logger/
│   │   │   └── logger.go              # Structured logging interface
│   │   ├── validator/
│   │   │   └── validator.go           # Input validation helpers
│   │   └── pagination/
│   │       └── pagination.go          # Pagination types
│   │
│   ├── modules/
│   │   │
│   │   ├── user/                      # User & Authentication Module
│   │   │   ├── domain/
│   │   │   │   ├── user.go            # User entity + business rules
│   │   │   │   ├── session.go         # Session entity
│   │   │   │   └── errors.go          # Module-specific errors
│   │   │   ├── ports/
│   │   │   │   ├── repository.go      # UserRepository interface
│   │   │   │   └── session_store.go   # SessionStore interface
│   │   │   ├── application/
│   │   │   │   ├── service.go         # User orchestration logic
│   │   │   │   ├── auth_service.go    # Register, Login, Logout
│   │   │   │   └── commands.go        # DTOs (RegisterUser, LoginUser)
│   │   │   ├── adapters/
│   │   │   │   ├── sqlite_repo.go     # SQL implementation
│   │   │   │   ├── session_store.go   # Cookie-based session store
│   │   │   │   └── oauth_adapter.go   # Google/GitHub OAuth (future)
│   │   │   ├── handlers/
│   │   │   │   ├── http.go            # HTTP handlers (thin)
│   │   │   │   └── routes.go          # Route registration
│   │   │   └── module.go              # Module constructor + wiring
│   │   │
│   │   ├── post/                      # Post Module
│   │   │   ├── domain/
│   │   │   │   ├── post.go            # Post entity + invariants
│   │   │   │   ├── category.go        # Category value object
│   │   │   │   └── errors.go
│   │   │   ├── ports/
│   │   │   │   ├── repository.go      # PostRepository interface
│   │   │   │   └── event_publisher.go # EventPublisher interface (future)
│   │   │   ├── application/
│   │   │   │   ├── service.go         # Post CRUD + filtering
│   │   │   │   ├── commands.go        # CreatePost, UpdatePost, DeletePost
│   │   │   │   └── queries.go         # ListPosts, GetPostByID
│   │   │   ├── adapters/
│   │   │   │   ├── sqlite_repo.go
│   │   │   │   └── image_store.go     # Image upload/storage (future)
│   │   │   ├── handlers/
│   │   │   │   ├── http.go
│   │   │   │   └── routes.go
│   │   │   └── module.go
│   │   │
│   │   ├── comment/                   # Comment Module
│   │   │   ├── domain/
│   │   │   │   ├── comment.go         # Comment entity
│   │   │   │   └── errors.go
│   │   │   ├── ports/
│   │   │   │   └── repository.go
│   │   │   ├── application/
│   │   │   │   ├── service.go
│   │   │   │   └── commands.go
│   │   │   ├── adapters/
│   │   │   │   └── sqlite_repo.go
│   │   │   ├── handlers/
│   │   │   │   ├── http.go
│   │   │   │   └── routes.go
│   │   │   └── module.go
│   │   │
│   │   ├── reaction/                  # Reaction (Like/Dislike) Module
│   │   │   ├── domain/
│   │   │   │   ├── reaction.go        # Reaction entity (type, target)
│   │   │   │   └── errors.go
│   │   │   ├── ports/
│   │   │   │   └── repository.go
│   │   │   ├── application/
│   │   │   │   ├── service.go
│   │   │   │   └── commands.go
│   │   │   ├── adapters/
│   │   │   │   └── sqlite_repo.go
│   │   │   ├── handlers/
│   │   │   │   ├── http.go
│   │   │   │   └── routes.go
│   │   │   └── module.go
│   │   │
│   │   ├── moderation/                # Moderation Module (future)
│   │   │   ├── domain/
│   │   │   │   ├── report.go          # Report entity
│   │   │   │   ├── role.go            # Role enum (Guest, User, Mod, Admin)
│   │   │   │   └── permission.go      # Permission checks
│   │   │   ├── ports/
│   │   │   │   ├── repository.go
│   │   │   │   └── notifier.go        # Notification interface
│   │   │   ├── application/
│   │   │   │   ├── service.go
│   │   │   │   └── commands.go
│   │   │   ├── adapters/
│   │   │   │   └── sqlite_repo.go
│   │   │   ├── handlers/
│   │   │   │   ├── http.go
│   │   │   │   └── routes.go
│   │   │   └── module.go
│   │   │
│   │   ├── notification/              # Notification Module (future)
│   │   │   ├── domain/
│   │   │   │   ├── notification.go
│   │   │   │   └── event.go
│   │   │   ├── ports/
│   │   │   │   ├── repository.go
│   │   │   │   └── delivery.go        # Email, WebSocket, etc.
│   │   │   ├── application/
│   │   │   │   └── service.go
│   │   │   ├── adapters/
│   │   │   │   ├── sqlite_repo.go
│   │   │   │   └── ws_delivery.go     # WebSocket delivery
│   │   │   ├── handlers/
│   │   │   │   └── http.go
│   │   │   └── module.go
│   │   │
│   │   └── security/                  # Security Module (future)
│   │       ├── domain/
│   │       │   ├── rate_limit.go      # Rate limiting logic
│   │       │   └── cipher.go          # Cipher suite config
│   │       ├── application/
│   │       │   └── service.go
│   │       ├── adapters/
│   │       │   ├── rate_limiter.go    # In-memory/Redis limiter
│   │       │   └── tls_config.go      # HTTPS setup
│   │       └── module.go
│   │
│   ├── web/                           # Web Presentation Layer
│   │   ├── middleware/
│   │   │   ├── auth.go                # Authentication check
│   │   │   ├── session.go             # Session injection
│   │   │   ├── ratelimit.go           # Rate limiting middleware
│   │   │   ├── logger.go              # Request logging
│   │   │   └── recovery.go            # Panic recovery
│   │   ├── templates/
│   │   │   ├── layout/
│   │   │   │   ├── base.html
│   │   │   │   └── nav.html
│   │   │   ├── user/
│   │   │   │   ├── login.html
│   │   │   │   ├── register.html
│   │   │   │   └── profile.html
│   │   │   ├── post/
│   │   │   │   ├── list.html
│   │   │   │   ├── view.html
│   │   │   │   └── create.html
│   │   │   └── error/
│   │   │       ├── 404.html
│   │   │       └── 500.html
│   │   └── renderer.go                # Template renderer helper
│   │
│   └── server/
│       ├── server.go                  # HTTP server lifecycle
│       └── router.go                  # Global route registration
│
├── static/
│   ├── css/
│   │   └── style.css
│   └── js/
│       └── main.js
│
├── migrations/                        # SQL migrations (flat)
│   ├── 001_initial_schema.sql
│   ├── 002_add_roles.sql
│   └── 003_add_notifications.sql
│
├── tests/
│   ├── integration/
│   │   ├── user_test.go               # End-to-end module tests
│   │   ├── post_test.go
│   │   └── helpers.go
│   └── unit/
│       ├── user/
│       │   └── service_test.go        # Application layer unit tests
│       └── post/
│           └── service_test.go
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## Module Anatomy

Each module is self-contained with clear boundaries:

### 1. **Domain Layer** (`domain/`)
- Pure business logic, zero dependencies on infrastructure
- Entities with behavior, value objects, domain errors
- **Example**: `user.Validate()`, `post.AddCategory()`

### 2. **Ports Layer** (`ports/`)
- Interfaces consumed by application layer
- Repository, external service contracts
- **Example**: `type UserRepository interface { Save(ctx, *User) error }`

### 3. **Application Layer** (`application/`)
- Use case orchestration (commands, queries)
- Depends only on `domain` and `ports`
- **Example**: `RegisterUser(cmd RegisterUserCmd) (*User, error)`

### 4. **Adapters Layer** (`adapters/`)
- Concrete implementations of ports
- SQL repos, HTTP clients, file storage
- **Example**: `type SQLiteUserRepo struct { db *sql.DB }`

### 5. **Handlers Layer** (`handlers/`)
- Thin HTTP handlers: parse request → call application → render response
- No business logic
- **Example**: `func (h *Handler) Register(w http.ResponseWriter, r *http.Request)`

### 6. **Module Constructor** (`module.go`)
- Wire dependencies (DI without frameworks)
- Return handler + service for registration in main
```go
func NewUserModule(db *sql.DB) *Module {
    repo := adapters.NewSQLiteUserRepo(db)
    sessionStore := adapters.NewCookieSessionStore()
    svc := application.NewAuthService(repo, sessionStore)
    handler := handlers.NewHandler(svc)
    return &Module{Handler: handler, Service: svc}
}
```

---

## Cross-Module Communication

### Rule: **No Direct Dependencies Between Modules**

Modules communicate via:

1. **Shared Database** (read-only joins acceptable for queries)
2. **Service Interfaces** (injected via ports)

   ```go
   // In post/application layer
   type UserService interface {
       GetUserByID(ctx, userID int) (*User, error)
   }
   ```

3. **Domain Events** (future: async via channels/queues)

   ```go
   // Post liked → Notification module listens
   event := NewPostLikedEvent(postID, userID)
   eventBus.Publish(event)
   ```

---

## Dependency Flow

```
Handlers → Application → Domain
    ↓          ↓
Adapters ← Ports (interfaces)
```

- **Inward**: `handlers` depend on `application`, `application` depends on `domain`
- **Outward**: `application` depends on `ports` interfaces; `adapters` implement them
- **Result**: Easy to swap DB, add caching, mock for tests

---

## Growth Roadmap

### Phase 1: Core (Current)

- `user` (auth, sessions)
- `post` (CRUD, categories, filtering)
- `comment` (CRUD)
- `reaction` (like/dislike)

### Phase 2: Security

- `security` module: HTTPS, rate limiting, UUID sessions
- Encrypt passwords (bcrypt) in `user/domain`

### Phase 3: Moderation

- `moderation` module: roles (Guest, User, Mod, Admin), reports
- Add `role` field to `user/domain/user.go`
- Permission checks in `moderation/domain/permission.go`

### Phase 4: Media

- Extend `post` module: add image upload adapter
- Store images in `/uploads` or object storage
- Validate file size/type in `post/domain/post.go`

### Phase 5: Social Auth

- Extend `user` module: OAuth adapters (Google, GitHub)
- Add `oauth_adapter.go` in `user/adapters/`

### Phase 6: Advanced Features

- `notification` module: WebSocket delivery, real-time events
- Extend `user` module: activity tracking queries
- Edit/delete functionality in post/comment application services

---

## Key Design Decisions

### 1. **Shared Database, Separate Schemas**

- Single SQLite file, logical separation via table prefixes
- `users`, `user_sessions` (user module)
- `posts`, `post_categories`, `categories` (post module)
- `comments` (comment module)
- `reactions` (reaction module)

### 2. **No Shared Models**

- Each module defines its own domain entities
- Cross-module reads via service interfaces or read models

### 3. **Middleware as Cross-Cutting Concerns**

- Authentication, logging, rate limiting live in `web/middleware/`
- Injected globally in `server/router.go`

### 4. **Template Organization**

- Templates grouped by module in `web/templates/{module}/`
- Shared layouts in `web/templates/layout/`

### 5. **Migrations**

- Flat directory: `migrations/001_xxx.sql`
- Executed sequentially by `shared/database/migrator.go`
- Each module contributes its schema changes

### 6. **Testing Strategy**

- **Unit tests**: Application layer (mock ports)
- **Integration tests**: Full stack with in-memory SQLite
- Tests live in `tests/unit/{module}/` and `tests/integration/`

---

## Wiring in `main.go`

```go
func main() {
    // 1. Init shared resources
    db := shared.InitDB("forum.db")
    logger := shared.NewLogger()

    // 2. Wire modules
    userModule := user.NewUserModule(db, logger)
    postModule := post.NewPostModule(db, logger, userModule.Service)
    commentModule := comment.NewCommentModule(db, logger, postModule.Service)
    reactionModule := reaction.NewReactionModule(db, logger)

    // 3. Setup router
    router := server.NewRouter(
        userModule.Handler,
        postModule.Handler,
        commentModule.Handler,
        reactionModule.Handler,
    )

    // 4. Start server
    srv := server.NewServer(":8080", router, logger)
    srv.Start()
}
```

---

## Benefits of This Architecture

1. **Independent Development**: Two devs can work on different modules without conflicts
2. **Testability**: Mock ports, test application logic in isolation
3. **Scalability**: Easy to extract module to microservice later (just change adapters)
4. **Readability**: Business logic in `domain`, infrastructure in `adapters`
5. **Maintainability**: Change DB? Swap adapter. Add cache? Decorate repository.
6. **Go Idiomatic**: Small interfaces, composition, explicit dependencies

---

## Migration Path from Current Structure

1. **Step 1**: Move `internal/models/` → `internal/modules/{module}/domain/`
2. **Step 2**: Extract repository methods from models → `adapters/sqlite_repo.go`
3. **Step 3**: Create `application/service.go` for each module (move handler logic here)
4. **Step 4**: Define `ports/repository.go` interfaces
5. **Step 5**: Update handlers to call application services
6. **Step 6**: Create `module.go` constructors
7. **Step 7**: Refactor `main.go` to wire modules

---

## Conclusion

This modular monolith balances simplicity with structure. Each module is a mini-application following Clean Architecture. As requirements grow (security, moderation, notifications), add modules without touching existing code. Perfect for two AI-enhanced developers: clear boundaries, explicit contracts, minimal cognitive load.
