# Code Review: Moderation Module - 2026-01-14

## Executive Summary

The `moderation` module is currently in a **scaffold state**. While the hexagonal structure is correctly established, the majority of the business logic remains unimplemented (placeholders). There is a significant discrepancy between the detailed design documentation in `flow.md` and the actual Go implementation. Critical security features like RBAC and full input validation are missing.

## Critical Issues (Must Fix)

- **ISSUE-1: Major Discrepancy Between Documentation and Implementation**

  - **Location:** `internal/modules/moderation/flow.md` vs multiple files.
  - **Probability:** High (misleads developers)
  - **Description:** `flow.md` describes a complete implementation including `RequireRole` middleware, specific SQL queries, and cross-module cascading deletes. However, the actual code (`sqlite_repository.go`, `service.go`, `http_handler_api.go`) contains only empty placeholders or `501 Not Implemented` responses. This creates a false sense of module readiness and can lead to integration bugs.
  - **Proposed Fix:** Synchronize `flow.md` with the current state of implementation, or prioritize the implementation of the features described in the documentation.

- **ISSUE-2: Silent Placeholder Failure in Service**

  - **Location:** `internal/modules/moderation/application/service.go`, Lines 23-30
  - **Probability:** High
  - **Description:** The `CreateReport` method returns `nil` (success) without performing any validation or persistence. If this method is called by other modules (e.g., during testing or early integration), it will falsely indicate success while the report is lost.
  - **Proposed Fix:** Use a proper "not implemented" error or implement the logic.
    ```go
    func (s *Service) CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) error {
        return errors.New("not implemented")
    }
    ```

- **ISSUE-3: Missing Role-Based Access Control (RBAC)**
  - **Location:** `internal/modules/moderation/adapters/http_handler_api.go`
  - **Probability:** High (Security)
  - **Description:** Moderation endpoints (List, Review) are registered without any authorization middleware. These endpoints should only be accessible by users with `Moderator` or `Administrator` roles.
  - **Proposed Fix:** Implement or use a standard middleware (like `auth.RequireRole`) as described in `flow.md` and apply it to sensitive routes.

## Performance & Optimization

- **PERF-1: N/A**
  - **Description:** Code is currently too minimal for performance analysis.

## Nitpicks & Best Practices

- **NIT-1: Weak Domain Validation**
  - **Location:** `internal/modules/moderation/domain/report.go`, Line 31
  - **Description:** `IsValid()` only checks `TargetType`. It should also check if `Reason` is empty and if `Status` is valid, as noted in its own TODO comments.
- **NIT-2: Missing Dependency Injection for Target Resolution**
  - **Location:** `internal/modules/moderation/application/service.go`
  - **Description:** To resolve `targetPublicID` to an internal ID (as planned in the comments), the service will need `PostService` and `CommentService` ports. These are not yet injected.
- **NIT-3: Inconsistent Repository Comments**
  - **Location:** `internal/modules/moderation/ports/repository.go` vs `application/service.go`
  - **Description:** One comment says the repo "Must generate and set PublicID", while the application service comment says "(repo generates PublicID)". While consistent in intent, the wording "Must generate" in the port interface is better practice to enforce implementation requirements.

---
