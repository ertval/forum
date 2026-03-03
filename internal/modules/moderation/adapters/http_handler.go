// INPUT ADAPTER - HTTP Handler Base
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements HTTP handlers for moderation endpoints.
package adapters

import (
	"context"
	"fmt"
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/moderation/ports"
	userDomain "forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
	platformTemplates "forum/internal/platform/templates"
	"net/http"
)

// HTTPHandler handles HTTP requests for moderation.
type HTTPHandler struct {
	moderationService  ports.ModerationService
	userService        userLookupService
	middlewareProvider authPorts.AuthMiddleware
	templates          *platformTemplates.Registry
}

type userLookupService interface {
	GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error)
	GetByID(ctx context.Context, userID int) (*userDomain.User, error)
	UpdateRole(ctx context.Context, userID int, newRole userDomain.Role) error
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Moderation() ports.ModerationService
	User() userPorts.UserService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for moderation with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *platformTemplates.Registry) *HTTPHandler {
	return &HTTPHandler{
		moderationService:  services.Moderation(),
		userService:        services.User(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
	}
}

// getInternalUserID converts a public UUID from context to internal INT ID.
func (h *HTTPHandler) getInternalUserID(ctx context.Context, userPublicID string) (int, error) {
	user, err := h.userService.GetByPublicID(ctx, userPublicID)
	if err != nil || user == nil {
		return 0, fmt.Errorf("user not found")
	}
	return user.ID, nil
}

// LookupInternalID resolves a public UUID to an internal database ID.
func (h *HTTPHandler) LookupInternalID(ctx context.Context, publicID string) (int, error) {
	user, err := h.userService.GetByPublicID(ctx, publicID)
	if err != nil || user == nil {
		return 0, fmt.Errorf("user not found")
	}
	return user.ID, nil
}

// RegisterRoutes registers all moderation routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)
}
