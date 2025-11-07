# Caching Layer Proposal

## Goals

- Reduce read latency for high-traffic endpoints (posts, comments, reaction counts) while respecting domain rules.
- Encapsulate caching concerns behind explicit ports so business logic stays storage-agnostic.
- Support both local in-process caching and remote distributed caching (Redis) without altering module APIs.

## Scope & Principles

- **Read-mostly optimisation**: cache data that is expensive to compute or fetch but inexpensive to invalidate.
- **Exactness over staleness**: prefer short TTLs and explicit invalidation hooks; no write-behind patterns in phase one.
- **Module autonomy**: each module decides what to cache via dedicated output ports; shared cache implementation lives in the platform layer.
- **Configurable**: enable/disable caching and switch providers via configuration without code changes.

## Architectural Placement

```text
internal/
  platform/
    cache/
      store.go          # Core interfaces (ports) + errors
      config.go         # Cache-specific config struct + defaults
      metrics.go        # Optional instrumentation helpers
      inmemory/
        store.go        # INPUT ADAPTER - Cache Provider (in-memory)
      redis/
        store.go        # INPUT ADAPTER - Cache Provider (Redis)
```

- `store.go` exposes interfaces such as `KeyValueStore`, `NamespacedStore`, and helpers for typed reads (JSON marshalling handled here).
- Each provider subpackage implements the shared interface. Files include header comments: `// INPUT ADAPTER - Cache Provider`.
- The `platform/cache` package is treated similarly to existing platform utilities (logger, validator) and injected via the wire package.

## Module Interaction

1. **Define cache ports**: modules needing caching add an output port for cache operations.
   - Example (`internal/modules/post/ports/cache_repository.go`):

     ```go
     // OUTPUT PORT - Cache Repository Interface
     type PostCacheRepository interface {
         FetchPost(ctx context.Context, id string) (*domain.Post, error)
         StorePost(ctx context.Context, post *domain.Post, ttl time.Duration) error
         InvalidatePost(ctx context.Context, id string) error
     }
     ```

2. **Application service orchestration**: the post service injects both the storage repository and cache repository, checking the cache before falling back to SQLite and rehydrating on misses.
3. **Adapters**: a new adapter in `internal/modules/post/adapters/post_cache_repository.go` wraps `platform/cache` to satisfy the port. Writes call `InvalidatePost` after persisting changes.

## In-Memory Provider

- Implementation: thread-safe map with TTL, backed by `sync.Map` plus a background janitor goroutine (interval configurable).
- Config struct:

  ```go
  type InMemoryConfig struct {
      MaxEntries int           `env:"CACHE_INMEMORY_MAX_ENTRIES" default:"1000"`
      DefaultTTL time.Duration `env:"CACHE_INMEMORY_DEFAULT_TTL" default:"60s"`
      SweepEvery time.Duration `env:"CACHE_INMEMORY_SWEEP_EVERY" default:"30s"`
  }
  ```

- Best suited for single-instance deployments or local development; production can fall back to it when Redis is unavailable.

## Redis Provider (External Example)

- Dependency: `github.com/redis/go-redis/v9` (minimal, well-supported).
- Config struct:

  ```go
  type RedisConfig struct {
      Addr         string        `env:"CACHE_REDIS_ADDR" default:"localhost:6379"`
      Username     string        `env:"CACHE_REDIS_USERNAME"`
      Password     string        `env:"CACHE_REDIS_PASSWORD"`
      DB           int           `env:"CACHE_REDIS_DB" default:"0"`
      DialTimeout  time.Duration `env:"CACHE_REDIS_DIAL_TIMEOUT" default:"5s"`
      ReadTimeout  time.Duration `env:"CACHE_REDIS_READ_TIMEOUT" default:"1s"`
      WriteTimeout time.Duration `env:"CACHE_REDIS_WRITE_TIMEOUT" default:"1s"`
  }
  ```

- Wiring (step 3 of DI) inside `cmd/forum/wire/repositories.go`:

  ```go
  cacheStore, err := cacheRedis.NewStore(cacheConfig.Redis)
  if err != nil {
      return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to connect redis")
  }
  ```

- Redis keys use module-aware prefixes (e.g. `post:{id}`, `post:list:{filtersHash}`). Batch reads can leverage pipelines for reaction counts or paginated post listings.

## Configuration Wiring

- Extend `internal/platform/config/config.go` with a `Cache` block exposing provider type (`none`, `inmemory`, `redis`) plus nested provider configs.
- Update `cmd/forum/wire/app.go` to parse cache config, instantiate the chosen provider, and fall back to a no-op store when disabled.
- Inject `cacheStore` into modules via new adapter constructors in `cmd/forum/wire/services.go` and `cmd/forum/wire/handlers.go` as needed.
- Respect `CACHE_PROVIDER=none` to turn caching off without code changes.

## Testing Strategy

- **Unit tests**: cover `platform/cache/inmemory` eviction, TTL handling, and serialization; for Redis, run against a disposable Redis instance using docker or `redis-server --save ''`.
- **Service tests**: mock cache repository to simulate hits/misses, ensuring invalidation happens on writes and updates.
- **Integration tests**: default to in-memory provider; optionally run a separate test job with Redis enabled in `docker-compose` (`redis:alpine`) to verify wiring.
- **Load testing**: extend `tests/integration` benchmarks or add scripts to confirm cache warm-up reduces repeated request latency; document thresholds alongside audit requirements.

## Rollout Plan

1. **Phase 1**: Implement platform cache package with in-memory and no-op stores. Add cache port for post-by-id reads.
2. **Phase 2**: Expand caching to comment threads and reaction counts; add hit/miss metrics and structured logging.
3. **Phase 3**: Introduce Redis provider, configuration, and docker-compose support; update deployment docs and monitoring plan (Redis INFO logs).
4. **Phase 4**: Evaluate advanced patterns (write-through for aggregates, background warming) once baseline stability is confirmed.

## Additional Considerations

- Security: do not cache sensitive data such as auth sessions or password reset tokens.
- Consistency: prefer explicit invalidation hooks triggered by writes rather than relying on long TTLs.
- Observability: expose cache metrics via existing logger; design metrics hooks for future Prometheus integration.
- Resilience: if Redis is unavailable, adapters wrap errors as `ErrCacheUnavailable` and allow the request to proceed against the database.

## Next Steps Checklist

- [ ] Create `internal/platform/cache` package skeleton with interfaces, configuration structs, and a no-op store.
- [ ] Implement in-memory provider and integrate with the post module via a cache adapter.
- [ ] Update wire package and configuration to instantiate and inject cache stores.
- [ ] Document new environment variables in `README.md` and `docs/IMPLEMENTATION_ROADMAP.md`.
- [ ] Add Redis provider, docker-compose service, and optional integration test path.
