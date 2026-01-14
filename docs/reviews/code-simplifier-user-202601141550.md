# Go Code Simplifier Review

**Folder/Module:** user
**Date:** 2026-01-14 15:50
**Files Reviewed:**

- `internal/modules/user/domain/user.go`
- `internal/modules/user/domain/errors.go`
- `internal/modules/user/ports/repository.go`
- `internal/modules/user/ports/service.go`
- `internal/modules/user/application/service.go`
- `internal/modules/user/adapters/sqlite_repository.go`
- `internal/modules/user/adapters/http_handler_api.go`

---

## Summary

The `user` module is well-structured and follows the project's hexagonal architecture. It correctly implements the internal INT/public UUID identifier pattern. However, there is significant code duplication in the SQLite repository and HTTP API handlers. The simplification focuses on extracting common database scanning logic, centralizing error handling in the API, and removing redundant validation.

---

## Findings

### 1. Repetitive Scanning Logic in SQLite Repository

**File:** `internal/modules/user/adapters/sqlite_repository.go`
**Line(s):** 66-207
**Category:** KISS Violation / Idiomatic Go
**Severity:** Medium

**Current Code:**
The methods `GetByID`, `GetByPublicID`, `GetByEmail`, and `GetByUsername` all contain near-identical logic for scanning result rows into the `domain.User` struct.

```go
	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)

	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.PostCount,
		&user.CommentCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&isActive,
	)
    // ... error handling and conversion ...
	user.IsActive = isActive == 1
```

**Suggested Improvement:**
Extract a private helper function `scanUser` to handle the scanning and conversion.

```go
func (r *SQLiteUserRepository) scanUser(scanner interface{ Scan(dest ...any) error }) (*domain.User, error) {
	var u domain.User
	var isActive int
	err := scanner.Scan(
		&u.ID, &u.PublicID, &u.Email, &u.Username, &u.PasswordHash,
		&u.Role, &u.PostCount, &u.CommentCount, &u.CreatedAt, &u.UpdatedAt, &isActive,
	)
	if err != nil {
		return nil, err
	}
	u.IsActive = isActive == 1
	return &u, nil
}
```

**Rationale:** Reduces code duplication, makes the repository easier to maintain, and ensures consistent behavior for all "Get" operations.

---

### 2. Redundant Role Validation

**File:** `internal/modules/user/adapters/http_handler_api.go` and `internal/modules/user/application/service.go`
**Line(s):** Handler: 85-90; Service: 126-133
**Category:** KISS Violation
**Severity:** Low

**Current Code:**
Both the HTTP handler and the Service validate the role.

```go
// Handler
role := domain.Role(req.Role)
if role != domain.RoleUser && role != domain.RoleModerator && role != domain.RoleAdmin {
    http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
    return
}

// Service
if !isValidRole(newRole) {
    return domain.ErrInvalidRole
}
```

**Suggested Improvement:**
Rely on the service to perform validation. The handler should only be responsible for parsing the request and mapping service errors to HTTP responses.

**Rationale:** Single source of truth for business rules (roles) should be in the domain or service layer, not duplicated in the delivery layer (HTTP).

---

### 3. Missing Query-Based Pagination

**File:** `internal/modules/user/adapters/http_handler_api.go`
**Line(s):** 47-63
**Category:** Architecture
**Severity:** Low

**Current Code:**
`ListUsersAPI` uses hardcoded pagination values.

```go
func (h *HTTPHandler) ListUsersAPI(w http.ResponseWriter, r *http.Request) {
	// Default pagination values
	offset := 0
	limit := 20
    // ...
```

**Suggested Improvement:**
Extract `offset` and `limit` from the query string with sensible defaults.

```go
func (h *HTTPHandler) ListUsersAPI(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	if offset < 0 { offset = 0 }
    // ...
```

**Rationale:** Provides functional pagination for API consumers.

---

### 4. Direct Public ID to Internal ID Mapping in Handlers

**File:** `internal/modules/user/adapters/http_handler_api.go`
**Line(s):** 93-100, 123-129, 148-154
**Category:** Architecture / KISS Violation
**Severity:** Medium

**Current Code:**
Handlers repeatedly fetch the user by Public ID just to get the Internal ID.

```go
	// Get user by public ID to get internal ID
	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	// Update role
	if err := h.userService.UpdateRole(r.Context(), user.ID, role); err != nil {
        // ...
```

**Suggested Improvement:**
Consider adding service methods that accept `publicID` directly, OR extract this "get or error" logic into a helper in the `HTTPHandler`. Even better, the `UserService` could expose methods that operate on Public IDs for external-facing operations.

**Rationale:** Reduces boilerplate in handlers and minimizes unnecessary database round-trips if optimized.

---

### 5. Proper Error Wrapping and Context

**File:** `internal/modules/user/adapters/sqlite_repository.go`
**Line(s):** Throughout
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**
Returns raw errors from `database/sql`.

```go
	if err != nil {
		return err
	}
```

**Suggested Improvement:**
Use `fmt.Errorf` with `%w` to provide context.

```go
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
```

**Rationale:** Better observability when errors bubble up to the log or error handler.

---

## Action Items

- [ ] Extract `scanUser` helper in `sqlite_repository.go`.
- [ ] Implement query string parsing for pagination in `ListUsersAPI`.
- [ ] Remove redundant role validation from `http_handler_api.go`.
- [ ] Create a `respondWithError(w, code, message)` helper in `http_handler.go` (or shared kernel) to standardize JSON error responses.
- [ ] Add error wrapping to database operations in the repository.
- [ ] Clean up redundant `TODO` comments in `application/service.go`.

---

## Notes

The `user` module is a core part of the system. While it's functional and mostly follows the patterns, these simplifications will make it more robust and easier to extend as more user-related features (like profile pages or complex permissions) are added.
