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
)

// RegisterAPIRoutes registers all reaction API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	// POST /api/reactions - Add or update reaction (requires auth)
	router.HandleFunc("POST /api/reactions", func(w http.ResponseWriter, r *http.Request) {
		h.middlewareProvider.RequireAuth()(http.HandlerFunc(h.AddReactionAPI)).ServeHTTP(w, r)
	})
	// DELETE /api/reactions - Remove reaction (requires auth)
	router.HandleFunc("DELETE /api/reactions", func(w http.ResponseWriter, r *http.Request) {
		h.middlewareProvider.RequireAuth()(http.HandlerFunc(h.RemoveReactionAPI)).ServeHTTP(w, r)
	})
	// GET /api/reactions/{targetType}/{targetId} - Get reactions for target (public)
	router.HandleFunc("GET /api/reactions/{targetType}/{targetId}", h.GetReactionsAPI)
	// GET /api/reactions/{targetType}/{targetId}/count - Count reactions (public)
	router.HandleFunc("GET /api/reactions/{targetType}/{targetId}/count", h.CountReactionsAPI)
}

// AddReactionAPI handles adding a reaction to a post or comment.
func (h *HTTPHandler) AddReactionAPI(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's internal ID
	user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	userID := user.ID

	// Parse the request body
	var req struct {
		TargetType string             `json:"target_type"`
		TargetID   string             `json:"target_id"`
		Type       domain.ReactionType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the target type
	if req.TargetType != "post" && req.TargetType != "comment" {
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	// Validate the reaction type
	if req.Type != domain.ReactionLike && req.Type != domain.ReactionDislike {
		http.Error(w, "Reaction type must be 'like' or 'dislike'", http.StatusBadRequest)
		return
	}

	// Call the service
	err = h.reactionService.React(r.Context(), userID, req.TargetID, req.TargetType, req.Type)
	if err != nil {
		if err == domain.ErrInvalidTarget ||
		   err == domain.ErrInvalidReactionType {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Reaction added successfully"}`)
}

// RemoveReactionAPI handles removing a reaction from a post or comment.
func (h *HTTPHandler) RemoveReactionAPI(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's internal ID
	user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil {
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the target type
	if req.TargetType != "post" && req.TargetType != "comment" {
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	// Call the service
	err = h.reactionService.RemoveReaction(r.Context(), userID, req.TargetID, req.TargetType)
	if err != nil {
		if err == domain.ErrReactionNotFound || err == domain.ErrInvalidTarget {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetReactionsAPI handles retrieving reactions for a post or comment.
func (h *HTTPHandler) GetReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Extract target type and target ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetType := pathParts[len(pathParts)-2]
	targetID := pathParts[len(pathParts)-1]

	// Validate the target type
	if targetType != "post" && targetType != "comment" {
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	// Call the service
	reactions, err := h.reactionService.GetReactions(r.Context(), targetID, targetType)
	if err != nil {
		if err == domain.ErrInvalidTarget {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return the reactions
	h.writeJSON(w, http.StatusOK, reactions)
}

// CountReactionsAPI handles counting reactions for a target.
func (h *HTTPHandler) CountReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Extract target type and target ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetType := pathParts[len(pathParts)-2]
	targetID := pathParts[len(pathParts)-1]

	// Validate the target type
	if targetType != "post" && targetType != "comment" {
		http.Error(w, "Target type must be 'post' or 'comment'", http.StatusBadRequest)
		return
	}

	// Call the service
	likes, dislikes, err := h.reactionService.CountReactions(r.Context(), targetID, targetType)
	if err != nil {
		if err == domain.ErrInvalidTarget {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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
