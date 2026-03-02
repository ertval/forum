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

type endpoint struct {
	method string
	path   string
}

type moduleCheck struct {
	resultKey string
	endpoints []endpoint
}

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
	modules := []moduleCheck{
		{
			resultKey: "auth_api",
			endpoints: []endpoint{
				{"POST", "/api/auth/register"},
				{"POST", "/api/auth/login"},
				{"POST", "/api/auth/logout"},
				{"GET", "/api/auth/session"},
			},
		},
		{
			resultKey: "post_api",
			endpoints: []endpoint{
				{"GET", "/"},
				{"GET", "/api/posts"},
				{"POST", "/api/posts"},
				{"GET", "/api/posts/{id}"},
				{"PUT", "/api/posts/{id}"},
				{"DELETE", "/api/posts/{id}"},
			},
		},
		{
			resultKey: "user_api",
			endpoints: []endpoint{
				{"GET", "/api/users/{id}"},
				{"GET", "/api/users"},
				{"PUT", "/api/users/{id}/role"},
				{"PUT", "/api/users/{id}/deactivate"},
			},
		},
		{
			resultKey: "comment_api",
			endpoints: []endpoint{
				{"POST", "/api/comments/posts/{post_id}"},
				{"GET", "/api/comments/{id}"},
				{"PUT", "/api/comments/{id}"},
				{"DELETE", "/api/comments/{id}"},
				{"GET", "/api/comments/posts/{post_id}"},
			},
		},
		{
			resultKey: "reaction_api",
			endpoints: []endpoint{
				{"POST", "/api/reactions"},
				{"DELETE", "/api/reactions"},
				{"GET", "/api/reactions/{targetType}/{targetId}"},
				{"GET", "/api/reactions/{targetType}/{targetId}/count"},
			},
		},
		{
			resultKey: "moderation_api",
			endpoints: []endpoint{
				{"POST", "/api/moderation/reports"},
				{"GET", "/api/moderation/reports"},
				{"PUT", "/api/moderation/reports/{id}"},
			},
		},
		{
			resultKey: "notification_api",
			endpoints: []endpoint{
				{"GET", "/api/notifications"},
				{"PUT", "/api/notifications/{id}/read"},
			},
		},
	}

	for _, module := range modules {
		allUp := c.areAllRoutesRegistered(ctx, module.endpoints)
		results[module.resultKey] = map[bool]string{true: "up", false: "down"}[allUp]
	}
}

// areAllRoutesRegistered checks if all routes in the list are registered in the router
func (c *Checker) areAllRoutesRegistered(ctx context.Context, endpoints []endpoint) bool {
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
