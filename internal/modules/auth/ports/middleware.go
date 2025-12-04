// INPUT PORT - Middleware Interface
// Package ports defines middleware contracts for the auth module.
package ports

import (
	"context"
	"net/http"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the context key for user ID (public UUID).
	UserIDKey ContextKey = "user_id"
	// UsernameKey is the context key for username.
	UsernameKey ContextKey = "username"
)

// AuthMiddleware provides authentication middleware.
// This interface allows modules to use auth middleware without importing auth/adapters.
type AuthMiddleware interface {
	// RequireAuth returns middleware that requires authentication.
	// It validates the session token and adds user PUBLIC ID (UUID) to the context.
	// If authentication fails, it returns 401 Unauthorized.
	RequireAuth() Middleware

	// OptionalAuth returns middleware that optionally validates authentication.
	// It validates the session token if present and adds user info to context.
	// If authentication fails or is not present, it continues without error.
	OptionalAuth() Middleware
}

// GetUserID extracts user PUBLIC ID (UUID) from request context.
// Returns empty string if not authenticated.
// SECURITY: Returns UUID string, never internal INT ID.
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUsername extracts username from request context.
// Returns empty string if not authenticated.
func GetUsername(ctx context.Context) string {
	username, ok := ctx.Value(UsernameKey).(string)
	if !ok {
		return ""
	}
	return username
}

// IsAuthenticated checks if the request is authenticated.
func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != ""
}
