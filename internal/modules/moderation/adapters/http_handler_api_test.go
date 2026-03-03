package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"forum/internal/modules/auth/ports"
	"forum/internal/modules/moderation/domain"
	userDomain "forum/internal/modules/user/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockModerationService struct {
	createFn                 func(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) (*domain.Report, error)
	listFn                   func(ctx context.Context, status string) ([]*domain.Report, error)
	reviewFn                 func(ctx context.Context, moderatorID int, reportPublicID string, status, response string) (*domain.Report, error)
	requestModeratorFn       func(ctx context.Context, requesterID int, message string) (*domain.ModeratorRequest, error)
	listModeratorRequestsFn  func(ctx context.Context, status string) ([]*domain.ModeratorRequest, error)
	reviewModeratorRequestFn func(ctx context.Context, reviewerID int, requestPublicID string, status, response string) (*domain.ModeratorRequest, error)
	getModeratorRequestFn    func(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error)
}

func (m *mockModerationService) CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) (*domain.Report, error) {
	if m.createFn != nil {
		return m.createFn(ctx, reporterID, targetPublicID, targetType, reason)
	}
	return nil, nil
}

func (m *mockModerationService) ListReports(ctx context.Context, status string) ([]*domain.Report, error) {
	if m.listFn != nil {
		return m.listFn(ctx, status)
	}
	return nil, nil
}

func (m *mockModerationService) ReviewReport(ctx context.Context, moderatorID int, reportPublicID string, status, response string) (*domain.Report, error) {
	if m.reviewFn != nil {
		return m.reviewFn(ctx, moderatorID, reportPublicID, status, response)
	}
	return nil, nil
}

func (m *mockModerationService) RequestModeratorRole(ctx context.Context, requesterID int, message string) (*domain.ModeratorRequest, error) {
	if m.requestModeratorFn != nil {
		return m.requestModeratorFn(ctx, requesterID, message)
	}
	return nil, nil
}

func (m *mockModerationService) ListModeratorRequests(ctx context.Context, status string) ([]*domain.ModeratorRequest, error) {
	if m.listModeratorRequestsFn != nil {
		return m.listModeratorRequestsFn(ctx, status)
	}
	return nil, nil
}

func (m *mockModerationService) ReviewModeratorRequest(ctx context.Context, reviewerID int, requestPublicID string, status, response string) (*domain.ModeratorRequest, error) {
	if m.reviewModeratorRequestFn != nil {
		return m.reviewModeratorRequestFn(ctx, reviewerID, requestPublicID, status, response)
	}
	return nil, nil
}

func (m *mockModerationService) GetModeratorRequestByPublicID(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error) {
	if m.getModeratorRequestFn != nil {
		return m.getModeratorRequestFn(ctx, requestPublicID)
	}
	return nil, nil
}

type mockUserLookupService struct {
	user        *userDomain.User
	userByID    *userDomain.User
	err         error
	updateErr   error
	updatedRole userDomain.Role
}

func (m *mockUserLookupService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil {
		return nil, errors.New("user not found")
	}
	return m.user, nil
}

func (m *mockUserLookupService) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.userByID == nil {
		return nil, errors.New("user not found")
	}
	return m.userByID, nil
}

func (m *mockUserLookupService) UpdateRole(ctx context.Context, userID int, newRole userDomain.Role) error {
	m.updatedRole = newRole
	return m.updateErr
}

func withUserID(r *http.Request, userPublicID string) *http.Request {
	ctx := context.WithValue(r.Context(), ports.UserIDKey, userPublicID)
	return r.WithContext(ctx)
}

func TestCreateReportAPI_Success(t *testing.T) {
	h := &HTTPHandler{
		moderationService: &mockModerationService{createFn: func(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) (*domain.Report, error) {
			return &domain.Report{
				PublicID:       "report-public-id",
				TargetType:     targetType,
				PublicTargetID: targetPublicID,
				Reason:         reason,
				Status:         domain.StatusPending,
				CreatedAt:      time.Now(),
			}, nil
		}},
		userService: &mockUserLookupService{user: &userDomain.User{ID: 11, PublicID: "user-public-id", Role: userDomain.RoleUser}},
	}

	body := bytes.NewBufferString(`{"target_type":"post","target_id":"post-public-id","reason":"spam"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/moderation/reports", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-public-id")

	w := httptest.NewRecorder()
	h.CreateReportAPI(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp domain.Report
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.PublicID != "report-public-id" {
		t.Fatalf("id = %q, want report-public-id", resp.PublicID)
	}
	if resp.PublicReporterID != "user-public-id" {
		t.Fatalf("reporter_id = %q, want user-public-id", resp.PublicReporterID)
	}
}

func TestListReportsAPI_ForbiddenForNonModerator(t *testing.T) {
	h := &HTTPHandler{
		moderationService: &mockModerationService{},
		userService:       &mockUserLookupService{user: &userDomain.User{ID: 12, PublicID: "user-public-id", Role: userDomain.RoleUser}},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/moderation/reports", nil)
	req = withUserID(req, "user-public-id")
	w := httptest.NewRecorder()

	h.ListReportsAPI(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestReviewReportAPI_Success(t *testing.T) {
	h := &HTTPHandler{
		moderationService: &mockModerationService{reviewFn: func(ctx context.Context, moderatorID int, reportPublicID string, status, response string) (*domain.Report, error) {
			return &domain.Report{
				PublicID:   reportPublicID,
				Status:     status,
				Response:   response,
				CreatedAt:  time.Now(),
				ReviewedAt: ptrTime(time.Now()),
			}, nil
		}},
		userService: &mockUserLookupService{user: &userDomain.User{ID: 99, PublicID: "mod-public-id", Role: userDomain.RoleModerator}},
	}

	body := bytes.NewBufferString(`{"decision":"resolved","response":"done"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/moderation/reports/report-id", body)
	req.SetPathValue("id", "report-id")
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "mod-public-id")
	w := httptest.NewRecorder()

	h.ReviewReportAPI(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp domain.Report
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Status != domain.StatusResolved {
		t.Fatalf("status = %q, want %q", resp.Status, domain.StatusResolved)
	}
	if resp.PublicModeratorID != "mod-public-id" {
		t.Fatalf("moderator_id = %q, want mod-public-id", resp.PublicModeratorID)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestRequestModeratorRoleAPI_UserSuccess(t *testing.T) {
	h := &HTTPHandler{
		moderationService: &mockModerationService{requestModeratorFn: func(ctx context.Context, requesterID int, message string) (*domain.ModeratorRequest, error) {
			return &domain.ModeratorRequest{PublicID: "req-public-id", Status: domain.RequestStatusPending, Message: message, CreatedAt: time.Now()}, nil
		}},
		userService: &mockUserLookupService{user: &userDomain.User{ID: 11, PublicID: "user-public-id", Role: userDomain.RoleUser}},
	}

	body := bytes.NewBufferString(`{"message":"I can help moderate spam"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/moderation/requests", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-public-id")

	w := httptest.NewRecorder()
	h.RequestModeratorRoleAPI(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestRequestModeratorRoleAPI_ForModeratorForbidden(t *testing.T) {
	h := &HTTPHandler{
		moderationService: &mockModerationService{},
		userService:       &mockUserLookupService{user: &userDomain.User{ID: 11, PublicID: "mod-public-id", Role: userDomain.RoleModerator}},
	}

	body := bytes.NewBufferString(`{"message":"already mod"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/moderation/requests", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "mod-public-id")

	w := httptest.NewRecorder()
	h.RequestModeratorRoleAPI(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestReviewModeratorRequestAPI_ApprovePromotesUser(t *testing.T) {
	userLookup := &mockUserLookupService{
		user:     &userDomain.User{ID: 1, PublicID: "admin-public-id", Role: userDomain.RoleAdmin},
		userByID: &userDomain.User{ID: 22, PublicID: "requester-public-id", Role: userDomain.RoleUser},
	}

	h := &HTTPHandler{
		moderationService: &mockModerationService{
			getModeratorRequestFn: func(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error) {
				return &domain.ModeratorRequest{ID: 5, PublicID: requestPublicID, RequesterID: 22, Status: domain.RequestStatusPending, CreatedAt: time.Now()}, nil
			},
			reviewModeratorRequestFn: func(ctx context.Context, reviewerID int, requestPublicID string, status, response string) (*domain.ModeratorRequest, error) {
				return &domain.ModeratorRequest{ID: 5, PublicID: requestPublicID, Status: status, Response: response, CreatedAt: time.Now()}, nil
			},
		},
		userService: userLookup,
	}

	body := bytes.NewBufferString(`{"status":"approved","response":"approved"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/moderation/requests/request-id", body)
	req.SetPathValue("id", "request-id")
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "admin-public-id")

	w := httptest.NewRecorder()
	h.ReviewModeratorRequestAPI(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if userLookup.updatedRole != userDomain.RoleModerator {
		t.Fatalf("updated role = %q, want %q", userLookup.updatedRole, userDomain.RoleModerator)
	}
}
