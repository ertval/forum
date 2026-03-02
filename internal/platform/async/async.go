// Package async provides utilities for running background operations.
package async

import (
	"context"
	"log"
	"time"
)

// Run executes fn in a background goroutine with the given timeout.
// If fn returns an error, it is logged with the provided label.
// This replaces the repeated fire-and-forget goroutine pattern used
// across service layers for stat-counter updates.
func Run(fn func(context.Context) error, timeout time.Duration, label string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := fn(ctx); err != nil {
			log.Printf("WARNING: %s: %v", label, err)
		}
	}()
}
