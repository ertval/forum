// INPUT ADAPTER - HTTP API Handler
// Package adapters implements HTTP API handlers for post endpoints.
// This adapter handles JSON API requests for post operations.
package adapters

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	postDomain "forum/internal/modules/post/domain"
	platformErrors "forum/internal/platform/errors"
	logger "forum/internal/platform/logger"
)

// RegisterAPIRoutes registers all post API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	// Public API routes (no auth required)
	router.HandleFunc("GET /api/posts", h.ListPostsAPI)
	router.HandleFunc("GET /api/posts/{id}", h.GetPostAPI)
	router.HandleFunc("GET /api/posts/load-more", h.LoadMorePostsAPI)

	// Protected API routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("POST /api/posts", authMiddleware(http.HandlerFunc(h.CreatePostAPI)))
	router.Handle("PUT /api/posts/{id}", authMiddleware(http.HandlerFunc(h.UpdatePostAPI)))
	router.Handle("DELETE /api/posts/{id}", authMiddleware(http.HandlerFunc(h.DeletePostAPI)))
}

// CreatePostAPI handles post creation requests.
func (h *HTTPHandler) CreatePostAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Parse request body (JSON or multipart form submissions from browser)
	const maxUploadSize = 20 << 20 // 20MB limit
	var (
		req struct {
			Title      string   `json:"title"`
			Content    string   `json:"content"`
			Categories []string `json:"categories"`
		}
		imageData []byte
	)

	contentType := r.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid form data")
			return
		}

		req.Title = strings.TrimSpace(r.FormValue("title"))
		req.Content = strings.TrimSpace(r.FormValue("content"))
		req.Categories = r.Form["categories[]"]
		if len(req.Categories) == 0 {
			req.Categories = r.Form["categories"]
		}
		if len(req.Categories) == 0 {
			if csv := strings.TrimSpace(r.FormValue("categories")); csv != "" {
				req.Categories = strings.Split(csv, ",")
				for i := range req.Categories {
					req.Categories[i] = strings.TrimSpace(req.Categories[i])
				}
			}
		}

		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			if header.Size > maxUploadSize {
				platformErrors.WriteErrorJSON(w, http.StatusRequestEntityTooLarge, "Image exceeds 20MB limit")
				return
			}
			imageData, err = io.ReadAll(io.LimitReader(file, maxUploadSize))
			if err != nil {
				platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Failed to read image upload")
				return
			}
		} else if err != http.ErrMissingFile {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid image upload")
			return
		}

	case strings.HasPrefix(contentType, "application/json"), strings.HasPrefix(contentType, "text/json"), contentType == "":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Log decode error to terminal
			cfg := &logger.Config{
				TimePrecision: logger.TimePrecisionSeconds,
				AllowedFields: []string{"url", "error", "errors"},
				MaxLineWidth:  120,
			}
			l := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
			l.Error("http.request.error",
				logger.String("url", r.URL.RequestURI()),
				logger.String("error", err.Error()),
			)
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	default:
		platformErrors.WriteErrorJSON(w, http.StatusUnsupportedMediaType, "Unsupported content type")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	if len(req.Categories) > 0 {
		filtered := make([]string, 0, len(req.Categories))
		for _, cat := range req.Categories {
			if trimmed := strings.TrimSpace(cat); trimmed != "" {
				filtered = append(filtered, trimmed)
			}
		}
		req.Categories = filtered
	}

	// Create post
	post, err := h.postService.CreatePost(r.Context(), userID, req.Title, req.Content, req.Categories, imageData)
	if err != nil {
		switch err {
		case postDomain.ErrEmptyTitle, postDomain.ErrEmptyContent, postDomain.ErrNoCategories,
			postDomain.ErrTitleTooLong, postDomain.ErrContentTooLong:
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case postDomain.ErrCategoryNotFound:
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create post")
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, post)
}

// GetPostAPI handles post retrieval requests.
func (h *HTTPHandler) GetPostAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Extract post ID from path variable (Go 1.22+ pattern)
	postID := r.PathValue("id")
	if postID == "" {
		// Fallback: try extracting from URL path
		postID = strings.TrimPrefix(r.URL.Path, "/api/posts/")
	}

	if postID == "" || postID == "/api/posts" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	// Get post
	post, err := h.postService.GetPost(ctx, postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve post", http.StatusInternalServerError)
		}
		return
	}

	// Return JSON for API requests
	h.writeJSON(w, http.StatusOK, post)
}

// UpdatePostAPI handles post update requests.
func (h *HTTPHandler) UpdatePostAPI(w http.ResponseWriter, r *http.Request) {
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

	// Extract post ID from URL
	postID := r.PathValue("id")
	if postID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID required")
		return
	}

	// Check ownership
	post, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Post not found")
		} else {
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve post")
		}
		return
	}

	if post.UserID != userID {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "You can only edit your own posts")
		return
	}

	// Parse request body
	var req struct {
		Title      string   `json:"title"`
		Content    string   `json:"content"`
		Categories []string `json:"categories"`
	}

	contentType := r.Header.Get("Content-Type")

	// Log incoming request for debugging
	cfg := &logger.Config{
		TimePrecision: logger.TimePrecisionSeconds,
		AllowedFields: []string{"url", "method", "content_type", "post_id"},
		MaxLineWidth:  200,
	}
	l := logger.NewWithConfig(logger.InfoLevel, os.Stderr, cfg)
	l.Info("http.post.update.request",
		logger.String("method", r.Method),
		logger.String("content_type", contentType),
		logger.String("post_id", postID))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		const maxUploadSize = 20 << 20
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid form data")
			return
		}
		req.Title = strings.TrimSpace(r.FormValue("title"))
		req.Content = strings.TrimSpace(r.FormValue("content"))
		req.Categories = r.Form["categories[]"]
		if len(req.Categories) == 0 {
			req.Categories = r.Form["categories"]
		}
		if len(req.Categories) == 0 {
			if csv := strings.TrimSpace(r.FormValue("categories")); csv != "" {
				req.Categories = strings.Split(csv, ",")
				for i := range req.Categories {
					req.Categories[i] = strings.TrimSpace(req.Categories[i])
				}
			}
		}

	case strings.HasPrefix(contentType, "application/json"), strings.HasPrefix(contentType, "text/json"), contentType == "":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
			return
		}

	default:
		if err := r.ParseForm(); err == nil {
			req.Title = strings.TrimSpace(r.FormValue("title"))
			req.Content = strings.TrimSpace(r.FormValue("content"))
			req.Categories = r.Form["categories[]"]
			if len(req.Categories) == 0 {
				req.Categories = r.Form["categories"]
			}
			if len(req.Categories) == 0 {
				if csv := strings.TrimSpace(r.FormValue("categories")); csv != "" {
					req.Categories = strings.Split(csv, ",")
					for i := range req.Categories {
						req.Categories[i] = strings.TrimSpace(req.Categories[i])
					}
				}
			}
			break
		}
		platformErrors.WriteErrorJSON(w, http.StatusUnsupportedMediaType, "Unsupported content type")
		return
	}

	// Update post including categories
	if err := h.postService.UpdatePost(r.Context(), postID, req.Title, req.Content, req.Categories); err != nil {
		l.Error("http.post.update.service_error",
			logger.String("error", err.Error()),
			logger.String("post_id", postID))

		switch err {
		case postDomain.ErrEmptyTitle, postDomain.ErrEmptyContent, postDomain.ErrNoCategories,
			postDomain.ErrTitleTooLong, postDomain.ErrContentTooLong:
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case postDomain.ErrPostNotFound:
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Post not found")
		case postDomain.ErrCategoryNotFound:
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to update post")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePostAPI handles post deletion requests.
func (h *HTTPHandler) DeletePostAPI(w http.ResponseWriter, r *http.Request) {
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

	// Extract post ID from URL
	postID := r.PathValue("id")
	if postID == "" {
		postID = strings.TrimPrefix(r.URL.Path, "/api/posts/")
	}
	if postID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID required")
		return
	}

	// Check ownership
	post, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Post not found")
		} else {
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve post")
		}
		return
	}

	if post.UserID != userID {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "You can only delete your own posts")
		return
	}

	// Delete post
	if err := h.postService.DeletePost(r.Context(), postID); err != nil {
		if err == postDomain.ErrPostNotFound {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Post not found")
		} else {
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to delete post")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListPostsAPI handles listing posts with filters.
func (h *HTTPHandler) ListPostsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	filter := postDomain.PostFilter{
		Limit:  50, // Default limit
		Offset: 0,
	}

	// Category filter
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Categories = []string{category}
	}

	// User's own posts filter (requires auth)
	if r.URL.Query().Get("my_posts") == "true" {
		userPublicID := authPorts.GetUserID(r.Context())
		if userPublicID != "" {
			filter.UserID = userPublicID
		}
	}

	// Liked posts filter (requires auth)
	if r.URL.Query().Get("liked_posts") == "true" {
		userPublicID := authPorts.GetUserID(r.Context())
		if userPublicID != "" {
			filter.LikedByUserID = userPublicID
		}
	}

	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Get posts
	posts, err := h.postService.ListPosts(r.Context(), filter)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	// Return posts wrapped in an object for consistent API response
	response := map[string]interface{}{
		"posts": posts,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// LoadMorePostsAPI handles loading additional posts for the homepage.
func (h *HTTPHandler) LoadMorePostsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from session cookie if available.
	var userPublicID string
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			user, err := h.userService.GetByID(ctx, session.UserID)
			if err == nil && user != nil {
				userPublicID = user.PublicID
			}
		}
	}

	// Parse query parameters
	filter := postDomain.PostFilter{
		Limit:  20, // Load 20 posts at a time
		Offset: 0,
	}

	// Category filter
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Categories = []string{category}
	}

	// User's own posts filter (requires auth)
	if r.URL.Query().Get("my_posts") == "true" {
		if userPublicID != "" {
			filter.UserID = userPublicID
		}
	}

	// Liked posts filter (requires auth)
	if r.URL.Query().Get("liked_posts") == "true" {
		if userPublicID != "" {
			filter.LikedByUserID = userPublicID
		}
	}

	// Pagination
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Date filter
	if dateFilter := r.URL.Query().Get("date_filter"); dateFilter != "" {
		filter.DateFilter = dateFilter
	}

	// Get posts
	posts, err := h.postService.ListPosts(r.Context(), filter)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	// Create preview content for posts
	previewPosts := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		previewPost := make(map[string]interface{})

		previewPost["ID"] = post.PublicID
		previewPost["PublicID"] = post.PublicID
		previewPost["UserID"] = post.UserPublicID
		previewPost["UserPublicID"] = post.UserPublicID
		previewPost["AuthorUsername"] = post.AuthorUsername
		previewPost["Author"] = post.Author
		previewPost["Title"] = post.Title
		previewPost["Content"] = createPostPreview(post.Content)
		previewPost["ImageURL"] = post.ImageURL
		previewPost["Categories"] = post.Categories
		previewPost["LikeCount"] = post.LikeCount
		previewPost["DislikeCount"] = post.DislikeCount
		previewPost["CommentCount"] = post.CommentCount
		previewPost["CreatedAt"] = post.CreatedAt
		previewPost["UpdatedAt"] = post.UpdatedAt

		previewPosts[i] = previewPost
	}

	h.writeJSON(w, http.StatusOK, previewPosts)
}
