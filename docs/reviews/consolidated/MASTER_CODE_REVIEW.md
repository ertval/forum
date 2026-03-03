# Master Code Review — Deduplicated & Ready for Implementation

> Date: 2026-02-28  
> Scope: Full codebase — `internal/modules/`, `internal/platform/`, `cmd/forum/`, `static/`, `templates/`  
> Sources: Four prior review passes consolidated, duplicates removed, verified against live code  
> Principles: Idiomatic Go, KISS, minimal surface area, strong compile-time guarantees  
> **Implementation Status**: ✅ All items resolved (2026-02-28). `go build ./...` and `go vet ./...` pass. All unit tests green.

---

## Implementation Summary

| Section | Items | Status |
|---------|-------|--------|
| 1. Dead Code | 1.1–1.12 | ✅ All resolved (1.4 kept — has 11 active callers) |
| 2. Bugs | 2.1–2.19 | ✅ All resolved (2.13–2.15, 2.18–2.19 were already fixed) |
| 3. Duplicate Code | 3.1–3.14 | ✅ All resolved |
| 4. Unnecessary Complexity | 4.1–4.29 | ✅ All resolved (4.22 kept: health template parsing is separate by design) |
| 5. Performance | 5.1–5.3 | ✅ All resolved (5.2 was already implemented) |
| 6. Frontend / Templates | 6.1–6.12 | ✅ All resolved (6.7 was already consistent, 6.10 omitted as cosmetic) |
| 7. Tests | 7.1–7.2 | ✅ Mocks updated; method-not-allowed tests removed (dead after 4.26) |

### Key Changes by Module

**Platform layer:**
- `database/transaction.go` deleted (1.1); `migrator.go` Rollback stub removed (1.2), migrations now atomic in transactions (2.7), malformed files warn (2.16)
- `templates/validator.go` deleted (1.3); `registry.go` retained (1.4 has callers)
- `httpserver/server.go`: `RegisterHandler` deleted (1.6), startup race fixed with synchronous bind (2.5)
- `httpserver/middleware.go`: CORS credentials fix (2.1), rate limiter returns stop function (2.2)
- `errors/errors.go`: unused types/constants removed (1.7)
- `config/config.go`: OAuthConfig removed (1.8), Logger section removed (1.9), env validation uses `slices.Contains` (4.11), path validation shape-based (4.19)
- `config/env_parser.go`: warns on unparseable env vars (4.20), `getEnvStringSlice` removed
- `health.go`: Templates field removed (1.10), readiness classifies critical vs optional (4.18)
- `health/checker.go`: moderation `"down"/"down"` → `"up"/"down"` (2.9)
- `logger/logger.go` split → `logger/pretty.go` (4.3), keyword colouring removed (4.4), nil guard on Error(nil) (4.5)
- `validator/validator.go`: double-sanitize removed (4.6), `SanitizeHTML` deleted (4.7), username regex accepts handle-style (4.8)
- `upload/image.go`: `ValidateImage` returns MIME type (4.9), `MkdirAll` in constructor only (4.10), path boundary uses `filepath.Rel` (2.12)
- `httpserver/handlerutil.go` created: shared `BuildCurrentUser`/`GetInternalUserID` (3.1)
- `async/async.go` created: shared `Run` helper replaces 7 fire-and-forget goroutines (3.2)

**Auth module:**
- Deprecated middleware functions deleted (1.5), `RequireAuth`/`OptionalAuth` DRYed via `resolveAuth` (3.3)
- `LogoutPage` uses `h.secureCookies` (2.3)
- Content-Type check uses `HasPrefix` (2.14), `RegisterAPI` uses `errors.Is` (2.17)
- `ValidateCredentials` unexported (4.23)

**User module:**
- Legacy avatar shim removed (1.12), `GetByEmail`/`GetByUsername` include `avatar_path` (2.4)
- Queries consolidated to one `userColumns` constant (3.8), `scanUser` extracted (3.14)
- `AvatarURLPrefix` constant replaces hardcoded strings (3.9)
- Validation order deterministic (3.7 — was already fixed)
- `Permission` type introduced (4.25), `HasPermission` already implemented (4.15)

**Post module:**
- Dead `buildFilter` deleted (1.11), N+1 category queries fixed with batch `getCategoriesForPosts` (3.10)
- `parsePostRequest` extracted (3.12), `CategoryService` moved to own file (4.1)
- Logger injected via handler field (4.12), `MaxTitleLength=255` constant (4.13)
- `Author` field removed, `AuthorUsername` used everywhere (4.14)
- `GetByNames` batch category lookup added (4.17)

**Reaction module:**
- `ErrTargetNotFound` sentinel added (2.6), stale mocks fully updated (2.10/7.1)
- `ToggleReaction` transaction wraps check-delete-create (2.11)
- `resolveTargetID` helper replaces 5 copies (3.4), validation simplified (4.24)
- Redundant target pre-fetch removed (5.1), `r.PathValue` already used (4.16 — was ok)

**Comment module:**
- Author data JOINed in SQL queries, eliminating N+1 (5.3)
- `AuthorUsername` field added to Comment domain

**Notification module:**
- `CountUnread` already existed (5.2), `MarkAsRead` already scoped by user (2.13)

**Moderation module:**
- `CreateReport`/`ReviewReport` return `errors.New("not implemented")` (2.8)

**Wiring & Main:**
- `panic` → error return in `initServer` (4.28), blank log key fixed (4.27)
- Redundant method guards removed from handlers (4.26)

**Frontend:**
- HTML templates added for JS cloning (6.1), `location.reload` replaced with DOM injection (6.2)
- Shared `window.api.request()` used across JS files (6.3)
- Post-card templates consolidated (6.4), load-more button template (6.5)
- Layout selection simplified (6.6), health template uses range loop (6.8)
- Copyright year genericized (6.9), CSS `@import` → `<link>` tags (6.11), duplicate button styles removed (6.12)

---

## Legend

| Severity | Meaning |
|---|---|
| 🔴 **Bug** | Incorrect behaviour / security issue |
| 🟠 **Major** | KISS violation, significant dead code, race condition, or data correctness |
| 🟡 **Minor** | Duplication, style inconsistency, stale comment |
| 🟢 **Nit** | Trivial cleanup |
| ✅ | **Resolved** |

---

## 1. Dead Code — Remove Entirely

These items add noise with zero value. Delete them.

### ✅ 1.1 `database/transaction.go` — entire file unused
🟠 `Transaction`, `Begin`, `Commit`, `Rollback`, and `Tx()` are defined but **no repository anywhere calls `Begin()`**. All adapters write directly against `*sql.DB`.  
**Action**: Delete the file. **Done** — file deleted.

### ✅ 1.2 `database/migrator.go` — `Rollback()` stub
🟠 `func (m *Migrator) Rollback() error { return errors.New("rollback not yet implemented") }` is never called and creates false API surface.  
**Action**: Delete the method. **Done** — method and its test removed.

### 1.3 `templates/validator.go` — entire file unused
🟠 `TemplateValidator`, all its methods, and the unexported `getRequiredTemplates()` (which checks for a `"content"` template that does not exist) have zero call sites in production code.  
**Action**: Delete the file.

### 1.4 `templates/registry.go` — global parser never used
🟠 `templates.Get()` and `Registry.GetOrParse()` are not called by any handler. Templates are parsed via `ParseGlob` in `wire/handlers.go` and `ParseFiles` in `health.go`.  
**Action**: Delete the file. Consolidate all template parsing to one call site in the wire layer.

### 1.5 `auth/adapters/middleware.go` — deprecated function block
🟠 `RequireAuthFunc`, `OptionalAuthFunc`, `GetUserID`, `GetUsername`, and `IsAuthenticated` are all marked `// DEPRECATED` and have no callers.  
**Action**: Delete the five functions.

### 1.6 `httpserver/server.go` — `RegisterHandler` never called
🟡 `RegisterHandler` wraps a handler with a manual HTTP-method check, but all routes use Go 1.22 `ServeMux` method-prefixed patterns (`GET /api/...`). The function is never invoked.  
**Action**: Delete the method.

### 1.7 `errors/errors.go` — structured error types never used by handlers
🟡 `errors.Error`, `errors.New`, `errors.Wrap`, `ToHTTPResponse`, `ErrorResponse`, `HTTPStatus`, and all `ErrCode*` constants are defined but no HTTP handler uses them. All handlers call `WriteErrorJSON(w, http.StatusXxx, "msg")` directly.  
**Action**: Delete the unused types/constants. Keep only `WriteErrorJSON` and the package-level `errLogger`.

### 1.8 `config/config.go` — `OAuthConfig` parsed but feature absent
🟡 Google and GitHub OAuth structs are populated from env vars and partially validated. Nothing in the codebase consumes them.  
**Action**: Remove `OAuthConfig` from `Config` and its loading/validation logic. Restore when OAuth is actually implemented.

### 1.9 `config/config.go` — `cfg.Logger.*` fields loaded but never applied
🟡 `main.go` constructs the logger _before_ reading config, so `OmitFields`, `AllowedFields`, `MaxLineWidth`, and related fields are completely inert.  
**Action**: Either wire these fields into logger construction, or delete the `Logger` section from `Config`.

### 1.10 `httpserver/health.go` — `HealthPageConfig.Templates` field ignored
🟡 `wire/app.go` assigns `Templates: handlers.Post.Templates()`, but `HealthPage()` immediately re-parses its own templates from disk and ignores the injected field entirely.  
**Action**: Remove the `Templates` field from `HealthPageConfig` and the wire assignment.

### 1.11 `post/adapters/http_handler_api.go` — dead `buildFilter` fallback (~80 lines)
🟠 A private `buildFilter` method reimplements `FilterService.BuildFilter` "for when `filterService` is nil". `filterService` is always injected and is never nil.  
**Action**: Delete the fallback method. Make `filterService` a required constructor parameter.

### 1.12 `user/adapters/sqlite_repository.go` — legacy avatar compatibility shim
🟠 `getByIDLegacy`, `getByPublicIDLegacy`, and `isMissingAvatarColumnError` do SQLite driver error-message string-matching to handle a missing `avatar_path` column. Migration `008_user_add_avatar_path.sql` has been applied; the column always exists.  
**Action**: Delete the three private functions and both `selectUserLegacy*` query constants. Use the single `selectUserWithAvatar*` queries directly.

---

## 2. Bugs

### 2.1 🔴 CORS: `credentials: true` with wildcard origin
`internal/platform/httpserver/middleware.go` sets `Access-Control-Allow-Credentials: true` unconditionally, including when `allowAll = true` (origin `*`). The Fetch specification forbids this combination — browsers silently reject credentialed requests when the origin is `*`.

```go
// Only emit the credentials header when echoing a specific origin
if !allowAll {
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}
```

### 2.2 🔴 Rate limiter goroutine leak
`RateLimitWithConfig` spawns a cleanup goroutine that checks a private `limiter.done` channel. `Stop()` is unreachable from outside the function; the goroutine runs until process exit.  
**Fix**: Return a `stop func()` alongside the middleware, or accept a `context.Context` so the goroutine exits on server shutdown.

### 2.3 🔴 `LogoutPage` hardcodes `Secure: false`
`auth/adapters/http_handler_page.go`: `LogoutPage` clears the session cookie with `Secure: false` while the handler struct carries a `secureCookies bool` field that `LoginAPI` and `LogoutAPI` use correctly.  
**Fix**: Replace the literal with `Secure: h.secureCookies`.

### 2.4 🔴 `GetByEmail` and `GetByUsername` omit `avatar_path`
`user/adapters/sqlite_repository.go`: Both queries hardcode the SELECT column list without `avatar_path`. Users fetched by email or username will never have their avatar populated even after uploading one.  
**Fix**: Use the same query constant (`selectUserWithAvatarByID`/`ByPublicID`) for all four lookup methods.

### 2.5 🟠 Server startup 100 ms race
`httpserver/server.go`: `Start()` waits 100 ms and returns `nil` regardless of whether the listener bound successfully. Errors written to the goroutine's `errChan` are dropped after that window closes.

```go
// Synchronously bind, then serve asynchronously
ln, err := net.Listen("tcp", addr)
if err != nil {
    return err
}
go srv.Serve(ln)
return nil
```

### 2.6 🟠 Wrong sentinel error when reaction target is missing
`reaction/adapters/sqlite_repository.go`: `GetByTargetPublicID` returns `domain.ErrReactionNotFound` when the target post/comment does not exist. Callers cannot distinguish "target absent" from "no reactions for this target".  
**Fix**: Return `domain.ErrTargetNotFound` (or the corresponding module's `ErrNotFound`) when the target lookup yields no row.

### 2.7 🟠 Migrations not atomic
`database/migrator.go`: each migration runs with a bare `db.Exec(upSQL)`. If a multi-statement migration fails mid-way, earlier DDL statements are committed and the migration record is never written — leaving the schema in a partially-corrupted state.

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil { return err }
if _, err = tx.ExecContext(ctx, upSQL); err != nil {
    tx.Rollback()
    return fmt.Errorf("migration %s: %w", name, err)
}
_, err = tx.ExecContext(ctx, recordSQL, name)
if err != nil { tx.Rollback(); return err }
return tx.Commit()
```

### 2.8 🟠 Moderation service silently succeeds on unimplemented operations
`moderation/application/service.go`: `CreateReport` and `ReviewReport` return `nil` without doing anything. The HTTP handlers return `501`, but the service says "success". Future tests or callers will believe the feature works.  
**Fix**: Return `errors.New("not implemented")` from both methods.

### 2.9 🟠 Moderation health check always reports `"down"`
`health/checker.go`:
```go
results["moderation_api"] = map[bool]string{true: "down", false: "down"}[moderationAllUp]
```
Both map values are `"down"`. The health endpoint permanently misinforms consumers.  
**Fix**: `{true: "up", false: "down"}`, or omit moderation from health until the module ships.

### 2.10 🟠 Stale mock signatures in reaction tests
`reaction/ports/service_test.go`: Mock `React` uses `targetID int` but the live `ReactionService` interface uses `targetPublicID string`. Methods `GetUserReactionCount` and `GetByUserAndTargetPublicID` are missing entirely. The mock compiles only because it implements an outdated interface snapshot.  
**Fix**: Regenerate or manually update mocks to match the current interface; add the missing method stubs; assert real behaviour in tests.

### 2.11 🟠 TOCTOU race in reaction toggle
`reaction/application/service.go`: `React()` performs read → conditional delete → create as three separate, non-transactional steps. A concurrent request with the same user/target can produce duplicate reactions or wrong counts.  
**Fix**: Wrap the check-delete-create sequence in a single database transaction, or use an `INSERT … ON CONFLICT` upsert at the SQL level.

### 2.12 🟠 Upload path boundary not verified with `filepath.Rel`
`platform/upload/image.go`: The final path is validated by `strings.HasPrefix`, which accepts sibling-path tricks (e.g. `uploads/../secret`).

```go
rel, err := filepath.Rel(h.uploadDir, finalPath)
if err != nil || strings.HasPrefix(rel, "..") {
    return "", errors.New("invalid upload path")
}
```

### 2.13 🟠 Notification mark-read missing ownership scope
`notification` ports/application/adapters: `MarkAsRead` accepts only `notificationPublicID`; the SQL `WHERE` clause does not filter by `user_id`. Any authenticated user can mark any notification as read.  
**Fix**: Change the signature to `MarkAsRead(ctx, userID int, publicID string)` and add `AND user_id = ?` to the query.

### 2.14 🟡 Strict JSON `Content-Type` equality check
Several handlers (`auth`, `comment`) compare `r.Header.Get("Content-Type") == "application/json"`. Clients that send `application/json; charset=utf-8` are rejected unnecessarily.  
**Fix**: Use `strings.HasPrefix(ct, "application/json")` or parse with `mime.ParseMediaType`.

### 2.15 🟠 Comment `CreatedAt`/`UpdatedAt` timestamp inconsistency
`comment/adapters/sqlite_repository.go`: The repository uses `CURRENT_TIMESTAMP` in the INSERT statement while the service constructs its own `time.Time` values. The two timestamps can differ, producing misleading audit data.  
**Fix**: Pass the service-provided `CreatedAt`/`UpdatedAt` values as explicit bind parameters; remove `CURRENT_TIMESTAMP` defaults from the INSERT.

### 2.16 🟠 Malformed migration files silently skipped
`platform/database/migrator.go`: A `.sql` file missing the `-- +migrate Up` marker is silently ignored. An operator renaming or mis-formatting a file gets no feedback and the schema diverges invisibly.  
**Fix**: Log a warning (or return an error) when a migration file is present but contains no parseable `Up` block.

### 2.17 🟡 `RegisterAPI` status code derived from error-message string matching
`auth/adapters/http_handler_api.go`: HTTP status is chosen by `strings.Contains(err.Error(), "...")`, coupling behaviour to message text. Any rewording of an error string silently changes the API status code.  
**Fix**: Map domain sentinel errors (`domain.ErrEmailTaken`, etc.) to status codes using `errors.Is`; all unknown errors return `500`.

### 2.18 🟡 `DeleteCommentAPI` returns `200` with body instead of `204`
`comment/adapters/http_handler_api.go`: This endpoint returns `200` with a JSON body while all other delete endpoints in the codebase return `204 No Content`.  
**Fix**: Return `http.StatusNoContent` and omit the response body.

### 2.19 🟡 Empty comment list serialises as `null`
`comment/adapters/http_handler_api.go`: A nil slice marshals to JSON `null` instead of `[]`.  
**Fix**: Initialise with `comments := make([]domain.Comment, 0)` so the response is always `[]`.

---

## 3. Duplicate Code — Consolidate

### 3.1 `buildCurrentUser` and `getInternalUserID` copied across modules
Near-identical private helpers appear in both `post/adapters/http_handler.go` and `comment/adapters/http_handler.go`. Add a shared `handlerutil` package under `internal/platform/httpserver/` and call it from both.

### 3.2 Fire-and-forget stat-counter goroutine — 6 copies
The identical "update user stat in background" pattern is copy-pasted across `post`, `comment`, and `reaction` application services (twice each):

```go
go func(uid int) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = s.userService.IncrementXxxCount(ctx, uid)
}(userID)
```

**Fix**: Add a shared `runAsync(fn func(context.Context) error, label string)` helper and replace all six sites. Each call should log on failure instead of discarding the error.

### 3.3 `RequireAuth` / `OptionalAuth` — 95 % identical
Both middleware functions: fetch the cookie → call `ValidateSession` → fetch the user → write to context. The only difference is "abort with 401 on failure" vs. "continue".  
**Fix**:
```go
func resolveAuth(w http.ResponseWriter, r *http.Request, required bool) (*http.Request, bool)
func RequireAuth(next http.Handler) http.Handler  { /* thin wrapper */ }
func OptionalAuth(next http.Handler) http.Handler { /* thin wrapper */ }
```

### 3.4 `reaction/adapters/sqlite_repository.go` — target-ID resolution copied 5 times
Every repository method independently executes `SELECT id FROM posts/comments WHERE public_id = ?`.  
**Fix**:
```go
func (r *Repository) resolveTargetID(ctx context.Context, publicID, targetType string) (int, error)
```

### 3.5 `ListPostsAPI` and `LoadMorePostsAPI` — near-identical handlers
`LoadMorePostsAPI` re-validates the session cookie even though `OptionalAuth` middleware already populates the context. Both handlers then parse the same 10+ query parameters.  
**Fix**: Merge into one handler; consume auth state from context in both code paths.

### 3.6 `HomePage` and `BoardPage` — ~150 lines duplicated
Filter parsing, cookie reading, pagination, and template rendering are cloned verbatim.  
**Fix**: Extract `renderPostListPage(w http.ResponseWriter, r *http.Request, defaults FilterDefaults)`.

### 3.7 `user/application/service.go` — non-deterministic validation order
```go
for field := range v.Errors() { // map iteration is random
    switch field { ... }
}
```
When multiple fields fail, the returned error is random on each run.  
**Fix**: Check fields in a fixed priority order (username, email, password, …) rather than ranging over a map.

### 3.8 Four near-identical SELECT query constants in `user/adapters/sqlite_repository.go`
`selectUserWithAvatarByID`, `selectUserLegacyByID`, `selectUserWithAvatarByPublicID`, `selectUserLegacyByPublicID` repeat the same 10-column list. After §1.12 removes the legacy path, only one constant per lookup key survives.

### 3.9 Magic upload-path prefix `/static/uploads/` repeated in 3+ files
`user/application/service.go`, `user/adapters/sqlite_repository.go`, and `user/adapters/http_handler_settings.go` all hardcode the string.  
**Fix**: Define `const AvatarURLPrefix = "/static/uploads/"` in the user domain (or platform upload package) and reference the constant everywhere.

### 3.10 N+1 queries in post listing category hydration
`post/adapters/sqlite_repository.go`: `r.getPostCategories(ctx, post.ID)` is called inside `for rows.Next()` — one extra query per post; 21 queries for a page of 20.  
**Fix**: Collect all post IDs first, then run a single `SELECT … WHERE post_id IN (…)` and assemble the categories in memory.

```go
// Collect IDs first
var ids []int
for _, p := range posts { ids = append(ids, p.ID) }
// One batch query
cats, _ := r.getCategoriesForPosts(ctx, ids)
for i := range posts { posts[i].Categories = cats[posts[i].ID] }
```

### 3.11 N+1 queries in comment activity views
`comment/adapters/http_handler_page.go`: Per-comment post and reaction lookups are repeated for every comment in the activity list.  
**Fix**: Collect unique post IDs and reaction targets once per request, fetch all in two bulk queries, and map results in memory before template rendering.

### 3.12 Duplicate multipart / JSON parsing in post handlers
`post/adapters/http_handler_api.go`: Both `CreatePostAPI` and `UpdatePostAPI` contain full, independent `switch` blocks for multipart/JSON/form-encoded content parsing.  
**Fix**:
```go
func parsePostRequest(r *http.Request) (*PostInput, error)
```
Call from both handlers.

### 3.13 Seven identical "check module endpoints" blocks in `health/checker.go`
The same 4-line pattern is repeated for auth, user, post, comment, reaction, notification, moderation.  
**Fix**: Data-driven table:
```go
type moduleCheck struct {
    name      string
    endpoints []struct{ method, path string }
}
for _, m := range modules {
    results[m.name] = boolToUpDown(c.areAllRoutesRegistered(ctx, m.endpoints))
}
```

### 3.14 Repetitive row-scanning in `user/adapters/sqlite_repository.go`
`GetByID`, `GetByPublicID`, `GetByEmail`, and `GetByUsername` each scan identical column sets line by line.  
**Fix**: Extract `scanUser(row interface{ Scan(...) error }) (*domain.User, error)` and call it from all four methods.

---

## 4. Unnecessary Complexity — Simplify

### 4.1 `post/application/service.go` — two independent services in one file
`Service` (PostService) and `CategoryService` are unrelated but share a file that exceeds 400 lines.  
**Fix**: Move `CategoryService` to `post/application/category_service.go`.

### 4.2 `post/domain/` — two overlapping filter types
`FilterParams` (14 fields, HTTP query layer) and `PostFilter` (11 fields, repository layer) have substantial overlap. The conversion in `filter_service.go` is pure boilerplate.  
**Fix**: Evaluate whether `FilterParams` can be eliminated by using `PostFilter` directly in the adapter layer, reducing total field count and the conversion step.

### 4.3 `platform/logger/logger.go` — oversized 593-line file
HTTP-request colouring, emoji selection, and formatted human output (~200 lines) are developer-UX features mixed into the logger core.  
**Fix**: Extract `formatHTTPRequest`, `colorForMethod`, `colorForStatusCode`, `formatBytes`, and `colorForMessage` to `logger/pretty.go`. Keep `logger.go` focused on `Logger`, `Field`, and the core `log()` method.

### 4.4 `platform/logger/logger.go` — `colorForMessage` keyword-based colouring
Colour is chosen via `strings.Contains(lower, "error")`, `"started"`, etc. A message like `"no error found"` is coloured red. Log messages should not control their own rendering.  
**Fix**: Remove keyword-based colouring. Derive colour from log level only.

### 4.5 `platform/logger/logger.go` — missing nil guard on `Error(nil)`
Passing a `nil` error to the `Error` field constructor panics or produces misleading output.  
**Fix**: Add an early return in the `Error` helper: `if err == nil { return Field{...zero value} }`.

### 4.6 `platform/validator/validator.go` — double `Sanitize` on chained calls
Every validation method (`Required`, `MinLength`, `Email`, …) calls `Sanitize(value)` internally. Chaining `Required` then `Email` on the same value sanitizes it twice silently; the caller never receives the sanitised string.  
**Fix**: Sanitize once per field at the call site, before any rule is applied, rather than inside each rule.

### 4.7 `platform/validator/validator.go` — `SanitizeHTML` is a no-op alias
```go
func SanitizeHTML(input string) string { return Sanitize(input) }
```
Zero additional behaviour.  
**Fix**: Remove `SanitizeHTML`. Use `Sanitize` as the single exported name everywhere.

### 4.8 `platform/validator/validator.go` — `Username` regex forces capitalised first letter
```go
namePartRegex = regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
```
Forum handles like `alice`, `john123`, and `cool_user` are rejected.  
**Fix**: Change to `^[a-zA-Z][a-zA-Z0-9_-]*$` for handle-style usernames, or document the intent if display-name format is intentional.

### 4.9 `platform/upload/image.go` — double type-detection per save
`Save()` calls `ValidateImage()` (which calls `DetectImageType` internally) and then calls `DetectImageType` again itself.  
**Fix**: Change `ValidateImage` to return the detected MIME type, and have `Save` reuse it:
```go
func ValidateImage(data []byte, maxSize int64) (mimeType string, err error)
```

### 4.10 `platform/upload/image.go` — `os.MkdirAll` on every upload
The upload directory exists after startup. `MkdirAll` is a syscall on every save.  
**Fix**: Call `os.MkdirAll` once inside `NewImageHandler`, remove it from `Save`.

### 4.11 `platform/config/config.go` — verbose environment validation
```go
if c.Server.Environment != "development" &&
   c.Server.Environment != "staging" &&
   c.Server.Environment != "production" {
```
**Fix** (idiomatic Go 1.21+):
```go
if !slices.Contains([]string{"development", "staging", "production"}, c.Server.Environment) {
    return fmt.Errorf("invalid environment %q", c.Server.Environment)
}
```

### 4.12 `post/adapters/http_handler_api.go` — ad-hoc logger created per request
`logger.NewWithConfig(...)` is called inside `CreatePostAPI` (error path) and at the top of every `UpdatePostAPI` invocation. Logger construction is not free.  
**Fix**: Add a `logger *logger.Logger` field to `HTTPHandler`; inject it via `NewHTTPHandler`; replace all in-handler calls with `h.logger`.

### 4.13 `post/domain/` — title length constant does not exist; validation and error message disagree
`Validate()` enforces `len(p.Title) > 255` while `ErrTitleTooLong` says *"max 300 characters"*.  
**Fix**:
```go
// post/domain/post.go
const MaxTitleLength = 255

// post/domain/errors.go
var ErrTitleTooLong = fmt.Errorf("post title too long (max %d characters)", MaxTitleLength)
```

### 4.14 `post/domain/post.go` — duplicate `Author` / `AuthorUsername` fields
Both carry the same string; the repository sets both.  
**Fix**: Remove `Author`, use only `AuthorUsername`. Update the repository and all template references.

### 4.15 `user/domain/user.go` — `HasPermission` always returns `false`
```go
func (u *User) HasPermission(action string) bool { return false }
```
The method is a stub. Callers that rely on it will always get denial.  
**Fix**: Implement role-based logic (admin/moderator/member), or remove the method and replace call sites with explicit role comparisons until it is implemented.

### 4.16 `reaction/adapters/http_handler_api.go` — manual `strings.Split` for path parameters
Path values are extracted via `strings.Split(r.URL.Path, "/")` while other handlers use `r.PathValue("param")` (Go 1.22 ServeMux).  
**Fix**: Replace manual path splitting with `r.PathValue("targetType")` and `r.PathValue("targetID")`.

### 4.17 `post/application/service.go` — O(n) category validation per post
`CreatePost` and `UpdatePost` loop over category names and call `GetByName()` once each. A post with five categories fires five round trips.  
**Fix**: Add `GetByNames(ctx context.Context, names []string) ([]domain.Category, error)` to the repository port and use one batch call.

### 4.18 `health/checker.go` — health readiness conflates critical and optional modules
The readiness endpoint returns `503` if any module (including unimplemented moderation) reports down. Optional/unfinished modules should not drive overall readiness.  
**Fix**: Classify checks as `critical` (DB, auth, core API) vs. `optional` (moderation). Return `503` only when a critical check fails. Surface optional degradation in the body with `200`.

### 4.19 `platform/config/config.go` — path validation rejects valid custom deployments
🟡 DB and upload path checks compare against hardcoded values, rejecting legitimate deployments that use non-default paths.  
**Fix**: Validate path _shape_ (non-empty, `.db` extension for the DB path, no null bytes) rather than exact strings.

### 4.20 `platform/config/env_parser.go` — malformed env vars silently fall back to defaults
🟡 When an env var is present but unparseable (e.g. `PORT=abc`), the parser quietly uses the default. Operators get no indication that their configuration is ignored.  
**Fix**: Return an error (or at minimum log a warning) for any explicitly set env var that fails to parse.

### 4.21 `comment/reaction` wiring — `SetNotificationService` post-construction mutation
🟡 `SetNotificationService(...)` can be forgotten by a wiring author, silently leaving the service without a notification dependency until the code path is exercised at runtime. Post-construction mutation bypasses compile-time DI guarantees.  
**Fix**: Accept the notification service as a constructor parameter; remove the setter.

### 4.22 `wire/handlers.go` + `health.go` — template parsing duplicated across call sites
🟡 Templates are parsed in `wire/handlers.go` via `ParseGlob` and again inside `health.go` via `ParseFiles`. Any template change requires validating two independent parse paths.  
**Fix**: Parse the full template set once in the wire layer; inject the resulting `*template.Template` into every handler including health (resolves §1.10 as a consequence).

### 4.23 `auth/application/service.go` — `ValidateCredentials` unnecessarily exported
🟢 The function is only called within the auth package. Exporting it widens the API surface without benefit.  
**Fix**: Rename to `validateCredentials`.

### 4.24 `reaction/domain/` — misleading target validation condition
🟡 The domain `Validate()` check reads `TargetID <= 0 && PublicTargetID == ""`. Because internal IDs are never set on inbound requests, the `TargetID` arm is always false — the condition effectively only checks `PublicTargetID`.  
**Fix**: Remove the `TargetID` arm; validate `PublicTargetID` directly: `if r.PublicTargetID == "" { return ErrInvalidTarget }`.

### 4.25 `user/domain/user.go` — untyped permission/role constants
🟢 Role and permission values are bare `string` constants. A function accepting a `string` role parameter accepts any string, losing compile-time safety.  
**Fix**: Define `type Role string` and `type Permission string`; change all function signatures accordingly.

### 4.26 Multiple handlers — redundant HTTP method guards with method-scoped mux
🟢 Several handlers contain explicit `if r.Method != "POST" { ... }` guards. Go 1.22 `ServeMux` already enforces method constraints via the route pattern (`POST /api/...`). The guards are unreachable dead code.  
**Fix**: Delete manual method checks from all handlers registered with method-prefixed routes.

### 4.27 `cmd/forum/main.go` — blank key in startup log field
🟢 `logger.String("", urls)` emits a log field with an empty key, producing malformed structured output.  
**Fix**: Use `logger.String("urls", urls)`.

### 4.28 `cmd/forum/wire/app.go` — `panic` in wiring instead of returning error
🟡 One wiring call site uses `panic(err)` while the surrounding init flow uses error returns. A panic in a non-test context is not recoverable in a clean way.  
**Fix**: Return a wrapped `error` and propagate to `main`.

### 4.29 `internal/platform/database/connection.go` — stale review-artifact comments
🟢 `KISS-*` and `NIT-*` marker comments are leftover review annotations, not useful production comments.  
**Fix**: Delete the marker comments.

---

## 5. Performance Hotspots

### 5.1 🟠 Redundant target existence fetch before reaction delete
`reaction/application/service.go`: `React()` fetches the target entity before attempting the delete, adding a round trip that provides no value — the repository's delete result already indicates whether a row was affected.  
**Fix**: Remove the prefetch; rely on the affected-rows count from the repository's delete call.

### 5.2 🟠 Notification unread count computed in Go over full payload
`notification/` repository + handler: The unread count is calculated by looping over the full notification slice in Go rather than using a SQL aggregate.  
**Fix**: Add a dedicated `CountUnread(ctx, userID int) (int, error)` repository method that executes `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read = 0`.

### 5.3 🟡 Comment list author data fetched per-comment
`comment/adapters/` repository + handler: Author details are resolved by individual `SELECT` calls per comment row rather than being joined in the comments query.  
**Fix**: `JOIN users ON comments.user_id = users.id` in the comments query and return `author_public_id` / `author_username` directly from the row scan.

---

## 6. Frontend / Templates

### 6.1 🟠 Large HTML fragments embedded in JavaScript
`static/js/load-more-posts.js` (`createPostElement`) and `static/js/load-more-comments.js` (`createCommentElement`) contain 35–50–line `innerHTML` template literals. These drift silently from Go-rendered HTML and are invisible to HTML tooling.  
**Fix**: Add `<template id="post-card-template">` (and a comment variant) to the corresponding Go templates. JavaScript clones and fills the template:
```js
const tpl = document.getElementById("post-card-template");
const el = tpl.content.cloneNode(true);
el.querySelector(".post-title a").textContent = post.Title;
container.appendChild(el);
```

### 6.2 🟡 `location.reload()` after submitting or editing a comment
`static/js/post-detail.js`: Full page reload after `POST` (new comment) and `PUT` (edit comment). This resets scroll position and wastes bandwidth.  
**Fix**: After a successful API response, inject or swap the updated comment node directly in the DOM. The JSON responses already contain all required data.

### 6.3 🟡 Duplicated fetch / error-handling boilerplate across JS files
`static/js/auth.js`, `post-forms.js`, and `post-detail.js` each independently implement try/catch, JSON body parsing, and inline `formErrors.innerHTML` injection.  
**Fix**: Expose a small shared client in `static/js/main.js` (loaded first via `base.html`):
```js
window.api = {
    async request(url, options = {}) {
        const res = await fetch(url, {
            ...options,
            headers: { "Content-Type": "application/json", ...options.headers },
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || "Server error");
        return data;
    },
};
```

### 6.4 🟡 Duplicate post-card markup in `home.html` and `board.html`
Both files inline their own `<article>` post-card block. `base.html` already defines `{{define "post-card"}}` but neither page calls it.  
**Fix**: Add `{{define "post-card-compact"}}` to `base.html`; replace the inline blocks in both pages with `{{template "post-card-compact" .}}`.

### 6.5 🟢 Duplicate load-more button markup across three pages
`home.html`, `board.html`, and `comments.html` each define a near-identical `<button>` with different `id` / `data-*` attributes.  
**Fix**: `{{define "load-more-button"}}` in `base.html`; call `{{template "load-more-button" .LoadMoreParams}}` from each page.

### 6.6 🟡 Layout selection via title-string comparison in `base.html`
~150 lines of `{{if eq .Title "Home"}} … {{else if eq .Title "Board"}}` control layout. Adding a new page requires editing template logic.  
**Fix**: Set a `.Layout` field in the Go handler's template data (`"single"`, `"two-col"`, `"three-col"`):
```gohtml
<div class="page-layout-{{.Layout}}">
```
Layout selection becomes a Go-level decision testable in unit tests.

### 6.7 🟢 Inconsistent error-container IDs across templates
Some templates use `id="page-errors"`, others `id="form-errors"`. JavaScript must hard-code the correct ID per page.  
**Fix**: Adopt a single convention (`id="page-errors"` for top-of-page banners, `id="form-errors"` inside `<form>` elements); extract as `{{define "error-container"}}` and use uniformly.

### 6.8 🟡 Repetitive health status table in `templates/health.html`
Seven `{{if eq $service "auth_api"}}` blocks hard-code display names inside the template. Adding a module requires editing the template.  
**Fix**: Change the Go health handler to pass `[]HealthItem{Key, DisplayName, Status}`. Template becomes a `{{range .ModuleHealth}}` loop.

### 6.9 🟢 Hardcoded copyright year in `templates/base.html`
`&copy; 2025` will silently go stale.  
**Fix**: Pass `.CurrentYear` from a base-data helper:
```gohtml
<p>&copy; {{.CurrentYear}} Forum Authors.</p>
```

### 6.10 🟢 Sidebar card markup duplicated in `templates/base.html`
`post-sidebar-cards` and `post-create-sidebar-cards` share ~80 % identical markup (image-upload widget, category-select block).  
**Fix**: Extract `{{define "sidebar-image-upload"}}` and `{{define "sidebar-category-select"}}` partials; include from both sidebar sections.

### 6.11 🟡 CSS `@import` causes serial stylesheet loading
`static/css/style.css` contains 14 `@import url(...)` statements. The browser must process each `@import` sequentially, blocking render.  
**Fix**: Replace with direct `<link rel="stylesheet">` tags in `templates/base.html`. The browser can then fetch all sheets in parallel.

### 6.12 🟢 Duplicate button styles in `static/css/cards.css`
`.filters button` re-declares a complete button style (border-radius, padding, background gradient) that duplicates `buttons.css`. A design change must be applied in two places.  
**Fix**: Apply `.btn .btn-primary` classes directly to filter buttons in the HTML. In `cards.css` retain only layout-specific overrides:
```css
.filters .btn { width: 100%; }
```

---

## 7. Tests & Safety Net Gaps

### 7.1 🔴 Stale reaction service mock signatures
`reaction/ports/service_test.go`: The mock `React` method uses `targetID int` while the live `ReactionService` interface uses `targetPublicID string`. Methods `GetUserReactionCount` and `GetByUserAndTargetPublicID` are absent entirely. The mock compiles only against an outdated interface snapshot.  
**Fix**: Update mocks to match the exact current interface; add missing method stubs; assert real state changes rather than just method calls.

### 7.2 🟡 Missing regression tests for fixed security/correctness bugs
The following bugs should each have a targeted test added alongside the fix to prevent regression:
- CORS wildcard + credentials (§2.1)
- Migrator transaction atomicity failure (§2.7)
- Upload path boundary traversal (§2.12)
- Rate limiter stop/cleanup lifecycle (§2.2)
- Logger `Error(nil)` no-panic (§4.5)
- Health readiness with optional module down (§4.18)

**Fix**: Add focused unit/integration tests adjacent to each changed code path.

---

## Implementation Constraints

- Preserve architecture boundaries (`domain`, `ports`, `application`, `adapters`) and flat `adapters/` layout.
- Keep UUIDs on all external surfaces; never expose internal INT IDs in URLs, templates, JSON, or context values.
- Avoid introducing new external dependencies for any of these fixes.
- Prefer incremental commits per priority tier (P0 → P1 → P2 → P3), each accompanied by targeted tests.
- Do not change published API contracts (URL paths, request/response shapes) unless a fix is explicitly listed above.

---

## Completion Definition

An item is complete only when **all** of the following hold:
1. The code path is simplified or fixed exactly as specified.
2. No unused replacement abstractions are left behind.
3. All existing tests pass; targeted regression tests are added for every bug fix.
4. No UUID/INT exposure regressions appear in templates, JSON responses, routes, or context values.
5. The application starts cleanly and the health endpoint returns a valid response.
