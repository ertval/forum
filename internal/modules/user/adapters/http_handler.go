// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements the HTTP handlers for user endpoints.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/user/ports"
	"forum/internal/platform/upload"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for user operations.
type HTTPHandler struct {
	userService        ports.UserService
	middlewareProvider authPorts.AuthMiddleware
	templates          *template.Template
	avatarImageHandler *upload.ImageHandler
	maxAvatarSize      int64
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	User() ports.UserService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for users with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template, uploadDir ...string) *HTTPHandler {
	dir := "./static/uploads"
	if len(uploadDir) > 0 && uploadDir[0] != "" {
		dir = uploadDir[0]
	}

	return &HTTPHandler{
		userService:        services.User(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
		avatarImageHandler: upload.NewImageHandler(dir, maxAvatarImageSize),
		maxAvatarSize:      maxAvatarImageSize,
	}
}

// RegisterRoutes registers all user routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes (none yet)
	h.RegisterPageRoutes(router)
}
