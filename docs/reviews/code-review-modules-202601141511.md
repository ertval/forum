# Code Review: Forum Modules

**Date:** 2026-01-14 15:11
**Reviewer:** Principal Software Engineer / Low-Level Systems Architect
**Scope:** `/internal/modules/` (auth, comment, post, reaction, user, moderation, notification)

---

## Executive Summary

The Forum modules demonstrate **strong adherence to hexagonal architecture** with proper separation of concerns (domain, ports, application, adapters). The codebase follows **Go idiomatic patterns** and the project's ID security policy (INT internally, UUID publicly) is consistently applied. However, there are **critical concurrency issues with fire-and-forget goroutines**, **silent error swallowing in several locations**, and **potential performance concerns with N+1 queries** in page handlers.

---

## Critical Issues (Must Fix)

### ISSUE-1: Fire-and-Forget Goroutines Without Error Handling or Synchronization

- **Location:** `comment/application/service.go`, Lines 58-60, 108-110; `reaction/application/service.go`, Lines 106-108, 147-149
- **Probability:** High
- **Description:** The code spawns goroutines to update user counts asynchronously but completely ignores errors and provides no mechanism to track completion. If these operations fail (e.g., database connection issues), the counts will drift out of sync with no visibility into the problem. In high-load scenarios, this could spawn unbounded goroutines.

```go
// Current problematic pattern (comment/application/service.go:58-60)
go func() {
    _ = s.userService.IncrementCommentCount(context.Background(), userID)
}()
```

- **Proposed Fix:** Use a proper worker pool pattern or at minimum log errors:

```go
// Option 1: Log errors (minimal fix)
go func(uid int) {
    if err := s.userService.IncrementCommentCount(context.Background(), uid); err != nil {
        log.Printf("WARNING: failed to increment comment count for user %d: %v", uid, err)
    }
}(userID)

// Option 2: Use a bounded worker pool (recommended for production)
type countWorker struct {
    jobs chan countJob
    wg   sync.WaitGroup
}
```

---

### ISSUE-2: Silent Error Swallowing in Session Cleanup

- **Location:** `auth/application/service.go`, Lines 147-151, 194, 213, 241
- **Probability:** Medium
- **Description:** Multiple places where errors are silently discarded with `_`, making debugging difficult and hiding potential database issues:

```go
// Line 147-151: Login session deletion
err = s.sessionRepo.DeleteByUserID(ctx, user.ID)
if err != nil {
    // If we can't delete existing sessions, continue anyway
    // This might result in multiple active sessions, but login should still work
}

// Line 194: Best effort cleanup
_ = s.sessionRepo.Delete(ctx, sessionToken) // Best effort cleanup
```

- **Proposed Fix:** At minimum, log these errors for debugging:

```go
if err := s.sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
    log.Printf("WARNING: failed to delete existing sessions for user %d: %v", user.ID, err)
    // Continue with login - not a blocking error
}

// For cleanup operations
if err := s.sessionRepo.Delete(ctx, sessionToken); err != nil && !errors.Is(err, domain.ErrSessionNotFound) {
    log.Printf("WARNING: failed to cleanup expired session: %v", err)
}
```

---

### ISSUE-3: Logout Error Silently Ignored in Page Handler

- **Location:** `auth/adapters/http_handler_page.go`, Lines 63-64
- **Probability:** Medium
- **Description:** The logout page handler ignores logout errors, which could leave sessions active server-side while the user appears logged out client-side (cookie cleared):

```go
// Line 63-64
_ = h.authService.Logout(r.Context(), cookie.Value) // We ignore the error for frontend UX
```

- **Proposed Fix:** Log the error while maintaining UX:

```go
if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
    // Log but don't block the logout UX
    log.Printf("WARNING: failed to invalidate session on logout: %v", err)
}
```

---

### ISSUE-4: Missing Input Validation on DeleteCommentForm Method Check

- **Location:** `comment/adapters/http_handler_form.go`, Lines 22-27
- **Probability:** Low
- **Description:** The `CreateCommentForm` handler is registered for `POST /posts/{post_id}/comments` but still manually checks for POST method at runtime. While this is defensive, the real issue is it returns `MethodNotAllowed` after the route already filtered for POST. This is dead code but suggests potential confusion about Go 1.22 route patterns.

```go
// CreateCommentForm is registered as: router.HandleFunc("POST /posts/{post_id}/comments", ...)
// But still checks:
if r.Method != http.MethodPost {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
}
```

- **Proposed Fix:** Remove redundant method checks since Go 1.22+ patterns handle this:

```go
// The "POST /posts/{post_id}/comments" pattern already enforces POST method
// Remove the manual check - it's dead code
func (h *HTTPHandler) CreateCommentForm(w http.ResponseWriter, r *http.Request) {
    // Get userID from session (no method check needed)
    userID, _ := h.GetCurrentUser(r)
    // ...
}
```

---

## Performance & Optimization

### PERF-1: N+1 Query Pattern in MyCommentsPage

- **Location:** `comment/adapters/http_handler_page.go`, Lines 89-115
- **Description:** For each comment, the handler makes separate database calls to fetch user info and post info. With 20 comments, this results in up to 41 queries (1 for comments + 20 for users + 20 for posts).

```go
for _, comment := range commentsFromService {
    // Query 1: Get author
    user, err := h.userService.GetByID(ctx, comment.UserID)

    // Query 2: Get post
    post, err := h.postService.GetPost(ctx, comment.PublicPostID)
}
```

- **Optimized Code:** Use batch fetching or JOIN queries:

```go
// Option 1: Add a method to fetch comments with eager-loaded relations
comments, err := h.commentService.ListCommentsByUserWithDetails(ctx, userPublicID, initialLimit+1, 0)

// Option 2: Batch fetch users and posts after getting comments
userIDs := collectUniqueUserIDs(comments)
users, _ := h.userService.GetByIDs(ctx, userIDs) // Add batch method
usersMap := mapByID(users)
```

---

### PERF-2: N+1 Query Pattern in LoadMoreCommentsAPI

- **Location:** `comment/adapters/http_handler_page.go`, Lines 254-298
- **Description:** Same N+1 pattern as PERF-1, compounded by the fact this is an API endpoint that may be called repeatedly for infinite scroll.

---

### PERF-3: Redundant Template Parsing on Every Request

- **Location:** `auth/adapters/http_handler_page.go`, Lines 24-29, 43-48
- **Probability:** Medium
- **Description:** Templates are parsed on every request instead of being cached:

```go
// LoginPage - parses templates on each request
tmpl, err := template.ParseFiles("templates/base.html", "templates/login.html")
if err != nil { ... }
```

- **Optimized Code:** Use the pre-parsed templates from the handler:

```go
// Use the cached templates instead
func (h *HTTPHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "Title": "Login",
    }

    if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
        http.Error(w, "Failed to render login page", http.StatusInternalServerError)
        return
    }
}
```

---

### PERF-4: Reaction Service Validates Target Existence Multiple Times

- **Location:** `reaction/application/service.go`, Lines 54-66, 167-177
- **Description:** When adding a reaction, the code validates the target exists, then the repository does it again. This doubles the database queries.

```go
// Service.React() calls:
switch targetType {
case "post":
    _, err := s.postRepo.GetByID(ctx, targetPublicID) // Query 1
    ...
}

// Then repository.Create() calls:
switch reaction.TargetType {
case "post":
    err = r.db.QueryRowContext(ctx, "SELECT id FROM posts WHERE public_id = ?", ...) // Query 2
```

- **Optimized Code:** The service layer validation is sufficient; repository should trust the input or use INSERT with subquery:

```go
// Option: Use INSERT with EXISTS check (single query)
query := `
    INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at)
    SELECT ?, ?, id, ?, ?, CURRENT_TIMESTAMP
    FROM posts WHERE public_id = ?
`
```

---

## Error Handling & Robustness

### ERR-1: Missing rows.Err() Check After Iteration

- **Location:** `user/adapters/sqlite_repository.go`, Line 295 (List method)
- **Description:** The `List` method iterates over rows but never checks `rows.Err()` after the loop:

```go
for rows.Next() {
    // ... scan ...
}
// Missing: if err := rows.Err(); err != nil { return nil, err }
return users, nil
```

- **Proposed Fix:**

```go
for rows.Next() {
    // ... existing scan logic ...
}

if err := rows.Err(); err != nil {
    return nil, err
}

return users, nil
```

---

### ERR-2: Same Issue in Multiple Repository Methods

- **Location:** Multiple locations:
  - `auth/adapters/sqlite_session_repository.go` (GetByUserID)
  - `comment/adapters/sqlite_repository.go` (ListByPostPublicID, ListByUser, ListByUserPaginated)
  - `reaction/adapters/sqlite_repository.go` (GetByTargetPublicID)

---

### ERR-3: Potential Nil Pointer in Reaction Handler

- **Location:** `reaction/adapters/http_handler_api.go`, Lines 184-231
- **Description:** `GetReactionsAPI` extracts path parts without validating they exist:

```go
pathParts := strings.Split(r.URL.Path, "/")
if len(pathParts) < 4 {
    // This check is correct, but...
}

targetType := pathParts[len(pathParts)-2]  // Could be index -2 if path ends with "/"
targetID := pathParts[len(pathParts)-1]    // Could be empty string
```

- **Proposed Fix:** Use Go 1.22+ `PathValue` instead:

```go
targetType := r.PathValue("targetType")
targetID := r.PathValue("targetId")

if targetType == "" || targetID == "" {
    http.Error(w, "Invalid path parameters", http.StatusBadRequest)
    return
}
```

---

## Nitpicks & Best Practices

### NIT-1: Inconsistent JSON Error Response Format

- **Location:** Various handlers
- **Description:** Some handlers use `platformErrors.WriteErrorJSON()` while others use `http.Error()`. This creates inconsistent API responses.
- **Recommendation:** Standardize on `platformErrors.WriteErrorJSON()` for all API handlers.

---

### NIT-2: Deprecated Functions Still Present

- **Location:** `auth/adapters/middleware.go`, Lines 64-70, 112-118, 120-139
- **Description:** Several functions are marked as deprecated but still exist:
  - `RequireAuth` (standalone function)
  - `OptionalAuth` (standalone function)
  - `GetUserID`, `GetUsername`, `IsAuthenticated` (in adapters package)
- **Recommendation:** Remove deprecated functions if no longer used, or add a deprecation timeline.

---

### NIT-3: Magic Numbers and Hardcoded Values

- **Location:** Various
- **Description:**
  - Comment max length: 5000 (comment/domain/comment.go:36)
  - Post title max: 255 (post/domain/post.go:31)
  - Post content max: 50000 (post/domain/post.go:35)
  - Session token in cookie named "session_token"
- **Recommendation:** Extract to constants with clear documentation.

---

### NIT-4: fmt.Printf Used for Logging

- **Location:** `comment/adapters/http_handler_page.go`, Lines 59, 81, 194
- **Description:** Uses `fmt.Printf` instead of the structured logger:

```go
fmt.Printf("Error fetching categories: %v\n", err)
fmt.Printf("Error fetching user comments: %v\n", err)
fmt.Printf("Template error: %v\n", err)
```

- **Recommendation:** Use the platform logger for consistency:

```go
h.logger.Error("Failed to fetch categories", logger.Error(err))
```

---

### NIT-5: writeJSON Error Ignored

- **Location:** `reaction/adapters/http_handler_api.go`, Line 300
- **Description:** JSON encoding errors are silently ignored:

```go
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)  // Error ignored
}
```

- **Recommendation:** At least log encoding errors (can't write to client after WriteHeader):

```go
if err := json.NewEncoder(w).Encode(data); err != nil {
    h.logger.Error("Failed to encode JSON response", logger.Error(err))
}
```

---

### NIT-6: Duplicated buildCurrentUser Implementation

- **Location:** `comment/adapters/http_handler.go` (Lines 59-92) and `post/adapters/http_handler.go` (Lines 66-99)
- **Description:** Nearly identical implementations of `buildCurrentUser` exist in both modules.
- **Recommendation:** Extract to a shared utility in the platform layer or create a shared template data builder.

---

### NIT-7: Context Background in Goroutines May Cause Timeout Issues

- **Location:** `comment/application/service.go`, `reaction/application/service.go`
- **Description:** Using `context.Background()` in fire-and-forget goroutines bypasses any parent context deadlines, which is intentional but should be documented:

```go
go func() {
    _ = s.userService.IncrementCommentCount(context.Background(), userID)
}()
```

- **Recommendation:** Add a comment explaining the intentional use of fresh context:

```go
// Use Background context intentionally - this operation should complete
// even if the parent request is cancelled
go func(uid int) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := s.userService.IncrementCommentCount(ctx, uid); err != nil {
        log.Printf("WARNING: failed to increment comment count: %v", err)
    }
}(userID)
```

---

## Security Observations

### SEC-1: Cookie Security Flags (Not Critical - Development Mode)

- **Location:** `auth/adapters/http_handler_api.go`, Lines 73, 127, 176
- **Description:** Cookies have `Secure: false` with a comment about production. Ensure this is configuration-driven.

---

### SEC-2: ID Security Policy Well-Implemented ✓

- **Observation:** The codebase correctly implements the INT/UUID separation:
  - Internal IDs (`ID int`) are never exposed in JSON (`json:"-"`)
  - Public UUIDs (`PublicID string`) are used in URLs and responses
  - Context stores PublicID (UUID), middleware correctly translates
  - Good security comments throughout (e.g., `// SECURITY: Stores PublicID (UUID) in context, never internal INT ID`)

---

## Summary of Action Items

| Priority    | Issue                                    | Fix Complexity |
| ----------- | ---------------------------------------- | -------------- |
| 🔴 Critical | ISSUE-1: Fire-and-forget goroutines      | Medium         |
| 🔴 Critical | ISSUE-2: Silent error swallowing in auth | Low            |
| 🟡 Medium   | PERF-1/2: N+1 queries in page handlers   | High           |
| 🟡 Medium   | PERF-3: Template re-parsing              | Low            |
| 🟡 Medium   | ERR-1/2: Missing rows.Err() checks       | Low            |
| 🟢 Low      | NIT-1 through NIT-7                      | Low-Medium     |

---

**Reviewed by:** Principal Software Engineer
**Date:** 2026-01-14
