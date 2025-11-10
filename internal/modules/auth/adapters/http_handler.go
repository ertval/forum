// INPUT ADAPTER - HTTP Handler
// Package adapters implements the HTTP handlers for authentication endpoints.
// This adapter translates HTTP requests into service calls.
package adapters

import (
	"encoding/json"
	"fmt"
	"forum/internal/modules/auth/ports"
	"net/http"
	"time"
)

// HTTPHandler handles HTTP requests for authentication.
// It receives HTTP requests, validates input, calls the service, and returns responses.
type HTTPHandler struct {
	authService ports.AuthService
}

// NewHTTPHandler creates a new HTTP handler for authentication.
func NewHTTPHandler(authService ports.AuthService) *HTTPHandler {
	return &HTTPHandler{
		authService: authService,
	}
}

// RegisterRoutes registers all authentication routes with the router.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /auth/register", h.Register)
	router.HandleFunc("POST /auth/login", h.Login)
	router.HandleFunc("POST /auth/logout", h.Logout)
	router.HandleFunc("GET /auth/session", h.GetSession)
}

// Register handles user registration requests.
func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := h.parseJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call the service to register the user
	userID, session, err := h.authService.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		h.writeError(w, http.StatusConflict, err.Error())
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Return success response
	resp := struct {
		UserID int    `json:"user_id"`
		Email  string `json:"email"`
		Token  string `json:"token"`
	}{
		UserID: userID,
		Email:  req.Email,
		Token:  session.Token,
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// Login handles user login requests.
func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := h.parseJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call the service to login the user
	session, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Return success response
	resp := struct {
		UserID int    `json:"user_id"`
		Email  string `json:"email"`
		Token  string `json:"token"`
	}{
		UserID: session.UserID,
		Email:  req.Email,
		Token:  session.Token,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// Logout handles user logout requests.
func (h *HTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "No session token found")
		return
	}

	// Call the service to logout the user
	err = h.authService.Logout(r.Context(), cookie.Value)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete the cookie
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Return success response
	h.writeJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "Successfully logged out",
	})
}

// GetSession retrieves the current session information.
func (h *HTTPHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "No session token found")
		return
	}

	// Call the service to validate the session
	session, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid or expired session")
		return
	}

	// Return session info
	resp := struct {
		UserID    int       `json:"user_id"`
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		UserID:    session.UserID,
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// writeJSON writes a JSON response.
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error, but don't send it to the client
		fmt.Printf("Error encoding JSON response: %v\n", err)
	}
}

// writeError writes an error response.
func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		// Log the error, but don't send it to the client
		fmt.Printf("Error encoding error response: %v\n", err)
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