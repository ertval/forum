// INPUT ADAPTER - HTTP Handler
// Package adapters implements the HTTP handlers for authentication endpoints.
// This adapter translates HTTP requests into service calls.
package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	authDomain "forum/internal/modules/auth/domain"
	"forum/internal/modules/auth/ports"
	userDomain "forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
	"html/template"
	"net/http"
	"time"
)

// HTTPHandler handles HTTP requests for authentication.
// It receives HTTP requests, validates input, calls the service, and returns responses.
type HTTPHandler struct {
	authService ports.AuthService
	userService userPorts.UserService
	templates   *template.Template
}

// ServiceContainer defines the minimal interface needed by this handler.
// This allows the handler to receive all services but only use what it needs.
type ServiceContainer interface {
	Auth() ports.AuthService
	User() userPorts.UserService
}

// NewHTTPHandler creates a new HTTP handler for authentication with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
	return &HTTPHandler{
		authService: services.Auth(),
		userService: services.User(),
		templates:   templates,
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

	// Fetch username from user service
	user, err := h.userService.GetByID(r.Context(), session.UserID)
	if err != nil || user == nil {
		return session.UserID, "" // Return ID even if username fetch fails
	}

	return session.UserID, user.Username
}

// RegisterRoutes registers all authentication routes with the router.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// API routes (return JSON)
	router.HandleFunc("POST /auth/register", h.RegisterAPI)
	router.HandleFunc("POST /auth/login", h.LoginAPI)
	router.HandleFunc("POST /auth/logout", h.LogoutAPI)
	router.HandleFunc("GET /auth/session", h.GetSessionAPI)

	// Page routes (render HTML or redirect)
	router.HandleFunc("GET /login", h.LoginPage)
	router.HandleFunc("GET /register", h.RegisterPage)
	router.HandleFunc("GET /logout", h.LogoutPage)
}

// RegisterAPI handles user registration API requests.
func (h *HTTPHandler) RegisterAPI(w http.ResponseWriter, r *http.Request) {
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
		// Differentiate between validation errors (400) and conflict errors (409)
		switch {
		case errors.Is(err, authDomain.ErrInvalidEmail),
			errors.Is(err, authDomain.ErrWeakPassword),
			errors.Is(err, authDomain.ErrInvalidUsername):
			h.writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, authDomain.ErrUserAlreadyExists),
			errors.Is(err, userDomain.ErrUserNotFound):
			h.writeError(w, http.StatusConflict, err.Error())
		default:
			// Check if error message contains validation keywords
			errMsg := err.Error()
			if strings.Contains(errMsg, "empty") || strings.Contains(errMsg, "invalid") ||
				strings.Contains(errMsg, "required") || strings.Contains(errMsg, "format") ||
				strings.Contains(errMsg, "too long") || strings.Contains(errMsg, "too short") {
				h.writeError(w, http.StatusBadRequest, errMsg)
			} else if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "taken") {
				h.writeError(w, http.StatusConflict, errMsg)
			} else {
				h.writeError(w, http.StatusConflict, errMsg)
			}
		}
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

	// Return success response with all required fields
	userIDStr := strconv.Itoa(userID)
	resp := struct {
		ID       string `json:"id"`
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Token    string `json:"token"`
	}{
		ID:       userIDStr,
		UserID:   userIDStr,
		Email:    req.Email,
		Username: req.Username,
		Token:    session.Token,
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// LoginAPI handles user login API requests.
func (h *HTTPHandler) LoginAPI(w http.ResponseWriter, r *http.Request) {
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

// LogoutAPI handles user logout requests.
func (h *HTTPHandler) LogoutAPI(w http.ResponseWriter, r *http.Request) {
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

// GetSessionAPI retrieves the current session information.
func (h *HTTPHandler) GetSessionAPI(w http.ResponseWriter, r *http.Request) {
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

// LoginPage renders the login page.
func (h *HTTPHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Execute the login template directly
	if err := h.templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render login page: %v", err), http.StatusInternalServerError)
		return
	}
}

// RegisterPage renders the registration page.
func (h *HTTPHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	// Execute the register template directly
	if err := h.templates.ExecuteTemplate(w, "register.html", nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render register page: %v", err), http.StatusInternalServerError)
		return
	}
}

// LogoutPage handles the frontend logout by invalidating the session and redirecting.
func (h *HTTPHandler) LogoutPage(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		// Call the service to logout the user (invalidate the session)
		_ = h.authService.Logout(r.Context(), cookie.Value) // We ignore the error for frontend UX
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

	// Redirect to home page after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
