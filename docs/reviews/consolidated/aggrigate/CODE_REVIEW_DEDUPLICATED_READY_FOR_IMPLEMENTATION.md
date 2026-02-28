# Code Review — Simplifications & Optimizations (Deduplicated, Implementation-Ready)

> Date: 2026-02-28  
> Scope: Full codebase (`internal/modules/`, `internal/platform/`, `cmd/forum/`, `templates/`, `static/`)  
> Sources merged:  
> - `CODE_REVIEW.md`  
> - `aggregate-review-2026-02-28.md`  
> - `code-review-kiss-idiomatic-go-2026-02-28.md`  
> - `code-simplifier-modules-consolidated.md`  
> Principles: Idiomatic Go, KISS, minimal surface area, strong compile-time guarantees, INT-internal / UUID-public invariant.

---

## Legend

| Severity | Meaning |
|---|---|
| 🔴 **Bug** | Incorrect behavior, security issue, or data integrity risk |
| 🟠 **Major** | Significant complexity, race condition risk, dead code, or architectural drift |
| 🟡 **Minor** | Duplication, maintainability issue, stale code path |
| 🟢 **Nit** | Trivial cleanup |

---

## 1) Dead Code & Unused Surface (Delete)

### 1.1 `internal/platform/database/transaction.go` unused abstraction
🟠 `Transaction`, `Begin`, `Commit`, `Rollback`, `Tx()` are not used by repositories.
**Action**: Delete file.

### 1.2 `internal/platform/database/migrator.go` unused `Rollback()` stub
🟡 Public method returns `"rollback not yet implemented"` and is never called.
**Action**: Delete method to reduce false API surface.

### 1.3 `internal/platform/templates/validator.go` unused validation path
🟠 `TemplateValidator` call path is unused in production; `getRequiredTemplates()` has no active consumer.
**Action**: Delete unused file/functions.

### 1.4 `internal/platform/templates/registry.go` global parser unused
🟠 `templates.Get()`/registry path is not used by module handlers.
**Action**: Delete file; keep one template-loading path (see §5.2).

### 1.5 `internal/modules/auth/adapters/middleware.go` deprecated helpers
🟠 Deprecated wrappers (`RequireAuthFunc`, `OptionalAuthFunc`, `GetUserID`, `GetUsername`, `IsAuthenticated`) are unused.
**Action**: Remove deprecated functions.

### 1.6 `internal/platform/httpserver/server.go` unused `RegisterHandler`
🟡 Redundant with Go 1.22 method-scoped mux patterns and not called.
**Action**: Delete method.

### 1.7 `internal/platform/errors/errors.go` unused typed error model
🟡 Structured error codes/types are not used by handlers; handlers rely on `WriteErrorJSON`.
**Action**: Remove unused types/constants; keep `WriteErrorJSON` + logger.

### 1.8 `internal/platform/config/config.go` unused OAuth config surface
🟡 OAuth structs loaded/validated but feature is not implemented.
**Action**: Remove OAuth config loading/validation until feature exists.

### 1.9 `internal/platform/config/config.go` inert logger config fields
🟡 Logger options are parsed but not wired to logger creation.
**Action**: Either wire fully or remove inert fields (prefer remove now).

### 1.10 `internal/platform/httpserver/health.go` unused `HealthPageConfig.Templates`
🟡 Field is assigned in wiring but ignored by health page code.
**Action**: Remove field and wire assignment.

### 1.11 `internal/modules/post/adapters/http_handler_api.go` dead fallback filter builder
🟠 Private fallback path for `nil filterService` is unreachable in DI setup.
**Action**: Delete fallback; make `filterService` required.

---

## 2) Correctness & Security Bugs (Fix First)

### 2.1 🔴 CORS invalid wildcard+credentials combination
**File**: `internal/platform/httpserver/middleware.go`  
`Access-Control-Allow-Credentials: true` must not be emitted with wildcard origin (`*`).
**Fix**: Set credentials header only when returning a specific allowed origin.

### 2.2 🔴 Rate limiter cleanup goroutine leak
**File**: `internal/platform/httpserver/middleware.go`  
Cleanup goroutine cannot be stopped from shutdown path.
**Fix**: Add explicit lifecycle (`Stop()`) and hook to server shutdown/context cancellation.

### 2.3 🔴 `LogoutPage` ignores secure cookie config
**File**: `internal/modules/auth/adapters/http_handler_page.go`  
Cookie clear uses `Secure: false` instead of configured `h.secureCookies`.
**Fix**: Use `Secure: h.secureCookies`.

### 2.4 🔴 User avatar not returned for email/username lookups
**File**: `internal/modules/user/adapters/sqlite_repository.go`  
`GetByEmail`/`GetByUsername` select list omits `avatar_path`.
**Fix**: Reuse same select projection as ID/publicID queries.

### 2.5 🔴 Upload path boundary check bypass risk
**File**: `internal/platform/upload/image.go`  
Path prefix checks can accept sibling-trick paths.
**Fix**: Use `filepath.Rel` boundary validation and reject rel paths that escape root (`..`).

### 2.6 🔴 Notification mark-read must enforce ownership in SQL
**Files**: notification ports/application/repository  
Mark-read path must always include requesting `user_id` ownership.
**Fix**: Method contract `MarkAsRead(userID, notificationPublicID)` + SQL `WHERE public_id=? AND user_id=?`.

### 2.7 🟠 Startup readiness race in server boot
**File**: `internal/platform/httpserver/server.go`  
`Start()` can return success before bind failure surfaces.
**Fix**: `net.Listen` synchronously, then serve in goroutine.

### 2.8 🟠 Wrong error type on missing reaction target
**File**: `internal/modules/reaction/adapters/sqlite_repository.go`  
Missing post/comment returns reaction-not-found, conflating target absence.
**Fix**: Return dedicated target-not-found sentinel.

### 2.9 🟠 Migrations not atomic per file
**File**: `internal/platform/database/migrator.go`  
Migration apply and migration-record should be in one transaction.
**Fix**: Wrap execution and schema_migrations insert in one tx per migration.

### 2.10 🟠 Moderation service false-success stubs
**File**: `internal/modules/moderation/application/service.go`  
Unimplemented methods return `nil`.
**Fix**: Return explicit not-implemented errors.

### 2.11 🟠 TOCTOU race in reaction toggle
**File**: `internal/modules/reaction/application/service.go`  
Read-delete-create sequence is non-transactional.
**Fix**: Transactional toggle or SQL upsert conflict strategy.

### 2.12 🟠 Non-deterministic validation error selection
**File**: `internal/modules/user/application/service.go`  
Map iteration order causes random error return when multiple fields fail.
**Fix**: Deterministic priority checks (email, username, ...).

### 2.13 🟠 Comment create timestamp inconsistency
**File**: `internal/modules/comment/adapters/sqlite_repository.go`  
Service timestamp and DB timestamp differ (`CURRENT_TIMESTAMP` vs Go time).
**Fix**: Persist service-provided `CreatedAt`/`UpdatedAt` values explicitly.

### 2.14 🟠 Malformed migration files silently skipped
**File**: `internal/platform/database/migrator.go`  
Missing `-- +migrate Up` marker is silently ignored.
**Fix**: Warn or fail for malformed `.sql` migration files.

### 2.15 🟡 Strict JSON content-type equality checks
**Files**: auth/comment handlers  
Exact `application/json` string check rejects valid `application/json; charset=utf-8`.
**Fix**: Parse media type (`mime.ParseMediaType`) and compare normalized type.

### 2.16 🟡 `RegisterAPI` brittle error-message string matching
**File**: `internal/modules/auth/adapters/http_handler_api.go`  
Fallback status by string-contains couples behavior to message text.
**Fix**: Remove fallback; unknown errors return 500.

### 2.17 🟡 `DeleteCommentAPI` contract inconsistent with REST style
**File**: `internal/modules/comment/adapters/http_handler_api.go`  
Returns 200 body while other delete endpoints return 204.
**Fix**: Standardize to 204 No Content.

### 2.18 🟡 Empty comments response serializes as `null`
**File**: `internal/modules/comment/adapters/http_handler_api.go`  
Nil slice JSON output is `null`.
**Fix**: Initialize empty slice for deterministic `[]` response.

---

## 3) Duplication to Consolidate (DRY)

### 3.1 Shared auth context helpers duplicated across modules
**Files**: post/comment handlers  
`buildCurrentUser` and `getInternalUserID` duplicated.
**Action**: Move to shared adapter helper in platform.

### 3.2 Repeated fire-and-forget stat update goroutine pattern (6 sites)
**Files**: post/comment/reaction services  
Same timeout+log wrapper repeated.
**Action**: Add small shared `asyncUpdate` helper (or switch to bounded synchronous update).

### 3.3 `RequireAuth` / `OptionalAuth` mostly duplicated flows
**File**: auth middleware  
Same cookie/session/user resolution with different failure behavior.
**Action**: Extract private `resolveAuth(required bool)`.

### 3.4 Reaction target ID resolution copied across repository methods
**File**: reaction sqlite repository  
`SELECT id FROM posts/comments WHERE public_id=?` repeated many times.
**Action**: Extract `resolveTargetID(ctx, publicID, targetType)` helper.

### 3.5 Post list APIs duplicate auth+filter parsing logic
**File**: post API handler  
`ListPostsAPI` and `LoadMorePostsAPI` near-identical.
**Action**: Single handler/path or shared parser/service call path.

### 3.6 `HomePage` and `BoardPage` duplicate full rendering logic
**File**: post page handler  
Session/filter/pagination/template logic duplicated.
**Action**: Extract `renderPostListPage(...)` helper.

### 3.7 User repository select constants duplicated
**File**: user sqlite repository  
Multiple near-identical select projections.
**Action**: Collapse to one canonical select query constant.

### 3.8 Repeated avatar URL prefix magic string
**Files**: user application/repository/handlers  
`"/static/uploads/"` duplicated.
**Action**: Centralize as one shared constant.

### 3.9 Static JS duplicated fetch/error plumbing
**Files**: `static/js/auth.js`, `post-forms.js`, `post-detail.js`  
Each reimplements request/error parsing.
**Action**: Shared `api.request(...)` utility in `static/js/main.js`.

### 3.10 Template duplication (post cards, load-more buttons, sidebar blocks)
**Files**: `templates/home.html`, `board.html`, `base.html`, `comments.html`  
Near-identical markup repeated.
**Action**: Extract reusable template partials and invoke from pages.

---

## 4) Simplification & Idiomatic Go

### 4.1 Remove legacy avatar-column fallback machinery
**File**: user sqlite repository  
String-matched missing-column compatibility path is obsolete after migration 008 baseline.
**Action**: Remove legacy fallbacks and direct-query only.

### 4.2 Split oversized post service file
**File**: `internal/modules/post/application/service.go`  
`Service` and `CategoryService` in one long file.
**Action**: Move `CategoryService` to `category_service.go`.

### 4.3 Reduce overlapping post filter types
**Files**: post domain/filter service  
`FilterParams` and `PostFilter` overlap heavily.
**Action**: Consolidate or clearly separate minimal responsibilities.

### 4.4 Logger file too large and mixed responsibilities
**File**: `internal/platform/logger/logger.go`  
Core logger mixed with pretty formatting and coloring concerns.
**Action**: Extract formatting/color helpers into `logger/pretty.go`.

### 4.5 Logger message-keyword color selection fragile
**File**: logger  
Color by `strings.Contains(msg, "error")` is semantically wrong.
**Action**: Color by level only.

### 4.6 Validator repeatedly sanitizes same field across chained rules
**File**: `internal/platform/validator/validator.go`  
Every rule sanitizes again.
**Action**: Sanitize once per field, then validate rules.

### 4.7 Remove redundant `SanitizeHTML` alias
**File**: validator  
Alias adds no behavior beyond `Sanitize`.
**Action**: Remove alias and keep single exported API.

### 4.8 Clarify username validation semantics
**File**: validator  
Current regex requires capitalized proper-name format.
**Action**: Decide display-name vs handle semantics; if handle, update regex accordingly.

### 4.9 Health checker repeats module route-check pattern
**File**: `internal/platform/health/checker.go`  
Seven copy-paste blocks for endpoint checks.
**Action**: Data-driven module definitions + loop.

### 4.10 Moderation health status hardcoded down
**File**: health checker  
Both true/false map branches return `"down"`.
**Action**: Use normal up/down mapping or omit until implemented.

### 4.11 Upload handler performs duplicate MIME detection
**File**: upload image handler  
`ValidateImage` and `Save` both detect type.
**Action**: Return MIME from validation and reuse.

### 4.12 Upload handler runs `MkdirAll` per save
**File**: upload image handler  
Per-request filesystem setup is unnecessary.
**Action**: Ensure directory once in constructor.

### 4.13 Config path validation too rigid
**File**: config  
Hardcoded accepted DB/upload paths reject valid custom deployments.
**Action**: Validate shape (`.db`, non-empty, sane path), not exact values.

### 4.14 Env parse failures silently fall back to defaults
**File**: `internal/platform/config/env_parser.go`  
Misconfigured env vars are ignored silently.
**Action**: Warn or fail-fast on malformed explicit env values.

### 4.15 Replace post-construction mutation hooks with constructor DI
**Files**: comment/reaction services + wiring  
`SetNotificationService(...)` can be forgotten and bypass compile-time guarantees.
**Action**: Inject optional notification dependency via constructor.

### 4.16 Consolidate template parsing into one wiring path
**Files**: wire + health + templates package  
Templates parsed in multiple places.
**Action**: Parse once in wiring, inject parsed template set everywhere.

### 4.17 Unexport internal-only function
**File**: auth application service  
`ValidateCredentials` exported but package-internal usage only.
**Action**: rename to `validateCredentials`.

### 4.18 Title max mismatch between validation and error text
**Files**: post domain post/errors  
Validation uses 255, error says 300.
**Action**: one constant `MaxTitleLength`, shared everywhere.

### 4.19 Duplicate `Author` and `AuthorUsername` fields
**File**: post domain model + consumers  
Both carry same data.
**Action**: keep one field, update templates/tests.

### 4.20 Misleading reaction target validation condition
**File**: reaction domain  
`TargetID <= 0 && PublicTargetID == ""` effectively only checks public ID.
**Action**: validate explicit public ID directly.

### 4.21 Untyped permission constants
**File**: user domain  
String constants lose compile-time safety.
**Action**: define `type Permission string`.

### 4.22 Redundant HTTP method guards in method-scoped routes
**Files**: multiple handlers  
Go 1.22 mux already enforces methods.
**Action**: remove manual guards.

### 4.23 Empty logger field key in startup log
**File**: `cmd/forum/main.go`  
`logger.String("", urls)` uses blank key.
**Action**: use `"urls"` key.

### 4.24 `panic` in wiring path instead of returning error
**File**: `cmd/forum/wire/app.go`  
Inconsistent with error-returning init flow.
**Action**: return wrapped error.

### 4.25 Remove stale review artifact comments in production code
**File**: `internal/platform/database/connection.go`  
`KISS-*`/`NIT-*` markers are no longer useful.
**Action**: delete artifact comments.

---

## 5) Performance Hotspots

### 5.1 N+1 in post list category hydration
**File**: post sqlite repository  
Per-post category query in list loop.
**Fix**: batch category lookup for all post IDs or aggregate in main query.

### 5.2 N+1 in comment activity page enrichment
**File**: comment page handler  
Repeated post/reaction enrichment per comment.
**Fix**: prefetch unique dependencies once per request into maps.

### 5.3 Redundant target existence fetch before reaction delete
**File**: reaction service  
Prefetch before delete adds query with little value.
**Fix**: remove prefetch and rely on repository-level delete result semantics.

### 5.4 Notification unread count computed in Go over full list
**File**: notification repository/handler  
Unread count loop over full payload is inefficient.
**Fix**: SQL count query or query-projected unread metric.

### 5.5 Comment API author ID enrichment still per-author querying
**File**: comment API handler + repository  
Handler does repeated user lookups.
**Fix**: join users in comments query and return `author_public_id` directly.

### 5.6 Per-request logger creation in post handlers
**File**: post API handler  
`logger.NewWithConfig(...)` inside request path.
**Fix**: inject logger into handler once.

### 5.7 Frontend stylesheet serial loading due to `@import`
**File**: `static/css/style.css` + `templates/base.html`  
Many `@import` lines cause serial fetch behavior.
**Fix**: use explicit `<link rel="stylesheet">` tags in base template.

### 5.8 Full-page reload after comment create/edit
**File**: `static/js/post-detail.js`  
`location.reload()` after actions hurts UX and bandwidth.
**Fix**: update/insert DOM nodes directly from API response.

---

## 6) Frontend Maintainability (Templates/JS/CSS)

### 6.1 Move large inline HTML strings out of JS into `<template>`
**Files**: `static/js/load-more-posts.js`, `static/js/load-more-comments.js`, page templates  
Long `innerHTML` literals are hard to lint/test and drift from server templates.
**Action**: define template blocks in HTML and clone/fill in JS.

### 6.2 Base layout controlled by title string comparisons
**File**: `templates/base.html`  
Layout selected via many `if eq .Title "..."` branches.
**Action**: pass `.Layout` from handlers and branch on layout enum.

### 6.3 Health table hard-coded service-name `if` chain
**File**: `templates/health.html` + health handler  
Display mapping embedded in template logic.
**Action**: pass structured `[]HealthItem` from handler and `range` directly.

### 6.4 Inconsistent error container ID conventions
**Files**: multiple templates + JS callers  
Mixed `page-errors` vs `form-errors` conventions without shared partial.
**Action**: standardize and extract reusable partial.

### 6.5 Hardcoded footer year
**File**: `templates/base.html`  
Year becomes stale.
**Action**: pass `CurrentYear` in base data model.

### 6.6 Duplicate button style definitions in cards CSS
**File**: `static/css/cards.css` vs shared buttons styles  
Same button styles repeated.
**Action**: reuse shared button classes; keep only local layout overrides.

---

## 7) Tests & Safety Net Gaps

### 7.1 🔴 Reaction service test mocks are stale/out-of-contract
**File**: `internal/modules/reaction/ports/service_test.go`  
Mocks use old signatures and miss methods from current interface.
**Action**: update mocks to exact current interface and assertions.

### 7.2 Add high-value regression tests for fixed bugs
**Areas**: CORS wildcard+credentials, migrator transaction atomicity, upload path boundary, rate limiter stop lifecycle, logger `Error(nil)` no panic, health optional-module behavior.
**Action**: add focused tests near changed code paths only.

---

## 8) Implementation Order (Ready-to-Execute)

### P0 — Correctness/Security (Do first)
1. §2.1 CORS wildcard+credentials
2. §2.3 secure cookie in logout page
3. §2.4 avatar query completeness
4. §2.5 upload path boundary check
5. §2.6 notification ownership for mark-read
6. §7.1 stale reaction mocks

### P1 — Data Integrity & Concurrency
1. §2.9 migration atomicity per file
2. §2.11 reaction toggle transaction/upsert
3. §2.7 deterministic server startup bind
4. §2.2 rate limiter stoppable lifecycle

### P2 — Performance/Hot Paths
1. §5.1 post category N+1 removal
2. §5.2 comment activity N+1 removal
3. §5.5 comment author public ID join
4. §5.4 unread count in SQL

### P3 — Maintainability & Simplification
1. §1 dead code removals
2. §3 duplication consolidation
3. §4 idiomatic/KISS cleanups
4. §6 template/js/css DRY cleanup

---

## 9) Constraints for Implementation

- Preserve architecture boundaries (`domain`, `ports`, `application`, `adapters`) and flat `adapters/` layout.
- Keep UUIDs on all external surfaces; never expose internal INT IDs.
- Avoid introducing new external dependencies for these fixes.
- Prefer incremental commits per section (P0/P1/P2/P3), each with targeted tests.
- Do not change API contracts unless explicitly listed above.

---

## 10) Completion Definition

A section is complete only when all items in that section satisfy all of:
1. Code path simplified/fixed as specified.
2. No unused replacement abstractions left behind.
3. Existing tests pass, and targeted regression tests added for bug fixes.
4. No UUID/INT exposure regressions in templates, JSON, routes, or context.
5. Startup and health endpoints remain functional.
