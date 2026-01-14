# Go Code Simplifier Review

**Folder/Module:** internal/modules (auth, user, post, comment, reaction, moderation, notification)
**Date:** 2026-01-14 15:16
**Files Reviewed:** 101 Go files across 7 modules

---

## Summary

This review analyzes all 7 modules in the `internal/modules` directory. The codebase demonstrates **strong adherence** to the project's modular monolith architecture with consistent structure across modules. Overall code quality is **good**, with consistent patterns for domain entities, ports/adapters separation, and HTTP handlers.

**Key Observations:**

- ✅ Excellent module isolation with clear boundaries
- ✅ Consistent 4-directory layout (domain, ports, application, adapters)
- ✅ Proper ID security (INT internal, UUID public) across all modules
- ⚠️ Some code duplication in validation patterns and helper functions
- ⚠️ Inconsistent error wrapping practices
- ⚠️ Several deprecated functions still present
- ⚠️ Minor KISS violations in some handlers

---

## Findings

### 1. Duplicated `min` Function Definition

**File:** `internal/modules/post/adapters/http_handler.go`
**Line(s):** 199-204
**Category:** KISS Violation
**Severity:** Low

**Current Code:**

```go
// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

**Suggested Improvement:**

```go
// Remove this function entirely - use built-in min() from Go 1.21+
```

**Rationale:** Go 1.21+ includes built-in `min()` and `max()` functions. Since the project uses Go 1.24+, this helper is redundant. The function is also unused in the file.

---

### 2. Redundant Content Validation in Comment

**File:** `internal/modules/comment/domain/comment.go`
**Line(s):** 24-38
**Category:** KISS Violation
**Severity:** Low

**Current Code:**

```go
func (c *Comment) Validate() error {
	// Check content is not empty
	if c.Content == "" {
		return ErrEmptyContent
	}

	// Check content is not just whitespace
	if len(c.Content) == 0 || len([]rune(c.Content)) == 0 || len(strings.TrimSpace(c.Content)) == 0 {
		return ErrEmptyContent
	}

	// Check content length limits (max 5000 characters)
	if len([]rune(c.Content)) > 5000 {
		return ErrContentTooLong
	}

	return nil
}
```

**Suggested Improvement:**

```go
func (c *Comment) Validate() error {
	// Check content is not empty or just whitespace
	trimmed := strings.TrimSpace(c.Content)
	if trimmed == "" {
		return ErrEmptyContent
	}

	// Check content length limits (max 5000 characters)
	if len([]rune(c.Content)) > 5000 {
		return ErrContentTooLong
	}

	return nil
}
```

**Rationale:** The first `if c.Content == ""` check is redundant since the second check already handles empty content. Also, `len(c.Content) == 0` is functionally equivalent to `c.Content == ""`. The simplified version is clearer and does the same job.

---

### 3. Deprecated Functions Should Be Removed

**File:** `internal/modules/auth/adapters/middleware.go`
**Line(s):** 64-139
**Category:** Architecture
**Severity:** Medium

**Current Code:**

```go
// RequireAuthFunc is a standalone function for backward compatibility.
// Prefer using MiddlewareProvider.RequireAuth() for new code.
// DEPRECATED: Use AuthMiddleware instead.
func RequireAuth(authService authPorts.AuthService, userService userPorts.UserService) func(http.Handler) http.Handler {
	provider := NewAuthMiddleware(authService, userService)
	return provider.RequireAuth()
}

// ... similar for OptionalAuthFunc, GetUserID, GetUsername, IsAuthenticated
```

**Suggested Improvement:**
Either:

1. Remove these deprecated functions if not used elsewhere
2. Or, if still used, update all call sites to use the new pattern and then remove

```go
// Remove deprecated functions after updating all callers to use:
// - AuthMiddleware.RequireAuth()
// - AuthMiddleware.OptionalAuth()
// - authPorts.GetUserID(ctx)
// - authPorts.GetUsername(ctx)
// - authPorts.IsAuthenticated(ctx)
```

**Rationale:** KISS principle - deprecated code adds maintenance burden and confusion. These wrapper functions just delegate to the new pattern and should be removed once all call sites are updated.

---

### 4. Inconsistent Error Wrapping Pattern

**File:** `internal/modules/post/adapters/sqlite_repository.go`
**Line(s):** 27-89
**Category:** Idiomatic Go
**Severity:** Medium

**Current Code:**

```go
func (r *SQLitePostRepository) Create(ctx context.Context, post *domain.Post) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate public UUID
	publicID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}
	// ...
}
```

**File:** `internal/modules/comment/adapters/sqlite_repository.go`
**Line(s):** 25-41

**Current Code (Inconsistent):**

```go
func (r *SQLiteCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	// Generate UUID for PublicID
	u, err := uuid.NewV4()
	if err != nil {
		return err  // Not wrapped
	}
	// ...
	_, err = r.db.ExecContext(ctx, `...`, ...)
	return err  // Not wrapped
}
```

**Suggested Improvement:**

```go
func (r *SQLiteCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	u, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}
	comment.PublicID = u.String()

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO comments (public_id, post_id, author_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, comment.PublicID, comment.PostID, comment.UserID, comment.Content)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	return nil
}
```

**Rationale:** Consistent error wrapping with `%w` verb improves debugging by providing context. The post repository wraps errors, but the comment repository does not. All repositories should follow the same pattern.

---

### 5. Goroutine Without Error Logging

**File:** `internal/modules/post/application/service.go`
**Line(s):** 131-133, 196-198
**Category:** Concurrency
**Severity:** Medium

**Current Code:**

```go
// Increment user's post count asynchronously (non-blocking)
go func() {
	_ = s.userService.IncrementPostCount(context.Background(), userID)
}()
```

**Suggested Improvement:**

```go
// Increment user's post count asynchronously (non-blocking)
go func() {
	if err := s.userService.IncrementPostCount(context.Background(), userID); err != nil {
		// Log error - silently failing counter increments can cause data inconsistency
		// Consider using a structured logger injected via dependency
	}
}()
```

**Rationale:** While the comment says "non-blocking", silently ignoring errors can lead to user stats being incorrect without any indication. At minimum, errors should be logged. Consider injecting a logger into the service.

---

### 6. Handler Logger Creation in Hot Path

**File:** `internal/modules/post/adapters/http_handler_api.go`
**Line(s):** 113-122, 254-264
**Category:** Performance
**Severity:** Low

**Current Code:**

```go
func (h *HTTPHandler) CreatePostAPI(w http.ResponseWriter, r *http.Request) {
	// ...
	case strings.HasPrefix(contentType, "application/json"):
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Log decode error to terminal
			cfg := &logger.Config{
				TimePrecision: logger.TimePrecisionSeconds,
				AllowedFields: []string{"url", "error", "errors"},
				MaxLineWidth:  120,
			}
			l := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
			l.Error("http.request.error",
				logger.String("url", r.URL.RequestURI()),
				logger.String("error", err.Error()),
			)
			// ...
		}
	// ...
}
```

**Suggested Improvement:**

```go
// In HTTPHandler struct, add a logger field:
type HTTPHandler struct {
	// ... existing fields
	logger *logger.Logger  // Inject at construction time
}

// Then use it in handlers:
func (h *HTTPHandler) CreatePostAPI(w http.ResponseWriter, r *http.Request) {
	// ...
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("http.request.error",
			logger.String("url", r.URL.RequestURI()),
			logger.String("error", err.Error()),
		)
		// ...
	}
	// ...
}
```

**Rationale:** Creating a new logger configuration on every error is wasteful. The logger should be injected via dependency injection and reused. This is already done correctly in the reaction handler.

---

### 7. Verbose String Building Pattern

**File:** `internal/modules/post/adapters/http_handler.go`
**Line(s):** 170-179
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
// Join parts with spaces
title := ""
for i, part := range parts {
	if i > 0 {
		title += " "
	}
	title += part
}

return title
```

**Suggested Improvement:**

```go
return strings.Join(parts, " ")
```

**Rationale:** Go's `strings.Join` is more idiomatic and efficient than manual string concatenation in a loop.

---

### 8. Repetitive Target Type Validation

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** Multiple (46-48, 55-66, 124-126, 129-140, etc.)
**Category:** KISS Violation
**Severity:** Medium

**Current Code:**

```go
// In React():
if targetType != "post" && targetType != "comment" {
	return domain.ErrInvalidTarget
}
// Validate that the target exists
switch targetType {
case "post":
	_, err := s.postRepo.GetByID(ctx, targetPublicID)
	if err != nil { return err }
case "comment":
	_, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
	if err != nil { return err }
}

// Same pattern repeated in RemoveReaction(), GetReactions(), CountReactions(), etc.
```

**Suggested Improvement:**

```go
// Extract to a helper method
func (s *Service) validateTarget(ctx context.Context, targetPublicID, targetType string) error {
	switch targetType {
	case "post":
		_, err := s.postRepo.GetByID(ctx, targetPublicID)
		return err
	case "comment":
		_, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
		return err
	default:
		return domain.ErrInvalidTarget
	}
}

// Then use in all methods:
func (s *Service) React(ctx context.Context, userID int, targetPublicID string, targetType string, reactionType domain.ReactionType) error {
	if err := s.validateTarget(ctx, targetPublicID, targetType); err != nil {
		return err
	}
	// ... rest of the logic
}
```

**Rationale:** The same validation logic is repeated 5+ times. Extracting it into a helper method reduces duplication and makes the code easier to maintain.

---

### 9. Missing `rows.Err()` Check

**File:** `internal/modules/comment/adapters/sqlite_repository.go`
**Line(s):** 91-123
**Category:** Error Handling
**Severity:** Medium

**Current Code:**

```go
func (r *SQLiteCommentRepository) ListByPostPublicID(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `...`, postPublicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		// ... scan logic
	}

	return comments, nil  // Missing rows.Err() check!
}
```

**Suggested Improvement:**

```go
func (r *SQLiteCommentRepository) ListByPostPublicID(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `...`, postPublicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		// ... scan logic
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}
```

**Rationale:** The `rows.Err()` check is essential for detecting errors that occur during row iteration. The post repository correctly includes this check, but the comment repository does not.

---

### 10. Placeholder TODO Methods Should Be Marked More Clearly

**File:** `internal/modules/moderation/application/service.go`
**Line(s):** 23-30, 34-40
**Category:** Architecture
**Severity:** Low

**Current Code:**

```go
// CreateReport creates a new moderation report.
// TODO: Implement report creation with validation.
func (s *Service) CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) error {
	// Implementation placeholder
	// 1. Validate target type and reason
	// 2. Resolve targetPublicID to internal target ID
	// 3. Create report entity with "pending" status
	// 4. Save to repository (repo generates PublicID)
	return nil
}
```

**Suggested Improvement:**

```go
// CreateReport creates a new moderation report.
func (s *Service) CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) error {
	// TODO: [OPTIONAL FEATURE: forum-moderation] Implement report creation
	return domain.ErrNotImplemented  // Or panic("not implemented")
}
```

**Rationale:** Returning `nil` silently succeeds, which could lead to unexpected behavior if this method is called. Either return an explicit error or panic to make it clear the feature is not implemented.

---

### 11. HasPermission Method Not Implemented

**File:** `internal/modules/user/domain/user.go`
**Line(s):** 44-48
**Category:** Architecture
**Severity:** Medium

**Current Code:**

```go
// HasPermission checks if the user has permission for an action based on their role.
// TODO: Implement permission logic.
func (u *User) HasPermission(action string) bool {
	// Implementation placeholder
	// Define permissions for each role
	return false
}
```

**Suggested Improvement:**
Either implement the method properly or remove it if not needed:

```go
// HasPermission checks if the user has permission for an action based on their role.
func (u *User) HasPermission(action string) bool {
	switch action {
	case "create_post", "create_comment", "react":
		return u.Role == RoleUser || u.Role == RoleModerator || u.Role == RoleAdmin
	case "moderate":
		return u.CanModerate()
	case "admin":
		return u.IsAdmin()
	default:
		return false
	}
}
```

**Rationale:** An unimplemented method that always returns `false` can lead to unexpected access denials if used. Either implement it or remove it until it's needed.

---

### 12. Path Parsing Instead of Using `r.PathValue()`

**File:** `internal/modules/reaction/adapters/http_handler_api.go`
**Line(s):** 185-201, 236-250
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
func (h *HTTPHandler) GetReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Extract target type and target ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.Error("Invalid path for getting reactions", logger.String("path", r.URL.Path))
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetType := pathParts[len(pathParts)-2]
	targetID := pathParts[len(pathParts)-1]
	// ...
}
```

**Suggested Improvement:**

```go
func (h *HTTPHandler) GetReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Use Go 1.22+ PathValue for clean path parameter extraction
	targetType := r.PathValue("targetType")
	targetID := r.PathValue("targetId")

	if targetType == "" || targetID == "" {
		h.logger.Error("Missing path parameters", logger.String("path", r.URL.Path))
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	// ...
}
```

**Rationale:** The post handler uses `r.PathValue()` (Go 1.22+ feature) correctly, but the reaction handler manually parses paths. This should be consistent across all handlers.

---

### 13. Duplicate Response Struct Declarations

**File:** `internal/modules/auth/adapters/http_handler_api.go`
**Line(s):** 86-98, 141-153
**Category:** KISS Violation
**Severity:** Low

**Current Code:**

```go
// In RegisterAPI:
resp := struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
}{...}

// In LoginAPI (same struct):
resp := struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
}{...}
```

**Suggested Improvement:**

```go
// Define once at package level or in a shared location
type authResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// Use in both handlers:
resp := authResponse{
	ID:       user.PublicID,
	UserID:   user.PublicID,
	Email:    req.Email,
	Username: req.Username,
	Token:    session.Token,
}
```

**Rationale:** Identical anonymous structs should be extracted into a named type to reduce duplication and improve maintainability.

---

## Action Items

- [ ] **High Priority:**

  - [ ] Add `rows.Err()` check to `ListByPostPublicID` in comment repository
  - [ ] Add error logging to goroutines in post/comment/reaction services
  - [ ] Standardize error wrapping across all SQLite repositories

- [ ] **Medium Priority:**

  - [ ] Extract target validation helper in reaction service to reduce duplication
  - [ ] Inject logger into post handler instead of creating in hot path
  - [ ] Remove deprecated middleware functions in auth after updating callers
  - [ ] Implement or remove `HasPermission()` method

- [ ] **Low Priority:**
  - [ ] Remove custom `min()` function (use built-in)
  - [ ] Simplify comment validation logic
  - [ ] Use `strings.Join()` for title building
  - [ ] Use `r.PathValue()` consistently in reaction handlers
  - [ ] Extract duplicate auth response struct
  - [ ] Make TODO methods return explicit errors

---

## Notes

### Positive Patterns Observed

1. **Consistent Module Structure**: All 7 modules follow the exact same 4-directory layout (`domain/`, `ports/`, `application/`, `adapters/`), making navigation and understanding easy.

2. **Proper ID Security**: All modules correctly use internal INT IDs for database operations and public UUIDs for API responses. JSON tags consistently hide internal IDs with `json:"-"`.

3. **Good Transaction Usage**: The post repository correctly uses transactions with `defer tx.Rollback()` for multi-step operations.

4. **Clean Service Interfaces**: Port interfaces are well-defined with clear method signatures and documentation.

5. **Proper HTTP Patterns**: Handlers correctly use Go 1.22+ routing patterns (`"GET /api/posts/{id}"`) in most places.

### Modules Status Summary

| Module       | Status        | Notes                                 |
| ------------ | ------------- | ------------------------------------- |
| auth         | ✅ Complete   | Reference implementation, well-tested |
| user         | ✅ Complete   | Good implementation with cached stats |
| post         | ✅ Complete   | Comprehensive with image handling     |
| comment      | ✅ Complete   | Solid implementation                  |
| reaction     | ✅ Complete   | Good toggle behavior implementation   |
| moderation   | ⚠️ Scaffolded | TODOs present, marked as optional     |
| notification | ⚠️ Scaffolded | TODOs present, marked as optional     |

### Recommendations for New Features

When implementing the scaffolded moderation and notification modules:

1. Follow the error wrapping pattern from the post repository
2. Include `rows.Err()` checks in all list methods
3. Use injected loggers instead of creating them per-request
4. Implement proper validation methods instead of returning false/nil
