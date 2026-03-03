// INPUT ADAPTER - HTTP API Handler
// Package adapters implements the HTTP API handlers for authentication endpoints.
// This adapter handles JSON API requests for authentication operations.
package adapters

import (
	"errors"
	"net/http"
	"time"

	authDomain "forum/internal/modules/auth/domain"
	"forum/internal/modules/shared/adapters/httpjson"
	platformErrors "forum/internal/platform/errors"
)

func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	authMiddleware := h.middlewareProvider.RequireAuth()

	// Public API routes (no authentication required)
	router.HandleFunc("POST /api/auth/register", h.RegisterAPI)
	router.HandleFunc("POST /api/auth/login", h.LoginAPI)

	// Protected API routes (require authentication)
	router.Handle("POST /api/auth/logout", authMiddleware(http.HandlerFunc(h.LogoutAPI)))
	router.Handle("GET /api/auth/session", authMiddleware(http.HandlerFunc(h.GetSessionAPI)))
}

// RegisterAPI handles user registration API requests.
func (h *HTTPHandler) RegisterAPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := httpjson.ParseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call the service to register the user
	userID, session, err := h.authService.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		// Differentiate between validation errors (400) and conflict errors (409)
		switch {
		case authDomain.IsPasswordValidationError(err):
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, authDomain.ErrInvalidEmail),
			errors.Is(err, authDomain.ErrWeakPassword),
			errors.Is(err, authDomain.ErrInvalidUsername):
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, authDomain.ErrEmailAlreadyExists),
			errors.Is(err, authDomain.ErrUsernameAlreadyExists):
			platformErrors.WriteErrorJSON(w, http.StatusConflict, err.Error())
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	// Fetch user to get PublicID
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve user information")
		return
	}

	// Return success response with public UUID (token is in HttpOnly cookie only)
	resp := struct {
		ID       string `json:"id"`
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Message  string `json:"message"`
	}{
		ID:       user.PublicID,
		UserID:   user.PublicID,
		Email:    req.Email,
		Username: req.Username,
		Message:  "Registration successful",
	}

	httpjson.WriteJSON(w, http.StatusCreated, resp)
}

// LoginAPI handles user login API requests.
func (h *HTTPHandler) LoginAPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := httpjson.ParseJSON(r, &req); err != nil {
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
		Name:     h.cookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	// Fetch user to get PublicID
	user, err := h.userService.GetByID(r.Context(), session.UserID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve user information")
		return
	}

	// Return success response with public UUID (token is in HttpOnly cookie only)
	resp := struct {
		ID       string `json:"id"`
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Message  string `json:"message"`
	}{
		ID:       user.PublicID,
		UserID:   user.PublicID,
		Email:    user.Email,
		Username: user.Username,
		Message:  "Login successful",
	}

	httpjson.WriteJSON(w, http.StatusOK, resp)
}

// LogoutAPI handles user logout requests.
func (h *HTTPHandler) LogoutAPI(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie(h.cookieName)
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
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete the cookie
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	// Return success response
	httpjson.WriteJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{
		Message: "Successfully logged out",
	})
}

// GetSessionAPI retrieves the current session information.
func (h *HTTPHandler) GetSessionAPI(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie(h.cookieName)
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

	// Return session info with public UUID (token is in HttpOnly cookie only)
	resp := struct {
		UserID    string    `json:"user_id"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		UserID:    user.PublicID,
		ExpiresAt: session.ExpiresAt,
	}

	httpjson.WriteJSON(w, http.StatusOK, resp)
}
