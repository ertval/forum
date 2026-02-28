# Moderation Module Schema Refactor Audit

## Overview
Audit of the moderation module (optional feature) against schema requirements.

## Findings

### ✅ Correctly Implemented
- **Migration (006_moderation_create_reports.sql)**: Has `id INTEGER PRIMARY KEY AUTOINCREMENT`, `public_id TEXT UNIQUE NOT NULL`

### ❌ Issues Found

#### 1. Domain Entity Missing PublicID
**File**: `internal/modules/moderation/domain/report.go`
**Issue**: `Report` struct has `ID int` but no `PublicID string`
**Required**:
```go
type Report struct {
    ID           int       `json:"-"`                    // Internal
    PublicID     string    `json:"id"`                   // Public UUID
    ReporterID   int       `json:"-"`                    // Internal
    TargetID     int       `json:"-"`                    // Internal
    TargetType   string    `json:"target_type"`
    Reason       string    `json:"reason"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
    // For API:
    PublicReporterID string `json:"reporter_id,omitempty"`
    PublicTargetID   string `json:"target_id,omitempty"`
}
```

#### 2. Service Interface Issues
**File**: `internal/modules/moderation/ports/service.go`
**Issues**:
- `CreateReport(ctx context.Context, reporterID, targetID int, ...)` → `targetID string`
- `ReviewReport(ctx context.Context, reportID int, ...)` → `reportID string`

#### 3. Repository Interface Issues
**File**: `internal/modules/moderation/ports/repository.go`
**Issues**:
- `GetByID(ctx context.Context, reportID int)` → `reportID string`
- `Update(ctx context.Context, report *domain.Report)` → OK
- `List(ctx context.Context, status string)` → OK

#### 4. All Implementations (TODO)
**Status**: All application, adapters are TODO placeholders

## Security Analysis

### High Risk Issues
1. **Report ID Exposure**: Sequential IDs in moderation reports could allow enumeration of all reports
2. **Target Privacy**: Reports contain sensitive information about inappropriate content
3. **Moderator Access Control**: Must ensure only moderators can access reports

### Medium Risk Issues
1. **Information Disclosure**: Report reasons and targets could leak if not properly secured
2. **Audit Trail**: Reports should be immutable once created

## Recommendations

### Implementation Requirements
1. Add PublicID to Report domain entity
2. Update all interfaces to use string public_ids for external access
3. Implement proper authorization (moderator-only access)
4. Add audit logging for report reviews

### Security Controls
- Rate limiting on report creation
- Moderator authentication for all endpoints
- Input validation for report reasons
- Audit trail for status changes

## Test Recommendations

### Unit Tests
1. **Domain**: Test Report validation
2. **Repository**: Test UUID generation
3. **Service**: Test authorization checks

### Integration Tests
1. **API**: Test report creation with public_id targets
2. **Security**: Test non-moderators cannot access reports
3. **Security**: Test report enumeration prevention
4. **Audit**: Test status change logging</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/moderation_module_audit.md