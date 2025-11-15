# Wire Package - Dependency Injection Structure

## Overview

The `wire` package centralizes all dependency injection and application wiring, keeping `main.go` clean and focused on lifecycle management.

## Directory Structure

```text
cmd/forum/
├── main.go              # Minimal entry point (40 lines)
│                        # - Load config
│                        # - Initialize logger
│                        # - Call wire.InitializeApp()
│                        # - Start/shutdown server
│
└── wire/                # Dependency injection package
    ├── app.go           # Main orchestration
    │                    # - App struct (Server, DB, Logger)
    │                    # - InitializeApp() - coordinates all initialization
    │                    # - initDatabase() - DB connection & migrations
    │                    # - initServer() - HTTP server setup
    │                    # - Start(), Shutdown(), Cleanup() methods
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
    └── README.md        # This documentation
```

## Dependency Flow

```text
main.go
   │
   ├─► wire.InitializeApp()
   │      │
   │      ├─► initDatabase() ──────► Database Connection + Migrations
   │      │
   │      ├─► initRepositories() ──► All Repository Instances
   │      │                          (Auth, User, Post, Comment, etc.)
   │      │
   │      ├─► initServices() ──────► ServiceContainer with all services
   │      │                          (Returns unified DI container)
   │      │
   │      ├─► initHandlers() ──────► All HTTP Handler Instances
   │      │                          (All handlers receive ServiceContainer + templates)
   │      │
   │      └─► initServer() ────────► HTTP Server with:
   │                                  - Middleware (Recovery, Logger, CORS, RateLimit)
   │                                  - All routes registered
   │                                  - Static file serving
   │
   └─► Returns *App with Start() and Shutdown() methods
```

## Unified Dependency Injection Pattern

### ServiceContainer

All HTTP handlers now receive dependencies through a **unified ServiceContainer**:

```go
// In wire/services.go
type ServiceContainer struct {
    auth         authPorts.AuthService
    user         userPorts.UserService
    post         postPorts.PostService
    category     postPorts.CategoryService
    comment      commentPorts.CommentService
    reaction     reactionPorts.ReactionService
    moderation   moderationPorts.ModerationService
    notification notificationPorts.NotificationService
}

// Accessor methods for each service
func (sc *ServiceContainer) Auth() authPorts.AuthService { return sc.auth }
func (sc *ServiceContainer) User() userPorts.UserService { return sc.user }
// ... etc
```

### Handler Constructors

**ALL handlers now have the SAME constructor signature**:

```go
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler
```

Each handler defines a **local interface** with only the services it needs:

```go
// In auth/adapters/http_handler.go
type ServiceContainer interface {
    Auth() authPorts.AuthService
}

func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
    return &HTTPHandler{
        authService: services.Auth(),
        templates:   templates,
    }
}
```

```go
// In post/adapters/http_handler.go
type ServiceContainer interface {
    Post() postPorts.PostService
    Category() postPorts.CategoryService
    Auth() authPorts.AuthService
    User() userPorts.UserService
}

func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
    return &HTTPHandler{
        postService:     services.Post(),
        categoryService: services.Category(),
        authService:     services.Auth(),
        userService:     services.User(),
        templates:       templates,
    }
}
```

### Benefits of This Pattern

1. **Uniform Interface**: All handlers have identical constructor signature
2. **Explicit Dependencies**: Each handler declares exactly what it needs via local interface
3. **Type Safety**: Compile-time verification of dependencies
4. **Easy Testing**: Mock only the services a handler actually uses
5. **No Circular Dependencies**: Container holds interfaces, not implementations
6. **Scalability**: Adding new services requires minimal changes
7. **Clean Separation**: Handlers remain independent while sharing common DI mechanism

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
  - `app.go` - Register routes (in initServer function)
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

### Handler Initialization Example

```go
// In wire/handlers.go
func initHandlers(services *ServiceContainer) *Handlers {
    templates, err := template.ParseGlob("templates/*.html")
    if err != nil {
        panic(err)
    }

    // All handlers have the same constructor signature!
    return &Handlers{
        Auth:         authAdapters.NewHTTPHandler(services, templates),
        User:         userAdapters.NewHTTPHandler(services, templates),
        Post:         postAdapters.NewHTTPHandler(services, templates),
        Comment:      commentAdapters.NewHTTPHandler(services, templates),
        Reaction:     reactionAdapters.NewHTTPHandler(services, templates),
        Moderation:   moderationAdapters.NewHTTPHandler(services, templates),
        Notification: notificationAdapters.NewHTTPHandler(services, templates),
    }
}
```

### Adding a New Module

1. **Create module** with hexagonal structure:

   ```text
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

3. **Update wire/services.go** (add to ServiceContainer):

   ```go
   // Add to ServiceContainer struct
   type ServiceContainer struct {
       // ... existing (lowercase fields)
       newmodule newmodulePorts.Service
   }
   
   // Add accessor method
   func (sc *ServiceContainer) NewModule() newmodulePorts.Service {
       return sc.newmodule
   }
   
   // Update initServices
   func initServices(repos *Repositories, sessionDuration time.Duration) *ServiceContainer {
       return &ServiceContainer{
           // ... existing
           newmodule: newmoduleApp.NewService(repos.NewModule),
       }
   }
   ```

4. **Create handler with unified constructor** (adapters/http_handler.go):

   ```go
   type HTTPHandler struct {
       newmoduleService ports.Service
       templates        *template.Template
   }
   
   // Define local interface with only needed services
   type ServiceContainer interface {
       NewModule() ports.Service
       Auth() authPorts.AuthService // If auth is needed
   }
   
   // SAME signature as all other handlers!
   func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
       return &HTTPHandler{
           newmoduleService: services.NewModule(),
           templates:        templates,
       }
   }
   ```

5. **Update wire/handlers.go**:

   ```go
   type Handlers struct {
       // ... existing
       NewModule *newmoduleAdapters.HTTPHandler
   }
   
   func initHandlers(services *ServiceContainer) *Handlers {
       templates, _ := template.ParseGlob("templates/*.html")
       
       return &Handlers{
           // ... existing
           NewModule: newmoduleAdapters.NewHTTPHandler(services, templates),
       }
   }
   ```

6. **Update wire/app.go** (initServer function):

   ```go
   handlers.NewModule.RegisterRoutes(server.Router())
   ```

Done! The new module follows the same unified pattern as all existing modules.

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
