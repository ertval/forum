# Post Module Schema Refactor Compliance Analysis

## Executive Summary

The post module has been partially refactored to use INT primary keys internally with UUID public IDs, but several critical issues remain that violate the security and architectural principles outlined in SCHEMA_REFACTOR_STATUS.md. The main problems are:

1. **HTTP handlers expose internal INT IDs in URLs and responses**
2. **Templates display internal IDs instead of public UUIDs**
3. **User module follows the correct pattern but has unimplemented handlers**
4. **All unit tests have compilation errors due to type mismatches**
5. **Integration tests are likely broken**

## Detailed Findings

### 1. Post Module Domain Layer ✅ COMPLIANT

**Status**: Fully compliant with schema refactor.

**Domain Entities**:
- `Post`: ID (int, json:"-" hidden), PublicID (string, json:"id" exposed)
- `Category`: ID (int, json:"-" hidden), PublicID (string, json:"id" exposed)

**Validation**: All entities properly hide internal IDs in JSON responses.

### 2. Post Module Ports Layer ✅ COMPLIANT

**Service Interfaces**:
- `CreatePost(userID int, ...)` - Correctly uses int for internal operations
- `GetPost(postID string)` - Correctly uses string (public_id) for external lookups
- `PostFilter.UserID string` - Uses public_id for filtering (correct for URL params)

**Repository Interfaces**:
- `GetByID(postID string)` - Uses public_id for external queries
- `Create(*domain.Post)` - Repository generates UUID internally

### 3. Post Module Application Layer ✅ COMPLIANT

**Service Implementation**:
- `CreatePost` takes `userID int` from session, calls repository which generates UUID
- `GetPost` takes `postID string` (public_id), passes to repository
- `UpdatePost` takes `postID string` (public_id)
- All business logic uses internal INT IDs where appropriate

### 4. Post Module Repository Layer ✅ COMPLIANT

**SQLite Implementation**:
- `Create`: Generates UUID for public_id, uses LastInsertId() for internal ID
- `GetByID`: Queries `WHERE public_id = ?`, scans both ID and PublicID
- `Update`: Uses `WHERE id = ?` with internal ID
- `Delete`: Uses `WHERE public_id = ?`
- `List`: Filters by `u.public_id = ?` for user filtering

**Category Repository**: Follows identical pattern.

### 5. Post Module HTTP Handler Layer ❌ NON-COMPLIANT

**Critical Security Issues**:

#### buildCurrentUser Function
```go
// WRONG: Exposes internal user ID
return map[string]interface{}{
    "ID": strconv.Itoa(userID),  // userID is int from session
    // ...
}
```
**Issue**: Should return `user.PublicID` instead of `strconv.Itoa(userID)`

#### Template Data Exposure
- Templates receive `.User.ID` as internal INT ID
- URLs like `/board?user={{.User.ID}}` expose sequential IDs
- This violates the core security principle of not exposing internal IDs

#### Post URL Exposure
In templates:
```html
<a href="/posts/{{.ID}}">{{.Title}}</a>  <!-- .ID is internal int -->
```
**Issue**: Should use `{{.PublicID}}`

#### Ownership Checks
```go
if eq .User.ID .Post.UserID  <!-- .User.ID is string, .Post.UserID is int -->
```
**Issue**: Type mismatch - comparison will always fail

#### Missing Author Links
```html
<span class="author">by {{.AuthorUsername}}</span>
```
**Issue**: Should be `<a href="/board?user={{.UserPublicID}}">{{.AuthorUsername}}</a>`

### 6. Template Layer ❌ NON-COMPLIANT

**base.html**:
- `{{.User.ID}}` in user card links exposes internal ID
- Should use PublicID

**post_detail.html**:
- `{{.Post.ID}}` in edit/delete links exposes internal ID
- `data-post-id="{{.Post.ID}}"` exposes internal ID to JavaScript
- `eq .User.ID .Post.UserID` type mismatch

**post-card.html**:
- `{{.ID}}` in post links exposes internal ID
- Missing author links with UserPublicID

### 7. User Module Analysis ✅ MOSTLY COMPLIANT

**Domain Layer ✅**: User struct correctly hides ID, exposes PublicID

**Repository Layer ✅**: 
- Create generates UUID, uses LastInsertId for internal ID
- GetByID takes int, queries by internal id, returns both IDs

**Service Layer ✅**: GetByID takes int (correct for internal operations)

**HTTP Handler Layer ⚠️**: Not implemented (placeholders only)

### 8. Auth Module Analysis ✅ COMPLIANT

**Domain Layer ✅**: Session struct hides ID, exposes PublicID

**Repository Layer ✅**: Follows correct UUID generation pattern

**Service Layer ✅**: Register returns userID int (correct)

**HTTP Handler Layer ✅**: GetCurrentUser returns userID int (correct)

### 9. Test Files ❌ BROKEN

**All unit tests have compilation errors**:

#### Post Service Tests
```go
// WRONG
post := &domain.Post{
    ID: "post-1",        // Should be ID: 1, PublicID: "post-1"
    UserID: "user-1",    // Should be UserID: 1
}

// WRONG
userID: "user-1",       // Should be userID: 1
```

#### Auth Tests
```go
// WRONG
session := &domain.Session{
    ID: "session-1",     // Should be ID: 1, PublicID: "session-1"
}
```

#### Repository Tests
- Mock repositories need to handle both ID types
- GetByID calls need string public_id instead of int ID

## Security Analysis

### Information Disclosure Vulnerabilities

1. **Internal ID Exposure in URLs**
   - URLs like `/posts/123`, `/board?user=456` expose sequential internal IDs
   - Allows attackers to enumerate resources and users
   - Predictable patterns enable brute force attacks

2. **User Enumeration**
   - Exposed user IDs allow building user profiles
   - Sequential IDs reveal registration order and user count
   - Privacy violation for user anonymity

3. **API Inconsistency**
   - JSON responses hide internal IDs (correct)
   - But HTML responses and URLs expose them (incorrect)
   - Mixed security model creates confusion

### Attack Vectors

1. **Resource Enumeration**
   ```
   GET /posts/1    → 200 OK
   GET /posts/2    → 200 OK
   GET /posts/1000 → 404 Not Found
   ```
   Attacker can discover all post IDs and existence.

2. **User Discovery**
   ```
   GET /board?user=1   → Alice's posts
   GET /board?user=2   → Bob's posts
   ```
   Attacker can discover all users and their activity.

3. **Predictable Attacks**
   - Sequential IDs make it trivial to guess valid resources
   - No need for complex enumeration techniques

### Compliance Violations

According to SCHEMA_REFACTOR_STATUS.md:
- "Public UUID TEXT (public_id) - for external API exposure"
- "Internal INT PRIMARY KEY - for DB performance"
- "JSON Response exposes public_id, hides internal ID"

**Current State**: JSON hides IDs ✅, but HTML/URLs expose them ❌

## Recommended Fixes

### 1. Fix buildCurrentUser Function
```go
func (h *HTTPHandler) buildCurrentUser(ctx context.Context, userID int) map[string]interface{} {
    // ... existing code ...
    
    if user, err := h.userService.GetByID(ctx, userID); err == nil && user != nil {
        return map[string]interface{}{
            "ID": user.PublicID,  // Use PublicID instead of strconv.Itoa(userID)
            "Username": user.Username,
            "Email": user.Email,
            // ...
        }
    }
    
    // Fallback if user fetch fails
    return map[string]interface{}{
        "ID": "",  // Empty string instead of exposing internal ID
        // ...
    }
}
```

### 2. Fix Template Usage
**post-card.html**:
```html
<a href="/posts/{{.PublicID}}">{{.Title}}</a>
<span class="author">
    <a href="/board?user={{.UserPublicID}}">{{.AuthorUsername}}</a>
</span>
```

**post_detail.html**:
```html
<a href="/posts/{{.Post.PublicID}}/edit">{{.Title}}</a>
data-post-id="{{.Post.PublicID}}"
{{if eq .User.ID .Post.UserPublicID}}  <!-- Both strings now -->
```

### 3. Fix Test Files
**Pattern for post tests**:
```go
post := &domain.Post{
    ID: 1,
    PublicID: "test-post-uuid",
    UserID: 1,
    Title: "Test Post",
    // ...
}

userID := 1  // int instead of "user-1"
```

**Pattern for repository tests**:
```go
// GetByID now takes string
post, err := repo.GetByID(ctx, "test-post-uuid")
```

### 4. Add Integration Tests
```go
func TestPostURLsExposePublicIDs(t *testing.T) {
    // Test that /posts/{id} uses public_id, not internal ID
    // Test that user links use public_id
    // Test that JSON responses don't include internal IDs
}
```

## Implementation Priority

1. **HIGH**: Fix buildCurrentUser to return PublicID
2. **HIGH**: Update templates to use PublicID in URLs
3. **HIGH**: Fix ownership checks in templates
4. **MEDIUM**: Update all unit tests
5. **MEDIUM**: Add integration tests for ID exposure
6. **LOW**: Implement user profile handlers (currently placeholders)

## Testing Strategy

### Unit Tests
- Update all mock objects to return correct types
- Fix struct literals to use int IDs and string PublicIDs
- Update assertions to check both ID types appropriately

### Integration Tests
- Test HTTP endpoints return public_ids in URLs
- Test JSON responses don't include internal IDs
- Test HTML templates render public_ids
- Test user enumeration prevention

### Security Tests
- Attempt to access resources with internal IDs (should fail)
- Verify UUID format in all public exposures
- Test that internal IDs are never leaked in responses

## Conclusion

The post module architecture is sound, but the HTTP layer and templates critically violate the security principles by exposing internal sequential IDs. This creates information disclosure vulnerabilities that could allow user enumeration and resource discovery attacks. Immediate fixes to handlers and templates are required, followed by comprehensive test updates.</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/post_module_schema_analysis.md