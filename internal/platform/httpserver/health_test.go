package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
