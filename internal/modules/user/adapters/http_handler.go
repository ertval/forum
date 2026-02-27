// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements the HTTP handlers for user endpoints.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/user/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for user operations.
type HTTPHandler struct {
	userService        ports.UserService
	middlewareProvider authPorts.AuthMiddleware
	templates          *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	User() ports.UserService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for users with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		userService:        services.User(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
	}
}

// RegisterRoutes registers all user routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes (none yet)
	h.RegisterPageRoutes(router)
}
