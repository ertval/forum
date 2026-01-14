# Code Review: Comment Module

## Executive Summary

The comment module is functional and follows the project's hexagonal architecture well. However, it suffers from significant performance issues due to N+1 query patterns in the HTTP handlers and lacks robust error handling in database iterators. There are also unmanaged goroutines in the service layer that pose a risk during graceful shutdown.

## Critical Issues (Must Fix)

- **ISSUE-1: Unmanaged Goroutines & Ignored Errors**

  - **Location:** `internal/modules/comment/application/service.go`, Lines 59-61, 111-113
  - **Probability:** Medium
  - **Description:** `CreateComment` and `DeleteComment` spawn detached goroutines (`go func() { ... }`) to update user statistics. These goroutines inherit `context.Background()` but are not tracked by any WaitGroup. This means they could be abruptly terminated during application shutdown, leading to inconsistent data. Furthermore, errors returned by `s.userService.IncrementCommentCount` are explicitly ignored (`_ = ...`), so failures to update stats will go unnoticed.
  - **Proposed Fix:** Use a worker pool or a proper background job system. For a simple fix, log the error instead of ignoring it.

- **ISSUE-2: Missing `rows.Err()` Check**
  - **Location:** `internal/modules/comment/adapters/sqlite_repository.go`, Lines 120, 160, 201
  - **Probability:** Medium
  - **Description:** When iterating over database rows using `for rows.Next() { ... }`, it is mandatory to check `rows.Err()` afterwards to catch any errors that occurred during iteration (e.g., network issues, broken pipe). The current implementation returns `nil` error even if the iteration was incomplete due to a failure.
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

## Performance & Optimization

- **PERF-1: N+1 Query Problem in Page Handlers**

  - **Location:** `internal/modules/comment/adapters/http_handler_page.go`, Lines 89-167
  - **Description:** The `MyCommentsPage` handler loops through each comment and performs multiple synchronous blocking calls to other services (`userService.GetByID`, `postService.GetPost`, `reactionService.CountReactions`). If a user has 20 comments, this results in 20 \* 3 = 60 additional database queries per page load. This will severely degrade performance under load.
  - **Optimized Code:** Implement "Batch Get" methods in the respective services (e.g., `GetPostsByIDs([]string)`) to fetch all related data in a single query per entity type, or use a JOIN at the repository level if boundaries allow (though typically strict modular monoliths discourage cross-module joins).

- **PERF-2: Recreating Logger on Every Error**
  - **Location:** `internal/modules/comment/adapters/http_handler.go`, Lines 129-132
  - **Description:** `writeJSON` creates a new `logger.Config` and `logger.Logger` instance every time a JSON encoding error occurs. While likely rare, this is unnecessary allocation.
  - **Optimized Code:** Initialize the logger once in the `HTTPHandler` struct (it is already absent from the struct definition, should be added) and reuse it.

## Nitpicks & Best Practices

- **Code Duplication**: `LoadMoreCommentsAPI` and `MyCommentsPage` share significant logic for enriching comment data (fetching author, post details, reactions). This logic should be extracted into a helper method like `enrichComments(ctx, comments)`.
- **Complex Handler**: `MyCommentsPage` is doing too much: auth validation, param parsing, fetching data, filtering data, and rendering. It violates SRP.
- **Silent Failures**: In `MyCommentsPage`, errors from `h.categoryService.List` and `h.commentService.ListCommentsByUserPaginated` are logged to stdout (`fmt.Printf`) but execution continues, potentially rendering a broken page.
