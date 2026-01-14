# Go Code Simplifier Review

**Folder/Module:** notification
**Date:** 2026-01-14 15:40
**Files Reviewed:**

- `internal/modules/notification/domain/notification.go`
- `internal/modules/notification/domain/errors.go`
- `internal/modules/notification/ports/service.go`
- `internal/modules/notification/ports/repository.go`
- `internal/modules/notification/application/service.go`
- `internal/modules/notification/adapters/http_handler.go`
- `internal/modules/notification/adapters/http_handler_api.go`
- `internal/modules/notification/adapters/sqlite_repository.go`
- `migrations/007_notification_create_notifications.sql`

---

## Summary

The `notification` module is currently in a scaffolded state with numerous placeholders and `TODO` comments. While the basic structure follows the project's hexagonal architecture, there are significant inconsistencies between the domain model, the database schema, and the implementation. The tests are present but largely function as "placeholder tests" that pass even with empty logic, which violates TDD principles.

---

## Findings

### 1. Inconsistency Between Domain Struct and DB Schema

**File:** `internal/modules/notification/domain/notification.go`, `migrations/007_notification_create_notifications.sql`
**Category:** Architecture / KISS Violation
**Severity:** High

**Current Code (domain/notification.go):**

```go
type Notification struct {
	ID        int       `json:"-"`
	PublicID  string    `json:"id"`
	UserID    int       `json:"-"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	TargetID  int       `json:"-"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	PublicTargetID string `json:"target_id,omitempty"`
}
```

**Current Code (migrations/007...sql):**

```sql
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    ...
);
```

**Suggested Improvement:**
Update the domain struct to include `ActorID` and `PublicActorID`, and align field names with the database where possible (e.g., `read` vs `IsRead`).

```go
type Notification struct {
	ID             int       `json:"-"`
	PublicID       string    `json:"id"`
	UserID         int       `json:"-"`
	ActorID        int       `json:"-"`
	PublicActorID  string    `json:"actor_id,omitempty"`
	Type           string    `json:"type"`
	Message        string    `json:"message"`
	TargetID       int       `json:"-"`
	PublicTargetID string    `json:"target_id,omitempty"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}
```

**Rationale:** The database schema tracks who performed the action (`actor_id`), but the domain model does not. This will lead to data loss or mapping errors during implementation.

---

### 2. Missing Domain Validation

**File:** `internal/modules/notification/domain/notification.go`
**Category:** Architecture
**Severity:** Medium

**Current Code:**

```go
// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}
```

**Suggested Improvement:**
Add a `Validate()` method to the `Notification` entity as required by `GEMINI.md`.

```go
func (n *Notification) Validate() error {
	if n.UserID <= 0 {
		return errors.New("invalid user id")
	}
	if n.Type == "" {
		return ErrInvalidNotificationType
	}
	if n.Message == "" {
		return errors.New("message is required")
	}
	return nil
}
```

**Rationale:** Every domain entity should be responsible for its own validity to ensure business rules are enforced consistently.

---

### 3. TDD Violation: Empty Tests Passing

**File:** `internal/modules/notification/application/service_test.go`, `internal/modules/notification/adapters/sqlite_repository_test.go`
**Category:** TDD
**Severity:** High

**Current Code (service_test.go):**

```go
func TestService_CreateNotification(t *testing.T) {
	// ... setup ...
	err := service.CreateNotification(ctx, 1, domain.TypeLike, "message", "target-10")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
```

**Rationale:** The tests for `CreateNotification` pass because the implementation is an empty function that returns `nil`. These tests provide zero confidence. Tests should be written to assert that a notification was actually created (e.g., by checking the mock repository's state).

---

### 4. Direct Use of `int` for Cross-Module References

**File:** `internal/modules/notification/domain/notification.go`
**Category:** Architecture
**Severity:** Medium

**Current Code:**

```go
	TargetID  int       `json:"-"`          // Internal ID of the related entity (post, comment, etc.)
```

**Suggested Improvement:**
Consider if `TargetID` is actually needed internally if the entity is from another module. If the modular monolith strictly isolates database tables, the `notification` module might only ever know the `PublicTargetID` (UUID) of the post/comment.

**Rationale:** Storing internal `int` IDs of entities owned by other modules couples the `notification` module to those modules' internal database details.

---

## Action Items

- [ ] Add `ActorID` and `PublicActorID` to the `Notification` domain struct.
- [ ] Implement the `Validate()` method in `domain/notification.go`.
- [ ] Update `CreateNotification` to actually implement the logic (UUID generation, repository call).
- [ ] Refactor tests to be meaningful (assert state changes, not just lack of errors from placeholders).
- [ ] Standardize field naming between SQL (`read`) and Go (`IsRead`) or use struct tags correctly.
- [ ] Implement `sqlite_repository.go` logic (at least the basic INSERT/SELECT).

---

## Notes

The module is marked as an `[OPTIONAL FEATURE]`, which explains its current scaffolded state. However, if implementation proceeds, these architectural alignment issues should be addressed first to avoid technical debt.
