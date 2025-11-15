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

// AddReaction handles reaction creation/update requests.
// TODO: Implement reaction handler.
func (h *HTTPHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	// 1. Parse request body (targetID, targetType, reactionType)
	// 2. Get userID from session
	// 3. Call reactionService.React
	// 4. Return 200 OK
}

// RemoveReaction handles reaction removal requests.
// TODO: Implement reaction removal handler.
func (h *HTTPHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// GetReactions handles retrieving reactions for a target.
// TODO: Implement get reactions handler.
func (h *HTTPHandler) GetReactions(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// CountReactions handles counting reactions for a target.
// TODO: Implement count reactions handler.
func (h *HTTPHandler) CountReactions(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}
