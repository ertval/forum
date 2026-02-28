# Code Review: Reaction, Moderation, and Notification Modules

**Date**: 2026-02-28  
**Scope**: Full review of all files in `internal/modules/{reaction,moderation,notification}/`  
**Principles**: Idiomatic Go, KISS, simplification, correctness

---

## Table of Contents

1. [Reaction Module](#1-reaction-module)
2. [Moderation Module](#2-moderation-module)
3. [Notification Module](#3-notification-module)
4. [Cross-Module Issues](#4-cross-module-issues)

---

## 1. Reaction Module

### 1.1 `domain/reaction.go` (54 lines)

**Structure**: Defines `ReactionType` (string enum), `Reaction` struct, and `Validate()`.

**Issues**:

**1. Validate() has mixed-mode logic (L47-51)** — It checks `r.TargetID <= 0 && r.PublicTargetID == ""`, mixing internal and public ID validation. Domain validation should not care about which ID transport layer is used.

```go
// CURRENT (L47-51): confusing dual-ID check
if r.TargetID <= 0 && r.PublicTargetID == "" {
    return ErrInvalidTargetID
}

// SUGGESTED: Domain should validate one concept
if r.TargetID <= 0 {
    return ErrInvalidTargetID
}
```

**2. `ReactionType` as `string` is fine, but `TargetType` is a raw `string`** — Inconsistency. If `ReactionType` gets a named type, `TargetType` should too, or neither should.

```go
// SUGGESTED: Either make both named types or neither
type TargetType string
const (
    TargetPost    TargetType = "post"
    TargetComment TargetType = "comment"
)
```

**3. `PublicUserID` and `PublicTargetID` on the domain entity** — These are API presentation concerns leaked into the domain. A DTO in the adapter layer would be cleaner (KISS: domain doesn't need to know about API serialization).

### 1.2 `domain/errors.go` (26 lines)

Clean. No issues.

### 1.3 `domain/reaction_test.go` (109 lines)

**Issues**:

**1. `TestReaction_StructFields` (L76-104) is a waste** — Testing that struct field assignment works is testing the Go language, not your code. Delete entirely.

**2. `TestReactionTypeConstants` (L106-113) is trivially testing constants** — Testing `ReactionLike != "like"` is testing that strings equal themselves. Delete.

**3. Test named `TestReaction_IsValid` but method is `Validate()`** — Minor naming mismatch that can confuse.

### 1.4 `ports/service.go` (35 lines)

Clean interface. No issues.

### 1.5 `ports/repository.go` (31 lines)

Clean interface. No issues.

### 1.6 `ports/service_test.go` (113 lines)

**CRITICAL: Mock implementations don't match the current interface signatures.**

The mock `mockReactionService.React()` takes `(ctx, userID, targetID int, targetType string, reactionType)` — using `int` for targetID. But the actual interface uses `string` for `targetPublicID`. Same for `RemoveReaction`, `GetReactions`, `CountReactions`.

```go
// CURRENT (L31-33): WRONG SIGNATURE — uses int for targetID
func (m *mockReactionService) React(ctx context.Context, userID, targetID int, ...) error {

// ACTUAL INTERFACE: uses string targetPublicID
React(ctx context.Context, userID int, targetPublicID string, ...) error
```

**This means the mock doesn't implement the interface and these tests should fail to compile.** The mock is never assigned to a `ReactionService` variable, so the compiler doesn't catch it. The `var reactionService ReactionService; _ = reactionService` trick only checks the variable declaration, not the mock.

Similarly `mockReactionRepository` has methods `Count`, `GetByTarget`, `Delete` that don't exist on the actual `ReactionRepository` interface (which has `CountByTargetPublicID`, `GetByTargetPublicID`, `DeleteByTargetPublicID`).

**Fix**: Use `var _ ReactionService = (*mockReactionService)(nil)` pattern (as notification does) to get compile-time interface checks. Then fix the signatures.

### 1.7 `application/service.go` (282 lines)

**Issues**:

**1. Massive repeated target-existence validation (DRY violation)** — The same `switch targetType { case "post": ... case "comment": ... }` block to verify target existence appears in `React()` (L70-80), `RemoveReaction()` (L155-165), `GetReactions()` (L197-207), `CountReactions()` (L218-228), and `GetByUserAndTargetPublicID()` (L257-267). That's **5 copies** of the same logic.

```go
// SUGGESTED: Extract once
func (s *Service) verifyTargetExists(ctx context.Context, targetPublicID, targetType string) (ownerID int, err error) {
    switch targetType {
    case "post":
        post, err := s.postRepo.GetByID(ctx, targetPublicID)
        if err != nil { return 0, err }
        return post.UserID, nil
    case "comment":
        _, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
        return 0, err
    default:
        return 0, domain.ErrInvalidTarget
    }
}
```

This would eliminate ~50 lines and make the validation consistent.

**2. Repeated input validation** — `targetType != "post" && targetType != "comment"` appears in every method. Extract to a helper:

```go
func validTargetType(t string) bool { return t == "post" || t == "comment" }
```

**3. Fire-and-forget goroutine for counter updates (L130-136, L173-179)** — Launching goroutines with `go func()` for `IncrementReactionCount`/`DecrementReactionCount` is risky:
  - No way to track errors in tests
  - The goroutine creates a new detached `context.Background()`, losing any tracing/cancellation
  - Race conditions: the user could remove and re-add a reaction before the goroutine runs
  - If the service is in a request-scoped lifecycle, the goroutine outlives the request

**Suggested**: Either do this synchronously (it's just a counter update, not expensive) or use a proper async mechanism (channel, work queue). For KISS, synchronous is better.

**4. `SetNotificationService` setter (L46-48)** — This is a code smell for circular dependency. The reaction service optionally depends on notifications. A better pattern: accept it as a constructor parameter (possibly nil) or use an event/observer pattern. The setter allows calling it at any time, which is unpredictable.

**5. Notification creation for reactions only fires for posts (L122-131)** — Comment reaction notifications are silently skipped. If this is intentional, add a comment. If not, it's a missing feature.

**6. `postOwnerID` captured only for "post" case (L70-80)** — If targetType is "comment", `postOwnerID` stays 0, which correctly skips notification. But this coupling between post ownership and notification logic in the reaction service is fragile.

### 1.8 `application/service_test.go` (506 lines)

**Issues**:

**1. Enormous mock implementations** — The `MockUserService` alone is ~60 lines of no-op methods. This is a sign the `UserService` interface is too fat. For this module, only `GetByID`, `GetByPublicID`, `IncrementReactionCount`, and `DecrementReactionCount` are needed. The handler already uses a local `ServiceContainer` interface — the mocks should do the same.

**2. Mock repositories are duplicated across test files** — `MockPostRepository`, `MockCommentRepository`, `MockUserService` re-implement repository interfaces from scratch in each module's tests. This violates DRY across the codebase.

### 1.9 `adapters/http_handler.go` (49 lines)

**Issues**:

**1. `Logger()` on `ServiceContainer` interface** — The reaction handler's `ServiceContainer` requires `Logger()` but moderation's doesn't. This inconsistency means the interface contracts differ per module. Consider making logger a standard field across all handler containers.

**2. Unused fields** — `templates *template.Template` is stored but never used (no page handlers in this module). Remove it.

### 1.10 `adapters/http_handler_api.go` (264 lines)

**Issues**:

**1. Excessive logging** — Every handler has 2-4 log calls (entry, exit, error). For production, structured request logging middleware would handle this. The per-handler logging adds ~40% of the code volume:

```go
// BEFORE: 8 lines of logging per handler
h.logger.Info("Processing reaction", ...)
// ... handler logic ...
h.logger.Info("Reaction added successfully", ...)

// AFTER: Middleware handles it, handler is just logic
```

**2. Redundant auth check in `AddReactionAPI` (L34-38)** — The route already uses `authMiddleware(http.HandlerFunc(...))`, so `authPorts.IsAuthenticated(r.Context())` will always be true. The double-check is defensive but unnecessary (same in `RemoveReactionAPI`).

**3. Manual JSON response (L104)** — `fmt.Fprintf(w, `{"message": "Reaction added successfully"}`)` is fragile (no Content-Type header set, no JSON encoding). Use `json.NewEncoder` or the existing `writeJSON` helper for consistency.

**4. Error comparison uses `==` instead of `errors.Is()` (L96-98)** — `if err == domain.ErrInvalidTarget` should be `errors.Is(err, domain.ErrInvalidTarget)` for proper error wrapping support.

### 1.11 `adapters/sqlite_repository.go` (268 lines)

**Issues**:

**1. CRITICAL: Repeated public-to-internal ID resolution** — Every single method (`Create`, `DeleteByTargetPublicID`, `GetByTargetPublicID`, `GetByUserAndTargetPublicID`, `CountByTargetPublicID`) repeats the same 10-15 line pattern to resolve `targetPublicID` → `targetID`:

```go
var targetID int
switch targetType {
case "post":
    query = "SELECT id FROM posts WHERE public_id = ?"
case "comment":
    query = "SELECT id FROM comments WHERE public_id = ?"
}
err := r.db.QueryRowContext(ctx, query, targetPublicID).Scan(&targetID)
```

This appears **5 times**. Extract to a helper:

```go
func (r *SQLiteReactionRepository) resolveTargetID(ctx context.Context, publicID, targetType string) (int, error) {
    var table string
    switch targetType {
    case "post":
        table = "posts"
    case "comment":
        table = "comments"
    default:
        return 0, domain.ErrInvalidTarget
    }
    var id int
    err := r.db.QueryRowContext(ctx, "SELECT id FROM "+table+" WHERE public_id = ?", publicID).Scan(&id)
    if err == sql.ErrNoRows {
        return 0, domain.ErrReactionNotFound
    }
    return id, err
}
```

This would eliminate ~60 lines.

**2. `CountByTargetPublicID` validates reaction type (L225-227)** — This validation is also done in the service layer. Decide on one layer for validation (prefer service). Repository should just query.

**3. Target type validation in repository** — Also duplicated from service layer. The repository shouldn't re-validate business rules.

### 1.12 `adapters/sqlite_repository_test.go` (270 lines)

**Issues**:

**1. Repeated test DB setup** — Each test function creates the same schema (posts, comments, reactions tables). Extract a `setupReactionTestDB(t *testing.T) *sql.DB` helper (notification module does this correctly).

```go
// CURRENT: ~20 lines of CREATE TABLE repeated 3 times
// SUGGESTED: One helper, called from each test
func setupReactionTestDB(t *testing.T) *sql.DB { ... }
```

---

## 2. Moderation Module

### 2.1 `domain/report.go` (40 lines)

**Issues**:

**1. `IsValid()` only checks target type** — The TODO says it should also check status and reason. The method name implies full validation but only does partial work. Meanwhile, the reaction module uses `Validate() error` which is more idiomatic (returns an error explaining *what* is invalid).

```go
// CURRENT: returns bool, incomplete
func (r *Report) IsValid() bool {
    return r.TargetType == "post" || r.TargetType == "comment"
}

// SUGGESTED: Match the pattern from reaction module
func (r *Report) Validate() error {
    if r.TargetType != "post" && r.TargetType != "comment" {
        return ErrInvalidTargetType
    }
    if r.Reason == "" {
        return ErrEmptyReason
    }
    switch r.Status {
    case StatusPending, StatusReviewed, StatusResolved:
    default:
        return ErrInvalidReportStatus
    }
    return nil
}
```

**2. Status constants are untyped strings** — Unlike `ReactionType`, these are plain `const` strings without a named type. Inconsistent with the reaction module pattern.

### 2.2 `domain/errors.go` (18 lines)

Clean. But missing errors for common cases (`ErrInvalidTargetType`, `ErrEmptyReason`).

### 2.3 `domain/report_test.go` (120 lines)

Same issues as reaction: `TestReport_StructFields` and `TestReportStatusConstants` test the Go language, not application logic.

### 2.4 `ports/service.go`, `ports/repository.go` (22, 24 lines)

Clean interfaces.

### 2.5 `ports/service_test.go` (115 lines)

**Issue**: Same as reaction — mock compatibility checks use nil assignment patterns that don't actually verify interface compliance for all mocks. The moderation mock does `if moderationService != nil` which is weaker than `var _ ModerationService = (*mockModerationService)(nil)`.

### 2.6 `application/service.go` (51 lines)

**CRITICAL: Two of three methods are unimplemented placeholders.**

```go
func (s *Service) CreateReport(...) error {
    // Implementation placeholder
    return nil
}
func (s *Service) ReviewReport(...) error {
    // Implementation placeholder
    return nil
}
```

`CreateReport` silently does nothing — it returns `nil` (success) without creating anything. This is worse than returning an error because callers think the operation succeeded. Same for `ReviewReport`.

**Suggested**: Either implement them or return a clear error:

```go
func (s *Service) CreateReport(...) error {
    return errors.New("moderation: CreateReport not implemented")
}
```

### 2.7 `application/service_test.go` (164 lines)

**Issue**: Tests pass because they test placeholder implementations that return `nil`. The tests are effectively testing `nil == nil`. They provide false confidence.

```go
func TestService_CreateReport(t *testing.T) {
    // ...
    err := service.CreateReport(ctx, 1, "pub-10", "post", "Inappropriate content")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    // This test passes because the function returns nil — not because it works!
}
```

### 2.8 `adapters/http_handler.go` (42 lines)

No issues.

### 2.9 `adapters/http_handler_api.go` (48 lines)

All three handlers return `StatusNotImplemented`. Consistent with the service being a placeholder, but the routes are registered and wired — meaning the API surface advertises endpoints that don't work.

### 2.10 `adapters/sqlite_repository.go` (64 lines)

**All four methods are unimplemented placeholders** returning `nil`. Same issue as the service — `Create` returns nil (success) without inserting anything.

### 2.11 `adapters/sqlite_repository_test.go` (208 lines)

**All tests are testing placeholder no-ops.** They create schemas, insert test data, then call methods that ignore the data and return nil. The test comments even acknowledge this:

```go
// Since the implementation is a placeholder, we expect this to return nil
// Since the implementation is a placeholder (returns nil, nil), we expect this to be nil
```

**These tests should be deleted or rewritten once the feature is implemented.** They provide zero value and noise in test output.

---

## 3. Notification Module

### 3.1 `domain/notification.go` (37 lines)

**Issues**:

**1. `MarkAsRead()` method is trivial** — It's a one-liner setter (`n.IsRead = true`). In Go, trivial setters are not idiomatic; just set the field directly. The method suggests there might be side effects or validation, but there are none.

```go
// CURRENT: Unnecessary method
func (n *Notification) MarkAsRead() {
    n.IsRead = true
}

// IDIOMATIC: Just set the field
notification.IsRead = true
```

**2. No `Validate()` method** — Unlike `Reaction`, `Notification` has no validation. The service does inline validation instead. Inconsistent with reaction's pattern.

**3. Type constants are untyped** — Same issue as moderation. `TypeLike = "like"` should be a named type for consistency with `ReactionType`.

### 3.2 `domain/errors.go` (19 lines)

Clean. But note `ErrInvalidUserID` and `ErrInvalidTarget` duplicate reaction's error names — potential confusion when both are imported.

### 3.3 `domain/notification_test.go` (83 lines)

Same issues: `TestNotification_StructFields` and `TestNotificationTypeConstants` are trivial. `TestNotification_MarkAsRead` tests a one-line setter.

### 3.4 `ports/service.go`, `ports/repository.go` (23, 22 lines)

Clean. No issues.

### 3.5 `ports/service_test.go` (90 lines)

Good — uses the `var _ NotificationService = (*mockNotificationService)(nil)` pattern correctly for compile-time interface checks. Other modules should follow this pattern.

### 3.6 `application/service.go` (62 lines)

Clean, well-implemented. Minor notes:

**1. Validation could be in domain** — The `switch notifType` validation (L32-36) and `userID <= 0` check could live in `Notification.Validate()` for consistency with the reaction domain pattern.

### 3.7 `application/service_test.go` (151 lines)

Well-structured. Good coverage of error cases. No major issues.

### 3.8 `adapters/http_handler.go` (45 lines)

No issues. Clean.

### 3.9 `adapters/http_handler_api.go` (81 lines)

**Issues**:

**1. `GetNotificationsAPI` doesn't verify auth** — Unlike reaction's handler which double-checks `authPorts.IsAuthenticated()`, notification uses `authPorts.GetUserID()` and checks for empty string. This is actually the better approach (less redundant), but the inconsistency between modules is notable.

**2. Manual JSON encoding with `map[string]interface{}`** — Should use a typed struct:

```go
// CURRENT: untyped map
json.NewEncoder(w).Encode(map[string]interface{}{
    "notifications": notifications,
    "count": len(notifications),
    "unread_count": unreadCount,
})

// SUGGESTED: typed response
type notificationsResponse struct {
    Notifications []*domain.Notification `json:"notifications"`
    Count         int                    `json:"count"`
    UnreadCount   int                    `json:"unread_count"`
}
```

### 3.10 `adapters/sqlite_repository.go` (125 lines)

**Issues**:

**1. `Create()` has an awkward zero-time branch (L43-55)** — The logic checks `createdAt.IsZero()`, then re-checks it after a no-op reassignment. This is overengineered:

```go
// CURRENT: Confusing double-check
createdAt := notification.CreatedAt
if createdAt.IsZero() {
    createdAt = sql.NullTime{}.Time  // This is STILL zero!
}
if createdAt.IsZero() {  // Always true if above was true
    // use CURRENT_TIMESTAMP
}

// SUGGESTED: Simple
if notification.CreatedAt.IsZero() {
    _, err = r.db.ExecContext(ctx, `INSERT ... CURRENT_TIMESTAMP`, ...)
} else {
    _, err = r.db.ExecContext(ctx, `INSERT ... ?`, ..., notification.CreatedAt)
}
```

Or better yet, always set `CreatedAt` in the service (which already does `time.Now()`) and eliminate the branch entirely — just one INSERT query.

**2. Target resolution only checks `posts` table (L38-44)** — Notifications could be about comments too (the domain defines `TypeComment` and `TypeReply`), but the target ID resolution only queries the `posts` table. This is a **bug** if notifications are ever created for comment-related events.

**3. `GetByUserID` uses LEFT JOIN** — Good. But the JOIN assumes all targets are posts (`LEFT JOIN posts p ON n.target_id = p.id`). Same bug as above — comment notifications would have no target public ID.

### 3.11 `adapters/sqlite_repository_test.go` (136 lines)

Well-structured with `setupNotificationTestDB` helper. Good pattern others should follow.

### 3.12 `adapters/http_handler_api_test.go` (135 lines)

**Issues**:

**1. Massive mock for `mockUserService`** — ~40 lines of no-op methods. Same fat-interface problem.

---

## 4. Cross-Module Issues

### 4.1 Duplicated "target type" validation

All three modules validate `targetType == "post" || targetType == "comment"`. This string-based dispatch appears 15+ times across the codebase. A shared utility or named type would centralize this:

```go
// platform/target.go
type TargetType string
const (
    TargetPost    TargetType = "post"
    TargetComment TargetType = "comment"
)
func (t TargetType) Valid() bool { return t == TargetPost || t == TargetComment }
```

### 4.2 Duplicated public-ID-to-internal-ID resolution

The reaction repository resolves `publicID → internalID` five times. The notification repository does it once (but only for posts). This is the core cost of the dual-ID architecture. A shared helper would reduce this:

```go
// platform/idresolver.go
func ResolveTargetID(ctx context.Context, db *sql.DB, publicID, targetType string) (int, error) { ... }
```

### 4.3 Fat mock implementations in tests

`MockUserService` is duplicated in `reaction/application/service_test.go` and `notification/adapters/http_handler_api_test.go` with identical 20+ no-op methods. Consider a shared test helper package:

```
tests/mocks/user_service.go    // One mock, used everywhere
tests/mocks/post_repository.go
```

### 4.4 Inconsistent validation patterns across domains

| Module       | Method       | Returns    | Validates             |
|-------------|-------------|------------|----------------------|
| Reaction    | `Validate()` | `error`    | targetType, type, IDs |
| Moderation  | `IsValid()`  | `bool`     | Only targetType       |
| Notification | *(none)*    | —          | —                    |

Should be unified: all use `Validate() error`.

### 4.5 Inconsistent `ServiceContainer` interfaces

| Module       | Has `Logger()`? | Has `User()`? |
|-------------|-----------------|---------------|
| Reaction    | Yes             | Yes           |
| Moderation  | No              | No            |
| Notification | No              | Yes           |

### 4.6 Inconsistent error handling in handlers

- Reaction uses `errors ==` comparison (not `errors.Is()`)
- Notification uses `errors ==` comparison
- Should use `errors.Is()` or `errors.As()` throughout for wrappable errors

### 4.7 Moderation module is 90% placeholder

The entire moderation module (service, repository, handlers) is unimplemented. The code is wired up, routes are registered, but everything returns nil/501. This adds compilation cost, test noise, and maintenance burden for zero functionality. Consider either:
- Implementing it fully
- Removing the wiring until it's ready (keep only domain/ports as design docs)

---

## Summary of Priority Fixes

### Critical (Bugs / Correctness)
1. **Reaction `ports/service_test.go`**: Mock signatures don't match interface — broken compliance test
2. **Notification `sqlite_repository.go`**: Target resolution only checks `posts` table — bug for comment notifications
3. **Moderation service**: Returns nil (success) for unimplemented operations — silent data loss

### High (Code Quality / DRY)
4. Extract `resolveTargetID()` helper in reaction repository (eliminates 5× duplication)
5. Extract `verifyTargetExists()` helper in reaction service (eliminates 5× duplication)
6. Unify domain validation pattern: all entities use `Validate() error`
7. Fire-and-forget goroutines in reaction service → make synchronous

### Medium (Idioms / Consistency)
8. Remove trivial tests (`TestXxx_StructFields`, `TestXxxConstants`)
9. Use `errors.Is()` instead of `==` for error comparisons
10. Extract shared `MockUserService` to test helper package
11. Remove unused `templates` field from reaction handler
12. Add `TargetType` named type to share across modules
13. Remove `MarkAsRead()` domain method — just set field directly

### Low (Cleanup)
14. Use typed response structs instead of `map[string]interface{}`
15. Fix `createdAt` double-zero-check in notification repository
16. Delete moderation placeholder tests that test nil==nil
17. Remove redundant auth checks in reaction handlers (middleware already guards)
18. Move `PublicUserID`/`PublicTargetID` from domain to adapter DTOs
