# Go Code Simplifier Review

**Folder/Module:** reaction
**Date:** 2026-01-14 14:56
**Files Reviewed:**

- `internal/modules/reaction/application/service.go`
- `internal/modules/reaction/adapters/sqlite_repository.go`
- `internal/modules/reaction/adapters/http_handler_api.go`
- `internal/platform/health/checker.go`

---

## Summary

The reaction module implementation follows the established hexagonal architecture but contains significant redundancy and non-idiomatic Go patterns (specifically for Go 1.22+). The primary areas for improvement are simplifying target resolution logic, leveraging modern `net/http` features, and streamlining cross-module dependency usage. Additionally, the health checker can be slightly refined for better parameter handling.

---

## Findings

### 1. Non-Idiomatic Path Parameter Extraction

**File:** `internal/modules/reaction/adapters/http_handler_api.go`
**Line(s):** 187-195, 236-244
**Category:** Idiomatic Go
**Severity:** Medium

**Current Code:**

```go
// Extract target type and target ID from path
pathParts := strings.Split(r.URL.Path, "/")
if len(pathParts) < 4 {
    h.logger.Error("Invalid path for getting reactions", logger.String("path", r.URL.Path))
    http.Error(w, "Invalid path", http.StatusBadRequest)
    return
}

targetType := pathParts[len(pathParts)-2]
targetID := pathParts[len(pathParts)-1]
```

**Suggested Improvement:**

```go
targetType := r.PathValue("targetType")
targetID := r.PathValue("targetId")

if targetType == "" || targetID == "" {
    h.logger.Error("Missing path parameters", logger.String("path", r.URL.Path))
    http.Error(w, "Invalid path parameters", http.StatusBadRequest)
    return
}
```

**Rationale:** The project uses Go 1.24+ and the routes are registered using wildcards (e.g., `{targetType}`). Using `r.PathValue` is the idiomatic way to extract these parameters since Go 1.22, making the code cleaner and less prone to errors than manual string splitting.

---

### 2. Redundant Target Resolution Logic

**File:** `internal/modules/reaction/adapters/sqlite_repository.go`
**Line(s):** 34-42, 72-87, 120-135, 187-202, 250-265
**Category:** KISS Violation / DRY
**Severity:** Medium

**Current Code:**

```go
var targetID int
switch reaction.TargetType {
case "post":
    err = r.db.QueryRowContext(ctx, "SELECT id FROM posts WHERE public_id = ?", reaction.PublicTargetID).Scan(&targetID)
case "comment":
    err = r.db.QueryRowContext(ctx, "SELECT id FROM comments WHERE public_id = ?", reaction.PublicTargetID).Scan(&targetID)
default:
    return domain.ErrInvalidTarget
}
```

**Suggested Improvement:**

```go
func (r *SQLiteReactionRepository) resolveTargetID(ctx context.Context, publicID string, targetType string) (int, error) {
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

**Rationale:** Every method in the repository repeats the same logic to convert a public UUID to an internal integer ID. Extracting this into a private helper method significantly reduces code duplication and makes the repository easier to maintain.

---

### 3. Redundant Existence Checks in Service

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 55-66, 129-140, 163-174, 236-247
**Category:** Architecture / Performance
**Severity:** Low

**Current Code:**

```go
// Validate that the target exists
switch targetType {
case "post":
    _, err := s.postRepo.GetByID(ctx, targetPublicID)
    if err != nil {
        return err
    }
case "comment":
    _, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
    if err != nil {
        return err
    }
}
```

**Suggested Improvement:**
Rely on the repository's `resolveTargetID` (which already checks for existence) or database foreign key constraints. If the repository operation fails because the target doesn't exist, it will return `domain.ErrReactionNotFound` anyway.

**Rationale:** The service layer currently performs a full database round-trip just to verify existence, and then the repository performs _another_ round-trip to get the internal ID. This doubles the database load for every reaction operation.

---

### 4. Fragile Background Operations

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 110-112, 148-150
**Category:** Concurrency / Reliability
**Severity:** High

**Current Code:**

```go
go func() {
    _ = s.userService.IncrementReactionCount(context.Background(), userID)
}()
```

**Suggested Improvement:**

```go
// Perform within the same request lifecycle if possible, or use a managed worker pool
err = s.userService.IncrementReactionCount(ctx, userID)
if err != nil {
    s.logger.Error("Failed to increment reaction count", logger.Int("user_id", userID), logger.Error(err))
}
```

**Rationale:** Using `context.Background()` and ignoring errors in a detached goroutine is dangerous. If the update fails, the data becomes permanently inconsistent. In a high-integrity system, these count updates should ideally be transactional or at least logged and retried.

---

### 5. Repetitive User Fetching in API Handler

**File:** `internal/modules/reaction/adapters/http_handler_api.go`
**Line(s):** 41-48, 125-132
**Category:** Performance
**Severity:** Low

**Current Code:**

```go
user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
if err != nil {
    // ... error handling
}
userID := user.ID
```

**Suggested Improvement:**
Store the internal `UserID` in the context during authentication if it is frequently needed, or allow the service to accept `PublicUserID` and resolve it internally if necessary.

**Rationale:** The API handler performs an extra database query on every reaction request just to translate the public ID to an internal ID, even though the user was already verified by the middleware.

---

### 6. Fragile Route Parameter Replacement in Health Checker

**File:** `internal/platform/health/checker.go`
**Line(s):** 134-143
**Category:** KISS Violation / Robustness
**Severity:** Low

**Current Code:**

```go
if strings.Contains(path, "{") && strings.Contains(path, "}") {
    // Handle common parameter names in routes
    testPath = strings.Replace(testPath, "{id}", "1", -1)
    testPath = strings.Replace(testPath, "{postId}", "1", -1)
    testPath = strings.Replace(testPath, "{targetType}", "post", -1)
    testPath = strings.Replace(testPath, "{targetId}", "1", -1)
    // Remove any remaining brackets that weren't matched by the specific replacements
    testPath = strings.ReplaceAll(testPath, "{", "1") // fallback for other parameter names
    testPath = strings.ReplaceAll(testPath, "}", "")
}
```

**Suggested Improvement:**

```go
re := regexp.MustCompile(`\{[^}]+\}`)
testPath = re.ReplaceAllString(path, "1")
// Then handle specific type-safe values if needed, or just use "1" for all checks
```

**Rationale:** The current approach uses multiple string replacements and a final `ReplaceAll` that might leave the path in an invalid state if parameters aren't strictly formatted. A regular expression is more robust for identifying and replacing any `{parameter}` block.

---

## Action Items

- [ ] Refactor `http_handler_api.go` to use `r.PathValue()` for parameter extraction.
- [ ] Implement `resolveTargetID` helper in `sqlite_repository.go` to eliminate code duplication.
- [ ] Remove redundant existence checks in `service.go` and let the repository handle it.
- [ ] Update background counter increments to use the request context and handle potential errors.
- [ ] Simplify route parameter replacement in `health/checker.go` using regular expressions.
- [ ] Consider adding a `writeErrorJSON` helper to the HTTP handler to standardize error responses.

---

## Notes

While the code is functional, these simplifications will make it more idiomatic and performant, especially as the system scales. The reduction in database round-trips for existence checks will have a noticeable impact under high load.
