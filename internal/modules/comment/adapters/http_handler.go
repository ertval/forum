// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for comment endpoints.
package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	authPorts "forum/internal/modules/auth/ports"
	commentDomain "forum/internal/modules/comment/domain"
	commentPorts "forum/internal/modules/comment/ports"
	platformErrors "forum/internal/platform/errors"
	logger "forum/internal/platform/logger"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTTPHandler handles HTTP requests for comments.
type HTTPHandler struct {
	commentService commentPorts.CommentService
	authService    authPorts.AuthService
	templates      *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Comment() commentPorts.CommentService
	Auth() authPorts.AuthService
}

// NewHTTPHandler creates a new HTTP handler for comments with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		commentService: services.Comment(),
		authService:    services.Auth(),
		templates:      templates,
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

// RegisterRoutes registers all comment routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// API routes (JSON responses)
	// POST /api/posts/{post_id}/comments - Create comment (requires auth)
	router.HandleFunc("POST /api/posts/{post_id}/comments", h.CreateCommentAPI)
	// GET /api/comments/{id} - Get comment (public)
	router.HandleFunc("GET /api/comments/{id}", h.GetCommentAPI)
	// PUT /api/comments/{id} - Update comment (requires auth + ownership)
	router.HandleFunc("PUT /api/comments/{id}", h.UpdateCommentAPI)
	// DELETE /api/comments/{id} - Delete comment (requires auth + ownership)
	router.HandleFunc("DELETE /api/comments/{id}", h.DeleteCommentAPI)
	// GET /api/posts/{post_id}/comments - List comments for post (public)
	router.HandleFunc("GET /api/posts/{post_id}/comments", h.ListCommentsByPostAPI)

	// Form routes (HTML responses)
	// POST /posts/{post_id}/comments - Create comment via form (requires auth)
	router.HandleFunc("POST /posts/{post_id}/comments", h.CreateCommentForm)
	// DELETE /comments/{id} - Delete comment via form (requires auth + ownership)
	router.HandleFunc("DELETE /comments/{id}", h.DeleteCommentForm)
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

// CreateCommentForm handles comment form submissions from the post detail page (HTML form).
func (h *HTTPHandler) CreateCommentForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get userID from session
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract post ID from URL path using PathValue (Go 1.22+ pattern)
	postPublicID := r.PathValue("post_id")
	if postPublicID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Parse form data (content)
	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	// Call service to create comment
	_, err := h.commentService.CreateComment(r.Context(), postPublicID, userID, content)
	if err != nil {
		// Map domain errors to HTTP status codes
		if errors.Is(err, commentDomain.ErrEmptyContent) || errors.Is(err, commentDomain.ErrContentTooLong) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page (same as the original request)
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

// DeleteCommentForm handles comment deletion from the post detail page (HTML form).
func (h *HTTPHandler) DeleteCommentForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get userID from session
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract comment ID from URL path using PathValue (Go 1.22+ pattern)
	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	// For authorization check, first get the existing comment to verify ownership
	existingComment, err := h.commentService.GetComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve comment", http.StatusInternalServerError)
		return
	}

	// Check if the user is the owner of the comment
	if existingComment.UserID != userID {
		http.Error(w, "Not authorized to delete this comment", http.StatusForbidden)
		return
	}

	// Call service to delete comment
	err = h.commentService.DeleteComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	// Return success response for AJAX request
	w.WriteHeader(http.StatusNoContent)
}
