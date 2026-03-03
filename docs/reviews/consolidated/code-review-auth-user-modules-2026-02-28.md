# Code Review: Auth & User Modules — Idiomatic Go & KISS

**Date**: 2026-02-28  
**Scope**: `internal/modules/auth/` and `internal/modules/user/` (all files)  
**Principles**: Idiomatic Go, KISS, DRY, correctness

---

## Table of Contents

1. [Auth Module](#1-auth-module)
   - [domain/session.go](#11-domainsessiongo)
   - [domain/errors.go](#12-domainerrorsgo)
   - [ports/service.go](#13-portsservicego)
   - [ports/repository.go](#14-portsrepositorygo)
   - [ports/middleware.go](#15-portsmiddlewarego)
   - [application/service.go](#16-applicationservicego)
   - [adapters/http_handler.go](#17-adaptershttp_handlergo)
   - [adapters/http_handler_api.go](#18-adaptershttp_handler_apigo)
   - [adapters/http_handler_page.go](#19-adaptershttp_handler_pagego)
   - [adapters/middleware.go](#110-adaptersmiddlewarego)
   - [adapters/sqlite_session_repository.go](#111-adapterssqlite_session_repositorygo)
2. [User Module](#2-user-module)
   - [domain/user.go](#21-domainusergo)
   - [domain/errors.go](#22-domainerrorsgo)
   - [ports/service.go](#23-portsservicego)
   - [ports/repository.go](#24-portsrepositorygo)
   - [application/service.go](#25-applicationservicego)
   - [adapters/http_handler.go](#26-adaptershttp_handlergo)
   - [adapters/http_handler_api.go](#27-adaptershttp_handler_apigo)
   - [adapters/http_handler_page.go](#28-adaptershttp_handler_pagego)
   - [adapters/http_handler_settings.go](#29-adaptershttp_handler_settingsgo)
   - [adapters/sqlite_repository.go](#210-adapterssqlite_repositorygo)
3. [Test Files](#3-test-files)
4. [Cross-Cutting Issues](#4-cross-cutting-issues)
5. [Summary & Priority Matrix](#5-summary--priority-matrix)

---

## 1. Auth Module

### 1.1 `domain/session.go`

**Structure**: Session entity, Credentials value object, `IsExpired()`, `IsValid()` methods.

#### Issues

**KISS-1: `IsValid()` duplicates `IsExpired()` internally**  
`IsValid()` calls `IsExpired()` but also checks `ID > 0 && UserID > 0`. This method is **never called anywhere in production code** — only from the test file. It's dead code.

```go
// DEAD CODE — grep the codebase: no caller outside the test file
func (s *Session) IsValid() bool {
    return s.ID > 0 && s.UserID > 0 && !s.IsExpired()
}
```
**Recommendation**: Remove `IsValid()`. The service layer already handles all validation (expired check, session existence via DB lookup). If needed later, add it back.

**GO-1: `Credentials` struct has no methods or validation**  
It's a plain data holder used only as a parameter-passing struct in `ValidateCredentials()`. Idiomatic Go would just pass `email, password string` directly (which the service already does). `Credentials` is an unnecessary indirection.

```go
// Instead of:
creds := &domain.Credentials{Email: email, Password: password}
err = ValidateCredentials(creds)

// Simpler:
err = ValidateCredentials(email, password)
```

---

### 1.2 `domain/errors.go`

**Structure**: 9 sentinel errors.

#### Issues

**OVERLAP-1: Duplicate errors across auth and user modules**  
Both `auth/domain` and `user/domain` define:
- `ErrInvalidEmail`
- `ErrWeakPassword`  
- `ErrInvalidUsername`
- `ErrEmailAlreadyExists`
- `ErrUsernameAlreadyExists`

These have **different message text** between modules (e.g., auth says `"invalid email format"` vs user says `"invalid email"`), which is bug-prone and violates DRY. Since auth delegates user creation to the user service, the user module's errors should be the canonical source, and auth should just re-expose or wrap them.

---

### 1.3 `ports/service.go`

**Structure**: `AuthService` interface with 6 methods.

#### Issues

**KISS-2: `GetSession()` vs `ValidateSession()` overlap**  
Both methods retrieve a session by token and check expiration. The application layer implementations are nearly identical (both call `GetByToken`, both check `IsExpired()`, both delete expired sessions). This is a KISS violation — one method suffices.

```go
// ValidateSession and GetSession are identical in implementation
// Keep only ValidateSession, remove GetSession
```

**GO-2: `RefreshSession()` is never called**  
Searching the codebase: `RefreshSession` has no callers outside the test file. It's dead code in the interface.

---

### 1.4 `ports/repository.go`

**Structure**: `SessionRepository` interface with 7 methods. Clean, minimal.

No significant issues.

---

### 1.5 `ports/middleware.go`

**Structure**: `AuthMiddleware` interface, context helpers, `ContextKey` type.

#### Issues

**GO-3: `Middleware` type alias is redundant**  
```go
type Middleware func(http.Handler) http.Handler
```
This is the standard `func(http.Handler) http.Handler` pattern. The type alias adds a layer of indirection without adding clarity. It's acceptable but adds a concept that callers must look up.

**POSITIVE**: The `GetUserID()`, `GetUsername()`, `IsAuthenticated()` helpers are well-designed and properly use typed context keys.

---

### 1.6 `application/service.go`

**Structure**: Auth service implementation, credential validation, bcrypt hashing.

#### Issues

**KISS-3: `ValidateCredentials` is an exported function taking a `*Credentials` pointer — over-engineered**  
Takes a pointer to a struct just to read two string fields. Could be `validateCredentials(email, password string) error` (unexported, since it's internal to the package).

**KISS-4: Validator error-to-domain-error mapping is fragile**  
```go
for field := range v.Errors() {
    switch field {
        case "email": return domain.ErrInvalidEmail
        case "password": return domain.ErrWeakPassword
    }
}
```
Map iteration order is random in Go. If both email and password fail, which error gets returned is **non-deterministic**. This should check fields in priority order:

```go
if errs := v.Errors(); len(errs) > 0 {
    if _, ok := errs["email"]; ok {
        return domain.ErrInvalidEmail
    }
    if _, ok := errs["password"]; ok {
        return domain.ErrWeakPassword
    }
}
```

**KISS-5: `Register()` validation is split across two validation phases**  
Lines 50–68 validate credentials, then separately validate username. These should be unified into a single validation pass:

```go
// Current: 2 validators, 2 separate checks
err = ValidateCredentials(creds)
validation := validator.New()
validation.Required("username", username)
// ...
```

**BUG-1: `Register()` sanitizes input AFTER duplicate checks**  
Lines 66–68 sanitize email/username, but the `ExistsByEmail`/`ExistsByUsername` checks at lines 70–82 happen AFTER sanitization. However, the `ValidateCredentials` call at line 50 validates the raw un-sanitized input. If the sanitizer alters the email (e.g., trimming spaces), the validation is checking different data than what gets stored. The sanitize step should come first.

**GO-4: `hashPassword` and `comparePassword` are methods on `*Service` but use no receiver fields**  
These should be package-level functions:
```go
func hashPassword(password string) (string, error) { ... }
func comparePassword(hash, password string) error { ... }
```

**PERF-1: `generateSessionToken` creates a new UUID each call — fine, but the method doesn't need `*Service` receiver**  
Same as above: make it a package-level function.

**KISS-6: Repeated expired-session cleanup pattern**  
`ValidateSession`, `RefreshSession`, and `GetSession` all contain the exact same pattern:
```go
if session.IsExpired() {
    if err := s.sessionRepo.Delete(ctx, sessionToken); err != nil {
        log.Printf("WARNING: ...")
    }
    return nil, domain.ErrSessionExpired
}
```
Extract to a helper:
```go
func (s *Service) validateNotExpired(ctx context.Context, session *domain.Session) error {
    if !session.IsExpired() { return nil }
    if err := s.sessionRepo.Delete(ctx, session.Token); err != nil {
        log.Printf("WARNING: failed to delete expired session: %v", err)
    }
    return domain.ErrSessionExpired
}
```

---

### 1.7 `adapters/http_handler.go`

**Structure**: Base handler, `ServiceContainer` interface, `RegisterRoutes()`, `GetCurrentUser()`.

#### Issues

**KISS-7: `GetCurrentUser()` is dead code or poorly placed**  
This method validates sessions and fetches user data, duplicating what the middleware already does. It returns `(userID int, username string)` using internal INT IDs, which violates the project's UUID-externally rule. No callers found in the auth module. If other modules need this, they should use the middleware context values.

**GO-5: `HTTPHandler` stores `templates *template.Template` but never uses it**  
Page handlers use `templates.Get()` (the platform template cache) instead of `h.templates`. The stored field is dead.

---

### 1.8 `adapters/http_handler_api.go`

**Structure**: API handlers for register, login, logout, session.

#### Issues

**KISS-8: `RegisterAPI` error handling has a massive, fragile fallback block**  
Lines 62–74: After matching known domain errors, there's a catch-all that pattern-matches error message substrings:
```go
if strings.Contains(errMsg, "empty") || strings.Contains(errMsg, "invalid") || ...
```
This is extremely fragile. New errors with "invalid" in their message will silently match. Replace with a simple `500 Internal Server Error` default, or use error wrapping with `errors.Is()`.

**DRY-1: Cookie-setting code is repeated 4 times**  
`RegisterAPI`, `LoginAPI`, `LogoutAPI`, `LogoutPage` all construct `http.Cookie` structs with mostly identical fields. Extract:
```go
func (h *HTTPHandler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
    http.SetCookie(w, &http.Cookie{
        Name: "session_token", Value: token, Path: "/",
        Expires: expiresAt, HttpOnly: true,
        Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
    })
}

func (h *HTTPHandler) clearSessionCookie(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{
        Name: "session_token", Value: "", Path: "/",
        MaxAge: -1, HttpOnly: true,
        Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
    })
}
```

**DRY-2: Response struct for register/login is duplicated**  
`RegisterAPI` and `LoginAPI` both define identical anonymous structs:
```go
resp := struct {
    ID       string `json:"id"`
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Token    string `json:"token"`
}{ ... }
```
Extract to a named type:
```go
type authResponse struct {
    ID       string `json:"id"`
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Token    string `json:"token"`
}
```

**GO-6: `parseJSON` enforces `DisallowUnknownFields()` — too strict**  
`DisallowUnknownFields()` means if a client sends `{"email":"...", "extra":"..."}`, the request fails. This is unusual for Go APIs and breaks forward compatibility. Remove `DisallowUnknownFields()` unless intentionally strict.

**GO-7: `parseJSON` checks `Content-Type` exactly equals `"application/json"`**  
Fails for `"application/json; charset=utf-8"`. Use `strings.HasPrefix`:
```go
if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
```

---

### 1.9 `adapters/http_handler_page.go`

**Structure**: Login, Register, Logout page handlers.

#### Issues

**BUG-2: `LogoutPage` hardcodes `Secure: false` instead of using `h.secureCookies`**  
Line 77:
```go
Secure: false, // Set to true in production with HTTPS
```
Other handlers correctly use `h.secureCookies`. This is a bug.

---

### 1.10 `adapters/middleware.go`

**Structure**: `AuthMiddleware` implementation, deprecated compatibility functions.

#### Issues

**DEAD-1: Three deprecated functions should be removed**  
Lines 73–82 (`RequireAuthFunc`), 109–112 (`OptionalAuthFunc`), 117–126 (`GetUserID`, `GetUsername`, `IsAuthenticated` proxies). These are marked `// DEPRECATED` but still exist. If truly deprecated, remove them to reduce surface area.

**GO-8: `RequireAuth()` and `OptionalAuth()` share 80% of their code**  
Both:
1. Get cookie
2. Validate session
3. Fetch user by ID
4. Set context with public ID

The only difference is error handling (401 vs continue). Extract the common logic:

```go
func (p *AuthMiddleware) extractUserFromSession(r *http.Request) (string, error) {
    cookie, err := r.Cookie("session_token")
    if err != nil { return "", err }
    session, err := p.authService.ValidateSession(r.Context(), cookie.Value)
    if err != nil { return "", err }
    user, err := p.userService.GetByID(r.Context(), session.UserID)
    if err != nil { return "", err }
    return user.PublicID, nil
}
```

**PERF-2: Every authenticated request makes 2 DB queries (session + user)**  
`ValidateSession` hits the sessions table, then `GetByID` hits the users table. Consider caching the user's PublicID in the session itself, or using a single JOIN query.

---

### 1.11 `adapters/sqlite_session_repository.go`

**Structure**: SQLite implementation of `SessionRepository`.

#### Issues

**GO-9: `Update` doesn't verify the row existed**  
`Update` runs `UPDATE sessions SET expires_at = ? WHERE token = ?` but doesn't check `RowsAffected()`. If the token doesn't exist, it silently succeeds. In contrast, `Delete` correctly checks this.

**POSITIVE**: Clean, well-structured SQL operations. Good use of `sql.ErrNoRows` mapping.

---

## 2. User Module

### 2.1 `domain/user.go`

**Structure**: User entity, Role type, permission system.

#### Issues

**GO-10: Permission constants are untyped strings**  
```go
const PermissionViewContent = "view"
```
These should be a typed constant:
```go
type Permission string
const PermissionViewContent Permission = "view"
```
This prevents accidental use of arbitrary strings and enables type-safe function signatures.

**KISS-9: `HasPermission()` switch-in-switch is complex**  
The nested switch has O(roles × permissions) cases. A map-based approach is simpler and data-driven:
```go
var rolePermissions = map[Role]map[Permission]bool{
    RoleAdmin: nil, // special: all permissions
    RoleModerator: {PermissionViewContent: true, ...},
    RoleUser: {PermissionViewContent: true, ...},
    RoleGuest: {PermissionViewContent: true},
}
```

**POSITIVE**: `CanModerate()` and `IsAdmin()` are clean convenience methods.

---

### 2.2 `domain/errors.go`

See [OVERLAP-1](#issues-1) — duplicates auth domain errors with different messages.

---

### 2.3 `ports/service.go`

**Structure**: `UserService` interface with 17 methods.

#### Issues

**KISS-10: Interface is too large (17 methods)**  
This is a "fat interface" anti-pattern. Per Go proverb: "The bigger the interface, the weaker the abstraction." The 6 Increment/Decrement methods are mechanical counter operations that could be a separate `UserStatsService` interface or folded into the repository directly without a service layer.

**GO-11: `CreateUser` returns `(userID int, err error)` — leaks internal ID**  
The return type exposes the internal integer ID across module boundaries. The auth module receives this int and uses it to create sessions. While this is internal-only, it weakens the UUID boundary. Consider returning the `*User` directly.

---

### 2.4 `ports/repository.go`

**Structure**: `UserRepository` interface with 16 methods.

#### Issues

**INTERFACE-1: Missing methods vs test mocks**  
The `ports/service_test.go` mock has `Get()` and `UpdatePassword()` methods that **don't exist** on the actual `UserRepository` interface (which has `GetByID()`, no `UpdatePassword()`). The test compiles because the mock is never assigned to a `UserRepository` variable — but the test is verifying the wrong interface. This is a correctness bug in the test.

---

### 2.5 `application/service.go`

**Structure**: User service implementation.

#### Issues

**KISS-11: 8 pure pass-through methods**  
These methods add zero business logic:
```go
func (s *Service) IncrementPostCount(ctx context.Context, userID int) error {
    return s.userRepo.IncrementPostCount(ctx, userID)
}
```
`GetByID`, `GetByPublicID`, `GetByUsername`, `GetByEmail`, `ListUsers`, `ExistsByEmail`, `ExistsByUsername` plus all 6 Increment/Decrement methods are pure delegation. Consider whether the service layer adds value here or just adds indirection.

**GO-12: Stale `TODO` comments**  
Lines 55, 65, 137 have `// TODO: Implement user retrieval.` and `// TODO: Implement username-based retrieval.` but the methods are already implemented (just delegate to repo). Remove the stale TODOs.

**KISS-12: `UpdateSettings` does double validation**  
`UpdateSettings` validates and sanitizes username/email, but the caller (`http_handler_settings.go`) already validates required fields and trims spaces before calling. The `strings.TrimSpace()` in the handler and in the service are redundant.

**BUG-3: `UpdateSettings` avatar URL construction is duplicated**  
Lines 227–231 build `user.AvatarURL = "/static/uploads/" + user.AvatarPath`, but `scanUserRowWithAvatar()` in the repository already does this when reading. This means the URL is set twice with potentially different logic (repo uses `user.AvatarPath`, service also uses `user.AvatarPath`). The URL construction should live in exactly one place.

**GO-13: `isValidRole()` is a package-level function but could be a method on `Role`**  
```go
func (r Role) IsValid() bool {
    switch r {
    case RoleGuest, RoleUser, RoleModerator, RoleAdmin:
        return true
    }
    return false
}
```

---

### 2.6 `adapters/http_handler.go`

**Structure**: Base handler, DI.

#### Issues

**KISS-13: `NewHTTPHandler` uses variadic `uploadDir ...string` for a single optional parameter**  
This is an unusual pattern. A struct-based options pattern or just requiring the parameter is cleaner:
```go
func NewHTTPHandler(services ServiceContainer, templates *template.Template, uploadDir string) *HTTPHandler {
    if uploadDir == "" { uploadDir = "./static/uploads" }
```

---

### 2.7 `adapters/http_handler_api.go`

**Structure**: User CRUD API handlers.

#### Issues

**KISS-14: `ListUsersAPI` hardcodes pagination and ignores query params**  
```go
offset := 0
limit := 20
```
The query params are never read. Either document this is intentional or read `r.URL.Query().Get("offset")`.

**BUG-4: `UpdateRoleAPI` doesn't check if the caller is an admin**  
The comment says "Requires admin permissions (checked via middleware in production)" but the middleware only checks authentication, not authorization. Any authenticated user can change any user's role. This is a security vulnerability.

**DRY-3: Repeated pattern in Deactivate/Activate/UpdateRole handlers**  
All three:
1. Get publicID from path
2. Get user by public ID
3. Call service method
4. Return JSON success

Extract a helper:
```go
func (h *HTTPHandler) withUser(w http.ResponseWriter, r *http.Request, fn func(*domain.User) error) {
    publicID := r.PathValue("id")
    user, err := h.userService.GetByPublicID(r.Context(), publicID)
    if err != nil || user == nil {
        platformErrors.WriteErrorJSON(w, http.StatusNotFound, "user not found")
        return
    }
    if err := fn(user); err != nil {
        platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, err.Error())
        return
    }
    // success response
}
```

---

### 2.8 `adapters/http_handler_page.go`

**Structure**: Settings page routes.

No major issues. Clean and minimal.

---

### 2.9 `adapters/http_handler_settings.go`

**Structure**: Settings update handlers for both HTML form and API.

#### Issues

**KISS-15: `parseSettingsUpdateInput` handles both multipart and form-encoded**  
This is reasonable but the function is 40+ lines. Consider splitting multipart parsing from url-encoded parsing.

**KISS-16: `renderSettingsPage` has redundant nil check**  
```go
if currentUser == nil {
    currentUser = &domain.User{}
}
if currentUser != nil && currentUser.AvatarPath != "" && ... {
```
The `currentUser != nil` check on the second `if` will always be true because the previous block guarantees non-nil. Remove the redundant nil check.

**GO-14: `updateCurrentUserSettings` returns `(*domain.User, int, string)` — stringly-typed errors**  
Returning `string` for error messages instead of `error` is not idiomatic Go. Consider wrapping with a custom error type:
```go
type handlerError struct {
    statusCode int
    message    string
}
func (e *handlerError) Error() string { return e.message }
```

---

### 2.10 `adapters/sqlite_repository.go`

**Structure**: SQLite user repository, legacy avatar-column fallback.

#### Issues

**KISS-17: Legacy fallback for missing `avatar_path` column — keep or remove?**  
The code has elaborate fallback logic (`isMissingAvatarColumnError`, `getByIDLegacy`, `getByPublicIDLegacy`, `scanUserRowLegacy`, legacy UPDATE query). This means ~80 extra lines and two code paths for every read/update operation. Since migration `008_user_add_avatar_path.sql` exists and auto-applies, the legacy path is dead code in any properly migrated database.

**Recommendation**: Remove all legacy fallback code. If migration hasn't run, the app should fail clearly rather than silently degrading.

**DRY-4: `GetByEmail` and `GetByUsername` have identical scan logic**  
Both methods:
1. Run a SELECT with different WHERE clause
2. Scan identical columns into `domain.User`
3. Convert `isActive int` to `bool`

Extract a shared scanner:
```go
func (r *SQLiteUserRepository) getUserByQuery(ctx context.Context, query string, arg any) (*domain.User, error) {
    row := r.db.QueryRowContext(ctx, query, arg)
    var user domain.User
    var isActive int
    err := row.Scan(&user.ID, &user.PublicID, ...)
    if err != nil {
        if err == sql.ErrNoRows { return nil, domain.ErrUserNotFound }
        return nil, err
    }
    user.IsActive = isActive == 1
    return &user, nil
}
```

**Note**: `GetByEmail` and `GetByUsername` don't include `avatar_path` in their SELECT, unlike `GetByID`/`GetByPublicID`. This means avatar data is silently lost when fetching by email or username. This is likely a bug.

**GO-15: `ExistsByEmail` and `ExistsByUsername` use `SELECT COUNT(*)` instead of `SELECT 1 ... LIMIT 1`**  
`SELECT COUNT(*)` scans all matching rows. `SELECT 1 FROM users WHERE email = ? LIMIT 1` stops at the first match and is more efficient:
```go
query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`
```

---

## 3. Test Files

### 3.1 `auth/ports/service_test.go` (284 lines)

**ISSUE-T1: Contains two duplicate mock implementations for the same interface**  
`MockSessionRepository` (capital M, line 60) and `mockSessionRepository` (lowercase m, line 257). The lowercase one implements an outdated interface signature. Only one is needed.

**ISSUE-T2: `MockUserRepository` uses `interface{}` instead of concrete types**  
Lines 22–30: All methods use `interface{}` as argument and return type:
```go
func (m *MockUserRepository) Create(ctx context.Context, user interface{}) error
func (m *MockUserRepository) Get(ctx context.Context, id int) (interface{}, error)
```
This mock doesn't actually implement any real interface — it's testing phantom types.

**ISSUE-T3: Interface verification tests are boilerplate**  
`TestAuthServiceInterface`, `TestSessionRepositoryInterface` etc. just declare `var x Interface` and assign nil. The Go compiler already catches unimplemented interfaces at assignment sites. These tests add no value.

### 3.2 `auth/domain/session_test.go`

**ISSUE-T4: `TestCredentials` is a trivial struct-field test**  
```go
creds := &Credentials{Email: "test@example.com", Password: "password123"}
if creds.Email != "test@example.com" { ... }
```
This tests Go's struct initialization, not application logic. Remove.

### 3.3 `user/ports/service_test.go` (266 lines)

**BUG-T1: `mockUserRepository` has wrong method signatures**  
- Has `Get()` — interface has `GetByID()` and `GetByPublicID()`, no `Get()`
- Has `UpdatePassword()` — interface has no `UpdatePassword()`
- Missing `Count()`, `IncrementPostCount`, `DecrementPostCount`, etc.

The mock is never assigned to a `UserRepository` variable, so it compiles but **doesn't verify interface compatibility** — which is the stated purpose of the test.

### 3.4 `user/domain/user_test.go`

**ISSUE-T5: `TestUser_StructFields` and `TestRoleConstants` test Go language features**  
These verify that struct fields hold values and string constants equal their literal values. No business logic tested.

**ISSUE-T6: `TestUser_PostAndCommentCounts` is another struct-field test**  
Verifies `user.PostCount != 5` after setting it to 5. Remove.

### 3.5 `user/adapters/sqlite_repository_test.go` (1072 lines)

**DRY-T1: `CREATE TABLE users (...)` is repeated 15+ times**  
Every test function has identical table creation. Extract a `setupTestDB(t *testing.T) *sql.DB` helper (the auth tests do this with `createSessionsTable`).

### 3.6 General Test Observations

**MOCK-EXPLOSION: UserService mock appears in 3 separate files**  
- `auth/application/service_test.go` — `MockUserService`
- `user/adapters/http_handler_api_test.go` — `MockUserService`
- `user/application/service_test.go` — `MockUserRepository`

Each re-implements 15+ methods. Consider a shared `testutil` package or code-generated mocks.

---

## 4. Cross-Cutting Issues

### 4.1 Duplicate Error Definitions (HIGH)

Both `auth/domain/errors.go` and `user/domain/errors.go` define:
| Error | Auth Message | User Message |
|-------|-------------|--------------|
| `ErrInvalidEmail` | `"invalid email format"` | `"invalid email"` |
| `ErrWeakPassword` | `"password doesn't meet security requirements"` | `"password must be at least 8 characters"` |
| `ErrInvalidUsername` | `"invalid username: must start with..."` | `"invalid username"` |
| `ErrEmailAlreadyExists` | `"user with this email already exists"` | `"email already exists"` |
| `ErrUsernameAlreadyExists` | `"user with this username already exists"` | `"username already exists"` |

**Impact**: Client-facing error messages are inconsistent depending on which code path is hit.  
**Fix**: User module owns identity errors. Auth module imports them or wraps them.

### 4.2 Fat Interfaces

`UserService` has 17 methods. Split suggestion:
- `UserService` (core): `CreateUser`, `GetByID`, `GetByPublicID`, `GetByEmail`, `GetByUsername`, `UpdateSettings`
- `UserQueryService` (reads): `ListUsers`, `ExistsByEmail`, `ExistsByUsername`
- `UserStatsService` (counters): 6× Increment/Decrement methods
- `UserAdminService` (admin): `UpdateRole`, `DeactivateUser`, `ActivateUser`

### 4.3 Avatar URL Construction Scattered

`"/static/uploads/" + path` appears in:
- `user/adapters/sqlite_repository.go` (`scanUserRowWithAvatar`)
- `user/application/service.go` (`UpdateSettings`)
- `user/adapters/http_handler_settings.go` (`renderSettingsPage`)

Define a single helper: `func AvatarURL(path string) string`.

### 4.4 `errors.Is()` vs `==` for Error Comparison

Several handlers use `==` for error comparison:
```go
if err == domain.ErrInvalidRole {  // adapters/http_handler_api.go
```
This breaks if errors are wrapped with `fmt.Errorf("...: %w", err)`. Always use `errors.Is()`.

---

## 5. Summary & Priority Matrix

| Priority | ID | Category | Description | Effort |
|----------|----|----------|-------------|--------|
| **P0** | BUG-4 | Security | `UpdateRoleAPI` has no authorization check | Low |
| **P0** | BUG-2 | Bug | `LogoutPage` hardcodes `Secure: false` | Trivial |
| **P1** | BUG-1 | Bug | Sanitization after validation in `Register()` | Low |
| **P1** | KISS-4 | Correctness | Non-deterministic map iteration in `ValidateCredentials` | Low |
| **P1** | OVERLAP-1 | DRY | Duplicate error definitions across modules | Medium |
| **P1** | BUG-T1 | Test Bug | Mock has wrong interface methods | Low |
| **P2** | KISS-8 | Maintainability | Fragile string-matching error classification | Low |
| **P2** | DRY-1 | DRY | Cookie code duplicated 4× | Low |
| **P2** | DRY-4 | DRY | Scan logic duplicated in GetByEmail/GetByUsername | Low |
| **P2** | KISS-17 | KISS | Legacy avatar fallback code (~80 dead lines) | Medium |
| **P2** | GO-7 | Bug | Content-Type check too strict | Trivial |
| **P3** | KISS-2 | KISS | `GetSession` duplicates `ValidateSession` | Low |
| **P3** | KISS-6 | DRY | Expired session cleanup repeated 3× | Low |
| **P3** | KISS-10 | Design | 17-method fat interface | Medium |
| **P3** | KISS-11 | KISS | 8 pure pass-through service methods | Low |
| **P3** | DEAD-1 | Dead Code | 3 deprecated wrapper functions | Trivial |
| **P3** | GO-12 | Cleanup | Stale TODO comments | Trivial |
| **P4** | PERF-2 | Performance | 2 DB queries per auth check | Medium |
| **P4** | GO-10 | Typing | Untyped permission constants | Low |
| **P4** | DRY-T1 | Tests | Table creation repeated 15× in tests | Low |
| **P4** | T3/T4/T5 | Tests | Trivial/valueless tests | Low |

**Top 3 Actions**:
1. Fix BUG-4 (security: add admin authorization check to role update endpoint)
2. Fix BUG-2 (hardcoded `Secure: false` in LogoutPage cookie)
3. Consolidate duplicate error definitions between auth and user modules
