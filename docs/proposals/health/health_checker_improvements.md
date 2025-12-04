# Health Checker Improvements for Go API Endpoints

## Overview

This document analyzes the current health checker implementation in the `internal/platform/httpserver` and `internal/platform/health` packages of the forum application. It evaluates whether the health check properly verifies API endpoint registration and responsiveness according to Go API health check best practices.

## Current Implementation

### Files Involved
- `internal/platform/httpserver/health.go`: Contains `HealthAPI` and `HealthPage` handlers
- `internal/platform/health/checker.go`: Contains the `Checker` struct and `Check` method

### Current Behavior
- **HealthAPI**: Calls `checker.Check(ctx)` and returns a JSON map of results. Sets HTTP 200 if all components are "up", otherwise 503.
- **HealthPage**: Renders an HTML page with health status.
- **Checker.Check**: Only performs a database ping (`db.PingContext`) and returns `{"database": "up|down"}`.

### Key Limitations
1. **No Route Verification**: The health check does not verify that HTTP routes/endpoints are registered or responding.
2. **No Readiness/Liveness Separation**: Only one health endpoint that mixes infrastructure checks with application readiness.
3. **No Timeouts**: Only a 5-second timeout on DB ping; no per-check timeouts.
4. **No Concurrency**: Checks run sequentially, potentially blocking.
5. **Simple Status Only**: Returns basic "up/down" strings without latency, error details, or structured data.
6. **No Configurability**: Hard-coded checks; no way to specify required endpoints.

## Industry Best Practices for Go API Health Checks

### Core Principles
1. **Liveness vs Readiness**:
   - **Liveness** (`/health/live`): Process is running (simple 200 response)
   - **Readiness** (`/health/ready`): Application is ready to serve traffic (dependencies healthy, routes registered)

2. **Dependency Checks**: Verify critical dependencies (DB, cache, external services) are accessible.

3. **Route Verification**: Ensure key API endpoints are registered and responding (not just 404).

4. **Timeouts and Resilience**: Each check should have bounded timeouts to prevent cascading failures.

5. **Concurrency**: Run independent checks concurrently to minimize total response time.

6. **Structured Response**: Return detailed status with latency, error messages, and component breakdown.

7. **Caching**: For expensive checks, use background polling with short TTL to avoid blocking requests.

### Go-Specific Best Practices
- Use `httptest` for in-process route testing instead of pattern string comparisons.
- Leverage `context.WithTimeout` for per-check timeouts.
- Use goroutines + `sync.WaitGroup` for concurrent checks.
- Return structured JSON with timestamps and component details.

## Issues Found

### 1. No API Endpoint Verification
The current implementation only checks database connectivity. It does not verify that HTTP routes are registered or responding, which is critical for API readiness.

### 2. Brittle Route Detection (If Implemented)
The code attempts to use `ServeMux.Handler()` pattern strings, which is fragile and router-specific. Better to perform actual HTTP requests.

### 3. Lack of Timeouts and Concurrency
No per-check timeouts or concurrent execution, leading to potential blocking and slow responses.

### 4. Poor Error Handling and Diagnostics
Simple "up/down" status without latency metrics or error details makes debugging difficult.

### 5. Not Configurable
Required endpoints are hard-coded, making it inflexible for different deployments.

## Suggested Improvements

### 1. Separate Liveness and Readiness Endpoints
- `/health/live`: Simple liveness probe (always 200 if process is running)
- `/health/ready` or `/health-api`: Comprehensive readiness check including routes

### 2. Enhanced Checker Design
```go
type Endpoint struct {
    Method string
    Path   string
    Name   string
}

type ComponentResult struct {
    Status    string        `json:"status"`
    LatencyMs int64         `json:"latency_ms"`
    Error     string        `json:"error,omitempty"`
    Detail    string        `json:"detail,omitempty"`
}

type Checker struct {
    db              *sql.DB
    router          http.Handler
    endpoints       []Endpoint
    perCheckTimeout time.Duration
}

func NewChecker(db *sql.DB, router http.Handler, endpoints []Endpoint) *Checker {
    return &Checker{
        db:              db,
        router:          router,
        endpoints:       endpoints,
        perCheckTimeout: 500 * time.Millisecond,
    }
}
```

### 3. Route Verification with httptest
```go
import (
    "context"
    "net/http"
    "net/http/httptest"
    "regexp"
    "time"
)

var placeholderRE = regexp.MustCompile(`\{[^/}]+\}`)

func (c *Checker) routeExists(ctx context.Context, method, pattern string) ComponentResult {
    start := time.Now()
    
    testPath := placeholderRE.ReplaceAllString(pattern, "1")
    req := httptest.NewRequest(method, testPath, nil).WithContext(ctx)
    rec := httptest.NewRecorder()
    
    done := make(chan struct{})
    go func() {
        c.router.ServeHTTP(rec, req)
        close(done)
    }()
    
    select {
    case <-done:
        latency := time.Since(start).Milliseconds()
        if rec.Code == http.StatusNotFound {
            return ComponentResult{
                Status:    "down",
                LatencyMs: latency,
                Error:     "route not found",
            }
        }
        return ComponentResult{
            Status:    "up",
            LatencyMs: latency,
            Detail:    fmt.Sprintf("%s %s -> %d", method, pattern, rec.Code),
        }
    case <-time.After(c.perCheckTimeout):
        return ComponentResult{
            Status:    "down",
            LatencyMs: int64(c.perCheckTimeout.Milliseconds()),
            Error:     "timeout",
        }
    }
}
```

### 4. Concurrent Check Execution
```go
func (c *Checker) Check(ctx context.Context) map[string]ComponentResult {
    results := make(map[string]ComponentResult)
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    // Database check
    wg.Add(1)
    go func() {
        defer wg.Done()
        result := c.checkDatabase(ctx)
        mu.Lock()
        results["database"] = result
        mu.Unlock()
    }()
    
    // Route checks
    for _, endpoint := range c.endpoints {
        wg.Add(1)
        go func(ep Endpoint) {
            defer wg.Done()
            result := c.routeExists(ctx, ep.Method, ep.Path)
            mu.Lock()
            results[ep.Name] = result
            mu.Unlock()
        }(endpoint)
    }
    
    wg.Wait()
    return results
}
```

### 5. Structured JSON Response
Update `HealthAPI` to return:
```json
{
  "status": "up",
  "checked_at": "2025-11-16T12:34:56Z",
  "components": {
    "database": {"status":"up","latency_ms":12},
    "auth_api": {"status":"up","latency_ms":45,"detail":"POST /auth/login -> 401"},
    "posts_api": {"status":"down","latency_ms":500,"error":"route timed out"}
  }
}
```

### 6. Configuration and Injection
- Pass required endpoints to `NewChecker` from `cmd/forum/wire/app.go`
- Make timeouts configurable via config

### 7. Optional: Background Caching
For production deployments, implement background polling:
```go
type CachedChecker struct {
    Checker
    cache     map[string]ComponentResult
    cacheTime time.Time
    ttl       time.Duration
    mu        sync.RWMutex
}

func (cc *CachedChecker) Check(ctx context.Context) map[string]ComponentResult {
    cc.mu.RLock()
    if time.Since(cc.cacheTime) < cc.ttl {
        defer cc.mu.RUnlock()
        return cc.cache
    }
    cc.mu.RUnlock()
    
    // Recalculate
    results := cc.Checker.Check(ctx)
    cc.mu.Lock()
    cc.cache = results
    cc.cacheTime = time.Now()
    cc.mu.Unlock()
    
    return results
}
```

## Implementation Priority

1. **High Priority**: Add route verification with `httptest` and basic concurrency
2. **Medium Priority**: Structured JSON response with latency/error details
3. **Low Priority**: Background caching and separate liveness endpoint

## Testing Recommendations

- Unit tests for `routeExists` with mocked router
- Integration tests that verify actual endpoints return expected status codes
- Load tests to ensure health checks don't impact performance

## Conclusion

The current health checker implementation is insufficient for production Go APIs. It only verifies database connectivity and lacks proper endpoint verification. Implementing the suggested improvements will provide robust health checking that follows industry best practices, ensuring the application is truly ready to serve traffic.

The key changes involve using `httptest` for route verification, adding concurrency and timeouts, and returning structured diagnostic information.