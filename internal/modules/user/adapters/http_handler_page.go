// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for user endpoints.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	platformErrors "forum/internal/platform/errors"
	"net/http"
)

// RegisterPageRoutes registers all user page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	// Protected page routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("GET /settings", authMiddleware(http.HandlerFunc(h.SettingsPage)))
	router.Handle("POST /settings", authMiddleware(http.HandlerFunc(h.UpdateSettingsPage)))
}

// SettingsPage handles rendering the account settings page.
func (h *HTTPHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authPorts.GetUserID(ctx)
	if userPublicID == "" {
		platformErrors.RenderErrorPage(w, http.StatusUnauthorized, "", nil)
		return
	}

	currentUser, err := h.userService.GetByPublicID(ctx, userPublicID)
	if err != nil || currentUser == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "Failed to load user settings.", nil)
		return
	}

	h.renderSettingsPage(w, http.StatusOK, currentUser, "", r.URL.Query().Get("updated") == "1")
}

// UpdateSettingsPage handles settings form submissions (HTML form POST).
func (h *HTTPHandler) UpdateSettingsPage(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.RenderErrorPage(w, http.StatusUnauthorized, "", nil)
		return
	}

	updatedUser, statusCode, errMessage := h.updateCurrentUserSettings(r, userPublicID)
	if errMessage != "" {
		h.renderSettingsPage(w, statusCode, updatedUser, errMessage, false)
		return
	}

	http.Redirect(w, r, "/settings?updated=1", http.StatusSeeOther)
}
