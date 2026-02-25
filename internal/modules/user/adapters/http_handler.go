// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements the HTTP handlers for user endpoints.
package adapters

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
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes (none yet)
	h.RegisterPageRoutes(router)
}
