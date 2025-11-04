package apikey

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// APIKey represents an API key
type APIKey struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Key         string    `gorm:"uniqueIndex;not null" json:"key"`
	Hash        string    `gorm:"not null" json:"-"`
	UserID      *uint     `gorm:"index" json:"user_id,omitempty"`
	Scopes      []string  `gorm:"type:json" json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	RateLimit   int       `gorm:"default:1000" json:"rate_limit"` // requests per hour
	UsageCount  int64     `gorm:"default:0" json:"usage_count"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APIKeyManager manages API keys
type APIKeyManager struct {
	db *gorm.DB
}

// NewAPIKeyManager creates a new API key manager
func NewAPIKeyManager(db *gorm.DB) *APIKeyManager {
	return &APIKeyManager{db: db}
}

// Create creates a new API key
func (m *APIKeyManager) Create(name string, userID *uint, scopes []string, expiresAt *time.Time, rateLimit int) (*APIKey, string, error) {
	// Generate key
	key, hash, err := generateAPIKey()
	if err != nil {
		return nil, "", err
	}

	apiKey := &APIKey{
		Name:      name,
		Key:       key,
		Hash:      hash,
		UserID:    userID,
		Scopes:    scopes,
		ExpiresAt: expiresAt,
		RateLimit: rateLimit,
		IsActive:  true,
	}

	if err := m.db.Create(apiKey).Error; err != nil {
		return nil, "", err
	}

	return apiKey, key, nil
}

// Validate validates an API key
func (m *APIKeyManager) Validate(key string) (*APIKey, error) {
	hash := hashAPIKey(key)

	var apiKey APIKey
	if err := m.db.Where("hash = ? AND is_active = ?", hash, true).First(&apiKey).Error; err != nil {
		return nil, errors.New("invalid API key")
	}

	// Check expiration
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("API key expired")
	}

	// Update last used
	now := time.Now()
	apiKey.LastUsedAt = &now
	apiKey.UsageCount++
	m.db.Save(&apiKey)

	return &apiKey, nil
}

// HasScope checks if an API key has a specific scope
func (m *APIKeyManager) HasScope(apiKey *APIKey, scope string) bool {
	for _, s := range apiKey.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// Revoke revokes an API key
func (m *APIKeyManager) Revoke(keyID uint) error {
	return m.db.Model(&APIKey{}).Where("id = ?", keyID).Update("is_active", false).Error
}

// List lists API keys for a user
func (m *APIKeyManager) List(userID *uint) ([]APIKey, error) {
	var keys []APIKey
	query := m.db.Where("is_active = ?", true)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if err := query.Find(&keys).Error; err != nil {
		return nil, err
	}

	// Clear key values for security
	for i := range keys {
		keys[i].Key = ""
		keys[i].Hash = ""
	}

	return keys, nil
}

// GetByID gets an API key by ID
func (m *APIKeyManager) GetByID(keyID uint) (*APIKey, error) {
	var apiKey APIKey
	if err := m.db.First(&apiKey, keyID).Error; err != nil {
		return nil, err
	}

	// Clear sensitive data
	apiKey.Key = ""
	apiKey.Hash = ""

	return &apiKey, nil
}

// Update updates an API key
func (m *APIKeyManager) Update(keyID uint, name string, scopes []string, rateLimit int) error {
	updates := map[string]interface{}{
		"name":       name,
		"scopes":     scopes,
		"rate_limit": rateLimit,
	}
	return m.db.Model(&APIKey{}).Where("id = ?", keyID).Updates(updates).Error
}

// GetUsageStats gets usage statistics for an API key
func (m *APIKeyManager) GetUsageStats(keyID uint) (map[string]interface{}, error) {
	var apiKey APIKey
	if err := m.db.First(&apiKey, keyID).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"usage_count":  apiKey.UsageCount,
		"last_used_at": apiKey.LastUsedAt,
		"created_at":   apiKey.CreatedAt,
		"expires_at":   apiKey.ExpiresAt,
	}

	return stats, nil
}

// generateAPIKey generates a new API key
func generateAPIKey() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("dolphin_%s", base64.URLEncoding.EncodeToString(bytes))
	hash := hashAPIKey(key)

	return key, hash, nil
}

// hashAPIKey hashes an API key for storage
func hashAPIKey(key string) string {
	// Use a simple hash for now - in production, use bcrypt or similar
	hash := make([]byte, 32)
	copy(hash, []byte(key))
	return base64.StdEncoding.EncodeToString(hash)
}

