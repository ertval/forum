# Code Review: Reaction Module

## Executive Summary

The reaction module adheres to the project's strict hexagonal architecture and naming conventions. However, it contains a **critical functional defect** where reaction lists do not expose the identity of the user who reacted, rendering the API incomplete. Additionally, the use of **unmanaged goroutines** for updating user statistics introduces potential for race conditions and data inconsistencies.

## Critical Issues (Must Fix)

- **ISSUE-1: Incomplete Data in Reaction List (Functional)**

  - **Location:** `adapters/sqlite_repository.go` (`GetByTargetPublicID`), Line 139
  - **Probability:** High (100% reproducible)
  - **Description:** The `GetByTargetPublicID` query fetches from the `reactions` table but fails to JOIN the `users` table to populate the `PublicUserID` field. Since the internal `UserID` is hidden (`json:"-"`) and `PublicUserID` is left empty, the API key `user_id` will be missing or empty in the response. Clients will see _that_ a reaction occurred, but not _who_ performed it.
  - **Proposed Fix:** JOIN with the `users` table in the SQL query.
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

- **ISSUE-2: Unmanaged Goroutines & Silent Failures (Concurrency/Robustness)**

  - **Location:** `application/service.go`, Lines 110 & 148
  - **Probability:** Medium (High risk at scale)
  - **Description:** The service uses `go func() { _ = s.userService.IncrementReactionCount(...) }()` to update user stats.
    1. **Data Consistency:** Errors are explicitly swallowed (`_ =`). If the DB update fails, the user's total reaction count remains incorrect indefinitely.
    2. **Resource Management:** These goroutines are not tracked (no `WaitGroup`), creating potential for leaks or abrupt termination on shutdown.
  - **Proposed Fix:** For this architecture, **KISS** (Keep It Simple, Stupid) applies. Run the update synchronously within the request. The performance cost is a single indexed row update (~1ms), which is negligible compared to the complexity of distributed consistency.

    ```go
    // Replace:
    // go func() { _ = s.userService.IncrementReactionCount(...) }()

    // With:
    if err := s.userService.IncrementReactionCount(ctx, userID); err != nil {
        // Log error but maybe don't fail the request since reaction was created?
        // Or return error. Given "Consistency" preference:
        return fmt.Errorf("failed to update user stats: %w", err)
    }
    ```

- **ISSUE-3: Missing Duplicate Handling (Robustness)**
  - **Location:** `adapters/sqlite_repository.go`, Line 61
  - **Probability:** Low (Race condition specific)
  - **Description:** While the service checks for existing reactions, race conditions can still allow concurrent requests to bypass the check. The repository's `Create` method returns the raw SQLite error instead of checking for a unique constraint violation and mapping it to `domain.ErrDuplicateReaction`.

## Performance & Optimization

- **PERF-1: Redundant ID Lookups**
  - **Description:** The `Create` workflow involves two separate DB round-trips to resolve the target ID:
    1. Service calls `postRepo.GetByID` (resolves Target)
    2. Repository calls `SELECT id FROM posts` (resolves Target again)
  - **Optimized Code:** While the current approach maintains strict decoupling, passing the resolved internal `TargetID` (if the architecture allowed internal IDs in domain structs) would save a query. Given the constraints, this is acceptable but worth noting.

## Nitpicks & Best Practices

- **Obsolete Path Parsing**:

  - `adapters/http_handler_api.go` uses manual `strings.Split(r.URL.Path, "/")` to extract parameters.
  - **Fix**: Since the project uses Go 1.24+ and registered patterns like `/api/reactions/{targetType}/{targetId}`, use the standard library's `r.PathValue("targetType")` and `r.PathValue("targetId")`.

- **Hardcoded SQL Strings**:
  - SQL queries in `sqlite_repository.go` are hardcoded strings. Consider moving them to constants or a separate SQL file if they grow complex.
