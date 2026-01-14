# Consolidated Code Review: Platform Layer

**Date:** 2026-01-14  
**Reviewers:** Principal Software Engineer / AI Audit  
**Scope:** `internal/platform/` (config, database, errors, health, httpserver, logger, templates, upload, validator)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Critical Issues (Must Fix)](#critical-issues-must-fix)
  - [CRIT-1: Goroutine Leak in RateLimit Middleware](#crit-1-goroutine-leak-in-ratelimit-middleware)
  - [CRIT-2: Regex Compilation on Every Sanitize() Call](#crit-2-regex-compilation-on-every-sanitize-call)
  - [CRIT-3: Silent Error Swallowing in WriteErrorJSON](#crit-3-silent-error-swallowing-in-writeerrorjson)
  - [CRIT-4: Potential Nil Pointer Dereference in Health Checker](#crit-4-potential-nil-pointer-dereference-in-health-checker)
  - [CRIT-5: Template Parsing on Every HealthPage Request](#crit-5-template-parsing-on-every-healthpage-request)
  - [CRIT-6: Memory Leaks / Unbounded Growth in Rate Limiter](#crit-6-memory-leaks--unbounded-growth-in-rate-limiter)
  - [CRIT-7: IP Spoofing via X-Forwarded-For Header](#crit-7-ip-spoofing-via-x-forwarded-for-header)
  - [CRIT-8: Cookie Security Flags Hardcoded to Insecure Values](#crit-8-cookie-security-flags-hardcoded-to-insecure-values)
- [Performance & Optimization](#performance--optimization)
  - [PERF-1: Email Regex Compiled Per Validation Call](#perf-1-email-regex-compiled-per-validation-call)
  - [PERF-2: Username Regex Compiled Per Validation Call](#perf-2-username-regex-compiled-per-validation-call)
  - [PERF-3: Logger Creates New Config Object on Every Error](#perf-3-logger-creates-new-config-object-on-every-error)
  - [PERF-4: Rate Limiter O(n) Scan and Lock Contention](#perf-4-rate-limiter-on-scan-and-lock-contention)
  - [PERF-5: Logger Reflection/Allocation Overhead](#perf-5-logger-reflectionallocation-overhead)
- [Security Issues](#security-issues)
  - [SEC-1: Session Secret Default Value is Weak](#sec-1-session-secret-default-value-is-weak)
- [Nitpicks & Best Practices](#nitpicks--best-practices)
  - [NIT-1: Redundant Check in getEnvStringSlice](#nit-1-redundant-check-in-getenvstringslice)
  - [NIT-2: Custom indexOf Function When stdlib Exists](#nit-2-custom-indexof-function-when-stdlib-exists)
  - [NIT-3: Custom itoa Function When strconv Exists](#nit-3-custom-itoa-function-when-strconv-exists)
  - [NIT-4: Missing Error Check After rows.Err()](#nit-4-missing-error-check-after-rowserr)
  - [NIT-5: Unused getRequiredTemplates Function](#nit-5-unused-getrequiredtemplates-function)
  - [NIT-6: Inconsistent Error Messages Case](#nit-6-inconsistent-error-messages-case)
  - [NIT-7: Migrator Methods Are Stubs](#nit-7-migrator-methods-are-stubs)
  - [NIT-8: TLS Config Uses Deprecated Field](#nit-8-tls-config-uses-deprecated-field)
  - [NIT-9: Database Connection Doesn't Apply Pool Settings](#nit-9-database-connection-doesnt-apply-pool-settings)
  - [NIT-10: Transaction Package Has No Tests](#nit-10-transaction-package-has-no-tests)
  - [NIT-11: Dangerous Database Journal Mode Default](#nit-11-dangerous-database-journal-mode-default)
  - [NIT-12: Hardcoded Rate Limiter Cleanup Ticker](#nit-12-hardcoded-rate-limiter-cleanup-ticker)
  - [NIT-13: Race Condition in Server Startup](#nit-13-race-condition-in-server-startup)
- [Summary Table](#summary-table)
- [Recommendations Priority](#recommendations-priority)

---

## Executive Summary

The platform layer is **well-structured and generally follows idiomatic Go patterns** with good separation of concerns. The implementation is high-quality, idiomatic Go, with attention to detail in logging (structured + human readable), security (TLS, headers), and configuration validation.

However, several issues exist:

- A **goroutine leak** in the rate limiter that can never be stopped
- **Unbounded memory growth** potential in the rate limiter (DoS vector)
- **Silent error handling** in multiple places
- **Performance inefficiencies** in the validator's regex usage (compiled on every call)
- **Security concerns** with IP spoofing via trusted headers and insecure cookie defaults
- A few **concurrency edge cases** with lock contention

**Most critical** is the orphaned cleanup goroutine and the `Sanitize()` function compiling regexes on every call.

---

## Critical Issues (Must Fix)

### CRIT-1: Goroutine Leak in RateLimit Middleware

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 172-186
- **Probability:** **High** (100% reproduction)
- **Description:** The `RateLimit()` middleware spawns a background goroutine for cleanup that runs indefinitely with `for range ticker.C`. This goroutine can never be stopped — there is no context, channel, or shutdown mechanism. Each time `RateLimit()` is called (e.g., in tests or server restarts within the same process), a new goroutine is spawned and the old ones remain alive, leaking memory and goroutines.

```go
// Current code (problematic)
go func() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        limiter.cleanup()  // Runs forever, never stops
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
                return  // Exit on shutdown signal
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

### CRIT-2: Regex Compilation on Every Sanitize() Call

- **Location:** `internal/platform/validator/validator.go`, Lines 143-182 (`Sanitize` function)
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

### CRIT-3: Silent Error Swallowing in WriteErrorJSON

- **Location:** `internal/platform/errors/errors.go`, Lines 149-153
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

### CRIT-4: Potential Nil Pointer Dereference in Health Checker

- **Location:** `internal/platform/health/checker.go`, Lines 23-40
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

### CRIT-5: Template Parsing on Every HealthPage Request

- **Location:** `internal/platform/httpserver/health.go`, Lines 76-82
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

### CRIT-6: Memory Leaks / Unbounded Growth in Rate Limiter

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 163-174
- **Probability:** **Medium**
- **Description:** The `rateLimiter` uses a map `requests map[string][]time.Time` keyed by IP address. Multiple issues exist:

  1. **Unbounded Growth:** While individual IP entries are cleaned up, the map itself can grow unbounded if an attacker spoofs a large number of distinct IP addresses (IP spoofing/Distributed DoS). This can lead to memory exhaustion.

  2. **Stop-the-World Cleanup:** The `cleanup` method locks the entire map (`rl.mu.Lock()`) for the duration of the iteration. For a large map (busy server), this will cause significant latency spikes for all incoming requests waiting on `rl.mu.Lock()` in `allow()`.

- **Proposed Fix:** Use sharded maps, an LRU cache with TTL, or the standard library's `golang.org/x/time/rate` token bucket algorithm:

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

### CRIT-7: IP Spoofing via X-Forwarded-For Header

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 191-194
- **Probability:** **High** (Security)
- **Description:** The rate limiter trusts the `X-Forwarded-For` header without validation, allowing attackers to bypass rate limiting by spoofing IPs.

```go
clientIP := r.RemoteAddr
if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
    clientIP = forwarded  // Trusts header unconditionally
}
```

- **Proposed Fix:** Only trust `X-Forwarded-For` if behind a known proxy:

```go
// Only trust X-Forwarded-For if behind a known proxy
func getClientIP(r *http.Request, trustProxy bool) string {
    if trustProxy {
        if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
            // Take the first IP in the chain (original client)
            if idx := strings.Index(xff, ","); idx != -1 {
                return strings.TrimSpace(xff[:idx])
            }
            return strings.TrimSpace(xff)
        }
        if xri := r.Header.Get("X-Real-IP"); xri != "" {
            return strings.TrimSpace(xri)
        }
    }
    // Strip port from RemoteAddr
    if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
        return host
    }
    return r.RemoteAddr
}
```

---

### CRIT-8: Cookie Security Flags Hardcoded to Insecure Values

- **Location:** `internal/modules/auth/adapters/http_handler_api.go`, Lines 74, 129, 181
- **Probability:** **High** (Security)
- **Description:** Session cookies have `Secure: false` hardcoded with comments saying "Set to true in production with HTTPS". This is **insecure by default** — the comment will be forgotten in production deployment.

```go
http.SetCookie(w, &http.Cookie{
    Name:     "session_token",
    Value:    session.Token,
    Path:     "/",
    Expires:  session.ExpiresAt,
    HttpOnly: true,
    Secure:   false, // Set to true in production with HTTPS  <-- DANGER
    SameSite: http.SameSiteLaxMode,
})
```

- **Proposed Fix:** Get security settings from config:

```go
// Get security settings from config
type HTTPHandler struct {
    authService   authPorts.AuthService
    userService   userPorts.UserService
    templates     *template.Template
    secureCookies bool  // From environment/config
}

// Or better yet, create a cookie helper:
func (h *HTTPHandler) setSessionCookie(w http.ResponseWriter, session *domain.Session) {
    http.SetCookie(w, &http.Cookie{
        Name:     "session_token",
        Value:    session.Token,
        Path:     "/",
        Expires:  session.ExpiresAt,
        HttpOnly: true,
        Secure:   h.isProduction(),  // Derived from config.Server.Environment
        SameSite: http.SameSiteLaxMode,
    })
}
```

---

## Performance & Optimization

### PERF-1: Email Regex Compiled Per Validation Call

- **Location:** `internal/platform/validator/validator.go`, Line 71
- **Description:** Similar to CRIT-2, the email regex is compiled inside `Email()`:

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

- **Location:** `internal/platform/validator/validator.go`, Line 99
- **Description:** Same pattern as email validation.

```go
namePartRegex := regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
```

- **Optimized Code:** Move to package level.

---

### PERF-3: Logger Creates New Config Object on Every Error

- **Location:** `internal/platform/errors/errors.go`, Lines 137-143
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

### PERF-4: Rate Limiter O(n) Scan and Lock Contention

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 216-231
- **Description:** The `allow()` method iterates through all timestamps for an IP on every request. For IPs with many requests, this is O(n) per request. The cleanup also does O(n\*m) where n is IPs and m is timestamps per IP. Additionally, the `sync.Mutex` in `rateLimiter` is a single point of contention for **every single HTTP request**.

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

Or use `sync.RWMutex`, channel-based token bucket, or atomic counters if strict precision isn't required.

---

### PERF-5: Logger Reflection/Allocation Overhead

- **Location:** `internal/platform/logger/logger.go`
- **Description:** The logger handles many types in `formatHTTPRequest` and `log` using type switches and reflection-like behavior (`fmt.Sprintf("%v")`). It also allocates maps for fields on every log call.
- **Optimization:** For a high-throughput forum, this is likely fine. If profiling shows issues, switch to `zerolog` or `zap` for zero-allocation logging.

---

## Security Issues

### SEC-1: Session Secret Default Value is Weak

- **Location:** `internal/platform/config/config.go`, Line 138
- **Description:**

```go
cfg.Session.Secret = getEnvString("SESSION_SECRET", "defaultsecret")
```

While validation catches this in production, in development the default is weak and could be committed to version control or used accidentally.

- **Proposed Fix:**

```go
// Generate a random secret if not provided in non-production:
if cfg.Server.Environment != "production" && cfg.Session.Secret == "defaultsecret" {
    // Generate random secret for development
    randomBytes := make([]byte, 32)
    rand.Read(randomBytes)
    cfg.Session.Secret = base64.StdEncoding.EncodeToString(randomBytes)
    log.Println("Warning: Using auto-generated session secret for development")
}
```

---

## Nitpicks & Best Practices

### NIT-1: Redundant Check in getEnvStringSlice

- **Location:** `internal/platform/config/env_parser.go`, Lines 46-58
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

### NIT-2: Custom indexOf Function When stdlib Exists

- **Location:** `internal/platform/database/connection.go`, Lines 95-103
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

### NIT-3: Custom itoa Function When strconv Exists

- **Location:** `internal/platform/httpserver/security_headers.go`, Lines 114-140
- **Description:** A custom integer-to-string conversion is implemented when `strconv.Itoa` exists.

- **Fix:** Use `strconv.Itoa(cfg.HSTSMaxAge)`.

---

### NIT-4: Missing Error Check After rows.Err()

- **Location:** `internal/platform/database/migrator.go`, Lines 56-62
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

### NIT-5: Unused getRequiredTemplates Function

- **Location:** `internal/platform/templates/validator.go`, Lines 116-122
- **Description:** The function `getRequiredTemplates()` is defined but never called (unexported helper).

---

### NIT-6: Inconsistent Error Messages Case

- **Location:** `internal/platform/config/config.go`
- **Description:** Error messages have inconsistent casing: some start with lowercase, others with uppercase.

  - Line 288: `"google OAuth client secret..."`
  - Line 251: `"TLS certificate file path..."`

- **Fix:** Use consistent formatting (typically lowercase for Go errors that may be wrapped).

---

### NIT-7: Migrator Methods Are Stubs

- **Location:** `internal/platform/database/migrator.go`, Lines 138-148
- **Description:** `Rollback()` and `Version()` return nil/0 without implementation. These should either be removed, panic with "not implemented", or have proper implementation.

---

### NIT-8: TLS Config Uses Deprecated Field

- **Location:** `internal/platform/httpserver/tls.go`, Line 23
- **Description:** `PreferServerCipherSuites` is deprecated in Go 1.17+ as Go now makes the optimal choice automatically for TLS 1.3.

```go
PreferServerCipherSuites: true, // Deprecated
```

- **Fix:** Can be removed for TLS 1.3; only keep for TLS 1.2 compatibility comment.

---

### NIT-9: Database Connection Doesn't Apply Pool Settings

- **Location:** `internal/platform/database/connection.go`
- **Description:** The `NewConnection` function accepts a DSN but doesn't apply connection pool settings (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`) that are defined in the config. These settings are never used.

- **Fix:** Either accept config parameters or document that pool settings must be applied by the caller.

---

### NIT-10: Transaction Package Has No Tests

- **Location:** `internal/platform/database/transaction.go`
- **Description:** The transaction wrapper has no dedicated tests. Comment mentions tests were moved to `transaction_test.go` but that file doesn't exist in the listing.

---

### NIT-11: Dangerous Database Journal Mode Default

- **Location:** `internal/platform/database/connection.go`
- **Description:** The code forces `PRAGMA journal_mode = MEMORY`. While fast, this guarantees execution relies on volatile RAM, significantly increasing the risk of database corruption in a crash.
- **Recommendation:** Use `WAL` mode for better durability with good performance.

---

### NIT-12: Hardcoded Rate Limiter Cleanup Ticker

- **Location:** `internal/platform/httpserver/middleware.go`
- **Description:** The rate limiter cleanup ticker is hardcoded to `1 minute`. This might not be aggressive enough for high traffic or too aggressive for low traffic. Consider making this configurable.

---

### NIT-13: Race Condition in Server Startup

- **Location:** `internal/platform/httpserver/server.go`
- **Description:** The server waits an arbitrary `100ms` to check for startup errors. This is flaky. Using a channel to signal the listener is ready is more robust.

---

## Summary Table

| Category                | Count | Severity     |
| ----------------------- | ----- | ------------ |
| Critical Issues         | 8     | Must Fix     |
| Performance Issues      | 5     | Should Fix   |
| Security Issues         | 1     | Should Fix   |
| Nitpicks/Best Practices | 13    | Nice to Have |

---

## Recommendations Priority

1. **Immediate (P0):**

   - Fix CRIT-1 (goroutine leak in rate limiter)
   - Fix CRIT-2 (regex compilation on every call)
   - Fix CRIT-7 (IP spoofing bypass)
   - Fix CRIT-8 (insecure cookie defaults)

2. **High (P1):**

   - Fix CRIT-5 (template parsing per request)
   - Fix CRIT-6 (memory leak/DoS in rate limiter)
   - Address PERF-1 through PERF-4 (regex and logger performance)

3. **Medium (P2):**

   - Fix CRIT-3 (silent error swallowing)
   - Fix CRIT-4 (nil pointer in health checker)
   - Fix SEC-1 (weak session secret default)

4. **Low (P3):**
   - Clean up nitpicks during regular maintenance
   - Add missing tests for transaction package

---

_Review consolidated from multiple audits conducted on 2026-01-14._
_All files in `internal/platform/` have been analyzed._
