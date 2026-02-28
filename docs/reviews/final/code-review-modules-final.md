# Consolidated Code Review: Forum Modules

**Date:** 2026-01-14  
**Scope:** All internal modules (`auth`, `comment`, `moderation`, `notification`, `post`, `reaction`, `user`)  
**Sources:** Code-review + code-simplifier documents merged and deduplicated

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Auth Module](#auth-module)
3. [Comment Module](#comment-module)
4. [Moderation Module](#moderation-module)
5. [Notification Module](#notification-module)
6. [Post Module](#post-module)
7. [Reaction Module](#reaction-module)
8. [User Module](#user-module)
9. [Cross-Module Issues](#cross-module-issues)
10. [Summary of Action Items](#summary-of-action-items)

---

## Executive Summary

The Forum modules demonstrate **strong adherence to hexagonal architecture** with proper separation of concerns. The codebase follows **Go idiomatic patterns** and the ID security policy (INT internally, UUID publicly) is consistently applied.

| Category                  | Count | Severity    |
| ------------------------- | ----- | ----------- |
| Critical Issues           | 25+   | 🔴 Must Fix |
| Performance Issues        | 15+   | 🟡 Medium   |
| Nitpicks & Best Practices | 30+   | 🟢 Low      |

**Common Patterns:**

- Fire-and-forget goroutines without error handling/synchronization
- Silent error swallowing with `_ =`
- N+1 query patterns
- Template parsing on every request
- Missing `rows.Err()` checks after DB iterations

---

## Auth Module

### Critical Issues

#### AUTH-1: Template Parsing on Every Request

- **Location:** `adapters/http_handler_page.go`, `LoginPage`/`RegisterPage`
- **Description:** `template.ParseFiles` on every request causes disk I/O, latency, potential panic.
- **Fix:** Parse once in `NewHTTPHandler`, use `h.templates.ExecuteTemplate()`.

#### AUTH-2: Ignored Errors in Session Cleanup

- **Location:** `application/service.go`, `Login`/`ValidateSession`
- **Description:** `DeleteByUserID`/`Delete` errors ignored. Risks multiple active sessions.
- **Fix:** Log errors: `log.Printf("WARNING: failed to delete sessions: %v", err)`

#### AUTH-3: Weak Password Policy

- **Location:** `application/service.go`, `ValidateCredentials`
- **Description:** Min length 6 chars is insufficient.
- **Fix:** Increase to 8-12 chars or add complexity requirements.

#### AUTH-4: Registration DoS via Bcrypt

- **Location:** `application/service.go`, `Register`
- **Description:** Bcrypt hashing is CPU-intensive; flood attacks possible.
- **Fix:** Rate limiting on `/api/auth/register` and `/api/auth/login`.

#### AUTH-5: Session Token Entropy

- **Location:** `application/service.go`
- **Description:** UUID v4 (122 bits) not designed for session tokens.
- **Fix:** Use `crypto/rand` 32-byte base64-encoded tokens.

#### AUTH-6: Cookie Security Flags Hardcoded

- **Location:** `adapters/http_handler_api.go`, Lines 74, 129, 181
- **Description:** `Secure: false` hardcoded. Dangerous for production.
- **Fix:** Make conditional via config: `Secure: h.isProduction()`.

### Performance

#### AUTH-PERF-1: Redundant User Fetching

- **Description:** `RegisterAPI`/`LoginAPI` call service then `userService.GetByID` again for `PublicID`.
- **Fix:** Return full `User` or `PublicID` from auth service methods.

#### AUTH-PERF-2: Duplicate User Lookups in Middleware

- **Description:** `OptionalAuth`/`RequireAuth` fetch full user each request.
- **Fix:** Store `PublicID` in sessions table (denormalization).

### Nitpicks

- **API Robustness:** `RegisterAPI` `strings.Contains` fallback is brittle.
- **Deprecated Functions:** `RequireAuthFunc`, `OptionalAuthFunc`, `GetUserID`, `GetUsername`, `IsAuthenticated` should be removed.
- **Transaction Safety:** `Register` has separate User/Session creation; partial failure leaves inconsistent state.
- **Duplicate Response Struct:** `RegisterAPI`/`LoginAPI` use identical response structures. Extract to `authResponse`.

---

## Comment Module

### Critical Issues

#### COMMENT-1: Unmanaged Goroutines & Ignored Errors

- **Location:** `application/service.go`, Lines 59-61, 111-113
- **Description:** `go func() { _ = s.userService.IncrementCommentCount(...) }()` - no tracking, no timeout, errors swallowed.
- **Fix:** Add timeout and logging:
  ```go
  go func(uid int) {
      ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
      defer cancel()
      if err := s.userService.IncrementCommentCount(ctx, uid); err != nil {
          log.Printf("WARNING: failed to increment comment count: %v", err)
      }
  }(userID)
  ```

#### COMMENT-2: Missing `rows.Err()` Check

- **Location:** `adapters/sqlite_repository.go`, Lines 120, 160, 201
- **Description:** Must check `rows.Err()` after iteration loop.
- **Fix:** Add `if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterating comments: %w", err) }`

#### COMMENT-3: Missing Transaction Boundary

- **Location:** `application/service.go`, `CreateComment`
- **Description:** Comment insert + user count update are separate; partial failure = inconsistent data.

### Performance

#### COMMENT-PERF-1: N+1 Query in Page Handlers

- **Location:** `adapters/http_handler_page.go`, Lines 89-167
- **Description:** `MyCommentsPage` loops with calls to `userService.GetByID`, `postService.GetPost`, `reactionService.CountReactions`. 20 comments = 60+ queries.
- **Fix:** Implement batch methods: `GetPostsByIDs([]string)`.

#### COMMENT-PERF-2: Offset Pagination

- **Description:** `LIMIT ? OFFSET ?` scans discards rows as offset grows.
- **Fix:** Use keyset pagination: `WHERE created_at < ?`.

#### COMMENT-PERF-3: Logger Recreated on Every Error

- **Location:** `adapters/http_handler.go`, Lines 129-132
- **Fix:** Initialize logger once in `HTTPHandler` struct.

### Nitpicks

- **Code Duplication:** `LoadMoreCommentsAPI`/`MyCommentsPage` share enrichment logic. Extract `enrichComments()`.
- **Complex Handler:** `MyCommentsPage` violates SRP.
- **Silent Failures:** `h.categoryService.List` errors logged with `fmt.Printf`, execution continues.
- **RESTful Design:** `POST /api/comments/posts/{post_id}` vs `POST /api/posts/{post_id}/comments`.
- **Redundant Content Validation:** `strings.TrimSpace(c.Content) == ""` covers empty+whitespace.
- **Inconsistent Error Wrapping:** Wrap like post repository.

---

## Moderation Module

> **Note:** Module is **scaffold only**, marked `[OPTIONAL FEATURE]`.

### Critical Issues

#### MOD-1: Module Incomplete

- **Description:** All methods return `nil` or `501 Not Implemented`.
- **Fix:** Implement per `flow.md` or return `errors.New("not implemented")`.

#### MOD-2: Silent Placeholder Failure

- **Location:** `application/service.go`, `CreateReport`
- **Description:** Returns `nil` (success) without action. Reports are lost.
- **Fix:** Return `domain.ErrNotImplemented`.

#### MOD-3: Missing RBAC

- **Location:** `adapters/http_handler_api.go`
- **Description:** Moderation endpoints lack authorization middleware.
- **Fix:** Implement `RequireRole` middleware.

#### MOD-4: Missing Content Validation

- **Location:** `application/service.go`:23
- **Description:** `CreateReport` doesn't verify target post/comment exists.
- **Fix:** Inject and call `PostService`/`CommentService` to validate.

#### MOD-5: Incomplete Domain Validation

- **Location:** `domain/report.go`, `IsValid()`
- **Description:** Placeholder only checks `TargetType`. Should validate `Reason`, `Status`.
- **Fix:** Implement full `Validate() error` method.

#### MOD-6: Documentation vs Implementation Mismatch

- **Location:** `flow.md` vs actual code
- **Description:** `flow.md` describes complete implementation; code is empty.
- **Fix:** Synchronize documentation with actual state.

### Performance

#### MOD-PERF-1: N+1 Hazard in Report Listing

- **Fix:** Use JOIN for reporter/moderator usernames.

#### MOD-PERF-2: External ID Resolution Bottleneck

- **Fix:** Ensure `GetInternalIDByPublicID` methods are efficient.

### Nitpicks

- **Inconsistent Status Naming:** `StatusResolved` vs `dismissed` in `flow.md`.
- **Missing Type Safety:** Use typed constants for `ReportStatus`, `TargetType`.
- **Primitive Obsession:** Use request structs instead of multiple parameters.

---

## Notification Module

> **Note:** Module is **scaffold only**, marked `[OPTIONAL FEATURE]`.

### Critical Issues

#### NOTIF-1: Incomplete Implementation

- **Description:** Methods are placeholders returning `nil`.
- **Fix:** Implement or return `errors.New("not implemented")`.

#### NOTIF-2: Domain/Schema Mismatch

- **Location:** `domain/notification.go` vs `migrations/007_*.sql`
- **Description:** Domain missing `ActorID` field present in schema as `NOT NULL`.
- **Fix:** Add `ActorID int` and `PublicActorID string` to struct.

#### NOTIF-3: Missing Service Validation

- **Location:** `application/service.go`, `CreateNotification`
- **Description:** TODO for validation does nothing.

#### NOTIF-4: Target ID Resolution Strategy

- **Description:** Accepts `targetPublicID` (UUID), domain requires `TargetID` (int).
- **Options:** Store UUID directly or inject services to resolve.

#### NOTIF-5: TDD Violation - Empty Tests Pass

- **Location:** `notification/application/service_test.go`
- **Description:** Tests pass because implementation returns `nil`.
- **Fix:** Assert actual state changes.

### Performance

#### NOTIF-PERF-1: Polling vs Push

- **Description:** `GET /api/notifications` implies polling.
- **Fix:** Consider SSE/WebSockets or "Since ID" incremental sync.

#### NOTIF-PERF-2: Sync vs Async Creation

- **Fix:** Use worker pool for notification creation.

### Nitpicks

- **Cross-Module ID Coupling:** Storing internal `int` IDs couples to other module internals.
- **Missing Domain Validation:** Add `Validate() error` method.

---

## Post Module

### Critical Issues

#### POST-1: N+1 Query Pattern

- **Location:** `adapters/sqlite_repository.go`, `List` (Line 432)
- **Description:** Separate category query per post. 50 posts = 51 queries.
- **Fix:** Use `GROUP_CONCAT` or batch query with `IN` clause.

#### POST-2: Template Parsing on Every Request

- **Location:** `adapters/http_handler_page.go`, Lines 130, 249, 346, 401, 476
- **Fix:** Parse once in `NewHTTPHandler`.

#### POST-3: Redundant Auth Logic

- **Location:** `adapters/http_handler_page.go`, Lines 39-46, 154-161, 292-299
- **Description:** Manual cookie parsing bypasses middleware.
- **Fix:** Use `OptionalAuth` middleware.

#### POST-4: Race Condition in Async Ops

- **Location:** `application/service.go`, `CreatePost`/`DeletePost`
- **Description:** `go func() { _ = s.userService.IncrementPostCount(...) }()` - errors swallowed, race on counter.
- **Fix:** Synchronous update or durable job queue.

#### POST-5: SQL Injection Risk

- **Location:** `adapters/sqlite_repository.go` (Line 329)
- **Description:** Dynamic `IN` clause construction. Risky pattern.
- **Fix:** Use query builder or rigorous testing.

#### POST-6: Missing Transaction Rollback on Category Failure

- **Location:** `adapters/sqlite_repository.go`, Lines 27-88
- **Description:** Error message leaks `sql.ErrNoRows`.
- **Fix:** Map to `domain.ErrCategoryNotFound`.

### Performance

#### POST-PERF-1: Suboptimal Counting

- **Location:** `adapters/sqlite_repository.go`, `GetByID`/`List`
- **Description:** `LEFT JOIN` on derived tables materializes counts for entire DB.
- **Fix:** Use correlated subqueries or denormalize counts.

#### POST-PERF-2: Excessive Allocations

- **Description:** Manual `map[string]interface{}` conversion.
- **Fix:** Use `PostPreview` DTO struct.

#### POST-PERF-3: No Body Limit

- **Description:** `json.NewDecoder(r.Body).Decode()` without size limit.
- **Fix:** Use `http.MaxBytesReader`.

#### POST-PERF-4: Inefficient Preview Creation

- **Location:** `adapters/http_handler.go`, Lines 207-232
- **Fix:** Simplify bounds checking.

### Nitpicks

- **Concurrency:** Async `context.Background()` may fail silently.
- **Error Silencing:** `GetImagePath` error ignored in `DeletePost`.
- **Duplicated Filtering Logic:** Delegate to `filterService.BuildFilter`.
- **Inconsistent JSON Naming:** `AuthorUsername` vs `Author`.
- **Case-Insensitivity Inconsistency:** Category lookups use `LOWER()`, `IN` clause does not.
- **Deeply Nested Code:** Large switch in `CreatePostAPI`. Extract parsing.
- **Rollback Logic:** Manual image deletion; orphaned files if crash.
- **Code Duplication:** Category parsing in `CreatePostAPI`/`UpdatePostAPI`. Extract `parseCategories()`.
- **Redundant Method Checks:** Go 1.22+ pattern routing makes method checks redundant.
- **Redundant `min` Function:** Go 1.21+ has built-in `min()`.
- **Verbose String Building:** Use `strings.Join(parts, " ")`.
- **Ad-hoc Logger:** Logger created per-request. Inject at construction.
- **Duplicate Page Logic:** `HomePage`/`BoardPage` share logic. Extract `renderPostList()`.
- **Validation/Error Mismatch:** Validates 255 chars, error says 300. Use constant.
- **Redundant Author Fields:** `AuthorUsername` and `Author` duplicate data.
- **Inefficient Category Validation:** Loop calls `GetByName()`. Use batch `GetByNames()`.

---

## Reaction Module

### Critical Issues

#### REACT-1: Incomplete Data in List

- **Location:** `adapters/sqlite_repository.go`, `GetByTargetPublicID`
- **Description:** Missing JOIN for `PublicUserID`; API returns empty `user_id`.
- **Fix:** JOIN `users` table for `u.public_id`.

#### REACT-2: Unmanaged Goroutines & Silent Failures

- **Location:** `application/service.go`, Lines 110, 148
- **Description:** `go func() { _ = ... }()` swallows errors, no tracking.
- **Fix:** Synchronous update or log errors.

#### REACT-3: TOCTOU Race Condition in Toggle

- **Location:** `application/service.go`, Lines 68-104
- **Description:** Check-then-act without atomicity. Concurrent requests can both insert.
- **Fix:** Use UPSERT, transaction with locking, or handle unique constraint error.

#### REACT-4: Non-Atomic Toggle Operations

- **Location:** `application/service.go`, `React`
- **Description:** Delete+Create sequence. If Create fails, reaction is lost.
- **Fix:** Wrap in transaction.

#### REACT-5: Broken Path Parsing

- **Location:** `adapters/http_handler_api.go`, Lines 243-244
- **Description:** `CountReactionsAPI` parses URL incorrectly.
- **Fix:** Use `r.PathValue("targetType")` (Go 1.22+).

#### REACT-6: Missing Duplicate Handling

- **Location:** `adapters/sqlite_repository.go`, Line 61
- **Description:** Returns raw SQLite error instead of domain error.

#### REACT-7: Missing Transactions

- **Location:** `adapters/sqlite_repository.go`, Lines 25-61
- **Description:** `Create` resolves target then inserts without transaction.
- **Fix:** Wrap in `sql.Tx`.

#### REACT-8: Stale Mock Implementations

- **Location:** `reaction/ports/service_test.go`, Lines 29-56
- **Description:** Mocks use old signatures (`targetID int` vs `targetPublicID string`).
- **Fix:** Update to match interface.

### Performance

#### REACT-PERF-1: Redundant ID Lookups

- **Description:** Two DB roundtrips to resolve target ID.
- **Fix:** Use `INSERT ... SELECT` or pass internal ID.

#### REACT-PERF-2: Duplicate ID Resolution Logic

- **Description:** Switch/case for `posts`/`comments` repeated 5+ times.
- **Fix:** Extract `resolveTargetID(ctx, publicID, targetType)`.

#### REACT-PERF-3: String-based Polymorphism

- **Description:** String discriminators prevent efficient FK constraints.
- **Fix:** Use separate link tables or optimize index.

#### REACT-PERF-4: Fragile Path Parsing

- **Description:** `strings.Split(r.URL.Path, "/")` is fragile.
- **Fix:** Use `r.PathValue()`.

### Nitpicks

- **Misleading Success Message:** Toggle-off returns "Reaction added successfully".
- **Hardcoded String Constants:** Use domain constants for `"post"`/`"comment"`.
- **Error Wrapping:** Use `fmt.Errorf("reaction service: %w", err)`.
- **Toggle Logic Complexity:** Break into `addReaction`, `removeReaction`, `switchReaction`.
- **API Path Design:** `/api/reactions/{targetType}/{targetId}` exposes implementation. Consider `/api/posts/{id}/reactions`.
- **Inconsistent Error Mapping:** Raw `sql.ErrNoRows` vs domain errors.
- **Inconsistent JSON Response:** Use `writeJSON()` instead of `fmt.Fprintf`.
- **Missing Error Context:** Wrap errors in repository.
- **Unchecked JSON Encoding:** Check error from `json.Encode()`.

---

## User Module

### Critical Issues

#### USER-1: Missing `reaction_count` Persistence

- **Location:** `adapters/sqlite_repository.go`
- **Description:** `reaction_count` in schema/domain but omitted from all SQL.
- **Fix:** Add to SELECT, INSERT, UPDATE queries.

#### USER-2: Race Condition in State Updates

- **Location:** `application/service.go`, `UpdateRole`/`DeactivateUser`/`ActivateUser`
- **Description:** Read-Modify-Write pattern. Concurrent updates overwrite each other.
- **Fix:** Atomic SQL updates or optimistic locking.

#### USER-3: Abstraction Leak

- **Location:** `application/service.go`
- **Description:** `sql.ErrNoRows` leaks to HTTP handler.
- **Fix:** Map to `domain.ErrUserNotFound`.

#### USER-4: Missing Domain Fields (OAuth)

- **Location:** `domain/user.go` vs schema
- **Description:** Schema has `oauth_provider`, `oauth_provider_id`; domain lacks them.
- **Fix:** Add to `User` struct.

#### USER-5: `HasPermission` Not Implemented

- **Location:** `user/domain/user.go`, Lines 44-48
- **Description:** Always returns `false`.
- **Fix:** Implement or remove.

### Performance

#### USER-PERF-1: Redundant DB Lookups

- **Description:** Handler fetches by PublicID, service fetches again by ID.
- **Fix:** Service methods accept `publicID` directly.

#### USER-PERF-2: Repetitive Scanning Logic

- **Location:** `adapters/sqlite_repository.go`, Lines 66-207
- **Description:** Scan logic repeated across Get methods.
- **Fix:** Extract `scanUser(scanner)` helper.

### Nitpicks

- **Validation:** `CreateUser` doesn't validate email/username format in Domain.
- **Error Messages:** No check if user exists before insert; opaque DB errors.
- **Hardcoded Ordering:** `ListUsers` hardcodes `ORDER BY created_at DESC`.
- **Semantic Content-Type:** `http.Error` uses `text/plain` but body is JSON.
- **Duplicated Role Validation:** In both Service and Handler. Keep in service only.
- **ID Security:** ✅ Correctly uses PublicID for external exposure.
- **File Headers:** ✅ Present.
- **Missing `rows.Err()`:** Line 295 (List method).
- **Missing Query Pagination:** Parse `offset`/`limit` from query string.
- **Public-to-Internal Lookup Repeated:** Extract to service accepting PublicID.
- **Missing Error Wrapping:** Use `fmt.Errorf("create user: %w", err)`.

---

## Cross-Module Issues

### CROSS-1: Fire-and-Forget Goroutines

- **Modules:** `comment`, `post`, `reaction`
- **Impact:** Data inconsistency, goroutine leaks, silent failures, no timeout.
- **Fix:** Worker pool, log errors, add timeout, or synchronous execution.

### CROSS-2: Missing `rows.Err()` Check

- **Modules:** `auth`, `comment`, `reaction`, `user`
- **Locations:** session repo, comment repo, reaction repo, user repo List.

### CROSS-3: Template Parsing on Every Request

- **Modules:** `auth`, `post`
- **Fix:** Parse once at startup.

### CROSS-4: Duplicated `buildCurrentUser`

- **Modules:** `comment`, `post`
- **Fix:** Extract to platform layer utility.

### CROSS-5: Inconsistent JSON Error Format

- **Description:** Some use `platformErrors.WriteErrorJSON()`, others `http.Error()`.
- **Fix:** Standardize on `WriteErrorJSON()`.

### CROSS-6: `fmt.Printf` for Logging

- **Modules:** `comment`
- **Fix:** Use platform logger.

### CROSS-7: Magic Numbers

- **Examples:** Comment max 5000, Post title 255, content 50000, pagination 50/20.
- **Fix:** Extract to documented constants.

### CROSS-8: Context Background in Goroutines

- **Modules:** `comment`, `reaction`, `post`
- **Fix:** Document or propagate with timeout.

### CROSS-9: Cookie Security Flags

- **Fix:** Configuration-driven `Secure` flag.

### CROSS-10: Inconsistent Error Mapping

- **Modules:** `comment`, `reaction`, `user`
- **Description:** Raw `sql.ErrNoRows` vs wrapped errors.
- **Fix:** Map DB errors to domain errors at repository boundary.

### CROSS-11: Logger Injection Inconsistent

- **Description:** Post creates per-request; reaction injects correctly.
- **Fix:** Standardize injection.

### CROSS-12: Path Parameter Extraction

- **Description:** Post uses `r.PathValue()`; reaction uses `strings.Split`.
- **Fix:** Use stdlib consistently.

### SEC-1: ID Security Policy ✓

- ✅ INT/UUID separation correctly implemented throughout.

---

## Summary of Action Items

| Priority    | Module       | Issue                                    | Complexity |
| ----------- | ------------ | ---------------------------------------- | ---------- |
| 🔴 Critical | All          | Fire-and-forget goroutines (CROSS-1)     | Medium     |
| 🔴 Critical | All          | Missing `rows.Err()` checks (CROSS-2)    | Low        |
| 🔴 Critical | Auth         | Template parsing every request (AUTH-1)  | Low        |
| 🔴 Critical | Auth         | Ignored session cleanup errors (AUTH-2)  | Low        |
| 🔴 Critical | Auth         | Weak password policy (AUTH-3)            | Low        |
| 🔴 Critical | Auth         | Cookie security hardcoded (AUTH-6)       | Low        |
| 🔴 Critical | Comment      | Unmanaged goroutines (COMMENT-1)         | Medium     |
| 🔴 Critical | Comment      | Missing `rows.Err()` (COMMENT-2)         | Low        |
| 🔴 Critical | Post         | N+1 query pattern (POST-1)               | Medium     |
| 🔴 Critical | Post         | Template parsing (POST-2)                | Low        |
| 🔴 Critical | Post         | Race condition async ops (POST-4)        | Medium     |
| 🔴 Critical | Post         | SQL performance in list (POST-PERF-1)    | Medium     |
| 🔴 Critical | Reaction     | Incomplete data in list (REACT-1)        | Low        |
| 🔴 Critical | Reaction     | TOCTOU race condition (REACT-3)          | Medium     |
| 🔴 Critical | Reaction     | Broken path parsing (REACT-5)            | Low        |
| 🔴 Critical | Reaction     | Missing transactions (REACT-7)           | Medium     |
| 🔴 Critical | Reaction     | Stale mock implementations (REACT-8)     | Low        |
| 🔴 Critical | User         | Missing reaction_count (USER-1)          | Low        |
| 🔴 Critical | User         | Race condition updates (USER-2)          | Medium     |
| 🔴 Critical | User         | SQL error abstraction leak (USER-3)      | Low        |
| 🔴 Critical | Moderation   | Module incomplete (MOD-1)                | High       |
| 🔴 Critical | Moderation   | Missing RBAC (MOD-3)                     | Medium     |
| 🔴 Critical | Notification | Module incomplete (NOTIF-1)              | High       |
| 🔴 Critical | Notification | Schema mismatch (NOTIF-2)                | Low        |
| 🔴 Critical | Notification | TDD violations (NOTIF-5)                 | Low        |
| 🟡 Medium   | Auth         | Rate limiting DoS (AUTH-4)               | Medium     |
| 🟡 Medium   | Comment      | N+1 in page handlers (COMMENT-PERF-1)    | High       |
| 🟡 Medium   | Post         | Suboptimal counting (POST-PERF-1)        | Medium     |
| 🟡 Medium   | Reaction     | Redundant ID lookups (REACT-PERF-1)      | Medium     |
| 🟡 Medium   | User         | Redundant DB lookups (USER-PERF-1)       | Medium     |
| 🟡 Medium   | User         | HasPermission not implemented (USER-5)   | Low        |
| 🟡 Medium   | All          | Inconsistent JSON error format (CROSS-5) | Low        |
| 🟡 Medium   | All          | Inconsistent error mapping (CROSS-10)    | Medium     |
| 🟡 Medium   | All          | Logger injection inconsistent (CROSS-11) | Low        |
| 🟡 Medium   | All          | Path extraction inconsistent (CROSS-12)  | Low        |
| 🟢 Low      | All          | Magic numbers (CROSS-7)                  | Low        |
| 🟢 Low      | All          | Deprecated functions (AUTH)              | Low        |
| 🟢 Low      | All          | Code duplication (CROSS-4)               | Medium     |
| 🟢 Low      | All          | `fmt.Printf` logging (CROSS-6)           | Low        |
| 🟢 Low      | Auth         | Session token entropy (AUTH-5)           | Low        |
| 🟢 Low      | Auth         | Duplicate response struct                | Low        |
| 🟢 Low      | Post         | Remove custom `min()` function           | Low        |
| 🟢 Low      | Post         | Use `strings.Join()`                     | Low        |
| 🟢 Low      | Comment      | Simplify content validation              | Low        |
| 🟢 Low      | Reaction     | Check JSON encoding errors               | Low        |
| 🟢 Low      | User         | Missing OAuth fields (USER-4)            | Low        |
| 🟢 Low      | User         | Query-based pagination                   | Low        |
| 🟢 Low      | User         | Error wrapping                           | Low        |

---

**Generated:** 2026-01-14  
**Sources:** code-review + code-simplifier documents merged and deduplicated
