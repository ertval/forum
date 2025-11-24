# Dockerfile for the forum application
# Uses multi-stage build for security and minimal image size

# == Build stage ==
# Use specific Go 1.25.0 and Alpine 3.20 versions for reproducible builds and security
FROM golang:1.25.0-alpine3.20 AS builder

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

# Change ownership of all files to the non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user for security (principle of least privilege)
USER appuser

# Expose port 8080 for HTTP traffic
EXPOSE 8080

# Run the application using exec form (preferred over shell form for signal handling)
CMD ["./forum"]
