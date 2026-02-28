# Caching Layer Implementation Proposal for Forum Application

## Table of Contents
1. [Introduction](#introduction)
2. [Architecture Overview](#architecture-overview)
3. [Module Structure](#module-structure)
4. [Integration with External Solutions](#integration-with-external-solutions)
5. [Implementation Examples](#implementation-examples)
6. [Usage in Other Modules](#usage-in-other-modules)
7. [Configuration](#configuration)
8. [Conclusion](#conclusion)

## Introduction

This proposal outlines the implementation of a caching layer for the forum application using a hexagonal architecture pattern. The caching layer will provide both in-memory and external caching capabilities (such as Redis) to improve application performance by reducing database queries and response times.

## Architecture Overview

The caching module will follow the hexagonal architecture pattern with the following structure:

```
internal/modules/cache/
├── domain/          # Cache operations interfaces and types
├── ports/           # Cache service and repository interfaces
├── application/     # Cache orchestration logic
└── adapters/        # In-memory cache, Redis adapter, etc.
```

### Key Design Principles
- **Interface Segregation**: Separate interfaces for different caching operations
- **Dependency Inversion**: Modules depend on cache abstractions, not implementations
- **Extensibility**: Easy to swap between different caching backends
- **Type Safety**: Strong typing for cache keys and value types

## Module Structure

### Domain Layer
The domain layer contains the core caching concepts:

```go
// internal/modules/cache/domain/cache.go
package domain

import "time"

// CacheKey represents a cache key with additional metadata
type CacheKey struct {
    Key       string
    Namespace string // Optional namespace for the key
}

// Create a full key string from the CacheKey
func (ck CacheKey) FullKey() string {
    if ck.Namespace != "" {
        return ck.Namespace + ":" + ck.Key
    }
    return ck.Key
}

// CacheOptions represents options for cache operations
type CacheOptions struct {
    TTL     time.Duration // Time to live for the cached item
    Tags    []string      // Optional tags for cache invalidation
    Encode  bool          // Whether to encode the value (for complex types)
}
```

### Ports Layer
The ports layer defines the interfaces that other modules will depend on:

```go
// internal/modules/cache/ports/cache_service.go
package ports

import (
    "context"
    "time"
    "forum/internal/modules/cache/domain"
)

// CacheService defines the main caching operations
type CacheService interface {
    // Get retrieves a value from cache
    Get(ctx context.Context, key domain.CacheKey) (interface{}, error)
    
    // GetWithDefault retrieves a value from cache or returns default if not found
    GetWithDefault(ctx context.Context, key domain.CacheKey, defaultValue interface{}) (interface{}, error)
    
    // Set stores a value in cache with optional TTL
    Set(ctx context.Context, key domain.CacheKey, value interface{}, opts *domain.CacheOptions) error
    
    // Delete removes a value from cache
    Delete(ctx context.Context, key domain.CacheKey) error
    
    // Exists checks if a key exists in cache
    Exists(ctx context.Context, key domain.CacheKey) (bool, error)
    
    // InvalidateByTag invalidates all cache entries with a given tag
    InvalidateByTag(ctx context.Context, tag string) error
    
    // InvalidateByNamespace invalidates all cache entries in a namespace
    InvalidateByNamespace(ctx context.Context, namespace string) error
    
    // Clear clears all cache entries
    Clear(ctx context.Context) error
}

// TypedCacheService provides type-safe cache operations
type TypedCacheService interface {
    CacheService
    
    // GetTyped retrieves a value from cache with type information
    GetTyped(ctx context.Context, key domain.CacheKey, dest interface{}) error
    
    // SetTyped stores a value in cache with type information
    SetTyped(ctx context.Context, key domain.CacheKey, value interface{}, opts *domain.CacheOptions) error
}
```

### Application Layer
The application layer orchestrates cache operations:

```go
// internal/modules/cache/application/service.go
package application

import (
    "context"
    "encoding/json"
    "forum/internal/modules/cache/domain"
    "forum/internal/modules/cache/ports"
)

type Service struct {
    cacheRepository ports.CacheRepository
}

func NewService(cacheRepo ports.CacheRepository) *Service {
    return &Service{
        cacheRepository: cacheRepo,
    }
}

func (s *Service) Get(ctx context.Context, key domain.CacheKey) (interface{}, error) {
    return s.cacheRepository.Get(ctx, key)
}

func (s *Service) GetWithDefault(ctx context.Context, key domain.CacheKey, defaultValue interface{}) (interface{}, error) {
    value, err := s.Get(ctx, key)
    if err != nil {
        return defaultValue, nil
    }
    if value == nil {
        return defaultValue, nil
    }
    return value, nil
}

func (s *Service) Set(ctx context.Context, key domain.CacheKey, value interface{}, opts *domain.CacheOptions) error {
    return s.cacheRepository.Set(ctx, key, value, opts)
}

func (s *Service) Delete(ctx context.Context, key domain.CacheKey) error {
    return s.cacheRepository.Delete(ctx, key)
}

func (s *Service) Exists(ctx context.Context, key domain.CacheKey) (bool, error) {
    return s.cacheRepository.Exists(ctx, key)
}

func (s *Service) InvalidateByTag(ctx context.Context, tag string) error {
    return s.cacheRepository.InvalidateByTag(ctx, tag)
}

func (s *Service) InvalidateByNamespace(ctx context.Context, namespace string) error {
    return s.cacheRepository.InvalidateByNamespace(ctx, namespace)
}

func (s *Service) Clear(ctx context.Context) error {
    return s.cacheRepository.Clear(ctx)
}
```

### Adapters Layer
The adapters layer contains the concrete implementations:

```go
// internal/modules/cache/adapters/in_memory_cache.go
package adapters

import (
    "context"
    "sync"
    "time"
    "forum/internal/modules/cache/domain"
    "forum/internal/modules/cache/ports"
)

type InMemoryCache struct {
    data      map[string]*cacheItem
    tags      map[string][]string // tag -> keys mapping
    mutex     sync.RWMutex
    defaultTTL time.Duration
}

type cacheItem struct {
    value     interface{}
    expiry    time.Time
    tags      []string
}

func NewInMemoryCache(defaultTTL time.Duration) *InMemoryCache {
    cache := &InMemoryCache{
        data:       make(map[string]*cacheItem),
        tags:       make(map[string][]string),
        defaultTTL: defaultTTL,
    }
    
    // Start cleanup goroutine to remove expired items
    go cache.startCleanup()
    
    return cache
}

func (c *InMemoryCache) Get(ctx context.Context, key domain.CacheKey) (interface{}, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    item, exists := c.data[key.FullKey()]
    if !exists || time.Now().After(item.expiry) {
        return nil, nil // Not found
    }
    
    return item.value, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key domain.CacheKey, value interface{}, opts *domain.CacheOptions) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    ttl := c.defaultTTL
    if opts != nil && opts.TTL > 0 {
        ttl = opts.TTL
    }
    
    var actualTags []string
    if opts != nil {
        actualTags = opts.Tags
    }
    
    expiry := time.Now().Add(ttl)
    
    c.data[key.FullKey()] = &cacheItem{
        value:  value,
        expiry: expiry,
        tags:   actualTags,
    }
    
    // Add to tag index
    if opts != nil {
        for _, tag := range opts.Tags {
            c.tags[tag] = append(c.tags[tag], key.FullKey())
        }
    }
    
    return nil
}

func (c *InMemoryCache) Delete(ctx context.Context, key domain.CacheKey) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    item, exists := c.data[key.FullKey()]
    if !exists {
        return nil
    }
    
    // Remove from tag index
    for _, tag := range item.tags {
        keys := c.tags[tag]
        newKeys := make([]string, 0, len(keys))
        for _, k := range keys {
            if k != key.FullKey() {
                newKeys = append(newKeys, k)
            }
        }
        c.tags[tag] = newKeys
    }
    
    delete(c.data, key.FullKey())
    return nil
}

func (c *InMemoryCache) Exists(ctx context.Context, key domain.CacheKey) (bool, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    item, exists := c.data[key.FullKey()]
    if !exists || time.Now().After(item.expiry) {
        return false, nil
    }
    
    return true, nil
}

func (c *InMemoryCache) InvalidateByTag(ctx context.Context, tag string) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    keys, exists := c.tags[tag]
    if !exists {
        return nil
    }
    
    for _, key := range keys {
        delete(c.data, key)
    }
    
    delete(c.tags, tag)
    return nil
}

func (c *InMemoryCache) InvalidateByNamespace(ctx context.Context, namespace string) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    for key, item := range c.data {
        if item != nil && len(key) > len(namespace) && key[:len(namespace)] == namespace {
            // Remove from tag index
            for _, tag := range item.tags {
                keys := c.tags[tag]
                newKeys := make([]string, 0, len(keys))
                for _, k := range keys {
                    if k != key {
                        newKeys = append(newKeys, k)
                    }
                }
                c.tags[tag] = newKeys
            }
            
            delete(c.data, key)
        }
    }
    
    return nil
}

func (c *InMemoryCache) Clear(ctx context.Context) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.data = make(map[string]*cacheItem)
    c.tags = make(map[string][]string)
    
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
            // Remove from tag index
            for _, tag := range item.tags {
                keys := c.tags[tag]
                newKeys := make([]string, 0, len(keys))
                for _, k := range keys {
                    if k != key {
                        newKeys = append(newKeys, k)
                    }
                }
                c.tags[tag] = newKeys
            }
            
            delete(c.data, key)
        }
    }
}
```

## Integration with External Solutions

### Redis Adapter Implementation

For external caching solutions like Redis, we'll create a Redis adapter:

```go
// internal/modules/cache/adapters/redis_cache.go
package adapters

import (
    "context"
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
    "forum/internal/modules/cache/domain"
    "forum/internal/modules/cache/ports"
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

func (r *RedisCache) Get(ctx context.Context, key domain.CacheKey) (interface{}, error) {
    result, err := r.client.Get(ctx, key.FullKey()).Result()
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

func (r *RedisCache) Set(ctx context.Context, key domain.CacheKey, value interface{}, opts *domain.CacheOptions) error {
    ttl := r.defaultTTL
    if opts != nil && opts.TTL > 0 {
        ttl = opts.TTL
    }
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return r.client.Set(ctx, key.FullKey(), data, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key domain.CacheKey) error {
    return r.client.Del(ctx, key.FullKey()).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key domain.CacheKey) (bool, error) {
    result, err := r.client.Exists(ctx, key.FullKey()).Result()
    if err != nil {
        return false, err
    }
    
    return result > 0, nil
}

func (r *RedisCache) InvalidateByTag(ctx context.Context, tag string) error {
    // Redis doesn't support tags natively, so we'd need to maintain our own tag index
    // This can be done using Redis sets with SADD/SREM operations
    return nil
}

func (r *RedisCache) InvalidateByNamespace(ctx context.Context, namespace string) error {
    // Find keys by pattern and delete them
    keys, err := r.client.Keys(ctx, namespace+":*").Result()
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
```

### Redis Configuration and Initialization

```go
// internal/modules/cache/adapters/redis_config.go
package adapters

import (
    "github.com/go-redis/redis/v8"
)

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

### Dependency Injection with Wire

```go
// internal/modules/cache/wire.go (or integrated into existing wire package)
package cache

import (
    "github.com/go-redis/redis/v8"
    "forum/internal/modules/cache/adapters"
    "forum/internal/modules/cache/application"
    "forum/internal/modules/cache/ports"
)

func InitializeInMemoryCache(defaultTTL int) ports.TypedCacheService {
    cache := adapters.NewInMemoryCache(time.Duration(defaultTTL) * time.Second)
    return application.NewService(cache)
}

func InitializeRedisCache(config adapters.RedisConfig, defaultTTL int) ports.TypedCacheService {
    client := adapters.NewRedisClient(config)
    cache := adapters.NewRedisCache(client, time.Duration(defaultTTL) * time.Second)
    return application.NewService(cache)
}
```

## Implementation Examples

### Using Cache in Post Module

```go
// internal/modules/post/application/service.go
package application

import (
    "context"
    "forum/internal/modules/cache/domain"
    "forum/internal/modules/cache/ports"
    "forum/internal/modules/post/domain"
    "forum/internal/modules/post/ports"
)

type Service struct {
    postRepository ports.PostRepository
    cacheService   ports.TypedCacheService
}

// GetPost retrieves a post by ID, with caching
func (s *Service) GetPost(ctx context.Context, postID int) (*domain.Post, error) {
    // Create cache key
    cacheKey := domain.CacheKey{
        Namespace: "posts",
        Key:       fmt.Sprintf("id:%d", postID),
    }
    
    // Try to get from cache first
    var post *domain.Post
    if err := s.cacheService.GetTyped(ctx, cacheKey, &post); err == nil && post != nil {
        return post, nil
    }
    
    // If not in cache, fetch from repository
    post, err := s.postRepository.GetByID(ctx, postID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for future requests
    opts := &domain.CacheOptions{
        TTL:  time.Hour, // Cache for 1 hour
        Tags: []string{"posts"}, // Add tag for bulk invalidation
    }
    
    if err := s.cacheService.SetTyped(ctx, cacheKey, post, opts); err != nil {
        // Log error but don't fail the operation
        log.Printf("Failed to cache post: %v", err)
    }
    
    return post, nil
}

// CreatePost creates a new post and invalidates related caches
func (s *Service) CreatePost(ctx context.Context, post *domain.Post) error {
    // Create the post in the repository
    err := s.postRepository.Create(ctx, post)
    if err != nil {
        return err
    }
    
    // Invalidate posts list cache
    if err := s.cacheService.InvalidateByTag(ctx, "posts-list"); err != nil {
        log.Printf("Failed to invalidate posts list cache: %v", err)
    }
    
    return nil
}
```

### Configuration

```go
// internal/platform/config/config.go (updated)
type CacheConfig struct {
    Type         string        // "inmemory" or "redis"
    DefaultTTL   time.Duration // Default time-to-live for cache entries
    InMemorySize int          // For in-memory cache configuration
    
    // Redis-specific configuration
    Redis RedisConfig
}

type RedisConfig struct {
    Address  string
    Password string
    DB       int
    PoolSize int
}

// Add to main Config struct
type Config struct {
    // ... other config fields
    Cache CacheConfig
}
```

## Usage in Other Modules

The cache module will be available for other modules to use while maintaining clean architecture principles:

1. **Posts Module**: Cache frequently accessed posts to reduce database load
2. **User Module**: Cache user profiles and permissions
3. **Comment Module**: Cache comment threads
4. **Reaction Module**: Cache reaction counts
5. **Auth Module**: Cache session information

Each module will depend on the cache module's ports interface, not its implementation, maintaining proper separation of concerns.

## Configuration

The caching implementation will be configurable through environment variables:

```env
# Cache configuration
CACHE_TYPE=redis          # "redis" or "inmemory"
CACHE_DEFAULT_TTL=3600    # Default TTL in seconds
CACHE_MEMORY_SIZE=1000    # For in-memory cache size

# Redis configuration
REDIS_ADDRESS=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

## Conclusion

This caching layer proposal provides a flexible, extensible caching solution that fits well with the existing hexagonal architecture of the forum application. The design allows for:

1. **Easy switching** between different caching backends (in-memory, Redis, etc.)
2. **Type safety** through the TypedCacheService interface
3. **Tag-based invalidation** for complex cache invalidation scenarios
4. **Namespace support** for organized cache key management
5. **Automatic cleanup** of expired entries
6. **Integration-ready** interfaces that other modules can depend on without tight coupling

The implementation follows the same architectural patterns as the rest of the application, ensuring consistency and maintainability. The cache module can be developed independently and integrated gradually into existing modules as needed.