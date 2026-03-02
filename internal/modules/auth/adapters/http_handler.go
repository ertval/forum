// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements the HTTP handlers for authentication endpoints.
// This file contains the base handler structure and shared utilities.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	userPorts "forum/internal/modules/user/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for authentication.
// It receives HTTP requests, validates input, calls the service, and returns responses.
type HTTPHandler struct {
	authService        authPorts.AuthService
	userService        userPorts.UserService
	middlewareProvider authPorts.AuthMiddleware
	templates          *template.Template
	secureCookies      bool // Whether to set Secure flag on cookies (true in production)
	cookieName         string
}

// ServiceContainer defines the minimal interface needed by this handler.
// This allows the handler to receive all services but only use what it needs.
type ServiceContainer interface {
	Auth() authPorts.AuthService
	User() userPorts.UserService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for authentication with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template, secureCookies bool, cookieName string) *HTTPHandler {
	if cookieName == "" {
		cookieName = "session_token"
	}

	return &HTTPHandler{
		authService:        services.Auth(),
		userService:        services.User(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
		secureCookies:      secureCookies,
		cookieName:         cookieName,
	}
}

// GetCurrentUser extracts user info from session cookie (helper for other handlers).
// Returns publicID (UUID) and username, or ("", "") if not authenticated.
// SECURITY: Returns PublicID (UUID), never the internal int ID.
func (h *HTTPHandler) GetCurrentUser(r *http.Request) (publicID string, username string) {
	cookie, err := r.Cookie(h.cookieName)
	if err != nil || cookie.Value == "" {
		return "", ""
	}

	session, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil || session == nil {
		return "", ""
	}

	// Fetch user to get PublicID and username
	user, err := h.userService.GetByID(r.Context(), session.UserID)
	if err != nil || user == nil {
		return "", "" // Cannot determine public ID without user record
	}

	return user.PublicID, user.Username
}

// RegisterRoutes registers all authentication routes with the router.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes
	h.RegisterPageRoutes(router)
}
