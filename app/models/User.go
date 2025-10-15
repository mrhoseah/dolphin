package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents an authenticatable user (Dolphin style)
type User struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	Email         string         `json:"email" gorm:"uniqueIndex;not null"`
	Password      string         `json:"-" gorm:"not null"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	RememberToken string         `json:"-" gorm:"column:remember_token"`
	LastLoginAt   *time.Time     `json:"last_login_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// GetID returns the user's ID
func (u *User) GetID() uint {
	return u.ID
}

// GetAuthIdentifierName returns the name of the unique identifier for the user
func (u *User) GetAuthIdentifierName() string {
	return "email"
}

// GetAuthIdentifier returns the unique identifier for the user
func (u *User) GetAuthIdentifier() string {
	return u.Email
}

// GetAuthPassword returns the password for the user
func (u *User) GetAuthPassword() string {
	return u.Password
}

// GetRememberToken returns the remember token for the user
func (u *User) GetRememberToken() string {
	return u.RememberToken
}

// SetRememberToken sets the remember token for the user
func (u *User) SetRememberToken(value string) {
	u.RememberToken = value
}

// GetRememberTokenName returns the name of the remember token field
func (u *User) GetRememberTokenName() string {
	return "remember_token"
}

// IsEmailVerified returns true if the user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// BeforeCreate is called before creating a new record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-create logic here
	// For example, hash the password
	return nil
}

// BeforeUpdate is called before updating a record
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// BeforeDelete is called before deleting a record
func (u *User) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}

// UserRepository handles user database operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(user *User) error {
	return r.db.Delete(user).Error
}

// List returns a list of users with pagination
func (r *UserRepository) List(limit, offset int) ([]*User, error) {
	var users []*User
	if err := r.db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
