# Health Checker Review

- File reviewed: `internal/platform/health/checker.go`
- Reviewer: GitHub Copilot (assistant)
- Date: 2025-11-16

## Summary

The current `Checker` does two things:

- Pings the SQL database via `db.PingContext()` (with a 1s timeout) — good.
- Verifies that expected API paths are registered on the `http.ServeMux` by constructing a synthetic `http.Request` and calling `router.Handler(req)` to examine the matched route pattern.

This approach correctly detects whether routes are registered in the `ServeMux`, but it does NOT verify that handlers actually execute and respond correctly. In other words, the code affirms registration, not runtime responsiveness.

## Findings (concise)

- ✅ Database ping is implemented and uses context/timeouts.
- ✅ Route enumeration covers core modules (auth, post, user, comment, reaction, moderation, notification).
- ⚠️ Route verification uses `router.Handler(req)` and compares patterns. This only tests registration (pattern matching), not whether the handler runs successfully or returns a healthy response.
- ⚠️ Parameterized route handling attempts to replace `{id}` etc., but fallback string replaces may produce odd paths for uncommon parameter names.
- ⚠️ There is no checking of handler execution (status code, panic recovery, middleware effects, auth/middlewares), so runtime handler failures will not be detected.

## Industry-aligned guidance / Recommendations

1. Keep the DB ping check as-is — it's appropriate for readiness checks.

2. Distinguish probe types:
   - Liveness probe: very lightweight (process alive). Usually only a minimal check is required.
   - Readiness probe: ensure service can accept traffic (DB connections, caches, and optionally a minimal handler invocation).

3. Improve endpoint checks to validate handler execution for a small set of safe, idempotent endpoints (prefer `GET` endpoints that do not require auth/state). Use `httptest.ResponseRecorder` to call the handler returned by `router.Handler(req)` and validate it returns a 2xx (or other expected) status. Do NOT call expensive or state-changing endpoints from health checks.

4. Make endpoint-testing configurable (toggle on/off and limit to specific endpoints) so you can disable deeper tests in environments where they cause issues.

5. Add basic logging for failed checks (so operators can see which specific route or dependency failed).

6. Consider exposing health results as Prometheus metrics (optional) for richer monitoring.

## Suggested code change (safe, minimal):

Replace the current `isRouteRegistered` behavior (pattern-only) with a two-step check:
1. Verify the router returns a matching pattern for the request.
2. Execute the handler using an `httptest.ResponseRecorder` and confirm the handler completes and returns a healthy status (2xx) for safe `GET` endpoints.

Example snippet (illustrative):

```go
// import net/http/httptest

func (c *Checker) isRouteRegistered(ctx context.Context, method, path string) bool {
    // existing parameter substitution for {id} etc. to produce testPath

    req, err := http.NewRequestWithContext(ctx, method, testPath, nil)
    if err != nil {
        return false
    }

    handler, pattern := c.router.Handler(req)
    if pattern != expectedPattern {
        return false
    }

    // Only run handler if method is safe (e.g., GET) to avoid side effects
    if method == "GET" {
        rr := httptest.NewRecorder()
        // run the handler - this exercises middleware and handler code
        handler.ServeHTTP(rr, req)
        // treat 2xx as healthy; you might extend acceptable codes as needed
        return rr.Code >= 200 && rr.Code < 300
    }

    // For non-GET methods default to pattern match only
    return true
}
```

Notes about the snippet:
- This uses `httptest` to execute the handler in-process (no network overhead).
- It exercises middleware and catches panics or runtime errors as handler would run; if middleware blocks due to auth, test requests must call unauthenticated endpoints only.
- Limit the number of endpoints executed to avoid performance impact.

## Operational suggestions

- Configure which endpoints are executed (a small list of safe `GET` endpoints like `/posts`, `/health`, `/auth/session` if safe) via config or environment.
- Maintain the current pattern-based check as a fast fallback for non-GET endpoints.
- Add a small timeout for handler execution controlled by context; keep health checks very short (e.g., 500ms - 1s per handler).
- Log the failing endpoint and the recorder status body for debugging.

## Example next steps

- Implement the `httptest`-based change for a handful of safe GET endpoints in `internal/platform/health/checker.go`.
- Add a configuration flag (e.g., `HEALTH_CHECK_DEEP=true`) to enable/disable handler execution.
- Optionally wire the Checker to accept a list of endpoints to actively call.

## Conclusion

The existing checker is good at detecting route registration and DB connectivity. To align with industry best practices for verifying that endpoints respond properly, add a limited, configurable in-process handler invocation for a small set of safe endpoints, keep the DB check, and expose useful logs/metrics for failed checks.

---

Saved: `docs/health_checker_review.md`
