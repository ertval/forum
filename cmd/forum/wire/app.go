package wire

// Package wire handles dependency injection and application wiring.
// It initializes all components and returns a fully configured App instance.

import (
	"database/sql"
	"fmt"
	"forum/internal/platform/config"
	"forum/internal/platform/database"
	"forum/internal/platform/health"
	"forum/internal/platform/httpserver"
	"forum/internal/platform/logger"
	"net/http"
	"os"
)

// App encapsulates the entire application with all its dependencies.
type App struct {
	Server   *httpserver.Server
	Database *database.Connection
	Logger   *logger.Logger
}

// Cleanup performs graceful cleanup of application resources.
func (a *App) Cleanup() error {
	a.Logger.Info("Cleaning up application resources")
	if err := a.Database.Close(); err != nil {
		a.Logger.Error("Failed to close database connection", logger.Error(err))
		return err
	}
	return nil
}

// Start starts the HTTP server.
func (a *App) Start() error {
	return a.Server.Start()
}

// Shutdown gracefully shuts down the HTTP server.
func (a *App) Shutdown() error {
	return a.Server.Shutdown()
}

// InitializeApp creates and wires all application components.
// This is the main dependency injection function called from main.go.
func InitializeApp(cfg *config.Config, lgr *logger.Logger) (*App, error) {
	lgr.Info("Initializing application components")

	// 1. Initialize Database
	db, err := initDatabase(cfg, lgr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 2. Initialize Repositories (Output Adapters)
	repos := initRepositories(db.DB())

	// 3. Initialize Services (Application Layer)
	services := initServices(repos, cfg.Session.Duration, lgr)

	// 4. Initialize HTTP Handlers (Input Adapters)
	handlers := initHandlers(services)

	// 5. Initialize HTTP Server
	server := initServer(cfg, lgr, handlers, db.DB())

	lgr.Info("Application initialization complete")

	return &App{
		Server:   server,
		Database: db,
		Logger:   lgr,
	}, nil
}

// initDatabase establishes database connection and runs migrations.
func initDatabase(cfg *config.Config, lgr *logger.Logger) (*database.Connection, error) {
	lgr.Info("Connecting to database")

	dbConn, err := database.NewConnection(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	lgr.Info("Running database migrations")
	migrator := database.NewMigrator(dbConn)
	if err := migrator.Migrate(cfg.Database.MigrationsDir); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbConn, nil
}

// initServer creates and configures the HTTP server with all routes and middleware.
func initServer(cfg *config.Config, lgr *logger.Logger, handlers *Handlers, db *sql.DB) *httpserver.Server {
	lgr.Info("Initializing HTTP server")

	// Create server with config as single source of truth
	server := httpserver.New(cfg)

	// Register global middleware in correct order: recovery -> logger -> security headers -> rate limit -> cors
	server.RegisterMiddleware(httpserver.Recovery(lgr))
	server.RegisterMiddleware(httpserver.Logger(lgr))
	server.RegisterMiddleware(httpserver.SecurityHeaders(httpserver.DefaultSecurityHeadersConfig()))
	server.RegisterMiddleware(httpserver.CORS([]string{"*"}))
	server.RegisterMiddleware(httpserver.RateLimit(
		cfg.Security.RateLimitRequests,
		int(cfg.Security.RateLimitWindow.Seconds()),
	))

	// Register module routes first so they are available for the health check
	handlers.Auth.RegisterRoutes(server.Router())
	handlers.User.RegisterRoutes(server.Router())
	handlers.Post.RegisterRoutes(server.Router())
	handlers.Comment.RegisterRoutes(server.Router())
	handlers.Reaction.RegisterRoutes(server.Router())
	handlers.Moderation.RegisterRoutes(server.Router())
	handlers.Notification.RegisterRoutes(server.Router())

	// Create health checker after routes are registered
	healthChecker := health.NewChecker(db, server.Router())

	// Register health check routes with proper configuration
	server.Router().Handle("GET /health", httpserver.HealthPage(httpserver.HealthPageConfig{
		Checker:   healthChecker,
		Templates: handlers.Post.Templates(), // Reuse shared templates
		AuthFunc:  handlers.Auth.GetCurrentUser,
	}))
	server.Router().Handle("GET /health-api", httpserver.HealthAPI(healthChecker))

	// Serve static files (optional - skip if directory doesn't exist)
	// This allows tests to run without static files
	lgr.Info("Checking for static directory")
	if _, err := os.Stat("./static"); err == nil {
		lgr.Info("Registering static file handler")
		server.Router().Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	}

	return server
}
