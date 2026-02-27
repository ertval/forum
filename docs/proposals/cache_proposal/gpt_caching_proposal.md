## Caching layer proposal for Forum

This document proposes a practical, low-risk caching layer for the Forum project. It covers
where the cache code should live, recommended API/contract, supported adapters (in-memory and Redis),
integration/wiring with existing DI, common caching patterns, invalidation strategies, testing, and monitoring.

## Summary / Recommendation

- Implement caching as a platform-level cross-cutting service under `internal/platform/cache`.
- Provide a single `Cache` port (interface) used by application services that need caching (posts, comments, reactions, sessions where appropriate).
- Implement at least two adapters:
  - `inmemory` (LRU, single-process) for local development and low-latency transient caching.
  - `redis` (distributed) for production and multi-instance deployments.
- Use cache-aside as the default pattern (read-through is optional for specific hot-warm flows).
- Keep serialization simple (JSON for typed objects) and avoid caching large raw file blobs (images kept in static/uploads).

## Why a platform package (not a module)

Caching is a cross-cutting concern used by multiple modules (auth, post, comment, reaction). The project already stores cross-cutting concerns in `internal/platform` (logger, httpserver, database). Placing caching under `internal/platform/cache` keeps it available to all modules without breaking the module boundaries (ports/adapters) used for domain logic.

If you prefer a module approach, create `internal/modules/cache` with the 4-layer pattern — but for simplicity and to match existing cross-cutting code, `internal/platform/cache` is preferred.

## Proposed package structure

internal/platform/cache/

- cache.go                // public cache port (interface) + types
- options.go              // configuration structs, default options
- inmemory_adapter.go     // lightweight LRU TTL adapter (single-process)
- redis_adapter.go        // redis-based adapter implementation
- metrics.go              // optional metrics hooks (prometheus counters/histograms)

Notes:

- Keep adapters as thin wrappers implementing the `Cache` interface. Domain modules should rely on the interface only.

## Contract / Interface (concept)

Cache responsibilities (high-level contract):

- Get(ctx, key) (value []byte, found bool, err error)
- Set(ctx, key, value []byte, ttl time.Duration) error

# Caching layer proposal for Forum

This document proposes a practical, low-risk caching layer for the Forum project. It describes placement, the public contract, adapters (in-memory and Redis), wiring, operations, invalidation strategies, testing and monitoring suggestions.

## Overview

- Place caching as a platform-level concern in `internal/platform/cache` so it is available to all modules (auth, post, comment, reaction) without breaking module boundaries.
- Provide a simple, well-documented interface (port) used by application services. Keep adapters (in-memory, Redis) thin and interchangeable.
- Default pattern: cache-aside. Use short TTLs for dynamic content and explicit invalidation on writes.

## Where to add code

Recommended path and files:

- `internal/platform/cache/cache.go` — Cache interface and common types
- `internal/platform/cache/options.go` — configuration and constructor options
- `internal/platform/cache/inmemory_adapter.go` — single-process LRU + TTL (dev / fallback)
- `internal/platform/cache/redis_adapter.go` — production Redis adapter
- `internal/platform/cache/metrics.go` — optional metrics hooks (Prometheus)

This keeps caching with other cross-cutting code (logger, httpserver, database) in `internal/platform`.

## Public contract (interface)

The cache interface should be small and explicit. Example:

```go
package cache

import (
  "context"
)

// Cache is a minimal cache abstraction used by application services.
type Cache interface {
  // Get returns the stored bytes, whether it was found, and an error if something went wrong.
  Get(ctx context.Context, key string) (value []byte, found bool, err error)

  // Set stores the value for key with an explicit TTL in seconds. ttlSeconds==0 means no expiration.
  Set(ctx context.Context, key string, value []byte, ttlSeconds int) error

  // Delete removes a key.
  Delete(ctx context.Context, key string) error

  // Optional atomic counters (implement only if needed by your use-case)
  Incr(ctx context.Context, key string) (int64, error)
  Decr(ctx context.Context, key string) (int64, error)
}
```

Design notes:

- Always accept `context.Context` for request-scoped deadlines and cancellation.
- Keep value type as `[]byte` (serialized JSON or gob) so adapters remain generic.

## Key naming and TTL strategy

- Use deterministic keys with namespaces: `namespace:entity:version:id[:sub]`. Example keys:
  - `post:item:v1:12345`
  - `post:list:category:news:v1:page:1`
  - `reaction:count:v1:post:12345`
- Version keys when serialization or structure changes to allow safe rollouts.
- TTL guidance (examples):
  - Post item: 300s (5m) or 60s for highly dynamic sites
  - Post list / feeds: 30s–120s
  - Reaction counts: 5s–60s (eventual consistency acceptable)
  - Sessions: prefer authoritative DB; cache token -> user-id only when helpful and with short TTL (< 60s)

## Caching patterns (recommended default)

- Cache-aside (application checks cache first; on miss, loads from DB and sets cache) — simplest and safest.
- Read-through (adapter fetches from store if miss) — centralizes logic but increases adapter complexity.
- Write-through/write-back — avoid unless you need synchronous guarantees; prefer write-through only for small items with strict consistency and when you control invalidation.

## Invalidation strategies

- Explicit invalidation on writes: after creating/updating/deleting domain objects, call `cache.Delete` for the affected keys (single-item, relevant list caches). This is explicit and easy to reason about.
- When many keys reference the same entity (e.g., lists & feeds), consider invalidating a small set (primary list keys) or use short TTLs for derived caches.
- For counters, consider using Redis `INCR` and a reconciliation job to persist to DB periodically.

## Adapters

1) In-memory adapter (development / fallback):

- Single-process LRU with TTL (e.g., `github.com/hashicorp/golang-lru` or a small map with eviction).
- Pros: zero external deps, simplest to run.
- Cons: not shared between instances.

2) Redis adapter (production):

- Use a stable client like `github.com/redis/go-redis/v9`.
- Pros: distributed, supports atomic ops (INCR), TTLs, persistence, and pub/sub for invalidation signals.
- Cons: operational complexity, network calls, availability concerns (use HA Redis or managed service).

### Example: Redis adapter snippet (concept)

```go
import (
  "context"
  "time"

  "github.com/redis/go-redis/v9"
)

type RedisAdapter struct{ client *redis.Client }

func NewRedisAdapter(opts *redis.Options) *RedisAdapter {
  return &RedisAdapter{client: redis.NewClient(opts)}
}

func (r *RedisAdapter) Get(ctx context.Context, key string) ([]byte, bool, error) {
  cmd := r.client.Get(ctx, key)
  b, err := cmd.Bytes()
  if err == redis.Nil {
    return nil, false, nil
  }
  if err != nil {
    return nil, false, err
  }
  return b, true, nil
}

func (r *RedisAdapter) Set(ctx context.Context, key string, value []byte, ttlSeconds int) error {
  ttl := time.Duration(ttlSeconds) * time.Second
  return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisAdapter) Delete(ctx context.Context, key string) error {
  return r.client.Del(ctx, key).Err()
}
```

## Wiring (DI) and configuration

Add config options to `internal/platform/config/config.go` (example fields):

- `Cache.Enabled` (bool)
- `Cache.Adapter` ("inmemory" | "redis")
- `Cache.Redis.Address`, `Cache.Redis.Password`, `Cache.Redis.DB`, `Cache.Redis.PoolSize`

Wiring in `cmd/forum/wire/app.go` (concept):

1. Read config → initialize logger
2. Initialize database
3. Initialize cache adapter (after DB and logger):

```go
// pseudo-code inside wire
var platformCache cache.Cache
if cfg.Cache.Adapter == "redis" {
  redisOpts := &redis.Options{Addr: cfg.Cache.Redis.Address, Password: cfg.Cache.Redis.Password, DB: cfg.Cache.Redis.DB}
  platformCache = cache.NewRedisAdapter(redisOpts)
} else {
  platformCache = cache.NewInMemoryAdapter(....)
}

postService := postApp.NewPostService(postRepo, platformCache, ...)
```

Follow the project's canonical wiring order in `cmd/forum/wire/` and register cache creation before service creation.

## Example integration: caching Post summaries

Service pseudocode (cache-aside):

1. Build key: `post:item:v1:<id>`
2. Try `cache.Get(ctx, key)`
3. If found, deserialize and return
4. Else, load from DB, serialize, `cache.Set(ctx, key, bytes, ttl)` and return

Invalidate on update/delete by calling `cache.Delete(ctx, key)` for the item and any related list keys.

## Tests & Validation

- Unit tests: mock the `Cache` interface to test application services behavior on cache hits/misses and failure modes.
- Integration tests: run with the in-memory adapter and a Redis instance (use `testcontainers` or a local Redis during CI). Test TTLs and invalidation behavior.
- Quality gates:
  - Build: PASS (`go build ./...`)
  - Lint/Format: PASS (`go fmt` / `golangci-lint` as configured)
  - Tests: add unit tests for cache-using services and adapter tests for Redis.

## Metrics & Observability

- Export cache hits/misses, latency, and error counts (Prometheus counters/histograms).
- Log cache errors at debug/info level; avoid returning cache errors as fatal to users — fall back to DB.
- For Redis, monitor connection pool, memory usage, evictions, and slow logs.

## Edge cases & pitfalls

- Don’t cache very large objects (images) — keep caching to computed views or small JSON blobs.
- Beware of stampeding herd on cache expiry — consider small random jitter on TTLs or use request coalescing for hot keys.
- Ensure cache failures degrade gracefully: if cache adapter errors, fall back to DB reads/writes.

## External Redis deployment options

- Managed Redis (e.g., AWS ElastiCache, Google Memorystore, Redis Cloud) — low ops overhead.
- Self-hosted Redis in HA mode (replication with sentinel or clustering) — more control, more ops.

Connection notes:

- Use TLS and auth for production Redis.
- Use client-side pooling and reasonable timeouts. Configure retry/backoff policies.

## Next steps / Implementation plan

1. Create `internal/platform/cache` package and implement in-memory adapter (dev / test) — quick win.
2. Add configuration and wiring in `cmd/forum/wire/` to construct adapter based on env/config.
3. Implement Redis adapter and integration tests.
4. Add cache usage to one service (e.g., `post` summaries) with unit + integration tests.
5. Add metrics and a monitoring dashboard for cache behavior.

## Appendix: Example dependencies

- Go module: `github.com/redis/go-redis/v9` for Redis client
- Optionally `github.com/hashicorp/golang-lru` for in-memory LRU cache

---

If you'd like, I can scaffold `internal/platform/cache` with the interface and the in-memory adapter and add a small wiring example in `cmd/forum/wire/` next. Which adapter should I implement first (in-memory or Redis)?