# Code Review - Moderation Module - 2026-01-14 15:35

## Executive Summary

The `moderation` module is currently in a **scaffolded state** and is non-functional. While the directory structure follows the project's hexagonal architecture and the database schema is correctly defined, the actual implementation of business logic (Application layer), data persistence (Adapters/SQLite layer), and API endpoints (Adapters/HTTP layer) is missing. The module contains numerous `TODO` placeholders and returns `Not Implemented` errors for all primary actions.

## Critical Issues (Must Fix)

### ISSUE-1: Module is Incomplete (Functional GAP)

- **Location:** `internal/modules/moderation/application/service.go`, `internal/modules/moderation/adapters/sqlite_repository.go`, `internal/modules/moderation/adapters/http_handler_api.go`
- **Probability:** High
- **Description:** All primary use cases (creating reports, reviewing reports, listing reports) are unimplemented. The service methods return `nil` or empty results, and the API handlers return `501 Not Implemented`.
- **Proposed Fix:** Implement the logic described in `internal/modules/moderation/flow.md`. Specifically:
  1. Implement `CreateReport` in `application/service.go` with target existence validation.
  2. Implement `ReviewReport` with actual status transitions and content deletion logic.
  3. Implement full SQL queries in `adapters/sqlite_repository.go`.

### ISSUE-2: Missing Role-Based Access Control (Security)

- **Location:** `internal/modules/moderation/adapters/http_handler_api.go`
- **Probability:** High
- **Description:** There are no permission checks on moderation endpoints. Even if the logic were implemented, any authenticated user could potentially access moderator-only endpoints if routes are registered without role verification.
- **Proposed Fix:** Implement a `RequireRole` middleware (as suggested in `flow.md`) or integrate with `user.UserService.CanModerate()` within the handlers or application service. The `auth.AuthMiddleware` currently only provides the `UserID`, so the role needs to be fetched or added to the context.

### ISSUE-3: Missing Cross-Module Integration for Content Validation

- **Location:** `internal/modules/moderation/application/service.go`:23
- **Probability:** High
- **Description:** `CreateReport` receives a `targetPublicID` but there is no logic to verify if the post or comment actually exists. A user could spam reports for non-existent content.
- **Proposed Fix:** Inject `post.PostService` and `comment.CommentService` into the `moderation.Service` to validate targets before creating a report.

## Performance & Optimization

### PERF-1: N+1 Hazard in Report Listing

- **Description:** The `List` method in `sqlite_repository.go` (placeholder) and the `domain.Report` struct don't yet account for fetching reporter/moderator usernames. If the UI needs to show usernames, fetching them one by one after listing reports will cause an N+1 query problem.
- **Optimized Code:** Update the repository to use SQL JOINs:

```sql
SELECT r.*, u.username as reporter_username, m.username as moderator_username
FROM reports r
LEFT JOIN users u ON r.reporter_id = u.id
LEFT JOIN users m ON r.moderator_id = m.id
```

## Nitpicks & Best Practices

1. **Inconsistent Status Naming**: `domain/report.go` defines `StatusResolved`, but `flow.md` uses `dismissed`. Stick to one terminology for clarity.
2. **Missing Validation**: `domain.Report.IsValid()` only checks the target type. It should also validate the `Reason` length and `Status` values.
3. **Internal IDs in Mocks**: `application/service_test.go` uses a manual ID-based mapping in its mock which won't reflect the production behavior of UUID-based lookups.
4. **Flow Documentation Divergence**: `internal/modules/moderation/flow.md` displays complex logic and middleware that simply does not exist in the codebase. This creates confusion for anyone joining the project.

---

**Reviewer:** Antigravity (Principal AI Engineer)
**Date:** 2026-01-14
