# User Module - ID Handling Analysis Findings

## Overview
The User module's ID handling implementation has been reviewed against the schema refactor guidelines that require using internal INT primary keys with public UUID identifiers for external exposure.

## Current Implementation Analysis

### Domain Layer (`domain/user.go`)
- ✅ **Correctly implemented**: Domain entity has both internal INT ID and public UUID fields
- ✅ `ID int` - Internal unique identifier (INT PRIMARY KEY)
- ✅ `PublicID string` - Public UUID identifier (exposed in API)
- ✅ Proper JSON tags configuration to hide internal ID and expose PublicID as "id"

### Ports Layer (`ports/repository.go` and `ports/service.go`)
- ❌ **Critical Issue**: Repository interface uses INT IDs instead of public UUIDs for external access
  - `GetByID(ctx context.Context, userID int)` should be `GetByID(ctx context.Context, userID string)` to use public_id
  - `Update(ctx context.Context, user *domain.User)` should operate with public_id
  - `Delete(ctx context.Context, userID int)` should be `Delete(ctx context.Context, userID string)`
  - `GetUserStats(ctx context.Context, userID int)` should be `GetUserStats(ctx context.Context, userID string)`

- ❌ **Critical Issue**: Service interface uses INT IDs instead of public UUIDs
  - `GetByID(ctx context.Context, userID int)` should be `GetByID(ctx context.Context, userID string)`
  - `GetProfile(ctx context.Context, userID int)` should be `GetProfile(ctx context.Context, userID string)`
  - `UpdateRole(ctx context.Context, userID int, ...)` should be `UpdateRole(ctx context.Context, userID string, ...)`
  - `DeactivateUser(ctx context.Context, userID int)` should be `DeactivateUser(ctx context.Context, userID string)`
  - `ActivateUser(ctx context.Context, userID int)` should be `ActivateUser(ctx context.Context, userID string)`
  - `GetUserStats(ctx context.Context, userID int)` should be `GetUserStats(ctx context.Context, userID string)`
  - `ListUsers` methods may need to consider internal vs. public ID usage

### Application Layer (`application/service.go`)
- ❌ **Propagates Interface Issues**: Follows the incorrect interface design
- ❌ Uses internal INT IDs in service operations rather than public UUIDs

### Adapter Layer (`adapters/sqlite_repository.go`)
- ✅ **Correctly implemented**: Repository properly handles both internal ID and public UUID
- ✅ `Create` method generates public UUID and stores internal INT ID
- ✅ Proper scanning of both internal ID and public UUID from database
- ❌ **Issue**: Repository methods use internal INT IDs instead of public UUIDs for external access
  - Should query by public_id for external operations
  - Internal usage for foreign key relationships is appropriate

### HTTP Handler (`adapters/http_handler.go`)
- ⚠️ **Potentially Non-Compliant**: Currently only has TODO implementations
- ❌ When implemented, handlers might use internal INT IDs if they follow service interface
- ❌ `GetUserAPI`, `ListUsersAPI`, `UpdateRoleAPI`, `DeactivateUserAPI` need to be reviewed for proper ID handling

## Security Analysis

### High-Risk Issues
- **ID Enumeration Vulnerability**: Using INT IDs in HTTP handlers exposes sequential internal identifiers that could be enumerated by attackers
- **User Information Disclosure**: Exposing internal database IDs in user-related endpoints provides sensitive information about the system
- **Direct Object Reference**: Using internal IDs directly in HTTP endpoints creates Insecure Direct Object Reference (IDOR) vulnerabilities, especially dangerous for user operations

### Medium-Risk Issues
- **Inconsistent ID Handling**: Mixing internal and public IDs in API contracts creates confusion and potential implementation errors
- **Lack of Input Validation**: When handlers are implemented, they may not properly validate UUID format
- **User Privacy**: Sequential IDs could reveal information about user registration patterns

## Recommendations

### Immediate Actions
1. **Update Repository Interface**: Change all repository methods to use public_id (string) for external access
2. **Update Service Interface**: Change all service methods to use public_id (string) for external access
3. **Update Repository Implementation**: Modify repository methods to query by public_id instead of internal ID for external operations
4. **Update Service Implementation**: Adapt service methods to accept public UUIDs and convert internally to INT as needed

### Implementation Guidelines
- Repository `GetByID` should query: `SELECT ... FROM users WHERE public_id = ?`
- Repository `Delete` should delete WHERE `public_id = ?`
- Repository `Update` should update WHERE `public_id = ?`
- Service methods should accept public UUID strings and convert internally to INT as needed
- Internal foreign key operations can continue using INT IDs for performance
- HTTP handlers should accept and return public UUIDs, never internal INT IDs

### Testing Considerations
- Write tests to verify that API endpoints return public UUIDs, not internal IDs
- Test that database queries use public_id for external lookups
- Verify that internal foreign keys still use INT for performance while maintaining security
- Ensure proper authentication/authorization checks use the converted internal IDs

## Compliance Status
**Current Status**: ❌ Non-Compliant
The User module does not follow the required INT primary key + UUID public ID pattern for external interfaces.

**Required Changes**: 
- Repository interface and implementation: Major changes needed
- Service interface and implementation: Major changes needed  
- HTTP handlers: Implementation needed with proper ID handling