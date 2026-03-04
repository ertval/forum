# Session handling and authentication middleware

## Overview
- Sessions are server-side entities (`domain.Session`) stored in SQLite. They include internal `ID`, public `PublicID` (UUID), internal `UserID` (int), an opaque `Token`, `ExpiresAt`, `IPAddress`, and `UserAgent`.
- Auth operations live behind the `AuthService` interface: `Register`, `Login`, `Logout`, `ValidateSession`, `RefreshSession`, `GetSession`.
- Sessions are persisted by `SQLiteSessionRepository` which generates `PublicID` UUIDs and returns internal IDs.

(See code: [internal/modules/auth/domain/session.go](internal/modules/auth/domain/session.go#L1-L50), [internal/modules/auth/ports/service.go](internal/modules/auth/ports/service.go#L1-L38), [internal/modules/auth/adapters/sqlite_session_repository.go](internal/modules/auth/adapters/sqlite_session_repository.go#L1-L120))

## Cookie & token details
- On successful `Login` or `Register`, the service creates an opaque session token and stores a `Session` row with `ExpiresAt = now + configured duration`.
- The HTTP handlers set an HttpOnly cookie (default name `session_token`) with the session token value, `Path: /`, `Expires` = session expiry, `HttpOnly: true`, `SameSite=Lax`, and `Secure` controlled by configuration.

(See cookie set/clear in [internal/modules/auth/adapters/http_handler_api.go](internal/modules/auth/adapters/http_handler_api.go#L99-L150) and logout handlers in [internal/modules/auth/adapters/http_handler_api.go](internal/modules/auth/adapters/http_handler_api.go#L154-L188) and [internal/modules/auth/adapters/http_handler_page.go](internal/modules/auth/adapters/http_handler_page.go#L70-L93))

## Lifecycle & policies
- On login the service deletes existing sessions for the user (enforcing one session per user by default).
- The repository supports `Create`, `GetByToken`, `GetByUserID`, `Update` (refresh expiry), and `Delete` (logout).
- `ValidateSession` (service) checks DB for token and expiry; `RefreshSession` extends expiry when applicable.

(See flow in [internal/modules/auth/application/service.go](internal/modules/auth/application/service.go#L34-L120))

## Middleware: `OptionalAuth` vs `RequireAuth`
- The project exposes two middleware behaviors via `AuthMiddleware`:
  - `RequireAuth()` — enforces authentication. If the session cookie is missing or invalid it responds with `401 Unauthorized` for API requests (and an HTTP error for pages). Use this when the route must be accessed only by authenticated users (example: creating/editing posts, settings endpoints, private APIs).
  - `OptionalAuth()` — attempts to validate authentication only if a session cookie is present; it does not error if there is no cookie or the token is invalid. Use this for public pages or APIs that should work for both anonymous and signed-in users (example: home page, post detail, comments listing, pages that show a "Login" CTA when anonymous and user info when authenticated).

Why both are necessary:
- Some endpoints require a user identity to proceed (authorization, write operations). `RequireAuth` ensures those handlers never run without a valid user.
- Many pages and APIs should be accessible to both anonymous and authenticated users; `OptionalAuth` lets handlers read user info when available (e.g., to render a "Like" button state or personalized UI) but leaves anonymous access intact.

How they behave internally:
- The middleware reads the cookie (`session_token`), calls `AuthService.ValidateSession(token)`, fetches the user record, and stores the user's public UUID in the request context under the `UserIDKey` context key (never the internal int ID). If validation succeeds the request context is enriched; otherwise:
  - `OptionalAuth()` simply continues without the user in context.
  - `RequireAuth()` returns `401` (for `/api/` paths) or an unauthorized page response.

(See implementation in [internal/modules/auth/adapters/middleware.go](internal/modules/auth/adapters/middleware.go#L1-L80) and middleware interface in [internal/modules/auth/ports/middleware.go](internal/modules/auth/ports/middleware.go#L1-L35))

## Where to change behavior
- Cookie name and `Secure` flag: configured via DI (`ServiceContainer.SessionCookieName()` and `ServiceContainer.SecureCookies()` used by `HTTPHandler`). See [internal/modules/auth/adapters/http_handler.go](internal/modules/auth/adapters/http_handler.go#L1-L55).
- Session duration: injected when creating the auth service (`sessionDuration`). See `NewService` in `application/service.go`.
- Token generation/validation: in the auth service (`generateSessionToken`, `ValidateSession`, `RefreshSession`). Inspect `internal/modules/auth/application/service.go` for details.

## Quick examples (common usage)
- Use `RequireAuth()` on API routes that perform mutations (POST /api/posts, POST /api/comments, user settings APIs).
- Use `OptionalAuth()` on public pages and read-only API endpoints where UI varies by signed-in state (home, post detail, comments listing).

---
Generated on: 2026-03-03
