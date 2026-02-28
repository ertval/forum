package health

import (
	"context"
	"net/http"
	"testing"
)

func TestCheckAPIEndpoints_NotificationAPIUpWhenAllRoutesRegistered(t *testing.T) {
	router := http.NewServeMux()
	router.Handle("GET /api/notifications", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/notifications/{id}/read", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	checker := NewChecker(nil, router)
	results := make(map[string]string)

	checker.checkAPIEndpoints(context.Background(), results)

	if results["notification_api"] != "up" {
		t.Fatalf("notification_api = %q, want %q", results["notification_api"], "up")
	}
}

func TestCheckAPIEndpoints_NotificationAPIDownWhenAnyRouteMissing(t *testing.T) {
	router := http.NewServeMux()
	router.Handle("GET /api/notifications", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	checker := NewChecker(nil, router)
	results := make(map[string]string)

	checker.checkAPIEndpoints(context.Background(), results)

	if results["notification_api"] != "down" {
		t.Fatalf("notification_api = %q, want %q", results["notification_api"], "down")
	}
}
