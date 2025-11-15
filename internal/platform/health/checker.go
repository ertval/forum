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
	// Check one representative endpoint per module to determine if the module is available

	// Check auth module by testing one of its endpoints
	authUp := c.isRouteRegistered(ctx, "POST", "/auth/register")
	results["auth_api"] = map[bool]string{true: "up", false: "down"}[authUp]

	// Check post module by testing one of its endpoints
	postUp := c.isRouteRegistered(ctx, "GET", "/posts")
	results["post_api"] = map[bool]string{true: "up", false: "down"}[postUp]

	// For other modules, check if they have any registered endpoints
	userUp := c.isRouteRegistered(ctx, "GET", "/users/1") // Testing a potential user endpoint
	results["user_api"] = map[bool]string{true: "up", false: "down"}[userUp]

	commentUp := c.isRouteRegistered(ctx, "POST", "/comments") // Testing a potential comment endpoint
	results["comment_api"] = map[bool]string{true: "up", false: "down"}[commentUp]

	reactionUp := c.isRouteRegistered(ctx, "POST", "/reactions") // Testing a potential reaction endpoint
	results["reaction_api"] = map[bool]string{true: "up", false: "down"}[reactionUp]

	moderationUp := c.isRouteRegistered(ctx, "POST", "/reports") // Testing a potential moderation endpoint
	results["moderation_api"] = map[bool]string{true: "up", false: "down"}[moderationUp]

	notificationUp := c.isRouteRegistered(ctx, "GET", "/notifications") // Testing a potential notification endpoint
	results["notification_api"] = map[bool]string{true: "up", false: "down"}[notificationUp]
}

// isRouteRegistered checks if a specific route is registered in the router
func (c *Checker) isRouteRegistered(ctx context.Context, method, path string) bool {
	// For parameterized routes like /posts/{id}, we'll try to make a test request
	// to an actual instance of the route (e.g., /posts/1)
	testPath := path
	if strings.Contains(path, "{") && strings.Contains(path, "}") {
		// Replace parameters with actual values for testing
		testPath = strings.Replace(path, "{id}", "1", -1)
		testPath = strings.Replace(testPath, "{", "1", -1) // Handle other parameter formats
		testPath = strings.Replace(testPath, "}", "", -1)
	}

	// Create a test request with the appropriate method and test path
	req, err := http.NewRequestWithContext(ctx, method, testPath, nil)
	if err != nil {
		return false
	}

	// Use ServeMux.Handler() to see what pattern would be matched
	_, pattern := c.router.Handler(req)

	// Create a request to a definitely non-existent path to get the default pattern
	testReq, _ := http.NewRequestWithContext(ctx, "GET", "/__nonexistent_route__", nil)
	_, defaultPattern := c.router.Handler(testReq)

	// If the pattern is not empty and not the same as the default, it means a route is registered
	return pattern != "" && pattern != defaultPattern
}

