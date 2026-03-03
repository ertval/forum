# Code Review — Simplifications & Optimizations

> Date: 2026-02-28  
> Scope: Full codebase (`internal/modules/`, `internal/platform/`, `cmd/forum/`)  
> Principles: Idiomatic Go, KISS, minimal surface area, strong compile-time guarantees

---

## Legend

| Severity | Meaning |
|---|---|
| 🔴 **Bug** | Incorrect behaviour / security issue |
| 🟠 **Major** | KISS violation, significant dead code, race condition, or data correctness |
| 🟡 **Minor** | Duplication, style inconsistency, stale comment |
| 🟢 **Nit** | Trivial cleanup |

---

## 1. Dead Code — Remove Entirely

These items add noise with zero value. Delete them.

### 1.1 `database/transaction.go` — entire file unused
🟠 `Transaction`, `Begin`, `Commit`, `Rollback`, `Tx()` are defined but **no repository anywhere calls `Begin()`**. All adapters write directly against `*sql.DB`.  
**Action**: Delete the file.

### 1.2 `database/migrator.go` — `Rollback()` stub
🟠 `func (m *Migrator) Rollback() error { return errors.New("rollback not yet implemented") }` is never called and creates false API surface.  
**Action**: Delete the method.

### 1.3 `templates/validator.go` — entire file unused
🟠 `TemplateValidator`, all its methods, and `getRequiredTemplates()` (which checks for a `"content"` template that doesn't exist) are never called.  
**Action**: Delete the file.

### 1.4 `templates/registry.go` — global parser never used
🟠 `templates.Get()` and `Registry.GetOrParse()` are not used by any handler; templates are parsed via `ParseGlob` in `wire/handlers.go` and `ParseFiles` in `health.go`.  
**Action**: Delete the file. Consolidate template parsing to one place (see §5.2).

### 1.5 `auth/adapters/middleware.go` — four deprecated functions
🟠 `RequireAuthFunc`, `OptionalAuthFunc`, `GetUserID`, `GetUsername`, `IsAuthenticated` are all marked `// DEPRECATED` at the bottom of the file and are not called anywhere.  
**Action**: Delete the four functions.

### 1.6 `httpserver/server.go` — `RegisterHandler` never called
🟡 `RegisterHandler` wraps a handler with a manual method check, but all routes use Go 1.22 `ServeMux` method-prefixed patterns (`GET /api/...`). The function is never invoked.  
**Action**: Delete the method.

### 1.7 `errors/errors.go` — structured error types never used by handlers
🟡 `errors.Error`, `errors.New`, `errors.Wrap`, `ToHTTPResponse`, `ErrorResponse`, `HTTPStatus`, and all `ErrCode*` constants are defined but no HTTP handler uses them. Handlers call `WriteErrorJSON(w, http.StatusXxx, "msg")` directly.  
**Action**: Delete the unused types. Keep only `WriteErrorJSON` and the package-level `errLogger`.

### 1.8 `config/config.go` — `OAuthConfig` loaded but feature not built
🟡 Both Google and GitHub OAuth structs are parsed from env and partially validated. Nothing in the codebase consumes them.  
**Action**: Remove `OAuthConfig` from `Config` and its loading/validation logic. Add back when OAuth is implemented.

### 1.9 `config/config.go` — `cfg.Logger.*` fields loaded but never applied
🟡 `main.go` creates the logger _before_ applying cfg, and config logger fields (`OmitFields`, `AllowedFields`, `MaxLineWidth`, etc.) are never passed to the logger. These config fields are completely inert.  
**Action**: Either wire the config into logger construction, or delete the `Logger` section from `Config`.

### 1.10 `httpserver/health.go` — `HealthPageConfig.Templates` field is never read
🟡 `wire/app.go` sets `Templates: handlers.Post.Templates()` but `HealthPage()` immediately re-parses its own templates and ignores this field.  
**Action**: Remove the `Templates` field from `HealthPageConfig` and the wire assignment.

### 1.11 `post/adapters/http_handler_api.go` — dead filter fallback (~80 lines)
🟠 Contains a private `buildFilter` method that re-implements `FilterService.BuildFilter` "for when `filterService is nil`". `filterService` is never nil (always injected).  
**Action**: Delete the fallback `buildFilter` method. Require `filterService` in the constructor.

---

## 2. Bugs

### 2.1 🔴 CORS: credentials + wildcard origin
`internal/platform/httpserver/middleware.go` sets `Access-Control-Allow-Credentials: true` unconditionally, including when `allowAll = true` (origin `*`). The Fetch specification forbids this combination — browsers reject credentialed cross-origin requests with a wildcard origin.  
**Fix**: Only set `Access-Control-Allow-Credentials: true` when reflecting a specific origin, not `*`.

```go
// Only set credentials header for non-wildcard origin responses
if origin != "" && allowAll == false {
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}
```

### 2.2 🔴 Rate limiter goroutine leak
`RateLimitWithConfig` spawns a cleanup goroutine that checks `limiter.done`. `limiter` is a private local variable; `limiter.Stop()` is unreachable. The goroutine runs indefinitely after application shutdown.  
**Fix**: Return the limiter from `RateLimitWithConfig` so callers can call `Stop()`, or hook `Stop()` into the server's shutdown sequence via context cancellation.

### 2.3 🔴 `LogoutPage` ignores `secureCookies` field
`auth/adapters/http_handler_page.go`: `LogoutPage` hardcodes `Secure: false` on the cookie clear, while the handler struct has a `secureCookies bool` field that `LoginAPI` and `LogoutAPI` use correctly.  
**Fix**: `Secure: h.secureCookies`.

### 2.4 🔴 `user/adapters/sqlite_repository.go` — `GetByEmail` and `GetByUsername` omit `avatar_path`
Both queries hardcode the SELECT field list without `avatar_path`. Users fetched by email or username will never have their avatar populated even if they uploaded one.  
**Fix**: Use the same query constant as `GetByID` / `GetByPublicID`.

### 2.5 🟠 `httpserver/server.go` — 100ms startup race
`Start()` returns `nil` after 100ms even if the listener fails later. Errors written to the goroutine's `errChan` are never consumed after `Start()` returns — they disappear.  
**Fix**: Call `net.Listen` synchronously before spawning the `http.Serve` goroutine:

```go
ln, err := net.Listen("tcp", addr)
if err != nil {
    return err
}
go srv.Serve(ln)
return nil
```

### 2.6 🟠 `reaction/adapters/sqlite_repository.go` — wrong error on missing target
`GetByTargetPublicID` returns `domain.ErrReactionNotFound` when the post/comment itself doesn't exist. Callers cannot distinguish "no such target" from "no reactions for this target".  
**Fix**: Return a sentinel `domain.ErrTargetNotFound` (or the post/comment's own `ErrNotFound`) when the target lookup fails.

### 2.7 🟠 Migrations not wrapped in transactions
`database/migrator.go`: each migration runs via bare `db.Exec(upSQL)`. If a multi-statement migration has two DDL statements and the second fails, the first is committed — partial schema corruption.  
**Fix**: Wrap each migration in a transaction:
```go
tx, _ := db.Begin()
_, err = tx.Exec(upSQL)
if err != nil { tx.Rollback(); return err }
tx.Commit()
```

### 2.8 🟠 Moderation service silently succeeds on unimplemented operations
`moderation/application/service.go`: `CreateReport` and `ReviewReport` return `nil` without doing anything. API handlers return `501`, but the service layer says "success". If the service is ever called directly (e.g., in future tests), it appears to work.  
**Fix**: Return `errors.New("not implemented")` from both methods so callers always know the feature is non-functional.

---

## 3. Duplicate Code — Consolidate

### 3.1 `buildCurrentUser` and `getInternalUserID` duplicated across modules
Identical or near-identical private methods in `post/adapters/http_handler.go` and `comment/adapters/http_handler.go`. A shared helper in `internal/platform/httpserver/` or a common `handlerutil` package would eliminate the copies.

```go
// present in BOTH post and comment http_handler.go:
func (h *HTTPHandler) buildCurrentUser(ctx, userID int) map[string]any { ... }
func (h *HTTPHandler) getInternalUserID(ctx, publicID string) (int, error) { ... }
```

### 3.2 Fire-and-forget user stat counter goroutine — 6 copies
The pattern below is copy-pasted identically in `post`, `comment`, and `reaction` application services (twice each):
```go
go func(uid int) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := s.userService.IncrementXCount(ctx, uid); err != nil {
        log.Printf("WARNING: ...")
    }
}(userID)
```
**Fix**: Add a `asyncUpdate(fn func(context.Context) error, label string)` helper to a shared location and call it from all six sites.

### 3.3 `RequireAuth` / `OptionalAuth` — 95% identical
Both middleware functions fetch the cookie, call `ValidateSession`, fetch the user, and write to context. The only difference is "continue on failure vs. return 401".  
**Fix**: Private `resolveAuth(w, r, required bool) bool` and thin public wrappers.

### 3.4 `reaction/adapters/sqlite_repository.go` — target ID resolution copied 5 times
Every repo method independently executes `SELECT id FROM posts/comments WHERE public_id = ?`. Extract:
```go
func (r *Repository) resolveTargetID(ctx context.Context, publicID, targetType string) (int, error)
```

### 3.5 `ListPostsAPI` and `LoadMorePostsAPI` are nearly identical
`LoadMorePostsAPI` manually reads the session cookie and re-validates it even though `OptionalAuth` middleware already populates the context. Both then parse the same 10+ query parameters.  
**Fix**: Unify into one handler; use context for auth in both paths.

### 3.6 `HomePage` and `BoardPage` are nearly identical
~150 lines of duplicated filter-parsing, cookie-reading, and template-rendering logic.  
**Fix**: Extract a shared `renderPostList(w, r, defaultLimit int)` function.

### 3.7 `user/application/service.go` — validation loop non-deterministic
```go
for field := range v.Errors() {  // map iteration order is random
    switch field {
    case "email":    return nil, ErrInvalidEmail
    case "username": return nil, ErrInvalidUsername
    }
}
```
If both email and username fail validation, the returned error is random.  
**Fix**: Check fields in a deterministic priority order (not a range over a map).

### 3.8 `user/adapters/sqlite_repository.go` — four near-identical SELECT constants
`selectUserWithAvatarByID`, `selectUserLegacyByID`, `selectUserWithAvatarByPublicID`, `selectUserLegacyByPublicID` repeat the same 10-column field list. After completing issue §4.1 (remove legacy path), only one constant is needed.

### 3.9 Magic string `/static/uploads/` repeated in 3+ files
`user/application/service.go`, `user/adapters/sqlite_repository.go`, `user/adapters/http_handler_settings.go` all hardcode this prefix when constructing `AvatarURL`.  
**Fix**: Define a single `const avatarURLPrefix = "/static/uploads/"` in the user domain or platform upload package.

---

## 4. Unnecessary Complexity — Simplify

### 4.1 `user/adapters/sqlite_repository.go` — legacy avatar column fallback
`isMissingAvatarColumnError` does string matching on SQLite driver messages. The `avatar_path` migration was applied in `008_user_add_avatar_path.sql`. This compatibility shim has been dead since migration 008 was applied.  
**Fix**: Remove `getByIDLegacy`, `getByPublicIDLegacy`, and `isMissingAvatarColumnError`. Use single direct queries.

### 4.2 `post/application/service.go` — two structs in one file
`Service` (PostService) and `CategoryService` both live in `service.go`. The file is long and the two services are independent.  
**Fix**: Move `CategoryService` to `category_service.go`.

### 4.3 `post/domain/` — two overlapping filter types
`FilterParams` (14 fields, from HTTP query) and `PostFilter` (11 fields, for repository) have substantial overlap. The conversion in `filter_service.go` is boilerplate.  
**Fix**: Evaluate whether `FilterParams` can be eliminated by using `PostFilter` directly, or reduce the field count of both.

### 4.4 `platform/logger/logger.go` — oversized file (593 lines)
HTTP request coloring, emoji selection, and formatted human output (~200 lines) are developer UX features mixed into the logger core.  
**Fix**: Extract `formatHTTPRequest`, `colorForMethod`, `colorForStatusCode`, `formatBytes`, and `colorForMessage` into a `logger/pretty.go` file, keeping `logger.go` focused on the `Logger`, `Field`, and core `log()` method.

### 4.5 `platform/logger/logger.go` — `colorForMessage` fragile keyword matching
Colors are chosen by `strings.Contains(lower, "error")`, `"started"`, etc. A log message saying `"no error found"` would be colored red. Log messages should not determine their own display color.  
**Fix**: Remove keyword-based coloring. Use only log level for color selection.

### 4.6 `platform/validator/validator.go` — double sanitize on chained calls
Every validation method (`Required`, `MinLength`, `Email`, etc.) calls `Sanitize(value)` internally. If a caller chains `Required` then `Email`, `Sanitize` runs twice silently. The caller never receives the sanitized string.  
**Fix**: Sanitize once at the call site (before any validation), or expose sanitize as a pre-step in the `Validator` struct and run it once per field, not per rule.

### 4.7 `platform/validator/validator.go` — `SanitizeHTML` is an alias for `Sanitize`
```go
func SanitizeHTML(input string) string { return Sanitize(input) }
```
Zero additional behavior.  
**Fix**: Remove `SanitizeHTML`, make `Sanitize` the single exported name (or vice-versa).

### 4.8 `platform/validator/validator.go` — `Username` regex requires capital first letter
```go
namePartRegex = regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
```
This enforces proper-name format (`Alice Smith`). Forum usernames like `alice`, `john123`, `cool_user` are rejected. This is almost certainly wrong for a forum.  
**Fix**: Clarify intended semantics with the team. If usernames are display names, current logic may be intentional. If they are handles, change to `^[a-zA-Z][a-zA-Z0-9_-]*$`.

### 4.9 `health/checker.go` — 7 identical "check module endpoints" blocks
Same 4-line pattern repeated 7 times: define endpoint slice → call `areAllRoutesRegistered` → set result in map.  

```go
// Repeated 7 times for auth, user, post, comment, reaction, notification, moderation
xyzEndpoints := []struct{ method, path string }{ ... }
xyzAllUp := c.areAllRoutesRegistered(ctx, xyzEndpoints)
results["xyz_api"] = map[bool]string{true: "up", false: "down"}[xyzAllUp]
```
**Fix**: Data-driven approach:
```go
type moduleHealth struct {
    name      string
    endpoints []struct{ method, path string }
}
modules := []moduleHealth{ {name: "auth_api", endpoints: [...]}, ... }
for _, m := range modules {
    up := c.areAllRoutesRegistered(ctx, m.endpoints)
    results[m.name] = boolToUpDown(up)
}
```

### 4.10 `health/checker.go` — moderation always reports "down"
```go
results["moderation_api"] = map[bool]string{true: "down", false: "down"}[moderationAllUp]
```
Both values are `"down"`. This permanently misleads the health endpoint.  
**Fix**: Either fix to `{true: "up", false: "down"}` or omit moderation from health until the module is implemented.

### 4.11 `upload/image.go` — double type detection per save
`Save()` calls `ValidateImage()` (which calls `DetectImageType`) then immediately calls `DetectImageType` again itself.  
**Fix**: Have `ValidateImage` return the detected MIME type so `Save` can reuse it:
```go
mimeType, err := ValidateImageAndGetType(data, h.maxSize)
```

### 4.12 `upload/image.go` — `MkdirAll` on every save
The upload directory almost certainly exists after startup. `MkdirAll` is a syscall on every upload operation.  
**Fix**: Call `MkdirAll` once in `NewImageHandler`, not in `Save`.

### 4.13 `config/config.go` — overly rigid path validation
```go
if !(c.Database.Path == "./data/forum.db" || c.Database.Path == "./db/forum.db" || dbBase == "forum.db") {
```
Blocks any custom database filename. Same pattern for upload directory.  
**Fix**: Require only that the path is non-empty and ends in `.db`:
```go
if !strings.HasSuffix(c.Database.Path, ".db") {
    return fmt.Errorf("database path must end in .db")
}
```

### 4.14 `env_parser.go` — silent parse failure on misconfigured env vars
`getEnvDuration`, `getEnvInt`, etc. silently use defaults on parse failure. `RATE_LIMIT_WINDOW=1hour` (space) is silently ignored.  
**Fix**: Log a warning when a set env var cannot be parsed:
```go
if value := os.Getenv(key); value != "" {
    d, err := time.ParseDuration(value)
    if err != nil {
        log.Printf("WARN: env %s=%q is not a valid duration, using default %v", key, value, defaultVal)
        return defaultVal
    }
    return d
}
```

---

## 5. Idiomatic Go Issues

### 5.1 Post-construction `SetNotificationService` mutation
`comment/application/service.go` and `reaction/application/service.go` both expose:
```go
func (s *Service) SetNotificationService(ns notificationPorts.NotificationService)
```
This is called in `wire/services.go` after construction. If this call is ever missed (e.g., a new test creates only the service), notification integration silently breaks with no compile error.  
**Fix**: Accept `notificationService` in both constructors. Use `nil` interface check if it's truly optional:
```go
func NewService(repo Repo, notif notificationPorts.NotificationService) *Service { ... }
```

### 5.2 Template parsing in three places
Templates are parsed in `wire/handlers.go` (via `template.ParseGlob`), in `httpserver/health.go` (via `template.ParseFiles`, ignoring the field from config), and `templates/registry.go` (never called).  
**Fix**: Parse once in `wire/handlers.go` and pass the result everywhere. Delete `registry.go` and `validator.go` from the templates package.

### 5.3 Exported function in application package leaks internals
`auth/application/service.go`: `ValidateCredentials` is an exported package-level function only called from within the same package.  
**Fix**: Rename to `validateCredentials`.

### 5.4 `post/domain/errors.go` — error message contradicts code
```go
ErrTitleTooLong = errors.New("post title too long (max 300 characters)")
```
But `Validate()` checks `len(p.Title) > 255`.  
**Fix**: Either change the constant to 300, or change the error message to "max 255 characters".

### 5.5 `post/domain/post.go` — `Author` alias field
```go
AuthorUsername string `json:"author_username,omitempty"`
Author         string `json:"author,omitempty"` // Alias for AuthorUsername (for compatibility)
```
Both are always set to the same value. Every API response sends this data twice.  
**Fix**: Remove `Author`. Fix the single consumer (template or test) that reads `.Author`.

### 5.6 `reaction/domain/reaction.go` — misleading `Validate()` condition
```go
if r.TargetID <= 0 && r.PublicTargetID == "" { return ErrInvalidTargetID }
```
`TargetID` is always 0 at creation time (filled by repo). The first condition is always true. The `&&` means this only validates `PublicTargetID`.  
**Fix**: Remove the `r.TargetID <= 0` clause; just check `r.PublicTargetID == ""`.

### 5.7 `user/domain/errors.go` — duplicate error definitions from auth module
`ErrInvalidEmail`, `ErrWeakPassword`, `ErrEmailAlreadyExists`, `ErrUsernameAlreadyExists` are defined in both `auth/domain/errors.go` and `user/domain/errors.go` with identical messages but different identity.  
If both modules are ever compared by callers, `auth.ErrEmailAlreadyExists != user.ErrEmailAlreadyExists`, which is surprising. Consider whether these belong in a shared `domain` package or if the split is intentional.

### 5.8 `user/domain/user.go` — permission constants are untyped strings
```go
const PermissionViewContent = "view"
```
Any `string` can be passed to `HasPermission`. A named type provides compile-time safety:
```go
type Permission string
const PermissionViewContent Permission = "view"
```

### 5.9 `comment/adapters/sqlite_repository.go` — `Create` discards `CreatedAt`
The service sets `comment.CreatedAt = time.Now()`, then the repository INSERT uses `CURRENT_TIMESTAMP` and ignores the Go time. The response timestamp is the Go time; the stored timestamp is SQLite's. These differ by milliseconds but the inconsistency is a correctness issue.  
**Fix**: Use the service-provided time in the INSERT:
```go
INSERT INTO comments (..., created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)
// pass comment.CreatedAt, comment.CreatedAt
```

### 5.10 `auth/adapters/http_handler_api.go` — brittle string-matching error fallback
After a `switch` on known domain errors, `RegisterAPI` falls back to string matching:
```go
if strings.Contains(errMsg, "empty") || strings.Contains(errMsg, "invalid") { ... }
```
Any future rename of an error message silently changes the HTTP status code returned.  
**Fix**: Remove the fallback. Unknown errors should unconditionally return 500.

### 5.11 `comment/adapters/http_handler_api.go` — delete returns 200, not 204
`DeleteCommentAPI` returns `200 {"message": "Comment deleted successfully"}` while `DeletePostAPI` returns `204 No Content`. REST convention is 204 for deletes.  
**Fix**: Standardize all delete endpoints to `204 No Content`.

### 5.12 `comment/adapters/http_handler_api.go` — null list on empty result
```go
var commentsResp []struct{ ... }  // nil slice → JSON null
```
**Fix**: `commentsResp := make([]struct{...}, 0)` → JSON `[]`.

### 5.13 Go 1.22 method-scoped routing makes HTTP method guards redundant
In several handlers (`GetPostAPI`, `ListPostsAPI`, `CreatePostAPI`, etc.) there are explicit guards:
```go
if r.Method != http.MethodGet { ... return }
```
Routes are registered as `GET /api/posts`, which already constrains the method.  
**Fix**: Remove these guards. The ServeMux returns 405 automatically for wrong-method requests.

### 5.14 `main.go` — empty logger key
```go
logger.String("", urls)  // key is ""
```
**Fix**: `logger.String("urls", urls)`.

### 5.15 `wire/app.go` — panic instead of error in `initServer`
```go
if err := os.MkdirAll(uploadDir, 0755); err != nil {
    panic(fmt.Sprintf("failed to create upload directory: %v", err))
}
```
All other `init*` functions return `error`.  
**Fix**: Return the error.

### 5.16 `database/migrator.go` — silent skip for malformed migration files
Files without the `-- +migrate Up` marker are skipped silently.  
**Fix**: Log a warning (or return an error) when a `.sql` file in the migrations directory lacks the expected marker.

### 5.17 `database/connection.go` — stale inline fix-tracking comments
`// KISS-1:`, `// NIT-6:` style comments are code-review artifact noise in production code.  
**Fix**: Remove them. If the fix is applied, the comment is worthless. If it isn't, open an issue.

### 5.18 `post/adapters/sqlite_repository.go` — `repeatPlaceholders` uses string concat loop
```go
for i := 0; i < count; i++ { result += ", ?" }
```
**Fix**: `strings.Repeat(", ?", count)` (or `strings.TrimPrefix(strings.Repeat(", ?", count), ", ")`).

---

## 6. Performance

### 6.1 N+1 queries in `post/adapters/sqlite_repository.go` — `List`
For every post in a list result, `getPostCategories` fires a separate `SELECT`:
```go
for rows.Next() {
    // scan post...
    categories, _ := r.getPostCategories(ctx, post.ID)  // ← 1 query per post
}
```
50 posts → 51 queries.  
**Fix**: After scanning all post IDs, fetch all categories in one `SELECT ... WHERE post_id IN (...)` and join in Go:
```go
postIDs := []int{ ... }
cats, _ := r.getCategoriesForPosts(ctx, postIDs)
// distribute cats to posts by post_id
```

### 6.2 `reaction/application/service.go` — target existence check before delete is redundant
`RemoveReaction` fetches the full post/comment to verify existence before calling `DeleteByTargetPublicID`. If the target doesn't exist, the reaction won't exist either, so the repository will return `ErrReactionNotFound`.  
**Fix**: Remove the pre-fetch. Let the repository return the appropriate error.

### 6.3 `notification/adapters/sqlite_repository.go` — unread count computed in Go
`GetUserNotifications` returns all notifications; the HTTP handler counts unread ones in a Go loop. A `WHERE is_read = 0` subquery or a separate `COUNT` query avoids transferring read notifications just to count unread ones.

### 6.4 `comment/adapters/http_handler_api.go` — per-author user lookup in N+1 loop
```go
for _, c := range comments {
    author, _ := h.userService.GetByID(r.Context(), c.UserID)
    c.PublicUserID = author.PublicID
}
```
A `userPublicIDs` cache reduces duplicate DB calls for the same user but still fires one query per unique author. The repository `ListByPost` query should JOIN `users` and return `author_public_id` directly.

---

## 7. Summary Table

| # | Category | Severity | File(s) |
|---|---|---|---|
| 1.1–1.11 | Dead code | 🟠/🟡 | transaction.go, templates/, middleware.go, etc. |
| 2.1 | CORS credential + wildcard | 🔴 | httpserver/middleware.go |
| 2.2 | Goroutine leak | 🔴 | httpserver/middleware.go |
| 2.3 | Secure cookie bug | 🔴 | auth/adapters/http_handler_page.go |
| 2.4 | Missing avatar in queries | 🔴 | user/adapters/sqlite_repository.go |
| 2.5 | Startup race condition | 🟠 | httpserver/server.go |
| 2.6 | Wrong error type | 🟠 | reaction/adapters/sqlite_repository.go |
| 2.7 | No-transaction migrations | 🟠 | database/migrator.go |
| 2.8 | Silent service success stub | 🟠 | moderation/application/service.go |
| 3.1–3.9 | Duplicate code | 🟠/🟡 | post, comment, reaction adapters / user |
| 4.1–4.14 | Unnecessary complexity | 🟠/🟡 | user repo, post service, logger, config, health |
| 5.1–5.18 | Idiomatic Go | 🟠/🟡/🟢 | widespread |
| 6.1–6.4 | Performance | 🟠/🟡 | post/comment/notification adapters |

**Recommended priority order**:
1. Fix all 🔴 bugs (§2.1–2.4) — correctness and security
2. Delete dead code (§1) — reduces surface area before any refactor
3. Fix 🟠 bugs (§2.5–2.8)
4. Consolidate duplicated logic (§3) — biggest readability gain
5. Simplify complexity (§4, §5)
6. Performance (§6) — after correctness is solid
