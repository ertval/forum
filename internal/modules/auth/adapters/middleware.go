// Package adapters provides HTTP middleware for authentication.
package adapters

import (
	"context"
	"fmt"
	"forum/internal/modules/auth/ports"
	"net/http"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// UserIDKey is the context key for user ID.
	UserIDKey contextKey = "user_id"
	// UsernameKey is the context key for username.
	UsernameKey contextKey = "username"
)

// RequireAuth is middleware that requires authentication.
// It validates the session token and adds user information to the context.
// If authentication fails, it returns 401 Unauthorized.
func RequireAuth(authService ports.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session token from cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate session
			session, err := authService.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user ID to context (convert int to string)
			ctx := context.WithValue(r.Context(), UserIDKey, fmt.Sprintf("%d", session.UserID))

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth is middleware that optionally validates authentication.
// It validates the session token if present and adds user information to the context.
// If authentication fails or is not present, it continues without error.
func OptionalAuth(authService ports.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get session token from cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				// No session token, continue as guest
				next.ServeHTTP(w, r)
				return
			}

			// Try to validate session
			session, err := authService.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				// Invalid session, continue as guest
				next.ServeHTTP(w, r)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, fmt.Sprintf("%d", session.UserID))

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from request context.
// Returns empty string if not authenticated.
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
