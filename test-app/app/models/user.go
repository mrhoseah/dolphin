package models

import (
	"time"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Name              string         `json:"name" gorm:"not null"`
	Email             string         `json:"email" gorm:"uniqueIndex;not null"`
	EmailVerifiedAt   *time.Time     `json:"email_verified_at"`
	Password          string         `json:"-" gorm:"not null"`
	RememberToken     string         `json:"-"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// SetPassword hashes the password before storing
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsEmailVerified checks if user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// MarkEmailAsVerified marks the user's email as verified
func (u *User) MarkEmailAsVerified() {
	now := time.Now()
	u.EmailVerifiedAt = &now
}
