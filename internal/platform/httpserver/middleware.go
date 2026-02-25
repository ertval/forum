// Package httpserver provides HTTP server setup and middleware management.
package httpserver

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"forum/internal/platform/logger"
)

// Middleware is a function that wraps an http.Handler.
// It can modify the request/response or perform actions before/after the handler.
type Middleware func(http.Handler) http.Handler

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader captures the status code and delegates to the underlying writer.
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write captures the number of bytes written and delegates to the underlying writer.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Chain creates a middleware chain from multiple middleware functions.
// Middleware is executed in the order provided.
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// Recovery middleware recovers from panics and returns a 500 error.
func Recovery(lgr *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Convert error to string for logging
					var errMsg string
					switch e := err.(type) {
					case error:
						errMsg = e.Error()
					case string:
						errMsg = e
					default:
						errMsg = fmt.Sprintf("%v", e)
					}

					lgr.Error("panic.recovered",
						logger.String("path", r.URL.Path),
						logger.String("method", r.Method),
						logger.String("remote", r.RemoteAddr),
						logger.String("error", errMsg),
						logger.String("stack", string(debug.Stack())),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logger middleware logs HTTP requests and responses.
func Logger(lgr *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			if rw.status == 0 {
				rw.status = http.StatusOK
			}

			// Determine protocol (http or https)
			proto := "http"
			if r.TLS != nil {
				proto = "https"
			}

			lgr.Info("http.request",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("query", r.URL.RawQuery),
				logger.Int("status", rw.status),
				logger.Int("size", rw.size),
				logger.Duration("duration_ms", time.Since(start)),
				logger.String("remote", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
				logger.String("proto", proto),
			)
		})
	}
}

// CORS middleware adds CORS headers to responses.
// This is a generic HTTP concern that doesn't require business logic.
func CORS(allowedOrigins []string) Middleware {
	// Build origin lookup map for O(1) checks
	originSet := make(map[string]bool)
	allowAll := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
		originSet[origin] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Determine if origin is allowed
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && originSet[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			// Set common CORS headers
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// rateLimiter tracks request counts per IP address.
type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

// RateLimit middleware limits the number of requests per time window.
// This is IP-based rate limiting that doesn't require user context.
// For user-based rate limiting, use auth module middleware.
func RateLimit(requests int, windowSeconds int) Middleware {
	limiter := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requests,
		window:   time.Duration(windowSeconds) * time.Second,
	}

	// Cleanup goroutine to prevent memory leak
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP (handle proxies)
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = forwarded
			}

			if !limiter.allow(clientIP) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", windowSeconds))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// allow checks if a request from the given IP is allowed.
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Filter out old requests
	var recent []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}

	// Check if under limit
	if len(recent) >= rl.limit {
		rl.requests[ip] = recent
		return false
	}

	// Record this request
	rl.requests[ip] = append(recent, now)
	return true
}

// cleanup removes expired entries to prevent memory growth.
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, times := range rl.requests {
		var recent []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = recent
		}
	}
}

// NOTE: Authentication and Authorization middleware are intentionally NOT here.
// They require business logic (AuthService, UserService) and belong in the
// auth module: internal/modules/auth/adapters/middleware.go
//
// This follows the hexagonal architecture principle:
//   - Platform packages must NOT import module packages
//   - Modules CAN import platform packages
//
// See: auth/adapters/middleware.go for RequireAuth() and OptionalAuth()
