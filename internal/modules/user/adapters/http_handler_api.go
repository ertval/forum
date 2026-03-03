// INPUT ADAPTER - HTTP API Handler
// Package adapters implements HTTP API handlers for user endpoints.
package adapters

import (
	"net/http"

	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/shared/adapters/httpjson"
	"forum/internal/modules/user/domain"
	platformErrors "forum/internal/platform/errors"
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
	// POST /api/users/settings - Update current user settings
	router.Handle("POST /api/users/settings", authMiddleware(http.HandlerFunc(h.UpdateSettingsAPI)))
}

// GetUserAPI handles user retrieval requests.
// Returns user profile by public ID (UUID).
func (h *HTTPHandler) GetUserAPI(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "user id is required")
		return
	}

	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		platformErrors.WriteErrorJSON(w, http.StatusNotFound, "user not found")
		return
	}

	// Return user profile (sensitive fields filtered by JSON tags)
	httpjson.WriteJSON(w, http.StatusOK, user)
}

// ListUsersAPI handles listing users with pagination.
func (h *HTTPHandler) ListUsersAPI(w http.ResponseWriter, r *http.Request) {
	// Default pagination values
	offset := 0
	limit := 20

	users, err := h.userService.ListUsers(r.Context(), offset, limit)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// updateRoleRequest represents the request body for role update.
type updateRoleRequest struct {
	Role string `json:"role"`
}

// UpdateRoleAPI handles updating a user's role.
// Requires admin permissions (checked via middleware in production).
func (h *HTTPHandler) UpdateRoleAPI(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireAdmin(r); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "admin role required")
		return
	}

	publicID := r.PathValue("id")
	if publicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "user id is required")
		return
	}

	var req updateRoleRequest
	if err := httpjson.ParseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate role
	role := domain.Role(req.Role)
	if role != domain.RoleUser && role != domain.RoleModerator && role != domain.RoleAdmin {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "invalid role")
		return
	}

	// Get user by public ID to get internal ID
	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		platformErrors.WriteErrorJSON(w, http.StatusNotFound, "user not found")
		return
	}

	// Update role
	if err := h.userService.UpdateRole(r.Context(), user.ID, role); err != nil {
		if err == domain.ErrInvalidRole {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "invalid role")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"message": "role updated successfully"})
}

// DeactivateUserAPI handles deactivating a user account.
func (h *HTTPHandler) DeactivateUserAPI(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireAdmin(r); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "admin role required")
		return
	}

	publicID := r.PathValue("id")
	if publicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "user id is required")
		return
	}

	// Get user by public ID to get internal ID
	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		platformErrors.WriteErrorJSON(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.userService.DeactivateUser(r.Context(), user.ID); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "failed to deactivate user")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"message": "user deactivated successfully"})
}

// ActivateUserAPI handles activating a user account.
func (h *HTTPHandler) ActivateUserAPI(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireAdmin(r); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "admin role required")
		return
	}

	publicID := r.PathValue("id")
	if publicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "user id is required")
		return
	}

	// Get user by public ID to get internal ID
	user, err := h.userService.GetByPublicID(r.Context(), publicID)
	if err != nil || user == nil {
		platformErrors.WriteErrorJSON(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.userService.ActivateUser(r.Context(), user.ID); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "failed to activate user")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"message": "user activated successfully"})
}

// UpdateSettingsAPI handles settings update requests (JSON API).
func (h *HTTPHandler) UpdateSettingsAPI(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}

	updatedUser, statusCode, errMessage := h.updateCurrentUserSettings(r, userPublicID)
	if errMessage != "" {
		platformErrors.WriteErrorJSON(w, statusCode, errMessage)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "settings updated successfully",
		"user":    updatedUser,
	})
}

func (h *HTTPHandler) requireAdmin(r *http.Request) (*domain.User, error) {
	requesterPublicID := authPorts.GetUserID(r.Context())
	if requesterPublicID == "" {
		return nil, domain.ErrUserNotFound
	}

	requester, err := h.userService.GetByPublicID(r.Context(), requesterPublicID)
	if err != nil || requester == nil {
		return nil, domain.ErrUserNotFound
	}
	if requester.Role != domain.RoleAdmin {
		return nil, domain.ErrInvalidRole
	}

	return requester, nil
}
