package graceful

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DrainConfig represents connection draining configuration
type DrainConfig struct {
	// Timeouts
	DrainTimeout  time.Duration `yaml:"drain_timeout" json:"drain_timeout"`
	MaxDrainWait  time.Duration `yaml:"max_drain_wait" json:"max_drain_wait"`
	CheckInterval time.Duration `yaml:"check_interval" json:"check_interval"`

	// Connection limits
	MaxConcurrent int           `yaml:"max_concurrent" json:"max_concurrent"`
	MaxIdleTime   time.Duration `yaml:"max_idle_time" json:"max_idle_time"`

	// Graceful close
	EnableGracefulClose bool          `yaml:"enable_graceful_close" json:"enable_graceful_close"`
	CloseDelay          time.Duration `yaml:"close_delay" json:"close_delay"`

	// Logging
	LogDrainEvents bool `yaml:"log_drain_events" json:"log_drain_events"`
}

// DefaultDrainConfig returns default drain configuration
func DefaultDrainConfig() *DrainConfig {
	return &DrainConfig{
		DrainTimeout:        5 * time.Second,
		MaxDrainWait:        30 * time.Second,
		CheckInterval:       100 * time.Millisecond,
		MaxConcurrent:       1000,
		MaxIdleTime:         30 * time.Second,
		EnableGracefulClose: true,
		CloseDelay:          1 * time.Second,
		LogDrainEvents:      true,
	}
}

// ConnectionTracker tracks active connections for draining
type ConnectionTracker struct {
	config *DrainConfig
	logger *zap.Logger

	// Connection tracking
	connections map[net.Conn]*ConnectionInfo
	mu          sync.RWMutex

	// State
	draining bool
	drainMu  sync.RWMutex

	// Channels
	drainChan chan struct{}
	doneChan  chan struct{}
}

// ConnectionInfo holds information about a tracked connection
type ConnectionInfo struct {
	Conn         net.Conn
	StartTime    time.Time
	LastActivity time.Time
	RequestCount int
	IsIdle       bool
}

// NewConnectionTracker creates a new connection tracker
func NewConnectionTracker(config *DrainConfig, logger *zap.Logger) *ConnectionTracker {
	if config == nil {
		config = DefaultDrainConfig()
	}

	return &ConnectionTracker{
		config:      config,
		logger:      logger,
		connections: make(map[net.Conn]*ConnectionInfo),
		drainChan:   make(chan struct{}),
		doneChan:    make(chan struct{}),
	}
}

// TrackConnection starts tracking a connection
func (ct *ConnectionTracker) TrackConnection(conn net.Conn) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	now := time.Now()
	ct.connections[conn] = &ConnectionInfo{
		Conn:         conn,
		StartTime:    now,
		LastActivity: now,
		RequestCount: 0,
		IsIdle:       false,
	}

	if ct.config.LogDrainEvents {
		ct.logger.Debug("Connection tracked",
			zap.String("remote_addr", conn.RemoteAddr().String()),
			zap.Time("start_time", now))
	}
}

// UntrackConnection stops tracking a connection
func (ct *ConnectionTracker) UntrackConnection(conn net.Conn) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if info, exists := ct.connections[conn]; exists {
		if ct.config.LogDrainEvents {
			ct.logger.Debug("Connection untracked",
				zap.String("remote_addr", conn.RemoteAddr().String()),
				zap.Duration("duration", time.Since(info.StartTime)),
				zap.Int("request_count", info.RequestCount))
		}
		delete(ct.connections, conn)
	}
}

// UpdateActivity updates the last activity time for a connection
func (ct *ConnectionTracker) UpdateActivity(conn net.Conn) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if info, exists := ct.connections[conn]; exists {
		info.LastActivity = time.Now()
		info.RequestCount++
		info.IsIdle = false
	}
}

// MarkIdle marks a connection as idle
func (ct *ConnectionTracker) MarkIdle(conn net.Conn) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if info, exists := ct.connections[conn]; exists {
		info.IsIdle = true
	}
}

// GetConnectionCount returns the number of active connections
func (ct *ConnectionTracker) GetConnectionCount() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return len(ct.connections)
}

// GetIdleConnectionCount returns the number of idle connections
func (ct *ConnectionTracker) GetIdleConnectionCount() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	count := 0
	for _, info := range ct.connections {
		if info.IsIdle {
			count++
		}
	}
	return count
}

// GetActiveConnectionCount returns the number of active (non-idle) connections
func (ct *ConnectionTracker) GetActiveConnectionCount() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	count := 0
	for _, info := range ct.connections {
		if !info.IsIdle {
			count++
		}
	}
	return count
}

// StartDraining starts the connection draining process
func (ct *ConnectionTracker) StartDraining(ctx context.Context) error {
	ct.drainMu.Lock()
	if ct.draining {
		ct.drainMu.Unlock()
		return fmt.Errorf("draining already in progress")
	}
	ct.draining = true
	ct.drainMu.Unlock()

	if ct.config.LogDrainEvents {
		ct.logger.Info("Starting connection draining",
			zap.Duration("drain_timeout", ct.config.DrainTimeout),
			zap.Duration("max_drain_wait", ct.config.MaxDrainWait))
	}

	// Start draining process
	go func() {
		defer close(ct.doneChan)
		ct.performDraining(ctx)
	}()

	return nil
}

// WaitForDraining waits for the draining process to complete
func (ct *ConnectionTracker) WaitForDraining(ctx context.Context) error {
	select {
	case <-ct.doneChan:
		if ct.config.LogDrainEvents {
			ct.logger.Info("Connection draining completed")
		}
		return nil
	case <-ctx.Done():
		ct.logger.Warn("Connection draining timed out",
			zap.Duration("timeout", ct.config.DrainTimeout))
		return fmt.Errorf("draining timeout after %v", ct.config.DrainTimeout)
	}
}

// IsDraining returns true if draining is in progress
func (ct *ConnectionTracker) IsDraining() bool {
	ct.drainMu.RLock()
	defer ct.drainMu.RUnlock()
	return ct.draining
}

// performDraining performs the actual connection draining
func (ct *ConnectionTracker) performDraining(ctx context.Context) {
	// Create drain context with timeout
	drainCtx, cancel := context.WithTimeout(ctx, ct.config.MaxDrainWait)
	defer cancel()

	// Start monitoring goroutine
	go ct.monitorConnections(drainCtx)

	// Wait for all connections to drain
	ct.waitForConnectionsToDrain(drainCtx)
}

// monitorConnections monitors connections during draining
func (ct *ConnectionTracker) monitorConnections(ctx context.Context) {
	ticker := time.NewTicker(ct.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ct.checkIdleConnections()
			ct.logDrainProgress()
		}
	}
}

// checkIdleConnections checks for idle connections and closes them
func (ct *ConnectionTracker) checkIdleConnections() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	now := time.Now()
	toClose := make([]net.Conn, 0)

	for conn, info := range ct.connections {
		// Check if connection is idle for too long
		if info.IsIdle && now.Sub(info.LastActivity) > ct.config.MaxIdleTime {
			toClose = append(toClose, conn)
		}
	}

	// Close idle connections
	for _, conn := range toClose {
		if ct.config.EnableGracefulClose {
			// Set read deadline to force connection close
			conn.SetReadDeadline(time.Now().Add(ct.config.CloseDelay))
		} else {
			conn.Close()
		}
		delete(ct.connections, conn)
	}
}

// waitForConnectionsToDrain waits for all connections to be drained
func (ct *ConnectionTracker) waitForConnectionsToDrain(ctx context.Context) {
	ticker := time.NewTicker(ct.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if ct.config.LogDrainEvents {
				ct.logger.Warn("Connection drain timeout reached",
					zap.Duration("timeout", ct.config.MaxDrainWait))
			}
			return
		case <-ticker.C:
			ct.mu.RLock()
			activeCount := ct.GetActiveConnectionCount()
			idleCount := ct.GetIdleConnectionCount()
			totalCount := len(ct.connections)
			ct.mu.RUnlock()

			if totalCount == 0 {
				if ct.config.LogDrainEvents {
					ct.logger.Info("All connections drained")
				}
				return
			}

			if ct.config.LogDrainEvents {
				ct.logger.Debug("Waiting for connections to drain",
					zap.Int("active_connections", activeCount),
					zap.Int("idle_connections", idleCount),
					zap.Int("total_connections", totalCount))
			}
		}
	}
}

// logDrainProgress logs the current draining progress
func (ct *ConnectionTracker) logDrainProgress() {
	ct.mu.RLock()
	activeCount := ct.GetActiveConnectionCount()
	idleCount := ct.GetIdleConnectionCount()
	totalCount := len(ct.connections)
	ct.mu.RUnlock()

	if ct.config.LogDrainEvents && totalCount > 0 {
		ct.logger.Info("Drain progress",
			zap.Int("active_connections", activeCount),
			zap.Int("idle_connections", idleCount),
			zap.Int("total_connections", totalCount))
	}
}

// ForceCloseAllConnections forcefully closes all tracked connections
func (ct *ConnectionTracker) ForceCloseAllConnections() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if ct.config.LogDrainEvents {
		ct.logger.Warn("Force closing all connections",
			zap.Int("count", len(ct.connections)))
	}

	for conn := range ct.connections {
		conn.Close()
	}

	ct.connections = make(map[net.Conn]*ConnectionInfo)
}

// GetConnectionStats returns statistics about tracked connections
func (ct *ConnectionTracker) GetConnectionStats() map[string]interface{} {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	activeCount := 0
	idleCount := 0
	totalRequests := 0
	oldestConnection := time.Now()

	for _, info := range ct.connections {
		if info.IsIdle {
			idleCount++
		} else {
			activeCount++
		}
		totalRequests += info.RequestCount
		if info.StartTime.Before(oldestConnection) {
			oldestConnection = info.StartTime
		}
	}

	return map[string]interface{}{
		"total_connections":  len(ct.connections),
		"active_connections": activeCount,
		"idle_connections":   idleCount,
		"total_requests":     totalRequests,
		"oldest_connection":  oldestConnection,
		"is_draining":        ct.IsDraining(),
	}
}

// DrainConfigFromEnv creates drain config from environment variables
func DrainConfigFromEnv() *DrainConfig {
	config := DefaultDrainConfig()

	// Override with environment variables if present
	if drainTimeout := os.Getenv("DRAIN_TIMEOUT"); drainTimeout != "" {
		if d, err := time.ParseDuration(drainTimeout); err == nil {
			config.DrainTimeout = d
		}
	}
	if maxDrainWait := os.Getenv("MAX_DRAIN_WAIT"); maxDrainWait != "" {
		if d, err := time.ParseDuration(maxDrainWait); err == nil {
			config.MaxDrainWait = d
		}
	}
	if checkInterval := os.Getenv("DRAIN_CHECK_INTERVAL"); checkInterval != "" {
		if d, err := time.ParseDuration(checkInterval); err == nil {
			config.CheckInterval = d
		}
	}
	if maxIdleTime := os.Getenv("MAX_IDLE_TIME"); maxIdleTime != "" {
		if d, err := time.ParseDuration(maxIdleTime); err == nil {
			config.MaxIdleTime = d
		}
	}

	return config
}
