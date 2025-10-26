package server

// router.go sets up all HTTP routes and middleware for the forum application.
// It connects handlers with their corresponding URL patterns.

import (
	"net/http"

	"forum/internal/handlers"
	"forum/internal/middleware"
)

// setupRouter creates and configures the application router with all routes and middleware
func setupRouter() http.Handler {
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public routes
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/register", handleRegister)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/post/", handlers.PostHandler) // View single post

	// Protected routes (require authentication)
	mux.HandleFunc("/post/create", handlers.CreatePostHandler)
	mux.HandleFunc("/comment/create", handlers.CreateCommentHandler)
	mux.HandleFunc("/reaction", handlers.ReactionHandler)

	// Filter routes
	mux.HandleFunc("/filter", handlers.FilterHandler)

	// Apply middleware chain
	// Order matters: session -> error handling -> routes
	handler := middleware.SessionMiddleware(mux)
	handler = middleware.ErrorHandler(handler)

	return handler
}

// handleRegister routes GET and POST requests for registration
func handleRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.RegisterHandlerGET(w, r)
	case http.MethodPost:
		handlers.RegisterHandlerPOST(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleLogin routes GET and POST requests for login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.LoginHandlerGET(w, r)
	case http.MethodPost:
		handlers.LoginHandlerPOST(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
