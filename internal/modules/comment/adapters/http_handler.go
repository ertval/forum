// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements HTTP handlers for comment endpoints.
// This file contains the base handler structure and shared utilities.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"

	authPorts "forum/internal/modules/auth/ports"
	commentPorts "forum/internal/modules/comment/ports"
	postPorts "forum/internal/modules/post/ports"
	reactionPorts "forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
	"forum/internal/platform/httpserver"
	logger "forum/internal/platform/logger"
	platformTemplates "forum/internal/platform/templates"
)

const (
	// DefaultPaginationLimit is the default number of items to show per page.
	DefaultPaginationLimit = 20
	// MaxPaginationLimit is the maximum number of items allowed to be requested per page.
	MaxPaginationLimit = 100
)

// HTTPHandler handles HTTP requests for comments.
type HTTPHandler struct {
	commentService     commentPorts.CommentService
	authService        authPorts.AuthService
	userService        userPorts.UserService
	postService        postPorts.PostService
	categoryService    postPorts.CategoryService
	reactionService    reactionPorts.ReactionService
	middlewareProvider authPorts.AuthMiddleware
	templates          *platformTemplates.Registry
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Comment() commentPorts.CommentService
	Auth() authPorts.AuthService
	User() userPorts.UserService
	Post() postPorts.PostService
	Category() postPorts.CategoryService
	Reaction() reactionPorts.ReactionService
	AuthMiddleware() authPorts.AuthMiddleware
}

// NewHTTPHandler creates a new HTTP handler for comments with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *platformTemplates.Registry) *HTTPHandler {
	return &HTTPHandler{
		commentService:     services.Comment(),
		authService:        services.Auth(),
		userService:        services.User(),
		postService:        services.Post(),
		categoryService:    services.Category(),
		reactionService:    services.Reaction(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
	}
}

// LookupUser returns user data by internal ID for template rendering.
func (h *HTTPHandler) LookupUser(ctx context.Context, userID int) (*httpserver.CurrentUserData, error) {
	user, err := h.userService.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return &httpserver.CurrentUserData{
		PublicID:     user.PublicID,
		Username:     user.Username,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		PostCount:    user.PostCount,
		CommentCount: user.CommentCount,
	}, nil
}

// LookupInternalID resolves a public UUID to an internal database ID.
func (h *HTTPHandler) LookupInternalID(ctx context.Context, publicID string) (int, error) {
	user, err := h.userService.GetByPublicID(ctx, publicID)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// LookupReactionCount returns the total reaction count for a user.
func (h *HTTPHandler) LookupReactionCount(ctx context.Context, userID int) (int, error) {
	if h.reactionService == nil {
		return 0, nil
	}
	return h.reactionService.GetUserReactionCount(ctx, userID)
}

// buildCurrentUser fetches full user info (including cached stats) and returns
// a map suitable for templates. It always returns a map (never nil).
func (h *HTTPHandler) buildCurrentUser(ctx context.Context, userID int) map[string]interface{} {
	return httpserver.BuildCurrentUser(ctx, userID, h)
}

// GetCurrentUser extracts user info from session cookie (helper for other handlers).
// Returns userID and username, or (0, "") if not authenticated.
func (h *HTTPHandler) GetCurrentUser(r *http.Request) (userID int, username string) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return 0, ""
	}

	session, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil || session == nil {
		return 0, ""
	}

	return session.UserID, "" // Return ID even if username fetch fails
}

// getInternalUserID converts a PublicID (UUID) from context to internal INT ID.
// This is used by handlers to convert the UUID stored in context by middleware
// to the internal INT ID needed for service layer calls.
// SECURITY: Ensures public UUID is never exposed, only used for lookups.
func (h *HTTPHandler) getInternalUserID(ctx context.Context, userPublicID string) (int, error) {
	return httpserver.GetInternalUserID(ctx, userPublicID, h)
}

// RegisterRoutes registers all comment routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes (includes form submission routes)
	h.RegisterPageRoutes(router)
}

// writeJSON writes a JSON response.
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error, but don't send it to the client
		cfg := &logger.Config{
			TimePrecision: logger.TimePrecisionSeconds,
		}
		lgr := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
		lgr.Error("Failed to encode JSON response",
			logger.Error(err),
			logger.String("method", "writeJSON"))
	}
}

// parseJSON parses JSON request body.
func (h *HTTPHandler) parseJSON(r *http.Request, v interface{}) error {
	// Check if content type is JSON
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return fmt.Errorf("content type is not application/json")
	}

	// Decode the JSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // This makes parsing stricter

	return decoder.Decode(v)
}
