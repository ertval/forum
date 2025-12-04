# User Module ID Exposure Audit

Date: 2025-11-17
Repo branch: `ekaramet/post-v5-schema`

Scope
- Manual audit of `internal/modules/user` package (domain, ports, application, adapters)
- Scan of templates and HTTP handlers that render user info or construct URLs
- Produce findings, security analysis, and suggested tests to verify no internal INT IDs are exposed in HTML/JSON/URLs

Reference rule (from `docs/SCHEMA_REFACTOR_STATUS.md`)
- Internal database IDs: `id INTEGER PRIMARY KEY AUTOINCREMENT` (internal, `int`) and must NOT be exposed to users or included in public URLs or API responses.
- Public IDs: `public_id TEXT UNIQUE NOT NULL` (UUID strings) are the only IDs exposed externally (in JSON and URLs).
- Handlers must accept public_id (string) in HTTP paths/queries and convert to internal `int` only for service/repository calls when necessary.
- Templates must display or emit only public UUIDs (or other safe identifiers), never the internal int ID.

Summary of findings
- Repositories: Good. `internal/modules/user/adapters/sqlite_repository.go` correctly generates `public_id` UUIDs on `Create`, scans `id` and `public_id`, and uses internal `id` for joins and counts. `GetUserStats` uses internal INTs for performance. (OK)

- Domain: `internal/modules/user/domain/user.go` defines `ID int` and `PublicID string`. The domain types currently lack JSON tags (this is acceptable but the higher layers must ensure only `PublicID` is rendered in JSON/templates). (Informational)

- Ports / Services:
  - `internal/modules/user/ports/service.go` and `application/service.go` expose service interfaces that take and return internal INT `userID` values (e.g., `GetByID(ctx, userID int)`). That is consistent with application-internal use, but there is no clear public-facing method that accepts `public_id string` for HTTP handlers to call directly. (Design note)

- HTTP Handlers (user module): `internal/modules/user/adapters/http_handler.go` is mostly placeholder (not implemented). No direct issues there.

- Templates and other handlers (critical issues found):
  - `templates/base.html` constructs links using `?user={{.User.ID}}` (line ~109). This places the internal integer ID into query params (exposes internal id). Example:
    - `/board?user={{.User.ID}}`
  - `templates/post_detail.html` references `{{.User.ID}}` and compares `{{if eq .User.ID .Post.UserID}}` (line ~41 and ~87). The template also uses `data-post-id="{{.Post.ID}}"` and link `/posts/{{.Post.ID}}/edit` which may or may not be public UUIDs depending on how the post object is prepared by handlers. The `User.ID` usage is definitely the internal int (from `buildCurrentUser` in `post` handler — see below).
  - `internal/modules/post/adapters/http_handler.go` (post handler) contains `buildCurrentUser(ctx, userID int)` which returns a map with the key `"ID": strconv.Itoa(userID)` — i.e., it exposes the internal integer ID to templates as `ID` (stringified). This function is directly used when rendering templates and is the root cause of `{{.User.ID}}` containing an internal integer in templates.

- Cross-module mismatch risks / code smell:
  - Some service/repository APIs use internal `int` IDs while HTTP filters and some filter structs (per `docs/SCHEMA_REFACTOR_STATUS.md`) expect public_id strings (e.g., `PostFilter.UserID` is a string). This produces friction and can introduce mistakes where internal `int` values are accidentally surfaced in templates or URLs.

Concrete locations (file / approximate line references)
- `templates/base.html` — uses `{{.User.ID}}` inside `/board?user={{.User.ID}}` (line 109 in current file). This exposes internal int.
- `templates/post_detail.html` — uses `{{.User.ID}}`, `{{.Post.ID}}`, and compares `eq .User.ID .Post.UserID` (line ~41, ~87). If `.Post.ID` is the public UUID, these comparisons could be mismatched types; if `.Post.ID` is internal int then it is inconsistent with design.
- `internal/modules/post/adapters/http_handler.go` — `buildCurrentUser(ctx context.Context, userID int)` returns `"ID": strconv.Itoa(userID)` (exposes int). This is used by multiple page renderers (board/home/post detail), making this a systemic exposure.
- `internal/modules/user/adapters/sqlite_repository.go` — (good) sets `user.PublicID = publicID.String()` on creation and scans both `id` and `public_id` from DB queries (e.g., `GetByID`, `List`).

Security analysis

1) Information Disclosure / Enumeration
- Exposing internal sequential integer IDs in HTML/URLs makes it trivial to enumerate users, posts, or other resources by iterating numeric IDs. Attackers can scrape `/board?user=1`, `/board?user=2`, etc., confirming activity or presence of accounts. Although many resources are public, the presence of internal IDs increases attack surface and allows correlating internal DB structure with public endpoints.

2) Insecure Direct Object References (IDOR)
- If any protected endpoints accept internal IDs and authorization checks are performed using values from the request (instead of derived from session), exposing internal IDs in URLs could allow attackers to directly call endpoints with other users' internal ids. Even if the current code compares `session.UserID` (internal) for authorization, incorrect type comparisons in templates/handlers (string vs int) may cause owner checks to be bypassed or fail unexpectedly.

3) Information Correlation & Privacy
- Internal IDs can be used to correlate records across different datasets (e.g., leaked DB snapshots or other systems). The design goal to expose only UUIDs helps prevent this correlation.

4) Logic type mismatches causing security logic bugs
- Templates comparing `.User.ID` (stringified `int`) to `.Post.UserID` (likely an `int`) may evaluate incorrectly depending on how template engine treats types. This can cause the template to hide or show owner actions incorrectly. More critically, if ownership checks are duplicated in both template (UI) and handler authorization logic, a mismatch may lead to UI showing unauthorized actions or handlers allowing unauthorized changes (if backend uses different types and comparisons).

5) CSRF/Injection surface
- Exposing internal IDs does not directly cause injection, but placing raw internal values into HTML attributes (e.g., `data-post-id`) means JavaScript code may rely on those numeric IDs when making API calls; if API endpoints expect public IDs, mismatch leads to errors; if they accept numeric IDs accidentally, it widens attack surface.

Recommendations (priority order)

Immediate (high priority)
- Stop exposing internal `int` IDs in templates and URLs.
  - Replace all template occurrences of `{{.User.ID}}` with `{{.User.PublicID}}` or ensure that the `User` view model property used by templates is a `PublicID` string (named consistently, e.g., `ID` may remain but should be public UUID, not internal int).
  - Update `post` handler's `buildCurrentUser` to expose `PublicID` instead of `strconv.Itoa(userID)`. Example: if you have the `user` object, set `"ID": user.PublicID` or `"PublicID": user.PublicID` and update templates accordingly.
- Make sure all links/URLs that reference users use the `public_id` (UUID) in query params and routes (e.g., `/board?user={public_uuid}`), and ensure the `FilterService` and downstream code accept public UUID strings and map them to internal `int` IDs in a single place (e.g., a repository helper `GetInternalIDFromPublicID`).

Short / medium term
- Add an explicit repository method `GetByPublicID(ctx, publicID string) (*domain.User, error)` (or `GetInternalIDFromPublicID(ctx, publicID string) (int, error)`) to centralize translation from public UUID → internal int. Ensure all handlers that receive public IDs use that method to obtain internal IDs for service calls.
- Update `ports` to clearly separate internal methods (taking `int`) from external lookup methods (taking `string publicID`) so intent is explicit.
- Update domain structs or create view-model structs for templates that guarantee the `ID` field is the public identifier. Avoid reusing domain structs directly in templates unless you can guarantee JSON tags and field naming reflect public/external view.

Longer term / hardening
- Add JSON struct tags in domain or dedicated DTO/view-models so that `json.Marshal` of user objects exposes `id` as `public_id` (uuid) and hides internal `id` (e.g., `ID int  json:"-"` and `PublicID string  json:"id"`). Follow the pattern in `docs/SCHEMA_REFACTOR_STATUS.md`.
- Review all handlers and template renderers across modules to ensure consistent naming: if templates expect `ID` to be the public id, ensure the code sets that consistently.

Suggested tests (to add to repo tests — *do not* add/change code now; below are test recipes and sample code to implement)

1) Static template scan (unit test) — assert templates do not contain patterns that expose INT IDs
- Purpose: fail early if a template contains `{{.User.ID}}` or emits numeric-only IDs in URLs
- Strategy: open each file in `templates/` directory and use regex checks

Sample test (Go):

```go
func TestTemplates_DoNotExposeInternalUserIDs(t *testing.T) {
    files, err := filepath.Glob("templates/**/*.html")
    if err != nil { t.Fatal(err) }

    numericIDPattern := regexp.MustCompile(`\?user=\d+|\{\{\.User\.ID\}\}`)
    for _, f := range files {
        b, err := os.ReadFile(f)
        require.NoError(t, err)
        if numericIDPattern.Match(b) {
            t.Errorf("template %s contains an internal user ID pattern", f)
        }
    }
}
```

2) HTML response inspection (integration test)
- Purpose: detect if server renders numeric internal IDs in HTML responses
- Strategy: start server (or use handler with httptest), request e.g., `/board` and `/posts/{public_id}`, search HTML for numeric IDs in `?user=` or `data-post-id="\d+"` or links to `/posts/\d+`.

Sample test (Go):

```go
func TestServerResponses_DoNotContainNumericInternalIDs(t *testing.T) {
    // start server or use existing mux
    resp := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/board", nil)
    mux.ServeHTTP(resp, req)

    body := resp.Body.String()
    if regexp.MustCompile(`\?user=\d+`).MatchString(body) {
        t.Fatal("found numeric internal user id in board HTML")
    }
}
```

3) JSON API contract tests
- Purpose: ensure JSON responses expose `id` as UUID (hyphenated) not numeric
- Strategy: call API endpoints that return users or posts and assert that `id` fields match UUID regex and not `^\d+$`.

Sample test snippet:

```go
uuidRe := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
// decode JSON from /api/posts or /api/users and verify id fields
if !uuidRe.MatchString(got.ID) {
    t.Fatalf("expected uuid id, got %s", got.ID)
}
```

4) Template equality / ownership checks
- Purpose: ensure owner-only UI elements are controlled by backend-authoritative checks, not just template comparisons that can be broken by types
- Strategy: Ensure server-side authorization in handler code uses `session.UserID` (internal int) compared against the authoritative `post.UserID` (internal int) before processing edit/delete operations. Tests should exercise unauthorized attempts with valid session cookies for another user and confirm 403/401 responses.

Implementation notes for tests
- Prefer integration tests using `httptest` with an in-memory DB or test DB seeded with the seed SQL. Tests that rely on exact HTML structure are brittle — prefer scanning for the risky patterns (numeric IDs) rather than asserting exact markup.
- Add helper functions:
  - `isUUID(string) bool`
  - `containsNumericIDInUserQuery(html string) bool`

Developer action items (ordered)
1. Change `post` handler `buildCurrentUser` so the map returned for templates contains `ID` as the user's `PublicID` (UUID) — or better: return `PublicID` explicitly and update templates to use `User.PublicID` or `User.ID` depending on chosen convention.
2. Replace `{{.User.ID}}` in templates with `{{.User.PublicID}}` or ensure the view-model supplies `ID` as a public UUID string — update all templates (`templates/base.html`, `templates/post_detail.html`, others found by scanning).
3. Add repository helper `GetByPublicID` or `GetInternalIDFromPublicID` to map a public UUID to the internal INT used by services.
4. Add tests described above and add a CI check for template scanning to prevent regressions.
5. Review all other modules (post, comment, reaction, moderation, notification) to ensure they do not expose internal ints. The `post` module already has similar issues in `buildCurrentUser`.

Appendix — exact matches found (grep output excerpts)
- `templates/base.html` — `?user={{.User.ID}}` (exposes internal id)
- `templates/post_detail.html` — `{{if eq .User.ID .Post.UserID}}` and `data-post-id="{{.Post.ID}}"` (may be internal ints depending on render)
- `internal/modules/post/adapters/http_handler.go` — `buildCurrentUser(...)` returns `"ID": strconv.Itoa(userID)` (exposes internal int)
- `internal/modules/user/adapters/sqlite_repository.go` — correct: sets `user.PublicID = publicID.String()` and scans `id, public_id` (OK)

Notes and caveats
- The repository layer appears correctly updated to the `INT` + `public_id` pattern. The primary remaining issue is presentation: templates and some handlers are exposing internal IDs. Fixing this requires coordinated changes to handlers and templates and consistent naming of view-models.
- There are some mismatches between `ports` and the HTTP-facing expectations (service methods accept `int` while HTTP input is `public_id` string). That is manageable if you centralize public->internal translation in handlers or add repository helpers.

If you'd like I can (next steps, optional):
- Create concrete PR patches to (1) change `buildCurrentUser` to return `PublicID`, (2) update the templates (`base.html`, `post_detail.html`) to use `User.PublicID`, and (3) add `GetByPublicID` in `user` repository and a small helper in the post handler to translate public user filters into internal IDs. I will not change code until you confirm.


