// Package main is the application entry point.
// It loads configuration, initializes the application via wire package,
// and manages the server lifecycle (start and graceful shutdown).
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// 2. Initialize Logger (level from LOG_LEVEL env var)
	logLevel := logger.InfoLevel
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		logLevel = logger.DebugLevel
	}
	lgr := logger.New(logLevel, os.Stdout)

	// 3. Initialize Application (all wiring happens in wire package)
	app, err := wire.InitializeApp(cfg, lgr)
	if err != nil {
		lgr.Error("Failed to initialize application", logger.Error(err))
		os.Exit(1)
	}
	defer app.Cleanup()

	// 4. Start Server
	// Call Start synchronously. Server.Start launches the HTTP(S) listeners
	// in their own goroutines and returns quickly, so this call won't block.
	if err := app.Start(); err != nil {
		lgr.Error("Server failed to start", logger.Error(err))
		os.Exit(1)
	}

	httpAddr := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpsAddr := fmt.Sprintf("https://%s:%d", cfg.Server.Host, cfg.Server.TLSPort)

	urls := httpAddr
	if cfg.Security.TLSCertFile != "" && cfg.Security.TLSKeyFile != "" {
		urls += " " + httpsAddr
	}

	lgr.Info("Forum server started", logger.String("urls", urls))

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println()
	lgr.Info("Shutting down server...")

	// Shutdown uses internal 30s timeout context
	if err := app.Shutdown(); err != nil {
		lgr.Error("Server forced to shutdown", logger.Error(err))
		return
	}

	lgr.Info("Server exited gracefully")
}
