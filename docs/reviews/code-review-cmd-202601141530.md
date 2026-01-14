# Code Review: cmd Folder (2026-01-14)

## Executive Summary

The `cmd` directory follows a clean and modular structure using the Ports & Adapters pattern. However, there is a significant performance and consistency issue regarding template management: templates are being parsed from disk on every request in handlers despite being pre-parsed in the wiring layer.

## Critical Issues (Must Fix)

- **ISSUE-1: Redundant Template Parsing & Disk I/O**

  - **Location:** `internal/modules/*/adapters/http_handler_page.go` and `cmd/forum/wire/handlers.go`
  - **Probability:** High (Performance Impact)
  - **Description:** `wire/handlers.go` performs a `template.ParseGlob("templates/*.html")` once at startup and passes the result to all handlers. However, the handlers ignore this shared `*template.Template` and call `template.ParseFiles` on every single HTTP request. This causes unnecessary disk I/O and CPU overhead.
  - **Proposed Fix:** Update the `HTTPHandler` in each module to use the `templates` field provided during initialization instead of re-parsing from disk. Use `h.templates.Lookup("name")` or similar patterns.

- **ISSUE-2: Platform-Business Coupling in Health Check**

  - **Location:** `cmd/forum/wire/app.go`, Line 128-130
  - **Probability:** Medium
  - **Description:** The health check route is configured using specific methods from the `Post` handler (`handlers.Post.Templates()` and `handlers.Post.GetUserWithStats`). The health check (a platform utility) should not depend on a specific business module like `Post`. If the `Post` module is modified or disabled, the core application wiring breaks.
  - **Proposed Fix:** Inject a generic template provider or move the health check's rendering requirements to the `health` package or a neutral shared adapter.

- **ISSUE-3: Fragile Template error handling**
  - **Location:** `cmd/forum/wire/handlers.go`, Line 32-36
  - **Probability:** Medium
  - **Description:** If `template.ParseGlob` fails due to a syntax error in any template, it silently sets `templates = nil`. While the comment claims "API-only mode works", the page handlers (as discussed in Issue 1) will still attempt to parse files and might fail at runtime with inconsistent errors.
  - **Proposed Fix:** If templates are present on disk, a parse error should likely be fatal at startup in development mode, or at least logged as a Critical error.

## Performance & Optimization

- **PERF-1: Hardcoded Log Level**
  - **Location:** `cmd/forum/main.go`, Line 28
  - **Description:** The logger is initialized with `logger.InfoLevel` regardless of the configuration in `cfg.Logger.Level`. This prevents users from enabling DEBUG logs via environment variables.
  - **Optimized Code:**
    ```go
    // In main.go
    logLevel := logger.InfoLevel // Default
    if cfg.Logger.Level == "DEBUG" {
        logLevel = logger.DebugLevel
    }
    // ... initialize logger with logLevel
    ```

## Nitpicks & Best Practices

- **Graceful Shutdown Redundancy**: Both `main.go` and `internal/platform/httpserver/server.go` define a 30-second timeout for graceful shutdown. While safe, it's redundant.
- **Static File Check**: `app.go` checks for `./static` existence. This is good, but it might be better to make the static path configurable in `UploadConfig` (which it partially is, but for the file server it's hardcoded).
- **Wiring File Headers**: Most files have `// INPUT ADAPTERS` or `// OUTPUT ADAPTERS` headers, which is good and follows the project rules.
