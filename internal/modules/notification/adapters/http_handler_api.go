// INPUT ADAPTER - HTTP API Handler
// [OPTIONAL FEATURE: forum-advanced-features]
// Package adapters implements HTTP API handlers for notification endpoints.
package adapters

import (
	"encoding/json"
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/notification/domain"
	"net/http"

	platformErrors "forum/internal/platform/errors"
)

// RegisterAPIRoutes registers all notification API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	authMiddleware := h.middlewareProvider.RequireAuth()

	// Protected API routes (require authentication)
	// GET /api/notifications - Get user's notifications
	router.Handle("GET /api/notifications", authMiddleware(http.HandlerFunc(h.GetNotificationsAPI)))
	// PUT /api/notifications/{id}/read - Mark notification as read
	router.Handle("PUT /api/notifications/{id}/read", authMiddleware(http.HandlerFunc(h.MarkAsReadAPI)))
}

// GetNotificationsAPI handles retrieving user notifications.
func (h *HTTPHandler) GetNotificationsAPI(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || user == nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	notifications, err := h.notificationService.GetUserNotifications(r.Context(), user.ID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve notifications")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	}); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to encode notifications")
	}
}

// MarkAsReadAPI handles marking notifications as read.
func (h *HTTPHandler) MarkAsReadAPI(w http.ResponseWriter, r *http.Request) {
	notificationPublicID := r.PathValue("id")
	if notificationPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Notification id is required")
		return
	}

	if err := h.notificationService.MarkAsRead(r.Context(), notificationPublicID); err != nil {
		if err == domain.ErrNotificationNotFound {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Notification not found")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
