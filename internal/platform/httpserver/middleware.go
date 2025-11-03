// Package httpserver - middleware
// This file provides common HTTP middleware for the application.
// Middleware includes logging, recovery, rate limiting, CORS, etc.
package httpserver

import (
	"net/http"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain chains multiple middleware together.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Recovery recovers from panics and returns 500 Internal Server Error.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement recovery middleware
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs incoming HTTP requests.
func RequestLogger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement request logging
			next.ServeHTTP(w, r)
		})
	}
}

// CORS handles Cross-Origin Resource Sharing.
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement CORS middleware
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit implements rate limiting per IP or user.
func RateLimit(requestsPerMinute int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement rate limiting
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds security-related HTTP headers.
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement security headers (CSP, X-Frame-Options, etc.)
			next.ServeHTTP(w, r)
		})
	}
}
