// INPUT ADAPTER - HTTP Handler
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements HTTP handlers for moderation endpoints.
package adapters

import (
	"forum/internal/modules/moderation/ports"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for moderation.
type HTTPHandler struct {
	moderationService ports.ModerationService
	templates         *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Moderation() ports.ModerationService
}

// NewHTTPHandler creates a new HTTP handler for moderation with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		moderationService: services.Moderation(),
		templates:         templates,
	}
}

// RegisterRoutes registers all moderation routes.
// TODO: Implement route registration.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Implementation placeholder
	// POST /reports - Create report
	// GET /reports - List reports (filtered by status)
	// PUT /reports/{id} - Review report
}

// CreateReportAPI handles creating a new report.
// TODO: Implement report creation handler.
func (h *HTTPHandler) CreateReportAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ListReportsAPI handles listing reports.
// TODO: Implement report listing handler.
func (h *HTTPHandler) ListReportsAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}

// ReviewReportAPI handles reviewing and updating a report.
// TODO: Implement report review handler.
func (h *HTTPHandler) ReviewReportAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
}
