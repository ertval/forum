# HTTP Handler Abstraction Proposal

## Overview

This document proposes abstracting HTTP handler implementations to improve testability, maintainability, and framework flexibility while maintaining hexagonal architecture principles.

## Current State Analysis

### Current Architecture

```
Module Layer (auth, user, etc.)
├── adapters/
│   └── http_handler.go          ← Concrete HTTP handlers
│       ├── HTTPHandler struct
│       ├── Register(w, r)       ← Direct http.ResponseWriter usage
│       ├── Login(w, r)          ← Direct http.Request usage
│       └── RegisterRoutes(mux)  ← Direct *http.ServeMux usage

Platform Layer
├── httpserver/
│   ├── server.go                ← Server setup
│   ├── middleware.go            ← Middleware chain
│   └── router.go                ← Route registration
```

**Current Issues:**
- ✅ Handlers are thin and focused on HTTP → business logic translation
- ❌ Direct coupling to `net/http` makes testing difficult
- ❌ JSON encoding/decoding repeated across all handlers
- ❌ Error response formatting duplicated
- ❌ No abstraction for different HTTP frameworks

## Proposed Architecture

### Target State

```
Module Layer (auth, user, etc.)
├── ports/
│   └── http_handler.go          ← INPUT PORT: Handler interface
│       └── AuthHandler interface {
│               Register(req) (resp, error)
│               Login(req) (resp, error)
│           }
├── adapters/
│   └── http_handler.go          ← INPUT ADAPTER: Thin HTTP adapter
│       └── HTTPHandler struct implements AuthHandler
│           └── Register(req) → calls service, returns response

Platform Layer
├── httpserver/
│   ├── server.go                ← Server setup
│   ├── middleware.go            ← Middleware chain
│   ├── router.go                ← Route registration
│   ├── response.go              ← HTTP response abstractions
│   ├── request.go               ← HTTP request abstractions
│   └── json.go                  ← JSON encoding/decoding utilities
```

### Key Changes

1. **Handler Interfaces in Ports**: Define what each module's HTTP interface looks like
2. **Thin HTTP Adapters**: Convert between HTTP and the handler interface
3. **Platform HTTP Utilities**: Common HTTP operations abstracted
4. **Testable Interfaces**: Mock HTTP responses in unit tests

## Implementation Strategy

### Phase 1: Platform HTTP Abstractions

Create HTTP abstractions in `internal/platform/httpserver/`:

#### 1. Response Abstractions (`response.go`)

```go
package httpserver

import (
    "encoding/json"
    "net/http"
)

// Response represents an HTTP response that can be written.
type Response interface {
    StatusCode() int
    Headers() map[string]string
    Body() interface{}
}

// JSONResponse represents a JSON HTTP response.
type JSONResponse struct {
    statusCode int
    data       interface{}
    headers    map[string]string
}

func NewJSONResponse(statusCode int, data interface{}) Response {
    return &JSONResponse{
        statusCode: statusCode,
        data:       data,
        headers:    map[string]string{"Content-Type": "application/json"},
    }
}

func NewErrorResponse(statusCode int, message string) Response {
    return NewJSONResponse(statusCode, map[string]string{"error": message})
}

// ResponseWriter abstracts http.ResponseWriter for testing.
type ResponseWriter interface {
    WriteResponse(resp Response) error
}

// HTTPResponseWriter implements ResponseWriter using http.ResponseWriter.
type HTTPResponseWriter struct {
    w http.ResponseWriter
}

func NewHTTPResponseWriter(w http.ResponseWriter) ResponseWriter {
    return &HTTPResponseWriter{w: w}
}

func (rw *HTTPResponseWriter) WriteResponse(resp Response) error {
    // Set status code
    rw.w.WriteHeader(resp.StatusCode())
    
    // Set headers
    for key, value := range resp.Headers() {
        rw.w.Header().Set(key, value)
    }
    
    // Write JSON body
    return json.NewEncoder(rw.w).Encode(resp.Body())
}
```

#### 2. Request Abstractions (`request.go`)

```go
package httpserver

import (
    "encoding/json"
    "net/http"
)

// Request represents an HTTP request that can be read.
type Request interface {
    Method() string
    Path() string
    QueryParam(key string) string
    Header(key string) string
    Body() interface{}
    Context() context.Context
}

// JSONRequest represents a JSON HTTP request.
type JSONRequest struct {
    method string
    path   string
    query  url.Values
    headers map[string]string
    body   interface{}
    ctx    context.Context
}

// HTTPRequest implements Request using http.Request.
type HTTPRequest struct {
    r *http.Request
}

func NewHTTPRequest(r *http.Request) Request {
    return &HTTPRequest{r: r}
}

func (r *HTTPRequest) Method() string {
    return r.r.Method
}

func (r *HTTPRequest) Path() string {
    return r.r.URL.Path
}

func (r *HTTPRequest) QueryParam(key string) string {
    return r.r.URL.Query().Get(key)
}

func (r *HTTPRequest) Header(key string) string {
    return r.r.Header.Get(key)
}

func (r *HTTPRequest) Body() interface{} {
    var body interface{}
    json.NewDecoder(r.r.Body).Decode(&body)
    return body
}

func (r *HTTPRequest) Context() context.Context {
    return r.r.Context()
}
```

#### 3. JSON Utilities (`json.go`)

```go
package httpserver

import (
    "encoding/json"
    "net/http"
)

// JSONCodec handles JSON encoding/decoding with error handling.
type JSONCodec struct{}

// DecodeRequest decodes JSON from request body into target struct.
func (c *JSONCodec) DecodeRequest(r *http.Request, target interface{}) error {
    defer r.Body.Close()
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields() // Strict decoding
    return decoder.Decode(target)
}

// EncodeResponse encodes data to JSON and writes to response writer.
func (c *JSONCodec) EncodeResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    return json.NewEncoder(w).Encode(data)
}

// EncodeError encodes error message to JSON response.
func (c *JSONCodec) EncodeError(w http.ResponseWriter, statusCode int, message string) error {
    return c.EncodeResponse(w, statusCode, map[string]string{"error": message})
}
```

### Phase 2: Module Handler Interface Refactoring

#### 1. Define Handler Interfaces in Ports

**File**: `internal/modules/auth/ports/http_handler.go`

```go
// INPUT PORT - HTTP Handler Interface
// Package ports defines the HTTP handler interfaces for the auth module.
// These interfaces define how HTTP requests are handled.
package ports

import (
    "context"
    "forum/internal/platform/httpserver"
)

// AuthHandler defines the HTTP interface for authentication operations.
// Implementations will handle HTTP requests and return HTTP responses.
type AuthHandler interface {
    // Register handles user registration requests.
    Register(ctx context.Context, req httpserver.Request) (httpserver.Response, error)
    
    // Login handles user login requests.
    Login(ctx context.Context, req httpserver.Request) (httpserver.Response, error)
    
    // Logout handles user logout requests.
    Logout(ctx context.Context, req httpserver.Request) (httpserver.Response, error)
    
    // GetSession retrieves current session information.
    GetSession(ctx context.Context, req httpserver.Request) (httpserver.Response, error)
}
```

#### 2. Implement Handler Interface in Adapters

**File**: `internal/modules/auth/adapters/http_handler.go`

```go
// INPUT ADAPTER - HTTP Handler Implementation
// Package adapters implements the HTTP handler interface for authentication.
// This adapter translates between HTTP requests/responses and the handler interface.
package adapters

import (
    "context"
    "forum/internal/modules/auth/ports"
    "forum/internal/platform/httpserver"
)

// HTTPHandler implements the AuthHandler interface.
type HTTPHandler struct {
    authService ports.AuthService
    jsonCodec   *httpserver.JSONCodec
}

// NewHTTPHandler creates a new HTTP handler for authentication.
func NewHTTPHandler(authService ports.AuthService) ports.AuthHandler {
    return &HTTPHandler{
        authService: authService,
        jsonCodec:   &httpserver.JSONCodec{},
    }
}

// Register handles user registration requests.
func (h *HTTPHandler) Register(ctx context.Context, req httpserver.Request) (httpserver.Response, error) {
    // Parse request body
    var registerReq struct {
        Email    string `json:"email"`
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    if err := h.jsonCodec.DecodeRequest(req, &registerReq); err != nil {
        return httpserver.NewErrorResponse(400, "Invalid request body"), nil
    }
    
    // Validate input
    if registerReq.Email == "" || registerReq.Username == "" || registerReq.Password == "" {
        return httpserver.NewErrorResponse(400, "Email, username, and password are required"), nil
    }
    
    // Call service
    user, err := h.authService.Register(ctx, registerReq.Email, registerReq.Username, registerReq.Password)
    if err != nil {
        // Handle domain errors
        return httpserver.NewErrorResponse(400, err.Error()), nil
    }
    
    // Return success response
    return httpserver.NewJSONResponse(201, map[string]interface{}{
        "user": user,
        "message": "User registered successfully",
    }), nil
}

// Login handles user login requests.
func (h *HTTPHandler) Login(ctx context.Context, req httpserver.Request) (httpserver.Response, error) {
    // Parse request body
    var loginReq struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := h.jsonCodec.DecodeRequest(req, &loginReq); err != nil {
        return httpserver.NewErrorResponse(400, "Invalid request body"), nil
    }
    
    // Call service
    session, err := h.authService.Login(ctx, loginReq.Email, loginReq.Password)
    if err != nil {
        return httpserver.NewErrorResponse(401, "Invalid credentials"), nil
    }
    
    // Return success response with session
    return httpserver.NewJSONResponse(200, map[string]interface{}{
        "session": session,
        "message": "Login successful",
    }), nil
}

// Logout handles user logout requests.
func (h *HTTPHandler) Logout(ctx context.Context, req httpserver.Request) (httpserver.Response, error) {
    // Get session token from request
    token := req.Header("Authorization")
    if token == "" {
        return httpserver.NewErrorResponse(401, "No session token provided"), nil
    }
    
    // Call service
    err := h.authService.Logout(ctx, token)
    if err != nil {
        return httpserver.NewErrorResponse(500, "Failed to logout"), nil
    }
    
    return httpserver.NewJSONResponse(200, map[string]string{
        "message": "Logout successful",
    }), nil
}

// GetSession retrieves current session information.
func (h *HTTPHandler) GetSession(ctx context.Context, req httpserver.Request) (httpserver.Response, error) {
    // Get session token from request
    token := req.Header("Authorization")
    if token == "" {
        return httpserver.NewErrorResponse(401, "No session token provided"), nil
    }
    
    // Call service
    session, err := h.authService.ValidateSession(ctx, token)
    if err != nil {
        return httpserver.NewErrorResponse(401, "Invalid session"), nil
    }
    
    return httpserver.NewJSONResponse(200, session), nil
}
```

#### 3. Create HTTP Adapter Layer

**File**: `internal/modules/auth/adapters/http_adapter.go`

```go
// INPUT ADAPTER - HTTP Adapter
// Package adapters provides the bridge between HTTP and the handler interface.
// This adapter converts raw HTTP requests/responses to/from the abstracted interface.
package adapters

import (
    "net/http"
    "forum/internal/modules/auth/ports"
    "forum/internal/platform/httpserver"
)

// HTTPAdapter adapts between net/http and the AuthHandler interface.
type HTTPAdapter struct {
    handler ports.AuthHandler
}

// NewHTTPAdapter creates a new HTTP adapter.
func NewHTTPAdapter(handler ports.AuthHandler) *HTTPAdapter {
    return &HTTPAdapter{handler: handler}
}

// Register adapts HTTP request to handler interface.
func (a *HTTPAdapter) Register(w http.ResponseWriter, r *http.Request) {
    req := httpserver.NewHTTPRequest(r)
    resp, err := a.handler.Register(r.Context(), req)
    
    writer := httpserver.NewHTTPResponseWriter(w)
    if err != nil {
        writer.WriteResponse(httpserver.NewErrorResponse(500, err.Error()))
        return
    }
    
    writer.WriteResponse(resp)
}

// Login adapts HTTP request to handler interface.
func (a *HTTPAdapter) Login(w http.ResponseWriter, r *http.Request) {
    req := httpserver.NewHTTPRequest(r)
    resp, err := a.handler.Login(r.Context(), req)
    
    writer := httpserver.NewHTTPResponseWriter(w)
    if err != nil {
        writer.WriteResponse(httpserver.NewErrorResponse(500, err.Error()))
        return
    }
    
    writer.WriteResponse(resp)
}

// Logout adapts HTTP request to handler interface.
func (a *HTTPAdapter) Logout(w http.ResponseWriter, r *http.Request) {
    req := httpserver.NewHTTPRequest(r)
    resp, err := a.handler.Logout(r.Context(), req)
    
    writer := httpserver.NewHTTPResponseWriter(w)
    if err != nil {
        writer.WriteResponse(httpserver.NewErrorResponse(500, err.Error()))
        return
    }
    
    writer.WriteResponse(resp)
}

// GetSession adapts HTTP request to handler interface.
func (a *HTTPAdapter) GetSession(w http.ResponseWriter, r *http.Request) {
    req := httpserver.NewHTTPRequest(r)
    resp, err := a.handler.GetSession(r.Context(), req)
    
    writer := httpserver.NewHTTPResponseWriter(w)
    if err != nil {
        writer.WriteResponse(httpserver.NewErrorResponse(500, err.Error()))
        return
    }
    
    writer.WriteResponse(resp)
}

// RegisterRoutes registers routes with the HTTP server.
func (a *HTTPAdapter) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("POST /api/auth/register", a.Register)
    mux.HandleFunc("POST /api/auth/login", a.Login)
    mux.HandleFunc("POST /api/auth/logout", a.Logout)
    mux.HandleFunc("GET /api/auth/session", a.GetSession)
}
```

### Phase 3: Update Wiring

#### 1. Update Wire Handlers

**File**: `cmd/forum/wire/handlers.go`

```go
// INPUT ADAPTERS - HTTP Handler Initialization
package wire

import (
    authAdapters "forum/internal/modules/auth/adapters"
    // ... other imports
)

// Handlers holds all HTTP adapter instances.
type Handlers struct {
    Auth *authAdapters.HTTPAdapter
    // ... other handlers
}

// initHandlers creates all HTTP adapter instances.
func initHandlers(services *Services) *Handlers {
    return &Handlers{
        Auth: authAdapters.NewHTTPAdapter(
            authAdapters.NewHTTPHandler(services.Auth),
        ),
        // ... other handlers
    }
}
```

## Benefits of This Approach

### 1. Improved Testability

**Before**: Hard to test HTTP handlers due to direct `http.ResponseWriter` coupling

```go
// Hard to test - requires httptest.ResponseRecorder
func TestRegister(t *testing.T) {
    handler := NewHTTPHandler(mockService)
    req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"email":"test"}`))
    w := httptest.NewRecorder()
    handler.Register(w, req)
    // Assert on w.Body, w.Code, etc.
}
```

**After**: Easy to test handler interface directly

```go
// Easy to test - pure interface testing
func TestRegister(t *testing.T) {
    handler := NewHTTPHandler(mockService)
    req := &mockRequest{body: `{"email":"test"}`}
    resp, err := handler.Register(context.Background(), req)
    // Assert on resp.StatusCode(), resp.Body(), err
}
```

### 2. Framework Flexibility

**Current**: Locked into `net/http`

**Future**: Could easily swap to Gin, Echo, Fiber, etc.

```go
// Could implement GinAdapter, EchoAdapter, etc.
type GinAdapter struct {
    handler AuthHandler
}

func (a *GinAdapter) Register(c *gin.Context) {
    req := httpserver.NewGinRequest(c)
    resp, err := a.handler.Register(c.Request.Context(), req)
    // Write response using Gin methods
}
```

### 3. Separation of Concerns

- **Handler Interface**: Business logic HTTP contract
- **HTTP Handler**: Implementation of the interface
- **HTTP Adapter**: Bridge between HTTP framework and interface
- **Platform Utilities**: Common HTTP operations

### 4. Consistency Across Modules

All modules follow the same HTTP handling pattern:
1. Define handler interface in `ports/`
2. Implement interface in `adapters/`
3. Create HTTP adapter in `adapters/`
4. Wire everything in `wire/`

## Implementation Timeline

### Week 1: Platform Abstractions
- [ ] Create `httpserver/response.go`
- [ ] Create `httpserver/request.go`
- [ ] Create `httpserver/json.go`
- [ ] Write tests for abstractions

### Week 2: Auth Module Refactor
- [ ] Create `auth/ports/http_handler.go`
- [ ] Refactor `auth/adapters/http_handler.go`
- [ ] Create `auth/adapters/http_adapter.go`
- [ ] Update wiring and tests

### Week 3: User Module Refactor
- [ ] Apply same pattern to user module
- [ ] Update tests

### Week 4: Remaining Modules
- [ ] Post, Comment, Reaction modules
- [ ] Optional modules (Moderation, Notification)

## Conclusion

This HTTP abstraction provides significant benefits for testability and maintainability while maintaining the hexagonal architecture. The abstraction is lighter than the database abstraction since HTTP is more standardized than SQL dialects.

**Key Benefits:**
- ✅ Much easier unit testing of handlers
- ✅ Consistent HTTP handling across modules
- ✅ Framework flexibility for future changes
- ✅ Clear separation between HTTP mechanics and business logic
- ✅ Maintains hexagonal architecture principles

**Trade-offs:**
- ⚠️ More files and interfaces to maintain
- ⚠️ Slight performance overhead from abstraction
- ⚠️ Learning curve for the new patterns

**Recommendation:** Yes, implement this abstraction. The testing and maintainability benefits outweigh the complexity costs, especially as the application grows.