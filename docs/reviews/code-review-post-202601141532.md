# Code Review: Post Module

**Date:** 2026-01-14 15:32
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `post` module manages the core forum content, including post creation, retrieval, updates, deletion, and listing. It handles complex features like image uploads, category associations, and advanced filtering. The code is modular and generally well-organized, but there are several critical issues related to concurrency, resource management, and potential SQL injection vulnerabilities in dynamic query construction.

## Critical Issues (Must Fix)

- **ISSUE-1: Potential Resource Leak in Repository List Method**

  - **Location:** `internal/modules/post/adapters/sqlite_repository.go`, `List` method (Lines 284-446)
  - **Probability:** Medium
  - **Description:** The `List` method loops through query results and calls `getPostCategories` (Line 432) for _each post_. `getPostCategories` executes a new SQL query (Lines 458-463). This is the "N+1 Select Problem". For a page of 50 posts, this results in 1 + 50 = 51 database queries. While SQLite is fast, this scales poorly and can be a significant bottleneck.
  - **Proposed Fix:** Fetch all categories for the retrieved posts in a single query (using `IN` clause with post IDs) and map them in memory, or simplify the main query to `GROUP_CONCAT` categories (though this requires parsing). A separate batch query is preferred.

- **ISSUE-2: SQL Injection Risk in Dynamic Query Building**

  - **Location:** `internal/modules/post/adapters/sqlite_repository.go` (Line 329)
  - **Probability:** Low (due to internal usage) but risky pattern
  - **Description:** `repeatPlaceholders` is used to build the `IN` clause for category filtering. While `?` placeholders are used, manually constructing SQL strings is fragile. More importantly, the `GetByID` and `List` queries are very large and complex, increasing the risk of subtle bugs or injection if modified carelessly.
  - **Proposed Fix:** Use a query builder or ensure rigorous testing of the dynamic SQL generation. Ensure `repeatPlaceholders` is never used with user input directly.

- **ISSUE-3: Race Condition in CreatePost/DeletePost Async Operations**
  - **Location:** `internal/modules/post/application/service.go` (`CreatePost`, `DeletePost`)
  - **Probability:** Medium
  - **Description:** The user's post count is incremented/decremented asynchronously in a goroutine (Line 131, Line 196).
    ```go
    go func() {
        _ = s.userService.IncrementPostCount(context.Background(), userID)
    }()
    ```
    If the application crashes or restarts immediately after the HTTP handler returns but before the goroutine executes, the post count will be out of sync with the actual number of posts. This disregards data consistency.
  - **Proposed Fix:** Perform the count update synchronously within the request context, or use a reliable event bus/queue if async is strictly required (which is likely not needed for a simple counter). The durability of data is more important than saving a few milliseconds here.

## Performance & Optimization

- **PERF-1: N+1 Select for Categories**

  - **Description:** As mentioned in ISSUE-1, fetching categories for each post in a loop is inefficient.
  - **Optimized Code:** Fetch posts first. Collect all `post.ID`s. Fetch all `post_categories` where `post_id IN (...)`. Map categories to posts in Go.

- **PERF-2: JSON Decoding of Request Body**
  - **Description:** In `CreatePostAPI` and `UpdatePostAPI`, `json.NewDecoder(r.Body).Decode(&req)` is used. If the body is large, this buffers everything. While `LimitReader` is used for files, the JSON decoder might still be vulnerable to large payloads.
  - **Optimization:** Use `http.MaxBytesReader` to enforce a hard limit on the request body size for JSON requests as well.

## Nitpicks & Best Practices

- **Deeply Nested Code:** `CreatePostAPI` has a very large `switch` statement for content types, with nested logic for multipart parsing. This makes the handler hard to read and test. Extracting the request parsing logic into a separate private method would improve readability.
- **Rollback Logic:** The "rollback" logic in `CreatePost` (delete image if DB fails) is manual. If the server crashes during this operation, orphaned files may remain. A periodic cleanup job for orphaned images is recommended.
- **Error Handling:** Ignoring errors in the async goroutines (`_ = s.userService...`) means failures to update stats are silently swallowed. These should strictly be logged.

---
