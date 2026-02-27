post

Summary
- Module: `internal/modules/post`
- Purpose: Verify that public-facing identifiers are UUID `public_id` strings, and internal INT `id` values are only used internally and never rendered in templates, JSON responses, or URLs.

Findings

1) Templates exposing internal INT `ID`
- Files and locations found (search results):
  - `templates/base.html` — uses `/posts/{{.ID}}` for post links.
  - `templates/board.html` — uses `/posts/{{.ID}}` for post links.
  - `templates/home.html` — uses `/posts/{{.ID}}` for post links.
  - `templates/post_detail.html` — uses `id="comment-{{.ID}}"`, `data-comment-id="{{.ID}}"` and buttons referencing `{{.ID}}` for comments.

Impact: Templates currently render `{{.ID}}` which (in the current code) maps to internal integer IDs supplied by handlers. This leaks internal DB identifiers in URLs, HTML element IDs and data attributes. This enables enumeration and potential profiling of the database schema, and violates the design in `SCHEMA_REFACTOR_STATUS.md` which mandates exposing `PublicID` UUIDs in external interfaces.

2) `internal/modules/post/adapters/http_handler.go` exposes internal INT IDs in template data and JSON preview objects
- Problems observed (representative code locations):
  - `previewPost["ID"] = post.ID` — preview maps sent to templates and JSON use `post.ID` (internal INT) instead of `post.PublicID`.
  - `previewPost["UserID"] = post.UserID` — user IDs are internal INTs placed into template/JSON data.
  - `CreatePostPage`, `BoardPage`, `HomePage`, `LoadMorePostsAPI` build preview maps using internal IDs.
  - `renderPostDetail` passes `Post` domain object directly to templates; if that object contains `ID int json:"-"` and `PublicID string json:"id"` the template may still reference `.ID` and reveal internal INT if handlers set that in context.

Impact: Because handlers set `ID` and `UserID` fields in maps that are used by templates, templates will render internal values. URLs generated in templates like `/posts/{{.ID}}` will contain INTs.

3) Ownership checks use internal IDs (correct) but the surrounding flow leaks internal IDs to the view layer
- Example: `if post.UserID != userID { ... }` is correct (both ints). Ownership checks should remain internal. However the handler must convert internal IDs to public UUIDs before rendering or building URLs.

Recommendations

- Templates: replace all uses of `{{.ID}}` (for post and comment links and data-attributes) with `{{.PublicID}}` or the appropriate `PublicID` field (for comments use comment.PublicID).
- Handlers: when preparing template data and JSON preview objects, populate `ID` keys with `post.PublicID` or rename to `PublicID`/`id` to avoid ambiguous `ID` usage. Example: `previewPost["id"] = post.PublicID` and `previewPost["user_id"] = post.UserPublicID`.
- JSON responses: ensure JSON marshaling uses domain struct tags so that the `id` field comes from `PublicID` and internal `ID int` is annotated `json:"-"` (already the pattern in `SCHEMA_REFACTOR_STATUS.md`). Avoid constructing custom maps that insert internal `ID` values.
- Comments: update comment template usage (element IDs and data-* attributes) to use `PublicID` so DOM elements use stable UUIDs.
- Tests: add automated checks (see `tests/id_exposure_test.go`) that fail on `{{.ID}}` in any template and on adapter code patterns that directly place `post.ID` or `post.UserID` into template data maps.

Notes
- Correct internal use: keeping `post.ID` and `post.UserID` as INTs in domain and repository layers is good for DB performance (joins, FKs).
- Public API: handlers must accept, parse and route using public UUIDs (e.g. path `/posts/{public_id}`) and the service layer must convert to internal IDs before business logic.

Suggested Implementation Steps
1. Update templates (`templates/*.html`) to use `PublicID`/`id` field names.
2. Update `internal/modules/post/adapters/http_handler.go` to populate template data with `post.PublicID` and `post.UserPublicID`.
3. Update any JavaScript that performs AJAX calls or uses data attributes to use public UUID values.
4. Run `make test` and the new id-exposure tests; iterate until clean.

References
- `docs/SCHEMA_REFACTOR_STATUS.md`
- Files flagged in this report: `templates/base.html`, `templates/board.html`, `templates/home.html`, `templates/post_detail.html`, `internal/modules/post/adapters/http_handler.go` (lines that set `previewPost["ID"]` / `previewPost["UserID"]`).
