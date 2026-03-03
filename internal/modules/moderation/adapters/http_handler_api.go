// INPUT ADAPTER - HTTP API Handler
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements HTTP API handlers for moderation endpoints.
package adapters

import (
	"errors"
	"net/http"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/moderation/domain"
	"forum/internal/modules/shared/adapters/httpjson"
	userDomain "forum/internal/modules/user/domain"
	platformErrors "forum/internal/platform/errors"
)

// RegisterAPIRoutes registers all moderation API routes with the router.
func (h *HTTPHandler) RegisterAPIRoutes(router *http.ServeMux) {
	authMiddleware := h.middlewareProvider.RequireAuth()

	// Protected API routes (require authentication)
	// POST /api/moderation/reports - Create report
	router.Handle("POST /api/moderation/reports", authMiddleware(http.HandlerFunc(h.CreateReportAPI)))
	// GET /api/moderation/reports - List reports (filtered by status)
	router.Handle("GET /api/moderation/reports", authMiddleware(http.HandlerFunc(h.ListReportsAPI)))
	// PUT /api/moderation/reports/{id} - Review report
	router.Handle("PUT /api/moderation/reports/{id}", authMiddleware(http.HandlerFunc(h.ReviewReportAPI)))

	// POST /api/moderation/requests - Request moderator role (user role only)
	router.Handle("POST /api/moderation/requests", authMiddleware(http.HandlerFunc(h.RequestModeratorRoleAPI)))
	// GET /api/moderation/requests - List moderator requests (admin only)
	router.Handle("GET /api/moderation/requests", authMiddleware(http.HandlerFunc(h.ListModeratorRequestsAPI)))
	// PUT /api/moderation/requests/{id} - Review moderator request (admin only)
	router.Handle("PUT /api/moderation/requests/{id}", authMiddleware(http.HandlerFunc(h.ReviewModeratorRequestAPI)))
	// Note: The route parameter is already aligned with the health checker expected {id} placeholder.
}

// CreateReportAPI handles creating a new report.
func (h *HTTPHandler) CreateReportAPI(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	userID, err := h.getInternalUserID(r.Context(), userPublicID)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}

	var req struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Reason     string `json:"reason"`
	}
	if err := httpjson.ParseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	report, err := h.moderationService.CreateReport(
		r.Context(),
		userID,
		req.TargetID,
		req.TargetType,
		req.Reason,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidTargetType), errors.Is(err, domain.ErrInvalidReason), errors.Is(err, domain.ErrInvalidTarget):
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrInsufficientPermissions):
			platformErrors.WriteErrorJSON(w, http.StatusForbidden, err.Error())
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create report")
		}
		return
	}

	report.PublicReporterID = userPublicID
	if report.PublicTargetID == "" {
		report.PublicTargetID = req.TargetID
	}

	httpjson.WriteJSON(w, http.StatusCreated, report)
}

// ListReportsAPI handles listing reports.
func (h *HTTPHandler) ListReportsAPI(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireModerator(r); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, err.Error())
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	reports, err := h.moderationService.ListReports(r.Context(), status)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidReportStatus) {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to list reports")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]any{
		"reports": reports,
		"count":   len(reports),
	})
}

// ReviewReportAPI handles reviewing and updating a report.
func (h *HTTPHandler) ReviewReportAPI(w http.ResponseWriter, r *http.Request) {
	moderator, err := h.requireModerator(r)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, err.Error())
		return
	}

	reportPublicID := r.PathValue("id")
	if strings.TrimSpace(reportPublicID) == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Report ID is required")
		return
	}

	var req struct {
		Status   string `json:"status"`
		Decision string `json:"decision"`
		Response string `json:"response"`
	}
	if err := httpjson.ParseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	status := req.Status
	if strings.TrimSpace(status) == "" {
		status = req.Decision
	}

	report, err := h.moderationService.ReviewReport(r.Context(), moderator.ID, reportPublicID, status, req.Response)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrReportNotFound):
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrInvalidReportStatus):
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrInsufficientPermissions):
			platformErrors.WriteErrorJSON(w, http.StatusForbidden, err.Error())
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to review report")
		}
		return
	}

	if report.PublicModeratorID == "" {
		report.PublicModeratorID = moderator.PublicID
	}

	httpjson.WriteJSON(w, http.StatusOK, report)
}

func (h *HTTPHandler) requireModerator(r *http.Request) (*userView, error) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		return nil, domain.ErrInsufficientPermissions
	}

	user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || user == nil {
		return nil, domain.ErrInsufficientPermissions
	}
	if !user.CanModerate() {
		return nil, domain.ErrInsufficientPermissions
	}

	return &userView{ID: user.ID, PublicID: user.PublicID}, nil
}

func (h *HTTPHandler) requireAdmin(r *http.Request) (*userDomain.User, error) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		return nil, domain.ErrInsufficientPermissions
	}

	user, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || user == nil {
		return nil, domain.ErrInsufficientPermissions
	}
	if !user.IsAdmin() {
		return nil, domain.ErrInsufficientPermissions
	}

	return user, nil
}

// RequestModeratorRoleAPI creates a moderator role request for the current user.
func (h *HTTPHandler) RequestModeratorRoleAPI(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	requester, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || requester == nil {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Invalid user")
		return
	}
	if requester.Role != userDomain.RoleUser {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "only regular users can request moderator role")
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := httpjson.ParseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	request, err := h.moderationService.RequestModeratorRole(r.Context(), requester.ID, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrModeratorRequestAlreadyPending):
			platformErrors.WriteErrorJSON(w, http.StatusConflict, err.Error())
		case errors.Is(err, domain.ErrInvalidRequester):
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create moderator request")
		}
		return
	}

	request.PublicRequesterID = requester.PublicID
	httpjson.WriteJSON(w, http.StatusCreated, request)
}

// ListModeratorRequestsAPI lists moderator role requests (admin only).
func (h *HTTPHandler) ListModeratorRequestsAPI(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireAdmin(r); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, err.Error())
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	requests, err := h.moderationService.ListModeratorRequests(r.Context(), status)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRequestStatus) {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to list moderator requests")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]any{
		"requests": requests,
		"count":    len(requests),
	})
}

// ReviewModeratorRequestAPI approves or denies a moderator role request (admin only).
func (h *HTTPHandler) ReviewModeratorRequestAPI(w http.ResponseWriter, r *http.Request) {
	admin, err := h.requireAdmin(r)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, err.Error())
		return
	}

	requestPublicID := strings.TrimSpace(r.PathValue("id"))
	if requestPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "request id is required")
		return
	}

	var req struct {
		Status   string `json:"status"`
		Decision string `json:"decision"`
		Response string `json:"response"`
	}
	if err := httpjson.ParseJSON(r, &req); err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	decision := strings.TrimSpace(req.Status)
	if decision == "" {
		decision = strings.TrimSpace(req.Decision)
	}
	decision = domain.NormalizeRequestStatus(decision)

	moderatorRequest, err := h.moderationService.GetModeratorRequestByPublicID(r.Context(), requestPublicID)
	if err != nil {
		if errors.Is(err, domain.ErrModeratorRequestNotFound) {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to load moderator request")
		return
	}

	if decision == domain.RequestStatusApproved {
		requester, getErr := h.userService.GetByID(r.Context(), moderatorRequest.RequesterID)
		if getErr != nil || requester == nil {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "request user not found")
			return
		}
		if !requester.CanModerate() {
			if updateErr := h.userService.UpdateRole(r.Context(), requester.ID, userDomain.RoleModerator); updateErr != nil {
				platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "failed to promote user to moderator")
				return
			}
		}
	}

	updatedRequest, err := h.moderationService.ReviewModeratorRequest(r.Context(), admin.ID, requestPublicID, decision, req.Response)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrModeratorRequestNotFound):
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrInvalidRequestStatus):
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrInsufficientPermissions):
			platformErrors.WriteErrorJSON(w, http.StatusForbidden, err.Error())
		default:
			platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to review moderator request")
		}
		return
	}

	updatedRequest.PublicReviewerID = admin.PublicID
	httpjson.WriteJSON(w, http.StatusOK, updatedRequest)
}

type userView struct {
	ID       int
	PublicID string
}
