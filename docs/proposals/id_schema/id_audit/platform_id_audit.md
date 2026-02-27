# Platform ID Handling Security Audit

## Overview

This audit examines the platform packages and all modules for compliance with the schema refactor rules outlined in `docs/SCHEMA_REFACTOR_STATUS.md`. The refactor mandates that:

- **Public IDs**: Must be UUID strings for external exposure
- **Internal IDs**: Must be integers used only internally and in database queries
- **Security**: Internal sequential IDs must never be exposed in URLs, responses, or templates

## Audit Scope

- **Platform packages**: `internal/platform/*` (config, database, errors, health, httpserver, logger, validator)
- **All modules**: `internal/modules/*` (auth, post, user, comment, reaction, etc.)
- **Templates**: `templates/*.html` for ID exposure in HTML
- **Static assets**: `static/*` for any ID usage in JS/CSS

## Findings

### 1. Platform Packages Compliance ✅

**Status**: COMPLIANT - No ID handling in platform packages

**Details**:
- `config/`: Only handles configuration values, no entity IDs
- `database/`: Migration handling uses version integers (not entity IDs)
- `errors/`: Error types, no ID exposure
- `health/`: Health checks, no ID exposure
- `httpserver/`: Generic middleware (auth/session middleware not yet implemented)
- `logger/`: Logging infrastructure, no ID exposure
- `validator/`: Input validation, no ID exposure

**Recommendation**: No changes needed for platform packages.

### 2. Template ID Exposure ❌

**Status**: NON-COMPLIANT - Internal integer IDs exposed in HTML templates

**Critical Issues**:

#### Post Templates
- `templates/post_detail.html`:
  - `data-post-id="{{.Post.ID}}"` (line 25, 29) - exposes internal int ID
  - `href="/posts/{{.Post.ID}}/edit"` (line 42) - exposes internal int ID in URL
  - `href="/posts/{{.Post.ID}}"` (line 44) - exposes internal int ID in URL

- `templates/home.html`:
  - `href="/posts/{{.ID}}"` (line 8) - exposes internal int ID in URL

- `templates/board.html`:
  - `href="/posts/{{.ID}}"` (line 10) - exposes internal int ID in URL

- `templates/post_edit.html`:
  - `data-post-id="{{.Post.ID}}"` (line 5) - exposes internal int ID
  - `href="/posts/{{.Post.ID}}"` (line 29) - exposes internal int ID in URL

#### User Templates
- `templates/base.html`:
  - `href="/board?user={{.User.ID}}"` (line 109) - exposes internal int ID in query param

**Security Impact**:
- **Enumeration attacks**: Sequential integer IDs allow attackers to guess valid post/user IDs
- **Information disclosure**: Internal database structure exposed
- **Privacy violation**: User IDs can be correlated across sessions

**Required Fix**:
```html
<!-- BEFORE -->
<a href="/posts/{{.ID}}">{{.Title}}</a>
<data-post-id="{{.Post.ID}}">

<!-- AFTER -->
<a href="/posts/{{.PublicID}}">{{.Title}}</a>
<data-post-id="{{.Post.PublicID}}">
```

### 3. Handler Template Data Construction ❌

**Status**: NON-COMPLIANT - Incorrect ID exposure in template data

**Issues**:

#### Post Handler (`internal/modules/post/adapters/http_handler.go`)
- `buildCurrentUser()` function (line 65):
  - Returns `"ID": strconv.Itoa(userID)` - exposes internal int as string
  - Should return user's `PublicID`

**Code Issue**:
```go
// CURRENT (WRONG)
return map[string]interface{}{
    "ID": strconv.Itoa(userID),  // Internal int exposed
    // ...
}

// FIXED
user, _ := h.userService.GetByID(ctx, userID)
return map[string]interface{}{
    "ID": user.PublicID,  // Public UUID exposed
    // ...
}
```

### 4. API Response ID Exposure ❌

**Status**: NON-COMPLIANT - JSON APIs return stringified internal integers instead of UUIDs

**Issues**:

#### Auth Handler (`internal/modules/auth/adapters/http_handler.go`)
- `RegisterAPI()` (line 135):
  - Returns `ID: userIDStr, UserID: userIDStr` where `userIDStr = strconv.Itoa(userID)`
  - Should return user's `PublicID`

- `LoginAPI()`: Similar issue (returns stringified int instead of PublicID)

**Security Impact**:
- API consumers see internal sequential IDs
- Breaks abstraction between internal/external ID spaces

**Required Fix**:
```go
// Service layer needs to return PublicID instead of int
// OR handler fetches user and returns PublicID
user, _ := h.userService.GetByID(ctx, userID)
resp := struct {
    ID       string `json:"id"`
    UserID   string `json:"user_id"`
    // ...
}{
    ID:     user.PublicID,
    UserID: user.PublicID,
    // ...
}
```

### 5. Service Interface Design ❌

**Status**: NON-COMPLIANT - Service ports return internal integers for external operations

**Issues**:

#### Auth Service (`internal/modules/auth/ports/service.go`)
- `Register()` returns `(userID int, session *domain.Session, error)`
- Should return `(userPublicID string, session *domain.Session, error)`

#### Post Service (`internal/modules/post/ports/service.go`)
- `CreatePost()` takes `userID int` (correct for internal use)
- But external callers need to convert string PublicID to int

**Design Flaw**:
The service interfaces mix internal/external concerns. Services should:
- Accept PublicIDs for external operations
- Use internal ints for business logic
- Return PublicIDs for API responses

### 6. Domain Entity JSON Tags ✅

**Status**: MOSTLY COMPLIANT

**Good**:
- Post: `ID int `json:"-"`, PublicID string `json:"id"`
- User: `ID int`, `PublicID string` (no JSON tags, correctly hidden)
- Session: `ID int`, `PublicID string` (no JSON tags, correctly hidden)

**Issue**: User and Session lack explicit `json:"-"` tags on ID fields, but since they're not tagged for JSON, they're hidden by default.

### 7. Repository Layer ✅

**Status**: COMPLIANT - Correctly uses internal ints for DB operations

**Verified**:
- All repositories use `int` for internal IDs
- Generate UUIDs for `public_id` fields
- Use internal IDs for foreign keys and joins

### 8. Database Schema ✅

**Status**: COMPLIANT - Correct INT primary keys with UUID public_ids

**Verified**:
- All tables have `id INTEGER PRIMARY KEY AUTOINCREMENT`
- All tables have `public_id TEXT UNIQUE NOT NULL`
- Foreign keys use internal INT IDs
- Indexes on `public_id` fields

## Security Analysis

### Threat Model

1. **ID Enumeration**: Attackers can guess valid resources by incrementing IDs
2. **Information Disclosure**: Internal DB structure exposed
3. **Correlation Attacks**: User IDs linkable across different contexts
4. **Privacy Violation**: Sequential IDs reveal registration order

### Attack Vectors

1. **URL Guessing**: `/posts/1`, `/posts/2`, `/posts/3`... reveals all posts
2. **User Enumeration**: `/board?user=1`, `/board?user=2`... reveals all users
3. **Data Attribute Scraping**: JS can extract `data-post-id` values
4. **API Response Analysis**: JSON responses leak internal IDs

### Risk Assessment

- **Likelihood**: HIGH - IDs exposed in multiple locations
- **Impact**: MEDIUM-HIGH - Privacy violation, enumeration possible
- **Exploitability**: EASY - No authentication required for many endpoints

### Compliance with SCHEMA_REFACTOR_STATUS.md

- ✅ Database schema: INT + UUID correctly implemented
- ✅ Repository layer: Uses ints internally, generates UUIDs
- ❌ Handler layer: Exposes internal ints in responses/templates
- ❌ Template layer: Uses internal ints in URLs/data attributes
- ❌ API responses: Return stringified ints instead of UUIDs

## Recommendations

### Immediate Actions (Security Critical)

1. **Fix Template ID Exposure**
   - Replace `{{.ID}}` with `{{.PublicID}}` in all templates
   - Replace `{{.Post.ID}}` with `{{.Post.PublicID}}`
   - Replace `{{.User.ID}}` with `{{.User.PublicID}}`

2. **Fix Handler Template Data**
   - Modify `buildCurrentUser()` to return `user.PublicID`
   - Ensure all template data uses PublicIDs

3. **Fix API Responses**
   - Auth handlers: Return `user.PublicID` instead of `strconv.Itoa(userID)`
   - Post handlers: Ensure JSON responses use PublicIDs

### Architectural Improvements

1. **Service Interface Updates**
   - Change `Register()` to return `userPublicID string`
   - Add `GetUserPublicID(ctx, userID int) (string, error)` methods

2. **Handler Pattern**
   - Implement consistent PublicID resolution
   - Add middleware to convert PublicIDs to internal IDs for service calls

3. **Template Data Structure**
   - Ensure all template models expose PublicIDs, not internal IDs

## Test Suite for ID Compliance

```go
// internal/platform/tests/id_compliance_test.go
package tests

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestIDExposureCompliance(t *testing.T) {
    // Test template rendering doesn't expose internal IDs
    t.Run("Templates_No_Internal_IDs", func(t *testing.T) {
        // Mock post with internal ID
        post := &domain.Post{
            ID: 123,
            PublicID: "550e8400-e29b-41d4-a716-446655440001",
            Title: "Test Post",
        }
        
        // Render template
        // Assert no "123" appears in output
        // Assert "550e8400-e29b-41d4-a716-446655440001" appears
    })
    
    t.Run("API_Responses_Use_PublicIDs", func(t *testing.T) {
        // Test register API returns PublicID, not int
        req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(`{
            "email": "test@example.com",
            "username": "testuser", 
            "password": "password123"
        }`))
        
        w := httptest.NewRecorder()
        // Call handler
        
        // Assert response contains UUID, not sequential number
        body := w.Body.String()
        assert.Contains(t, body, `"id":`)
        assert.NotContains(t, body, `"id":"1"`) // Not sequential
        // Validate UUID format
    })
    
    t.Run("URLs_Use_PublicIDs", func(t *testing.T) {
        // Test generated URLs contain UUIDs
        // Parse HTML, extract hrefs, validate UUID format
    })
    
    t.Run("Data_Attributes_Use_PublicIDs", func(t *testing.T) {
        // Test data-post-id contains UUID
        // Parse HTML, check data attributes
    })
}

// Helper function to validate UUID format
func isValidUUID(s string) bool {
    // UUID v4 regex validation
    return regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(s)
}
```

## Implementation Checklist

- [ ] Update all templates to use `.PublicID`
- [ ] Fix `buildCurrentUser()` to return PublicID
- [ ] Update auth API handlers to return PublicIDs
- [ ] Add service methods to resolve PublicIDs
- [ ] Update post handlers for consistency
- [ ] Add comprehensive test suite
- [ ] Verify no internal IDs in HTML/JSON responses
- [ ] Test enumeration attack prevention

## Conclusion

The platform packages are compliant, but the module implementations have critical security issues with ID exposure. Internal sequential integers are exposed in templates, URLs, and API responses, violating the security principles of the schema refactor. Immediate fixes are required to prevent enumeration attacks and information disclosure.</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/platform_id_audit.md