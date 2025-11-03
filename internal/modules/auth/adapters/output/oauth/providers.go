package oauth
// Package oauth provides OAuth provider implementations.
// These are adapters for the auth module's OAuth outbound port.
package oauth

import (
	"context"

	"forum/internal/modules/auth/domain"
	"forum/internal/modules/auth/ports/output"
)

// GoogleProvider implements OAuth for Google.
type GoogleProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

// NewGoogleProvider creates a new Google OAuth provider.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) output.OAuthProvider {
	return &GoogleProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
	}
}

// GetAuthURL returns the Google OAuth authorization URL.
func (p *GoogleProvider) GetAuthURL(state string) string {
	// TODO: Implement Google OAuth URL generation
	return ""
}

// ExchangeCode exchanges an authorization code for an access token.
func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	// TODO: Implement Google OAuth code exchange
	return "", nil
}

// GetUserInfo retrieves user information from Google.
func (p *GoogleProvider) GetUserInfo(ctx context.Context, accessToken string) (*domain.OAuthCredentials, error) {
	// TODO: Implement Google user info retrieval
	return nil, nil
}

// GitHubProvider implements OAuth for GitHub.
type GitHubProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

// NewGitHubProvider creates a new GitHub OAuth provider.
func NewGitHubProvider(clientID, clientSecret, redirectURL string) output.OAuthProvider {
	return &GitHubProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL.
func (p *GitHubProvider) GetAuthURL(state string) string {
	// TODO: Implement GitHub OAuth URL generation
	return ""
}

// ExchangeCode exchanges an authorization code for an access token.
func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	// TODO: Implement GitHub OAuth code exchange
	return "", nil
}

// GetUserInfo retrieves user information from GitHub.
func (p *GitHubProvider) GetUserInfo(ctx context.Context, accessToken string) (*domain.OAuthCredentials, error) {
	// TODO: Implement GitHub user info retrieval
	return nil, nil
}
