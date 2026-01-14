// INPUT ADAPTER - HTTP Handler Base
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements HTTP handlers for moderation endpoints.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/moderation/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for moderation.
type HTTPHandler struct {
	moderationService  ports.ModerationService
	middlewareProvider authPorts.AuthMiddleware
	templates          *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Moderation() ports.ModerationService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for moderation with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		moderationService:  services.Moderation(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
	}
}

// RegisterRoutes registers all moderation routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)
}
