// Package adapters provides HTTP middleware for authentication.
package adapters

import (
	"context"
	"net/http"

	authPorts "forum/internal/modules/auth/ports"
	userPorts "forum/internal/modules/user/ports"
)

// MiddlewareProvider implements authPorts.MiddlewareProvider.
// It provides authentication middleware using auth and user services.
type MiddlewareProvider struct {
	authService authPorts.AuthService
	userService userPorts.UserService
}

// NewMiddlewareProvider creates a new MiddlewareProvider.
func NewMiddlewareProvider(authService authPorts.AuthService, userService userPorts.UserService) *MiddlewareProvider {
	return &MiddlewareProvider{
		authService: authService,
		userService: userService,
	}
}

// RequireAuth returns middleware that requires authentication.
// It validates the session token and adds user PUBLIC ID (UUID) to the context.
// If authentication fails, it returns 401 Unauthorized.
// SECURITY: Stores PublicID (UUID) in context, never internal INT ID.
func (p *MiddlewareProvider) RequireAuth() authPorts.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session token from cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate session
			session, err := p.authService.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// SECURITY: Fetch user to get PublicID (UUID), never expose internal INT ID
			user, err := p.userService.GetByID(r.Context(), session.UserID)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user PUBLIC ID (UUID) to context using ports key
			ctx := context.WithValue(r.Context(), authPorts.UserIDKey, user.PublicID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthFunc is a standalone function for backward compatibility.
// Prefer using MiddlewareProvider.RequireAuth() for new code.
// DEPRECATED: Use MiddlewareProvider instead.
func RequireAuth(authService authPorts.AuthService, userService userPorts.UserService) func(http.Handler) http.Handler {
	provider := NewMiddlewareProvider(authService, userService)
	return provider.RequireAuth()
}

// OptionalAuth returns middleware that optionally validates authentication.
// It validates the session token if present and adds user PUBLIC ID (UUID) to the context.
// If authentication fails or is not present, it continues without error.
// SECURITY: Stores PublicID (UUID) in context, never internal INT ID.
func (p *MiddlewareProvider) OptionalAuth() authPorts.Middleware {
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
			session, err := p.authService.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				// Invalid session, continue as guest
				next.ServeHTTP(w, r)
				return
			}

			// SECURITY: Fetch user to get PublicID (UUID), never expose internal INT ID
			user, err := p.userService.GetByID(r.Context(), session.UserID)
			if err != nil {
				// User not found, continue as guest
				next.ServeHTTP(w, r)
				return
			}

			// Add user PUBLIC ID (UUID) to context using ports key
			ctx := context.WithValue(r.Context(), authPorts.UserIDKey, user.PublicID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthFunc is a standalone function for backward compatibility.
// Prefer using MiddlewareProvider.OptionalAuth() for new code.
// DEPRECATED: Use MiddlewareProvider instead.
func OptionalAuth(authService authPorts.AuthService, userService userPorts.UserService) func(http.Handler) http.Handler {
	provider := NewMiddlewareProvider(authService, userService)
	return provider.OptionalAuth()
}

// GetUserID extracts user PUBLIC ID (UUID) from request context.
// Returns empty string if not authenticated.
// SECURITY: Returns UUID string, never internal INT ID.
// DEPRECATED: Use authPorts.GetUserID instead.
func GetUserID(ctx context.Context) string {
	return authPorts.GetUserID(ctx)
}

// GetUsername extracts username from request context.
// Returns empty string if not authenticated.
// DEPRECATED: Use authPorts.GetUsername instead.
func GetUsername(ctx context.Context) string {
	return authPorts.GetUsername(ctx)
}

// IsAuthenticated checks if the request is authenticated.
// DEPRECATED: Use authPorts.IsAuthenticated instead.
func IsAuthenticated(ctx context.Context) bool {
	return authPorts.IsAuthenticated(ctx)
}
