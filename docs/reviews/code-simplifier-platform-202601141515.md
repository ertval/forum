# Go Code Simplifier Review

**Folder/Module:** platform  
**Date:** 2026-01-14 15:15  
**Files Reviewed:**

- `config/config.go`
- `config/env_parser.go`
- `database/connection.go`
- `database/migrator.go`
- `database/transaction.go`
- `errors/errors.go`
- `health/checker.go`
- `httpserver/server.go`
- `httpserver/middleware.go`
- `httpserver/security_headers.go`
- `httpserver/health.go`
- `httpserver/tls.go`
- `logger/logger.go`
- `templates/validator.go`
- `upload/image.go`
- `validator/validator.go`

---

## Summary

The platform module is well-structured and follows many Go best practices. It provides foundational infrastructure for the forum application including configuration, database, logging, HTTP server, and validation. The code is generally clean and well-documented.

**Key Strengths:**

- Excellent documentation with package and function comments
- Good separation of concerns across packages
- Proper error handling with `%w` wrapping
- Table-driven tests in several places
- Good use of dependency injection patterns

**Areas for Improvement:**

- Several custom utility functions duplicate stdlib functionality
- Some redundant code patterns that can be simplified
- A few functions have TODO comments indicating incomplete implementation
- Minor KISS violations in complex conditionals
- Regex compilation should be moved to package level for performance

---

## Findings

### 1. Custom `indexOf` Function Duplicates `strings.IndexByte`

**File:** `database/connection.go`  
**Line(s):** 95-103  
**Category:** Idiomatic Go | KISS Violation  
**Severity:** Low

**Current Code:**

```go
// indexOf returns the index of ch in s or -1 if not present.
func indexOf(s string, ch byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ch {
			return i
		}
	}
	return -1
}
```

**Suggested Improvement:**

```go
// Use strings.IndexByte from the standard library instead:
// import "strings"
//
// In NewConnection:
// if idx := strings.IndexByte(dsn, '?'); idx != -1 {
```

**Rationale:** Go's standard library provides `strings.IndexByte(s string, c byte) int` which does exactly this. "Leverage the standard library" is a core Go idiom. Removing custom implementations reduces maintenance burden and potential bugs.

---

### 2. Custom `itoa` Function Duplicates `strconv.Itoa`

**File:** `httpserver/security_headers.go`  
**Line(s):** 113-140  
**Category:** Idiomatic Go | KISS Violation  
**Severity:** Low

**Current Code:**

```go
// itoa converts an integer to a string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	// Maximum int64 is 19 digits
	buf := make([]byte, 20)
	pos := len(buf)

	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}

	if negative {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}
```

**Suggested Improvement:**

```go
// Replace with strconv.Itoa:
import "strconv"

// In SecurityHeaders:
hstsValue := "max-age=" + strconv.Itoa(cfg.HSTSMaxAge)
```

**Rationale:** The comment "without importing strconv" suggests avoiding dependencies, but `strconv` is part of Go's standard library and has zero cost. The custom implementation adds 27 lines of code that must be maintained and tested. This violates "a little copying is better than a little dependency" principle—the stdlib is not a dependency.

---

### 3. Redundant Nil Check in `getEnvStringSlice`

**File:** `config/env_parser.go`  
**Line(s):** 46-59  
**Category:** KISS Violation  
**Severity:** Low

**Current Code:**

```go
func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		if value != "" {  // <-- Redundant check
			// Split by comma and trim spaces
			parts := strings.Split(value, ",")
			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}
			return parts
		}
	}
	return defaultValue
}
```

**Suggested Improvement:**

```go
func getEnvStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}
```

**Rationale:** The inner `if value != ""` is redundant since it's already checked in the outer condition. Early return pattern improves readability.

---

### 4. Regex Compilation in `Sanitize` Should Be at Package Level

**File:** `validator/validator.go`  
**Line(s):** 143-183  
**Category:** Performance  
**Severity:** Medium

**Current Code:**

```go
func Sanitize(input string) string {
	if input == "" {
		return ""
	}

	s := html.UnescapeString(input)

	// Remove script blocks and style blocks (case-insensitive)
	reScript := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	s = reScript.ReplaceAllString(s, "")
	reStyle := regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
	s = reStyle.ReplaceAllString(s, "")

	// Strip remaining tags
	reTags := regexp.MustCompile(`<[^>]+>`)
	s = reTags.ReplaceAllString(s, "")

	// ... more code ...

	// Collapse all whitespace sequences to a single space
	reSpace := regexp.MustCompile(`\s+`)
	s = reSpace.ReplaceAllString(s, " ")
	// ...
}
```

**Suggested Improvement:**

```go
// Package-level compiled regexes (compiled once at init)
var (
	reScript = regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	reStyle  = regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
	reTags   = regexp.MustCompile(`<[^>]+>`)
	reSpace  = regexp.MustCompile(`\s+`)
)

func Sanitize(input string) string {
	if input == "" {
		return ""
	}

	s := html.UnescapeString(input)
	s = reScript.ReplaceAllString(s, "")
	s = reStyle.ReplaceAllString(s, "")
	s = reTags.ReplaceAllString(s, "")

	// ... control char removal ...

	s = reSpace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}
```

**Rationale:** `regexp.MustCompile` is expensive. Compiling the regex on every call to `Sanitize` wastes CPU cycles. Move to package-level variables—they're compiled once at startup and safely reused (regexes are thread-safe for matching).

---

### 5. Same Regex Compilation Issue in `Username` and `Email` Validation

**File:** `validator/validator.go`  
**Line(s):** 68-106  
**Category:** Performance  
**Severity:** Medium

**Current Code:**

```go
func (v *Validator) Email(field, value string) {
	value = strings.ToLower(strings.TrimSpace(Sanitize(value)))
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	// ...
}

func (v *Validator) Username(field, value string) {
	// ...
	namePartRegex := regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
	// ...
}
```

**Suggested Improvement:**

```go
var (
	emailRegex    = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	namePartRegex = regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
)

func (v *Validator) Email(field, value string) {
	value = strings.ToLower(strings.TrimSpace(Sanitize(value)))
	if !emailRegex.MatchString(value) {
		v.AddError(field, "Must be a valid email address")
	}
}
```

**Rationale:** Same as above—compile once, reuse forever.

---

### 6. Complex Environment Validation Could Use Slice Contains

**File:** `config/config.go`  
**Line(s):** 189-191  
**Category:** Idiomatic Go  
**Severity:** Low

**Current Code:**

```go
if c.Server.Environment != "development" && c.Server.Environment != "staging" && c.Server.Environment != "production" {
	return fmt.Errorf("invalid environment: %s", c.Server.Environment)
}
```

**Suggested Improvement:**

```go
validEnvironments := []string{"development", "staging", "production"}
if !slices.Contains(validEnvironments, c.Server.Environment) {
	return fmt.Errorf("invalid environment: %s (valid: %v)", c.Server.Environment, validEnvironments)
}
```

**Rationale:** Using `slices.Contains` (available since Go 1.21) is more readable and makes the valid values self-documenting. The error message can also be improved to show valid options.

---

### 7. `WriteErrorJSON` Creates Logger on Every Call

**File:** `errors/errors.go`  
**Line(s):** 124-154  
**Category:** Performance | Architecture  
**Severity:** Medium

**Current Code:**

```go
func WriteErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	// Log error for debugging (human-readable to stderr)
	cfg := &logger.Config{
		TimePrecision: logger.TimePrecisionSeconds,
		AllowedFields: []string{"status", "error"},
		MaxLineWidth:  200,
		Colorize:      true,
	}
	lgr := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
	lgr.Error("http.error",
		logger.Int("status", status),
		logger.String("error", message))

	// Write JSON response
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		lgr.Error("failed to encode error response",
			logger.Error(err))
	}
}
```

**Suggested Improvement:**

```go
var errorLogger = logger.NewWithConfig(logger.ErrorLevel, os.Stderr, &logger.Config{
	TimePrecision: logger.TimePrecisionSeconds,
	AllowedFields: []string{"status", "error"},
	MaxLineWidth:  200,
	Colorize:      true,
})

func WriteErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	errorLogger.Error("http.error",
		logger.Int("status", status),
		logger.String("error", message))

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		errorLogger.Error("failed to encode error response",
			logger.Error(err))
	}
}
```

**Rationale:** Creating a new logger with configuration on every error response is wasteful. Use a package-level logger initialized once. Alternatively, consider accepting a logger as a parameter for better dependency injection.

---

### 8. Template Re-parsing in `HealthPage`

**File:** `httpserver/health.go`  
**Line(s):** 76-82  
**Category:** Performance  
**Severity:** Medium

**Current Code:**

```go
func HealthPage(cfg HealthPageConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ...
		var tmpl *template.Template
		var err error
		tmpl, err = template.ParseFiles("templates/base.html", "templates/health.html")
		if err != nil {
			http.Error(w, "Could not parse templates", http.StatusInternalServerError)
			return
		}
		// ...
	}
}
```

**Suggested Improvement:**

```go
func HealthPage(cfg HealthPageConfig) http.HandlerFunc {
	// Parse templates once at handler creation time
	tmpl, err := template.ParseFiles("templates/base.html", "templates/health.html")
	if err != nil {
		// Log and panic during startup if templates are missing
		panic(fmt.Sprintf("failed to parse health templates: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Use the pre-parsed template
		results := cfg.Checker.Check(r.Context())
		// ...
		err := tmpl.ExecuteTemplate(w, "base", data)
		if err != nil {
			http.Error(w, "Could not execute template", http.StatusInternalServerError)
		}
	}
}
```

**Rationale:** Parsing templates on every request is expensive and unnecessary. Parse once when the handler is created. If templates are missing, fail fast at startup rather than on every request.

---

### 9. Unused Function `getRequiredTemplates`

**File:** `templates/validator.go`  
**Line(s):** 115-122  
**Category:** Dead Code  
**Severity:** Low

**Current Code:**

```go
// getRequiredTemplates returns the list of templates required for the application.
func getRequiredTemplates() []string {
	return []string{
		"base",    // Base layout template
		"content", // Content block (defined in various templates)
		// Note: Other templates are checked dynamically
	}
}
```

**Suggested Improvement:**

```go
// Remove this function if unused, or export it if it should be part of the public API:
// RequiredTemplates returns the list of templates required for the application.
func RequiredTemplates() []string {
	return []string{
		"base",
		"content",
	}
}
```

**Rationale:** This unexported function appears to be unused. Either remove it (dead code) or export it if it has a purpose. If it's intended for future use, add a TODO comment explaining when it will be needed.

---

### 10. TODO Comments Indicate Incomplete Implementation

**File:** `database/migrator.go`  
**Line(s):** 137-148  
**Category:** TDD | Architecture  
**Severity:** Medium

**Current Code:**

```go
// Rollback rolls back the last migration.
// TODO: Implement rollback logic.
func (m *Migrator) Rollback() error {
	// Implementation placeholder
	return nil
}

// Version returns the current database schema version.
// TODO: Implement version tracking.
func (m *Migrator) Version() (int, error) {
	// Implementation placeholder
	return 0, nil
}
```

**Suggested Improvement:**

```go
// Rollback rolls back the last migration.
// Returns ErrNotImplemented until rollback support is added.
func (m *Migrator) Rollback() error {
	return errors.New("rollback not yet implemented")
}

// Version returns the current database schema version.
// Returns the highest applied migration version, or 0 if none applied.
func (m *Migrator) Version() (int, error) {
	var version int
	err := m.conn.DB().QueryRow(
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations",
	).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}
	return version, nil
}
```

**Rationale:** Functions with TODO placeholders that silently return success can mask issues. Either implement them properly or return an error indicating the feature is not implemented. The `Version()` function is straightforward to implement using the existing `schema_migrations` table.

---

### 11. Unbounded Goroutine in Rate Limiter Cleanup

**File:** `httpserver/middleware.go`  
**Line(s):** 179-186  
**Category:** Concurrency | Resource Management  
**Severity:** Low

**Current Code:**

```go
func RateLimit(requests int, windowSeconds int) Middleware {
	limiter := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requests,
		window:   time.Duration(windowSeconds) * time.Second,
	}

	// Cleanup goroutine to prevent memory leak
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()
	// ...
}
```

**Suggested Improvement:**

```go
func RateLimitWithContext(ctx context.Context, requests int, windowSeconds int) Middleware {
	limiter := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requests,
		window:   time.Duration(windowSeconds) * time.Second,
	}

	// Cleanup goroutine with graceful shutdown support
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				limiter.cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
	// ...
}
```

**Rationale:** The cleanup goroutine has no exit condition and will run until the process terminates. While this is acceptable for long-running servers, it prevents clean testing and doesn't follow the pattern of "explicitly define exit conditions for goroutines." Consider accepting a context for graceful shutdown.

---

### 12. Magic Number for Graceful Shutdown Timeout

**File:** `httpserver/server.go`  
**Line(s):** 124-126  
**Category:** KISS | Configuration  
**Severity:** Low

**Current Code:**

```go
func (s *Server) Shutdown() error {
	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// ...
}
```

**Suggested Improvement:**

```go
const defaultShutdownTimeout = 30 * time.Second

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()
	// ...
}

// Or even better, make it configurable:
type ServerConfig struct {
	// ... existing fields ...
	ShutdownTimeout time.Duration
}
```

**Rationale:** Magic numbers should be named constants for clarity. Better yet, make this configurable through the config system like other timeouts.

---

### 13. Simplify HTTP Status Mapping with Map Lookup

**File:** `errors/errors.go`  
**Line(s):** 79-99  
**Category:** Idiomatic Go | KISS  
**Severity:** Low

**Current Code:**

```go
func HTTPStatus(err error) int {
	if e, ok := err.(*Error); ok {
		switch e.Code {
		case ErrCodeValidation, ErrCodeBadRequest:
			return http.StatusBadRequest
		case ErrCodeUnauthorized:
			return http.StatusUnauthorized
		case ErrCodeForbidden:
			return http.StatusForbidden
		case ErrCodeNotFound:
			return http.StatusNotFound
		case ErrCodeConflict:
			return http.StatusConflict
		case ErrCodeTooManyRequests:
			return http.StatusTooManyRequests
		case ErrCodeInternal:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}
```

**Suggested Improvement:**

```go
var codeToStatus = map[string]int{
	ErrCodeValidation:      http.StatusBadRequest,
	ErrCodeBadRequest:      http.StatusBadRequest,
	ErrCodeUnauthorized:    http.StatusUnauthorized,
	ErrCodeForbidden:       http.StatusForbidden,
	ErrCodeNotFound:        http.StatusNotFound,
	ErrCodeConflict:        http.StatusConflict,
	ErrCodeTooManyRequests: http.StatusTooManyRequests,
	ErrCodeInternal:        http.StatusInternalServerError,
}

func HTTPStatus(err error) int {
	if e, ok := err.(*Error); ok {
		if status, ok := codeToStatus[e.Code]; ok {
			return status
		}
	}
	return http.StatusInternalServerError
}
```

**Rationale:** Map lookup is cleaner when mapping from one set of values to another. Adding new error codes becomes a single line addition to the map instead of modifying the switch statement.

---

### 14. Health Checker Creates Hardcoded Test Values

**File:** `health/checker.go`  
**Line(s):** 136-143  
**Category:** KISS | Flexibility  
**Severity:** Low

**Current Code:**

```go
if strings.Contains(path, "{") && strings.Contains(path, "}") {
	// Handle common parameter names in routes
	testPath = strings.Replace(testPath, "{id}", "1", -1)
	testPath = strings.Replace(testPath, "{postId}", "1", -1)
	testPath = strings.Replace(testPath, "{targetType}", "post", -1)
	testPath = strings.Replace(testPath, "{targetId}", "1", -1)
	// Remove any remaining brackets that weren't matched by the specific replacements
	testPath = strings.ReplaceAll(testPath, "{", "1") // fallback for other parameter names
	testPath = strings.ReplaceAll(testPath, "}", "")
}
```

**Suggested Improvement:**

```go
// Use a regex to replace all path parameters with a test value
var pathParamRegex = regexp.MustCompile(`\{[^}]+\}`)

func replacePathParams(path string) string {
	return pathParamRegex.ReplaceAllString(path, "test-value")
}

// In isRouteRegistered:
if strings.Contains(path, "{") {
	testPath = replacePathParams(path)
}
```

**Rationale:** The explicit replacement of each parameter name is brittle and requires updates when new parameter names are added. A regex-based approach handles any parameter name automatically.

---

## Action Items

- [ ] Replace custom `indexOf` with `strings.IndexByte` in `database/connection.go`
- [ ] Replace custom `itoa` with `strconv.Itoa` in `httpserver/security_headers.go`
- [ ] Remove redundant nil check in `getEnvStringSlice` in `config/env_parser.go`
- [ ] Move regex compilations to package level in `validator/validator.go`
- [ ] Move error logger to package level in `errors/errors.go`
- [ ] Parse templates once in `HealthPage` handler creation in `httpserver/health.go`
- [ ] Implement or properly error `Migrator.Version()` and `Migrator.Rollback()`
- [ ] Add context support to rate limiter cleanup goroutine
- [ ] Extract magic numbers to named constants (shutdown timeout, etc.)
- [ ] Consider using `slices.Contains` for environment validation
- [ ] Review and remove/export `getRequiredTemplates` function
- [ ] Consider map-based HTTP status mapping for extensibility
- [ ] Simplify path parameter replacement in health checker with regex

---

## Notes

1. **Overall Code Quality**: The platform module demonstrates good Go practices with comprehensive documentation, proper error wrapping, and clean separation of concerns.

2. **Testing Coverage**: Tests exist for most packages with table-driven patterns. Consider adding benchmarks for performance-critical code like `Sanitize`.

3. **Architecture**: The module correctly avoids importing business modules, maintaining the hexagonal architecture boundary. The note in `middleware.go` about auth middleware placement is a good example of architectural documentation.

4. **TODO Items**: Several functions have TODO placeholders. These should be tracked in an issue system or implementation roadmap.

5. **Go Version**: Some suggestions use `slices.Contains` which requires Go 1.21+. Verify the project's minimum Go version before applying.

6. **Performance Impact**: The regex pre-compilation changes in `validator.go` will have measurable performance improvements when validating many inputs (e.g., form submissions, API requests).
