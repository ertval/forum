# Forum Project Structure

## Overview

This document describes the complete project structure of the Forum application. The project follows a **Modular Monolith** architecture with **Hexagonal Architecture** (Ports and Adapters) for each module.

## Complete Directory Tree

```
forum/
├── cmd/forum/main.go                         # Application entry point & DI
├── internal/
│   ├── modules/                              # Business modules
│   │   ├── auth/                            # [CORE] Authentication
│   │   │   ├── domain/
│   │   │   │   ├── session.go
│   │   │   │   └── errors.go
│   │   │   ├── ports/
│   │   │   │   ├── service.go              # INPUT PORT
│   │   │   │   └── repository.go           # OUTPUT PORT
│   │   │   ├── application/
│   │   │   │   └── service.go
│   │   │   └── adapters/
│   │   │       ├── http_handler.go         # INPUT ADAPTER
│   │   │       ├── sqlite_session_repository.go   # OUTPUT ADAPTER
│   │   │       └── sqlite_user_repository.go      # OUTPUT ADAPTER
│   │   ├── user/                            # [CORE] User management
│   │   ├── post/                            # [CORE] Posts & categories
│   │   ├── comment/                         # [CORE] Comments
│   │   ├── reaction/                        # [CORE] Likes/dislikes
│   │   ├── moderation/                      # [OPTIONAL: forum-moderation]
│   │   └── notification/                    # [OPTIONAL: forum-advanced-features]
│   └── platform/                            # Shared infrastructure
│       ├── database/
│       ├── config/
│       ├── logger/
│       ├── httpserver/
│       ├── errors/
│       └── validator/
├── migrations/                               # SQL migrations
├── static/                                   # Static assets
│   ├── css/style.css
│   ├── js/app.js
│   └── uploads/
├── templates/                                # HTML templates
├── tests/
│   ├── integration/
│   └── unit/
├── docs/
│   ├── ARCHITECTURE.md
│   ├── PROJECT_STRUCTURE.md
│   ├── IMPLEMENTATION_ROADMAP.md
│   └── issues_tracker.md
├── ARCHITECTURE.md
├── README.md
├── LICENSE
├── .gitignore
├── go.mod
├── Dockerfile
└── docker-compose.yml
```

## Module Pattern (Flattened Hexagonal)

Every module has EXACTLY 4 directories:

```
module/
├── domain/          # Entities, business rules, errors
├── ports/           # service.go (INPUT), repository.go (OUTPUT)
├── application/     # service.go implementation
└── adapters/        # http_handler.go (INPUT), sqlite_repository.go (OUTPUT)
```

## File Type Annotations

- `// INPUT PORT - Service Interface` - Use case definitions
- `// OUTPUT PORT - Repository Interface` - Data access contracts
- `// INPUT ADAPTER - HTTP Handler` - HTTP request handlers
- `// OUTPUT ADAPTER - SQLite Repository` - Database implementations

## Module Categories

### Core Modules (Required)
1. auth - Authentication & sessions
2. user - User management & roles
3. post - Posts & categories
4. comment - Comments
5. reaction - Likes & dislikes

### Optional Modules (Extra Features)
6. moderation - [OPTIONAL: forum-moderation]
7. notification - [OPTIONAL: forum-advanced-features]

---

**Last Updated**: November 3, 2025
