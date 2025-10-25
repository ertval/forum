# Dockerfile for the forum application

# == Build stage ==
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o forum ./cmd/forum

# == Runtime stage ==
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

# Create a non-root user
RUN adduser -D appuser

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/forum .

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/templates ./internal/templates
COPY --from=builder /app/internal/database/schema.sql ./internal/database/schema.sql

# Change ownership to the non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./forum"]
