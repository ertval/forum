// Package httpserver provides HTTP server setup and middleware management.
package httpserver

import (
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
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
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Only emit credentials header when echoing a specific origin.
			// The Fetch spec forbids credentials: true with wildcard origin.
			if !allowAll {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiterConfig holds configuration for the rate limiter.
type RateLimiterConfig struct {
	Requests        int           // Maximum requests per window
	Window          time.Duration // Time window for rate limiting
	CleanupInterval time.Duration // How often to clean up expired entries
	MaxEntries      int           // Maximum number of IP entries to prevent DoS
	TrustProxy      bool          // Whether to trust X-Forwarded-For header
}

// DefaultRateLimiterConfig returns sensible defaults for rate limiting.
func DefaultRateLimiterConfig(requests int, windowSeconds int) RateLimiterConfig {
	return RateLimiterConfig{
		Requests:        requests,
		Window:          time.Duration(windowSeconds) * time.Second,
		CleanupInterval: time.Minute,
		MaxEntries:      10000, // Prevent memory exhaustion from IP spoofing
		TrustProxy:      false, // Don't trust proxy headers by default
	}
}

// rateLimiter tracks request counts per IP address using a sliding window.
type rateLimiter struct {
	entries    sync.Map // map[string]*ipEntry - lock-free for reads
	limit      int
	window     time.Duration
	maxEntries int
	trustProxy bool
	entryCount int64 // atomic counter for entries
	done       chan struct{}
}

// ipEntry tracks requests for a single IP with its own lock.
type ipEntry struct {
	mu       sync.Mutex
	requests []time.Time
}

// RateLimit middleware limits the number of requests per time window.
// This is IP-based rate limiting that doesn't require user context.
// For user-based rate limiting, use auth module middleware.
// Returns the middleware and a stop function to shut down the cleanup goroutine.
func RateLimit(requests int, windowSeconds int) (Middleware, func()) {
	return RateLimitWithConfig(DefaultRateLimiterConfig(requests, windowSeconds))
}

// RateLimitWithConfig creates a rate limiter with custom configuration.
// Returns the middleware and a stop function that shuts down the cleanup goroutine.
// The stop function should be called during graceful shutdown.
func RateLimitWithConfig(cfg RateLimiterConfig) (Middleware, func()) {
	limiter := &rateLimiter{
		limit:      cfg.Requests,
		window:     cfg.Window,
		maxEntries: cfg.MaxEntries,
		trustProxy: cfg.TrustProxy,
		done:       make(chan struct{}),
	}

	// Cleanup goroutine with graceful shutdown support
	go func() {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				limiter.cleanup()
			case <-limiter.done:
				return // Exit on shutdown signal
			}
		}
	}()

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r, limiter.trustProxy)

			if !limiter.allow(clientIP) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(cfg.Window.Seconds())))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	return middleware, limiter.Stop
}

// getClientIP extracts the client IP address from the request.
// Only trusts X-Forwarded-For if trustProxy is true (behind known proxy).
func getClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		// X-Forwarded-For may contain multiple IPs: client, proxy1, proxy2
		// The first IP is the original client
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		// X-Real-IP is set by some proxies (nginx)
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}
	// Strip port from RemoteAddr (format: "IP:port" or "[IPv6]:port")
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// allow checks if a request from the given IP is allowed.
func (rl *rateLimiter) allow(ip string) bool {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get or create entry for this IP
	val, loaded := rl.entries.LoadOrStore(ip, &ipEntry{
		requests: []time.Time{now},
	})

	if !loaded {
		// New entry - check if we're at capacity
		newCount := atomic.AddInt64(&rl.entryCount, 1)
		if newCount > int64(rl.maxEntries) {
			// At capacity - remove the entry we just added and reject
			rl.entries.Delete(ip)
			atomic.AddInt64(&rl.entryCount, -1)
			return false // Reject to prevent DoS via IP spoofing
		}
		return true // First request for this IP, always allowed
	}

	entry := val.(*ipEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	// Filter out old requests (sliding window)
	recent := entry.requests[:0] // Reuse backing array
	for _, t := range entry.requests {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}

	// Check if under limit
	if len(recent) >= rl.limit {
		entry.requests = recent
		return false
	}

	// Record this request
	entry.requests = append(recent, now)
	return true
}

// cleanup removes expired entries to prevent memory growth.
func (rl *rateLimiter) cleanup() {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	rl.entries.Range(func(key, value interface{}) bool {
		entry := value.(*ipEntry)
		entry.mu.Lock()

		// Filter to only recent requests
		recent := entry.requests[:0]
		for _, t := range entry.requests {
			if t.After(windowStart) {
				recent = append(recent, t)
			}
		}

		if len(recent) == 0 {
			entry.mu.Unlock()
			rl.entries.Delete(key)
			atomic.AddInt64(&rl.entryCount, -1)
		} else {
			entry.requests = recent
			entry.mu.Unlock()
		}

		return true // Continue iteration
	})
}

// Stop gracefully shuts down the rate limiter's cleanup goroutine.
func (rl *rateLimiter) Stop() {
	close(rl.done)
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
