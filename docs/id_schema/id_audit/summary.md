ID Handling Security Audit — Summary & Remediation
===============================================

Scope
- Modules inspected: `comment`, `moderation`, `notification`, `reaction`, `user` and project templates.
- Focus: Ensure internal INT primary keys are never exposed to clients (URLs, JSON, templates, DOM attributes), and that every public entity has a UUID `public_id` persisted by repositories.

High-level Findings
- Several modules (notably `moderation` and `notification`) are missing a `PublicID` field on the domain entity. This violates the repository schema contract (INT primary key + public UUID) described in `docs/SCHEMA_REFACTOR_STATUS.md`.
- Some modules (e.g. `comment`, `reaction`) have `PublicID` on the domain type, but the repository implementations are incomplete/TODOs and do not show evidence of generating or persisting `public_id` values.
- Templates across the app (e.g. `post_detail.html`, `base.html`, `board.html`, `home.html`, `post_edit.html`) reference internal `.ID` fields directly in links and data-attributes (e.g. `/posts/{{.ID}}`, `data-post-id="{{.Post.ID}}"`, `comment-{{.ID}}`). These produce direct exposure of internal integer IDs to clients.

Security Implications
1. ID Enumeration and Resource Discovery: Exposing sequential or predictable internal integer IDs makes it trivial for attackers or crawlers to enumerate resources (posts, comments, users) and test for existence, author relationships, or access control gaps.
2. Horizontal Privilege Escoping: A user discovering internal IDs might attempt API calls or crafted URLs using integer IDs to perform operations against resources they should not be able to access.
3. Linkability Across Systems: Internal database IDs may leak implementation details that facilitate cross-site correlation if the same integer IDs are used elsewhere.
4. SQL/Business Logic Assumptions: If handlers accept integer IDs from clients (via templates/JS), they bypass the public_id abstraction and can create inconsistent flows where internal IDs are mistakenly trusted.

Immediate Remediation Steps (High Priority)
1. Templates: Replace all template occurrences that render or include `.ID` in public contexts with the corresponding `PublicID` fields. For example:
   - `href="/posts/{{.ID}}"` -> `href="/posts/{{.PublicID}}"`
   - `data-post-id="{{.Post.ID}}"` -> `data-post-id="{{.Post.PublicID}}"`
   - DOM element IDs: `id="comment-{{.ID}}"` -> `id="comment-{{.PublicID}}"`
2. Domain: Add `PublicID string` to `Report` (moderation) and `Notification` (notification) structs. Tag `PublicID` for JSON exposure (e.g. `json:"id"`).
3. Repository: Update `Create` methods to generate UUIDs (e.g. `uuid.NewV4()` or `uuid.New()`) and persist the `public_id` column. Provide repository lookup by `public_id` (string) for all externally addressed methods.
4. Handlers: Accept `public_id` (string) in URLs and query parameters and only convert to internal INT IDs server-side for repository/service calls. Never accept or trust integer IDs from clients in public endpoints.
5. Tests: Keep and extend the `tests/id_audit` detector added in this audit — make it part of CI to block PRs that reintroduce `.ID` exposure in templates or forget to generate `public_id` in repositories.

Longer-term Hardening
- Add a template linter rule in the build pipeline that flags `{{.*\.ID.*}}` occurrences.
- Add unit tests for each repository to assert that `Create` populates `PublicID` and that public endpoints return only `public_id` values in JSON.
- Review any JavaScript client code that reads `data-*-id` attributes and ensure it uses `public_id` strings. If internal IDs must be used client-side, ensure they are never communicated to the server as-is from the client.

How to run the detector test
1. From the repository root run:

```bash
go test ./tests/id_audit -v
```

This test is intentionally conservative and will fail if it finds template patterns that reference `.ID` or repository files that do not show any `public_id` handling.

Files Added
- `docs/id_audit/{comment,moderation,notification,reaction,user,summary}.md` — per-module findings and recommendations.
- `tests/id_audit/id_exposure_test.go` — conservative detector test.

If you want, I can now:
- Open PR patches to update templates to use `PublicID` (small mechanical changes)
- Add repository UUID generation and DB insert SQL for missing modules
- Update handlers to accept `public_id` in URLs and convert server-side
