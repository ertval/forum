// Package application implements the auth service business logic.
// It orchestrates domain entities and repository operations to fulfill use cases.
package application

import (
	"context"
	"errors"
	"forum/internal/modules/auth/domain"
	authPort "forum/internal/modules/auth/ports"
	userDomain "forum/internal/modules/user/domain"
	userPort "forum/internal/modules/user/ports"
	"forum/internal/platform/validator"
	"time"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service implements the AuthService interface.
// It coordinates authentication operations using domain logic and repositories.
type Service struct {
	sessionRepo     authPort.SessionRepository
	userRepo        userPort.UserRepository
	sessionDuration time.Duration
}

// NewService creates a new auth service with the required dependencies.
func NewService(
	sessionRepo authPort.SessionRepository,
	userRepo userPort.UserRepository,
	sessionDuration time.Duration,
) *Service {
	return &Service{
		sessionRepo:     sessionRepo,
		userRepo:        userRepo,
		sessionDuration: sessionDuration,
	}
}

// Register creates a new user account.
// It validates input, checks for duplicates, hashes the password, and creates a session.
func (s *Service) Register(ctx context.Context, email, username, password string) (userID int, session *domain.Session, err error) {
	// 1. Validate input
	creds := &domain.Credentials{Email: email, Password: password}
	err = ValidateCredentials(creds)
	if err != nil {
		return 0, nil, err
	}

	validation := validator.New()
	validation.Required("username", username)
	if username != "" {
		validation.Username("username", username)
	}
	if !validation.Valid() {
		for field := range validation.Errors() {
			if field == "username" {
				return 0, nil, domain.ErrInvalidUsername
			}
		}
	}

	// 2. Check if email or username already exists
	emailExists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return 0, nil, err
	}
	if emailExists {
		return 0, nil, domain.ErrUserAlreadyExists
	}

	usernameExists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return 0, nil, err
	}
	if usernameExists {
		return 0, nil, domain.ErrUserAlreadyExists
	}

	// 3. Hash the password
	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return 0, nil, err
	}

	// 4. Create user in repository
	user := &userDomain.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         userDomain.RoleUser, // Default role is User
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return 0, nil, err
	}

	// 5. Create session for the new user
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		return 0, nil, err
	}

	session = &domain.Session{
		ID:        sessionToken, // Use the token as the ID for simplicity
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(s.sessionDuration),
		CreatedAt: time.Now(),
		IPAddress: "", // Will be set from request in the HTTP handler
		UserAgent: "", // Will be set from request in the HTTP handler
	}

	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return 0, nil, err
	}

	return user.ID, session, nil
}

// Login authenticates a user with email and password.
// It validates credentials and creates a new session on success.
func (s *Service) Login(ctx context.Context, email, password string) (*domain.Session, error) {
	// 1. Validate input
	creds := &domain.Credentials{Email: email, Password: password}
	err := ValidateCredentials(creds)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 2. Retrieve user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// If user doesn't exist, return invalid credentials to avoid user enumeration
		return nil, domain.ErrInvalidCredentials
	}

	// 3. Compare password hash using bcrypt
	err = s.comparePassword(user.PasswordHash, password)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 4. If valid, delete any existing sessions for this user (one session per user)
	err = s.sessionRepo.DeleteByUserID(ctx, user.ID)
	if err != nil {
		// If we can't delete existing sessions, continue anyway
		// This might result in multiple active sessions, but login should still work
	}

	// 5. Create new session
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        sessionToken, // Use the token as the ID for simplicity
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(s.sessionDuration),
		CreatedAt: time.Now(),
		IPAddress: "", // Will be set from request in the HTTP handler
		UserAgent: "", // Will be set from request in the HTTP handler
	}

	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Logout invalidates the session with the given token.
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	// 1. Delete session from repository
	return s.sessionRepo.Delete(ctx, sessionToken)
}

// ValidateSession checks if a session token is valid and not expired.
func (s *Service) ValidateSession(ctx context.Context, sessionToken string) (*domain.Session, error) {
	// 1. Retrieve session by token
	session, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// 2. Check if session is expired
	if session.IsExpired() {
		// Clean up expired session
		_ = s.sessionRepo.Delete(ctx, sessionToken) // Best effort cleanup
		return nil, domain.ErrSessionExpired
	}

	// 3. Return session
	return session, nil
}

// RefreshSession extends the expiration time of an existing session.
func (s *Service) RefreshSession(ctx context.Context, sessionToken string) (*domain.Session, error) {
	// 1. Retrieve session by token
	session, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// 2. Check if session is expired
	if session.IsExpired() {
		// Clean up expired session
		_ = s.sessionRepo.Delete(ctx, sessionToken) // Best effort cleanup
		return nil, domain.ErrSessionExpired
	}

	// 3. Update expiration time
	session.ExpiresAt = time.Now().Add(s.sessionDuration)

	// 4. Save updated session
	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, err
	}

	// 5. Return updated session
	return session, nil
}

// GetSession retrieves a session by its token.
func (s *Service) GetSession(ctx context.Context, sessionToken string) (*domain.Session, error) {
	// 1. Retrieve session from repository
	session, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// 2. Check if session is expired
	if session.IsExpired() {
		// Clean up expired session
		_ = s.sessionRepo.Delete(ctx, sessionToken) // Best effort cleanup
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}

// generateSessionToken generates a unique session token.
func (s *Service) generateSessionToken() (string, error) {
	token, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

// hashPassword hashes a plaintext password using bcrypt.
func (s *Service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// comparePassword compares a plaintext password with a hash.
func (s *Service) comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidateCredentials validates the credentials.
// Returns an error if the credentials are invalid.
func ValidateCredentials(c *domain.Credentials) error {
	validator := validator.New()

	validator.Required("email", c.Email)
	if c.Email != "" {
		validator.Email("email", c.Email)
	}

	validator.Required("password", c.Password)
	if c.Password != "" {
		// Using minimum 6 characters for basic validation, can be increased for production
		validator.Password("password", c.Password, 6)
	}

	if !validator.Valid() {
		// Convert validator errors to domain-specific errors
		for field, msg := range validator.Errors() {
			if field == "email" {
				if msg == "This field is required" {
					return domain.ErrInvalidEmail
				}
				return domain.ErrInvalidEmail
			}
			if field == "password" {
				if msg == "This field is required" {
					return domain.ErrWeakPassword
				}
				return domain.ErrWeakPassword
			}
			// Fallback to generic error
			return errors.New(msg)
		}
	}

	return nil
}
