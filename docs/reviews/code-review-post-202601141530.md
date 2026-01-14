## Executive Summary

The `post` module implements a robust post management system with support for images, categories, and complex filtering. While the architectural separation is clean, the implementation suffers from significant performance bottlenecks in the database layer (N+1 queries) and inefficient resource management in the HTTP layer (repeated template parsing and redundant authentication checks).

## Critical Issues (Must Fix)

- **ISSUE-1: N+1 Query Pattern in Repository**

  - **Location:** `internal/modules/post/adapters/sqlite_repository.go`, `List` method (Line 432)
  - **Probability:** High
  - **Description:** The `List` method executes a separate SQL query to fetch categories for _every_ post in the result set. For a list of 50 posts, this results in 51 database roundtrips. This will cause significant latency as the traffic or result set size increases.
  - **Proposed Fix:** Use a single query with `GROUP_CONCAT` or perform a separate join to fetch all categories for all retrieved posts in one go.
    ```go
    // Better approach: Join and group categories in the main query
    query := `SELECT p.*, GROUP_CONCAT(c.name) as category_names ...`
    ```

- **ISSUE-2: Inefficient Template Parsing**

  - **Location:** `internal/modules/post/adapters/http_handler_page.go`, multiple handlers (Lines 130, 249, 346, 401, 476)
  - **Probability:** High
  - **Description:** HTML templates are parsed from disk on _every_ request. File I/O and template parsing are expensive operations that should happen once during application startup.
  - **Proposed Fix:** Parse templates once in the `NewHTTPHandler` or a configuration step and store the `*template.Template` in the handler struct.

- **ISSUE-3: Redundant Authentication Logic**
  - **Location:** `internal/modules/post/adapters/http_handler_page.go` (Lines 39-46, 154-161, 292-299)
  - **Probability:** Medium
  - **Description:** Handlers manually parse the `session_token` cookie and call `authService.ValidateSession` even when they are not protected by middleware (e.g., `HomePage`). This bypasses the centralized middleware logic and increases the risk of inconsistent auth state or security holes.
  - **Proposed Fix:** Use the `OptionalAuth` middleware for public pages to populate the user context, then retrieve the user from the context in the handler.

## Performance & Optimization

- **PERF-1: Suboptimal Counting in Queries**

  - **Location:** `internal/modules/post/adapters/sqlite_repository.go`, `GetByID` and `List` (Lines 103-119, 295-311)
  - **Description:** Counts for likes, dislikes, and comments are calculated using subqueries in `LEFT JOIN` for every request. As the `reactions` and `comments` tables grow, these joins will become increasingly expensive.
  - **Optimized Code:** Consider caching counts in the `posts` table (e.g., `like_count`, `comment_count` columns) and updating them via triggers or application logic (partially implemented in user counts but not post counts).

- **PERF-2: Excessive Field Mapping and Allocations**
  - **Location:** `internal/modules/post/adapters/http_handler_api.go` (Lines 545-566) and `http_handler_page.go` (Lines 81-102)
  - **Description:** Posts are manually converted to `map[string]interface{}` and then to JSON/Template data. This creates many small heap allocations.
  - **Optimized Code:** Define a `PostPreview` domain struct or a DTO struct that can be serialized/used directly.

## Nitpicks & Best Practices

- **Concurrency Robustness**: In `application/service.go`, the async increment/decrement of post counts using `context.Background()` may fail silently or leak if the application shuts down before completion. Use a specialized background context or a wait group for clean shutdowns.
- **Error Silencing**: In `DeletePost`, `s.postRepo.GetImagePath` error is ignored (Line 183). While "best effort", logging the error would help identify storage issues.
- **Duplicated Filtering Logic**: The logic to build filters from query parameters is duplicated across multiple handlers. This should be fully delegated to `filterService.BuildFilter`.
- **Inconsistent JSON naming**: `domain/post.go` uses `AuthorUsername` but also has a compatible `Author` field. Standardizing on one would simplify the codebase.
- **Case-Insensitivity Inconsistency**: Category lookups use `LOWER(name) = LOWER(?)`, but `PostFilter` implementation for `IN` clause (Line 329) does not appear to be case-insensitive.
