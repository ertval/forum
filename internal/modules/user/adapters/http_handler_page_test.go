// INPUT ADAPTER TEST - HTTP Page Handler Tests
package adapters

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
	platformTemplates "forum/internal/platform/templates"
)

type pageTestMiddlewareProvider struct {
	authenticated bool
	userPublicID  string
}

func (m *pageTestMiddlewareProvider) RequireAuth() authPorts.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.authenticated {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), authPorts.UserIDKey, m.userPublicID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *pageTestMiddlewareProvider) OptionalAuth() authPorts.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

type pageTestServiceContainer struct {
	userService        userPorts.UserService
	middlewareProvider authPorts.AuthMiddleware
}

func (m *pageTestServiceContainer) User() userPorts.UserService {
	return m.userService
}

func (m *pageTestServiceContainer) AuthMiddleware() authPorts.AuthMiddleware {
	return m.middlewareProvider
}

func (m *pageTestServiceContainer) UploadDir() string {
	return "./static/uploads"
}

func TestHTTPHandler_SettingsPage_Authenticated(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(currentDir, "../../../../"))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}
	defer func() {
		_ = os.Chdir(currentDir)
	}()

	userPublicID := "8f2ce2a5-7ac2-4afb-8db8-c8f597cd17b9"

	mockUserService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			if publicID != userPublicID {
				t.Fatalf("expected publicID %s, got %s", userPublicID, publicID)
			}
			return &domain.User{
				PublicID: userPublicID,
				Username: "testuser",
				Email:    "test@example.com",
				Role:     domain.RoleUser,
			}, nil
		},
	}

	container := &pageTestServiceContainer{
		userService: mockUserService,
		middlewareProvider: &pageTestMiddlewareProvider{
			authenticated: true,
			userPublicID:  userPublicID,
		},
	}

	registry := platformTemplates.NewRegistry()
	if _, err := registry.GetOrParse("settings", "templates/base.html", "templates/settings.html"); err != nil {
		t.Fatalf("failed to parse settings template: %v", err)
	}

	handler := NewHTTPHandler(container, registry)
	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Account Settings") {
		t.Fatalf("expected settings page content")
	}
	if !strings.Contains(body, userPublicID) {
		t.Fatalf("expected page to render public UUID")
	}
	if !strings.Contains(body, "testuser") {
		t.Fatalf("expected page to render username")
	}
}

func TestHTTPHandler_SettingsPage_Unauthenticated(t *testing.T) {
	container := &pageTestServiceContainer{
		userService: &MockUserService{},
		middlewareProvider: &pageTestMiddlewareProvider{
			authenticated: false,
		},
	}

	handler := NewHTTPHandler(container, nil)
	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdateSettingsPage_Success(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(currentDir, "../../../../"))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}
	defer func() { _ = os.Chdir(currentDir) }()

	userPublicID := "8f2ce2a5-7ac2-4afb-8db8-c8f597cd17b9"

	mockUserService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return &domain.User{ID: 1, PublicID: userPublicID, Username: "Old Name", Email: "old@example.com"}, nil
		},
		updateSettingsFn: func(ctx context.Context, publicID, username, email, newPassword, avatarPath string) (*domain.User, error) {
			if username != "Alice Smith" {
				t.Fatalf("expected updated username, got %s", username)
			}
			if email != "alice@example.com" {
				t.Fatalf("expected updated email, got %s", email)
			}
			if newPassword != "StrongPass123" {
				t.Fatalf("expected password to be forwarded")
			}
			return &domain.User{ID: 1, PublicID: publicID, Username: username, Email: email}, nil
		},
	}

	container := &pageTestServiceContainer{
		userService: mockUserService,
		middlewareProvider: &pageTestMiddlewareProvider{
			authenticated: true,
			userPublicID:  userPublicID,
		},
	}

	handler := NewHTTPHandler(container, nil)
	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("username", "Alice Smith")
	_ = writer.WriteField("email", "alice@example.com")
	_ = writer.WriteField("new_password", "StrongPass123")
	_ = writer.WriteField("confirm_password", "StrongPass123")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/settings", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", w.Code)
	}
	if location := w.Header().Get("Location"); location != "/settings?updated=1" {
		t.Fatalf("expected redirect to updated settings page, got %s", location)
	}
}

func TestHTTPHandler_UpdateSettingsPage_PasswordMismatch(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(currentDir, "../../../../"))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}
	defer func() { _ = os.Chdir(currentDir) }()

	userPublicID := "8f2ce2a5-7ac2-4afb-8db8-c8f597cd17b9"
	mockUserService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return &domain.User{ID: 1, PublicID: userPublicID, Username: "Alice", Email: "alice@example.com"}, nil
		},
	}

	container := &pageTestServiceContainer{
		userService: mockUserService,
		middlewareProvider: &pageTestMiddlewareProvider{
			authenticated: true,
			userPublicID:  userPublicID,
		},
	}

	registry := platformTemplates.NewRegistry()
	if _, err := registry.GetOrParse("settings", "templates/base.html", "templates/settings.html"); err != nil {
		t.Fatalf("failed to parse settings template: %v", err)
	}

	handler := NewHTTPHandler(container, registry)
	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("username", "Alice")
	_ = writer.WriteField("email", "alice@example.com")
	_ = writer.WriteField("new_password", "StrongPass123")
	_ = writer.WriteField("confirm_password", "DifferentPass123")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/settings", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "password confirmation does not match") {
		t.Fatalf("expected password mismatch error")
	}
}

func TestHTTPHandler_UpdateSettingsPage_UnauthorizedRendersErrorPage(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(currentDir, "../../../../"))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}
	defer func() { _ = os.Chdir(currentDir) }()

	handler := &HTTPHandler{userService: &MockUserService{}}

	req := httptest.NewRequest(http.MethodPost, "/settings", nil)
	w := httptest.NewRecorder()

	handler.UpdateSettingsPage(w, req)

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
