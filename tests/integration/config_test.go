package integration

import (
	"os"
	"testing"

	"forum/internal/platform/config"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load error: %v", err)
	}

	t.Logf("Config loaded: DB Path=%s UploadDir=%s SessionSecretLen=%d Env=%s",
		cfg.Database.Path, cfg.Upload.UploadDir, len(cfg.Session.Secret), cfg.Server.Environment)

	if cfg.Database.Path == "" {
		t.Error("Database.Path is empty")
	}
	if cfg.Upload.UploadDir == "" {
		t.Error("Upload.UploadDir is empty")
	}

	// In production, session secret should be at least 16 characters
	// In development, we allow the default secret but warn about it
	if cfg.Server.Environment == "production" {
		if len(cfg.Session.Secret) < 16 {
			t.Errorf("Production: Session.Secret length too short: %d (need >= 16)", len(cfg.Session.Secret))
		}
	} else if os.Getenv("SESSION_SECRET") == "" {
		t.Logf("Warning: Using default session secret in %s environment. Set SESSION_SECRET env var for production.", cfg.Server.Environment)
	}
}
