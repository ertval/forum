// Package http - middleware
// This file provides authentication middleware for protecting routes.
package http

import (
	"net/http"

	"forum/internal/modules/auth/ports/input"
)

// AuthMiddleware provides authentication middleware.
type AuthMiddleware struct {
	authService input.AuthService
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(authService input.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth is a middleware that requires authentication.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement auth middleware
		// 1. Get session token from cookie
		// 2. Validate session
		// 3. Add user ID to request context
		// 4. Call next handler or return 401
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth is a middleware that loads user info if authenticated but doesn't require it.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement optional auth middleware
		next.ServeHTTP(w, r)
	})
}
