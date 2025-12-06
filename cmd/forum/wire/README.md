# Wire Package - Dependency Injection

## Overview

The `wire` package centralizes all dependency injection, keeping `main.go` clean and focused on lifecycle management.

## Structure

```text
cmd/forum/
├── main.go              # Entry point (~40 lines)
│                        # - Load config
│                        # - Initialize logger  
│                        # - Call wire.InitializeApp()
│                        # - Start/shutdown server
│
└── wire/                # Dependency injection
    ├── app.go           # Main orchestration
    ├── repos.go         # Repository initialization
    ├── services.go      # Service initialization + config injection
    ├── handlers.go      # Handler initialization
    └── README.md        # This file
```

## Key Principle: Config Injection at Service Level

**ALL configuration values are injected at the SERVICE layer, NOT at handlers.**

### Why Services, Not Handlers?

- **Single Responsibility**: Handlers route HTTP requests; services implement business logic
- **Type Safety**: Services enforce constraints through their interface contracts
- **Testability**: Mock services with different configs without touching handlers
- **Consistency**: All handlers follow the same constructor signature

### Example: Image Upload Config

```go
// ❌ WRONG: Config in handler constructor
func NewHTTPHandler(services ServiceContainer, templates *template.Template, maxImageSize int64)

// ✅ CORRECT: Config in service constructor  
func NewService(repo Repository, imageHandler ImageHandler, maxImageSize int64) *Service

// Handler gets config from service when needed
maxSize := h.postService.MaxImageSize()
```

## Unified Dependency Injection Pattern

### ServiceContainer

All HTTP handlers receive dependencies through a **unified ServiceContainer**:

```go
// In wire/services.go
type ServiceContainer struct {
    auth         authPorts.AuthService
    user         userPorts.UserService
    post         postPorts.PostService
    // ... other services
}

// Accessor methods
func (sc *ServiceContainer) Auth() authPorts.AuthService { return sc.auth }
func (sc *ServiceContainer) Post() postPorts.PostService { return sc.post }
```

### Universal Handler Constructor Signature

**ALL handlers use the SAME signature:**

```go
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler
```

Each handler defines a **local interface** declaring its dependencies:

```go
// In post/adapters/http_handler.go
type ServiceContainer interface {
    Post() postPorts.PostService
    Category() postPorts.CategoryService
    Auth() authPorts.AuthService
    // ... only what this handler needs
}
```

### Benefits

1. **Uniform Interface**: All handlers have identical constructor signature
2. **Config at Right Layer**: Business rules enforced in services, not handlers
3. **Explicit Dependencies**: Each handler declares exactly what it needs
4. **Type Safety**: Compile-time verification
5. **Easy Testing**: Mock only required services
6. **No Config Import in Handlers**: Handlers never import `config` package

## Dependency Flow

```text
main.go
   │
   ├─► Load Config
   ├─► Initialize Logger
   │
   └─► wire.InitializeApp(cfg, logger)
          │
          ├─► Database + Migrations
          │
          ├─► initRepositories(db)
          │      └─► Create all repository instances
          │
          ├─► initServices(repos, cfg, logger)
          │      └─► ✨ CONFIG INJECTED HERE ✨
          │          • Session duration → AuthService
          │          • Max image size → PostService
          │          • Upload directory → ImageHandler
          │
          ├─► initHandlers(services, cfg)
          │      └─► Create handlers with ServiceContainer ONLY
          │          • NO config parameters
          │          • Get config from services when needed
          │
          └─► initServer(cfg, logger, handlers, db)
                 └─► Configure routes + middleware
```

## Adding a New Module

### 1. Create Module Structure

```text
internal/modules/yourmodule/
├── domain/         # Business entities + validation
├── ports/          # Interface contracts
├── application/    # Service implementation
└── adapters/       # HTTP handlers + DB repos
```

### 2. Update wire/repos.go

```go
type Repositories struct {
    // ... existing
    YourModule yourPorts.YourRepository
}

repos.YourModule = yourAdapters.NewRepository(db)
```

### 3. Update wire/services.go

**Inject config at service initialization:**

```go
type ServiceContainer struct {
    // ... existing
    yourModule yourPorts.YourService
}

func (sc *ServiceContainer) YourModule() yourPorts.YourService { 
    return sc.yourModule 
}

func initServices(repos *Repositories, cfg *config.Config, lgr *logger.Logger) *ServiceContainer {
    // Inject config values from cfg.YourModule.*
    yourService := yourApp.NewService(
        repos.YourModule,
        cfg.YourModule.SomeSetting,  // ✅ Config injected here
        cfg.YourModule.AnotherSetting,
    )
    
    return &ServiceContainer{
        // ... existing
        yourModule: yourService,
    }
}
```

### 4. Create Handler (adapters/http_handler.go)

**NO config imports, use unified signature:**

```go
package adapters

// Local interface - declare only what you need
type ServiceContainer interface {
    YourModule() yourPorts.YourService
    Auth() authPorts.AuthService  // if needed
}

type HTTPHandler struct {
    yourService yourPorts.YourService
    authService authPorts.AuthService
    templates   *template.Template
}

// ✅ UNIVERSAL SIGNATURE - no config parameter
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
    return &HTTPHandler{
        yourService: services.YourModule(),
        authService: services.Auth(),
        templates:   templates,
    }
}

// Get config from service when needed
func (h *HTTPHandler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    maxSize := h.yourService.GetConfigValue()  // ✅ From service
    // ... use it
}
```

### 5. Update wire/handlers.go

```go
type Handlers struct {
    // ... existing
    YourModule *yourAdapters.HTTPHandler
}

func initHandlers(services *ServiceContainer, cfg *config.Config) *Handlers {
    templates, _ := template.ParseGlob("templates/*.html")
    
    return &Handlers{
        // ... existing
        YourModule: yourAdapters.NewHTTPHandler(services, templates),  // ✅ No config
    }
}
```

### 6. Update wire/app.go

```go
func initServer(...) *httpserver.Server {
    // ... middleware setup
    
    handlers.YourModule.RegisterRoutes(server.Router())
    
    // ... rest of setup
}
```

Done! The module follows the unified pattern with config at service level.

## Architecture Alignment

- **Domain Layer**: Pure business logic (no changes)
- **Ports Layer**: Interface contracts (add MaxSomething() methods if needed)
- **Application Layer**: Services receive config in constructor
- **Adapters Layer**: Handlers get config from services via interface methods
- **Wire Package**: Orchestrates everything, injects config at service initialization

## Key Takeaway

**Config flows: main.go → wire/services.go → Service constructors → Service methods**

Handlers are config-agnostic and depend only on service interfaces.
