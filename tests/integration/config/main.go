package main

import (
	"fmt"
	"os"

	"forum/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config.Load error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config loaded:\n DB Path=%s\n UploadDir=%s\n SessionSecretLen=%d\n Env=%s\n",
		cfg.Database.Path, cfg.Upload.UploadDir, len(cfg.Session.Secret), cfg.Server.Environment)
}
