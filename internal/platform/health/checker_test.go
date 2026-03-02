package health

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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

func TestCheckAPIEndpoints_ModerationAPIUpWhenAllRoutesRegistered(t *testing.T) {
	router := http.NewServeMux()
	router.Handle("POST /api/moderation/reports", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/moderation/reports", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/moderation/reports/{id}", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	checker := NewChecker(nil, router)
	results := make(map[string]string)

	checker.checkAPIEndpoints(context.Background(), results)

	if results["moderation_api"] != "up" {
		t.Fatalf("moderation_api = %q, want %q", results["moderation_api"], "up")
	}
}

func TestCheckAPIEndpoints_ModerationAPIDownWhenAnyRouteMissing(t *testing.T) {
	router := http.NewServeMux()
	router.Handle("POST /api/moderation/reports", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("GET /api/moderation/reports", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	checker := NewChecker(nil, router)
	results := make(map[string]string)

	checker.checkAPIEndpoints(context.Background(), results)

	if results["moderation_api"] != "down" {
		t.Fatalf("moderation_api = %q, want %q", results["moderation_api"], "down")
	}
}

// TestChecker_Check_EndToEnd calls checker.Check end-to-end with a nil DB
// (configured as "not configured") and a minimal router.
func TestChecker_Check_EndToEnd(t *testing.T) {
	router := http.NewServeMux()
	// Register both notification routes so notification_api reports "up".
	router.Handle("GET /api/notifications", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	router.Handle("PUT /api/notifications/{id}/read", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	checker := NewChecker(nil, router)
	results := checker.Check(context.Background())

	if results == nil {
		t.Fatal("Check() returned nil results map")
	}

	// Nil DB should produce the "not configured" sentinel.
	if want := "down (not configured)"; results["database"] != want {
		t.Errorf("database = %q, want %q", results["database"], want)
	}

	// The two registered notification routes should make the API report "up".
	if results["notification_api"] != "up" {
		t.Errorf("notification_api = %q, want %q", results["notification_api"], "up")
	}
}

// TestChecker_Check_DBDown verifies that when the database connection is
// broken (closed), Check reports database as "down".
func TestChecker_Check_DBDown(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	// Close the connection so that PingContext returns an error.
	db.Close()

	checker := NewChecker(db, http.NewServeMux())
	results := checker.Check(context.Background())

	if results["database"] != "down" {
		t.Errorf("database = %q, want %q", results["database"], "down")
	}
}
