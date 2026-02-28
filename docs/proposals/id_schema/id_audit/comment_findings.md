# Comment Module - ID Handling Analysis Findings

## Overview
The Comment module's ID handling implementation has been reviewed against the schema refactor guidelines that require using internal INT primary keys with public UUID identifiers for external exposure.

## Current Implementation Analysis

### Domain Layer (`domain/comment.go`)
- ✅ **Correctly implemented**: Domain entity has both internal INT ID and public UUID fields
- ✅ `ID int` - Internal unique identifier (INT PRIMARY KEY)
- ✅ `PublicID string` - Public UUID identifier (exposed in API)
- ✅ `PostID int` and `UserID int` - Internal foreign keys
- ✅ JSON tags properly configured to hide internal IDs

### Ports Layer (`ports/repository.go` and `ports/service.go`)
- ❌ **Critical Issue**: Repository interface uses INT IDs instead of public UUIDs for external access
  - `GetByID(ctx context.Context, commentID int)` should be `GetByID(ctx context.Context, commentID string)` to use public_id
  - `Update(ctx context.Context, comment *domain.Comment)` should operate on public_id
  - `Delete(ctx context.Context, commentID int)` should be `Delete(ctx context.Context, commentID string)`
  - `ListByPostID(ctx context.Context, postID int)` should potentially use public post UUID

- ❌ **Critical Issue**: Service interface uses INT IDs instead of public UUIDs
  - `GetComment(ctx context.Context, commentID int)` should be `GetComment(ctx context.Context, commentID string)`
  - `UpdateComment(ctx context.Context, commentID int, content string)` should be `UpdateComment(ctx context.Context, commentID string, content string)`
  - `DeleteComment(ctx context.Context, commentID int)` should be `DeleteComment(ctx context.Context, commentID string)`
  - `ListCommentsByPost(ctx context.Context, postID int)` should potentially use public post UUID

### Application Layer (`application/service.go`)
- ❌ **Propagates Interface Issues**: Follows the incorrect interface design
- ❌ Uses internal INT IDs in service operations rather than public UUIDs

### Adapter Layer (`adapters/sqlite_repository.go`)
- ❌ **Critical Issue**: Repository methods operate on internal INT IDs rather than public UUIDs
- ❌ `GetByID(ctx context.Context, commentID int)` queries by internal ID instead of public_id
- ❌ `Delete(ctx context.Context, commentID int)` deletes by internal ID instead of public_id
- ❌ Other methods use internal INT IDs for database operations

### HTTP Handler (`adapters/http_handler.go`)
- ⚠️ **Potentially Non-Compliant**: Currently only has TODO implementations
- ❌ When implemented, handlers might use internal INT IDs if they follow service interface
- ❌ `CreateCommentAPI`, `GetCommentAPI`, `UpdateCommentAPI`, `DeleteCommentAPI` need to be reviewed for proper ID handling

## Security Analysis

### High-Risk Issues
- **ID Enumeration Vulnerability**: Using INT IDs in HTTP handlers exposes sequential internal identifiers that could be enumerated by attackers
- **Information Disclosure**: Exposing internal database IDs in any form provides information about the system's internal structure
- **Direct Object Reference**: Using internal IDs directly in HTTP endpoints creates Insecure Direct Object Reference (IDOR) vulnerabilities

### Medium-Risk Issues
- **Inconsistent ID Handling**: Mixing internal and public IDs in API contracts creates confusion and potential implementation errors
- **Lack of Input Validation**: When handlers are implemented, they may not properly validate UUID format

## Recommendations

### Immediate Actions
1. **Update Repository Interface**: Change all repository methods to use public_id (string) for external access
2. **Update Service Interface**: Change all service methods to use public_id (string) for external access
3. **Update Repository Implementation**: Modify repository methods to query by public_id instead of internal ID
4. **Review Foreign Keys**: Keep internal INT IDs for foreign key relationships but ensure they're not exposed externally

### Implementation Guidelines
- Repository `GetByID` should query: `SELECT ... FROM comments WHERE public_id = ?`
- Repository `Delete` should query: `DELETE FROM comments WHERE public_id = ?`
- Repository `Update` should update WHERE `public_id = ?`
- Service methods should accept public UUID strings and convert internally to INT as needed
- HTTP handlers should accept and return public UUIDs, never internal INT IDs

### Testing Considerations
- Write tests to verify that API endpoints return public UUIDs, not internal IDs
- Test that database queries use public_id for external lookups
- Verify that internal foreign keys still use INT for performance while maintaining security

## Compliance Status
**Current Status**: ❌ Non-Compliant
The Comment module does not follow the required INT primary key + UUID public ID pattern for external interfaces.

**Required Changes**: 
- Repository interface and implementation: Major changes needed
- Service interface and implementation: Major changes needed  
- HTTP handlers: Implementation needed with proper ID handling