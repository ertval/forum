package domain

import (
	"testing"
	"time"
)

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected bool
	}{
		{
			name: "session not expired",
			session: &Session{
				ExpiresAt: time.Now().Add(1 * time.Hour), // Expires in 1 hour
			},
			expected: false,
		},
		{
			name: "session expired",
			session: &Session{
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			},
			expected: true,
		},
		{
			name: "session expires now",
			session: &Session{
				ExpiresAt: time.Now().Add(-1 * time.Nanosecond), // Just expired
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.session.IsExpired()
			if result != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSession_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected bool
	}{
		{
			name: "valid session",
			session: &Session{
				ID:        1,
				UserID:    1,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "missing ID",
			session: &Session{
				ID:        0,
				UserID:    1,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "invalid UserID",
			session: &Session{
				ID:        1,
				UserID:    0,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "expired session",
			session: &Session{
				ID:        1,
				UserID:    1,
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.session.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCredentials(t *testing.T) {
	// Test that Credentials can be created and used as expected
	creds := &Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	if creds.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got '%s'", creds.Email)
	}

	if creds.Password != "password123" {
		t.Errorf("Expected Password to be 'password123', got '%s'", creds.Password)
	}
}