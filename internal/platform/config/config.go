// Package config provides configuration management for the forum application.
// It loads configuration from environment variables, config files, and provides
// default values for all settings.
package config

import (
	"fmt"
	"path/filepath"
	"time"
)

// Config represents the application configuration.
// All configuration values are loaded from environment variables with sensible defaults.
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// Session configuration
	Session SessionConfig

	// Security configuration
	Security SecurityConfig

	// File upload configuration
	Upload UploadConfig

	// Logger configuration
	Logger LoggerConfig

	// OAuth configuration (optional)
	OAuth OAuthConfig
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Port         int           // HTTP port (default: 8080)
	TLSPort      int           // HTTPS port (default: 8443)
	Host         string        // Server host (default: "0.0.0.0")
	Environment  string        // Environment: development, staging, production
	ReadTimeout  time.Duration // HTTP read timeout
	WriteTimeout time.Duration // HTTP write timeout
	IdleTimeout  time.Duration // HTTP idle timeout
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	Path            string        // SQLite database file path
	MigrationsDir   string        // Database migrations directory
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
}

// SessionConfig contains session management settings.
type SessionConfig struct {
	Secret     string        // Session encryption secret
	Duration   time.Duration // Session lifetime (default: 24h)
	CookieName string        // Session cookie name
	Secure     bool          // Use secure cookies (HTTPS only)
	HttpOnly   bool          // HTTP only cookies (no JavaScript access)
}

// SecurityConfig contains security settings.
type SecurityConfig struct {
	// TLS configuration
	TLSCertFile string // TLS certificate file path
	TLSKeyFile  string // TLS private key file path

	// Rate limiting
	RateLimitRequests int           // Max requests per window
	RateLimitWindow   time.Duration // Rate limit time window

	// Password requirements
	MinPasswordLength int // Minimum password length
}

// UploadConfig contains file upload settings.
type UploadConfig struct {
	MaxSize      int64    // Maximum file size in bytes (default: 20MB)
	AllowedTypes []string // Allowed file types (JPEG, PNG, GIF)
	UploadDir    string   // Upload directory path
}

// LoggerConfig contains logging settings.
type LoggerConfig struct {
	Level         string   // Log level: DEBUG, INFO, WARN, ERROR
	TimePrecision string   // Time precision: seconds, nano
	OmitFields    []string // Fields to omit from human output
	AllowedFields []string // Fields to allow in human output (empty = all)
	MaxLineWidth  int      // Maximum line width for human output
	Colorize      bool     // Enable ANSI colors in human output
}

// OAuthConfig contains OAuth provider settings (optional).
type OAuthConfig struct {
	Google GoogleOAuthConfig
	GitHub GitHubOAuthConfig
}

// GoogleOAuthConfig contains Google OAuth settings.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// GitHubOAuthConfig contains GitHub OAuth settings.
type GitHubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Load loads configuration from environment variables and config files.
// It returns a Config struct with all values populated.
// TODO: Implement configuration loading logic.
func Load() (*Config, error) {
	// Initialize Config with default values
	cfg := &Config{}

	cfg.Server.Host = getEnvString("SERVER_HOST", "localhost")
	cfg.Server.Port = getEnvInt("SERVER_PORT", 8080)
	cfg.Server.TLSPort = getEnvInt("SERVER_TLS_PORT", 8443)
	cfg.Server.Environment = getEnvString("SERVER_ENVIRONMENT", "development")
	cfg.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second)
	cfg.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second)
	cfg.Server.IdleTimeout = getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second)

	cfg.Database.Path = getEnvString("DATABASE_PATH", "./data/forum.db")
	cfg.Database.MigrationsDir = getEnvString("DATABASE_MIGRATIONS_DIR", "./migrations")
	cfg.Database.MaxOpenConns = getEnvInt("DATABASE_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getEnvInt("DATABASE_MAX_IDLE_CONNS", 25)
	cfg.Database.ConnMaxLifetime = getEnvDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute)

	cfg.Session.Secret = getEnvString("SESSION_SECRET", "defaultsecret")
	cfg.Session.Duration = getEnvDuration("SESSION_DURATION", 24*time.Hour)
	cfg.Session.CookieName = getEnvString("SESSION_COOKIE_NAME", "forum_session")
	cfg.Session.Secure = getEnvBool("SESSION_SECURE", false)
	cfg.Session.HttpOnly = getEnvBool("SESSION_HTTP_ONLY", true)

	cfg.Security.TLSCertFile = getEnvString("TLS_CERT_FILE", "./certs/cert.pem")
	cfg.Security.TLSKeyFile = getEnvString("TLS_KEY_FILE", "./certs/key.pem")
	cfg.Security.RateLimitRequests = getEnvInt("RATE_LIMIT_REQUESTS", 100)
	cfg.Security.RateLimitWindow = getEnvDuration("RATE_LIMIT_WINDOW", time.Minute)
	cfg.Security.MinPasswordLength = getEnvInt("MIN_PASSWORD_LENGTH", 8)

	cfg.Upload.MaxSize = int64(getEnvInt("UPLOAD_MAX_SIZE_MB", 20)) * 1024 * 1024
	cfg.Upload.AllowedTypes = []string{"image/jpeg", "image/png", "image/gif"}
	cfg.Upload.UploadDir = getEnvString("UPLOAD_DIR", "./static/uploads")

	cfg.Logger.Level = getEnvString("LOG_LEVEL", "INFO")
	cfg.Logger.TimePrecision = getEnvString("LOG_TIME_PRECISION", "seconds")
	cfg.Logger.OmitFields = getEnvStringSlice("LOG_OMIT_FIELDS", []string{"user_agent"})
	cfg.Logger.AllowedFields = getEnvStringSlice("LOG_ALLOWED_FIELDS", []string{"method", "path", "query", "status", "size", "duration_ms", "remote", "url", "response", "error", "errors"})
	cfg.Logger.MaxLineWidth = getEnvInt("LOG_MAX_LINE_WIDTH", 200)
	cfg.Logger.Colorize = getEnvBool("LOG_COLORIZE", true)

	cfg.OAuth.Google.ClientID = getEnvString("GOOGLE_OAUTH_CLIENT_ID", "")
	cfg.OAuth.Google.ClientSecret = getEnvString("GOOGLE_OAUTH_CLIENT_SECRET", "")
	cfg.OAuth.Google.RedirectURL = getEnvString("GOOGLE_OAUTH_REDIRECT_URL", "")

	cfg.OAuth.GitHub.ClientID = getEnvString("GITHUB_OAUTH_CLIENT_ID", "")
	cfg.OAuth.GitHub.ClientSecret = getEnvString("GITHUB_OAUTH_CLIENT_SECRET", "")
	cfg.OAuth.GitHub.RedirectURL = getEnvString("GITHUB_OAUTH_REDIRECT_URL", "")

	// Return the populated configuration

	err := cfg.Validate()

	return cfg, err
}

// Validate validates the configuration values.
// Returns an error if any required value is missing or invalid.
func (c *Config) Validate() error {
	// Validate Server configuration
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Server.TLSPort <= 0 || c.Server.TLSPort > 65535 {
		return fmt.Errorf("invalid TLS port: %d", c.Server.TLSPort)
	}
	if c.Server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}
	if c.Server.Environment != "development" && c.Server.Environment != "staging" && c.Server.Environment != "production" {
		return fmt.Errorf("invalid environment: %s", c.Server.Environment)
	}
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	// Validate Database configuration
	// Allow common developer paths: ./data/forum.db, ./forum.db, or any
	// path whose base name is forum.db (absolute or relative).
	dbBase := filepath.Base(c.Database.Path)
	if !(c.Database.Path == "./data/forum.db" || c.Database.Path == "./db/forum.db" || dbBase == "forum.db") {
		return fmt.Errorf("database path must point to a forum.db file (e.g. './data/forum.db' or './db/forum.db')")
	}
	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}
	if c.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("max idle connections must be positive")
	}
	if c.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("connection max lifetime must be positive")
	}

	// Validate Session configuration
	// In production require a strong secret (32+). In other environments allow
	// a shorter dev secret but it must be non-empty and at least 8 chars.
	if c.Server.Environment == "production" {
		if len(c.Session.Secret) < 32 {
			return fmt.Errorf("session secret must be at least 32 characters long in production")
		}
	} else {
		if len(c.Session.Secret) < 8 {
			return fmt.Errorf("session secret must be at least 8 characters long for non-production environments")
		}
	}
	if c.Session.Duration <= 0 {
		return fmt.Errorf("session duration must be positive")
	}
	if c.Session.CookieName == "" {
		return fmt.Errorf("session cookie name cannot be empty")
	}

	// Validate Security configuration
	if c.Security.MinPasswordLength < 8 {
		return fmt.Errorf("minimum password length must be at least 8 characters")
	}
	if c.Security.RateLimitRequests <= 0 {
		return fmt.Errorf("rate limit requests must be positive")
	}
	if c.Security.RateLimitWindow <= 0 {
		return fmt.Errorf("rate limit window must be positive")
	}
	// Validate TLS configuration if environment is production
	if c.Server.Environment == "production" {
		if c.Security.TLSCertFile == "" {
			return fmt.Errorf("TLS certificate file path cannot be empty in production")
		}
		if c.Security.TLSKeyFile == "" {
			return fmt.Errorf("TLS key file path cannot be empty in production")
		}
	}

	// Validate Upload configuration
	if c.Upload.MaxSize <= 0 {
		return fmt.Errorf("upload max size must be positive")
	}
	if len(c.Upload.AllowedTypes) == 0 {
		return fmt.Errorf("at least one allowed file type must be specified")
	}
	// Validate Upload configuration
	// Allow either ./static/uploads or ./uploads or any path ending in 'uploads'.
	uploadBase := filepath.Base(c.Upload.UploadDir)
	if !(c.Upload.UploadDir == "./static/uploads" || c.Upload.UploadDir == "./uploads" || uploadBase == "uploads") {
		return fmt.Errorf("upload directory path must point to an 'uploads' directory (e.g. './static/uploads' or './uploads')")
	}

	// Validate Logger configuration
	validLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "WARNING": true, "ERROR": true}
	if !validLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s (must be DEBUG, INFO, WARN, or ERROR)", c.Logger.Level)
	}
	validTimePrecisions := map[string]bool{"seconds": true, "nano": true}
	if !validTimePrecisions[c.Logger.TimePrecision] {
		return fmt.Errorf("invalid log time precision: %s (must be 'seconds' or 'nano')", c.Logger.TimePrecision)
	}
	if c.Logger.MaxLineWidth < 0 {
		return fmt.Errorf("log max line width must be non-negative")
	}

	// OAuth configuration is optional, but if provided, validate it
	if c.OAuth.Google.ClientID != "" {
		if c.OAuth.Google.ClientSecret == "" {
			return fmt.Errorf("google OAuth client secret cannot be empty when client ID is provided")
		}
		if c.OAuth.Google.RedirectURL == "" {
			return fmt.Errorf("google OAuth redirect URL cannot be empty when client ID is provided")
		}
	}
	if c.OAuth.GitHub.ClientID != "" {
		if c.OAuth.GitHub.ClientSecret == "" {
			return fmt.Errorf("gitHub OAuth client secret cannot be empty when client ID is provided")
		}
		if c.OAuth.GitHub.RedirectURL == "" {
			return fmt.Errorf("gitHub OAuth redirect URL cannot be empty when client ID is provided")
		}
	}

	return nil
}
