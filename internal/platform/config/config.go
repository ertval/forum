// Package config provides configuration management for the forum application.
// It loads configuration from environment variables, config files, and provides
// default values for all settings.
package config

import (
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
	// Implementation placeholder
	// 1. Load from .env file
	// 2. Load from environment variables
	// 3. Apply default values
	// 4. Validate configuration

	// Initialize Config with default values
	cfg := &Config{}

	cfg.Server.Host = getEnvString("SERVER_HOST", "localhost")
	cfg.Server.Port = getEnvInt("SERVER_PORT", 8080)
	cfg.Server.TLSPort = getEnvInt("SERVER_TLS_PORT", 8443)
	cfg.Server.Environment = getEnvString("SERVER_ENVIRONMENT", "development")
	cfg.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second)
	cfg.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second)
	cfg.Server.IdleTimeout = getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second)

	cfg.Database.Path = getEnvString("DATABASE_PATH", "forum.db")
	cfg.Database.MaxOpenConns = getEnvInt("DATABASE_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getEnvInt("DATABASE_MAX_IDLE_CONNS", 25)
	cfg.Database.ConnMaxLifetime = getEnvDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute)

	cfg.Session.Secret = getEnvString("SESSION_SECRET", "defaultsecret")
	cfg.Session.Duration = getEnvDuration("SESSION_DURATION", 24*time.Hour)
	cfg.Session.CookieName = getEnvString("SESSION_COOKIE_NAME", "forum_session")
	cfg.Session.Secure = getEnvBool("SESSION_SECURE", false)
	cfg.Session.HttpOnly = getEnvBool("SESSION_HTTP_ONLY", true)	

	cfg.Security.TLSCertFile = getEnvString("TLS_CERT_FILE", "cert.pem")
	cfg.Security.TLSKeyFile = getEnvString("TLS_KEY_FILE", "key.pem")
	cfg.Security.RateLimitRequests = getEnvInt("RATE_LIMIT_REQUESTS", 100)
	cfg.Security.RateLimitWindow = getEnvDuration("RATE_LIMIT_WINDOW", time.Minute)
	cfg.Security.MinPasswordLength = getEnvInt("MIN_PASSWORD_LENGTH", 8)	

	cfg.Upload.MaxSize = int64(getEnvInt("UPLOAD_MAX_SIZE_MB", 20)) * 1024 * 1024
	cfg.Upload.AllowedTypes = []string{"image/jpeg", "image/png", "image/gif"}
	cfg.Upload.UploadDir = getEnvString("UPLOAD_DIR", "./uploads")	
	
	cfg.OAuth.Google.ClientID = getEnvString("GOOGLE_OAUTH_CLIENT_ID", "")
	cfg.OAuth.Google.ClientSecret = getEnvString("GOOGLE_OAUTH_CLIENT_SECRET", "")
	cfg.OAuth.Google.RedirectURL = getEnvString("GOOGLE_OAUTH_REDIRECT_URL", "")

	cfg.OAuth.GitHub.ClientID = getEnvString("GITHUB_OAUTH_CLIENT_ID", "")
	cfg.OAuth.GitHub.ClientSecret = getEnvString("GITHUB_OAUTH_CLIENT_SECRET", "")
	cfg.OAuth.GitHub.RedirectURL = getEnvString("GITHUB_OAUTH_REDIRECT_URL", "")


	// Return the populated configuration
	return cfg, nil
}

// Validate validates the configuration values.
// Returns an error if any required value is missing or invalid.
// TODO: Implement validation logic.
func (c *Config) Validate() error {
	// Implementation placeholder
	return nil
}
