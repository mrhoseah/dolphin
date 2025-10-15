package maintenance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager handles maintenance mode operations
type Manager struct {
	filePath     string
	mu           sync.RWMutex
	enabled      bool
	message      string
	retryAfter   int
	allowedIPs   []string
	bypassSecret string
}

// MaintenanceInfo holds maintenance mode configuration
type MaintenanceInfo struct {
	Enabled      bool      `json:"enabled"`
	Message      string    `json:"message"`
	RetryAfter   int       `json:"retry_after"`
	AllowedIPs   []string  `json:"allowed_ips"`
	BypassSecret string    `json:"bypass_secret"`
	StartedAt    time.Time `json:"started_at"`
	EndsAt       time.Time `json:"ends_at,omitempty"`
}

// NewManager creates a new maintenance manager
func NewManager(filePath string) *Manager {
	if filePath == "" {
		filePath = "storage/framework/maintenance.json"
	}

	return &Manager{
		filePath:     filePath,
		enabled:      false,
		message:      "Application is currently under maintenance. Please try again later.",
		retryAfter:   3600, // 1 hour
		allowedIPs:   []string{},
		bypassSecret: "",
	}
}

// Enable puts the application in maintenance mode
func (m *Manager) Enable(message string, retryAfter int, allowedIPs []string, bypassSecret string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := MaintenanceInfo{
		Enabled:      true,
		Message:      message,
		RetryAfter:   retryAfter,
		AllowedIPs:   allowedIPs,
		BypassSecret: bypassSecret,
		StartedAt:    time.Now(),
	}

	if retryAfter > 0 {
		info.EndsAt = time.Now().Add(time.Duration(retryAfter) * time.Second)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create maintenance directory: %w", err)
	}

	// Write maintenance file
	if err := m.writeMaintenanceFile(info); err != nil {
		return fmt.Errorf("failed to write maintenance file: %w", err)
	}

	m.enabled = true
	m.message = message
	m.retryAfter = retryAfter
	m.allowedIPs = allowedIPs
	m.bypassSecret = bypassSecret

	return nil
}

// Disable removes maintenance mode
func (m *Manager) Disable() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove maintenance file
	if err := os.Remove(m.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove maintenance file: %w", err)
	}

	m.enabled = false
	m.message = ""
	m.retryAfter = 0
	m.allowedIPs = []string{}
	m.bypassSecret = ""

	return nil
}

// IsEnabled checks if maintenance mode is active
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check file existence first
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return false
	}

	// Load from file to get latest state
	info, err := m.loadMaintenanceFile()
	if err != nil {
		return false
	}

	// Check if maintenance has expired
	if !info.EndsAt.IsZero() && time.Now().After(info.EndsAt) {
		// Auto-disable expired maintenance
		m.Disable()
		return false
	}

	return info.Enabled
}

// GetInfo returns current maintenance information
func (m *Manager) GetInfo() (*MaintenanceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.loadMaintenanceFile()
}

// IsIPAllowed checks if an IP is allowed during maintenance
func (m *Manager) IsIPAllowed(ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, err := m.loadMaintenanceFile()
	if err != nil {
		return false
	}

	for _, allowedIP := range info.AllowedIPs {
		if allowedIP == ip {
			return true
		}
	}

	return false
}

// IsBypassSecretValid checks if the bypass secret is valid
func (m *Manager) IsBypassSecretValid(secret string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, err := m.loadMaintenanceFile()
	if err != nil {
		return false
	}

	return info.BypassSecret != "" && info.BypassSecret == secret
}

// GetRetryAfter returns the retry-after header value
func (m *Manager) GetRetryAfter() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, err := m.loadMaintenanceFile()
	if err != nil {
		return 0
	}

	return info.RetryAfter
}

// GetMessage returns the maintenance message
func (m *Manager) GetMessage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, err := m.loadMaintenanceFile()
	if err != nil {
		return "Application is currently under maintenance."
	}

	return info.Message
}

// loadMaintenanceFile loads maintenance info from file
func (m *Manager) loadMaintenanceFile() (*MaintenanceInfo, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, err
	}

	var info MaintenanceInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// writeMaintenanceFile writes maintenance info to file
func (m *Manager) writeMaintenanceFile(info MaintenanceInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// Status returns maintenance status information
func (m *Manager) Status() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, err := m.loadMaintenanceFile()
	if err != nil {
		return map[string]interface{}{
			"enabled": false,
			"error":   err.Error(),
		}
	}

	status := map[string]interface{}{
		"enabled":       info.Enabled,
		"message":       info.Message,
		"retry_after":   info.RetryAfter,
		"allowed_ips":   info.AllowedIPs,
		"started_at":    info.StartedAt,
		"bypass_secret": info.BypassSecret != "",
	}

	if !info.EndsAt.IsZero() {
		status["ends_at"] = info.EndsAt
		status["expires_in"] = int(time.Until(info.EndsAt).Seconds())
	}

	return status
}

// Cleanup removes expired maintenance files
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, err := m.loadMaintenanceFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No maintenance file
		}
		return err
	}

	// Check if maintenance has expired
	if !info.EndsAt.IsZero() && time.Now().After(info.EndsAt) {
		return m.Disable()
	}

	return nil
}
