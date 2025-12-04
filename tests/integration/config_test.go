package integration

import (
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
    if len(cfg.Session.Secret) < 16 {
        t.Errorf("Session.Secret length too short: %d", len(cfg.Session.Secret))
    }
}
