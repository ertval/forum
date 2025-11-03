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
	return nil, nil
}

// Validate validates the configuration values.
// Returns an error if any required value is missing or invalid.
// TODO: Implement validation logic.
func (c *Config) Validate() error {
	// Implementation placeholder
	return nil
}
