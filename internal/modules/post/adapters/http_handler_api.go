// INPUT ADAPTER - HTTP API Handler
// Package adapters implements HTTP API handlers for post endpoints.
// This adapter handles JSON API requests for post operations.
package adapters

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	postDomain "forum/internal/modules/post/domain"
	platformErrors "forum/internal/platform/errors"
	"forum/internal/platform/httpjson"
	logger "forum/internal/platform/logger"
	"forum/internal/platform/upload"
)

// postRequestData holds the parsed fields from a create/update post request.
type postRequestData struct {
	Title       string
	Content     string
	Categories  []string
	ImageData   []byte
	RemoveImage bool
}

// parsePostRequest extracts title, content, categories, image data, and remove-image
// flag from an HTTP request regardless of content type (multipart, JSON, form-encoded).
func parsePostRequest(r *http.Request, maxUploadSize int64) (*postRequestData, error) {
	data := &postRequestData{}
	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			return nil, &parseError{status: http.StatusBadRequest, msg: "Invalid form data"}
		}

		data.Title = strings.TrimSpace(r.FormValue("title"))
		data.Content = strings.TrimSpace(r.FormValue("content"))
		data.RemoveImage = r.FormValue("remove_image") == "true"
		data.Categories = r.Form["categories[]"]
		if len(data.Categories) == 0 {
			data.Categories = r.Form["categories"]
		}
		if len(data.Categories) == 0 {
			if csv := strings.TrimSpace(r.FormValue("categories")); csv != "" {
				data.Categories = strings.Split(csv, ",")
				for i := range data.Categories {
					data.Categories[i] = strings.TrimSpace(data.Categories[i])
				}
			}
		}

		// Try reading image from multipart
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			if header.Size > maxUploadSize {
				return nil, &parseError{status: http.StatusRequestEntityTooLarge, msg: upload.FormatImageSizeError(maxUploadSize)}
			}
			data.ImageData, err = io.ReadAll(io.LimitReader(file, maxUploadSize))
			if err != nil {
				return nil, &parseError{status: http.StatusBadRequest, msg: "Failed to read image upload"}
			}
		} else if err != http.ErrMissingFile {
			return nil, &parseError{status: http.StatusBadRequest, msg: "Invalid image upload"}
		}

	case strings.HasPrefix(contentType, "application/json"),
		strings.HasPrefix(contentType, "text/json"),
		contentType == "":
		var req struct {
			Title       string   `json:"title"`
			Content     string   `json:"content"`
			Categories  []string `json:"categories"`
			RemoveImage bool     `json:"remove_image"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, &parseError{status: http.StatusBadRequest, msg: "Invalid request body", decodeErr: err}
		}
		data.Title = strings.TrimSpace(req.Title)
		data.Content = strings.TrimSpace(req.Content)
		data.Categories = req.Categories
		data.RemoveImage = req.RemoveImage

	default:
		// Try form-encoded as fallback for application/x-www-form-urlencoded
		if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err == nil {
				data.Title = strings.TrimSpace(r.FormValue("title"))
				data.Content = strings.TrimSpace(r.FormValue("content"))
				data.RemoveImage = r.FormValue("remove_image") == "true"
				data.Categories = r.Form["categories[]"]
				if len(data.Categories) == 0 {
					data.Categories = r.Form["categories"]
				}
				if len(data.Categories) == 0 {
					if csv := strings.TrimSpace(r.FormValue("categories")); csv != "" {
						data.Categories = strings.Split(csv, ",")
						for i := range data.Categories {
							data.Categories[i] = strings.TrimSpace(data.Categories[i])
						}
					}
				}
			}
		} else {
			return nil, &parseError{status: http.StatusUnsupportedMediaType, msg: "Unsupported content type"}
		}
	}

	// Trim category names
	if len(data.Categories) > 0 {
		filtered := make([]string, 0, len(data.Categories))
		for _, cat := range data.Categories {
			if trimmed := strings.TrimSpace(cat); trimmed != "" {
				filtered = append(filtered, trimmed)
			}
		}
		data.Categories = filtered
	}

	return data, nil
}

// parseError is a typed error wrapping HTTP status and message for parse failures.
type parseError struct {
	status    int
	msg       string
	decodeErr error // optional underlying JSON decode error
}

func (e *parseError) Error() string { return e.msg }

// RegisterAPIRoutes registers all post API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	// Public API routes with optional auth (needed for my_posts/liked_posts filters)
	optionalAuth := h.middlewareProvider.OptionalAuth()
	router.Handle("GET /api/posts", optionalAuth(http.HandlerFunc(h.ListPostsAPI)))
	router.Handle("GET /api/posts/load-more", optionalAuth(http.HandlerFunc(h.LoadMorePostsAPI)))

	// Public API route without auth
	router.HandleFunc("GET /api/posts/{id}", h.GetPostAPI)

	// Protected API routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("POST /api/posts", authMiddleware(http.HandlerFunc(h.CreatePostAPI)))
	router.Handle("PUT /api/posts/{id}", authMiddleware(http.HandlerFunc(h.UpdatePostAPI)))
	router.Handle("DELETE /api/posts/{id}", authMiddleware(http.HandlerFunc(h.DeletePostAPI)))
}

// CreatePostAPI handles post creation requests.
func (h *HTTPHandler) CreatePostAPI(w http.ResponseWriter, r *http.Request) {
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
	data, err := parsePostRequest(r, h.postService.MaxImageSize())
	if err != nil {
		if pe, ok := err.(*parseError); ok {
			if pe.decodeErr != nil {
				h.logger.Error("http.request.error",
					logger.String("url", r.URL.RequestURI()),
					logger.String("error", pe.decodeErr.Error()),
				)
			}
			platformErrors.WriteErrorJSON(w, pe.status, pe.msg)
		} else {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// Create post
	post, err := h.postService.CreatePost(r.Context(), userID, data.Title, data.Content, data.Categories, data.ImageData)
	if err != nil {
		switch err {
		case postDomain.ErrEmptyTitle, postDomain.ErrEmptyContent, postDomain.ErrNoCategories,
			postDomain.ErrTitleTooLong, postDomain.ErrContentTooLong:
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case postDomain.ErrCategoryNotFound:
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		case upload.ErrInvalidImageType, postDomain.ErrInvalidImageType:
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid image type, must be JPEG, PNG, GIF, or WebP")
		case postDomain.ErrInvalidImage:
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid image file")
		case upload.ErrImageTooLarge, postDomain.ErrImageTooLarge:
			platformErrors.WriteErrorJSON(w, http.StatusRequestEntityTooLarge, upload.FormatImageSizeError(h.postService.MaxImageSize()))
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create post")
		}
		return
	}

	httpjson.WriteJSON(w, http.StatusCreated, post)
}

// GetPostAPI handles post retrieval requests.
func (h *HTTPHandler) GetPostAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract post ID from path variable (Go 1.22+ pattern)
	postID := r.PathValue("id")
	if postID == "" {
		// Fallback: try extracting from URL path
		postID = strings.TrimPrefix(r.URL.Path, "/api/posts/")
	}

	if postID == "" || postID == "/api/posts" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID required")
		return
	}

	// Get post
	post, err := h.postService.GetPost(ctx, postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Post not found")
		} else {
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve post")
		}
		return
	}

	// Return JSON for API requests
	httpjson.WriteJSON(w, http.StatusOK, post)
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
	currentUser, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || currentUser == nil {
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

	// Log incoming request for debugging
	contentType := r.Header.Get("Content-Type")
	h.logger.Info("http.post.update.request",
		logger.String("method", r.Method),
		logger.String("content_type", contentType),
		logger.String("post_id", postID))

	// Parse request body
	data, err := parsePostRequest(r, h.postService.MaxImageSize())
	if err != nil {
		if pe, ok := err.(*parseError); ok {
			platformErrors.WriteErrorJSON(w, pe.status, pe.msg)
		} else {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// For multipart update, also check the dedicated image upload helper for advanced
	// image handling (e.g. ParseImageUpload with remove detection).
	if strings.HasPrefix(contentType, "multipart/form-data") {
		imageResult, err := ParseImageUpload(r, "image", h.postService.MaxImageSize())
		if err != nil {
			if err == upload.ErrImageTooLarge {
				platformErrors.WriteErrorJSON(w, http.StatusRequestEntityTooLarge, FormatImageError(err, h.postService.MaxImageSize()))
			} else {
				platformErrors.WriteErrorJSON(w, http.StatusBadRequest, FormatImageError(err, h.postService.MaxImageSize()))
			}
			return
		}
		if len(imageResult.Data) > 0 {
			data.ImageData = imageResult.Data
		}
		if imageResult.RemoveImage {
			data.RemoveImage = true
		}
	}

	// Update post including categories
	if err := h.postService.UpdatePost(r.Context(), postID, data.Title, data.Content, data.Categories); err != nil {
		h.logger.Error("http.post.update.service_error",
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

	// Handle image update if there's new image data or removal request
	if len(data.ImageData) > 0 || data.RemoveImage {
		if err := h.postService.UpdatePostImage(r.Context(), postID, data.ImageData, data.RemoveImage); err != nil {
			h.logger.Error("http.post.update.image_error",
				logger.String("error", err.Error()),
				logger.String("post_id", postID))
			// Don't fail the whole request, post content was already updated
			// Just log the error
		}
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

	currentUser, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || currentUser == nil {
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

	canDeleteAny := currentUser.Role == "moderator" || currentUser.Role == "admin"
	if post.UserID != userID && !canDeleteAny {
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

func parseListPagination(r *http.Request, defaultLimit int) (int, int) {
	limit := defaultLimit
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}

func buildListFilterOptions(r *http.Request, currentUserPublicID string, defaultLimit int) listFilterOptions {
	limit, offset := parseListPagination(r, defaultLimit)

	filterOptions := listFilterOptions{
		Category:       r.URL.Query().Get("category"),
		UserID:         r.URL.Query().Get("user"),
		ActivityType:   r.URL.Query().Get("activity_type"),
		ReactionType:   r.URL.Query().Get("reaction_type"),
		MyPosts:        r.URL.Query().Get("my_posts") == "true",
		LikedPosts:     r.URL.Query().Get("liked_posts") == "true",
		DislikedPosts:  r.URL.Query().Get("disliked_posts") == "true",
		CommentedPosts: r.URL.Query().Get("commented_posts") == "true",
		Commenter:      r.URL.Query().Get("commenter"),
		DateFilter:     r.URL.Query().Get("date_filter"),
		Limit:          limit,
		Offset:         offset,
		CurrentUserID:  currentUserPublicID,
	}
	if filterOptions.Commenter == "" && filterOptions.CommentedPosts && currentUserPublicID != "" {
		filterOptions.Commenter = currentUserPublicID
	}

	return filterOptions
}

func (h *HTTPHandler) listPostsForRequest(r *http.Request, currentUserPublicID string, defaultLimit int) ([]*postDomain.Post, error) {
	filterOptions := buildListFilterOptions(r, currentUserPublicID, defaultLimit)
	filter := buildPostFilter(filterOptions)
	return h.postService.ListPosts(r.Context(), filter)
}

// ListPostsAPI handles listing posts with filters.
func (h *HTTPHandler) ListPostsAPI(w http.ResponseWriter, r *http.Request) {
	currentUserPublicID := authPorts.GetUserID(r.Context())
	posts, err := h.listPostsForRequest(r, currentUserPublicID, 50)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	// Return posts wrapped in an object for consistent API response
	response := map[string]interface{}{
		"posts": posts,
	}
	httpjson.WriteJSON(w, http.StatusOK, response)
}

// LoadMorePostsAPI handles loading additional posts for the homepage.
func (h *HTTPHandler) LoadMorePostsAPI(w http.ResponseWriter, r *http.Request) {
	currentUserPublicID := authPorts.GetUserID(r.Context())
	posts, err := h.listPostsForRequest(r, currentUserPublicID, 20)
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

	httpjson.WriteJSON(w, http.StatusOK, previewPosts)
}
