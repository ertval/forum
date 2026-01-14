# Code Audit Review - Forum Codebase

**Date:** 2026-01-14 14:53  
**Reviewer:** Principal Software Engineer / Low-Level Systems Architect  
**Scope:** Full codebase audit (`internal/`, `cmd/`, platform layer)

---

## Executive Summary

The forum codebase demonstrates **solid architectural foundations** with well-structured hexagonal (ports & adapters) architecture, proper separation of concerns, and thoughtful security measures (UUID public IDs, input sanitization, TLS configuration). However, there are **several critical issues** around goroutine management, error handling gaps, and potential resource leaks that require immediate attention. The code quality ranges from excellent (auth module, security headers) to scaffolded/incomplete (moderation, notification modules).

---

## Critical Issues (Must Fix)

### ISSUE-1: Detached Goroutines with Background Context

- **Location:** Multiple files:

  - `internal/modules/post/application/service.go`, Lines 131-133, 196-198
  - `internal/modules/comment/application/service.go`, Lines 59-61, 111-113
  - `internal/modules/reaction/application/service.go`, Lines 110-112, 148-150

- **Probability:** High

- **Description:** The service layers spawn goroutines with `context.Background()` for incrementing/decrementing user stats. These goroutines:

  1. **Cannot be cancelled** when the application shuts down, leading to potential goroutine leaks
  2. **Lose request tracing/logging context** (if tracing is added in future)
  3. **Silently swallow errors** with `_ = s.userService.IncrementPostCount(...)`
  4. **No timeout protection** - if the database is slow, these goroutines hang indefinitely

- **Proposed Fix:**

```go
// Instead of:
go func() {
    _ = s.userService.IncrementPostCount(context.Background(), userID)
}()

// Use a pattern with timeout and proper error logging:
func (s *Service) CreatePost(ctx context.Context, userID int, /* ... */) (*domain.Post, error) {
    // ... existing logic ...

    // Use a detached context with short timeout for background stat updates
    // These are non-critical and should not block the main request
    go func(userID int) {
        bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := s.userService.IncrementPostCount(bgCtx, userID); err != nil {
            // Log error - don't ignore silently
            // Consider using the logger from service container
            log.Printf("warning: failed to increment post count for user %d: %v", userID, err)
        }
    }(userID)

    return post, nil
}

// Better yet: Use a worker pool pattern or async job queue for non-critical operations
```

---

### ISSUE-2: Rate Limiter Cleanup Goroutine Never Stops

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 179-186

- **Probability:** High

- **Description:** The rate limiter's cleanup goroutine runs forever (`for range ticker.C`) without any mechanism to stop it during graceful shutdown. This creates a **goroutine leak** and prevents clean application shutdown.

```go
// Current problematic code:
go func() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        limiter.cleanup()  // Runs forever, never stops
    }
}()
```

- **Proposed Fix:**

```go
// RateLimit middleware limits the number of requests per time window.
func RateLimit(requests int, windowSeconds int) Middleware {
    limiter := &rateLimiter{
        requests: make(map[string][]time.Time),
        limit:    requests,
        window:   time.Duration(windowSeconds) * time.Second,
        stopCh:   make(chan struct{}),  // Add stop channel
    }

    // Cleanup goroutine with stop mechanism
    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                limiter.cleanup()
            case <-limiter.stopCh:
                return  // Exit on shutdown signal
            }
        }
    }()

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... existing logic ...
        })
    }
}

// Add to rateLimiter struct:
type rateLimiter struct {
    requests map[string][]time.Time
    mu       sync.Mutex
    limit    int
    window   time.Duration
    stopCh   chan struct{}  // NEW: for graceful shutdown
}

// Add Stop method and call it during server shutdown:
func (rl *rateLimiter) Stop() {
    close(rl.stopCh)
}
```

---

### ISSUE-3: Template Parsing on Every Page Request

- **Location:** `internal/modules/auth/adapters/http_handler_page.go`, Lines 25, 44

- **Probability:** High (Performance + Potential Crash)

- **Description:** Templates are parsed **on every single HTTP request** instead of being parsed once at startup:

```go
func (h *HTTPHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
    // This is called on EVERY request!
    tmpl, err := template.ParseFiles("templates/base.html", "templates/login.html")
    // ...
}
```

This causes:

1. **Performance degradation** - Disk I/O + parsing on every page load
2. **Potential panic** if templates are missing (file not found during high traffic)
3. **Memory churn** - Repeated allocations for the same templates

- **Proposed Fix:**

```go
// Parse templates once in NewHTTPHandler or at module initialization:
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
    // Templates should already be shared via the templates parameter
    // The handler already has h.templates - use it!
    return &HTTPHandler{
        authService: services.Auth(),
        userService: services.User(),
        templates:   templates,
    }
}

// In page handlers, use the pre-parsed templates:
func (h *HTTPHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "Title": "Login",
    }

    // Use pre-parsed templates
    if h.templates == nil {
        http.Error(w, "Templates not configured", http.StatusInternalServerError)
        return
    }

    if err := h.templates.ExecuteTemplate(w, "login", data); err != nil {
        http.Error(w, "Failed to render login page", http.StatusInternalServerError)
        return
    }
}
```

---

### ISSUE-4: Missing Transaction Rollback on Post Category Insert Failure

- **Location:** `internal/modules/post/adapters/sqlite_repository.go`, Lines 27-88

- **Probability:** Medium (Data Inconsistency)

- **Description:** While the transaction has `defer tx.Rollback()`, when category insertion fails, the post is already inserted. The error handling is correct, but there's a subtle issue: if `GetByName` fails for a non-existent category, the error message leaks internal details:

```go
err := tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE LOWER(name) = LOWER(?)", categoryName).Scan(&categoryID)
if err != nil {
    return fmt.Errorf("category %s not found: %w", categoryName, err)  // Leaks sql.ErrNoRows
}
```

- **Proposed Fix:**

```go
// Map database errors to domain errors consistently:
if err != nil {
    if err == sql.ErrNoRows {
        return domain.ErrCategoryNotFound  // Domain error, not wrapped DB error
    }
    return fmt.Errorf("failed to lookup category: %w", err)  // Don't expose category name in production
}
```

---

### ISSUE-5: Cookie Security Flags Hardcoded to Insecure Values

- **Location:** Multiple files:

  - `internal/modules/auth/adapters/http_handler_api.go`, Lines 74, 129, 181

- **Probability:** High (Security)

- **Description:** Session cookies have `Secure: false` hardcoded with comments saying "Set to true in production with HTTPS":

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

This is **insecure by default** - the comment will be forgotten in production deployment.

- **Proposed Fix:**

```go
// Get security settings from config
type HTTPHandler struct {
    authService authPorts.AuthService
    userService userPorts.UserService
    templates   *template.Template
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

### PERF-1: N+1 Query Problem in Post Listing

- **Location:** `internal/modules/post/adapters/sqlite_repository.go`, Lines 431-436

- **Description:** For each post in the list, a separate query fetches categories:

```go
for rows.Next() {
    // ... scan post ...

    // This is called for EACH post - N+1 problem!
    categories, err := r.getPostCategories(ctx, post.ID)
}
```

With 50 posts (default limit), this executes **51 queries** instead of 2.

- **Optimized Code:**

```go
// Batch fetch all categories for the posts in a single query:
func (r *SQLitePostRepository) List(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error) {
    // First query: get posts
    posts, postIDs := r.fetchPosts(ctx, filter)

    // Second query: get all categories for all posts at once
    categoryMap := r.batchGetPostCategories(ctx, postIDs)

    // Assign categories to posts
    for _, post := range posts {
        post.Categories = categoryMap[post.ID]
    }

    return posts, nil
}

// Batch fetch categories:
func (r *SQLitePostRepository) batchGetPostCategories(ctx context.Context, postIDs []int) map[int][]string {
    if len(postIDs) == 0 {
        return map[int][]string{}
    }

    placeholders := strings.Repeat(",?", len(postIDs)-1)
    query := fmt.Sprintf(`
        SELECT pc.post_id, c.name
        FROM categories c
        INNER JOIN post_categories pc ON c.id = pc.category_id
        WHERE pc.post_id IN (?%s)
        ORDER BY c.name
    `, placeholders)

    args := make([]interface{}, len(postIDs))
    for i, id := range postIDs {
        args[i] = id
    }

    rows, err := r.db.QueryContext(ctx, query, args...)
    // ... process and return map[postID][]categoryNames
}
```

---

### PERF-2: Repeated Regex Compilation in Sanitize Function

- **Location:** `internal/platform/validator/validator.go`, Lines 143-182

- **Description:** The `Sanitize()` function compiles regexes **on every call**:

```go
func Sanitize(input string) string {
    // These are compiled EVERY time Sanitize() is called!
    reScript := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
    reStyle := regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
    reTags := regexp.MustCompile(`<[^>]+>`)
    reSpace := regexp.MustCompile(`\s+`)
    // ...
}
```

Regex compilation is **expensive** and this function is called for every user input.

- **Optimized Code:**

```go
// Compile regexes once at package initialization:
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

### PERF-3: Unnecessary Memory Allocation in CreatePostPreview

- **Location:** `internal/modules/post/adapters/http_handler.go`, Lines 207-232

- **Description:** The function performs redundant bounds checking and string slicing:

```go
func createPostPreview(content string) string {
    const previewLength = 100
    if len(content) <= previewLength {
        return content
    }

    preview := content[:previewLength]  // First slice
    if len(content) > previewLength {   // Redundant check - already true from above
        // ...
    }
}
```

- **Optimized Code:**

```go
func createPostPreview(content string) string {
    const previewLength = 100
    if len(content) <= previewLength {
        return content
    }

    // Find last space before previewLength to avoid cutting words
    lastSpace := strings.LastIndexByte(content[:previewLength], ' ')
    if lastSpace > previewLength/2 {
        return content[:lastSpace] + "..."
    }
    return content[:previewLength] + "..."
}
```

---

## Error Handling & Robustness Issues

### ERR-1: Silent Error Swallowing in Multiple Locations

- **Location:** Multiple files

- **Description:** Many places ignore errors with `_`:

```go
// auth/adapters/http_handler_page.go:62
_ = h.authService.Logout(r.Context(), cookie.Value) // We ignore the error for frontend UX

// auth/application/service.go:194
_ = s.sessionRepo.Delete(ctx, sessionToken) // Best effort cleanup

// Multiple services with go func() { _ = ... }()
```

While some are intentional "best effort" operations, they should at minimum be logged for debugging.

- **Proposed Fix:**

```go
// Use a consistent pattern for non-critical errors:
if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
    // Log but don't fail the user experience
    log.Printf("debug: logout cleanup failed: %v", err)
}
```

---

### ERR-2: Method Check After Pattern-Based Routing

- **Location:** `internal/modules/post/adapters/http_handler_api.go`, Lines 40-43, 168-172

- **Description:** Handlers manually check HTTP methods even though routes use pattern-based routing (`POST /api/posts`):

```go
func (h *HTTPHandler) CreatePostAPI(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {  // Redundant - route is "POST /api/posts"
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    // ...
}
```

With Go 1.22+ pattern routing (`router.HandleFunc("POST /api/posts", ...)`), this check is redundant.

- **Proposed Fix:**
  Remove redundant method checks when using pattern-based routing. The router already enforces the method.

---

## Security Analysis

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

### SEC-2: IP Spoofing via X-Forwarded-For Header

- **Location:** `internal/platform/httpserver/middleware.go`, Lines 191-194

- **Description:**

```go
clientIP := r.RemoteAddr
if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
    clientIP = forwarded  // Trusts header unconditionally
}
```

This trusts the `X-Forwarded-For` header without validation, allowing attackers to bypass rate limiting by spoofing IPs.

- **Proposed Fix:**

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

## Nitpicks & Best Practices

1. **Dead Code:** The `min` function in `http_handler.go` is defined but never used. Go 1.21+ has built-in `min`.

2. **Inconsistent Error Wrapping:** Some errors use `fmt.Errorf("... %w", err)` while others return raw errors. Standardize on wrapped errors with context.

3. **TODO Comments Still Present:** Many `// TODO:` comments indicate incomplete implementations (moderation, notification modules). These should be tracked in issue tracker.

4. **JSON Encoding Error Ignored:**

   ```go
   if err := json.NewEncoder(w).Encode(data); err != nil {
       fmt.Printf("Error encoding JSON response: %v\n", err)  // Uses fmt.Printf instead of logger
   }
   ```

5. **Deprecated Functions Warning:** `http.Handler` wrapped functions in middleware.go use deprecated patterns. Consider using the modern approach.

6. **Hardcoded Pagination Limits:** Default limits (50, 20) are scattered across handlers. Centralize in config.

7. **Missing Context Timeout in Main:** The shutdown context in `main.go` is created but the `ctx.Done()` select case may never trigger since we wait on `app.Shutdown()` first.

8. **Unused `templates` Field:** The `HTTPHandler.templates` field in auth module receives templates but the page handlers don't use it, instead re-parsing templates.

9. **PreferServerCipherSuites Deprecated:** In Go 1.22+, `PreferServerCipherSuites` is deprecated and has no effect.

10. **Database Journal Mode:** Using `PRAGMA journal_mode = MEMORY` in production could cause data loss on crash. Consider WAL mode for better durability with good performance.

---

## Summary

| Category       | Critical | Medium | Low    |
| -------------- | -------- | ------ | ------ |
| Concurrency    | 2        | 0      | 0      |
| Security       | 2        | 1      | 1      |
| Performance    | 0        | 3      | 2      |
| Error Handling | 0        | 2      | 3      |
| Best Practices | 0        | 0      | 10     |
| **Total**      | **4**    | **6**  | **16** |

**Priority Recommendations:**

1. **Immediate:** Fix goroutine management (ISSUE-1, ISSUE-2)
2. **High:** Fix cookie security flags (ISSUE-5) and template parsing (ISSUE-3)
3. **Medium:** Address N+1 queries and regex compilation
4. **Low:** Refactor for consistency and remove dead code

---

_Review completed: 2026-01-14_
