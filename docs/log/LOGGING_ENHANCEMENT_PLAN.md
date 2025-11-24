# Logging Enhancement Plan

## Current Issues
- Logger middleware is not implemented, so no HTTP request/response logging.
- Recovery middleware not implemented, panics not logged.
- Template rendering errors not logged in handlers.
- No logging of API calls and response codes.
- Errors in general not visible in terminal.

## Proposed Fixes

### 1. Implement Logger Middleware
- Modify `httpserver.Logger()` to accept a `*logger.Logger` parameter.
- Log incoming requests (method, URL, user agent, etc.) at Info level.
- Log outgoing responses (status code, duration) at Info level.
- Use a response writer wrapper to capture status code.

### 2. Implement Recovery Middleware
- Modify `httpserver.Recovery()` to accept a `*logger.Logger` parameter.
- Recover from panics, log the panic details at Error level, and return 500.

### 3. Inject Logger into Handlers
- Add `Logger() *logger.Logger` to the `ServiceContainer` interface in each handler.
- Update `wire.ServiceContainer` to include the logger.
- In handlers, log errors (e.g., template rendering failures) at Error level.
- Log successful operations at Debug/Info as needed.

### 4. Update Wire Package
- Pass logger to `initServer` and use it in middleware registration.
- Update `initHandlers` to pass logger to ServiceContainer.

### 5. Configuration
- Ensure logger level is set appropriately (e.g., Info for development).
- Output to stdout/stderr for terminal visibility.

### Implementation Steps
1. Modify middleware functions in `internal/platform/httpserver/middleware.go`.
2. Update `cmd/forum/wire/app.go` to pass logger to middleware.
3. Update ServiceContainer in `cmd/forum/wire/services.go` and handler interfaces.
4. Update all handlers to log errors appropriately.
5. Test logging output in terminal.

### Benefits
- Full visibility into HTTP traffic.
- Error tracking for debugging.
- Better monitoring and troubleshooting.</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/docs/LOGGING_ENHANCEMENT_PLAN.md