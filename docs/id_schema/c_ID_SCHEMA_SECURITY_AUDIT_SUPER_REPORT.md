# ID Schema Security Audit — Super Report

Date: 2025-11-17
Repo branch: `ekaramet/post-v5-schema`

## Executive Summary

- Scope: consolidated findings from `docs/id_schema/*` and `docs/id_schema/id_audit/*` audits (auth, post, user, comment, reaction, moderation, notification, platform, templates, tests).
- Root issue: database and repository layers were refactored to use INT primary keys + UUID `public_id`, but many adapters/handlers/templates/JS still expose internal integer `id` values to clients (URLs, JSON, DOM attributes). This causes ID enumeration, information disclosure, and increased IDOR risk.
- Severity: Critical for auth and user/profile exposure; High for post/comment templates & JS; High/Medium for modules with missing repo implementations (comment, reaction, moderation, notification).

## High-level Findings (deduplicated)

- Repositories and migrations: Generally compliant — tables contain `id INTEGER PRIMARY KEY AUTOINCREMENT` and `public_id TEXT UNIQUE NOT NULL` and many repositories generate `public_id`.
- Domain structs: most modules include both `ID int` and `PublicID string` but JSON tags and view-model usage are inconsistent.
- Adapters / Handlers: many HTTP handlers return or pass internal `ID` values to templates and JSON responses instead of `PublicID` (auth handlers and post handlers are prominent offenders).
- Templates & JS: templates contain `{{.ID}}`, `{{.Post.ID}}`, `{{.User.ID}}`, `data-post-id` attributes set from internal IDs; JS reads these attributes and calls APIs with internal ints.
- Ports / Services: service and repository interfaces often accept/return `int` where external callers should use `string` public_id (user, comment, reaction, moderation, notification modules).
- Tests: several tests and test schemas use or assume internal ints (or lack `public_id`), so they don't detect regressions and some are broken.

## Concrete Per-Module Findings & Immediate Fixes

- **Auth**
  - Finding: `RegisterAPI`, `LoginAPI`, `GetSessionAPI` return internal user ID (often stringified int).
  - Risk: High — enables user enumeration via API.
  - Fix: Return `user.PublicID` (UUID). Either change service signature to return publicID or have handler fetch user by internal ID and emit `PublicID`.

- **Post**
  - Finding: Handlers populate template maps with `post.ID` / `post.UserID`; templates use `{{.ID}}` and `{{.Post.ID}}` in URLs and `data-*` attributes.
  - Risk: High — content enumeration, predictable URLs.
  - Fix: Handlers must pass `post.PublicID` and `post.UserPublicID` to templates and APIs; update templates to use `{{.PublicID}}` or `{{.Post.PublicID}}` consistently.

- **User**
  - Finding: Templates use `{{.User.ID}}` and service ports lack `GetByPublicID(string)` for external lookups.
  - Risk: Critical — profile enumeration and privacy leakage.
  - Fix: Add / implement `GetByPublicID(string)` in repository/service; update handlers and templates to accept and use public UUIDs for URLs and query params.

- **Comment**
  - Finding: Domain has `PublicID` but repo methods and ports still use `int` and repo `Create` may not persist `public_id` yet.
  - Risk: High when handlers are implemented.
  - Fix: Repository `Create` must generate and persist `public_id`; ports/services/handlers should accept `public_id` strings for external access; templates/JS must use comment `PublicID`.

- **Reaction**
  - Finding: Reaction entity has `PublicID`, but reaction APIs use internal target ids; repository TODOs exist.
  - Risk: High — target enumeration and mismatched lookups.
  - Fix: Accept target `public_id` in HTTP layer and resolve to internal ID server-side before repository queries; implement repository UUID handling.

- **Moderation**
  - Finding: `Report` domain lacks `PublicID` and repo/handlers use ints.
  - Risk: High — report enumeration and disclosure of sensitive data.
  - Fix: Add `PublicID` to `Report`, persist `public_id` in repo, and update interfaces to use `public_id` for external operations.

- **Notification**
  - Finding: Domain lacks `PublicID`; handlers/repo are TODOs.
  - Risk: High — notification enumeration and privacy issues.
  - Fix: Add `PublicID`, persist `public_id`, update handlers to use public IDs for client-side operations (mark-as-read, listing).

- **Platform / Templates / JS**
  - Finding: Templates contain many `{{.ID}}` patterns and JS expects numeric IDs from `data-*` attributes.
  - Risk: High — broad exposure across UI and client behavior.
  - Fix: Add view-model mapping so `ID` in templates is a public UUID (or rename to `PublicID` and update templates). Update JS to use UUIDs.

## Security Impact / Threats

- ID Enumeration: exposing sequential internal integers allows easy discovery of resources (posts, users, comments).
- Insecure Direct Object Reference (IDOR): internal IDs used in endpoints can be abused to access or modify other users' resources.
- Information leakage: sequential IDs reveal system scale, registration order, and allow correlation across logs/services.

## Priority Remediation Plan (ordered)

1. Immediate (Critical) — do before public deployment
   - Update auth handlers to return `PublicID` instead of internal IDs.
   - Update post handlers to populate template data with `PublicID` values.
   - Replace `{{.ID}}` occurrences in templates used in public contexts with `{{.PublicID}}` (or ensure view-model `ID` is public UUID).

2. Short-term (days)
   - Add/implement `GetByPublicID(string)` in repositories and services for `user`, `comment`, `moderation`, `notification` modules.
   - Ensure repository `Create()` methods generate and persist `public_id` where missing.
   - Update JS to accept and use UUIDs from `data-*` attributes.

3. Medium-term (1–2 weeks)
   - Update ports/service signatures for external-facing methods to use `string public_id` where appropriate (keep internal int signatures for internal calls if needed).
   - Add middleware or central helper that validates public_id format and converts to internal ID for service calls.
   - Add authorization checks that operate on internal IDs (server-side) after conversion.

4. Longer-term
   - Add CI checks: template-scan linter and tests to prevent regressions.
   - Add audit logging for public→internal ID conversions and rate limiting for endpoints susceptible to enumeration.

## Concrete Code Patterns / Examples

- Domain JSON tags (recommended):

```go
type Post struct {
  ID       int    `json:"-"`
  PublicID string `json:"id"`
  // other fields
}
```

- Auth handler pattern (send public id):

```go
userID, session, err := h.authService.Register(ctx, ...)
user, _ := h.userService.GetByID(ctx, userID)
resp := struct { ID string `json:"id"` /* ... */ }{ ID: user.PublicID }
```

- Template mapping pattern (server-side):

```go
preview := map[string]interface{}{
  "ID": post.PublicID, // NOT post.ID
  "Title": post.Title,
}
```

## Tests to Add (examples)

- Template scan unit test: fail if templates contain `{{.User.ID}}`, `{{.Post.ID}}`, `data-post-id="{{.Post.ID}}"`, `href="/posts/{{.ID}}"`, or similar patterns.
- API contract tests: assert `id` fields in JSON responses are UUIDs (regex) and not numeric.
- Enumeration tests: hitting `/posts/1` or `/users/1` should return 404 or invalid format.
- Repo tests: `Create()` populates `PublicID` and `GetByPublicID()` returns the created entity.

## Files Likely Needing Edits (high-impact list)

- `internal/modules/auth/adapters/http_handler.go`
- `internal/modules/auth/adapters/middleware.go`
- `internal/modules/post/adapters/http_handler.go`
- `internal/modules/post/*` templates: `templates/base.html`, `templates/board.html`, `templates/home.html`, `templates/post_detail.html`, `templates/post_edit.html`, `templates/post_create.html`
- `internal/modules/user/ports/*` and `internal/modules/user/adapters/sqlite_repository.go`
- `internal/modules/comment/*` (domain JSON tags, repository Create, ports, handlers)
- `internal/modules/reaction/*` (resolve public target IDs)
- `internal/modules/moderation/*` (add `PublicID` to `report.go`, repo/service/handlers)
- `internal/modules/notification/*` (add `PublicID` and repo changes)
- `static/js/*` — update to use UUIDs
- `tests/*` — add template-scan and API contract tests

## Implementation Guidance & Notes

- Do minimal, coordinated changes: either update templates and handlers together to use `PublicID`, or standardize that the view-model `ID` field is always a public UUID so templates need minimal edits.
- Keep internal INT IDs for DB joins and performance; convert public UUID → internal int at handler layer (centralize this logic).
- Add `GetByPublicID` helpers in repositories to keep conversion logic consistent and testable.
- Add the template-scan test to CI to block regressions.

## Suggested Next Steps (pick one)

1. Create a small patch replacing template `{{.XXX.ID}}` occurrences with `{{.XXX.PublicID}}` and update the few handlers that populate those fields (fast, high impact).
2. Add the detector tests (template-scan and API contract) so future PRs are blocked if they reintroduce leaks.
3. Implement `GetByPublicID` methods for `user/comment/moderation/notification` and update service interfaces.

If you want, I can start with Option 1 (mechanical templates + handler mapping updates) and create a PR-ready patch. Tell me which option to implement and I will apply changes incrementally and run tests where possible.

---
Report generated by consolidating `docs/id_schema` audit files on 2025-11-17.
