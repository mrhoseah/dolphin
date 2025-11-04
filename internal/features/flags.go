package features

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Flag represents a feature flag
type Flag struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Manager manages feature flags
type Manager struct {
	flags     map[string]*Flag
	mu        sync.RWMutex
	filePath  string
	autoSave  bool
	listeners []func(string, bool)
}

// NewManager creates a new feature flag manager
func NewManager(filePath string) *Manager {
	m := &Manager{
		flags:     make(map[string]*Flag),
		filePath:  filePath,
		autoSave:  true,
		listeners: make([]func(string, bool), 0),
	}
	m.loadFromFile()
	return m
}

// Register registers a new feature flag
func (m *Manager) Register(name, description string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.flags[name]; exists {
		return fmt.Errorf("feature flag '%s' already exists", name)
	}

	now := time.Now()
	m.flags[name] = &Flag{
		Name:        name,
		Enabled:     enabled,
		Description: description,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if m.autoSave {
		return m.saveToFile()
	}
	return nil
}

// Enable enables a feature flag
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, exists := m.flags[name]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", name)
	}

	oldValue := flag.Enabled
	flag.Enabled = true
	flag.UpdatedAt = time.Now()

	if m.autoSave {
		if err := m.saveToFile(); err != nil {
			flag.Enabled = oldValue
			return err
		}
	}

	m.notifyListeners(name, true)
	return nil
}

// Disable disables a feature flag
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, exists := m.flags[name]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", name)
	}

	oldValue := flag.Enabled
	flag.Enabled = false
	flag.UpdatedAt = time.Now()

	if m.autoSave {
		if err := m.saveToFile(); err != nil {
			flag.Enabled = oldValue
			return err
		}
	}

	m.notifyListeners(name, false)
	return nil
}

// IsEnabled checks if a feature flag is enabled
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, exists := m.flags[name]
	if !exists {
		return false
	}
	return flag.Enabled
}

// Toggle toggles a feature flag
func (m *Manager) Toggle(name string) error {
	if m.IsEnabled(name) {
		return m.Disable(name)
	}
	return m.Enable(name)
}

// Get retrieves a feature flag
func (m *Manager) Get(name string) (*Flag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, exists := m.flags[name]
	if !exists {
		return nil, fmt.Errorf("feature flag '%s' not found", name)
	}

	// Return a copy to prevent external modification
	flagCopy := *flag
	return &flagCopy, nil
}

// GetAll retrieves all feature flags
func (m *Manager) GetAll() map[string]*Flag {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Flag)
	for name, flag := range m.flags {
		flagCopy := *flag
		result[name] = &flagCopy
	}
	return result
}

// SetMetadata sets metadata for a feature flag
func (m *Manager) SetMetadata(name string, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, exists := m.flags[name]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", name)
	}

	if flag.Metadata == nil {
		flag.Metadata = make(map[string]interface{})
	}
	flag.Metadata[key] = value
	flag.UpdatedAt = time.Now()

	if m.autoSave {
		return m.saveToFile()
	}
	return nil
}

// Delete deletes a feature flag
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.flags[name]; !exists {
		return fmt.Errorf("feature flag '%s' not found", name)
	}

	delete(m.flags, name)

	if m.autoSave {
		return m.saveToFile()
	}
	return nil
}

// OnChange registers a listener for flag changes
func (m *Manager) OnChange(listener func(flagName string, enabled bool)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// notifyListeners notifies all listeners of a flag change
func (m *Manager) notifyListeners(name string, enabled bool) {
	m.mu.RLock()
	listeners := make([]func(string, bool), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, listener := range listeners {
		go listener(name, enabled)
	}
}

// loadFromFile loads flags from file
func (m *Manager) loadFromFile() error {
	if m.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var flags map[string]*Flag
	if err := json.Unmarshal(data, &flags); err != nil {
		return err
	}

	m.mu.Lock()
	m.flags = flags
	m.mu.Unlock()

	return nil
}

// saveToFile saves flags to file
func (m *Manager) saveToFile() error {
	if m.filePath == "" {
		return nil
	}

	m.mu.RLock()
	data, err := json.MarshalIndent(m.flags, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// Save manually saves flags to file
func (m *Manager) Save() error {
	return m.saveToFile()
}

// Load manually loads flags from file
func (m *Manager) Load() error {
	return m.loadFromFile()
}

