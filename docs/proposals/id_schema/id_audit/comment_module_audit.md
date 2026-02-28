# Comment Module Schema Refactor Audit

## Overview
Audit of the comment module against SCHEMA_REFACTOR_STATUS.md requirements for INT primary keys + UUID public IDs.

## Findings

### ✅ Correctly Implemented
- **Migration (004_comment_create_comments.sql)**: Correct schema with `id INTEGER PRIMARY KEY AUTOINCREMENT`, `public_id TEXT UNIQUE NOT NULL`
- **Domain Entity (comment.go)**: Has `ID int` and `PublicID string` fields

### ❌ Issues Found

#### 1. Missing JSON Tags in Domain Entity
**File**: `internal/modules/comment/domain/comment.go`
**Issue**: No JSON tags on struct fields
**Required**:
```go
type Comment struct {
    ID        int       `json:"-"`                    // Internal INT
    PublicID  string    `json:"id"`                   // Public UUID
    PostID    int       `json:"-"`                    // Internal
    UserID    int       `json:"-"`                    // Internal
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    // Need to add for API responses:
    PublicPostID   string `json:"post_id,omitempty"`    // Public UUID of post
    PublicUserID   string `json:"user_id,omitempty"`    // Public UUID of author
}
```

#### 2. Service Interface Uses INT Instead of String for External IDs
**File**: `internal/modules/comment/ports/service.go`
**Issues**:
- `GetComment(ctx context.Context, commentID int)` → Should be `commentID string` (public_id)
- `UpdateComment(ctx context.Context, commentID int, ...)` → Should be `commentID string`
- `DeleteComment(ctx context.Context, commentID int)` → Should be `commentID string`
- `ListCommentsByPost(ctx context.Context, postID int)` → Should be `postID string` (public_id)

#### 3. Repository Interface Uses INT Instead of String
**File**: `internal/modules/comment/ports/repository.go`
**Issues**:
- `GetByID(ctx context.Context, commentID int)` → Should be `commentID string`
- `Update(ctx context.Context, comment *domain.Comment)` → OK (uses internal ID)
- `Delete(ctx context.Context, commentID int)` → Should be `commentID string`
- `ListByPostID(ctx context.Context, postID int)` → Should be `postID string`

#### 4. Application Service Implementation
**File**: `internal/modules/comment/application/service.go`
**Issues**: Mirrors the interface issues above. All methods using int IDs need to be updated to string public_ids.

#### 5. Repository Implementation (TODO)
**File**: `internal/modules/comment/adapters/sqlite_repository.go`
**Status**: Not implemented (all methods are TODO placeholders)
**Required Changes**:
- Generate UUID in `Create()` method
- `GetByID()` should query by `public_id = ?` and take `commentID string`
- All queries should use internal INT IDs for joins, but accept public_id for lookups

#### 6. HTTP Handler (TODO)
**File**: `internal/modules/comment/adapters/http_handler.go`
**Status**: Not implemented
**Required Security Checks**:
- URLs must use public_id strings, not expose internal INT IDs
- No INT IDs in JSON responses
- Convert public_id strings to internal IDs for service calls
- Authorization checks must use internal IDs

## Security Analysis

### High Risk Issues
1. **ID Exposure**: Once implemented, if handlers accidentally return internal `ID` fields in JSON (missing `json:"-"` tags), it would expose sequential IDs
2. **URL Parameter Injection**: If handlers accept INT IDs in URLs instead of UUIDs, attackers could enumerate resources by guessing sequential numbers
3. **Authorization Bypass**: If service methods take int IDs but don't validate ownership properly, users could access others' comments

### Medium Risk Issues
1. **Information Disclosure**: Missing JSON tags could leak internal database structure
2. **Performance**: Using INT IDs in URLs would be less secure than UUIDs

## Recommendations

### Immediate Fixes
1. Add JSON tags to domain entity
2. Update all interfaces to use `string` for external IDs, `int` for internal
3. Implement repository with UUID generation and proper queries
4. Implement handlers with public_id validation

### Implementation Pattern
```go
// Service method
func (s *Service) GetComment(ctx context.Context, commentID string) (*domain.Comment, error) {
    return s.commentRepo.GetByID(ctx, commentID)
}

// Repository method
func (r *SQLiteCommentRepository) GetByID(ctx context.Context, commentID string) (*domain.Comment, error) {
    query := `SELECT id, public_id, post_id, user_id, content, created_at, updated_at 
              FROM comments WHERE public_id = ?`
    // Scan into struct with both IDs
}

// Handler pattern
func (h *HTTPHandler) GetCommentAPI(w http.ResponseWriter, r *http.Request) {
    commentID := getFromURL(r, "id") // public_id string
    comment, err := h.commentService.GetComment(r.Context(), commentID)
    // Response includes public_id, hides internal ID
}
```

## Test Recommendations

### Unit Tests
1. **Domain Validation**: Test Comment.Validate() with various inputs
2. **Repository**: Test UUID generation in Create()
3. **Repository**: Test GetByID with public_id string returns correct entity
4. **Service**: Test GetComment with invalid public_id returns error

### Integration Tests
1. **API Endpoints**: Test GET /comments/{public_id} returns comment with public_id in JSON
2. **Security**: Test accessing non-existent public_id returns 404
3. **Security**: Test URL enumeration resistance (UUIDs vs sequential IDs)
4. **Authorization**: Test users cannot access others' comments

### Handler Tests
1. **URL Parsing**: Ensure handlers extract public_id correctly
2. **JSON Response**: Verify internal IDs are not in responses
3. **Error Handling**: Invalid UUID format returns proper error</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/comment_module_audit.md