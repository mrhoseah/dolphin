package auth

import (
	"dolphin/internal/auth"
)

// AuthManager represents the authentication manager
type AuthManager = auth.AuthManager

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return auth.NewAuthManager()
}
