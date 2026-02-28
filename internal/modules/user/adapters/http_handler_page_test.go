// INPUT ADAPTER TEST - HTTP Page Handler Tests
package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
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

	handler := NewHTTPHandler(container, nil)
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
