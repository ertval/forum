package httpserver

import (
	"encoding/json"
	"html/template"
	"net/http"

	"forum/internal/platform/health"
)

func HealthHandler(checker *health.Checker) http.HandlerFunc {
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

// HealthUIHandler renders an HTML page with the system's health status.
func HealthUIHandler(checker *health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform checks
		results := checker.Check(r.Context())

		// Prepare data for the template
		data := map[string]interface{}{
			"Title":  "System Health",
			"Health": results,
			"User":   nil, // No user context for health page
		}

		// Parse templates
		tmpl, err := template.ParseFiles("templates/base.html", "templates/health.html")
		if err != nil {
			http.Error(w, "Could not parse templates", http.StatusInternalServerError)
			return
		}

		// Execute the template
		w.Header().Set("Content-Type", "text/html")
		err = tmpl.ExecuteTemplate(w, "base", data)
		if err != nil {
			http.Error(w, "Could not execute template", http.StatusInternalServerError)
		}
	}
}
