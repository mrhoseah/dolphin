package auth

import (
	"sync"
)

// MemorySessionStore implements SessionStore using in-memory storage
type MemorySessionStore struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

// NewMemorySessionStore creates a new memory session store
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value from the session
func (s *MemorySessionStore) Get(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data[key]
}

// Put stores a value in the session
func (s *MemorySessionStore) Put(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
}

// Forget removes a value from the session
func (s *MemorySessionStore) Forget(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, key)
}

// Regenerate regenerates the session ID
func (s *MemorySessionStore) Regenerate() {
	// No-op for in-memory demo to avoid losing session data on login.
}
