# Aggregate Codebase Review — Verified Open Issues

**Date:** 2026-02-28  
**Method:** Five parallel subagents verified each prior review report against the live codebase. Only issues confirmed as **STILL EXISTS** are listed here.  
**Principle:** All fixes should follow idiomatic Go and the KISS principle — simple, compact, optimal.

---

## Summary

| Module | Open Issues |
|--------|-------------|
| `internal/modules/post` | 6 |
| `internal/modules/reaction` | 2 |
| `internal/platform` | 2 |
| `static/css` | 2 |
| `static/js` | 3 |
| `templates` | 7 |
| **Total** | **22** |

---

## 1. Post Module (`internal/modules/post`)

### 1.1 N+1 Query in Post Listing — HIGH

**File:** `internal/modules/post/adapters/sqlite_repository.go` line ~456  
**Problem:** `r.getPostCategories(ctx, post.ID)` is called inside `for rows.Next()` — one extra query per post. On a page of 20 posts this is 21 queries instead of 2.

**Fix:** Use a single `GROUP_CONCAT` in the main query or a batch `SELECT … WHERE post_id IN (…)` lookup after collecting all IDs.

```go
// Current (bad)
for rows.Next() {
    // ...
    post.Categories, err = r.getPostCategories(ctx, post.ID) // N+1
}

// Fix: collect IDs first, then one batch query
```

---

### 1.2 Ad-hoc Logger Created Per Request — MEDIUM

**File:** `internal/modules/post/adapters/http_handler_api.go` lines ~119, ~261  
**Problem:** `logger.NewWithConfig(...)` called inside `CreatePostAPI` (on error path) and at the top of every `UpdatePostAPI` call. Logger creation is expensive; it should be injected once at construction.

**Fix:** Add `logger *logger.Logger` field to `HTTPHandler`, inject via `NewHTTPHandler`, replace ad-hoc calls with `h.logger`.

---

### 1.3 Duplicate Multipart/JSON Parsing Logic — MEDIUM

**File:** `internal/modules/post/adapters/http_handler_api.go`  
**Problem:** Both `CreatePostAPI` and `UpdatePostAPI` contain their own full `switch` block for multipart/JSON/form-encoded content parsing with no shared helper.

**Fix:** Extract `parsePostRequest(r *http.Request) (*PostInput, error)` helper called by both handlers.

---

### 1.4 `HomePage` and `BoardPage` Duplicate Logic — MEDIUM

**File:** `internal/modules/post/adapters/http_handler_page.go` line ~163  
**Problem:** `BoardPage` carries the inline comment *"identical to homepage"* and fully duplicates all session, filter, pagination, and template rendering logic from `HomePage`.

**Fix:** Extract shared `renderPostListPage(w, r, filters)` helper. Both handlers call it with different default filter values.

---

### 1.5 Title Length Validation/Error Mismatch — LOW

**File:** `internal/modules/post/domain/post.go` line ~32 vs `internal/modules/post/domain/errors.go` line ~30  
**Problem:** `Validate()` enforces `len(p.Title) > 255` but `ErrTitleTooLong` error message says *"max 300 characters"*.

**Fix:** Introduce a single constant and use it in both places:
```go
const MaxTitleLength = 255

// In Validate:
if len(p.Title) > MaxTitleLength { return ErrTitleTooLong }

// In errors.go:
ErrTitleTooLong = errors.New(fmt.Sprintf("post title too long (max %d characters)", MaxTitleLength))
```

---

### 1.6 Duplicate `Author` / `AuthorUsername` Fields — LOW

**File:** `internal/modules/post/domain/post.go` lines ~12–13  
**Problem:** `Post` struct has both `AuthorUsername string` and `Author string` carrying the same data. Repository sets both.

**Fix:** Remove `Author`, use only `AuthorUsername` (or rename once to `Author`). Update repository and all template references.

---

## 2. Reaction Module (`internal/modules/reaction`)

### 2.1 TOCTOU Race Condition in Reaction Toggle — MEDIUM

**File:** `internal/modules/reaction/application/service.go` lines ~83–122  
**Problem:** `React()` performs read (`GetByUserAndTargetPublicID`) → conditional delete → create as three separate, non-transactional operations. A concurrent request with the same user/target can produce duplicate reactions or wrong counts.

**Fix:** Wrap the entire check-delete-create sequence in a database transaction, or use an `INSERT … ON CONFLICT` upsert pattern at the SQL level.

---

### 2.2 Stale Mock Implementations — HIGH

**File:** `internal/modules/reaction/ports/service_test.go` lines ~27–39  
**Problem:** Mock signatures use `targetID int` but the current `ReactionService` interface uses `targetPublicID string`. Mocks also lack `GetUserReactionCount` and `GetByUserAndTargetPublicID` methods. Tests compile only because the mock implements an outdated interface snapshot.

**Fix:** Regenerate or hand-update mocks to match the current interface; add missing method stubs; verify tests assert actual behaviour.

---

## 3. Platform (`internal/platform`)

### 3.1 Complex Environment Validation — LOW

**File:** `internal/platform/config/config.go` line ~189  
**Problem:** Three-way `&&` chain is harder to extend and read than a slice membership check.
```go
// Current
if c.Server.Environment != "development" && c.Server.Environment != "staging" && c.Server.Environment != "production" {
```

**Fix:**
```go
// idiomatic Go 1.21+
if !slices.Contains([]string{"development", "staging", "production"}, c.Server.Environment) {
    return fmt.Errorf("invalid environment %q (valid: development, staging, production)", c.Server.Environment)
}
```

---

### 3.2 Unused `getRequiredTemplates` Function — LOW

**File:** `internal/platform/templates/validator.go` lines ~115–122  
**Problem:** Unexported `getRequiredTemplates() []string` has zero call sites in production code. Dead code increases maintenance cost.

**Fix:** Remove the function. If it becomes needed, re-add it (or export it) at that point.

---

## 4. Static CSS (`static/css`)

### 4.1 Duplicate Button Styles in `cards.css` — MEDIUM

**File:** `static/css/cards.css` lines ~193–215  
**Problem:** `.filters button` re-declares a complete button style (border-radius, padding, background gradient, etc.) that duplicates `buttons.css`. Any button style update must be made in two places.

**Fix:** Apply `.btn` (and optionally `.btn-primary`) classes directly to filter buttons in the HTML, then only override layout-specific properties in `cards.css`:
```css
.filters .btn {
    width: 100%;
}
```

---

### 4.2 CSS `@import` Serial Loading — LOW

**File:** `static/css/style.css`  
**Problem:** 14 `@import url(...)` statements load stylesheets serially, creating a cascade of render-blocking requests on first load.

**Fix:** Replace with direct `<link rel="stylesheet">` tags in `templates/base.html`. The browser can then fetch all sheets in parallel. Alternatively, concatenate files at build time.

---

## 5. Static JavaScript (`static/js`)

### 5.1 Large HTML Fragments Embedded in JS — HIGH

**Files:** `static/js/load-more-posts.js` line ~62, `static/js/load-more-comments.js` line ~37  
**Problem:** `createPostElement()` and `createCommentElement()` contain 35–50 line `innerHTML` template literals. These are invisible to IDE HTML tooling, get out of sync with Go-rendered HTML, and are impossible to test in isolation.

**Fix:** Add `<template id="post-card-template">` (and comment variant) to the relevant Go templates. JS clones and fills the template:
```js
const tpl = document.getElementById("post-card-template");
const el = tpl.content.cloneNode(true);
el.querySelector(".post-title a").textContent = post.Title;
// ...
container.appendChild(el);
```

---

### 5.2 `location.reload()` After User Actions — MEDIUM

**File:** `static/js/post-detail.js` lines ~104, ~250  
**Problem:** Page is fully reloaded after posting a comment and after editing one. This resets scroll position, wastes bandwidth, and feels like legacy web behaviour.

**Fix:** After a successful POST/PUT, inject or update the comment DOM element directly instead of reloading. Reaction counts on the same page already return JSON suited for this pattern.

---

### 5.3 Duplicated Fetch / Error-Handling Boilerplate — MEDIUM

**Files:** `static/js/auth.js`, `static/js/post-forms.js`, `static/js/post-detail.js`  
**Problem:** Each file independently implements try/catch, JSON parsing, and inline `formErrors.innerHTML` error injection. No shared utility exists.

**Fix:** Add a small API client to `static/js/main.js` (loaded first via `base.html`):
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
Replace per-file boilerplate with `api.request(...)`.

---

## 6. Templates (`templates/`)

### 6.1 Duplicate Post Card Markup — MEDIUM

**Files:** `templates/home.html`, `templates/board.html`  
**Problem:** Both files contain fully inline `<article>` post card markup (`post-card-compact` and `post-card` variants). `base.html` defines `{{define "post-card"}}` but neither page uses it. A UI change requires editing two files.

**Fix:** Define `{{define "post-card-compact"}}` in `base.html` and replace the inline markup in `home.html` with `{{template "post-card-compact" .}}`.

---

### 6.2 Duplicate Load-More Button Markup — LOW

**Files:** `templates/home.html` lines ~70–82, `templates/board.html` lines ~55–68, `templates/comments.html`  
**Problem:** Each page defines its own load-more `<button>` with different `id` and `data-*` attributes but near-identical structure.

**Fix:** Define `{{define "load-more-button"}}` in `base.html` and call `{{template "load-more-button" .LoadMoreParams}}` from each page.

---

### 6.3 Complex Layout Logic in `base.html` — MEDIUM

**File:** `templates/base.html` lines ~13, 85–160  
**Problem:** Layout is determined by six `if/else if` template branches driven by title-string comparisons like `(eq .Title "Home")`. Adding a new page requires modifying template logic instead of just setting a field in the handler.

**Fix:** Set a `.Layout` field in the Go handler's template data (`"single"`, `"right"`, `"three-col"`, etc.). Template becomes:
```gohtml
<div class="page-layout-{{.Layout}}">
```
This makes layout decisions testable in Go unit tests.

---

### 6.4 Inconsistent Error Container IDs — LOW

**Files:** Multiple templates  
**Problem:** Page-level templates use `id="page-errors"` while form templates use `id="form-errors"` with no documented convention. JavaScript must know which ID to target per page.

**Fix:** Document and enforce a single convention: `id="page-errors"` for top-of-page notifications, `id="form-errors"` inside `<form>` elements. Extract as a reusable `{{define "error-container"}}` partial.

---

### 6.5 Repetitive Health Status Table Logic — MEDIUM

**File:** `templates/health.html` lines ~42–96  
**Problem:** Seven separate `{{if eq $service "auth_api"}}` blocks hard-code display names inside the template. Adding a new module requires a template edit.

**Fix:** Change the Go health handler to pass a `[]HealthItem{Key, DisplayName, Status}` slice. Template becomes a simple `{{range .ModuleHealth}}` loop.

---

### 6.6 Hardcoded Copyright Year — LOW

**File:** `templates/base.html` line ~152  
**Problem:** `&copy; 2025` will become stale each year.

**Fix:** Pass `.CurrentYear` from a base template data helper:
```gohtml
<p>&copy; {{.CurrentYear}} Ertval Karameta & Magnus Edvall.</p>
```

---

### 6.7 Sidebar Cards Markup Duplication — LOW

**File:** `templates/base.html` lines ~336–424  
**Problem:** `post-sidebar-cards` and `post-create-sidebar-cards` share ~80% identical markup (image upload widget, category selection block) with no extracted sub-templates.

**Fix:** Extract `{{define "sidebar-image-upload"}}` and `{{define "sidebar-category-select"}}` partials, include them from both sidebar templates.

---

## Prioritised Action Plan

### P0 — Correctness / Test Integrity
| # | Action | File |
|---|--------|------|
| R2.2 | Fix stale reaction mock signatures | `reaction/ports/service_test.go` |

### P1 — Performance
| # | Action | File |
|---|--------|------|
| R1.1 | Eliminate N+1 query in post listing | `post/adapters/sqlite_repository.go` |
| JS5.1 | Move inline HTML to `<template>` tags | `load-more-posts.js`, `load-more-comments.js`, Go templates |

### P2 — Maintainability / DRY
| # | Action | File |
|---|--------|------|
| R1.2 | Inject logger into post handler | `post/adapters/http_handler_api.go` |
| R1.3 | Extract `parsePostRequest` helper | `post/adapters/http_handler_api.go` |
| R1.4 | Extract `renderPostListPage` helper | `post/adapters/http_handler_page.go` |
| R2.1 | Wrap reaction toggle in transaction | `reaction/application/service.go` |
| T6.3 | Replace layout conditionals with `.Layout` field | `base.html` + all page handlers |
| T6.5 | Replace health table `if` chain with structured data | `health.html` + health handler |
| JS5.3 | Add shared `api.request` utility | `static/js/main.js` |
| JS5.2 | Replace `location.reload()` with DOM updates | `post-detail.js` |

### P3 — Minor / Cleanup
| # | Action | File |
|---|--------|------|
| R1.5 | Introduce `MaxTitleLength` constant | `post/domain/post.go` + `errors.go` |
| R1.6 | Remove duplicate `Author`/`AuthorUsername` | `post/domain/post.go` |
| P3.1 | Use `slices.Contains` for env validation | `platform/config/config.go` |
| P3.2 | Delete unused `getRequiredTemplates` | `platform/templates/validator.go` |
| CSS4.1 | Remove duplicate button styles from `cards.css` | `static/css/cards.css` |
| CSS4.2 | Replace `@import` with `<link>` tags | `static/css/style.css` + `base.html` |
| T6.1 | Extract `post-card-compact` template | `base.html`, `home.html` |
| T6.2 | Extract `load-more-button` template | `base.html`, page templates |
| T6.4 | Standardise error container IDs | all templates |
| T6.6 | Dynamic copyright year | `base.html` + base handler |
| T6.7 | Extract sidebar sub-templates | `base.html` |
