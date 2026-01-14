# Code Review: Forum Application

**Date:** 2026-01-14  
**Scope:** `internal/platform`, `internal/modules/post`, `cmd/forum`

## Executive Summary

The codebase demonstrates a strong adherence to Hexagonal Architecture (Ports & Adapters) and Go best practices, with clear separation of concerns and consistent error handling. However, there are significant reliability risks involving **goroutine leaks** in middleware and **data consistency** issues in async operations. Furthermore, the primary post listing query contains a **severe performance bottleneck** due to inefficient full-table aggregations that will degrade O(N) with database size.

## Critical Issues (Must Fix)

### ISSUE-1: Goroutine Leak in RateLimit Middleware

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 172-186
- **Probability:** High
- **Description:**
  The `RateLimit` factory function starts a background goroutine (`go func() { ... limiter.cleanup() }`) every time it is called. This goroutine runs a ticker forever. Since `RateLimit` might be called multiple times (e.g., during tests, re-configuration, or if applied dynamically), these goroutines are never signaled to stop, leading to a permanent resource leak.
- **Proposed Fix:**
  Add a `Stop()` method to the middleware or use a centralized cleaner. For a simple fix, attach the context to the cleaner or use a `Quit` channel.

```go
type RateLimiter struct {
    // ... fields
    stopChan chan struct{}
}

func (rl *RateLimiter) Stop() {
    close(rl.stopChan)
}

// In factory:
go func() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            limiter.cleanup()
        case <-limiter.stopChan: // Clean exit
            return
        }
    }
}()
```

### ISSUE-2: Unbounded Memory Growth Potential (DoS Vector)

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 163, 174
- **Probability:** Medium
- **Description:**
  The `rateLimiter` uses a map `requests map[string][]time.Time` keyed by IP address. While individual IP entries are cleaned up, the map itself can grow unbounded if an attacker spoofs a large number of distinct IP addresses (IP spoofing/Distributed DOS). This can lead to memory exhaustion.
- **Proposed Fix:**
  Implement a hard limit on the map size (e.g., LRU cache) or use a fixed-size ring buffer/Token Bucket algorithm (e.g., `golang.org/x/time/rate`).

### ISSUE-3: Data Consistency Risk in Post Creation/Deletion

- **Location:** `internal/modules/post/application/service.go`, Lines 131-133 & 196-198
- **Probability:** Medium
- **Description:**
  The service updates user post counts asynchronously:
  ```go
  go func() { _ = s.userService.IncrementPostCount(context.Background(), userID) }()
  ```
  This implementation:
  1. **Swallows errors:** If the DB update fails, the count is permanently out of sync.
  2. **Race condition:** If the application crashes immediately after the HTTP response, the goroutine may never execute.
  3. **Untraceable:** Uses `context.Background()`, losing request tracing/logging context.
- **Proposed Fix:**
  Execute the update synchronously within the transaction (ideal) or use a durable job queue. Given the current monolith structure, synchronous execution is preferred for consistency.

## Performance & Optimization

### PERF-1: SQL Performance Killer in Post Listing

- **Location:** `internal/modules/post/adapters/sqlite_repository.go`, Lines 285-312
- **Description:**
  The `List` query performs three separate `LEFT JOIN`s on derived tables that calculate counts for **ALL** posts in the system:
  ```sql
  LEFT JOIN (SELECT target_id, COUNT(*) ... GROUP BY target_id) ...
  ```
  This forces SQLite to materialize the count of likes/dislikes/comments for the _entire database_ on every single list request, resulting in O(N) performance where N is the total number of reactions/comments.
- **Optimized Code:**
  Use **correlated subqueries**, which SQLite optimizes effectively to run only for the filtered posts:

```sql
SELECT
    p.id, ...,
    (SELECT COUNT(*) FROM reactions WHERE target_id = p.id AND type='like') as like_count,
    (SELECT COUNT(*) FROM reactions WHERE target_id = p.id AND type='dislike') as dislike_count,
    (SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
FROM posts p
LEFT JOIN ...
WHERE ...
```

_Alternatively: Store `like_count` and `comment_count` as columns in the `posts` table (Denormalization) and update them incrementally._

## Nitpicks & Best Practices

- **Dangerous Default Config:** `internal/platform/database/connection.go` forces `PRAGMA journal_mode = MEMORY`. While fast, this guarantees execution relies on volatile RAM, significantly increasing the risk of database corruption in a crash. Recommended: `WAL` mode.
- **Code Duplication:** Argument parsing logic for Categories (CSV splitting, array handling) is duplicated almost verbatim between `CreatePostAPI` and `UpdatePostAPI`. Suggest extracting to a helper function `parseCategories(r *http.Request) []string`.
- **Race Condition in Startup:** `internal/platform/httpserver/server.go` waits an arbitrary `100ms` to check for startup errors. This is flaky. Using a channel to signal the listener is ready is more robust.
