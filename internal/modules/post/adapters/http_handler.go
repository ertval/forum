// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements HTTP handlers for post endpoints.
// This file contains the base handler structure and shared utilities.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	commentPorts "forum/internal/modules/comment/ports"
	postPorts "forum/internal/modules/post/ports"
	reactionPorts "forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
	"forum/internal/platform/httpserver"
	logger "forum/internal/platform/logger"
	platformTemplates "forum/internal/platform/templates"
)

// HTTPHandler handles HTTP requests for posts.
type HTTPHandler struct {
	postService        postPorts.PostService
	categoryService    postPorts.CategoryService
	authService        authPorts.AuthService
	userService        userPorts.UserService
	middlewareProvider authPorts.AuthMiddleware
	commentService     commentPorts.CommentService
	reactionService    reactionPorts.ReactionService
	templates          *platformTemplates.Registry
	logger             *logger.Logger
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Post() postPorts.PostService
	Category() postPorts.CategoryService
	Auth() authPorts.AuthService
	User() userPorts.UserService
	AuthMiddleware() authPorts.AuthMiddleware
	Comment() commentPorts.CommentService
	Reaction() reactionPorts.ReactionService
}

// NewHTTPHandler creates a new HTTP handler for posts with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *platformTemplates.Registry, log ...*logger.Logger) *HTTPHandler {
	var l *logger.Logger
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	} else {
		l = logger.New(logger.InfoLevel, os.Stderr)
	}
	return &HTTPHandler{
		postService:        services.Post(),
		categoryService:    services.Category(),
		authService:        services.Auth(),
		userService:        services.User(),
		middlewareProvider: services.AuthMiddleware(),
		commentService:     services.Comment(),
		reactionService:    services.Reaction(),
		templates:          templates,
		logger:             l,
	}
}

// Templates returns the injected template registry.
func (h *HTTPHandler) Templates() *platformTemplates.Registry {
	return h.templates
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

// GetUserWithStats extracts user info with stats from session cookie (for external handlers).
// Returns a map with full user data including stats, or nil if not authenticated.
func (h *HTTPHandler) GetUserWithStats(r *http.Request) map[string]interface{} {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return nil
	}

	session, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil || session == nil {
		return nil
	}

	return h.buildCurrentUser(r.Context(), session.UserID)
}

// getInternalUserID converts a PublicID (UUID) from context to internal INT ID.
// This is used by handlers to convert the UUID stored in context by middleware
// to the internal INT ID needed for service layer calls.
// SECURITY: Ensures public UUID is never exposed, only used for lookups.
func (h *HTTPHandler) getInternalUserID(ctx context.Context, userPublicID string) (int, error) {
	return httpserver.GetInternalUserID(ctx, userPublicID, h)
}

// buildPageTitle creates a dynamic page title based on active filters.
func (h *HTTPHandler) buildPageTitle(filterParams listFilterOptions) string {
	// Build title parts: [My] [Category] Posts [TimePeriod]
	var parts []string
	activityType, reactionType := resolveBoardActivityFilters(filterParams)

	// Add activity-specific prefix
	if activityType == "my_posts" || filterParams.MyPosts || filterParams.UserID != "" {
		parts = append(parts, "My")
	} else if activityType == "reactions" {
		switch reactionType {
		case "dislike":
			parts = append(parts, "My Disliked")
		case "like":
			parts = append(parts, "My Liked")
		default:
			parts = append(parts, "My Reacted")
		}
	} else if filterParams.LikedPosts {
		parts = append(parts, "My Liked")
	} else if filterParams.DislikedPosts {
		parts = append(parts, "My Disliked")
	} else if activityType == "commented_posts" || filterParams.CommentedPosts || filterParams.Commenter != "" {
		parts = append(parts, "Commented")
	}

	// Add category if selected
	if filterParams.Category != "" {
		parts = append(parts, filterParams.Category)
	}

	// Always include "Posts"
	parts = append(parts, "Posts")

	// Add time period if selected (and not "all")
	switch filterParams.DateFilter {
	case "today":
		parts = append(parts, "Today")
	case "week":
		parts = append(parts, "This Week")
	case "month":
		parts = append(parts, "This Month")
	}

	// Default title if no filters
	if len(parts) == 1 && parts[0] == "Posts" {
		return "All Posts"
	}

	return strings.Join(parts, " ")
}

func normalizeBoardActivityType(activityType string) string {
	normalized := strings.ToLower(strings.TrimSpace(activityType))
	switch normalized {
	case "my_posts", "reactions", "commented_posts":
		return normalized
	default:
		return "all"
	}
}

func normalizeReactionType(reactionType string) string {
	normalized := strings.ToLower(strings.TrimSpace(reactionType))
	switch normalized {
	case "like", "dislike":
		return normalized
	default:
		return "all"
	}
}

func resolveBoardActivityFilters(params listFilterOptions) (string, string) {
	activityType := normalizeBoardActivityType(params.ActivityType)
	reactionType := normalizeReactionType(params.ReactionType)

	if activityType == "all" {
		switch {
		case params.MyPosts || params.UserID != "":
			activityType = "my_posts"
		case params.CommentedPosts || params.Commenter != "":
			activityType = "commented_posts"
		case params.LikedPosts || params.DislikedPosts:
			activityType = "reactions"
		}
	}

	if activityType == "reactions" && reactionType == "all" {
		switch {
		case params.LikedPosts && !params.DislikedPosts:
			reactionType = "like"
		case params.DislikedPosts && !params.LikedPosts:
			reactionType = "dislike"
		}
	}

	return activityType, reactionType
}

// RegisterRoutes registers all post routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes
	h.RegisterPageRoutes(router)
}

// writeJSON writes a JSON response.
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Error encoding JSON response", logger.Error(err))
	}
}

// createPostPreview creates a preview of the post content with a fixed length.
func createPostPreview(content string) string {
	const previewLength = 100 // Characters to show in preview
	if len(content) <= previewLength {
		return content
	}

	// Ensure we don't cut in the middle of a word if possible
	preview := content[:previewLength]
	if len(content) > previewLength {
		// Find the last space to avoid cutting in the middle of a word
		lastSpaceIndex := previewLength
		for i := previewLength - 1; i >= 0; i-- {
			if content[i] == ' ' {
				lastSpaceIndex = i
				break
			}
		}

		// If we found a space and it's not too close to the beginning, use it
		if lastSpaceIndex > previewLength/2 {
			preview = content[:lastSpaceIndex]
		}
	}

	return preview + "..."
}
