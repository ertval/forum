# Platform Package ID Schema Compliance Analysis

## Overview

This document analyzes the compliance of platform packages with the new ID schema refactoring that separates internal INT primary keys from external UUID public IDs. The refactoring aims to improve database performance using INT primary keys while maintaining security by exposing only UUIDs to the external API and UI layers.

## Schema Refactoring Requirements

The new schema pattern follows this design:

```
HTTP Request (public_id string "750e8400-...")
    ↓
Middleware extracts session.UserID (int)
    ↓
Handler converts string → int for service calls
    ↓
Service uses int for business logic
    ↓
Repository:
  - Generates UUID for public_id on Create
  - Uses INT for joins/foreign keys
  - Queries by public_id for external lookups
    ↓
JSON Response exposes public_id, hides internal ID
```

## Findings Summary

### ✅ Compliant Platform Packages

The following platform packages correctly follow the refactoring requirements and don't handle ID exposure directly:
- `internal/platform/config` - Configuration management
- `internal/platform/database` - Database connection and migration tools
- `internal/platform/errors` - Error handling utilities
- `internal/platform/health` - Health check functionality
- `internal/platform/httpserver` - HTTP server and middleware framework
- `internal/platform/logger` - Structured logging
- `internal/platform/validator` - Input validation

### ⚠️ Critical Non-Compliance Issues Found

#### 1. HTTP Handlers (auth module)

**Issue**: The RegisterAPI method in `internal/modules/auth/adapters/http_handler.go` returns internal INT IDs as strings in JSON responses:

```go
// RegisterAPI - Lines 124-137
resp := struct {
    ID       string `json:"id"`      // ❌ Should use PublicID instead of internal ID
    UserID   string `json:"user_id"` // ❌ Should use PublicID instead of internal ID
    Email    string `json:"email"`
    Username string `json:"username"`
    Token    string `json:"token"`
}{
    ID:       userIDStr,  // ❌ Exposes internal INT ID as string
    UserID:   userIDStr,  // ❌ Exposes internal INT ID as string
    Email:    req.Email,
    Username: req.Username,
    Token:    session.Token,
}
```

**Impact**: Direct exposure of internal sequential IDs in API responses.

#### 2. HTTP Handlers (post module)

**Issue**: The `buildCurrentUser` function in `internal/modules/post/adapters/http_handler.go` uses internal ID as string:

```go
// buildCurrentUser - Line 100
return map[string]interface{}{
    "ID":           strconv.Itoa(userID), // ❌ Exposes internal INT ID to templates
    "Username":     username,
    "Email":        email,
    "PostCount":    postCount,
    "CommentCount": commentCount,
}
```

**Impact**: Internal IDs are passed to templates where they can be rendered in HTML.

#### 3. Template Files

**Critical Issues Found**:
- `templates/post_detail.html`: Uses `{{.Post.ID}}` in URLs, JavaScript attributes, and forms
- `templates/post_edit.html`: Uses `{{.Post.ID}}` in URLs and forms
- `templates/home.html`: Uses `{{.Post.ID}}` in URLs
- `templates/board.html`: Uses `{{.Post.ID}}` in URLs

**Examples of Non-Compliance**:
```html
<!-- templates/post_detail.html -->
<a href="/posts/{{.Post.ID}}/edit" class="btn btn-secondary">Edit</a>  <!-- ❌ Exposes internal ID -->
<button class="btn-delete-post" data-post-id="{{.Post.ID}}">Delete</button>  <!-- ❌ Exposes internal ID -->
```

**Impact**: Direct exposure of internal IDs in URLs and JavaScript, allowing enumeration and potential security vulnerabilities.

#### 4. JavaScript Files

**Issue**: JavaScript files receive and use internal IDs from templates:

```javascript
// static/js/post-detail.js
const postId = e.target.getAttribute('data-post-id'); // ❌ Gets internal ID from template
const response = await fetch(`/posts/${postId}`, {  // ❌ Uses internal ID in API call
```

**Impact**: Client-side JavaScript operates with internal IDs instead of UUIDs.

#### 5. Auth Middleware

**Issue**: Authentication middleware converts internal IDs to strings:

```go
// internal/modules/auth/adapters/middleware.go - Lines 40, 63
ctx := context.WithValue(r.Context(), UserIDKey, fmt.Sprintf("%d", session.UserID))
```

**Impact**: Internal IDs are passed through the request context to other handlers.

## Security Analysis

### Vulnerabilities

1. **Direct Object Reference (AOR)**: Exposing internal sequential IDs allows users to enumerate objects by incrementing IDs.

2. **Information Disclosure**: Sequential internal IDs reveal system usage patterns and total number of records.

3. **Business Logic Exposure**: Internal IDs provide insights into creation order and business metrics.

### Risk Assessment

- **High Risk**: Direct exposure of internal IDs in URLs and templates
- **Medium Risk**: Internal IDs in API responses and JavaScript
- **Low Risk**: Internal IDs in server-side context passing (though still non-compliant)

## Recommendations

### 1. Immediate Fixes Required

1. **Update API Response Format**:
   - In `auth/adapters/http_handler.go`, modify RegisterAPI to use user PublicID:
   ```go
   resp := struct {
       ID       string `json:"id"`      // Use user.PublicID instead of int ID
       UserID   string `json:"user_id"` // Use user.PublicID instead of int ID
       Email    string `json:"email"`
       Username string `json:"username"`
       Token    string `json:"token"`
   }{
       ID:       user.PublicID,   // ✅ Use UUID
       UserID:   user.PublicID,   // ✅ Use UUID
       Email:    req.Email,
       Username: req.Username,
       Token:    session.Token,
   }
   ```

2. **Update Template Data**:
   - Modify `buildCurrentUser` to pass PublicID instead of internal ID
   - Ensure all domain entities use the correct JSON tags:
     - `ID int` with `json:"-"` (internal only)
     - `PublicID string` with `json:"id"` (for external API)

3. **Update Templates**:
   - Replace `{{.Post.ID}}` with `{{.Post.PublicID}}` in all templates
   - Update all URLs and JavaScript attributes to use UUIDs

4. **Update JavaScript**:
   - Update JavaScript to use UUIDs passed from templates instead of internal IDs

### 2. Implementation Strategy

1. **Domain Layer**: Ensure all entity structs have proper JSON tags:
   ```go
   type Post struct {
       ID       int    `json:"-"`          // Internal INT ID
       PublicID string `json:"id"`         // External UUID (exposed in JSON)
       // ... other fields
   }
   ```

2. **Service Layer**: Update service methods to return entities with properly populated PublicID fields

3. **Repository Layer**: Verify repositories are generating UUIDs for PublicID fields during creation

4. **Handler Layer**: Update handlers to pass UUIDs to templates instead of internal IDs

5. **Template Layer**: Update all templates to use `PublicID` fields instead of `ID` fields

### 3. Testing Strategy

1. **Add compliance tests** to verify all API responses use UUIDs in "id" field
2. **Template security scanning** to detect internal ID usage
3. **Integration tests** to verify URLs use UUIDs
4. **Security tests** to detect any exposed internal sequential IDs

## Follow-up Actions

1. Review all other modules (comment, reaction, moderation, notification) for similar issues
2. Implement comprehensive test coverage for ID schema compliance
3. Add automated scanning for ID exposure in CI/CD pipeline
4. Update documentation to reflect the correct ID usage patterns