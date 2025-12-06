Moderation Module - ID Handling Audit
====================================

Summary:
- The `moderation` module domain (`internal/modules/moderation/domain/report.go`) currently defines `Report` without a `PublicID` field. This deviates from the repository-level schema refactor which mandates a `public_id` UUID for all public entities.
- Repository and handler code use internal integer `id` values (method signatures accept `int` ids).

Findings:
- Domain: `Report` lacks `PublicID string`. The struct only contains internal IDs (int) for `ReporterID` and `TargetID`.
- Repository: `internal/modules/moderation/adapters/sqlite_repository.go` uses internal-int identifiers and contains TODOs — no evidence of `public_id` insertion or generation.
- Handlers: `internal/modules/moderation/adapters/http_handler.go` are placeholders and do not yet expose behaviour.

Risks and Security Issues:
- Without a `PublicID` field and repository support, any future HTTP handler or template may be tempted to use internal `id` integers in URLs or JSON responses, leaking internal DB IDs and making the system susceptible to ID enumeration and inference.

Recommendations:
1. Domain: Add `PublicID string` to the `Report` struct and make it the exposed identifier in JSON (e.g. `json:"id"`).
2. Repository: On `Create`, generate a UUID and persist it to the `public_id` column. Provide lookup methods by `public_id` string for external requests.
3. Handlers: Accept `public_id` as URL parameters for report lookups. Never render internal `id` values in HTML or JSON.

Next Steps:
- Add `PublicID` to the domain type, update migrations if necessary, implement repository `Create` to insert `public_id`, and update handlers and tests accordingly.
