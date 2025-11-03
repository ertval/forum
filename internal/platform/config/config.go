// Package config provides configuration management for the forum application.
// It loads configuration from environment variables, config files, and provides
// default values for all settings.
package config

import (
	"fmt"
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
	if c.Database.Path != "./db/forum.db" {
		return fmt.Errorf("database path must be 'db/forum.db'")
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
	if len(c.Session.Secret) < 32 {
		return fmt.Errorf("session secret must be at least 32 characters long")
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
	if c.Upload.UploadDir != "./static/uploads" {
		return fmt.Errorf("upload directory path must be './static/uploads'")
	}

	// OAuth configuration is optional, but if provided, validate it
	if c.OAuth.Google.ClientID != "" {
		if c.OAuth.Google.ClientSecret == "" {
			return fmt.Errorf("Google OAuth client secret cannot be empty when client ID is provided")
		}
		if c.OAuth.Google.RedirectURL == "" {
			return fmt.Errorf("Google OAuth redirect URL cannot be empty when client ID is provided")
		}
	}
	if c.OAuth.GitHub.ClientID != "" {
		if c.OAuth.GitHub.ClientSecret == "" {
			return fmt.Errorf("GitHub OAuth client secret cannot be empty when client ID is provided")
		}
		if c.OAuth.GitHub.RedirectURL == "" {
			return fmt.Errorf("GitHub OAuth redirect URL cannot be empty when client ID is provided")
		}
	}

	return nil
}
