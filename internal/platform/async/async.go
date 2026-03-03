// Package async provides utilities for running background operations.
package async

import (
	"context"
	"log"
	"runtime/debug"
	"time"
)

// defaultTimeout is the fixed timeout for all async operations.
const defaultTimeout = 5 * time.Second

// Run executes fn in a background goroutine with a fixed 5-second timeout.
// If fn returns an error, it is logged with the provided label.
// This replaces the repeated fire-and-forget goroutine pattern used
// across service layers for stat-counter updates.
func Run(fn func(context.Context) error, label string) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("ERROR: async panic recovered in %s: %v\n%s", label, recovered, debug.Stack())
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		if err := fn(ctx); err != nil {
			log.Printf("WARNING: %s: %v", label, err)
		}
	}()
}
