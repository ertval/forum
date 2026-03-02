// Package config provides configuration management for the forum application.
// It loads configuration from environment variables, config files, and provides
// default values for all settings.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// generateDevSecret creates a random development secret.
func generateDevSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a reasonable default if crypto/rand fails (shouldn't happen)
		return "dev-fallback-secret-do-not-use-in-production"
	}
	return hex.EncodeToString(b)
}

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



// Load loads configuration from environment variables and config files.
// It returns a Config struct with all values populated.
func Load() (*Config, error) {
	// Initialize Config with default values
	cfg := &Config{}

	cfg.Server.Host = getEnvString("SERVER_HOST", "0.0.0.0")
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

	sessionSecret := getEnvString("SESSION_SECRET", "")
	if sessionSecret == "" {
		sessionSecret = generateDevSecret()
	}
	cfg.Session.Secret = sessionSecret
	cfg.Session.Duration = getEnvDuration("SESSION_DURATION", 24*time.Hour)
	cfg.Session.CookieName = getEnvString("SESSION_COOKIE_NAME", "session_token")
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
	if !slices.Contains([]string{"development", "staging", "production"}, c.Server.Environment) {
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
	// Accept any non-empty path with a .db extension and no null bytes.
	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}
	if strings.ContainsRune(c.Database.Path, 0) {
		return fmt.Errorf("database path contains null bytes")
	}
	if filepath.Ext(c.Database.Path) != ".db" {
		return fmt.Errorf("database path must have .db extension, got %q", filepath.Base(c.Database.Path))
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
	if c.Upload.UploadDir == "" {
		return fmt.Errorf("upload directory path cannot be empty")
	}
	if strings.ContainsRune(c.Upload.UploadDir, 0) {
		return fmt.Errorf("upload directory path contains null bytes")
	}

	return nil
}

// IsDefaultSecret returns true if the session secret was not explicitly configured.
// Callers can use this to log warnings.
func (c *Config) IsDefaultSecret() bool {
	return os.Getenv("SESSION_SECRET") == ""
}
