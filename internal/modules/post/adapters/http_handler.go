// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for post endpoints.
package adapters

import (
	"forum/internal/modules/post/ports"
	"net/http"
)

// HTTPHandler handles HTTP requests for posts.
type HTTPHandler struct {
	postService ports.PostService
}

// NewHTTPHandler creates a new HTTP handler for posts.
func NewHTTPHandler(postService ports.PostService) *HTTPHandler {
	return &HTTPHandler{
		postService: postService,
	}
}

// RegisterRoutes registers all post routes.
// TODO: Implement route registration.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Implementation placeholder
	// POST /posts - Create post (with multipart/form-data for image)
	// GET /posts/{id} - Get post
	// PUT /posts/{id} - Update post
	// DELETE /posts/{id} - Delete post
	// GET /posts - List posts with filters (query params: category, userID, likedBy, offset, limit)
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
