package health

import (
	"context"
	"database/sql"
	"time"
)

// Checker is responsible for running health checks.
type Checker struct {
	db *sql.DB
}

// NewChecker creates a new Checker.
func NewChecker(db *sql.DB) *Checker {
	return &Checker{db: db}
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

	// Simulate checking other dependencies, like the status of core API endpoints
	results["api_endpoints"] = "up"

	return results
}
