# HTTP Handler Abstraction Refactor Guide

## Overview

This document outlines the refactoring plan to abstract HTTP framework specifics from module handlers into the platform layer. This approach maintains hexagonal architecture principles while enabling framework independence and reducing boilerplate code.

## Goals

1. **Framework Independence**: Switch HTTP frameworks (stdlib → Chi → Gin) without changing module code
2. **Module Autonomy**: Modules define their own API endpoints and business logic routing
3. **Reduced Boilerplate**: Abstract common HTTP tasks (JSON parsing, error responses, parameter extraction)
4. **Easy Testing**: Mock HTTP abstractions for unit testing handlers
5. **Hexagonal Compliance**: HTTP framework is infrastructure (platform), API design is domain concern (module)
6. **Consistent API**: Standardize request/response patterns across all modules

## Current State

```
Module HTTP Handlers (Currently):
├── internal/modules/auth/adapters/
│   └── http_handler.go               ← Direct http.ResponseWriter, *http.Request usage
├── internal/modules/user/adapters/
│   └── http_handler.go               ← Direct stdlib HTTP usage
├── internal/modules/post/adapters/
│   └── http_handler.go               ← Direct stdlib HTTP usage
└── ... (other modules)

Platform HTTP Server (Currently):
├── internal/platform/httpserver/
│   ├── server.go                     ← Wraps http.ServeMux
│   └── middleware.go                 ← Middleware functions
```

**Current Handler Example**:

```go
// Current: Direct dependency on net/http
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Manual parameter extraction
    userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Call service
    user, err := h.service.GetByID(r.Context(), userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Manual JSON encoding
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(user)
}
```

**Problems with Current Approach**:
- ❌ Tied to stdlib `http.ResponseWriter` and `*http.Request`
- ❌ Repetitive error handling code
- ❌ Manual JSON encoding/decoding
- ❌ Manual parameter extraction and validation
- ❌ Inconsistent error response formats
- ❌ Difficult to test without HTTP test servers
- ❌ Cannot switch to alternative frameworks (Chi, Gin, Echo)

## Target State

```
Module HTTP Handlers (After Refactor):
├── internal/modules/auth/adapters/
│   └── http_handler.go               ← Uses httpserver.Router, httpserver.Context
├── internal/modules/user/adapters/
│   └── http_handler.go               ← Framework-agnostic
├── internal/modules/post/adapters/
│   └── http_handler.go               ← Framework-agnostic
└── ... (other modules)

Platform HTTP Server (After Refactor):
├── internal/platform/httpserver/
│   ├── router.go                     ← Router interface
│   ├── context.go                    ← Context interface (request/response)
│   ├── response.go                   ← Standard response helpers
│   ├── stdlib_router.go              ← Standard library implementation
│   ├── chi_router.go                 ← Chi framework implementation (optional)
│   ├── gin_router.go                 ← Gin framework implementation (optional)
│   ├── server.go                     ← Server wrapper
│   └── middleware.go                 ← Middleware abstraction
```

**Target Handler Example**:

```go
// After: Framework-agnostic using platform abstractions
func (h *Handler) GetUser(ctx httpserver.Context) error {
    // Automatic parameter extraction with validation
    userID, err := ctx.ParamInt64("id")
    if err != nil {
        return ctx.BadRequest("Invalid user ID")
    }
    
    // Call service
    user, err := h.service.GetByID(ctx.Request().Context(), userID)
    if err != nil {
        return ctx.InternalError(err)
    }
    
    // Automatic JSON response
    return ctx.JSON(http.StatusOK, user)
}
```

**Benefits**:

- ✅ No direct dependency on net/http
- ✅ Clean, concise handler code
- ✅ Consistent error handling
- ✅ Automatic JSON encoding/decoding
- ✅ Type-safe parameter extraction
- ✅ Easy to mock for testing
- ✅ Can switch HTTP frameworks easily

---

## Architecture Principles

### Hexagonal Architecture Compliance

```
┌─────────────────────────────────────────────────────────────┐
│                    Module Layer (Adapters)                   │
│                                                              │
│  HTTP Handlers (INPUT ADAPTERS)                             │
│  ├── Define API endpoints                                   │
│  ├── Route requests to service methods                      │
│  ├── Validate domain-specific input                         │
│  └── Format domain-specific responses                       │
│                                                              │
│  Depend on: httpserver.Router, httpserver.Context           │
│              (platform interfaces)                           │
└──────────────────────────────────────────────────────────────┘
                                  │
                                  │ Uses abstractions
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Platform Layer                            │
│                                                              │
│  HTTP Server Abstractions                                   │
│  ├── Define Router interface (route registration)           │
│  ├── Define Context interface (request/response)            │
│  ├── Provide stdlib implementation                          │
│  ├── Provide alternative implementations (Chi, Gin, etc.)   │
│  └── Handle HTTP protocol specifics                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **Modules Own API Design**: Each module decides what endpoints to expose and how to map them to services
2. **Platform Owns HTTP Protocol**: Platform handles HTTP-specific concerns (headers, status codes, encoding)
3. **Interface-Based**: Modules depend on `Router` and `Context` interfaces, not concrete implementations
4. **Handler Signature**: Use `func(httpserver.Context) error` instead of `func(http.ResponseWriter, *http.Request)`
5. **Error Propagation**: Handlers return errors, platform converts them to HTTP responses
6. **No Framework Lock-in**: Can swap Chi/Gin/Echo/stdlib without touching module code

### What Belongs Where

**Platform Layer Responsibilities**:

- ✅ Route registration and matching
- ✅ HTTP method enforcement
- ✅ Request parsing (JSON, form data, files)
- ✅ Response encoding (JSON, XML, plain text)
- ✅ Parameter extraction (path, query, headers)
- ✅ Error response formatting
- ✅ Middleware management
- ✅ Content negotiation

**Module Layer Responsibilities**:

- ✅ Endpoint definitions (URL patterns)
- ✅ Business logic routing (service method calls)
- ✅ Domain-specific validation
- ✅ Domain-specific error handling
- ✅ Response structure (what data to return)
- ✅ Authorization checks (via middleware or service)

---

## Technical Specification

### 1. Router Interface (router.go)

The Router interface abstracts route registration and HTTP method routing.

```go
package httpserver

import (
    "net/http"
)

// Router provides framework-agnostic route registration.
// Implementations can use stdlib ServeMux, Chi, Gin, Echo, etc.
type Router interface {
    // Handle registers a handler for a specific HTTP method and pattern.
    // Example: Handle("GET", "/api/users/{id}", handler)
    Handle(method, pattern string, handler HandlerFunc)
    
    // GET is a shorthand for Handle("GET", pattern, handler)
    GET(pattern string, handler HandlerFunc)
    
    // POST is a shorthand for Handle("POST", pattern, handler)
    POST(pattern string, handler HandlerFunc)
    
    // PUT is a shorthand for Handle("PUT", pattern, handler)
    PUT(pattern string, handler HandlerFunc)
    
    // PATCH is a shorthand for Handle("PATCH", pattern, handler)
    PATCH(pattern string, handler HandlerFunc)
    
    // DELETE is a shorthand for Handle("DELETE", pattern, handler)
    DELETE(pattern string, handler HandlerFunc)
    
    // Group creates a route group with a common prefix.
    // Example: Group("/api/v1").GET("/users", handler)
    Group(prefix string) Router
    
    // Use attaches middleware to this router or route group.
    // Middleware applies to all routes registered after this call.
    Use(middleware ...MiddlewareFunc)
    
    // ServeHTTP exposes the underlying HTTP handler for the server.
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// HandlerFunc is the signature for HTTP handlers using the abstraction.
// Handlers receive a Context and return an error.
// The error is automatically converted to an HTTP response by the platform.
type HandlerFunc func(Context) error

// MiddlewareFunc wraps a HandlerFunc with additional behavior.
type MiddlewareFunc func(HandlerFunc) HandlerFunc
```

**Design Rationale**:

- **Method-specific shortcuts**: `GET()`, `POST()`, etc. for cleaner code
- **Route grouping**: Support for API versioning and prefixes
- **Middleware chaining**: Apply middleware at router or group level
- **Error-based handlers**: Handlers return errors instead of managing responses directly
- **ServeHTTP compatibility**: Can still integrate with stdlib `http.Server`

### 2. Context Interface (context.go)

The Context interface abstracts request/response operations.

```go
package httpserver

import (
    "context"
    "io"
    "mime/multipart"
    "net/http"
    "net/url"
)

// Context provides a framework-agnostic interface for HTTP request/response handling.
// It wraps the underlying HTTP request and response objects.
type Context interface {
    // Request returns the underlying HTTP request.
    // Use this to access the Go context, headers, or other raw data.
    Request() *http.Request
    
    // Context returns the request context for cancellation and values.
    Context() context.Context
    
    // SetContext sets a new context on the request (for middleware).
    SetContext(ctx context.Context)
    
    // --- Parameter Extraction ---
    
    // Param returns a path parameter by name.
    // Example: "/users/{id}" → ctx.Param("id")
    Param(name string) string
    
    // ParamInt returns a path parameter as an integer.
    ParamInt(name string) (int, error)
    
    // ParamInt64 returns a path parameter as an int64.
    ParamInt64(name string) (int64, error)
    
    // Query returns a query parameter by name.
    // Example: "/users?page=2" → ctx.Query("page")
    Query(name string) string
    
    // QueryDefault returns a query parameter or a default value if not present.
    QueryDefault(name, defaultValue string) string
    
    // QueryInt returns a query parameter as an integer.
    QueryInt(name string) (int, error)
    
    // QueryInt64 returns a query parameter as an int64.
    QueryInt64(name string) (int64, error)
    
    // Header returns a request header by name.
    Header(name string) string
    
    // Cookie returns a cookie by name.
    Cookie(name string) (*http.Cookie, error)
    
    // --- Request Body Parsing ---
    
    // Bind parses the request body into the provided struct.
    // Supports JSON, form data, and XML based on Content-Type.
    Bind(v interface{}) error
    
    // BindJSON parses JSON request body into the provided struct.
    BindJSON(v interface{}) error
    
    // BindForm parses form data into the provided struct.
    BindForm(v interface{}) error
    
    // FormValue returns a form value by name.
    FormValue(name string) string
    
    // FormFile returns a file from a multipart form.
    FormFile(name string) (*multipart.FileHeader, error)
    
    // Body returns the raw request body reader.
    Body() io.ReadCloser
    
    // --- Response Methods ---
    
    // JSON sends a JSON response with the given status code.
    JSON(code int, v interface{}) error
    
    // String sends a plain text response.
    String(code int, s string) error
    
    // HTML sends an HTML response.
    HTML(code int, html string) error
    
    // Data sends raw bytes with a content type.
    Data(code int, contentType string, data []byte) error
    
    // NoContent sends a 204 No Content response.
    NoContent() error
    
    // Redirect sends a redirect response.
    Redirect(code int, url string) error
    
    // --- Error Response Helpers ---
    
    // BadRequest sends a 400 Bad Request with a message.
    BadRequest(message string) error
    
    // Unauthorized sends a 401 Unauthorized with a message.
    Unauthorized(message string) error
    
    // Forbidden sends a 403 Forbidden with a message.
    Forbidden(message string) error
    
    // NotFound sends a 404 Not Found with a message.
    NotFound(message string) error
    
    // Conflict sends a 409 Conflict with a message.
    Conflict(message string) error
    
    // InternalError sends a 500 Internal Server Error.
    // The actual error is logged but not exposed to the client.
    InternalError(err error) error
    
    // Error sends a custom error response with status code and message.
    Error(code int, message string) error
    
    // --- Response Manipulation ---
    
    // SetHeader sets a response header.
    SetHeader(name, value string)
    
    // SetCookie sets a response cookie.
    SetCookie(cookie *http.Cookie)
    
    // Status sets the HTTP response status code.
    Status(code int)
    
    // --- Context Values ---
    
    // Set stores a value in the request context.
    Set(key string, value interface{})
    
    // Get retrieves a value from the request context.
    Get(key string) interface{}
    
    // GetString retrieves a string value from context.
    GetString(key string) string
    
    // GetInt retrieves an int value from context.
    GetInt(key string) int
}
```

**Design Rationale**:

- **Rich API**: Covers common HTTP tasks without requiring stdlib imports
- **Type-safe extraction**: Helpers for int/int64 conversion with error handling
- **Automatic encoding**: JSON/HTML/String responses with proper content types
- **Error helpers**: Consistent error response format across modules
- **Context integration**: Works with Go's context.Context for cancellation
- **Cookie/header access**: Common operations abstracted away

### 3. Standard Library Implementation (stdlib_router.go)

Implementation using Go's standard library `http.ServeMux`.

```go
package httpserver

import (
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "strconv"
    "strings"
)

// StdlibRouter implements Router using standard library http.ServeMux.
type StdlibRouter struct {
    mux         *http.ServeMux
    prefix      string
    middlewares []MiddlewareFunc
}

// NewStdlibRouter creates a new router using stdlib ServeMux.
func NewStdlibRouter() Router {
    return &StdlibRouter{
        mux:         http.NewServeMux(),
        prefix:      "",
        middlewares: []MiddlewareFunc{},
    }
}

// Handle registers a handler for a specific method and pattern.
func (r *StdlibRouter) Handle(method, pattern string, handler HandlerFunc) {
    // Apply middleware chain
    finalHandler := handler
    for i := len(r.middlewares) - 1; i >= 0; i-- {
        finalHandler = r.middlewares[i](finalHandler)
    }
    
    // Convert to stdlib handler
    httpHandler := func(w http.ResponseWriter, req *http.Request) {
        ctx := NewStdlibContext(w, req)
        if err := finalHandler(ctx); err != nil {
            // Error already handled by Context methods
            // or needs default handling
            if !ctx.(*stdlibContext).written {
                ctx.InternalError(err)
            }
        }
    }
    
    // Register with method prefix (stdlib ServeMux pattern)
    fullPattern := method + " " + r.prefix + pattern
    r.mux.HandleFunc(fullPattern, httpHandler)
}

// GET registers a GET handler.
func (r *StdlibRouter) GET(pattern string, handler HandlerFunc) {
    r.Handle("GET", pattern, handler)
}

// POST registers a POST handler.
func (r *StdlibRouter) POST(pattern string, handler HandlerFunc) {
    r.Handle("POST", pattern, handler)
}

// PUT registers a PUT handler.
func (r *StdlibRouter) PUT(pattern string, handler HandlerFunc) {
    r.Handle("PUT", pattern, handler)
}

// PATCH registers a PATCH handler.
func (r *StdlibRouter) PATCH(pattern string, handler HandlerFunc) {
    r.Handle("PATCH", pattern, handler)
}

// DELETE registers a DELETE handler.
func (r *StdlibRouter) DELETE(pattern string, handler HandlerFunc) {
    r.Handle("DELETE", pattern, handler)
}

// Group creates a route group with a prefix.
func (r *StdlibRouter) Group(prefix string) Router {
    return &StdlibRouter{
        mux:         r.mux,
        prefix:      r.prefix + prefix,
        middlewares: append([]MiddlewareFunc{}, r.middlewares...),
    }
}

// Use adds middleware to the router.
func (r *StdlibRouter) Use(middleware ...MiddlewareFunc) {
    r.middlewares = append(r.middlewares, middleware...)
}

// ServeHTTP implements http.Handler.
func (r *StdlibRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.mux.ServeHTTP(w, req)
}

// --- Context Implementation ---

// stdlibContext implements Context using stdlib request/response.
type stdlibContext struct {
    writer  http.ResponseWriter
    request *http.Request
    written bool
}

// NewStdlibContext creates a new context from stdlib request/response.
func NewStdlibContext(w http.ResponseWriter, r *http.Request) Context {
    return &stdlibContext{
        writer:  w,
        request: r,
        written: false,
    }
}

// Request returns the underlying HTTP request.
func (c *stdlibContext) Request() *http.Request {
    return c.request
}

// Context returns the request context.
func (c *stdlibContext) Context() context.Context {
    return c.request.Context()
}

// SetContext sets a new context on the request.
func (c *stdlibContext) SetContext(ctx context.Context) {
    c.request = c.request.WithContext(ctx)
}

// Param returns a path parameter.
func (c *stdlibContext) Param(name string) string {
    return c.request.PathValue(name)
}

// ParamInt returns a path parameter as integer.
func (c *stdlibContext) ParamInt(name string) (int, error) {
    v := c.Param(name)
    if v == "" {
        return 0, fmt.Errorf("parameter %s not found", name)
    }
    return strconv.Atoi(v)
}

// ParamInt64 returns a path parameter as int64.
func (c *stdlibContext) ParamInt64(name string) (int64, error) {
    v := c.Param(name)
    if v == "" {
        return 0, fmt.Errorf("parameter %s not found", name)
    }
    return strconv.ParseInt(v, 10, 64)
}

// Query returns a query parameter.
func (c *stdlibContext) Query(name string) string {
    return c.request.URL.Query().Get(name)
}

// QueryDefault returns a query parameter with default.
func (c *stdlibContext) QueryDefault(name, defaultValue string) string {
    if v := c.Query(name); v != "" {
        return v
    }
    return defaultValue
}

// QueryInt returns a query parameter as integer.
func (c *stdlibContext) QueryInt(name string) (int, error) {
    v := c.Query(name)
    if v == "" {
        return 0, fmt.Errorf("query parameter %s not found", name)
    }
    return strconv.Atoi(v)
}

// QueryInt64 returns a query parameter as int64.
func (c *stdlibContext) QueryInt64(name string) (int64, error) {
    v := c.Query(name)
    if v == "" {
        return 0, fmt.Errorf("query parameter %s not found", name)
    }
    return strconv.ParseInt(v, 10, 64)
}

// Header returns a request header.
func (c *stdlibContext) Header(name string) string {
    return c.request.Header.Get(name)
}

// Cookie returns a cookie.
func (c *stdlibContext) Cookie(name string) (*http.Cookie, error) {
    return c.request.Cookie(name)
}

// BindJSON parses JSON request body.
func (c *stdlibContext) BindJSON(v interface{}) error {
    decoder := json.NewDecoder(c.request.Body)
    return decoder.Decode(v)
}

// Bind parses request body based on Content-Type.
func (c *stdlibContext) Bind(v interface{}) error {
    contentType := c.Header("Content-Type")
    if strings.Contains(contentType, "application/json") {
        return c.BindJSON(v)
    }
    // Add form binding support here if needed
    return fmt.Errorf("unsupported content type: %s", contentType)
}

// BindForm parses form data.
func (c *stdlibContext) BindForm(v interface{}) error {
    if err := c.request.ParseForm(); err != nil {
        return err
    }
    // Use reflection to map form values to struct
    // Implementation omitted for brevity
    return nil
}

// FormValue returns a form value.
func (c *stdlibContext) FormValue(name string) string {
    return c.request.FormValue(name)
}

// FormFile returns a file from multipart form.
func (c *stdlibContext) FormFile(name string) (*multipart.FileHeader, error) {
    _, fileHeader, err := c.request.FormFile(name)
    return fileHeader, err
}

// Body returns the request body.
func (c *stdlibContext) Body() io.ReadCloser {
    return c.request.Body
}

// JSON sends a JSON response.
func (c *stdlibContext) JSON(code int, v interface{}) error {
    c.writer.Header().Set("Content-Type", "application/json")
    c.writer.WriteHeader(code)
    c.written = true
    return json.NewEncoder(c.writer).Encode(v)
}

// String sends a plain text response.
func (c *stdlibContext) String(code int, s string) error {
    c.writer.Header().Set("Content-Type", "text/plain")
    c.writer.WriteHeader(code)
    c.written = true
    _, err := c.writer.Write([]byte(s))
    return err
}

// HTML sends an HTML response.
func (c *stdlibContext) HTML(code int, html string) error {
    c.writer.Header().Set("Content-Type", "text/html")
    c.writer.WriteHeader(code)
    c.written = true
    _, err := c.writer.Write([]byte(html))
    return err
}

// Data sends raw bytes.
func (c *stdlibContext) Data(code int, contentType string, data []byte) error {
    c.writer.Header().Set("Content-Type", contentType)
    c.writer.WriteHeader(code)
    c.written = true
    _, err := c.writer.Write(data)
    return err
}

// NoContent sends 204 No Content.
func (c *stdlibContext) NoContent() error {
    c.writer.WriteHeader(http.StatusNoContent)
    c.written = true
    return nil
}

// Redirect sends a redirect response.
func (c *stdlibContext) Redirect(code int, url string) error {
    http.Redirect(c.writer, c.request, url, code)
    c.written = true
    return nil
}

// BadRequest sends 400 with message.
func (c *stdlibContext) BadRequest(message string) error {
    return c.JSON(http.StatusBadRequest, map[string]string{"error": message})
}

// Unauthorized sends 401 with message.
func (c *stdlibContext) Unauthorized(message string) error {
    return c.JSON(http.StatusUnauthorized, map[string]string{"error": message})
}

// Forbidden sends 403 with message.
func (c *stdlibContext) Forbidden(message string) error {
    return c.JSON(http.StatusForbidden, map[string]string{"error": message})
}

// NotFound sends 404 with message.
func (c *stdlibContext) NotFound(message string) error {
    return c.JSON(http.StatusNotFound, map[string]string{"error": message})
}

// Conflict sends 409 with message.
func (c *stdlibContext) Conflict(message string) error {
    return c.JSON(http.StatusConflict, map[string]string{"error": message})
}

// InternalError sends 500 and logs the error.
func (c *stdlibContext) InternalError(err error) error {
    // Log the error (would use platform logger here)
    return c.JSON(http.StatusInternalServerError, map[string]string{
        "error": "Internal server error",
    })
}

// Error sends custom error response.
func (c *stdlibContext) Error(code int, message string) error {
    return c.JSON(code, map[string]string{"error": message})
}

// SetHeader sets a response header.
func (c *stdlibContext) SetHeader(name, value string) {
    c.writer.Header().Set(name, value)
}

// SetCookie sets a response cookie.
func (c *stdlibContext) SetCookie(cookie *http.Cookie) {
    http.SetCookie(c.writer, cookie)
}

// Status sets the response status code.
func (c *stdlibContext) Status(code int) {
    c.writer.WriteHeader(code)
    c.written = true
}

// Set stores a value in request context.
func (c *stdlibContext) Set(key string, value interface{}) {
    ctx := context.WithValue(c.request.Context(), key, value)
    c.request = c.request.WithContext(ctx)
}

// Get retrieves a value from request context.
func (c *stdlibContext) Get(key string) interface{} {
    return c.request.Context().Value(key)
}

// GetString retrieves a string from context.
func (c *stdlibContext) GetString(key string) string {
    if v := c.Get(key); v != nil {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

// GetInt retrieves an int from context.
func (c *stdlibContext) GetInt(key string) int {
    if v := c.Get(key); v != nil {
        if i, ok := v.(int); ok {
            return i
        }
    }
    return 0
}
```

**Key Features**:

- ✅ Uses Go 1.22+ `http.ServeMux` with method and path parameter support
- ✅ Middleware chaining at router and group level
- ✅ Consistent error response format (JSON with `{"error": "message"}`)
- ✅ Type-safe parameter extraction with error handling
- ✅ Automatic content-type setting
- ✅ Response written tracking to avoid double writes

---

## Module Handler Refactoring

### Before and After: User Module Handler

#### Before (Current Implementation)

**File**: `internal/modules/user/adapters/http_handler.go`

```go
package adapters

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "forum/internal/modules/user/ports"
    "forum/internal/platform/logger"
)

type Handler struct {
    service ports.UserService
    logger  *logger.Logger
}

func NewHTTPHandler(service ports.UserService, logger *logger.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

// RegisterRoutes registers all user endpoints.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /api/users/{id}", h.GetUser)
    mux.HandleFunc("PUT /api/users/{id}", h.UpdateUser)
    mux.HandleFunc("GET /api/users/{id}/activity", h.GetActivity)
}

// GetUser handles GET /api/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Extract and validate user ID
    userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        h.logger.Error("Invalid user ID", logger.Error(err))
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Call service
    user, err := h.service.GetByID(r.Context(), userID)
    if err != nil {
        h.logger.Error("Failed to get user", logger.Error(err))
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    // Encode response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(user); err != nil {
        h.logger.Error("Failed to encode response", logger.Error(err))
    }
}

// UpdateUser handles PUT /api/users/{id}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    // Extract and validate user ID
    userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        h.logger.Error("Invalid user ID", logger.Error(err))
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Parse request body
    var updateData struct {
        Username string `json:"username"`
        Bio      string `json:"bio"`
    }
    if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
        h.logger.Error("Invalid request body", logger.Error(err))
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Call service
    user, err := h.service.Update(r.Context(), userID, updateData)
    if err != nil {
        h.logger.Error("Failed to update user", logger.Error(err))
        http.Error(w, "Failed to update user", http.StatusInternalServerError)
        return
    }
    
    // Encode response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(user); err != nil {
        h.logger.Error("Failed to encode response", logger.Error(err))
    }
}

// GetActivity handles GET /api/users/{id}/activity
func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
    // Extract and validate user ID
    userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        h.logger.Error("Invalid user ID", logger.Error(err))
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Call service
    activity, err := h.service.GetActivity(r.Context(), userID)
    if err != nil {
        h.logger.Error("Failed to get activity", logger.Error(err))
        http.Error(w, "Failed to get activity", http.StatusInternalServerError)
        return
    }
    
    // Encode response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(activity); err != nil {
        h.logger.Error("Failed to encode response", logger.Error(err))
    }
}
```

**Problems**:
- ❌ Repetitive error handling code (80+ lines for 3 handlers)
- ❌ Manual JSON encoding/decoding everywhere
- ❌ Inconsistent error responses (some use http.Error, some return JSON)
- ❌ Direct dependency on `http.ResponseWriter` and `*http.Request`
- ❌ Cannot easily mock for testing
- ❌ Logger calls scattered throughout

#### After (Refactored with Abstraction)

**File**: `internal/modules/user/adapters/http_handler.go`

```go
package adapters

import (
    "forum/internal/modules/user/ports"
    "forum/internal/platform/httpserver"
    "forum/internal/platform/logger"
)

type Handler struct {
    service ports.UserService
    logger  *logger.Logger
}

func NewHTTPHandler(service ports.UserService, logger *logger.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

// RegisterRoutes registers all user endpoints.
func (h *Handler) RegisterRoutes(router httpserver.Router) {
    api := router.Group("/api/users")
    api.GET("/{id}", h.GetUser)
    api.PUT("/{id}", h.UpdateUser)
    api.GET("/{id}/activity", h.GetActivity)
}

// GetUser handles GET /api/users/{id}
func (h *Handler) GetUser(ctx httpserver.Context) error {
    // Extract and validate user ID
    userID, err := ctx.ParamInt64("id")
    if err != nil {
        return ctx.BadRequest("Invalid user ID")
    }
    
    // Call service
    user, err := h.service.GetByID(ctx.Context(), userID)
    if err != nil {
        h.logger.Error("Failed to get user", logger.Error(err))
        return ctx.NotFound("User not found")
    }
    
    // Return JSON response
    return ctx.JSON(200, user)
}

// UpdateUser handles PUT /api/users/{id}
func (h *Handler) UpdateUser(ctx httpserver.Context) error {
    // Extract and validate user ID
    userID, err := ctx.ParamInt64("id")
    if err != nil {
        return ctx.BadRequest("Invalid user ID")
    }
    
    // Parse request body
    var updateData struct {
        Username string `json:"username"`
        Bio      string `json:"bio"`
    }
    if err := ctx.BindJSON(&updateData); err != nil {
        return ctx.BadRequest("Invalid request body")
    }
    
    // Call service
    user, err := h.service.Update(ctx.Context(), userID, updateData)
    if err != nil {
        h.logger.Error("Failed to update user", logger.Error(err))
        return ctx.InternalError(err)
    }
    
    // Return JSON response
    return ctx.JSON(200, user)
}

// GetActivity handles GET /api/users/{id}/activity
func (h *Handler) GetActivity(ctx httpserver.Context) error {
    // Extract and validate user ID
    userID, err := ctx.ParamInt64("id")
    if err != nil {
        return ctx.BadRequest("Invalid user ID")
    }
    
    // Call service
    activity, err := h.service.GetActivity(ctx.Context(), userID)
    if err != nil {
        h.logger.Error("Failed to get activity", logger.Error(err))
        return ctx.InternalError(err)
    }
    
    // Return JSON response
    return ctx.JSON(200, activity)
}
```

**Improvements**:
- ✅ 40+ lines shorter (90 → 48 lines for handlers)
- ✅ Consistent error handling using context methods
- ✅ Automatic JSON encoding/decoding
- ✅ No direct stdlib dependencies
- ✅ Easy to mock `httpserver.Context` for testing
- ✅ Cleaner route registration with grouping
- ✅ More readable handler code

**Line Count Comparison**:
- **Before**: ~130 lines total (including RegisterRoutes)
- **After**: ~75 lines total
- **Reduction**: 42% fewer lines, same functionality

### Before and After: Auth Module Handler

#### Before (Current Implementation)

```go
// File: internal/modules/auth/adapters/http_handler.go
package adapters

import (
    "encoding/json"
    "net/http"
    
    "forum/internal/modules/auth/ports"
    "forum/internal/platform/logger"
)

type Handler struct {
    service ports.AuthService
    logger  *logger.Logger
}

func NewHTTPHandler(service ports.AuthService, logger *logger.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("POST /api/auth/register", h.Register)
    mux.HandleFunc("POST /api/auth/login", h.Login)
    mux.HandleFunc("POST /api/auth/logout", h.Logout)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("Invalid request", logger.Error(err))
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    user, err := h.service.Register(r.Context(), req.Username, req.Email, req.Password)
    if err != nil {
        h.logger.Error("Registration failed", logger.Error(err))
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("Invalid request", logger.Error(err))
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    session, err := h.service.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        h.logger.Error("Login failed", logger.Error(err))
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    // Set session cookie
    cookie := &http.Cookie{
        Name:     "session_token",
        Value:    session.Token,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    http.SetCookie(w, cookie)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_token")
    if err != nil {
        http.Error(w, "Not authenticated", http.StatusUnauthorized)
        return
    }
    
    if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
        h.logger.Error("Logout failed", logger.Error(err))
        http.Error(w, "Logout failed", http.StatusInternalServerError)
        return
    }
    
    // Clear cookie
    http.SetCookie(w, &http.Cookie{
        Name:   "session_token",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    })
    
    w.WriteHeader(http.StatusNoContent)
}
```

#### After (Refactored with Abstraction)

```go
// File: internal/modules/auth/adapters/http_handler.go
package adapters

import (
    "net/http"
    
    "forum/internal/modules/auth/ports"
    "forum/internal/platform/httpserver"
    "forum/internal/platform/logger"
)

type Handler struct {
    service ports.AuthService
    logger  *logger.Logger
}

func NewHTTPHandler(service ports.AuthService, logger *logger.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

func (h *Handler) RegisterRoutes(router httpserver.Router) {
    auth := router.Group("/api/auth")
    auth.POST("/register", h.Register)
    auth.POST("/login", h.Login)
    auth.POST("/logout", h.Logout)
}

func (h *Handler) Register(ctx httpserver.Context) error {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := ctx.BindJSON(&req); err != nil {
        return ctx.BadRequest("Invalid request body")
    }
    
    user, err := h.service.Register(ctx.Context(), req.Username, req.Email, req.Password)
    if err != nil {
        h.logger.Error("Registration failed", logger.Error(err))
        return ctx.BadRequest(err.Error())
    }
    
    return ctx.JSON(http.StatusCreated, user)
}

func (h *Handler) Login(ctx httpserver.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := ctx.BindJSON(&req); err != nil {
        return ctx.BadRequest("Invalid request body")
    }
    
    session, err := h.service.Login(ctx.Context(), req.Email, req.Password)
    if err != nil {
        h.logger.Error("Login failed", logger.Error(err))
        return ctx.Unauthorized("Invalid credentials")
    }
    
    // Set session cookie
    ctx.SetCookie(&http.Cookie{
        Name:     "session_token",
        Value:    session.Token,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    })
    
    return ctx.JSON(http.StatusOK, map[string]string{"message": "Login successful"})
}

func (h *Handler) Logout(ctx httpserver.Context) error {
    cookie, err := ctx.Cookie("session_token")
    if err != nil {
        return ctx.Unauthorized("Not authenticated")
    }
    
    if err := h.service.Logout(ctx.Context(), cookie.Value); err != nil {
        h.logger.Error("Logout failed", logger.Error(err))
        return ctx.InternalError(err)
    }
    
    // Clear cookie
    ctx.SetCookie(&http.Cookie{
        Name:   "session_token",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    })
    
    return ctx.NoContent()
}
```

**Improvements**:

- ✅ 50+ lines shorter (120 → 70 lines)
- ✅ Cleaner cookie handling with context methods
- ✅ Consistent error responses
- ✅ Simpler route grouping
- ✅ Still explicit about cookie operations (security-critical)

---

## Wiring and Configuration

### Update Server Initialization (wire/app.go)

**Before**:

```go
package wire

import (
    "net/http"
    
    "forum/internal/platform/config"
    "forum/internal/platform/logger"
)

func NewHTTPServer(cfg *config.Config, handlers *Handlers, lgr *logger.Logger) *http.Server {
    mux := http.NewServeMux()
    
    // Register module routes
    handlers.Auth.RegisterRoutes(mux)
    handlers.User.RegisterRoutes(mux)
    handlers.Post.RegisterRoutes(mux)
    // ... other handlers
    
    server := &http.Server{
        Addr:         cfg.Server.Address,
        Handler:      mux,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }
    
    return server
}
```

**After**:

```go
package wire

import (
    "net/http"
    
    "forum/internal/platform/config"
    "forum/internal/platform/httpserver"
    "forum/internal/platform/logger"
)

func NewHTTPServer(cfg *config.Config, handlers *Handlers, lgr *logger.Logger) *http.Server {
    // Create router with abstraction
    router := httpserver.NewStdlibRouter()
    
    // Apply global middleware
    router.Use(
        httpserver.RecoveryMiddleware(lgr),
        httpserver.LoggerMiddleware(lgr),
        httpserver.CORSMiddleware(cfg.CORS),
    )
    
    // Register module routes
    handlers.Auth.RegisterRoutes(router)
    handlers.User.RegisterRoutes(router)
    handlers.Post.RegisterRoutes(router)
    handlers.Comment.RegisterRoutes(router)
    handlers.Reaction.RegisterRoutes(router)
    
    server := &http.Server{
        Addr:         cfg.Server.Address,
        Handler:      router,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }
    
    return server
}
```

**Key Changes**:

- ✅ Use `httpserver.NewStdlibRouter()` instead of `http.NewServeMux()`
- ✅ Apply middleware using `router.Use()`
- ✅ Handlers now receive `httpserver.Router` interface
- ✅ Can easily swap router implementation (stdlib → Chi → Gin)

### Update Handler Interfaces

All module handlers should now accept `httpserver.Router` in `RegisterRoutes`:

```go
// Pattern for all modules
type Handler interface {
    RegisterRoutes(router httpserver.Router)
}
```

**Example**: Update handler constructors in `wire/handlers.go`:

```go
package wire

import (
    authAdapters "forum/internal/modules/auth/adapters"
    userAdapters "forum/internal/modules/user/adapters"
    // ... other imports
    
    "forum/internal/platform/httpserver"
    "forum/internal/platform/logger"
)

type Handlers struct {
    Auth     *authAdapters.Handler
    User     *userAdapters.Handler
    Post     *postAdapters.Handler
    Comment  *commentAdapters.Handler
    Reaction *reactionAdapters.Handler
}

func NewHandlers(services *Services, lgr *logger.Logger) *Handlers {
    return &Handlers{
        Auth:     authAdapters.NewHTTPHandler(services.Auth, lgr),
        User:     userAdapters.NewHTTPHandler(services.User, lgr),
        Post:     postAdapters.NewHTTPHandler(services.Post, lgr),
        Comment:  commentAdapters.NewHTTPHandler(services.Comment, lgr),
        Reaction: reactionAdapters.NewHTTPHandler(services.Reaction, lgr),
    }
}
```

---

## Testing Strategy

### 1. Mock Context for Unit Tests

Create a mock implementation of `httpserver.Context` for testing handlers in isolation.

**File**: `tests/unit/mock_http_context.go`

```go
package unit

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "mime/multipart"
    "net/http"
    
    "forum/internal/platform/httpserver"
)

// MockContext is a mock implementation of httpserver.Context for testing.
type MockContext struct {
    ParamFunc         func(name string) string
    ParamIntFunc      func(name string) (int, error)
    ParamInt64Func    func(name string) (int64, error)
    QueryFunc         func(name string) string
    QueryIntFunc      func(name string) (int, error)
    BindJSONFunc      func(v interface{}) error
    JSONFunc          func(code int, v interface{}) error
    BadRequestFunc    func(message string) error
    UnauthorizedFunc  func(message string) error
    NotFoundFunc      func(message string) error
    InternalErrorFunc func(err error) error
    
    request    *http.Request
    statusCode int
    response   interface{}
    errMessage string
}

// NewMockContext creates a new mock context.
func NewMockContext() *MockContext {
    req, _ := http.NewRequest("GET", "/", nil)
    return &MockContext{
        request: req,
    }
}

// WithRequest sets the underlying request.
func (m *MockContext) WithRequest(r *http.Request) *MockContext {
    m.request = r
    return m
}

// WithJSONBody sets JSON body for testing.
func (m *MockContext) WithJSONBody(v interface{}) *MockContext {
    data, _ := json.Marshal(v)
    m.request.Body = io.NopCloser(bytes.NewReader(data))
    return m
}

// Request returns the mock request.
func (m *MockContext) Request() *http.Request {
    return m.request
}

// Context returns the request context.
func (m *MockContext) Context() context.Context {
    return m.request.Context()
}

// SetContext sets a new context.
func (m *MockContext) SetContext(ctx context.Context) {
    m.request = m.request.WithContext(ctx)
}

// Param returns a path parameter.
func (m *MockContext) Param(name string) string {
    if m.ParamFunc != nil {
        return m.ParamFunc(name)
    }
    return ""
}

// ParamInt returns a path parameter as int.
func (m *MockContext) ParamInt(name string) (int, error) {
    if m.ParamIntFunc != nil {
        return m.ParamIntFunc(name)
    }
    return 0, nil
}

// ParamInt64 returns a path parameter as int64.
func (m *MockContext) ParamInt64(name string) (int64, error) {
    if m.ParamInt64Func != nil {
        return m.ParamInt64Func(name)
    }
    return 0, nil
}

// Query returns a query parameter.
func (m *MockContext) Query(name string) string {
    if m.QueryFunc != nil {
        return m.QueryFunc(name)
    }
    return m.request.URL.Query().Get(name)
}

// QueryDefault returns a query parameter with default.
func (m *MockContext) QueryDefault(name, defaultValue string) string {
    if v := m.Query(name); v != "" {
        return v
    }
    return defaultValue
}

// QueryInt returns a query parameter as int.
func (m *MockContext) QueryInt(name string) (int, error) {
    if m.QueryIntFunc != nil {
        return m.QueryIntFunc(name)
    }
    return 0, nil
}

// QueryInt64 returns a query parameter as int64.
func (m *MockContext) QueryInt64(name string) (int64, error) {
    return 0, nil
}

// Header returns a request header.
func (m *MockContext) Header(name string) string {
    return m.request.Header.Get(name)
}

// Cookie returns a cookie.
func (m *MockContext) Cookie(name string) (*http.Cookie, error) {
    return m.request.Cookie(name)
}

// BindJSON parses JSON body.
func (m *MockContext) BindJSON(v interface{}) error {
    if m.BindJSONFunc != nil {
        return m.BindJSONFunc(v)
    }
    return json.NewDecoder(m.request.Body).Decode(v)
}

// Bind parses request body.
func (m *MockContext) Bind(v interface{}) error {
    return m.BindJSON(v)
}

// BindForm parses form data.
func (m *MockContext) BindForm(v interface{}) error {
    return nil
}

// FormValue returns a form value.
func (m *MockContext) FormValue(name string) string {
    return ""
}

// FormFile returns a file from form.
func (m *MockContext) FormFile(name string) (*multipart.FileHeader, error) {
    return nil, nil
}

// Body returns the request body.
func (m *MockContext) Body() io.ReadCloser {
    return m.request.Body
}

// JSON sends a JSON response.
func (m *MockContext) JSON(code int, v interface{}) error {
    if m.JSONFunc != nil {
        return m.JSONFunc(code, v)
    }
    m.statusCode = code
    m.response = v
    return nil
}

// String sends a text response.
func (m *MockContext) String(code int, s string) error {
    m.statusCode = code
    m.response = s
    return nil
}

// HTML sends an HTML response.
func (m *MockContext) HTML(code int, html string) error {
    m.statusCode = code
    m.response = html
    return nil
}

// Data sends raw bytes.
func (m *MockContext) Data(code int, contentType string, data []byte) error {
    m.statusCode = code
    m.response = data
    return nil
}

// NoContent sends 204.
func (m *MockContext) NoContent() error {
    m.statusCode = http.StatusNoContent
    return nil
}

// Redirect sends a redirect.
func (m *MockContext) Redirect(code int, url string) error {
    m.statusCode = code
    return nil
}

// BadRequest sends 400.
func (m *MockContext) BadRequest(message string) error {
    if m.BadRequestFunc != nil {
        return m.BadRequestFunc(message)
    }
    m.statusCode = http.StatusBadRequest
    m.errMessage = message
    return nil
}

// Unauthorized sends 401.
func (m *MockContext) Unauthorized(message string) error {
    if m.UnauthorizedFunc != nil {
        return m.UnauthorizedFunc(message)
    }
    m.statusCode = http.StatusUnauthorized
    m.errMessage = message
    return nil
}

// Forbidden sends 403.
func (m *MockContext) Forbidden(message string) error {
    m.statusCode = http.StatusForbidden
    m.errMessage = message
    return nil
}

// NotFound sends 404.
func (m *MockContext) NotFound(message string) error {
    if m.NotFoundFunc != nil {
        return m.NotFoundFunc(message)
    }
    m.statusCode = http.StatusNotFound
    m.errMessage = message
    return nil
}

// Conflict sends 409.
func (m *MockContext) Conflict(message string) error {
    m.statusCode = http.StatusConflict
    m.errMessage = message
    return nil
}

// InternalError sends 500.
func (m *MockContext) InternalError(err error) error {
    if m.InternalErrorFunc != nil {
        return m.InternalErrorFunc(err)
    }
    m.statusCode = http.StatusInternalServerError
    m.errMessage = "Internal server error"
    return nil
}

// Error sends custom error.
func (m *MockContext) Error(code int, message string) error {
    m.statusCode = code
    m.errMessage = message
    return nil
}

// SetHeader sets a response header.
func (m *MockContext) SetHeader(name, value string) {
    // No-op in mock
}

// SetCookie sets a cookie.
func (m *MockContext) SetCookie(cookie *http.Cookie) {
    // No-op in mock
}

// Status sets response status.
func (m *MockContext) Status(code int) {
    m.statusCode = code
}

// Set stores a value.
func (m *MockContext) Set(key string, value interface{}) {
    ctx := context.WithValue(m.request.Context(), key, value)
    m.request = m.request.WithContext(ctx)
}

// Get retrieves a value.
func (m *MockContext) Get(key string) interface{} {
    return m.request.Context().Value(key)
}

// GetString retrieves a string.
func (m *MockContext) GetString(key string) string {
    if v := m.Get(key); v != nil {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

// GetInt retrieves an int.
func (m *MockContext) GetInt(key string) int {
    if v := m.Get(key); v != nil {
        if i, ok := v.(int); ok {
            return i
        }
    }
    return 0
}

// GetStatusCode returns the response status code (for testing).
func (m *MockContext) GetStatusCode() int {
    return m.statusCode
}

// GetResponse returns the response data (for testing).
func (m *MockContext) GetResponse() interface{} {
    return m.response
}

// GetErrorMessage returns the error message (for testing).
func (m *MockContext) GetErrorMessage() string {
    return m.errMessage
}
```

### 2. Unit Test Example

**File**: `internal/modules/user/adapters/http_handler_test.go`

```go
package adapters

import (
    "context"
    "errors"
    "net/http"
    "testing"
    
    "forum/internal/modules/user/domain"
    "forum/tests/unit"
)

// MockUserService for testing
type MockUserService struct {
    GetByIDFunc func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *MockUserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, nil
}

func TestHandler_GetUser(t *testing.T) {
    tests := []struct {
        name           string
        userID         string
        mockService    func() *MockUserService
        mockContext    func() *unit.MockContext
        expectedStatus int
        expectError    bool
    }{
        {
            name:   "successful get user",
            userID: "123",
            mockService: func() *MockUserService {
                return &MockUserService{
                    GetByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
                        if id != 123 {
                            t.Errorf("expected id 123, got %d", id)
                        }
                        return &domain.User{
                            ID:       123,
                            Username: "testuser",
                            Email:    "test@example.com",
                        }, nil
                    },
                }
            },
            mockContext: func() *unit.MockContext {
                ctx := unit.NewMockContext()
                ctx.ParamInt64Func = func(name string) (int64, error) {
                    if name == "id" {
                        return 123, nil
                    }
                    return 0, errors.New("param not found")
                }
                return ctx
            },
            expectedStatus: http.StatusOK,
            expectError:    false,
        },
        {
            name:   "invalid user id",
            userID: "invalid",
            mockService: func() *MockUserService {
                return &MockUserService{}
            },
            mockContext: func() *unit.MockContext {
                ctx := unit.NewMockContext()
                ctx.ParamInt64Func = func(name string) (int64, error) {
                    return 0, errors.New("invalid parameter")
                }
                ctx.BadRequestFunc = func(message string) error {
                    ctx.Status(http.StatusBadRequest)
                    return nil
                }
                return ctx
            },
            expectedStatus: http.StatusBadRequest,
            expectError:    false,
        },
        {
            name:   "user not found",
            userID: "999",
            mockService: func() *MockUserService {
                return &MockUserService{
                    GetByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
                        return nil, errors.New("user not found")
                    },
                }
            },
            mockContext: func() *unit.MockContext {
                ctx := unit.NewMockContext()
                ctx.ParamInt64Func = func(name string) (int64, error) {
                    return 999, nil
                }
                ctx.NotFoundFunc = func(message string) error {
                    ctx.Status(http.StatusNotFound)
                    return nil
                }
                return ctx
            },
            expectedStatus: http.StatusNotFound,
            expectError:    false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := tt.mockService()
            mockCtx := tt.mockContext()
            
            handler := NewHTTPHandler(service, nil)
            err := handler.GetUser(mockCtx)
            
            if tt.expectError && err == nil {
                t.Error("expected error but got nil")
            }
            if !tt.expectError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            
            if mockCtx.GetStatusCode() != tt.expectedStatus {
                t.Errorf("expected status %d, got %d", tt.expectedStatus, mockCtx.GetStatusCode())
            }
        })
    }
}

func TestHandler_UpdateUser(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    map[string]interface{}
        mockService    func() *MockUserService
        mockContext    func() *unit.MockContext
        expectedStatus int
    }{
        {
            name: "successful update",
            requestBody: map[string]interface{}{
                "username": "newusername",
                "bio":      "New bio",
            },
            mockService: func() *MockUserService {
                return &MockUserService{
                    UpdateFunc: func(ctx context.Context, id int64, data interface{}) (*domain.User, error) {
                        return &domain.User{
                            ID:       123,
                            Username: "newusername",
                            Bio:      "New bio",
                        }, nil
                    },
                }
            },
            mockContext: func() *unit.MockContext {
                ctx := unit.NewMockContext()
                ctx.ParamInt64Func = func(name string) (int64, error) {
                    return 123, nil
                }
                ctx.WithJSONBody(map[string]interface{}{
                    "username": "newusername",
                    "bio":      "New bio",
                })
                return ctx
            },
            expectedStatus: http.StatusOK,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := tt.mockService()
            mockCtx := tt.mockContext()
            
            handler := NewHTTPHandler(service, nil)
            err := handler.UpdateUser(mockCtx)
            
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            
            if mockCtx.GetStatusCode() != tt.expectedStatus {
                t.Errorf("expected status %d, got %d", tt.expectedStatus, mockCtx.GetStatusCode())
            }
        })
    }
}
```

**Benefits of Mock-Based Testing**:

- ✅ No need for HTTP test servers
- ✅ Fast execution (no network overhead)
- ✅ Easy to test edge cases
- ✅ Clear test intent with mock functions
- ✅ Can verify exact parameters passed to services

---

## Implementation Roadmap

### Phase 1: Platform Layer Foundation (Week 1)

**Priority**: HIGH - Required for all other work

- [ ] Create `internal/platform/httpserver/router.go` - Define Router interface
- [ ] Create `internal/platform/httpserver/context.go` - Define Context interface
- [ ] Create `internal/platform/httpserver/stdlib_router.go` - Standard library implementation
- [ ] Create `internal/platform/httpserver/stdlib_context.go` - Standard library context
- [ ] Create `internal/platform/httpserver/response.go` - Response helpers
- [ ] Write unit tests for router and context implementations
- [ ] Test middleware chaining

**Success Criteria**:

- ✅ Router interface compiles and is well-documented
- ✅ Context interface covers all common HTTP operations
- ✅ Stdlib implementation passes all tests
- ✅ Middleware can be applied at router and group level
- ✅ Error responses are consistent

### Phase 2: Test Infrastructure (Week 1-2)

**Priority**: HIGH - Needed before refactoring modules

- [ ] Create `tests/unit/mock_http_context.go` - Mock Context implementation
- [ ] Create `tests/unit/mock_http_router.go` - Mock Router implementation (if needed)
- [ ] Write example tests demonstrating mock usage
- [ ] Document testing patterns for module handlers

**Success Criteria**:

- ✅ Mock context implements all interface methods
- ✅ Handlers can be tested without HTTP servers
- ✅ Clear examples for other developers

### Phase 3: Auth Module (Week 2)

**Priority**: HIGH - Authentication is critical

- [ ] Refactor `internal/modules/auth/adapters/http_handler.go`
  - [ ] Change `RegisterRoutes` to accept `httpserver.Router`
  - [ ] Update handler signatures to `func(httpserver.Context) error`
  - [ ] Replace `http.ResponseWriter` and `*http.Request` usage
  - [ ] Use context helpers for parameter extraction
  - [ ] Use context methods for JSON responses
- [ ] Write unit tests with mock context
- [ ] Test cookie handling (login/logout)
- [ ] Update `cmd/forum/wire/app.go` to use new router

**Success Criteria**:

- ✅ Auth handlers compile without `net/http` imports (except for constants)
- ✅ All tests pass
- ✅ Login/logout flow works correctly
- ✅ Cookie handling is secure

### Phase 4: User Module (Week 2-3)

**Priority**: HIGH - Core functionality

- [ ] Refactor `internal/modules/user/adapters/http_handler.go`
  - [ ] Update RegisterRoutes
  - [ ] Refactor all handler methods
  - [ ] Use route grouping (`/api/users`)
- [ ] Write unit tests
- [ ] Test parameter extraction (user ID)
- [ ] Test cross-module activity aggregation

**Success Criteria**:

- ✅ User handlers are framework-agnostic
- ✅ All endpoints work correctly
- ✅ Tests cover all scenarios

### Phase 5: Post Module (Week 3)

**Priority**: MEDIUM - Core content

- [ ] Refactor `internal/modules/post/adapters/http_handler.go`
  - [ ] Update for httpserver abstraction
  - [ ] Handle file uploads through context
  - [ ] Test multipart form data
- [ ] Write unit tests
- [ ] Test image upload handling

**Success Criteria**:

- ✅ Post creation/editing works
- ✅ Image uploads handled correctly
- ✅ Category associations work

### Phase 6: Comment Module (Week 3-4)

**Priority**: MEDIUM - Social features

- [ ] Refactor `internal/modules/comment/adapters/http_handler.go`
  - [ ] Update for httpserver abstraction
- [ ] Write unit tests
- [ ] Test nested comment responses

**Success Criteria**:

- ✅ Comment CRUD operations work
- ✅ Associated with correct posts

### Phase 7: Reaction Module (Week 4)

**Priority**: MEDIUM - Social features

- [ ] Refactor `internal/modules/reaction/adapters/http_handler.go`
  - [ ] Update for httpserver abstraction
- [ ] Write unit tests
- [ ] Test like/dislike toggle behavior

**Success Criteria**:

- ✅ Reactions work correctly
- ✅ Toggle behavior prevents double reactions

### Phase 8: Optional Modules (Week 4-5)

**Priority**: LOW - Can be deferred

- [ ] Moderation module handlers
- [ ] Notification module handlers
- [ ] Write tests for optional modules

### Phase 9: Integration Testing (Week 5)

**Priority**: HIGH - Verification

- [ ] Write end-to-end tests with real HTTP requests
- [ ] Test all audit scenarios
- [ ] Test middleware interactions
- [ ] Test error handling across modules
- [ ] Performance testing

**Success Criteria**:

- ✅ All audit requirements pass
- ✅ No HTTP-related bugs
- ✅ Consistent API behavior

### Phase 10: Alternative Implementations (Week 5-6)

**Priority**: LOW - Optional enhancement

- [ ] Create `internal/platform/httpserver/chi_router.go` - Chi implementation
- [ ] Create `internal/platform/httpserver/gin_router.go` - Gin implementation
- [ ] Test modules with alternative routers
- [ ] Document framework switching

**Success Criteria**:

- ✅ Modules work without changes on Chi
- ✅ Modules work without changes on Gin
- ✅ Performance comparison documented

---

## Benefits Summary

### Development Benefits

1. **Framework Independence**: Switch from stdlib to Chi/Gin/Echo without changing module code
2. **Reduced Boilerplate**: 40-50% less code in handlers
3. **Consistent Patterns**: All handlers use same error handling and response format
4. **Easy Testing**: Mock context for unit tests, no HTTP servers needed
5. **Type Safety**: Parameter extraction helpers prevent runtime errors

### Architecture Benefits

1. **Hexagonal Compliance**: HTTP framework is infrastructure (platform), not domain concern
2. **Module Autonomy**: Modules define their own APIs independently
3. **Clear Boundaries**: Platform handles protocol, modules handle business logic
4. **Dependency Inversion**: Modules depend on abstractions, not implementations

### Operational Benefits

1. **Consistent API**: All endpoints return errors in same format
2. **Better Logging**: Centralized error handling makes logging consistent
3. **Easier Debugging**: Clear separation of concerns
4. **Performance**: Can choose fastest framework for production

### Maintenance Benefits

1. **Single Update Point**: Change router implementation once, affects all modules
2. **Easy Refactoring**: Modules can be refactored without worrying about HTTP details
3. **Clear Contracts**: Interfaces document what handlers can do
4. **Future-Proof**: New HTTP features added to platform, not scattered across modules

---

## Migration Checklist

Use this checklist when refactoring each module handler:

- [ ] Update imports: Remove `net/http` (except constants), add `httpserver`
- [ ] Change `RegisterRoutes(mux *http.ServeMux)` to `RegisterRoutes(router httpserver.Router)`
- [ ] Update all handler signatures: `func(w http.ResponseWriter, r *http.Request)` → `func(ctx httpserver.Context) error`
- [ ] Replace parameter extraction: `r.PathValue("id")` → `ctx.Param("id")` or `ctx.ParamInt64("id")`
- [ ] Replace JSON decoding: `json.NewDecoder(r.Body).Decode(&v)` → `ctx.BindJSON(&v)`
- [ ] Replace JSON encoding: `json.NewEncoder(w).Encode(v)` → `ctx.JSON(code, v)`
- [ ] Replace error responses: `http.Error(w, msg, code)` → `ctx.BadRequest(msg)` / `ctx.InternalError(err)`
- [ ] Replace cookie operations: `r.Cookie(name)` → `ctx.Cookie(name)`, `http.SetCookie(w, c)` → `ctx.SetCookie(c)`
- [ ] Update route registration: Use `router.GET()`, `router.POST()`, etc.
- [ ] Use route grouping: `router.Group("/api/users")` for common prefixes
- [ ] Write unit tests with mock context
- [ ] Write integration tests if needed
- [ ] Update wiring in `cmd/forum/wire/`

---

## Conclusion

This refactoring provides **HTTP framework abstraction** while maintaining **hexagonal architecture**. The platform layer handles all HTTP protocol specifics, allowing modules to focus on business logic and API design.

**Key Takeaways**:

✅ **Modules own API design**, platform owns HTTP protocol  
✅ **40-50% less boilerplate** in handler code  
✅ **Framework independence** - switch stdlib/Chi/Gin easily  
✅ **Easy testing** with mock context  
✅ **Consistent error handling** across all modules  
✅ **Hexagonal compliance** - HTTP is infrastructure concern

**Next Steps**:

1. Review this proposal with the team
2. Start with Phase 1 (Platform Layer Foundation)
3. Build test infrastructure before refactoring modules
4. Refactor one module at a time, starting with Auth
5. Verify with integration tests after each module
6. Consider alternative implementations once stdlib version is stable

**Questions or Issues?**

- Check existing httpserver code in `internal/platform/httpserver/`
- Review hexagonal architecture docs in `docs/ARCHITECTURE.md`
- Test each change thoroughly before moving to next module
- Ask for code review before merging major changes

