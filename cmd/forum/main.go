// Package main is the application entry point.
// It initializes all modules, wires dependencies, and starts the HTTP server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Platform/Infrastructure
	"forum/internal/platform/config"
	"forum/internal/platform/database"
	"forum/internal/platform/httpserver"
	"forum/internal/platform/logger"

	// Module: Auth
	authAdapters "forum/internal/modules/auth/adapters"
	authApp "forum/internal/modules/auth/application"

	// Module: User
	userAdapters "forum/internal/modules/user/adapters"
	userApp "forum/internal/modules/user/application"

	// Module: Post
	postAdapters "forum/internal/modules/post/adapters"
	postApp "forum/internal/modules/post/application"

	// Module: Comment
	commentAdapters "forum/internal/modules/comment/adapters"
	commentApp "forum/internal/modules/comment/application"

	// Module: Reaction
	reactionAdapters "forum/internal/modules/reaction/adapters"
	reactionApp "forum/internal/modules/reaction/application"

	// Module: Moderation [OPTIONAL FEATURE]
	moderationAdapters "forum/internal/modules/moderation/adapters"
	moderationApp "forum/internal/modules/moderation/application"

	// Module: Notification [OPTIONAL FEATURE]
	notificationAdapters "forum/internal/modules/notification/adapters"
	notificationApp "forum/internal/modules/notification/application"
)

func main() {

	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Logger
	lgr := logger.New(logger.InfoLevel, os.Stdout)
	lgr.Info("Starting Forum Application")

	// 3. Initialize Database Connection
	dbConn, err := database.NewConnection(cfg.Database.Path)
	if err != nil {
		lgr.Error("Failed to connect to database", logger.Error(err))
		os.Exit(1)
	}
	defer dbConn.Close()

	// 4. Run Database Migrations
	migrator := database.NewMigrator(dbConn)
	if err := migrator.Migrate("./migrations"); err != nil {
		lgr.Error("Failed to run migrations", logger.Error(err))
		os.Exit(1)
	}

	// 5. Initialize Repositories (Output Adapters)
	// Auth repositories
	sessionRepo := authAdapters.NewSQLiteSessionRepository(dbConn.DB())
	authUserRepo := authAdapters.NewSQLiteUserRepository(dbConn.DB())

	// User repositories
	userRepo := userAdapters.NewSQLiteUserRepository(dbConn.DB())

	// Post repositories
	postRepo := postAdapters.NewSQLitePostRepository(dbConn.DB())
	// categoryRepo := postAdapters.NewSQLiteCategoryRepository(dbConn.DB())

	// Comment repositories
	commentRepo := commentAdapters.NewSQLiteCommentRepository(dbConn.DB())

	// Reaction repositories
	reactionRepo := reactionAdapters.NewSQLiteReactionRepository(dbConn.DB())

	// Moderation repositories [OPTIONAL]
	moderationRepo := moderationAdapters.NewSQLiteReportRepository(dbConn.DB())

	// Notification repositories [OPTIONAL]
	notificationRepo := notificationAdapters.NewSQLiteNotificationRepository(dbConn.DB())

	// 6. Initialize Services (Application Layer)
	// Auth service
	authService := authApp.NewService(
		sessionRepo,
		authUserRepo,
		cfg.Session.Duration,
	)

	// User service
	userService := userApp.NewService(userRepo)

	// Post service (note: categoryRepo would be added here when implemented)
	postService := postApp.NewService(postRepo, nil) // TODO: Add categoryRepo

	// Comment service
	commentService := commentApp.NewService(commentRepo)

	// Reaction service
	reactionService := reactionApp.NewService(reactionRepo)

	// Moderation service [OPTIONAL]
	moderationService := moderationApp.NewService(moderationRepo)

	// Notification service [OPTIONAL]
	notificationService := notificationApp.NewService(notificationRepo)

	// 7. Initialize HTTP Handlers (Input Adapters)
	authHandler := authAdapters.NewHTTPHandler(authService)
	userHandler := userAdapters.NewHTTPHandler(userService)
	postHandler := postAdapters.NewHTTPHandler(postService)
	commentHandler := commentAdapters.NewHTTPHandler(commentService)
	reactionHandler := reactionAdapters.NewHTTPHandler(reactionService)
	moderationHandler := moderationAdapters.NewHTTPHandler(moderationService)
	notificationHandler := notificationAdapters.NewHTTPHandler(notificationService)

	// 8. Initialize HTTP Server
	serverCfg := httpserver.Config{
		Host:         cfg.Server.Host,
		Port:         cfg.Server.Port,
		TLSPort:      cfg.Server.TLSPort,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		TLSCertFile:  cfg.Security.TLSCertFile,
		TLSKeyFile:   cfg.Security.TLSKeyFile,
	}
	server := httpserver.New(serverCfg)

	// 9. Register Global Middleware
	server.RegisterMiddleware(httpserver.Recovery())
	server.RegisterMiddleware(httpserver.Logger())
	server.RegisterMiddleware(httpserver.CORS([]string{"*"}))
	server.RegisterMiddleware(httpserver.RateLimit(
		cfg.Security.RateLimitRequests,
		int(cfg.Security.RateLimitWindow.Seconds()),
	))

	// 10. Register Routes
	authHandler.RegisterRoutes(server.Router())
	userHandler.RegisterRoutes(server.Router())
	postHandler.RegisterRoutes(server.Router())
	commentHandler.RegisterRoutes(server.Router())
	reactionHandler.RegisterRoutes(server.Router())
	moderationHandler.RegisterRoutes(server.Router())
	notificationHandler.RegisterRoutes(server.Router())

	fmt.Println("Registering static file handler")	

	// Serve static files
	server.Router().Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// 11. Start Server
	go func() {
		if err := server.Start(); err != nil {
			lgr.Error("Server failed to start", logger.Error(err))
			os.Exit(1)
		}
	}()

	fmt.Println("Registering static file handler 2")	

	lgr.Info(fmt.Sprintf("Forum server started on port %d (HTTP) and %d (HTTPS)", cfg.Server.Port, cfg.Server.TLSPort))

	// 12. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	lgr.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(); err != nil {
		lgr.Error("Server forced to shutdown", logger.Error(err))
	}

	select {
	case <-ctx.Done():
		lgr.Info("Timeout of 30 seconds exceeded")
	default:
		lgr.Info("Server exited gracefully")
	}
}
