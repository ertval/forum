// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements HTTP handlers for comment endpoints.
// This file contains the base handler structure and shared utilities.
package adapters

import (
	"encoding/json"
	"fmt"
	"os"

	authPorts "forum/internal/modules/auth/ports"
	commentPorts "forum/internal/modules/comment/ports"
	logger "forum/internal/platform/logger"
	"html/template"
	"net/http"
)

// HTTPHandler handles HTTP requests for comments.
type HTTPHandler struct {
	commentService commentPorts.CommentService
	authService    authPorts.AuthService
	templates      *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	Comment() commentPorts.CommentService
	Auth() authPorts.AuthService
}

// NewHTTPHandler creates a new HTTP handler for comments with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		commentService: services.Comment(),
		authService:    services.Auth(),
		templates:      templates,
	}
}

// GetCurrentUser extracts user info from session cookie (helper for other handlers).
// Returns userID and username, or (0, "") if not authenticated.
func (h *HTTPHandler) GetCurrentUser(r *http.Request) (userID int, username string) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return 0, ""
	}

	session, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil || session == nil {
		return 0, ""
	}

	return session.UserID, "" // Return ID even if username fetch fails
}

// RegisterRoutes registers all comment routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page/form routes
	h.RegisterFormRoutes(router)
}

// writeJSON writes a JSON response.
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error, but don't send it to the client
		cfg := &logger.Config{
			TimePrecision: logger.TimePrecisionSeconds,
		}
		lgr := logger.NewWithConfig(logger.ErrorLevel, os.Stderr, cfg)
		lgr.Error("Failed to encode JSON response",
			logger.Error(err),
			logger.String("method", "writeJSON"))
	}
}

// parseJSON parses JSON request body.
func (h *HTTPHandler) parseJSON(r *http.Request, v interface{}) error {
	// Check if content type is JSON
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("content type is not application/json")
	}

	// Decode the JSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // This makes parsing stricter

	return decoder.Decode(v)
}
