# Implementation Roadmap

This document outlines the step-by-step implementation plan for the Forum project.

## Current Status Summary

**Project Phase**: Initial Scaffolding (10% Complete)
- ✅ Project structure created
- ✅ Module scaffolding complete
- ✅ Database migrations defined
- ✅ Dependency injection wiring in main.go
- ⚠️ Most implementations are placeholders with TODO comments

## Phase 1: Foundation (Platform Services) ⏳ IN PROGRESS

### 1.1 Configuration Management - NOT STARTED
- [ ] Implement `config.Load()` function (currently placeholder)
- [ ] Add environment variable parsing
- [ ] Create default configuration values
- [ ] Add validation for required config (currently placeholder)

**Status**: Skeleton exists in `internal/platform/config/config.go` with TODO markers

### 1.2 Database Layer - NOT STARTED
- [ ] Implement `database.New()` connection (currently placeholder)
- [ ] Add connection pooling configuration
- [ ] Implement `BeginTx()` for transactions (currently placeholder)
- [ ] Create migration runner logic (currently placeholder)
- [ ] Test database connection and migrations

**Status**: Skeleton exists with TODO markers in:
- `internal/platform/database/connection.go`
- `internal/platform/database/migrator.go`
- `internal/platform/database/transaction.go`

### 1.3 Logging - NOT STARTED
- [ ] Implement structured logger (currently placeholder)
- [ ] Add log levels (Debug, Info, Warn, Error, Fatal) - interfaces defined
- [ ] Add context-aware logging
- [ ] Configure log output format

**Status**: Interface defined in `internal/platform/logger/logger.go`, all methods have TODO placeholders

### 1.4 HTTP Server - NOT STARTED
- [ ] Implement server creation and startup (currently placeholder)
- [ ] Add graceful shutdown (currently placeholder)
- [ ] Configure TLS/HTTPS
- [ ] Implement middleware chain

**Status**: Skeleton exists in `internal/platform/httpserver/server.go` with TODO markers

### 1.5 Middleware - NOT STARTED
- [ ] Recovery middleware (panic handling) - placeholder
- [ ] Request logging middleware - placeholder
- [ ] CORS middleware - placeholder
- [ ] Rate limiting middleware - placeholder
- [ ] Session validation middleware - placeholder
- [ ] Authorization middleware - placeholder

**Status**: All middleware stubs exist in `internal/platform/httpserver/middleware.go` with TODO markers

### 1.6 Utilities - PARTIAL
- [ ] Input validation functions - some implemented
- [ ] Error handling utilities - error types defined
- [ ] Response helpers - not started
- [ ] Email validation - TODO marker
- [ ] Password strength validation - TODO marker
- [ ] HTML sanitization - TODO marker

**Status**: `internal/platform/validator/validator.go` has partial implementation with TODOs
**Status**: `internal/platform/errors/errors.go` has error types defined

**Estimated Time**: 2-3 days

---

## Phase 2: Authentication Module 🔐 NOT STARTED

### 2.1 Domain Layer - PARTIAL

- [x] Session entity defined
- [ ] Finalize session validation logic (has TODO marker)
- [ ] Validate business rules
- [ ] Add domain tests

**Status**: `internal/modules/auth/domain/session.go` has entity structure but validation is TODO

### 2.2 Repository (Output Adapter) - NOT STARTED

- [ ] Implement SQLite session repository (all methods are placeholders)
- [ ] Add session CRUD operations
- [ ] Implement session cleanup
- [ ] Add repository tests

**Status**: Skeleton in `internal/modules/auth/adapters/sqlite_session_repository.go` - ALL methods have TODO markers
**Status**: Skeleton in `internal/modules/auth/adapters/sqlite_user_repository.go` - ALL methods have TODO markers

### 2.3 Password Hasher - NOT STARTED

- [ ] Implement bcrypt password hashing (currently placeholder)
- [ ] Add cost configuration
- [ ] Add tests

**Status**: Helper functions in `internal/modules/auth/application/service.go` have TODO markers

### 2.4 Application Service - NOT STARTED

- [ ] Implement Register use case (currently placeholder)
- [ ] Implement Login use case (currently placeholder)
- [ ] Implement Logout use case (currently placeholder)
- [ ] Implement ValidateSession use case (currently placeholder)
- [ ] Implement RefreshSession use case (currently placeholder)
- [ ] Implement GetSession use case (currently placeholder)
- [ ] Add service tests

**Status**: `internal/modules/auth/application/service.go` - ALL use cases are TODO placeholders

### 2.5 HTTP Handlers - NOT STARTED

- [ ] Implement registration handler (currently placeholder)
- [ ] Implement login handler (currently placeholder)
- [ ] Implement logout handler (currently placeholder)
- [ ] Implement session retrieval handler (currently placeholder)
- [ ] Add cookie management
- [ ] Add request validation
- [ ] Add handler tests
- [ ] Implement route registration (currently placeholder)

**Status**: `internal/modules/auth/adapters/http_handler.go` - ALL handlers are TODO placeholders

### 2.6 Middleware - NOT STARTED

- [ ] Implement RequireAuth middleware (in platform middleware with TODO)
- [ ] Implement OptionalAuth middleware
- [ ] Add user context injection
- [ ] Add middleware tests

**Status**: Planned but not started

### 2.7 OAuth (Bonus) - NOT STARTED

- [ ] Implement Google OAuth provider
- [ ] Implement GitHub OAuth provider
- [ ] Add OAuth callback handler
- [ ] Add OAuth tests

**Status**: Not started

**Estimated Time**: 3-4 days

---

## Phase 3: User Module 👤 NOT STARTED

### 3.1 Repository - NOT STARTED

- [ ] Implement SQLite user repository (all methods are placeholders)
- [ ] Add user CRUD operations
- [ ] Add email/username existence checks
- [ ] Add repository tests

**Status**: `internal/modules/user/adapters/sqlite_repository.go` - ALL 10 methods have TODO placeholders

### 3.2 Application Service - NOT STARTED

- [ ] Implement GetByID
- [ ] Implement UpdateProfile
- [ ] Implement PromoteToModerator
- [ ] Implement DemoteFromModerator
- [ ] Add authorization checks
- [ ] Add service tests

**Status**: Service stub exists but no implementations

### 3.3 HTTP Handlers - NOT STARTED

- [ ] Profile view handler (currently placeholder)
- [ ] Profile edit handler
- [ ] User list handler (admin) (currently placeholder)
- [ ] Role management handlers (admin) (currently placeholder)
- [ ] User deactivation handler (currently placeholder)
- [ ] Route registration (currently placeholder)
- [ ] Add handler tests

**Status**: `internal/modules/user/adapters/http_handler.go` - ALL handlers are TODO placeholders

**Estimated Time**: 2 days

---

## Phase 4: Post Module 📝 PARTIAL

### 4.1 Repositories - PARTIAL

- [x] Post repository skeleton created
- [ ] Implement post repository methods (stubs exist, need implementation)
- [ ] Implement category repository (not yet created)
- [ ] Add post filtering logic
- [ ] Add repository tests

**Status**: `internal/modules/post/adapters/sqlite_repository.go` exists with stub methods
**Status**: Category repository not yet created - noted as TODO in main.go

### 4.2 Image Storage - NOT STARTED

- [ ] Implement image upload
- [ ] Add image validation (format, size)
- [ ] Add image storage logic
- [ ] Add image deletion
- [ ] Add tests

**Status**: Not started

### 4.3 Application Services - PARTIAL

- [x] PostService skeleton exists
- [ ] Implement PostService methods
- [ ] Implement CategoryService (not yet created)
- [ ] Add authorization checks
- [ ] Add service tests

**Status**: Service structure exists but implementations incomplete

### 4.4 HTTP Handlers - PARTIAL

- [x] HTTP handler skeleton exists
- [ ] Post creation handler (with image upload)
- [ ] Post view handler
- [ ] Post list handler
- [ ] Post edit handler
- [ ] Post delete handler
- [ ] Category management handlers
- [ ] Filter handlers
- [ ] Route registration (currently placeholder)
- [ ] Add handler tests

**Status**: `internal/modules/post/adapters/http_handler.go` has skeleton with stub methods

**Estimated Time**: 3-4 days

---

## Phase 5: Comment Module 💬 PARTIAL

### 5.1 Repository - PARTIAL

- [x] Comment repository skeleton created
- [ ] Implement comment repository methods (stubs exist)
- [ ] Add comment CRUD operations
- [ ] Add repository tests

**Status**: `internal/modules/comment/adapters/sqlite_repository.go` exists with stub methods

### 5.2 Application Service - PARTIAL

- [x] CommentService skeleton exists
- [ ] Implement CommentService methods
- [ ] Add authorization checks
- [ ] Add service tests

**Status**: Service structure exists but implementations incomplete

### 5.3 HTTP Handlers - PARTIAL

- [x] HTTP handler skeleton exists with TODO note
- [ ] Comment creation handler
- [ ] Comment list handler
- [ ] Comment edit handler
- [ ] Comment delete handler
- [ ] Route registration (has TODO: POST /comments, GET /comments/{id}, etc.)
- [ ] Add handler tests

**Status**: `internal/modules/comment/adapters/http_handler.go` has skeleton with TODO marker for routes

**Estimated Time**: 2 days

---

## Phase 6: Reaction Module 👍👎 PARTIAL

### 6.1 Repository - PARTIAL

- [x] Reaction repository skeleton created
- [ ] Implement reaction repository methods (stubs exist)
- [ ] Add reaction toggle logic
- [ ] Add reaction counting
- [ ] Add liked posts retrieval
- [ ] Add repository tests

**Status**: `internal/modules/reaction/adapters/sqlite_repository.go` exists with stub methods

### 6.2 Application Service - PARTIAL

- [x] ReactionService skeleton exists
- [ ] Implement ReactTo method (returns nil with TODO comment)
- [ ] Implement GetReactionCounts method (returns 0, 0, nil with TODO comment)
- [ ] Add authorization checks
- [ ] Add service tests

**Status**: `internal/modules/reaction/application/service.go` has methods returning TODO placeholders

### 6.3 HTTP Handlers - PARTIAL

- [x] HTTP handler skeleton exists with TODO note
- [ ] Add/toggle reaction handler
- [ ] Remove reaction handler
- [ ] Get reaction counts handler
- [ ] Route registration (has TODO: POST /reactions, DELETE /reactions, GET /reactions)
- [ ] Add handler tests

**Status**: `internal/modules/reaction/adapters/http_handler.go` has skeleton with TODO marker for routes

### 6.4 Integration - NOT STARTED

- [ ] Integrate with Post module
- [ ] Integrate with Comment module
- [ ] Trigger notifications on reactions

**Status**: Not started

**Estimated Time**: 2 days

---

## Phase 7: Notification Module 🔔 [OPTIONAL] PARTIAL

### 7.1 Repository - PARTIAL

- [x] Notification repository skeleton created
- [ ] Implement notification repository methods (stubs exist)
- [ ] Add notification CRUD operations
- [ ] Add unread count query
- [ ] Add repository tests

**Status**: `internal/modules/notification/adapters/sqlite_repository.go` exists with stub methods

### 7.2 Application Service - PARTIAL

- [x] NotificationService skeleton exists
- [ ] Implement notification creation logic (currently returns nil with TODO)
- [ ] Implement GetUserNotifications method (currently returns nil with TODO)
- [ ] Add service tests

**Status**: `internal/modules/notification/application/service.go` has stub methods with TODO markers

### 7.3 HTTP Handlers - PARTIAL

- [x] HTTP handler skeleton exists
- [ ] Notification list handler
- [ ] Mark as read handler
- [ ] Mark all as read handler
- [ ] Notification count handler
- [ ] Route registration
- [ ] Add handler tests

**Status**: `internal/modules/notification/adapters/http_handler.go` exists with basic structure

### 7.4 Integration - NOT STARTED

- [ ] Trigger notifications on post reactions
- [ ] Trigger notifications on comment reactions
- [ ] Trigger notifications on new comments

**Status**: Not started

**Estimated Time**: 2 days

---

## Phase 8: Moderation Module 🛡️ [OPTIONAL] PARTIAL

### 8.1 Repository - PARTIAL

- [x] Report repository skeleton created
- [ ] Implement report repository methods (stubs exist)
- [ ] Add report CRUD operations
- [ ] Add repository tests

**Status**: `internal/modules/moderation/adapters/sqlite_repository.go` exists with stub methods

### 8.2 Application Service - PARTIAL

- [x] ModerationService skeleton exists
- [ ] Implement CreateReport method (currently returns nil with TODO)
- [ ] Implement GetReport method (currently returns nil with TODO)
- [ ] Add report review logic
- [ ] Add content deletion logic
- [ ] Add authorization checks (moderator/admin only)
- [ ] Add service tests

**Status**: `internal/modules/moderation/application/service.go` has stub methods with TODO markers

### 8.3 HTTP Handlers - PARTIAL

- [x] HTTP handler skeleton exists with TODO note
- [ ] Create report handler
- [ ] List reports handler (moderators)
- [ ] Review report handler (moderators/admins)
- [ ] Delete content handler (moderators)
- [ ] Route registration (has TODO: POST /reports, GET /reports, PUT /reports/{id})
- [ ] Add handler tests

**Status**: `internal/modules/moderation/adapters/http_handler.go` has skeleton with TODO marker for routes

**Estimated Time**: 2-3 days

---

## Phase 9: Frontend (Templates & Static Assets) 🎨

### 9.1 Templates
- [ ] Create base template with layout
- [ ] Home page with post list
- [ ] Registration page
- [ ] Login page
- [ ] Post creation page
- [ ] Post view page with comments
- [ ] User profile page
- [ ] Activity page
- [ ] Notification page
- [ ] Admin/moderator pages

### 9.2 Styling
- [ ] Create responsive CSS
- [ ] Add mobile support
- [ ] Style forms
- [ ] Style navigation
- [ ] Style notifications

### 9.3 JavaScript
- [ ] Form validation
- [ ] AJAX for reactions
- [ ] Real-time notification updates (optional)
- [ ] Image preview before upload
- [ ] Character counters

**Estimated Time**: 3-4 days

---

## Phase 10: Security & Rate Limiting 🔒

### 10.1 HTTPS/TLS
- [ ] Generate SSL certificates
- [ ] Configure TLS in server
- [ ] Add cipher suite configuration
- [ ] Test HTTPS connections

### 10.2 Rate Limiting
- [ ] Implement per-IP rate limiting
- [ ] Implement per-user rate limiting
- [ ] Add rate limit middleware
- [ ] Configure rate limit rules per endpoint

### 10.3 Security Headers
- [ ] Add CSP headers
- [ ] Add X-Frame-Options
- [ ] Add X-Content-Type-Options
- [ ] Add Strict-Transport-Security

### 10.4 CSRF Protection
- [ ] Implement CSRF tokens
- [ ] Add CSRF middleware
- [ ] Validate CSRF on forms

**Estimated Time**: 2 days

---

## Phase 11: Testing 🧪 NOT STARTED

### 11.1 Unit Tests - NOT STARTED

- [ ] Domain layer tests
- [ ] Application service tests
- [ ] Validator tests
- [ ] Utility tests

**Status**: Test stubs exist in `tests/unit/unit_test.go` with message "Unit tests will be implemented following TDD principles"

### 11.2 Integration Tests - NOT STARTED

- [ ] Repository tests with real database
- [ ] Handler tests with HTTP requests
- [ ] Module integration tests

**Status**: Test stubs exist in `tests/integration/integration_test.go` with message "Integration tests will be implemented after core functionality"

### 11.3 End-to-End Tests - NOT STARTED

- [ ] User registration and login flow
- [ ] Post creation and commenting flow
- [ ] Reaction flow
- [ ] Moderation flow

**Status**: Not started

**Estimated Time**: 3-4 days

---

## Phase 12: Documentation & Deployment 📚

### 12.1 Documentation
- [ ] Update README with complete setup
- [ ] Add API documentation
- [ ] Add deployment guide
- [ ] Add contribution guidelines

### 12.2 Docker
- [ ] Optimize Dockerfile
- [ ] Update docker-compose.yml
- [ ] Test containerized deployment
- [ ] Add health checks

### 12.3 Final Polish
- [ ] Code review and refactoring
- [ ] Performance optimization
- [ ] Error handling improvements
- [ ] Logging improvements

**Estimated Time**: 2 days

---

## Total Estimated Time

**Core Features**: 25-30 days  
**With Testing & Polish**: 30-35 days

## Current Progress Summary

**Overall Completion: ~10%**

| Phase | Status | Completion |
|-------|--------|------------|
| Phase 1: Foundation | ⏳ In Progress (Scaffolding) | 5% |
| Phase 2: Authentication | ❌ Not Started | 5% (structure only) |
| Phase 3: User Module | ❌ Not Started | 5% (structure only) |
| Phase 4: Post Module | ⏳ Partial | 10% (structure + stubs) |
| Phase 5: Comment Module | ⏳ Partial | 10% (structure + stubs) |
| Phase 6: Reaction Module | ⏳ Partial | 10% (structure + stubs) |
| Phase 7: Notification Module | ⏳ Partial | 10% (structure + stubs) |
| Phase 8: Moderation Module | ⏳ Partial | 10% (structure + stubs) |
| Phase 9: Frontend | ❌ Not Started | 0% |
| Phase 10: Security | ❌ Not Started | 0% |
| Phase 11: Testing | ❌ Not Started | 0% |
| Phase 12: Documentation | ⏳ Partial | 20% (docs exist) |

## Known TODO Items by File

### Platform Layer
- `internal/platform/config/config.go`: Load() and Validate() are placeholders
- `internal/platform/database/connection.go`: NewConnection() is placeholder
- `internal/platform/database/migrator.go`: All methods (Migrate, Rollback, Version) are placeholders
- `internal/platform/database/transaction.go`: BeginTx() is placeholder
- `internal/platform/logger/logger.go`: All methods (Debug, Info, Warn, Error) are placeholders
- `internal/platform/httpserver/server.go`: New(), RegisterRoutes(), Start(), Shutdown() are placeholders
- `internal/platform/httpserver/middleware.go`: All 6 middleware functions are placeholders
- `internal/platform/validator/validator.go`: Email validation, password strength, Sanitize(), SanitizeHTML() are TODOs
- `internal/platform/errors/errors.go`: Error types defined, implementation complete

### Auth Module
- `internal/modules/auth/domain/session.go`: Validate() method is TODO
- `internal/modules/auth/application/service.go`: ALL 8 methods are TODO placeholders
- `internal/modules/auth/adapters/http_handler.go`: ALL 4 handlers + helper methods are TODO placeholders
- `internal/modules/auth/adapters/sqlite_session_repository.go`: ALL 7 methods are TODO placeholders
- `internal/modules/auth/adapters/sqlite_user_repository.go`: ALL 4 methods are TODO placeholders

### User Module
- `internal/modules/user/domain/user.go`: HasPermission() is TODO
- `internal/modules/user/adapters/sqlite_repository.go`: ALL 10 methods are TODO placeholders
- `internal/modules/user/adapters/http_handler.go`: ALL 4 handlers are TODO placeholders

### Post Module
- `cmd/forum/main.go`: Line 112 - TODO: Add categoryRepo
- `internal/modules/post/adapters/`: Methods exist but need implementation
- `internal/modules/post/adapters/http_handler.go`: Methods exist but need implementation

### Comment Module
- `internal/modules/comment/adapters/http_handler.go`: TODO note for routes: "POST /comments, GET /comments/{id}, PUT /comments/{id}, DELETE /comments/{id}"
- `internal/modules/comment/adapters/sqlite_repository.go`: Methods exist but need implementation

### Reaction Module
- `internal/modules/reaction/application/service.go`: ReactTo() returns nil (TODO), GetReactionCounts() returns 0, 0, nil (TODO)
- `internal/modules/reaction/adapters/http_handler.go`: TODO note for routes: "POST /reactions, DELETE /reactions, GET /reactions"
- `internal/modules/reaction/adapters/sqlite_repository.go`: Methods exist but need implementation

### Moderation Module [OPTIONAL]
- `internal/modules/moderation/application/service.go`: CreateReport() returns nil (TODO), GetReport() returns nil (TODO)
- `internal/modules/moderation/adapters/http_handler.go`: TODO note for routes: "POST /reports, GET /reports, PUT /reports/{id}"
- `internal/modules/moderation/adapters/sqlite_repository.go`: Methods exist but need implementation

### Notification Module [OPTIONAL]
- `internal/modules/notification/application/service.go`: Methods exist but return nil with TODO
- `internal/modules/notification/adapters/sqlite_repository.go`: Methods exist but need implementation

### Test Files
- `tests/unit/unit_test.go`: Stub with message "Unit tests will be implemented following TDD principles"
- `tests/integration/integration_test.go`: Stub with message "Integration tests will be implemented after core functionality"

## Priority Order

1. **Must Have** (Core Requirements):
   - Phase 1: Foundation
   - Phase 2: Authentication
   - Phase 3: User Management
   - Phase 4: Posts
   - Phase 5: Comments
   - Phase 6: Reactions
   - Phase 9: Frontend

2. **Should Have** (Enhanced Features):
   - Phase 7: Notifications
   - Phase 8: Moderation
   - Phase 10: Security
   - Phase 11: Testing

3. **Nice to Have** (Bonus Features):
   - OAuth integration
   - Real-time notifications
   - Advanced moderation tools
   - Activity analytics

## Implementation Tips

1. **Test as you go**: Write tests immediately after implementing features
2. **Commit frequently**: Small, focused commits with clear messages
3. **Review architecture**: Ensure modules maintain clean boundaries
4. **Use dependency injection**: Wire dependencies in main.go
5. **Follow SOLID principles**: Keep classes/functions focused
6. **Keep it simple**: Follow KISS principle
7. **Document as you code**: Update documentation with changes
8. **Run the app frequently**: Test manually during development

## Getting Started

Start with Phase 1 (Foundation) as all other phases depend on it. Then proceed sequentially, but you can parallelize some work:

- Phases 4, 5, 6 can be partially parallelized (they're independent)
- Phase 9 (Frontend) can start once Phase 4 (Posts) is functional
- Phase 11 (Testing) should be done incrementally with each phase

Good luck! 🚀
