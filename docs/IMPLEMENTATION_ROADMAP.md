# Implementation Roadmap

This document outlines the step-by-step implementation plan for the Forum project.

## Phase 1: Foundation (Platform Services) ⏳

### 1.1 Configuration Management
- [ ] Implement `config.Load()` function
- [ ] Add environment variable parsing
- [ ] Create default configuration values
- [ ] Add validation for required config

### 1.2 Database Layer
- [ ] Implement `database.New()` connection
- [ ] Add connection pooling configuration
- [ ] Implement `BeginTx()` for transactions
- [ ] Create migration runner logic
- [ ] Test database connection and migrations

### 1.3 Logging
- [ ] Implement structured logger
- [ ] Add log levels (Debug, Info, Warn, Error, Fatal)
- [ ] Add context-aware logging
- [ ] Configure log output format

### 1.4 HTTP Server
- [ ] Implement server creation and startup
- [ ] Add graceful shutdown
- [ ] Configure TLS/HTTPS
- [ ] Implement middleware chain

### 1.5 Middleware
- [ ] Recovery middleware (panic handling)
- [ ] Request logging middleware
- [ ] CORS middleware
- [ ] Rate limiting middleware
- [ ] Security headers middleware

### 1.6 Utilities
- [ ] Input validation functions
- [ ] Error handling utilities
- [ ] Response helpers

**Estimated Time**: 2-3 days

---

## Phase 2: Authentication Module 🔐

### 2.1 Domain Layer
- [ ] Finalize Session entity
- [ ] Validate business rules
- [ ] Add domain tests

### 2.2 Repository (Output Adapter)
- [ ] Implement SQLite session repository
- [ ] Add session CRUD operations
- [ ] Implement session cleanup
- [ ] Add repository tests

### 2.3 Password Hasher
- [ ] Finalize bcrypt implementation
- [ ] Add cost configuration
- [ ] Add tests

### 2.4 Application Service
- [ ] Implement Register use case
- [ ] Implement Login use case
- [ ] Implement Logout use case
- [ ] Implement ValidateSession use case
- [ ] Implement RefreshSession use case
- [ ] Add service tests

### 2.5 HTTP Handlers
- [ ] Implement registration handler
- [ ] Implement login handler
- [ ] Implement logout handler
- [ ] Add cookie management
- [ ] Add request validation
- [ ] Add handler tests

### 2.6 Middleware
- [ ] Implement RequireAuth middleware
- [ ] Implement OptionalAuth middleware
- [ ] Add user context injection
- [ ] Add middleware tests

### 2.7 OAuth (Bonus)
- [ ] Implement Google OAuth provider
- [ ] Implement GitHub OAuth provider
- [ ] Add OAuth callback handler
- [ ] Add OAuth tests

**Estimated Time**: 3-4 days

---

## Phase 3: User Module 👤

### 3.1 Repository
- [ ] Implement SQLite user repository
- [ ] Add user CRUD operations
- [ ] Add email/username existence checks
- [ ] Add repository tests

### 3.2 Application Service
- [ ] Implement GetByID
- [ ] Implement UpdateProfile
- [ ] Implement PromoteToModerator
- [ ] Implement DemoteFromModerator
- [ ] Add authorization checks
- [ ] Add service tests

### 3.3 HTTP Handlers
- [ ] Profile view handler
- [ ] Profile edit handler
- [ ] User list handler (admin)
- [ ] Role management handlers (admin)
- [ ] Add handler tests

**Estimated Time**: 2 days

---

## Phase 4: Post Module 📝

### 4.1 Repositories
- [ ] Implement post repository
- [ ] Implement category repository
- [ ] Add post filtering logic
- [ ] Add repository tests

### 4.2 Image Storage
- [ ] Implement image upload
- [ ] Add image validation (format, size)
- [ ] Add image storage logic
- [ ] Add image deletion
- [ ] Add tests

### 4.3 Application Services
- [ ] Implement PostService
- [ ] Implement CategoryService
- [ ] Add authorization checks
- [ ] Add service tests

### 4.4 HTTP Handlers
- [ ] Post creation handler (with image upload)
- [ ] Post view handler
- [ ] Post list handler
- [ ] Post edit handler
- [ ] Post delete handler
- [ ] Category management handlers
- [ ] Filter handlers
- [ ] Add handler tests

**Estimated Time**: 3-4 days

---

## Phase 5: Comment Module 💬

### 5.1 Repository
- [ ] Implement comment repository
- [ ] Add comment CRUD operations
- [ ] Add repository tests

### 5.2 Application Service
- [ ] Implement CommentService
- [ ] Add authorization checks
- [ ] Add service tests

### 5.3 HTTP Handlers
- [ ] Comment creation handler
- [ ] Comment list handler
- [ ] Comment edit handler
- [ ] Comment delete handler
- [ ] Add handler tests

**Estimated Time**: 2 days

---

## Phase 6: Reaction Module 👍👎

### 6.1 Repository
- [ ] Implement reaction repository
- [ ] Add reaction toggle logic
- [ ] Add reaction counting
- [ ] Add liked posts retrieval
- [ ] Add repository tests

### 6.2 Application Service
- [ ] Implement ReactionService
- [ ] Add authorization checks
- [ ] Add service tests

### 6.3 HTTP Handlers
- [ ] Add/toggle reaction handler
- [ ] Remove reaction handler
- [ ] Get reaction counts handler
- [ ] Add handler tests

### 6.4 Integration
- [ ] Integrate with Post module
- [ ] Integrate with Comment module
- [ ] Trigger notifications on reactions

**Estimated Time**: 2 days

---

## Phase 7: Notification Module 🔔

### 7.1 Repository
- [ ] Implement notification repository
- [ ] Add notification CRUD operations
- [ ] Add unread count query
- [ ] Add repository tests

### 7.2 Application Service
- [ ] Implement NotificationService
- [ ] Add notification creation logic
- [ ] Add service tests

### 7.3 HTTP Handlers
- [ ] Notification list handler
- [ ] Mark as read handler
- [ ] Mark all as read handler
- [ ] Notification count handler
- [ ] Add handler tests

### 7.4 Integration
- [ ] Trigger notifications on post reactions
- [ ] Trigger notifications on comment reactions
- [ ] Trigger notifications on new comments

**Estimated Time**: 2 days

---

## Phase 8: Moderation Module 🛡️

### 8.1 Repository
- [ ] Implement report repository
- [ ] Add report CRUD operations
- [ ] Add repository tests

### 8.2 Application Service
- [ ] Implement ModerationService
- [ ] Add report creation
- [ ] Add report review logic
- [ ] Add content deletion logic
- [ ] Add authorization checks (moderator/admin only)
- [ ] Add service tests

### 8.3 HTTP Handlers
- [ ] Create report handler
- [ ] List reports handler (moderators)
- [ ] Review report handler (moderators/admins)
- [ ] Delete content handler (moderators)
- [ ] Add handler tests

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

## Phase 11: Testing 🧪

### 11.1 Unit Tests
- [ ] Domain layer tests
- [ ] Application service tests
- [ ] Validator tests
- [ ] Utility tests

### 11.2 Integration Tests
- [ ] Repository tests with real database
- [ ] Handler tests with HTTP requests
- [ ] Module integration tests

### 11.3 End-to-End Tests
- [ ] User registration and login flow
- [ ] Post creation and commenting flow
- [ ] Reaction flow
- [ ] Moderation flow

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
