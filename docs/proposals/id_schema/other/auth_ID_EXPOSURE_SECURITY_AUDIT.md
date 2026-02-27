# Security Audit: ID Exposure Analysis

## Executive Summary

This audit examines the forum codebase for compliance with the SCHEMA_REFACTOR_STATUS.md requirements regarding ID exposure. The refactor mandates that public APIs and user interfaces expose UUID-based public IDs while keeping internal integer IDs hidden for security and performance reasons.

**Critical Finding**: Multiple components are currently exposing internal integer IDs in public responses, URLs, and HTML attributes, violating the security-by-obscurity principle and potentially enabling user enumeration attacks.

## Detailed Findings

### 1. Authentication Module Issues

#### Auth HTTP Handlers (`internal/modules/auth/adapters/http_handler.go`)

**RegisterAPI** (lines ~120-140):
- **Issue**: Returns `user_id` as `strconv.Itoa(userID)` (integer converted to string)
- **Security Risk**: Exposes internal sequential user IDs
- **Code**:
```go
resp := struct {
    ID       string `json:"id"`
    UserID   string `json:"user_id"`
    // ...
}{
    ID:       userIDStr,  // Internal ID exposed
    UserID:   userIDStr,  // Internal ID exposed
    // ...
}
```

**LoginAPI** (lines ~170-190):
- **Issue**: Returns `user_id` as integer directly
- **Security Risk**: Direct exposure of internal user ID
- **Code**:
```go
resp := struct {
    UserID int    `json:"user_id"`  // Internal ID exposed
    // ...
}{
    UserID: session.UserID,  // Internal ID
    // ...
}
```

**Root Cause**: Auth service `Register` returns `userID int` instead of `publicID string`. Handler has no access to user's public ID without additional database query.

### 2. Post Module Issues

#### Post HTTP Handlers (`internal/modules/post/adapters/http_handler.go`)

**buildCurrentUser** (lines ~60-80):
- **Issue**: Returns `"ID": strconv.Itoa(userID)` for template data
- **Security Risk**: Exposes internal user ID in template context
- **Impact**: Used in sidebar user card and filter links

**BoardPage** (lines ~320-340):
- **Issue**: Creates `previewPost["ID"] = post.ID` (integer)
- **Security Risk**: Passes internal post IDs to templates
- **Impact**: Used in post listing URLs

**renderPostDetail** (lines ~1020-1040):
- **Issue**: Passes full `post` object with `ID int` to template
- **Security Risk**: Template accesses `{{.Post.ID}}` in URLs and data attributes

#### Template Usage Analysis

**Critical URLs exposing internal IDs**:
- `/posts/{{.ID}}` - Post detail pages
- `/posts/{{.ID}}/edit` - Post edit pages
- `data-post-id="{{.ID}}"` - JavaScript data attributes
- `/board?user={{.User.ID}}` - User-specific post filtering

**Files affected**:
- `templates/base.html`: User post links
- `templates/board.html`: Post listing links
- `templates/home.html`: Post preview links
- `templates/post_detail.html`: Edit/delete/comment actions
- `templates/post_edit.html`: Form data and cancel links

### 3. User Module Analysis

**Status**: ✅ Compliant
- No HTTP handlers implemented yet (marked as TODO in roadmap)
- Repository correctly generates UUID public_ids
- Domain entities properly structured with separate ID/PublicID fields
- No exposure points currently exist

### 4. Filter System Issues

**FilterParams.UserID** (string type):
- **Issue**: Sometimes populated with `strconv.Itoa(userID)` (string representation of internal ID)
- **Inconsistency**: Mix of public UUIDs and internal integer strings
- **Impact**: Filtering logic may fail when expecting public IDs

## Security Analysis

### Threat Vectors

1. **User Enumeration**: Sequential integer IDs allow attackers to discover valid user accounts by iterating through IDs
2. **Information Leakage**: Internal ID patterns may reveal system information (e.g., registration order, total users)
3. **Predictable URLs**: `/users/123` vs `/users/550e8400-e29b-41d4-a716-446655440001` - integers are guessable
4. **API Abuse**: Exposed IDs can be used for unauthorized data access or brute-force attacks

### Compliance with SCHEMA_REFACTOR_STATUS.md

**✅ Completed**:
- Database schema uses INT primary keys + UUID public_ids
- Domain entities have both ID types with proper JSON tags
- Repositories generate UUIDs and handle conversions

**❌ Not Compliant**:
- HTTP responses expose internal IDs
- Templates use internal IDs in public URLs
- Service interfaces return internal IDs for operations that should use public IDs

### Risk Assessment

| Component | Risk Level | Impact | Likelihood |
|-----------|------------|--------|------------|
| Auth API responses | High | User enumeration, data leakage | High |
| Post URLs | High | Content enumeration, unauthorized access | High |
| Template data attributes | Medium | JavaScript-based attacks | Medium |
| Filter parameters | Low | Inconsistent filtering | Low |

## Recommended Fixes

### Immediate Actions (Security Critical)

1. **Update Auth Service Interface**:
   ```go
   // Change Register signature
   Register(ctx context.Context, email, username, password string) (publicID string, session *domain.Session, err error)
   ```

2. **Update Auth Handlers**:
   - RegisterAPI: Return `user.PublicID` instead of `strconv.Itoa(userID)`
   - LoginAPI: Fetch user by `session.UserID` and return `user.PublicID`

3. **Update Post Templates**:
   - Change `{{.ID}}` to `{{.PublicID}}` in all templates
   - Update data attributes: `data-post-id="{{.PublicID}}"`

4. **Update Post Handlers**:
   - Modify `buildCurrentUser` to return `user.PublicID`
   - Update `previewPost` to use `post.PublicID`

### Long-term Improvements

1. **Service Layer Refactor**:
   - Create `GetUserByPublicID(string)` method
   - Update post creation to accept `userPublicID string` instead of `userID int`

2. **URL Structure Standardization**:
   - Ensure all public URLs use UUIDs: `/posts/{uuid}`, `/users/{uuid}`

3. **Template Data Structure**:
   - Always pass `PublicID` fields to templates
   - Remove internal `ID` fields from template contexts

## Test Suite Recommendations

### Unit Tests

```go
func TestAuthHandler_RegisterAPI_ExposesPublicID(t *testing.T) {
    // Test that RegisterAPI response contains UUID, not integer
    resp := registerUser(t, "test@example.com", "testuser", "password")
    
    // Assert ID is valid UUID format
    assert.Regexp(t, uuidRegex, resp.ID)
    assert.Regexp(t, uuidRegex, resp.UserID)
    
    // Assert ID is not a simple integer string
    _, err := strconv.Atoi(resp.ID)
    assert.Error(t, err, "ID should not be convertible to integer")
}

func TestPostHandler_BoardPage_UsesPublicIDs(t *testing.T) {
    // Test that rendered HTML contains UUIDs in URLs
    html := renderBoardPage(t)
    
    // Assert no integer IDs in post links
    assert.NotRegexp(t, `href="/posts/\d+"`, html)
    assert.Regexp(t, `href="/posts/[a-f0-9-]+`, html)
}
```

### Integration Tests

```go
func TestPublicAPI_NoInternalIDExposure(t *testing.T) {
    // Register user
    resp := registerUser(t, "test@example.com", "testuser", "password")
    userID := resp.ID
    
    // Create post
    postResp := createPost(t, userID, "Test Post", "Content")
    
    // Assert all IDs are UUIDs
    assertValidUUID(t, postResp.ID)
    assertValidUUID(t, postResp.UserID)
    
    // Test post URL is accessible
    url := fmt.Sprintf("/posts/%s", postResp.ID)
    http.Get(t, url) // Should work
    
    // Test integer URL fails
    intURL := "/posts/123"
    resp, err := http.Get(intURL)
    assert.Equal(t, 404, resp.StatusCode)
}
```

### Security Tests

```go
func TestIDEnumerationPrevention(t *testing.T) {
    // Create known user
    resp1 := registerUser(t, "user1@example.com", "user1", "pass")
    
    // Attempt to access sequential IDs
    for i := 1; i <= 100; i++ {
        url := fmt.Sprintf("/users/%d", i)
        resp, _ := http.Get(url)
        // Should return 404 or 401, not user data
        assert.NotEqual(t, 200, resp.StatusCode)
    }
    
    // Valid UUID should work
    url := fmt.Sprintf("/users/%s", resp1.ID)
    resp, _ := http.Get(url)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Implementation Plan

### Phase 1: Critical Fixes (Week 1)
1. Update auth service and handlers
2. Update post templates to use PublicID
3. Update post handlers for template data

### Phase 2: Testing (Week 2)
1. Implement test suite
2. Run security tests
3. Validate no ID exposure

### Phase 3: Refinement (Week 3)
1. Update filter system consistency
2. Add API versioning if needed
3. Documentation updates

## Conclusion

The current implementation exposes internal integer IDs in multiple public interfaces, creating significant security risks. Immediate remediation is required to prevent user enumeration and data leakage. The recommended fixes align with the established schema refactor goals and will improve the system's security posture.

**Priority**: Critical - Implement Phase 1 fixes immediately before production deployment.</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/ID_EXPOSURE_SECURITY_AUDIT.md