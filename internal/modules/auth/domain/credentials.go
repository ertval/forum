// Package domain - credentials
// This file contains value objects related to authentication credentials.
package domain

// Credentials represents user authentication credentials.
type Credentials struct {
	Email    string
	Password string
}

// Validate checks if the credentials are valid.
func (c Credentials) Validate() error {
	// TODO: Implement validation
	return nil
}

// OAuthCredentials represents OAuth authentication credentials.
type OAuthCredentials struct {
	Provider     string // "google", "github"
	ProviderID   string // User ID from OAuth provider
	Email        string
	Name         string
	AccessToken  string
	RefreshToken string
}
