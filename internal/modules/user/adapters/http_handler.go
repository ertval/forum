package adapters

// INPUT ADAPTER - HTTP Handler
// Package adapters implements the HTTP handlers for user endpoints.

import (
	"forum/internal/modules/user/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for user operations.
type HTTPHandler struct {
	userService ports.UserService
	templates   *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	User() ports.UserService
}

// NewHTTPHandler creates a new HTTP handler for users with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		userService: services.User(),
		templates:   templates,
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

// GetUserAPI handles user retrieval requests.
// TODO: Implement user retrieval handler.
func (h *HTTPHandler) GetUserAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ListUsersAPI handles listing users.
// TODO: Implement user listing handler.
func (h *HTTPHandler) ListUsersAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// UpdateRoleAPI handles updating a user's role.
// TODO: Implement role update handler.
func (h *HTTPHandler) UpdateRoleAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// DeactivateUserAPI handles deactivating a user account.
// TODO: Implement user deactivation handler.
func (h *HTTPHandler) DeactivateUserAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}
