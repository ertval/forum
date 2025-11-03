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
    // POST /posts - Create post
    // GET /posts/{id} - Get post
    // PUT /posts/{id} - Update post
    // DELETE /posts/{id} - Delete post
    // GET /posts - List posts with filters
}
