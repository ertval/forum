// INPUT ADAPTER TEST - HTTP Middleware Tests
package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authDomain "forum/internal/modules/auth/domain"
	userDomain "forum/internal/modules/user/domain"
)

type middlewareAuthServiceStub struct{}

func (s *middlewareAuthServiceStub) Register(ctx context.Context, email, username, password string) (int, *authDomain.Session, error) {
	return 0, nil, nil
}

func (s *middlewareAuthServiceStub) Login(ctx context.Context, email, password string) (*authDomain.Session, error) {
	return nil, nil
}

func (s *middlewareAuthServiceStub) Logout(ctx context.Context, sessionToken string) error {
	return nil
}

func (s *middlewareAuthServiceStub) ValidateSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, authDomain.ErrSessionNotFound
}

func (s *middlewareAuthServiceStub) RefreshSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, authDomain.ErrSessionNotFound
}

func (s *middlewareAuthServiceStub) GetSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, authDomain.ErrSessionNotFound
}

type middlewareUserServiceStub struct{}

func (s *middlewareUserServiceStub) CreateUser(ctx context.Context, email, username, passwordHash string) (int, error) {
	return 0, nil
}

func (s *middlewareUserServiceStub) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	return nil, nil
}

func (s *middlewareUserServiceStub) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	return nil, nil
}

func (s *middlewareUserServiceStub) GetByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	return nil, nil
}

func (s *middlewareUserServiceStub) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}

func (s *middlewareUserServiceStub) UpdateRole(ctx context.Context, userID int, newRole userDomain.Role) error {
	return nil
}

func (s *middlewareUserServiceStub) DeactivateUser(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) ActivateUser(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) ListUsers(ctx context.Context, offset, limit int) ([]*userDomain.User, error) {
	return nil, nil
}

func (s *middlewareUserServiceStub) IncrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) DecrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) IncrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) DecrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (s *middlewareUserServiceStub) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (s *middlewareUserServiceStub) IncrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) DecrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

func (s *middlewareUserServiceStub) UpdateSettings(ctx context.Context, publicID, username, email, newPassword, avatarPath string) (*userDomain.User, error) {
	return nil, nil
}

func TestAuthMiddleware_RequireAuth_UnauthorizedPageRendersHTML(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(currentDir, "../../../../"))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}
	defer func() { _ = os.Chdir(currentDir) }()

	middleware := NewAuthMiddleware(&middlewareAuthServiceStub{}, &middlewareUserServiceStub{}, "session_token")
	handler := middleware.RequireAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("expected HTML content type, got %q", got)
	}
	body := w.Body.String()
	if !strings.Contains(body, "error-page") {
		t.Fatalf("expected styled error page content")
	}
	if !strings.Contains(body, "Unauthorized") {
		t.Fatalf("expected unauthorized title in error page")
	}
}

func TestAuthMiddleware_RequireAuth_UnauthorizedAPIStaysJSON(t *testing.T) {
	middleware := NewAuthMiddleware(&middlewareAuthServiceStub{}, &middlewareUserServiceStub{}, "session_token")
	handler := middleware.RequireAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	if body := strings.TrimSpace(w.Body.String()); body != `{"error":"Unauthorized"}` {
		t.Fatalf("unexpected JSON response body: %q", body)
	}
}
