// INPUT ADAPTER - HTTP API Handler
// [OPTIONAL FEATURE: forum-advanced-features]
// Package adapters implements HTTP API handlers for notification endpoints.
package adapters

import (
	"net/http"
)

// RegisterAPIRoutes registers all notification API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	// GET /api/notifications - Get user's notifications
	router.HandleFunc("GET /api/notifications", h.GetNotificationsAPI)
	// PUT /api/notifications/{id}/read - Mark notification as read
	router.HandleFunc("PUT /api/notifications/{id}/read", h.MarkAsReadAPI)
}

// GetNotificationsAPI handles retrieving user notifications.
// TODO: Implement notification retrieval handler.
func (h *HTTPHandler) GetNotificationsAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// MarkAsReadAPI handles marking notifications as read.
// TODO: Implement mark as read handler.
func (h *HTTPHandler) MarkAsReadAPI(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
