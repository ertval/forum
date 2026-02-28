Reaction Module - ID Handling Audit
==================================

Summary:
- The `reaction` domain entity contains both an `ID int` and `PublicID string` field, aligned with the schema refactor.
- Repository code (`internal/modules/reaction/adapters/sqlite_repository.go`) is a TODO and does not show `public_id` generation or `public_id` column usage.
- Handlers are placeholders and not yet emitting responses.

Findings:
- Domain: `internal/modules/reaction/domain/reaction.go` includes `PublicID string` and `ID int` correctly.
- Repository: Implementation placeholders do not yet generate or persist `public_id` values. SQL comments show inserts without `public_id` column.
- Handlers: Not implemented; when implemented they should use `PublicID` for public endpoints.

Risks and Security Issues:
- If repository `Create` does not persist `public_id`, or if handlers/JS use internal `ID` for client-side attributes, internal DB IDs could be leaked.

Recommendations:
1. Repository: Generate and persist `public_id` on creation. Use `public_id` for any external lookups.
2. Handlers: Accept `public_id` in URLs and do not include internal `ID` in JSON or HTML attributes.
3. Templates/JS: Prefer `Reaction.PublicID` as data attributes if reactions are exposed in the UI.

Next Steps:
- Implement repo create with UUID generation, create unit tests verifying `public_id` present, update handlers and client code to use `public_id`.
