// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for post endpoints.
package adapters

import (
	"fmt"
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/post/domain"
	postPorts "forum/internal/modules/post/ports"
	userPorts "forum/internal/modules/user/ports"
	"html/template"
	"net/http"
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

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	var currentUser interface{} = nil // This will hold user info if logged in

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
	var categories []*domain.Category
	if h.categoryService != nil {
		categories, err = h.categoryService.List(ctx)
		if err != nil {
			// Log error but continue - categories are not critical
			categories = []*domain.Category{}
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
