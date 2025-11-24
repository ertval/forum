# User Module ID Security Audit Report

## Overview

This report analyzes the user module's compliance with the schema refactor rules from `SCHEMA_REFACTOR_STATUS.md`. The refactor mandates:
- Public IDs: UUID strings exposed in APIs and URLs
- Internal IDs: INT used only internally and in database queries
- Handlers and templates must not expose or use INT IDs in responses/URLs

## Audit Findings

### ✅ Compliant Areas

1. **Domain Entity (`domain/user.go`)**:
   - `User` struct correctly has `ID int` (internal) and `PublicID string` (public)
   - No JSON tags expose the internal ID (`ID` has no json tag, so `json:"-"`)

2. **Ports (`ports/service.go`, `ports/repository.go`)**:
   - Service interface uses `userID int` for internal operations
   - Repository interface uses `userID int` for data access
   - Correct separation of concerns

3. **Application Service (`application/service.go`)**:
   - Uses `int` userID parameters throughout
   - No exposure of internal IDs

4. **SQLite Repository (`adapters/sqlite_repository.go`)**:
   - `Create()` generates UUID for `PublicID` and sets `ID` from `LastInsertId()`
   - All query methods use `int` IDs for database operations
   - `GetUserStats()` correctly uses `author_id` (int) in queries
   - No public exposure of internal IDs

### ❌ Non-Compliant Areas

1. **HTTP Handlers (`adapters/http_handler.go`)**:
   - **Status**: Not implemented (all methods are TODO placeholders)
   - **Risk**: When implemented, routes like `GET /users/{id}` must use `public_id` (UUID), not internal `id` (int)
   - **Finding**: Route comments show `GET /users/{id}` - this `{id}` should be the public UUID

2. **Unit Tests (`adapters/sqlite_repository_test.go`)**:
   - **Critical Issue**: Test schema missing `public_id` column
   - **Issue**: Tests set `ID: 1` manually, but `Create()` should generate it
   - **Risk**: Tests don't validate UUID generation or public ID exposure

3. **Test Schema Mismatch**:
   - Test creates table without `public_id` column
   - Actual migration includes `public_id TEXT UNIQUE NOT NULL`
   - Tests will fail against real database schema

## Security Analysis

### ID Exposure Risks

1. **Enumeration Attacks**:
   - If handlers expose internal INT IDs in URLs/responses, attackers can enumerate users by incrementing IDs
   - UUIDs prevent enumeration due to their randomness

2. **Information Disclosure**:
   - Sequential INT IDs reveal user registration order and total user count
   - UUIDs provide no such information

3. **URL Predictability**:
   - INT IDs in URLs are predictable (`/users/1`, `/users/2`, etc.)
   - UUIDs are not predictable, reducing targeted attacks

### Current Risk Assessment

- **Low Risk**: Handlers not implemented, so no exposure yet
- **Medium Risk**: Broken tests may lead to incorrect implementations
- **High Risk**: When handlers are implemented, developers might accidentally use internal IDs

### Authentication/Authorization Concerns

- User module handles sensitive operations (role updates, deactivation)
- Ensure all handlers validate user permissions before operations
- Role-based access control must be implemented in service layer

## Recommendations

### Immediate Actions

1. **Fix Test Schema**:
   - Update `sqlite_repository_test.go` to include `public_id` column
   - Modify tests to not pre-set `ID`, let repository generate it
   - Add assertions for `PublicID` generation

2. **Handler Implementation Guidelines**:
   - URLs must use `{public_id}` (UUID string) in path parameters
   - Parse `public_id` from URL, resolve to internal `userID` via repository
   - Never expose internal `ID` in JSON responses
   - JSON responses should include `PublicID` as `"id"`

3. **Add Repository Method**:
   - Add `GetByPublicID(ctx context.Context, publicID string) (*domain.User, error)` to repository interface
   - Implement in SQLite repository
   - Use this for URL-based lookups

### Code Examples

#### Handler Pattern (When Implemented)
```go
func (h *HTTPHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
    // Extract public_id from URL
    publicID := chi.URLParam(r, "public_id")
    
    // Resolve to internal ID
    user, err := h.userService.GetByPublicID(r.Context(), publicID)
    if err != nil {
        // handle error
        return
    }
    
    // Response exposes PublicID, not ID
    response := map[string]interface{}{
        "id": user.PublicID,  // UUID
        "username": user.Username,
        // ... other public fields
    }
    
    json.NewEncoder(w).Encode(response)
}
```

#### Repository Addition
```go
// In ports/repository.go
GetByPublicID(ctx context.Context, publicID string) (*domain.User, error)

// In adapters/sqlite_repository.go
func (r *SQLiteUserRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
    query := `SELECT id, public_id, email, username, password_hash, role, created_at, updated_at, is_active
              FROM users WHERE public_id = ?`
    // ... scan logic
}
```

## Test Suite for ID Security

### Unit Tests for ID Handling

```go
func TestSQLiteUserRepository_IDHandling(t *testing.T) {
    db := setupTestDB(t) // Include public_id in schema
    repo := NewSQLiteUserRepository(db)
    ctx := context.Background()

    // Test Create generates UUID and sets ID
    user := &domain.User{
        Email:        "test@example.com",
        Username:     "testuser",
        PasswordHash: "hash",
        Role:         domain.RoleUser,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
        IsActive:     true,
    }

    err := repo.Create(ctx, user)
    require.NoError(t, err)
    
    // Assert ID is set (from LastInsertId)
    assert.NotZero(t, user.ID)
    
    // Assert PublicID is UUID
    assert.NotEmpty(t, user.PublicID)
    _, err = uuid.Parse(user.PublicID)
    assert.NoError(t, err)
}

func TestSQLiteUserRepository_GetByPublicID(t *testing.T) {
    db := setupTestDB(t)
    repo := NewSQLiteUserRepository(db)
    ctx := context.Background()

    // Create user
    user := createTestUser(t, repo)

    // Get by public ID
    retrieved, err := repo.GetByPublicID(ctx, user.PublicID)
    require.NoError(t, err)
    assert.Equal(t, user.ID, retrieved.ID)
    assert.Equal(t, user.PublicID, retrieved.PublicID)
}

func TestUserService_NoIDExposure(t *testing.T) {
    mockRepo := &MockUserRepository{}
    service := NewService(mockRepo)

    // Mock returns user with both IDs
    mockRepo.getByIDFn = func(ctx context.Context, id int) (*domain.User, error) {
        return &domain.User{
            ID:       123,  // internal
            PublicID: "uuid-123",
            Username: "test",
        }, nil
    }

    user, err := service.GetByID(context.Background(), 123)
    require.NoError(t, err)
    
    // Service should return full user (for internal use)
    // But handlers must not expose ID in responses
    assert.Equal(t, 123, user.ID)        // internal use OK
    assert.Equal(t, "uuid-123", user.PublicID) // public use OK
}
```

### Integration Tests for Handler Security

```go
func TestHTTPHandler_UserProfile_NoIDExposure(t *testing.T) {
    // Setup test server with user handler
    app := setupTestApp(t)
    
    // Create test user
    user := createTestUser(t, app.userService)
    
    // Make request to /users/{public_id}
    req := httptest.NewRequest("GET", "/users/"+user.PublicID, nil)
    w := httptest.NewRecorder()
    
    // Call handler
    app.userHandler.GetUserProfile(w, req)
    
    // Parse response
    var response map[string]interface{}
    err := json.NewDecoder(w.Body).Decode(&response)
    require.NoError(t, err)
    
    // Assert response contains public_id as "id"
    assert.Equal(t, user.PublicID, response["id"])
    
    // Assert internal ID is NOT present
    _, hasInternalID := response["internal_id"]
    assert.False(t, hasInternalID, "Response should not contain internal ID")
}

func TestHTTPHandler_UserProfile_InvalidPublicID(t *testing.T) {
    app := setupTestApp(t)
    
    // Request with invalid UUID
    req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
    w := httptest.NewRecorder()
    
    app.userHandler.GetUserProfile(w, req)
    
    // Should return 404 or appropriate error
    assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_UserProfile_IntIDNotAccepted(t *testing.T) {
    app := setupTestApp(t)
    
    // Request with internal INT ID (should not work)
    req := httptest.NewRequest("GET", "/users/123", nil) // 123 is int, not UUID
    w := httptest.NewRecorder()
    
    app.userHandler.GetUserProfile(w, req)
    
    // Should return 404 (no user with public_id="123")
    assert.Equal(t, 404, w.Code)
}
```

### Security Test Helpers

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    
    // Create table with correct schema (including public_id)
    _, err = db.Exec(`
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            public_id TEXT UNIQUE NOT NULL,
            email TEXT UNIQUE NOT NULL,
            username TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            role TEXT NOT NULL DEFAULT 'user',
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            is_active INTEGER NOT NULL DEFAULT 1
        )
    `)
    require.NoError(t, err)
    
    return db
}

func createTestUser(t *testing.T, repo ports.UserRepository) *domain.User {
    user := &domain.User{
        Email:        "test@example.com",
        Username:     "testuser", 
        PasswordHash: "hash",
        Role:         domain.RoleUser,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
        IsActive:     true,
    }
    
    err := repo.Create(context.Background(), user)
    require.NoError(t, err)
    
    return user
}
```

## Implementation Priority

1. **High**: Fix test schemas and assertions
2. **High**: Implement `GetByPublicID` repository method
3. **Medium**: Implement handlers with UUID-based URLs
4. **Medium**: Add comprehensive test suite
5. **Low**: Add rate limiting for user profile endpoints

## Conclusion

The user module's core architecture is compliant with the ID refactor rules. The main issues are:
- Incomplete handler implementations
- Broken unit tests due to schema mismatches

With the recommended fixes and test suite, the module will be secure against ID enumeration and information disclosure attacks.</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/docs/user_module_id_security_audit.md