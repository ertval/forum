# Caching Strategy Proposal for Forum Application

## Executive Summary

This document proposes a flexible, idiomatic caching layer for the forum application that adheres to our hexagonal architecture principles. The design supports multiple backend implementations (in-memory, Redis, Memcached) while maintaining clean boundaries and testability.

## Design Philosophy

### Core Principles

- **Interface-first**: Cache abstraction as a platform service
- **Backend-agnostic**: Support multiple cache implementations via adapters
- **Non-invasive**: Caching is a cross-cutting concern, not a domain concern
- **Decorator pattern**: Wrap repository implementations with caching logic
- **Fail-safe**: Cache failures never break application functionality
- **Observable**: Built-in metrics and logging for cache performance

### Why NOT a Module?

Caching is **infrastructure**, not a business domain. Unlike `auth`, `post`, or `user`, caching has:

- No domain entities
- No business rules
- No use cases specific to forum logic
- Cross-cutting nature (affects multiple modules)

**Decision**: Implement caching as a **platform service** in `internal/platform/cache/`.

---

## Architecture Overview

### Directory Structure

```text
internal/platform/cache/
├── cache.go                    # Core interface and types
├── decorator.go                # Repository decorator pattern
├── metrics.go                  # Cache performance metrics
├── inmemory/
│   └── inmemory.go            # In-memory cache implementation
├── redis/
│   └── redis.go               # Redis cache implementation
└── memcached/
│   └── memcached.go           # Memcached implementation (optional)
```

### Component Diagram

```
┌───────────────────────────────────────────────────────────┐
│                    HTTP Handlers                          │
│                   (Input Adapters)                        │
└────────────────────────┬──────────────────────────────────┘
                         │
                         ▼
┌───────────────────────────────────────────────────────────┐
│              Application Services                         │
│          (auth, post, comment services)                   │
└────────────────────────┬──────────────────────────────────┘
                         │
                         ▼
┌───────────────────────────────────────────────────────────┐
│            Cached Repository Decorator                    │
│  ┌─────────────────────────────────────────────────┐     │
│  │  Cache Check → Hit: Return                      │     │
│  │            → Miss: Call Repository → Cache It   │     │
│  └─────────────────────────────────────────────────┘     │
└────────────────────────┬──────────────────────────────────┘
                         │
        ┌────────────────┴────────────────┐
        ▼                                  ▼
┌──────────────────┐            ┌──────────────────┐
│  Cache Backend   │            │  Repository      │
│  (Redis/Memory)  │            │  (SQLite)        │
└──────────────────┘            └──────────────────┘
```

---

## Implementation Details

### 1. Core Cache Interface

**File**: `internal/platform/cache/cache.go`

```go
// Package cache provides a generic caching abstraction for the forum application.
// It supports multiple backend implementations while maintaining a consistent interface.
package cache

import (
 "context"
 "time"
)

// Cache defines the contract for cache implementations.
// All cache operations are fail-safe - errors are logged but don't break functionality.
type Cache interface {
	// Get retrieves a value from cache by key.
	// Returns ErrCacheMiss if key doesn't exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with TTL.
	// If ttl is 0, uses default TTL configured for the cache.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key from cache.
	Delete(ctx context.Context, key string) error

	// DeletePattern removes all keys matching a pattern (e.g., "user:*").
	// Not all backends support this efficiently (e.g., Redis SCAN vs full scan).
	DeletePattern(ctx context.Context, pattern string) error

    // TagKey associates a cache key with one or more tags for bulk invalidation.
    // Used primarily by HTTP caching layer for efficient cache invalidation.
    TagKey(ctx context.Context, key string, tags []string) error

    // UntagKey removes tag associations for a cache key.
    UntagKey(ctx context.Context, key string, tags []string) error

    // DeleteByTag removes all cache entries associated with a specific tag.
    DeleteByTag(ctx context.Context, tag string) error

    // GetKeysByTag returns all cache keys associated with a specific tag.
    // Primarily used for debugging and monitoring.
    GetKeysByTag(ctx context.Context, tag string) ([]string, error)

    // Exists checks if a key exists without retrieving its value.
    Exists(ctx context.Context, key string) (bool, error)

    // Clear removes all keys from the cache.
    // Use with caution in production.
    Clear(ctx context.Context) error

    // Ping checks cache backend health.
    Ping(ctx context.Context) error

    // Close cleanly shuts down cache connections.
    Close() error
}

// Common errors
var (
	ErrCacheMiss     = errors.New("cache: key not found")
	ErrCacheDisabled = errors.New("cache: caching is disabled")
)

// Config holds unified cache configuration for both repository and HTTP caching layers.
// This config is stored in internal/platform/config/cache.go
type Config struct {
    // Global cache settings
    Enabled    bool          // Master switch for all caching
    Backend    string        // "inmemory", "redis", "memcached", "none"
    DefaultTTL time.Duration // Default TTL for cached items

    // Layer-specific enablement
    RepositoryEnabled bool // Enable repository-level caching
    HTTPEnabled       bool // Enable HTTP-level caching

    // Backend-specific settings
    RedisAddr     string   // Redis server address
    RedisPassword string   // Redis password (optional)
    RedisDB       int      // Redis database number
    MemcachedServers []string // Memcached server list
    MaxEntries    int      // Max items in memory cache (LRU eviction)

    // HTTP caching strategies (only used when HTTPEnabled=true)
    HTTPRoutes map[string]CacheStrategy // Route-specific caching strategies

    // Observability
    EnableMetrics bool // Enable cache metrics collection
}

// CacheStrategy defines how HTTP routes should be cached.
// Used only when HTTPEnabled=true.
type CacheStrategy struct {
    Enabled      bool          // Enable caching for this route
    TTL          time.Duration // TTL for cached responses
    Tags         []string      // Tags for bulk invalidation
    SkipAuth     bool          // Cache responses for unauthenticated users only
    VaryHeaders  []string      // Headers to vary cache by (e.g., "Accept-Language")
}

// KeyBuilder helps construct consistent cache keys.
type KeyBuilder struct {
	prefix string
}

// NewKeyBuilder creates a key builder with a module prefix.
func NewKeyBuilder(module string) *KeyBuilder {
	return &KeyBuilder{prefix: module}
}

// Build creates a cache key from parts.
// Example: KeyBuilder("post").Build("id", "123") -> "post:id:123"
func (kb *KeyBuilder) Build(parts ...string) string {
	key := kb.prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// CacheMetrics provides unified metrics for both repository and HTTP caching layers.
type CacheMetrics struct {
    Layer            string        // "repository" or "http"
    Hits             int64         // Cache hits
    Misses           int64         // Cache misses
    Sets             int64         // Items stored in cache
    Deletes          int64         // Items deleted from cache
    Errors           int64         // Cache operation errors
    HitRate          float64       // Calculated: hits / (hits + misses)
    AvgGetLatency    time.Duration // Average time for cache gets

    // HTTP-specific metrics
    ResponsesServed  int64 // Complete responses served from cache
    BytesServed      int64 // Total bytes served from HTTP cache

    // Tag-specific metrics
    TagInvalidations int64 // Number of tag-based invalidations
}
```

### 2. Repository Decorator Pattern

**File**: `internal/platform/cache/decorator.go`

```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CachedRepository wraps any repository with caching logic.
// This is a generic pattern that can be applied to any repository interface.

// Example: Cached User Repository
type CachedUserRepository struct {
	repo  ports.UserRepository  // Original repository
	cache Cache                  // Cache backend
	kb    *KeyBuilder           // Key builder for consistent keys
	ttl   time.Duration         // TTL for cached items
}

// NewCachedUserRepository creates a caching decorator for user repository.
func NewCachedUserRepository(repo ports.UserRepository, cache Cache, ttl time.Duration) *CachedUserRepository {
	return &CachedUserRepository{
		repo:  repo,
		cache: cache,
		kb:    NewKeyBuilder("user"),
		ttl:   ttl,
	}
}

// GetByID retrieves a user by ID, checking cache first.
func (r *CachedUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	// 1. Build cache key
	key := r.kb.Build("id", id)

	// 2. Try cache first
	data, err := r.cache.Get(ctx, key)
	if err == nil {
		// Cache hit - deserialize and return
		var user domain.User
		if err := json.Unmarshal(data, &user); err == nil {
			return &user, nil
		}
		// Deserialization failed - fall through to DB
	}

	// 3. Cache miss - fetch from repository
	user, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 4. Store in cache (async, don't block on cache errors)
	go func() {
		data, _ := json.Marshal(user)
		_ = r.cache.Set(context.Background(), key, data, r.ttl)
	}()

	return user, nil
}

// Update invalidates cache entries when user is updated.
func (r *CachedUserRepository) Update(ctx context.Context, user *domain.User) error {
	// 1. Update in database first
	if err := r.repo.Update(ctx, user); err != nil {
		return err
	}

	// 2. Invalidate cache entries
	_ = r.cache.Delete(ctx, r.kb.Build("id", user.ID))
	_ = r.cache.Delete(ctx, r.kb.Build("email", user.Email))
	_ = r.cache.Delete(ctx, r.kb.Build("username", user.Username))

	return nil
}

// Generic helper for list caching with pagination
func (r *CachedUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	key := r.kb.Build("list", fmt.Sprintf("limit:%d:offset:%d", limit, offset))

	// Check cache
	data, err := r.cache.Get(ctx, key)
	if err == nil {
		var users []*domain.User
		if err := json.Unmarshal(data, &users); err == nil {
			return users, nil
		}
	}

	// Fetch from DB
	users, err := r.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Cache result
	go func() {
		data, _ := json.Marshal(users)
		_ = r.cache.Set(context.Background(), key, data, r.ttl)
	}()

	return users, nil
}
```

### 3. In-Memory Cache Implementation

**File**: `internal/platform/cache/inmemory/inmemory.go`

```go
package inmemory

import (
	"context"
	"sync"
	"time"
	"forum/internal/platform/cache"
)

// InMemoryCache is a thread-safe LRU cache with TTL support.
// Suitable for single-instance deployments or development.
type InMemoryCache struct {
	mu         sync.RWMutex
	items      map[string]*item
	maxEntries int
	defaultTTL time.Duration
	janitor    *janitor // Background cleanup goroutine
}

type item struct {
	value      []byte
	expiration int64 // Unix timestamp (0 = no expiration)
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache(maxEntries int, defaultTTL time.Duration) *InMemoryCache {
	c := &InMemoryCache{
		items:      make(map[string]*item),
		maxEntries: maxEntries,
		defaultTTL: defaultTTL,
	}
	
	// Start background cleanup
	c.janitor = newJanitor(c, 1*time.Minute)
	
	return c
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, cache.ErrCacheMiss
	}

	// Check expiration
	if item.expiration > 0 && time.Now().Unix() > item.expiration {
		return nil, cache.ErrCacheMiss
	}

	return item.value, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity (simple LRU: remove oldest)
	if len(c.items) >= c.maxEntries {
		c.evictOldest()
	}

	if ttl == 0 {
		ttl = c.defaultTTL
	}

	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).Unix()
	}

	c.items[key] = &item{
		value:      value,
		expiration: exp,
	}

	return nil
}

func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

func (c *InMemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	// Simple prefix matching (not full glob)
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if matchesPattern(key, pattern) {
			delete(c.items, key)
		}
	}
	return nil
}

func (c *InMemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*item)
	return nil
}

func (c *InMemoryCache) Ping(ctx context.Context) error {
	return nil // Always healthy
}

func (c *InMemoryCache) Close() error {
	c.janitor.stop()
	return nil
}

// Background cleanup of expired items
type janitor struct {
	interval time.Duration
	stop     chan struct{}
}

func newJanitor(c *InMemoryCache, interval time.Duration) *janitor {
	j := &janitor{
		interval: interval,
		stop:     make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.deleteExpired()
			case <-j.stop:
				return
			}
		}
	}()

	return j
}

func (j *janitor) stop() {
	close(j.stop)
}

func (c *InMemoryCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().Unix()
	for key, item := range c.items {
		if item.expiration > 0 && now > item.expiration {
			delete(c.items, key)
		}
	}
}

func (c *InMemoryCache) evictOldest() {
	// Simple implementation: delete first key found
	// Production: use proper LRU with access tracking
	for key := range c.items {
		delete(c.items, key)
		return
	}
}

func matchesPattern(key, pattern string) bool {
	// Simple prefix match for patterns like "user:*"
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
	return key == pattern
}
```

### 4. Redis Cache Implementation

**File**: `internal/platform/cache/redis/redis.go`

```go
package redis

import (
	"context"
	"fmt"
	"time"
	"forum/internal/platform/cache"
	
	"github.com/redis/go-redis/v9"
)

// RedisCache wraps Redis client with our Cache interface.
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// NewRedisCache creates a new Redis cache instance.
func NewRedisCache(cfg cache.Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisCache{
		client:     client,
		defaultTTL: cfg.DefaultTTL,
	}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, cache.ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	return result, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.defaultTTL
	}
	
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	// Use SCAN to avoid blocking Redis (better than KEYS)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("redis delete pattern failed: %w", err)
		}
	}
	return iter.Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return result > 0, nil
}

func (r *RedisCache) Clear(ctx context.Context) error {
	if err := r.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("redis clear failed: %w", err)
	}
	return nil
}

func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
```

---

## Integration with Application

### Configuration Changes

**File**: `internal/platform/config/config.go`

Add to `Config` struct:

```go
// Cache configuration
Cache CacheConfig
```

Add new config section:

```go
// CacheConfig contains caching settings.
type CacheConfig struct {
	Enabled    bool          // Enable/disable caching globally
	Backend    string        // "inmemory", "redis", "memcached", "none"
	DefaultTTL time.Duration // Default TTL for cached items

	// Redis settings
	RedisAddr     string // Redis server address (e.g., "localhost:6379")
	RedisPassword string // Redis password (if required)
	RedisDB       int    // Redis database number

	// In-memory settings
	MaxEntries int // Maximum entries in memory cache (LRU)

	// Memcached settings (if implemented)
	MemcachedServers []string // Memcached server list
}
```

### Dependency Injection

**File**: `cmd/forum/wire/app.go`

```go
// 2. Initialize Cache (after database, before repositories)
cache, err := initCache(cfg, lgr)
if err != nil {
	return nil, fmt.Errorf("failed to initialize cache: %w", err)
}

// Add to App struct
type App struct {
	Server *httpserver.Server
	DB     *database.Connection
	Cache  cache.Cache  // Add this
	Logger *logger.Logger
}

// Cleanup cache on shutdown
func (a *App) Cleanup() error {
	a.Logger.Info("Cleaning up application resources")
	
	// Close cache
	if err := a.Cache.Close(); err != nil {
		a.Logger.Error("Failed to close cache", logger.Error(err))
	}
	
	// Close database
	if err := a.DB.Close(); err != nil {
		a.Logger.Error("Failed to close database connection", logger.Error(err))
		return err
	}
	return nil
}
```

**New file**: `cmd/forum/wire/cache.go`

```go
package wire

import (
	"fmt"
	"forum/internal/platform/cache"
	"forum/internal/platform/cache/inmemory"
	"forum/internal/platform/cache/redis"
	"forum/internal/platform/config"
	"forum/internal/platform/logger"
)

// initCache creates the appropriate cache backend based on configuration.
func initCache(cfg *config.Config, lgr *logger.Logger) (cache.Cache, error) {
	if !cfg.Cache.Enabled {
		lgr.Info("Cache disabled")
		return cache.NewNoOpCache(), nil
	}

	cacheConfig := cache.Config{
		Backend:          cfg.Cache.Backend,
		DefaultTTL:       cfg.Cache.DefaultTTL,
		RedisAddr:        cfg.Cache.RedisAddr,
		RedisPassword:    cfg.Cache.RedisPassword,
		RedisDB:          cfg.Cache.RedisDB,
		MaxEntries:       cfg.Cache.MaxEntries,
		MemcachedServers: cfg.Cache.MemcachedServers,
		EnableMetrics:    true,
	}

	switch cfg.Cache.Backend {
	case "inmemory":
		lgr.Info("Initializing in-memory cache", 
			logger.Int("maxEntries", cfg.Cache.MaxEntries),
			logger.Duration("defaultTTL", cfg.Cache.DefaultTTL))
		return inmemory.NewInMemoryCache(cfg.Cache.MaxEntries, cfg.Cache.DefaultTTL), nil

	case "redis":
		lgr.Info("Initializing Redis cache",
			logger.String("addr", cfg.Cache.RedisAddr),
			logger.Duration("defaultTTL", cfg.Cache.DefaultTTL))
		return redis.NewRedisCache(cacheConfig)

	case "memcached":
		// TODO: Implement memcached adapter
		return nil, fmt.Errorf("memcached backend not yet implemented")

	case "none":
		lgr.Info("Cache explicitly disabled (backend=none)")
		return cache.NewNoOpCache(), nil

	default:
		return nil, fmt.Errorf("unknown cache backend: %s", cfg.Cache.Backend)
	}
}
```

### Repository Wiring

**File**: `cmd/forum/wire/repositories.go`

```go
// Wrap repositories with caching decorators
func initRepositories(db *database.Connection, cache cache.Cache, cfg *config.Config) *Repositories {
	// Create base repositories
	baseUserRepo := userAdapters.NewSQLiteUserRepository(db.DB())
	basePostRepo := postAdapters.NewSQLitePostRepository(db.DB())
	// ... other repos

	// Wrap with caching if enabled
	var userRepo userPorts.UserRepository
	if cfg.Cache.Enabled {
		userRepo = cache.NewCachedUserRepository(baseUserRepo, cache, 5*time.Minute)
	} else {
		userRepo = baseUserRepo
	}

	var postRepo postPorts.PostRepository
	if cfg.Cache.Enabled {
		postRepo = cache.NewCachedPostRepository(basePostRepo, cache, 2*time.Minute)
	} else {
		postRepo = basePostRepo
	}

	return &Repositories{
		User: userRepo,
		Post: postRepo,
		// ...
	}
}
```

---

## Environment Variables

Add to `.env` or environment:

```bash
# Unified Caching Configuration
CACHE_ENABLED=true                    # Master switch for all caching
CACHE_BACKEND=redis                   # Options: inmemory, redis, memcached, none
CACHE_DEFAULT_TTL=5m                  # Default TTL for cached items

# Layer-specific enablement
CACHE_REPOSITORY_ENABLED=true         # Enable repository-level caching
CACHE_HTTP_ENABLED=true               # Enable HTTP-level caching

# Backend Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=                       # Optional
REDIS_DB=0                            # Redis database number
CACHE_MAX_ENTRIES=10000               # Maximum items in memory cache

# Observability
CACHE_ENABLE_METRICS=true             # Enable cache metrics collection
```

---

## Docker Compose with Redis

**File**: `docker-compose.yml`

```yaml
version: '3.8'

services:
  forum:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CACHE_ENABLED=true
      - CACHE_BACKEND=redis
      - CACHE_REPOSITORY_ENABLED=true
      - CACHE_HTTP_ENABLED=true
      - REDIS_ADDR=redis:6379
      - CACHE_ENABLE_METRICS=true
      - DATABASE_PATH=/data/forum.db
    volumes:
      - ./data:/data
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

volumes:
  redis_data:
```

---

## Caching Strategies by Module

### 1. User Module

**Cache Keys:**

- `user:id:{user_id}` - User by ID
- `user:email:{email}` - User by email
- `user:username:{username}` - User by username

**TTL**: 5 minutes (users don't change frequently)

**Invalidation**:
- On user update/delete: invalidate all user keys
- On role change: invalidate user keys

### 2. Post Module

**Cache Keys:**
- `post:id:{post_id}` - Single post
- `post:list:page:{page}:category:{category_id}` - Post listings
- `post:user:{user_id}:page:{page}` - User's posts
- `post:liked:{user_id}:page:{page}` - User's liked posts

**TTL**: 2-3 minutes (posts change more frequently)

**Invalidation**:
- On post create/update/delete: invalidate post and list keys
- On reaction change: invalidate post details (includes reaction counts)

### 3. Comment Module

**Cache Keys:**
- `comment:post:{post_id}:page:{page}` - Comments for a post

**TTL**: 1-2 minutes (comments are real-time)

**Invalidation**:
- On comment create/update/delete: invalidate comment list for that post

### 4. Reaction Module

**Cache Strategy**: Cache-aside for aggregated counts

**Cache Keys:**
- `reaction:post:{post_id}:counts` - Like/dislike counts for post
- `reaction:comment:{comment_id}:counts` - Like/dislike counts for comment

**TTL**: 1 minute

**Invalidation**:
- On reaction create/delete: invalidate counts for target entity

### 5. Session Module (Auth)

**Cache Keys:**
- `session:token:{token}` - Session by token
- `session:user:{user_id}` - Active session for user

**TTL**: Match session expiration (24 hours default)

**Invalidation**:
- On logout: delete session keys
- On new login: invalidate old session for that user

---

## Cache Warming Strategies

### On Application Startup

```go
// Optional: Pre-warm frequently accessed data
func warmCache(cache cache.Cache, repos *Repositories) error {
	// Cache popular categories
	categories, _ := repos.Category.List(context.Background())
	// ... cache them

	// Cache homepage posts
	posts, _ := repos.Post.List(context.Background(), 20, 0)
	// ... cache them

	return nil
}
```

### Scheduled Refresh

```go
// Background goroutine to refresh stale cache entries
func startCacheRefresh(cache cache.Cache, repos *Repositories, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			// Refresh popular items
			refreshPopularPosts(cache, repos)
		}
	}()
}
```

---

## Monitoring & Observability

### Cache Metrics Endpoint

**File**: `internal/modules/admin/adapters/http_handler.go` (or new admin module)

```go
// GET /admin/cache/stats
func (h *AdminHandler) CacheStats(w http.ResponseWriter, r *http.Request) {
	metrics := h.cache.GetMetrics()
	json.NewEncoder(w).Encode(metrics)
}
```

**Response:**
```json
{
  "hits": 15234,
  "misses": 1023,
  "hit_rate": 93.7,
  "sets": 1045,
  "deletes": 234,
  "errors": 2,
  "avg_get_latency_ms": 0.3
}
```

### Logging

All cache operations should log at debug level:

```go
lgr.Debug("Cache miss", 
	logger.String("key", key),
	logger.String("operation", "Get"))

lgr.Error("Cache operation failed",
	logger.String("key", key),
	logger.String("operation", "Set"),
	logger.Error(err))
```

---

## Testing Strategy

### Unit Tests

**File**: `internal/platform/cache/inmemory/inmemory_test.go`

```go
func TestInMemoryCache_SetGet(t *testing.T) {
	cache := inmemory.NewInMemoryCache(100, 5*time.Minute)
	defer cache.Close()

	ctx := context.Background()
	key := "test:key"
	value := []byte("test value")

	// Test Set
	err := cache.Set(ctx, key, value, 1*time.Minute)
	assert.NoError(t, err)

	// Test Get
	result, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, result)
}

func TestInMemoryCache_Expiration(t *testing.T) {
	cache := inmemory.NewInMemoryCache(100, 100*time.Millisecond)
	defer cache.Close()

	ctx := context.Background()
	key := "test:expiring"
	value := []byte("expires soon")

	cache.Set(ctx, key, value, 100*time.Millisecond)
	
	// Should exist immediately
	result, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, result)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)
	
	// Should be gone
	_, err = cache.Get(ctx, key)
	assert.ErrorIs(t, err, cache.ErrCacheMiss)
}
```

### Integration Tests

**File**: `tests/integration/cache_test.go`

```go
func TestCachedUserRepository(t *testing.T) {
	// Setup: DB + Cache
	db := setupTestDB(t)
	cache := inmemory.NewInMemoryCache(100, 5*time.Minute)
	
	baseRepo := adapters.NewSQLiteUserRepository(db)
	cachedRepo := cache.NewCachedUserRepository(baseRepo, cache, 1*time.Minute)

	// Create user
	user := &domain.User{ID: "1", Email: "test@example.com"}
	err := cachedRepo.Create(context.Background(), user)
	require.NoError(t, err)

	// First read - should hit DB and cache
	u1, err := cachedRepo.GetByID(context.Background(), "1")
	require.NoError(t, err)
	assert.Equal(t, user.Email, u1.Email)

	// Second read - should hit cache (verify by mocking or timing)
	u2, err := cachedRepo.GetByID(context.Background(), "1")
	require.NoError(t, err)
	assert.Equal(t, user.Email, u2.Email)

	// Update - should invalidate cache
	u1.Email = "updated@example.com"
	err = cachedRepo.Update(context.Background(), u1)
	require.NoError(t, err)

	// Next read should reflect update
	u3, err := cachedRepo.GetByID(context.Background(), "1")
	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", u3.Email)
}
```

---

## Performance Considerations

### When to Cache

**✅ Good candidates:**

- Frequently read, rarely written (user profiles, categories)
- Expensive queries (joins, aggregations)
- External API responses
- Session data
- Post/comment counts

**❌ Bad candidates:**

- Data that changes constantly (real-time notifications)
- Very large objects (images - use CDN instead)
- User-specific data with low reuse
- Write-heavy operations

### Cache TTL Guidelines

| Data Type | Recommended TTL | Reasoning |
|-----------|----------------|-----------|
| User profiles | 5-10 minutes | Rarely change |
| Posts | 2-5 minutes | Moderate updates |
| Comments | 1-2 minutes | More dynamic |
| Reaction counts | 30-60 seconds | Frequently updated |
| Session tokens | Match session lifetime | Must stay in sync |
| Static data (categories) | 1 hour | Almost never changes |

### Memory Considerations

**In-Memory Cache:**

- ~1KB per cached user
- ~2KB per cached post
- 10,000 entries ≈ 20-30MB memory
- Set `CACHE_MAX_ENTRIES` based on available RAM

**Redis:**

- Redis memory = dataset size + overhead (~10-20%)
- Use `maxmemory` policy (e.g., `allkeys-lru`)
- Monitor with `redis-cli INFO memory`

---

## Migration Plan

### Phase 1: Infrastructure Setup (Week 1)

1. Implement core cache interface and types
2. Implement in-memory cache
3. Add configuration support
4. Write unit tests

### Phase 2: Redis Integration (Week 2)

1. Implement Redis adapter
2. Update Docker Compose
3. Add connection pooling and retry logic
4. Write integration tests

### Phase 3: Repository Decorators (Week 3)

1. Create cached user repository
2. Create cached post repository
3. Create cached session repository
4. Update wire package

### Phase 4: Monitoring & Optimization (Week 4)

1. Add metrics collection
2. Add admin endpoints for cache stats
3. Implement cache warming
4. Performance testing and tuning

---

## Production Deployment Checklist

- [ ] Choose cache backend (Redis recommended for multi-instance)
- [ ] Set appropriate TTL values per data type
- [ ] Configure Redis persistence (AOF + RDB)
- [ ] Set up Redis monitoring (memory, hit rate, latency)
- [ ] Implement cache invalidation strategy
- [ ] Test cache failures (ensure fail-safe behavior)
- [ ] Document cache key patterns
- [ ] Set up cache alerts (high miss rate, connection failures)
- [ ] Plan for cache warm-up on deployment
- [ ] Test with production-like data volume

---

## Alternative: Cache-Aside vs Write-Through

**Current Proposal**: Cache-Aside (Lazy Loading)

- Read: Check cache → Miss → Read DB → Write cache
- Write: Update DB → Invalidate cache

**Alternative**: Write-Through

- Read: Check cache → Miss → Read DB → Write cache
- Write: Update cache → Update DB

**Recommendation**: Stick with **cache-aside** because:

1. Simpler implementation
2. Better for read-heavy forum workload
3. Handles cache failures gracefully
4. Less risk of cache/DB inconsistency

---

## Conclusion

This caching strategy provides:

- **Flexibility**: Multiple backend support (in-memory, Redis, Memcached)
- **Clean architecture**: Cache as platform service, not domain concern
- **Fail-safe**: Cache failures don't break functionality
- **Testability**: Mock-friendly interfaces
- **Observability**: Built-in metrics and logging
- **Scalability**: Easy to scale from in-memory to distributed cache

**Next Steps:**

1. Review this proposal with the team
2. Choose initial cache backend (recommend starting with in-memory)
3. Implement Phase 1 (core infrastructure)
4. Gradually roll out to modules (start with sessions/users)
5. Monitor and tune based on real-world performance

**Questions to Decide:**

- Which modules to cache first? (Recommend: auth sessions, then users, then posts)
- Redis vs in-memory for initial deployment?
- Cache metrics dashboard requirements?
- Cache invalidation: immediate vs eventual consistency?
