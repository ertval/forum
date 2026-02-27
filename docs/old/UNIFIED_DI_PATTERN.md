# Unified Dependency Injection Pattern

## Overview

This document describes the unified dependency injection pattern implemented across all HTTP handlers in the Forum application.

## Problem Solved

Previously, each handler had a different constructor signature:
- `auth`: `NewHTTPHandler(authService, templates)`
- `user`: `NewHTTPHandler(userService)`
- `post`: `NewHTTPHandler(postService, authService, userService)` + `SetTemplates()` + `SetCategoryService()`
- Others: Various combinations

This inconsistency made the codebase harder to maintain, test, and extend.

## Solution: ServiceContainer Pattern

### Core Concept

All handlers now receive dependencies through a **unified ServiceContainer** with accessor methods:

```go
// In cmd/forum/wire/services.go
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

// Accessor methods
func (sc *ServiceContainer) Auth() authPorts.AuthService { return sc.auth }
func (sc *ServiceContainer) User() userPorts.UserService { return sc.user }
// ... etc for all services
```

### Unified Constructor Signature

**ALL handlers now have identical constructors**:

```go
func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler
```

### Handler-Specific Interfaces

Each handler defines its own minimal `ServiceContainer` interface:

```go
// In auth/adapters/http_handler.go
type ServiceContainer interface {
    Auth() authPorts.AuthService
}

// In post/adapters/http_handler.go
type ServiceContainer interface {
    Post() postPorts.PostService
    Category() postPorts.CategoryService
    Auth() authPorts.AuthService
    User() userPorts.UserService
}
```

This approach provides:
1. **Explicit dependencies**: Each handler declares what it needs
2. **Type safety**: Compiler verifies dependencies are available
3. **Interface satisfaction**: wire.ServiceContainer satisfies all handler interfaces

## Implementation Details

### 1. Wire Package (cmd/forum/wire/services.go)

```go
// ServiceContainer holds all services (only one struct needed!)
type ServiceContainer struct {
    auth         authPorts.AuthService
    user         userPorts.UserService
    // ... all services (lowercase, private)
}

// Accessor methods (public API)
func (sc *ServiceContainer) Auth() authPorts.AuthService { return sc.auth }
func (sc *ServiceContainer) User() userPorts.UserService { return sc.user }
// ... etc

// initServices creates and returns ServiceContainer directly
func initServices(repos *Repositories, sessionDuration time.Duration) *ServiceContainer {
    return &ServiceContainer{
        auth:         authApp.NewService(repos.Session, repos.User, sessionDuration),
        user:         userApp.NewService(repos.User),
        // ... all services
    }
}
```

### 2. Handler Initialization (cmd/forum/wire/handlers.go)

```go
func initHandlers(services *ServiceContainer) *Handlers {
    templates, err := template.ParseGlob("templates/*.html")
    if err != nil {
        panic(err)
    }

    // All handlers use same pattern - directly pass ServiceContainer!
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

### 3. Handler Implementation Example

```go
// In internal/modules/post/adapters/http_handler.go
type HTTPHandler struct {
    postService     postPorts.PostService
    categoryService postPorts.CategoryService
    authService     authPorts.AuthService
    userService     userPorts.UserService
    templates       *template.Template
}

// Local interface declaring dependencies
type ServiceContainer interface {
    Post() postPorts.PostService
    Category() postPorts.CategoryService
    Auth() authPorts.AuthService
    User() userPorts.UserService
}

// Unified constructor signature
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

## Benefits

### 1. Consistency
- All handlers have identical constructor signatures
- Predictable pattern across entire codebase
- Easier onboarding for new developers

### 2. Maintainability
- Adding a new service requires changes in one place (ServiceContainer)
- No need to update individual handler constructors
- Clear separation of concerns

### 3. Testability
- Mock only the services a handler actually uses
- Handler interfaces make dependencies explicit
- No need to mock unused services

### 4. Type Safety
- Compile-time verification of dependencies
- No runtime dependency resolution errors
- Interface satisfaction checked by compiler

### 5. Scalability
- Adding new handlers follows established pattern
- No special cases or exceptions
- Uniform approach scales with codebase

### 6. Explicit Dependencies
- Each handler declares what it needs via interface
- No hidden dependencies or magic
- Easy to see cross-module dependencies

## Adding a New Handler

When creating a new module handler:

1. **Define local ServiceContainer interface** with needed services:
   ```go
   type ServiceContainer interface {
       NewService() ports.Service
       Auth() authPorts.AuthService // If auth needed
   }
   ```

2. **Implement unified constructor**:
   ```go
   func NewHTTPHandler(services ServiceContainer, templates *template.Template) *HTTPHandler {
       return &HTTPHandler{
           newService: services.NewService(),
           templates:  templates,
       }
   }
   ```

3. **Update wire/services.go**:
   - Add service to `ServiceContainer` struct (lowercase field)
   - Add accessor method
   - Update `initServices()` to instantiate new service

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

4. **Update wire/handlers.go**:
   ```go
   NewModule: newmoduleAdapters.NewHTTPHandler(services, templates),
   ```

Done! The pattern is consistent and predictable.

## Testing Example

```go
// Mock container for testing
type mockContainer struct {
    authService *mockAuthService
}

func (m *mockContainer) Auth() authPorts.AuthService { return m.authService }

// Test handler
func TestHandler(t *testing.T) {
    mockAuth := &mockAuthService{/* ... */}
    container := &mockContainer{authService: mockAuth}
    templates := template.Must(template.New("test").Parse(""))
    
    handler := NewHTTPHandler(container, templates)
    
    // Test handler...
}
```

## Migration Notes

All handlers have been updated to use this pattern:
- ✅ `auth/adapters/http_handler.go`
- ✅ `user/adapters/http_handler.go`
- ✅ `post/adapters/http_handler.go`
- ✅ `comment/adapters/http_handler.go`
- ✅ `reaction/adapters/http_handler.go`
- ✅ `moderation/adapters/http_handler.go`
- ✅ `notification/adapters/http_handler.go`

All existing functionality preserved, only constructor signatures changed.

## References

- Implementation: `cmd/forum/wire/services.go`
- Handler examples: `internal/modules/*/adapters/http_handler.go`
- Documentation: `cmd/forum/wire/README.md`
- Architecture: `.github/copilot-instructions.md`

---

**Last Updated**: November 12, 2025
**Status**: ✅ Fully Implemented
