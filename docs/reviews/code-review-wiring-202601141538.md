# Code Review: Modules Wiring

**Date:** 2026-01-14 15:38
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The wiring logic in `cmd/forum/wire` is clean and centralized. It follows the Dependency Injection (DI) pattern manually (without a DI framework), which is idiomatic and simple for Go. The separation of concerns is excellent:

- `app.go`: Lifecycle (Start/Stop) and high-level orchestration.
- `repositories.go`: Initializes DB adapters.
- `services.go`: Wires services together (handling dependency layers).
- `handlers.go`: Injects services into HTTP handlers.

This structure allows for easy testing and swapping of implementations.

## Critical Issues (Must Fix)

- **None Identified.** The wiring code is robust. It properly propagates errors during initialization (e.g., DB connection failure) and ensures graceful shutdown.

## Performance & Optimization

- **PERF-1: Template Parsing**
  - **Description:** `template.ParseGlob("templates/*.html")` parses all templates at startup. This is correct for performance (fail fast, parse once).
  - **Note:** If the template set grows huge, this might slow down startup slightly, but for a forum, it's negligible.

## Nitpicks & Best Practices

- **Hardcoded Paths:**

  - `template.ParseGlob("templates/*.html")`
  - `os.Stat("./static")`
  - `http.FileServer(http.Dir("./static"))`
  - `cfg.Database.MigrationsDir` usage in `app.go`
    **Recommendation:** Make these paths configurable via `config.Config` to support different deployment environments (e.g., Docker containers where paths might differ).

- **ServiceContainer:** The `ServiceContainer` is a large struct implementing accessors for every service. This acts somewhat like a "Service Locator" anti-pattern but restricted to the composition root, which is acceptable. Ideally, handlers should only receive the specific interfaces they need, not the whole container. However, `handlers.go` does pass the container to `NewHTTPHandler`, which then likely (in specific module adapters) takes just what it needs or the whole container. Looking at valid module code (e.g., `user/adapters/http_handler.go`), it takes `ServiceContainer` interface. This couples handlers to the container shape.
  - **Refactor (Low Priority):** Inject specific interfaces into `NewHTTPHandler` (e.g., `NewHTTPHandler(s.User(), s.Auth())`) instead of the whole container.

---
