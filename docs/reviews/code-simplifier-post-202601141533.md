# Go Code Simplifier Review

**Folder/Module:** post
**Date:** 2026-01-14 15:33
**Files Reviewed:**

- `internal/modules/post/domain/post.go`
- `internal/modules/post/domain/errors.go`
- `internal/modules/post/domain/filter.go`
- `internal/modules/post/domain/category.go`
- `internal/modules/post/ports/repository.go`
- `internal/modules/post/ports/service.go`
- `internal/modules/post/application/service.go`
- `internal/modules/post/application/filter_service.go`
- `internal/modules/post/adapters/http_handler.go`
- `internal/modules/post/adapters/http_handler_api.go`
- `internal/modules/post/adapters/http_handler_page.go`
- `internal/modules/post/adapters/sqlite_repository.go`
- `internal/modules/post/adapters/image_upload.go`

---

## Summary

The `post` module is well-structured and follows the modular monolith architecture. However, several performance and maintainability issues were identified, including inefficient template parsing, an N+1 query problem in the repository, and duplicated logic in HTTP handlers. Applying Go idiomatic patterns and the KISS principle would significantly improve the codebase.

---

## Findings

### 1. Inefficient Template Parsing

**File:** `internal/modules/post/adapters/http_handler_page.go`
**Line(s):** 130, 249, 346, 401, 476
**Category:** Architecture / Performance
**Severity:** High

**Current Code:**

```go
tmpl, err := template.ParseFiles("templates/base.html", "templates/home.html")
if err != nil {
    http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
    return
}
```

**Suggested Improvement:**
Parse templates once during initialization and store them in the `HTTPHandler` or a global template registry.

```go
// In NewHTTPHandler or a dedicated template loader
parsedTemplates := template.Must(template.ParseFiles("templates/base.html", "templates/home.html"))

// In handler
if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
    // ...
}
```

**Rationale:** Parsing template files from disk on every single request is extremely slow and resource-intensive. It should be done once at startup.

---

### 2. N+1 Query Problem in Post Listing

**File:** `internal/modules/post/adapters/sqlite_repository.go`
**Line(s):** 431-436
**Category:** Performance / Architecture
**Severity:** High

**Current Code:**

```go
for rows.Next() {
    // ... scan post ...
    categories, err := r.getPostCategories(ctx, post.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get categories for post %d: %w", post.ID, err)
    }
    post.Categories = categories
    posts = append(posts, &post)
}
```

**Suggested Improvement:**
Use a single query with `GROUP_CONCAT` (if using SQLite) or fetch all categories for all retrieved posts in one separate query using the `IN` clause.

```go
// Option 1: Join with GROUP_CONCAT
// SELECT p.*, GROUP_CONCAT(c.name) as categories ... GROUP BY p.id

// Option 2: Second query
// ids := []int{...}
// SELECT post_id, c.name FROM post_categories pc JOIN categories c ... WHERE post_id IN (...)
```

**Rationale:** Calling `getPostCategories` for every post in a list (e.g., 50 posts) results in 51 database queries instead of 1 or 2. This significantly degrades performance as the number of posts grows.

---

### 3. Discrepancy in Validation Logic

**File:** `internal/modules/post/domain/post.go` and `internal/modules/post/domain/errors.go`
**Line(s):** `post.go:30`, `errors.go:33`
**Category:** Idiomatic Go / KISS
**Severity:** Low

**Current Code:**

```go
// post.go
if len(p.Title) > 255 {
    return ErrTitleTooLong
}

// errors.go
ErrTitleTooLong = errors.New("post title too long (max 300 characters)")
```

**Suggested Improvement:**
Align the validation threshold with the error message.

```go
const MaxTitleLength = 300
if len(p.Title) > MaxTitleLength {
    return ErrTitleTooLong
}
```

**Rationale:** Mismatched validation logic and error messages confuse developers and users. Using constants for such limits is also more idiomatic.

---

### 4. Duplicate Page Handler Logic

**File:** `internal/modules/post/adapters/http_handler_page.go`
**Line(s):** 29-142 (HomePage) and 144-260 (BoardPage)
**Category:** KISS Violation / DRY
**Severity:** Medium

**Current Code:**
`HomePage` and `BoardPage` contain nearly identical code for parsing filters, fetching posts, building previews, and preparing template data.

**Suggested Improvement:**
Extract the shared logic into a helper method.

```go
func (h *HTTPHandler) renderPostList(w http.ResponseWriter, r *http.Request, templateName string, limit int) {
    // Shared filter parsing and post fetching logic
}
```

**Rationale:** Maintaining duplicate code increases the risk of bugs and makes updates harder.

---

### 5. Inefficient Category Validation

**File:** `internal/modules/post/application/service.go`
**Line(s):** 103-108, 163-168
**Category:** Architecture / Performance
**Severity:** Medium

**Current Code:**

```go
for _, categoryName := range categories {
    _, err := s.categoryRepo.GetByName(ctx, categoryName)
    if err != nil {
        return nil, err
    }
}
```

**Suggested Improvement:**
Add a `GetByNames(ctx, names []string)` method to `CategoryRepository` to verify all categories in a single query.

**Rationale:** Multiple repository calls in a loop can be slow. A single bulk query is more efficient.

---

### 6. Use of `context.Background()` in Async Tasks

**File:** `internal/modules/post/application/service.go`
**Line(s):** 132, 197
**Category:** Idiomatic Go / Concurrency
**Severity:** Medium

**Current Code:**

```go
go func() {
    _ = s.userService.IncrementPostCount(context.Background(), userID)
}()
```

**Suggested Improvement:**
While these are intended to be "fire and forget", using `context.Background()` makes them detached from the application lifecycle. Consider using a long-lived application context or a task runner that ensures completion before shutdown.

**Rationale:** Detached goroutines using `context.Background()` can lead to leaked resources or lost updates if the application restarts.

---

### 7. Redundant Fields in Post Entity

**File:** `internal/modules/post/domain/post.go`
**Line(s):** 12-13
**Category:** KISS Violation
**Severity:** Low

**Current Code:**

```go
AuthorUsername string `json:"author_username,omitempty"`
Author         string `json:"author,omitempty"` // Alias for AuthorUsername
```

**Suggested Improvement:**
Pick one (ideally `AuthorUsername` or just `Author` if that's the public field) and use it consistently.

**Rationale:** Having two fields representing the same data adds clutter and confusion.

---

## Action Items

- [ ] Refactor template parsing to happen once at startup.
- [ ] Optimize `PostRepository.List` to avoid N+1 queries for categories.
- [ ] Align validation logic in `post.go` with error messages in `errors.go`.
- [ ] Refactor `HomePage` and `BoardPage` to share logic.
- [ ] Implement bulk category lookup in `CategoryRepository`.
- [ ] Clean up redundant fields in the `Post` entity.

---

## Notes

The usage of `getInternalUserID` in the adapter is necessary due to the "INT internally, UUID publicly" architectural decision. While it adds a bit of overhead (extra lookup), it's a trade-off for security and internal performance. However, this could be optimized by caching frequent lookups or using a more integrated approach in the session management.
