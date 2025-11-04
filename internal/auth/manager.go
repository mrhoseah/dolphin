package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents an authenticated user
type User struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	Password        string     `gorm:"not null" json:"-"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	RememberToken   string     `json:"remember_token"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Guard represents an authentication guard
type Guard interface {
	Attempt(credentials map[string]string) (bool, error)
	Login(user interface{}) error
	Logout() error
	Check() bool
	User() interface{}
	ID() interface{}
	Guest() bool
}

// Provider represents a user provider
type Provider interface {
	RetrieveByID(id interface{}) (interface{}, error)
	RetrieveByCredentials(credentials map[string]string) (interface{}, error)
	ValidateCredentials(user interface{}, credentials map[string]string) bool
}

// DatabaseProvider implements Provider using database
type DatabaseProvider struct {
	db    *gorm.DB
	model interface{}
}

// NewDatabaseProvider creates a new database provider
func NewDatabaseProvider(db *gorm.DB, model interface{}) *DatabaseProvider {
	return &DatabaseProvider{
		db:    db,
		model: model,
	}
}

// RetrieveByID retrieves a user by ID
func (dp *DatabaseProvider) RetrieveByID(id interface{}) (interface{}, error) {
	// Create a new instance of the model type using reflection
	modelType := reflect.TypeOf(dp.model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	// Create a new instance of the model
	userValue := reflect.New(modelType).Interface()
	err := dp.db.First(userValue, id).Error
	return userValue, err
}

// RetrieveByCredentials retrieves a user by credentials
func (dp *DatabaseProvider) RetrieveByCredentials(credentials map[string]string) (interface{}, error) {
	// Create a new instance of the model type using reflection
	modelType := reflect.TypeOf(dp.model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	// Create a new instance of the model
	userValue := reflect.New(modelType).Interface()
	query := dp.db

	for key, value := range credentials {
		if key != "password" {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	err := query.First(userValue).Error
	return userValue, err
}

// ValidateCredentials validates user credentials
func (dp *DatabaseProvider) ValidateCredentials(user interface{}, credentials map[string]string) bool {
	if password, exists := credentials["password"]; exists {
		// Get password field from user struct
		userValue := reflect.ValueOf(user)
		if userValue.Kind() == reflect.Ptr {
			userValue = userValue.Elem()
		}

		passwordField := userValue.FieldByName("Password")
		if passwordField.IsValid() {
			storedPassword := passwordField.String()
			return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) == nil
		}
	}
	return false
}

// SessionGuard implements Guard using sessions
type SessionGuard struct {
	name     string
	provider Provider
	session  SessionStore
	user     interface{}
}

// NewSessionGuard creates a new session guard
func NewSessionGuard(name string, provider Provider, session SessionStore) *SessionGuard {
	return &SessionGuard{
		name:     name,
		provider: provider,
		session:  session,
	}
}

// Attempt attempts to authenticate a user
func (sg *SessionGuard) Attempt(credentials map[string]string) (bool, error) {
	user, err := sg.provider.RetrieveByCredentials(credentials)
	if err != nil {
		return false, err
	}

	if sg.provider.ValidateCredentials(user, credentials) {
		sg.Login(user)
		return true, nil
	}

	return false, nil
}

// Login logs in a user
func (sg *SessionGuard) Login(user interface{}) error {
	sg.user = user

	// Get user ID
	userValue := reflect.ValueOf(user)
	if userValue.Kind() == reflect.Ptr {
		userValue = userValue.Elem()
	}

	idField := userValue.FieldByName("ID")
	if !idField.IsValid() {
		return errors.New("user must have an ID field")
	}

	userID := idField.Interface()

	// Store in session
	sg.session.Put(fmt.Sprintf("auth.%s", sg.name), userID)
	sg.session.Put(fmt.Sprintf("auth.%s.login", sg.name), time.Now().Unix())

	return nil
}

// Logout logs out the current user
func (sg *SessionGuard) Logout() error {
	sg.user = nil
	sg.session.Forget(fmt.Sprintf("auth.%s", sg.name))
	sg.session.Forget(fmt.Sprintf("auth.%s.login", sg.name))
	return nil
}

// Check checks if user is authenticated
func (sg *SessionGuard) Check() bool {
	if sg.user != nil {
		return true
	}

	userID := sg.session.Get(fmt.Sprintf("auth.%s", sg.name))
	if userID == nil {
		return false
	}

	user, err := sg.provider.RetrieveByID(userID)
	if err != nil {
		return false
	}

	sg.user = user
	return true
}

// User returns the authenticated user
func (sg *SessionGuard) User() interface{} {
	if !sg.Check() {
		return nil
	}
	return sg.user
}

// ID returns the authenticated user's ID
func (sg *SessionGuard) ID() interface{} {
	if !sg.Check() {
		return nil
	}

	userValue := reflect.ValueOf(sg.user)
	if userValue.Kind() == reflect.Ptr {
		userValue = userValue.Elem()
	}

	idField := userValue.FieldByName("ID")
	if idField.IsValid() {
		return idField.Interface()
	}

	return nil
}

// Guest checks if user is a guest
func (sg *SessionGuard) Guest() bool {
	return !sg.Check()
}

// JWTPayload represents JWT payload
type JWTPayload struct {
	UserID    interface{} `json:"user_id"`
	Email     string      `json:"email"`
	ExpiresAt int64       `json:"exp"`
	IssuedAt  int64       `json:"iat"`
}

// Valid implements jwt.Claims interface
func (p JWTPayload) Valid() error {
	if time.Now().Unix() > p.ExpiresAt {
		return fmt.Errorf("token expired")
	}
	return nil
}

// GetAudience implements jwt.Claims interface
func (p JWTPayload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}

// GetExpirationTime implements jwt.Claims interface
func (p JWTPayload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(p.ExpiresAt, 0)), nil
}

// GetIssuedAt implements jwt.Claims interface
func (p JWTPayload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(p.IssuedAt, 0)), nil
}

// GetIssuer implements jwt.Claims interface
func (p JWTPayload) GetIssuer() (string, error) {
	return "", nil
}

// GetNotBefore implements jwt.Claims interface
func (p JWTPayload) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(p.IssuedAt, 0)), nil
}

// GetSubject implements jwt.Claims interface
func (p JWTPayload) GetSubject() (string, error) {
	return p.Email, nil
}

// JWTGuard implements Guard using JWT tokens
type JWTGuard struct {
	name     string
	provider Provider
	secret   string
	user     interface{}
}

// NewJWTGuard creates a new JWT guard
func NewJWTGuard(name string, provider Provider, secret string) *JWTGuard {
	return &JWTGuard{
		name:     name,
		provider: provider,
		secret:   secret,
	}
}

// Attempt attempts to authenticate a user
func (jg *JWTGuard) Attempt(credentials map[string]string) (bool, error) {
	user, err := jg.provider.RetrieveByCredentials(credentials)
	if err != nil {
		return false, err
	}

	if jg.provider.ValidateCredentials(user, credentials) {
		jg.Login(user)
		return true, nil
	}

	return false, nil
}

// Login logs in a user and returns a token
func (jg *JWTGuard) Login(user interface{}) error {
	jg.user = user

	// Get user data
	userValue := reflect.ValueOf(user)
	if userValue.Kind() == reflect.Ptr {
		userValue = userValue.Elem()
	}

	idField := userValue.FieldByName("ID")
	emailField := userValue.FieldByName("Email")

	if !idField.IsValid() {
		return errors.New("user must have an ID field")
	}

	userID := idField.Interface()
	email := ""
	if emailField.IsValid() {
		email = emailField.String()
	}

	// Create JWT token
	now := time.Now()
	payload := JWTPayload{
		UserID:    userID,
		Email:     email,
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
		IssuedAt:  now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(jg.secret))
	if err != nil {
		return err
	}

	// Store token in user struct if possible
	tokenField := userValue.FieldByName("RememberToken")
	if tokenField.IsValid() && tokenField.CanSet() {
		tokenField.SetString(tokenString)
	}

	return nil
}

// Logout logs out the current user
func (jg *JWTGuard) Logout() error {
	jg.user = nil
	return nil
}

// Check checks if user is authenticated
func (jg *JWTGuard) Check() bool {
	return jg.user != nil
}

// User returns the authenticated user
func (jg *JWTGuard) User() interface{} {
	return jg.user
}

// ID returns the authenticated user's ID
func (jg *JWTGuard) ID() interface{} {
	if jg.user == nil {
		return nil
	}

	userValue := reflect.ValueOf(jg.user)
	if userValue.Kind() == reflect.Ptr {
		userValue = userValue.Elem()
	}

	idField := userValue.FieldByName("ID")
	if idField.IsValid() {
		return idField.Interface()
	}

	return nil
}

// Guest checks if user is a guest
func (jg *JWTGuard) Guest() bool {
	return jg.user == nil
}

// ValidateToken validates a JWT token
func (jg *JWTGuard) ValidateToken(tokenString string) (*JWTPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTPayload{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jg.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTPayload); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AuthenticateFromToken authenticates a user from a JWT token
func (jg *JWTGuard) AuthenticateFromToken(tokenString string) error {
	payload, err := jg.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	user, err := jg.provider.RetrieveByID(payload.UserID)
	if err != nil {
		return err
	}

	jg.user = user
	return nil
}

// AuthManager manages authentication
type AuthManager struct {
	guards       map[string]Guard
	providers    map[string]Provider
	defaultGuard string
}

// NewAuthManager creates a new auth manager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		guards:       make(map[string]Guard),
		providers:    make(map[string]Provider),
		defaultGuard: "web",
	}
}

// Guard returns a guard by name
func (am *AuthManager) Guard(name string) Guard {
	if guard, exists := am.guards[name]; exists {
		return guard
	}
	return nil
}

// DefaultGuard returns the default guard
func (am *AuthManager) DefaultGuard() Guard {
	guard := am.Guard(am.defaultGuard)
	if guard == nil {
		// Return a no-op guard if no guard is registered to prevent nil pointer panics
		// This allows the app to start even if guards aren't configured yet
		return &noOpGuard{}
	}
	return guard
}

// noOpGuard is a guard that always returns false/empty to prevent nil pointer panics
type noOpGuard struct{}

func (g *noOpGuard) Attempt(credentials map[string]string) (bool, error) {
	return false, nil
}

func (g *noOpGuard) Login(user interface{}) error {
	return nil
}

func (g *noOpGuard) Logout() error {
	return nil
}

func (g *noOpGuard) Check() bool {
	return false
}

func (g *noOpGuard) User() interface{} {
	return nil
}

func (g *noOpGuard) ID() interface{} {
	return nil
}

func (g *noOpGuard) Guest() bool {
	return true
}

// RegisterGuard registers a guard
func (am *AuthManager) RegisterGuard(name string, guard Guard) {
	am.guards[name] = guard
}

// RegisterProvider registers a provider
func (am *AuthManager) RegisterProvider(name string, provider Provider) {
	am.providers[name] = provider
}

// SetDefaultGuard sets the default guard
func (am *AuthManager) SetDefaultGuard(name string) {
	am.defaultGuard = name
}

// Attempt attempts to authenticate using the default guard
func (am *AuthManager) Attempt(credentials map[string]string) (bool, error) {
	return am.DefaultGuard().Attempt(credentials)
}

// Login logs in a user using the default guard
func (am *AuthManager) Login(user interface{}) error {
	return am.DefaultGuard().Login(user)
}

// Logout logs out the current user using the default guard
func (am *AuthManager) Logout() error {
	return am.DefaultGuard().Logout()
}

// Check checks if user is authenticated using the default guard
func (am *AuthManager) Check() bool {
	return am.DefaultGuard().Check()
}

// User returns the authenticated user using the default guard
func (am *AuthManager) User() interface{} {
	return am.DefaultGuard().User()
}

// ID returns the authenticated user's ID using the default guard
func (am *AuthManager) ID() interface{} {
	return am.DefaultGuard().ID()
}

// Guest checks if user is a guest using the default guard
func (am *AuthManager) Guest() bool {
	return am.DefaultGuard().Guest()
}

// SessionStore represents a session store interface
type SessionStore interface {
	Get(key string) interface{}
	Put(key string, value interface{})
	Forget(key string)
	Flush()
}

// PasswordHasher handles password hashing
type PasswordHasher struct{}

// Hash hashes a password
func (ph *PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// Check checks a password against a hash
func (ph *PasswordHasher) Check(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateRememberToken generates a remember token
func (ph *PasswordHasher) GenerateRememberToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// Middleware represents authentication middleware
type Middleware struct {
	authManager *AuthManager
}

// NewMiddleware creates new authentication middleware
func NewMiddleware(authManager *AuthManager) *Middleware {
	return &Middleware{
		authManager: authManager,
	}
}

// Authenticate middleware checks if user is authenticated
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.authManager.Check() {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Guest middleware checks if user is a guest
func (m *Middleware) Guest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.authManager.Check() {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Guard middleware checks authentication for a specific guard
func (m *Middleware) Guard(guardName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			guard := m.authManager.Guard(guardName)
			if guard == nil || !guard.Check() {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ExtractTokenFromRequest extracts JWT token from request
func ExtractTokenFromRequest(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Check query parameter
	return r.URL.Query().Get("token")
}

// HashPassword hashes a password
func HashPassword(password string) (string, error) {
	hasher := &PasswordHasher{}
	return hasher.Hash(password)
}

// CheckPassword checks a password against a hash
func CheckPassword(password, hash string) bool {
	hasher := &PasswordHasher{}
	return hasher.Check(password, hash)
}

// GenerateRememberToken generates a remember token
func GenerateRememberToken() string {
	hasher := &PasswordHasher{}
	return hasher.GenerateRememberToken()
}
