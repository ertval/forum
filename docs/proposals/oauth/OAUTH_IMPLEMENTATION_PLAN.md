# OAuth Authentication Implementation Plan

## Overview

This document outlines the implementation plan for adding Google and GitHub OAuth authentication to supplement the existing email/password authentication system in the forum application.

**Requirement**: [forum-authentication.md](requirements/forum-authentication.md)  
**Audit Checklist**: [audit-authentication.md](requirements/audit-authentication.md)

---

## Current State Analysis

### Existing Infrastructure ✅

1. **Config already supports OAuth** (`internal/platform/config/config.go`):
   ```go
   type OAuthConfig struct {
       Google GoogleOAuthConfig
       GitHub GitHubOAuthConfig
   }
   ```
   - Environment variables: `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, etc.
   - Validation for OAuth config already exists

2. **Database schema supports OAuth** (`migrations/002_user_create_users.sql`):
   ```sql
   oauth_provider TEXT,
   oauth_provider_id TEXT,
   CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_provider_id);
   ```

3. **User domain already has OAuth fields** ready (need to verify/add to struct)

4. **Auth module structure** follows hexagonal architecture:
   - `domain/` - Entities and errors
   - `ports/` - Service and repository interfaces
   - `application/` - Business logic
   - `adapters/` - HTTP handlers, SQLite repositories

---

## Implementation Plan

### Phase 1: Domain Layer Updates

**Files to modify/create:**

#### 1.1 Update Domain Entities

**File**: `internal/modules/user/domain/user.go`

Add OAuth fields to User struct:
```go
type User struct {
    // ... existing fields ...
    OAuthProvider   string `json:"-"` // "google", "github", or empty for email/password
    OAuthProviderID string `json:"-"` // Provider's unique user ID
}
```

#### 1.2 Add OAuth-specific Domain Errors

**File**: `internal/modules/auth/domain/errors.go`

Add new errors:
```go
var (
    // ErrOAuthProviderError is returned when OAuth provider communication fails.
    ErrOAuthProviderError = errors.New("oauth provider error")
    
    // ErrOAuthStateInvalid is returned when OAuth state parameter doesn't match.
    ErrOAuthStateInvalid = errors.New("invalid oauth state")
    
    // ErrOAuthEmailNotVerified is returned when OAuth email is not verified.
    ErrOAuthEmailNotVerified = errors.New("oauth email not verified")
    
    // ErrAccountAlreadyLinked is returned when trying to link an already linked account.
    ErrAccountAlreadyLinked = errors.New("account already linked to another user")
)
```

#### 1.3 Create OAuth Domain Types

**File**: `internal/modules/auth/domain/oauth.go` (NEW)

```go
package domain

// OAuthProvider represents supported OAuth providers.
type OAuthProvider string

const (
    OAuthProviderGoogle OAuthProvider = "google"
    OAuthProviderGitHub OAuthProvider = "github"
)

// OAuthUserInfo represents user information from an OAuth provider.
type OAuthUserInfo struct {
    Provider    OAuthProvider
    ProviderID  string // Unique ID from the provider
    Email       string
    Username    string // Display name from provider
    AvatarURL   string // Profile picture URL (optional, for future use)
    IsVerified  bool   // Email verification status
}

// Validate validates the OAuth user info.
func (o *OAuthUserInfo) Validate() error {
    if o.Provider == "" {
        return ErrOAuthProviderError
    }
    if o.ProviderID == "" {
        return ErrOAuthProviderError
    }
    if o.Email == "" {
        return ErrInvalidEmail
    }
    if !o.IsVerified {
        return ErrOAuthEmailNotVerified
    }
    return nil
}
```

---

### Phase 2: Port Layer Updates

#### 2.1 Update User Repository Interface

**File**: `internal/modules/user/ports/repository.go`

Add OAuth-specific methods:
```go
type UserRepository interface {
    // ... existing methods ...
    
    // GetByOAuthProvider retrieves a user by their OAuth provider and provider ID.
    GetByOAuthProvider(ctx context.Context, provider, providerID string) (*domain.User, error)
    
    // LinkOAuthAccount links an OAuth account to an existing user.
    LinkOAuthAccount(ctx context.Context, userID int, provider, providerID string) error
}
```

#### 2.2 Update Auth Service Interface

**File**: `internal/modules/auth/ports/service.go`

Add OAuth methods:
```go
type AuthService interface {
    // ... existing methods ...
    
    // LoginWithOAuth authenticates or registers a user via OAuth.
    // If user exists with this OAuth, logs them in.
    // If user exists with same email (password auth), links accounts and logs in.
    // If user doesn't exist, creates new account and logs in.
    LoginWithOAuth(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.Session, error)
}
```

#### 2.3 Create OAuth Provider Port

**File**: `internal/modules/auth/ports/oauth.go` (NEW)

```go
package ports

import (
    "context"
    "forum/internal/modules/auth/domain"
)

// OAuthProvider defines the interface for OAuth providers.
type OAuthProvider interface {
    // GetAuthURL returns the URL to redirect users to for authentication.
    // The state parameter is used to prevent CSRF attacks.
    GetAuthURL(state string) string
    
    // ExchangeCode exchanges an authorization code for user information.
    ExchangeCode(ctx context.Context, code string) (*domain.OAuthUserInfo, error)
    
    // GetProviderName returns the provider name (e.g., "google", "github").
    GetProviderName() domain.OAuthProvider
}

// OAuthStateStore manages OAuth state tokens for CSRF protection.
type OAuthStateStore interface {
    // GenerateState creates and stores a new state token.
    GenerateState(ctx context.Context) (string, error)
    
    // ValidateState checks if a state token is valid and deletes it.
    ValidateState(ctx context.Context, state string) (bool, error)
}
```

---

### Phase 3: Application Layer Updates

#### 3.1 Create OAuth Service Implementation

**File**: `internal/modules/auth/application/oauth_service.go` (NEW)

```go
package application

import (
    "context"
    "forum/internal/modules/auth/domain"
    authPort "forum/internal/modules/auth/ports"
    userDomain "forum/internal/modules/user/domain"
    userPort "forum/internal/modules/user/ports"
    "time"
)

// LoginWithOAuth handles OAuth authentication flow.
func (s *Service) LoginWithOAuth(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.Session, error) {
    // 1. Validate OAuth info
    if err := oauthInfo.Validate(); err != nil {
        return nil, err
    }
    
    // 2. Check if user exists with this OAuth provider
    user, err := s.userRepo.GetByOAuthProvider(ctx, string(oauthInfo.Provider), oauthInfo.ProviderID)
    if err == nil && user != nil {
        // User exists with OAuth, create session
        return s.createSessionForUser(ctx, user.ID)
    }
    
    // 3. Check if user exists with same email (existing password account)
    user, err = s.userRepo.GetByEmail(ctx, oauthInfo.Email)
    if err == nil && user != nil {
        // Link OAuth to existing account
        err = s.userRepo.LinkOAuthAccount(ctx, user.ID, string(oauthInfo.Provider), oauthInfo.ProviderID)
        if err != nil {
            return nil, err
        }
        return s.createSessionForUser(ctx, user.ID)
    }
    
    // 4. Create new user
    user = &userDomain.User{
        Email:           oauthInfo.Email,
        Username:        s.generateUniqueUsername(ctx, oauthInfo.Username),
        PasswordHash:    "", // No password for OAuth users
        OAuthProvider:   string(oauthInfo.Provider),
        OAuthProviderID: oauthInfo.ProviderID,
        Role:            userDomain.RoleUser,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
        IsActive:        true,
    }
    
    err = s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, err
    }
    
    return s.createSessionForUser(ctx, user.ID)
}

// createSessionForUser creates a new session for the given user ID.
func (s *Service) createSessionForUser(ctx context.Context, userID int) (*domain.Session, error) {
    // Delete existing sessions
    _ = s.sessionRepo.DeleteByUserID(ctx, userID)
    
    // Generate new session
    token, err := s.generateSessionToken()
    if err != nil {
        return nil, err
    }
    
    session := &domain.Session{
        UserID:    userID,
        Token:     token,
        ExpiresAt: time.Now().Add(s.sessionDuration),
        CreatedAt: time.Now(),
    }
    
    err = s.sessionRepo.Create(ctx, session)
    if err != nil {
        return nil, err
    }
    
    return session, nil
}

// generateUniqueUsername creates a unique username from OAuth display name.
func (s *Service) generateUniqueUsername(ctx context.Context, baseName string) string {
    // Sanitize and ensure uniqueness
    // Implementation: try baseName, then baseName_1, baseName_2, etc.
    // ...
}
```

---

### Phase 4: Adapter Layer - OAuth Providers

#### 4.1 Google OAuth Provider

**File**: `internal/modules/auth/adapters/oauth_google.go` (NEW)

```go
package adapters

import (
    "context"
    "encoding/json"
    "fmt"
    "forum/internal/modules/auth/domain"
    "forum/internal/modules/auth/ports"
    "forum/internal/platform/config"
    "io"
    "net/http"
    "net/url"
)

// GoogleOAuthProvider implements OAuth for Google.
type GoogleOAuthProvider struct {
    clientID     string
    clientSecret string
    redirectURL  string
    httpClient   *http.Client
}

// NewGoogleOAuthProvider creates a new Google OAuth provider.
func NewGoogleOAuthProvider(cfg config.GoogleOAuthConfig) *GoogleOAuthProvider {
    return &GoogleOAuthProvider{
        clientID:     cfg.ClientID,
        clientSecret: cfg.ClientSecret,
        redirectURL:  cfg.RedirectURL,
        httpClient:   &http.Client{},
    }
}

// GetAuthURL returns the Google OAuth authorization URL.
func (p *GoogleOAuthProvider) GetAuthURL(state string) string {
    params := url.Values{
        "client_id":     {p.clientID},
        "redirect_uri":  {p.redirectURL},
        "response_type": {"code"},
        "scope":         {"email profile"},
        "state":         {state},
        "access_type":   {"offline"},
        "prompt":        {"select_account"},
    }
    return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// ExchangeCode exchanges the authorization code for user info.
func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthUserInfo, error) {
    // 1. Exchange code for access token
    token, err := p.exchangeCodeForToken(ctx, code)
    if err != nil {
        return nil, err
    }
    
    // 2. Fetch user info from Google
    return p.fetchUserInfo(ctx, token)
}

func (p *GoogleOAuthProvider) GetProviderName() domain.OAuthProvider {
    return domain.OAuthProviderGoogle
}

// exchangeCodeForToken exchanges auth code for access token.
func (p *GoogleOAuthProvider) exchangeCodeForToken(ctx context.Context, code string) (string, error) {
    resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
        "client_id":     {p.clientID},
        "client_secret": {p.clientSecret},
        "code":          {code},
        "grant_type":    {"authorization_code"},
        "redirect_uri":  {p.redirectURL},
    })
    if err != nil {
        return "", domain.ErrOAuthProviderError
    }
    defer resp.Body.Close()
    
    var tokenResp struct {
        AccessToken string `json:"access_token"`
        Error       string `json:"error"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return "", domain.ErrOAuthProviderError
    }
    if tokenResp.Error != "" {
        return "", domain.ErrOAuthProviderError
    }
    
    return tokenResp.AccessToken, nil
}

// fetchUserInfo fetches user information from Google.
func (p *GoogleOAuthProvider) fetchUserInfo(ctx context.Context, token string) (*domain.OAuthUserInfo, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    resp, err := p.httpClient.Do(req)
    if err != nil {
        return nil, domain.ErrOAuthProviderError
    }
    defer resp.Body.Close()
    
    var userResp struct {
        ID            string `json:"id"`
        Email         string `json:"email"`
        Name          string `json:"name"`
        Picture       string `json:"picture"`
        VerifiedEmail bool   `json:"verified_email"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
        return nil, domain.ErrOAuthProviderError
    }
    
    return &domain.OAuthUserInfo{
        Provider:   domain.OAuthProviderGoogle,
        ProviderID: userResp.ID,
        Email:      userResp.Email,
        Username:   userResp.Name,
        AvatarURL:  userResp.Picture,
        IsVerified: userResp.VerifiedEmail,
    }, nil
}
```

#### 4.2 GitHub OAuth Provider

**File**: `internal/modules/auth/adapters/oauth_github.go` (NEW)

```go
package adapters

import (
    "context"
    "encoding/json"
    "forum/internal/modules/auth/domain"
    "forum/internal/platform/config"
    "net/http"
    "net/url"
)

// GitHubOAuthProvider implements OAuth for GitHub.
type GitHubOAuthProvider struct {
    clientID     string
    clientSecret string
    redirectURL  string
    httpClient   *http.Client
}

// NewGitHubOAuthProvider creates a new GitHub OAuth provider.
func NewGitHubOAuthProvider(cfg config.GitHubOAuthConfig) *GitHubOAuthProvider {
    return &GitHubOAuthProvider{
        clientID:     cfg.ClientID,
        clientSecret: cfg.ClientSecret,
        redirectURL:  cfg.RedirectURL,
        httpClient:   &http.Client{},
    }
}

// GetAuthURL returns the GitHub OAuth authorization URL.
func (p *GitHubOAuthProvider) GetAuthURL(state string) string {
    params := url.Values{
        "client_id":    {p.clientID},
        "redirect_uri": {p.redirectURL},
        "scope":        {"user:email"},
        "state":        {state},
    }
    return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// ExchangeCode exchanges the authorization code for user info.
func (p *GitHubOAuthProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthUserInfo, error) {
    // 1. Exchange code for access token
    token, err := p.exchangeCodeForToken(ctx, code)
    if err != nil {
        return nil, err
    }
    
    // 2. Fetch user info from GitHub
    userInfo, err := p.fetchUserInfo(ctx, token)
    if err != nil {
        return nil, err
    }
    
    // 3. Fetch email if not included in user info
    if userInfo.Email == "" {
        email, err := p.fetchPrimaryEmail(ctx, token)
        if err != nil {
            return nil, err
        }
        userInfo.Email = email
    }
    
    return userInfo, nil
}

func (p *GitHubOAuthProvider) GetProviderName() domain.OAuthProvider {
    return domain.OAuthProviderGitHub
}

// exchangeCodeForToken exchanges auth code for access token.
func (p *GitHubOAuthProvider) exchangeCodeForToken(ctx context.Context, code string) (string, error) {
    req, _ := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", nil)
    q := req.URL.Query()
    q.Add("client_id", p.clientID)
    q.Add("client_secret", p.clientSecret)
    q.Add("code", code)
    req.URL.RawQuery = q.Encode()
    req.Header.Set("Accept", "application/json")
    
    resp, err := p.httpClient.Do(req)
    if err != nil {
        return "", domain.ErrOAuthProviderError
    }
    defer resp.Body.Close()
    
    var tokenResp struct {
        AccessToken string `json:"access_token"`
        Error       string `json:"error"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return "", domain.ErrOAuthProviderError
    }
    if tokenResp.Error != "" {
        return "", domain.ErrOAuthProviderError
    }
    
    return tokenResp.AccessToken, nil
}

// fetchUserInfo fetches user information from GitHub.
func (p *GitHubOAuthProvider) fetchUserInfo(ctx context.Context, token string) (*domain.OAuthUserInfo, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Accept", "application/vnd.github+json")
    
    resp, err := p.httpClient.Do(req)
    if err != nil {
        return nil, domain.ErrOAuthProviderError
    }
    defer resp.Body.Close()
    
    var userResp struct {
        ID        int64  `json:"id"`
        Login     string `json:"login"`
        Email     string `json:"email"`
        Name      string `json:"name"`
        AvatarURL string `json:"avatar_url"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
        return nil, domain.ErrOAuthProviderError
    }
    
    username := userResp.Name
    if username == "" {
        username = userResp.Login
    }
    
    return &domain.OAuthUserInfo{
        Provider:   domain.OAuthProviderGitHub,
        ProviderID: fmt.Sprintf("%d", userResp.ID),
        Email:      userResp.Email,
        Username:   username,
        AvatarURL:  userResp.AvatarURL,
        IsVerified: true, // GitHub emails in user profile are verified
    }, nil
}

// fetchPrimaryEmail fetches user's primary email from GitHub.
func (p *GitHubOAuthProvider) fetchPrimaryEmail(ctx context.Context, token string) (string, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Accept", "application/vnd.github+json")
    
    resp, err := p.httpClient.Do(req)
    if err != nil {
        return "", domain.ErrOAuthProviderError
    }
    defer resp.Body.Close()
    
    var emails []struct {
        Email    string `json:"email"`
        Primary  bool   `json:"primary"`
        Verified bool   `json:"verified"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
        return "", domain.ErrOAuthProviderError
    }
    
    for _, e := range emails {
        if e.Primary && e.Verified {
            return e.Email, nil
        }
    }
    
    return "", domain.ErrOAuthEmailNotVerified
}
```

#### 4.3 OAuth State Store (In-Memory)

**File**: `internal/modules/auth/adapters/oauth_state_store.go` (NEW)

```go
package adapters

import (
    "context"
    "sync"
    "time"
    
    "github.com/gofrs/uuid/v5"
)

// InMemoryStateStore stores OAuth states in memory with expiration.
type InMemoryStateStore struct {
    states map[string]time.Time
    mu     sync.RWMutex
    ttl    time.Duration
}

// NewInMemoryStateStore creates a new in-memory state store.
func NewInMemoryStateStore(ttl time.Duration) *InMemoryStateStore {
    store := &InMemoryStateStore{
        states: make(map[string]time.Time),
        ttl:    ttl,
    }
    // Start cleanup goroutine
    go store.cleanupLoop()
    return store
}

// GenerateState creates and stores a new state token.
func (s *InMemoryStateStore) GenerateState(ctx context.Context) (string, error) {
    token, err := uuid.NewV4()
    if err != nil {
        return "", err
    }
    
    state := token.String()
    
    s.mu.Lock()
    s.states[state] = time.Now().Add(s.ttl)
    s.mu.Unlock()
    
    return state, nil
}

// ValidateState checks if a state token is valid and deletes it.
func (s *InMemoryStateStore) ValidateState(ctx context.Context, state string) (bool, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    expiry, exists := s.states[state]
    if !exists {
        return false, nil
    }
    
    delete(s.states, state)
    
    if time.Now().After(expiry) {
        return false, nil
    }
    
    return true, nil
}

// cleanupLoop periodically removes expired states.
func (s *InMemoryStateStore) cleanupLoop() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        s.mu.Lock()
        now := time.Now()
        for state, expiry := range s.states {
            if now.After(expiry) {
                delete(s.states, state)
            }
        }
        s.mu.Unlock()
    }
}
```

---

### Phase 5: HTTP Handler Updates

#### 5.1 Add OAuth HTTP Routes

**File**: `internal/modules/auth/adapters/http_handler.go`

Add new routes and handlers:
```go
// Add to HTTPHandler struct
type HTTPHandler struct {
    authService  ports.AuthService
    userService  userPorts.UserService
    templates    *template.Template
    googleOAuth  ports.OAuthProvider  // NEW
    githubOAuth  ports.OAuthProvider  // NEW
    stateStore   ports.OAuthStateStore // NEW
}

// Add new routes in RegisterRoutes
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
    // ... existing routes ...
    
    // OAuth routes
    router.HandleFunc("GET /auth/google", h.GoogleAuthRedirect)
    router.HandleFunc("GET /auth/google/callback", h.GoogleAuthCallback)
    router.HandleFunc("GET /auth/github", h.GitHubAuthRedirect)
    router.HandleFunc("GET /auth/github/callback", h.GitHubAuthCallback)
}

// GoogleAuthRedirect redirects to Google OAuth.
func (h *HTTPHandler) GoogleAuthRedirect(w http.ResponseWriter, r *http.Request) {
    if h.googleOAuth == nil {
        http.Error(w, "Google OAuth not configured", http.StatusServiceUnavailable)
        return
    }
    
    state, err := h.stateStore.GenerateState(r.Context())
    if err != nil {
        http.Error(w, "Failed to generate state", http.StatusInternalServerError)
        return
    }
    
    http.Redirect(w, r, h.googleOAuth.GetAuthURL(state), http.StatusTemporaryRedirect)
}

// GoogleAuthCallback handles Google OAuth callback.
func (h *HTTPHandler) GoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
    if h.googleOAuth == nil {
        http.Error(w, "Google OAuth not configured", http.StatusServiceUnavailable)
        return
    }
    
    // Validate state
    state := r.URL.Query().Get("state")
    valid, _ := h.stateStore.ValidateState(r.Context(), state)
    if !valid {
        http.Error(w, "Invalid state parameter", http.StatusBadRequest)
        return
    }
    
    // Get authorization code
    code := r.URL.Query().Get("code")
    if code == "" {
        http.Error(w, "No authorization code", http.StatusBadRequest)
        return
    }
    
    // Exchange code for user info
    oauthInfo, err := h.googleOAuth.ExchangeCode(r.Context(), code)
    if err != nil {
        http.Error(w, "Failed to authenticate with Google", http.StatusUnauthorized)
        return
    }
    
    // Login or register user
    session, err := h.authService.LoginWithOAuth(r.Context(), oauthInfo)
    if err != nil {
        http.Error(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
        return
    }
    
    // Set session cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "session_token",
        Value:    session.Token,
        Path:     "/",
        Expires:  session.ExpiresAt,
        HttpOnly: true,
        Secure:   false, // Set to true in production
        SameSite: http.SameSiteLaxMode,
    })
    
    // Redirect to home page
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GitHubAuthRedirect redirects to GitHub OAuth.
func (h *HTTPHandler) GitHubAuthRedirect(w http.ResponseWriter, r *http.Request) {
    // Similar to GoogleAuthRedirect
}

// GitHubAuthCallback handles GitHub OAuth callback.
func (h *HTTPHandler) GitHubAuthCallback(w http.ResponseWriter, r *http.Request) {
    // Similar to GoogleAuthCallback
}
```

---

### Phase 6: User Repository Updates

#### 6.1 Update SQLite Repository

**File**: `internal/modules/user/adapters/sqlite_repository.go`

Add OAuth methods:
```go
// GetByOAuthProvider retrieves a user by OAuth provider.
func (r *SQLiteUserRepository) GetByOAuthProvider(ctx context.Context, provider, providerID string) (*domain.User, error) {
    query := `
        SELECT id, public_id, email, username, password_hash, role,
               oauth_provider, oauth_provider_id, post_count, comment_count,
               created_at, updated_at, is_active
        FROM users
        WHERE oauth_provider = ? AND oauth_provider_id = ?
    `
    
    var user domain.User
    err := r.db.QueryRowContext(ctx, query, provider, providerID).Scan(
        &user.ID, &user.PublicID, &user.Email, &user.Username,
        &user.PasswordHash, &user.Role, &user.OAuthProvider,
        &user.OAuthProviderID, &user.PostCount, &user.CommentCount,
        &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
    )
    if err == sql.ErrNoRows {
        return nil, domain.ErrUserNotFound
    }
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

// LinkOAuthAccount links an OAuth account to an existing user.
func (r *SQLiteUserRepository) LinkOAuthAccount(ctx context.Context, userID int, provider, providerID string) error {
    query := `
        UPDATE users
        SET oauth_provider = ?, oauth_provider_id = ?, updated_at = ?
        WHERE id = ?
    `
    _, err := r.db.ExecContext(ctx, query, provider, providerID, time.Now(), userID)
    return err
}
```

---

### Phase 7: Wiring Updates

#### 7.1 Update ServiceContainer

**File**: `cmd/forum/wire/services.go`

```go
type ServiceContainer struct {
    // ... existing fields ...
    googleOAuth ports.OAuthProvider      // NEW
    githubOAuth ports.OAuthProvider      // NEW
    stateStore  ports.OAuthStateStore    // NEW
}

// Add accessor methods
func (sc *ServiceContainer) GoogleOAuth() ports.OAuthProvider    { return sc.googleOAuth }
func (sc *ServiceContainer) GitHubOAuth() ports.OAuthProvider    { return sc.githubOAuth }
func (sc *ServiceContainer) StateStore() ports.OAuthStateStore   { return sc.stateStore }
```

#### 7.2 Update initServices

```go
func initServices(repos *Repositories, cfg *config.Config, lgr *logger.Logger) *ServiceContainer {
    // ... existing initialization ...
    
    // Initialize OAuth providers (only if configured)
    var googleOAuth *authAdapters.GoogleOAuthProvider
    var githubOAuth *authAdapters.GitHubOAuthProvider
    
    if cfg.OAuth.Google.ClientID != "" {
        googleOAuth = authAdapters.NewGoogleOAuthProvider(cfg.OAuth.Google)
        lgr.Info("Google OAuth provider initialized")
    }
    
    if cfg.OAuth.GitHub.ClientID != "" {
        githubOAuth = authAdapters.NewGitHubOAuthProvider(cfg.OAuth.GitHub)
        lgr.Info("GitHub OAuth provider initialized")
    }
    
    // Initialize state store
    stateStore := authAdapters.NewInMemoryStateStore(10 * time.Minute)
    
    return &ServiceContainer{
        // ... existing fields ...
        googleOAuth: googleOAuth,
        githubOAuth: githubOAuth,
        stateStore:  stateStore,
    }
}
```

---

### Phase 8: Frontend Updates

#### 8.1 Update Login Template

**File**: `templates/login.html`

```html
{{define "content"}}
<div class="auth-container">
    <div class="auth-form">
        <h2>Login</h2>
        <form id="loginForm" method="POST" action="/auth/login">
            <div class="form-group">
                <label for="email">Email:</label>
                <input type="email" id="email" name="email" required>
            </div>

            <div class="form-group">
                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required>
            </div>

            <button type="submit">Login</button>
        </form>

        <!-- OAuth Buttons -->
        <div class="oauth-divider">
            <span>or continue with</span>
        </div>
        
        <div class="oauth-buttons">
            {{if .GoogleEnabled}}
            <a href="/auth/google" class="oauth-btn oauth-google">
                <svg viewBox="0 0 24 24" width="18" height="18">
                    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                </svg>
                Sign in with Google
            </a>
            {{end}}
            
            {{if .GitHubEnabled}}
            <a href="/auth/github" class="oauth-btn oauth-github">
                <svg viewBox="0 0 24 24" width="18" height="18">
                    <path fill="currentColor" d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                </svg>
                Sign in with GitHub
            </a>
            {{end}}
        </div>

        <p class="auth-link">Don't have an account? <a href="/register">Register here</a></p>
    </div>
</div>
{{end}}
```

#### 8.2 Add OAuth CSS

**File**: `static/css/auth.css` (add to existing or create)

```css
.oauth-divider {
    display: flex;
    align-items: center;
    margin: 1.5rem 0;
}

.oauth-divider::before,
.oauth-divider::after {
    content: '';
    flex: 1;
    border-bottom: 1px solid #ddd;
}

.oauth-divider span {
    padding: 0 1rem;
    color: #666;
    font-size: 0.875rem;
}

.oauth-buttons {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.oauth-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 500;
    transition: background-color 0.2s, border-color 0.2s;
}

.oauth-google {
    color: #333;
    background: #fff;
}

.oauth-google:hover {
    background: #f5f5f5;
    border-color: #ccc;
}

.oauth-github {
    color: #fff;
    background: #24292e;
    border-color: #24292e;
}

.oauth-github:hover {
    background: #2f363d;
}
```

---

## File Summary

### New Files to Create

| File | Description |
|------|-------------|
| `internal/modules/auth/domain/oauth.go` | OAuth domain types |
| `internal/modules/auth/ports/oauth.go` | OAuth provider interfaces |
| `internal/modules/auth/adapters/oauth_google.go` | Google OAuth implementation |
| `internal/modules/auth/adapters/oauth_github.go` | GitHub OAuth implementation |
| `internal/modules/auth/adapters/oauth_state_store.go` | CSRF state management |
| `internal/modules/auth/application/oauth_service.go` | OAuth business logic |

### Files to Modify

| File | Changes |
|------|---------|
| `internal/modules/auth/domain/errors.go` | Add OAuth errors |
| `internal/modules/user/domain/user.go` | Add OAuth fields |
| `internal/modules/auth/ports/service.go` | Add `LoginWithOAuth` method |
| `internal/modules/user/ports/repository.go` | Add OAuth methods |
| `internal/modules/auth/adapters/http_handler.go` | Add OAuth routes & handlers |
| `internal/modules/user/adapters/sqlite_repository.go` | Implement OAuth queries |
| `cmd/forum/wire/services.go` | Wire OAuth providers |
| `templates/login.html` | Add OAuth buttons |
| `templates/register.html` | Add OAuth buttons |
| `static/css/auth.css` | OAuth button styles |

---

## Environment Variables

```env
# Google OAuth
GOOGLE_OAUTH_CLIENT_ID=your-google-client-id
GOOGLE_OAUTH_CLIENT_SECRET=your-google-client-secret
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:8080/auth/google/callback

# GitHub OAuth
GITHUB_OAUTH_CLIENT_ID=your-github-client-id
GITHUB_OAUTH_CLIENT_SECRET=your-github-client-secret
GITHUB_OAUTH_REDIRECT_URL=http://localhost:8080/auth/github/callback
```

---

## OAuth Provider Setup

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Navigate to "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth client ID"
5. Configure consent screen if needed
6. Select "Web application"
7. Add authorized redirect URIs: `http://localhost:8080/auth/google/callback`
8. Copy Client ID and Client Secret

### GitHub OAuth Setup

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in application details:
   - Application name: Your Forum
   - Homepage URL: `http://localhost:8080`
   - Authorization callback URL: `http://localhost:8080/auth/github/callback`
4. Copy Client ID and generate Client Secret

---

## Testing Plan

### Unit Tests

1. **Domain tests**: Test `OAuthUserInfo.Validate()`
2. **Service tests**: Test `LoginWithOAuth` with various scenarios
3. **Repository tests**: Test OAuth queries

### Integration Tests

1. Mock OAuth provider responses
2. Test full OAuth flow with mocked external calls
3. Test account linking scenarios

### Manual Testing Checklist (from audit)

- [ ] Login with GitHub successfully
- [ ] Login with Google successfully  
- [ ] Create post with OAuth user, logout, verify post visible
- [ ] Login again with OAuth, verify all registered user rights
- [ ] Try creating account twice with same OAuth → should work (same user)
- [ ] Login without credentials shows error
- [ ] Registration still requires email and password for standard auth

---

## Timeline Estimate

| Phase | Estimated Time |
|-------|---------------|
| Phase 1: Domain Layer | 1-2 hours |
| Phase 2: Port Layer | 1 hour |
| Phase 3: Application Layer | 2-3 hours |
| Phase 4: OAuth Providers | 3-4 hours |
| Phase 5: HTTP Handlers | 2 hours |
| Phase 6: Repository Updates | 1-2 hours |
| Phase 7: Wiring | 1 hour |
| Phase 8: Frontend | 2 hours |
| Testing & Debugging | 3-4 hours |
| **Total** | **2-3 days** |

---

## Security Considerations

1. **State parameter**: Prevents CSRF attacks on OAuth flow
2. **Email verification**: Only accept verified emails from providers
3. **Token security**: Never log or expose OAuth tokens
4. **HTTPS in production**: Set `Secure: true` on cookies
5. **Rate limiting**: Apply to OAuth endpoints
6. **Account takeover prevention**: Require email verification before linking

---

## Future Enhancements

1. **Account unlinking**: Allow users to disconnect OAuth providers
2. **Multiple providers per account**: Support linking multiple OAuth accounts
3. **Additional providers**: Facebook, Discord, Twitter
4. **Profile picture sync**: Import avatar from OAuth provider
5. **Refresh tokens**: Handle token refresh for long sessions
