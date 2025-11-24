package health

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"
)

// Checker is responsible for running health checks.
type Checker struct {
	db     *sql.DB
	router *http.ServeMux
}

// NewChecker creates a new Checker.
func NewChecker(db *sql.DB, router *http.ServeMux) *Checker {
	return &Checker{db: db, router: router}
}

// Check performs all health checks and returns a map of results.
func (c *Checker) Check(ctx context.Context) map[string]string {
	results := make(map[string]string)

	// Check database connection
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		results["database"] = "down"
	} else {
		results["database"] = "up"
	}

	// Check API endpoints for each module that has them
	c.checkAPIEndpoints(ctx, results)

	return results
}

// checkAPIEndpoints checks the availability of API endpoints for each module
func (c *Checker) checkAPIEndpoints(ctx context.Context, results map[string]string) {
	// Check all endpoints per module and only mark the module as "up" if ALL endpoints are accessible

	// Auth module endpoints
	authEndpoints := []struct{method, path string}{
		{"POST", "/auth/register"},
		{"POST", "/auth/login"},
		{"POST", "/auth/logout"},
		{"GET", "/auth/session"},
	}
	authAllUp := c.areAllRoutesRegistered(ctx, authEndpoints)
	results["auth_api"] = map[bool]string{true: "up", false: "down"}[authAllUp]

	// Post module endpoints
	postEndpoints := []struct{method, path string}{
		{"GET", "/"}, // homepage
		{"GET", "/posts"}, // list posts
		{"POST", "/posts"}, // create post
		{"GET", "/posts/{id}"}, // get post (parameterized)
		{"PUT", "/posts/{id}"}, // update post (parameterized)
		{"DELETE", "/posts/{id}"}, // delete post (parameterized)
	}
	postAllUp := c.areAllRoutesRegistered(ctx, postEndpoints)
	results["post_api"] = map[bool]string{true: "up", false: "down"}[postAllUp]

	// User module endpoints
	userEndpoints := []struct{method, path string}{
		{"GET", "/users/{id}"},
		{"GET", "/users"},
		{"PUT", "/users/{id}/role"},
		{"PUT", "/users/{id}/deactivate"},
	}
	userAllUp := c.areAllRoutesRegistered(ctx, userEndpoints)
	results["user_api"] = map[bool]string{true: "up", false: "down"}[userAllUp]

	// Comment module endpoints
	commentEndpoints := []struct{method, path string}{
		{"POST", "/comments"},
		{"GET", "/comments/{id}"},
		{"PUT", "/comments/{id}"},
		{"DELETE", "/comments/{id}"},
		{"GET", "/posts/{postId}/comments"},
	}
	commentAllUp := c.areAllRoutesRegistered(ctx, commentEndpoints)
	results["comment_api"] = map[bool]string{true: "up", false: "down"}[commentAllUp]

	// Reaction module endpoints
	reactionEndpoints := []struct{method, path string}{
		{"POST", "/reactions"},
		{"DELETE", "/reactions"},
		{"GET", "/reactions/{targetType}/{targetId}"},
		{"GET", "/reactions/{targetType}/{targetId}/count"},
	}
	reactionAllUp := c.areAllRoutesRegistered(ctx, reactionEndpoints)
	results["reaction_api"] = map[bool]string{true: "up", false: "down"}[reactionAllUp]

	// Moderation module endpoints
	moderationEndpoints := []struct{method, path string}{
		{"POST", "/reports"},
		{"GET", "/reports"},
		{"PUT", "/reports/{id}"},
	}
	moderationAllUp := c.areAllRoutesRegistered(ctx, moderationEndpoints)
	results["moderation_api"] = map[bool]string{true: "up", false: "down"}[moderationAllUp]

	// Notification module endpoints
	notificationEndpoints := []struct{method, path string}{
		{"GET", "/notifications"},
		{"PUT", "/notifications/{id}/read"},
	}
	notificationAllUp := c.areAllRoutesRegistered(ctx, notificationEndpoints)
	results["notification_api"] = map[bool]string{true: "up", false: "down"}[notificationAllUp]
}

// areAllRoutesRegistered checks if all routes in the list are registered in the router
func (c *Checker) areAllRoutesRegistered(ctx context.Context, endpoints []struct{method, path string}) bool {
	for _, endpoint := range endpoints {
		if !c.isRouteRegistered(ctx, endpoint.method, endpoint.path) {
			return false
		}
	}
	return true
}

// isRouteRegistered checks if a specific route is registered in the router
func (c *Checker) isRouteRegistered(ctx context.Context, method, path string) bool {
	// For parameterized routes, we'll try to make a test request
	// to an actual instance of the route (e.g., /posts/1 for /posts/{id})
	testPath := path
	expectedPattern := method + " " + path

	if strings.Contains(path, "{") && strings.Contains(path, "}") {
		// Handle common parameter names in routes
		testPath = strings.Replace(testPath, "{id}", "1", -1)
		testPath = strings.Replace(testPath, "{postId}", "1", -1)
		testPath = strings.Replace(testPath, "{targetType}", "post", -1)
		testPath = strings.Replace(testPath, "{targetId}", "1", -1)
		// Remove any remaining brackets that weren't matched by the specific replacements
		testPath = strings.ReplaceAll(testPath, "{", "1") // fallback for other parameter names
		testPath = strings.ReplaceAll(testPath, "}", "")
	}

	// Create a test request with the appropriate method and test path
	req, err := http.NewRequestWithContext(ctx, method, testPath, nil)
	if err != nil {
		return false
	}

	// Use ServeMux.Handler() to see what pattern would be matched
	_, pattern := c.router.Handler(req)

	// Check if the matched pattern matches our expected pattern
	// For example, "GET /posts" should match "GET /posts" exactly
	// or "GET /posts/{id}" for parameterized routes
	return pattern == expectedPattern
}

