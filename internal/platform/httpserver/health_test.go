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
	handler := HealthAPI(checker)

	req := httptest.NewRequest(http.MethodGet, "/health-api", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

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
	handler := HealthAPI(checker)

	req := httptest.NewRequest(http.MethodGet, "/health-api", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

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
