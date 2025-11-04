package auth

import (
	"dolphin/internal/auth"
	"gorm.io/gorm"
)

// OAuthManager represents the OAuth manager
type OAuthManager = auth.OAuthManager

// NewOAuthManager creates a new OAuth manager
func NewOAuthManager(db *gorm.DB) *OAuthManager {
	return auth.NewOAuthManager(db)
}

// OAuthProvider represents an OAuth provider interface
type OAuthProvider = auth.OAuthProvider

// SocialUser represents a user from a social provider
type SocialUser = auth.SocialUser

// SocialAccount represents a social account
type SocialAccount = auth.SocialAccount

// GoogleProvider represents a Google OAuth2 provider
type GoogleProvider = auth.GoogleProvider

// NewGoogleProvider creates a new Google OAuth2 provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) *GoogleProvider {
	return auth.NewGoogleProvider(clientID, clientSecret, redirectURL, scopes)
}

// GitHubProvider represents a GitHub OAuth2 provider
type GitHubProvider = auth.GitHubProvider

// NewGitHubProvider creates a new GitHub OAuth2 provider
func NewGitHubProvider(clientID, clientSecret, redirectURL string, scopes []string) *GitHubProvider {
	return auth.NewGitHubProvider(clientID, clientSecret, redirectURL, scopes)
}

// FacebookProvider represents a Facebook OAuth2 provider
type FacebookProvider = auth.FacebookProvider

// NewFacebookProvider creates a new Facebook OAuth2 provider
func NewFacebookProvider(clientID, clientSecret, redirectURL string, scopes []string) *FacebookProvider {
	return auth.NewFacebookProvider(clientID, clientSecret, redirectURL, scopes)
}

// MicrosoftProvider represents a Microsoft OAuth2 provider
type MicrosoftProvider = auth.MicrosoftProvider

// NewMicrosoftProvider creates a new Microsoft OAuth2 provider
func NewMicrosoftProvider(clientID, clientSecret, redirectURL string, scopes []string) *MicrosoftProvider {
	return auth.NewMicrosoftProvider(clientID, clientSecret, redirectURL, scopes)
}

