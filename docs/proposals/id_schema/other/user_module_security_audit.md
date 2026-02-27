# User Module ID Handling - Security Audit Report

## Overview

This report presents a comprehensive analysis of the user module's compliance with the schema refactor requirements. The schema refactor mandates using INT primary keys internally for performance while exposing UUID public IDs externally for security.

## Findings Summary

### ✅ Proper Implementation
- **Domain Layer**: The `User` struct correctly implements both internal int ID and public UUID:
  ```go
  type User struct {
      ID       int    // Internal unique identifier (INT PRIMARY KEY)
      PublicID string // Public UUID identifier (exposed in API)
      // ... other fields
  }
  ```

- **Repository Layer**: The SQLite repository properly generates UUIDs on creation and stores both IDs:
  - `Create()` method generates public_id UUID and sets internal ID from database
  - Query methods return both internal ID (int) and public ID (string)
  - Database schema uses INT PRIMARY KEY with TEXT public_id field

### ⚠️ Critical Issues Identified

#### 1. Auth Handler Returns Internal IDs as Strings (Security Risk)
**Location**: `internal/modules/auth/adapters/http_handler.go` - `RegisterAPI` method

**Issue**: 
```go
// Current problematic code:
userIDStr := strconv.Itoa(userID)
resp := struct {
    ID       string `json:"id"`      // Internal int ID converted to string!
    UserID   string `json:"user_id"` // Internal int ID converted to string!
    Email    string `json:"email"`
    Username string `json:"username"`
    Token    string `json:"token"`
}{
    ID:       userIDStr,  // This is internal ID
    UserID:   userIDStr,  // This is internal ID  
    Email:    req.Email,
    Username: req.Username,
    Token:    session.Token,
}
```

**Security Impact**: Exposes sequential internal IDs to clients, enabling enumeration attacks and IDOR vulnerabilities.

**Expected Fix**:
```go
resp := struct {
    ID       string `json:"id"`      // Should be user.PublicID
    UserID   string `json:"user_id"` // Should be user.PublicID
    Email    string `json:"email"`
    Username string `json:"username"`
    Token    string `json:"token"`
}{
    ID:       user.PublicID,    // Public UUID
    UserID:   user.PublicID,    // Public UUID
    Email:    req.Email,
    Username: req.Username,
    Token:    session.Token,
}
```

#### 2. Template URLs Use Internal IDs
**Locations**: 
- `templates/home.html` - `<a href="/posts/{{.ID}}">`
- `templates/post_detail.html` - `href="/posts/{{.Post.ID}}/edit"`
- `templates/post_detail.html` - `data-post-id="{{.Post.ID}}"`
- `templates/post_edit.html` - `href="/posts/{{.Post.ID}}"`
- `templates/post_edit.html` - `data-post-id="{{.Post.ID}}"`

**Security Impact**: Exposes internal database structure and enables ID enumeration.

**Expected Fix**: Use public UUIDs in all URLs and data attributes.

#### 3. JavaScript Code Uses Internal IDs in API Calls
**Location**: `static/js/post-detail.js`

**Issue**:
```js
// Using internal IDs in URLs
const response = await fetch(`/posts/${postId}/comments`, {...});
const response = await fetch(`/posts/${postId}`, {method: 'DELETE'});
const response = await fetch(`/comments/${commentId}`, {method: 'DELETE'});
```

**Security Impact**: Same as above - exposes internal IDs to client-side code.

#### 4. Test File Inconsistency
**Location**: `internal/modules/user/adapters/sqlite_repository_test.go`

**Issue**: Test calls non-existent `Get` method instead of `GetByID`:
```go
result, err := repo.Get(ctx, user.ID)  // Should be GetByID
```

## Security Recommendations

### High Priority (Security Critical)
1. **Fix Auth Handler Response**: Modify `RegisterAPI`, `LoginAPI`, and `GetSessionAPI` to return public UUIDs instead of internal IDs
2. **Update Template URLs**: Replace all internal ID references with public UUIDs in templates
3. **Update JavaScript**: Modify all API calls to use public UUIDs instead of internal IDs
4. **Fix Test Files**: Correct the repository test to call `GetByID` instead of `Get`

### Medium Priority 
1. **Complete User Module Implementation**: Implement the placeholder HTTP handlers in `http_handler.go`
2. **Add JSON Tags to Domain Structs**: Add proper JSON tags to ensure internal IDs are not exposed:
   ```go
   type User struct {
       ID       int    `json:"-"`                    // Internal, not in JSON
       PublicID string `json:"id"`                   // Exposed as "id" in JSON
       // ... other fields
   }
   ```
3. **Update UserProfile Structure**: Add PublicID field to UserProfile for consistency

## Implementation Guidelines

### ID Flow Pattern (As per schema refactor):
```
HTTP Request (public_id string "750e8400-...")
    ↓
Middleware extracts session.UserID (int) for secure operations
    ↓
Handler converts string → int for service calls (when needed for internal operations)
    ↓
Service layer uses int internally for performance
    ↓
Repository:
  - Generates UUID for public_id on Create operations
  - Uses INT for joins/foreign key relationships
  - Queries by public_id for external lookups
    ↓
Domain Entity has both:
  - ID int (internal, tagged `json:"-"`) 
  - PublicID string (external, tagged `json:"id"`)
```

### JSON Response Pattern:
```json
{
  "id": "750e8400-e29b-41d4-a716-446655440001",  // PublicID
  "username": "john_doe",
  "email": "john@example.com"
}
```

## Compliance Status
- **Domain Layer**: ✅ 100% Compliant
- **Repository Layer**: ✅ 100% Compliant  
- **Service Layer**: ⚠️ 90% Compliant (placeholder implementations exist)
- **HTTP Handlers**: ❌ 0% Compliant (auth handler has critical security issue, user handlers mostly empty)
- **Templates**: ❌ 0% Compliant (uses internal IDs)
- **Frontend JS**: ❌ 0% Compliant (uses internal IDs)

## Conclusion

The user module's core infrastructure (domain, repository) properly implements the schema refactor. However, critical security issues exist in the HTTP layer that expose internal IDs to clients. These issues must be addressed immediately to maintain the security benefits of the UUID public ID system.

The main problem centers around the auth module's response formatting, which defeats the purpose of having public UUIDs by returning internal sequential IDs as strings. This creates security vulnerabilities that must be fixed before the system is deployed.