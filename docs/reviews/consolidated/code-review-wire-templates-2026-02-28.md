# Code Review: Wire, Entry Point, Templates, Build Files
## KISS & Idiomatic Go Analysis — 2026-02-28

---

## Table of Contents
1. [main.go](#1-maingo)
2. [wire/app.go](#2-wireappgo)
3. [wire/doc.go](#3-wiredocgo)
4. [wire/repositories.go](#4-wirerepositoriesgo)
5. [wire/services.go](#5-wireservicesgo)
6. [wire/handlers.go](#6-wirehandlersgo)
7. [go.mod](#7-gomod)
8. [Makefile](#8-makefile)
9. [Dockerfile](#9-dockerfile)
10. [Templates](#10-templates)
11. [Cross-Cutting Concerns](#11-cross-cutting-concerns)
12. [Summary Priority Matrix](#12-summary-priority-matrix)

---

## 1. main.go

**Structure:** 77 lines. Loads config, creates logger, calls `wire.InitializeApp`, starts server, blocks on SIGINT/SIGTERM, shuts down.

### Issues

#### 1.1 Logger level parsing is hard-coded to only "DEBUG" (L28-31)
```go
logLevel := logger.InfoLevel
if cfg.Logger.Level == "DEBUG" {
    logLevel = logger.DebugLevel
}
```
**Problem:** Any other level string (WARN, ERROR) is silently ignored. This is a latent bug.
**Fix:** Use a `ParseLevel(string) Level` function in the logger package, or at minimum a switch:
```go
logLevel, err := logger.ParseLevel(cfg.Logger.Level)
if err != nil {
    log.Fatalf("Invalid log level %q: %v", cfg.Logger.Level, err)
}
```

#### 1.2 Inconsistent error handling: `log.Fatalf` vs `lgr.Error` + `os.Exit(1)` (L23 vs L36-37)
- Line 23: `log.Fatalf` (stdlib logger) for config errors — because custom logger doesn't exist yet.
- Lines 36-37: `lgr.Error` + `os.Exit(1)` — because custom logger exists.

This is actually **correct** given the initialization order, but the `os.Exit(1)` pattern appears twice (L37, L43). Consider a tiny helper:
```go
fatal := func(msg string, err error) {
    lgr.Error(msg, logger.Error(err))
    os.Exit(1)
}
```

#### 1.3 URL construction with `fmt.Sprintf` is slightly wasteful (L47-48)
```go
httpAddr := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
httpsAddr := fmt.Sprintf("https://%s:%d", cfg.Server.Host, cfg.Server.TLSPort)
```
`httpsAddr` is always computed even when TLS is disabled. Minor, but `httpsAddr` could be computed inside the `if` block. Trivial optimization.

#### 1.4 Empty logger key on L54
```go
lgr.Info("Forum server started", logger.String("", urls))
```
The key `""` is an empty string — this produces poorly structured log output. Should be `logger.String("urls", urls)` or `logger.String("listen", urls)`.

#### 1.5 `fmt.Println()` on L61 is raw stdout, not going through the logger
This blank line between "signal received" and "Shutting down" breaks structured logging discipline.

### Verdict: **Good.** Clean, short, well-structured. Only minor polish needed.

---

## 2. wire/app.go

**Structure:** ~155 lines. Defines `App` struct, `InitializeApp` orchestrator, `initDatabase`, `initServer`.

### Issues

#### 2.1 `panic()` in `initServer` (L110) — KISS violation
```go
if err := os.MkdirAll(cfg.Upload.UploadDir, 0755); err != nil {
    lgr.Error("Failed to initialize upload directory", logger.Error(err))
    panic(fmt.Errorf("failed to initialize upload directory: %w", err))
}
```
**Problem:** Every other initialization function returns errors. This one panics, creating an inconsistent error-handling contract. The caller `InitializeApp` already handles errors gracefully.
**Fix:** Make `initServer` return `(*httpserver.Server, error)` and propagate the error.

#### 2.2 Health route registration breaks handler encapsulation (L130-146)
```go
server.Router().Handle("GET /health", httpserver.HealthPage(httpserver.HealthPageConfig{
    Checker:          healthChecker,
    Templates:        handlers.Post.Templates(),
    AuthFunc:         handlers.Auth.GetCurrentUser,
    GetUserWithStats: handlers.Post.GetUserWithStats,
}))
```
**Problems:**
- Reaches into `handlers.Post.Templates()` and `handlers.Post.GetUserWithStats` — cross-cutting accessor methods that leak implementation details.
- `handlers.Auth.GetCurrentUser` is accessed as a free function reference rather than through the service interface.
- Health configuration is scattered: the health HTTP handler is in `httpserver` package, the checker is in `health` package, and the wiring dips into two different module handlers.

**Fix:** Either create a dedicated health handler module following the same pattern as other modules, or move this config into `initHandlers` and return it as `Handlers.Health`.

#### 2.3 Hardcoded `"./static"` path (L151)
```go
if info, err := os.Stat("./static"); err == nil && info.IsDir() {
```
This should come from config (`cfg.Server.StaticDir` or similar) for consistency with other paths like `cfg.Upload.UploadDir`.

#### 2.4 Middleware registered by name but CORS allows all origins `"*"` (L125)
```go
server.RegisterMiddleware(httpserver.CORS([]string{"*"}))
```
CORS `*` is fine for development but should be configurable. Also, rate-limit parameters are extracted from config but CORS origins are hardcoded.

#### 2.5 Duplicate error logging in `initDatabase` (implicit)
The function logs errors AND returns them. The caller (`InitializeApp`) will also log or propagate. This can lead to double-logged errors. In Go, prefer: **return the error, let the caller decide how to log it.**

### Verdict: **Solid structure, but the `panic` and health-route wiring are notable violations.**

---

## 3. wire/doc.go

**Structure:** Package documentation only (23 lines).

**Verdict:** Good. Clear and useful. No issues.

---

## 4. wire/repositories.go

**Structure:** ~52 lines. Defines `Repositories` struct, `initRepositories` function.

### Issues

#### 4.1 Over-importing: port packages imported solely for type declarations
```go
import (
    authPorts "forum/internal/modules/auth/ports"
    commentPorts "forum/internal/modules/comment/ports"
    // ... 7 port packages
    // ... 7 adapter packages
)
```
14 imports for a single function that creates 8 structs. This is **inherent** to the hex architecture approach and is acceptable for a composition root, but it's worth noting the cognitive cost.

#### 4.2 `Repositories` fields typed as interfaces, constructed from concrete types
```go
type Repositories struct {
    Session      authPorts.SessionRepository  // interface
    // ...
}
// ...
Session: authAdapters.NewSQLiteSessionRepository(db), // returns concrete
```
This is **correct idiomatic Go** — accept interfaces, return structs. Fields typed as interfaces in the container allow swapping implementations. **No issue here.**

### Verdict: **Clean.** No actionable issues.

---

## 5. wire/services.go

**Structure:** ~100 lines. `ServiceContainer` with 11 private fields and 11 public accessor methods, plus `initServices`.

### Issues

#### 5.1 Setter-based initialization for circular deps is a smell (L89-90)
```go
reactionService.SetNotificationService(notificationService)
commentService.SetNotificationService(notificationService)
```
**Problem:** This means `reactionService` and `commentService` have a period after construction where they are in an **invalid state** (nil notification service). If any code path calls notification-related functionality before `Set*` is called, it will nil-pointer panic.

**Fix (KISS):** If notification is optional, use a functional option pattern or accept it in the constructor. If required, pass it in the constructor. The "Layer 2 depends on Layer 1" comment already documents the dependency — just wire it directly:
```go
reactionService := reactionApp.NewService(repos.Reaction, repos.Post, repos.Comment, userService, notificationService)
```

#### 5.2 11 accessor methods that are one-line getters (L53-64)
This is a lot of boilerplate. These exist to satisfy handler-local interfaces via interface segregation. This is **architecturally intentional** per the project's DI pattern.

However, an alternative KISS approach: since every handler interface is a subset, you could use a simpler pattern where each handler takes only the services it needs as constructor arguments, eliminating the container entirely:
```go
// Instead of:
authAdapters.NewHTTPHandler(services, templates, secureCookies)
// Consider:
authAdapters.NewHTTPHandler(authService, userService, templates, secureCookies)
```
This would eliminate `ServiceContainer`, all 11 accessors, and all 7 handler-local interface definitions. The trade-off is slightly longer constructor signatures, but **explicit is better than implicit** in Go.

That said, the current pattern does work and is well-documented. This is a design-philosophy choice, not a bug.

#### 5.3 `imageHandler` is created in services but it's infrastructure, not a service
```go
imageHandler := upload.NewImageHandler(cfg.Upload.UploadDir, cfg.Upload.MaxSize)
```
This is fine in practice. The layered comments correctly categorize it as "Layer 1b: Infrastructure adapters."

### Verdict: **Works well. The setter-injection pattern is the main concern.**

---

## 6. wire/handlers.go

**Structure:** ~57 lines. `Handlers` struct with 7 handler fields, `initHandlers` function.

### Issues

#### 6.1 Template parsing with `ParseGlob` is fragile (L40-44)
```go
templates, err = template.ParseGlob("templates/*.html")
if err != nil {
    return nil, err
}
```
**Problems:**
- Hardcoded path `"templates/*.html"` — should come from config.
- `ParseGlob` does not recurse into subdirectories. If templates are ever nested, they'll be silently skipped.
- No custom function map registered — the templates use `urlquery` (built-in) but if custom functions are ever needed, this is the place to add `template.New("").Funcs(funcMap).ParseGlob(...)`.

#### 6.2 Template error wrapping is missing context
```go
if err != nil {
    return nil, err  // raw error from template.ParseGlob
}
```
Should be:
```go
return nil, fmt.Errorf("parse templates: %w", err)
```

#### 6.3 `secureCookies` is only passed to `Auth` handler
```go
Auth: authAdapters.NewHTTPHandler(services, templates, secureCookies),
```
Other handlers get `(services, templates)` — consistent, but it means cookie concerns are isolated to auth. This is actually **correct** separation of concerns.

### Verdict: **Clean and simple. Minor hardcoding and error wrapping issues.**

---

## 7. go.mod

**Structure:** 4 dependencies (uuid, sqlite3, bcrypt) + Go 1.24.

### Issues

#### 7.1 Minimal dependency set — excellent
Only 3 direct dependencies. This is exemplary for a Go project.

#### 7.2 `go 1.24` but Dockerfile uses `golang:1.24-alpine`
Consistent. No issue.

#### 7.3 Missing `go.sum` in the reviewed files
Not a code issue — `go.sum` is auto-generated. Just noting it exists in the repo.

### Verdict: **Excellent. Textbook minimal dependencies.**

---

## 8. Makefile

**Structure:** 340 lines, 30+ targets.

### Issues

#### 8.1 Redundancy between `test` and `tests` targets (L63-98)
These two targets duplicate ~90% of their logic, differing only in quiet vs verbose output. This violates DRY.

**Fix:** Use a variable or mode argument:
```makefile
test:
	@$(MAKE) _run-tests QUIET=1

tests:
	@$(MAKE) _run-tests QUIET=0

_run-tests:
	# ... shared logic using $(QUIET) to control output
```

#### 8.2 `test` quietly swallows output then reruns on failure (L68-69)
```makefile
@$(GOTEST) ./... > /dev/null 2>&1 && echo "..." || (echo "..." && $(GOTEST) ./... && exit 1)
```
This **runs tests twice** on failure — once silently, once verbosely. Wasteful. Better to capture output to a temp file:
```makefile
@$(GOTEST) ./... > /tmp/test.out 2>&1 && echo "..." || (cat /tmp/test.out && exit 1)
```

#### 8.3 `BINARY_UNIX` is never used by any target other than `build-linux` (L15)
```makefile
BINARY_UNIX=bin/forum_unix
```
Only `build-linux` references it. If `build-linux` is rarely used, this is dead configuration. Minor clutter.

#### 8.4 `migration` target references a template that may not exist (L196)
```makefile
cp $(MIGRATIONS_DIR)/template/000_template_migration.sql $${filename};
```
If `migrations/template/000_template_migration.sql` doesn't exist, this fails silently. Should check for existence first.

#### 8.5 Color codes don't work on all terminals
```makefile
RED=\033[0;31m
```
Could use `tput` for portability, but ANSI codes are standard enough for development Makefiles.

#### 8.6 `all` target runs `mod` before `test` (L31)
```makefile
all: clean mod fmt vet test build
```
Running `mod download` and `mod verify` every time is slow. In CI this is fine; locally, `tidy` is usually sufficient. Consider making `mod` optional.

### Verdict: **Functional but has DRY violations. The double-test-on-failure pattern is the most impactful issue.**

---

## 9. Dockerfile

**Structure:** Multi-stage build (builder + runtime), 53 lines.

### Issues

#### 9.1 Comment says "Go 1.25" but uses `golang:1.24-alpine` (L6-7)
```dockerfile
# Use a stable Go 1.25 Alpine image for reproducible builds and security
FROM golang:1.24-alpine AS builder
```
**Bug:** Comment says 1.25, image says 1.24. Stale comment.

#### 9.2 Static linking with musl is correct but the `-a` flag is unnecessary with `-ldflags`
```dockerfile
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-w -s -extldflags '-static'" -o forum ./cmd/forum
```
The `-a` flag forces rebuilding all packages. This is rarely needed and slows builds. Only needed if switching between CGO_ENABLED states inside the same layer cache — unlikely in Docker.

#### 9.3 No `.dockerignore` referenced
Without `.dockerignore`, `COPY . .` copies everything including `.git/`, `bin/`, `data/`, docs, etc. This bloats the build context significantly.

**Fix:** Create `.dockerignore`:
```
.git
bin/
data/
docs/
reports/
*.md
full_test_output.txt
```

#### 9.4 `EXPOSE 8080` but the app may also serve TLS on another port
The config supports `TLSPort`. Should either expose both or document this.

#### 9.5 No `HEALTHCHECK` instruction
Docker Compose and orchestrators benefit from:
```dockerfile
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:8080/health-api || exit 1
```

### Verdict: **Good multi-stage build. Fix the stale comment and add `.dockerignore`.**

---

## 10. Templates

### 10.1 base.html (~300 lines)

#### Massive layout logic in template (L76-130)
The base template has a 5-branch conditional for layout selection:
```
if (showLeftSidebar AND showUserSidebar)        → three-col
else if (showUserSidebar AND isPostFormPage)     → three-col with post sidebar
else if (showUserSidebar)                        → right sidebar
else if (showSidebar)                            → left sidebar
else                                             → single column
```
**Problem:** This is business logic embedded in a template. 5 nested `if/else` branches make this hard to maintain and test.

**KISS Fix:** Use a single `.Layout` string field in the template data (`"three-col"`, `"right"`, `"left"`, `"single"`) and switch on it:
```html
{{if eq .Layout "three-col"}}
...
{{else if eq .Layout "right"}}
...
{{end}}
```
This moves the layout-selection logic to Go handlers where it can be tested.

#### User card duplicated between dropdown menu and sidebar
The nav dropdown (L30-67) and `user-card` component (L140-186) render nearly identical content: avatar, username, email, stats (posts/comments/reactions), and action links. This is **significant duplication** (~50 lines repeated).

**Fix:** Extract a shared `user-info` sub-template for the common parts.

#### Filter card is overly complex (L189-280)
The `filter-card` template has another large `if/else` block splitting between "My Activity" filters and board filters. Within each branch, there are hidden input fields for state preservation.

**Multiple hidden inputs for filter state** (L259-270):
```html
{{if .MyPosts}}<input type="hidden" name="my_posts" value="true">{{end}}
{{if .LikedPosts}}<input type="hidden" name="liked_posts" value="true">{{end}}
{{if .DislikedPosts}}<input type="hidden" name="disliked_posts" value="true">{{end}}
{{if .CommentedPosts}}<input type="hidden" name="commented_posts" value="true">{{end}}
```
These four booleans could be collapsed since `activity_type` select already represents the same state. Simplify: remove hidden inputs and derive from `activity_type`.

#### Post sidebar cards: edit and create are nearly identical
`post-sidebar-cards` and `post-create-sidebar-cards` share 80% of their HTML. The only difference is edit shows "current image" with remove button and has `checked` logic for categories.

**Fix:** Single template with conditional:
```html
{{define "post-sidebar-cards"}}
{{if and .Post .Post.ImageURL}}
  <!-- current image with remove -->
{{end}}
<!-- shared image upload UI -->
<!-- shared category checkboxes, with checked logic when .Post exists -->
{{end}}
```

### 10.2 home.html

#### Post card in `home.html` duplicates `board.html` post card
The home page uses `post-card-compact` classes while board uses `post-card` classes, but the structural HTML is nearly identical. Both have: title, meta (author, date), optional image, content, categories, reactions.

**KISS Fix:** Use the shared `post-card` template defined in `base.html` (which already exists!) in `board.html`, and create a `post-card-compact` variant that shares structure but applies different CSS. Or better — use the same template everywhere and control appearance via CSS class on the container.

### 10.3 health.html — most egregious KISS violation

#### Hardcoded service names instead of iterating (L1-110)
```html
{{range $service, $status := .Health}}
    {{if eq $service "auth_api"}}
        <tr><td>Authentication Module API</td>...
    {{end}}
    {{if eq $service "post_api"}}
        <tr><td>Post Module API</td>...
    {{end}}
    {{if eq $service "comment_api"}}
        ...
    {{end}}
    ...
{{end}}
```
This **iterates over the entire map 7+ times**, checking each key with `eq` individually. This is O(n×m) when it should be O(n).

**The "Other Services" section** (L83-107) repeats ALL service names in `ne` conditions:
```html
{{if and (ne $service "database") (ne $service "auth_api") (ne $service "post_api") (ne $service "user_api") (ne $service "comment_api") (ne $service "reaction_api") (ne $service "moderation_api") (ne $service "notification_api")}}
```

**Fix:** Structure the health data in Go, not in the template:
```go
type HealthData struct {
    CoreServices  []ServiceStatus  // [{Name: "Database", Status: "up"}]
    ModuleAPIs    []ServiceStatus  // [{Name: "Authentication Module API", Status: "up"}]
    OtherServices []ServiceStatus
}
```
Then the template becomes:
```html
{{range .CoreServices}}
<tr><td>{{.Name}}</td><td><span class="status-badge status-{{.Status}}">{{.Status}}</span></td></tr>
{{end}}
```
This reduces ~110 lines to ~30.

### 10.4 settings.html

#### Using `<article class="comment">` for non-comment content
```html
<article class="comment">
    <div class="comment-content">
        <form class="settings-form" ...>
```
Using comment styling for a settings form is a CSS hack. Should use a dedicated class.

#### No client-side validation for password match
The form has `new_password` and `confirm_password` but no JS to verify they match before submit. Settings page doesn't load any JS (`scripts` block is missing).

### 10.5 activity.html

#### Six nearly identical `{{range}}` blocks
Three sections (Created Posts, Reactions, Comments) each have: heading with link, `{{if .Items}}` `{{range .Items}}` article block `{{end}}` `{{else}}` empty message `{{end}}`.

The structure is sound but repetitive. This is acceptable for templates.

#### `HideCreatedPosts`, `HideReactions`, `HideComments` — negative booleans
```html
{{if not .HideCreatedPosts}}
```
**Idiomatic templates** prefer positive booleans: `ShowCreatedPosts`, `ShowReactions`, `ShowComments`. Negative booleans require mental double-negation.

### 10.6 comments.html

#### Missing owner check — all comments show edit/delete buttons
```html
<div class="comment-owner-actions">
    <button class="btn btn-secondary btn-edit-comment"...>Edit</button>
    <button class="btn btn-danger btn-delete-comment"...>Delete</button>
</div>
```
Unlike `post_detail.html` which checks `{{if eq $.User.PublicID .AuthorPublicID}}`, this template shows edit/delete for **every** comment without an ownership check. Either the server enforces this (which it should), or there's a UX issue where users see buttons that won't work.

**However**, this is the "My Comments" page which only shows the current user's comments, so all comments are owned by the user. Still, adding the guard is defensive best practice and documents intent.

### 10.7 post_create.html and post_edit.html

**Nearly identical.** The only differences are:
- Title: "Create New Post" vs "Edit Post"
- Form ID: `post-create-form` vs `post-edit-form`
- Edit has `data-post-id` and pre-filled `value`/`textarea`
- Submit button text differs
- Cancel link destination differs

Could be a single template parameterized by `.IsEdit`.

### 10.8 login.html and register.html

Clean and simple. No issues. Both correctly use the same `auth.js` script.

---

## 11. Cross-Cutting Concerns

### 11.1 No CSRF protection
Forms use `method="POST"` (settings, login, register) but there's no CSRF token. The API endpoints use cookies for auth, making them vulnerable to CSRF. The login/register forms POST to `/api/auth/*` URLs.

### 11.2 No Content Security Policy (CSP)
`SecurityHeaders` middleware is registered but CSP headers aren't mentioned. Inline scripts like emoji rendering could be affected.

### 11.3 Template data contracts are implicit
There's no Go struct type for template data visible in these files. Each handler presumably creates `map[string]interface{}` or ad-hoc structs. Without a defined contract, templates can reference fields that don't exist (causing runtime panics) with no compile-time safety.

**Improvement:** Define typed page data structs:
```go
type BoardPageData struct {
    Title            string
    User             *UserWithStats
    Posts            []PostSummary
    Categories       []Category
    SelectedCategory string
    // ...
}
```

### 11.4 Raw content rendering (XSS risk in post_detail.html)
```html
<div class="post-detail-content">
    {{.Post.Content}}
</div>
```
Go's `html/template` auto-escapes by default, so `{{.Post.Content}}` is safe. However, if anyone changes this to `{{.Post.Content | safeHTML}}` in the future without sanitization, it becomes an XSS vector. The current code is **safe but worth documenting**.

Comment content in multiple templates also uses `{{.Content}}` directly — same analysis applies.

---

## 12. Summary Priority Matrix

### Critical (Fix Now)
| # | File | Issue |
|---|------|-------|
| 1 | `wire/app.go` L110 | `panic()` in `initServer` — inconsistent with error-return pattern |
| 2 | `health.html` | O(n²) iteration with hardcoded service names — massive KISS violation |
| 3 | `Dockerfile` L6 | Stale comment says Go 1.25, image is 1.24 |

### High (Fix Soon)
| # | File | Issue |
|---|------|-------|
| 4 | `wire/services.go` L89-90 | Setter injection creates invalid-state window |
| 5 | `base.html` | User card info duplicated in dropdown and sidebar (~50 lines) |
| 6 | `base.html` L76-130 | 5-branch layout logic in template — move to Go |
| 7 | `comments.html` | Missing ownership guard on edit/delete buttons |
| 8 | `main.go` L54 | Empty logger key `""` produces malformed structured logs |

### Medium (Improve When Touching)
| # | File | Issue |
|---|------|-------|
| 9 | `main.go` L28-31 | Logger level only handles "DEBUG", ignores others |
| 10 | `Makefile` L63-98 | `test`/`tests` targets duplicate 90% of logic |
| 11 | `Makefile` L68 | Tests run twice on failure |
| 12 | `wire/handlers.go` L40 | Hardcoded `"templates/*.html"` path |
| 13 | `wire/app.go` L151 | Hardcoded `"./static"` path |
| 14 | `wire/app.go` L125 | CORS `"*"` hardcoded instead of configurable |
| 15 | `base.html` | `post-sidebar-cards` and `post-create-sidebar-cards` 80% identical |
| 16 | `Dockerfile` | Missing `.dockerignore` |
| 17 | `activity.html` | Negative booleans `HideX` should be positive `ShowX` |

### Low (Nice to Have)
| # | File | Issue |
|---|------|-------|
| 18 | `Dockerfile` L26 | `-a` flag unnecessary in Docker build |
| 19 | `Dockerfile` | Missing `HEALTHCHECK` instruction |
| 20 | `post_create.html` / `post_edit.html` | Nearly identical, could be merged |
| 21 | `settings.html` | Uses `comment` class for non-comment content |
| 22 | `settings.html` | No client-side password-match validation |
| 23 | `Makefile` L15 | `BINARY_UNIX` only used by one target |
| 24 | `home.html` / `board.html` | Post card HTML structurally duplicated |

---

### Overall Assessment

**Architecture: B+** — The hexagonal architecture with wire composition root is well-executed. The separation of concerns is clear and the DI pattern is consistent across all modules.

**KISS: B-** — The main offenders are `health.html` (hardcoded service iteration), `base.html` (layout logic and duplication), and the template sidebar card duplication. The Go code is generally simple and direct.

**Idiomatic Go: A-** — Minimal dependencies, proper error wrapping (mostly), interfaces used correctly. The `panic()` in `initServer` and setter injection are the notable exceptions.

**Build/Deploy: B** — Makefile is comprehensive but has DRY issues. Dockerfile is solid but needs `.dockerignore` and comment fix.
