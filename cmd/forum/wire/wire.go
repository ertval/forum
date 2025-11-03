// Main wiring logic - orchestrates initialization of all components
package wire

import (
	"fmt"
	"net/http"

	"forum/internal/platform/config"
	"forum/internal/platform/database"
	"forum/internal/platform/httpserver"
	"forum/internal/platform/logger"
)

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
	services := initServices(repos, cfg.Session.Duration)

	// 4. Initialize HTTP Handlers (Input Adapters)
	handlers := initHandlers(services)

	// 5. Initialize HTTP Server
	server := initServer(cfg, handlers, lgr)

	lgr.Info("Application initialization complete")

	return &App{
		Server: server,
		DB:     db,
		Logger: lgr,
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
	if err := migrator.Migrate("./migrations"); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbConn, nil
}

// initServer creates and configures the HTTP server with all routes and middleware.
func initServer(cfg *config.Config, handlers *Handlers, lgr *logger.Logger) *httpserver.Server {
	lgr.Info("Initializing HTTP server")

	// Create server
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

	// Register global middleware
	server.RegisterMiddleware(httpserver.Recovery())
	server.RegisterMiddleware(httpserver.Logger())
	server.RegisterMiddleware(httpserver.CORS([]string{"*"}))
	server.RegisterMiddleware(httpserver.RateLimit(
		cfg.Security.RateLimitRequests,
		int(cfg.Security.RateLimitWindow.Seconds()),
	))

	// Register module routes
	handlers.Auth.RegisterRoutes(server.Router())
	handlers.User.RegisterRoutes(server.Router())
	handlers.Post.RegisterRoutes(server.Router())
	handlers.Comment.RegisterRoutes(server.Router())
	handlers.Reaction.RegisterRoutes(server.Router())
	handlers.Moderation.RegisterRoutes(server.Router())
	handlers.Notification.RegisterRoutes(server.Router())

	// Serve static files
	lgr.Info("Registering static file handler")
	server.Router().Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return server
}
