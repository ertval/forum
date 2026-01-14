# Go Code Simplifier Review

**Folder/Module:** post
**Date:** 2026-01-14 14:56
**Files Reviewed:**

- `internal/modules/post/application/service.go`
- `internal/modules/post/adapters/http_handler_api.go`
- `internal/modules/post/adapters/sqlite_repository.go`

---

## Summary

The `post` module implements a vertical slice for forum posts, including creation, listing (with complex filters), and management. The code generally follows the project's Modular Monolith architecture. However, there are opportunities to simplify the HTTP request parsing logic, improve basic performance by injecting the logger, and enhance robustness in background tasks.

---

## Findings

### 1. Ad-hoc Logger Instantiation

**File:** `internal/modules/post/adapters/http_handler_api.go`
**Line(s):** 118, 260, 332
**Category:** Performance / Architecture
**Severity:** Medium

**Current Code:**

```go
l := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
l.Error(...)
```

**Situation:**
The code creates a new logger instance (and likely opens/accesses stderr) inside the request handler, sometimes multiple times or only on error. This is inefficient and decentralizes configuration.

**Suggested Improvement:**
Inject the logger into the `HTTPHandler` struct during initialization (in `internal/modules/post/adapters/http_handler.go`), similar to how `services` are injected.

```go
// In HTTPHandler struct
type HTTPHandler struct {
    // ... other fields
    logger *logger.Logger
}

// In NewHTTPHandler
func NewHTTPHandler(services ServiceContainer, templates *template.Template, logger *logger.Logger) *HTTPHandler {
    return &HTTPHandler{
        // ...
        logger: logger,
    }
}

// Usage in handlers
h.logger.Error(...)
```

**Rationale:**

- **Performance**: Removes overhead of creating loggers per request.
- **Consistency**: Centralizes log configuration (level, format, output) at the application level.

---

### 2. Complex Request Parsing Logic

**File:** `internal/modules/post/adapters/http_handler_api.go`
**Line(s):** 71-129 (Create), 266-328 (Update)
**Category:** KISS Violation / Readability
**Severity:** Medium

**Current Code:**
The `CreatePostAPI` and `UpdatePostAPI` handlers contain large `switch` blocks duplicating logic to handle both `multipart/form-data` and JSON/`application/x-www-form-urlencoded`.

**Suggested Improvement:**
Extract this logic into a private helper method specific to the handler or a distinct internal utility.

```go
func (h *HTTPHandler) parsePostRequest(r *http.Request) (req createPostRequest, img []byte, err error) {
    // Shared parsing logic handling Multipart, JSON, or Form
    // returns a normalized struct and image data
}
```

**Rationale:**

- **Readability**: Reduces the handler size significantly, letting the reader focus on the flow (Auth -> Parse -> Service -> Response).
- **DRY**: Logic for parsing categories (array vs CSV) is duplicated creates risk of inconsistency.

---

### 3. Unchecked Errors in Goroutines

**File:** `internal/modules/post/application/service.go`
**Line(s):** 130-133, 196-198
**Category:** Robustness
**Severity:** Low

**Current Code:**

```go
go func() {
    _ = s.userService.IncrementPostCount(context.Background(), userID)
}()
```

**Situation:**
The service fires a goroutine to update user stats and explicitly ensures no error is checked (`_ =`). While this prevents the main request from failing due to a stat update, it swallows potential issues (e.g., database locks, connection failures) making debugging impossible.

**Suggested Improvement:**
At minimum, log the error if it occurs.

```go
go func() {
    if err := s.userService.IncrementPostCount(context.Background(), userID); err != nil {
        // Assuming logger is available or use stdlib for now if not injected in service
        // fmt.Printf("Failed to increment post count: %v\n", err)
        // Ideally: s.logger.Warn("failed to increment post count", "error", err)
    }
}()
```

**Rationale:**

- **Observability**: "Fire and forget" should not mean "fire and ignore." You need to know if stats are drifting.

---

## Action Items

- [ ] Inject `logger` into `HTTPHandler` via `NewHTTPHandler`.
- [ ] Refactor request parsing in `CreatePostAPI` and `UpdatePostAPI` into a helper function (e.g., `parsePostRequest`).
- [ ] Add error logging to the background goroutines in `Service`.

---

## Notes

- The `SQLitePostRepository.List` method constructs a very dynamic query. While functional, ensure that as filters grow, this complexity remains managed. The current string concatenation is safe but requires careful maintenance.
- `GetByID` performs multiple aggregations (likes, dislikes, comments) on the fly. Validate performance on larger datasets; sticking to the current counter-cache approach (updating user counts) is good, consider doing the same for post reaction counts if read-heavy.
