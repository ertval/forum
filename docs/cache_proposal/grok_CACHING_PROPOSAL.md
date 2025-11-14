# Caching Layer Implementation Proposal

## Overview

This proposal outlines the implementation of a caching layer for the forum application to improve performance, reduce database load, and enhance user experience. The caching layer will follow the established Hexagonal Architecture pattern and integrate seamlessly with the existing modular monolith structure.

## Why Caching?

- **Performance**: Reduce response times for frequently accessed data
- **Scalability**: Decrease database load during peak usage
- **User Experience**: Faster page loads and API responses
- **Cost Efficiency**: Lower infrastructure costs through reduced database queries

## Architecture Decision: Platform vs Module

### Option 1: Platform Package (Recommended)
Place caching in `internal/platform/cache/` as infrastructure concern:

```
internal/platform/cache/
├── cache.go          # Core caching interfaces and types
├── memory.go         # In-memory implementation
├── redis.go          # Redis implementation
└── cache_test.go     # Unit tests
```

**Pros:**
- Cross-cutting concern fits platform layer
- Reusable across all modules
- Follows infrastructure separation
- Easier dependency injection

**Cons:**
- Less explicit business domain separation

### Option 2: Dedicated Cache Module
Create `internal/modules/cache/` following full module structure:

```
internal/modules/cache/
├── domain/
│   ├── cache_entry.go
│   └── errors.go
├── ports/
│   ├── service.go      // INPUT PORT
│   └── repository.go   // OUTPUT PORT
├── application/
│   └── service.go
└── adapters/
    ├── memory_repository.go
    └── redis_repository.go
```

**Pros:**
- Consistent with module architecture
- Clear separation of concerns
- Full hexagonal pattern compliance

**Cons:**
- Overkill for infrastructure concern
- Additional complexity for simple caching

**Recommendation:** Use Option 1 (Platform Package) for simplicity and alignment with infrastructure concerns, while maintaining clean interfaces.

## Implementation Strategy

### Core Interfaces

```go
// internal/platform/cache/cache.go
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
    Exists(ctx context.Context, key string) bool
}

type CacheConfig struct {
    Type     string        // "memory" or "redis"
    TTL      time.Duration // Default TTL
    Redis    RedisConfig
}

type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
    PoolSize int
}
```

### In-Memory Implementation

Simple thread-safe in-memory cache using sync.Map:

```go
type MemoryCache struct {
    data sync.Map
    ttl  time.Duration
}

func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
    // Implementation with TTL checking
}

func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // Implementation with expiration
}
```

### Redis Implementation

External Redis integration:

```go
import "github.com/go-redis/redis/v8"

type RedisCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewRedisCache(config RedisConfig, defaultTTL time.Duration) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
        Password: config.Password,
        DB:       config.DB,
        PoolSize: config.PoolSize,
    })
    return &RedisCache{client: rdb, ttl: defaultTTL}
}
```

## Integration with Existing Modules

### Dependency Injection Pattern

Update `cmd/forum/wire/` to include cache:

```go
// cmd/forum/wire/app.go
cache := platformCache.NewMemoryCache(config.Cache.TTL)
// OR for Redis:
cache := platformCache.NewRedisCache(config.Cache.Redis, config.Cache.TTL)

// Inject into services that need caching
postService := postApp.NewPostService(postRepo, cache)
userService := userApp.NewUserService(userRepo, cache)
```

### Service Layer Integration

Modify existing services to use caching:

```go
// Example: internal/modules/post/application/service.go
type PostService struct {
    repo  ports.PostRepository
    cache platform.Cache
}

func (s *PostService) GetPost(ctx context.Context, id string) (*domain.Post, error) {
    cacheKey := fmt.Sprintf("post:%s", id)
    
    // Try cache first
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached.(*domain.Post), nil
    }
    
    // Cache miss - fetch from database
    post, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    s.cache.Set(ctx, cacheKey, post, s.defaultTTL)
    return post, nil
}
```

### Cache Key Strategy

Standardized cache key patterns:
- `post:{id}` - Individual posts
- `posts:category:{category_id}:page:{page}` - Paginated category posts
- `user:{id}` - User profiles
- `comments:post:{post_id}:page:{page}` - Post comments
- `reactions:post:{post_id}` - Post reaction counts

### Cache Invalidation Strategy

**Write-Through Pattern:**
- Update database first
- Invalidate/update cache immediately
- Ensures consistency

```go
func (s *PostService) UpdatePost(ctx context.Context, id string, updates domain.PostUpdates) error {
    // Update database
    err := s.repo.Update(ctx, id, updates)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("post:%s", id)
    s.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

## Configuration

Add to `internal/platform/config/config.go`:

```go
type Config struct {
    // ... existing fields ...
    Cache CacheConfig
}

type CacheConfig struct {
    Type string        `env:"CACHE_TYPE" default:"memory"`
    TTL  time.Duration `env:"CACHE_TTL" default:"10m"`
    Redis struct {
        Host     string `env:"REDIS_HOST" default:"localhost"`
        Port     int    `env:"REDIS_PORT" default:"6379"`
        Password string `env:"REDIS_PASSWORD"`
        DB       int    `env:"REDIS_DB" default:"0"`
        PoolSize int    `env:"REDIS_POOL_SIZE" default:"10"`
    }
}
```

## Migration Strategy

### Phase 1: Core Infrastructure
1. Add cache package to platform
2. Implement in-memory cache
3. Add configuration
4. Wire into dependency injection
5. Add basic cache methods to services

### Phase 2: Service Integration
1. Identify high-traffic read operations
2. Add caching to post retrieval
3. Add caching to user profiles
4. Add caching to comment threads
5. Implement cache invalidation

### Phase 3: Redis Integration
1. Add Redis dependency
2. Implement Redis adapter
3. Add Redis configuration
4. Update wire package for Redis option
5. Performance testing and optimization

### Phase 4: Advanced Features
1. Cache warming strategies
2. Distributed cache invalidation
3. Cache analytics/metrics
4. Circuit breaker patterns

## Testing Strategy

### Unit Tests
- Test cache implementations in isolation
- Mock external dependencies (Redis)
- Test TTL expiration logic
- Test concurrent access patterns

### Integration Tests
- Test full request/response cycles with caching
- Verify cache hit/miss ratios
- Test cache invalidation scenarios
- Performance benchmarks

### Cache-Specific Tests
```go
func TestPostService_GetPost_CacheHit(t *testing.T) {
    // Setup mock cache with expected post
    // Verify database not called
    // Verify correct post returned
}

func TestPostService_GetPost_CacheMiss(t *testing.T) {
    // Setup empty cache
    // Verify database called
    // Verify cache populated
    // Verify correct post returned
}
```

## Performance Considerations

### Cache Hit Ratios
Monitor and optimize:
- Target >80% hit ratio for hot data
- Implement cache warming for critical data
- Use appropriate TTL values per data type

### Memory Management
- Set reasonable TTL values
- Implement cache size limits
- Monitor memory usage in production

### Redis Cluster Considerations
- Connection pooling configuration
- Failover and reconnection logic
- Serialization strategy for complex objects

## Monitoring and Observability

Add metrics to track:
- Cache hit/miss ratios
- Cache operation latency
- Memory usage
- Error rates

Integration with existing logger:
```go
lgr.Info("Cache operation",
    logger.String("operation", "get"),
    logger.String("key", cacheKey),
    logger.Bool("hit", true),
    logger.Duration("latency", time.Since(start)))
```

## Security Considerations

- Cache data may contain sensitive information
- Implement proper serialization
- Consider cache encryption for Redis
- Rate limiting for cache operations
- Access control for cache management endpoints

## Rollback Strategy

- Feature flags for cache enablement
- Gradual rollout with monitoring
- Quick disable option if issues arise
- Database load monitoring during rollout

## Conclusion

Implementing a caching layer using the platform package approach provides the best balance of simplicity, performance, and architectural consistency. Start with in-memory caching for immediate benefits, then add Redis for scalability. The hexagonal architecture ensures clean separation and testability while maintaining the modular monolith structure.

## Implementation Priority

1. ✅ Core cache interfaces and in-memory implementation
2. ✅ Configuration and dependency injection
3. ✅ Basic service integration (posts, users)
4. 🔄 Redis adapter implementation
5. 🔄 Advanced features (warming, metrics)
6. 🔄 Production monitoring and optimization