package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager manages multiple circuit breakers
type Manager struct {
	circuits map[string]*CircuitBreaker
	metrics  *MetricsCollector
	logger   *zap.Logger
	mu       sync.RWMutex

	// Configuration
	defaultConfig *Config

	// Monitoring
	enableMonitoring bool
	monitorInterval  time.Duration
	stopMonitor      chan struct{}
}

// ManagerConfig represents manager configuration
type ManagerConfig struct {
	DefaultConfig    *Config       `yaml:"default_config" json:"default_config"`
	EnableMonitoring bool          `yaml:"enable_monitoring" json:"enable_monitoring"`
	MonitorInterval  time.Duration `yaml:"monitor_interval" json:"monitor_interval"`
}

// DefaultManagerConfig returns default manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultConfig:    DefaultConfig(),
		EnableMonitoring: true,
		MonitorInterval:  30 * time.Second,
	}
}

// NewManager creates a new circuit breaker manager
func NewManager(config *ManagerConfig, logger *zap.Logger) *Manager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	manager := &Manager{
		circuits:         make(map[string]*CircuitBreaker),
		metrics:          NewMetricsCollector(logger),
		logger:           logger,
		defaultConfig:    config.DefaultConfig,
		enableMonitoring: config.EnableMonitoring,
		monitorInterval:  config.MonitorInterval,
		stopMonitor:      make(chan struct{}),
	}

	// Start monitoring if enabled
	if manager.enableMonitoring {
		go manager.startMonitoring()
	}

	return manager
}

// CreateCircuit creates a new circuit breaker
func (m *Manager) CreateCircuit(name string, config *Config) (*CircuitBreaker, error) {
	if name == "" {
		return nil, fmt.Errorf("circuit name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if circuit already exists
	if _, exists := m.circuits[name]; exists {
		return nil, fmt.Errorf("circuit breaker %s already exists", name)
	}

	// Use default config if none provided
	if config == nil {
		config = m.defaultConfig
	}

	// Create circuit breaker
	circuit := NewCircuitBreaker(name, config, m.logger)
	m.circuits[name] = circuit

	// Register with metrics collector
	m.metrics.RegisterCircuit(name, circuit.GetMetrics())

	if m.logger != nil {
		m.logger.Info("Circuit breaker created",
			zap.String("circuit", name),
			zap.Int("failure_threshold", config.FailureThreshold),
			zap.Duration("open_timeout", config.OpenTimeout))
	}

	return circuit, nil
}

// GetCircuit returns a circuit breaker by name
func (m *Manager) GetCircuit(name string) (*CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	circuit, exists := m.circuits[name]
	return circuit, exists
}

// RemoveCircuit removes a circuit breaker
func (m *Manager) RemoveCircuit(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	circuit, exists := m.circuits[name]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	// Unregister from metrics
	m.metrics.UnregisterCircuit(name)

	// Remove from circuits
	delete(m.circuits, name)

	if m.logger != nil {
		m.logger.Info("Circuit breaker removed",
			zap.String("circuit", name))
	}

	return nil
}

// Execute executes a function with circuit breaker protection
func (m *Manager) Execute(ctx context.Context, circuitName string, fn func() (interface{}, error)) (interface{}, error) {
	circuit, exists := m.GetCircuit(circuitName)
	if !exists {
		return nil, fmt.Errorf("circuit breaker %s not found", circuitName)
	}

	return circuit.Execute(ctx, fn)
}

// ExecuteAsync executes a function asynchronously with circuit breaker protection
func (m *Manager) ExecuteAsync(ctx context.Context, circuitName string, fn func() (interface{}, error)) (<-chan Result, error) {
	circuit, exists := m.GetCircuit(circuitName)
	if !exists {
		return nil, fmt.Errorf("circuit breaker %s not found", circuitName)
	}

	return circuit.ExecuteAsync(ctx, fn), nil
}

// GetCircuitNames returns all circuit breaker names
func (m *Manager) GetCircuitNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.circuits))
	for name := range m.circuits {
		names = append(names, name)
	}
	return names
}

// GetCircuitStats returns statistics for a specific circuit
func (m *Manager) GetCircuitStats(circuitName string) (Stats, error) {
	circuit, exists := m.GetCircuit(circuitName)
	if !exists {
		return Stats{}, fmt.Errorf("circuit breaker %s not found", circuitName)
	}

	return circuit.GetStats(), nil
}

// GetAllStats returns statistics for all circuits
func (m *Manager) GetAllStats() map[string]Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]Stats)
	for name, circuit := range m.circuits {
		stats[name] = circuit.GetStats()
	}
	return stats
}

// GetAggregatedStats returns aggregated statistics for all circuits
func (m *Manager) GetAggregatedStats() AggregatedStats {
	return m.metrics.GetAggregatedStats()
}

// ResetCircuit resets a specific circuit breaker
func (m *Manager) ResetCircuit(circuitName string) error {
	circuit, exists := m.GetCircuit(circuitName)
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", circuitName)
	}

	circuit.Reset()

	if m.logger != nil {
		m.logger.Info("Circuit breaker reset",
			zap.String("circuit", circuitName))
	}

	return nil
}

// ResetAll resets all circuit breakers
func (m *Manager) ResetAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, circuit := range m.circuits {
		circuit.Reset()

		if m.logger != nil {
			m.logger.Info("Circuit breaker reset",
				zap.String("circuit", name))
		}
	}

	// Reset metrics
	m.metrics.ResetAll()
}

// ForceOpen forces a circuit breaker to open state
func (m *Manager) ForceOpen(circuitName string) error {
	circuit, exists := m.GetCircuit(circuitName)
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", circuitName)
	}

	circuit.ForceOpen()

	if m.logger != nil {
		m.logger.Info("Circuit breaker forced open",
			zap.String("circuit", circuitName))
	}

	return nil
}

// ForceClose forces a circuit breaker to closed state
func (m *Manager) ForceClose(circuitName string) error {
	circuit, exists := m.GetCircuit(circuitName)
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", circuitName)
	}

	circuit.ForceClose()

	if m.logger != nil {
		m.logger.Info("Circuit breaker forced closed",
			zap.String("circuit", circuitName))
	}

	return nil
}

// GetMetrics returns the metrics collector
func (m *Manager) GetMetrics() *MetricsCollector {
	return m.metrics
}

// startMonitoring starts the monitoring goroutine
func (m *Manager) startMonitoring() {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.monitorCircuits()
		case <-m.stopMonitor:
			return
		}
	}
}

// monitorCircuits monitors all circuit breakers
func (m *Manager) monitorCircuits() {
	m.mu.RLock()
	circuits := make(map[string]*CircuitBreaker)
	for name, circuit := range m.circuits {
		circuits[name] = circuit
	}
	m.mu.RUnlock()

	for name, circuit := range circuits {
		stats := circuit.GetStats()

		// Log circuit state if it's open
		if stats.State == StateOpen && m.logger != nil {
			m.logger.Warn("Circuit breaker is open",
				zap.String("circuit", name),
				zap.Float64("failure_rate", stats.FailureRate),
				zap.Int("failure_count", stats.FailureCount))
		}

		// Update metrics
		if circuit.GetMetrics() != nil {
			circuit.GetMetrics().RecordStateChange(stats.State)
		}
	}
}

// Stop stops the manager and its monitoring
func (m *Manager) Stop() {
	if m.enableMonitoring {
		close(m.stopMonitor)
	}

	if m.logger != nil {
		m.logger.Info("Circuit breaker manager stopped")
	}
}

// GetManagerStats returns manager statistics
func (m *Manager) GetManagerStats() ManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	circuitCount := len(m.circuits)
	openCircuits := 0
	closedCircuits := 0
	halfOpenCircuits := 0

	for _, circuit := range m.circuits {
		switch circuit.GetState() {
		case StateOpen:
			openCircuits++
		case StateClosed:
			closedCircuits++
		case StateHalfOpen:
			halfOpenCircuits++
		}
	}

	return ManagerStats{
		CircuitCount:      circuitCount,
		OpenCircuits:      openCircuits,
		ClosedCircuits:    closedCircuits,
		HalfOpenCircuits:  halfOpenCircuits,
		MonitoringEnabled: m.enableMonitoring,
		MonitorInterval:   m.monitorInterval,
	}
}

// ManagerStats represents manager statistics
type ManagerStats struct {
	CircuitCount      int           `json:"circuit_count"`
	OpenCircuits      int           `json:"open_circuits"`
	ClosedCircuits    int           `json:"closed_circuits"`
	HalfOpenCircuits  int           `json:"half_open_circuits"`
	MonitoringEnabled bool          `json:"monitoring_enabled"`
	MonitorInterval   time.Duration `json:"monitor_interval"`
}

// ManagerConfigFromEnv creates manager config from environment variables
func ManagerConfigFromEnv() *ManagerConfig {
	config := DefaultManagerConfig()

	// Override with environment variables if present
	// This would typically read from environment variables
	// For now, return the default config
	return config
}
