# Go Code Simplifier Review

**Folder/Module:** comment
**Date:** 2026-01-14 15:40
**Files Reviewed:**

- `internal/modules/comment/domain/comment.go`
- `internal/modules/comment/domain/errors.go`
- `internal/modules/comment/ports/service.go`
- `internal/modules/comment/application/service.go`
- `internal/modules/comment/adapters/http_handler.go`
- `internal/modules/comment/adapters/http_handler_api.go`
- `internal/modules/comment/adapters/http_handler_page.go`
- `internal/modules/comment/adapters/sqlite_repository.go`

---

## Summary

The `comment` module follows the established Hexagonal Architecture (Ports & Adapters) and Modular Monolith principles. However, there are several areas where the code can be simplified by reducing redundancy, improving idiomatic Go usage, and strictly adhering to the project's ID security rules. Key findings include redundant validation logic, N+1 query patterns in page enrichment, repeated template parsing, and exposure of internal sequential IDs in API responses.

---

## Findings

### 1. Simplify Domain Validation

**File:** `internal/modules/comment/domain/comment.go`
**Line(s):** 31-33
**Category:** KISS Violation
**Severity:** Low

**Current Code:**

```go
	// Check content is not just whitespace
	if len(c.Content) == 0 || len([]rune(c.Content)) == 0 || len(strings.TrimSpace(c.Content)) == 0 {
		return ErrEmptyContent
	}
```

**Suggested Improvement:**

```go
	// Check content is not just whitespace
	if strings.TrimSpace(c.Content) == "" {
		return ErrEmptyContent
	}
```

**Rationale:** `strings.TrimSpace(c.Content) == ""` is a more concise and idiomatic way to check if a string consists only of whitespace or is empty, covering all the individual checks perform in the current code.

---

### 2. ID Security Violation in API Responses

**File:** `internal/modules/comment/adapters/http_handler_api.go`
**Line(s):** 76, 114, 194, 282, 311
**Category:** Architecture / Security
**Severity:** High

**Current Code:**

```go
	resp := struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    int    `json:"user_id"` // <--- Exposure of internal ID
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}{
		// ...
		UserID:    comment.UserID,
	}
```

**Suggested Improvement:**
Define a shared response struct in `http_handler.go` and use the public user ID:

```go
type CommentResponse struct {
	ID        string `json:"id"`
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"` // Public UUID
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
```

**Rationale:** The `GEMINI.md` strictly forbids exposing sequential internal IDs in JSON. API responses must use the public UUID for users.

---

### 3. Avoid Template Re-parsing

**File:** `internal/modules/comment/adapters/http_handler_page.go`
**Line(s):** 185
**Category:** Performance
**Severity:** Medium

**Current Code:**

```go
	// Parse templates individually for this page
	tmpl, err := template.ParseFiles("templates/base.html", "templates/comments.html")
```

**Suggested Improvement:**
Use the pre-parsed templates from the `HTTPHandler` struct or initialize them once in `NewHTTPHandler`.

```go
	if err := h.templates.ExecuteTemplate(&buf, "base", data); err != nil {
		// ...
	}
```

**Rationale:** Parsing templates on every request is inefficient and slows down the application. Templates should be parsed once at startup.

---

### 4. Consolidated Comment Enrichment and N+1 Prevention

**File:** `internal/modules/comment/adapters/http_handler_page.go`
**Line(s):** 89-167, 254-298
**Category:** KISS Violation / Performance
**Severity:** Medium

**Current Code:**
Loops through comments and fetches author/post details individually:

```go
	for _, comment := range commentsFromService {
		user, err := h.userService.GetByID(ctx, comment.UserID)
		post, err := h.postService.GetPost(ctx, comment.PublicPostID)
		// ...
	}
```

**Suggested Improvement:**
Extract an `enrichComments` helper and implement batch fetching for users and posts to avoid N+1 database calls.

**Rationale:** Individual fetches in a loop lead to many database queries. Batching these reduces overhead. Consolidation reduces code duplication.

---

### 5. Consolidate SQLite Row Scanning

**File:** `internal/modules/comment/adapters/sqlite_repository.go`
**Line(s):** 107-115, 145-154, 186-196
**Category:** KISS Violation / Code Duplication
**Severity:** Low

**Current Code:**
Repeated `Scan` calls across multiple listing methods.

**Suggested Improvement:**

```go
func (r *SQLiteCommentRepository) scanComment(scanner interface{ Scan(...interface{}) error }) (*domain.Comment, error) {
	var comment domain.Comment
	// ... scan logic ...
	return &comment, nil
}
```

**Rationale:** Reduces boilerplate and ensures consistency when scanning rows from the database.

---

### 6. Use Middleware-Provided Auth Info

**File:** `internal/modules/comment/adapters/http_handler_page.go`
**Line(s):** 33-47, 212-222
**Category:** Idiomatic Go / Clean Architecture
**Severity:** Low

**Current Code:**
Manually extracting session from cookies and validating it in handler methods.

**Suggested Improvement:**
Since the routes are protected by `RequireAuth`, the user session or ID should be available in the context via the middleware. If not, use the `h.GetCurrentUser(r)` helper consistently.

**Rationale:** Handlers should focus on orchestration, leaving cross-cutting concerns like authentication to middleware. This simplifies the handler logic.

---

## Action Items

- [ ] Refactor `domain/comment.go` to simplify `Validate()`.
- [ ] Fix `adapters/http_handler_api.go` to use public UUIDs for `user_id` in JSON.
- [ ] Consolidate JSON response structs to reduce duplication.
- [ ] Move template execution to use pre-parsed templates in `http_handler_page.go`.
- [ ] Implement `enrichComments` helper to centralize enrichment and prevent N+1 queries.
- [ ] Extract `scanComment` in `sqlite_repository.go` to reduce code duplication.

---

## Notes

The comment module is currently scaffolded. Implementing these simplifications now will establish a robust foundation for future features and ensure alignment with the reference `auth` implementation.
