# Code Review: Reaction Module

**Date:** 2026-01-14 15:36
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `reaction` module allows users to like/dislike posts and comments. It handles the "toggle" logic (removing an existing like if clicked again, or switching from like to dislike). The implementation relies heavily on string discriminators ("post" vs "comment") which creates repetitive code paths. There is a missing foreign key integrity check in the application logic, and race conditions in stats updates.

## Critical Issues (Must Fix)

- **ISSUE-1: Race Condition in Stats Updates**

  - **Location:** `internal/modules/reaction/application/service.go` (`React`, `RemoveReaction`)
  - **Probability:** Medium
  - **Description:** Similar to other modules, `IncrementReactionCount` and `DecrementReactionCount` are called in unchecked goroutines.
  - **Proposed Fix:** Synchronous execution or reliable event queue.

- **ISSUE-2: Duplicate ID Resolution Logic**
  - **Location:** `internal/modules/reaction/adapters/sqlite_repository.go` (`Create`, `DeleteByTargetPublicID`, `GetByTargetPublicID`, etc.)
  - **Probability:** Low (Maintenance burden)
  - **Description:** Every method in the repository repeats the switch/case logic to query `posts` or `comments` table to resolve `public_id` -> `id`.
    ```go
    switch targetType {
    case "post": query = "SELECT id FROM posts..."
    case "comment": query = "SELECT id FROM comments..."
    }
    ```
    This violates DRY.
  - **Proposed Fix:** Extract this resolution logic into a private helper method `resolveTargetID(ctx, publicID, targetType) (int, error)` to reduce code duplication and the risk of bugs in one specific method.

## Performance & Optimization

- **PERF-1: String-based Polymorphism**
  - **Description:** Using `"post"` and `"comment"` strings as discriminators in the DB `reactions` table (`target_type` column) prevents efficient foreign key constraints and requires string comparisons.
  - **Optimized Model:** Use separate link tables `post_reactions` and `comment_reactions`, OR keep the current single table but optimize the index to be `(target_type, target_id, user_id)`. Ensure this composite index exists.

## Nitpicks & Best Practices

- **API Path Design:** The API paths `/api/reactions/{targetType}/{targetId}` expose implementation details (polymorphism). A more RESTful approach might be `/api/posts/{id}/reactions` and `/api/comments/{id}/reactions`. However, the current design is functional.
- **Toggle Logic Complexity:** The `React` method in the service is quite complex because it handles creating, removing (toggle), and switching reaction types all in one function. Breaking this down into `addReaction`, `removeReaction`, `switchReaction` might make it more testable.

---
