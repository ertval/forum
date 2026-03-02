// Package application implements the user service business logic.
package application

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"forum/internal/modules/user/domain"
	"forum/internal/modules/user/ports"

	"golang.org/x/crypto/bcrypt"
)

// Package-level regexes for input validation (stdlib only).
var (
	userEmailRegex    = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	userNamePartRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
)

// Service implements the UserService interface.
type Service struct {
	userRepo ports.UserRepository
}

// NewService creates a new user service.
func NewService(userRepo ports.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with the given details.
// Returns the created user's internal ID or an error.
func (s *Service) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	user := &domain.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         domain.RoleUser, // Default role is User
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("creating user: %w", err)
	}

	return user.ID, nil
}

// GetByID retrieves a user by their internal ID (for internal use only).
// TODO: Implement user retrieval.
func (s *Service) GetByID(ctx context.Context, userID int) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetByPublicID retrieves a user by their public UUID (for external API access).
func (s *Service) GetByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	return s.userRepo.GetByPublicID(ctx, publicID)
}

// GetByUsername retrieves a user by their username.
// TODO: Implement username-based retrieval.
func (s *Service) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// GetByEmail retrieves a user by their email address.
// TODO: Implement email-based retrieval.
func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// UpdateRole updates a user's role.
// Validates the new role and updates the user in the repository.
func (s *Service) UpdateRole(ctx context.Context, userID int, newRole domain.Role) error {
	// Validate role
	if !isValidRole(newRole) {
		return domain.ErrInvalidRole
	}

	// Get the user first to ensure they exist
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user for role update: %w", err)
	}

	// Update role and timestamp
	user.Role = newRole
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// DeactivateUser deactivates a user account.
// Sets IsActive to false and updates the repository.
func (s *Service) DeactivateUser(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user for deactivation: %w", err)
	}

	if !user.IsActive {
		return nil // Already inactive
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// ActivateUser reactivates a user account.
// Sets IsActive to true and updates the repository.
func (s *Service) ActivateUser(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user for activation: %w", err)
	}

	if user.IsActive {
		return nil // Already active
	}

	user.IsActive = true
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// isValidRole checks if a role string is a valid role.
func isValidRole(role domain.Role) bool {
	switch role {
	case domain.RoleGuest, domain.RoleUser, domain.RoleModerator, domain.RoleAdmin:
		return true
	default:
		return false
	}
}

// userValidateUsername validates a username against format rules.
func userValidateUsername(username string) error {
	parts := strings.Fields(username)
	if len(parts) == 0 {
		return domain.ErrInvalidUsername
	}
	for _, part := range parts {
		if !userNamePartRegex.MatchString(part) {
			return domain.ErrInvalidUsername
		}
	}
	return nil
}

// userValidatePassword checks password strength.
// Returns (message, invalid) where invalid is true if the password does not meet requirements.
func userValidatePassword(password string, minLen int) (string, bool) {
	if utf8.RuneCountInString(password) < minLen {
		return fmt.Sprintf("Password must be at least %d characters long", minLen), true
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	var missing []string
	if !hasUpper {
		missing = append(missing, "an uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "a lowercase letter")
	}
	if !hasDigit {
		missing = append(missing, "a digit")
	}
	if len(missing) > 0 {
		return "Password must contain at least " + strings.Join(missing, ", "), true
	}
	return "", false
}

// userSanitize removes potentially dangerous characters from user-supplied input.
var (
	userReScript = regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	userReStyle  = regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
	userReTags   = regexp.MustCompile(`<[^>]+>`)
	userReSpace  = regexp.MustCompile(`\s+`)
)

func userSanitize(input string) string {
	if input == "" {
		return ""
	}
	s := html.UnescapeString(input)
	s = userReScript.ReplaceAllString(s, "")
	s = userReStyle.ReplaceAllString(s, "")
	s = userReTags.ReplaceAllString(s, "")
	var b strings.Builder
	for _, r := range s {
		if r >= 0x20 || r == '\t' || r == '\n' || r == '\r' {
			b.WriteRune(r)
		}
	}
	s = b.String()
	s = userReSpace.ReplaceAllString(strings.TrimSpace(s), " ")
	return s
}

// ListUsers returns a paginated list of users.
// TODO: Implement user listing.
func (s *Service) ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	return s.userRepo.List(ctx, offset, limit)
}

// IncrementPostCount atomically increments the user's post count.
func (s *Service) IncrementPostCount(ctx context.Context, userID int) error {
	return s.userRepo.IncrementPostCount(ctx, userID)
}

// DecrementPostCount atomically decrements the user's post count.
func (s *Service) DecrementPostCount(ctx context.Context, userID int) error {
	return s.userRepo.DecrementPostCount(ctx, userID)
}

// IncrementCommentCount atomically increments the user's comment count.
func (s *Service) IncrementCommentCount(ctx context.Context, userID int) error {
	return s.userRepo.IncrementCommentCount(ctx, userID)
}

// DecrementCommentCount atomically decrements the user's comment count.
func (s *Service) DecrementCommentCount(ctx context.Context, userID int) error {
	return s.userRepo.DecrementCommentCount(ctx, userID)
}

// ExistsByEmail checks if a user with the given email exists.
func (s *Service) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.userRepo.ExistsByEmail(ctx, email)
}

// ExistsByUsername checks if a user with the given username exists.
func (s *Service) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return s.userRepo.ExistsByUsername(ctx, username)
}

// IncrementReactionCount atomically increments the user's reaction count.
func (s *Service) IncrementReactionCount(ctx context.Context, userID int) error {
	return s.userRepo.IncrementReactionCount(ctx, userID)
}

// DecrementReactionCount atomically decrements the user's reaction count.
func (s *Service) DecrementReactionCount(ctx context.Context, userID int) error {
	return s.userRepo.DecrementReactionCount(ctx, userID)
}

// UpdateSettings updates profile settings for a user identified by public UUID.
func (s *Service) UpdateSettings(ctx context.Context, publicID, username, email, newPassword, avatarPath string) (*domain.User, error) {
	if publicID == "" {
		return nil, domain.ErrUserNotFound
	}

	user, err := s.userRepo.GetByPublicID(ctx, publicID)
	if err != nil || user == nil {
		return nil, domain.ErrUserNotFound
	}

	username = strings.TrimSpace(userSanitize(username))
	email = strings.ToLower(strings.TrimSpace(userSanitize(email)))

	// Validate username
	if strings.TrimSpace(username) == "" {
		return nil, domain.ErrInvalidUsername
	}
	if err := userValidateUsername(username); err != nil {
		return nil, err
	}
	// Validate email
	if strings.TrimSpace(email) == "" {
		return nil, domain.ErrInvalidEmail
	}
	if !userEmailRegex.MatchString(strings.ToLower(strings.TrimSpace(email))) {
		return nil, domain.ErrInvalidEmail
	}
	// Validate password if provided
	if newPassword != "" {
		if msg, invalid := userValidatePassword(newPassword, 8); invalid {
			return nil, &domain.PasswordValidationError{Message: msg}
		}
	}

	if username != user.Username {
		existingUser, getErr := s.userRepo.GetByUsername(ctx, username)
		if getErr == nil && existingUser != nil && existingUser.ID != user.ID {
			return nil, domain.ErrUsernameAlreadyExists
		}
	}

	if email != user.Email {
		existingUser, getErr := s.userRepo.GetByEmail(ctx, email)
		if getErr == nil && existingUser != nil && existingUser.ID != user.ID {
			return nil, domain.ErrEmailAlreadyExists
		}
	}

	user.Username = username
	user.Email = email
	user.AvatarPath = avatarPath
	user.UpdatedAt = time.Now()

	if newPassword != "" {
		passwordHash, hashErr := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, fmt.Errorf("hashing password: %w", hashErr)
		}
		user.PasswordHash = string(passwordHash)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user settings: %w", err)
	}

	if user.AvatarPath != "" {
		user.AvatarURL = domain.AvatarURLPrefix + user.AvatarPath
	} else {
		user.AvatarURL = ""
	}

	return user, nil
}
