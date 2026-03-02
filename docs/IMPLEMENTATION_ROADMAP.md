# Implementation Roadmap

Fast path to functional forum MVP following core requirements, then complete remaining features, finally add bonus features.

## Current Status

**Project Phase**: Production-ready MVP — core features complete, optional modules scaffolded, and test coverage is broad with occasional regressions under active maintenance.

### ✅ Completed Features

- **Platform layer**: config, database (SQLite), logger, httpserver, errors, validator, upload
- **Shared module adapters**: `internal/modules/shared/adapters/httpjson` for JSON handler utilities used across module adapters
- **Auth module**: Registration, login, sessions (one per user), logout, session validation
- **User module**: Domain, repository, stats (post/comment counts cached in users table)
- **Post module**: Full CRUD, categories, filtering (category, user, liked posts, date range)
- **Comment module**: Full CRUD with ownership validation, pagination for "My Comments" page
- **Reaction module**: Full implementation for posts and comments (like/dislike toggles)
- **Activity**: Unified `/activity` page and `/api/activity` endpoint for created posts, likes/dislikes, and comments with post context
- **Notification module**: End-to-end notifications for post owner on comment/like/dislike with list/read API
- **Image upload**: PNG/JPEG/GIF support, 20MB limit, validation, persistence
- **Filtering**: By category, My Posts, Liked Posts, date range (today/week/month/all)
- **UI Enhancements**: Hover effects on reaction buttons, "Show More" pagination for comments
- **Settings page**: Protected `GET /settings` page added and linked from existing navigation
- **Docker reliability**: `make up/down` supports modern `docker compose` and container startup now normalizes mounted volume permissions before dropping privileges

### ⚠️ Scaffolded (Needs Implementation)

- **Moderation module**: Domain/ports/adapters structure exists, minimal implementation

### 🧪 Test Status

- ✅ Broad unit/integration/E2E coverage exists across core modules
- ✅ `test_audit_advanced.sh` passes
- ⚠️ Pending optional audits: moderation and authentication (OAuth extensions)
- ℹ️ Use `make test` to verify the current repository state on your environment

---

## PART 1: MVP - CORE REQUIREMENTS ✅ COMPLETE

### Phase 1: Platform Basics ✅

- [x] Config loading from environment variables
- [x] Database connection (SQLite with mattn/go-sqlite3)
- [x] Database migrator - auto-apply migrations on startup
- [x] Manual migration script (`make migrate` / `scripts/seed/run_migrations.sh`)
- [x] HTTP server with standard lib http.ServeMux
- [x] Structured logger with levels
- [x] Logger middleware (request logging)
- [x] Error responses with HTTP status mapping
- [x] Input validator (email, password)

### Phase 2: Authentication ✅

- [x] Session entity with validation
- [x] SQLite session repository (create, get, delete, delete expired)
- [x] SQLite user repository (create, get by email, exists checks)
- [x] Register use case (email validation, bcrypt hashing)
- [x] Login use case (password verification, single session per user)
- [x] Logout use case
- [x] ValidateSession use case
- [x] HTTP handlers (register, login, logout, session)
- [x] RequireAuth and OptionalAuth middleware

### Phase 3: Posts & Categories ✅

- [x] Post entity with validation (title max 300, content max 50000)
- [x] Category entity with validation
- [x] SQLite post repository (full CRUD with categories)
- [x] SQLite category repository (CRUD + exists checks)
- [x] Post service (create, get, list, update, delete)
- [x] FilterService for complex post filtering
- [x] HTTP handlers (API + page handlers)
- [x] Templates (home, board, post_create, post_detail, post_edit)

### Phase 4: Comments ✅

- [x] Comment entity with validation
- [x] SQLite comment repository
- [x] Comment service (create, get, update, delete, list by post/user)
- [x] HTTP handlers (API handlers for all operations)
- [x] User comment count tracking (async updates)

### Phase 5: Filtering ✅

- [x] Filter by category
- [x] Filter by user (My Posts)
- [x] Filter by liked posts
- [x] Filter by date range (Today, This Week, This Month, All Time)
- [x] FilterService with BuildFilter and ApplyDateFilter
- [x] Filter state preservation in UI

### Phase 6: Image Upload ✅ (Bonus - Implemented Early)

- [x] Magic bytes validation (PNG, JPEG, GIF only)
- [x] Size validation (max 20MB)
- [x] Secure filename generation (UUID-based)
- [x] Storage in static/uploads/
- [x] Image deletion on post delete
- [x] UpdatePostImage service method
- [x] E2E tests for all image scenarios

### Phase 7: Docker ✅

- [x] Dockerfile with multi-stage build (CGO_ENABLED=1)
- [x] docker-compose.yml for deployment
- [x] Proper SQLite support in container

---

## PART 2: REMAINING FEATURES ✅ COMPLETE

### Phase 8: Reactions ✅ COMPLETE

**Status**: Fully implemented and tested.

- [x] Implement React service method (toggle logic)
- [x] Implement CountReactions service method
- [x] Implement AddReactionAPI handler
- [x] Implement RemoveReactionAPI handler
- [x] Implement GetReactionsAPI handler
- [x] Implement CountReactionsAPI handler
- [x] Add reaction buttons to post detail template
- [x] Add reaction buttons to comments

### Phase 9: Testing & Polish ✅ COMPLETE

- [x] Domain layer tests
- [x] Application service tests
- [x] Repository tests
- [x] HTTP handler tests
- [x] Integration tests
- [x] E2E test scripts
- [x] Additional edge case coverage
- [x] Performance optimization

---

## PART 3: OPTIONAL MODULES

### Phase 10: Moderation [OPTIONAL]

**Status**: Scaffolded, minimal implementation

- [ ] Report entity and repository
- [ ] Moderator role checks
- [ ] Admin role checks
- [ ] Report creation/review workflow
- [ ] Content deletion by moderators

### Phase 11: Notifications [OPTIONAL]

**Status**: Complete for advanced objective requirements

- [x] Notification entity and repository
- [x] Notification on post comment
- [x] Notification on post like/dislike
- [x] Mark as read functionality
- [x] Notification list endpoint

### Phase 12: Security Hardening ✅

- [x] HTTPS/TLS configuration (TLS 1.2+, strong cipher suites)
- [x] Rate limiting middleware (per-IP/per-user)
- [x] Security headers (CSP, HSTS, X-Frame-Options, X-XSS-Protection, Referrer-Policy)
- [x] Certificate generation script (`scripts/seed/generate_certs.sh`)
- [x] UUID-format session cookie token generation (SEC-07)
- [x] UUID-only outward ID exposure in API responses for security-sensitive entities
- [x] Unified JSON API error schema via `platform/errors.WriteErrorJSON`
- [x] Security headers tests

---

## Module Status Summary

| Module       | Domain | Ports | Application | Adapters | Tests | Status     |
| ------------ | ------ | ----- | ----------- | -------- | ----- | ---------- |
| auth         | ✅     | ✅    | ✅          | ✅       | ✅    | Complete   |
| user         | ✅     | ✅    | ✅          | ✅       | ✅    | Complete   |
| post         | ✅     | ✅    | ✅          | ✅       | ✅    | Complete   |
| comment      | ✅     | ✅    | ✅          | ✅       | ✅    | Complete   |
| reaction     | ✅     | ✅    | ✅          | ✅       | ✅    | Complete   |
| moderation   | ✅     | ✅    | ⚠️          | ⚠️       | ⚠️    | Scaffolded |
| notification | ✅     | ✅    | ✅          | ✅       | ✅    | Complete   |

---

## Technical Debt & Known Issues

1. **CSRF protection**: Should add CSRF tokens for state-changing operations (optional enhancement)

---

## Commands Reference

```bash
# Run locally
make go

# Run all tests
make test

# Go tests only
make test-go

# E2E scripts only
make test-script

# Build binary
go build -o bin/forum cmd/forum/main.go

# Docker
docker compose up --build
```

---

## File Locations

| Purpose       | Location                     |
| ------------- | ---------------------------- |
| Entry point   | `cmd/forum/main.go`          |
| DI wiring     | `cmd/forum/wire/`            |
| Modules       | `internal/modules/{module}/` |
| Platform      | `internal/platform/`         |
| Migrations    | `migrations/*.sql`           |
| Templates     | `templates/*.html`           |
| Static assets | `static/`                    |
| Tests         | `tests/`, `scripts/tests/`   |
