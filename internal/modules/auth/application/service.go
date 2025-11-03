package application
// Package application contains the application services that orchestrate
// domain logic and coordinate with external dependencies through ports.
package application

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"

	"forum/internal/modules/auth/domain"
	"forum/internal/modules/auth/ports/input"
	"forum/internal/modules/auth/ports/output"
	"forum/internal/platform/validator"
)

// Service implements the AuthService interface.
type Service struct {
	sessionRepo    output.SessionRepository
	userRepo       output.UserRepository
	passwordHasher output.PasswordHasher
	oauthProviders map[string]output.OAuthProvider
	sessionTTL     time.Duration
}

// NewService creates a new auth service instance.
func NewService(
	sessionRepo output.SessionRepository,
	userRepo output.UserRepository,
	passwordHasher output.PasswordHasher,
	sessionTTL time.Duration,
) input.AuthService {
	return &Service{
		sessionRepo:    sessionRepo,
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		oauthProviders: make(map[string]output.OAuthProvider),
		sessionTTL:     sessionTTL,
	}
}

// AddOAuthProvider registers an OAuth provider.
func (s *Service) AddOAuthProvider(name string, provider output.OAuthProvider) {
	s.oauthProviders[name] = provider
}

// Register implements the Register use case.
func (s *Service) Register(ctx context.Context, email, username, password string) error {
	// TODO: Implement registration logic
	// 1. Validate input
	// 2. Check if email/username exists
	// 3. Hash password
	// 4. Create user
	return nil
}

// Login implements the Login use case.
func (s *Service) Login(ctx context.Context, email, password string) (*domain.Session, error) {
	// TODO: Implement login logic
	// 1. Validate input
	// 2. Get user by email
	// 3. Verify password
	// 4. Create session
	return nil, nil
}

// Logout implements the Logout use case.
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	// TODO: Implement logout logic
	return nil
}

// ValidateSession implements the ValidateSession use case.
func (s *Service) ValidateSession(ctx context.Context, sessionToken string) (string, error) {
	// TODO: Implement session validation
	return "", nil
}

// RefreshSession implements the RefreshSession use case.
func (s *Service) RefreshSession(ctx context.Context, sessionToken string) (*domain.Session, error) {
	// TODO: Implement session refresh
	return nil, nil
}

// LoginWithOAuth implements the LoginWithOAuth use case.
func (s *Service) LoginWithOAuth(ctx context.Context, provider string, code string) (*domain.Session, error) {
	// TODO: Implement OAuth login
	return nil, nil
}

// createSession creates a new session for a user.
func (s *Service) createSession(ctx context.Context, userID, ipAddress, userAgent string) (*domain.Session, error) {
	token, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        uuid.Must(uuid.NewV4()),
		UserID:    userID,
		Token:     token.String(),
		ExpiresAt: time.Now().Add(s.sessionTTL),
		CreatedAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// validateInput validates registration input.
func validateInput(email, username, password string) error {
	var errs validator.ValidationErrors

	if !validator.Required(email) {
		errs.Add("email", "email is required")
	} else if !validator.IsValidEmail(email) {
		errs.Add("email", "invalid email format")
	}

	if !validator.Required(username) {
		errs.Add("username", "username is required")
	} else if !validator.IsValidUsername(username) {
		errs.Add("username", "username must be 3-30 characters and contain only letters, numbers, and underscores")
	}

	if !validator.Required(password) {
		errs.Add("password", "password is required")
	} else if !validator.IsValidPassword(password) {
		errs.Add("password", "password must be at least 8 characters and contain uppercase, lowercase, digit, and special character")
	}

	if errs.HasErrors() {
		return errs
	}

	return nil
}
