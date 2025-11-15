// Package main is the application entry point.
// It loads configuration, initializes the application via wire package,
// and manages the server lifecycle (start and graceful shutdown).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"forum/cmd/forum/wire"
	"forum/internal/platform/config"
	"forum/internal/platform/logger"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Logger
	log := logger.New(logger.InfoLevel, os.Stdout)
	log.Info("Starting Forum Application")

	// 3. Initialize Application (all wiring happens in wire package)
	app, err := wire.InitializeApp(cfg, log)
	if err != nil {
		log.Error("Failed to initialize application", logger.Error(err))
		os.Exit(1)
	}
	defer app.Cleanup()

	// 4. Start Server
	// Call Start synchronously. Server.Start launches the HTTP(S) listeners
	// in their own goroutines and returns quickly, so this call won't block.
	if err := app.Start(); err != nil {
		log.Error("Server failed to start", logger.Error(err))
		os.Exit(1)
	}

	log.Info(fmt.Sprintf("Forum server started on port %d (HTTP) and %d (HTTPS)",
		cfg.Server.Port, cfg.Server.TLSPort))
	log.Info(fmt.Sprintf("Access the forum at: http://localhost:%d", cfg.Server.Port))

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Shutdown(); err != nil {
		log.Error("Server forced to shutdown", logger.Error(err))
	}

	select {
	case <-ctx.Done():
		log.Info("Timeout of 30 seconds exceeded")
	default:
		log.Info("Server exited gracefully")
	}
}
