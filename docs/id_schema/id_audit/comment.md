Comment Module - ID Handling Audit
=================================

Summary:
- The `comment` domain entity contains both `ID int` and `PublicID string` fields as required by the schema refactor.
- The `comment` HTTP handler (`internal/modules/comment/adapters/http_handler.go`) is currently a placeholder and does not implement route handling. No direct exposure of internal INT IDs appears in that file.
- The SQLite repository (`internal/modules/comment/adapters/sqlite_repository.go`) uses method signatures with `int` for IDs (e.g. `GetByID(ctx, commentID int)`), which is appropriate for internal queries. However, the repository implementation is a TODO and does not currently show generation or persistence of `public_id` values.

Findings:
- Domain: `internal/modules/comment/domain/comment.go` correctly defines both `ID int` and `PublicID string`.
- Repository: Implementation is incomplete and lacks the SQL statements that would insert or query `public_id`. There is no evidence of `public_id` generation (e.g. UUID) in the repo.
- Handlers: Not implemented; currently safe because they do not emit responses, but will require careful implementation to use `PublicID` in URLs and JSON.

Risks and Security Issues:
- Because the repository does not yet assign `PublicID` when creating comments, a future implementation could mistakenly return or use the internal `ID` in responses or URLs.
- Templates or frontend JS that rely on integer IDs must be updated to use `PublicID` (string UUID) for any public-facing data attributes or links.

Recommendations:
1. Repository: On `Create`, generate a UUID for `PublicID` (prefer `github.com/gofrs/uuid` or `github.com/google/uuid`) and insert `public_id` into the database. Query functions exposed to the outside (GetByPublicID) should accept `public_id string`.
2. Handlers: Accept `public_id` strings from URL paths and convert to internal IDs only when calling service/repository layer if necessary. Never render internal `ID` in templates or JSON.
3. Templates/JS: Use `Comment.PublicID` (or exposed `id`) in data attributes and DOM ids: e.g. `data-comment-id="{{.PublicID}}"` and `id="comment-{{.PublicID}}"`.
4. Tests: Add unit tests that assert repository creates `public_id` and that HTTP handlers render/produce only `public_id` in responses.

Next Steps:
- Implement repository SQL with `public_id` and UUID generation.
- Implement handlers that read `public_id` from URLs and use int `UserID` from session for auth.
