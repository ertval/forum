// INPUT ADAPTER - HTTP Handler Base
// [OPTIONAL FEATURE: forum-advanced-features]
// Package adapters implements HTTP handlers for notification endpoints.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/notification/ports"
	userPorts "forum/internal/modules/user/ports"
	platformTemplates "forum/internal/platform/templates"
	"net/http"
)

// HTTPHandler handles HTTP requests for notifications.
type HTTPHandler struct {
	notificationService ports.NotificationService
	userService         userPorts.UserService
	middlewareProvider  authPorts.AuthMiddleware
	templates           *platformTemplates.Registry
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Notification() ports.NotificationService
	User() userPorts.UserService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for notifications with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *platformTemplates.Registry) *HTTPHandler {
	return &HTTPHandler{
		notificationService: services.Notification(),
		userService:         services.User(),
		middlewareProvider:  services.AuthMiddleware(),
		templates:           templates,
	}
}

// RegisterRoutes registers all notification routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)
}
