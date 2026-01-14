# Code Review: Platform Layer

**Date:** 2026-01-14 15:40
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `platform` layer provides the foundational infrastructure (Database, Logging, Config, HTTP Server). The implementation is high-quality, idiomatic Go, with attention to detail in logging (structured + human readable), security (TLS, headers), and configuration validation.

## Critical Issues (Must Fix)

- **ISSUE-1: Memory Leaks in Rate Limiter**
  - **Location:** `internal/platform/httpserver/middleware.go`
  - **Probability:** Medium
  - **Description:** The `cleanup` method for `RateLimit` iterates `rl.requests` and removes expired entries _inside_ the map iteration but only deletes if the slice is empty. If users make 1 request and disappear, they are removed. If they are active, the slice is filtered.
    However, the cleanup ticker runs every minute. If thousands of unique IPs hit the server (DoS or crawler), `requests` map grows unbounded until cleanup runs.
  - **Observation:** `limiter.cleanup` locks the entire map (`rl.mu.Lock()`) for the duration of the iteration. For a large map (busy server), this will cause significant latency spikes (Stop-the-World effect) for all incoming requests waiting on `rl.mu.Lock()` in `allow()`.
  - **Proposed Fix:** Use sharded maps or an LRU cache with TTL for the rate limiter to avoid monolythic locks, or process cleanup in smaller batches.

## Performance & Optimization

- **PERF-1: Logger Reflection/Allocation**

  - **Location:** `internal/platform/logger/logger.go`
  - **Description:** The logger handles many types in `formatHTTPRequest` and `log` using type switches and reflection-like behavior (`fmt.Sprintf("%v")`). It also allocates maps for fields on every log call.
  - **Optimization:** For a high-throughput forum, this is likely fine. If profiling shows issues, switch to `zerolog` or `zap` for zero-allocation logging.

- **PERF-2: Rate Limiter LocK Contention**
  - **Description:** As mentioned in ISSUE-1, the `sync.Mutex` in `rateLimiter` is a single point of contention for **every single HTTP request**.
  - **Fix:** Use `sync.RWMutex` (though writes happen often), or better, channel-based token bucket or atomic counters if strict precision isn't required.

## Nitpicks & Best Practices

- **Security - Upload Path Traversal:** `upload/image.go` correctly checks for path traversal using `filepath.Clean` / `Abs` checks. This is excellent.
- **TLS Configuration:** `httpserver/tls.go` correctly disables old TLS versions and weak ciphers.
- **Config Validation:** `config/config.go` has thorough validation logic.
- **Hardcoded Ticker:** The rate limiter cleanup ticker is hardcoded to `1 minute`. This might not be aggressive enough for high traffic or too aggressive for low traffic.

---
