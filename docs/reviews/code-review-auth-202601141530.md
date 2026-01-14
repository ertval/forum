# Code Review: Auth Module

## Executive Summary

The `auth` module is structurally sound, following the Hexagonal Architecture with clear separation of concerns. However, there is a **critical performance issue** in the page handlers (templates parsed on every request) and some questionable error handling practices (silencing errors during session cleanup) that need immediate attention. Security is generally handled well (UUIDs, bcrypt), but the password policy is weak.

## Critical Issues (Must Fix)

- **ISSUE-1: Template Parsing on Every Request (Performance Killer)**

  - **Location:** `adapters/http_handler_page.go`, `LoginPage` (line 25), `RegisterPage` (line 44)
  - **Probability:** High (100% of page load requests)
  - **Description:** The code calls `template.ParseFiles` on every single HTTP request. This triggers disk I/O and expensive parsing logic, severely limiting throughput and increasing latency.
  - **Proposed Fix:** Parse templates _once_ during application startup in `NewHTTPHandler` or `main.go` and store them in the `HTTPHandler` struct. Use `t.Lookup("login.html").Execute(...)` to render.

- **ISSUE-2: Ignored Errors in Session Cleanup**

  - **Location:** `application/service.go`, `Login` (line 149), `ValidateSession` (line 194)
  - **Probability:** Medium
  - **Description:**
    - In `Login`, `s.sessionRepo.DeleteByUserID` is called to invalidate old sessions. The error is explicitly ignored with a comment.
    - In `ValidateSession`, `s.sessionRepo.Delete` is called for expired sessions, and the error is ignored (`_ = ...`).
    - **Risk:** If the DB is locked or failing, these write operations fail silently. In `Login`, this could mean a user ends up with multiple active sessions against the "one session per user" policy logic. In `ValidateSession`, expired sessions remain in the DB, potentially clogging it.
  - **Proposed Fix:** At minimum, these errors should be logged using a logger instance. For `Login`, if `DeleteByUserID` fails, we should consider if we want to allow the login to proceed (soft failure) or fail hard.

- **ISSUE-3: Weak Password Policy**
  - **Location:** `application/service.go`, `ValidateCredentials` (line 284)
  - **Probability:** High (Security Risk)
  - **Description:** The password minimum length is set to 6 characters (`v.Password("password", c.Password, 6)`). This is insufficient for modern security standards.
  - **Proposed Fix:** Increase minimum length to at least 8, preferably 12, or implement complexity requirements.

## Performance & Optimization

- **PERF-1: Redundant User Fetching**
  - **Description:**
    - `RegisterAPI` calls `Register` (which creates user and session), then calls `userService.GetByID` to fetch the User just to get the `PublicID`.
    - `LoginAPI` calls `Login` (which gets user), then calls `userService.GetByID` again just to get the `PublicID`.
  - **Optimized Code:** Update `authService.Register` and `authService.Login` to return the full `User` entity or at least the `PublicID`. `Login` already fetches the user internally; returning it would save a database round-trip.

## Nitpicks & Best Practices

- **API Robustness:** In `RegisterAPI` (line 41), error handling relies on `errors.Is` which is good, but the fallback `strings.Contains` logic (lines 53-60) is brittle. New validation error strings could break this mapping.
- **Context Usage:** `adapters/sqlite_session_repository.go` correctly uses `QueryContext` and `ExecContext`. This is good practice.
- **Transaction Safety:** `Register` operation involves multiple steps (User creation, Session creation). These are currently separate DB calls. If `sessionRepo.Create` fails after `userService.CreateUser` succeeds, the user exists but the client gets an error and can't log in immediately without retrying "login" instead of "register". Ideally, these should be in a transaction, but that crosses module boundaries (User vs Auth), which is tricky in this architecture. Accepted as a trade-of for modularity.
- **Deprecation:** `adapters/middleware.go` contains deprecated wrappers (`RequireAuthFunc`, `OptionalAuthFunc`). These should be removed to encourage using the interface-based approach.
