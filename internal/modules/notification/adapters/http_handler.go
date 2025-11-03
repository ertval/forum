// INPUT ADAPTER - HTTP Handler
// [OPTIONAL FEATURE: forum-advanced-features]
// Package adapters implements HTTP handlers for notification endpoints.
package adapters

import (
	"forum/internal/modules/notification/ports"
	"net/http"
)

// HTTPHandler handles HTTP requests for notifications.
type HTTPHandler struct {
	notificationService ports.NotificationService
}

// NewHTTPHandler creates a new HTTP handler for notifications.
func NewHTTPHandler(notificationService ports.NotificationService) *HTTPHandler {
	return &HTTPHandler{notificationService: notificationService}
}

// RegisterRoutes registers all notification routes.
// TODO: Implement route registration.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Implementation placeholder
	// GET /notifications - Get user's notifications
	// PUT /notifications/{id}/read - Mark notification as read
}

// GetNotifications handles retrieving user notifications.
// TODO: Implement notification retrieval handler.
func (h *HTTPHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	// 1. Get userID from session
	// 2. Call notificationService.GetUserNotifications
	// 3. Return notifications list
}

// MarkAsRead handles marking a notification as read.
// TODO: Implement mark as read handler.
func (h *HTTPHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	// 1. Parse notification ID from URL
	// 2. Call notificationService.MarkAsRead
	// 3. Return 200 OK
}
