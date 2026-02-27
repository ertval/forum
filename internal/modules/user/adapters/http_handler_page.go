// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for user endpoints.
package adapters

import (
	"net/http"
)

// RegisterPageRoutes registers all user page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	// TODO: Add user profile page routes when implemented
	// router.HandleFunc("GET /users/{id}", h.UserProfilePage)
}
