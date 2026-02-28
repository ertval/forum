# Logging Implementation Plan

## Current State Analysis

The forum application currently has the following logging setup:

1. **Logger Package**: A structured logger exists at `internal/platform/logger/logger.go` with support for different log levels (DEBUG, INFO, WARN, ERROR) and structured field logging.

2. **Initialization**: Logger is properly initialized in `cmd/forum/main.go` and passed to the dependency injection system.

3. **Missing Functionality**:
   - HTTP request/response logging middleware is just a placeholder
   - Actual log output is not visible in terminal during development
   - Errors in HTTP handlers are logged using `fmt.Printf` instead of the structured logger
   - No logging of API calls and their response codes

## Issues Identified

1. **No HTTP Logging**: The `Logger()` middleware in `internal/platform/httpserver/middleware.go` doesn't actually log anything - it's just a placeholder implementation.

2. **Improper Error Logging**: Some HTTP handlers use `fmt.Printf` to log errors instead of the structured logger, as seen in the auth module's handler.

3. **No Visibility**: The logger is not properly connected to show output in the terminal during development.

## Proposed Solution

### 1. Implement HTTP Request/Response Logging Middleware

Create a proper logging middleware that captures:
- Request method, path, and headers
- Response status code
- Response time
- Client IP address
- Request body (for debugging, with size limits)

### 2. Integrate Logger with HTTP Server

Modify the `RegisterMiddleware` function in `server.go` to apply middleware correctly.

### 3. Replace fmt.Printf with Structured Logger

Update all HTTP handlers to use the structured logger instead of `fmt.Printf` for error messages.

### 4. Ensure Logger Output Visibility

Ensure the logger is configured to output to stdout during development so that logs are visible in the terminal.

## Implementation Steps

### Step 1: Enhance the Logger Middleware
- Implement the `Logger()` function in `internal/platform/httpserver/middleware.go`
- Add logging of request details (method, URL, IP, UA)
- Add logging of response details (status code, response time)

### Step 2: Update RegisterMiddleware Function
- Implement the middleware registration to properly chain middleware

### Step 3: Update HTTP Handlers
- Replace all `fmt.Printf` calls with structured logger calls
- Ensure proper error logging with context

### Step 4: Test Logging Functionality
- Verify that HTTP requests are properly logged
- Verify that errors are properly logged with context
- Verify that logs appear in the terminal during development

## Files to Modify

1. `internal/platform/httpserver/middleware.go` - Implement the Logger middleware
2. `internal/platform/httpserver/server.go` - Implement RegisterMiddleware function
3. `internal/modules/*/adapters/http_handler.go` - Replace fmt.Printf with logger
4. Potentially `internal/platform/logger/logger.go` - Ensure proper output configuration

## Expected Outcomes

After implementation:
- All HTTP requests and responses will be logged with method, path, status code, and response time
- Error messages will be properly structured and visible in the terminal
- Debug information will be available for troubleshooting
- Better visibility into application behavior during development