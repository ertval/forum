platform-packages

Scope
- Inspect `internal/platform` packages and adapters to ensure they do not expose internal INT IDs publicly and follow the INT-internal + PublicID-UUID external pattern described in `SCHEMA_REFACTOR_STATUS.md`.

Manual Findings (per-package)

1) `internal/platform/httpserver` (if present)
- Look for middleware or URL-building helpers that might insert internal IDs into routes or cookies.
- Recommendation: middleware that maps `session_token -> session.UserID` is fine, but any helper that converts user IDs to strings for cookies or JSON must use `PublicID` when used in external responses (cookies should use session token, not numeric ID).

2) `internal/platform/templates` / `templates/` directory
- Templates currently include `{{.ID}}` in multiple places (see `docs/reviews/post.md`). Template engine is part of `platform` responsibilities; ensure templates reference `PublicID` when generating URLs and data attributes.

3) `internal/platform/logger` and platform utilities
- Logging of IDs to stdout/stderr may print internal IDs for debugging. Avoid logging internal IDs in places that could be captured and exposed (e.g. error messages returned to clients). Keep logs using internal IDs only when necessary; consider logging `PublicID` alongside `ID` for correlation when logs are public.

4) `internal/platform/config` and other infra packages
- No direct exposure problems found; these packages rarely deal with domain IDs. Ensure config values used in templates / client responses do not embed internal IDs.

Security Recommendations for platform packages
- Sanitize any template data: handlers should map domain objects to view models specifically exposing only `PublicID` for identifiers and not include `ID` fields in the view model or JSON map.
- Middleware should not set cookies or headers that contain internal integers. Use session tokens or public UUIDs only.
- Centralize view model creation in a platform helper that enforces `ID` -> `PublicID` mapping to reduce human error.

Quick Mitigations
- Add a compile/test-time checker (see `tests/id_exposure_test.go`) that fails on common patterns that leak `ID` or `UserID` internal integers to templates or JSON maps.

Next Steps
- Run the automated tests added in `tests/id_exposure_test.go` and fix exposed templates/handlers accordingly.

