# Integration Tests

This directory contains integration tests for the Forum application. These tests verify that multiple components work together correctly, including database operations, service layer logic, and HTTP request/response cycles.

## Test Files

### `user_stats_test.go`

Tests for user statistics calculation and retrieval.

**Functions**:
- `TestUserStats_PostAndCommentCounts` - Verifies correct counting of posts and comments for users
- `TestUserStats_EmptyStats` - Verifies zero counts for users with no activity  
- `TestBuildCurrentUser_IntegrationWithStats` - Tests the full buildCurrentUser flow

**Coverage**:
- Repository layer: `GetUserStats` method
- Service layer: User service stats retrieval
- Template data preparation

### `user_card_test.go`

Integration tests for the user-card component display.

**Functions**:
- `TestUserCard_PostAndCommentCountsDisplay` - Full HTTP flow test with session validation
- `TestUserCard_HTMLRendering` - HTML template rendering verification

**Coverage**:
- Auth service: Register, login, session validation
- User service: User retrieval, stats calculation
- Post service: Post creation
- HTTP handlers: Request/response cycle
- Template rendering: Stats display in HTML

### `auth_test.go`

Tests for authentication flows (existing).

### `post_test.go`

Tests for post CRUD operations (existing).

## Running Tests

### Run all integration tests:
```bash
go test -v ./tests/integration/
```

### Run specific test pattern:
```bash
# User stats tests only
go test -v ./tests/integration/ -run "UserStats"

# User card tests only
go test -v ./tests/integration/ -run "UserCard"

# Combined user-related tests
go test -v ./tests/integration/ -run "UserStats|UserCard"
```

### Run with coverage:
```bash
go test -v -cover ./tests/integration/
```

### Run with race detection:
```bash
go test -v -race ./tests/integration/
```

## Test Database Schema

Integration tests use in-memory SQLite databases with the same schema as production:

**Required Tables**:
- `users` - User accounts
- `sessions` - Authentication sessions
- `posts` - User posts with categories
- `comments` - Comments on posts
- `reactions` - Likes/dislikes on posts and comments
- `categories` - Post categories
- `post_categories` - Many-to-many relationship

## Test Data

Tests create their own isolated test data:
- Unique users per test (using timestamps)
- Sample posts and comments
- Sample reactions (likes/dislikes)
- Sample categories

**No shared state between tests** - each test is independent.

## Assertions

Tests verify:
- ✅ Correct data returned from repositories
- ✅ Correct calculations in service layer
- ✅ Proper session handling
- ✅ Accurate HTTP response codes and bodies
- ✅ Valid JSON structure
- ✅ Correct HTML rendering
- ✅ Template variable access patterns

## Performance

Integration tests are fast:
- Most tests complete in <10ms
- Full test suite runs in <1 second
- Uses in-memory databases (no disk I/O)

## CI/CD Integration

These tests are included in:
- `make test` - Full test suite
- `scripts/tests/run_all_tests.sh` - Automated test runner
- GitHub Actions (if configured)

## Adding New Tests

When adding integration tests:

1. Create test file: `<feature>_test.go`
2. Use `package integration`
3. Import required modules
4. Setup in-memory database with schema
5. Create test data
6. Execute operations
7. Verify results with assertions

**Example**:
```go
func TestMyFeature(t *testing.T) {
    // Setup
    db, _ := sql.Open("sqlite3", ":memory:")
    defer db.Close()
    
    // Create schema
    _, _ = db.Exec(createTablesSQL)
    
    // Initialize repos and services
    repo := adapters.NewRepo(db)
    service := app.NewService(repo)
    
    // Test
    result, err := service.DoSomething(ctx, input)
    
    // Verify
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

## Troubleshooting

### Test fails with "table not found"
- Ensure all required tables are created in `createTablesSQL`
- Check schema matches migrations

### Test fails with "foreign key constraint"
- Create parent records before child records
- Check relationship setup order

### Test hangs or times out
- Check for goroutine leaks
- Ensure database connections are closed
- Use `defer db.Close()`

## Related Documentation

- `/docs/USER_CARD_STATS_FIX.md` - User card stats implementation details
- `/docs/ARCHITECTURE.md` - Overall architecture
- `/docs/IMPLEMENTATION_ROADMAP.md` - Development progress
- `/internal/modules/*/flow.md` - Module-specific flows
