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
	authService authPorts.AuthService
	userService userPorts.UserService
	templates   *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
// This allows the handler to receive all services but only use what it needs.
type ServiceContainer interface {
	Auth() authPorts.AuthService
	User() userPorts.UserService
}

// NewHTTPHandler creates a new HTTP handler for authentication with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		authService: services.Auth(),
		userService: services.User(),
		templates:   templates,
	}
}

// GetCurrentUser extracts user info from session cookie (helper for other handlers).
// Returns userID and username, or (0, "") if not authenticated.
func (h *HTTPHandler) GetCurrentUser(r *http.Request) (userID int, username string) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return 0, ""
	}

	session, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil || session == nil {
		return 0, ""
	}

	// Fetch username from user service
	user, err := h.userService.GetByID(r.Context(), session.UserID)
	if err != nil || user == nil {
		return session.UserID, "" // Return ID even if username fetch fails
	}

	return session.UserID, user.Username
}

// RegisterRoutes registers all authentication routes with the router.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes
	h.RegisterPageRoutes(router)
}
