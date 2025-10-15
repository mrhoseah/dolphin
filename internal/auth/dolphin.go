package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// DTOs for authentication requests
type LoginRequestDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequestDto struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UserProvider defines the interface for user providers
type UserProvider interface {
	RetrieveByID(ctx context.Context, id uint) (Authenticatable, error)
	RetrieveByCredentials(ctx context.Context, credentials map[string]string) (Authenticatable, error)
	ValidateCredentials(user Authenticatable, credentials map[string]string) bool
}

// Authenticatable defines the interface for authenticatable users
type Authenticatable interface {
	GetID() uint
	GetAuthIdentifierName() string
	GetAuthIdentifier() string
	GetAuthPassword() string
	GetRememberToken() string
	SetRememberToken(value string)
	GetRememberTokenName() string
}

// Guard defines the interface for authentication guards
type Guard interface {
	Check() bool
	Guest() bool
	User() Authenticatable
	ID() uint
	Login(user Authenticatable) error
	LoginUsingID(id uint) error
	LoginWithCredentials(credentials map[string]string) error
	Logout()
	Once(user Authenticatable) error
	OnceUsingID(id uint) error
	Validate(credentials map[string]string) bool
	Attempt(credentials map[string]string) bool
	AttemptWithRemember(credentials map[string]string, remember bool) bool
}

// User represents an authenticatable user (Dolphin style)
type User struct {
	ID            uint       `json:"id" gorm:"primarykey"`
	Email         string     `json:"email" gorm:"uniqueIndex;not null"`
	Password      string     `json:"-" gorm:"not null"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	RememberToken string     `json:"-" gorm:"column:remember_token"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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

// DatabaseUserProvider implements UserProvider using database
type DatabaseUserProvider struct {
	db *gorm.DB
}

// NewDatabaseUserProvider creates a new database user provider
func NewDatabaseUserProvider(db *gorm.DB) *DatabaseUserProvider {
	return &DatabaseUserProvider{db: db}
}

// RetrieveByID retrieves a user by their ID
func (p *DatabaseUserProvider) RetrieveByID(ctx context.Context, id uint) (Authenticatable, error) {
	var user User
	if err := p.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// RetrieveByCredentials retrieves a user by their credentials
func (p *DatabaseUserProvider) RetrieveByCredentials(ctx context.Context, credentials map[string]string) (Authenticatable, error) {
	email, exists := credentials["email"]
	if !exists {
		return nil, errors.New("email credential is required")
	}

	var user User
	if err := p.db.WithContext(ctx).Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// ValidateCredentials validates the given credentials against the user
func (p *DatabaseUserProvider) ValidateCredentials(user Authenticatable, credentials map[string]string) bool {
	password, exists := credentials["password"]
	if !exists {
		return false
	}

	// Simple password validation (replace with proper bcrypt in production)
	// This is a placeholder - you should use bcrypt.CompareHashAndPassword
	return user.GetAuthPassword() == password
}

// SessionGuard implements Guard using sessions
type SessionGuard struct {
	name      string
	provider  UserProvider
	session   SessionStore
	user      Authenticatable
	loggedOut bool
}

// NewSessionGuard creates a new session guard
func NewSessionGuard(name string, provider UserProvider, session SessionStore) *SessionGuard {
	return &SessionGuard{
		name:      name,
		provider:  provider,
		session:   session,
		loggedOut: false,
	}
}

// Check determines if the current user is authenticated
func (g *SessionGuard) Check() bool {
	return !g.Guest()
}

// Guest determines if the current user is a guest
func (g *SessionGuard) Guest() bool {
	return g.User() == nil
}

// User returns the currently authenticated user
func (g *SessionGuard) User() Authenticatable {
	if g.loggedOut {
		return nil
	}

	if g.user != nil {
		return g.user
	}

	id := g.session.Get(g.getName())
	if id == nil {
		return nil
	}

	userID, ok := id.(uint)
	if !ok {
		return nil
	}

	user, err := g.provider.RetrieveByID(context.Background(), userID)
	if err != nil {
		return nil
	}

	g.user = user
	return user
}

// ID returns the ID of the currently authenticated user
func (g *SessionGuard) ID() uint {
	if user := g.User(); user != nil {
		return user.GetID()
	}
	return 0
}

// Login logs in a user
func (g *SessionGuard) Login(user Authenticatable) error {
	g.user = user
	g.session.Put(g.getName(), user.GetID())
	g.session.Regenerate()
	return nil
}

// LoginUsingID logs in a user by ID
func (g *SessionGuard) LoginUsingID(id uint) error {
	user, err := g.provider.RetrieveByID(context.Background(), id)
	if err != nil {
		return err
	}
	return g.Login(user)
}

// LoginWithCredentials logs in a user with credentials
func (g *SessionGuard) LoginWithCredentials(credentials map[string]string) error {
	user, err := g.provider.RetrieveByCredentials(context.Background(), credentials)
	if err != nil {
		return err
	}

	if !g.provider.ValidateCredentials(user, credentials) {
		return errors.New("invalid credentials")
	}

	return g.Login(user)
}

// Logout logs out the current user
func (g *SessionGuard) Logout() {
	g.user = nil
	g.loggedOut = true
	g.session.Forget(g.getName())
	g.session.Regenerate()
}

// Once logs in a user for a single request
func (g *SessionGuard) Once(user Authenticatable) error {
	g.user = user
	return nil
}

// OnceUsingID logs in a user by ID for a single request
func (g *SessionGuard) OnceUsingID(id uint) error {
	user, err := g.provider.RetrieveByID(context.Background(), id)
	if err != nil {
		return err
	}
	return g.Once(user)
}

// Validate validates the given credentials
func (g *SessionGuard) Validate(credentials map[string]string) bool {
	user, err := g.provider.RetrieveByCredentials(context.Background(), credentials)
	if err != nil {
		return false
	}
	return g.provider.ValidateCredentials(user, credentials)
}

// Attempt attempts to authenticate a user with the given credentials
func (g *SessionGuard) Attempt(credentials map[string]string) bool {
	return g.AttemptWithRemember(credentials, false)
}

// AttemptWithRemember attempts to authenticate a user with the given credentials and remember option
func (g *SessionGuard) AttemptWithRemember(credentials map[string]string, remember bool) bool {
	if g.Validate(credentials) {
		user, _ := g.provider.RetrieveByCredentials(context.Background(), credentials)
		g.Login(user)

		if remember {
			// Implement remember me functionality
			// This would typically set a remember token cookie
		}

		return true
	}
	return false
}

// getName returns the session key name for this guard
func (g *SessionGuard) getName() string {
	return "login_" + g.name + "_" + "user_id"
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	Get(key string) interface{}
	Put(key string, value interface{})
	Forget(key string)
	Regenerate()
}
