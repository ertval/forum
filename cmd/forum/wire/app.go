package wire

// Package wire handles dependency injection and application wiring.
// It initializes all components and returns a fully configured App instance.

import (
	"forum/internal/platform/database"
	"forum/internal/platform/httpserver"
	"forum/internal/platform/logger"
)

// App encapsulates the entire application with all its dependencies.
type App struct {
	Server *httpserver.Server
	DB     *database.Connection
	Logger *logger.Logger
}

// Cleanup performs graceful cleanup of application resources.
func (a *App) Cleanup() error {
	a.Logger.Info("Cleaning up application resources")
	if err := a.DB.Close(); err != nil {
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
