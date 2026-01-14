// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements HTTP handlers for comment endpoints.
// This file contains the base handler structure and shared utilities.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	authPorts "forum/internal/modules/auth/ports"
	commentPorts "forum/internal/modules/comment/ports"
	postPorts "forum/internal/modules/post/ports"
	reactionPorts "forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
	logger "forum/internal/platform/logger"
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
	templates          *template.Template
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
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
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

// buildCurrentUser fetches full user info (including cached stats) and returns
// a map suitable for templates. It always returns a map (never nil).
func (h *HTTPHandler) buildCurrentUser(ctx context.Context, userID int) map[string]interface{} {
	// Fetch user with all fields including cached stats
	user, err := h.userService.GetByID(ctx, userID)
	if err != nil || user == nil {
		// Return empty map if user not found
		return map[string]interface{}{
			"PublicID":      "",
			"Username":      "",
			"Email":         "",
			"PostCount":     0,
			"CommentCount":  0,
			"ReactionCount": 0,
		}
	}

	// Get reaction count from reaction service
	reactionCount := 0
	if h.reactionService != nil {
		if count, err := h.reactionService.GetUserReactionCount(ctx, userID); err == nil {
			reactionCount = count
		}
	}

	return map[string]interface{}{
		"PublicID":      user.PublicID,
		"Username":      user.Username,
		"Email":         user.Email,
		"PostCount":     user.PostCount,
		"CommentCount":  user.CommentCount,
		"ReactionCount": reactionCount,
	}
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
	if userPublicID == "" {
		return 0, fmt.Errorf("user ID required")
	}

	// Fetch user by PublicID to get internal INT ID
	user, err := h.userService.GetByPublicID(ctx, userPublicID)
	if err != nil {
		return 0, fmt.Errorf("user not found")
	}

	return user.ID, nil
}

// RegisterRoutes registers all comment routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes
	h.RegisterPageRoutes(router)

	// Register form routes
	h.RegisterFormRoutes(router)
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
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("content type is not application/json")
	}

	// Decode the JSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // This makes parsing stricter

	return decoder.Decode(v)
}
