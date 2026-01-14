# Code Review: Platform Package

**Date:** 2026-01-14 15:11  
**Reviewer:** Principal Software Engineer / AI Audit  
**Scope:** `internal/platform/` (config, database, errors, health, httpserver, logger, templates, upload, validator)

---

## Executive Summary

The platform layer is **well-structured and generally follows idiomatic Go patterns** with good separation of concerns. However, several issues exist: a **goroutine leak** in the rate limiter that can never be stopped, **silent error handling** in multiple places, a few **concurrency edge cases**, and some **performance inefficiencies** in the validator's regex usage. Most critical is the orphaned cleanup goroutine and the `Sanitize()` function compiling regexes on every call.

---

## Critical Issues (Must Fix)

### ISSUE-1: Goroutine Leak in RateLimit Middleware

- **Location:** `httpserver/middleware.go`, Lines 179-186
- **Probability:** **High** (100% reproduction)
- **Description:** The `RateLimit()` middleware spawns a background goroutine for cleanup that runs indefinitely with `for range ticker.C`. This goroutine can never be stopped — there is no context, channel, or shutdown mechanism. Each time `RateLimit()` is called (e.g., in tests or server restarts within the same process), a new goroutine is spawned and the old ones remain alive, leaking memory and goroutines.

```go
// Current code (problematic)
go func() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        limiter.cleanup()
    }
}()
```

- **Proposed Fix:** Add a context or done channel to allow graceful shutdown:

```go
type rateLimiter struct {
    requests map[string][]time.Time
    mu       sync.Mutex
    limit    int
    window   time.Duration
    done     chan struct{} // ADD: shutdown signal
}

func RateLimit(requests int, windowSeconds int) (Middleware, func()) {
    limiter := &rateLimiter{
        requests: make(map[string][]time.Time),
        limit:    requests,
        window:   time.Duration(windowSeconds) * time.Second,
        done:     make(chan struct{}),
    }

    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                limiter.cleanup()
            case <-limiter.done:
                return
            }
        }
    }()

    stop := func() { close(limiter.done) }

    return func(next http.Handler) http.Handler {
        // ... middleware logic
    }, stop
}
```

---

### ISSUE-2: Regex Compilation on Every Sanitize() Call (Performance & Memory)

- **Location:** `validator/validator.go`, Lines 143-182 (`Sanitize` function)
- **Probability:** **High** (affects every request with user input)
- **Description:** The `Sanitize()` function compiles **4 regular expressions on every call**:

  - `(?i)<script[^>]*>[\s\S]*?</script>`
  - `(?i)<style[^>]*>[\s\S]*?</style>`
  - `<[^>]+>`
  - `\s+`

  This is extremely inefficient for a function that is called on virtually every user input. Each regex compilation allocates memory and consumes CPU. Under high load, this becomes a significant bottleneck.

```go
// Current problematic code (inside function):
reScript := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
reStyle := regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
reTags := regexp.MustCompile(`<[^>]+>`)
reSpace := regexp.MustCompile(`\s+`)
```

- **Proposed Fix:** Compile regexes once at package initialization:

```go
// Package-level compiled regexes (initialized once)
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
    // ... rest of function
}
```

---

### ISSUE-3: Silent Error Swallowing in WriteErrorJSON

- **Location:** `errors/errors.go`, Lines 149-153
- **Probability:** **Medium**
- **Description:** When `json.NewEncoder(w).Encode()` fails, the error is only logged but not handled. The client may receive a partial or corrupted response. More critically, the logger is created fresh on every error, which is inefficient.

```go
// Current code
if err := json.NewEncoder(w).Encode(errResp); err != nil {
    // If JSON encoding fails, log but don't expose to client
    lgr.Error("failed to encode error response",
        logger.Error(err))
}
```

- **Impact:** If encoding fails (rare but possible with io issues), the HTTP connection may be left in an undefined state.

- **Proposed Fix:** Consider using a simpler fallback:

```go
if err := json.NewEncoder(w).Encode(errResp); err != nil {
    // Fallback to plain text if JSON fails
    http.Error(w, message, status)
}
```

---

### ISSUE-4: Potential Nil Pointer Dereference in Health Checker

- **Location:** `health/checker.go`, Lines 23-40
- **Probability:** **Low** (depends on initialization order)
- **Description:** The `Checker.Check()` method does not validate that `c.db` or `c.router` are non-nil before use. If `NewChecker` is called with nil arguments, the code will panic.

```go
func (c *Checker) Check(ctx context.Context) map[string]string {
    results := make(map[string]string)

    ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
    defer cancel()

    if err := c.db.PingContext(ctx); err != nil { // PANIC if c.db is nil
        results["database"] = "down"
    }
    // ...
}
```

- **Proposed Fix:** Add nil checks or validate in constructor:

```go
func NewChecker(db *sql.DB, router *http.ServeMux) (*Checker, error) {
    if db == nil {
        return nil, errors.New("database connection cannot be nil")
    }
    if router == nil {
        return nil, errors.New("router cannot be nil")
    }
    return &Checker{db: db, router: router}, nil
}
```

---

### ISSUE-5: Template Parsing on Every HealthPage Request

- **Location:** `httpserver/health.go`, Lines 76-82
- **Probability:** **High** (affects every health page render)
- **Description:** The `HealthPage` handler parses template files on every request (`template.ParseFiles(...)`). This is inefficient and can cause file system I/O on every hit. Templates should be parsed once at startup.

```go
// Current problematic code (inside handler):
tmpl, err = template.ParseFiles("templates/base.html", "templates/health.html")
if err != nil {
    http.Error(w, "Could not parse templates", http.StatusInternalServerError)
    return
}
```

- **Proposed Fix:** Pass pre-parsed templates via config or parse once:

```go
// Option 1: Use the templates passed in config
func HealthPage(cfg HealthPageConfig) http.HandlerFunc {
    // Parse templates ONCE at handler creation
    tmpl, err := template.ParseFiles("templates/base.html", "templates/health.html")
    if err != nil {
        // Log error and return handler that always errors
        return func(w http.ResponseWriter, r *http.Request) {
            http.Error(w, "Template initialization failed", http.StatusInternalServerError)
        }
    }

    return func(w http.ResponseWriter, r *http.Request) {
        // Use pre-parsed tmpl
        // ...
    }
}
```

---

## Performance & Optimization

### PERF-1: Email Regex Compiled Per Validation Call

- **Location:** `validator/validator.go`, Line 71
- **Description:** Similar to ISSUE-2, the email regex is compiled inside `Email()`:

```go
func (v *Validator) Email(field, value string) {
    // Compiled on EVERY call
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
    // ...
}
```

- **Optimized Code:**

```go
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

func (v *Validator) Email(field, value string) {
    value = strings.ToLower(strings.TrimSpace(Sanitize(value)))
    if !emailRegex.MatchString(value) {
        v.AddError(field, "Must be a valid email address")
    }
}
```

---

### PERF-2: Username Regex Compiled Per Validation Call

- **Location:** `validator/validator.go`, Line 99
- **Description:** Same pattern as email validation.

```go
namePartRegex := regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
```

- **Optimized Code:** Move to package level.

---

### PERF-3: Logger Creates New Config Object on Every Error

- **Location:** `errors/errors.go`, Lines 137-143
- **Description:** `WriteErrorJSON` creates a new `logger.Config` and `logger.Logger` on every call. This is wasteful.

```go
cfg := &logger.Config{
    TimePrecision: logger.TimePrecisionSeconds,
    AllowedFields: []string{"status", "error"},
    MaxLineWidth:  200,
    Colorize:      true,
}
lgr := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
```

- **Optimized Code:** Use a package-level singleton logger or accept logger as parameter:

```go
var errLogger = logger.NewWithConfig(logger.ErrorLevel, os.Stderr, &logger.Config{
    TimePrecision: logger.TimePrecisionSeconds,
    AllowedFields: []string{"status", "error"},
    MaxLineWidth:  200,
    Colorize:      true,
})

func WriteErrorJSON(w http.ResponseWriter, status int, message string) {
    // ... use errLogger instead of creating new one
}
```

---

### PERF-4: Rate Limiter O(n) Scan on Every Request

- **Location:** `httpserver/middleware.go`, Lines 216-231
- **Description:** The `allow()` method iterates through all timestamps for an IP on every request. For IPs with many requests, this is O(n) per request. The cleanup also does O(n\*m) where n is IPs and m is timestamps per IP.

- **Optimized Code:** Consider using a sliding window counter or ring buffer for O(1) operations:

```go
// Alternative: Use atomic counter with periodic reset
type rateLimiter struct {
    counts sync.Map // map[string]*ipCounter
    limit  int
    window time.Duration
}

type ipCounter struct {
    count     int64
    windowEnd time.Time
    mu        sync.Mutex
}
```

---

## Nitpicks & Best Practices

### NIT-1: Redundant Check in getEnvStringSlice

- **Location:** `config/env_parser.go`, Lines 46-58
- **Description:** The function checks `if value != ""` twice unnecessarily:

```go
func getEnvStringSlice(key string, defaultValue []string) []string {
    if value := os.Getenv(key); value != "" {  // First check
        if value != "" {                        // Redundant second check
            parts := strings.Split(value, ",")
            // ...
        }
    }
    return defaultValue
}
```

- **Fix:** Remove the inner check.

---

### NIT-2: Custom `indexOf` Function When stdlib Exists

- **Location:** `database/connection.go`, Lines 95-103
- **Description:** A custom `indexOf` function is implemented when `strings.IndexByte` exists in the standard library.

```go
// Current custom implementation
func indexOf(s string, ch byte) int {
    for i := 0; i < len(s); i++ {
        if s[i] == ch { return i }
    }
    return -1
}
```

- **Fix:** Use `strings.IndexByte(dsn, '?')`.

---

### NIT-3: Custom `itoa` Function When strconv Exists

- **Location:** `httpserver/security_headers.go`, Lines 114-140
- **Description:** A custom integer-to-string conversion is implemented when `strconv.Itoa` exists.

- **Fix:** Use `strconv.Itoa(cfg.HSTSMaxAge)`.

---

### NIT-4: Missing Error Check After rows.Err()

- **Location:** `database/migrator.go`, Lines 56-62
- **Description:** After iterating `rows.Next()`, `rows.Err()` should be checked for iterator errors.

```go
for rows.Next() {
    var v int
    if err := rows.Scan(&v); err != nil {
        return err
    }
    applied[v] = true
}
// Missing: if err := rows.Err(); err != nil { return err }
```

---

### NIT-5: Unused `getRequiredTemplates` Function

- **Location:** `templates/validator.go`, Lines 116-122
- **Description:** The function `getRequiredTemplates()` is defined but never called (unexported helper).

---

### NIT-6: Inconsistent Error Messages Case

- **Location:** `config/config.go`
- **Description:** Error messages have inconsistent casing: some start with lowercase, others with uppercase.

  - Line 288: `"google OAuth client secret..."`
  - Line 251: `"TLS certificate file path..."`

- **Fix:** Use consistent formatting (typically lowercase for Go errors that may be wrapped).

---

### NIT-7: Migrator Methods Are Stubs

- **Location:** `database/migrator.go`, Lines 138-148
- **Description:** `Rollback()` and `Version()` return nil/0 without implementation. These should either be removed, panic with "not implemented", or have proper implementation.

---

### NIT-8: TLS Config Uses Deprecated Field

- **Location:** `httpserver/tls.go`, Line 23
- **Description:** `PreferServerCipherSuites` is deprecated in Go 1.17+ as Go now makes the optimal choice automatically for TLS 1.3.

```go
PreferServerCipherSuites: true, // Deprecated
```

- **Fix:** Can be removed for TLS 1.3; only keep for TLS 1.2 compatibility comment.

---

### NIT-9: Database Connection Doesn't Apply Pool Settings

- **Location:** `database/connection.go`
- **Description:** The `NewConnection` function accepts a DSN but doesn't apply connection pool settings (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`) that are defined in the config. These settings are never used.

- **Fix:** Either accept config parameters or document that pool settings must be applied by the caller.

---

### NIT-10: Transaction Package Has No Tests

- **Location:** `database/transaction.go`
- **Description:** The transaction wrapper has no dedicated tests. Comment mentions tests were moved to `transaction_test.go` but that file doesn't exist in the listing.

---

## Summary Table

| Category                | Count | Severity     |
| ----------------------- | ----- | ------------ |
| Critical Issues         | 5     | Must Fix     |
| Performance Issues      | 4     | Should Fix   |
| Nitpicks/Best Practices | 10    | Nice to Have |

---

## Recommendations Priority

1. **Immediate:** Fix ISSUE-1 (goroutine leak) and ISSUE-2 (regex compilation)
2. **High:** Fix ISSUE-5 (template parsing per request)
3. **Medium:** Address PERF-1 through PERF-4
4. **Low:** Clean up nitpicks during regular maintenance

---

_Review complete. All files in `internal/platform/` have been analyzed._
