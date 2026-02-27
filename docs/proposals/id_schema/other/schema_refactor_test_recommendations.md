# Schema Refactor Testing Recommendations

## Overview
Comprehensive test suite to validate the INT + UUID schema implementation and security controls across all modules.

## Test Categories

### 1. Domain Entity Tests

#### JSON Serialization Tests
```go
func TestUserJSONSerialization(t *testing.T) {
    user := &domain.User{
        ID: 1,
        PublicID: "550e8400-e29b-41d4-a716-446655440001",
        Email: "test@example.com",
        Username: "testuser",
        PasswordHash: "hash",
        Role: domain.RoleUser,
    }

    jsonData, err := json.Marshal(user)
    assert.NoError(t, err)

    // Should NOT contain internal ID
    assert.NotContains(t, string(jsonData), `"id":1`)
    // Should contain public ID
    assert.Contains(t, string(jsonData), `"id":"550e8400-e29b-41d4-a716-446655440001"`)
    // Should NOT contain password hash
    assert.NotContains(t, string(jsonData), `"password_hash"`)
}

func TestCommentJSONSerialization(t *testing.T) {
    comment := &domain.Comment{
        ID: 1,
        PublicID: "750e8400-e29b-41d4-a716-446655440001",
        PostID: 1,
        PublicPostID: "750e8400-e29b-41d4-a716-446655440002",
        UserID: 1,
        PublicUserID: "550e8400-e29b-41d4-a716-446655440001",
        Content: "Test comment",
    }

    jsonData, err := json.Marshal(comment)
    assert.NoError(t, err)

    var result map[string]interface{}
    json.Unmarshal(jsonData, &result)

    // Public IDs should be present
    assert.Equal(t, "750e8400-e29b-41d4-a716-446655440001", result["id"])
    assert.Equal(t, "750e8400-e29b-41d4-a716-446655440002", result["post_id"])
    assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", result["user_id"])

    // Internal IDs should NOT be present
    assert.NotContains(t, result, "ID")
    assert.NotContains(t, result, "PostID")
    assert.NotContains(t, result, "UserID")
}
```

#### Domain Validation Tests
```go
func TestCommentValidation(t *testing.T) {
    tests := []struct {
        name    string
        comment *domain.Comment
        wantErr bool
    }{
        {"valid comment", &domain.Comment{Content: "Valid content"}, false},
        {"empty content", &domain.Comment{Content: ""}, true},
        {"too long content", &domain.Comment{Content: strings.Repeat("a", 1001)}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.comment.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Repository Tests

#### UUID Generation Tests
```go
func TestUserRepositoryCreateGeneratesUUID(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := adapters.NewSQLiteUserRepository(db)

    user := &domain.User{
        Email: "test@example.com",
        Username: "testuser",
        PasswordHash: "hash",
        Role: domain.RoleUser,
    }

    err := repo.Create(context.Background(), user)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.PublicID)
    assert.NotEqual(t, 0, user.ID)

    // Verify UUID format
    _, err = uuid.Parse(user.PublicID)
    assert.NoError(t, err)
}

func TestRepositoryGetByPublicID(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := adapters.NewSQLiteUserRepository(db)

    // Create user
    user := createTestUser(t, repo)
    publicID := user.PublicID

    // Retrieve by public ID
    retrieved, err := repo.GetByPublicID(context.Background(), publicID)
    assert.NoError(t, err)
    assert.Equal(t, user.ID, retrieved.ID)
    assert.Equal(t, user.PublicID, retrieved.PublicID)
    assert.Equal(t, user.Email, retrieved.Email)
}

func TestRepositoryGetByPublicIDNotFound(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := adapters.NewSQLiteUserRepository(db)

    _, err := repo.GetByPublicID(context.Background(), "non-existent-uuid")
    assert.Error(t, err)
    assert.Equal(t, sql.ErrNoRows, err)
}
```

#### Internal ID Usage Tests
```go
func TestRepositoryUsesInternalIDsForJoins(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    userRepo := adapters.NewSQLiteUserRepository(db)
    postRepo := adapters.NewSQLitePostRepository(db)

    // Create user
    user := createTestUser(t, userRepo)

    // Create post
    post := &domain.Post{
        UserID: user.ID, // Internal ID
        Title: "Test Post",
        Content: "Content",
    }
    err := postRepo.Create(context.Background(), post)
    assert.NoError(t, err)

    // Verify post is linked to user via internal ID
    retrieved, err := postRepo.GetByID(context.Background(), post.PublicID)
    assert.NoError(t, err)
    assert.Equal(t, user.ID, retrieved.UserID)
    assert.Equal(t, user.PublicID, retrieved.UserPublicID)
}
```

### 3. Service Tests

#### Public ID Interface Tests
```go
func TestUserServiceGetByPublicID(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := adapters.NewSQLiteUserRepository(db)
    service := application.NewService(repo)

    // Create user
    user := createTestUser(t, repo)

    // Get by public ID
    retrieved, err := service.GetByPublicID(context.Background(), user.PublicID)
    assert.NoError(t, err)
    assert.Equal(t, user.ID, retrieved.ID)
    assert.Equal(t, user.PublicID, retrieved.PublicID)
}

func TestCommentServiceCreateWithPublicIDs(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    commentRepo := adapters.NewSQLiteCommentRepository(db)
    service := application.NewService(commentRepo)

    // Create comment with internal IDs
    comment, err := service.CreateComment(context.Background(), 1, 1, "Test comment")
    assert.NoError(t, err)
    assert.NotEmpty(t, comment.PublicID)

    // Verify UUID format
    _, err = uuid.Parse(comment.PublicID)
    assert.NoError(t, err)
}
```

### 4. Handler Tests

#### URL Parameter Validation Tests
```go
func TestGetUserAPIInvalidUUID(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()

    app := setupTestApp(t, db)
    server := setupTestServer(t, app)

    // Make request with invalid UUID
    req := httptest.NewRequest("GET", "/api/users/invalid-uuid", nil)
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    assert.Equal(t, 400, w.Code)
    assert.Contains(t, w.Body.String(), "invalid UUID")
}

func TestGetUserAPINotFound(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    app := setupTestApp(t, db)
    server := setupTestServer(t, app)

    validUUID := "550e8400-e29b-41d4-a716-446655440000"
    req := httptest.NewRequest("GET", "/api/users/"+validUUID, nil)
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    assert.Equal(t, 404, w.Code)
}
```

#### Authorization Tests
```go
func TestGetUserAPIUnauthorized(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    app := setupTestApp(t, db)
    userRepo := adapters.NewSQLiteUserRepository(db)

    // Create two users
    user1 := createTestUser(t, userRepo)
    user2 := createTestUserWithEmail(t, userRepo, "other@example.com", "otheruser")

    // Login as user1
    session := loginAsUser(t, app, user1)

    // Try to access user2's profile
    req := httptest.NewRequest("GET", "/api/users/"+user2.PublicID, nil)
    req.AddCookie(&http.Cookie{Name: "session_id", Value: session.PublicID})
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    // Should be forbidden (403) or not found (404) depending on implementation
    assert.True(t, w.Code == 403 || w.Code == 404)
}

func TestGetUserAPIOwnProfile(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    app := setupTestApp(t, db)
    userRepo := adapters.NewSQLiteUserRepository(db)

    user := createTestUser(t, userRepo)
    session := loginAsUser(t, app, user)

    req := httptest.NewRequest("GET", "/api/users/"+user.PublicID, nil)
    req.AddCookie(&http.Cookie{Name: "session_id", Value: session.PublicID})
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    // Should contain public ID, not internal ID
    assert.Equal(t, user.PublicID, response["id"])
    assert.NotContains(t, response, "ID")
}
```

### 5. Security Integration Tests

#### ID Enumeration Prevention
```go
func TestIDEnumerationPrevention(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    app := setupTestApp(t, db)
    userRepo := adapters.NewSQLiteUserRepository(db)

    // Create a user
    user := createTestUser(t, userRepo)

    // Try to access using internal ID (should fail)
    req := httptest.NewRequest("GET", fmt.Sprintf("/api/users/%d", user.ID), nil)
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    // Should return 404 or 400 (invalid UUID)
    assert.True(t, w.Code == 404 || w.Code == 400)
}

func TestSequentialIDGuessing(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    app := setupTestApp(t, db)

    // Try various sequential numbers
    for i := 1; i <= 100; i++ {
        req := httptest.NewRequest("GET", fmt.Sprintf("/api/users/%d", i), nil)
        w := httptest.NewRecorder()

        app.ServeHTTP(w, req)

        // Should not return user data
        assert.True(t, w.Code == 404 || w.Code == 400,
            "Sequential ID %d should not return user data", i)
    }
}
```

#### JSON Response Security
```go
func TestJSONResponseDoesNotLeakInternalIDs(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    app := setupTestApp(t, db)
    userRepo := adapters.NewSQLiteUserRepository(db)
    postRepo := adapters.NewSQLitePostRepository(db)

    user := createTestUser(t, userRepo)
    session := loginAsUser(t, app, user)

    // Create post
    post := &domain.Post{
        UserID: user.ID,
        Title: "Test Post",
        Content: "Content",
    }
    err := postRepo.Create(context.Background(), post)
    assert.NoError(t, err)

    // Get post via API
    req := httptest.NewRequest("GET", "/api/posts/"+post.PublicID, nil)
    req.AddCookie(&http.Cookie{Name: "session_id", Value: session.PublicID})
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    // Check that internal IDs are not present
    assert.NotContains(t, response, "ID")
    assert.NotContains(t, response, "UserID")
    assert.NotContains(t, response, "PostID")

    // Check that public IDs are present
    assert.Contains(t, response, "id")
    assert.Contains(t, response, "user_id")
}
```

### 6. Performance Tests

#### Query Performance Comparison
```go
func BenchmarkGetByID(b *testing.B) {
    db := setupTestDB(b)
    defer db.Close()

    repo := adapters.NewSQLiteUserRepository(db)
    user := createTestUser(b, repo)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := repo.GetByID(context.Background(), user.ID)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGetByPublicID(b *testing.B) {
    db := setupTestDB(b)
    defer db.Close()

    repo := adapters.NewSQLiteUserRepository(db)
    user := createTestUser(b, repo)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := repo.GetByPublicID(context.Background(), user.PublicID)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Test Organization

### Directory Structure
```
tests/
├── unit/
│   ├── domain/
│   │   ├── user_json_test.go
│   │   ├── comment_validation_test.go
│   │   └── reaction_test.go
│   ├── repository/
│   │   ├── user_repository_test.go
│   │   ├── post_repository_test.go
│   │   └── comment_repository_test.go
│   └── service/
│       ├── user_service_test.go
│       └── post_service_test.go
├── integration/
│   ├── api/
│   │   ├── user_api_test.go
│   │   ├── post_api_test.go
│   │   └── comment_api_test.go
│   └── security/
│       ├── id_enumeration_test.go
│       ├── authorization_test.go
│       └── json_leakage_test.go
└── performance/
    ├── query_performance_test.go
    └── join_performance_test.go
```

### Test Helpers
```go
func setupTestDB(t testing.TB) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    assert.NoError(t, err)

    // Run migrations
    runMigrations(t, db)

    return db
}

func createTestUser(t testing.TB, repo ports.UserRepository) *domain.User {
    user := &domain.User{
        Email: "test@example.com",
        Username: "testuser",
        PasswordHash: "hash",
        Role: domain.RoleUser,
    }
    err := repo.Create(context.Background(), user)
    assert.NoError(t, err)
    return user
}

func loginAsUser(t testing.TB, app *http.ServeMux, user *domain.User) *domain.Session {
    // Implementation for logging in and returning session
}
```

## CI/CD Integration

### Test Commands
```bash
# Run all tests
make test

# Run security tests only
go test ./tests/integration/security/... -v

# Run performance tests
go test ./tests/performance/... -bench=. -benchmem

# Run with coverage
make test-coverage
```

### Coverage Requirements
- Overall coverage: >80%
- Security tests: 100% coverage
- Repository layer: >90% coverage
- Handler layer: >85% coverage

## Conclusion

This comprehensive test suite ensures that:
1. Internal IDs are never exposed externally
2. Public UUIDs are properly validated and used
3. Authorization is enforced
4. Information disclosure is prevented
5. Performance is maintained
6. Security vulnerabilities are caught

All tests should be run before deploying any changes to the schema refactor.</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/schema_refactor_test_recommendations.md