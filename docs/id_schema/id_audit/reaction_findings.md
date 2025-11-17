# Reaction Module - ID Handling Analysis Findings

## Overview
The Reaction module's ID handling implementation has been reviewed against the schema refactor guidelines that require using internal INT primary keys with public UUID identifiers for external exposure.

## Current Implementation Analysis

### Domain Layer (`domain/reaction.go`)
- ✅ **Correctly implemented**: Domain entity has both internal INT ID and public UUID fields
- ✅ `ID int` - Internal unique identifier (INT PRIMARY KEY)
- ✅ `PublicID string` - Public UUID identifier (exposed in API)
- ✅ `UserID int` and `TargetID int` - Internal foreign keys
- ✅ `TargetType string` - Specifies whether target is "post" or "comment"
- ✅ JSON tags properly configured to hide internal IDs

### Ports Layer (`ports/repository.go` and `ports/service.go`)
- ❌ **Critical Issue**: Repository interface uses INT IDs instead of public UUIDs for external access
  - `GetByUserAndTarget(ctx context.Context, userID, targetID int, targetType string)` - Uses internal INT IDs
  - `Count(ctx context.Context, targetID int, ...)` - Uses internal INT for target ID
  - `GetByTarget(ctx context.Context, targetID int, ...)` - Uses internal INT for target ID
  - `Delete(ctx context.Context, userID, targetID int, ...)` - Uses internal INT IDs

- ❌ **Critical Issue**: Service interface uses INT IDs instead of public UUIDs for external references
  - `GetReactions(ctx context.Context, targetID int, ...)` - Uses internal INT for target ID
  - `CountReactions(ctx context.Context, targetID int, ...)` - Uses internal INT for target ID
  - However, this may be acceptable internally since reactions are tied to posts/comments that have their own public IDs

### Application Layer (`application/service.go`)
- ❌ **Propagates Interface Issues**: Follows the incorrect interface design
- ❌ Uses internal INT IDs in service operations rather than public UUIDs where appropriate

### Adapter Layer (`adapters/sqlite_repository.go`)
- ❌ **Critical Issue**: Repository methods operate on internal INT IDs rather than public UUIDs for external access
- ❌ `GetByTarget`, `GetByUserAndTarget`, `Count`, `Delete` methods use internal INT IDs
- ❌ Missing implementation for proper ID handling

### HTTP Handler (`adapters/http_handler.go`)
- ⚠️ **Potentially Non-Compliant**: Currently only has TODO implementations
- ❌ When implemented, handlers might use internal INT IDs if they follow service interface
- ❌ `AddReactionAPI`, `RemoveReactionAPI`, `GetReactionsAPI`, `CountReactionsAPI` need to be reviewed for proper ID handling

## Security Analysis

### High-Risk Issues
- **ID Enumeration Vulnerability**: Using INT IDs for target references (posts/comments) could expose internal structure
- **Information Disclosure**: Exposing internal database IDs in any form provides information about the system's internal structure
- **Direct Object Reference**: Using internal IDs directly in HTTP endpoints creates Insecure Direct Object Reference (IDOR) vulnerabilities
- **Improper Target Identification**: Reactions reference posts/comments by internal INT IDs rather than public UUIDs

### Medium-Risk Issues
- **Inconsistent ID Handling**: Target references use internal IDs while reaction itself has public UUID
- **Lack of Input Validation**: When handlers are implemented, they may not properly validate UUID format for target IDs

## Recommendations

### Immediate Actions
1. **Consider Target Reference Approach**: Determine if targetID should be public UUID or if internal INT is acceptable for performance
2. **Update Service Interface**: If external targets should use UUIDs, update service methods accordingly
3. **Update Repository Interface**: If needed, modify repository methods to handle public UUIDs for targets
4. **Review Foreign Key Strategy**: Ensure foreign keys use internal INTs while maintaining security and performance

### Implementation Guidelines
- For target references (posts/comments), there may be two approaches:
  - Internal approach: Use internal INT IDs for performance (current approach)
  - External approach: Use public UUIDs for targets, requiring conversion
- If using external approach:
  - HTTP handlers should accept target public UUIDs and convert to internal IDs
  - Repository methods should map public UUIDs to internal IDs before queries
- The Reaction's own ID should always use public UUID in responses

### Testing Considerations
- Write tests to verify that API endpoints return Reaction public UUIDs, not internal IDs
- If using public UUIDs for targets, test that conversion from UUID to internal ID works properly
- Test that database queries use appropriate ID types for internal performance
- Verify that internal foreign keys still use INT for performance while maintaining security

## Compliance Status
**Current Status**: ⚠️ Partially Compliant
The Reaction module partially follows the required INT primary key + UUID public ID pattern. The Reaction entity itself is properly implemented with both internal ID and public UUID, but the interface for accessing reactions by target still uses internal INT IDs.

**Required Changes**: 
- Consider if target references should use public UUIDs instead of internal IDs
- Repository interface and implementation: Potential changes needed if target references use UUIDs
- HTTP handlers: Implementation needed with proper ID handling