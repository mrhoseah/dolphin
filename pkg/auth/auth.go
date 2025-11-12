package auth

import (
	"github.com/mrhoseah/dolphin/internal/auth"
	"gorm.io/gorm"
)

// AuthManager represents the authentication manager
type AuthManager = auth.AuthManager

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return auth.NewAuthManager()
}

// SessionStore represents a session store interface
type SessionStore = auth.SessionStore

// MemorySessionStore represents an in-memory session store
type MemorySessionStore = auth.MemorySessionStore

// NewMemorySessionStore creates a new in-memory session store
func NewMemorySessionStore() *MemorySessionStore {
	return auth.NewMemorySessionStore()
}

// DatabaseProvider represents a database user provider
type DatabaseProvider = auth.DatabaseProvider

// NewDatabaseProvider creates a new database provider
func NewDatabaseProvider(db *gorm.DB, model interface{}) *DatabaseProvider {
	return auth.NewDatabaseProvider(db, model)
}

// SessionGuard represents a session-based authentication guard
type SessionGuard = auth.SessionGuard

// NewSessionGuard creates a new session guard
func NewSessionGuard(name string, provider Provider, session SessionStore) *SessionGuard {
	return auth.NewSessionGuard(name, provider, session)
}

// Provider represents a user provider interface
type Provider = auth.Provider
