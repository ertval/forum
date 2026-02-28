# Moderation Module - ID Handling Analysis Findings

## Overview
The Moderation module's ID handling implementation has been reviewed against the schema refactor guidelines that require using internal INT primary keys with public UUID identifiers for external exposure.

## Current Implementation Analysis

### Domain Layer (`domain/report.go`)
- ❌ **Partially Implemented**: Domain entity has only internal INT ID, no public UUID field
- ❌ Missing `PublicID string` field that should be exposed in API responses
- ❌ `ID int` - Internal unique identifier (INT PRIMARY KEY) ✓
- ❌ `ReporterID int` and `TargetID int` - Internal foreign keys (should be fine internally)
- ❌ JSON tags likely missing proper configuration to hide internal ID

### Ports Layer (`ports/repository.go` and `ports/service.go`)
- ❌ **Critical Issue**: Repository interface uses INT IDs instead of public UUIDs for external access
  - `GetByID(ctx context.Context, reportID int)` should be `GetByID(ctx context.Context, reportID string)` to use public_id
  - `Update(ctx context.Context, report *domain.Report)` should operate with public_id

- ❌ **Critical Issue**: Service interface uses INT IDs instead of public UUIDs
  - No explicit Get method, but if added would need to use public_id

### Application Layer (`application/service.go`)
- ❌ **Propagates Interface Issues**: Follows the incorrect interface design
- ❌ No explicit Get method would need to use public_id when implemented

### Adapter Layer (`adapters/sqlite_repository.go`)
- ❌ **Critical Issue**: Repository methods operate on internal INT IDs rather than public UUIDs
- ❌ `GetByID(ctx context.Context, reportID int)` queries by internal ID instead of public_id
- ❌ Other methods need to be updated to handle public_id

### HTTP Handler (`adapters/http_handler.go`)
- ⚠️ **Potentially Non-Compliant**: Currently only has TODO implementations
- ❌ When implemented, handlers might use internal INT IDs if they follow service interface
- ❌ `CreateReportAPI`, `ListReportsAPI`, `ReviewReportAPI` need to be reviewed for proper ID handling

## Security Analysis

### High-Risk Issues
- **ID Enumeration Vulnerability**: Using INT IDs in HTTP handlers exposes sequential internal identifiers that could be enumerated by attackers
- **Information Disclosure**: Exposing internal database IDs in any form provides information about the system's internal structure
- **Direct Object Reference**: Using internal IDs directly in HTTP endpoints creates Insecure Direct Object Reference (IDOR) vulnerabilities
- **Missing Public ID**: No public UUID implementation at all for reports

### Medium-Risk Issues
- **Inconsistent ID Handling**: Not following the established pattern of INT + UUID for other modules
- **Lack of Input Validation**: When handlers are implemented, they may not properly validate UUID format

## Recommendations

### Immediate Actions
1. **Update Domain Entity**: Add `PublicID string` field to `Report` struct
2. **Update Repository Interface**: Change all repository methods to use public_id (string) for external access
3. **Update Service Interface**: Consider if service methods need to use public_id (string) for external access
4. **Update Repository Implementation**: Modify repository methods to query by public_id instead of internal ID
5. **Add JSON tags**: Ensure internal ID is hidden in JSON responses

### Implementation Guidelines
- Add `PublicID string \`json:"id"\`` to Report struct
- Repository `GetByID` should query: `SELECT ... FROM reports WHERE public_id = ?`
- Repository `Update` should update WHERE `public_id = ?`
- Ensure internal foreign keys (ReporterID, TargetID) can still map to user/posts with proper UUID handling
- HTTP handlers should accept and return public UUIDs, never internal INT IDs

### Testing Considerations
- Write tests to verify that API endpoints return public UUIDs, not internal IDs
- Test that database queries use public_id for external lookups
- Verify that internal foreign keys still use INT for performance while maintaining security

## Compliance Status
**Current Status**: ❌ Non-Compliant
The Moderation module does not follow the required INT primary key + UUID public ID pattern for external interfaces.

**Required Changes**: 
- Domain entity: Add PublicID field
- Repository interface and implementation: Major changes needed
- Service interface: Minor changes needed  
- HTTP handlers: Implementation needed with proper ID handling