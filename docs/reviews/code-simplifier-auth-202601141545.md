# Go Code Simplifier Review

**Folder/Module:** auth
**Date:** 2026-01-14 15:45
**Files Reviewed:**

- `internal/modules/auth/application/service.go`
- `internal/modules/auth/adapters/http_handler_api.go`
- `internal/modules/auth/adapters/http_handler.go`
- `internal/modules/auth/adapters/middleware.go`
- `internal/modules/auth/domain/session.go`

---

## Summary

The `auth` module serves as the project's reference implementation and is already well-structured. It correctly implements the Hexagonal Architecture and adheres to the "INT internally, UUID publicly" security requirement. However, there are opportunities to simplify the code by reducing duplication in session management logic, refining error mapping in HTTP handlers, and consolidating shared HTTP utilities.

---

## Findings

### 1. Refactor Duplicated Session Validation and Cleanup

**File:** `internal/modules/auth/application/service.go`
**Line(s):** 184-246
**Category:** KISS Violation | DRY Principle
**Severity:** Low

**Current Code:**

```go
// ValidateSession snippet (repeated in RefreshSession and GetSession)
if session.IsExpired() {
    // Clean up expired session
    _ = s.sessionRepo.Delete(ctx, sessionToken) // Best effort cleanup
    return nil, domain.ErrSessionExpired
}
```

**Suggested Improvement:**

```go
// Private helper to consolidate check and cleanup
func (s *Service) getAndValidateSession(ctx context.Context, token string) (*domain.Session, error) {
    session, err := s.sessionRepo.GetByToken(ctx, token)
    if err != nil {
        return nil, err
    }
    if session.IsExpired() {
        _ = s.sessionRepo.Delete(ctx, token)
        return nil, domain.ErrSessionExpired
    }
    return session, nil
}
```

**Rationale:** This logic is repeated three times. Extracting it into a helper method reduces duplication and ensures consistent behavior for session cleanup.

---

### 2. Simplify HTTP API Error Mapping

**File:** `internal/modules/auth/adapters/http_handler_api.go`
**Line(s):** 41-64
**Category:** Idiomatic Go | KISS Violation
**Severity:** Medium

**Current Code:**

```go
if err != nil {
    switch {
    case errors.Is(err, authDomain.ErrInvalidEmail), ...:
        platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
    case errors.Is(err, authDomain.ErrEmailAlreadyExists), ...:
        platformErrors.WriteErrorJSON(w, http.StatusConflict, err.Error())
    default:
        errMsg := err.Error()
        if strings.Contains(errMsg, "empty") || ... {
            platformErrors.WriteErrorJSON(w, http.StatusBadRequest, errMsg)
        } else if ... {
            platformErrors.WriteErrorJSON(w, http.StatusConflict, errMsg)
        } else {
            platformErrors.WriteErrorJSON(w, http.StatusConflict, errMsg)
        }
    }
    return
}
```

**Suggested Improvement:**

```go
if err != nil {
    status := http.StatusInternalServerError
    switch {
    case errors.Is(err, authDomain.ErrInvalidEmail),
         errors.Is(err, authDomain.ErrWeakPassword),
         errors.Is(err, authDomain.ErrInvalidUsername):
        status = http.StatusBadRequest
    case errors.Is(err, authDomain.ErrEmailAlreadyExists),
         errors.Is(err, authDomain.ErrUsernameAlreadyExists):
        status = http.StatusConflict
    case errors.Is(err, authDomain.ErrInvalidCredentials):
        status = http.StatusUnauthorized
    }
    platformErrors.WriteErrorJSON(w, status, err.Error())
    return
}
```

**Rationale:** The current handle depends on string matching (`strings.Contains`), which is fragile. Relying on sentinel errors is more robust and idiomatic. If `Register` returns a validation error from the platform, it should still be handled consistently without resorting to string inspection.

---

### 3. Consolidate HTTP Utilities in Base Handler

**File:** `internal/modules/auth/adapters/http_handler_api.go` (and others)
**Line(s):** 230-254
**Category:** Architecture | KISS Violation
**Severity:** Low

**Current Code:**
The `writeJSON` and `parseJSON` methods are defined in `http_handler_api.go` but are effectively shared by the whole `HTTPHandler`.

**Suggested Improvement:**
Move `writeJSON` and `parseJSON` to `http_handler.go` (the base handler file). Even better, since these are likely used across all modules, they could be moved to a platform utility package, but keeping them in the base handler within the module is a good first step for local simplification.

**Rationale:** Helps keep the specific handler files (API vs Page) focused only on their routes while sharing common logic.

---

### 4. Reduce Duplication in Auth Middleware

**File:** `internal/modules/auth/adapters/middleware.go`
**Line(s):** 31-110
**Category:** KISS Violation | DRY Principle
**Severity:** Low

**Current Code:**
`RequireAuth` and `OptionalAuth` both manually extract the cookie, validate the session, and fetch the user.

**Suggested Improvement:**

```go
func (p *AuthMiddleware) authenticate(r *http.Request) (*userDomain.User, error) {
    cookie, err := r.Cookie("session_token")
    if err != nil {
        return nil, err
    }
    session, err := p.authService.ValidateSession(r.Context(), cookie.Value)
    if err != nil {
        return nil, err
    }
    return p.userService.GetByID(r.Context(), session.UserID)
}
```

**Rationale:** Centralizing the authentication logic makes the middleware functions much simpler and easier to maintain.

---

### 5. Unified Input Validation in Service

**File:** `internal/modules/auth/application/service.go`
**Line(s):** 43-61
**Category:** KISS Violation
**Severity:** Low

**Current Code:**
Validation is split between `ValidateCredentials(creds)` and a manual `validator.New()` block for the username.

**Suggested Improvement:**
Extend `ValidateCredentials` to handle the username as well, or create a `RegisterRequest` domain object with its own `Validate()` method as per `GEMINI.md` feature 1.

**Rationale:** Keeps validation logic consistent and centralized.

---

## Action Items

- [ ] Extract `getAndValidateSession` private helper in `application/Service`.
- [ ] Simplify error mapping in `RegisterAPI` by removing string-based checks.
- [ ] Move `writeJSON` and `parseJSON` to `internal/modules/auth/adapters/http_handler.go`.
- [ ] Refactor common authentication logic in `AuthMiddleware`.
- [ ] Unify registration validation logic.

---

## Notes

The `auth` module is very solid. These suggestions are aimed at reaching "platinum" level code quality by eliminating micro-repetitions and making the handlers more declarative.
