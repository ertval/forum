# Code Review: Comment Module

**Date:** 2026-01-14 15:35
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `comment` module provides standard CRUD functionality for comments on posts. It correctly integrates with the `post` module to verify parent existence and the `user` module for authorship. The implementation is generally clean, but there are race conditions in asynchronous counter updates and potential performance issues with pagination (offset-based) on large datasets.

## Critical Issues (Must Fix)

- **ISSUE-1: Race Condition in Async Stats Updates**

  - **Location:** `internal/modules/comment/application/service.go` (`CreateComment`, `DeleteComment`)
  - **Probability:** Medium
  - **Description:** Similar to the Post module, user comment counts are updated in a background goroutine without error handling or guarantees.
    ```go
    go func() {
        _ = s.userService.IncrementCommentCount(context.Background(), userID)
    }()
    ```
    If the server shuts down, these updates are lost.
  - **Proposed Fix:** Execute these synchronous to the request transaction, or use a proper background job queue with retry logic. For a monolithic app, synchronous updates in a transaction are preferred for data integrity.

- **ISSUE-2: Missing Transaction Boundary**
  - **Location:** `internal/modules/comment/application/service.go` (`CreateComment`)
  - **Probability:** Low
  - **Description:** When a comment is created, we insert the comment record AND increment the user's comment count. These are two separate DB operations (in different modules/repositories). If the second fails (even if synchronous), the data is inconsistent (comment exists but count is wrong).
  - **Proposed Fix:** In a strictly modular monolith, this is hard without distributed transactions. However, if using a shared DB instance, wrap the logic in a transaction passed via `context` or use an event-driven architecture where the `User` module listens for `CommentCreated` events. For KISS: Accept acceptable inconsistency or make updates synchronous and log errors fatallly.

## Performance & Optimization

- **PERF-1: Offset Pagination**
  - **Description:** `ListByUserPaginated` uses `LIMIT ? OFFSET ?`. As the offset grows (e.g., user has 10k comments), the DB must scan and discard `OFFSET` rows.
  - **Optimized Code:** Use Keyset Pagination (Seek Method) if possible, filtering by `created_at < ?` of the last item on the previous page.

## Nitpicks & Best Practices

- **Validation:** `c.Content` validation checks length of runes (`[]rune(c.Content)`). This is correct for Unicode, which is good.
- **Error Mapping:** In `http_handler_api.go`, explicit error mapping is done. This is good practice.
- **RESTful Design:** Routes like `POST /api/comments/posts/{post_id}` are slightly inconsistent with `GET /api/comments/{id}`. Usually, standard REST prefers `POST /api/posts/{post_id}/comments` for nested resources, but the current flat design is also acceptable.

---
