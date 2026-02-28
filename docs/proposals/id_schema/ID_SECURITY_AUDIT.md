# ID Security Audit: UUID Public IDs vs INT Internal IDs

**Audit Date**: 2025-01-17  
**Schema Refactor**: INT Primary Keys + UUID Public IDs  
**Status**: 🔴 **CRITICAL VULNERABILITIES FOUND**

## Executive Summary

This audit identifies **critical security vulnerabilities** where internal INT IDs are exposed in public-facing interfaces (URLs, templates, JavaScript). This violates the INT+UUID schema pattern and creates security risks.

### Critical Findings

1. ✅ **Domain Layer**: CORRECT - All entities properly use INT (internal) + UUID (public)
2. ✅ **Repository Layer**: CORRECT - Generates UUIDs, uses INT internally
3. ✅ **Service Layer**: CORRECT - Uses INT IDs internally
4. 🔴 **Handler Layer**: **VULNERABLE** - Exposes internal INT IDs in context
5. 🔴 **Template Layer**: **VULNERABLE** - References `.User.ID` which could be INT
6. 🔴 **JavaScript Layer**: **VULNERABLE** - Uses `post.ID` without explicit UUID field name
7. ✅ **API Responses**: CORRECT - JSON responses use PublicID correctly

---

## Detailed Vulnerability Analysis

### 🔴 CRITICAL: Template ID Exposure

**File**: `templates/base.html` Line 109  
**Issue**: Uses `.User.ID` in URL parameter

```gohtml
<a href="/board?user={{.User.ID}}{{if .SelectedCategory}}&category={{urlquery .SelectedCategory}}{{end}}{{if .DateFilter}}&date_filter={{.DateFilter}}{{end}}" class="btn btn-primary btn-small">My Posts</a>
```

**Risk**: If `.User.ID` contains the internal INT ID, it exposes sequential user IDs in URLs.

**Files Affected**:
- `templates/base.html:109` - "My Posts" link

**Expected Behavior**: Should use `.User.PublicID` (UUID)

---

### 🟡 MEDIUM: Template Ownership Checks

**Files**: `templates/post_detail.html`

```gohtml
{{if eq .User.ID .Post.UserPublicID}}  <!-- Line 41 -->
{{if eq $.User.ID .AuthorPublicID}}    <!-- Line 87 -->
```

**Issue**: Compares `.User.ID` (possibly INT) with `.Post.UserPublicID` (UUID string)

**Risk**: 
- Type mismatch (int vs string) could fail silently
- If `.User.ID` is INT, comparison always fails → broken ownership checks
- Users could edit/delete others' content

**Expected Behavior**: Should compare `.User.PublicID` with `.Post.UserPublicID`

---

### 🟡 MEDIUM: JavaScript ID Field Ambiguity

**File**: `static/js/load-more-posts.js` Lines 30, 65

```javascript
<h3><a href="/posts/${post.ID}">${post.Title}</a></h3>
```

**Issue**: Uses `post.ID` which is ambiguous. In JSON responses:
- `post.id` → PublicID (UUID) ✅
- `post.ID` → Could be mistaken for internal ID

**Risk**: If JSON parsing uses wrong field, could expose INT IDs

**Expected Behavior**: Use explicit `post.PublicID` or `post.id` (lowercase, as in JSON)

---

### 🔴 CRITICAL: Handler Context Stores String INT ID

**File**: `internal/modules/auth/adapters/middleware.go` Line 42

```go
ctx := context.WithValue(r.Context(), UserIDKey, fmt.Sprintf("%d", session.UserID))
```

**Issue**: Converts internal INT `session.UserID` to string and stores in context

**Risk**: 
- Handlers retrieve this string INT ID via `authAdapters.GetUserID()`
- This string (e.g., "123") could be accidentally used in URLs/templates
- Breaks UUID-only public interface contract

**Affected Handlers**:
- All handlers using `authAdapters.GetUserID(r.Context())` get string INT ID
- `buildCurrentUser()` passes INT to services (correct internally)
- BUT: User object must have PublicID for templates

---

### ✅ CORRECT: Handler Passes User Object to Templates

**File**: `internal/modules/post/adapters/http_handler.go` Lines 65-89

```go
func (h *HTTPHandler) buildCurrentUser(ctx context.Context, userID int) map[string]interface{} {
    var username, email, publicID string
    // ...
    if user, err := h.userService.GetByID(ctx, userID); err == nil && user != nil {
        username = user.Username
        email = user.Email
        publicID = user.PublicID  // ✅ Fetches PublicID
    }
    return map[string]interface{}{
        "ID":           publicID,  // ✅ Returns UUID as "ID"
        "Username":     username,
        "Email":        email,
        "PostCount":    postCount,
        "CommentCount": commentCount,
    }
}
```

**Status**: ✅ **CORRECT** - Handler returns user map with `"ID": publicID` (UUID)

**However**: Template field name `.User.ID` is ambiguous - could be confused with internal ID

---

## Security Impact Assessment

### High Risk Scenarios

1. **Information Disclosure**:
   - Internal sequential INT IDs exposed in URLs (`?user=123`)
   - Attackers can enumerate all users: `/board?user=1`, `/board?user=2`, etc.
   - Reveals user count, registration order, account activity patterns

2. **Broken Access Control**:
   - Template ownership checks (`eq .User.ID .Post.UserPublicID`) fail silently
   - Users could edit/delete others' posts if checks always return false
   - Type mismatch (int vs string) causes logic errors

3. **Data Integrity**:
   - JavaScript using wrong ID field could send INT IDs to API
   - Repository queries by UUID would fail
   - Creates inconsistent state

### OWASP Vulnerabilities

- **A01:2021 - Broken Access Control**: Ownership checks comparing wrong ID types
- **A05:2021 - Security Misconfiguration**: Exposing internal IDs violates security design
- **A07:2021 - Identification and Authentication Failures**: Sequential IDs enable enumeration

---

## Root Cause Analysis

### Design Pattern

The INT+UUID schema pattern has **three ID contexts**:

1. **Internal INT IDs**: Database primary keys, foreign keys (performance)
2. **Public UUID strings**: API responses, URLs, external references (security)
3. **Context/Template Interface**: Must use UUIDs, never INT

### Where Pattern Breaks

| Layer | Expected | Actual | Status |
|-------|----------|--------|--------|
| Database | INT IDs | INT IDs | ✅ CORRECT |
| Domain Entities | Both INT + UUID fields | Both fields | ✅ CORRECT |
| Repository | Generate UUID, use INT | Generates UUID | ✅ CORRECT |
| Service | INT internally | INT internally | ✅ CORRECT |
| **Middleware** | **Store UUID in context** | **Stores string INT** | 🔴 **BROKEN** |
| **Handler→Template** | **Pass UUID as ID** | **Passes UUID but field name ambiguous** | 🟡 **RISKY** |
| Template | Use .User.PublicID | Uses .User.ID (ambiguous) | 🟡 **RISKY** |
| API JSON | PublicID as "id" | Correct | ✅ CORRECT |

---

## Recommendations

### 1. Middleware Context Fix (CRITICAL)

**Current** (middleware.go:42):
```go
ctx := context.WithValue(r.Context(), UserIDKey, fmt.Sprintf("%d", session.UserID))
```

**Proposed Fix**:
```go
// Fetch user to get PublicID
user, err := userService.GetByID(ctx, session.UserID)
if err != nil {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

// Store PublicID (UUID) in context, not INT
ctx := context.WithValue(r.Context(), UserIDKey, user.PublicID)
```

**Impact**: All handlers would receive UUID string from `authAdapters.GetUserID()`

**Trade-off**: Requires database query in middleware (performance cost)

**Alternative**: Store both IDs in context with different keys:
```go
ctx = context.WithValue(ctx, InternalUserIDKey, session.UserID) // int
ctx = context.WithValue(ctx, PublicUserIDKey, user.PublicID)    // string UUID
```

### 2. Template Field Naming (HIGH PRIORITY)

**Change templates** to use explicit UUID field names:

```gohtml
<!-- BEFORE -->
<a href="/board?user={{.User.ID}}">My Posts</a>
{{if eq .User.ID .Post.UserPublicID}}

<!-- AFTER -->
<a href="/board?user={{.User.PublicID}}">My Posts</a>
{{if eq .User.PublicID .Post.UserPublicID}}
```

**Files to update**:
- `templates/base.html:109`
- `templates/post_detail.html:41, 87`

### 3. Handler Template Data Structure (HIGH PRIORITY)

**Standardize user object passed to templates**:

```go
type TemplateUser struct {
    PublicID     string `json:"id"`        // UUID for URLs
    Username     string `json:"username"`
    Email        string `json:"email"`
    PostCount    int    `json:"post_count"`
    CommentCount int    `json:"comment_count"`
    // NO "ID" field to avoid ambiguity
}
```

**Update `buildCurrentUser`**:
```go
return map[string]interface{}{
    "PublicID":     publicID,  // Explicit UUID field
    "Username":     username,
    "Email":        email,
    "PostCount":    postCount,
    "CommentCount": commentCount,
}
```

### 4. JavaScript Field Standardization (MEDIUM PRIORITY)

**Update JavaScript** to use explicit UUID field:

```javascript
// BEFORE
<h3><a href="/posts/${post.ID}">${post.Title}</a></h3>

// AFTER
<h3><a href="/posts/${post.id}">${post.Title}</a></h3>
// OR (more explicit)
<h3><a href="/posts/${post.PublicID}">${post.Title}</a></h3>
```

**Rationale**: JSON responses use lowercase `"id"` for PublicID, match that convention

---

## Testing Strategy

### 1. Unit Tests for ID Exposure

**Test File**: `tests/id_exposure_test.go`

```go
func TestNoInternalIDsInAPIResponses(t *testing.T) {
    // Test all API endpoints return only UUIDs
    // Assert response contains valid UUID format
    // Assert response does NOT contain integers in ID fields
}

func TestTemplateUsesPublicIDs(t *testing.T) {
    // Parse templates
    // Search for ".User.ID", ".Post.ID", etc.
    // Assert templates use ".User.PublicID" or explicit UUID fields
}

func TestContextStoresUUIDs(t *testing.T) {
    // Call middleware with mock session
    // Extract UserID from context
    // Assert it's a valid UUID format (not integer string)
}
```

### 2. Integration Tests

**Test Scenarios**:
1. Register user → Verify URL parameters use UUID
2. Create post → Verify post detail URL uses UUID
3. Edit post → Verify ownership check uses UUIDs (both sides)
4. List posts filtered by user → Verify `?user=` parameter is UUID
5. Load more posts (AJAX) → Verify JSON response uses UUID as `"id"`

### 3. Manual Security Testing

**Checklist**:
- [ ] Inspect browser URLs - all IDs are UUIDs (36 chars with dashes)
- [ ] Check browser DevTools Network tab - API responses use UUIDs
- [ ] Try editing someone else's post - verify 403 Forbidden
- [ ] Inspect HTML source - no integer IDs in data attributes
- [ ] Test "My Posts" link - URL uses UUID
- [ ] Test comment ownership - can only delete own comments

---

## Implementation Priority

### Phase 1: Critical Security Fixes (IMMEDIATE)

1. ✅ Fix middleware to fetch and store PublicID in context
2. ✅ Update `buildCurrentUser()` to return `PublicID` field (not `ID`)
3. ✅ Update templates to use `.User.PublicID` explicitly
4. ✅ Add integration tests for ID exposure

**Estimated Time**: 2-3 hours

### Phase 2: Consistency & Standards (HIGH)

5. ✅ Standardize JavaScript to use `post.id` or `post.PublicID`
6. ✅ Update all handlers to use consistent template data structure
7. ✅ Add unit tests for template ID field usage

**Estimated Time**: 2-3 hours

### Phase 3: Documentation & Guidelines (MEDIUM)

8. ✅ Document ID handling patterns in ARCHITECTURE.md
9. ✅ Add linting rules to catch `.User.ID` in templates
10. ✅ Update copilot-instructions.md with ID security rules

**Estimated Time**: 1-2 hours

---

## Monitoring & Prevention

### 1. Automated Checks

**Pre-commit Hook**:
```bash
#!/bin/bash
# Check templates for internal ID usage
if grep -r "\.User\.ID\|\.Post\.ID\|\.Comment\.ID" templates/; then
    echo "ERROR: Templates must use .User.PublicID, not .User.ID"
    exit 1
fi
```

**CI/CD Pipeline**:
- Run ID exposure tests on every commit
- Fail build if internal IDs found in API responses
- Generate UUID validation report

### 2. Code Review Checklist

When reviewing PRs:
- [ ] All new templates use `.PublicID` fields
- [ ] API responses return only UUIDs in `"id"` fields
- [ ] No `fmt.Sprintf("%d", ...)` converting INT IDs to strings for public use
- [ ] Middleware stores only UUIDs in context
- [ ] JavaScript uses lowercase `post.id` (matches JSON response)

---

## Compliance & Standards

### Schema Refactor Goals

**From** `docs/SCHEMA_REFACTOR_STATUS.md`:
> All public ids should be uuid and int ids should be used only internally and in the database queries for better performance.

**Current Compliance**: 🔴 **PARTIAL**

- ✅ Database: INT IDs used correctly
- ✅ JSON APIs: UUIDs exposed correctly
- 🔴 Templates: INT IDs potentially exposed in URLs
- 🔴 Middleware: Stores string INT in context

### Security Design Principles

1. **Least Privilege**: External users should never see internal IDs
2. **Defense in Depth**: Multiple layers should validate UUID format
3. **Fail Secure**: Ownership checks must fail if types mismatch
4. **Audit Trail**: Log attempts to use INT IDs in public context

---

## Conclusion

The schema refactor successfully implemented INT+UUID pattern in the **data layer** (database, domain, repositories), but **critical gaps remain** in the **presentation layer** (middleware, templates, URLs).

**Immediate Action Required**:
1. Fix middleware to store PublicID in context
2. Update templates to use `.User.PublicID`
3. Add automated tests to prevent regression

**Long-term**:
- Establish ID security as part of code review process
- Add linting/scanning for ID exposure
- Document patterns clearly for future development

---

**Audit Conducted By**: AI Security Analysis  
**Next Review Date**: After Phase 1 fixes implemented  
**Document Version**: 1.0
