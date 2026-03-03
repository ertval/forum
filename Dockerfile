# Dockerfile for the forum application
# Uses multi-stage build for security and minimal image size

# == Build stage ==
# Use a stable Go 1.24 Alpine image for reproducible builds and security
FROM golang:1.24-alpine AS builder

# Install build dependencies required for CGO and SQLite compilation
# --no-cache: Don't store package manager cache to reduce image size
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory for all subsequent commands
WORKDIR /app

# Copy Go module files first for better Docker layer caching
COPY go.mod go.sum ./

# Download Go dependencies separately to leverage Docker layer caching
RUN go mod download

# Copy source code after dependencies to avoid re-downloading on code changes
COPY . .

# Build the application with optimizations
# CGO_ENABLED=1 for SQLite, GOOS=linux target, -ldflags="-w -s" strips debug info
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-w -s -extldflags '-static'" -o forum ./cmd/forum

# == Runtime stage ==
# Use specific Alpine 3.20 for minimal attack surface and reproducible builds
FROM alpine:3.20

# Install minimal runtime dependencies and create non-root user
# ca-certificates: HTTPS/TLS, sqlite-libs: database, tzdata: timezones
# --no-cache reduces image size by not storing package manager cache
RUN apk add --no-cache ca-certificates sqlite-libs tzdata && \
    adduser -D -s /bin/sh appuser

# Set working directory for the application
WORKDIR /app

# Copy the compiled binary from build stage
COPY --from=builder /app/forum .

# Copy static assets, HTML templates, and database migrations
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/migrations ./migrations

# Create data and upload directories and set ownership at build time.
# Since docker-compose mounts these as bind volumes the ownership is preserved.
RUN mkdir -p data static/uploads && \
    chown -R appuser:appuser /app

# Declare volumes for data persistence across container restarts
VOLUME ["/app/data", "/app/static/uploads"]

# Drop to non-root user for all subsequent commands and at runtime
USER appuser

# Expose HTTP and HTTPS ports
EXPOSE 8080 8443

# Run the application using exec form (preferred over shell form for signal handling)
CMD ["./forum"]
