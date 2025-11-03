package http
// Package http provides HTTP handlers for authentication endpoints.
// This is the HTTP adapter for the auth module's inbound port.
package http

import (
	"net/http"

	"forum/internal/modules/auth/ports/input"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	authService input.AuthService
}

// NewHandler creates a new HTTP handler for authentication.
func NewHandler(authService input.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

// RegisterRoutes registers the authentication routes with the router.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// TODO: Register routes
	mux.HandleFunc("/register", h.HandleRegister)
	mux.HandleFunc("/login", h.HandleLogin)
	mux.HandleFunc("/logout", h.HandleLogout)
	mux.HandleFunc("/oauth/callback", h.HandleOAuthCallback)
}

// HandleRegister handles user registration requests.
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement registration handler
}

// HandleLogin handles user login requests.
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement login handler
}

// HandleLogout handles user logout requests.
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logout handler
}

// HandleOAuthCallback handles OAuth callback requests.
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement OAuth callback handler
}
