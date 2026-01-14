# Code Review: User Module

**Date:** 2026-01-14 15:31
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `user` module is well-structured following the Hexagonal Architecture pattern. The distinction between internal integer IDs and public UUIDs is strictly enforced, which is excellent for security. However, there are minor abstraction leaks where database errors bubble up to the handler, and some potential race conditions in state updates.

## Critical Issues (Must Fix)

- **ISSUE-1: Race Condition in State Updates**

  - **Location:** `internal/modules/user/application/service.go` (`UpdateRole`, `DeactivateUser`, `ActivateUser`)
  - **Probability:** Low (Admin actions)
  - **Description:** The Update operations follow a "Check-Then-Act" pattern. The service fetches the user, modifies the struct in memory, and then saves it.
    ```go
    user, _ := repo.GetByID(id) // Read
    user.IsActive = false       // Modify
    repo.Update(user)           // Write
    ```
    If two concurrent requests occur, one might overwrite the other.
  - **Proposed Fix:** Use atomic SQL updates in the repository for specific status changes, or use optimistic locking (versioning).
    ```go
    // In Repository
    func (r *SQLiteUserRepository) UpdateStatus(ctx context.Context, userID int, isActive bool) error {
        query := `UPDATE users SET is_active = ?, updated_at = ? WHERE id = ?`
        // ...
    }
    ```

- **ISSUE-2: Abstraction Leak (SQL Errors)**
  - **Location:** `internal/modules/user/application/service.go`
  - **Probability:** High
  - **Description:** The Service layer methods (e.g., `GetByID`, `GetByEmail`) return errors directly from the repository. If the repository returns `sql.ErrNoRows`, this driver-specific error leaks to the HTTP handler. While the handler currently checks `err != nil`, relying on implementation details violates the architectural boundaries.
  - **Proposed Fix:** Wrap `sql.ErrNoRows` in the Repository or Service layer.
    ```go
    // In Service.GetByID
    user, err := s.userRepo.GetByID(ctx, userID)
    if err == sql.ErrNoRows {
        return nil, domain.ErrUserNotFound
    }
    return user, err
    ```

## Performance & Optimization

- **PERF-1: SQLite Boolean Handling**
  - **Description:** The `Scan` logic in `sqlite_repository.go` repeatedly handles the `isActive` int-to-bool conversion manually in every method (`GetByID`, `List`, etc.).
  - **Optimized Code:** Consolidate the scanning logic into a helper method or private struct method to reduce code duplication and potential errors.

## Nitpicks & Best Practices

- **Validation:** `CreateUser` in the Service does not validate email format or username length. This should be added, preferably in the Domain layer (`user.Validate()`).
- **Error Messages:** `CreateUser` does not check if the user already exists before attempting insertion. Relying solely on DB uniqueness constraints creates opaque errors. It is better to check `ExistsByEmail` first or handle the specific constraint violation error.
- **Hardcoded Ordering:** `ListUsers` in the repository hardcodes `ORDER BY created_at DESC`. It might be beneficial to allow sorting options in the future.

---
