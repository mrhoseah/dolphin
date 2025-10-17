package graceful

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// ShutdownConfig represents graceful shutdown configuration
type ShutdownConfig struct {
	// Timeouts
	ShutdownTimeout    time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
	ReadTimeout        time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout       time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	
	// Connection draining
	DrainTimeout       time.Duration `yaml:"drain_timeout" json:"drain_timeout"`
	MaxDrainWait       time.Duration `yaml:"max_drain_wait" json:"max_drain_wait"`
	
	// Signal handling
	EnableSignalHandling bool `yaml:"enable_signal_handling" json:"enable_signal_handling"`
	Signals             []os.Signal `yaml:"signals" json:"signals"`
	
	// Health checks
	EnableHealthCheck  bool          `yaml:"enable_health_check" json:"enable_health_check"`
	HealthCheckPath    string        `yaml:"health_check_path" json:"health_check_path"`
	HealthCheckTimeout time.Duration `yaml:"health_check_timeout" json:"health_check_timeout"`
	
	// Logging
	LogShutdownEvents bool `yaml:"log_shutdown_events" json:"log_shutdown_events"`
}

// DefaultShutdownConfig returns default shutdown configuration
func DefaultShutdownConfig() *ShutdownConfig {
	return &ShutdownConfig{
		ShutdownTimeout:     30 * time.Second,
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		IdleTimeout:         60 * time.Second,
		DrainTimeout:        5 * time.Second,
		MaxDrainWait:        30 * time.Second,
		EnableSignalHandling: true,
		Signals:             []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		EnableHealthCheck:   true,
		HealthCheckPath:     "/health",
		HealthCheckTimeout:  5 * time.Second,
		LogShutdownEvents:   true,
	}
}

// ShutdownManager manages graceful shutdown of services
type ShutdownManager struct {
	config     *ShutdownConfig
	logger     *zap.Logger
	
	// Services to shutdown
	httpServers []*http.Server
	services    []Shutdownable
	
	// State management
	shuttingDown bool
	mu           sync.RWMutex
	
	// Channels for coordination
	shutdownChan chan struct{}
	doneChan     chan struct{}
	
	// Connection tracking
	activeConns  map[net.Conn]bool
	connMu       sync.RWMutex
	
	// Health check
	healthStatus bool
	healthMu     sync.RWMutex
}

// Shutdownable represents a service that can be gracefully shutdown
type Shutdownable interface {
	Shutdown(ctx context.Context) error
	Name() string
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(config *ShutdownConfig, logger *zap.Logger) *ShutdownManager {
	if config == nil {
		config = DefaultShutdownConfig()
	}
	
	return &ShutdownManager{
		config:       config,
		logger:       logger,
		httpServers:  make([]*http.Server, 0),
		services:     make([]Shutdownable, 0),
		shutdownChan: make(chan struct{}),
		doneChan:     make(chan struct{}),
		activeConns:  make(map[net.Conn]bool),
		healthStatus: true,
	}
}

// RegisterHTTPServer registers an HTTP server for graceful shutdown
func (sm *ShutdownManager) RegisterHTTPServer(server *http.Server) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sm.httpServers = append(sm.httpServers, server)
	
	// Set timeouts
	server.ReadTimeout = sm.config.ReadTimeout
	server.WriteTimeout = sm.config.WriteTimeout
	server.IdleTimeout = sm.config.IdleTimeout
	
	sm.logger.Info("HTTP server registered for graceful shutdown",
		zap.String("addr", server.Addr))
}

// RegisterService registers a service for graceful shutdown
func (sm *ShutdownManager) RegisterService(service Shutdownable) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sm.services = append(sm.services, service)
	
	sm.logger.Info("Service registered for graceful shutdown",
		zap.String("service", service.Name()))
}

// Start starts the shutdown manager and signal handling
func (sm *ShutdownManager) Start() error {
	if !sm.config.EnableSignalHandling {
		return nil
	}
	
	// Start signal handling goroutine
	go sm.handleSignals()
	
	// Start health check if enabled
	if sm.config.EnableHealthCheck {
		go sm.startHealthCheck()
	}
	
	sm.logger.Info("Graceful shutdown manager started",
		zap.Duration("shutdown_timeout", sm.config.ShutdownTimeout),
		zap.Duration("drain_timeout", sm.config.DrainTimeout))
	
	return nil
}

// Shutdown initiates graceful shutdown
func (sm *ShutdownManager) Shutdown(ctx context.Context) error {
	sm.mu.Lock()
	if sm.shuttingDown {
		sm.mu.Unlock()
		return fmt.Errorf("shutdown already in progress")
	}
	sm.shuttingDown = true
	sm.mu.Unlock()
	
	if sm.config.LogShutdownEvents {
		sm.logger.Info("Graceful shutdown initiated")
	}
	
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, sm.config.ShutdownTimeout)
	defer cancel()
	
	// Start shutdown process
	go func() {
		defer close(sm.doneChan)
		sm.performShutdown(shutdownCtx)
	}()
	
	// Wait for shutdown to complete or timeout
	select {
	case <-sm.doneChan:
		if sm.config.LogShutdownEvents {
			sm.logger.Info("Graceful shutdown completed")
		}
		return nil
	case <-shutdownCtx.Done():
		sm.logger.Error("Graceful shutdown timed out",
			zap.Duration("timeout", sm.config.ShutdownTimeout))
		return fmt.Errorf("shutdown timeout after %v", sm.config.ShutdownTimeout)
	}
}

// Wait blocks until shutdown is complete
func (sm *ShutdownManager) Wait() {
	<-sm.doneChan
}

// IsShuttingDown returns true if shutdown is in progress
func (sm *ShutdownManager) IsShuttingDown() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.shuttingDown
}

// SetHealthStatus sets the health status for health checks
func (sm *ShutdownManager) SetHealthStatus(healthy bool) {
	sm.healthMu.Lock()
	defer sm.healthMu.Unlock()
	sm.healthStatus = healthy
}

// GetHealthStatus returns the current health status
func (sm *ShutdownManager) GetHealthStatus() bool {
	sm.healthMu.RLock()
	defer sm.healthMu.RUnlock()
	return sm.healthStatus
}

// performShutdown performs the actual shutdown process
func (sm *ShutdownManager) performShutdown(ctx context.Context) {
	// Step 1: Stop accepting new connections
	sm.stopAcceptingConnections()
	
	// Step 2: Drain existing connections
	sm.drainConnections(ctx)
	
	// Step 3: Shutdown HTTP servers
	sm.shutdownHTTPServers(ctx)
	
	// Step 4: Shutdown services
	sm.shutdownServices(ctx)
	
	if sm.config.LogShutdownEvents {
		sm.logger.Info("All services shutdown completed")
	}
}

// stopAcceptingConnections stops accepting new connections
func (sm *ShutdownManager) stopAcceptingConnections() {
	if sm.config.LogShutdownEvents {
		sm.logger.Info("Stopping acceptance of new connections")
	}
	
	// This is handled by the HTTP server's Shutdown method
	// We just log the event
}

// drainConnections drains existing connections
func (sm *ShutdownManager) drainConnections(ctx context.Context) {
	if sm.config.LogShutdownEvents {
		sm.logger.Info("Draining existing connections",
			zap.Duration("timeout", sm.config.DrainTimeout))
	}
	
	// Create drain context with timeout
	drainCtx, cancel := context.WithTimeout(ctx, sm.config.DrainTimeout)
	defer cancel()
	
	// Wait for connections to drain or timeout
	sm.waitForConnectionsToDrain(drainCtx)
}

// waitForConnectionsToDrain waits for active connections to close
func (sm *ShutdownManager) waitForConnectionsToDrain(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			if sm.config.LogShutdownEvents {
				sm.logger.Warn("Connection drain timeout reached",
					zap.Duration("timeout", sm.config.DrainTimeout))
			}
			return
		case <-ticker.C:
			sm.connMu.RLock()
			activeCount := len(sm.activeConns)
			sm.connMu.RUnlock()
			
			if activeCount == 0 {
				if sm.config.LogShutdownEvents {
					sm.logger.Info("All connections drained")
				}
				return
			}
			
			if sm.config.LogShutdownEvents {
				sm.logger.Debug("Waiting for connections to drain",
					zap.Int("active_connections", activeCount))
			}
		}
	}
}

// shutdownHTTPServers shuts down all HTTP servers
func (sm *ShutdownManager) shutdownHTTPServers(ctx context.Context) {
	sm.mu.RLock()
	servers := make([]*http.Server, len(sm.httpServers))
	copy(servers, sm.httpServers)
	sm.mu.RUnlock()
	
	for _, server := range servers {
		if sm.config.LogShutdownEvents {
			sm.logger.Info("Shutting down HTTP server",
				zap.String("addr", server.Addr))
		}
		
		if err := server.Shutdown(ctx); err != nil {
			sm.logger.Error("Failed to shutdown HTTP server",
				zap.String("addr", server.Addr),
				zap.Error(err))
		} else {
			sm.logger.Info("HTTP server shutdown completed",
				zap.String("addr", server.Addr))
		}
	}
}

// shutdownServices shuts down all registered services
func (sm *ShutdownManager) shutdownServices(ctx context.Context) {
	sm.mu.RLock()
	services := make([]Shutdownable, len(sm.services))
	copy(services, sm.services)
	sm.mu.RUnlock()
	
	for _, service := range services {
		if sm.config.LogShutdownEvents {
			sm.logger.Info("Shutting down service",
				zap.String("service", service.Name()))
		}
		
		if err := service.Shutdown(ctx); err != nil {
			sm.logger.Error("Failed to shutdown service",
				zap.String("service", service.Name()),
				zap.Error(err))
		} else {
			sm.logger.Info("Service shutdown completed",
				zap.String("service", service.Name()))
		}
	}
}

// handleSignals handles OS signals for graceful shutdown
func (sm *ShutdownManager) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sm.config.Signals...)
	
	sig := <-sigChan
	
	if sm.config.LogShutdownEvents {
		sm.logger.Info("Received shutdown signal",
			zap.String("signal", sig.String()))
	}
	
	// Set health status to unhealthy
	sm.SetHealthStatus(false)
	
	// Trigger shutdown
	close(sm.shutdownChan)
}

// startHealthCheck starts the health check server
func (sm *ShutdownManager) startHealthCheck() {
	mux := http.NewServeMux()
	mux.HandleFunc(sm.config.HealthCheckPath, sm.healthCheckHandler)
	
	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}
	
	sm.logger.Info("Health check server started",
		zap.String("path", sm.config.HealthCheckPath))
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sm.logger.Error("Health check server error", zap.Error(err))
	}
}

// healthCheckHandler handles health check requests
func (sm *ShutdownManager) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	healthy := sm.GetHealthStatus()
	
	if healthy {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	}
}

// TrackConnection tracks an active connection
func (sm *ShutdownManager) TrackConnection(conn net.Conn) {
	sm.connMu.Lock()
	defer sm.connMu.Unlock()
	sm.activeConns[conn] = true
}

// UntrackConnection removes a connection from tracking
func (sm *ShutdownManager) UntrackConnection(conn net.Conn) {
	sm.connMu.Lock()
	defer sm.connMu.Unlock()
	delete(sm.activeConns, conn)
}

// GetActiveConnectionCount returns the number of active connections
func (sm *ShutdownManager) GetActiveConnectionCount() int {
	sm.connMu.RLock()
	defer sm.connMu.RUnlock()
	return len(sm.activeConns)
}

// ShutdownConfigFromEnv creates shutdown config from environment variables
func ShutdownConfigFromEnv() *ShutdownConfig {
	config := DefaultShutdownConfig()
	
	// Override with environment variables if present
	if timeout := os.Getenv("SHUTDOWN_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.ShutdownTimeout = d
		}
	}
	if drainTimeout := os.Getenv("DRAIN_TIMEOUT"); drainTimeout != "" {
		if d, err := time.ParseDuration(drainTimeout); err == nil {
			config.DrainTimeout = d
		}
	}
	if enableSignals := os.Getenv("ENABLE_SIGNAL_HANDLING"); enableSignals == "false" {
		config.EnableSignalHandling = false
	}
	if enableHealth := os.Getenv("ENABLE_HEALTH_CHECK"); enableHealth == "false" {
		config.EnableHealthCheck = false
	}
	
	return config
}
