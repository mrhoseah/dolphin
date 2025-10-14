package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// SessionManager handles session operations
type SessionManager struct {
	store *sessions.CookieStore
}

// NewSessionManager creates a new session manager
func NewSessionManager(secretKey string) *SessionManager {
	store := sessions.NewCookieStore([]byte(secretKey))

	// Configure session options
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	return &SessionManager{
		store: store,
	}
}

// GetSession retrieves a session
func (sm *SessionManager) GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return sm.store.Get(r, name)
}

// SaveSession saves a session
func (sm *SessionManager) SaveSession(session *sessions.Session, w http.ResponseWriter, r *http.Request) error {
	return session.Save(r, w)
}

// Set stores a value in session
func (sm *SessionManager) Set(session *sessions.Session, key string, value interface{}) {
	session.Values[key] = value
}

// Get retrieves a value from session
func (sm *SessionManager) Get(session *sessions.Session, key string) (interface{}, bool) {
	value, exists := session.Values[key]
	return value, exists
}

// GetString retrieves a string value from session
func (sm *SessionManager) GetString(session *sessions.Session, key string) (string, bool) {
	value, exists := session.Values[key]
	if !exists {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

// GetInt retrieves an integer value from session
func (sm *SessionManager) GetInt(session *sessions.Session, key string) (int, bool) {
	value, exists := session.Values[key]
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a boolean value from session
func (sm *SessionManager) GetBool(session *sessions.Session, key string) (bool, bool) {
	value, exists := session.Values[key]
	if !exists {
		return false, false
	}

	boolVal, ok := value.(bool)
	return boolVal, ok
}

// Delete removes a value from session
func (sm *SessionManager) Delete(session *sessions.Session, key string) {
	delete(session.Values, key)
}

// Clear removes all values from session
func (sm *SessionManager) Clear(session *sessions.Session) {
	for key := range session.Values {
		delete(session.Values, key)
	}
}

// Flash stores a flash message in session
func (sm *SessionManager) Flash(session *sessions.Session, key string, value interface{}) {
	sm.Set(session, "flash_"+key, value)
}

// GetFlash retrieves and removes a flash message from session
func (sm *SessionManager) GetFlash(session *sessions.Session, key string) (interface{}, bool) {
	flashKey := "flash_" + key
	value, exists := session.Values[flashKey]
	if exists {
		delete(session.Values, flashKey)
	}
	return value, exists
}

// GenerateSessionID generates a unique session ID
func (sm *SessionManager) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SessionData represents session data structure
type SessionData struct {
	UserID    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Role      string                 `json:"role"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SetUserData stores user data in session
func (sm *SessionManager) SetUserData(session *sessions.Session, userData *SessionData) {
	sm.Set(session, "user_data", userData)
}

// GetUserData retrieves user data from session
func (sm *SessionManager) GetUserData(session *sessions.Session) (*SessionData, bool) {
	value, exists := sm.Get(session, "user_data")
	if !exists {
		return nil, false
	}

	userData, ok := value.(*SessionData)
	return userData, ok
}

// IsAuthenticated checks if user is authenticated
func (sm *SessionManager) IsAuthenticated(session *sessions.Session) bool {
	userData, exists := sm.GetUserData(session)
	return exists && userData.UserID != ""
}

// Logout clears user data from session
func (sm *SessionManager) Logout(session *sessions.Session) {
	sm.Delete(session, "user_data")
}

// SessionMiddleware provides session handling middleware
func SessionMiddleware(sessionManager *SessionManager, sessionName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessionManager.GetSession(r, sessionName)
			if err != nil {
				// Log error but continue
				fmt.Printf("Session error: %v\n", err)
			}

			// Add session to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "session", session)
			ctx = context.WithValue(ctx, "session_manager", sessionManager)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSessionFromContext retrieves session from request context
func GetSessionFromContext(ctx context.Context) (*sessions.Session, bool) {
	session, ok := ctx.Value("session").(*sessions.Session)
	return session, ok
}

// GetSessionManagerFromContext retrieves session manager from request context
func GetSessionManagerFromContext(ctx context.Context) (*SessionManager, bool) {
	manager, ok := ctx.Value("session_manager").(*SessionManager)
	return manager, ok
}

// SessionStore interface for different session storage backends
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	Save(s *sessions.Session, w http.ResponseWriter, r *http.Request) error
}

// DatabaseSessionStore implements session storage in database
type DatabaseSessionStore struct {
	// This would be implemented with actual database operations
	// For now, it's a placeholder
}

func (d *DatabaseSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	// Implement database session retrieval
	return nil, fmt.Errorf("not implemented")
}

func (d *DatabaseSessionStore) Save(s *sessions.Session, w http.ResponseWriter, r *http.Request) error {
	// Implement database session saving
	return fmt.Errorf("not implemented")
}

// RedisSessionStore implements session storage in Redis
type RedisSessionStore struct {
	// This would be implemented with Redis operations
	// For now, it's a placeholder
}

func (r *RedisSessionStore) Get(req *http.Request, name string) (*sessions.Session, error) {
	// Implement Redis session retrieval
	return nil, fmt.Errorf("not implemented")
}

func (r *RedisSessionStore) Save(s *sessions.Session, w http.ResponseWriter, req *http.Request) error {
	// Implement Redis session saving
	return fmt.Errorf("not implemented")
}
