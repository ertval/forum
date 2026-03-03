# Modules Layer (Bounded Contexts)

This directory contains the feature modules of the application following the **Hexagonal Architecture (Ports and Adapters)** pattern within a **Modular Monolith**.

## Bounded Contexts & Responsibilities

Each module encapsulates a specific domain, owning its data and exposing an interface for other modules to interact with it.

| Bounded Context | Aggregate Root | Main Entity | Responsibilities |
|-----------------|----------------|--------------|------------------|
| **Auth** | Session | `Session`, `Credentials` | Authentication, session lifecycle, credential validation, active tokens |
| **User** | User | `User`, `UserStats` | User identity, profile, cached statistics (post/comment counts) |
| **Post** | Post | `Post`, `Category` | Content creation, categorization, complex filtering, image handling |
| **Comment** | Comment | `Comment` | Threaded discussions, comment CRUD, author context |
| **Reaction** | Reaction | `Reaction` | Like/dislike tracking for posts and comments, reaction counts |
| **Moderation** | Report | `Report`, `Action` | Content reporting by users, moderation actions by admins |
| **Notification** | Notification | `Notification` | User notifications, read/unread state |

## Module Structure

Every module MUST follow this exact 4-directory layout internally:

```text
internal/modules/{module}/
├── domain/           # Pure business logic (stdlib only)
│   ├── entity.go     # Domain entities with Validate() methods
│   └── errors.go     # Domain-specific errors (e.g., ErrNotFound)
│
├── ports/            # Interface definitions
│   ├── service.go    # INPUT PORT - Use case interface
│   └── repository.go # OUTPUT PORT - Data access contract
│
├── application/      # Business logic implementation
│   └── service.go    # Implements ports/service.go, orchestrates domain
│
└── adapters/         # Technical implementations (FLAT directory)
    ├── http_handler.go       # Base handler, ServiceContainer interface, RegisterRoutes
    ├── http_handler_api.go   # JSON API handlers (/api/...)
    ├── http_handler_page.go  # HTML page handlers
    └── sqlite_repository.go  # Database access/persistence 
```

## Anti-Pattern Rules (Crucial)

1. **NO Cross-Module Adapter Imports**: Modules MUST only communicate via `ports/service.go` interfaces. Do NOT import an `application` or `adapters` package from one module into another.
2. **NO Domain External Deps**: The `domain/` directory must only import from the Go standard library (`stdlib`).
3. **NO Sequential IDs Outward**: Expose ONLY `PublicID` (UUID) in JSON responses and templates. Internal integer `ID`s are only stringed internally.

For dependency injection details across modules, see `cmd/forum/wire/README.md`.
