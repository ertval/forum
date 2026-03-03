# Full Codebase Review — Idiomatic Go & KISS Simplifications

**Date**: 2026-02-28  
**Scope**: All modules (`auth`, `user`, `post`, `comment`, `reaction`, `moderation`, `notification`), platform packages, wire/DI, templates, build  
**Principles**: Idiomatic Go, KISS, DRY, minimal abstractions, correctness

---

## Executive Summary

The codebase is well-structured with clear module boundaries and consistent hexagonal architecture. However, significant complexity has accumulated in cross-cutting patterns, over-abstracted layers, and duplicated logic. This review identifies **~65 issues** across 4 severity levels, organized by theme.

**Top wins (effort-to-impact ratio)**:
1. Extract 6 duplicated helpers into shared utilities (~200 lines saved)
2. Consolidate near-identical page handlers (~400 lines saved)
3. Replace custom logger with `slog` (~500 lines saved)
4. Remove dead/scaffolded code in moderation module (~300 lines saved)
5. Fix 3 correctness bugs (sanitization order, mock drift, silent no-ops)

---

## P0 — Bugs & Security

### S-1. Sanitizer unescape-before-strip creates bypass vector
**File**: `internal/platform/validator/validator.go`  
**Issue**: `Sanitize()` calls `html.UnescapeString()` before stripping HTML tags. Input `&lt;script&gt;alert(1)&lt;/script&gt;` gets unescaped to `<script>alert(1)</script>`, then tags are stripped. However, double-encoded input `&amp;lt;` only gets one pass of unescaping, leaving `&lt;` intact — inconsistent behavior. Either loop unescape until stable, or don't unescape at all (just strip tags on raw input).

### S-2. `UpdateRoleAPI` lacks authorization
**File**: `internal/modules/user/adapters/http_handler_api.go`  
**Issue**: Any authenticated user can call the role-update endpoint. There's no check that the caller is an admin. Add role-based authorization guard.

### S-3. `LogoutPage` hardcodes `Secure: false`
**File**: `internal/modules/auth/adapters/http_handler_page.go`  
**Issue**: Cookie deletion sets `Secure: false` instead of using `h.secureCookies`, allowing cookie to be sent over HTTP.

### S-4. Health page leaks internal integer user IDs
**File**: `internal/platform/httpserver/health.go`  
**Issue**: `AuthFunc` returns `userID int` (internal sequential ID) which is placed into template data as `map[string]interface{}{"ID": userID}`. Per project rules: never expose sequential IDs. Use UUID string instead.

---

## P1 — Correctness

### C-1. `Register()` sanitizes after validation
**File**: `internal/modules/auth/application/service.go`  
**Issue**: Email/username are validated first, then sanitized. If sanitization changes the value (e.g., stripping tags from a name), the validated value differs from the stored value. Sanitize first, then validate.

### C-2. Moderation service is a silent no-op
**File**: `internal/modules/moderation/application/service.go`  
**Issue**: `CreateReport()` and `ReviewReport()` return `nil` (success) without doing anything. Callers believe the operation succeeded. Either implement or return `ErrNotImplemented`.

### C-3. Notification target resolution only checks posts table
**File**: `internal/modules/notification/adapters/sqlite_repository.go`  
**Issue**: `Create()` resolves target IDs by querying only the `posts` table. Comment notifications (`TypeComment`, `TypeReply`) would fail silently or get wrong IDs.

### C-4. `HTTPStatus` / `ToHTTPResponse` use type assertion, not `errors.As`
**File**: `internal/platform/errors/errors.go`  
**Issue**: `err.(*Error)` assertion fails silently on wrapped errors. Use `errors.As`:
```go
var e *Error
if errors.As(err, &e) { ... }
```

### C-5. Mock signatures drift from interfaces
**Files**: `reaction/ports/service_test.go`, `comment/ports/service_test.go`, `user/ports/service_test.go`  
**Issue**: Mock methods use different parameter types (`int` vs `string`, missing params) than the actual interfaces. Mocks never compile-check against interfaces. Fix: add `var _ ports.XxxService = (*MockXxxService)(nil)` assertions.

### C-6. `GetByEmail`/`GetByUsername` missing columns in SELECT
**File**: `internal/modules/user/adapters/sqlite_repository.go`  
**Issue**: `avatar_path` not included in SELECT for `GetByEmail` and `GetByUsername` queries — silent data loss when these methods are used.

### C-7. Moderation health check always returns "down"
**File**: `internal/platform/health/checker.go`  
**Issue**: `map[bool]string{true: "down", false: "down"}[moderationAllUp]` — both branches return "down". Dead code.

---

## P2 — DRY Violations (Cross-Module Duplication)

### D-1. `buildCurrentUser()` duplicated in 4+ handlers
**Files**: `post/adapters`, `comment/adapters`, `reaction/adapters`, `notification/adapters`  
**Pattern**: Each handler has a ~15-line function that extracts user ID from context and builds a `currentUser` map. Extract to a shared `platform/httputil` helper:
```go
func CurrentUserFromContext(r *http.Request) map[string]interface{} { ... }
```

### D-2. `getInternalUserID()` duplicated in 3+ handlers
**Files**: `post/adapters`, `comment/adapters`, `reaction/adapters`  
**Pattern**: Each handler resolves UUID → internal int ID via user service. Extract to shared helper or middleware that adds internal ID to context.

### D-3. `writeJSON()` duplicated in every handler package
**Files**: All `http_handler_api.go` files  
**Pattern**: Each module has its own identical `writeJSON` helper. Extract to `platform/httputil`:
```go
func WriteJSON(w http.ResponseWriter, status int, v interface{}) { ... }
```

### D-4. Cookie construction repeated 4 times
**File**: `internal/modules/auth/adapters/http_handler_api.go`, `http_handler_page.go`  
**Pattern**: `http.Cookie{Name: "session_token", Path: "/", HttpOnly: true, ...}` built identically in login, register, logout (API + page). Extract `newSessionCookie()` and `newDeleteCookie()`.

### D-5. `ServiceContainer` interface pattern boilerplate
**Files**: Every handler's `http_handler.go`  
**Issue**: Each module declares a local `ServiceContainer` interface with method accessors. The pattern is correct (interface segregation), but the boilerplate could be reduced with Go generics or by passing service directly instead of through a container for single-service handlers.

### D-6. Mock structs duplicated across test files
**Files**: All `*_test.go` across modules  
**Pattern**: `MockUserService`, `MockPostService`, etc. are copy-pasted with ~60 lines of no-op methods in each module's tests. Create a shared `internal/testutil/mocks.go` package.

---

## P2 — DRY Violations (Intra-Module Duplication)

### D-7. `HomePage` and `BoardPage` are 90% identical
**File**: `internal/modules/post/adapters/http_handler_page.go`  
**Issue**: Both load posts with filters, build template data, render a page. Only difference is template name and minor data setup. Extract shared `renderPostListPage()`.

### D-8. `ListPostsAPI` and `LoadMorePostsAPI` are 80% identical
**File**: `internal/modules/post/adapters/http_handler_api.go`  
**Issue**: Same filter building, same post fetching, same JSON response. `LoadMore` just adds pagination. Consolidate into one handler with optional pagination params.

### D-9. `ListByUser` / `ListByUserPaginated` near-copies in repository
**File**: `internal/modules/comment/adapters/sqlite_repository.go`  
**Issue**: Two methods with nearly identical SQL — one adds `LIMIT/OFFSET`. Use optional params on one method.

### D-10. Target-exists validation repeated 5× in reaction service
**File**: `internal/modules/reaction/application/service.go`  
**Issue**: The same `switch targetType { case "post": ... case "comment": ... }` block appears in every method. Extract `verifyTargetExists(targetType, targetID)`.

### D-11. Public→internal ID resolution repeated 5× in reaction repo
**File**: `internal/modules/reaction/adapters/sqlite_repository.go`  
**Issue**: Same query pattern for resolving public UUID to internal int ID. Extract `resolveTargetID(table, publicID)`.

---

## P2 — KISS / Over-Abstraction

### K-1. Custom logger is 593 lines — replace with `slog`
**File**: `internal/platform/logger/logger.go`  
**Issue**: ~300 lines of human-readable formatting with color management, emoji indicators, content sniffing, HTTP-specific formatting. Go 1.21+ `slog` handles all of this natively. Replace with:
```go
var Logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
```
For development, use `slog.NewTextHandler`. This eliminates the entire file.

### K-2. `FilterService` as a stateless struct behind an interface
**File**: `internal/modules/post/application/filter_service.go`  
**Issue**: `FilterService` has zero fields and its methods are pure functions that transform `FilterParams` → `PostFilter`. This is over-abstraction. Make them plain functions:
```go
func BuildFilter(params FilterParams) PostFilter { ... }
```

### K-3. Dual `FilterParams` / `PostFilter` with redundant booleans
**Files**: `internal/modules/post/domain/filter.go`, `ports/service.go`  
**Issue**: `FilterParams` has `HasLiked bool` / `LikedByUserID string` — one boolean controls whether the string is used. Just check if the string is empty. Same for `HasCreated` / `CreatedByUserID`.

### K-4. `Transaction` wrapper adds indirection with no benefit
**File**: `internal/platform/database/transaction.go`  
**Issue**: The wrapper just exposes `Tx()` so callers use `*sql.Tx` directly anyway. Replace with a `RunInTx` helper function:
```go
func (c *Connection) RunInTx(ctx context.Context, fn func(*sql.Tx) error) error { ... }
```

### K-5. `UserService` is a 17-method fat interface
**File**: `internal/modules/user/ports/service.go`  
**Issue**: 8 of 17 methods are pure pass-throughs to repository. Split into `UserService` (business logic) and use repository directly where no business logic is needed.

### K-6. `GetSession` and `ValidateSession` are functionally identical
**File**: `internal/modules/auth/ports/service.go`  
**Issue**: Both retrieve and validate a session. Remove one.

### K-7. `RefreshSession` has no callers outside tests
**File**: `internal/modules/auth/application/service.go`  
**Issue**: Dead code. Remove or document as intentionally reserved.

### K-8. Health checker hard-codes every API endpoint
**File**: `internal/platform/health/checker.go`  
**Issue**: 70+ lines of hard-coded endpoint lists that must be manually updated on every route change. Simplify to DB connectivity check + module self-registration pattern.

### K-9. `map[string]interface{}` everywhere for template data
**Files**: All `http_handler_page.go` files  
**Issue**: Template data built as untyped maps is error-prone (typos in keys cause silent template failures). Use typed structs:
```go
type PostListPageData struct {
    Posts      []domain.Post
    User       *UserInfo
    Categories []domain.Category
    // ...
}
```

### K-10. Config `Validate()` fails on first error
**File**: `internal/platform/config/config.go`  
**Issue**: 100+ line sequential if-return chain. Users get one error at a time. Use `errors.Join` to collect all validation errors:
```go
var errs []error
if c.Server.Port <= 0 { errs = append(errs, fmt.Errorf("invalid port")) }
return errors.Join(errs...)
```

---

## P2 — Idiomatic Go

### G-1. Error comparison uses `==` instead of `errors.Is()`
**Files**: Multiple across modules  
**Issue**: `if err == domain.ErrNotFound` won't work with wrapped errors. Use `errors.Is(err, domain.ErrNotFound)`.

### G-2. Inconsistent `Validate()` vs `IsValid() bool`
**Files**: `reaction/domain` uses `Validate() error`, `moderation/domain` uses `IsValid() bool`  
**Issue**: Inconsistent API across modules. Standardize on `Validate() error` (returns reason for failure).

### G-3. `WriteErrorJSON` logs all errors at ERROR level
**File**: `internal/platform/errors/errors.go`  
**Issue**: 4xx client errors (bad request, not found) are normal and shouldn't be ERROR. Log 5xx at ERROR, 4xx at INFO/DEBUG.

### G-4. `var migrations = []migrationFile{}` should be nil slice
**File**: `internal/platform/database/migrator.go`  
**Issue**: `var migrations []migrationFile` (nil) is idiomatic. Empty literal `[]migrationFile{}` is only needed when marshaling to JSON where `null` vs `[]` matters.

### G-5. Package `errors` shadows stdlib
**File**: `internal/platform/errors/`  
**Issue**: Importing requires aliasing every time. Rename to `apperrors` or `httperrors`.

### G-6. Fire-and-forget goroutines for counter updates
**File**: `internal/modules/reaction/application/service.go`  
**Issue**: `go func() { updateCounters() }()` is untestable and risks races. Use synchronous calls or a proper background worker with error reporting.

### G-7. `PreferServerCipherSuites` deprecated since Go 1.17
**File**: `internal/platform/httpserver/tls.go`  
**Issue**: Go's TLS stack handles cipher preference automatically. Remove the deprecated field.

### G-8. `colorForMessage` sniffs message content for logging color
**File**: `internal/platform/logger/logger.go`  
**Issue**: Checking substrings like "started", "error", "failed" to pick color is brittle. Log level already conveys severity.

---

## P3 — Minor / Low Priority

### M-1. `responseWriter` wrapper doesn't implement `http.Flusher`/`http.Hijacker`
**File**: `internal/platform/httpserver/middleware.go`  
**Issue**: Breaks SSE and WebSocket upgrades. Implement via delegation or use `http.ResponseController`.

### M-2. Rate limiter goroutine leaks
**File**: `internal/platform/httpserver/middleware.go`  
**Issue**: `RateLimit()` creates a background cleanup goroutine that's never stopped. Either expose `Stop()` or use lazy cleanup.

### M-3. `Start()` uses 100ms `time.After` for error detection
**File**: `internal/platform/httpserver/server.go`  
**Issue**: Race condition — if server fails after 100ms, error is lost. Use sync mechanism.

### M-4. Double-close bug in database connection
**File**: `internal/platform/database/connection.go`  
**Issue**: After `Close()`, `c.db` is not set to `nil`. Second `Close()` calls `db.Close()` on already-closed DB. Set `c.db = nil` after close.

### M-5. Migrations not wrapped in transactions
**File**: `internal/platform/database/migrator.go`  
**Issue**: Partial migration failure leaves DB inconsistent (SQL executed but version not recorded).

### M-6. `SanitizeHTML` identical to `Sanitize`
**File**: `internal/platform/validator/validator.go`  
**Issue**: Dead code duplication. Remove `SanitizeHTML`.

### M-7. Tests that test stdlib (`strings.IndexByte`, `strconv.Itoa`)
**Files**: `database_test.go`, `security_headers_test.go`  
**Issue**: Testing Go's stdlib, not project code. Remove.

### M-8. `ErrorResponse` type is dead code
**File**: `internal/platform/errors/errors.go`  
**Issue**: `ToHTTPResponse` returns `ErrorResponse` but `WriteErrorJSON` uses anonymous struct. Pick one.

### M-9. `getRequiredTemplates()` unexported dead code
**File**: `internal/platform/templates/validator.go`  
**Issue**: Never called. Remove.

### M-10. N+1 query patterns in page handlers
**Files**: `post/adapters/http_handler_page.go`, `comment/adapters/http_handler_page.go`  
**Issue**: `getPostCategories` called per-post in loops; per-comment user/reaction lookups. Batch into single queries with `WHERE id IN (...)`.

### M-11. `post_create.html` and `post_edit.html` nearly identical
**Files**: `templates/post_create.html`, `templates/post_edit.html`  
**Issue**: ~80% shared structure. Use a single `post_form.html` partial with conditional logic for create vs edit.

### M-12. `home.html` and `board.html` post card duplication
**Files**: `templates/home.html`, `templates/board.html`  
**Issue**: Post card HTML repeated. Extract to a `_post_card.html` partial.

### M-13. Setter injection creates nil-dependency window
**File**: `cmd/forum/wire/services.go`  
**Issue**: `SetNotificationService()` called after construction means services temporarily have nil dependencies. Use constructor injection or builder pattern.

### M-14. `panic()` in `initServer` for upload dir failure
**File**: `cmd/forum/wire/app.go`  
**Issue**: Every other init function returns errors; this one panics. Return `(*httpserver.Server, error)` for consistency.

### M-15. Stale Dockerfile comment
**File**: `Dockerfile`  
**Issue**: Comment says "Go 1.25" but image is `golang:1.24-alpine`.

### M-16. ~80 lines of legacy avatar-column fallback code
**File**: `internal/modules/user/adapters/sqlite_repository.go`  
**Issue**: Dead code handling pre-migration schema. The migration has been applied. Remove.

### M-17. Avatar URL constructed in 3 different places
**Files**: `user/adapters`, `post/adapters/http_handler_page.go`, `comment/adapters`  
**Issue**: `"/static/uploads/" + path` repeated. Extract to `domain.AvatarURL(path) string`.

### M-18. Empty logger key `""` in main.go
**File**: `cmd/forum/main.go`  
**Issue**: Produces malformed structured log output.

### M-19. CORS `*` not configurable
**File**: `internal/platform/httpserver/middleware.go`  
**Issue**: `Access-Control-Allow-Origin: *` is hardcoded. Should come from config for production deployments.

---

## Summary Table

| Priority | Count | Theme |
|----------|-------|-------|
| **P0 — Security** | 4 | XSS bypass, missing authz, cookie security, ID leak |
| **P1 — Correctness** | 7 | Silent no-ops, mock drift, missing data, wrong error handling |
| **P2 — DRY** | 11 | Cross-module duplication, intra-module duplication |
| **P2 — KISS** | 10 | Over-abstraction, fat interfaces, unnecessary complexity |
| **P2 — Idiomatic Go** | 8 | Error handling, naming, deprecated APIs |
| **P3 — Minor** | 19 | Dead code, N+1 queries, template duplication, build issues |

## Recommended Action Order

1. **Fix P0 security issues** (S-1 through S-4) — small, targeted fixes
2. **Fix P1 correctness bugs** (C-1 through C-7) — prevent data loss and silent failures
3. **Extract shared helpers** (D-1 through D-6) — biggest DRY win across modules
4. **Consolidate duplicate handlers** (D-7 through D-11) — ~400 lines saved
5. **Replace logger with slog** (K-1) — ~500 lines eliminated
6. **Remove dead code** (M-6, M-8, M-9, M-16, C-2) — reduce maintenance surface
7. **Standardize error patterns** (G-1 through G-5) — consistency across codebase
8. **Template partials** (M-11, M-12) — DRY up HTML
9. **Address remaining KISS** (K-2 through K-10) — simplify abstractions
10. **Minor fixes** (M-1 through M-19) — polish

---

## Estimated Impact

| Metric | Current | After Review |
|--------|---------|-------------|
| Total Go LOC (est.) | ~12,000 | ~10,500 (-12%) |
| Duplicated patterns | ~30 instances | ~8 instances |
| Dead code files/functions | ~15 | 0 |
| External dependencies | Minimal | Minimal (unchanged) |
| Test mock duplication | ~6 copies | 1 shared package |
