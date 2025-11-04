package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
	"gorm.io/gorm"
)

// OAuthProvider represents an OAuth2 provider
type OAuthProvider interface {
	GetAuthURL(state string) string
	GetUserInfo(ctx context.Context, code string) (*SocialUser, error)
}

// SocialUser represents a user from a social provider
type SocialUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Avatar        string `json:"avatar"`
	Provider      string `json:"provider"`
	ProviderID    string `json:"provider_id"`
	EmailVerified bool   `json:"email_verified"`
}

// GoogleProvider implements Google OAuth2
type GoogleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a new Google OAuth2 provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) *GoogleProvider {
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleProvider{config: config}
}

// GetAuthURL returns the OAuth authorization URL
func (g *GoogleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetUserInfo retrieves user information from Google
func (g *GoogleProvider) GetUserInfo(ctx context.Context, code string) (*SocialUser, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &SocialUser{
		ID:            userInfo.ID,
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		FirstName:     userInfo.GivenName,
		LastName:      userInfo.FamilyName,
		Avatar:        userInfo.Picture,
		Provider:      "google",
		ProviderID:    userInfo.ID,
		EmailVerified: userInfo.VerifiedEmail,
	}, nil
}

// GitHubProvider implements GitHub OAuth2
type GitHubProvider struct {
	config *oauth2.Config
}

// NewGitHubProvider creates a new GitHub OAuth2 provider
func NewGitHubProvider(clientID, clientSecret, redirectURL string, scopes []string) *GitHubProvider {
	if len(scopes) == 0 {
		scopes = []string{"user:email"}
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     github.Endpoint,
	}

	return &GitHubProvider{config: config}
}

// GetAuthURL returns the OAuth authorization URL
func (g *GitHubProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

// GetUserInfo retrieves user information from GitHub
func (g *GitHubProvider) GetUserInfo(ctx context.Context, code string) (*SocialUser, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := g.config.Client(ctx, token)

	// Get user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID     int    `json:"id"`
		Login  string `json:"login"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// Get email if not in user info
	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, email := range emails {
					if email.Primary {
						userInfo.Email = email.Email
						break
					}
				}
			}
		}
	}

	return &SocialUser{
		ID:            fmt.Sprintf("%d", userInfo.ID),
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		Avatar:        userInfo.Avatar,
		Provider:      "github",
		ProviderID:    fmt.Sprintf("%d", userInfo.ID),
		EmailVerified: userInfo.Email != "",
	}, nil
}

// FacebookProvider implements Facebook OAuth2
type FacebookProvider struct {
	config *oauth2.Config
}

// NewFacebookProvider creates a new Facebook OAuth2 provider
func NewFacebookProvider(clientID, clientSecret, redirectURL string, scopes []string) *FacebookProvider {
	if len(scopes) == 0 {
		scopes = []string{"email", "public_profile"}
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/v18.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v18.0/oauth/access_token",
		},
	}

	return &FacebookProvider{config: config}
}

// GetAuthURL returns the OAuth authorization URL
func (f *FacebookProvider) GetAuthURL(state string) string {
	return f.config.AuthCodeURL(state)
}

// GetUserInfo retrieves user information from Facebook
func (f *FacebookProvider) GetUserInfo(ctx context.Context, code string) (*SocialUser, error) {
	token, err := f.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := f.config.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,first_name,last_name,picture")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Picture   struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &SocialUser{
		ID:            userInfo.ID,
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		FirstName:     userInfo.FirstName,
		LastName:      userInfo.LastName,
		Avatar:        userInfo.Picture.Data.URL,
		Provider:      "facebook",
		ProviderID:    userInfo.ID,
		EmailVerified: userInfo.Email != "",
	}, nil
}

// MicrosoftProvider implements Microsoft OAuth2
type MicrosoftProvider struct {
	config *oauth2.Config
}

// NewMicrosoftProvider creates a new Microsoft OAuth2 provider
func NewMicrosoftProvider(clientID, clientSecret, redirectURL string, scopes []string) *MicrosoftProvider {
	if len(scopes) == 0 {
		scopes = []string{"User.Read"}
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}

	return &MicrosoftProvider{config: config}
}

// GetAuthURL returns the OAuth authorization URL
func (m *MicrosoftProvider) GetAuthURL(state string) string {
	return m.config.AuthCodeURL(state)
}

// GetUserInfo retrieves user information from Microsoft
func (m *MicrosoftProvider) GetUserInfo(ctx context.Context, code string) (*SocialUser, error) {
	token, err := m.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := m.config.Client(ctx, token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID                string `json:"id"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
		DisplayName       string `json:"displayName"`
		GivenName         string `json:"givenName"`
		Surname           string `json:"surname"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	email := userInfo.Mail
	if email == "" {
		email = userInfo.UserPrincipalName
	}

	return &SocialUser{
		ID:            userInfo.ID,
		Email:         email,
		Name:          userInfo.DisplayName,
		FirstName:     userInfo.GivenName,
		LastName:      userInfo.Surname,
		Provider:      "microsoft",
		ProviderID:    userInfo.ID,
		EmailVerified: email != "",
	}, nil
}

// OAuthManager manages OAuth2 providers
type OAuthManager struct {
	providers map[string]OAuthProvider
	db        *gorm.DB
}

// NewOAuthManager creates a new OAuth manager
func NewOAuthManager(db *gorm.DB) *OAuthManager {
	return &OAuthManager{
		providers: make(map[string]OAuthProvider),
		db:        db,
	}
}

// RegisterProvider registers an OAuth provider
func (om *OAuthManager) RegisterProvider(name string, provider OAuthProvider) {
	om.providers[name] = provider
}

// GetProvider returns an OAuth provider by name
func (om *OAuthManager) GetProvider(name string) (OAuthProvider, error) {
	provider, exists := om.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// SocialAccount represents a social account linked to a user
type SocialAccount struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	Provider       string     `gorm:"not null;index" json:"provider"`
	ProviderID     string     `gorm:"not null" json:"provider_id"`
	Email          string     `json:"email"`
	Name           string     `json:"name"`
	Avatar         string     `json:"avatar"`
	AccessToken    string     `gorm:"type:text" json:"-"`
	RefreshToken   string     `gorm:"type:text" json:"-"`
	TokenExpiresAt *time.Time `json:"token_expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// FindOrCreateUser finds or creates a user from a social account
func (om *OAuthManager) FindOrCreateUser(socialUser *SocialUser) (uint, error) {
	var account SocialAccount
	err := om.db.Where("provider = ? AND provider_id = ?", socialUser.Provider, socialUser.ProviderID).
		First(&account).Error

	if err == nil {
		// Account exists, return user ID
		return account.UserID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return 0, err
	}

	// Create new account and user
	// This would typically create a User record as well
	// For now, we'll just create the social account
	account = SocialAccount{
		Provider:   socialUser.Provider,
		ProviderID: socialUser.ProviderID,
		Email:      socialUser.Email,
		Name:       socialUser.Name,
		Avatar:     socialUser.Avatar,
	}

	if err := om.db.Create(&account).Error; err != nil {
		return 0, err
	}

	return account.UserID, nil
}
