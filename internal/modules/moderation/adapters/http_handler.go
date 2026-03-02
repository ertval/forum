// INPUT ADAPTER - HTTP Handler Base
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements HTTP handlers for moderation endpoints.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/moderation/ports"
	userDomain "forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
	"html/template"
	"mime"
	"net/http"
)

// HTTPHandler handles HTTP requests for moderation.
type HTTPHandler struct {
	moderationService  ports.ModerationService
	userService        userLookupService
	middlewareProvider authPorts.AuthMiddleware
	templates          *template.Template
}

type userLookupService interface {
	GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error)
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Moderation() ports.ModerationService
	User() userPorts.UserService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for moderation with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
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

func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *HTTPHandler) parseJSON(r *http.Request, v any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return fmt.Errorf("content type is not application/json")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// RegisterRoutes registers all moderation routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)
}
