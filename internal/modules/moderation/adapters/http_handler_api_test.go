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
	createFn func(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) (*domain.Report, error)
	listFn   func(ctx context.Context, status string) ([]*domain.Report, error)
	reviewFn func(ctx context.Context, moderatorID int, reportPublicID string, status, response string) (*domain.Report, error)
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

type mockUserLookupService struct {
	user *userDomain.User
	err  error
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
