User Module - ID Handling Audit
===============================

Summary:
- The `user` domain exposes both `ID int` and `PublicID string` fields as expected.
- Several templates and HTML generation points in the project still reference `.User.ID` and other `.ID` fields which are internal integers. These must be replaced with `PublicID` where they are used in URLs or DOM attributes.

Findings:
- Domain: `internal/modules/user/domain/user.go` correctly contains `PublicID string`.
- Handlers: Many user-related handlers are placeholders, but templates (global app templates) include `{{.User.ID}}` in query parameters or links (e.g. `base.html` uses `/board?user={{.User.ID}}`).

Risks and Security Issues:
- Templates that include `User.ID` in URLs or JS pass the internal integer to clients, exposing an internal identifier and enabling enumeration.

Recommendations:
1. Update templates to use `User.PublicID` for any query parameters, links, or data attributes that are visible to clients.
2. Update handler code (if accepting `user` query params in `board` and similar endpoints) to accept `public_id` and, if needed, translate to internal ID server-side.
3. Add template linter unit tests to detect uses of internal `.ID` in templates (see tests added in `tests/id_audit/`).

Next Steps:
- Replace all template occurrences of `{{.User.ID}}` with `{{.User.PublicID}}` and ensure server endpoints accept public ids for user-based filtering.
