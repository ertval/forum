package httpserver

import (
	"encoding/json"
	"net/http"

	platformErrors "forum/internal/platform/errors"
	"forum/internal/platform/health"
	"forum/internal/platform/templates"
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

// HealthHandler handles health check HTTP requests.
// It follows the same RegisterRoutes pattern as module handlers.
type HealthHandler struct {
	checker          *health.Checker
	templates        *templates.Registry
	authFunc         func(r *http.Request) (publicID string, username string)
	getUserWithStats func(r *http.Request) map[string]interface{}
}

// NewHealthHandler creates a new health handler with the given dependencies.
func NewHealthHandler(
	checker *health.Checker,
	tmpl *templates.Registry,
	authFunc func(r *http.Request) (string, string),
	getUserWithStats func(r *http.Request) map[string]interface{},
) *HealthHandler {
	return &HealthHandler{
		checker:          checker,
		templates:        tmpl,
		authFunc:         authFunc,
		getUserWithStats: getUserWithStats,
	}
}

// RegisterRoutes registers all health check routes on the given router.
func (h *HealthHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /health", h.HealthPage)
	router.HandleFunc("GET /health-api", h.HealthAPI)
	router.HandleFunc("GET /health/errors/400", h.HealthError400Page)
	router.HandleFunc("GET /health/errors/404", h.HealthError404Page)
	router.HandleFunc("GET /health/errors/500", h.HealthError500Page)
}

// HealthAPI returns health check results as JSON.
func (h *HealthHandler) HealthAPI(w http.ResponseWriter, r *http.Request) {
	results := h.checker.Check(r.Context())

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

// HealthPage renders an HTML page with the system's health status.
func (h *HealthHandler) HealthPage(w http.ResponseWriter, r *http.Request) {
	if h.templates == nil || h.templates.Lookup("health") == nil {
		http.Error(w, "Health templates not available", http.StatusInternalServerError)
		return
	}

	results := h.checker.Check(r.Context())

	// Get current user if logged in
	var currentUser interface{}
	if h.getUserWithStats != nil {
		currentUser = h.getUserWithStats(r)
	} else if h.authFunc != nil {
		if publicID, username := h.authFunc(r); publicID != "" {
			currentUser = map[string]interface{}{
				"PublicID": publicID,
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
		"ShowSidebar":     false,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "health", "base", data); err != nil {
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
	}
}

// HealthError400Page renders a styled 400 error page for manual testing.
func (h *HealthHandler) HealthError400Page(w http.ResponseWriter, r *http.Request) {
	platformErrors.RenderErrorPage(w, http.StatusBadRequest, "Health test route: simulated bad request.", nil)
}

// HealthError404Page renders a styled 404 error page for manual testing.
func (h *HealthHandler) HealthError404Page(w http.ResponseWriter, r *http.Request) {
	platformErrors.RenderErrorPage(w, http.StatusNotFound, "Health test route: simulated missing page.", nil)
}

// HealthError500Page renders a styled 500 error page for manual testing.
func (h *HealthHandler) HealthError500Page(w http.ResponseWriter, r *http.Request) {
	platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "Health test route: simulated internal server error.", nil)
}
