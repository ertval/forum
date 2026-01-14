// INPUT ADAPTER - HTTP API Handler
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements HTTP API handlers for moderation endpoints.
package adapters

import (
	"net/http"
)

// RegisterAPIRoutes registers all moderation API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	authMiddleware := h.middlewareProvider.RequireAuth()

	// Protected API routes (require authentication)
	// POST /api/moderation/reports - Create report
	router.Handle("POST /api/moderation/reports", authMiddleware(http.HandlerFunc(h.CreateReportAPI)))
	// GET /api/moderation/reports - List reports (filtered by status)
	router.Handle("GET /api/moderation/reports", authMiddleware(http.HandlerFunc(h.ListReportsAPI)))
	// PUT /api/moderation/reports/{id} - Review report
	router.Handle("PUT /api/moderation/reports/{id}", authMiddleware(http.HandlerFunc(h.ReviewReportAPI)))
}

// CreateReportAPI handles creating a new report.
// TODO: Implement report creation handler.
func (h *HTTPHandler) CreateReportAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// ListReportsAPI handles listing reports.
// TODO: Implement report listing handler.
func (h *HTTPHandler) ListReportsAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// ReviewReportAPI handles reviewing and updating a report.
// TODO: Implement report review handler.
func (h *HTTPHandler) ReviewReportAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
