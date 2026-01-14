// INPUT ADAPTER - HTTP API Handler
// Package adapters implements HTTP API handlers for reaction endpoints.
package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/reaction/domain"
	"forum/internal/platform/logger"
)

// RegisterAPIRoutes registers all reaction API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	authMiddleware := h.middlewareProvider.RequireAuth()

	// Protected API routes (require authentication)
	router.Handle("POST /api/reactions", authMiddleware(http.HandlerFunc(h.AddReactionAPI)))
	router.Handle("DELETE /api/reactions", authMiddleware(http.HandlerFunc(h.RemoveReactionAPI)))

	// Public API routes (no authentication required)
	// GET /api/reactions/{targetType}/{targetId} - Get reactions for target (public)
	router.HandleFunc("GET /api/reactions/{targetType}/{targetId}", h.GetReactionsAPI)
	// GET /api/reactions/{targetType}/{targetId}/count - Count reactions (public)
	router.HandleFunc("GET /api/reactions/{targetType}/{targetId}/count", h.CountReactionsAPI)
}

// AddReactionAPI handles adding a reaction to a post or comment.
func (h *HTTPHandler) AddReactionAPI(w http.ResponseWriter, r *http.Request) {
	// Verify authentication first
	if !authPorts.IsAuthenticated(r.Context()) {
		h.logger.Error("Unauthorized reaction attempt", logger.String("method", "POST"), logger.String("path", "/api/reactions"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract user ID from context
	userPublicID := authPorts.GetUserID(r.Context())

	// Get user's internal ID
	user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil {
		h.logger.Error("User not found for reaction", logger.String("user_id", userPublicID), logger.Error(err))
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	userID := user.ID

	// Parse the request body
	var req struct {
		TargetType string              `json:"target_type"`
		TargetID   string              `json:"target_id"`
		Type       domain.ReactionType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for reaction", logger.String("user_id", userPublicID), logger.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the target type
	if req.TargetType != "post" && req.TargetType != "comment" {
		h.logger.Error("Invalid target type for reaction", logger.String("target_type", req.TargetType), logger.String("user_id", userPublicID))
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	// Validate the reaction type
	if req.Type != domain.ReactionLike && req.Type != domain.ReactionDislike {
		h.logger.Error("Invalid reaction type", logger.String("reaction_type", string(req.Type)), logger.String("user_id", userPublicID))
		http.Error(w, "Reaction type must be 'like' or 'dislike'", http.StatusBadRequest)
		return
	}

	h.logger.Info("Processing reaction",
		logger.String("user_id", userPublicID),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
		logger.String("reaction_type", string(req.Type)))

	// Call the service
	err = h.reactionService.React(r.Context(), userID, req.TargetID, req.TargetType, req.Type)
	if err != nil {
		h.logger.Error("Failed to add reaction",
			logger.String("user_id", userPublicID),
			logger.String("target_type", req.TargetType),
			logger.String("target_id", req.TargetID),
			logger.String("reaction_type", string(req.Type)),
			logger.Error(err))

		if err == domain.ErrInvalidTarget || err == domain.ErrInvalidReactionType {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Reaction added successfully",
		logger.String("user_id", userPublicID),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
		logger.String("reaction_type", string(req.Type)))

	// Return success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Reaction added successfully"}`)
}

// RemoveReactionAPI handles removing a reaction from a post or comment.
func (h *HTTPHandler) RemoveReactionAPI(w http.ResponseWriter, r *http.Request) {
	// Verify authentication first
	if !authPorts.IsAuthenticated(r.Context()) {
		h.logger.Error("Unauthorized reaction removal attempt", logger.String("method", "DELETE"), logger.String("path", "/api/reactions"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract user ID from context
	userPublicID := authPorts.GetUserID(r.Context())

	// Get user's internal ID
	user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil {
		h.logger.Error("User not found for reaction removal", logger.String("user_id", userPublicID), logger.Error(err))
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	userID := user.ID

	// Parse the request body
	var req struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for reaction removal", logger.String("user_id", userPublicID), logger.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the target type
	if req.TargetType != "post" && req.TargetType != "comment" {
		h.logger.Error("Invalid target type for reaction removal", logger.String("target_type", req.TargetType), logger.String("user_id", userPublicID))
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	h.logger.Info("Processing reaction removal",
		logger.String("user_id", userPublicID),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID))

	// Call the service
	err = h.reactionService.RemoveReaction(r.Context(), userID, req.TargetID, req.TargetType)
	if err != nil {
		h.logger.Error("Failed to remove reaction",
			logger.String("user_id", userPublicID),
			logger.String("target_type", req.TargetType),
			logger.String("target_id", req.TargetID),
			logger.Error(err))

		if err == domain.ErrReactionNotFound || err == domain.ErrInvalidTarget {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Reaction removed successfully",
		logger.String("user_id", userPublicID),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID))

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetReactionsAPI handles retrieving reactions for a post or comment.
func (h *HTTPHandler) GetReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Extract target type and target ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.logger.Error("Invalid path for getting reactions", logger.String("path", r.URL.Path))
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetType := pathParts[len(pathParts)-2]
	targetID := pathParts[len(pathParts)-1]

	// Validate the target type
	if targetType != "post" && targetType != "comment" {
		h.logger.Error("Invalid target type for getting reactions", logger.String("target_type", targetType))
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting reactions for target",
		logger.String("target_type", targetType),
		logger.String("target_id", targetID))

	// Call the service
	reactions, err := h.reactionService.GetReactions(r.Context(), targetID, targetType)
	if err != nil {
		h.logger.Error("Failed to get reactions",
			logger.String("target_type", targetType),
			logger.String("target_id", targetID),
			logger.Error(err))

		if err == domain.ErrInvalidTarget {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully retrieved reactions",
		logger.String("target_type", targetType),
		logger.String("target_id", targetID),
		logger.Int("reaction_count", len(reactions)))

	// Return the reactions
	h.writeJSON(w, http.StatusOK, reactions)
}

// CountReactionsAPI handles counting reactions for a target.
func (h *HTTPHandler) CountReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Extract target type and target ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		h.logger.Error("Invalid path for counting reactions", logger.String("path", r.URL.Path))
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetType := pathParts[len(pathParts)-2]
	targetID := pathParts[len(pathParts)-1]

	// Validate the target type
	if targetType != "post" && targetType != "comment" {
		h.logger.Error("Invalid target type for counting reactions", logger.String("target_type", targetType))
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	h.logger.Info("Counting reactions for target",
		logger.String("target_type", targetType),
		logger.String("target_id", targetID))

	// Call the service
	likes, dislikes, err := h.reactionService.CountReactions(r.Context(), targetID, targetType)
	if err != nil {
		h.logger.Error("Failed to count reactions",
			logger.String("target_type", targetType),
			logger.String("target_id", targetID),
			logger.Error(err))

		if err == domain.ErrInvalidTarget {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully counted reactions",
		logger.String("target_type", targetType),
		logger.String("target_id", targetID),
		logger.Int("likes", likes),
		logger.Int("dislikes", dislikes))

	// Return the counts
	response := struct {
		TargetID   string `json:"target_id"`
		TargetType string `json:"target_type"`
		Likes      int    `json:"likes"`
		Dislikes   int    `json:"dislikes"`
	}{
		TargetID:   targetID,
		TargetType: targetType,
		Likes:      likes,
		Dislikes:   dislikes,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response.
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
