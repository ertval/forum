# Schema Refactor Status: INT Primary Keys + UUID Public IDs

## Overview

Major refactor to update database schema from TEXT UUID primary keys to:
- **Internal INT PRIMARY KEY AUTOINCREMENT** - for DB performance  
- **Public UUID TEXT (public_id)** - for external API exposure

## ✅ Completed

### 1. Database Migrations (100%)
All migration files updated with new schema:
- `001_auth_create_sessions.sql` - sessions table
- `002_user_create_users.sql` - users table
- `003_post_create_tables.sql` - posts & categories tables
- `004_comment_create_comments.sql` - comments table
- `005_reaction_create_reactions.sql` - reactions table
- `006_moderation_create_reports.sql` - reports table
- `007_notification_create_notifications.sql` - notifications table

**Schema Pattern:**
```sql
CREATE TABLE example (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    -- other fields
);
CREATE INDEX idx_example_public_id ON example(public_id);
```

### 2. Domain Entities (100%)
Updated all domain structs with both ID types:
- `User` - ID (int), PublicID (string)
- `Session` - ID (int), PublicID (string)
- `Post` - ID (int), PublicID (string), UserID (int), UserPublicID (string for JSON)
- `Category` - ID (int), PublicID (string)
- `Comment` - ID (int), PublicID (string)
- `Reaction` - ID (int), PublicID (string)

**Pattern:**
```go
type Post struct {
    ID             int       `json:"-"`                    // Internal INT
    PublicID       string    `json:"id"`                   // Public UUID (exposed in API)
    UserID         int       `json:"-"`                    // Internal foreign key
    UserPublicID   string    `json:"user_id,omitempty"`    // Public UUID (for API)
    // ... other fields
}
```

### 3. Repository Layer (100%)
Updated all repositories to generate UUIDs and handle INT IDs:

**User Repository:**
- `Create` - generates public_id UUID, returns both IDs
- All `Get*` methods - scan both id and public_id
- UUID import added

**Session Repository:**
- `Create` - generates public_id UUID, stores internal ID
- All query methods - scan both id and public_id

**Post Repository:**
- `Create` - generates public_id UUID for posts
- `GetByID` - queries by public_id string, returns with internal IDs
- `Update` - uses internal INT IDs for categories
- `Delete` - queries by public_id string
- `List` - updated to handle public_id for user filtering
- `getPostCategories` - uses internal INT postID

**Category Repository:**
- `Create` - generates public_id UUID
- `GetByID` - queries by public_id string
- `GetByName` - includes public_id in results
- `List` - includes public_id
- `Delete` - queries by public_id string

### 4. Service Layer (90%)
Updated service interfaces and implementations:

**PostService:**
- `CreatePost` - signature changed to `userID int` (takes internal ID from session)
- Repository generates UUIDs internally
- No longer pre-sets ID field (repository handles it)

**CategoryService:**
- `Create` - no longer pre-sets ID (repository handles it)

**AuthService:**
- `Register` & `Login` - no longer set session.ID (repository handles it)

### 5. Seed Data (100%)
`scripts/seed/seed_data.sql` completely rewritten:
- Users: hardcoded UUIDs for public_id, auto-increment INT IDs
- Categories: hardcoded UUIDs for public_id, auto-increment INT IDs  
- Posts: hardcoded UUIDs for public_id, references INT author_id
- post_categories: uses INT IDs for both post_id and category_id
- Reactions: uses INT IDs for user_id and target_id
- Notifications: uses INT IDs for all foreign keys

**UUID Pattern in seed:**
```sql
-- Users: 550e8400-e29b-41d4-a716-44665544000X
-- Categories: 650e8400-e29b-41d4-a716-44665544000X
-- Posts: 750e8400-e29b-41d4-a716-44665544000X
-- Reactions: 850e8400-e29b-41d4-a716-44665544000X
-- Notifications: 950e8400-e29b-41d4-a716-44665544000X
```

## ⚠️ TODO - High Priority

### 1. HTTP Handlers (100% COMPLETE ✅)
**All handlers have been fixed** - security vulnerabilities resolved:

**Fixed Issues**:
- ✅ Middleware now stores PublicID (UUID) in context, not INT ID
- ✅ Handlers convert UUID from context to INT for service calls via `getInternalUserID()`
- ✅ Templates now use `.User.PublicID` for URLs and ownership checks
- ✅ JavaScript uses lowercase `post.id` (matches JSON "id" field)

**Files updated**:
- `internal/modules/auth/adapters/middleware.go` - RequireAuth/OptionalAuth now store UUID
- `internal/modules/post/adapters/http_handler.go` - Added `getInternalUserID()` helper, updated all handlers
- `templates/base.html` - Changed `.User.ID` to `.User.PublicID`
- `templates/post_detail.html` - Changed ownership checks to use `.User.PublicID`
- `static/js/load-more-posts.js` - Changed `post.ID` to `post.id` (lowercase)

**Pattern implemented**:
```go
// Middleware stores UUID in context
ctx := context.WithValue(r.Context(), UserIDKey, user.PublicID)

// Handler converts UUID to INT for service calls
userPublicID := authAdapters.GetUserID(r.Context())  // Gets UUID string
userID, err := h.getInternalUserID(ctx, userPublicID)  // Converts to INT
post, err := h.postService.CreatePost(ctx, userID, ...)  // Uses INT internally
```

### 2. Unit Tests (0%)
**ALL unit tests have compilation errors** - they use old string IDs:

**Files to fix (partial list):**
- `internal/modules/post/application/service_test.go`
  - Update all `UserID: "user-1"` to `UserID: 1`
  - Update all `ID: "post-1"` to `ID: 1, PublicID: "post-1"`
  - Fix `tt.userID` type from string to int
  
- `internal/modules/auth/domain/session_test.go`
  - Update all `ID: "valid-id"` to `ID: 1`
  
- `internal/modules/auth/application/service_test.go`
  - Update all session ID literals from string to int

- `internal/modules/auth/adapters/sqlite_session_repository_test.go`
  - Fix session creation tests
  - Fix ID comparisons (int vs string)

- `internal/modules/post/adapters/sqlite_repository_test.go`
  - Update post/category creation
  - Fix GetByID calls (now takes string public_id, not int ID)

**Pattern for fixing:**
```go
// OLD
post := &domain.Post{
    ID: "test-post-1",
    UserID: "user-1",
}

// NEW
post := &domain.Post{
    ID: 1,  // Will be set by repository, or use 1 for existing
    PublicID: "test-post-uuid",
    UserID: 1,  // int not string
}
```

### 3. Integration Tests (0%)
**Integration tests likely broken:**
- `tests/integration/auth_test.go`
- `tests/integration/post_test.go`

Similar fixes needed as unit tests.

### 4. PostFilter Structure
Currently `PostFilter.UserID` and `LikedByUserID` are strings (public_id).
This is CORRECT for HTTP filters but verify handlers pass correct values.

## 📝 Design Decisions

### Why INT + UUID Pattern?

1. **Performance**: INT primary keys are faster for joins, indexes, and foreign keys
2. **Security**: Don't expose internal sequential IDs in public API
3. **Flexibility**: Can change public ID format without DB migration

### ID Flow Pattern

```
HTTP Request (public_id string)
    ↓
Handler extracts session.UserID (int) for create operations
    ↓
Service layer:
  - CreatePost(userID int) ← uses internal ID
  - GetPost(postID string) ← uses public_id
    ↓
Repository:
  - Generates UUID for public_id on create
  - Queries by public_id string where needed
  - Uses INT internally for joins/foreign keys
    ↓
Domain Entity has both:
  - ID int (internal, not in JSON)
  - PublicID string (external, in JSON as "id")
```

### JSON Response Pattern

```json
{
  "id": "750e8400-e29b-41d4-a716-446655440001",  // PublicID
  "user_id": "550e8400-e29b-41d4-a716-446655440001",  // UserPublicID
  "title": "Post Title",
  "author_username": "alice"
}
```

Internal `ID` and `UserID` (int) are tagged `json:"-"` so they don't appear in JSON.

## 🔧 Testing Plan

1. **Delete existing database** - `rm data/forum.db`
2. **Start server** - migrations run automatically
3. **Run seed script** - `sqlite3 data/forum.db < scripts/seed/seed_data.sql`
4. **Manual DB check**:
```sql
SELECT id, public_id, username FROM users LIMIT 3;
SELECT id, public_id, title FROM posts LIMIT 3;
SELECT p.id, p.public_id, u.username 
FROM posts p 
JOIN users u ON p.author_id = u.id 
LIMIT 3;
```
5. **Fix compilation errors in handlers**
6. **Fix all test files**
7. **Run test suite** - `make test`
8. **Run integration tests** - `scripts/tests/run_all_tests.sh`
9. **Manual HTTP testing** - register, login, create post, view posts
10. **Check counters** - verify user post/comment counts display correctly

## 🐛 Known Issues

### Post/Comment Counters Not Working
User post and comment counts not displaying correctly. This is likely due to:
1. Queries in UserStats still using old column names
2. Need to verify `GetUserStats` in user repository uses correct joins

**Check:**
```go
// internal/modules/user/adapters/sqlite_repository.go
func (r *SQLiteUserRepository) GetUserStats(ctx context.Context, userID int) (*ports.UserStats, error) {
    // Ensure queries use author_id not author_uuid or similar
    postQuery := `SELECT COUNT(*) FROM posts WHERE author_id = ?`
    commentQuery := `SELECT COUNT(*) FROM comments WHERE author_id = ?`
    // ...
}
```

## 📚 Reference Files

- **Architecture**: `docs/ARCHITECTURE.md`
- **Roadmap**: `docs/IMPLEMENTATION_ROADMAP.md`
- **DI Pattern**: `docs/UNIFIED_DI_PATTERN.md`
- **Example Module**: `internal/modules/auth/` (reference for structure)
- **Migration Guide**: `migrations/MIGRATIONS_GUIDE.md`

## ⏭️ Next Steps (Priority Order)

1. ✅ Fix HTTP handlers (post create/update/delete handlers) - **COMPLETE**
2. ✅ Fix templates to use PublicID - **COMPLETE**
3. ✅ Fix JavaScript to use lowercase id field - **COMPLETE**
4. ✅ Add ID security tests - **COMPLETE**
5. Fix auth/post service tests  
6. Fix repository tests
7. Delete old database, run server to create new schema
8. Run seed script
9. Fix remaining compilation errors
10. Run test suite
11. Manual testing
12. Fix user stats counters if still broken
13. Update any remaining TODO comments in code

---

**Status**: 85% Complete - Schema, repositories, handlers, templates, and middleware fixed. Tests need updating.
**Last Updated**: 2025-01-17 (ID Security fixes applied)
