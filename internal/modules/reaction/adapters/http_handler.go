// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements HTTP handlers for reaction endpoints.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	reactionPorts "forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
	"forum/internal/platform/logger"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for reactions.
type HTTPHandler struct {
	reactionService    reactionPorts.ReactionService
	userService        userPorts.UserService
	middlewareProvider authPorts.AuthMiddleware
	templates          *template.Template
	logger             *logger.Logger
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Reaction() reactionPorts.ReactionService
	User() userPorts.UserService
	AuthMiddleware() authPorts.AuthMiddleware
	Logger() *logger.Logger  // Add this to the interface to access logger
}

// NewHTTPHandler creates a new HTTP handler for reactions with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		reactionService:    services.Reaction(),
		userService:        services.User(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
		logger:             services.Logger(),
	}
}

// RegisterRoutes registers all reaction routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)
}
