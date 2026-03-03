# Platform Layer (Shared Infrastructure)

The platform directory provides shared cross-cutting concerns and infrastructure components for the entire application.

Unlike `internal/modules`, platform services are purely infrastructure-focused and have NO business logic.

## Services

| Component | Responsibility |
|-----------|----------------|
| **async** | Worker pools and background task processing orchestration. |
| **config** | Environment variable loading, validation, and typed configuration mapping via `.env`. |
| **database** | SQLite connection lifecycle management and automatic/manual migration runner. |
| **errors** | Common HTTP error handling, standardized JSON error shape payload generation. |
| **health** | Health check status validation endpoints for runtime readiness signals. |
| **httpserver**| HTTP server multiplexer, request timeouts, gracefully shutdown, TLS, security headers middleware, rate limiting. |
| **logger** | Highly-structured JSON logging service injected throughout the app layers. |
| **templates**| HTML template compilation, layout application, and rendering helpers. |
| **upload** | File/Image upload processing mechanisms, static persistence, image resizing/validation. |
| **validator**| Reusable validation mechanisms (e.g., email struct validation, password complexity). |

## Usage Principles

- **Abstract When Necessary**: Modules should define an interface for these capabilities if they intend to be decoupled from the framework (e.g., `ports/logger.go`).
- **Initialization**: Platform services are initialized exactly ONE time in `cmd/forum/main.go` or `cmd/forum/wire/app.go` and injected downwards into the `adapters` or `application` layers.
- **Stateless Components**: Methods in `platform` should ideally be thread-safe or manage synchronization over singleton resources.
