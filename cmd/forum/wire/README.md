# Wire Package - Dependency Injection Structure

## Overview

The `wire` package centralizes all dependency injection and application wiring, keeping `main.go` clean and focused on lifecycle management.

## Directory Structure

```
cmd/forum/
├── main.go              # Minimal entry point (40 lines)
│                        # - Load config
│                        # - Initialize logger
│                        # - Call wire.InitializeApp()
│                        # - Start/shutdown server
│
└── wire/                # Dependency injection package
    ├── wire.go          # Main orchestration
    │                    # - InitializeApp() - coordinates all initialization
    │                    # - initDatabase() - DB connection & migrations
    │                    # - initServer() - HTTP server setup
    │
    ├── repos.go         # Repository initialization (OUTPUT ADAPTERS)
    │                    # - Repositories struct
    │                    # - initRepositories() - creates all SQLite repos
    │
    ├── services.go      # Service initialization (APPLICATION LAYER)
    │                    # - Services struct
    │                    # - initServices() - wires all domain services
    │
    ├── handlers.go      # Handler initialization (INPUT ADAPTERS)
    │                    # - Handlers struct
    │                    # - initHandlers() - creates all HTTP handlers
    │
    └── app.go           # Application struct
                         # - App struct (Server, DB, Logger)
                         # - Start(), Shutdown(), Cleanup() methods
```

## Dependency Flow

```
main.go
   │
   ├─► wire.InitializeApp()
   │      │
   │      ├─► initDatabase() ──────► Database Connection + Migrations
   │      │
   │      ├─► initRepositories() ──► All Repository Instances
   │      │                          (Auth, User, Post, Comment, etc.)
   │      │
   │      ├─► initServices() ──────► All Service Instances
   │      │                          (Auth, User, Post, Comment, etc.)
   │      │
   │      ├─► initHandlers() ──────► All HTTP Handler Instances
   │      │                          (Auth, User, Post, Comment, etc.)
   │      │
   │      └─► initServer() ────────► HTTP Server with:
   │                                  - Middleware (Recovery, Logger, CORS, RateLimit)
   │                                  - All routes registered
   │                                  - Static file serving
   │
   └─► Returns *App with Start() and Shutdown() methods
```

## Key Benefits

### 1. **Separation of Concerns**
- `main.go` - Lifecycle management only (config, start, shutdown)
- `wire/` - All dependency construction and wiring
- Clean, focused responsibilities

### 2. **Testability**
- Can test initialization logic independently
- Can create test apps with different configurations
- Mock dependencies easily in tests

### 3. **Explicitness**
- All dependencies visible in one place
- Clear initialization order
- No hidden magic or reflection

### 4. **Maintainability**
- Adding a new module? Update 4 files in `wire/`:
  - `repos.go` - Add repository
  - `services.go` - Add service
  - `handlers.go` - Add handler
  - `wire.go` - Register routes
- `main.go` never changes

### 5. **Idiomatic Go**
- Manual dependency injection (no frameworks)
- Explicit over implicit
- Simple, readable code

## Usage Example

### main.go (Simplified)

```go
func main() {
    cfg := config.MustLoad()
    lgr := logger.New(logger.InfoLevel, os.Stdout)
    
    // All wiring happens here
    app, err := wire.InitializeApp(cfg, lgr)
    if err != nil {
        lgr.Fatal("Failed to initialize", logger.Error(err))
    }
    defer app.Cleanup()
    
    // Simple lifecycle management
    app.Start()
    // ... graceful shutdown logic
}
```

### Adding a New Module

1. **Create module** with hexagonal structure:
   ```
   internal/modules/newmodule/
   ├── domain/
   ├── ports/
   ├── application/
   └── adapters/
   ```

2. **Update wire/repos.go**:
   ```go
   type Repositories struct {
       // ... existing
       NewModule newmodulePorts.Repository
   }
   
   func initRepositories(db *sql.DB) *Repositories {
       return &Repositories{
           // ... existing
           NewModule: newmoduleAdapters.NewSQLiteRepository(db),
       }
   }
   ```

3. **Update wire/services.go**:
   ```go
   type Services struct {
       // ... existing
       NewModule newmodulePorts.Service
   }
   
   func initServices(repos *Repositories) *Services {
       return &Services{
           // ... existing
           NewModule: newmoduleApp.NewService(repos.NewModule),
       }
   }
   ```

4. **Update wire/handlers.go**:
   ```go
   type Handlers struct {
       // ... existing
       NewModule *newmoduleAdapters.HTTPHandler
   }
   
   func initHandlers(services *Services) *Handlers {
       return &Handlers{
           // ... existing
           NewModule: newmoduleAdapters.NewHTTPHandler(services.NewModule),
       }
   }
   ```

5. **Update wire/wire.go** (initServer function):
   ```go
   handlers.NewModule.RegisterRoutes(server.Router())
   ```

Done! No changes to `main.go` needed.

## Architecture Alignment

This wire package perfectly aligns with the hexagonal architecture:

- **Domain Layer**: Remains pure (no changes)
- **Ports Layer**: Defines contracts (no changes)
- **Application Layer**: Service implementations (no changes)
- **Adapters Layer**: Technical implementations (no changes)
- **Wire Package**: Orchestrates everything together

The wire package sits **outside** the hexagon, connecting all the pieces together while keeping the core business logic clean and dependency-free.

## Comparison with Original main.go

### Before (154 lines in main.go)
- All imports in main.go
- All repository initialization in main.go
- All service initialization in main.go
- All handler initialization in main.go
- All middleware configuration in main.go
- All route registration in main.go

### After (47 lines in main.go + organized wire package)
- `main.go`: Config, logger, app lifecycle only
- `wire/`: All dependency injection logic, organized by concern
- Clear separation, better testing, easier maintenance

## Future Enhancements

### Optional: Google Wire Integration

For even more automation, you could use [google/wire](https://github.com/google/wire):

```go
// wire_gen.go (auto-generated by wire tool)
//go:build wireinject

func InitializeApp(cfg *config.Config, lgr *logger.Logger) (*App, error) {
    wire.Build(
        provideDatabase,
        provideRepositories,
        provideServices,
        provideHandlers,
        provideServer,
    )
    return &App{}, nil
}
```

However, **manual wiring is recommended** for this project because:
- More explicit and easier to understand
- No build-time code generation
- Follows Go's simplicity principle
- Easier for learning and education

---

**Result**: Clean, maintainable, testable, and idiomatic Go dependency injection.
