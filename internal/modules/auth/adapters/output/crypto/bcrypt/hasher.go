package bcrypt
// Package bcrypt provides bcrypt implementation of the password hasher.
// This is an adapter for the auth module's outbound port.
package bcrypt

import (
	"golang.org/x/crypto/bcrypt"

	"forum/internal/modules/auth/ports/output"
)

// PasswordHasher implements the PasswordHasher interface using bcrypt.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new bcrypt password hasher.
func NewPasswordHasher() output.PasswordHasher {
	return &PasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// Hash generates a bcrypt hash from a plain text password.
func (h *PasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare compares a plain text password with a bcrypt hash.
func (h *PasswordHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
