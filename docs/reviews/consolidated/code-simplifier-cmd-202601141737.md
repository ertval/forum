# Go Code Simplifier Review

**Folder/Module:** cmd
**Date:** 2026-01-14 17:37
**Files Reviewed:**

- `cmd/forum/main.go`
- `cmd/forum/wire/app.go`
- `cmd/forum/wire/handlers.go`
- `cmd/forum/wire/repositories.go`
- `cmd/forum/wire/services.go`

---

## Summary

The `cmd` directory contains the application entry point and dependency injection wiring. Overall, the code is well-structured and follows idiomatic Go patterns. However, there are several opportunities for improvement around error handling consistency, unused context, redundant wrapping, and minor KISS violations. The wiring code is clean and modular, but some simplifications can reduce cognitive load.

---

## Findings

### 1. Unused Context Variable in Shutdown

**File:** `cmd/forum/main.go`
**Line(s):** 60-72
**Category:** KISS Violation | Dead Code
**Severity:** Medium

**Current Code:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := app.Shutdown(); err != nil {
	log.Error("Server forced to shutdown", logger.Error(err))
}

select {
case <-ctx.Done():
	log.Info("Timeout of 30 seconds exceeded")
default:
	log.Info("Server exited gracefully")
}
```

**Suggested Improvement:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Pass context to Shutdown for proper timeout handling
if err := app.Shutdown(ctx); err != nil {
	log.Error("Server forced to shutdown", logger.Error(err))
	return
}

log.Info("Server exited gracefully")
```

**Rationale:** The context is created but never used - `app.Shutdown()` doesn't receive it. The `select` block is misleading because `ctx.Done()` will never be triggered by the shutdown operation. Either pass the context to `Shutdown(ctx)` for proper timeout semantics, or remove the context entirely. The current code gives false confidence about timeout handling.

---

### 2. Package Comment Misplaced

**File:** `cmd/forum/wire/app.go`
**Line(s):** 1-4
**Category:** Idiomatic Go
**Severity:** Low

**Current Code:**

```go
package wire

// Package wire handles dependency injection and application wiring.
// It initializes all components and returns a fully configured App instance.

import (
```

**Suggested Improvement:**

```go
// Package wire handles dependency injection and application wiring.
// It initializes all components and returns a fully configured App instance.
package wire

import (
```

**Rationale:** Go convention places the package doc comment immediately before the `package` declaration, not after. This ensures `go doc` and IDE tools display the documentation correctly.

---

### 3. Redundant Error Wrapping with Duplicate Context

**File:** `cmd/forum/wire/app.go`
**Line(s):** 81-84
**Category:** Idiomatic Go | Error Handling
**Severity:** Low

**Current Code:**

```go
dbConn, err := database.NewConnection(cfg.Database.Path)
if err != nil {
	return nil, fmt.Errorf("failed to connect to database: %w", err)
}
```

**Suggested Improvement:**

```go
dbConn, err := database.NewConnection(cfg.Database.Path)
if err != nil {
	return nil, fmt.Errorf("database connection: %w", err)
}
```

**Rationale:** The phrase "failed to" is redundant in error messages - errors inherently represent failures. A shorter message is clearer and follows Go conventions where error messages are lowercase and concise. This pattern appears throughout the file.

---

### 4. Ignored Template Parse Error Silently

**File:** `cmd/forum/wire/handlers.go`
**Line(s):** 32-36
**Category:** Error Handling
**Severity:** Medium

**Current Code:**

```go
templates, err := template.ParseGlob("templates/*.html")
if err != nil {
	// Templates are optional - API-only mode works without them
	templates = nil
}
```

**Suggested Improvement:**

```go
templates, err := template.ParseGlob("templates/*.html")
if err != nil {
	// Templates are optional for API-only mode.
	// In production with templates, a parse error is a bug.
	if !os.IsNotExist(err) {
		// Log actual parse errors vs. missing directory
		// Consider: log.Warn("template parse error: %v", err)
	}
	templates = nil
}
```

**Rationale:** Currently, both "templates directory doesn't exist" and "template has a syntax error" are treated identically. In production, a template parse error is a critical bug that should be logged. Differentiating between "no templates" and "broken templates" helps with debugging.

---

### 5. Inconsistent File Header Comments

**File:** `cmd/forum/wire/app.go`
**Line(s):** 1
**Category:** Idiomatic Go | Consistency
**Severity:** Low

**Current Code:**

```go
package wire

// Package wire handles...
```

Compare with `handlers.go`:

```go
// INPUT ADAPTERS - HTTP Handler Initialization
package wire
```

And `repositories.go`:

```go
// OUTPUT ADAPTERS - Repository Initialization
package wire
```

**Suggested Improvement:**
Standardize to one pattern. Recommended approach for consistency:

```go
// Package wire handles dependency injection and application wiring.
// File: app.go - Core application initialization
package wire
```

**Rationale:** The GEMINI.md specifies file headers like `// INPUT PORT - Service Interface`. Within the same package, headers should follow a consistent pattern. Either all files use the `// LAYER - Description` format, or all use standard package doc comments.

---

### 6. Defensive Check Uses os.Stat Instead of Embedded FS Pattern

**File:** `cmd/forum/wire/app.go`
**Line(s):** 137-140
**Category:** KISS Violation | Robustness
**Severity:** Low

**Current Code:**

```go
lgr.Info("Checking for static directory")
if _, err := os.Stat("./static"); err == nil {
	lgr.Info("Registering static file handler")
	server.Router().Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}
```

**Suggested Improvement:**

```go
// Static files are optional - skip if directory doesn't exist
if info, err := os.Stat("./static"); err == nil && info.IsDir() {
	lgr.Info("Registering static file handler")
	server.Router().Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}
```

**Rationale:** The extra log line "Checking for static directory" is unnecessary - logging should focus on outcomes, not intentions. Also, checking `IsDir()` ensures we don't try to serve a file named "static" as a directory.

---

### 7. Service Container Methods Could Use Embedding

**File:** `cmd/forum/wire/services.go`
**Line(s):** 45-57
**Category:** KISS Violation | Architecture
**Severity:** Low

**Current Code:**

```go
func (sc *ServiceContainer) Auth() authPorts.AuthService                   { return sc.auth }
func (sc *ServiceContainer) AuthMiddleware() authPorts.AuthMiddleware      { return sc.authMiddleware }
func (sc *ServiceContainer) User() userPorts.UserService                   { return sc.user }
func (sc *ServiceContainer) Post() postPorts.PostService                   { return sc.post }
func (sc *ServiceContainer) Category() postPorts.CategoryService           { return sc.category }
func (sc *ServiceContainer) Filter() postPorts.FilterService               { return sc.filter }
func (sc *ServiceContainer) Comment() commentPorts.CommentService          { return sc.comment }
func (sc *ServiceContainer) Reaction() reactionPorts.ReactionService       { return sc.reaction }
func (sc *ServiceContainer) Moderation() moderationPorts.ModerationService { return sc.moderation }
func (sc *ServiceContainer) Notification() notificationPorts.NotificationService {
	return sc.notification
}
func (sc *ServiceContainer) Logger() *logger.Logger { return sc.logger }
```

**Suggested Improvement:**
Keep as-is, but add a clarifying comment:

```go
// Service accessor methods satisfy handler-specific interfaces.
// Each handler declares only the services it needs, e.g.:
//   type ServiceContainer interface { Auth() authPorts.AuthService }
// This pattern enables compile-time dependency verification.
func (sc *ServiceContainer) Auth() authPorts.AuthService { return sc.auth }
// ... rest unchanged
```

**Rationale:** While this seems verbose, it's the correct pattern per GEMINI.md. The accessor methods enable interface segregation - each handler defines a minimal interface. Adding a comment clarifies the design rationale for future maintainers.

---

### 8. Layered Service Initialization Comments Could Be More Precise

**File:** `cmd/forum/wire/services.go`
**Line(s):** 61-78
**Category:** Idiomatic Go | Documentation
**Severity:** Low

**Current Code:**

```go
// Layer 1: Services with no dependencies
userService := userApp.NewService(repos.User)

// Initialize image handler for post uploads (config-driven)
imageHandler := upload.NewImageHandler(cfg.Upload.UploadDir, cfg.Upload.MaxSize)
categoryService := postApp.NewCategoryService(repos.Category)
```

**Suggested Improvement:**

```go
// Layer 1: Foundation services (no inter-service dependencies)
userService := userApp.NewService(repos.User)
categoryService := postApp.NewCategoryService(repos.Category)
filterService := postApp.NewFilterService()

// Layer 1b: Infrastructure adapters
imageHandler := upload.NewImageHandler(cfg.Upload.UploadDir, cfg.Upload.MaxSize)

// Layer 2: Domain services (depend on Layer 1)
reactionService := reactionApp.NewService(repos.Reaction, repos.Post, repos.Comment, userService)
moderationService := moderationApp.NewService(repos.Moderation)
notificationService := notificationApp.NewService(repos.Notification)
authService := authApp.NewService(repos.Session, userService, cfg.Session.Duration)
postService := postApp.NewService(repos.Post, repos.Category, userService, imageHandler, cfg.Upload.MaxSize)
commentService := commentApp.NewService(repos.Comment, postService, userService)

// Layer 3: Cross-cutting adapters
authMiddleware := authAdapters.NewAuthMiddleware(authService, userService)
```

**Rationale:** The current Layer 1 comment says "no dependencies" but `imageHandler` is mixed in. Grouping services by their actual dependency relationships makes the initialization order clearer and helps prevent cyclic dependency issues during refactoring.

---

### 9. main.go Log Message Typo

**File:** `cmd/forum/main.go`
**Line(s):** 48
**Category:** Typo
**Severity:** Low

**Current Code:**

```go
log.Info(fmt.Sprintf("HTTPS access: http://localhost:%d", cfg.Server.Port))
```

**Suggested Improvement:**

```go
log.Info(fmt.Sprintf("HTTP access: http://localhost:%d", cfg.Server.Port))
```

**Rationale:** The message says "HTTPS access" but shows an HTTP URL. This is confusing for operators.

---

### 10. Cleanup Method Should Not Return Error From Deferred Call

**File:** `cmd/forum/wire/app.go`
**Line(s):** 26-33
**Category:** Idiomatic Go | Error Handling
**Severity:** Medium

**Current Code:**

```go
func (a *App) Cleanup() error {
	a.Logger.Info("Cleaning up application resources")
	if err := a.Database.Close(); err != nil {
		a.Logger.Error("Failed to close database connection", logger.Error(err))
		return err
	}
	return nil
}
```

**Suggested Improvement:**

```go
// Cleanup releases application resources. Errors are logged but not returned
// since callers typically cannot meaningfully handle cleanup failures.
func (a *App) Cleanup() {
	a.Logger.Info("Cleaning up application resources")
	if err := a.Database.Close(); err != nil {
		a.Logger.Error("Failed to close database connection", logger.Error(err))
	}
}
```

**Rationale:** In `main.go`, `Cleanup()` is called via `defer app.Cleanup()` - the return value is ignored. Returning an error that's never checked is misleading. Either the caller should handle the error, or the method should be `void`. Since cleanup errors rarely have meaningful recovery paths, logging is sufficient.

---

## Action Items

- [ ] Fix unused context in shutdown - either pass to `Shutdown(ctx)` or remove context creation
- [ ] Move package doc comment before `package wire` in `app.go`
- [ ] Simplify error messages (remove "failed to" prefix) for consistency
- [ ] Differentiate template parse errors from missing template directory
- [ ] Standardize file header comments across wire package files
- [ ] Add `IsDir()` check for static directory verification
- [ ] Add clarifying comment for ServiceContainer accessor methods
- [ ] Reorganize service initialization with clearer layer groupings
- [ ] Fix "HTTPS access" typo to "HTTP access"
- [ ] Change `Cleanup()` to return void since error is never handled

---

## Notes

1. **Overall Quality**: The `cmd` package follows clean architecture principles well. The separation of concerns (repositories, services, handlers) is clear and the wiring is explicit.

2. **Testing Considerations**: The `wire` package could benefit from constructor injection for the template path to improve testability. Currently, `template.ParseGlob("templates/*.html")` uses a hardcoded path.

3. **Potential Future Improvement**: Consider using the functional options pattern for `InitializeApp` if more configuration options are needed:

   ```go
   type Option func(*App)
   func WithLogger(l *logger.Logger) Option { ... }
   func InitializeApp(cfg *config.Config, opts ...Option) (*App, error)
   ```

4. **GEMINI.md Compliance**: The code follows the specified patterns for DI wiring and file organization. The `ServiceContainer` pattern correctly implements interface segregation.
