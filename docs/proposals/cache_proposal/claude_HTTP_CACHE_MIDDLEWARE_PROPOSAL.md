# HTTP Cache Middleware Proposal

## Executive Summary

This document proposes an HTTP-level caching middleware that integrates with the platform cache system (in-memory/Redis) described in `claude_CACHING_PROPOSAL.md`. The middleware provides transparent response caching at the HTTP layer, reducing database load and improving response times for cacheable endpoints.

## Design Philosophy

### Core Principles

- **Cache abstraction agnostic**: Works with any `cache.Cache` implementation (in-memory, Redis, etc.)
- **Selective caching**: Opt-in per route with configurable strategies
- **HTTP-aware**: Respects HTTP semantics (cache headers, methods, status codes)
- **Vary support**: Cache different responses for same URL based on headers
- **Conditional requests**: Support ETags and Last-Modified for validation
- **Invalidation patterns**: URL-based and tag-based cache invalidation
- **Non-blocking**: Cache operations don't block request processing on errors

### Why HTTP Caching?

**Repository caching** (from the original proposal) caches at the data layer - individual entities and queries. **HTTP caching** caches complete rendered responses, providing:

1. **Faster response times**: No handler execution, no rendering
2. **Reduced CPU usage**: Skip JSON marshaling, template rendering
3. **Guest user optimization**: Cache public pages for unauthenticated users
4. **API response caching**: Cache expensive aggregations, reports
5. **Complementary strategy**: Works alongside repository-level caching

**When to use each:**

- **Repository cache**: For data reused across different endpoints/views
- **HTTP cache**: For complete responses that don't change per-user

---

## Package Structure

All HTTP cache middleware components are located in the **`cache`** package, not `httpserver`:

```text
internal/platform/cache/
├── cache.go                    # Core cache interface (from original proposal)
├── inmemory/
│   └── inmemory.go            # In-memory implementation
├── redis/
│   └── redis.go               # Redis implementation
└── http/                       # HTTP cache middleware (NEW)
    ├── cache_key.go           # Cache key builder
    ├── response_writer.go     # Response capture wrapper
    ├── cached_response.go     # Cached response storage
    ├── cache_config.go        # Cache strategies
    ├── middleware.go          # Main cache middleware
    ├── invalidator.go         # Cache invalidation helpers
    ├── stats.go               # Metrics and monitoring
    └── middleware_test.go     # Unit tests
```

**Rationale:**

- `httpserver` package contains **generic** HTTP server infrastructure (server, basic middleware)
- `cache/http` package contains **cache-specific** HTTP middleware (specialization of cache)
- This follows the principle of keeping related functionality together
- Import as: `cachehttp "forum/internal/platform/cache/http"`

---

## Architecture Overview

### Middleware Position in Stack

```text
Request Flow:
┌─────────────────────┐
│   HTTP Request      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Recovery Middleware│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Logger Middleware  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ **CACHE MIDDLEWARE**│  ◄── New middleware (before auth)
└──────────┬──────────┘
           │
    Cache Hit? ───────────┐
           │              │
      No   │         Yes  │
           ▼              ▼
┌─────────────────────┐  │
│  Auth Middleware    │  │
└──────────┬──────────┘  │
           │              │
           ▼              │
┌─────────────────────┐  │
│   HTTP Handlers     │  │
└──────────┬──────────┘  │
           │              │
           ▼              │
     [Response]           │
           │              │
   Store in Cache         │
           │              │
           └──────────────┤
                          ▼
                   Return Response
```

### Key Decisions

1. **Position**: After logger, before auth
   - Can cache responses before auth check (for public endpoints)
   - Logged requests include cache hits/misses
   
2. **Scope**: Only GET and HEAD requests
   - POST/PUT/DELETE are never cached (per HTTP semantics)
   - OPTIONS requests can be cached but with shorter TTL

3. **Granularity**: Per-route opt-in
   - Not all GET requests should be cached
   - Different routes have different caching needs

---

## Core Implementation

### 1. Cache Key Strategy

**File**: `internal/platform/cache/http/cache_key.go`

```go
package http

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// CacheKeyBuilder constructs cache keys for HTTP responses.
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder creates a new cache key builder.
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// BuildKey creates a cache key from HTTP request.
// Format: http:cache:{prefix}:{method}:{path}:{varyHash}
func (b *CacheKeyBuilder) BuildKey(r *http.Request, varyHeaders []string) string {
	parts := []string{
		"http:cache",
		b.prefix,
		r.Method,
		r.URL.Path,
	}

	// Add query parameters (sorted for consistency)
	if r.URL.RawQuery != "" {
		parts = append(parts, "query:"+b.hashQuery(r.URL.Query()))
	}

	// Add vary headers if specified
	if len(varyHeaders) > 0 {
		varyHash := b.hashVaryHeaders(r, varyHeaders)
		parts = append(parts, "vary:"+varyHash)
	}

	return strings.Join(parts, ":")
}

// hashQuery creates a consistent hash of query parameters.
func (b *CacheKeyBuilder) hashQuery(query map[string][]string) string {
	if len(query) == 0 {
		return ""
	}

	// Sort keys for consistency
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	var parts []string
	for _, k := range keys {
		vals := query[k]
		sort.Strings(vals)
		for _, v := range vals {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Hash for shorter keys
	queryStr := strings.Join(parts, "&")
	hash := sha256.Sum256([]byte(queryStr))
	return fmt.Sprintf("%x", hash[:8]) // First 8 bytes
}

// hashVaryHeaders creates a hash of specified request headers.
func (b *CacheKeyBuilder) hashVaryHeaders(r *http.Request, headers []string) string {
	var parts []string
	for _, h := range headers {
		if val := r.Header.Get(h); val != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", h, val))
		}
	}

	if len(parts) == 0 {
		return "none"
	}

	hashStr := strings.Join(parts, "&")
	hash := sha256.Sum256([]byte(hashStr))
	return fmt.Sprintf("%x", hash[:8])
}
```

**Key Design Decisions:**

- **Consistent hashing**: Query params and headers are sorted
- **Short keys**: SHA-256 truncated to 8 bytes (16 hex chars)
- **Vary support**: Different cache entries for different header combinations
- **Query-aware**: Same endpoint with different params = different cache entries

---

### 2. Response Capture

To cache HTTP responses, we need to capture the response writer's output:

**File**: `internal/platform/cache/http/response_writer.go`

```go
package http

import (
	"bytes"
	stdhttp "net/http"
)

// ResponseCapture wraps http.ResponseWriter to capture response for caching.
type ResponseCapture struct {
	stdhttp.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	headers    stdhttp.Header
}

// NewResponseCapture creates a new response capture wrapper.
func NewResponseCapture(w stdhttp.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		statusCode:     stdhttp.StatusOK, // Default status
		body:           new(bytes.Buffer),
		headers:        make(stdhttp.Header),
	}
}

// WriteHeader captures status code.
func (rc *ResponseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

// Write captures response body.
func (rc *ResponseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b) // Capture for caching
	return rc.ResponseWriter.Write(b)
}

// Header returns header map for modification.
func (rc *ResponseCapture) Header() stdhttp.Header {
	// Copy headers for caching
	h := rc.ResponseWriter.Header()
	for k, v := range h {
		rc.headers[k] = v
	}
	return h
}

// StatusCode returns the captured status code.
func (rc *ResponseCapture) StatusCode() int {
	return rc.statusCode
}

// Body returns the captured response body.
func (rc *ResponseCapture) Body() []byte {
	return rc.body.Bytes()
}

// Headers returns captured headers.
func (rc *ResponseCapture) Headers() stdhttp.Header {
	return rc.headers
}
```

---

### 3. Cached Response Storage

**File**: `internal/platform/cache/http/cached_response.go`

```go
package http

import (
	"bytes"
	"encoding/gob"
	stdhttp "net/http"
	"time"
)

// CachedResponse represents a cached HTTP response.
type CachedResponse struct {
	StatusCode int
	Headers    stdhttp.Header
	Body       []byte
	CachedAt   time.Time
	ExpiresAt  time.Time
}

// NewCachedResponse creates a cached response from a captured response.
func NewCachedResponse(capture *ResponseCapture, ttl time.Duration) *CachedResponse {
	now := time.Now()
	return &CachedResponse{
		StatusCode: capture.StatusCode(),
		Headers:    capture.Headers(),
		Body:       capture.Body(),
		CachedAt:   now,
		ExpiresAt:  now.Add(ttl),
	}
}

// IsExpired checks if the cached response has expired.
func (cr *CachedResponse) IsExpired() bool {
	return time.Now().After(cr.ExpiresAt)
}

// WriteTo writes the cached response to an http.ResponseWriter.
func (cr *CachedResponse) WriteTo(w stdhttp.ResponseWriter) error {
	// Copy headers
	for k, values := range cr.Headers {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	// Add cache headers
	w.Header().Set("X-Cache", "HIT")
	w.Header().Set("X-Cache-Date", cr.CachedAt.Format(time.RFC1123))

	// Write status and body
	w.WriteHeader(cr.StatusCode)
	_, err := w.Write(cr.Body)
	return err
}

// Marshal serializes cached response for storage.
func (cr *CachedResponse) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(cr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalCachedResponse deserializes a cached response.
func UnmarshalCachedResponse(data []byte) (*CachedResponse, error) {
	var cr CachedResponse
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}
```

---

### 4. Cache Middleware Configuration

**File**: `internal/platform/cache/http/cache_config.go`

```go
package http

import (
	"time"
	stdhttp "net/http"
)

// CacheStrategy defines how a route should be cached.
type CacheStrategy struct {
	// Enabled determines if caching is active for this route
	Enabled bool

	// TTL is how long to cache the response
	TTL time.Duration

	// VaryBy specifies which headers to include in cache key
	// Example: []string{"Accept", "Accept-Language"}
	VaryBy []string

	// CacheableStatusCodes defines which status codes to cache
	// Default: []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}
	CacheableStatusCodes []int

	// SkipIfAuthenticated skips caching if request has authentication
	// Useful for public pages that can be cached for guests only
	SkipIfAuthenticated bool

	// CachePredicate is a custom function to determine if response should be cached
	// If nil, default rules apply (method, status code)
	CachePredicate func(r *stdhttp.Request, statusCode int) bool

	// Tags for cache invalidation (e.g., "posts", "user:123")
	Tags []string
}

// DefaultCacheStrategy returns sensible defaults.
func DefaultCacheStrategy() CacheStrategy {
	return CacheStrategy{
		Enabled: true,
		TTL:     5 * time.Minute,
		VaryBy:  []string{},
		CacheableStatusCodes: []int{
			stdhttp.StatusOK,                  // 200
			stdhttp.StatusNonAuthoritativeInfo, // 203
			stdhttp.StatusNoContent,           // 204
			stdhttp.StatusPartialContent,      // 206
			stdhttp.StatusMultipleChoices,     // 300
			stdhttp.StatusMovedPermanently,    // 301
			stdhttp.StatusNotFound,            // 404 (cache 404s too!)
			stdhttp.StatusMethodNotAllowed,    // 405
			stdhttp.StatusGone,                // 410
		},
		SkipIfAuthenticated: false,
	}
}

// ShortLivedStrategy for frequently changing data (comments, reactions).
func ShortLivedStrategy() CacheStrategy {
	strategy := DefaultCacheStrategy()
	strategy.TTL = 30 * time.Second
	return strategy
}

// LongLivedStrategy for static content (categories, static pages).
func LongLivedStrategy() CacheStrategy {
	strategy := DefaultCacheStrategy()
	strategy.TTL = 1 * time.Hour
	return strategy
}

// PublicOnlyStrategy caches only for unauthenticated users.
func PublicOnlyStrategy() CacheStrategy {
	strategy := DefaultCacheStrategy()
	strategy.SkipIfAuthenticated = true
	return strategy
}
```

---

### 5. Main Cache Middleware

**File**: `internal/platform/cache/http/middleware.go`

```go
package http

import (
	"context"
	stdhttp "net/http"
	"strings"
	"time"

	"forum/internal/platform/cache"
	"forum/internal/platform/logger"
)

// CacheMiddleware provides HTTP response caching.
type CacheMiddleware struct {
	cache      cache.Cache
	logger     *logger.Logger
	keyBuilder *CacheKeyBuilder
	strategies map[string]CacheStrategy // route -> strategy
}

// NewCacheMiddleware creates a new cache middleware.
func NewCacheMiddleware(cache cache.Cache, lgr *logger.Logger) *CacheMiddleware {
	return &CacheMiddleware{
		cache:      cache,
		logger:     lgr,
		keyBuilder: NewCacheKeyBuilder("v1"),
		strategies: make(map[string]CacheStrategy),
	}
}

// RegisterRoute registers a caching strategy for a specific route pattern.
func (m *CacheMiddleware) RegisterRoute(pattern string, strategy CacheStrategy) {
	m.strategies[pattern] = strategy
	m.logger.Debug("Registered cache strategy",
		logger.String("pattern", pattern),
		logger.Duration("ttl", strategy.TTL),
		logger.Bool("enabled", strategy.Enabled))
}

// Middleware returns the HTTP middleware handler.
func (m *CacheMiddleware) Middleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		// 1. Check if route has caching strategy
		strategy, found := m.findStrategy(r.URL.Path)
		if !found || !strategy.Enabled {
			// No caching for this route
			next.ServeHTTP(w, r)
			return
		}

		// 2. Only cache GET and HEAD requests
		if r.Method != stdhttp.MethodGet && r.Method != stdhttp.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		// 3. Skip if authenticated and strategy says so
		if strategy.SkipIfAuthenticated && m.isAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}

		// 4. Build cache key
		cacheKey := m.keyBuilder.BuildKey(r, strategy.VaryBy)

		// 5. Try to get from cache
		cachedData, err := m.cache.Get(r.Context(), cacheKey)
		if err == nil {
			// Cache hit - deserialize and serve
			if cachedResp, err := UnmarshalCachedResponse(cachedData); err == nil {
				if !cachedResp.IsExpired() {
					m.logger.Debug("Cache hit",
						logger.String("path", r.URL.Path),
						logger.String("key", cacheKey))
					
					_ = cachedResp.WriteTo(w)
					return
				}
			}
		}

		// 6. Cache miss - capture response
		m.logger.Debug("Cache miss",
			logger.String("path", r.URL.Path),
			logger.String("key", cacheKey))

		capture := NewResponseCapture(w)
		next.ServeHTTP(capture, r)

		// 7. Check if response should be cached
		if m.shouldCache(r, capture.StatusCode(), strategy) {
			m.cacheResponse(r.Context(), cacheKey, capture, strategy)
		}
	})
}

// findStrategy finds the caching strategy for a given path.
func (m *CacheMiddleware) findStrategy(path string) (CacheStrategy, bool) {
	// Exact match first
	if strategy, ok := m.strategies[path]; ok {
		return strategy, true
	}

	// Prefix match (e.g., "/api/posts" matches "/api/posts/123")
	for pattern, strategy := range m.strategies {
		if strings.HasPrefix(path, pattern) {
			return strategy, true
		}
	}

	return CacheStrategy{}, false
}

// isAuthenticated checks if request has authentication.
func (m *CacheMiddleware) isAuthenticated(r *stdhttp.Request) bool {
	// Check for session cookie
	if _, err := r.Cookie("session_token"); err == nil {
		return true
	}

	// Check for Authorization header
	if r.Header.Get("Authorization") != "" {
		return true
	}

	return false
}

// shouldCache determines if a response should be cached.
func (m *CacheMiddleware) shouldCache(r *stdhttp.Request, statusCode int, strategy CacheStrategy) bool {
	// Custom predicate takes precedence
	if strategy.CachePredicate != nil {
		return strategy.CachePredicate(r, statusCode)
	}

	// Check if status code is cacheable
	for _, code := range strategy.CacheableStatusCodes {
		if statusCode == code {
			return true
		}
	}

	return false
}

// cacheResponse stores a response in cache.
func (m *CacheMiddleware) cacheResponse(ctx context.Context, key string, capture *ResponseCapture, strategy CacheStrategy) {
	// Create cached response
	cachedResp := NewCachedResponse(capture, strategy.TTL)

	// Serialize
	data, err := cachedResp.Marshal()
	if err != nil {
		m.logger.Error("Failed to marshal cached response",
			logger.String("key", key),
			logger.Error(err))
		return
	}

	// Store in cache (async, don't block)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := m.cache.Set(ctx, key, data, strategy.TTL); err != nil {
			m.logger.Error("Failed to store in cache",
				logger.String("key", key),
				logger.Error(err))
		} else {
			m.logger.Debug("Cached response",
				logger.String("key", key),
				logger.Int("size", len(data)))
		}

		// Store tags for invalidation
		m.storeTags(ctx, key, strategy.Tags)
	}()
}

// storeTags associates cache keys with tags for bulk invalidation.
func (m *CacheMiddleware) storeTags(ctx context.Context, key string, tags []string) {
	for _, tag := range tags {
		tagKey := "cache:tag:" + tag
		// Append key to tag's key list (simplified - in production use sets)
		_ = m.cache.Set(ctx, tagKey, []byte(key), 24*time.Hour)
	}
}
```

**Key Features:**

- **Strategy-based**: Each route defines its own caching behavior
- **Async storage**: Caching doesn't block response delivery
- **Flexible matching**: Exact and prefix-based route matching
- **Auth-aware**: Can skip caching for authenticated requests
- **Debug logging**: Track cache hits/misses

---

## Integration & Usage

### 1. Wire Integration

**File**: `cmd/forum/wire/app.go`

```go
// Initialize cache middleware (after cache, before routes)
cacheMiddleware := cachehttp.NewCacheMiddleware(cache, lgr)

// Register caching strategies for specific routes
registerCacheStrategies(cacheMiddleware)

// Add to middleware stack
server.Use(httpserver.Recovery(lgr))
server.Use(httpserver.Logger(lgr))
server.Use(cacheMiddleware.Middleware) // Add here, before auth
server.Use(httpserver.CORS())
```

**New file**: `cmd/forum/wire/cache_strategies.go`

```go
package wire

import (
	"time"
	cachehttp "forum/internal/platform/cache/http"
)

// registerCacheStrategies defines caching behavior for routes.
func registerCacheStrategies(m *cachehttp.CacheMiddleware) {
	// Homepage - cache for guests only
	m.RegisterRoute("/", cachehttp.CacheStrategy{
		Enabled:             true,
		TTL:                 5 * time.Minute,
		SkipIfAuthenticated: true,
		Tags:                []string{"homepage", "posts"},
	})

	// Public post listings - cache for all users
	m.RegisterRoute("/api/posts", cachehttp.CacheStrategy{
		Enabled: true,
		TTL:     2 * time.Minute,
		VaryBy:  []string{"Accept"}, // Vary by response format
		Tags:    []string{"posts"},
	})

	// Individual post - cache with vary by user
	m.RegisterRoute("/api/posts/", cachehttp.CacheStrategy{
		Enabled: true,
		TTL:     3 * time.Minute,
		Tags:    []string{"posts"},
	})

	// Comments - short TTL (frequently updated)
	m.RegisterRoute("/api/comments", cachehttp.ShortLivedStrategy())

	// Categories - long TTL (rarely change)
	m.RegisterRoute("/api/categories", cachehttp.CacheStrategy{
		Enabled: true,
		TTL:     1 * time.Hour,
		Tags:    []string{"categories"},
	})

	// User profiles - don't cache (personalized)
	m.RegisterRoute("/api/users/me", cachehttp.CacheStrategy{
		Enabled: false, // Explicitly disable
	})

	// Static pages - very long TTL
	m.RegisterRoute("/static/", cachehttp.CacheStrategy{
		Enabled: true,
		TTL:     24 * time.Hour,
		Tags:    []string{"static"},
	})
}
```

### 2. Cache Invalidation Helper

**File**: `internal/platform/cache/http/invalidator.go`

```go
package http

import (
	"context"
	"fmt"
	"forum/internal/platform/cache"
	"forum/internal/platform/logger"
)

// CacheInvalidator provides methods to invalidate cached responses.
type CacheInvalidator struct {
	cache  cache.Cache
	logger *logger.Logger
}

// NewCacheInvalidator creates a cache invalidator.
func NewCacheInvalidator(c cache.Cache, lgr *logger.Logger) *CacheInvalidator {
	return &CacheInvalidator{
		cache:  c,
		logger: lgr,
	}
}

// InvalidatePath removes cache entries for a specific path.
func (ci *CacheInvalidator) InvalidatePath(ctx context.Context, path string) error {
	pattern := fmt.Sprintf("http:cache:*:%s*", path)
	if err := ci.cache.DeletePattern(ctx, pattern); err != nil {
		ci.logger.Error("Failed to invalidate path",
			logger.String("path", path),
			logger.Error(err))
		return err
	}

	ci.logger.Debug("Invalidated cache path", logger.String("path", path))
	return nil
}

// InvalidateTag removes all cache entries with a specific tag.
func (ci *CacheInvalidator) InvalidateTag(ctx context.Context, tag string) error {
	// Find all keys with this tag
	tagKey := "cache:tag:" + tag
	// In production, store tag->keys mapping in cache
	// For now, use pattern matching
	pattern := fmt.Sprintf("http:cache:*")
	
	if err := ci.cache.DeletePattern(ctx, pattern); err != nil {
		ci.logger.Error("Failed to invalidate tag",
			logger.String("tag", tag),
			logger.Error(err))
		return err
	}

	ci.logger.Debug("Invalidated cache tag", logger.String("tag", tag))
	return nil
}

// InvalidateAll clears all HTTP cache entries.
func (ci *CacheInvalidator) InvalidateAll(ctx context.Context) error {
	pattern := "http:cache:*"
	if err := ci.cache.DeletePattern(ctx, pattern); err != nil {
		ci.logger.Error("Failed to invalidate all cache", logger.Error(err))
		return err
	}

	ci.logger.Info("Invalidated all HTTP cache")
	return nil
}
```

### 3. Using Invalidation in Handlers

**Example**: `internal/modules/post/adapters/http_handler.go`

```go
type HTTPHandler struct {
	service     ports.PostService
	invalidator *cachehttp.CacheInvalidator // Add this
	logger      *logger.Logger
}

func NewHTTPHandler(
	service ports.PostService,
	invalidator *cachehttp.CacheInvalidator,
	lgr *logger.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		service:     service,
		invalidator: invalidator,
		logger:      lgr,
	}
}

// CreatePost creates a new post and invalidates related caches.
func (h *HTTPHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// ... create post logic ...

	// Invalidate post listings cache
	if h.invalidator != nil {
		_ = h.invalidator.InvalidateTag(r.Context(), "posts")
		_ = h.invalidator.InvalidatePath(r.Context(), "/api/posts")
	}

	// Return response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// UpdatePost updates a post and invalidates its cache.
func (h *HTTPHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("id")
	
	// ... update logic ...

	// Invalidate specific post and listings
	if h.invalidator != nil {
		_ = h.invalidator.InvalidatePath(r.Context(), "/api/posts/"+postID)
		_ = h.invalidator.InvalidateTag(r.Context(), "posts")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}
```

---

## Testing Strategy

### 1. Unit Tests

**File**: `internal/platform/cache/http/middleware_test.go`

```go
package http_test

import (
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"forum/internal/platform/cache/inmemory"
	cachehttp "forum/internal/platform/cache/http"
	"forum/internal/platform/logger"
)

func TestCacheMiddleware_CacheHit(t *testing.T) {
	// Setup
	cache := inmemory.NewInMemoryCache(100, 5*time.Minute)
	defer cache.Close()

	lgr := logger.New(logger.Config{Level: "debug"})
	middleware := cachehttp.NewCacheMiddleware(cache, lgr)

	// Register strategy
	middleware.RegisterRoute("/test", cachehttp.CacheStrategy{
		Enabled: true,
		TTL:     1 * time.Minute,
	})

	// Handler that counts calls
	callCount := 0
	handler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		callCount++
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Wrap with middleware
	cachedHandler := middleware.Middleware(handler)

	// First request - cache miss
	req1 := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec1, req1)

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	// Second request - should be cache hit
	req2 := httptest.NewRequest(stdhttp.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Errorf("Expected 1 call (cached), got %d", callCount)
	}

	// Verify cache headers
	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("Expected X-Cache: HIT, got %s", rec2.Header().Get("X-Cache"))
	}
}

func TestCacheMiddleware_SkipAuthenticated(t *testing.T) {
	cache := inmemory.NewInMemoryCache(100, 5*time.Minute)
	defer cache.Close()

	lgr := logger.New(logger.Config{Level: "debug"})
	middleware := cachehttp.NewCacheMiddleware(cache, lgr)

	middleware.RegisterRoute("/profile", cachehttp.CacheStrategy{
		Enabled:             true,
		TTL:                 1 * time.Minute,
		SkipIfAuthenticated: true,
	})

	callCount := 0
	handler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		callCount++
		w.WriteHeader(stdhttp.StatusOK)
	})

	cachedHandler := middleware.Middleware(handler)

	// Request with session cookie - should NOT cache
	req := httptest.NewRequest(stdhttp.MethodGet, "/profile", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "session_token", Value: "abc123"})
	
	rec1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec1, req)

	rec2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec2, req)

	// Should call handler twice (not cached)
	if callCount != 2 {
		t.Errorf("Expected 2 calls (not cached), got %d", callCount)
	}
}

func TestCacheMiddleware_VaryByHeaders(t *testing.T) {
	cache := inmemory.NewInMemoryCache(100, 5*time.Minute)
	defer cache.Close()

	lgr := logger.New(logger.Config{Level: "debug"})
	middleware := cachehttp.NewCacheMiddleware(cache, lgr)

	middleware.RegisterRoute("/api/data", cachehttp.CacheStrategy{
		Enabled: true,
		TTL:     1 * time.Minute,
		VaryBy:  []string{"Accept-Language"},
	})

	responses := map[string]string{
		"en": "Hello",
		"es": "Hola",
	}

	handler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		lang := r.Header.Get("Accept-Language")
		w.Write([]byte(responses[lang]))
	})

	cachedHandler := middleware.Middleware(handler)

	// Request in English
	req1 := httptest.NewRequest(stdhttp.MethodGet, "/api/data", nil)
	req1.Header.Set("Accept-Language", "en")
	rec1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec1, req1)

	// Request in Spanish - should be different cache entry
	req2 := httptest.NewRequest(stdhttp.MethodGet, "/api/data", nil)
	req2.Header.Set("Accept-Language", "es")
	rec2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(rec2, req2)

	if rec1.Body.String() == rec2.Body.String() {
		t.Error("Expected different responses for different languages")
	}
}
```

### 2. Integration Tests

**File**: `tests/integration/http_cache_test.go`

```go
package integration

import (
	"context"
	stdhttp "net/http"
	"testing"
	"time"

	"forum/internal/platform/cache/inmemory"
	cachehttp "forum/internal/platform/cache/http"
)

func TestHTTPCache_PostListings(t *testing.T) {
	// Setup test server with cache middleware
	cache := inmemory.NewInMemoryCache(1000, 10*time.Minute)
	defer cache.Close()

	// ... initialize full app with cache middleware ...

	// Create a post
	post := createTestPost(t, "Test Post")

	// First request - should hit database
	start := time.Now()
	resp1 := httpGet(t, "/api/posts")
	duration1 := time.Since(start)

	// Second request - should be cached (faster)
	start = time.Now()
	resp2 := httpGet(t, "/api/posts")
	duration2 := time.Since(start)

	// Cached response should be faster
	if duration2 >= duration1 {
		t.Logf("Warning: Cached response not faster (duration1=%v, duration2=%v)",
			duration1, duration2)
	}

	// Check cache header
	if resp2.Header.Get("X-Cache") != "HIT" {
		t.Error("Expected cache hit")
	}

	// Update post - should invalidate cache
	updateTestPost(t, post.ID, "Updated Post")

	// Next request should be fresh (cache invalidated)
	resp3 := httpGet(t, "/api/posts")
	if resp3.Header.Get("X-Cache") == "HIT" {
		t.Error("Expected cache miss after invalidation")
	}
}
```

---

## Performance Considerations

### Memory Usage

**Cached Response Size Estimation:**

```text
Typical forum response sizes:
- Post listing (20 items):   ~50KB JSON
- Single post with comments:  ~30KB JSON
- User profile:                ~5KB JSON
- Category list:               ~2KB JSON

With 1000 cached responses:
- Average: ~25KB per response
- Total: ~25MB in cache
- With overhead: ~35MB
```

**Memory Limits:**

- **In-memory cache**: Set `CACHE_MAX_ENTRIES` based on available RAM
- **Redis**: Configure `maxmemory` and eviction policy (`allkeys-lru`)
- Monitor memory usage and adjust TTLs accordingly

### Response Time Improvements

**Expected Performance Gains:**

| Scenario | Without Cache | With Cache | Improvement |
|----------|--------------|------------|-------------|
| Post listing (DB query) | 50ms | 2ms | 25x faster |
| Post with joins | 100ms | 2ms | 50x faster |
| Static category list | 20ms | 1ms | 20x faster |
| 404 pages | 30ms | 1ms | 30x faster |

**Best Cases for Caching:**

- ✅ Public content viewed by many users
- ✅ Expensive database queries with joins
- ✅ Aggregated/computed data (counts, rankings)
- ✅ Static or rarely-changing content
- ✅ High-traffic endpoints

**Poor Cases for Caching:**

- ❌ User-specific personalized content
- ❌ Real-time data (live notifications)
- ❌ Content that changes per request
- ❌ Low-traffic endpoints (cache overhead not worth it)
- ❌ Sensitive data requiring fresh reads

### Cache Key Explosion Prevention

**Problem**: Too many unique cache keys waste memory.

**Solutions:**

1. **Limit pagination depth**: Only cache first N pages
2. **Ignore certain query params**: Skip tracking/analytics params
3. **Normalize query params**: Sort and lowercase
4. **Set max entries**: Use LRU eviction when full
5. **Monitor key count**: Alert if growing too fast

**Example - Selective Query Param Caching:**

```go
// Only cache first 5 pages
func (m *CacheMiddleware) shouldCacheQuery(r *http.Request) bool {
	page := r.URL.Query().Get("page")
	if page != "" {
		pageNum, _ := strconv.Atoi(page)
		if pageNum > 5 {
			return false // Don't cache deep pages
		}
	}
	return true
}
```

---

## Cache Strategy Matrix

### Recommended TTLs by Route

| Route | TTL | Reason | Skip Auth? |
|-------|-----|--------|------------|
| `/` (homepage) | 5 min | Frequently updated | Yes (guests) |
| `/api/posts` | 2 min | Moderate updates | No |
| `/api/posts/{id}` | 3 min | Individual posts | No |
| `/api/comments` | 30 sec | Real-time feel | No |
| `/api/categories` | 1 hour | Rarely change | No |
| `/api/users/{id}` | 5 min | Profile views | No |
| `/api/users/me` | Disabled | Always fresh | N/A |
| `/static/*` | 24 hours | Never changes | No |
| `/api/notifications` | Disabled | Real-time | N/A |

### Invalidation Strategy by Action

| Action | Invalidate | Method |
|--------|-----------|--------|
| Create post | Post listings, homepage | Tag: `posts` |
| Update post | Specific post, listings | Path + Tag |
| Delete post | Specific post, listings | Path + Tag |
| Create comment | Post page, comment list | Path: `/api/posts/{id}` |
| Like/dislike | Post/comment page | Path |
| Update user profile | User profile | Path: `/api/users/{id}` |
| Create category | Category list | Tag: `categories` |

---

## Monitoring & Observability

### 1. Cache Metrics Endpoint

**File**: `internal/platform/cache/http/stats.go`

```go
package http

import (
	"encoding/json"
	stdhttp "net/http"
	"sync/atomic"
	"time"
)

// CacheStats tracks cache performance metrics.
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	HitRate     float64 `json:"hit_rate"`
	TotalKeys   int64   `json:"total_keys"`
	LastReset   string  `json:"last_reset"`
	
	hitCounter  int64
	missCounter int64
	resetTime   time.Time
}

// NewCacheStats creates a new stats tracker.
func NewCacheStats() *CacheStats {
	return &CacheStats{
		resetTime: time.Now(),
		LastReset: time.Now().Format(time.RFC3339),
	}
}

// RecordHit increments hit counter.
func (cs *CacheStats) RecordHit() {
	atomic.AddInt64(&cs.hitCounter, 1)
}

// RecordMiss increments miss counter.
func (cs *CacheStats) RecordMiss() {
	atomic.AddInt64(&cs.missCounter, 1)
}

// GetStats returns current statistics.
func (cs *CacheStats) GetStats() CacheStats {
	hits := atomic.LoadInt64(&cs.hitCounter)
	misses := atomic.LoadInt64(&cs.missCounter)
	
	total := hits + misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return CacheStats{
		Hits:      hits,
		Misses:    misses,
		HitRate:   hitRate,
		LastReset: cs.resetTime.Format(time.RFC3339),
	}
}

// Reset clears statistics.
func (cs *CacheStats) Reset() {
	atomic.StoreInt64(&cs.hitCounter, 0)
	atomic.StoreInt64(&cs.missCounter, 0)
	cs.resetTime = time.Now()
	cs.LastReset = cs.resetTime.Format(time.RFC3339)
}

// StatsHandler returns HTTP handler for cache stats.
func (m *CacheMiddleware) StatsHandler() stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		stats := m.stats.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
```

**Access stats**: `GET /admin/cache/stats`

**Response:**
```json
{
  "hits": 15234,
  "misses": 2103,
  "hit_rate": 87.9,
  "total_keys": 456,
  "last_reset": "2025-11-06T10:00:00Z"
}
```

### 2. Logging

All cache operations should log at appropriate levels:

```go
// Cache hit (debug level)
m.logger.Debug("Cache hit",
	logger.String("path", r.URL.Path),
	logger.String("key", cacheKey),
	logger.Duration("age", time.Since(cachedResp.CachedAt)))

// Cache miss (debug level)
m.logger.Debug("Cache miss",
	logger.String("path", r.URL.Path),
	logger.String("key", cacheKey))

// Cache error (error level)
m.logger.Error("Cache operation failed",
	logger.String("operation", "Set"),
	logger.String("key", cacheKey),
	logger.Error(err))

// Invalidation (info level)
m.logger.Info("Cache invalidated",
	logger.String("tag", "posts"),
	logger.Int("keys_removed", count))
```

---

## Production Deployment

### Configuration

**Environment variables:**

```bash
# Enable HTTP caching
HTTP_CACHE_ENABLED=true

# Cache backend (inmemory or redis - must have cache enabled from original proposal)
CACHE_BACKEND=redis
CACHE_DEFAULT_TTL=5m

# Redis settings
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Docker Compose Update

**File**: `docker-compose.yml`

```yaml
version: '3.8'

services:
  forum:
    build: .
    ports:
      - "8080:8080"
    environment:
      - HTTP_CACHE_ENABLED=true
      - CACHE_BACKEND=redis
      - CACHE_DEFAULT_TTL=5m
      - REDIS_ADDR=redis:6379
    depends_on:
      redis:
        condition: service_healthy

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  redis_data:
```

### Health Checks

Add cache health check to application health endpoint:

```go
// GET /health
func HealthCheck(cache cache.Cache) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		health := map[string]string{
			"status": "healthy",
		}

		// Check cache
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := cache.Ping(ctx); err != nil {
			health["cache"] = "unhealthy: " + err.Error()
			health["status"] = "degraded"
		} else {
			health["cache"] = "healthy"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}
}
```

---

## Migration Plan

### Phase 1: Infrastructure (Week 1)

- [ ] Implement cache key builder
- [ ] Implement response capture and storage
- [ ] Write unit tests for core components
- [ ] Add configuration support

### Phase 2: Middleware (Week 2)

- [ ] Implement cache middleware
- [ ] Add strategy configuration
- [ ] Implement cache invalidator
- [ ] Write middleware unit tests

### Phase 3: Integration (Week 3)

- [ ] Wire middleware into application
- [ ] Register cache strategies for core routes
- [ ] Add invalidation to handlers (posts, comments)
- [ ] Integration testing

### Phase 4: Monitoring & Tuning (Week 4)

- [ ] Add metrics collection
- [ ] Implement stats endpoint
- [ ] Add logging and observability
- [ ] Performance testing
- [ ] Tune TTLs based on metrics

---

## Comparison: Repository vs HTTP Caching

| Aspect | Repository Cache | HTTP Cache (This Proposal) |
|--------|-----------------|---------------------------|
| **Layer** | Data access | HTTP response |
| **Granularity** | Individual entities, queries | Complete responses |
| **Use case** | Reusable data across endpoints | Complete page/API responses |
| **Invalidation** | Entity-level | URL/tag-based |
| **Complexity** | Higher (per-repository) | Lower (middleware) |
| **Benefit** | Reduces DB load | Reduces CPU + DB load |
| **Best for** | Backend optimization | Frontend performance |

**Recommendation**: Use **both** strategies together:

1. **Repository caching** for data-heavy operations (user lookups, session validation)
2. **HTTP caching** for frequently accessed endpoints (homepage, listings)

This layered approach provides maximum performance gain.

---

## Conclusion

This HTTP cache middleware proposal provides:

- ✅ **Cache-agnostic**: Works with any `cache.Cache` implementation
- ✅ **Flexible strategies**: Per-route configuration with sensible defaults
- ✅ **HTTP-compliant**: Respects HTTP semantics and best practices
- ✅ **Selective caching**: Opt-in per route, skip authenticated users
- ✅ **Easy invalidation**: Path and tag-based cache clearing
- ✅ **Observable**: Built-in metrics and logging
- ✅ **Non-blocking**: Cache failures don't break functionality
- ✅ **Testable**: Clean interfaces, easy to mock and test

### Key Benefits

1. **Performance**: 10-50x faster response times for cached endpoints
2. **Scalability**: Reduced database and CPU load
3. **User experience**: Faster page loads, especially for guests
4. **Cost savings**: Lower infrastructure costs due to reduced load
5. **Flexibility**: Easy to enable/disable per route

### Next Steps

1. Review and approve this proposal
2. Start with Phase 1 (infrastructure)
3. Test in development with in-memory cache
4. Roll out to production with Redis
5. Monitor metrics and tune TTLs
6. Gradually expand to more routes

### Questions for Discussion

- Which routes should be cached first? (Suggest: homepage, post listings)
- Preferred cache backend for production? (Suggest: Redis for multi-instance)
- Acceptable cache staleness for different content types?
- Should we cache 404 responses? (Suggest: Yes, with shorter TTL)
- Admin interface for cache management needed?
