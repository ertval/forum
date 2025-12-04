# Caching Layer Implementation Proposal

## Overview

This document outlines a comprehensive caching strategy for the forum application that follows our Hexagonal Architecture patterns. The caching layer will be implemented as a dedicated **cache module** that can be consumed by other modules through well-defined interfaces.

## Architecture Decision

### Should We Create a New Module?

**Yes, absolutely.** Caching is a cross-cutting concern that affects multiple modules (posts, users, sessions, reactions). Following the Modular Monolith pattern, we should create a dedicated `cache` module that:

1. Provides a **consistent caching interface** for all modules
2. Supports **multiple implementations** (in-memory, Redis, Memcached)
3. Enables **easy testing** with in-memory implementations
4. Follows the **same 4-layer structure** (domain, ports, application, adapters)
5. Can be **swapped** without affecting business logic

## Module Structure

```
cache/
├── domain/          # Pure caching logic
│   ├── entity.go    # Cache entities (if needed)
│   └── errors.go    # Cache-specific errors
│
├── ports/           # Cache interfaces
│   ├── service.go   # INPUT PORT - Cache service interface
│   └── repository.go # OUTPUT PORT - Cache storage interface
│
├── application/     # Cache orchestration
│   └── service.go   # Implements service.go with cache logic
│
└── adapters/        # Cache implementations
    ├── memory_repository.go   # OUTPUT ADAPTER - In-memory cache
    ├── redis_repository.go    # OUTPUT ADAPTER - Redis cache
    └── cache_decorator.go     # INPUT ADAPTER - Caching decorator
```

## Implementation Details

### 1. Domain Layer

**`internal/modules/cache/domain/errors.go`**
```go
package domain

import "errors"

var (
    ErrCacheMiss     = errors.New("cache key not found")
    ErrCacheInvalid  = errors.New("cache value is invalid")
    ErrCacheWrite    = errors.New("failed to write to cache")
    ErrCacheRead     = errors.New("failed to read from cache")
    ErrCacheDelete   = errors.New("failed to delete from cache")
)
```

**`internal/modules/cache/domain/entity.go`**
```go
package domain

import "time"

// CacheEntry represents a cached value with metadata.
type CacheEntry struct {
    Key        string
    Value      []byte
    Expiration time.Duration
    CreatedAt  time.Time
}
```

### 2. Ports Layer

**`internal/modules/cache/ports/repository.go` - OUTPUT PORT**
```go
// OUTPUT PORT - Cache Storage Interface
package ports

import (
   "context"
    "forum/internal/modules/cache/domain"
)

type CacheRepository interface {
    // Set stores a value in the cache with an expiration time.
    Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

    // Get retrieves a value from the cache.
    Get(ctx context.Context, key string) ([]byte, error)

    // Delete removes a value from the cache.
    Delete(ctx context.Context, key string) error

    // Clear removes all values from the cache.
    Clear(ctx context.Context) error

    // GetWithTTL retrieves a value and its remaining TTL.
    GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error)
}
```

**`internal/modules/cache/ports/service.go` - INPUT PORT**
```go
// INPUT PORT - Cache Service Interface
package ports

import "context"

type CacheService interface {
    // Get retrieves a value from cache. Returns domain.ErrCacheMiss if not found.
    Get(ctx context.Context, key string) ([]byte, error)

    // Set stores a value in cache with expiration.
    Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

    // Delete removes a value from cache.
    Delete(ctx context.Context, key string) error

    // GetOrSet retrieves from cache, or sets using the provided function if missing.
    GetOrSet(ctx context.Context, key string, value []byte, expiration time.Duration, setFn func() ([]byte, error)) ([]byte, error)

    // InvalidatePattern deletes all keys matching a pattern.
    InvalidatePattern(ctx context.Context, pattern string) error
}
```

### 3. Application Layer

**`internal/modules/cache/application/service.go`**
```go
// Package application implements cache service business logic.
package application

import (
    "context"
    "forum/internal/modules/cache/domain"
    "forum/internal/modules/cache/ports"
)

type Service struct {
    cacheRepo ports.CacheRepository
}

func NewService(cacheRepo ports.CacheRepository) *Service {
    return &Service{
        cacheRepo: cacheRepo,
    }
}

func (s *Service) Get(ctx context.Context, key string) ([]byte, error) {
    value, err := s.cacheRepo.Get(ctx, key)
    if err != nil {
        return nil, domain.ErrCacheMiss
    }
    return value, nil
}

func (s *Service) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
    if err := s.cacheRepo.Set(ctx, key, value, expiration); err != nil {
        return domain.ErrCacheWrite
    }
    return nil
}

func (s *Service) Delete(ctx context context.Context, key string) error {
    if err := s.cacheRepo.Delete(ctx, key); err != nil {
        return domain.ErrCacheDelete
    }
    return nil
}

func (s *Service) GetOrSet(ctx context.Context, key string, value []byte, expiration time.Duration, setFn func() ([]byte, error)) ([]byte, error) {
    // Try to get from cache first
    if cached, err := s.cacheRepo.Get(ctx, key); err == nil {
        return cached, nil
    }

    // Cache miss - call the function to generate value
    newValue, err := setFn()
    if err != nil {
        return nil, err
    }

    // Store in cache for next time
    if err := s.cacheRepo.Set(ctx, key, newValue, expiration); err == nil {
        return newValue, nil
    }

    // If cache write fails, return the value anyway
    return newValue, nil
}

func (s *Service) InvalidatePattern(ctx context.Context, pattern string) error {
    // Implementation depends on the cache backend
    // Memory: iterate and delete matching keys
    // Redis: use SCAN + DEL pattern
    return nil
}
```

### 4. Adapters Layer

**In-Memory Implementation - `internal/modules/cache/adapters/memory_repository.go`**
```go
// OUTPUT ADAPTER - In-Memory Cache Repository
package adapters

import (
    "context"
    "sync"
    "time"
    "forum/internal/modules/cache/ports"
)

type MemoryRepository struct {
    data map[string]cacheEntry
    mu   sync.RWMutex
}

type cacheEntry struct {
    value     []byte
    expiresAt time.Time
}

func NewMemoryRepository() *MemoryRepository {
    repo := &MemoryRepository{
        data: make(map[string]cacheEntry),
    }

    // Start cleanup goroutine
    go repo.cleanup()

    return repo
}

func (r *MemoryRepository) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.data[key] = cacheEntry{
        value:     value,
        expiresAt: time.Now().Add(expiration),
    }
    return nil
}

func (r *MemoryRepository) Get(ctx context.Context, key string) ([]byte, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    entry, exists := r.data[key]
    if !exists {
        return nil, ports.ErrCacheMiss
    }

    if time.Now().After(entry.expiresAt) {
        // Expired - delete and return miss
        delete(r.data, key)
        return nil, ports.ErrCacheMiss
    }

    return entry.value, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, key string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.data, key)
    return nil
}

func (r *MemoryRepository) Clear(ctx context.Context) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.data = make(map[string]cacheEntry)
    return nil
}

func (r *MemoryRepository) GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    entry, exists := r.data[key]
    if !exists {
        return nil, 0, ports.ErrCacheMiss
    }

    ttl := time.Until(entry.expiresAt)
    if ttl <= 0 {
        delete(r.data, key)
        return nil, 0, ports.ErrCacheMiss
    }

    return entry.value, ttl, nil
}

func (r *MemoryRepository) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        r.mu.Lock()
        now := time.Now()
        for key, entry := range r.data {
            if now.After(entry.expiresAt) {
                delete(r.data, key)
            }
        }
        r.mu.Unlock()
    }
}
```

**Redis Implementation - `internal/modules/cache/adapters/redis_repository.go`**
```go
// OUTPUT ADAPTER - Redis Cache Repository
package adapters

import (
    "context"
    "time"

    "github.com/go-redis/redis/v8"
    "forum/internal/modules/cache/ports"
)

type RedisRepository struct {
    client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
    return &RedisRepository{
        client: client,
    }
}

func (r *RedisRepository) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
    return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisRepository) Get(ctx context.Context, key string) ([]byte, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, ports.ErrCacheMiss
    }
    if err != nil {
        return nil, err
    }
    return []byte(val), nil
}

func (r *RedisRepository) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}

func (r *RedisRepository) Clear(ctx context.Context) error {
    return r.client.FlushAll(ctx).Err()
}

func (r *RedisRepository) GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error) {
    pipe := r.client.Pipeline()
    getCmd := pipe.Get(ctx, key)
    ttlCmd := pipe.TTL(ctx, key)

    _, err := pipe.Exec(ctx)
    if err == redis.Nil {
        return nil, 0, ports.ErrCacheMiss
    }
    if err != nil {
        return nil, 0, err
    }

    val, err := getCmd.Result()
    if err != nil {
        return nil, 0, err
    }

    ttl, err := ttlCmd.Result()
    if err != nil {
        return nil, 0, err
    }

    return []byte(val), ttl, nil
}
```

**Cache Decorator Pattern - `internal/modules/cache/adapters/cache_decorator.go`**
```go
// INPUT ADAPTER - Cache Decorator for wrapping repositories
package adapters

import (
    "context"
    "forum/internal/modules/cache/ports"
)

// RepositoryDecorator wraps a repository with caching capability.
type RepositoryDecorator struct {
    cache  ports.CacheService
    repo   interface{}
    prefix string
    ttl    time.Duration
}

func NewRepositoryDecorator(cache ports.CacheService, repo interface{}, prefix string, ttl time.Duration) *RepositoryDecorator {
    return &RepositoryDecorator{
        cache:  cache,
        repo:   repo,
        prefix: prefix,
        ttl:    ttl,
    }
}

// GetPost is an example of how to use the decorator.
// This would be called from the post service.
func (d *RepositoryDecorator) GetPost(ctx context.Context, postID int) (*domain.Post, error) {
    cacheKey := d.key("post", postID)

    // Try cache first
    if cached, err := d.cache.Get(ctx, cacheKey); err == nil {
        // Deserialize and return
        return deserializePost(cached)
    }

    // Cache miss - get from repository
    post, err := d.getFromRepo(ctx, postID)
    if err != nil {
        return nil, err
    }

    // Store in cache
    serialized, err := serializePost(post)
    if err == nil {
        d.cache.Set(ctx, cacheKey, serialized, d.ttl)
    }

    return post, nil
}

func (d *RepositoryDecorator) InvalidatePost(ctx context.Context, postID int) error {
    cacheKey := d.key("post", postID)
    return d.cache.Delete(ctx, cacheKey)
}

func (d *RepositoryDecorator) key(parts ...interface{}) string {
    // Build cache key: prefix:part1:part2:...
    // Example: "post:123"
    return "cache:" + d.prefix + ":" + fmt.Sprintf("%v", parts...)
}
```

## Integration with Existing Modules

### Option 1: Repository Wrapper Pattern

Modify existing modules to use a cache-wrapped repository.

**`internal/modules/post/application/service.go` (Modified)**
```go
package application

import (
    "context"
    "forum/internal/modules/post/domain"
    "forum/internal/modules/post/ports"
    "forum/internal/modules/cache/ports" as cachePorts
)

type Service struct {
    postRepo     ports.PostRepository
    categoryRepo ports.CategoryRepository
    cache        cachePorts.CacheService // New dependency
}

// NewService creates a new post service.
func NewService(
    postRepo ports.PostRepository,
    categoryRepo ports.CategoryRepository,
    cacheService cachePorts.CacheService,
) *Service {
    return &Service{
        postRepo:     postRepo,
        categoryRepo: categoryRepo,
        cache:        cacheService,
    }
}

// GetPost with caching
func (s *Service) GetPost(ctx context.Context, postID int) (*domain.Post, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("post:%d", postID)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        // Deserialize cached value
        return deserializePost(cached)
    }

    // Cache miss - get from repository
    post, err := s.postRepo.GetByID(ctx, postID)
    if err != nil {
        return nil, err
    }

    // Cache the result
    if serialized, err := serializePost(post); err == nil {
        s.cache.Set(ctx, cacheKey, serialized, 30*time.Minute)
    }

    return post, nil
}

// UpdatePost with cache invalidation
func (s *Service) UpdatePost(ctx context.Context, postID int, title, content string) error {
    // Update in database
    err := s.postRepo.Update(ctx, postID, title, content)
    if err != nil {
        return err
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("post:%d", postID)
    s.cache.Delete(ctx, cacheKey)

    return nil
}
```

### Option 2: Cache Decorator Pattern

Keep services clean by wrapping repositories.

**`cmd/forum/wire/repos.go` (Modified)**
```go
package wire

// Existing code...

// Create cache service
cacheRepo := cacheAdapters.NewMemoryRepository() // or NewRedisRepository(redisClient)
cacheService := cacheApp.NewService(cacheRepo)

// Wrap post repository with cache decorator
postRepo := postAdapters.NewSQLiteRepository(dbConn.DB())
postRepo = cacheAdapters.NewRepositoryDecorator(cacheService, postRepo, "post", 30*time.Minute)

// Wire the services
postService := postApp.NewService(postRepo, categoryRepo)
```

## Dependency Injection Setup

**`cmd/forum/wire/repos.go`**
```go
package wire

import (
    "forum/internal/modules/cache/adapters"
    "forum/internal/modules/cache/application"
)

var cacheRepo ports.CacheRepository

func initCacheRepo(cfg config.Config) {
    if cfg.Cache.RedisURL != "" {
        // Use Redis
        redisClient := redis.NewClient(&redis.Options{
            Addr: cfg.Cache.RedisURL,
        })
        cacheRepo = adapters.NewRedisRepository(redisClient)
    } else {
        // Use in-memory
        cacheRepo = adapters.NewMemoryRepository()
    }
}
```

**`cmd/forum/wire/services.go`**
```go
package wire

import (
    "forum/internal/modules/cache/adapters"
    "forum/internal/modules/cache/application"
    "forum/internal/modules/cache/ports"
)

var cacheService ports.CacheService

func initCacheService() {
    cacheService = application.NewService(cacheRepo)
}
```

**`cmd/forum/wire/services.go` (Post Service)**
```go
package wire

import (
    "forum/internal/modules/post/adapters"
    "forum/internal/modules/post/application"
)

func initPostService() {
    // Get cached repository
    postRepo := postAdapters.NewSQLiteRepository(dbConn.DB())

    // Create cached version
    cachedPostRepo := adapters.NewCachedRepository(postRepo, cacheService, "post", 30*time.Minute)

    // Create service with cached repository
    postService := postApp.NewService(cachedPostRepo, categoryRepo)
}
```

## Configuration

**`internal/platform/config/config.go`**
```go
type Config struct {
    // ... existing fields

    Cache CacheConfig `env:"CACHE"`
}

type CacheConfig struct {
    Type     string        `env:"CACHE_TYPE" default:"memory"`
    RedisURL string        `env:"REDIS_URL"`
    TTL      time.Duration `env:"CACHE_TTL" default:"30m"`
}
```

**Environment Variables**
```env
# Cache configuration
CACHE_TYPE=memory        # or "redis"
REDIS_URL=localhost:6379 # only if CACHE_TYPE=redis
CACHE_TTL=30m           # default TTL
```

## Caching Strategy by Module

### Posts Module
- **ListPosts**: Cache for 5 minutes (page-level caching)
- **GetPost**: Cache individual posts for 30 minutes
- **ListPostsByCategory**: Cache for 10 minutes
- **Invalidate**: On create, update, delete

### User Module
- **GetUser**: Cache user profiles for 60 minutes
- **GetUserActivity**: Cache for 10 minutes
- **Invalidate**: On profile update

### Auth Module
- **Session Validation**: Cache for session duration
- **User Lookup**: Cache for 30 minutes
- **Invalidate**: On logout

### Reaction Module
- **Reaction Counts**: Cache for 5 minutes
- **User's Liked Posts**: Cache for 15 minutes

## Key Design Decisions

### 1. **Repository Decoration vs Service Modification**
- **Recommendation**: Use repository decorator pattern
- **Reason**: Keeps business logic clean, caching becomes transparent
- **Trade-off**: Adds complexity to wiring

### 2. **In-Memory vs Redis**
- **Development**: In-memory (no external dependency)
- **Production**: Redis (shared cache across instances, persistence)
- **Fallback**: Application should work without cache if Redis is down

### 3. **Cache Key Naming Convention**
```
Format: cache:{module}:{entity}:{id}
Example: cache:post:123, cache:user:456
```

### 4. **TTL (Time To Live)**
- **Default**: 30 minutes
- **Adjust per use case**:
  - Frequent changes: 5-10 minutes
  - Rare changes: 1-2 hours
  - User sessions: Session duration

### 5. **Cache Invalidation Strategy**
- **Write-through**: Update DB → Invalidate cache
- **Write-behind**: Update DB → Async cache update (not recommended for this use case)
- **Write-around**: Update DB only (cache miss on next read)

**Recommendation**: Write-through for consistency.

## Redis Integration Details

### Setup Requirements

**`go.mod` additions:**
```go
require (
    github.com/go-redis/redis/v8 v8.11.5
)
```

### Connection Management

**`internal/platform/database/redis.go`**
```go
package database

import (
    "github.com/go-redis/redis/v8"
)

type RedisConnection struct {
    client *redis.Client
}

func NewRedisConnection(url string) (*RedisConnection, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     url,
        Password: "",
        DB:       0,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }

    return &RedisConnection{
        client: client,
    }, nil
}

func (r *RedisConnection) Client() *redis.Client {
    return r.client
}

func (r *RedisConnection) Close() error {
    return r.client.Close()
}
```

### Docker Compose for Redis

**`docker-compose.yml`**
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CACHE_TYPE=redis
      - REDIS_URL=redis:6379
    depends_on:
      - redis
    volumes:
      - ./static/uploads:/app/static/uploads

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

## Performance Considerations

### 1. **Cache Hit Rate**
- Monitor hit rate with metrics
- Target: >80% for frequently accessed data
- Adjust TTL if hit rate is too low

### 2. **Memory Usage**
- In-memory: Monitor process memory
- Redis: Set maxmemory limit
- Consider eviction policies (LRU for Redis)

### 3. **Connection Pooling**
- Redis client should use connection pooling
- Set MaxRetries, PoolSize, MinIdleConns

### 4. **Serialization**
- Use efficient format (JSON, MsgPack, or Gob)
- Consider compression for large objects

## Testing Strategy

### 1. **Unit Tests**
- Test cache service with in-memory repository
- Test repository decorators
- Verify TTL expiration

### 2. **Integration Tests**
- Test with real Redis (in Docker)
- Test cache invalidation
- Test error handling (cache miss, connection failure)

### 3. **Performance Tests**
- Benchmark cache vs database access
- Load test with high cache usage
- Memory usage profiling

## Example: Complete Flow

### User Profile Caching

**1. Initial Request (Cache Miss)**
```
GET /api/users/123
→ Check cache: miss
→ Query database: 50ms
→ Store in cache
→ Return to user
```

**2. Subsequent Request (Cache Hit)**
```
GET /api/users/123
→ Check cache: hit
→ Return to user: 2ms
→ Save 48ms, 96% faster
```

**3. Profile Update (Cache Invalidation)**
```
PUT /api/users/123
→ Update database: 50ms
→ Delete from cache
→ Return success
```

**4. Next Request After Update**
```
GET /api/users/123
→ Check cache: miss
→ Query database
→ Store in cache
→ Return fresh data
```

## Recommended Implementation Order

1. **Phase 1**: Create cache module structure (domain, ports, application)
2. **Phase 2**: Implement in-memory repository
3. **Phase 3**: Integrate with post module (ListPosts, GetPost)
4. **Phase 4**: Add cache invalidation on writes
5. **Phase 5**: Implement Redis repository
6. **Phase 6**: Extend to other modules (user, auth, reaction)
7. **Phase 7**: Add metrics and monitoring

## Conclusion

This caching layer design:
- ✅ Follows Hexagonal Architecture
- ✅ Enables easy testing with in-memory implementation
- ✅ Supports multiple cache backends
- ✅ Maintains clean separation of concerns
- ✅ Integrates seamlessly with existing modules
- ✅ Can be incrementally adopted
- ✅ Provides significant performance improvements

**Expected Impact**: 80-95% reduction in database load for read-heavy operations, with response times improving from 50-100ms to 1-5ms for cached data.
