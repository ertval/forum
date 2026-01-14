package httpserver

import (
	"encoding/json"
	"html/template"
	"net/http"

	"forum/internal/platform/health"
)

func HealthAPI(checker *health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform checks
		results := checker.Check(r.Context())

		// Determine overall status
		isHealthy := true
		for _, status := range results {
			if status != "up" {
				isHealthy = false
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !isHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(results)
	}
}

// HealthPageConfig holds dependencies for the health UI page handler.
type HealthPageConfig struct {
	Checker   *health.Checker
	Templates *template.Template
	AuthFunc  func(r *http.Request) (userID int, username string)
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
			"Title":  "Health Status",
			"Health": results,
			"User":   currentUser,
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
