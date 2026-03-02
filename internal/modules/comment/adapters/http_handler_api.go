// INPUT ADAPTER - HTTP API Handler
// Package adapters implements HTTP API handlers for comment endpoints.
// This adapter handles JSON API requests for comment operations.
package adapters

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	authPorts "forum/internal/modules/auth/ports"
	commentDomain "forum/internal/modules/comment/domain"
	platformErrors "forum/internal/platform/errors"
	"forum/internal/platform/httpjson"
)

// RegisterAPIRoutes registers all comment API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	authMiddleware := h.middlewareProvider.RequireAuth()
	optionalAuth := h.middlewareProvider.OptionalAuth()

	// Protected API routes (require authentication)
	router.Handle("POST /api/comments/posts/{post_id}", authMiddleware(http.HandlerFunc(h.CreateCommentAPI)))
	router.Handle("PUT /api/comments/{id}", authMiddleware(http.HandlerFunc(h.UpdateCommentAPI)))
	router.Handle("DELETE /api/comments/{id}", authMiddleware(http.HandlerFunc(h.DeleteCommentAPI)))
	router.Handle("GET /api/activity", authMiddleware(http.HandlerFunc(h.GetActivityAPI)))
	router.Handle("GET /api/comments/load-more", optionalAuth(http.HandlerFunc(h.LoadMoreCommentsAPI)))

	// Public API routes (no authentication required)
	// GET /api/comments/{id} - Get comment (public)
	router.HandleFunc("GET /api/comments/{id}", h.GetCommentAPI)
	// GET /api/comments/posts/{post_id} - List comments for post (public)
	router.HandleFunc("GET /api/comments/posts/{post_id}", h.ListCommentsByPostAPI)
}

// GetActivityAPI returns unified activity data for the authenticated user.
func (h *HTTPHandler) GetActivityAPI(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	filters := parseActivityFilters(r)

	activity, err := h.aggregateUserActivity(r.Context(), userPublicID, filters)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve activity")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, activity)
}

// CreateCommentAPI handles comment creation requests.
func (h *HTTPHandler) CreateCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	postPublicID := r.PathValue("post_id")
	if postPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(r.Context(), userPublicID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := httpjson.ParseJSON(r, &req); err != nil {
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

	comment.PublicUserID = userPublicID

	// Return success response
	resp := struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    string `json:"user_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}{
		ID:        comment.PublicID,
		PostID:    comment.PublicPostID,
		UserID:    comment.PublicUserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}

	httpjson.WriteJSON(w, http.StatusCreated, resp)
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

	// Author public ID is already populated by the repository JOIN query

	// Return success response
	resp := struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    string `json:"user_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        comment.PublicID,
		PostID:    comment.PublicPostID,
		UserID:    comment.PublicUserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
	}

	httpjson.WriteJSON(w, http.StatusOK, resp)
}

// UpdateCommentAPI handles comment update requests.
func (h *HTTPHandler) UpdateCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from URL path
	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	// Get user PUBLIC ID (UUID) from context
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(r.Context(), userPublicID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}
	currentUser, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || currentUser == nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := httpjson.ParseJSON(r, &req); err != nil {
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

	existingComment.PublicUserID = userPublicID

	// Return success response
	resp := struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    string `json:"user_id"`
		Content   string `json:"content"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        existingComment.PublicID,
		PostID:    existingComment.PublicPostID,
		UserID:    existingComment.PublicUserID,
		Content:   req.Content,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	httpjson.WriteJSON(w, http.StatusOK, resp)
}

// DeleteCommentAPI handles comment deletion requests.
func (h *HTTPHandler) DeleteCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from URL path
	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	// Get user PUBLIC ID (UUID) from context
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(r.Context(), userPublicID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	currentUser, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || currentUser == nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
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
	canDeleteAny := currentUser.Role == "moderator" || currentUser.Role == "admin"
	if existingComment.UserID != userID && !canDeleteAny {
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

	// Return 204 No Content on successful deletion
	w.WriteHeader(http.StatusNoContent)
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

	// Author public IDs are already populated by the repository JOIN query

	// Prepare response - initialize as empty slice to avoid null in JSON
	commentsResp := make([]struct {
		ID        string `json:"id"`
		PostID    string `json:"post_id"`
		UserID    string `json:"user_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}, 0)

	for _, comment := range comments {
		commentResp := struct {
			ID        string `json:"id"`
			PostID    string `json:"post_id"`
			UserID    string `json:"user_id"`
			Content   string `json:"content"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}{
			ID:        comment.PublicID,
			PostID:    comment.PublicPostID,
			UserID:    comment.PublicUserID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
		}
		commentsResp = append(commentsResp, commentResp)
	}

	httpjson.WriteJSON(w, http.StatusOK, struct {
		Comments []struct {
			ID        string `json:"id"`
			PostID    string `json:"post_id"`
			UserID    string `json:"user_id"`
			Content   string `json:"content"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		} `json:"comments"`
	}{
		Comments: commentsResp,
	})
}

// LoadMoreCommentsAPI handles loading additional comments for the My Comments page.
func (h *HTTPHandler) LoadMoreCommentsAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from session cookie.
	var userPublicID string
	cookie, err := r.Cookie(h.cookieName)
	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			user, err := h.userService.GetByID(ctx, session.UserID)
			if err == nil && user != nil {
				userPublicID = user.PublicID
			}
		}
	}

	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	limit := DefaultPaginationLimit
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= MaxPaginationLimit {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	comments, err := h.commentService.ListCommentsByUserPaginated(ctx, userPublicID, limit, offset)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}

	commentsData := make([]map[string]interface{}, 0, len(comments))

	uniquePostIDs := make(map[string]struct{})
	for _, comment := range comments {
		if comment.PublicPostID != "" {
			uniquePostIDs[comment.PublicPostID] = struct{}{}
		}
	}

	type postInfo struct {
		Title          string
		AuthorUsername string
	}
	postCache := make(map[string]postInfo, len(uniquePostIDs))
	for pid := range uniquePostIDs {
		post, err := h.postService.GetPost(ctx, pid)
		if err == nil && post != nil {
			postCache[pid] = postInfo{Title: post.Title, AuthorUsername: post.AuthorUsername}
		}
	}

	type reactionCounts struct {
		likes    int
		dislikes int
	}
	reactionCache := make(map[string]reactionCounts, len(comments))
	if h.reactionService != nil && len(comments) > 0 {
		commentIDs := make([]string, 0, len(comments))
		for _, comment := range comments {
			commentIDs = append(commentIDs, comment.PublicID)
		}
		batchCounts, err := h.reactionService.CountReactionsBatch(ctx, commentIDs, "comment")
		if err != nil {
			log.Printf("Error batch counting reactions for load-more comments: %v", err)
		} else {
			for id, counts := range batchCounts {
				reactionCache[id] = reactionCounts{
					likes:    counts["like"],
					dislikes: counts["dislike"],
				}
			}
		}
	}

	for _, comment := range comments {
		authorUsername := comment.AuthorUsername

		postTitle := "Post not found"
		postAuthorUsername := "Unknown"
		if pi, ok := postCache[comment.PublicPostID]; ok {
			postTitle = pi.Title
			postAuthorUsername = pi.AuthorUsername
		} else if comment.PublicPostID == "" {
			postTitle = "Post ID unknown"
		}

		rc := reactionCache[comment.PublicID]

		commentData := map[string]interface{}{
			"PublicID":           comment.PublicID,
			"AuthorUsername":     authorUsername,
			"Content":            comment.Content,
			"PostPublicID":       comment.PublicPostID,
			"PostTitle":          postTitle,
			"PostAuthorUsername": postAuthorUsername,
			"CreatedAt":          comment.CreatedAt,
			"UpdatedAt":          comment.UpdatedAt,
			"Likes":              rc.likes,
			"Dislikes":           rc.dislikes,
		}
		commentsData = append(commentsData, commentData)
	}

	httpjson.WriteJSON(w, http.StatusOK, commentsData)
}
