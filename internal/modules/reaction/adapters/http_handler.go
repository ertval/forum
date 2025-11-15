// INPUT ADAPTER - HTTP Handler
// Package adapters implements HTTP handlers for reaction endpoints.
package adapters

import (
	"forum/internal/modules/reaction/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for reactions.
type HTTPHandler struct {
	reactionService ports.ReactionService
	templates       *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Reaction() ports.ReactionService
}

// NewHTTPHandler creates a new HTTP handler for reactions with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		reactionService: services.Reaction(),
		templates:       templates,
	}
}

// RegisterRoutes registers all reaction routes.
// TODO: Implement route registration.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Implementation placeholder
	// POST /reactions - Add or update reaction
	// DELETE /reactions - Remove reaction
	// GET /reactions/{targetType}/{targetId} - Get reactions for target
	// GET /reactions/{targetType}/{targetId}/count - Count reactions
}

// AddReactionAPI handles adding a reaction to a post or comment.
// TODO: Implement reaction addition handler.
func (h *HTTPHandler) AddReactionAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	// 1. Parse request body (targetID, targetType, reactionType)
	// 2. Get userID from session
	// 3. Call reactionService.React
	// 4. Return 200 OK
}

// RemoveReactionAPI handles removing a reaction from a post or comment.
// TODO: Implement reaction removal handler.
func (h *HTTPHandler) RemoveReactionAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// GetReactionsAPI handles retrieving reactions for a post or comment.
// TODO: Implement reaction retrieval handler.
func (h *HTTPHandler) GetReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// CountReactionsAPI handles counting reactions for a target.
// TODO: Implement count reactions handler.
func (h *HTTPHandler) CountReactionsAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}
