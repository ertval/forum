// Package wire provides dependency injection and application wiring.
//
// This package is the composition root of the forum application. It initializes
// all components (repositories, services, handlers) and wires them together
// following the Ports & Adapters (Hexagonal) architecture pattern.
//
// Structure:
//   - app.go: Core App type, lifecycle (Start/Stop), and high-level orchestration
//   - repositories.go: Initializes database adapters (output adapters)
//   - services.go: Wires application services with their dependencies
//   - handlers.go: Injects services into HTTP handlers (input adapters)
//
// Usage:
//
//	app, err := wire.InitializeApp(cfg, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer app.Cleanup()
//
//	if err := app.Start(); err != nil {
//	    log.Fatal(err)
//	}
package wire
