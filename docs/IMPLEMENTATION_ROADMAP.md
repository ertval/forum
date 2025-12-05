# Implementation Roadmap

Fast path to functional forum MVP following core requirements, then complete remaining features, finally add bonus features.

## Current Status

**Project Phase**: Production-ready MVP — core features complete, all tests pass.

### ✅ Completed Features
- **Platform layer**: config, database (SQLite), logger, httpserver, errors, validator, upload
- **Auth module**: Registration, login, sessions (one per user), logout, session validation
- **User module**: Domain, repository, stats (post/comment counts cached in users table)
- **Post module**: Full CRUD, categories, filtering (category, user, liked posts, date range)
- **Comment module**: Full CRUD with ownership validation, pagination for "My Comments" page
- **Image upload**: PNG/JPEG/GIF support, 20MB limit, validation, persistence
- **Filtering**: By category, My Posts, Liked Posts, date range (today/week/month/all)
- **UI Enhancements**: Hover effects on reaction buttons, "Show More" pagination for comments

### ⚠️ Scaffolded (Needs Implementation)
- **Reaction module**: Routes defined, but handlers return 501 Not Implemented
- **Moderation module**: Domain/ports/adapters structure exists, minimal implementation
- **Notification module**: Domain/ports/adapters structure exists, minimal implementation

### 🧪 Test Status
- ✅ All Go unit tests pass
- ✅ All integration tests pass
- ✅ All E2E test scripts pass (API, audit, image upload, pages)

---

## PART 1: MVP - CORE REQUIREMENTS ✅ COMPLETE

### Phase 1: Platform Basics ✅
- [x] Config loading from environment variables
- [x] Database connection (SQLite with mattn/go-sqlite3)
- [x] Database migrator - auto-apply migrations on startup
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

## PART 2: REMAINING FEATURES

### Phase 8: Reactions ⚠️ SCAFFOLDED
**Status**: Routes defined, handlers return 501 Not Implemented

**Remaining Work**:
- [ ] Implement React service method (toggle logic)
- [ ] Implement CountReactions service method
- [ ] Implement AddReactionAPI handler
- [ ] Implement RemoveReactionAPI handler
- [ ] Implement GetReactionsAPI handler
- [ ] Implement CountReactionsAPI handler
- [ ] Add reaction buttons to post detail template
- [ ] Add reaction buttons to comments

**Time Estimate**: 1-2 days

### Phase 9: Testing & Polish
- [x] Domain layer tests
- [x] Application service tests
- [x] Repository tests
- [x] HTTP handler tests
- [x] Integration tests
- [x] E2E test scripts
- [ ] Additional edge case coverage
- [ ] Performance optimization

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
**Status**: Scaffolded, minimal implementation

- [ ] Notification entity and repository
- [ ] Notification on comment reply
- [ ] Notification on post reaction
- [ ] Mark as read functionality
- [ ] Notification list endpoint

### Phase 12: Security Hardening ✅
- [x] HTTPS/TLS configuration (TLS 1.2+, strong cipher suites)
- [x] Rate limiting middleware (per-IP/per-user)
- [x] Security headers (CSP, HSTS, X-Frame-Options, X-XSS-Protection, Referrer-Policy)
- [x] Certificate generation script (`scripts/generate_certs.sh`)
- [x] Security headers tests

---

## Module Status Summary

| Module | Domain | Ports | Application | Adapters | Tests | Status |
|--------|--------|-------|-------------|----------|-------|--------|
| auth | ✅ | ✅ | ✅ | ✅ | ✅ | Complete |
| user | ✅ | ✅ | ✅ | ✅ | ✅ | Complete |
| post | ✅ | ✅ | ✅ | ✅ | ✅ | Complete |
| comment | ✅ | ✅ | ✅ | ✅ | ✅ | Complete |
| reaction | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | Scaffolded |
| moderation | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | Scaffolded |
| notification | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | Scaffolded |

---

## Technical Debt & Known Issues

1. **Reaction module incomplete**: Handlers return 501, needs full implementation
2. **CSRF protection**: Should add CSRF tokens for state-changing operations (optional enhancement)

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
docker-compose up --build
```

---

## File Locations

| Purpose | Location |
|---------|----------|
| Entry point | `cmd/forum/main.go` |
| DI wiring | `cmd/forum/wire/` |
| Modules | `internal/modules/{module}/` |
| Platform | `internal/platform/` |
| Migrations | `migrations/*.sql` |
| Templates | `templates/*.html` |
| Static assets | `static/` |
| Tests | `tests/`, `scripts/tests/` |
