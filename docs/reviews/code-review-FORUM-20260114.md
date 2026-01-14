## Executive Summary

The codebase demonstrates a solid architectural foundation using Hexagonal Architecture and robust security practices (UUIDs for public IDs). However, it contains several critical concurrency issues, specifically TOCTOU race conditions in reaction logic and non-deterministic background updates that risk long-term data inconsistency.

## Critical Issues (Must Fix)

- **ISSUE-1: TOCTOU Race Condition in Reaction Logic**

  - **Location:** `internal/modules/reaction/application/service.go`, Line 68-104
  - **Probability:** High
  - **Description:** The `React` method checks for an existing reaction and then performs a delete/insert. Under concurrent requests (e.g., a user double-clicking a like button), two goroutines can both see "no reaction" and attempt to insert. While the database unique constraint prevents duplicate rows, the service will return an unhandled database error to the user or result in inconsistent counter updates.
  - **Proposed Fix:** Use a database transaction with `Upsert` logic or explicit row locking. Move the resolution of `target_id` and the reaction toggle into a single atomic database operation.

- **ISSUE-2: Fragile Background Counter Updates**

  - **Location:** `internal/modules/reaction/application/service.go`, Line 110 & 148
  - **Probability:** High
  - **Description:** Reaction counts are updated via detached goroutines using `context.Background()`. These goroutines swallow errors and do not honor the request context. If an update fails (e.g., DB busy), the user's reaction counter will be permanently out of sync. Furthermore, there is no synchronization between the reaction insertion and the counter increment, leading to potential race conditions on the counter itself.
  - **Proposed Fix:** Perform counter updates within the same transaction as the reaction change. If asynchronous processing is required, use a reliable task queue or a persistent "outbox" pattern rather than naked goroutines.

- **ISSUE-3: Missing Transactions in Multi-Step Repository Operations**
  - **Location:** `internal/modules/reaction/adapters/sqlite_repository.go`, Line 25-61
  - **Probability:** Medium
  - **Description:** The `Create` method resolves a `target_id` via a `SELECT` and then performs an `INSERT`. These are two separate database calls without a transaction. If the target post/comment is deleted between these calls, the insert will fail or create inconsistent data.
  - **Proposed Fix:** Wrap the target resolution and insertion in a `sql.Tx`.

## Performance & Optimization

- **PERF-1: Redundant Target Existence Checks**

  - **Description:** The `ReactionService` calls `s.postRepo.GetByID` to verify existence, and then `reactionRepo.Create` immediately calls `SELECT id FROM posts` again. This doubles the database round-trips for every reaction.
  - **Optimized Code:** Pass the internal `ID` from the service to the repository, or let the repository's foreign key constraints handle existence validation during the insert.

- **PERF-2: Subquery Bottleneck in Post List**
  - **Description:** `PostRepository.List` uses three subqueries in `LEFT JOIN`s to calculate counts for likes, dislikes, and comments. This will scale poorly (O(N\*M)) as the number of posts and reactions grows.
  - **Optimized Code:** Use denormalized counter columns in the `posts` table that are updated transactionally, or use a more efficient join/aggregation strategy.

## Nitpicks & Best Practices

- **Context.Background() Usage:** Avoid using `context.Background()` inside service methods. Always propagate the request context even to "background" tasks (unless they truly must outlive the request, in which case use a structured worker pool).
- **Inconsistent Error Mapping:** Repositories such as `comment` and `reaction` often return raw `sql.ErrNoRows` or other database errors instead of mapping them to domain errors consistently.
- **Middleware Overhead:** The `AuthMiddleware` fetches the full user record from the database on every authenticated request. Consider extracting only the `PublicID` if that's all that's needed, or implement a short-lived cache.
