# Remaining Audit Findings — Actionable Report

Date: 2026-03-02

This document lists remaining audit findings from the repository that were NOT changed automatically (require design/implementation decisions or larger refactors). Each item includes a short description, severity, suggested fix, and suggested next steps / priority.

---

## 1. N+1 queries: comment reaction counts on post detail
- Severity: HIGH
- Files/areas: `internal/modules/post/adapters/http_handler_page.go` (post detail rendering), `internal/modules/reaction/application/service.go`
- Description: Rendering the post detail iterates comments and calls `CountReactions` per comment, producing many DB round-trips (N comments → many queries). This causes high latency for posts with many comments.
- Suggested fix: Add a batch repository/service API such as `CountReactionsBatch(ctx, []publicIDs, targetType)` that returns counts for many targets in one query. Update the post handler to call it once per page.
- Priority: P0 — implement as a single service/repo change.

## 2. `CountReactions` issues: multiple redundant queries per call
- Severity: HIGH
- Files/areas: `internal/modules/reaction/application/service.go`, reaction repository
- Description: `CountReactions` resolves the target ID repeatedly and executes separate queries for likes and dislikes, leading to 3–5 queries per logical count request.
- Suggested fix: Implement a single repository method `CountLikesAndDislikes(ctx, targetPublicID, targetType)` that resolves target once and counts both reaction types in one aggregated query, or a batch variant for many targets.
- Priority: P0 — refactor repository + service, update callers.

## 3. Full-table subqueries in post list/get queries
- Severity: MEDIUM → HIGH (depending on scale)
- Files/areas: `internal/modules/post/adapters/sqlite_repository.go` (List/Get queries using grouped subqueries for likes/dislikes/comments)
- Description: Current approach does GROUP BY on whole `reactions`/`comments` tables and then JOIN, causing full-table scans even for single post fetches or small paginated results.
- Suggested fix: Replace with correlated subqueries or targeted aggregates per returned row, or use indexed joins. Example: correlated `SELECT COUNT(*) FROM reactions WHERE target_id = p.id AND type='like'` for a single post; for lists use query patterns that let SQLite use indexes, or precompute counters in a denormalized column and update transactionally.
- Priority: P1 — optimize query patterns and add tests for performance regression.

## 4. Database connection pool configuration not applied
- Severity: HIGH
- Files/areas: `internal/platform/database/connection.go`, `internal/platform/config/config.go`
- Description: Config defines connection pool settings (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`) but these settings are not applied to `*sql.DB` (no `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`).
- Suggested fix: Apply the config values during DB initialization (`db.SetMaxOpenConns(cfg.Database.MaxOpenConns)` etc.). Ensure `NewConnection` accepts config or that caller applies settings immediately after connection creation.
- Priority: P0 — simple config plumbing and test.

## 5. Prepared statements / statement reuse
- Severity: MEDIUM
- Files/areas: repository implementations across modules (many `db.QueryRowContext`/`ExecContext` calls)
- Description: Frequently-executed SQL strings are executed without prepared statements. SQLite driver may reparse and re-plan queries each execution which can be expensive.
- Suggested fix: Prepare commonly-used statements once (e.g., session lookup, post list/get, auth validate) and reuse them via `*sql.Stmt` fields on repository structs. Add graceful Close() of prepared statements on shutdown.
- Priority: P1 — targeted set of statements to prepare initially.

## 6. Missing or suboptimal indexes referenced by queries
- Severity: LOW → MEDIUM (depends on data size)
- Files/areas: migration files `migrations/*.sql` (e.g., notifications, categories), post/reaction queries
- Description: Some queries use columns that lack indexes (e.g., `notifications.actor_id`, `categories.name` used in case-insensitive lookups). Missing indexes will cause full table scans as data grows.
- Suggested fix: Add targeted indexes in new migration(s): `idx_notifications_actor (actor_id)`, `idx_categories_name (name COLLATE NOCASE)`, and any indexed access paths identified via EXPLAIN.
- Priority: P2 — add migrations and test.

## 7. `writeJSON` / `parseJSON` duplication across handlers
- Severity: MEDIUM
- Files/areas: handler packages across modules (auth, comment, post, reaction, moderation etc.)
- Description: Utility functions for writing JSON responses and parsing JSON requests are duplicated with differing error logging and signatures. This leads to inconsistent behavior and makes changes tedious.
- Suggested fix: Extract a single shared utility in `internal/platform/httpserver` (or `internal/platform/httpjson`) providing consistent `WriteJSON(w, status, data)` and `ParseJSON(r, v)` helpers. Migrate callers module-by-module.
- Priority: P2 — refactor for consistency; low risk but wide touch surface.

## 8. Session cookie name hard-coded in some handlers
- Severity: MEDIUM
- Files/areas: several handler adapters (comment, post, others)
- Description: Some handlers use a hardcoded cookie name like `session_token` instead of pulling from the `ServiceContainer` (`SessionCookieName()`), producing inconsistency and potential bugs if cookie name changes.
- Suggested fix: Update module `ServiceContainer` interfaces (where appropriate) to expose `SessionCookieName()` and have handlers read from the container (or centralize auth middleware to provide current user in context so handlers don't read cookies directly).
- Priority: P1 — moderate effort across handlers.

## 9. Session token included in JSON responses (security risk)
- Severity: MEDIUM (security)
- Files/areas: `internal/modules/auth/adapters/http_handler_api.go` (Register/Login handlers)
- Description: API responses include the session token in the JSON body while also setting an HttpOnly cookie. Returning tokens in JSON increases XSS risk.
- Suggested fix: Stop returning the raw session token in JSON responses; rely on HttpOnly cookie. If token return is required for API clients, gate it behind an explicit configuration and secure usage (e.g., `response.tokens=true` for machine clients).
- Priority: P0 (security) — discuss client compatibility before change.

## 10. `int64` → `int` conversions from `LastInsertId()`
- Severity: LOW
- Files/areas: repositories that cast `LastInsertId()` to `int` (auth, post, user, moderation)
- Description: `LastInsertId()` returns `int64` and code casts it to `int`. On 32-bit builds this can overflow. While the project targets 64-bit, it's safer to use `int64` for DB IDs or validate the cast.
- Suggested fix: Use `int64` for repository/internal IDs, or add a checked conversion helper that fails on overflow.
- Priority: P3 — low immediate risk but easy to fix if desired.

## 11. Missing `RowsAffected` checks on Update/Delete repository methods
- Severity: LOW
- Files/areas: several repository methods (comment, session, notification update/delete)
- Description: Update/Delete queries execute successfully even if they affected zero rows (no RowsAffected check). This hides TOCTOU conditions where resource was deleted between check and update.
- Suggested fix: Check `res.RowsAffected()` and return a domain `ErrNotFound` when 0. Add tests.
- Priority: P2 — small code changes.

## 12. Missing pagination (LIMIT/OFFSET) on list endpoints
- Severity: LOW → MEDIUM
- Files/areas: comments listing and notifications listing repositories and APIs
- Description: Some list methods lack `LIMIT`/`OFFSET`, potentially returning large result sets into memory for heavy users.
- Suggested fix: Add pagination parameters to repository and service methods, update API handlers and templates to use paginated loads.
- Priority: P1 — implement for endpoints that can return many rows.

## 13. Background goroutines without shutdown/timeout
- Severity: LOW
- Files/areas: reaction/comment services `runInBackground` usage
- Description: Some services spawn goroutines (background counters, async tasks) using `context.Background()` with no cancellation or timeout. These can leak or operate after DB shutdown.
- Suggested fix: Use a shared `async.Run(ctx, ...)` helper with context derived from application shutdown or use the platform `async` package with timeouts and panic recovery. Ensure goroutines honor shutdown signals.
- Priority: P2 — medium effort to standardize.

## 14. Rate limiter micro-race (accept as-is or refine)
- Severity: LOW
- Files/areas: `internal/platform/httpserver/middleware.go` (rate limiter implementation)
- Description: Minor race in entry creation path was mitigated; if more strict correctness is desired, a rework to a token-bucket or atomic counter approach is recommended.
- Suggested fix: Consider replacing with a token-bucket algorithm with atomic counters for higher scale.
- Priority: P3 — optional.

## 15. Template rendering partial writes on error (best practice)
- Severity: LOW
- Files/areas: some handlers render templates directly to `http.ResponseWriter` — e.g., `post` pages
- Description: Rendering directly to the ResponseWriter allows partial HTML to be sent if template execution fails.
- Suggested fix: Render to a buffer (`bytes.Buffer`) first, then copy to `w` on success. Apply to critical page handlers.
- Priority: P2 — low effort fix.

## 16. Prepared statement and query optimization backlog
- Severity: MEDIUM
- Files/areas: all repository files
- Description: There are multiple opportunities to reduce CPU and parsing overhead by preparing frequently-used statements and consolidating repeated query patterns.
- Suggested fix: Identify top 10 most-executed queries (via profiling or tests) and prepare them; consolidate repeated SQL strings into constants; consider a small prepared-statement helper on repository structs.
- Priority: P2 — medium effort.

## 17. JS & CSS suggestions not yet actioned (higher-level)
- Severity: LOW → MEDIUM (maintainability)
- Files/areas: `static/js/` (template fallbacks, duplicate functions), `static/css/` (many hardcoded colors), templates (`base.html` duplicated blocks)
- Description: Several frontend simplifications were recommended (remove template DOM fallbacks, unify reaction handlers, use CSS variables, extract shared templates). These were NOT applied automatically.
- Suggested fix: Prioritize highest-value frontend changes: remove large unused fallback builders (JS), consolidate CSS color usage to variables, and extract shared template partials used in multiple places.
- Priority: P1 — do the large JS fallback removal and CSS variable cleanup first.

## 18. Logger replacement (large refactor)
- Severity: HIGH (maintenance/duplication)
- Files/areas: `internal/platform/logger/*` (custom logger ~600 lines)
- Description: The codebase contains a hand-rolled structured logger that reimplements standard library facilities (`log/slog` in Go 1.21+). Replacing with `slog` reduces maintenance and line count.
- Suggested fix: Plan a careful refactor: replace custom logger API with an adapter to `slog.Logger`, update uses across packages, and keep a small compatibility wrapper for any missing features (HTTP formatting, colorization in dev). Run tests and ensure logging behavior remains acceptable.
- Priority: P1 — high-impact but requires a staged approach.

## 19. Security: default session secret & config guidance
- Severity: MEDIUM
- Files/areas: `internal/platform/config/config.go`
- Description: Default session secret is weak in development; presence of the secret field suggests signing but current tokens are UUIDs. Improve guidance and enforce warnings for insecure defaults.
- Suggested fix: Generate a secure development secret automatically or log a warning if default secret is used in non-dev environments. Audit session signing usage and ensure secret is used meaningfully or removed.
- Priority: P1 — security-focused.

## 20. Misc smaller findings
- Integer overflow / 32-bit safety (see item 10)
- `repeatPlaceholders` already fixed; review other micro-allocations (string building, slice preallocation) for hot paths
- Consolidate small doc/comment consistency issues (done in many locations but spot checks remain)

---