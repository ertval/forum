## Executive Summary

The `user` module provides a solid foundation for user management with a clean Hexagonal Architecture and proper separation of concerns. However, it suffers from several critical data synchronization issues—specifically, missing fields in the repository layer—and performance inefficiencies due to redundant database lookups in the API handlers.

## Critical Issues (Must Fix)

- **ISSUE-1: Incomplete Data Persistence (Missing reaction_count)**

  - **Location:** `internal/modules/user/adapters/sqlite_repository.go`, multiple lines (Scan, Insert, Update)
  - **Probability:** High
  - **Description:** While the `domain.User` struct and the SQLite table schema both include `reaction_count`, the repository's `SELECT`, `INSERT`, and `UPDATE` queries completely omit this field. This means the user's reaction count is never loaded into memory, and newly created users or general updates do not persist this value, leading to data loss or "zeroed" counts in the UI/API.
  - **Proposed Fix:** Update all SQL queries in `sqlite_repository.go` to include the `reaction_count` column and update the corresponding `Scan` calls and `ExecContext` arguments.

- **ISSUE-2: Data Race and Potential Data Loss in Service Updates**

  - **Location:** `internal/modules/user/application/service.go`, Lines 70-123
  - **Probability:** Medium
  - **Description:** Methods like `UpdateRole`, `DeactivateUser`, and `ActivateUser` follow a "Read-Modify-Write" pattern: they fetch the user, modify a field, and then call `Update`. If two concurrent requests update different fields (e.g., one updates the role while another updates the username—if implemented), the second `Update` will overwrite the first one's changes because it was based on stale data.
  - **Proposed Fix:** Use more granular repository methods for specific updates (e.g., `UpdateRole(ctx, userID, role)`) or implement optimistic locking (e.g., `WHERE id = ? AND updated_at = ?`).

- **ISSUE-3: Missing Domain Fields (OAuth)**
  - **Location:** `internal/modules/user/domain/user.go` vs `migrations/002_user_create_users.sql`
  - **Probability:** Low (Future-proofing)
  - **Description:** The database schema includes `oauth_provider` and `oauth_provider_id`, but these are missing from the `domain.User` entity. This prevents the application from supporting OAuth-based authentication even though the schema is ready.
  - **Proposed Fix:** Add `OAuthProvider` and `OAuthProviderID` fields to the `User` struct in `domain/user.go`.

## Performance & Optimization

- **PERF-1: Redundant Database Lookups in API Handlers**

  - **Description:** Each update API (`UpdateRoleAPI`, `DeactivateUserAPI`, etc.) fetches the user by `PublicID` in the handler just to find the internal `ID`, and then the service layer fetches the user _again_ by `ID`. This results in 2 reads and 1 write for every single update.
  - **Optimized Code:**
    ```go
    // In application/service.go, change method signature to accept PublicID or pass the whole user
    func (s *Service) UpdateRole(ctx context.Context, publicID string, newRole domain.Role) error {
        user, err := s.userRepo.GetByPublicID(ctx, publicID)
        // ... update and save
    }
    ```
    This avoids the extra trip to the DB in the handler.

- **PERF-2: Use of MAX(0, count - 1) in SQL**
  - **Description:** The current implementation uses `MAX(0, count - 1)` which is excellent for safety. However, ensures the DB index is utilized if these counts ever become part of a filter/sort. (Note: This is actually a positive highlight of the current code).

## Nitpicks & Best Practices

- **NIT-1: Semantic Content-Type in Error Responses**
  - **Location:** `adapters/http_handler_api.go`
  - **Description:** `http.Error` defaults to `text/plain`. Since the response body is explicitly JSON (`{"error": "..."}`), the `Content-Type` should be set to `application/json` before calling `WriteHeader` and writing the body.
- **NIT-2: Duplicated Logic**
  - **Description:** Role validation logic exists in both the Service and the HTTP Handler. This should be unified in the Domain layer (e.g., a `Role.IsValid()` method).
- **NIT-3: Missing Input Validation**
  - **Description:** `CreateUser` does not validate email format or username length/characters. This should be handled in the service or domain layer before hitting the database.
- **NIT-4: ID Security Rule Compliance**
  - **Observation:** The code correctly follows the instruction of using `PublicID` (UUID) for external exposure and `ID` (INT) internally. Good job on `GetByPublicID`.
- **NIT-5: File Headers**
  - **Observation:** All files correctly include the required headers (e.g., `// INPUT PORT - Service Interface`).
