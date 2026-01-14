# Go Code Simplifier Review

**Folder/Module:** reaction
**Date:** 2026-01-14 14:56
**Files Reviewed:**

- `internal/modules/reaction/domain/reaction.go`
- `internal/modules/reaction/domain/errors.go`
- `internal/modules/reaction/domain/reaction_test.go`
- `internal/modules/reaction/ports/service.go`
- `internal/modules/reaction/ports/repository.go`
- `internal/modules/reaction/ports/service_test.go`
- `internal/modules/reaction/application/service.go`
- `internal/modules/reaction/application/service_test.go`
- `internal/modules/reaction/adapters/http_handler.go`
- `internal/modules/reaction/adapters/http_handler_api.go`
- `internal/modules/reaction/adapters/sqlite_repository.go`

---

## Summary

The reaction module is well-structured overall, following the project's hexagonal architecture with clear separation between domain, ports, application, and adapters. However, several opportunities for simplification and improvement were identified:

1. **Concurrency concerns** with fire-and-forget goroutines that silently swallow errors
2. **Code duplication** in target validation across multiple methods
3. **Inconsistent error handling** patterns between API handlers
4. **Stale mock implementations** in ports test file that don't match current interface signatures
5. **Missing Content-Type headers** in some JSON responses
6. **URL path parsing fragility** that could be improved with stdlib features

---

## Findings

### 1. Fire-and-Forget Goroutines Silently Swallow Errors

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 105-107, 147-149
**Category:** Concurrency
**Severity:** Medium

**Current Code:**

```go
// Increment user's reaction count asynchronously (non-blocking)
go func() {
	_ = s.userService.IncrementReactionCount(context.Background(), userID)
}()
```

**Suggested Improvement:**

```go
// Increment user's reaction count asynchronously (non-blocking)
// Note: Error is logged but not returned as this is a non-critical side effect
go func() {
	if err := s.userService.IncrementReactionCount(context.Background(), userID); err != nil {
		// Consider adding structured logging here
		// log.Printf("failed to increment reaction count for user %d: %v", userID, err)
	}
}()
```

**Rationale:** While fire-and-forget for non-critical operations is acceptable, silently ignoring errors with `_` masks potential issues. At minimum, errors should be logged for debugging purposes. Consider injecting a logger into the service and logging failures. Alternatively, if the reaction count must be accurate, consider making this synchronous or using a background worker queue.

---

### 2. Repeated Target Validation Logic

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 46-47, 120-121, 159-160, 184-185, 231-232
**Category:** KISS Violation
**Severity:** Low

**Current Code:**

```go
if targetType != "post" && targetType != "comment" {
	return domain.ErrInvalidTarget
}
```

**Suggested Improvement:**

```go
// Add a helper method to the service or domain package
func isValidTargetType(targetType string) bool {
	return targetType == "post" || targetType == "comment"
}

// Usage in service methods:
if !isValidTargetType(targetType) {
	return domain.ErrInvalidTarget
}
```

**Rationale:** The same validation logic is repeated 5+ times across service methods. Extracting this to a helper function centralizes the valid target types in one location, making future additions (e.g., "poll", "thread") a single-point change. This follows DRY principle and improves maintainability.

---

### 3. Repeated Target Existence Verification Pattern

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 54-65, 127-140, 163-176, 190-201, 237-248
**Category:** KISS Violation
**Severity:** Medium

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

```go
// Add a private helper method to the service
func (s *Service) validateTargetExists(ctx context.Context, targetPublicID, targetType string) error {
	switch targetType {
	case "post":
		if _, err := s.postRepo.GetByID(ctx, targetPublicID); err != nil {
			return err
		}
	case "comment":
		if _, err := s.commentRepo.GetByPublicID(ctx, targetPublicID); err != nil {
			return err
		}
	default:
		return domain.ErrInvalidTarget
	}
	return nil
}

// Usage:
if err := s.validateTargetExists(ctx, targetPublicID, targetType); err != nil {
	return err
}
```

**Rationale:** This exact switch pattern appears 5 times in the service. Extracting it to a helper method reduces duplication by ~50 lines and ensures consistent error handling across all methods. The helper can also include the target type validation, consolidating both checks.

---

### 4. Inconsistent JSON Response Formatting

**File:** `internal/modules/reaction/adapters/http_handler_api.go`
**Line(s):** 108-109
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
// Return success
w.WriteHeader(http.StatusOK)
fmt.Fprintf(w, `{"message": "Reaction added successfully"}`)
```

**Suggested Improvement:**

```go
// Return success - use consistent JSON encoding
h.writeJSON(w, http.StatusOK, map[string]string{
	"message": "Reaction added successfully",
})
```

**Rationale:** The handler already has a `writeJSON` helper method that sets the `Content-Type` header properly, but this success response uses `fmt.Fprintf` instead. This means the response is missing the `Content-Type: application/json` header. Using the existing helper ensures consistent headers and encoding across all responses.

---

### 5. Stale Mock Implementations in Ports Tests

**File:** `internal/modules/reaction/ports/service_test.go`
**Line(s):** 29-56
**Category:** TDD
**Severity:** High

**Current Code:**

```go
type mockReactionService struct{}

func (m *mockReactionService) React(ctx context.Context, userID, targetID int, targetType string, reactionType domain.ReactionType) error {
	return nil
}

func (m *mockReactionService) RemoveReaction(ctx context.Context, userID, targetID int, targetType string) error {
	return nil
}

func (m *mockReactionService) GetReactions(ctx context.Context, targetID int, targetType string) ([]*domain.Reaction, error) {
	return nil, nil
}
```

**Suggested Improvement:**

```go
type mockReactionService struct{}

func (m *mockReactionService) React(ctx context.Context, userID int, targetPublicID string, targetType string, reactionType domain.ReactionType) error {
	return nil
}

func (m *mockReactionService) RemoveReaction(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	return nil
}

func (m *mockReactionService) GetReactions(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	return nil, nil
}

func (m *mockReactionService) CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	return 0, 0, nil
}

func (m *mockReactionService) GetUserReactionCount(ctx context.Context, userID int) (int, error) {
	return 0, nil
}

func (m *mockReactionService) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error) {
	return nil, nil
}
```

**Rationale:** The mock implementations in `ports/service_test.go` use the old interface signatures (`targetID int` instead of `targetPublicID string`) and are missing newer methods (`GetUserReactionCount`, `GetByUserAndTargetPublicID`). This causes the tests to not actually verify the current interface contract. These mocks should match the actual `ReactionService` interface in `service.go`.

---

### 6. Fragile URL Path Parsing

**File:** `internal/modules/reaction/adapters/http_handler_api.go`
**Line(s):** 186-195, 236-244
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
// Extract target type and target ID using stdlib route parameters
targetType := r.PathValue("targetType")
targetID := r.PathValue("targetId")

if targetType == "" || targetID == "" {
	h.logger.Error("Missing path parameters for getting reactions", logger.String("path", r.URL.Path))
	http.Error(w, "Invalid path", http.StatusBadRequest)
	return
}
```

**Rationale:** Go 1.22+ introduced `r.PathValue()` for extracting named path parameters from routes registered with patterns like `GET /api/reactions/{targetType}/{targetId}`. The current manual string splitting is fragile (off-by-one errors with trailing slashes, inconsistent index calculations between `GetReactionsAPI` check for `< 4` and `CountReactionsAPI` check for `< 5`). Using the stdlib method is cleaner and more reliable.

---

### 7. Inconsistent Error Handling for Not Found vs Invalid Target

**File:** `internal/modules/reaction/adapters/sqlite_repository.go`
**Line(s):** 44-47
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
if err != nil {
	if err == sql.ErrNoRows {
		return domain.ErrReactionNotFound
	}
	return err
}
```

**Suggested Improvement:**

```go
if err != nil {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrTargetNotFound // Consider a more specific error
	}
	return fmt.Errorf("looking up target ID: %w", err)
}
```

**Rationale:** Two issues here: (1) Use `errors.Is()` for error comparison to handle wrapped errors correctly. (2) When the target post/comment doesn't exist, returning `ErrReactionNotFound` is semantically incorrect—the reaction wasn't found because the target doesn't exist. Consider adding `ErrTargetNotFound` to the domain errors for clarity.

---

### 8. Missing Error Context in Repository

**File:** `internal/modules/reaction/adapters/sqlite_repository.go`
**Line(s):** 60, 92, 109
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
_, err = r.db.ExecContext(ctx, query, reaction.PublicID, reaction.UserID, reaction.TargetID, reaction.TargetType, reaction.Type)
return err
```

**Suggested Improvement:**

```go
_, err = r.db.ExecContext(ctx, query, reaction.PublicID, reaction.UserID, reaction.TargetID, reaction.TargetType, reaction.Type)
if err != nil {
	return fmt.Errorf("inserting reaction: %w", err)
}
return nil
```

**Rationale:** Wrapping errors with context using `fmt.Errorf` and `%w` provides better debugging information when errors propagate up the call stack. This makes it easier to identify which database operation failed without having to trace through the code.

---

### 9. Potential Race Condition in Reaction Toggle

**File:** `internal/modules/reaction/application/service.go`
**Line(s):** 67-89
**Category:** Concurrency
**Severity:** Medium

**Current Code:**

```go
// Check if user already has a reaction on this target
existingReaction, err := s.reactionRepo.GetByUserAndTargetPublicID(
	ctx, userID, targetPublicID, targetType,
)
// ... time gap where another request could modify state ...
if existingReaction != nil {
	if existingReaction.Type == reactionType {
		return s.RemoveReaction(ctx, userID, targetPublicID, targetType)
	}
	err = s.reactionRepo.DeleteByTargetPublicID(ctx, userID, targetPublicID, targetType)
	// ...
}
// Create new reaction
err = s.reactionRepo.Create(ctx, reaction)
```

**Suggested Improvement:**

```go
// Consider using database transaction for atomicity
tx, err := s.reactionRepo.BeginTx(ctx)
if err != nil {
	return fmt.Errorf("starting transaction: %w", err)
}
defer tx.Rollback()

// Check if user already has a reaction on this target (within transaction)
existingReaction, err := tx.GetByUserAndTargetPublicID(ctx, userID, targetPublicID, targetType)
// ... rest of logic using tx ...

if err := tx.Commit(); err != nil {
	return fmt.Errorf("committing transaction: %w", err)
}
return nil
```

**Rationale:** The current implementation has a time-of-check to time-of-use (TOCTOU) race condition. If the same user sends two rapid reaction requests, both might see no existing reaction and both might attempt to create one, resulting in duplicate reactions or constraint violations. Wrapping the check-and-modify in a database transaction would ensure atomicity. This requires adding transaction support to the repository interface.

---

### 10. Domain Validation Could Be Stricter

**File:** `internal/modules/reaction/domain/reaction.go`
**Line(s):** 44-47
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
if r.TargetID <= 0 && r.PublicTargetID == "" {
	return ErrInvalidTargetID
}
```

**Suggested Improvement:**

```go
// Be explicit about what's required
if r.PublicTargetID == "" {
	return ErrInvalidTargetID
}
// TargetID can be 0 as it's resolved by the repository
```

**Rationale:** Since the architecture uses `PublicTargetID` (UUID) for all operations and `TargetID` (int) is resolved internally by the repository, the validation should focus on `PublicTargetID` being required. The current OR condition is confusing and allows either to be set, but the service layer only ever sets `PublicTargetID`. This could be simplified to match actual usage patterns.

---

## Action Items

- [ ] Add error logging to fire-and-forget goroutines in `application/service.go` (lines 105-107, 147-149)
- [ ] Extract `isValidTargetType()` helper to reduce code duplication
- [ ] Extract `validateTargetExists()` helper method to centralize target validation
- [ ] Fix JSON response in `AddReactionAPI` to use `writeJSON` helper for consistent `Content-Type` headers
- [ ] **CRITICAL:** Update mock implementations in `ports/service_test.go` to match current interface signatures
- [ ] Consider using `r.PathValue()` for path parameter extraction (Go 1.22+)
- [ ] Use `errors.Is()` for error comparisons in repository
- [ ] Add error wrapping with context using `fmt.Errorf(...: %w, err)`
- [ ] Evaluate adding transaction support to prevent race conditions in toggle behavior
- [ ] Simplify domain validation to focus on `PublicTargetID` requirement

---

## Notes

### Positive Observations

1. **Good module structure**: The reaction module follows the project's hexagonal architecture conventions with clear file headers indicating adapter types (INPUT/OUTPUT).

2. **Consistent ID security**: Public UUIDs are correctly used in API layer while internal IDs are kept hidden (following `GEMINI.md` guidelines).

3. **Comprehensive test coverage**: The `application/service_test.go` has good table-driven tests with proper mocks, though the ports tests need updating.

4. **Good separation of concerns**: HTTP handlers are properly split between base handler (`http_handler.go`) and API endpoints (`http_handler_api.go`).

### Architecture Compliance

The module correctly implements:

- ✅ Flat adapters directory structure
- ✅ File headers for port/adapter classification
- ✅ ServiceContainer pattern for dependency injection
- ✅ PublicID/UUID for external interfaces
- ✅ Proper import restrictions (domain has no project imports)

### Testing Recommendations

1. Consider adding integration tests for the HTTP handlers using `httptest`
2. Add test cases for error paths in `application/service_test.go`
3. Consider using fuzzing for input validation testing on the domain `Validate()` method
