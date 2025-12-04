// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements HTTP handlers for reaction endpoints.
package adapters

import (
	"forum/internal/modules/reaction/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for reactions.
type HTTPHandler struct {
	reactionService ports.ReactionService
	templates       *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Reaction() ports.ReactionService
}

// NewHTTPHandler creates a new HTTP handler for reactions with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		reactionService: services.Reaction(),
		templates:       templates,
	}
}

// RegisterRoutes registers all reaction routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)
}
