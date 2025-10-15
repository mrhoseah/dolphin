package auth

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// AuthManager manages authentication guards and providers
type AuthManager struct {
	guards       map[string]Guard
	providers    map[string]UserProvider
	defaultGuard string
	mutex        sync.RWMutex
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		guards:       make(map[string]Guard),
		providers:    make(map[string]UserProvider),
		defaultGuard: "web",
	}
}

// Guard returns the specified guard
func (m *AuthManager) Guard(name string) Guard {
	if name == "" {
		name = m.defaultGuard
	}

	m.mutex.RLock()
	guard, exists := m.guards[name]
	m.mutex.RUnlock()

	if !exists {
		panic(fmt.Sprintf("Auth guard [%s] is not defined", name))
	}

	return guard
}

// DefaultGuard returns the default guard
func (m *AuthManager) DefaultGuard() Guard {
	return m.Guard(m.defaultGuard)
}

// Check determines if the current user is authenticated
func (m *AuthManager) Check() bool {
	return m.DefaultGuard().Check()
}

// Guest determines if the current user is a guest
func (m *AuthManager) Guest() bool {
	return m.DefaultGuard().Guest()
}

// User returns the currently authenticated user
func (m *AuthManager) User() Authenticatable {
	return m.DefaultGuard().User()
}

// ID returns the ID of the currently authenticated user
func (m *AuthManager) ID() uint {
	return m.DefaultGuard().ID()
}

// Login logs in a user
func (m *AuthManager) Login(user Authenticatable) error {
	return m.DefaultGuard().Login(user)
}

// LoginUsingID logs in a user by ID
func (m *AuthManager) LoginUsingID(id uint) error {
	return m.DefaultGuard().LoginUsingID(id)
}

// LoginWithCredentials logs in a user with credentials
func (m *AuthManager) LoginWithCredentials(credentials map[string]string) error {
	return m.DefaultGuard().LoginWithCredentials(credentials)
}

// Logout logs out the current user
func (m *AuthManager) Logout() {
	m.DefaultGuard().Logout()
}

// Once logs in a user for a single request
func (m *AuthManager) Once(user Authenticatable) error {
	return m.DefaultGuard().Once(user)
}

// OnceUsingID logs in a user by ID for a single request
func (m *AuthManager) OnceUsingID(id uint) error {
	return m.DefaultGuard().OnceUsingID(id)
}

// Validate validates the given credentials
func (m *AuthManager) Validate(credentials map[string]string) bool {
	return m.DefaultGuard().Validate(credentials)
}

// Attempt attempts to authenticate a user with the given credentials
func (m *AuthManager) Attempt(credentials map[string]string) bool {
	return m.DefaultGuard().Attempt(credentials)
}

// AttemptWithRemember attempts to authenticate a user with the given credentials and remember option
func (m *AuthManager) AttemptWithRemember(credentials map[string]string, remember bool) bool {
	return m.DefaultGuard().AttemptWithRemember(credentials, remember)
}

// RegisterGuard registers a guard
func (m *AuthManager) RegisterGuard(name string, guard Guard) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.guards[name] = guard
}

// RegisterProvider registers a user provider
func (m *AuthManager) RegisterProvider(name string, provider UserProvider) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.providers[name] = provider
}

// SetDefaultGuard sets the default guard
func (m *AuthManager) SetDefaultGuard(name string) {
	m.defaultGuard = name
}

// AuthFacade provides a static interface to the authentication manager
type AuthFacade struct {
	manager *AuthManager
}

// NewAuthFacade creates a new auth facade
func NewAuthFacade(manager *AuthManager) *AuthFacade {
	return &AuthFacade{manager: manager}
}

// Check determines if the current user is authenticated
func (f *AuthFacade) Check() bool {
	return f.manager.Check()
}

// Guest determines if the current user is a guest
func (f *AuthFacade) Guest() bool {
	return f.manager.Guest()
}

// User returns the currently authenticated user
func (f *AuthFacade) User() Authenticatable {
	return f.manager.User()
}

// ID returns the ID of the currently authenticated user
func (f *AuthFacade) ID() uint {
	return f.manager.ID()
}

// Login logs in a user
func (f *AuthFacade) Login(user Authenticatable) error {
	return f.manager.Login(user)
}

// LoginUsingID logs in a user by ID
func (f *AuthFacade) LoginUsingID(id uint) error {
	return f.manager.LoginUsingID(id)
}

// LoginWithCredentials logs in a user with credentials
func (f *AuthFacade) LoginWithCredentials(credentials map[string]string) error {
	return f.manager.LoginWithCredentials(credentials)
}

// Logout logs out the current user
func (f *AuthFacade) Logout() {
	f.manager.Logout()
}

// Once logs in a user for a single request
func (f *AuthFacade) Once(user Authenticatable) error {
	return f.manager.Once(user)
}

// OnceUsingID logs in a user by ID for a single request
func (f *AuthFacade) OnceUsingID(id uint) error {
	return f.manager.OnceUsingID(id)
}

// Validate validates the given credentials
func (f *AuthFacade) Validate(credentials map[string]string) bool {
	return f.manager.Validate(credentials)
}

// Attempt attempts to authenticate a user with the given credentials
func (f *AuthFacade) Attempt(credentials map[string]string) bool {
	return f.manager.Attempt(credentials)
}

// AttemptWithRemember attempts to authenticate a user with the given credentials and remember option
func (f *AuthFacade) AttemptWithRemember(credentials map[string]string, remember bool) bool {
	return f.manager.AttemptWithRemember(credentials, remember)
}

// Guard returns the specified guard
func (f *AuthFacade) Guard(name string) Guard {
	return f.manager.Guard(name)
}

// DefaultGuard returns the default guard
func (f *AuthFacade) DefaultGuard() Guard {
	return f.manager.DefaultGuard()
}

// SetupAuth configures the authentication system
func SetupAuth(db *gorm.DB, sessionStore SessionStore) *AuthManager {
	manager := NewAuthManager()

	// Register user provider
	userProvider := NewDatabaseUserProvider(db)
	manager.RegisterProvider("users", userProvider)

	// Register web guard
	webGuard := NewSessionGuard("web", userProvider, sessionStore)
	manager.RegisterGuard("web", webGuard)

	// Register api guard (for API authentication)
	apiGuard := NewSessionGuard("api", userProvider, sessionStore)
	manager.RegisterGuard("api", apiGuard)

	// Set default guard
	manager.SetDefaultGuard("web")

	return manager
}
