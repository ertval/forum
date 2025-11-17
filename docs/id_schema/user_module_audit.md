# User Module Schema Refactor Audit

## Overview
Comprehensive audit of the user module against SCHEMA_REFACTOR_STATUS.md requirements.

## Findings

### ✅ Correctly Implemented
- **Migration (002_user_create_users.sql)**: Correct schema with `id INTEGER PRIMARY KEY AUTOINCREMENT`, `public_id TEXT UNIQUE NOT NULL`
- **Domain Entity (user.go)**: Has `ID int` and `PublicID string` fields
- **Repository Implementation**: Generates UUID in Create(), scans both IDs in queries

### ❌ Issues Found

#### 1. Missing JSON Tags in Domain Entities
**File**: `internal/modules/user/domain/user.go`
**Issue**: No JSON tags on User struct
**Required**:
```go
type User struct {
    ID           int       `json:"-"`                    // Internal
    PublicID     string    `json:"id"`                   // Public UUID
    Email        string    `json:"email"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"`                    // Never expose
    Role         Role      `json:"role"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    IsActive     bool      `json:"is_active"`
}
```

**UserProfile struct**:
```go
type UserProfile struct {
    UserID       int       `json:"-"`                    // Internal
    PublicUserID string    `json:"id"`                   // Public UUID
    Username     string    `json:"username"`
    Role         Role      `json:"role"`
    PostCount    int       `json:"post_count"`
    CommentCount int       `json:"comment_count"`
    CreatedAt    time.Time `json:"created_at"`
}
```

#### 2. Service Interface Issues
**File**: `internal/modules/user/ports/service.go`
**Critical Issues**:
- `GetByID(ctx context.Context, userID int)` → Should be `GetByPublicID(ctx context.Context, userID string)` for external API access
- `GetProfile(ctx context.Context, userID int)` → Should be `userID string`
- `UpdateRole(ctx context.Context, userID int, ...)` → Should be `userID string`
- `DeactivateUser(ctx context.Context, userID int)` → Should be `userID string`
- `ActivateUser(ctx context.Context, userID int)` → Should be `userID string`

**Note**: Keep `GetByID(int)` for internal use (auth service), but add `GetByPublicID(string)` for external API.

#### 3. Repository Interface Issues
**File**: `internal/modules/user/ports/repository.go`
**Issues**:
- Need to add `GetByPublicID(ctx context.Context, userID string) (*domain.User, error)`
- `GetByID` should remain for internal use
- `Update` and `Delete` use internal `user.ID int` → OK
- `GetUserStats` takes `userID int` → OK (internal)

#### 4. Repository Implementation Issues
**File**: `internal/modules/user/adapters/sqlite_repository.go`
**Issues**:
- Missing `GetByPublicID` implementation
- `GetUserStats` query uses `reaction_type` but migration has `type` column
- `GetUserStats` uses `author_id` for posts/comments → OK

#### 5. Application Service Issues
**File**: `internal/modules/user/application/service.go`
**Issues**: Mirrors interface issues, needs `GetByPublicID` method

#### 6. HTTP Handler (TODO)
**File**: `internal/modules/user/adapters/http_handler.go`
**Status**: Not implemented
**Required**: Use public_id in URLs, return public_id in JSON

## Security Analysis

### Critical Security Issues

#### 1. ID Enumeration Vulnerability
**Current Risk**: `GetByID(int)` allows enumeration of all users by guessing sequential IDs
**Impact**: Attackers can discover all user accounts, even private ones
**Fix**: Use `GetByPublicID(string)` for all external API access

#### 2. Information Disclosure
**Current Risk**: Missing `json:"-"` tags could expose internal IDs and password hashes
**Impact**: Leaks database structure and sensitive data
**Fix**: Add proper JSON tags

#### 3. Authorization Bypass
**Current Risk**: Service methods using int IDs may not validate user permissions properly
**Impact**: Users could access/modify other users' data
**Fix**: Implement proper authorization checks

#### 4. Profile Privacy
**Current Risk**: `GetProfile` with int ID could allow accessing any user's profile
**Impact**: Privacy violation, stalking
**Fix**: Ensure users can only access their own or public profiles

### Medium Risk Issues
1. **Role Elevation**: `UpdateRole` needs admin-only access validation
2. **Account Deactivation**: `DeactivateUser` needs proper authorization
3. **Stats Disclosure**: `GetUserStats` could leak activity information

## Recommendations

### Immediate Security Fixes
1. Add JSON tags to prevent information leakage
2. Implement `GetByPublicID` in repository and service
3. Update handlers to use public_id in URLs
4. Add authorization checks to all user operations

### Implementation Pattern
```go
// Add to repository interface
GetByPublicID(ctx context.Context, userID string) (*domain.User, error)

// Repository implementation
func (r *SQLiteUserRepository) GetByPublicID(ctx context.Context, userID string) (*domain.User, error) {
    query := `SELECT id, public_id, email, username, password_hash, role, created_at, updated_at, is_active
              FROM users WHERE public_id = ?`
    // Scan into User struct
}

// Service
func (s *Service) GetByPublicID(ctx context.Context, userID string) (*domain.User, error) {
    return s.userRepo.GetByPublicID(ctx, userID)
}

// Handler
func (h *HTTPHandler) GetUserAPI(w http.ResponseWriter, r *http.Request) {
    userID := getFromURL(r, "id") // public_id string
    user, err := h.userService.GetByPublicID(r.Context(), userID)
    // Check if current user can view this profile
    // Return JSON with public_id
}
```

### Authorization Pattern
```go
func (h *HTTPHandler) GetUserAPI(w http.ResponseWriter, r *http.Request) {
    requestedUserID := getFromURL(r, "id")
    currentUserID := getUserFromSession(r) // internal int
    
    user, err := h.userService.GetByPublicID(r.Context(), requestedUserID)
    if err != nil {
        // handle error
    }
    
    // Allow access to own profile or if admin
    if user.ID != currentUserID && !isAdmin(currentUser) {
        http.Error(w, "Forbidden", 403)
        return
    }
    
    // Return profile JSON
}
```

## Test Recommendations

### Unit Tests
1. **Repository**: Test GetByPublicID returns correct user
2. **Repository**: Test GetByPublicID with invalid UUID returns error
3. **Service**: Test GetByPublicID authorization
4. **Domain**: Test JSON marshaling hides internal fields

### Integration Tests
1. **API**: GET /users/{public_id} returns user profile
2. **Security**: Test users cannot access others' profiles
3. **Security**: Test invalid UUID returns 404
4. **Security**: Test ID enumeration prevention
5. **Authorization**: Test admin can access all profiles
6. **JSON**: Verify internal IDs not in responses

### Handler Tests
1. **URL Parsing**: Extract public_id from URL parameters
2. **Session**: Get current user from session
3. **Authorization**: Check permissions before access
4. **Response**: Ensure public_id in JSON, internal fields hidden</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/user_module_audit.md