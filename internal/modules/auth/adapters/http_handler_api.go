// INPUT ADAPTER - HTTP API Handler
// Package adapters implements the HTTP API handlers for authentication endpoints.
// This adapter handles JSON API requests for authentication operations.
package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	authDomain "forum/internal/modules/auth/domain"
	platformErrors "forum/internal/platform/errors"
)

// RegisterAPIRoutes registers all authentication API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /api/auth/register", h.RegisterAPI)
	router.HandleFunc("POST /api/auth/login", h.LoginAPI)
	router.HandleFunc("POST /api/auth/logout", h.LogoutAPI)
	router.HandleFunc("GET /api/auth/session", h.GetSessionAPI)
}

// RegisterAPI handles user registration API requests.
func (h *HTTPHandler) RegisterAPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := h.parseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
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
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, authDomain.ErrEmailAlreadyExists),
			errors.Is(err, authDomain.ErrUsernameAlreadyExists):
			platformErrors.WriteErrorJSON(w, http.StatusConflict, err.Error())
		default:
			// Check if error message contains validation keywords
			errMsg := err.Error()
			if strings.Contains(errMsg, "empty") || strings.Contains(errMsg, "invalid") ||
				strings.Contains(errMsg, "required") || strings.Contains(errMsg, "format") ||
				strings.Contains(errMsg, "too long") || strings.Contains(errMsg, "too short") {
				platformErrors.WriteErrorJSON(w, http.StatusBadRequest, errMsg)
			} else if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "taken") {
				platformErrors.WriteErrorJSON(w, http.StatusConflict, errMsg)
			} else {
				platformErrors.WriteErrorJSON(w, http.StatusConflict, errMsg)
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

	// Fetch user to get PublicID
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve user information")
		return
	}

	// Return success response with public UUID
	resp := struct {
		ID       string `json:"id"`
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Token    string `json:"token"`
	}{
		ID:       user.PublicID,
		UserID:   user.PublicID,
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
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call the service to login the user
	session, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid email or password")
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

	// Fetch user to get PublicID
	user, err := h.userService.GetByID(r.Context(), session.UserID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve user information")
		return
	}

	// Return success response with public UUID
	resp := struct {
		ID       string `json:"id"`
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Token    string `json:"token"`
	}{
		ID:       user.PublicID,
		UserID:   user.PublicID,
		Email:    user.Email,
		Username: user.Username,
		Token:    session.Token,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// LogoutAPI handles user logout requests.
func (h *HTTPHandler) LogoutAPI(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "No session token found")
		return
	}

	// Call the service to logout the user
	err = h.authService.Logout(r.Context(), cookie.Value)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to logout")
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
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "No session token found")
		return
	}

	// Call the service to validate the session
	session, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid or expired session")
		return
	}

	// Fetch user to get PublicID
	user, err := h.userService.GetByID(r.Context(), session.UserID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve user information")
		return
	}

	// Return session info with public UUID
	resp := struct {
		UserID    string    `json:"user_id"`
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		UserID:    user.PublicID,
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
