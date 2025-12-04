# Unified Logger Improvement Plan

## Current State Analysis & Issues Identified

The forum application currently has the following logging setup:

1. **Logger Package**: A structured logger exists at `internal/platform/logger/logger.go` with support for different log levels (DEBUG, INFO, WARN, ERROR) and structured field logging.
2. **Initialization**: Logger is properly initialized in `cmd/forum/main.go` and passed to the dependency injection system.
3. **Missing Functionality**:
   - HTTP request/response logging middleware is just a placeholder with no actual logging
   - Recovery middleware not implemented, so panics are not logged
   - Template rendering errors are not logged in handlers
   - No logging of API calls and response codes
   - Errors in general are not visible in terminal (some handlers use `fmt.Printf` instead of structured logger)
   - Actual log output is not visible in terminal during development

## Goals

- Ensure the standard logger writes to console (stdout) by default, at a configurable log level
- Add structured HTTP request/response logging middleware that records method, path, status, duration, and remote IP
- Implement recovery middleware to log panics
- Log API handlers' important events and returned HTTP status codes
- Capture and log template render errors with template name and request context
- Replace all `fmt.Printf` calls with structured logger calls throughout the project
- Ensure consistent logging implementation across all modules
- Provide quick verification steps and tests for the changes

## Unified Implementation Plan

### 1. Audit and Enhance Logger Configuration
- Locate logger setup (in `internal/platform/logger` and `cmd/forum/wire` where it's constructed)
- Ensure logger defaults to writing to `os.Stdout` and respects `LOG_LEVEL` and `LOG_FORMAT` env vars
- Make sure the standard logger is configured to output to stdout during development for terminal visibility
- Ensure all modules consistently use the same logger instance and configuration

### 2. Implement Logging Middleware Components
- Implement a proper `Logger()` middleware in `internal/platform/httpserver/middleware.go` that:
  - Wraps `http.ResponseWriter` to capture status code and bytes written
  - Logs method, path, query, status, duration, remote address, and request-id (if available)
  - Logs incoming requests (method, URL, user agent, etc.) at Info level
  - Logs outgoing responses (status code, duration) at Info level
- Implement a proper `Recovery()` middleware in `internal/platform/httpserver/middleware.go` that:
  - Recovers from panics
  - Logs the panic details at Error level
  - Returns 500 status code

### 3. Update HTTP Server Integration
- Implement the `RegisterMiddleware` function in `internal/platform/httpserver/server.go` to apply middleware correctly
- Register the new middleware in the HTTP server pipeline (see `internal/platform/httpserver` and `cmd/forum/wire/app.go`)
- Ensure middleware runs in the correct order: recovery -> logging -> handlers
- Pass logger to `initServer` and use it in middleware registration

### 4. Enhance Dependency Injection for Logger
- Update `wire.ServiceContainer` to include the logger
- Add `Logger() *logger.Logger` method to the `ServiceContainer` interface
- Update `initHandlers` to pass logger to ServiceContainer

### 5. Improve Template Render Error Logging
- Wherever templates are rendered (in `internal/platform/httpserver` helpers or handler code), ensure template errors are logged with:
  - Template name
  - Request path and user id (if available)
  - Error text/trace
- Use the standard logger consistently with structured fields

### 6. Update All HTTP Handlers for Consistent Logging
- Replace all `fmt.Printf` calls with structured logger calls throughout all modules
- In API handlers (files under `internal/modules/*/adapters/http_handler.go`) add explicit log entries for:
  - Start of important operations
  - Errors (with stack/err context)
  - Final response status for non-2xx outcomes (the middleware will log status for all requests; but handlers should log domain-level failures with more detail)
- Ensure consistent logging format across all modules using the standard logger
- Follow the same pattern for error logging throughout all handlers

## Suggested Concrete Changes and Example Code

### 1) Console logger defaults and config

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

### 2) Request and Recovery logging middleware (implement in `internal/platform/httpserver/middleware.go`)

Key parts:
- ResponseWriter wrapper to capture status
- Logger middleware that logs at the end of request
- Recovery middleware that catches panics and logs details

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

func Logger(lgr logger.Logger) func(http.Handler) http.Handler {
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

func Recovery(lgr logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    lgr.Error("panic.recovered", 
                        logger.String("path", r.URL.Path),
                        logger.String("method", r.Method),
                        logger.Any("error", err),
                    )
                    http.Error(w, "internal server error", http.StatusInternalServerError)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

### 3) Consistent logging in API handlers

Recommendations:
- In all API handlers, emit contextual logs for errors using the standard logger. Example pattern:

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

### 4) Template rendering errors

Wherever templates are executed, ensure errors are logged with template name and request context using the standard logger. Example:

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

## Implementation Steps (in order)

1. **Audit current logger initialization and wire locations** (`cmd/forum/wire`) and confirm how logger is injected - ensure consistent usage of the standard logger
2. **Implement responseWriter wrapper** and `Logger` + `Recovery` middleware and wire into server
3. **Update RegisterMiddleware function** in server.go to properly chain middleware in correct order
4. **Update ServiceContainer interface** in wire package to include logger access
5. **Update all handlers** to use structured logger consistently instead of fmt.Printf
6. **Add template error logging points** in template render helper(s)
7. **Add tests and perform manual verification**

## Files to Modify

1. `internal/platform/httpserver/middleware.go` - Implement proper Logger and Recovery middleware
2. `internal/platform/httpserver/server.go` - Implement RegisterMiddleware function
3. `internal/modules/*/adapters/http_handler.go` - Replace fmt.Printf with standard logger, ensure consistent formatting
4. `cmd/forum/wire/services.go` - Update ServiceContainer interface and implementation to include logger
5. `cmd/forum/wire/app.go` - Pass logger to initServer and middleware
6. Potentially `internal/platform/logger/logger.go` - Ensure proper output configuration

## Testing & Verification

- Unit tests:
  - Test `responseWriter` sets status correctly and middleware logs expected fields
  - Mock logger (or capture stdout) and assert key-value pairs are present
  
- Manual checks:
  - Start server locally `go run cmd/forum/main.go` with `LOG_LEVEL=debug`
  - Run a few requests and verify log output contains `http.request` lines and template errors:

```bash
LOG_LEVEL=debug go run cmd/forum/main.go
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/api/posts -d '{...}'
```

## Expected Outcomes

After implementation:
- All HTTP requests and responses will be logged with method, path, status code, and response time
- Panic recovery will be properly logged with context
- Error messages will be properly structured using the standard logger consistently throughout the project
- Template errors will be captured with context and logged
- Debug information will be available for troubleshooting
- Better visibility into application behavior during development
- Consistent logging format and approach across all modules

## Rollout and Safety
- Feature toggle: consider enabling verbose request logging only when `LOG_LEVEL=debug` to avoid excessive logs in production
- Backwards compatibility: middleware should be additive and not change handler behavior except capturing status
- Ensure all modules consistently use the same standard logger implementation

## Estimated Effort
- Audit & logger defaults: 1-2 hours
- Middleware implementations: 3-5 hours
- Template error propagation & handler updates: 3-4 hours (spread across modules)
- Update ServiceContainer and DI: 1-2 hours
- Tests & verification: 2-3 hours

## References
- Project's DI wiring: `cmd/forum/wire` (where services/handlers are constructed)
- HTTP server wrapper: `internal/platform/httpserver`
- Logger usage pattern in repo: `internal/platform/logger`

---

## Key Emphasis: Consistent Standard Logger Usage

Throughout this implementation, it is crucial that:
1. ALL modules use the standard logger provided through the dependency injection system
2. The same structured logging format is applied across all modules with consistent field names
3. All error logging uses the same pattern to ensure uniformity in logs
4. The same log levels (DEBUG, INFO, WARN, ERROR) are used consistently across the application
5. All modules follow the same approach for including context in log messages

This will ensure that the logging system provides a consistent, reliable, and maintainable approach to observability throughout the entire application.