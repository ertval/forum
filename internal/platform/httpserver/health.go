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
}

// HealthPage renders an HTML page with the system's health status.
// Now accepts shared templates and auth function to preserve session.
func HealthPage(cfg HealthPageConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform checks
		results := cfg.Checker.Check(r.Context())

		// Get current user if logged in
		var currentUser interface{}
		if cfg.AuthFunc != nil {
			if userID, username := cfg.AuthFunc(r); userID > 0 {
				currentUser = map[string]interface{}{
					"ID":       userID,
					"Username": username,
				}
			}
		}

		// Prepare data for the template
		data := map[string]interface{}{
			"Title":  "System Health",
			"Health": results,
			"User":   currentUser,
		}

		// Use shared templates if available, otherwise parse on demand
		var tmpl *template.Template
		var err error
		if cfg.Templates != nil {
			tmpl = cfg.Templates
		} else {
			tmpl, err = template.ParseFiles("templates/base.html", "templates/health.html")
			if err != nil {
				http.Error(w, "Could not parse templates", http.StatusInternalServerError)
				return
			}
		}

		// Execute the template
		w.Header().Set("Content-Type", "text/html")
		err = tmpl.ExecuteTemplate(w, "base", data)
		if err != nil {
			http.Error(w, "Could not execute template", http.StatusInternalServerError)
		}
	}
}
