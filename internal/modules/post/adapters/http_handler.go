// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for post endpoints.
package adapters

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

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
	authService     authPorts.AuthService
	userService     userPorts.UserService
	templates       *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Post() postPorts.PostService
	Category() postPorts.CategoryService
	Auth() authPorts.AuthService
	User() userPorts.UserService
}

// NewHTTPHandler creates a new HTTP handler for posts with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		postService:     services.Post(),
		categoryService: services.Category(),
		authService:     services.Auth(),
		userService:     services.User(),
		templates:       templates,
	}
}

// RegisterRoutes registers all post routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Public routes (no auth required)
	router.HandleFunc("GET /", h.HomePage)
	router.HandleFunc("GET /posts", h.ListPosts)
	router.HandleFunc("GET /posts/{id}", h.GetPost)

	// Protected routes (require authentication)
	// Wrap handlers with RequireAuth middleware
	authMiddleware := authAdapters.RequireAuth(h.authService)
	router.Handle("POST /posts", authMiddleware(http.HandlerFunc(h.CreatePost)))
	router.Handle("PUT /posts/{id}", authMiddleware(http.HandlerFunc(h.UpdatePost)))
	router.Handle("DELETE /posts/{id}", authMiddleware(http.HandlerFunc(h.DeletePost)))
}

// HomePage handles the homepage rendering with post list.

// handlePostRoutes handles routes with post ID parameter.
func (h *HTTPHandler) handlePostRoutes(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from path
	// For now, just handle basic CRUD operations
	// TODO: Implement proper routing for /posts/{id}
	switch r.Method {
	case http.MethodGet:
		h.GetPost(w, r)
	case http.MethodPut:
		h.UpdatePost(w, r)
	case http.MethodDelete:
		h.DeletePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HomePage handles the homepage rendering with post list.
func (h *HTTPHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	var currentUser any = nil // This will hold user info if logged in

	if err == nil && cookie.Value != "" {
		// Validate the session using the auth service
		session, err := h.authService.ValidateSession(ctx, cookie.Value)
		if err == nil && session != nil {
			// Get actual user details using user service
			user, err := h.userService.GetByID(ctx, session.UserID)
			if err == nil && user != nil {
				currentUser = map[string]interface{}{
					"ID":       user.ID,
					"Username": user.Username,
				}
			} else {
				// If we can't get user details, still create with minimal info
				currentUser = map[string]interface{}{
					"ID":       session.UserID,
					"Username": "user" + fmt.Sprintf("%d", session.UserID),
				}
			}
		}
	}

	// Parse filter parameters
	category := r.URL.Query().Get("category")
	myPosts := r.URL.Query().Get("my_posts") == "true"
	likedPosts := r.URL.Query().Get("liked_posts") == "true"

	// Build filter
	filter := postPorts.PostFilter{
		Limit: 50, // Default limit
	}

	if category != "" {
		filter.Categories = []string{category}
	}

	if myPosts && currentUser != nil {
		// filter.UserID = currentUser.ID
	}

	if likedPosts && currentUser != nil {
		// filter.LikedByUserID = currentUser.ID
	}

	// Fetch posts
	posts, err := h.postService.ListPosts(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
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
		"Posts":            posts,
		"Categories":       categories,
		"SelectedCategory": category,
		"MyPosts":          myPosts,
		"LikedPosts":       likedPosts,
		"User":             currentUser,
	}

	// Render template
	if err := h.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// CreatePost handles post creation requests.
func (h *HTTPHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by RequireAuth middleware)
	userID := authAdapters.GetUserID(r.Context())
	if userID == "" {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse JSON request
	var req struct {
		Title      string   `json:"title"`
		Content    string   `json:"content"`
		Categories []string `json:"categories"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create post
	post, err := h.postService.CreatePost(r.Context(), userID, req.Title, req.Content, req.Categories, nil)
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

// GetPost handles post retrieval requests.
func (h *HTTPHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from path variable (Go 1.22+ pattern)
	postID := r.PathValue("id")
	if postID == "" {
		// Fallback: try extracting from URL path
		postID = strings.TrimPrefix(r.URL.Path, "/posts/")
	}

	if postID == "" || postID == "/posts" {
		h.writeError(w, http.StatusBadRequest, "Post ID required")
		return
	}

	// Get post
	post, err := h.postService.GetPost(r.Context(), postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			h.writeError(w, http.StatusNotFound, "Post not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, "Failed to retrieve post")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, post)
}

// UpdatePost handles post update requests.
func (h *HTTPHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID := authAdapters.GetUserID(r.Context())
	if userID == "" {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract post ID from URL
	postID := strings.TrimPrefix(r.URL.Path, "/posts/")
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

	// Parse JSON request
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update post
	if err := h.postService.UpdatePost(r.Context(), postID, req.Title, req.Content); err != nil {
		switch err {
		case postDomain.ErrEmptyTitle, postDomain.ErrEmptyContent,
			postDomain.ErrTitleTooLong, postDomain.ErrContentTooLong:
			h.writeError(w, http.StatusBadRequest, err.Error())
		case postDomain.ErrPostNotFound:
			h.writeError(w, http.StatusNotFound, "Post not found")
		default:
			h.writeError(w, http.StatusInternalServerError, "Failed to update post")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePost handles post deletion requests.
func (h *HTTPHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID := authAdapters.GetUserID(r.Context())
	if userID == "" {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
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

// ListPosts handles listing posts with filters.
func (h *HTTPHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
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
		userID := authAdapters.GetUserID(r.Context())
		if userID != "" {
			filter.UserID = userID
		}
	}

	// Liked posts filter (requires auth)
	if r.URL.Query().Get("liked_posts") == "true" {
		userID := authAdapters.GetUserID(r.Context())
		if userID != "" {
			filter.LikedByUserID = userID
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

	h.writeJSON(w, http.StatusOK, posts)
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
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
