# Go Code Simplifier Review

**Folder/Module:** moderation
**Date:** 2026-01-14 15:35
**Files Reviewed:**

- `internal/modules/moderation/domain/report.go`
- `internal/modules/moderation/domain/errors.go`
- `internal/modules/moderation/ports/service.go`
- `internal/modules/moderation/ports/repository.go`
- `internal/modules/moderation/application/service.go`
- `internal/modules/moderation/adapters/http_handler.go`
- `internal/modules/moderation/adapters/http_handler_api.go`
- `internal/modules/moderation/adapters/sqlite_repository.go`

---

## Summary

The `moderation` module is currently in a scaffolded state with most implementations being placeholders or minimal. While the structure follows the project's hexagonal architecture and modular monolith patterns, there are several opportunities to simplify and improve the code as implementation proceeds. The primary focus should be on consistent ID handling, robust validation, and idiomatic Go patterns in the upcoming implementation.

---

## Findings

### Incomplete Domain Validation

**File:** `internal/modules/moderation/domain/report.go`
**Line(s):** 31-36
**Category:** KISS Violation
**Severity:** Medium

**Current Code:**

```go
func (r *Report) IsValid() bool {
	// Check target type is "post" or "comment"
	// Check status is valid
	// Check reason is not empty
	return r.TargetType == "post" || r.TargetType == "comment"
}
```

**Suggested Improvement:**

```go
func (r *Report) IsValid() bool {
	if r.Reason == "" {
		return false
	}
	if r.TargetType != "post" && r.TargetType != "comment" {
		return false
	}
	switch r.Status {
	case StatusPending, StatusReviewed, StatusResolved:
		return true
	default:
		return false
	}
}
```

**Rationale:** The current implementation is a placeholder. Completing it with explicit checks and using a `switch` for status validation makes it clearer and safer.

---

### Inconsistent ID Handling in Service Interface

**File:** `internal/modules/moderation/ports/service.go`
**Line(s):** 15
**Category:** Architecture
**Severity:** Medium

**Current Code:**

```go
CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) error
```

**Suggested Improvement:**

```go
CreateReport(ctx context.Context, reporterPublicID string, targetPublicID string, targetType, reason string) error
```

**Rationale:** The `AuthMiddleware` provides the user's Public ID (UUID) in the context. Using Public IDs for all external entity references in the service interface maintains consistency and makes the handler's job easier (no need to resolve the reporter UUID to an internal INT ID before calling the service). The service should handle internal ID resolution if needed.

---

### Redundant Public ID fields in Domain Entity

**File:** `internal/modules/moderation/domain/report.go`
**Line(s):** 18-19
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
	// For API responses - public UUIDs of related entities
	PublicReporterID string `json:"reporter_id,omitempty"` // Public UUID of reporter
	PublicTargetID   string `json:"target_id,omitempty"`   // Public UUID of reported content
```

**Suggested Improvement:**
Keep the fields but ensure the service or repository populates them only when needed for API responses. Alternatively, use a separate DTO for API responses if the domain entity becomes too cluttered with "Public" mirror fields.

**Rationale:** While these fields are necessary for JSON responses (as internal IDs are ignored), having them in the core domain entity creates a "double bookkeeping" problem. If the project's pattern is to include them in domain entities for simplicity, ensure they are consistently populated.

---

### Generic Error Handling in Handlers

**File:** `internal/modules/moderation/adapters/http_handler_api.go`
**Line(s):** 22, 29, 36
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
// Implementation placeholder
http.Error(w, "Not implemented", http.StatusNotImplemented)
```

**Suggested Improvement:**
When implementing, use the existing `platform/errors` package to return consistent JSON error responses:

```go
platformErrors.WriteErrorJSON(w, http.StatusNotImplemented, "Endpoint not yet implemented")
```

**Rationale:** Consistency in API error responses is crucial for client-side handling.

---

### Missing Repository Implementation Details

**File:** `internal/modules/moderation/adapters/sqlite_repository.go`
**Line(s):** 25-57
**Category:** KISS Violation
**Severity:** Medium

**Current Code:**
Placeholders returning `nil`.

**Suggested Improvement:**
Implement using standard library `sql` patterns. For example, in `Create`:

```go
func (r *SQLiteReportRepository) Create(ctx context.Context, report *domain.Report) error {
    if report.PublicID == "" {
        report.PublicID = uuid.Must(uuid.NewV4()).String()
    }

    query := `INSERT INTO reports (public_id, reporter_id, target_id, target_type, reason, status, created_at)
              VALUES (?, ?, ?, ?, ?, ?, ?)`

    _, err := r.db.ExecContext(ctx, query,
        report.PublicID, report.ReporterID, report.TargetID,
        report.TargetType, report.Reason, report.Status, time.Now())

    if err != nil {
        return fmt.Errorf("failed to create report: %w", err)
    }
    return nil
}
```

**Rationale:** Proper error wrapping with `%w` and using `ExecContext` for cancellations/timeouts are idiomatic Go practices.

---

## Action Items

- [ ] Complete `domain.Report.IsValid()` implementation.
- [ ] Align `ModerationService` interface with the project's "UUID for public API" strategy by using Public IDs for reporter identification.
- [ ] Implement `application.Service` logic, including validation and ID resolution.
- [ ] Implement `adapters.SQLiteReportRepository` with proper error handling and UUID generation.
- [ ] Update `adapters.HTTPHandler` to use the `AuthMiddleware` to extract the reporter's Public ID.
- [ ] Flesh out tests in `application/service_test.go` and `adapters/sqlite_repository_test.go` to cover edge cases and error paths.

---

## Notes

The module is well-structured according to the modular monolith pattern. The use of `[OPTIONAL FEATURE: forum-moderation]` headers indicates clear feature flagging. Once the placeholders are filled, the module will be a solid addition to the codebase.
