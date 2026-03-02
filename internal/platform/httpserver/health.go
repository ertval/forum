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

type healthTableRow struct {
	Label  string
	Status string
}

type healthTableSection struct {
	Title   string
	Rows    []healthTableRow
	ColName string
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
	Templates *template.Template
	AuthFunc func(r *http.Request) (userID int, username string)
	// GetUserWithStats returns full user data including stats for template rendering.
	// If nil, only basic auth info (ID, Username) will be shown.
	GetUserWithStats func(r *http.Request) map[string]interface{}
}

// HealthPage renders an HTML page with the system's health status.
// Now accepts shared templates and auth function to preserve session.
func HealthPage(cfg HealthPageConfig) http.HandlerFunc {
	if cfg.Templates == nil {
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
		moduleRows := []healthTableRow{
			{Label: "auth_api", Status: results["auth_api"]},
			{Label: "post_api", Status: results["post_api"]},
			{Label: "user_api", Status: results["user_api"]},
			{Label: "comment_api", Status: results["comment_api"]},
			{Label: "reaction_api", Status: results["reaction_api"]},
			{Label: "moderation_api", Status: results["moderation_api"]},
			{Label: "notification_api", Status: results["notification_api"]},
		}

		healthSections := []healthTableSection{
			{
				Title:   "Core Services",
				ColName: "Service",
				Rows: []healthTableRow{
					{Label: "Database", Status: results["database"]},
				},
			},
			{
				Title:   "Module API Status",
				ColName: "Module",
				Rows:    moduleRows,
			},
		}

		data := map[string]interface{}{
			"Title":           "Health Status",
			"HideUserSidebar": true,
			"Health":          results,
			"HealthSections":  healthSections,
			"User":            currentUser,
			// Health page should not show the sidebar even when user is present
			"ShowSidebar": false,
		}

		// Execute the shared pre-parsed template set
		w.Header().Set("Content-Type", "text/html")
		if err := cfg.Templates.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Could not execute template", http.StatusInternalServerError)
		}
	}
}
