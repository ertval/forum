// INPUT ADAPTER - HTTP Handler
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements HTTP handlers for moderation endpoints.
package adapters

import (
	"forum/internal/modules/moderation/ports"
	"net/http"
)

// HTTPHandler handles HTTP requests for moderation.
type HTTPHandler struct {
	moderationService ports.ModerationService
}

// NewHTTPHandler creates a new HTTP handler for moderation.
func NewHTTPHandler(moderationService ports.ModerationService) *HTTPHandler {
	return &HTTPHandler{moderationService: moderationService}
}

// RegisterRoutes registers all moderation routes.
// TODO: Implement route registration.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Implementation placeholder
	// POST /reports - Create report
	// GET /reports - List reports (filtered by status)
	// PUT /reports/{id} - Review report
}

// CreateReport handles report creation requests.
// TODO: Implement report creation handler.
func (h *HTTPHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ListReports handles listing reports.
// TODO: Implement report listing handler.
func (h *HTTPHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ReviewReport handles report review requests.
// TODO: Implement report review handler.
func (h *HTTPHandler) ReviewReport(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}
