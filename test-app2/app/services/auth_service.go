package services

import (
	"errors"
	"time"
	"crypto/rand"
	"encoding/hex"
	
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	
	"test-app2/models"
)

type AuthService struct {
	db *gorm.DB
}

type Session struct {
	Token     string
	UserID    uint
	ExpiresAt time.Time
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// Attempt attempts to authenticate a user
func (as *AuthService) Attempt(email, password string) (*models.User, error) {
	user := &models.User{}
	if err := as.db.Where("email = ?", email).First(user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.CheckPassword(password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// Register creates a new user
func (as *AuthService) Register(name, email, password string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := as.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists")
	}

	user := &models.User{
		Name:  name,
		Email: email,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	if err := as.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// CreateSession creates a new session for the user
func (as *AuthService) CreateSession(user *models.User) *Session {
	token := generateRandomToken()
	session := &Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7), // 7 days
	}

	// In a real implementation, you would store this in Redis or database
	// For now, we'll just return the session
	
	return session
}

// ValidateSession validates a session token
func (as *AuthService) ValidateSession(token string) (*models.User, error) {
	// In a real implementation, you would validate the token from Redis/database
	// For now, we'll just return a mock user
	
	// This is a placeholder - you should implement proper session validation
	return nil, errors.New("session validation not implemented")
}

// Helper function to generate random token
func generateRandomToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
