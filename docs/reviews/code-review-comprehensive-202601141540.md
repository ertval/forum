# Comprehensive Code Review: Forum Modules

**Date:** 2026-01-14  
**Scope:** All internal modules (`auth`, `comment`, `moderation`, `notification`, `post`, `reaction`, `user`)  
**Reviewers:** Multiple Principal Software Engineers

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

The Forum modules demonstrate **strong adherence to hexagonal architecture** with proper separation of concerns (domain, ports, application, adapters). The codebase follows **Go idiomatic patterns** and the project's ID security policy (INT internally, UUID publicly) is consistently applied.

**Key Findings Across All Modules:**

| Category                  | Count | Severity    |
| ------------------------- | ----- | ----------- |
| Critical Issues           | 25+   | 🔴 Must Fix |
| Performance Issues        | 15+   | 🟡 Medium   |
| Nitpicks & Best Practices | 30+   | 🟢 Low      |

**Common Patterns Identified:**

- **Fire-and-Forget Goroutines**: Multiple modules spawn unmanaged goroutines for async operations without error handling or synchronization
- **Silent Error Swallowing**: Errors are explicitly ignored with `_ =` in numerous locations
- **N+1 Query Patterns**: Page handlers and repositories fetch related data in loops instead of batch operations
- **Template Parsing on Every Request**: Performance killer in page handlers
- **Missing `rows.Err()` Checks**: After database iterations in repositories

---

## Auth Module

### Critical Issues

#### AUTH-1: Template Parsing on Every Request (Performance Killer)

- **Location:** `adapters/http_handler_page.go`, `LoginPage` (line 25), `RegisterPage` (line 44)
- **Probability:** High (100% of page load requests)
- **Description:** The code calls `template.ParseFiles` on every single HTTP request. This triggers disk I/O and expensive parsing logic, severely limiting throughput and increasing latency.
- **Proposed Fix:** Parse templates _once_ during application startup in `NewHTTPHandler` or `main.go` and store them in the `HTTPHandler` struct. Use `t.Lookup("login.html").Execute(...)` to render.

#### AUTH-2: Ignored Errors in Session Cleanup

- **Location:** `application/service.go`, `Login` (line 147-151), `ValidateSession` (line 194)
- **Probability:** Medium
- **Description:**
  - In `Login`, `s.sessionRepo.DeleteByUserID` is called to invalidate old sessions. The error is explicitly ignored.
  - In `ValidateSession`, `s.sessionRepo.Delete` is called for expired sessions, and the error is ignored (`_ = ...`).
  - **Risk:** If the DB is locked or failing, these write operations fail silently. In `Login`, this could mean a user ends up with multiple active sessions against the "one session per user" policy.
- **Proposed Fix:**
  ```go
  if err := s.sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
      log.Printf("WARNING: failed to delete existing sessions for user %d: %v", user.ID, err)
  }
  ```

#### AUTH-3: Weak Password Policy

- **Location:** `application/service.go`, `ValidateCredentials` (line 284)
- **Probability:** High (Security Risk)
- **Description:** The password minimum length is set to 6 characters. This is insufficient for modern security standards.
- **Proposed Fix:** Increase minimum length to at least 8, preferably 12, or implement complexity requirements.

#### AUTH-4: Registration DoS via Bcrypt

- **Location:** `application/service.go` (`Register`)
- **Probability:** Medium
- **Description:** `Register` performs a bcrypt hash (expensive operation) synchronously in the request handler. An attacker could flood the `/register` endpoint with unique emails/usernames to exhaust CPU resources.
- **Proposed Fix:** Implement rate limiting on the `/api/auth/register` and `/api/auth/login` endpoints at the infrastructure or middleware level.

#### AUTH-5: Session Token Entropy

- **Location:** `application/service.go`
- **Probability:** Low
- **Description:** Session tokens are UUID v4. While generally unique, UUIDs are not designed to be cryptographically secure session identifiers (entropy is 122 bits).
- **Proposed Fix:** Use `crypto/rand` to generate a 32-byte random string (base64 encoded) for session tokens instead of UUIDs to ensure maximum entropy and resistance to prediction.

### Performance & Optimization

#### AUTH-PERF-1: Redundant User Fetching

- **Description:** `RegisterAPI` and `LoginAPI` call their respective service methods, then call `userService.GetByID` again just to get the `PublicID`.
- **Optimized Code:** Update `authService.Register` and `authService.Login` to return the full `User` entity or at least the `PublicID`.

#### AUTH-PERF-2: Duplicate User Lookups in Middleware

- **Description:** `OptionalAuth` and `RequireAuth` middleware fetch the full user entity from the database on every single request to populate the context.
- **Optimized Code:** Store `PublicID` in the `sessions` table (denormalization) to avoid the second DB roundtrip to the `users` table.

### Nitpicks & Best Practices

- **API Robustness:** In `RegisterAPI` (line 41), the fallback `strings.Contains` logic (lines 53-60) is brittle. New validation error strings could break this mapping.
- **Cookie Security:** The `Secure` flag is explicitly set to `false`. This MUST be conditional based on the environment (e.g., via config flag `IsProduction`). Hardcoding `false` is dangerous if deployed to prod.
- **Deprecated Functions:** `adapters/middleware.go` contains deprecated wrappers (`RequireAuthFunc`, `OptionalAuthFunc`, `GetUserID`, `GetUsername`, `IsAuthenticated`). These should be removed or have a deprecation timeline.
- **Transaction Safety:** `Register` operation involves multiple steps (User creation, Session creation). These are currently separate DB calls. If `sessionRepo.Create` fails after `userService.CreateUser` succeeds, the user exists but the client gets an error.

---

## Comment Module

### Critical Issues

#### COMMENT-1: Unmanaged Goroutines & Ignored Errors

- **Location:** `application/service.go`, Lines 59-61, 111-113
- **Probability:** Medium
- **Description:** `CreateComment` and `DeleteComment` spawn detached goroutines (`go func() { ... }`) to update user statistics. These goroutines inherit `context.Background()` but are not tracked by any WaitGroup. Errors are explicitly ignored (`_ = ...`), so failures to update stats will go unnoticed.
- **Proposed Fix:** Use a worker pool or a proper background job system. For a simple fix:
  ```go
  go func(uid int) {
      if err := s.userService.IncrementCommentCount(context.Background(), uid); err != nil {
          log.Printf("WARNING: failed to increment comment count for user %d: %v", uid, err)
      }
  }(userID)
  ```

#### COMMENT-2: Missing `rows.Err()` Check

- **Location:** `adapters/sqlite_repository.go`, Lines 120, 160, 201
- **Probability:** Medium
- **Description:** When iterating over database rows using `for rows.Next() { ... }`, it is mandatory to check `rows.Err()` afterwards to catch any errors that occurred during iteration.
- **Proposed Fix:**
  ```go
  for rows.Next() {
      // ... scan ...
  }
  if err := rows.Err(); err != nil {
      return nil, err
  }
  return comments, nil
  ```

#### COMMENT-3: Missing Transaction Boundary

- **Location:** `application/service.go` (`CreateComment`)
- **Probability:** Low
- **Description:** When a comment is created, we insert the comment record AND increment the user's comment count. These are two separate DB operations. If the second fails, the data is inconsistent.
- **Proposed Fix:** Accept acceptable inconsistency, make updates synchronous and log errors, or use an event-driven architecture.

### Performance & Optimization

#### COMMENT-PERF-1: N+1 Query Problem in Page Handlers

- **Location:** `adapters/http_handler_page.go`, Lines 89-167
- **Description:** The `MyCommentsPage` handler loops through each comment and performs multiple synchronous blocking calls to other services (`userService.GetByID`, `postService.GetPost`, `reactionService.CountReactions`). If a user has 20 comments, this results in 60+ additional database queries per page load.
- **Optimized Code:** Implement "Batch Get" methods in the respective services (e.g., `GetPostsByIDs([]string)`) to fetch all related data in a single query per entity type.

#### COMMENT-PERF-2: Offset Pagination

- **Description:** `ListByUserPaginated` uses `LIMIT ? OFFSET ?`. As the offset grows, the DB must scan and discard `OFFSET` rows.
- **Optimized Code:** Use Keyset Pagination (Seek Method), filtering by `created_at < ?` of the last item on the previous page.

#### COMMENT-PERF-3: Recreating Logger on Every Error

- **Location:** `adapters/http_handler.go`, Lines 129-132
- **Description:** `writeJSON` creates a new `logger.Config` and `logger.Logger` instance every time a JSON encoding error occurs.
- **Optimized Code:** Initialize the logger once in the `HTTPHandler` struct and reuse it.

### Nitpicks & Best Practices

- **Code Duplication:** `LoadMoreCommentsAPI` and `MyCommentsPage` share significant logic for enriching comment data. Extract into a helper method like `enrichComments(ctx, comments)`.
- **Complex Handler:** `MyCommentsPage` is doing too much: auth validation, param parsing, fetching data, filtering data, and rendering. It violates SRP.
- **Silent Failures:** Errors from `h.categoryService.List` are logged to stdout (`fmt.Printf`) but execution continues, potentially rendering a broken page.
- **RESTful Design:** Routes like `POST /api/comments/posts/{post_id}` are slightly inconsistent with `GET /api/comments/{id}`. Standard REST prefers `POST /api/posts/{post_id}/comments` for nested resources.

---

## Moderation Module

> **Note:** This module is currently in a **scaffold state** and marked as `[OPTIONAL FEATURE]`. Most methods are placeholders.

### Critical Issues

#### MOD-1: Module is Incomplete (Functional GAP)

- **Location:** `application/service.go`, `adapters/sqlite_repository.go`, `adapters/http_handler_api.go`
- **Probability:** High (Guaranteed failure)
- **Description:** All primary use cases (creating reports, reviewing reports, listing reports) are unimplemented. Service methods return `nil` or empty results, and API handlers return `501 Not Implemented`.
- **Proposed Fix:** Implement the logic described in `internal/modules/moderation/flow.md` when the feature is prioritized.

#### MOD-2: Silent Placeholder Failure in Service

- **Location:** `application/service.go`, Lines 23-30
- **Probability:** High
- **Description:** The `CreateReport` method returns `nil` (success) without performing any validation or persistence. If called, it will falsely indicate success while the report is lost.
- **Proposed Fix:**
  ```go
  func (s *Service) CreateReport(...) error {
      return errors.New("method CreateReport not implemented")
  }
  ```

#### MOD-3: Missing Role-Based Access Control (RBAC)

- **Location:** `adapters/http_handler_api.go`
- **Probability:** High (Security)
- **Description:** Moderation endpoints (List, Review) are registered without any authorization middleware. These endpoints should only be accessible by users with `Moderator` or `Administrator` roles.
- **Proposed Fix:** Implement a `RequireRole` middleware and apply it to sensitive routes.

#### MOD-4: Missing Cross-Module Integration for Content Validation

- **Location:** `application/service.go`:23
- **Probability:** High
- **Description:** `CreateReport` receives a `targetPublicID` but there is no logic to verify if the post or comment actually exists. A user could spam reports for non-existent content.
- **Proposed Fix:** Inject `post.PostService` and `comment.CommentService` into `moderation.Service` to validate targets.

#### MOD-5: Incomplete Domain Validation

- **Location:** `domain/report.go`, Line 31
- **Probability:** Low
- **Description:** The `IsValid()` method only checks the `TargetType`. It should also validate the `Reason` (not empty) and the `Status` (matches defined constants).
- **Proposed Fix:**
  ```go
  func (r *Report) IsValid() bool {
      if r.Reason == "" { return false }
      if r.TargetType != "post" && r.TargetType != "comment" { return false }
      switch r.Status {
      case StatusPending, StatusReviewed, StatusResolved:
          return true
      default:
          return false
      }
  }
  ```

#### MOD-6: Major Discrepancy Between Documentation and Implementation

- **Location:** `flow.md` vs multiple implementation files
- **Probability:** High (misleads developers)
- **Description:** `flow.md` describes a complete implementation including `RequireRole` middleware, specific SQL queries, and cross-module cascading deletes. However, the actual code contains only empty placeholders. This creates a false sense of module readiness.
- **Proposed Fix:** Synchronize `flow.md` with the current state of implementation.

### Performance & Optimization

#### MOD-PERF-1: N+1 Hazard in Report Listing

- **Description:** When `ListReports` is implemented, it will likely need to enrich reports with Reporter/Moderator usernames for the UI. Doing this in a loop will result in N+1 queries.
- **Optimized Code:**
  ```sql
  SELECT r.*, u.username as reporter_username, m.username as moderator_username
  FROM reports r
  LEFT JOIN users u ON r.reporter_id = u.id
  LEFT JOIN users m ON r.moderator_id = m.id
  ```

#### MOD-PERF-2: External ID Resolution Bottleneck

- **Description:** The `CreateReport` service accepts a `targetPublicID` (UUID). Resolving this to an internal integer ID requires calling `PostService` or `CommentService`.
- **Optimization:** Ensure that `PostService` and `CommentService` have efficient `GetInternalIDByPublicID` methods to avoid full entity fetches just for ID translation.

### Nitpicks & Best Practices

- **Inconsistent Status Naming:** `domain/report.go` defines `StatusResolved`, but `flow.md` uses `dismissed`. Stick to one terminology.
- **TargetID Type:** The `TargetID` is an `int` (internal ID). Ensure that when resolving `targetPublicID` to `TargetID`, the correct table is queried based on `TargetType`. This "polymorphic association" can be fragile.
- **Missing Dependency Injection for Target Resolution:** The service will need `PostService` and `CommentService` ports. These are not yet injected.

---

## Notification Module

> **Note:** This module is currently in a **scaffold state** and marked as `[OPTIONAL FEATURE]`. Most methods are placeholders.

### Critical Issues

#### NOTIF-1: Incomplete Implementation (Module Scaffold Only)

- **Location:** `application/service.go`, `adapters/sqlite_repository.go`, `adapters/http_handler_api.go`
- **Probability:** High (Functional failure)
- **Description:** The core notification logic is not implemented. Methods like `CreateNotification`, `GetByUserID`, and `MarkAsReadByPublicID` are placeholders. No notifications are actually generated or persisted.
- **Proposed Fix:** Implement the logic or return `errors.New("not implemented")` to fail fast.

#### NOTIF-2: Domain Model and Database Schema Mismatch

- **Location:** `domain/notification.go` vs `migrations/007_notification_create_notifications.sql`
- **Probability:** High
- **Description:** The `Notification` struct is missing the `ActorID` field which is defined as `NOT NULL` in the database schema. Any attempt to insert data will fail or lead to data loss.
- **Proposed Fix:**
  ```go
  type Notification struct {
      // ... existing fields
      ActorID       int    `json:"-"`
      PublicActorID string `json:"actor_id,omitempty"`
  }
  ```

#### NOTIF-3: Missing Validation in Service Layer

- **Location:** `application/service.go`, Line 23
- **Probability:** Medium
- **Description:** The `CreateNotification` method has a `TODO` for validation but currently does nothing. It should validate the `notifType` against allowed constants.

#### NOTIF-4: Target ID Resolution Strategy

- **Location:** `application/service.go`, Line 26
- **Description:** The `notification` module accepts a `targetPublicID` (string/UUID) but the domain entity requires a `TargetID` (int). The notification module should not directly have access to `post` or `comment` tables.
- **Recommendation:**
  - **Option A (Loose Coupling):** Change `TargetID` to `string` and store the UUID directly.
  - **Option B (Service Dependencies):** Inject `PostService` or `CommentService` to resolve IDs.

### Performance & Optimization

#### NOTIF-PERF-1: Polling vs Push

- **Description:** The current design (`GET /api/notifications`) implies a polling architecture. If client-side polling is aggressive, this will pound the database.
- **Optimization:** Consider Server-Sent Events (SSE) or WebSockets. If polling, ensure `List` endpoints support "Since ID" or "After Timestamp" for incremental sync.

#### NOTIF-PERF-2: Sync vs Async Notification Creation

- **Description:** When fully implemented, `CreateNotification` might be called synchronously by critical paths. If the notification system is slow, it will slow down user actions.
- **Optimized Code:** Use a worker pool or a background goroutine (fire-and-forget) for creating notifications, or ensure the DB write is extremely fast.

### Nitpicks & Best Practices

- **Use `ErrNotImplemented` for Placeholders:** Instead of returning `nil`, return a specific error during development to avoid silent failures.
- **TargetID Resolution:** Like the `moderation` module, resolving these for the UI will require polymorphic queries or multiple round trips.
- **Database Indexing:** The current migration provides good indexing with `idx_notifications_user` on `(user_id, read)`.

---

## Post Module

### Critical Issues

#### POST-1: N+1 Query Pattern in Repository

- **Location:** `adapters/sqlite_repository.go`, `List` method (Line 432)
- **Probability:** High
- **Description:** The `List` method executes a separate SQL query to fetch categories for _every_ post in the result set. For a list of 50 posts, this results in 51 database roundtrips.
- **Proposed Fix:**
  ```go
  // Better approach: Join and group categories in the main query
  query := `SELECT p.*, GROUP_CONCAT(c.name) as category_names ...`
  ```
  Or fetch all categories for retrieved posts in one query using `IN` clause with post IDs.

#### POST-2: Inefficient Template Parsing

- **Location:** `adapters/http_handler_page.go`, multiple handlers (Lines 130, 249, 346, 401, 476)
- **Probability:** High
- **Description:** HTML templates are parsed from disk on _every_ request. File I/O and template parsing are expensive operations that should happen once during application startup.
- **Proposed Fix:** Parse templates once in `NewHTTPHandler` and store the `*template.Template` in the handler struct.

#### POST-3: Redundant Authentication Logic

- **Location:** `adapters/http_handler_page.go` (Lines 39-46, 154-161, 292-299)
- **Probability:** Medium
- **Description:** Handlers manually parse the `session_token` cookie and call `authService.ValidateSession` even when not protected by middleware. This bypasses the centralized middleware logic and increases the risk of inconsistent auth state.
- **Proposed Fix:** Use the `OptionalAuth` middleware for public pages to populate the user context.

#### POST-4: Race Condition in CreatePost/DeletePost Async Operations

- **Location:** `application/service.go` (`CreatePost`, `DeletePost`)
- **Probability:** Medium
- **Description:** The user's post count is incremented/decremented asynchronously in a goroutine (Line 131, Line 196).
  ```go
  go func() {
      _ = s.userService.IncrementPostCount(context.Background(), userID)
  }()
  ```
  If the application crashes after the HTTP handler returns but before the goroutine executes, the post count will be out of sync.
- **Proposed Fix:** Perform the count update synchronously within the request context, or use a reliable event bus/queue.

#### POST-5: SQL Injection Risk in Dynamic Query Building

- **Location:** `adapters/sqlite_repository.go` (Line 329)
- **Probability:** Low (due to internal usage) but risky pattern
- **Description:** `repeatPlaceholders` is used to build the `IN` clause for category filtering. While `?` placeholders are used, manually constructing SQL strings is fragile.
- **Proposed Fix:** Use a query builder or ensure rigorous testing of the dynamic SQL generation.

### Performance & Optimization

#### POST-PERF-1: Suboptimal Counting in Queries

- **Location:** `adapters/sqlite_repository.go`, `GetByID` and `List` (Lines 103-119, 295-311)
- **Description:** Counts for likes, dislikes, and comments are calculated using subqueries for every request. As the tables grow, these joins will become expensive.
- **Optimized Code:** Cache counts in the `posts` table (`like_count`, `comment_count` columns) and update via triggers or application logic.

#### POST-PERF-2: Excessive Field Mapping and Allocations

- **Location:** `adapters/http_handler_api.go` (Lines 545-566) and `http_handler_page.go` (Lines 81-102)
- **Description:** Posts are manually converted to `map[string]interface{}` then to JSON/Template data. This creates many small heap allocations.
- **Optimized Code:** Define a `PostPreview` domain struct or DTO struct that can be serialized directly.

#### POST-PERF-3: JSON Decoding without Body Limit

- **Description:** In `CreatePostAPI` and `UpdatePostAPI`, `json.NewDecoder(r.Body).Decode(&req)` is used. If the body is large, this may be vulnerable.
- **Optimization:** Use `http.MaxBytesReader` to enforce a hard limit on the request body size for JSON requests.

### Nitpicks & Best Practices

- **Concurrency Robustness:** The async increment/decrement using `context.Background()` may fail silently or leak if the application shuts down before completion.
- **Error Silencing:** In `DeletePost`, `s.postRepo.GetImagePath` error is ignored. Logging would help identify storage issues.
- **Duplicated Filtering Logic:** Logic to build filters from query parameters is duplicated. Delegate fully to `filterService.BuildFilter`.
- **Inconsistent JSON Naming:** `domain/post.go` uses `AuthorUsername` but also has a compatible `Author` field.
- **Case-Insensitivity Inconsistency:** Category lookups use `LOWER(name)`, but `PostFilter` implementation for `IN` clause does not.
- **Deeply Nested Code:** `CreatePostAPI` has a very large `switch` statement for content types. Extract the request parsing logic into a separate private method.
- **Rollback Logic:** The "rollback" logic in `CreatePost` (delete image if DB fails) is manual. If the server crashes, orphaned files may remain.

---

## Reaction Module

### Critical Issues

#### REACT-1: Incomplete Data in Reaction List (Functional)

- **Location:** `adapters/sqlite_repository.go` (`GetByTargetPublicID`), Line 139
- **Probability:** High (100% reproducible)
- **Description:** The query fetches from `reactions` table but fails to JOIN `users` table to populate the `PublicUserID` field. The API key `user_id` will be missing or empty in the response.
- **Proposed Fix:**
  ```go
  selectQuery := `
      SELECT r.id, r.public_id, r.user_id, r.target_id, r.target_type, r.type, r.created_at, u.public_id
      FROM reactions r
      JOIN users u ON r.user_id = u.id
      WHERE r.target_id = ? AND r.target_type = ?
      ORDER BY r.created_at DESC
  `
  // Scan into &reaction.PublicUserID as well
  ```

#### REACT-2: Unmanaged Goroutines & Silent Failures

- **Location:** `application/service.go`, Lines 110 & 148
- **Probability:** Medium (High risk at scale)
- **Description:** The service uses `go func() { _ = s.userService.IncrementReactionCount(...) }()` to update user stats.
  1. **Data Consistency:** Errors are explicitly swallowed. If the DB update fails, the user's total reaction count remains incorrect indefinitely.
  2. **Resource Management:** These goroutines are not tracked.
- **Proposed Fix:** Run the update synchronously:
  ```go
  if err := s.userService.IncrementReactionCount(ctx, userID); err != nil {
      return fmt.Errorf("failed to update user stats: %w", err)
  }
  ```

#### REACT-3: Non-Atomic Reaction Toggle (Race Condition)

- **Location:** `application/service.go`, Lines 69-91
- **Probability:** Medium
- **Description:** The `React` method implements a "toggle" logic by first checking for an existing reaction and then performing delete/insert. This is not atomic. Concurrent requests might see no reaction exists and both attempt to create one. The database unique constraint will prevent duplicates, but one request will fail with a raw SQL error.
- **Proposed Fix:** Use a single "UPSERT" style operation, wrap in a database transaction, or handle the unique constraint error and map it to `domain.ErrDuplicateReaction`.

#### REACT-4: Lack of Atomicity for Toggle Operations

- **Location:** `application/service.go`, Method `React`
- **Probability:** Low
- **Description:** Changing a reaction type (e.g., from like to dislike) involves `DeleteByTargetPublicID` followed by `Create`. If `Create` fails, the user's previous reaction is lost entirely.
- **Proposed Fix:** Wrap the "switch reaction" operations in a transaction.

#### REACT-5: Broken Path Parsing in CountReactionsAPI

- **Location:** `adapters/http_handler_api.go`, Lines 243-244
- **Probability:** High
- **Description:** The `CountReactionsAPI` handler parses the target type and ID incorrectly from the URL path. For the route `GET /api/reactions/{targetType}/{targetId}/count`, `strings.Split` with `/` results in incorrect indices.
- **Proposed Fix:**
  ```go
  // Correct parsing:
  targetType := pathParts[len(pathParts)-3]
  targetID := pathParts[len(pathParts)-2]
  ```
  Or use Go 1.22+ `r.PathValue("targetType")` and `r.PathValue("targetId")`.

#### REACT-6: Missing Duplicate Handling

- **Location:** `adapters/sqlite_repository.go`, Line 61
- **Probability:** Low (Race condition specific)
- **Description:** While the service checks for existing reactions, race conditions can still occur. The repository's `Create` method returns the raw SQLite error instead of checking for a unique constraint violation.

### Performance & Optimization

#### REACT-PERF-1: Redundant ID Lookups

- **Description:** The `Create` workflow involves two separate DB round-trips to resolve the target ID:
  1. Service calls `postRepo.GetByID` (resolves Target)
  2. Repository calls `SELECT id FROM posts` (resolves Target again)
- **Optimized Code:**
  ```go
  // Use INSERT with EXISTS check (single query)
  query := `
      INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at)
      SELECT ?, ?, id, ?, ?, CURRENT_TIMESTAMP
      FROM posts WHERE public_id = ?
  `
  ```

#### REACT-PERF-2: Duplicate ID Resolution Logic

- **Location:** `adapters/sqlite_repository.go`
- **Description:** Every method in the repository repeats the switch/case logic to query `posts` or `comments` tables to resolve `public_id` -> `id`. This violates DRY (repeated 5+ times).
- **Proposed Fix:** Extract into a private helper method `resolveTargetID(ctx, publicID, targetType) (int, error)`.

#### REACT-PERF-3: String-based Polymorphism

- **Description:** Using `"post"` and `"comment"` strings as discriminators requires string comparisons and prevents efficient foreign key constraints.
- **Optimized Model:** Use separate link tables `post_reactions` and `comment_reactions`, OR optimize the index to be `(target_type, target_id, user_id)`.

#### REACT-PERF-4: Fragile Path Parsing in API handlers

- **Description:** Handlers use `strings.Split(r.URL.Path, "/")` to extract parameters. This is fragile and unnecessary.
- **Optimized Code:** Use `r.PathValue("targetType")` and `r.PathValue("targetId")`.

### Nitpicks & Best Practices

- **Obsolete Path Parsing:** Use Go 1.24's `r.PathValue()` instead of manual path splitting.
- **Misleading Success Message on Toggle:** When a reaction is "toggled off", the API still returns `{"message": "Reaction added successfully"}`.
- **Hardcoded String Constants:** Use domain constants for `post` and `comment` strings to avoid typos.
- **Error Wrapping:** Errors from the repository are returned directly without wrapping. Use `fmt.Errorf("reaction service: %w", err)`.
- **Hardcoded SQL Strings:** Consider moving them to constants if they grow complex.
- **Toggle Logic Complexity:** The `React` method handles creating, removing, and switching in one function. Breaking down into `addReaction`, `removeReaction`, `switchReaction` would improve testability.
- **API Path Design:** The paths `/api/reactions/{targetType}/{targetId}` expose implementation details. A more RESTful approach might be `/api/posts/{id}/reactions`.

---

## User Module

### Critical Issues

#### USER-1: Incomplete Data Persistence (Missing reaction_count)

- **Location:** `adapters/sqlite_repository.go`, multiple lines (Scan, Insert, Update)
- **Probability:** High
- **Description:** While the `domain.User` struct and the SQLite table schema both include `reaction_count`, the repository's `SELECT`, `INSERT`, and `UPDATE` queries completely omit this field. This means the user's reaction count is never loaded into memory, and updates do not persist this value.
- **Proposed Fix:** Update all SQL queries in `sqlite_repository.go` to include the `reaction_count` column.

#### USER-2: Race Condition in State Updates

- **Location:** `application/service.go` (`UpdateRole`, `DeactivateUser`, `ActivateUser`)
- **Probability:** Low (Admin actions), Medium (General updates)
- **Description:** The Update operations follow a "Read-Modify-Write" pattern:
  ```go
  user, _ := repo.GetByID(id) // Read
  user.IsActive = false       // Modify
  repo.Update(user)           // Write
  ```
  If two concurrent requests update different fields, the second will overwrite the first's changes.
- **Proposed Fix:** Use atomic SQL updates or implement optimistic locking:
  ```go
  func (r *SQLiteUserRepository) UpdateStatus(ctx context.Context, userID int, isActive bool) error {
      query := `UPDATE users SET is_active = ?, updated_at = ? WHERE id = ?`
      // ...
  }
  ```

#### USER-3: Abstraction Leak (SQL Errors)

- **Location:** `application/service.go`
- **Probability:** High
- **Description:** The Service layer returns errors directly from the repository. If the repository returns `sql.ErrNoRows`, this driver-specific error leaks to the HTTP handler, violating architectural boundaries.
- **Proposed Fix:**
  ```go
  // In Service.GetByID
  user, err := s.userRepo.GetByID(ctx, userID)
  if err == sql.ErrNoRows {
      return nil, domain.ErrUserNotFound
  }
  return user, err
  ```

#### USER-4: Missing Domain Fields (OAuth)

- **Location:** `domain/user.go` vs `migrations/002_user_create_users.sql`
- **Probability:** Low (Future-proofing)
- **Description:** The database schema includes `oauth_provider` and `oauth_provider_id`, but these are missing from the `domain.User` entity.
- **Proposed Fix:** Add `OAuthProvider` and `OAuthProviderID` fields to the `User` struct.

### Performance & Optimization

#### USER-PERF-1: Redundant Database Lookups in API Handlers

- **Description:** Each update API fetches the user by `PublicID` in the handler to find the internal `ID`, then the service layer fetches the user _again_ by `ID`. This results in 2 reads and 1 write for every update.
- **Optimized Code:**
  ```go
  func (s *Service) UpdateRole(ctx context.Context, publicID string, newRole domain.Role) error {
      user, err := s.userRepo.GetByPublicID(ctx, publicID)
      // ... update and save
  }
  ```

#### USER-PERF-2: SQLite Boolean Handling

- **Description:** The `Scan` logic in `sqlite_repository.go` repeatedly handles the `isActive` int-to-bool conversion manually in every method.
- **Optimized Code:** Consolidate the scanning logic into a helper method to reduce code duplication.

### Nitpicks & Best Practices

- **Validation:** `CreateUser` in the Service does not validate email format or username length. Add this in the Domain layer.
- **Error Messages:** `CreateUser` does not check if the user already exists before attempting insertion. Relying solely on DB uniqueness constraints creates opaque errors.
- **Hardcoded Ordering:** `ListUsers` hardcodes `ORDER BY created_at DESC`. Allow sorting options in the future.
- **Semantic Content-Type:** `http.Error` defaults to `text/plain`, but response body is JSON. Set `Content-Type` to `application/json`.
- **Duplicated Role Validation Logic:** Exists in both Service and HTTP Handler. Unify in the Domain layer.
- **ID Security Rule Compliance:** ✅ The code correctly follows the instruction of using `PublicID` (UUID) for external exposure.
- **File Headers:** ✅ All files correctly include the required headers.
- **Missing `rows.Err()` Check:** In `sqlite_repository.go`, Line 295 (List method)

---

## Cross-Module Issues

These issues appear across multiple modules and should be addressed systematically.

### CROSS-1: Fire-and-Forget Goroutines Without Error Handling or Synchronization

- **Affected Modules:** `comment`, `post`, `reaction`
- **Pattern:**
  ```go
  go func() {
      _ = s.userService.IncrementCommentCount(context.Background(), userID)
  }()
  ```
- **Impact:** Data consistency issues, unbounded goroutines in high-load scenarios, silent failures
- **Recommended Fix:** Use a worker pool pattern or at minimum log errors:
  ```go
  go func(uid int) {
      ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
      defer cancel()
      if err := s.userService.IncrementCommentCount(ctx, uid); err != nil {
          log.Printf("WARNING: failed to increment comment count: %v", err)
      }
  }(userID)
  ```

### CROSS-2: Missing `rows.Err()` Check After Iteration

- **Affected Modules:** `auth`, `comment`, `reaction`, `user`
- **Locations:**
  - `auth/adapters/sqlite_session_repository.go` (GetByUserID)
  - `comment/adapters/sqlite_repository.go` (ListByPostPublicID, ListByUser, ListByUserPaginated)
  - `reaction/adapters/sqlite_repository.go` (GetByTargetPublicID)
  - `user/adapters/sqlite_repository.go` (List)

### CROSS-3: Template Parsing on Every Request

- **Affected Modules:** `auth`, `post`
- **Impact:** Severe performance degradation on page load requests
- **Recommended Fix:** Parse templates once during application startup and store in handler struct.

### CROSS-4: Duplicated `buildCurrentUser` Implementation

- **Affected Modules:** `comment`, `post`
- **Location:** `comment/adapters/http_handler.go` (Lines 59-92) and `post/adapters/http_handler.go` (Lines 66-99)
- **Recommended Fix:** Extract to a shared utility in the platform layer.

### CROSS-5: Inconsistent JSON Error Response Format

- **Affected Modules:** Various handlers
- **Description:** Some handlers use `platformErrors.WriteErrorJSON()` while others use `http.Error()`.
- **Recommended Fix:** Standardize on `platformErrors.WriteErrorJSON()` for all API handlers.

### CROSS-6: `fmt.Printf` Used for Logging

- **Affected Modules:** `comment`
- **Location:** `comment/adapters/http_handler_page.go`, Lines 59, 81, 194
- **Recommended Fix:** Use the platform logger for consistency.

### CROSS-7: Magic Numbers and Hardcoded Values

- **Locations:**
  - Comment max length: 5000 (`comment/domain/comment.go:36`)
  - Post title max: 255 (`post/domain/post.go:31`)
  - Post content max: 50000 (`post/domain/post.go:35`)
  - Session token cookie name: "session_token"
- **Recommended Fix:** Extract to constants with clear documentation.

### CROSS-8: Context Background in Goroutines

- **Affected Modules:** `comment`, `reaction`
- **Description:** Using `context.Background()` in fire-and-forget goroutines bypasses any parent context deadlines. This is intentional but should be documented.
- **Recommended Fix:** Add comments explaining the intentional use of fresh context.

### CROSS-9: Cookie Security Flags

- **Location:** `auth/adapters/http_handler_api.go`, Lines 73, 127, 176
- **Description:** Cookies have `Secure: false` with a comment about production. This MUST be configuration-driven.

### SEC-1: ID Security Policy Well-Implemented ✓

- **Observation:** The codebase correctly implements the INT/UUID separation:
  - Internal IDs (`ID int`) are never exposed in JSON (`json:"-"`)
  - Public UUIDs (`PublicID string`) are used in URLs and responses
  - Context stores PublicID (UUID), middleware correctly translates
  - Good security comments throughout

---

## Summary of Action Items

| Priority    | Module       | Issue                                         | Fix Complexity |
| ----------- | ------------ | --------------------------------------------- | -------------- |
| 🔴 Critical | All          | Fire-and-forget goroutines (CROSS-1)          | Medium         |
| 🔴 Critical | All          | Missing `rows.Err()` checks (CROSS-2)         | Low            |
| 🔴 Critical | Auth         | Template parsing on every request (AUTH-1)    | Low            |
| 🔴 Critical | Auth         | Ignored errors in session cleanup (AUTH-2)    | Low            |
| 🔴 Critical | Auth         | Weak password policy (AUTH-3)                 | Low            |
| 🔴 Critical | Comment      | Unmanaged goroutines (COMMENT-1)              | Medium         |
| 🔴 Critical | Comment      | Missing `rows.Err()` (COMMENT-2)              | Low            |
| 🔴 Critical | Post         | N+1 query pattern (POST-1)                    | Medium         |
| 🔴 Critical | Post         | Template parsing (POST-2)                     | Low            |
| 🔴 Critical | Post         | Race condition in async ops (POST-4)          | Medium         |
| 🔴 Critical | Reaction     | Incomplete data in list (REACT-1)             | Low            |
| 🔴 Critical | Reaction     | Non-atomic toggle (REACT-3)                   | Medium         |
| 🔴 Critical | Reaction     | Broken path parsing (REACT-5)                 | Low            |
| 🔴 Critical | User         | Missing reaction_count (USER-1)               | Low            |
| 🔴 Critical | User         | Race condition in updates (USER-2)            | Medium         |
| 🔴 Critical | Moderation   | Module incomplete (MOD-1)                     | High           |
| 🔴 Critical | Moderation   | Missing RBAC (MOD-3)                          | Medium         |
| 🔴 Critical | Notification | Module incomplete (NOTIF-1)                   | High           |
| 🔴 Critical | Notification | Schema mismatch (NOTIF-2)                     | Low            |
| 🟡 Medium   | Auth         | Rate limiting for DoS (AUTH-4)                | Medium         |
| 🟡 Medium   | Comment      | N+1 queries in page handlers (COMMENT-PERF-1) | High           |
| 🟡 Medium   | Post         | Suboptimal counting (POST-PERF-1)             | Medium         |
| 🟡 Medium   | Reaction     | Redundant ID lookups (REACT-PERF-1)           | Medium         |
| 🟡 Medium   | User         | Redundant DB lookups (USER-PERF-1)            | Medium         |
| 🟡 Medium   | All          | Inconsistent JSON error format (CROSS-5)      | Low            |
| 🟢 Low      | All          | Magic numbers (CROSS-7)                       | Low            |
| 🟢 Low      | All          | Deprecated functions (AUTH-deprecated)        | Low            |
| 🟢 Low      | All          | Code duplication (CROSS-4)                    | Medium         |
| 🟢 Low      | All          | `fmt.Printf` for logging (CROSS-6)            | Low            |

---

**Generated:** 2026-01-14 15:40  
**Source Reviews:** 20 individual review files merged and deduplicated
