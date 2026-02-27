post

# post module — ID exposure audit

Summary
- Goal: verify that all external-facing endpoints, templates and JS use the public UUID (`PublicID` / `public_id`) and never leak or render internal integer DB IDs (`id` INT).
- Scope: `internal/modules/post` handlers and templates; then `internal/modules/user` package (repo & adapters) because handlers often need the user's public id.
- Action: manual code scan (templates + handlers + static js) and recommendations. No code changes were made — this file contains findings, security analysis and test snippets you can run.

Findings (high-level)
- The DB schema, domain entities and repositories already follow the INT internal ID + UUID public_id pattern (migrations and repositories generate and persist `public_id` and use integer `id` internally).
- Multiple server-side templates and handlers currently render or reference `.ID` or `.Post.ID` in URLs and in element data attributes, which are the internal integer IDs and therefore leak internal IDs in HTML and JS.
- Handlers and helpers sometimes convert internal ints to strings (e.g. `strconv.Itoa(userID)`) and pass those into template data or filter params used in URLs — again exposing internal IDs in URLs and as `user` query params.
- The `user` repository/adapters persist and return `PublicID` correctly, but handlers do not always fetch or populate `PublicID` for templates; they instead pass internal `UserID` values into template data.

Concrete occurrences (examples)
- Templates that render internal IDs (path + snippet):
  - `templates/post_edit.html`
    - `data-post-id="{{.Post.ID}}"` (data attribute) and `href="/posts/{{.Post.ID}}"` — line examples found during scan.
  - `templates/post_detail.html`
    - `data-post-id="{{.Post.ID}}"` (like/dislike buttons), `{{if eq .User.ID .Post.UserID}}` ownership checks in template, `href="/posts/{{.Post.ID}}/edit"`, `button ... data-post-id="{{.Post.ID}}"`, comment form: `data-post-id="{{.Post.ID}}"`.
    - Comment items: `id="comment-{{.ID}}"`, `data-comment-id="{{.ID}}"` — note that comment `ID` is an internal integer too.
  - `templates/board.html` and `templates/home.html`
    - Post links: `<a href="/posts/{{.ID}}">` inside post card templates. These use `.ID` not `.PublicID`.
  - `templates/base.html`
    - `href="/board?user={{.User.ID}}` in the user card "My Posts" link — this passes an internal ID as a query param.
  - `templates/post_edit.html` and `templates/post_create.html` use `data-post-id` attributes with `.Post.ID`.
- Static JS that expects `id` property in API responses (ok if API returns UUID), but also client-side code that builds links using server-side template variables: `static/js/post-forms.js` and `static/js/load-more-posts.js` use `result.id` and `${post.ID}` respectively. Server-side `${post.ID}` inserted into template HTML will be the integer ID if templates use `.ID`.
- Handler helper `buildCurrentUser(ctx, userID int)` in `internal/modules/post/adapters/http_handler.go` returns template data where `"ID": strconv.Itoa(userID)` therefore exposing the internal user ID in templates (string conversion of internal int). The templates and JS will then include internal IDs in URLs and data attributes.

Why this is a security/privacy problem
- Internal integer IDs are predictable and enumerable. Exposing them in URLs, HTML attributes or API responses can enable:
  - ID enumeration and probing (scanning sequential IDs to find resources and attempt unauthorized access)
  - Profile harvesting and easier correlation across services or logs
  - Making it easier for attackers to craft targeted requests or to correlate internal DB records
- It may also break the intention behind the public UUID abstraction (which hides internal DB internals and mitigates enumeration and linkability).

Severity & exploitability
- Severity: High for information exposure and potential enumeration (exposing internal IDs in URLs/HTML makes it trivial to discover sequential resources).
- Exploitability: Low to medium complexity — an attacker only needs to view HTML/JS to find IDs or derive them programmatically. If access controls are correct on server side then leaking an internal ID may not directly grant access, but it greatly facilitates automated attacks and increases risk surface.

Detailed recommendations (fixes, ordered by priority)
1. Templates: Replace all `{{.ID}}` / `{{.Post.ID}}` / `data-post-id="{{.Post.ID}}"` and any use of `.User.ID` in public HTML with the public identifier field (e.g. `{{.PublicID}}`, `{{.Post.PublicID}}`, `{{.User.PublicID}}`).
   - Example changes:
     - `href="/posts/{{.ID}}"` → `href="/posts/{{.PublicID}}"`
     - `data-post-id="{{.Post.ID}}"` → `data-post-id="{{.Post.PublicID}}"`
     - `id="comment-{{.ID}}"` → `id="comment-{{.PublicID}}"` (update comment JS accordingly).
2. Handler helper(s): Update `buildCurrentUser` and any handler data maps to return `PublicID` instead of `strconv.Itoa(userID)`.
   - Prefer returning both `PublicID` and `Username` in the template data: `"PublicID": user.PublicID`, and keep internal `userID` only in server-side logic.
   - If the handler only has `session.UserID` (internal int), it must call `userService.GetByID(ctx, session.UserID)` to fetch the `PublicID` before passing values to templates.
3. Filter params and query strings: `FilterParams.UserID` is a string (public id) and must not be populated with `strconv.Itoa(userID)` (stringified internal ID). Ensure the code sets `params.CurrentUserID` or `params.UserID` to `user.PublicID` when building filters for `My Posts` or `LikedPosts`.
4. API responses: Ensure JSON responses expose `id` (public UUID) and `user_id` (public UUID of author), not internal numeric IDs. Domain structs already have `PublicID` tagging `json:"id"` — confirm the code sets `Post.PublicID` before JSON encoding (repos should set it on create and Get queries should populate both fields).
5. JavaScript: Confirm client-side code expects the UUID `id` field (e.g. `result.id`) and not a numeric ID. Update server-side rendering to ensure `post.ID` is not inserted into JS templates — use `post.PublicID`.
6. Tests: Add tests that fail if a template contains any patterns that would expose internal numeric IDs in URLs or data attributes, and add integration tests that assert returned HTML contains UUIDs in post URLs.
7. Comments & other modules: comment templates and reaction templates may also reference `.ID` for comments and reactions — treat similarly (replace with `PublicID`).

Files & locations to change (recommended edits)
- `internal/modules/post/adapters/http_handler.go`
  - `buildCurrentUser()` — currently returns `"ID": strconv.Itoa(userID)`. Change to fetch user's public id and return `"PublicID": user.PublicID` and `"ID": user.PublicID` if templates expect `ID` key — better: standardize on `PublicID` and update templates.
  - Any places where handler code sets filter params using `strconv.Itoa(userID)` — replace with `user.PublicID`.
- Templates:
  - `templates/base.html` — `?user={{.User.ID}}` → `?user={{.User.PublicID}}`
  - `templates/home.html`, `templates/board.html`, `templates/post_detail.html`, `templates/post_edit.html`, `templates/post_create.html` — replace `{{.ID}}`, `{{.Post.ID}}`, `{{.User.ID}}`, `{{.Comment.ID}}` with `PublicID` variants
- Static JS:
  - `static/js/load-more-posts.js` — currently uses `${post.ID}` when server-side rendering posts; ensure all server-side rendering uses `PublicID` and JS uses `post.id` returned from JSON (uuid). If HTML is server-rendered, use `post.PublicID`.
  - `static/js/post-forms.js` — its `window.location.href = "/posts/${result.id}"` expects `result.id` to be UUID; ensure API returns public id.

Notes about data model & repositories
- Repositories and migrations appear correct: `public_id` columns are present and indexes created (see `migrations/003_post_create_tables.sql` and other migration files). `internal/modules/user/adapters/sqlite_repository.go` sets `user.PublicID = publicID.String()` when creating users.
- Domain structs include both `ID int` and `PublicID string` fields and tag `PublicID` for JSON as `id` which is correct.

Security analysis (detailed)
- Threat: ID enumeration
  - With internal integer IDs leaked in URLs or HTML, an attacker can estimate neighboring resource IDs (e.g. posts 1..N) and attempt to access them. If server-side ACLs rely on obscurity (which they shouldn't), this becomes an immediate access vector. Even when ACLs are enforced, enumeration assists scraping and data mining.
- Threat: Cross-module correlation
  - Internal int IDs exposed across different pages or API endpoints allow cross-correlation and profiling of users across logs or services that leak integer IDs as well (linkability). UUID public IDs decouple internal database identity from public identity and avoid deterministic relationships.
- Threat: Client-side logic confusion
  - If templates embed internal ints in JS and API returns UUIDs (or vice versa), inconsistencies will break client behavior and might accidentally cause the client to call an endpoint with an internal ID.
- Threat: Logs leakage / backward compatibility
  - URLs with integer IDs may be stored in logs, caches, analytics, or external CDN logs — these are harder to purge than UUIDs and reveal internal DB structure long-term.

Attack surface severity matrix (quick)
- Data enumeration / scraping: HIGH
- Unauthorized access (if ACLs weak): HIGH
- Correlation across services: MEDIUM
- Minor breakage (template/JS mismatch): LOW to MEDIUM

Conservative mitigation timeline
1. Immediate (days): Stop rendering integer IDs into HTML/JS. Update `buildCurrentUser` and other template data functions to include `PublicID` and update templates to use `PublicID` in URLs and data-attributes.
2. Short term (1-2 weeks): Update filter building and any query string usage to use `PublicID` for `user` query param. Add unit tests that scan template source and fail if patterns leaking `{{.ID}}` are present.
3. Medium term (2-4 weeks): Run integration tests, re-seed DB, and manually check critical flows (register/login/create/view/edit/delete post) for correct IDs in URLs and API responses.
4. Long term: Add automated checks as part of CI to scan built templates or rendered HTML for internal numeric IDs.

Suggested tests (copy these into your test suite). These tests do not change business logic — they assert correct behavior and will fail until templates/handlers are updated.

1) Template static-scan unit test (go test): ensures templates don't include dangerous patterns
```go
// internal/tests/template_id_leak_test.go
package tests

import (
    "io/ioutil"
    "path/filepath"
    "strings"
    "testing"
)

func TestTemplates_DoNotExposeInternalIDs(t *testing.T) {
    tplDir := "templates"
    // patterns that indicate an internal ID is being rendered into public HTML/JS
    unsafePatterns := []string{
        "{{.ID}}",
        "{{.Post.ID}}",
        "{{.User.ID}}",
        "data-post-id=\"{{.Post.ID}}\"",
        "href=\"/posts/{{.ID}}\"",
        "href=\"/posts/{{.Post.ID}}\"",
        "?user={{.User.ID}}",
    }

    files, err := filepath.Glob(filepath.Join(tplDir, "*.html"))
    if err != nil {
        t.Fatalf("glob templates: %v", err)
    }
    for _, f := range files {
        b, err := ioutil.ReadFile(f)
        if err != nil {
            t.Fatalf("read %s: %v", f, err)
        }
        s := string(b)
        for _, p := range unsafePatterns {
            if strings.Contains(s, p) {
                t.Errorf("template %s contains unsafe pattern %q — replace with PublicID variant", f, p)
            }
        }
    }
}
```

2) Integration test (httptest) — ensure HTML links contain UUIDs
```go
// internal/tests/post_publicid_in_html_test.go
package tests

import (
    "net/http/httptest"
    "regexp"
    "testing"
)

func TestPostDetailPage_UsesUUIDInUrl(t *testing.T) {
    // start the real handler (or a minimal app) and request a known post public id
    // This snippet assumes you have a way to create test data and start the server handler

    // Example regex for UUID v4
    uuidRx := regexp.MustCompile(` + "`" + `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}` + "`" + `)

    req := httptest.NewRequest("GET", "/posts/750e8400-e29b-41d4-a716-446655440001", nil)
    w := httptest.NewRecorder()

    // replace `YourMuxOrHandler` with actual handler
    // YourMuxOrHandler.ServeHTTP(w, req)

    res := w.Result()
    if res.StatusCode != 200 {
        t.Fatalf("unexpected status: %d", res.StatusCode)
    }

    body := w.Body.String()
    if !uuidRx.MatchString(body) {
        t.Errorf("expected page to contain a UUID somewhere, got: %s", body[:200])
    }

    // check there are no occurrences of "/posts/" followed by a numeric ID
    numIDRegex := regexp.MustCompile(`/posts/\d+`) 
    if numIDRegex.MatchString(body) {
        t.Errorf("found numeric internal IDs in post links: %v", numIDRegex.FindString(body))
    }
}
```

3) API response test — ensure JSON `id` is UUID and `user_id` is UUID
```go
// internal/tests/post_api_publicid_test.go
package tests

import (
    "encoding/json"
    "net/http/httptest"
    "regexp"
    "testing"
)

func TestCreatePostAPI_ReturnsUUIDs(t *testing.T) {
    // Construct authenticated request that creates a post using the app handler
    // and verify the JSON response `id` and `user_id` are UUIDs
    uuidRx := regexp.MustCompile(` + "`" + `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}` + "`" + `)

    // req := httptest.NewRequest(...)
    w := httptest.NewRecorder()
    // YourCreatePostHandler.ServeHTTP(w, req)
    if w.Result().StatusCode != 201 {
        t.Fatalf("unexpected status: %d", w.Result().StatusCode)
    }

    var resp struct {
        ID     string `json:"id"`
        UserID string `json:"user_id"`
    }
    if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
        t.Fatalf("decode: %v", err)
    }
    if !uuidRx.MatchString(resp.ID) {
        t.Errorf("post id not a uuid: %s", resp.ID)
    }
    if !uuidRx.MatchString(resp.UserID) {
        t.Errorf("user_id not a uuid: %s", resp.UserID)
    }
}
```

How to run the suggested tests
- Place the test files under a test package (example: `internal/tests` or `tests/integration`) and then run:

```bash
# run the single package tests (adjust path as needed)
go test ./internal/tests -v

# or run all tests (may include unrelated failures)
go test ./... -v
```

Implementation notes and pitfalls
- If you change template variable names (e.g. from `ID` to `PublicID`) make sure to update every template and JS usage site.
- The `RequireAuth` middleware provides `session.UserID` (internal int). To populate `User.PublicID` into templates you must call `userService.GetByID(ctx, session.UserID)` before rendering templates. This adds a DB call; cache or include `PublicID` in the session object if acceptable and secure.
- Do not change the database schema: keep `id INTEGER PRIMARY KEY` for performance and `public_id TEXT UNIQUE` for external use.

Checklist to complete the hardening
- [ ] Replace all `{{.ID}}` in templates with `{{.PublicID}}` or `{{.Post.PublicID}}`.
- [ ] Update `buildCurrentUser` to set `User.PublicID` (fetch user by `session.UserID`).
- [ ] Update filter building so `FilterParams.UserID` uses public IDs.
- [ ] Update any JS that reads `data-post-id` to expect UUID.
- [ ] Add the template-scan unit test to CI to catch regressions.
- [ ] Run full integration tests after changes and re-seed DB if necessary.

Appendix: quick grep commands you can run locally
```bash
# Find templates rendering `.ID` or `.Post.ID`
grep -R "{{\.?\(Post\|User\)\.ID}}" templates || true

# Find server code converting ints to strings for template data
grep -R "strconv.Itoa" -n internal || true

# Find `data-post-id` usages
grep -R "data-post-id" -n || true
```

If you want, I can next:
- Produce a concrete patch (in small, reviewable commits) to update templates and handlers to use `PublicID` and wire `buildCurrentUser` to fetch the `User.PublicID` (I will not change database or repository code). OR
- Generate the test files in the repo so you can run them immediately. (You asked not to alter codebase; I respected that. If you want me to add the tests, tell me and I will create them.)

---
Report generated: automated scan + manual verification of matched templates/handlers. If you want, I can now either (A) prepare the fix patch (templates + handler updates) or (B) add the suggested test files to the repo. Which would you prefer?