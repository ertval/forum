# Go Code Simplifier Review

**Folder/Module:** reaction
**Date:** 2026-01-14 15:45
**Files Reviewed:**

- `internal/modules/reaction/domain/reaction.go`
- `internal/modules/reaction/domain/errors.go`
- `internal/modules/reaction/ports/service.go`
- `internal/modules/reaction/ports/repository.go`
- `internal/modules/reaction/application/service.go`
- `internal/modules/reaction/adapters/http_handler.go`
- `internal/modules/reaction/adapters/http_handler_api.go`
- `internal/modules/reaction/adapters/sqlite_repository.go`

---

## Summary

The `reaction` module follows the project's hexagonal architecture and modular monolith principles well. It correctly handles the separation between internal (INT) and public (UUID) identifiers. However, there are several areas where code can be simplified by reducing redundancy, leveraging modern Go features, and consolidating validation logic.

---

## Findings

### 1. Repeated Target ID Resolution Logic

**File:** `internal/modules/reaction/adapters/sqlite_repository.go`
**Line(s):** 33-49, 71-87, 119-135, 186-202, 249-265
**Category:** KISS Violation | Persistence
**Severity:** Medium

**Current Code:**
The logic to convert a `targetPublicID` and `targetType` into an internal `targetID` is repeated in almost every method of the repository.

```go
// Get internal target ID based on targetPublicID and targetType
var targetID int
var query string
switch targetType {
case "post":
    query = "SELECT id FROM posts WHERE public_id = ?"
case "comment":
    query = "SELECT id FROM comments WHERE public_id = ?"
}

err := r.db.QueryRowContext(ctx, query, targetPublicID).Scan(&targetID)
if err != nil {
    if err == sql.ErrNoRows {
        return domain.ErrReactionNotFound
    }
    return err
}
```

**Suggested Improvement:**
Extract this into a private helper method within the repository.

```go
func (r *SQLiteReactionRepository) getTargetID(ctx context.Context, publicID string, targetType string) (int, error) {
	var query string
	switch targetType {
	case "post":
		query = "SELECT id FROM posts WHERE public_id = ?"
	case "comment":
		query = "SELECT id FROM comments WHERE public_id = ?"
	default:
		return 0, domain.ErrInvalidTarget
	}

	var id int
	err := r.db.QueryRowContext(ctx, query, publicID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrReactionNotFound
		}
		return 0, err
	}
	return id, nil
}
```

**Rationale:** Reduces code duplication, improves maintainability, and ensures consistent error handling for target resolution across all repository methods.

---

### 2. Manual Path Parsing instead of PathValue

**File:** `internal/modules/reaction/adapters/http_handler_api.go`
**Line(s):** 187-196, 236-245
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**
The code manually splits the URL path to extract parameters, despite using wildcard patterns in the router registration.

```go
// Extract target type and target ID from path
pathParts := strings.Split(r.URL.Path, "/")
// ... validation ...
targetType := pathParts[len(pathParts)-2]
targetID := pathParts[len(pathParts)-1]
```

**Suggested Improvement:**
Use `r.PathValue()` which was introduced in Go 1.22 and is more robust.

```go
targetType := r.PathValue("targetType")
targetID := r.PathValue("targetId")
```

**Rationale:** Leveraging standard library features for path parameter extraction is safer, more readable, and aligns with modern Go practices.

---

### 3. Redundant and Distributed Validation

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 41-53, 119-126, 157-160, 181-184, 217-219, 227-233
**Category:** KISS Violation
**Severity:** Low

**Current Code:**
Validation of `userID`, `targetType`, and `reactionType` is repeated in multiple service methods.

**Suggested Improvement:**
Consolidate validation into small helper functions or leverage the `domain.Reaction.Validate()` method more effectively. While manual validation is fine for simple fields, repeated checks for things like `targetType != "post" && targetType != "comment"` should be centralized.

**Rationale:** Centralizing validation logic makes it easier to update (e.g., adding a new target type) and ensures consistency across the service layer.

---

### 4. Unchecked JSON Encoding Error

**File:** `internal/modules/reaction/adapters/http_handler_api.go`
**Line(s):** 297-301
**Category:** Error Handling
**Severity:** Low

**Current Code:**

```go
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
```

**Suggested Improvement:**
Check the error from `Encode`. Although unlikely to fail for simple structs, it's good practice.

```go
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", logger.Error(err))
	}
}
```

**Rationale:** Never ignore errors. Logging an encoding failure can help debug issues with response structures or circular dependencies.

---

### 5. Goroutine Context Usage

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 110-112, 148-150
**Category:** Concurrency | Cleanup
**Severity:** Low

**Current Code:**

```go
go func() {
    _ = s.userService.IncrementReactionCount(context.Background(), userID)
}()
```

**Suggested Improvement:**
While using a background context is necessary if the original request context might be cancelled (as is the case with HTTP requests), it's often better to use a "detachable" context that preserves trace IDs, or at least acknowledge why `Background()` is used. Additionally, consider if these updates absolutely must be async or if they could be part of the transaction (if the DB supported cross-module transactions, which it doesn't here).

**Rationale:** Discarding context values can break observability (tracing). If the project has a standard way to spawn "fire-and-forget" tasks with context propagation, it should be used.

---

## Action Items

- [ ] Refactor `internal/modules/reaction/adapters/sqlite_repository.go` to use a `getTargetID` helper method.
- [ ] Update `internal/modules/reaction/adapters/http_handler_api.go` to use `r.PathValue()` for extracting URL parameters.
- [ ] Add error checking and logging to the `writeJSON` helper in `http_handler_api.go`.
- [ ] Centralize common validation checks in `internal/modules/reaction/application/service.go`.
- [ ] Review the use of raw goroutines for user count updates; consider if a simple internal event system or worker pool is needed as the application grows.

---

## Notes

The module is overall very clean and follows the "INT internally, UUID publicly" rule strictly. The suggested improvements are primarily focused on removing boilerplate and using more idiomatic Go 1.22+ features.
