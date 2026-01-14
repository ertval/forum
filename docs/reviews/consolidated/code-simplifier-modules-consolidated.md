# Consolidated Modules Code Simplifier Review

**Date:** 2026-01-14
**Scope:** `internal/modules/` (post, reaction, moderation, notification, user, comment, auth)

---

## Table of Contents

1. [Post Module](#post-module)
2. [Reaction Module](#reaction-module)
3. [Moderation Module](#moderation-module)
4. [Notification Module](#notification-module)
5. [User Module](#user-module)
6. [Comment Module](#comment-module)
7. [Auth Module](#auth-module)
8. [Cross-Module Issues](#cross-module-issues)
9. [Action Items](#action-items)

---

## Post Module

### 1. Inefficient Template Parsing (High)

**File:** `post/adapters/http_handler_page.go`

Templates are parsed from disk on every request. Parse once at startup.

```go
// Current: template.ParseFiles(...) inside handler
// Fix: Store in HTTPHandler, use h.templates.ExecuteTemplate()
```

### 2. N+1 Query Problem (High)

**File:** `post/adapters/sqlite_repository.go` (Line 431-436)

`getPostCategories()` called per post in loop. Use `GROUP_CONCAT` or batch query with `IN` clause.

### 3. Ad-hoc Logger Instantiation (Medium)

**File:** `post/adapters/http_handler_api.go` (Lines 118, 260, 332)

Logger created per-request. Inject logger into `HTTPHandler` at construction.

### 4. Complex Request Parsing Logic (Medium)

**File:** `post/adapters/http_handler_api.go` (Lines 71-129, 266-328)

Duplicate parsing for multipart/JSON. Extract to `parsePostRequest()` helper.

### 5. Duplicate Page Handler Logic (Medium)

**File:** `post/adapters/http_handler_page.go` (Lines 29-260)

`HomePage` and `BoardPage` share identical logic. Extract to `renderPostList()` helper.

### 6. Inefficient Category Validation (Medium)

**File:** `post/application/service.go` (Lines 103-108, 163-168)

Loop calls `GetByName()`. Add `GetByNames(ctx, []string)` for batch lookup.

### 7. Validation/Error Message Mismatch (Low)

**File:** `post/domain/post.go` vs `post/domain/errors.go`

Validates at 255 chars but error says 300. Use constant `MaxTitleLength`.

### 8. Redundant Author Fields (Low)

**File:** `post/domain/post.go` (Lines 12-13)

`AuthorUsername` and `Author` duplicate same data. Pick one.

### 9. Redundant `min` Function (Low)

**File:** `post/adapters/http_handler.go` (Lines 199-204)

Go 1.21+ has built-in `min()`. Remove custom function.

### 10. Verbose String Building (Low)

**File:** `post/adapters/http_handler.go` (Lines 170-179)

Use `strings.Join(parts, " ")` instead of manual loop.

---

## Reaction Module

### 1. Fire-and-Forget Goroutines Without Logging (High)

**File:** `reaction/application/service.go` (Lines 105-107, 147-149)

```go
// Current
go func() { _ = s.userService.IncrementReactionCount(...) }()

// Fix: Log errors
go func() {
    if err := s.userService.IncrementReactionCount(...); err != nil {
        s.logger.Error("failed to increment count", logger.Error(err))
    }
}()
```

### 2. TOCTOU Race Condition in Toggle (Medium)

**File:** `reaction/application/service.go` (Lines 67-89)

Check-then-act without transaction. Wrap in database transaction.

### 3. Repeated Target ID Resolution (Medium)

**File:** `reaction/adapters/sqlite_repository.go` (5+ locations)

Extract `getTargetID(ctx, publicID, targetType)` helper method.

### 4. Repeated Target Validation (Medium)

**File:** `reaction/application/service.go` (5+ locations)

Extract `validateTargetExists()` helper combining type check and existence.

### 5. Non-Idiomatic Path Parsing (Medium)

**File:** `reaction/adapters/http_handler_api.go` (Lines 187-195, 236-244)

Use `r.PathValue("targetType")` instead of manual `strings.Split`.

### 6. Stale Mock Implementations (High)

**File:** `reaction/ports/service_test.go` (Lines 29-56)

Mocks use old signatures (`targetID int` vs `targetPublicID string`). Update to match interface.

### 7. Inconsistent JSON Response (Low)

**File:** `reaction/adapters/http_handler_api.go` (Lines 108-109)

Use `writeJSON()` helper instead of `fmt.Fprintf` for consistent Content-Type.

### 8. Missing Error Context in Repository (Low)

**File:** `reaction/adapters/sqlite_repository.go`

Wrap errors: `return fmt.Errorf("inserting reaction: %w", err)`

### 9. Unchecked JSON Encoding Error (Low)

**File:** `reaction/adapters/http_handler_api.go` (Lines 297-301)

Check and log error from `json.Encode()`.

---

## Moderation Module

### 1. Incomplete Domain Validation (High)

**File:** `moderation/domain/report.go` (Lines 31-36)

`IsValid()` is placeholder. Implement full `Validate() error` method.

### 2. Missing Type Safety for Status/TargetType (Medium)

**File:** `moderation/domain/report.go`

```go
// Current: string fields
// Fix: Define typed constants
type ReportStatus string
type TargetType string

const (
    StatusPending  ReportStatus = "pending"
    TargetPost     TargetType   = "post"
)
```

### 3. Primitive Obsession in Service Interface (Medium)

**File:** `moderation/ports/service.go` (Lines 12-22)

Use request structs instead of multiple primitives.

### 4. Placeholder TODOs Return nil (Medium)

**File:** `moderation/application/service.go` (Lines 23-40)

Return `domain.ErrNotImplemented` instead of silent `nil`.

### 5. Missing Repository Implementation (Medium)

**File:** `moderation/adapters/sqlite_repository.go`

Implement with proper error wrapping and UUID generation.

---

## Notification Module

### 1. Domain/Schema Mismatch (High)

**File:** `notification/domain/notification.go` vs `migrations/007_*.sql`

Domain missing `ActorID` and `PublicActorID` present in schema.

### 2. TDD Violation: Empty Tests Pass (High)

**File:** `notification/application/service_test.go`

Tests pass because implementation returns `nil`. Assert actual state changes.

### 3. Missing Domain Validation (Medium)

**File:** `notification/domain/notification.go`

Add `Validate() error` method per project conventions.

### 4. Cross-Module ID Coupling (Medium)

**File:** `notification/domain/notification.go`

Storing internal `int` IDs from other modules couples to their internals.

---

## User Module

### 1. Repetitive Scanning Logic (Medium)

**File:** `user/adapters/sqlite_repository.go` (Lines 66-207)

Extract `scanUser(scanner)` helper for `GetByID`, `GetByPublicID`, `GetByEmail`, `GetByUsername`.

### 2. Redundant Role Validation (Low)

**File:** `user/adapters/http_handler_api.go` + `user/application/service.go`

Handler and service both validate role. Keep only in service layer.

### 3. Missing Query-Based Pagination (Low)

**File:** `user/adapters/http_handler_api.go` (Lines 47-63)

Parse `offset`/`limit` from query string instead of hardcoding.

### 4. Repeated Public-to-Internal ID Lookup (Medium)

**File:** `user/adapters/http_handler_api.go` (Lines 93-100, 123-129, 148-154)

Create service methods accepting `publicID` directly.

### 5. Missing Error Wrapping (Low)

**File:** `user/adapters/sqlite_repository.go`

Use `fmt.Errorf("create user: %w", err)` for context.

### 6. HasPermission Not Implemented (Medium)

**File:** `user/domain/user.go` (Lines 44-48)

Method always returns `false`. Implement or remove.

---

## Comment Module

### 1. Missing rows.Err() Check (Medium)

**File:** `comment/adapters/sqlite_repository.go` (Lines 91-123)

```go
// Add after loop:
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterating comments: %w", err)
}
```

### 2. Redundant Content Validation (Low)

**File:** `comment/domain/comment.go` (Lines 24-38)

Simplify: `strings.TrimSpace(c.Content) == ""` covers empty and whitespace.

### 3. Inconsistent Error Wrapping (Medium)

**File:** `comment/adapters/sqlite_repository.go`

Wrap errors like post repository does.

---

## Auth Module

### 1. Deprecated Functions (Medium)

**File:** `auth/adapters/middleware.go` (Lines 64-139)

Remove `RequireAuth`, `OptionalAuthFunc`, etc. after updating callers to use `AuthMiddleware`.

### 2. Duplicate Response Struct (Low)

**File:** `auth/adapters/http_handler_api.go` (Lines 86-98, 141-153)

Extract `authResponse` struct used by both `RegisterAPI` and `LoginAPI`.

---

## Cross-Module Issues

### 1. Inconsistent Error Wrapping

Post repository wraps errors; comment/reaction/user do not. Standardize with `%w`.

### 2. Goroutines Ignoring Errors

Post, comment, reaction services all use `_ =` in goroutines. Log errors.

### 3. Logger Injection

Post handler creates logger per-request. Reaction handler injects it correctly. Standardize.

### 4. Path Parameter Extraction

Post uses `r.PathValue()`. Reaction uses manual `strings.Split`. Use stdlib consistently.

---

## Action Items

### Critical

- [ ] Fix stale mocks in `reaction/ports/service_test.go`
- [ ] Fix notification domain/schema mismatch (add `ActorID`)
- [ ] Fix TDD violations in notification tests
- [ ] Add `rows.Err()` check in comment repository

### High Priority

- [ ] Parse templates once at startup (post module)
- [ ] Fix N+1 query in post listing
- [ ] Log errors in all fire-and-forget goroutines
- [ ] Implement moderation domain validation

### Medium Priority

- [ ] Extract target validation helpers (reaction service)
- [ ] Extract `getTargetID` helper (reaction repository)
- [ ] Extract `scanUser` helper (user repository)
- [ ] Inject logger into post handler
- [ ] Use `r.PathValue()` consistently
- [ ] Add transaction support for reaction toggle
- [ ] Implement or remove `HasPermission()`
- [ ] Remove deprecated auth middleware functions

### Low Priority

- [ ] Remove custom `min()` function
- [ ] Use `strings.Join()` for title building
- [ ] Simplify comment validation
- [ ] Extract duplicate auth response struct
- [ ] Add error wrapping across all repositories
- [ ] Add query-based pagination to user API
- [ ] Check JSON encoding errors in writeJSON helpers

---

## Source Documents

- `code-simplifier-moderation-202601141535.md`
- `code-simplifier-moderation-202601141555.md`
- `code-simplifier-modules-202601141516.md`
- `code-simplifier-notification-202601141540.md`
- `code-simplifier-post-20260114-1456.md`
- `code-simplifier-post-202601141533.md`
- `code-simplifier-reaction-20260114-1456.md`
- `code-simplifier-reaction-202601141456.md`
- `code-simplifier-reaction-202601141545.md`
- `code-simplifier-user-202601141550.md`
