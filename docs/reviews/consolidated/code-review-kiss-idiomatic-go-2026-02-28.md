# Code Review: Post & Comment Modules – Idiomatic Go & KISS

**Date**: 2026-02-28  
**Scope**: All 37 files in `internal/modules/post/` and `internal/modules/comment/`  
**Principles**: Idiomatic Go, KISS (Keep It Simple, Stupid), DRY, performance

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Cross-Module Issues](#cross-module-issues)
3. [Post Module](#post-module)
   - [domain/](#post-domain)
   - [ports/](#post-ports)
   - [application/](#post-application)
   - [adapters/](#post-adapters)
4. [Comment Module](#comment-module)
   - [domain/](#comment-domain)
   - [ports/](#comment-ports)
   - [application/](#comment-application)
   - [adapters/](#comment-adapters)
5. [Priority Action Items](#priority-action-items)

---

## Executive Summary

The post and comment modules are functional and well-structured at the hexagonal-architecture level. However, they share several systemic issues:

| Category | Count | Severity |
|----------|-------|----------|
| Code duplication (cross-module) | 6 | High |
| Code duplication (intra-module) | 8 | Medium |
| N+1 query patterns | 4 | High |
| Type-safety violations | 3 | Medium |
| Stale/broken test code | 2 | Medium |
| Bug (validation mismatch) | 1 | Medium |
| KISS violations (over-abstraction) | 3 | Low |
| Dead/unused code | 2 | Low |

---

## Cross-Module Issues

### CM-1. Massive Handler Duplication (HIGH)

The following functions are **copy-pasted** nearly identically between both modules' `adapters/http_handler.go`:

| Function | Post handler | Comment handler |
|----------|-------------|-----------------|
| `buildCurrentUser()` | Lines 136–171 | Lines 78–117 |
| `getInternalUserID()` | Lines 174–186 | Lines 120–133 |
| `writeJSON()` | Lines 189–196 | Lines 136–143 |
| `ServiceContainer` interface | Lines 20–30 | Lines 24–37 |

**Fix**: Extract to a shared `internal/platform/handler` package:

```go
// internal/platform/handler/helpers.go
package handler

type CurrentUser struct {
    PublicID     string
    Username     string
    Email        string
    AvatarPath   string
    PostCount    int
    CommentCount int
}

func BuildCurrentUser(ctx context.Context, userSvc UserService, userID int) CurrentUser { ... }
func GetInternalUserID(ctx context.Context, userSvc UserService, publicID string) (int, error) { ... }
func WriteJSON(w http.ResponseWriter, status int, data interface{}) { ... }
```

### CM-2. `map[string]interface{}` for Template Data (MEDIUM)

Both modules construct template data as `map[string]interface{}`, which is:
- Not type-safe (typos in keys are silent bugs)
- Hard to refactor
- Not idiomatic Go

**Locations**: All page handlers, `buildCurrentUser()`, `createPostPreview()`, all `commentData` construction in comment handler.

**Fix**: Define typed structs:

```go
type CurrentUser struct {
    PublicID     string
    Username     string
    Email        string
    AvatarPath   string
    PostCount    int
    CommentCount int
}

type PostPreview struct {
    PublicID    string
    Title       string
    Preview     string
    // ...
}
```

### CM-3. Fire-and-Forget Goroutines Without Error Tracking (MEDIUM)

Both `post/application/service.go` and `comment/application/service.go` use the same pattern:

```go
go func() {
    if err := s.userService.IncrementPostCount(ctx, userID); err != nil {
        log.Printf("Warning: failed to increment post count: %v", err)
    }
}()
```

Issues:
- `ctx` may be cancelled before goroutine runs (request-scoped context)
- Errors are only logged, never tracked
- No way to test or observe failures

**Fix**: Use `context.WithoutCancel(ctx)` (Go 1.21+) and consider a simple async worker:

```go
go func() {
    bgCtx := context.WithoutCancel(ctx)
    if err := s.userService.IncrementPostCount(bgCtx, userID); err != nil {
        log.Printf("Warning: failed to increment post count: %v", err)
    }
}()
```

### CM-4. Inconsistent Validation Between Modules (LOW)

| Aspect | Post module | Comment module |
|--------|------------|----------------|
| Empty check | `p.Title == ""` | `strings.TrimSpace(c.Content) == ""` |
| Length check | `len(p.Title)` (byte count) | `len([]rune(c.Content))` (char count) |

The comment module is more correct (handles Unicode & whitespace). The post module should adopt the same approach.

---

## Post Module

### Post Domain

#### `domain/post.go` (53 lines)

**Summary**: Defines the `Post` struct with dual-ID system, `Validate()`, and `HasImage()`.

**Issues**:

1. **BUG – Title length mismatch** (Line 35 vs `errors.go` Line 35):
   - `Validate()` checks `len(p.Title) > 255`
   - `ErrTitleTooLong` message says "max 300 characters"
   - One of these is wrong. Fix to be consistent.

2. **Redundant field** – `Author` field (Line 15) duplicates `AuthorUsername` (Line 14):
   ```go
   AuthorUsername string `json:"author_username,omitempty"`
   Author         string `json:"author,omitempty"` // "for compatibility"
   ```
   If backward compatibility requires both JSON keys, use a custom `MarshalJSON` instead of carrying two fields through the entire codebase.

3. **Validation uses byte length, not rune length** – `len(p.Title)` counts bytes. A 100-character Chinese title uses 300 bytes. Use `utf8.RuneCountInString()` if the intent is character limits (like the comment module does).

#### `domain/category.go` (~28 lines)

**Summary**: `Category` struct with `Validate()`. Clean and correct.

No issues.

#### `domain/filter.go` (40 lines)

**Summary**: Two parallel structs: `PostFilter` (query-level) and `FilterParams` (HTTP-level).

**Issues**:

1. **KISS violation – Two overlapping filter structs**: `FilterParams` has highly redundant fields that mirror `PostFilter`:
   - `MyPosts bool` → same as `UserID == CurrentUserID`
   - `CommentedPosts bool` → same as `Commenter == CurrentUserID`
   - `LikedPosts bool` → same as `LikedByUserID == CurrentUserID`
   
   The conversion logic in `FilterService.BuildFilter` is complex specifically because these booleans exist. Simplify `FilterParams` to only carry raw HTTP values, or eliminate it entirely and parse directly into `PostFilter`.

2. **Missing closing paren in doc comment** (Line 39): `// Filter by posts commented on by this user (public ID` — missing `)`.

#### `domain/errors.go` (49 lines)

**Summary**: Standard sentinel errors.

**Issues**:

1. **ErrTitleTooLong says "max 300"** (Line 35) but `Validate()` enforces 255. Fix one or the other.

#### `domain/post_test.go` & `domain/category_test.go`

**Summary**: Well-structured table-driven tests. Correct.

No issues.

---

### Post Ports

#### `ports/service.go` (53 lines)

**Summary**: Three interfaces: `PostService`, `CategoryService`, `FilterService`.

**Issues**:

1. **Over-abstraction – `FilterService` as interface** (Lines 45–53): `FilterService` is a pure data transformer with no dependencies. It doesn't need to be an interface—plain functions would be simpler and equally testable:
   ```go
   func BuildFilter(params domain.FilterParams) domain.PostFilter { ... }
   ```

#### `ports/repository.go` (~40 lines)

Clean. No issues.

#### `ports/image.go` (33 lines)

**Summary**: `ImageHandler` interface + `ImageUploadRequest` struct.

**Issues**:

1. **Likely dead code – `ImageUploadRequest`**: The struct and its methods (`IsEmpty()`, `HasNewImage()`) are defined but don't appear to be used in any handler or service. Verify with `grep -r "ImageUploadRequest" internal/` and remove if unused.

#### `ports/service_test.go` (~148 lines)

**Summary**: Interface compilation tests using mock structs.

**Issues**:

1. **Stale mock** – `mockPostService` may be missing the `MaxImageSize()` method if the interface has been updated. These tests only verify compilation, which the compiler already does—consider removing them.

---

### Post Application

#### `application/service.go` (~237 lines)

**Summary**: Core service implementing `PostService` and `CategoryService`.

**Issues**:

1. **N+1 category verification** (CreatePost, ~Line 85): Categories are verified one-by-one in a loop:
   ```go
   for _, catName := range categories {
       _, err := s.categoryRepo.GetByName(ctx, catName)
       if err != nil { return nil, domain.ErrCategoryNotFound }
   }
   ```
   **Fix**: Add a `GetByNames(ctx, names []string) ([]*Category, error)` repository method for batch lookup.

2. **CategoryService in same file** (~Line 180+): `CategoryService` is a separate type in the same file. It should be in its own `category_service.go` file for discoverability, or justify co-location with a comment.

#### `application/filter_service.go` (103 lines)

**Summary**: Stateless `FilterService` struct implementing filter conversion.

**Issues**:

1. **Unused context parameter** (Line 22): `_ = ctx` confirms ctx is unused. Remove it from the method signature or the interface.

2. **Stateless struct** – `FilterService` has zero fields. The constructor `NewFilterService()` returns `&FilterService{}`. This is over-engineering for what should be a package-level function.

3. **Complex branching in `BuildFilter`** (Lines 30–95): The method has ~8 conditional branches converting `FilterParams` → `PostFilter`. This complexity exists because `FilterParams` carries redundant boolean flags. Simplify `FilterParams` (see CM filter issues above) and this function becomes trivial.

#### `application/service_test.go` (~1149 lines) & `application/service_image_test.go`

**Summary**: Thorough tests with inline mock implementations.

**Issues**:

1. **Mock explosion** – Each test function defines its own mock struct inline. Consider a shared `testutil` file within the package with configurable mocks using function fields (which the file already partially uses). The pattern is good; it just needs consolidation to reduce file length.

---

### Post Adapters

#### `adapters/http_handler.go` (283 lines)

**Summary**: Base handler, ServiceContainer interface, helper methods.

**Issues**:

1. **`buildCurrentUser` returns `map[string]interface{}`** (Line 136): See CM-2.

2. **`normalizeBoardActivityType` and `normalizeReactionType`** (Lines 198–230): These functions switch on strings and default to "all". They're fine but the string-based filtering system is error-prone—consider using an enum/`const` block with typed constants.

3. **Large ServiceContainer** (Lines 20–30): 9 methods. This handler depends on too many services. Consider breaking the handler into smaller handlers per concern (posts, board, categories).

#### `adapters/http_handler_api.go` (622 lines)

**Summary**: CRUD API handlers for posts.

**Issues**:

1. **DRY – previewPost map construction duplicated** (Lines ~130, ~350, ~440): The same `map[string]interface{}` construction for post previews appears in `CreatePostAPI`, `ListPostsAPI`, and `LoadMorePostsAPI`. Extract to `createPostPreview` (which already exists for content truncation but not the full map).

2. **Category parsing has 3 fallback strategies** in `CreatePostAPI` (Lines ~80–120):
   ```go
   // Try 1: JSON string array
   // Try 2: Form values
   // Try 3: Comma-separated string
   ```
   This complexity should be in a dedicated `parseCategoriesFromRequest()` helper.

3. **`buildFilter` method duplicates `FilterService.BuildFilter`** (Lines 580–622): The handler contains its own inline `buildFilter` method as a "fallback." This violates DRY and means filter logic has two sources of truth.

   **Fix**: Always use `FilterService.BuildFilter`; remove the handler's `buildFilter`.

4. **`ListPostsAPI` and `LoadMorePostsAPI` are ~80% identical** (Lines 280–440 and 440–580): Same filter construction, same post map building, same category fetching. Only pagination differs.

   **Fix**: Extract shared logic into a private `listPosts(ctx, filter) ([]map[string]interface{}, error)` method.

5. **UpdatePostAPI doesn't support image updates** but has logging with inline config creation (Lines ~215–240). The logging setup is verbose and shouldn't be inline.

#### `adapters/http_handler_page.go` (513 lines)

**Summary**: HTML page handlers (home, board, detail, create, edit).

**Issues**:

1. **`HomePage` and `BoardPage` are ~90% identical** (Lines 30–210 and 210–400): Both build filter params, fetch posts, construct preview maps, fetch categories, render template. The only differences are: template name, URL path check, and page title.

   **Fix**: Extract shared logic into `listPostsPage(w, r, templateName, pageTitle)`.

2. **N+1 in `PostDetailPage`** (Lines ~420–480): For each comment, a separate user lookup is made:
   ```go
   for _, comment := range comments {
       author, err := h.userService.GetByID(ctx, comment.UserID)
   ```
   **Fix**: Collect unique user IDs first, batch-fetch, then map.

3. **Preview map construction duplicated again** – same `map[string]interface{}` with keys `PublicID`, `Title`, `Preview`, `AuthorUsername`, etc. appears for the third time in page handlers.

#### `adapters/image_upload.go` (120 lines)

**Summary**: Image parsing and validation helpers.

**Issues**:

1. **Two similar parse functions** – `ParseImageUpload` (from `r.Body`) and `ParseMultipartImageUpload` (from `multipart.File`) share validation logic but are separate functions. Consider unifying: both ultimately produce `[]byte` from an `io.Reader`.

#### `adapters/sqlite_repository.go` (679 lines)

**Summary**: SQLite implementations for `PostRepository` and `CategoryRepository`.

**Issues**:

1. **N+1 in `List()`** (Lines ~200–250): `getPostCategories()` is called once per post in the loop. For a page of 20 posts, this is 20 extra queries.

   **Fix**: After fetching posts, collect all post IDs, do a single `SELECT post_id, c.name FROM post_categories JOIN categories ...  WHERE post_id IN (?,?)`, then group results in-memory:
   ```go
   func (r *SQLitePostRepository) getPostCategoriesBatch(ctx context.Context, postIDs []int64) (map[int64][]string, error) { ... }
   ```

2. **Scan logic duplicated between `GetByID` and `List`** (Lines ~80 and ~180): Both scan the same columns from the posts table with overlapping code. Extract a `scanPost(rows *sql.Rows) (*domain.Post, error)` helper.

3. **`repeatPlaceholders` uses string concatenation in loop** (Lines ~650–665):
   ```go
   func repeatPlaceholders(count int) string {
       result := ""
       for i := 0; i < count; i++ {
           if i > 0 { result += "," }
           result += "?"
       }
       return result
   }
   ```
   **Fix**: Use `strings.Repeat` or `strings.Join`:
   ```go
   func repeatPlaceholders(n int) string {
       return strings.Join(slices.Repeat([]string{"?"}, n), ",")
   }
   ```

4. **`normalizeImagePath`** (Lines ~630–645): Multiple `strings.TrimPrefix` calls could be simplified with a single helper or a loop over prefixes.

5. **Dynamic SQL building with string concatenation** (Lines ~130–200): The `List()` method builds SQL with `conditions = append(conditions, ...)` and `args = append(args, ...)`. This is functional but consider a lightweight query builder pattern for readability.

#### `adapters/http_handler_test.go` (2305 lines)

**Summary**: Extensive handler tests with mock structs and table-driven tests.

**Issues**:

1. **Mock structs duplicated from comment module tests** – Both modules define nearly identical `mockAuthService`, `mockUserService`, etc. Extract to a shared `internal/testutil/` package.

2. **File is very long** (2305 lines): Consider splitting into `http_handler_api_test.go` and `http_handler_page_test.go` to match the handler file structure.

#### `adapters/image_upload_test.go` (~336 lines) & `adapters/sqlite_repository_test.go` (~1686 lines)

Clean, thorough test files. The repository tests are well-structured integration tests using in-memory SQLite.

No significant issues.

---

## Comment Module

### Comment Domain

#### `domain/comment.go` (40 lines)

**Summary**: `Comment` struct with `Validate()` using `strings.TrimSpace` and `len([]rune(c.Content))`.

**Issues**:

1. **Better validation than post module** – This is the *correct* approach (trim whitespace, count runes). Post module should adopt this.

No other issues; clean file.

#### `domain/errors.go` (~20 lines)

Clean sentinel errors. No issues.

#### `domain/comment_test.go`

Well-structured table-driven tests including struct field existence test.

No issues.

---

### Comment Ports

#### `ports/service.go` (~30 lines)

**Summary**: `CommentService` interface with CRUD + pagination methods.

No issues.

#### `ports/repository.go` (~25 lines)

Clean interface. No issues.

#### `ports/service_test.go` (~145 lines)

**Summary**: Interface compilation tests with mock structs.

**Issues**:

1. **CRITICAL – Stale mocks**: The `mockCommentService` uses `int` parameters:
   ```go
   func (m *mockCommentService) GetComment(ctx context.Context, commentID int) (*domain.Comment, error) {
   ```
   But the actual `CommentService` interface uses `string` (public IDs):
   ```go
   GetComment(ctx context.Context, publicID string) (*domain.Comment, error)
   ```
   
   These tests **compile** because Go interfaces are structurally typed and the mock is never actually assigned to the interface type in a way the compiler would check. The `var _ ports.CommentService = &mockCommentService{}` line (if present) would catch this, but it may be missing or the mocks were updated incorrectly.

   **Fix**: Update all mock method signatures to match the current interface, or delete these tests (the compiler already verifies interface satisfaction).

---

### Comment Application

#### `application/service.go` (~167 lines)

**Summary**: Service implementing `CommentService` with notification integration.

**Issues**:

1. **`UpdateComment` reconstructs the entire comment** (Lines ~90–110):
   ```go
   updatedComment := &domain.Comment{
       ID:           existingComment.ID,
       PublicID:      existingComment.PublicID,
       PostID:        existingComment.PostID,
       // ... copies all fields
       Content:       content,
       UpdatedAt:     time.Now(),
   }
   return s.repo.Update(ctx, updatedComment)
   ```
   **Fix**: Modify in place:
   ```go
   existingComment.Content = content
   existingComment.UpdatedAt = time.Now()
   return s.repo.Update(ctx, existingComment)
   ```

2. **`ListCommentsByUser` and `ListCommentsByUserPaginated` are near-duplicates** (Lines ~120–160): Both do user lookup → repo call. The only difference is the repo method called. Consolidate:
   ```go
   func (s *Service) ListCommentsByUserPaginated(ctx context.Context, userPublicID string, limit, offset int) ([]domain.Comment, error) {
       // if limit == 0, use ListByUser; else use ListByUserPaginated
   }
   ```
   Or better: always paginate (with a high default limit).

3. **Same fire-and-forget goroutine issue as post module** – See CM-3.

#### `application/service_test.go` (~610 lines)

**Summary**: Tests with mock implementations.

Well-structured, reasonable length. No significant issues beyond the mock duplication noted in CM-1.

---

### Comment Adapters

#### `adapters/http_handler.go` (198 lines)

**Summary**: Base handler with duplicated helpers from post module.

**Issues**:

1. **Complete duplication of post handler helpers** – See CM-1.

2. **`parseJSON` with `DisallowUnknownFields`** (Lines ~145–160): Stricter than the post module's JSON parsing. This inconsistency means the post API silently ignores unknown fields while the comment API rejects them.

   **Fix**: Pick one approach and use it consistently.

#### `adapters/http_handler_api.go` (392 lines)

**Summary**: API handlers for comment CRUD + listing.

**Issues**:

1. **Double-fetch in `UpdateCommentAPI`** (~Lines 150–180): Handler fetches the comment for ownership checking, then calls `s.commentService.UpdateComment()` which fetches the comment *again* internally.

   **Fix**: Either pass the already-fetched comment to the service, or have the service accept a user ID for authorization and do the fetch only once.

2. **Inline struct definition repeated 3 times in `ListCommentsByPostAPI`** (Lines 350–390):
   ```go
   var commentsResp []struct {
       ID string `json:"id"`; PostID string `json:"post_id"` // ...
   }
   for _, comment := range comments {
       commentResp := struct {
           ID string `json:"id"`; PostID string `json:"post_id"` // ...
       }{ ... }
       commentsResp = append(commentsResp, commentResp)
   }
   h.writeJSON(w, http.StatusOK, struct {
       Comments []struct {
           ID string `json:"id"`; PostID string `json:"post_id"` // ...
       } `json:"comments"`
   }{ Comments: commentsResp })
   ```
   **Fix**: Define the struct once:
   ```go
   type commentResponse struct {
       ID        string `json:"id"`
       PostID    string `json:"post_id"`
       UserID    string `json:"user_id"`
       Content   string `json:"content"`
       CreatedAt string `json:"created_at"`
       UpdatedAt string `json:"updated_at"`
   }
   ```

3. **User ID enrichment in `ListCommentsByPostAPI`** – Caches user public IDs per internal ID but mutates `comment.PublicUserID` which may be a copy (range produces copies of value types). Verify that `comments` is `[]domain.Comment` (values) vs `[]*domain.Comment` (pointers).

#### `adapters/http_handler_form.go` (~115 lines)

**Summary**: HTML form submission handlers for create/delete.

Clean. No significant issues.

#### `adapters/http_handler_page.go` (622 lines)

**Summary**: Activity page, my-comments page, load-more API, activity aggregation.

**Issues**:

1. **Massive N+1 in `MyCommentsPage`** (Lines ~270–400): For each comment, separate lookups for:
   - User info (`h.userService.GetByID`)
   - Post info (`h.postService.GetPost`)
   - Reaction counts (`h.reactionService.CountReactions`)
   
   For 20 comments, this is 60 extra queries. Same issue in `LoadMoreCommentsAPI`.

   **Fix**: Batch-fetch users and posts by collecting unique IDs first.

2. **In-memory filtering after DB fetch** (Lines ~330–370): Category and date filters are applied *in Go after fetching all comments*, rather than at the database level. This means fetching all data then discarding most of it.

   **Fix**: Push filters into the repository query.

3. **`aggregateUserActivity` fetches 4 lists serially** (Lines ~490–540):
   ```go
   createdPosts, _ := h.postService.ListPosts(ctx, ...)
   likedPosts, _ := h.postService.ListPosts(ctx, ...)
   dislikedPosts, _ := h.postService.ListPosts(ctx, ...)
   comments, _ := h.commentService.ListCommentsByUserPaginated(ctx, ...)
   ```
   These are independent queries that could run concurrently with `errgroup`:
   ```go
   g, ctx := errgroup.WithContext(ctx)
   g.Go(func() error { createdPosts, err = ...; return err })
   g.Go(func() error { likedPosts, err = ...; return err })
   // ...
   if err := g.Wait(); err != nil { return nil, err }
   ```

4. **`filterCreatedPostItems`, `filterReactionItems`, `filterCommentItems`** (Lines 200–250): Three nearly identical filter functions that check time and category. Extract a generic filter:
   ```go
   func filterItems(items []map[string]interface{}, filters activityFilters, now time.Time, extraCheck func(map[string]interface{}) bool) []map[string]interface{} { ... }
   ```

5. **`cutoffForTimeFilter` duplicated** – Same switch-on-date-filter logic appears in `MyCommentsPage` (inline) and in `cutoffForTimeFilter` helper. Use the helper everywhere.

#### `adapters/sqlite_repository.go` (~280 lines)

**Summary**: SQLite comment repository.

**Issues**:

1. **`ListByUser` and `ListByUserPaginated` are near-copies** (Lines ~100–200): Same query, same scan logic, only LIMIT/OFFSET differs.

   **Fix**: Merge into one function. `ListByUser` can call `ListByUserPaginated(ctx, userID, math.MaxInt32, 0)`.

2. **`DeleteByPublicID` doesn't check rows affected** (~Line 230): If the comment doesn't exist, the DELETE succeeds silently (0 rows affected).

   **Fix**:
   ```go
   res, err := r.db.ExecContext(ctx, "DELETE FROM comments WHERE public_id = ?", publicID)
   if err != nil { return err }
   n, _ := res.RowsAffected()
   if n == 0 { return domain.ErrCommentNotFound }
   return nil
   ```

#### `adapters/sqlite_repository_test.go` & `adapters/activity_test.go`

Clean tests, well-structured. No significant issues.

---

## Priority Action Items

### P0 – Bugs
| # | Issue | File | Fix |
|---|-------|------|-----|
| 1 | Title max length: Validate() checks 255, error says "max 300" | post/domain/post.go:35, post/domain/errors.go:35 | Align to one value |
| 2 | Stale comment mocks use `int` instead of `string` params | comment/ports/service_test.go | Update or delete |

### P1 – Performance (N+1 Queries)
| # | Issue | File | Fix |
|---|-------|------|-----|
| 3 | `getPostCategories` per-post in `List()` | post/adapters/sqlite_repository.go:~200 | Batch query |
| 4 | Per-comment user+post+reaction lookup | comment/adapters/http_handler_page.go:~270 | Batch fetch |
| 5 | Per-comment user lookup in PostDetailPage | post/adapters/http_handler_page.go:~420 | Batch fetch |
| 6 | In-memory filtering after full DB fetch | comment/adapters/http_handler_page.go:~330 | Push to SQL |

### P2 – DRY (Code Duplication)
| # | Issue | Files | Fix |
|---|-------|-------|-----|
| 7 | `buildCurrentUser`/`getInternalUserID`/`writeJSON` | Both http_handler.go | Extract to platform/handler |
| 8 | `previewPost` map construction (3+ copies) | post/adapters/http_handler_*.go | Extract helper |
| 9 | `HomePage`/`BoardPage` ~90% identical | post/adapters/http_handler_page.go | Extract shared method |
| 10 | `ListPostsAPI`/`LoadMorePostsAPI` ~80% identical | post/adapters/http_handler_api.go | Extract shared method |
| 11 | `ListByUser`/`ListByUserPaginated` duplicated | comment/adapters/sqlite_repository.go | Merge |
| 12 | Handler `buildFilter` duplicates `FilterService` | post/adapters/http_handler_api.go:580 | Remove handler version |

### P3 – Idiomatic Go & KISS
| # | Issue | File | Fix |
|---|-------|------|-----|
| 13 | `map[string]interface{}` for template data | All page handlers | Use typed structs |
| 14 | `FilterService` as stateless struct with interface | post/application/filter_service.go | Convert to functions |
| 15 | `UpdateComment` reconstructs full struct | comment/application/service.go:~90 | Modify in place |
| 16 | `ImageUploadRequest` likely unused | post/ports/image.go:19 | Verify and remove |
| 17 | Inconsistent JSON strictness (comment strict, post lenient) | Both http_handler.go | Unify |
| 18 | Inline struct repeated 3× in `ListCommentsByPostAPI` | comment/adapters/http_handler_api.go:350 | Define once |
| 19 | `repeatPlaceholders` string concatenation | post/adapters/sqlite_repository.go:~650 | Use strings.Join |
| 20 | Serial queries in `aggregateUserActivity` | comment/adapters/http_handler_page.go:~490 | Use errgroup |

---

*End of review.*
