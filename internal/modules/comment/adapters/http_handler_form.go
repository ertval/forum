// INPUT ADAPTER - HTTP Form Handler
// Package adapters implements HTTP form handlers for comment endpoints.
// This adapter handles HTML form submissions for comment operations.
package adapters

import (
	"errors"
	"net/http"
	"strings"

	commentDomain "forum/internal/modules/comment/domain"
	platformErrors "forum/internal/platform/errors"
)

// RegisterFormRoutes registers all comment form routes with the router.
func (h *HTTPHandler) RegisterFormRoutes(router *http.ServeMux) {
	// POST /posts/{post_id}/comments - Create comment via form (requires auth)
	router.HandleFunc("POST /posts/{post_id}/comments", h.CreateCommentForm)
	// DELETE /comments/{id} - Delete comment via form (requires auth + ownership)
	router.HandleFunc("DELETE /comments/{id}", h.DeleteCommentForm)
}

// CreateCommentForm handles comment form submissions from the post detail page (HTML form).
func (h *HTTPHandler) CreateCommentForm(w http.ResponseWriter, r *http.Request) {
	// Get userID from session
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract post ID from URL path using PathValue (Go 1.22+ pattern)
	postPublicID := r.PathValue("post_id")
	if postPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	// Parse form data (content)
	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment content is required")
		return
	}

	// Call service to create comment
	_, err := h.commentService.CreateComment(r.Context(), postPublicID, userID, content)
	if err != nil {
		// Map domain errors to HTTP status codes
		if errors.Is(err, commentDomain.ErrEmptyContent) || errors.Is(err, commentDomain.ErrContentTooLong) {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	// Redirect back to the post page (same as the original request)
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

// DeleteCommentForm handles comment deletion from the post detail page (HTML form).
func (h *HTTPHandler) DeleteCommentForm(w http.ResponseWriter, r *http.Request) {
	// Get userID from session
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract comment ID from URL path using PathValue (Go 1.22+ pattern)
	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment ID is required")
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

	// Return success response for AJAX request
	w.WriteHeader(http.StatusNoContent)
}
