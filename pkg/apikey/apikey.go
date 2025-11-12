package apikey

import (
	"github.com/mrhoseah/dolphin/internal/apikey"
	"gorm.io/gorm"
)

// APIKeyManager represents the API key manager
type APIKeyManager = apikey.APIKeyManager

// NewAPIKeyManager creates a new API key manager
func NewAPIKeyManager(db *gorm.DB) *APIKeyManager {
	return apikey.NewAPIKeyManager(db)
}

// APIKey represents an API key
type APIKey = apikey.APIKey

