# Code Review - Reaction Module - 202601141540

## Executive Summary

The `reaction` module implementation is functional and follows the Modular Monolith/Hexagonal architecture. However, there are significant bugs in the HTTP API path parsing for counting reactions, and potential data consistency issues due to non-atomic operations and ignored errors in asynchronous background tasks.

## Critical Issues (Must Fix)

- **ISSUE-1: Broken Path Parsing in CountReactionsAPI**

  - **Location:** `internal/modules/reaction/adapters/http_handler_api.go`, Lines 243-244
  - **Probability:** High
  - **Description:** The `CountReactionsAPI` handler parses the target type and ID incorrectly from the URL path. For the route `GET /api/reactions/{targetType}/{targetId}/count`, `strings.Split` with `/` results in `count` being at the last index.
  - **Proposed Fix:**
    ```go
    // Correct parsing:
    targetType := pathParts[len(pathParts)-3]
    targetID := pathParts[len(pathParts)-2]
    ```

- **ISSUE-2: Data Inconsistency Risk in Async Tasks**

  - **Location:** `internal/modules/reaction/application/service.go`, Lines 110-112 and 148-150
  - **Probability:** Medium
  - **Description:** Reaction counts are updated asynchronously using a new `context.Background()` and ignoring errors. If the `userService` fails to update the count (e.g., database lock, server shutdown), the user's reaction count will become permanently out of sync with the actual reaction records.
  - **Proposed Fix:** Update the count synchronously within a transaction (if possible across modules) or at least log the error and use a properly derived context.
    ```go
    // Better: use a logger to at least know if it failed
    go func(ctx context.Context, uid int) {
        if err := s.userService.IncrementReactionCount(ctx, uid); err != nil {
            // Log error
        }
    }(ctx, userID) // Pass derived context or at least a timeout-shortened one
    ```

- **ISSUE-3: Non-Atomic "Toggle" Logic**
  - **Location:** `internal/modules/reaction/application/service.go`, Lines 80-91
  - **Probability:** Medium
  - **Description:** The "React" logic (toggle or change) involves multiple repository calls (Get, then Delete, then Create). If the server crashes or the database fails between Delete and Create, the user's previous reaction is lost without the new one being created.
  - **Proposed Fix:** Use a database transaction to ensure atomicity of the reaction update.

## Performance & Optimization

- **PERF-1: Redundant Target Existence Checks**
  - **Description:** The service layer checks if a post/comment exists (Lines 55-66), and then the repository checks again to resolve the internal ID (Lines 34-42 in `sqlite_repository.go`). This results in 2-3 database queries where 1 would suffice.
  - **Optimized Code:** The repository's `Create` and `Delete` methods could return a specific "TargetNotFound" error if the resolution fails, or the service could pass the internal ID if it's already known (though passing internal IDs between layers is generally discouraged in this project). A better approach is to merge the existence check into the repository's main operation.

## Nitpicks & Best Practices

- **NIT-1: Misleading Success Message on Toggle**
  - **Location:** `internal/modules/reaction/adapters/http_handler_api.go`, Line 109
  - **Description:** When a reaction is "toggled off", the API still returns `{"message": "Reaction added successfully"}`. It should probably return a message reflecting that it was removed or use a generic "Reaction updated".
- **NIT-2: Hardcoded String Constants**
  - **Location:** `internal/modules/reaction/application/service.go`
  - **Description:** Use the domain constants for `post` and `comment` strings if they exist, or define them in the domain to avoid typos.
- **NIT-3: Error Wrapping**
  - **Description:** Errors from the repository and secondary services (userService, postRepo) are returned directly without wrapping. Using `fmt.Errorf("reaction service: %w", err)` would improve debugging.

---
