// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for post endpoints.
package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	logger "forum/internal/platform/logger"

	authAdapters "forum/internal/modules/auth/adapters"
	authPorts "forum/internal/modules/auth/ports"
	userPorts "forum/internal/modules/user/ports"

	postDomain "forum/internal/modules/post/domain"
	postPorts "forum/internal/modules/post/ports"
)

// HTTPHandler handles HTTP requests for posts.
type HTTPHandler struct {
	postService     postPorts.PostService
	categoryService postPorts.CategoryService
	filterService   postPorts.FilterService
	authService     authPorts.AuthService
	userService     userPorts.UserService
	templates       *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Post() postPorts.PostService
	Category() postPorts.CategoryService
	Filter() postPorts.FilterService
	Auth() authPorts.AuthService
	User() userPorts.UserService
}

// NewHTTPHandler creates a new HTTP handler for posts with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		postService:     services.Post(),
		categoryService: services.Category(),
		filterService:   services.Filter(),
		authService:     services.Auth(),
		userService:     services.User(),
		templates:       templates,
	}
}

// Templates returns the shared templates (helper for other handlers).
func (h *HTTPHandler) Templates() *template.Template {
	return h.templates
}

// buildCurrentUser fetches full user info and activity stats and returns
// a map suitable for templates. It always returns a map (never nil).
func (h *HTTPHandler) buildCurrentUser(ctx context.Context, userID int) map[string]interface{} {
	var username, email, publicID string
	var postCount, commentCount int

	if h.userService != nil {
		if user, err := h.userService.GetByID(ctx, userID); err == nil && user != nil {
			username = user.Username
			email = user.Email
			publicID = user.PublicID
		}

		if stats, err := h.userService.GetUserStats(ctx, userID); err == nil && stats != nil {
			postCount = stats.PostCount
			commentCount = stats.CommentCount
		}
	}

	return map[string]interface{}{
		"PublicID":     publicID, // Use explicit PublicID field for templates
		"Username":     username,
		"Email":        email,
		"PostCount":    postCount,
		"CommentCount": commentCount,
	}
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

// buildPageTitle creates a dynamic page title based on active filters.
func (h *HTTPHandler) buildPageTitle(filterParams postPorts.FilterParams) string {
	// Build title parts: [My] [Category] Posts [TimePeriod]
	var parts []string

	// Add "My" if showing user's own posts or liked posts
	if filterParams.MyPosts || filterParams.UserID != "" {
		parts = append(parts, "My")
	} else if filterParams.LikedPosts {
		parts = append(parts, "My Liked")
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

	// Join parts with spaces
	title := ""
	for i, part := range parts {
		if i > 0 {
			title += " "
		}
		title += part
	}

	return title
}

// RegisterRoutes registers all post routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Public routes (no auth required)
	router.HandleFunc("GET /", h.HomePage)
	router.HandleFunc("GET /board", h.BoardPage)
	router.HandleFunc("GET /posts", h.ListPostsAPI)
	router.HandleFunc("GET /posts/{id}", h.GetPostAPI)
	router.HandleFunc("GET /api/posts/load-more", h.LoadMorePostsAPI)

	// Protected routes (require authentication)
	// Wrap handlers with RequireAuth middleware
	authMiddleware := authAdapters.RequireAuth(h.authService, h.userService)
	router.Handle("GET /posts/new", authMiddleware(http.HandlerFunc(h.CreatePostPage)))
	router.Handle("GET /posts/{id}/edit", authMiddleware(http.HandlerFunc(h.EditPostPage)))
	router.Handle("POST /posts", authMiddleware(http.HandlerFunc(h.CreatePostAPI)))
	router.Handle("PUT /posts/{id}", authMiddleware(http.HandlerFunc(h.UpdatePostAPI)))
	router.Handle("DELETE /posts/{id}", authMiddleware(http.HandlerFunc(h.DeletePostAPI)))
}

// HomePage handles the homepage rendering with post list.

// HomePage handles the homepage rendering with post list.
func (h *HTTPHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()

	// Get session token from cookie and build full user info when available
	cookie, err := r.Cookie("session_token")
	var currentUser any = nil // This will hold user info if logged in

	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			currentUser = h.buildCurrentUser(ctx, session.UserID)
		}
	}

	// Parse filter parameters
	var currentUserID string
	if currentUser != nil {
		if userMap, ok := currentUser.(map[string]interface{}); ok {
			if uid, ok := userMap["ID"].(string); ok {
				currentUserID = uid
			}
		}
	}

	// Build filter using FilterService
	filterParams := postPorts.FilterParams{
		Category:      r.URL.Query().Get("category"),
		UserID:        r.URL.Query().Get("user"),
		MyPosts:       r.URL.Query().Get("my_posts") == "true",
		LikedPosts:    r.URL.Query().Get("liked_posts") == "true",
		DateFilter:    r.URL.Query().Get("date_filter"),
		Limit:         12,
		Offset:        0,
		CurrentUserID: currentUserID,
	}

	filter := h.filterService.BuildFilter(ctx, filterParams)

	// Fetch posts
	posts, err := h.postService.ListPosts(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Create preview content for posts on home page
	previewPosts := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		previewPost := make(map[string]interface{})

		// Copy all fields from the original post
		previewPost["ID"] = post.PublicID
		previewPost["PublicID"] = post.PublicID
		previewPost["UserID"] = post.UserPublicID
		previewPost["UserPublicID"] = post.UserPublicID
		previewPost["AuthorUsername"] = post.AuthorUsername
		previewPost["Author"] = post.Author
		previewPost["Title"] = post.Title
		previewPost["Content"] = createPostPreview(post.Content) // Use preview instead of full content
		previewPost["ImageURL"] = post.ImageURL
		previewPost["Categories"] = post.Categories
		previewPost["LikeCount"] = post.LikeCount
		previewPost["DislikeCount"] = post.DislikeCount
		previewPost["CommentCount"] = post.CommentCount
		previewPost["CreatedAt"] = post.CreatedAt
		previewPost["UpdatedAt"] = post.UpdatedAt

		previewPosts[i] = previewPost
	}

	// Fetch all categories for filter dropdown
	var categories []*postDomain.Category
	if h.categoryService != nil {
		categories, err = h.categoryService.List(ctx)
		if err != nil {
			// Log error but continue - categories are not critical
			categories = []*postDomain.Category{}
		}
	}

	// Prepare template data for home page
	data := map[string]interface{}{
		"Title":            "Home",
		"Posts":            previewPosts,
		"Categories":       categories,
		"SelectedCategory": filterParams.Category,
		"DateFilter":       filterParams.DateFilter,
		"MyPosts":          filterParams.MyPosts,
		"LikedPosts":       filterParams.LikedPosts,
		"UserFilter":       filterParams.UserID,
		"User":             currentUser,
		"FilterAction":     "/",
		"ShowFilter":       false,
		"ShowSidebar":      false,
	}

	// Parse templates individually for this page
	tmpl, err := template.ParseFiles("templates/base.html", "templates/home.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// BoardPage handles the board page rendering with post list (identical to homepage).
func (h *HTTPHandler) BoardPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/board" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	var currentUser any = nil // This will hold user info if logged in

	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			currentUser = h.buildCurrentUser(ctx, session.UserID)
		}
	}

	// Parse filter parameters
	var currentUserID string
	if currentUser != nil {
		if userMap, ok := currentUser.(map[string]interface{}); ok {
			if uid, ok := userMap["ID"].(string); ok {
				currentUserID = uid
			}
		}
	}

	// Build filter using FilterService
	filterParams := postPorts.FilterParams{
		Category:      r.URL.Query().Get("category"),
		UserID:        r.URL.Query().Get("user"),
		MyPosts:       r.URL.Query().Get("my_posts") == "true",
		LikedPosts:    r.URL.Query().Get("liked_posts") == "true",
		DateFilter:    r.URL.Query().Get("date_filter"),
		Limit:         10,
		Offset:        0,
		CurrentUserID: currentUserID,
	}

	filter := h.filterService.BuildFilter(ctx, filterParams)

	// Fetch posts
	posts, err := h.postService.ListPosts(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Create preview content for posts on board page
	previewPosts := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		previewPost := make(map[string]interface{})

		// Copy all fields from the original post
		previewPost["ID"] = post.PublicID
		previewPost["PublicID"] = post.PublicID
		previewPost["UserID"] = post.UserPublicID
		previewPost["UserPublicID"] = post.UserPublicID
		previewPost["AuthorUsername"] = post.AuthorUsername
		previewPost["Author"] = post.Author
		previewPost["Title"] = post.Title
		previewPost["Content"] = createPostPreview(post.Content) // Use preview instead of full content
		previewPost["ImageURL"] = post.ImageURL
		previewPost["Categories"] = post.Categories
		previewPost["LikeCount"] = post.LikeCount
		previewPost["DislikeCount"] = post.DislikeCount
		previewPost["CommentCount"] = post.CommentCount
		previewPost["CreatedAt"] = post.CreatedAt
		previewPost["UpdatedAt"] = post.UpdatedAt

		previewPosts[i] = previewPost
	}

	// Fetch all categories for filter dropdown
	var categories []*postDomain.Category
	if h.categoryService != nil {
		categories, err = h.categoryService.List(ctx)
		if err != nil {
			// Log error but continue - categories are not critical
			categories = []*postDomain.Category{}
		}
	}

	// Build dynamic page title based on active filters
	pageTitle := h.buildPageTitle(filterParams)

	// Prepare template data for board page
	data := map[string]interface{}{
		"Title":            "Board",
		"PageTitle":        pageTitle,
		"Posts":            previewPosts,
		"Categories":       categories,
		"SelectedCategory": filterParams.Category,
		"DateFilter":       filterParams.DateFilter,
		"FilterAction":     "/board",
		"ShowFilter":       true,
		"ShowSidebar":      true,
		"MyPosts":          filterParams.MyPosts,
		"LikedPosts":       filterParams.LikedPosts,
		"UserFilter":       filterParams.UserID,
		"User":             currentUser,
	}

	// Render template using the board template
	// Parse templates individually for this page
	tmpl, err := template.ParseFiles("templates/base.html", "templates/board.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// CreatePostAPI handles post creation requests.
func (h *HTTPHandler) CreatePostAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authAdapters.GetUserID(r.Context())
	if userPublicID == "" {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(r.Context(), userPublicID)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user")
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
			h.writeError(w, http.StatusBadRequest, "Invalid form data")
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
				h.writeError(w, http.StatusRequestEntityTooLarge, "Image exceeds 20MB limit")
				return
			}
			imageData, err = io.ReadAll(io.LimitReader(file, maxUploadSize))
			if err != nil {
				h.writeError(w, http.StatusBadRequest, "Failed to read image upload")
				return
			}
		} else if err != http.ErrMissingFile {
			h.writeError(w, http.StatusBadRequest, "Invalid image upload")
			return
		}

	case strings.HasPrefix(contentType, "application/json"), strings.HasPrefix(contentType, "text/json"), contentType == "":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Log decode error to terminal so human output always shows errors.
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
			h.writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	default:
		h.writeError(w, http.StatusUnsupportedMediaType, "Unsupported content type")
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
			h.writeError(w, http.StatusBadRequest, err.Error())
		case postDomain.ErrCategoryNotFound:
			h.writeError(w, http.StatusNotFound, err.Error())
		default:
			h.writeError(w, http.StatusInternalServerError, "Failed to create post")
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
		postID = strings.TrimPrefix(r.URL.Path, "/posts/")
	}

	if postID == "" || postID == "/posts" {
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

	// Check if this is an HTML request (browser) or API request
	// Default to HTML if templates are available and no explicit JSON request
	accept := r.Header.Get("Accept")
	wantsJSON := strings.Contains(accept, "application/json")

	// Render HTML by default if templates exist and not explicitly requesting JSON
	if h.templates != nil && !wantsJSON {
		// Render HTML template for browsers
		h.renderPostDetail(w, r, post)
		return
	}

	// Return JSON for API requests
	h.writeJSON(w, http.StatusOK, post)
}

// UpdatePostAPI handles post update requests.
func (h *HTTPHandler) UpdatePostAPI(w http.ResponseWriter, r *http.Request) {
	// Get user PUBLIC ID (UUID) from context
	userPublicID := authAdapters.GetUserID(r.Context())
	if userPublicID == "" {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(r.Context(), userPublicID)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	// Extract post ID from URL
	postID := r.PathValue("id")
	if postID == "" {
		h.writeError(w, http.StatusBadRequest, "Post ID required")
		return
	}

	// Check ownership
	post, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			h.writeError(w, http.StatusNotFound, "Post not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, "Failed to retrieve post")
		}
		return
	}

	if post.UserID != userID {
		h.writeError(w, http.StatusForbidden, "You can only edit your own posts")
		return
	}

	// Parse request body. The edit form sends a PUT with multipart/form-data (FormData),
	// so support multipart parsing as well as JSON bodies.
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
		// Allow moderately large uploads for edit (same 20MB limit)
		const maxUploadSize = 20 << 20
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid form data")
			return
		}
		req.Title = strings.TrimSpace(r.FormValue("title"))
		req.Content = strings.TrimSpace(r.FormValue("content"))
		// categories may be posted as categories[] or categories
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
			h.writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

	default:
		// Fallback: try reading form values for urlencoded requests
		if err := r.ParseForm(); err == nil {
			req.Title = strings.TrimSpace(r.FormValue("title"))
			req.Content = strings.TrimSpace(r.FormValue("content"))
			// categories may be posted as categories[] or categories
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
		h.writeError(w, http.StatusUnsupportedMediaType, "Unsupported content type")
		return
	}

	// Log parsed data for debugging
	l.Info("http.post.update.parsed",
		logger.String("title", req.Title),
		logger.String("content", req.Content[:min(50, len(req.Content))]), // First 50 chars of content
		logger.Int("category_count", len(req.Categories)))

	if len(req.Categories) > 0 {
		l.Info("http.post.update.categories", logger.String("categories", fmt.Sprintf("%v", req.Categories)))
	}

	// Update post including categories
	if err := h.postService.UpdatePost(r.Context(), postID, req.Title, req.Content, req.Categories); err != nil {
		// Log the actual error before mapping to HTTP error
		l.Error("http.post.update.service_error",
			logger.String("error", err.Error()),
			logger.String("post_id", postID))

		switch err {
		case postDomain.ErrEmptyTitle, postDomain.ErrEmptyContent, postDomain.ErrNoCategories,
			postDomain.ErrTitleTooLong, postDomain.ErrContentTooLong:
			h.writeError(w, http.StatusBadRequest, err.Error())
		case postDomain.ErrPostNotFound:
			h.writeError(w, http.StatusNotFound, "Post not found")
		case postDomain.ErrCategoryNotFound:
			h.writeError(w, http.StatusNotFound, err.Error())
		default:
			h.writeError(w, http.StatusInternalServerError, "Failed to update post")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePostAPI handles post deletion requests.
func (h *HTTPHandler) DeletePostAPI(w http.ResponseWriter, r *http.Request) {
	// Get user PUBLIC ID (UUID) from context
	userPublicID := authAdapters.GetUserID(r.Context())
	if userPublicID == "" {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(r.Context(), userPublicID)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	// Extract post ID from URL
	postID := r.PathValue("id")
	if postID == "" {
		// Fallback: try extracting from URL path
		postID = strings.TrimPrefix(r.URL.Path, "/posts/")
	}
	if postID == "" {
		h.writeError(w, http.StatusBadRequest, "Post ID required")
		return
	}

	// Check ownership
	post, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			h.writeError(w, http.StatusNotFound, "Post not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, "Failed to retrieve post")
		}
		return
	}

	if post.UserID != userID {
		h.writeError(w, http.StatusForbidden, "You can only delete your own posts")
		return
	}

	// Delete post
	if err := h.postService.DeletePost(r.Context(), postID); err != nil {
		if err == postDomain.ErrPostNotFound {
			h.writeError(w, http.StatusNotFound, "Post not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, "Failed to delete post")
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
	filter := postPorts.PostFilter{
		Limit:  50, // Default limit
		Offset: 0,
	}

	// Category filter
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Categories = []string{category}
	}

	// User's own posts filter (requires auth)
	if r.URL.Query().Get("my_posts") == "true" {
		userPublicID := authAdapters.GetUserID(r.Context())
		if userPublicID != "" {
			filter.UserID = userPublicID // Use PublicID (UUID) for filtering
		}
	}

	// Liked posts filter (requires auth)
	if r.URL.Query().Get("liked_posts") == "true" {
		userPublicID := authAdapters.GetUserID(r.Context())
		if userPublicID != "" {
			filter.LikedByUserID = userPublicID // Use PublicID (UUID) for filtering
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
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve posts")
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

	// Parse query parameters
	filter := postPorts.PostFilter{
		Limit:  20, // Load 20 posts at a time
		Offset: 0,
	}

	// Category filter (for when user has filtered by category)
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Categories = []string{category}
	}

	// User's own posts filter (requires auth)
	if r.URL.Query().Get("my_posts") == "true" {
		userPublicID := authAdapters.GetUserID(r.Context())
		if userPublicID != "" {
			filter.UserID = userPublicID // Use PublicID (UUID) for filtering
		}
	}

	// Liked posts filter (requires auth)
	if r.URL.Query().Get("liked_posts") == "true" {
		userPublicID := authAdapters.GetUserID(r.Context())
		if userPublicID != "" {
			filter.LikedByUserID = userPublicID // Use PublicID (UUID) for filtering
		}
	}

	// Pagination - get offset from query parameter (how many posts have been loaded already)
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Get posts
	posts, err := h.postService.ListPosts(r.Context(), filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	// Create preview content for posts (similar to HomePage function)
	previewPosts := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		previewPost := make(map[string]interface{})

		// Copy all fields from the original post
		previewPost["ID"] = post.PublicID
		previewPost["PublicID"] = post.PublicID
		previewPost["UserID"] = post.UserPublicID
		previewPost["UserPublicID"] = post.UserPublicID
		previewPost["AuthorUsername"] = post.AuthorUsername
		previewPost["Author"] = post.Author
		previewPost["Title"] = post.Title
		previewPost["Content"] = createPostPreview(post.Content) // Use preview instead of full content
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

// CreatePostPage renders the post creation form.
func (h *HTTPHandler) CreatePostPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get current user (populate full profile/stats)
	userIDStr := authAdapters.GetUserID(ctx)
	var currentUser interface{}
	if userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			currentUser = h.buildCurrentUser(ctx, userID)
		}
	}

	// Fetch all categories
	categories, err := h.categoryService.List(ctx)
	if err != nil {
		categories = []*postDomain.Category{}
	}

	data := map[string]interface{}{
		"Title":           "Create Post",
		"User":            currentUser,
		"Categories":      categories,
		"ShowSidebar":     true,
		"ShowPostSidebar": true,
	}

	// Parse templates individually for this page
	tmpl, err := template.ParseFiles("templates/base.html", "templates/post_create.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// EditPostPage renders the post edit form.
func (h *HTTPHandler) EditPostPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userIDStr := authAdapters.GetUserID(ctx)
	if userIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to int
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Extract post ID
	postID := r.PathValue("id")
	if postID == "" {
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

	// Check ownership
	if post.UserID != userID {
		http.Error(w, "You can only edit your own posts", http.StatusForbidden)
		return
	}

	// Get current user info (full profile/stats)
	currentUser := h.buildCurrentUser(ctx, userID)

	// Fetch all categories
	categories, err := h.categoryService.List(ctx)
	if err != nil {
		categories = []*postDomain.Category{}
	}

	data := map[string]interface{}{
		"Title":           "Edit Post",
		"User":            currentUser,
		"Post":            post,
		"Categories":      categories,
		"ShowSidebar":     true,
		"ShowPostSidebar": true,
	}

	// Parse templates individually for this page
	tmpl, err := template.ParseFiles("templates/base.html", "templates/post_edit.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// renderPostDetail renders the post detail page with comments.
func (h *HTTPHandler) renderPostDetail(w http.ResponseWriter, r *http.Request, post *postDomain.Post) {
	ctx := r.Context()

	// Get current user if logged in (full profile/stats)
	var currentUser interface{}
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			currentUser = h.buildCurrentUser(ctx, session.UserID)
		}
	}

	// TODO: Fetch comments for this post when comment service is implemented
	var comments []interface{}

	data := map[string]interface{}{
		"Title":    post.Title,
		"User":     currentUser,
		"Post":     post,
		"Comments": comments,
	}

	// Parse templates individually for this page
	tmpl, err := template.ParseFiles("templates/base.html", "templates/post_detail.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		// Log the actual template error for debugging
		fmt.Printf("Template error: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to render page: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// writeJSON writes a JSON response.
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response.
func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Also log the error to the terminal so human output always includes error messages.
	cfg := &logger.Config{
		TimePrecision: logger.TimePrecisionSeconds,
		AllowedFields: []string{"error", "errors", "status"},
		MaxLineWidth:  200,
	}
	l := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
	l.Error("http.handler.error", logger.String("error", message), logger.Int("status", status))

	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
