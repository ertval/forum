# Caching Platform Infrastructure Implementation Proposal for Forum Application

## Table of Contents
1. [Introduction](#introduction)
2. [Rationale for Platform Infrastructure Approach](#rationale-for-platform-infrastructure-approach)
3. [Architecture Overview](#architecture-overview)
4. [Integration with External Solutions](#integration-with-external-solutions)
5. [Implementation Examples](#implementation-examples)
6. [Usage in Business Modules](#usage-in-business-modules)
7. [Configuration](#configuration)
8. [Conclusion](#conclusion)

## Introduction

This proposal outlines the implementation of a caching layer for the forum application as a platform infrastructure component. The caching solution will be placed in the `internal/platform` directory, positioned as a cross-cutting concern that provides caching capabilities to multiple business modules while maintaining clean architectural boundaries.

## Rationale for Platform Infrastructure Approach

Caching is a classic example of a cross-cutting concern - a functionality that cuts across multiple business domains and is needed by the entire application. Here's why it should be positioned as platform infrastructure rather than as a business module:

1. **Cross-cutting Nature**: Caching is needed by multiple modules (posts, users, comments, sessions) rather than belonging to a specific business domain.

2. **Technical Concern**: Caching is primarily a technical concern focused on performance optimization rather than business logic.

3. **Separation of Concerns**: Business modules should focus on their specific domain logic rather than infrastructure concerns.

4. **Consistency**: Centralizing caching in platform infrastructure ensures consistent caching patterns across all modules.

5. **Reusability**: Other infrastructure needs (logging, error handling, validation) follow the same pattern in the platform layer.

6. **Dependency Direction**: Business modules depend on platform infrastructure, not the other way around.

7. **Hexagonal Architecture Alignment**: Caching serves as an adapter for performance optimization between the application core and external systems.

## Architecture Overview

The caching infrastructure will be added to the existing platform layer:

```
internal/platform/
├── cache/
│   ├── cache.go           # Cache interface and implementations
│   ├── in_memory.go       # In-memory cache implementation
│   ├── redis.go           # Redis cache implementation
│   └── middleware.go      # Optional caching middleware
├── config/                # Configuration management
├── database/              # SQLite connection & migrations
├── logger/                # Structured logging
├── httpserver/            # HTTP server & middleware
├── errors/                # Error handling
└── validator/             # Input validation
```

### Cache Interface Definition

```go
// internal/platform/cache/cache.go
package cache

import (
    "context"
    "time"
)

// Cache represents a generic caching interface
type Cache interface {
    // Get retrieves a value from cache
    Get(ctx context.Context, key string) (interface{}, error)
    
    // GetWithDefault retrieves a value from cache or returns default if not found
    GetWithDefault(ctx context.Context, key string, defaultValue interface{}) (interface{}, error)
    
    // Set stores a value in cache with optional TTL
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    
    // Delete removes a value from cache
    Delete(ctx context.Context, key string) error
    
    // Exists checks if a key exists in cache
    Exists(ctx context.Context, key string) (bool, error)
    
    // InvalidateByPattern invalidates cache entries by pattern matching (for Redis)
    InvalidateByPattern(ctx context.Context, pattern string) error
    
    // Clear clears all cache entries
    Clear(ctx context.Context) error
}

// TypedCache provides type-safe cache operations
type TypedCache interface {
    Cache
    
    // GetTyped retrieves a value from cache with type information
    GetTyped(ctx context.Context, key string, dest interface{}) error
    
    // SetTyped stores a value in cache with type information
    SetTyped(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
```

### In-Memory Cache Implementation

```go
// internal/platform/cache/in_memory.go
package cache

import (
    "context"
    "sync"
    "time"
)

type InMemoryCache struct {
    data      map[string]*cacheItem
    mutex     sync.RWMutex
    defaultTTL time.Duration
}

type cacheItem struct {
    value  interface{}
    expiry time.Time
}

func NewInMemoryCache(defaultTTL time.Duration) *InMemoryCache {
    cache := &InMemoryCache{
        data:       make(map[string]*cacheItem),
        defaultTTL: defaultTTL,
    }
    
    // Start cleanup goroutine to remove expired items
    go cache.startCleanup()
    
    return cache
}

func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    item, exists := c.data[key]
    if !exists || time.Now().After(item.expiry) {
        return nil, nil // Not found
    }
    
    return item.value, nil
}

func (c *InMemoryCache) GetWithDefault(ctx context.Context, key string, defaultValue interface{}) (interface{}, error) {
    value, err := c.Get(ctx, key)
    if err != nil {
        return defaultValue, nil
    }
    if value == nil {
        return defaultValue, nil
    }
    return value, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if ttl <= 0 {
        ttl = c.defaultTTL
    }
    
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    expiry := time.Now().Add(ttl)
    c.data[key] = &cacheItem{
        value:  value,
        expiry: expiry,
    }
    
    return nil
}

func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    delete(c.data, key)
    return nil
}

func (c *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    item, exists := c.data[key]
    if !exists || time.Now().After(item.expiry) {
        return false, nil
    }
    
    return true, nil
}

func (c *InMemoryCache) InvalidateByPattern(ctx context.Context, pattern string) error {
    // In-memory cache doesn't support pattern matching,
    // so we'll clear the entire cache as a fallback
    return c.Clear(ctx)
}

func (c *InMemoryCache) Clear(ctx context.Context) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.data = make(map[string]*cacheItem)
    return nil
}

func (c *InMemoryCache) startCleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.cleanupExpired()
    }
}

func (c *InMemoryCache) cleanupExpired() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    now := time.Now()
    for key, item := range c.data {
        if now.After(item.expiry) {
            delete(c.data, key)
        }
    }
}

func (c *InMemoryCache) GetTyped(ctx context.Context, key string, dest interface{}) error {
    value, err := c.Get(ctx, key)
    if err != nil || value == nil {
        return err
    }
    
    // In a real implementation, you would handle type assertion properly
    // Here we're assuming the value is already the correct type
    // or needs to be unmarshaled from a stored format
    return nil
}

func (c *InMemoryCache) SetTyped(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return c.Set(ctx, key, value, ttl)
}
```

## Integration with External Solutions

### Redis Cache Implementation

```go
// internal/platform/cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
)

type RedisCache struct {
    client *redis.Client
    defaultTTL time.Duration
}

func NewRedisCache(client *redis.Client, defaultTTL time.Duration) *RedisCache {
    return &RedisCache{
        client:     client,
        defaultTTL: defaultTTL,
    }
}

func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    result, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, nil // Key not found
        }
        return nil, err
    }
    
    var value interface{}
    if err := json.Unmarshal([]byte(result), &value); err != nil {
        return nil, err
    }
    
    return value, nil
}

func (r *RedisCache) GetWithDefault(ctx context.Context, key string, defaultValue interface{}) (interface{}, error) {
    value, err := r.Get(ctx, key)
    if err != nil {
        return defaultValue, nil
    }
    if value == nil {
        return defaultValue, nil
    }
    return value, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if ttl <= 0 {
        ttl = r.defaultTTL
    }
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
    result, err := r.client.Exists(ctx, key).Result()
    if err != nil {
        return false, err
    }
    
    return result > 0, nil
}

func (r *RedisCache) InvalidateByPattern(ctx context.Context, pattern string) error {
    // Find keys by pattern and delete them
    keys, err := r.client.Keys(ctx, pattern).Result()
    if err != nil {
        return err
    }
    
    if len(keys) > 0 {
        return r.client.Del(ctx, keys...).Err()
    }
    
    return nil
}

func (r *RedisCache) Clear(ctx context.Context) error {
    return r.client.FlushDB(ctx).Err()
}

func (r *RedisCache) GetTyped(ctx context.Context, key string, dest interface{}) error {
    result, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil // Key not found
        }
        return err
    }
    
    return json.Unmarshal([]byte(result), dest)
}

func (r *RedisCache) SetTyped(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if ttl <= 0 {
        ttl = r.defaultTTL
    }
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return r.client.Set(ctx, key, data, ttl).Err()
}
```

### Redis Configuration

```go
// internal/platform/cache/config.go
package cache

import (
    "github.com/go-redis/redis/v8"
)

type CacheType string

const (
    InMemoryCacheType CacheType = "inmemory"
    RedisCacheType    CacheType = "redis"
)

type Config struct {
    Type         CacheType
    DefaultTTL   time.Duration
    InMemorySize int // For future in-memory size limits
    
    Redis RedisConfig
}

type RedisConfig struct {
    Address  string
    Password string
    DB       int
    PoolSize int
}

func NewRedisClient(config RedisConfig) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     config.Address,
        Password: config.Password,
        DB:       config.DB,
        PoolSize: config.PoolSize,
    })
}
```

## Implementation Examples

### Integration in Wire Package

```go
// cmd/forum/wire/app.go (updated)
package wire

import (
    "forum/internal/platform/cache"
    // ... other imports
)

// InitializeApp creates and wires all application components.
func InitializeApp(cfg *config.Config, lgr *logger.Logger) (*App, error) {
    lgr.Info("Initializing application components")

    // Create cache based on configuration
    cache := initCache(cfg.Cache)
    
    // 1. Initialize Database
    db, err := initDatabase(cfg, lgr)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize database: %w", err)
    }

    // 2. Initialize Repositories (Output Adapters)
    repos := initRepositories(db.DB())

    // 3. Initialize Services (Application Layer)
    services := initServices(repos, cfg.Session.Duration, cache) // Pass cache to services

    // 4. Initialize HTTP Handlers (Input Adapters)
    handlers := initHandlers(services)

    // 5. Initialize HTTP Server
    server := initServer(cfg, handlers, lgr)

    lgr.Info("Application initialization complete")

    return &App{
        Server: server,
        DB:     db,
        Logger: lgr,
        Cache:  cache, // Add cache to the app struct
    }, nil
}

func initCache(cfg cache.Config) cache.TypedCache {
    switch cfg.Type {
    case cache.RedisCacheType:
        client := cache.NewRedisClient(cfg.Redis)
        return cache.NewRedisCache(client, cfg.DefaultTTL)
    case cache.InMemoryCacheType:
        fallthrough
    default:
        return cache.NewInMemoryCache(cfg.DefaultTTL)
    }
}
```

### Using Cache in Business Modules

```go
// internal/modules/post/application/service.go
package application

import (
    "context"
    "fmt"
    "forum/internal/modules/post/domain"
    "forum/internal/modules/post/ports"
    "forum/internal/platform/cache"
)

type Service struct {
    postRepository ports.PostRepository
    cache          cache.TypedCache
    defaultTTL     time.Duration
}

func NewService(postRepo ports.PostRepository, cache cache.TypedCache, defaultTTL time.Duration) *Service {
    return &Service{
        postRepository: postRepo,
        cache:          cache,
        defaultTTL:     defaultTTL,
    }
}

// GetPost retrieves a post by ID, with caching
func (s *Service) GetPost(ctx context.Context, postID int) (*domain.Post, error) {
    cacheKey := fmt.Sprintf("post:%d", postID)
    
    // Try to get from cache first
    var post domain.Post
    if err := s.cache.GetTyped(ctx, cacheKey, &post); err == nil {
        return &post, nil
    }
    
    // If not in cache, fetch from repository
    post, err := s.postRepository.GetByID(ctx, postID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for future requests
    if err := s.cache.SetTyped(ctx, cacheKey, post, s.defaultTTL); err != nil {
        // Log error but don't fail the operation
        // log.Printf("Failed to cache post: %v", err)
    }
    
    return &post, nil
}

// CreatePost creates a new post and invalidates related caches
func (s *Service) CreatePost(ctx context.Context, post *domain.Post) error {
    // Create the post in the repository
    err := s.postRepository.Create(ctx, post)
    if err != nil {
        return err
    }
    
    // Invalidate posts list cache
    if err := s.cache.InvalidateByPattern(ctx, "posts:list:*"); err != nil {
        // Log error but don't fail the operation
        // log.Printf("Failed to invalidate posts list cache: %v", err)
    }
    
    return nil
}

// GetPosts retrieves posts with caching
func (s *Service) GetPosts(ctx context.Context, filter *domain.PostFilter) ([]*domain.Post, error) {
    cacheKey := fmt.Sprintf("posts:list:page:%d:limit:%d:category:%s", 
        filter.Page, filter.Limit, filter.Category)
    
    // Try to get from cache first
    var posts []*domain.Post
    if err := s.cache.GetTyped(ctx, cacheKey, &posts); err == nil {
        return posts, nil
    }
    
    // If not in cache, fetch from repository
    posts, err := s.postRepository.GetByFilter(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for future requests
    if err := s.cache.SetTyped(ctx, cacheKey, posts, s.defaultTTL); err != nil {
        // Log error but don't fail the operation
        // log.Printf("Failed to cache posts list: %v", err)
    }
    
    return posts, nil
}
```

### Cache Middleware (Optional Enhancement)

```go
// internal/platform/cache/middleware.go
package cache

import (
    "context"
    "net/http"
    "time"
)

// HTTPCacheMiddleware provides HTTP-level caching
type HTTPCacheMiddleware struct {
    cache    TypedCache
    duration time.Duration
}

func NewHTTPCacheMiddleware(cache TypedCache, duration time.Duration) *HTTPCacheMiddleware {
    return &HTTPCacheMiddleware{
        cache:    cache,
        duration: duration,
    }
}

func (m *HTTPCacheMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // This is a simplified example - real implementation would need to handle
        // more complex caching strategies
        cacheKey := r.URL.Path + r.URL.RawQuery
        
        // Check if we have a cached response
        var cachedResponse string
        if err := m.cache.GetTyped(r.Context(), cacheKey, &cachedResponse); err == nil && cachedResponse != "" {
            w.Header().Set("X-Cache", "HIT")
            w.Write([]byte(cachedResponse))
            return
        }
        
        // Execute the next handler and capture the response
        // (This would require a ResponseWriter wrapper to capture output)
        next.ServeHTTP(w, r)
        
        // Cache the response for future requests
        // This would be implemented with a custom ResponseWriter that captures
        // the response body before it's sent to the client
    })
}
```

## Usage in Business Modules

The cache infrastructure will be available to all business modules through dependency injection:

1. **Posts Module**: Cache frequently accessed posts and post lists
2. **User Module**: Cache user profiles and permissions
3. **Comment Module**: Cache comment threads
4. **Reaction Module**: Cache reaction counts
5. **Auth Module**: Cache session information

Each business module will receive the cache as a dependency through its constructor, maintaining proper separation of concerns while providing caching capabilities where needed.

## Configuration

The caching implementation will be configurable through environment variables:

```env
# Cache configuration
CACHE_TYPE=redis          # "redis" or "inmemory"
CACHE_DEFAULT_TTL=3600    # Default TTL in seconds

# Redis configuration
REDIS_ADDRESS=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

And in the application config:

```go
// internal/platform/config/config.go (updated)
type Config struct {
    // ... other config fields
    Cache cache.Config
}
```

## Conclusion

This updated caching infrastructure proposal positions caching as a platform-level concern, which is the appropriate architectural approach for cross-cutting functionality like caching. Key benefits of this approach include:

1. **Proper Separation of Concerns**: Business logic remains focused on its specific domain without infrastructure concerns

2. **Reusability**: The same caching infrastructure can be used by multiple business modules

3. **Consistency**: All modules use the same caching interface and patterns

4. **Maintainability**: Infrastructure changes are centralized in one location

5. **Flexibility**: Easy to switch between caching implementations (in-memory, Redis) without changing business modules

6. **Hexagonal Architecture Compliance**: Maintains the proper dependency directions with business modules depending on platform infrastructure rather than the other way around

7. **Scalability**: External caching solutions like Redis can be easily integrated when scaling needs arise

This approach aligns with the existing architecture of your forum application by placing the caching infrastructure alongside other cross-cutting concerns like logging, configuration, and database connections in the platform layer.