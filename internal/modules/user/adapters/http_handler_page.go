// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for user endpoints.
package adapters

import (
	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/platform/templates"
	"net/http"
)

// RegisterPageRoutes registers all user page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	// Protected page routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("GET /settings", authMiddleware(http.HandlerFunc(h.SettingsPage)))
}

// SettingsPage handles rendering the account settings page.
func (h *HTTPHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authPorts.GetUserID(ctx)
	if userPublicID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currentUser, err := h.userService.GetByPublicID(ctx, userPublicID)
	if err != nil || currentUser == nil {
		http.Error(w, "Failed to load user settings", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":       "Settings",
		"User":        currentUser,
		"ShowFilter":  false,
		"ShowSidebar": false,
	}

	tmpl, err := templates.Get("settings", "templates/base.html", "templates/settings.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
