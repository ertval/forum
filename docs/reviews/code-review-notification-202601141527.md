## Executive Summary

The `notification` module is currently in a scaffold state. It defines the necessary interfaces and directory structure according to the project's architecture (Hexagonal/Modular Monolith), but the actual implementation of services and repositories is missing. Most methods are placeholders returning `nil` or `http.StatusNotImplemented`. Additionally, there is a discrepancy between the domain entity and the database schema regarding the `actor_id` field.

## Critical Issues (Must Fix)

- **ISSUE-1: Incomplete Implementation (Module Scaffold Only)**

  - **Location:** Multiple files (`application/service.go`, `adapters/sqlite_repository.go`, `adapters/http_handler_api.go`)
  - **Probability:** High (Functional failure)
  - **Description:** The core notification logic is not implemented. Methods like `CreateNotification`, `GetByUserID`, and `MarkAsReadByPublicID` are placeholders. This means no notifications are actually generated or persisted.
  - **Proposed Fix:** Implement the logic in `application/service.go` and `adapters/sqlite_repository.go` using the SQL queries suggested in the comments.

- **ISSUE-2: Domain Model and Database Schema Mismatch**

  - **Location:** `domain/notification.go` vs `migrations/007_notification_create_notifications.sql`
  - **Probability:** High
  - **Description:** The `Notification` struct is missing the `ActorID` field which is defined as `NOT NULL` in the database schema (`actor_id INTEGER NOT NULL`). Any attempt to insert data using the current struct will likely fail or lead to data loss of who triggered the notification.
  - **Proposed Fix:** Add `ActorID` (internal) and `PublicActorID` (public) to the `Notification` struct.
    ```go
    type Notification struct {
        // ... existing fields
        ActorID       int    `json:"-"`
        PublicActorID string `json:"actor_id,omitempty"`
    }
    ```

- **ISSUE-3: Missing Validation in Service Layer**
  - **Location:** `application/service.go`, Line 23
  - **Probability:** Medium
  - **Description:** The `CreateNotification` method has a `TODO` for validation but currently does nothing. It should validate the `notifType` against allowed constants.
  - **Proposed Fix:** Implement validation using a `Validate()` method on the domain entity or directly in the service.

## Performance & Optimization

- **PERF-1: Database Indexing**
  - **Description:** The current migration provides good indexing for retrieving notifications by user ID and read status, which is the most common use case.
  - **Optimized Code:** (Migrations are already well-optimized with `idx_notifications_user` on `(user_id, read)`).

## Nitpicks & Best Practices

- **NIT-1: Use `ErrNotImplemented` for Placeholders**
  - Instead of returning `nil`, it's better to return a specific error during development to avoid silent failures if a method is called unexpectedly.
- **NIT-2: Consistent Field Naming**
  - The database uses `read` but the struct uses `IsRead`. While common in Go, keeping them aligned in terms of meaning is good. The JSON tag `is_read` is consistent with modern API practices.
- **NIT-3: MarkAsRead Return Value**
  - The `application/service.go` `MarkAsRead` method returns any error from the repo, which is good. Ensure the repo returns `domain.ErrNotificationNotFound` if the ID doesn't exist.

---
