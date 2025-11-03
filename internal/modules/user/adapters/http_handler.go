package adapters
// INPUT ADAPTER - HTTP Handler
// Package adapters implements the HTTP handlers for user endpoints.
package adapters

import (
	"forum/internal/modules/user/ports"
	"net/http"
)

// HTTPHandler handles HTTP requests for user operations.
type HTTPHandler struct {
	userService ports.UserService
}

// NewHTTPHandler creates a new HTTP handler for users.
func NewHTTPHandler(userService ports.UserService) *HTTPHandler {
	return &HTTPHandler{
		userService: userService,
	}
}

// RegisterRoutes registers all user routes.
// TODO: Implement route registration.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// GET /users/{id} - Get user profile
	// GET /users - List users
	// PUT /users/{id}/role - Update user role (admin only)
	// PUT /users/{id}/deactivate - Deactivate user
}

// GetUser retrieves a user's profile.
// TODO: Implement user profile handler.
func (h *HTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ListUsers lists all users (paginated).
// TODO: Implement user listing handler.
func (h *HTTPHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// UpdateRole updates a user's role.
// TODO: Implement role update handler.
func (h *HTTPHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// DeactivateUser deactivates a user account.
// TODO: Implement deactivation handler.
func (h *HTTPHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}
