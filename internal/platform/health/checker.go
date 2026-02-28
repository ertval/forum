package health

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Checker is responsible for running health checks.
type Checker struct {
	db     *sql.DB
	router *http.ServeMux
}

// NewChecker creates a new Checker.
// Returns an error if required dependencies are nil.
func NewChecker(db *sql.DB, router *http.ServeMux) *Checker {
	// Note: We allow nil checks to happen at Check() time to provide
	// meaningful error messages in health check results rather than panicking
	return &Checker{db: db, router: router}
}

// NewCheckerWithValidation creates a new Checker with strict validation.
// Returns an error if required dependencies are nil.
func NewCheckerWithValidation(db *sql.DB, router *http.ServeMux) (*Checker, error) {
	if db == nil {
		return nil, errors.New("database connection cannot be nil")
	}
	if router == nil {
		return nil, errors.New("router cannot be nil")
	}
	return &Checker{db: db, router: router}, nil
}

// Check performs all health checks and returns a map of results.
func (c *Checker) Check(ctx context.Context) map[string]string {
	results := make(map[string]string)

	// Check database connection (with nil check to prevent panic - CRIT-4)
	if c.db == nil {
		results["database"] = "down (not configured)"
	} else {
		dbCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		if err := c.db.PingContext(dbCtx); err != nil {
			results["database"] = "down"
		} else {
			results["database"] = "up"
		}
	}

	// Check API endpoints for each module that has them
	if c.router != nil {
		c.checkAPIEndpoints(ctx, results)
	}

	return results
}

// checkAPIEndpoints checks the availability of API endpoints for each module
func (c *Checker) checkAPIEndpoints(ctx context.Context, results map[string]string) {
	// Check all endpoints per module and only mark the module as "up" if ALL endpoints are accessible

	// Auth module endpoints
	authEndpoints := []struct{ method, path string }{
		{"POST", "/api/auth/register"},
		{"POST", "/api/auth/login"},
		{"POST", "/api/auth/logout"},
		{"GET", "/api/auth/session"},
	}
	authAllUp := c.areAllRoutesRegistered(ctx, authEndpoints)
	results["auth_api"] = map[bool]string{true: "up", false: "down"}[authAllUp]

	// Post module endpoints
	postEndpoints := []struct{ method, path string }{
		{"GET", "/"},                  // homepage
		{"GET", "/api/posts"},         // list posts
		{"POST", "/api/posts"},        // create post
		{"GET", "/api/posts/{id}"},    // get post (parameterized)
		{"PUT", "/api/posts/{id}"},    // update post (parameterized)
		{"DELETE", "/api/posts/{id}"}, // delete post (parameterized)
	}
	postAllUp := c.areAllRoutesRegistered(ctx, postEndpoints)
	results["post_api"] = map[bool]string{true: "up", false: "down"}[postAllUp]

	// User module endpoints
	userEndpoints := []struct{ method, path string }{
		{"GET", "/api/users/{id}"},
		{"GET", "/api/users"},
		{"PUT", "/api/users/{id}/role"},
		{"PUT", "/api/users/{id}/deactivate"},
	}
	userAllUp := c.areAllRoutesRegistered(ctx, userEndpoints)
	results["user_api"] = map[bool]string{true: "up", false: "down"}[userAllUp]

	// Comment module endpoints
	commentEndpoints := []struct{ method, path string }{
		{"POST", "/api/comments/posts/{post_id}"},
		{"GET", "/api/comments/{id}"},
		{"PUT", "/api/comments/{id}"},
		{"DELETE", "/api/comments/{id}"},
		{"GET", "/api/comments/posts/{post_id}"},
	}
	commentAllUp := c.areAllRoutesRegistered(ctx, commentEndpoints)
	results["comment_api"] = map[bool]string{true: "up", false: "down"}[commentAllUp]

	// Reaction module endpoints
	reactionEndpoints := []struct{ method, path string }{
		{"POST", "/api/reactions"},
		{"DELETE", "/api/reactions"},
		{"GET", "/api/reactions/{targetType}/{targetId}"},
		{"GET", "/api/reactions/{targetType}/{targetId}/count"},
	}
	reactionAllUp := c.areAllRoutesRegistered(ctx, reactionEndpoints)
	results["reaction_api"] = map[bool]string{true: "up", false: "down"}[reactionAllUp]

	// Moderation module endpoints - NOT YET IMPLEMENTED
	moderationEndpoints := []struct{ method, path string }{
		{"POST", "/api/moderation/reports"},
		{"GET", "/api/moderation/reports"},
		{"PUT", "/api/moderation/reports/{id}"},
	}
	moderationAllUp := c.areAllRoutesRegistered(ctx, moderationEndpoints)
	results["moderation_api"] = map[bool]string{true: "up", false: "down"}[moderationAllUp]

	// Notification module endpoints
	notificationEndpoints := []struct{ method, path string }{
		{"GET", "/api/notifications"},
		{"PUT", "/api/notifications/{id}/read"},
	}
	notificationAllUp := c.areAllRoutesRegistered(ctx, notificationEndpoints)
	results["notification_api"] = map[bool]string{true: "up", false: "down"}[notificationAllUp]
}

// areAllRoutesRegistered checks if all routes in the list are registered in the router
func (c *Checker) areAllRoutesRegistered(ctx context.Context, endpoints []struct{ method, path string }) bool {
	for _, endpoint := range endpoints {
		if !c.isRouteRegistered(ctx, endpoint.method, endpoint.path) {
			return false
		}
	}
	return true
}

// pathParamRegex matches path parameters like {id}, {postId}, etc. (KISS-7)
var pathParamRegex = regexp.MustCompile(`\{[^}]+\}`)

// isRouteRegistered checks if a specific route is registered in the router
func (c *Checker) isRouteRegistered(ctx context.Context, method, path string) bool {
	// For parameterized routes, replace all path parameters with test values
	testPath := path
	expectedPattern := method + " " + path

	if strings.Contains(path, "{") {
		// Use regex to replace all path parameters with a test value (KISS-7)
		testPath = pathParamRegex.ReplaceAllString(path, "test-value-1")
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
