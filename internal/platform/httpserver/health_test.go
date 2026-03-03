package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forum/internal/platform/database"
	"forum/internal/platform/health"
)

func TestHealthAPI_IncludesNotificationStatusWhenRoutesRegistered(t *testing.T) {
	router := http.NewServeMux()
	router.Handle("GET /api/notifications", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/notifications/{id}/read", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	checker := health.NewChecker(nil, router)
	handler := NewHealthHandler(checker, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health-api", nil)
	rec := httptest.NewRecorder()

	handler.HealthAPI(rec, req)

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode health api response: %v", err)
	}

	if body["notification_api"] != "up" {
		t.Fatalf("notification_api = %q, want %q", body["notification_api"], "up")
	}
}

func TestHealthAPI_ReadinessIgnoresOptionalChecks(t *testing.T) {
	conn, err := database.NewConnection(":memory:")
	if err != nil {
		t.Fatalf("failed to create db connection: %v", err)
	}
	defer conn.Close()

	router := http.NewServeMux()

	// Register all critical module routes
	router.Handle("POST /api/auth/register", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("POST /api/auth/login", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("POST /api/auth/logout", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/auth/session", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	router.Handle("GET /", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/posts", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("POST /api/posts", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/posts/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/posts/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("DELETE /api/posts/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	router.Handle("GET /api/users/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/users", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/users/{id}/role", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/users/{id}/deactivate", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	router.Handle("POST /api/comments/posts/{post_id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/comments/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/comments/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("DELETE /api/comments/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/comments/posts/{post_id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	router.Handle("POST /api/reactions", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("DELETE /api/reactions", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/reactions/{targetType}/{targetId}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/reactions/{targetType}/{targetId}/count", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	router.Handle("GET /api/notifications", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/notifications/{id}/read", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	checker := health.NewChecker(conn.DB(), router)
	handler := NewHealthHandler(checker, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health-api", nil)
	rec := httptest.NewRecorder()
	handler.HealthAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode health api response: %v", err)
	}

	if body["moderation_api"] != "down" {
		t.Fatalf("moderation_api = %q, want %q", body["moderation_api"], "down")
	}
}

func TestHealthErrorTestRoutes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "400 route", path: "/health/errors/400", wantStatus: http.StatusBadRequest},
		{name: "404 route", path: "/health/errors/404", wantStatus: http.StatusNotFound},
		{name: "500 route", path: "/health/errors/500", wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := http.NewServeMux()
			handler := NewHealthHandler(nil, nil, nil, nil)
			handler.RegisterRoutes(router)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
