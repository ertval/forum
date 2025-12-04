// INPUT ADAPTER - HTTP API Handler
// Package adapters implements HTTP API handlers for comment endpoints.
// This adapter handles JSON API requests for comment operations.
package adapters

import (
	"errors"
	"net/http"
	"time"

	commentDomain "forum/internal/modules/comment/domain"
	platformErrors "forum/internal/platform/errors"
)

// RegisterAPIRoutes registers all comment API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	// POST /api/comments/posts/{post_id} - Create comment (requires auth)
	router.HandleFunc("POST /api/comments/posts/{post_id}", h.CreateCommentAPI)
	// GET /api/comments/{id} - Get comment (public)
	router.HandleFunc("GET /api/comments/{id}", h.GetCommentAPI)
	// PUT /api/comments/{id} - Update comment (requires auth + ownership)
	router.HandleFunc("PUT /api/comments/{id}", h.UpdateCommentAPI)
	// DELETE /api/comments/{id} - Delete comment (requires auth + ownership)
	router.HandleFunc("DELETE /api/comments/{id}", h.DeleteCommentAPI)
	// GET /api/comments/posts/{post_id} - List comments for post (public)
	router.HandleFunc("GET /api/comments/posts/{post_id}", h.ListCommentsByPostAPI)
}

// CreateCommentAPI handles comment creation requests.
func (h *HTTPHandler) CreateCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	postPublicID := r.PathValue("post_id")
	if postPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	// Get userID from session
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := h.parseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate content
	if req.Content == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment content is required")
		return
	}

	// Call service to create comment
	comment, err := h.commentService.CreateComment(r.Context(), postPublicID, userID, req.Content)
	if err != nil {
		// Map domain errors to HTTP status codes
		if errors.Is(err, commentDomain.ErrEmptyContent) || errors.Is(err, commentDomain.ErrContentTooLong) {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	// Return success response
	resp := struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    int    `json:"user_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}{
		ID:        comment.PublicID,
		PostID:    comment.PublicPostID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// GetCommentAPI handles comment retrieval requests.
func (h *HTTPHandler) GetCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from URL path
	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	// Call service to get comment
	comment, err := h.commentService.GetComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Comment not found")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve comment")
		return
	}

	// Return success response
	resp := struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    int    `json:"user_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        comment.PublicID,
		PostID:    comment.PublicPostID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// UpdateCommentAPI handles comment update requests.
func (h *HTTPHandler) UpdateCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from URL path
	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	// Get userID from session
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := h.parseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate content
	if req.Content == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment content is required")
		return
	}

	// For authorization check, first get the existing comment to verify ownership
	existingComment, err := h.commentService.GetComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Comment not found")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve comment")
		return
	}

	// Check if the user is the owner of the comment
	if existingComment.UserID != userID {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "Not authorized to edit this comment")
		return
	}

	// Call service to update comment
	err = h.commentService.UpdateComment(r.Context(), commentPublicID, req.Content)
	if err != nil {
		// Map domain errors to HTTP status codes
		if errors.Is(err, commentDomain.ErrEmptyContent) || errors.Is(err, commentDomain.ErrContentTooLong) {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	// Return success response
	resp := struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    int    `json:"user_id"`
		Content   string `json:"content"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        existingComment.PublicID,
		PostID:    existingComment.PublicPostID,
		UserID:    existingComment.UserID,
		Content:   req.Content,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// DeleteCommentAPI handles comment deletion requests.
func (h *HTTPHandler) DeleteCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from URL path
	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	// Get userID from session
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// For authorization check, first get the existing comment to verify ownership
	existingComment, err := h.commentService.GetComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Comment not found")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve comment")
		return
	}

	// Check if the user is the owner of the comment
	if existingComment.UserID != userID {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "Not authorized to delete this comment")
		return
	}

	// Call service to delete comment
	err = h.commentService.DeleteComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Comment not found")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	// Return success response
	resp := struct {
		Message string `json:"message"`
	}{
		Message: "Comment deleted successfully",
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// ListCommentsByPostAPI handles listing comments for a post.
func (h *HTTPHandler) ListCommentsByPostAPI(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	postPublicID := r.PathValue("post_id")
	if postPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	// Call service to list comments
	comments, err := h.commentService.ListCommentsByPost(r.Context(), postPublicID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}

	// Prepare response
	var commentsResp []struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    int    `json:"user_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	for _, comment := range comments {
		commentResp := struct {
			ID        string `json:"id"`
			PostID    string `json:"post_id"`
			UserID    int    `json:"user_id"`
			Content   string `json:"content"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}{
			ID:        comment.PublicID,
			PostID:    comment.PublicPostID,
			UserID:    comment.UserID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
		}
		commentsResp = append(commentsResp, commentResp)
	}

	h.writeJSON(w, http.StatusOK, struct {
		Comments []struct {
			ID        string `json:"id"`
			PostID    string `json:"post_id"`
			UserID    int    `json:"user_id"`
			Content   string `json:"content"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		} `json:"comments"`
	}{
		Comments: commentsResp,
	})
}
