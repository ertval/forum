// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for post endpoints.
package adapters

import (
	"fmt"
	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for posts.
type HTTPHandler struct {
	postService     ports.PostService
	categoryService ports.CategoryService
	templates       *template.Template
}

// NewHTTPHandler creates a new HTTP handler for posts.
func NewHTTPHandler(postService ports.PostService) *HTTPHandler {
	// Parse templates
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		panic(fmt.Sprintf("failed to parse templates: %v", err))
	}

	return &HTTPHandler{
		postService: postService,
		templates:   tmpl,
	}
}

// SetCategoryService sets the category service (optional dependency).
func (h *HTTPHandler) SetCategoryService(categoryService ports.CategoryService) {
	h.categoryService = categoryService
}

// RegisterRoutes registers all post routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/", h.HomePage)
	router.HandleFunc("/posts", h.ListPosts)
	router.HandleFunc("/posts/", h.handlePostRoutes)
}

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

	// Parse filter parameters
	category := r.URL.Query().Get("category")
	myPosts := r.URL.Query().Get("my_posts") == "true"
	likedPosts := r.URL.Query().Get("liked_posts") == "true"

	// Build filter
	filter := ports.PostFilter{
		Limit: 50, // Default limit
	}

	if category != "" {
		filter.Categories = []string{category}
	}

	// TODO: Get user from session context
	// For now, we'll skip user-specific filters
	var currentUser interface{}
	// currentUser = ctx.Value("user")

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
	var categories []*domain.Category
	if h.categoryService != nil {
		categories, err = h.categoryService.List(ctx)
		if err != nil {
			// Log error but continue - categories are not critical
			categories = []*domain.Category{}
		}
	}

	// Prepare template data
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
// TODO: Implement post creation handler with image upload.
func (h *HTTPHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	// 1. Parse multipart form data
	// 2. Extract title, content, categories from form
	// 3. Extract optional image file
	// 4. Get userID from session
	// 5. Call postService.CreatePost
	// 6. Return 201 Created with post data
}

// GetPost handles post retrieval requests.
// TODO: Implement post retrieval handler.
func (h *HTTPHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// UpdatePost handles post update requests.
// TODO: Implement post update handler.
func (h *HTTPHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// DeletePost handles post deletion requests.
// TODO: Implement post deletion handler.
func (h *HTTPHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ListPosts handles listing posts with filters.
// TODO: Implement post listing handler with filters.
func (h *HTTPHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	// Parse query parameters for filters
	// Call postService.ListPosts with filter
	// Return posts array
}
