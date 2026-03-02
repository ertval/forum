// Package adapters provides HTTP middleware for authentication.
package adapters

import (
	"context"
	"net/http"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	userPorts "forum/internal/modules/user/ports"
	platformErrors "forum/internal/platform/errors"
)

// AuthMiddleware implements authPorts.AuthMiddleware.
// It provides authentication middleware using auth and user services.
type AuthMiddleware struct {
	authService authPorts.AuthService
	userService userPorts.UserService
	cookieName  string
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(authService authPorts.AuthService, userService userPorts.UserService, cookieName string) *AuthMiddleware {
	if cookieName == "" {
		cookieName = "session_token"
	}

	return &AuthMiddleware{
		authService: authService,
		userService: userService,
		cookieName:  cookieName,
	}
}

// resolveAuth attempts to resolve the authenticated user from the session cookie.
// It returns the enriched context with the user's PublicID if authentication succeeds,
// or the original context and ok=false otherwise.
// SECURITY: Stores PublicID (UUID) in context, never internal INT ID.
func (p *AuthMiddleware) resolveAuth(r *http.Request) (context.Context, bool) {
	cookie, err := r.Cookie(p.cookieName)
	if err != nil {
		return r.Context(), false
	}

	session, err := p.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		return r.Context(), false
	}

	user, err := p.userService.GetByID(r.Context(), session.UserID)
	if err != nil {
		return r.Context(), false
	}

	ctx := context.WithValue(r.Context(), authPorts.UserIDKey, user.PublicID)
	return ctx, true
}

// RequireAuth returns middleware that requires authentication.
// It validates the session token and adds user PUBLIC ID (UUID) to the context.
// If authentication fails, it returns 401 Unauthorized.
func (p *AuthMiddleware) RequireAuth() authPorts.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, ok := p.resolveAuth(r)
			if !ok {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
					return
				}

				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns middleware that optionally validates authentication.
// It validates the session token if present and adds user PUBLIC ID (UUID) to the context.
// If authentication fails or is not present, it continues without error.
func (p *AuthMiddleware) OptionalAuth() authPorts.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, _ := p.resolveAuth(r)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
