# Logger & Request-Template Observability Improvement Plan

Status: draft

Purpose
-------
The application currently does not produce helpful logs for runtime errors, HTTP requests/responses, or template rendering failures. This plan outlines concrete changes to make logs visible (console + optional file), capture API calls and response codes, and record template rendering errors with context.

Goals
-----
- Ensure the logger writes to console (stdout) by default, at a configurable log level.
- Add structured HTTP request/response logging middleware that records method, path, status, duration, and remote IP.
- Log API handlers' important events and returned HTTP status codes.
- Capture and log template render errors with template name and request context.
- Provide quick verification steps and tests for the changes.

High-level Plan (implementation steps)
-------------------------------------
1. Audit logger initialization and configuration.
   - Locate logger setup (likely in `internal/platform/logger` and `cmd/forum/wire` where it's constructed).
   - Make sure logger defaults to writing to `os.Stdout` and respects a `LOG_LEVEL` env var.

2. Add a Request/Response logging middleware.
   - Implement a middleware in `internal/platform/httpserver` (or add to existing middleware collection) that:
     - Wraps `http.ResponseWriter` to capture status code and bytes written.
     - Logs method, path, query, status, duration, remote address, and request-id (if available).
   - Ensure middleware runs after recovery and before the handlers.

3. Log handler responses and errors.
   - In API handlers (files under `internal/modules/*/adapters/http_handler.go`) add explicit log entries for:
     - Start of important operations
     - Errors (with stack/err context)
     - Final response status for non-2xx outcomes (the middleware will log status for all requests; but handlers should log domain-level failures with more detail).

4. Improve template render error logging.
   - Wherever templates are rendered (likely `internal/platform/httpserver` helpers or handler code), ensure template errors are logged with:
     - Template name
     - Request path and user id (if available)
     - Error text/trace

5. Tests and manual verification.
   - Unit tests for response-writer wrapper and middleware that assert log entries.
   - Manual checks using `curl` to verify log lines appear and include expected fields.

6. Documentation & rollout.
   - Update `docs/` with these changes (this file)
   - Optional: add configuration notes to `README.md` and `internal/platform/config/config.go`.

Suggested Concrete Changes and Example Code
-----------------------------------------

1) Console logger defaults and config

Recommendation: Ensure the logger is constructed to write to `os.Stdout` by default and accepts `LOG_LEVEL` and `LOG_FORMAT` env vars. Example pseudo-init (conceptual):

```go
// in internal/platform/logger/init.go (or where logger is built)
func NewLogger(level string, format string) logger.Logger {
    out := os.Stdout
    // build logger with chosen format (json / console)
    l := logger.New(logger.WithOutput(out), logger.WithLevel(level), logger.WithFormat(format))
    return l
}
```

Notes:
- Default level should be `info` in production, `debug` available for local development.
- Keep structured logging (key/value) for easier parsing.

2) Request logging middleware (implement in `internal/platform/httpserver/middleware.go`)

Key parts:
- responseWriter wrapper to capture status
- middleware that logs at the end of request

Example:

```go
type responseWriter struct {
    http.ResponseWriter
    status int
    size   int
}

func (rw *responseWriter) WriteHeader(status int) {
    rw.status = status
    rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    if rw.status == 0 {
        rw.status = http.StatusOK
    }
    n, err := rw.ResponseWriter.Write(b)
    rw.size += n
    return n, err
}

func RequestLogger(lgr logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            rw := &responseWriter{ResponseWriter: w}
            next.ServeHTTP(rw, r)
            if rw.status == 0 {
                rw.status = http.StatusOK
            }
            lgr.Info("http.request",
                logger.String("method", r.Method),
                logger.String("path", r.URL.Path),
                logger.Int("status", rw.status),
                logger.Int("size", rw.size),
                logger.Duration("duration_ms", time.Since(start)),
                logger.String("remote", r.RemoteAddr),
            )
        })
    }
}
```

Integration:
- Register this middleware in the HTTP server pipeline (see `internal/platform/httpserver` and `cmd/forum/wire/app.go`). Make sure it runs after recovery middleware so that panics are handled first and have logs too.

3) Log API handlers & responses

Recommendations:
- In API handlers, emit contextual logs for errors. Example pattern:

```go
func (h *HTTPHandler) CreatePostAPI(w http.ResponseWriter, r *http.Request) {
    // parse req
    if err := validate(...); err != nil {
        h.logger.Info("request.validation_failed", logger.String("path", r.URL.Path), logger.Error(err))
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    if err := h.service.Create(...); err != nil {
        h.logger.Error("post.create_failed", logger.Error(err), logger.String("user", userID))
        http.Error(w, "internal", http.StatusInternalServerError)
        return
    }

    h.logger.Info("post.created", logger.String("post_id", id), logger.String("user", userID))
    w.WriteHeader(http.StatusCreated)
}
```

Note: the middleware will still emit a top-level `http.request` log for every request; handler logs should add domain-level context for failures and important events.

4) Template rendering errors

Wherever templates are executed, ensure errors are logged with template name and request context. Example:

```go
if err := templates.ExecuteTemplate(w, "post_detail", data); err != nil {
    lgr.Error("template.render_failed",
        logger.String("template", "post_detail"),
        logger.String("path", r.URL.Path),
        logger.Error(err),
    )
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
}
```

5) Tests & manual verification

- Unit tests:
  - Test `responseWriter` sets status correctly and middleware logs expected fields.
  - Mock logger (or capture stdout) and assert key-value pairs are present.

- Manual checks:
  - Start server locally `go run cmd/forum/main.go` with `LOG_LEVEL=debug`.
  - Run a few requests and verify log output contains `http.request` lines and template errors:

```bash
LOG_LEVEL=debug go run cmd/forum/main.go
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/api/posts -d '{...}'
```

Rollout and Safety
-------------------
- Feature toggle: consider enabling verbose request logging only when `LOG_LEVEL=debug` to avoid excessive logs in production.
- Backwards compatibility: middleware should be additive and not change handler behavior except capturing status.

Estimated Effort
----------------
- Audit & logger defaults: 1-2 hours
- Middleware + response-writer: 2-4 hours
- Template error propagation & handler updates: 2-3 hours (spread across modules)
- Tests & verification: 2-3 hours

Next Steps (implementation order)
--------------------------------
1. Audit current logger init and wire locations (`cmd/forum/wire`) and confirm how logger is injected.
2. Implement `responseWriter` + `RequestLogger` middleware and wire into server.
3. Add template error logging points in template render helper(s).
4. Add a few key handler logs for critical failure paths.
5. Add tests and perform manual verification.

References
----------
- Project's DI wiring: `cmd/forum/wire` (where services/handlers are constructed)
- HTTP server wrapper: `internal/platform/httpserver`
- Logger usage pattern in repo: `internal/platform/logger`

--
Generated: by dev plan for logging and observability improvements
