# cmd/forum

This directory contains the main application entry point and dependency injection configuration for the forum.

## Overview

The `main.go` file acts strictly as the application bootstrapper:
1. Loads environment configuration.
2. Initializes the core logger.
3. Defers execution to the completely isolated `wire` package for dependency injection mapping.
4. Spawns the HTTP server natively and listens for graceful shutdown signals (OS Interrupts/SIGTERM).

## Directory Layout

```text
cmd/forum/
├── main.go         # Bootstrapper (Config -> Logger -> Wire Init)
└── wire/           # The composition root
    ├── README.md   # Extensive documentation on DI constraints
    ├── app.go             # Global HTTP route definitions + middlewares
    ├── handlers.go        # Instantiation of all API and Page Handlers
    ├── repositories.go    # SQLite repository instantiations
    └── services.go        # Application service implementations mapping
```

## Why `wire/` is separate

The `wire` directory separates business logic setup from the infrastructure daemon. See `wire/README.md` for specific rules: **All configuration parameters MUST be injected into Services, not Handlers.**
