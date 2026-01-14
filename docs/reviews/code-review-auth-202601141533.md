# Code Review: Auth Module

**Date:** 2026-01-14 15:33
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `auth` module implements session-based authentication using cookies. It strictly separates internal integer IDs from public UUIDs, complying with security requirements. The module uses `bcrypt` for password hashing and proper context propagation for user identity in middleware. However, there are concurrency issues in session validation and potential denial-of-service vectors in the registration flow.

## Critical Issues (Must Fix)

- **ISSUE-1: Concurrent Map Read/Write in Middleware (Unlikely but Context Related)**

  - **Location:** `internal/modules/auth/adapters/middleware.go`
  - **Probability:** Low (Standard library Context is thread-safe, but usage pattern is important)
  - **Description:** The middleware correctly uses `context.WithValue` which is immutable and thread-safe. However, if any other part of the application attempts to _modify_ values inside the context (unlikely here but possible in larger apps), it would race. The current implementation looks safe, but watch out for mutable objects stored in context.
  - **Correction:** The current implementation is actually **Correct**. `context.WithValue` returns a _new_ context.

- **ISSUE-2: Registration DoS via Bcrypt**

  - **Location:** `internal/modules/auth/application/service.go` (`Register`)
  - **Probability:** Medium
  - **Description:** `Register` performs a bcrypt hash (expensive operation) _after_ checking for duplicates but _synchronously_ in the request handler. An attacker could flood the `/register` endpoint with unique emails/usernames to exhaust CPU resources, as bcrypt is designed to be slow. Use of `bcrypt.DefaultCost` (usually 10) is standard, but high volume can still be DoS.
  - **Proposed Fix:** Implement rate limiting on the `/api/auth/register` endpoint (and `/login`) at the infrastructure or middleware level. This is an architectural fix rather than a code logic fix, but critical for auth modules.

- **ISSUE-3: Session Fixation / Token Entropy**
  - **Location:** `internal/modules/auth/application/service.go`
  - **Probability:** Low
  - **Description:** Session tokens are UUID v4. While generally unique, UUIDs are not designed to be cryptographically secure session identifiers (though v4 is random, its entropy is 122 bits).
  - **Proposed Fix:** Use `crypto/rand` to generate a 32-byte random string (base64 encoded) for session tokens instead of UUIDs to ensure maximum entropy and resistance to prediction.

## Performance & Optimization

- **PERF-1: Duplicate User Lookups in Middleware**
  - **Description:** `OptionalAuth` and `RequireAuth` middleware fetch the _full user entity_ from the database (`userService.GetByID`) on _every single request_ to populate the context.
  - **Optimized Code:** if `ValidateSession` already returns `UserID`, and the goal is just to put `PublicID` in the context, consider caching the `IntID -> PublicID` mapping or storing `PublicID` in the `sessions` table (denormalization) to avoid the second DB roundtrip to the `users` table.

## Nitpicks & Best Practices

- **Cookie Security:** The `Secure` flag is explicit set to `false` in `http_handler_api.go`.
  ```go
  Secure:   false, // Set to true in production with HTTPS
  ```
  This MUST be conditional based on the environment (e.g., via config flag `IsProduction`). Hardcoding `false` is dangerous if deployed to prod.
- **Error Leakage:** In `RegisterAPI` lines 53-61, logic attempts to parse error strings to decide the HTTP status code. This is brittle. Comparing error values/types (`errors.Is`) is preferred over string matching.
- **Hardcoded Session Duration:** Session duration is passed to the service info, but check if "Keep me logged in" functionality is needed or if the duration is fixed for everyone.

---
