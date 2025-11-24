// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for comment endpoints.
package adapters

import (
	"forum/internal/modules/comment/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for comments.
type HTTPHandler struct {
	commentService ports.CommentService
	templates      *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Comment() ports.CommentService
}

// NewHTTPHandler creates a new HTTP handler for comments with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		commentService: services.Comment(),
		templates:      templates,
	}
}

// RegisterRoutes registers all comment routes.
// TODO: Implement route registration.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Implementation placeholder
	// POST /comments - Create comment
	// GET /comments/{id} - Get comment
	// PUT /comments/{id} - Update comment
	// DELETE /comments/{id} - Delete comment
	// GET /posts/{postId}/comments - List comments for post
}

// CreateCommentAPI handles comment creation requests.
// TODO: Implement comment creation handler.
func (h *HTTPHandler) CreateCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	// 1. Parse request body (postID, content)
	// 2. Get userID from session
	// 3. Call commentService.CreateComment
	// 4. Return 201 Created with comment data
}

// GetCommentAPI handles comment retrieval requests.
// TODO: Implement comment retrieval handler.
func (h *HTTPHandler) GetCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// UpdateCommentAPI handles comment update requests.
// TODO: Implement comment update handler.
func (h *HTTPHandler) UpdateCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// DeleteCommentAPI handles comment deletion requests.
// TODO: Implement comment deletion handler.
func (h *HTTPHandler) DeleteCommentAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ListCommentsByPostAPI handles listing comments for a post.
// TODO: Implement comment listing handler.
func (h *HTTPHandler) ListCommentsByPostAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}
