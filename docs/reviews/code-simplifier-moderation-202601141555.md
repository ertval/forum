# Go Code Simplifier Review

**Folder/Module:** moderation
**Date:** 2026-01-14 15:55
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

The moderation module is currently in a scaffolded state with several placeholders ("TODO") and unimplemented logic. The structure follows the clean modular monolith architecture (Hexagonal/Ports & Adapters) as defined in the project rules. This review focuses on refining the domain models, improving type safety, and ensuring consistency across the module's interfaces to prepare for full implementation.

---

## Findings

### 1. Type Safety for Reports (Status and TargetType)

**File:** `internal/modules/moderation/domain/report.go`
**Line(s):** 13, 15, 23-27
**Category:** Idiomatic Go | KISS Violation
**Severity:** Medium

**Current Code:**

```go
type Report struct {
    // ...
    TargetType string    `json:"target_type"` // Type of target: "post" or "comment"
    Status     string    `json:"status"`      // Report status: "pending", "reviewed", "resolved"
    // ...
}

const (
    StatusPending  = "pending"
    StatusReviewed = "reviewed"
    StatusResolved = "resolved"
)
```

**Suggested Improvement:**

```go
type ReportStatus string
type TargetType string

const (
    StatusPending  ReportStatus = "pending"
    StatusReviewed ReportStatus = "reviewed"
    StatusResolved ReportStatus = "resolved"

    TargetPost    TargetType = "post"
    TargetComment TargetType = "comment"
)

type Report struct {
    // ...
    TargetType TargetType   `json:"target_type"`
    Status     ReportStatus `json:"status"`
    // ...
}
```

**Rationale:** Using custom types for status and target types improves code clarity and enables the compiler to catch errors if someone tries to pass an arbitrary string where a status is expected. This is idiomatic in Go for small sets of predefined values.

---

### 2. Incomplete Domain Validation

**File:** `internal/modules/moderation/domain/report.go`
**Line(s):** 31-36
**Category:** TDD | Architecture
**Severity:** High

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
func (r *Report) Validate() error {
    if r.TargetType != TargetPost && r.TargetType != TargetComment {
        return ErrInvalidTargetType
    }
    if r.Status != StatusPending && r.Status != StatusReviewed && r.Status != StatusResolved {
        return ErrInvalidReportStatus
    }
    if strings.TrimSpace(r.Reason) == "" {
        return ErrEmptyReason
    }
    if r.ReporterID <= 0 {
        return ErrInvalidReporter
    }
    return nil
}
```

**Rationale:** Domain entities should validate their own state. Returning an `error` rather than a `bool` provides more context about why validation failed. This aligns with the "Refactor" phase of TDD once implementation starts.

---

### 3. Service Interface Simplification

**File:** `internal/modules/moderation/ports/service.go`
**Line(s):** 12-22
**Category:** KISS Violation | Architecture
**Severity:** Medium

**Current Code:**

```go
type ModerationService interface {
	CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) error
	ReviewReport(ctx context.Context, reportPublicID string, decision string) error
	ListReports(ctx context.Context, status string) ([]*domain.Report, error)
}
```

**Suggested Improvement:**

```go
type CreateReportRequest struct {
    ReporterID     int
    TargetPublicID string
    TargetType     domain.TargetType
    Reason         string
}

type ReviewReportRequest struct {
    ReportPublicID string
    ModeratorID    int
    Decision       string // Consider making this an enum too
    Notes          string
}

type ModerationService interface {
	CreateReport(ctx context.Context, req CreateReportRequest) error
	ReviewReport(ctx context.Context, req ReviewReportRequest) error
	ListReports(ctx context.Context, status domain.ReportStatus) ([]*domain.Report, error)
}
```

**Rationale:** Passing primitive arguments to service methods (Primitive Obsession) can lead to large argument lists as requirements grow. Using request structs makes the API more flexible and easier to extend without breaking the interface.

---

### 4. Alignment of Adapter Placeholders with GEMINI.md

**File:** `internal/modules/moderation/adapters/http_handler_api.go`
**Line(s):** 11-18
**Category:** Architecture | IDE Security
**Severity:** Medium

**Current Code:**

```go
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	// POST /api/moderation/reports - Create report
	router.HandleFunc("POST /api/moderation/reports", h.CreateReportAPI)
	// GET /api/moderation/reports - List reports (filtered by status)
	router.HandleFunc("GET /api/moderation/reports", h.ListReportsAPI)
	// PUT /api/moderation/reports/{id} - Review report
	router.HandleFunc("PUT /api/moderation/reports/{id}", h.ReviewReportAPI)
}
```

**Suggested Improvement:**
Ensure the routes follow the established pattern `/api/{module}/{action}` if applicable, although `/api/moderation/reports` is acceptable since "reports" is the resource. However, `GEMINI.md` suggests `/api/{module}/{action}`.

**Rationale:** Consistency in API design helps with frontend integration and overall system predictability.

---

### 5. Repository Error Wrapping

**File:** `internal/modules/moderation/adapters/sqlite_repository.go`
**Line(s):** Entire file (placeholders)
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:** (Mental model of future implementation)

```go
func (r *SQLiteReportRepository) Create(ctx context.Context, report *domain.Report) error {
    // ...
    return nil
}
```

**Suggested Improvement:**
When implementing, ensure all database errors are wrapped to provide context.

```go
func (r *SQLiteReportRepository) Create(ctx context.Context, report *domain.Report) error {
    // ...
    if err := r.db.ExecContext(...); err != nil {
        return fmt.Errorf("create report: %w", err)
    }
    return nil
}
```

**Rationale:** Idiomatic Go error handling involves wrapping errors to provide a trace of what was happening when the error occurred.

---

## Action Items

- [ ] Define `ReportStatus` and `TargetType` as custom types in `domain/report.go`.
- [ ] Implement a full `Validate() error` method in `domain/report.go`.
- [ ] Add missing error constants to `domain/errors.go` (e.g., `ErrInvalidTargetType`, `ErrEmptyReason`).
- [ ] Refactor `ModerationService` and `ReportRepository` to use domain types and possibly request structs.
- [ ] Remove TODOs by implementing the logic in `application/service.go` and `adapters/sqlite_repository.go`.
- [ ] Add unit tests for `domain.Report.Validate()` to follow TDD principles.

---

## Notes

The module is well-structured from a Hexagonal perspective. The main task is moving from "Scaffold" to "Implementation" while maintaining the strict boundary between domain/ports/application/adapters. The `flow.md` file provides an excellent blueprint, but care should be taken to synchronize it with the specific ID naming conventions (PublicID vs ID) defined in `GEMINI.md`.
