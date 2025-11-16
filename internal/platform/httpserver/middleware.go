// Package httpserver provides HTTP server setup and middleware management.
package httpserver

import (
	"fmt"
	"net/http"
	"runtime/debug"
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

			lgr.Info("http.request",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("query", r.URL.RawQuery),
				logger.Int("status", rw.status),
				logger.Int("size", rw.size),
				logger.Duration("duration_ms", time.Since(start)),
				logger.String("remote", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// CORS middleware adds CORS headers to responses.
// TODO: Implement CORS handling.
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Implementation placeholder
			// Set CORS headers
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit middleware limits the number of requests per time window.
// TODO: Implement rate limiting.
func RateLimit(requests int, window int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Implementation placeholder
			// Check rate limit and reject if exceeded
			next.ServeHTTP(w, r)
		})
	}
}

// Authentication middleware checks if the request has a valid session.
// TODO: Implement session validation.
func Authentication() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Implementation placeholder
			// Validate session cookie
			// Add user info to request context
			next.ServeHTTP(w, r)
		})
	}
}

// Authorization middleware checks if the user has the required permissions.
// TODO: Implement authorization checks.
func Authorization(requiredRole string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Implementation placeholder
			// Check user role from context
			// Reject if insufficient permissions
			next.ServeHTTP(w, r)
		})
	}
}
