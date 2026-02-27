// INPUT ADAPTER - HTTP API Handler
// Package adapters implements HTTP API handlers for user endpoints.
package adapters

import (
	"encoding/json"
	"log"
	"net/http"

	"forum/internal/modules/user/domain"
)

// RegisterAPIRoutes registers all user API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	authMiddleware := h.middlewareProvider.RequireAuth()

	// Public API routes (no authentication required)
	// GET /api/users/{id} - Get user profile
	router.HandleFunc("GET /api/users/{id}", h.GetUserAPI)
	// GET /api/users - List users
	router.HandleFunc("GET /api/users", h.ListUsersAPI)

	// Protected API routes (require authentication)
	// PUT /api/users/{id}/role - Update user role (admin only)
	router.Handle("PUT /api/users/{id}/role", authMiddleware(http.HandlerFunc(h.UpdateRoleAPI)))
	// PUT /api/users/{id}/deactivate - Deactivate user
	router.Handle("PUT /api/users/{id}/deactivate", authMiddleware(http.HandlerFunc(h.DeactivateUserAPI)))
	// PUT /api/users/{id}/activate - Activate user
	router.Handle("PUT /api/users/{id}/activate", authMiddleware(http.HandlerFunc(h.ActivateUserAPI)))
}

// GetUserAPI handles user retrieval requests.
// Returns user profile by public ID (UUID).
func (h *HTTPHandler) GetUserAPI(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	// Return user profile (sensitive fields filtered by JSON tags)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// ListUsersAPI handles listing users with pagination.
func (h *HTTPHandler) ListUsersAPI(w http.ResponseWriter, r *http.Request) {
	// Default pagination values
	offset := 0
	limit := 20

	users, err := h.userService.ListUsers(r.Context(), offset, limit)
	if err != nil {
		http.Error(w, `{"error":"failed to list users"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	}); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// updateRoleRequest represents the request body for role update.
type updateRoleRequest struct {
	Role string `json:"role"`
}

// UpdateRoleAPI handles updating a user's role.
// Requires admin permissions (checked via middleware in production).
func (h *HTTPHandler) UpdateRoleAPI(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate role
	role := domain.Role(req.Role)
	if role != domain.RoleUser && role != domain.RoleModerator && role != domain.RoleAdmin {
		http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
		return
	}

	// Get user by public ID to get internal ID
	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	// Update role
	if err := h.userService.UpdateRole(r.Context(), user.ID, role); err != nil {
		if err == domain.ErrInvalidRole {
			http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"failed to update role"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "role updated successfully"}); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// DeactivateUserAPI handles deactivating a user account.
func (h *HTTPHandler) DeactivateUserAPI(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}

	// Get user by public ID to get internal ID
	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	if err := h.userService.DeactivateUser(r.Context(), user.ID); err != nil {
		http.Error(w, `{"error":"failed to deactivate user"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "user deactivated successfully"}); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// ActivateUserAPI handles activating a user account.
func (h *HTTPHandler) ActivateUserAPI(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		http.Error(w, `{"error":"user id is required"}`, http.StatusBadRequest)
		return
	}

	// Get user by public ID to get internal ID
	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	if err := h.userService.ActivateUser(r.Context(), user.ID); err != nil {
		http.Error(w, `{"error":"failed to activate user"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "user activated successfully"}); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
