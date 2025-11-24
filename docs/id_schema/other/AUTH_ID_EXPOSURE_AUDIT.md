# Auth & User ID Exposure Security Audit

Date: 2025-11-17

Summary
-------
This audit inspects the `auth` module (handlers, middleware, service, repository) and related `post`/`user` usages to verify compliance with the project's SCHEMA_REFACTOR_STATUS rules: public-facing interfaces (APIs and HTML) MUST expose UUID public IDs, and internal integer IDs must remain internal (DB joins, FK, performance).

Executive finding: The codebase currently exposes internal integer IDs in multiple public surfaces (JSON responses, HTML templates, URL paths, and JavaScript data attributes). This is a security issue (user enumeration, information leakage) and violates the re-factor rules.

Files inspected
---------------
- `internal/modules/auth/adapters/http_handler.go`
- `internal/modules/auth/adapters/middleware.go`
- `internal/modules/auth/application/service.go`
- `internal/modules/auth/adapters/sqlite_session_repository.go`
- `internal/modules/user/adapters/sqlite_repository.go`
- `internal/modules/post/adapters/http_handler.go`
- `internal/modules/post/adapters/sqlite_repository.go`
- templates: `templates/home.html`, `templates/board.html`, `templates/base.html`, `templates/post_detail.html`, `templates/post_edit.html`, `templates/post_create.html`
- static JS: `static/js/post-detail.js`, `static/js/post-forms.js`

Detailed Findings
-----------------

1) Authentication handlers expose internal IDs

- `internal/modules/auth/adapters/http_handler.go`
  - `RegisterAPI` returns a response object with fields `ID` and `UserID` set to `strconv.Itoa(userID)`. `userID` is an internal integer (DB PK). This exposes the internal integer ID through the public API.
  - `LoginAPI` returns `UserID int` in JSON responses (directly `session.UserID`).
  - `GetSessionAPI` returns `UserID int` in JSON responses.

  Security impact: High — these responses leak sequential/incrementing integer IDs that enable user enumeration and information leakage.

2) Middleware places internal IDs into template-accessible context

- `internal/modules/auth/adapters/middleware.go`
  - `RequireAuth` and `OptionalAuth` put the internal `session.UserID` into the request context under key `UserIDKey` as `fmt.Sprintf("%d", session.UserID)` (stringified internal ID). That makes an internal numeric ID available to handlers and templates via `authAdapters.GetUserID()`.

  Security/functional impact:
  - Templates and handlers now often use `authAdapters.GetUserID()` output in HTML contexts or as filter params. Even if converted to string, that value is an internal ID.
  - Using the internal ID in template contexts ends up in HTML links and data attributes (see below).

3) Post handlers & templates use internal IDs publicly

- `internal/modules/post/adapters/http_handler.go`
  - `buildCurrentUser` exposes `"ID": strconv.Itoa(userID)` in the `User` map passed to templates.
  - In `HomePage`, `BoardPage`, `LoadMorePostsAPI` and similar functions, the code explicitly sets `previewPost["ID"] = post.ID` and `previewPost["UserID"] = post.UserID` (both internal ints) when creating the data passed to templates or APIs.

- Templates rely on `.ID` for links and attributes:
  - `templates/home.html`, `templates/board.html`, `templates/base.html` use `<a href="/posts/{{.ID}}">` which will produce `/posts/<internal-int>`.
  - `templates/post_detail.html` uses `data-post-id="{{.Post.ID}}"`, `href="/posts/{{.Post.ID}}/edit"`, and comment `id` attributes like `id="comment-{{.ID}}"`.

  Security/functional impact:
  - Internal numeric IDs are exposed in URL paths and JS data attributes. Attackers can iterate integer IDs to discover posts and users (enumeration).
  - JavaScript code (`static/js/post-detail.js`, `static/js/post-forms.js`) reads `data-post-id` and will send that to APIs; if it sends internal ints publicly, that continues exposure via AJAX.

4) Repository layer is correct; adapters mostly generate UUIDs

- `internal/modules/post/adapters/sqlite_repository.go` and `internal/modules/user/adapters/sqlite_repository.go` correctly generate `public_id` UUIDs and store internal `id` ints as DB primary keys. Queries that are externally addressing entities use `public_id` where appropriate (for example, posts repository `GetByID` uses `WHERE p.public_id = ?`).

  This means the domain layer has `PublicID` fields available, but handlers/templates are not consistently using them.

5) Comment module (partial) and other parts

- `templates/post_detail.html` references comment IDs as `{{.ID}}` and `data-comment-id` — the comment domain also contains `PublicID` but the templates expect `.ID`, introducing similar risk for comments when implemented.

Root cause
----------
- Domain and repository layers were migrated to include `PublicID` (UUID) fields and repositories generate them correctly.
- However many adapters/handlers/templates continue to treat `ID` as the public identifier (often copying internal `ID` into variables labelled `ID` that are passed to templates or JSON responses). The mapping from internal ID → public ID hasn't been applied consistently at the adapter level.

Security Analysis & Risk
-----------------------

- Exposing internal integer IDs permits:
  - User enumeration (high likelihood, high impact).
  - Predictable URL discovery and data scraping.
  - Leaking information about total records and creation order.

- Repositories are using INT for joins and UUIDs for public references — that is good. The main security risk is the mismatch between repository design and adapter usage.

Severity: Critical for auth APIs and high for templates/URLs.

Recommendations (Concrete)
------------------------

Immediate fixes (highest priority)

1. Auth handlers must not return internal IDs
   - `RegisterAPI` should return the user's `PublicID` (UUID) instead of `strconv.Itoa(userID)`.
   - `LoginAPI` and `GetSessionAPI` should return `user.PublicID` or omit the user ID entirely. If the client needs an identifier, return a stable `public_id` UUID only.
   - Suggested change: Update `AuthService.Register` signature to return `(userPublicID string, session *domain.Session, err error)` OR have `Register` continue returning `userID int` but make the handler call `userService.GetByID(ctx, userID)` to fetch `user.PublicID` before producing JSON.

2. Middleware should avoid placing internal IDs into template-visible context keys
   - Keep an internal-only context key for service calls (e.g., `contextKeyInternalUserID`) and store the `int` there.
   - If templates/controller logic require a user identifier for links, expose a `UserPublicID` or `Username` separately — but do not expose the internal integer.

3. Post & template changes
   - In all handlers that construct maps for templates (e.g., `buildCurrentUser`, `previewPost`), use `post.PublicID` and `user.PublicID` **for any field named `ID` that will be used in URLs or JS data attributes**.
   - Example: `previewPost["ID"] = post.PublicID` and `previewPost["UserID"] = post.UserPublicID`.
   - Update templates to use `{{.ID}}` expecting it to be a UUID OR migrate templates to be explicit: `{{.PublicID}}`.

4. JS changes
   - Ensure client-side JS expects UUIDs in `data-post-id` and sends them in AJAX payloads.

5. Filters and query params
   - Use public UUIDs for `user`/`liked_by`/`filter.UserID` query parameters. Ensure `FilterService` and repository filters join on `users.public_id` when using those external values (the code already uses `u.public_id = ?` in many places, adjust callers to pass UUIDs, not stringified ints).

Suggested code snippets (non-invasive)
-------------------------------------

1) Auth Register handler (pseudo-fix)

```go
// After register returns internal ID
userID, session, err := h.authService.Register(ctx, email, username, password)
if err != nil { /* handle */ }

// Fetch user to get public id
user, err := h.userService.GetByID(ctx, userID)
if err != nil { /* handle */ }

resp := struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Token    string `json:"token"`
}{
    ID:       user.PublicID,
    Email:    user.Email,
    Username: user.Username,
    Token:    session.Token,
}
```

2) Post preview mapping (pseudo-fix)

```go
previewPost["ID"] = post.PublicID
previewPost["UserID"] = post.UserPublicID
```

Tests (recommended) — add these to the test suite
-------------------------------------------------

Below are test ideas and example snippets you can add to `internal/.../` tests. These tests are written as suggestions (do not modify production code here):

Unit test: ensure Auth API responses contain UUID, not integer

```go
func TestRegisterAPI_ExposesPublicID_NotInternalInt(t *testing.T) {
    // Arrange: start handler with in-memory DB / mocks
    // Act: call RegisterAPI
    // Assert: response 'id' matches UUID regex and is not a decimal integer
}
```

Integration test: rendered templates contain UUIDs in post links

```go
func TestHomePage_PostLinksUseUUIDs(t *testing.T) {
    res := httpGet(t, serverURL+"/")
    body := readBody(t, res)
    // Ensure no hrefs with /posts/<decimal-int>
    assert.NotRegexp(t, `href="/posts/\d+"`, body)
    // Ensure href with /posts/<uuid>
    assert.Regexp(t, `href="/posts/[a-fA-F0-9\-]{36}"`, body)
}
```

Checklist for implementation
----------------------------
 - [ ] Stop returning internal ints in auth JSON responses
 - [ ] Ensure middleware does not expose internal int IDs to template contexts
 - [ ] Use `PublicID` (UUID) when composing URLs, `data-*` attributes and JSON responses
 - [ ] Update client-side JS to assume UUIDs in `data-*` attributes
 - [ ] Add tests described above and run `go test ./...`

Appendix: concrete code locations to change
-----------------------------------------
- `internal/modules/auth/adapters/http_handler.go` — RegisterAPI, LoginAPI, GetSessionAPI
- `internal/modules/auth/adapters/middleware.go` — context keys and Set/Expose behavior
- `internal/modules/post/adapters/http_handler.go` — previewPost creation, buildCurrentUser
- `templates/*.html` — update uses of `{{.ID}}`/`{{.Post.ID}}` → ensure they refer to UUIDs
- `static/js/*.js` — read/send UUIDs from `data-post-id` and `data-comment-id`

If you want, I can also:
- Produce exact patch diffs for minimal changes (handlers + template variable mapping)
- Create the suggested tests as runnable files in `tests/integration/` and `internal/.../` (I will not modify business logic; tests only)

End of report.
