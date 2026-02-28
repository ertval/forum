Notification Module - ID Handling Audit
======================================

Summary:
- The `notification` domain (`internal/modules/notification/domain/notification.go`) currently does not include a `PublicID` field; notifications are modelled only with internal `ID int` values.
- Repository and handler code operate with internal integer IDs and the repository implementations are TODO placeholders.

Findings:
- Domain: `Notification` struct lacks `PublicID string`.
- Repository: `internal/modules/notification/adapters/sqlite_repository.go` does not show generation or persistence of `public_id` values.
- Handlers: `internal/modules/notification/adapters/http_handler.go` are placeholders, which reduces immediate exposure risk but are incomplete.

Risks and Security Issues:
- Notifications should still have a public identifier when exposed via APIs (e.g., to mark a notification read by the client); without `public_id` the app might use internal `ID` in client-side code, leaking DB internals.

Recommendations:
1. Domain: Add `PublicID string` to the `Notification` struct and ensure JSON responses expose only the `PublicID`.
2. Repository: Persist `public_id` and provide lookups/modifications by `public_id` when handling API requests.
3. Handlers & Templates: When rendering notification-related links or data attributes, use `PublicID`.

Next Steps:
- Update domain, repository, handlers and tests to ensure `public_id` is created and used for public endpoints.
