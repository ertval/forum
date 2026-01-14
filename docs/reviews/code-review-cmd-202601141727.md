# Code Review: `cmd/` Directory

**Date:** 2026-01-14 17:27  
**Reviewer:** Antigravity AI  
**Scope:** `cmd/forum/main.go`, `cmd/forum/wire/*.go`

---

## Executive Summary

The `cmd/` directory demonstrates **solid structural design** with clean separation between entry point (`main.go`) and dependency injection (`wire/`). The code follows idiomatic Go patterns with proper error handling and a layered initialization approach. However, there are **several critical issues** around shutdown logic, resource management, and error handling that need attention.

---

## Table of Contents

- [Critical Issues (Must Fix)](#critical-issues-must-fix)
- [Performance & Optimization](#performance--optimization)
- [Nitpicks & Best Practices](#nitpicks--best-practices)

---

## Critical Issues (Must Fix)

### ISSUE-1: Graceful Shutdown Context Not Used

- **Location:** `cmd/forum/main.go`, Lines 60-71
- **Probability:** High
- **Description:** A timeout context is created for graceful shutdown but **never passed to `app.Shutdown()`**. The shutdown immediately returns, and the `select` block only checks if context was cancelled—but nothing waits on the actual shutdown completion. The 30-second timeout is effectively dead code.

```go
// Current (broken):
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := app.Shutdown(); err != nil {  // ctx not passed!
    log.Error("Server forced to shutdown", logger.Error(err))
}

select {
case <-ctx.Done():          // Only fires if we called cancel() or timeout
    log.Info("Timeout of 30 seconds exceeded")
default:                     // Always immediately falls through
    log.Info("Server exited gracefully")
}
```

- **Proposed Fix:** Pass the context to `Shutdown()` and wait properly:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := app.Shutdown(ctx); err != nil {
    log.Error("Server forced to shutdown", logger.Error(err))
}

log.Info("Server exited gracefully")
```

This requires updating `App.Shutdown()` and `httpserver.Server.Shutdown()` to accept a context.

---

### ISSUE-2: Log Message Contradiction (HTTP vs HTTPS)

- **Location:** `cmd/forum/main.go`, Line 48
- **Probability:** High
- **Description:** The log message says "HTTPS access" but prints the HTTP port. Copy-paste error.

```go
log.Info(fmt.Sprintf("HTTPS access: http://localhost:%d", cfg.Server.Port))  // Wrong!
```

- **Proposed Fix:**

```go
log.Info(fmt.Sprintf("HTTP access: http://localhost:%d", cfg.Server.Port))
```

---

### ISSUE-3: Template Parsing Failure Silently Ignored

- **Location:** `cmd/forum/wire/handlers.go`, Lines 32-36
- **Probability:** Medium
- **Description:** If template parsing fails for any reason other than "files don't exist" (e.g., syntax error in template), the error is silently swallowed and `templates` is set to `nil`. Handlers will later panic or return confusing errors when they try to execute templates. This masks legitimate template bugs during development.

```go
templates, err := template.ParseGlob("templates/*.html")
if err != nil {
    // Templates are optional - API-only mode works without them
    templates = nil  // Silent failure - template syntax errors hidden!
}
```

- **Proposed Fix:** Distinguish between "directory not found" and "parse error":

```go
templates, err := template.ParseGlob("templates/*.html")
if err != nil {
    // Check if it's a genuine parse error vs missing directory
    if _, statErr := os.Stat("templates"); !os.IsNotExist(statErr) {
        // Templates directory exists, but parsing failed - this is a bug
        return nil, fmt.Errorf("template parsing failed: %w", err)
    }
    // Templates are optional - API-only mode works without them
    templates = nil
}
```

---

### ISSUE-4: `Cleanup()` Error Return Ignored in Defer

- **Location:** `cmd/forum/main.go`, Line 36
- **Probability:** Medium
- **Description:** `Cleanup()` returns an error, but when called via `defer`, the error is silently discarded. If database close fails (e.g., pending transactions), the application exits without knowing.

```go
defer app.Cleanup()  // Error return value ignored
```

- **Proposed Fix:**

```go
defer func() {
    if err := app.Cleanup(); err != nil {
        log.Error("Cleanup failed", logger.Error(err))
    }
}()
```

---

### ISSUE-5: Double Error on Database Migration Failure

- **Location:** `cmd/forum/wire/app.go`, Lines 88-91
- **Probability:** Low
- **Description:** If migration fails, the code attempts `dbConn.Close()` but ignores any close error. While the migration error is correctly returned, a close failure would be masked.

```go
if err := migrator.Migrate(cfg.Database.MigrationsDir); err != nil {
    dbConn.Close()  // Error ignored
    return nil, fmt.Errorf("failed to run migrations: %w", err)
}
```

- **Proposed Fix:** Log the close error before returning migration error:

```go
if err := migrator.Migrate(cfg.Database.MigrationsDir); err != nil {
    if closeErr := dbConn.Close(); closeErr != nil {
        // Log but don't mask the original error
        // Note: lgr should be passed to this function
    }
    return nil, fmt.Errorf("failed to run migrations: %w", err)
}
```

---

## Performance & Optimization

### PERF-1: Static File Handler Created Regardless of Usage

- **Location:** `cmd/forum/wire/app.go`, Lines 137-140
- **Probability:** Low
- **Description:** The static file handler performs an `os.Stat()` on every initialization. This is fine for startup but could be optimized to happen once and cache the result if the application supports hot-reloading.

No fix needed unless hot-reload is implemented.

---

### PERF-2: Template Parsing Could Be Parallelized

- **Location:** `cmd/forum/wire/handlers.go`, Line 32
- **Probability:** Low
- **Description:** Template parsing is synchronous during startup. For large template directories, this could slow down cold starts. Not a concern for current scale but worth noting.

No immediate fix needed.

---

## Nitpicks & Best Practices

### NIT-1: Hardcoded CORS Wildcard

- **Location:** `cmd/forum/wire/app.go`, Line 107
- **Probability:** N/A (Security Consideration)
- **Description:** CORS is configured with `["*"]` which allows any origin. This should come from configuration for production deployments.

```go
server.RegisterMiddleware(httpserver.CORS([]string{"*"}))
```

**Suggestion:** Move to configuration:

```go
server.RegisterMiddleware(httpserver.CORS(cfg.Security.AllowedOrigins))
```

---

### NIT-2: Inconsistent Package Comment Placement

- **Location:** `cmd/forum/wire/app.go`, Lines 1-4
- **Probability:** N/A (Style)
- **Description:** The package comment appears after `package wire` instead of before it, which is non-idiomatic.

```go
package wire

// Package wire handles dependency injection...
```

**Suggestion:** Move comment before package declaration:

```go
// Package wire handles dependency injection and application wiring.
// It initializes all components and returns a fully configured App instance.
package wire
```

---

### NIT-3: Variable Shadowing with `log`

- **Location:** `cmd/forum/main.go`, Lines 10, 28
- **Probability:** N/A (Readability)
- **Description:** The standard `log` package is imported (line 10), then a variable named `log` is created (line 28), shadowing the import. This works but can cause confusion.

```go
import (
    "log"  // Standard library
    ...
)

func main() {
    ...
    log := logger.New(logger.InfoLevel, os.Stdout)  // Shadows import
}
```

**Suggestion:** Use `lgr` for the variable to match other files:

```go
lgr := logger.New(logger.InfoLevel, os.Stdout)
```

---

### NIT-4: Magic String for Templates Path

- **Location:** `cmd/forum/wire/handlers.go`, Line 32
- **Probability:** N/A (Maintainability)
- **Description:** The templates path `"templates/*.html"` is hardcoded. This should come from configuration to support alternative deployments.

```go
templates, err := template.ParseGlob("templates/*.html")
```

**Suggestion:** Add to config:

```go
templates, err := template.ParseGlob(cfg.Templates.Path)
```

---

### NIT-5: Missing `doc.go` or README for Wire Package

- **Location:** `cmd/forum/wire/`
- **Probability:** N/A (Documentation)
- **Description:** While there's a `README.md`, Go convention is to have a `doc.go` file for package-level documentation that appears in godoc.

**Suggestion:** Create `doc.go`:

```go
// Package wire provides dependency injection and application wiring.
// It initializes all components and returns a fully configured App instance.
package wire
```

---

### NIT-6: Use `%s` Instead of `fmt.Sprintf` with `log.Info`

- **Location:** `cmd/forum/main.go`, Lines 46-50
- **Probability:** N/A (Style)
- **Description:** Using `fmt.Sprintf` inside log calls allocates an extra string. If the logger supports format arguments, use them directly.

```go
// Current:
log.Info(fmt.Sprintf("Forum server started on port %d", cfg.Server.Port))

// Better (if logger supports it):
log.Info("Forum server started", logger.Int("port", cfg.Server.Port))
```

---

## Summary Table

| ID      | Severity | Type            | Location          | Description                              |
| ------- | -------- | --------------- | ----------------- | ---------------------------------------- |
| ISSUE-1 | Critical | Concurrency     | main.go:60-71     | Shutdown context not used                |
| ISSUE-2 | High     | Bug             | main.go:48        | Wrong log message (HTTP/HTTPS)           |
| ISSUE-3 | Medium   | Error Handling  | handlers.go:32-36 | Template parse errors silently ignored   |
| ISSUE-4 | Medium   | Resource Mgmt   | main.go:36        | Cleanup error discarded in defer         |
| ISSUE-5 | Low      | Error Handling  | app.go:88-91      | Close error ignored on migration failure |
| NIT-1   | Low      | Security        | app.go:107        | Hardcoded CORS wildcard                  |
| NIT-2   | Low      | Style           | app.go:1-4        | Package comment placement                |
| NIT-3   | Low      | Readability     | main.go:10,28     | Variable shadows import                  |
| NIT-4   | Low      | Maintainability | handlers.go:32    | Hardcoded templates path                 |
| NIT-5   | Low      | Documentation   | wire/             | Missing doc.go                           |
| NIT-6   | Low      | Style           | main.go:46-50     | Unnecessary fmt.Sprintf                  |

---

**Overall Assessment:** The code is well-structured with good separation of concerns. The critical issues around shutdown logic should be addressed before production deployment. The template error handling (ISSUE-3) is particularly important for developer experience.
