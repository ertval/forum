package main

// main.go is the entry point of the forum application.
// It orchestrates dependency injection, module wiring, and server startup.
// This file follows the dependency injection pattern to wire all modules together.

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Platform imports
	"forum/internal/platform/config"
	"forum/internal/platform/database"
	"forum/internal/platform/httpserver"
	"forum/internal/platform/logger"

	// Module application services
	authApp "forum/internal/modules/auth/application"
	userApp "forum/internal/modules/user/application"

	// postApp "forum/internal/modules/post/application"
	// commentApp "forum/internal/modules/comment/application"
	// reactionApp "forum/internal/modules/reaction/application"
	// moderationApp "forum/internal/modules/moderation/application"
	// notificationApp "forum/internal/modules/notification/application"

	// Module adapters
	authHttp "forum/internal/modules/auth/adapters/input/http"
	authBcrypt "forum/internal/modules/auth/adapters/output/crypto/bcrypt"
	authSqlite "forum/internal/modules/auth/adapters/output/persistence/sqlite"

	// authOAuth "forum/internal/modules/auth/adapters/output/oauth"

	userSqlite "forum/internal/modules/user/adapters/output/persistence/sqlite"
)

func main() {
	// Initialize logger
	log := logger.New("info")
	log.Info("Starting forum application...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", logger.Error(err))
	}

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database", logger.Error(err))
	}
	defer db.Close()

	// Run database migrations
	migrator := database.NewMigrator(db)
	if err := migrator.Up(); err != nil {
		log.Fatal("Failed to run migrations", logger.Error(err))
	}

	// Initialize repositories (adapters)
	sessionRepo := authSqlite.NewSessionRepository(db)
	userRepo := userSqlite.NewUserRepository(db)
	passwordHasher := authBcrypt.NewPasswordHasher()

	// Initialize application services
	authService := authApp.NewService(
		sessionRepo,
		userRepo,
		passwordHasher,
		24*time.Hour, // Session TTL
	)

	userService := userApp.NewService(userRepo)

	// TODO: Initialize other services (post, comment, reaction, etc.)

	// Initialize HTTP handlers
	authHandler := authHttp.NewHandler(authService)
	authMiddleware := authHttp.NewAuthMiddleware(authService)

	// TODO: Initialize other handlers

	// Setup HTTP server and register routes
	router := http.NewServeMux()

	// Register auth routes
	authHandler.RegisterRoutes(router)

	// TODO: Register other module routes

	// Apply middleware
	handler := httpserver.Chain(
		httpserver.Recovery(),
		httpserver.RequestLogger(),
		httpserver.SecurityHeaders(),
		httpserver.RateLimit(100), // 100 requests per minute
	)(router)

	// Create and start HTTP server
	server := httpserver.New(httpserver.Config{
		Port:         cfg.Port,
		TLSPort:      cfg.TLSPort,
		ReadTimeout:  15,
		WriteTimeout: 15,
		IdleTimeout:  60,
		CertFile:     cfg.CertFile,
		KeyFile:      cfg.KeyFile,
	}, handler)

	// Start server in goroutine
	go func() {
		log.Info("Server starting", logger.String("port", cfg.Port))
		if err := server.Start(); err != nil {
			log.Fatal("Server failed to start", logger.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Shutdown server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(); err != nil {
		log.Error("Server forced to shutdown", logger.Error(err))
	}

	log.Info("Server exited")
}
