// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements the HTTP page handlers for authentication endpoints.
// This adapter handles HTML page requests for authentication operations.
package adapters

import (
	"net/http"

	platformErrors "forum/internal/platform/errors"
	platformTemplates "forum/internal/platform/templates"
)

// RegisterPageRoutes registers all authentication page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /login", h.LoginPage)
	router.HandleFunc("GET /register", h.RegisterPage)
	router.HandleFunc("GET /logout", h.LogoutPage)
}

// LoginPage renders the login page.
func (h *HTTPHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Login",
	}

	tmpl, err := platformTemplates.Get("login", "templates/base.html", "templates/login.html")
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", nil)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", nil)
		return
	}
}

// RegisterPage renders the registration page.
func (h *HTTPHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Register",
	}

	tmpl, err := platformTemplates.Get("register", "templates/base.html", "templates/register.html")
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", nil)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", nil)
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
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to home page after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
