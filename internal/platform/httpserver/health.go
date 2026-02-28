package httpserver

import (
	"encoding/json"
	"html/template"
	"net/http"

	"forum/internal/platform/health"
)

// criticalChecks lists health check keys that must be "up" for the
// readiness endpoint to return 200. Everything else is optional and
// will be surfaced in the body but won't cause a 503.
var criticalChecks = map[string]bool{
	"database":         true,
	"auth_api":         true,
	"post_api":         true,
	"comment_api":      true,
	"reaction_api":     true,
	"notification_api": true,
	"user_api":         true,
}

func HealthAPI(checker *health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform checks
		results := checker.Check(r.Context())

		// Determine overall status — only critical checks affect readiness
		criticalHealthy := true
		for key, status := range results {
			if criticalChecks[key] && status != "up" {
				criticalHealthy = false
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !criticalHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(results)
	}
}

// HealthPageConfig holds dependencies for the health UI page handler.
type HealthPageConfig struct {
	Checker  *health.Checker
	AuthFunc func(r *http.Request) (userID int, username string)
	// GetUserWithStats returns full user data including stats for template rendering.
	// If nil, only basic auth info (ID, Username) will be shown.
	GetUserWithStats func(r *http.Request) map[string]interface{}
}

// HealthPage renders an HTML page with the system's health status.
// Now accepts shared templates and auth function to preserve session.
func HealthPage(cfg HealthPageConfig) http.HandlerFunc {
	// Parse templates ONCE at handler creation time (not on every request)
	healthTemplate, parseErr := template.ParseFiles("templates/base.html", "templates/health.html")
	if parseErr != nil {
		// If templates can't be parsed at startup, return a handler that reports the error
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Health templates not available", http.StatusInternalServerError)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Perform checks
		results := cfg.Checker.Check(r.Context())

		// Get current user if logged in
		var currentUser interface{}
		if cfg.GetUserWithStats != nil {
			// Use full user data with stats if available
			currentUser = cfg.GetUserWithStats(r)
		} else if cfg.AuthFunc != nil {
			// Fall back to basic auth info
			if userID, username := cfg.AuthFunc(r); userID > 0 {
				currentUser = map[string]interface{}{
					"ID":       userID,
					"Username": username,
				}
			}
		}

		// Prepare data for the template
		data := map[string]interface{}{
			"Title":           "Health Status",
			"HideUserSidebar": true,
			"Health":          results,
			"User":            currentUser,
			// Health page should not show the sidebar even when user is present
			"ShowSidebar": false,
		}

		// Execute the pre-parsed template
		w.Header().Set("Content-Type", "text/html")
		if err := healthTemplate.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Could not execute template", http.StatusInternalServerError)
		}
	}
}
