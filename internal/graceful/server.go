package graceful

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// GracefulServer wraps an HTTP server with graceful shutdown capabilities
type GracefulServer struct {
	server   *http.Server
	config   *ShutdownConfig
	logger   *zap.Logger
	
	// Connection tracking
	tracker  *ConnectionTracker
	
	// State
	listener net.Listener
	mu       sync.RWMutex
}

// NewGracefulServer creates a new graceful HTTP server
func NewGracefulServer(server *http.Server, config *ShutdownConfig, logger *zap.Logger) *GracefulServer {
	if config == nil {
		config = DefaultShutdownConfig()
	}
	
	// Create connection tracker
	tracker := NewConnectionTracker(&DrainConfig{
		DrainTimeout:       config.DrainTimeout,
		MaxDrainWait:       config.MaxDrainWait,
		CheckInterval:      100 * time.Millisecond,
		MaxConcurrent:      1000,
		MaxIdleTime:        30 * time.Second,
		EnableGracefulClose: true,
		CloseDelay:         1 * time.Second,
		LogDrainEvents:     config.LogShutdownEvents,
	}, logger)
	
	// Wrap the handler with connection tracking
	originalHandler := server.Handler
	server.Handler = &connectionTrackingHandler{
		handler: originalHandler,
		tracker: tracker,
		logger:  logger,
	}
	
	return &GracefulServer{
		server:  server,
		config:  config,
		logger:  logger,
		tracker: tracker,
	}
}

// ListenAndServe starts the server with graceful shutdown support
func (gs *GracefulServer) ListenAndServe() error {
	// Create listener
	listener, err := net.Listen("tcp", gs.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	
	gs.mu.Lock()
	gs.listener = listener
	gs.mu.Unlock()
	
	gs.logger.Info("Starting graceful HTTP server",
		zap.String("addr", gs.server.Addr))
	
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- gs.server.Serve(gs.trackedListener())
	}()
	
	// Wait for server to start or fail
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	case <-time.After(100 * time.Millisecond):
		// Server started successfully
		gs.logger.Info("Graceful HTTP server started successfully")
		return nil
	}
}

// Shutdown gracefully shuts down the server
func (gs *GracefulServer) Shutdown(ctx context.Context) error {
	gs.logger.Info("Initiating graceful server shutdown")
	
	// Start connection draining
	if err := gs.tracker.StartDraining(ctx); err != nil {
		gs.logger.Error("Failed to start connection draining", zap.Error(err))
	} else {
		// Wait for connections to drain
		drainCtx, cancel := context.WithTimeout(ctx, gs.config.DrainTimeout)
		defer cancel()
		
		if err := gs.tracker.WaitForDraining(drainCtx); err != nil {
			gs.logger.Warn("Connection draining timed out", zap.Error(err))
		}
	}
	
	// Shutdown the HTTP server
	if err := gs.server.Shutdown(ctx); err != nil {
		gs.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		return err
	}
	
	gs.logger.Info("Graceful server shutdown completed")
	return nil
}

// GetConnectionStats returns connection statistics
func (gs *GracefulServer) GetConnectionStats() map[string]interface{} {
	return gs.tracker.GetConnectionStats()
}

// GetActiveConnectionCount returns the number of active connections
func (gs *GracefulServer) GetActiveConnectionCount() int {
	return gs.tracker.GetActiveConnectionCount()
}

// GetIdleConnectionCount returns the number of idle connections
func (gs *GracefulServer) GetIdleConnectionCount() int {
	return gs.tracker.GetIdleConnectionCount()
}

// IsDraining returns true if the server is draining connections
func (gs *GracefulServer) IsDraining() bool {
	return gs.tracker.IsDraining()
}

// trackedListener returns a listener that tracks connections
func (gs *GracefulServer) trackedListener() net.Listener {
	return &trackedListener{
		Listener: gs.listener,
		tracker:  gs.tracker,
		logger:   gs.logger,
	}
}

// trackedListener wraps a net.Listener to track connections
type trackedListener struct {
	net.Listener
	tracker *ConnectionTracker
	logger  *zap.Logger
}

// Accept tracks accepted connections
func (tl *trackedListener) Accept() (net.Conn, error) {
	conn, err := tl.Listener.Accept()
	if err != nil {
		return nil, err
	}
	
	// Track the connection
	tl.tracker.TrackConnection(conn)
	
	return &trackedConn{
		Conn:    conn,
		tracker: tl.tracker,
		logger:  tl.logger,
	}, nil
}

// trackedConn wraps a net.Conn to track connection activity
type trackedConn struct {
	net.Conn
	tracker *ConnectionTracker
	logger  *zap.Logger
}

// Read tracks read activity
func (tc *trackedConn) Read(b []byte) (int, error) {
	n, err := tc.Conn.Read(b)
	if n > 0 {
		tc.tracker.UpdateActivity(tc.Conn)
	}
	return n, err
}

// Write tracks write activity
func (tc *trackedConn) Write(b []byte) (int, error) {
	n, err := tc.Conn.Write(b)
	if n > 0 {
		tc.tracker.UpdateActivity(tc.Conn)
	}
	return n, err
}

// Close untracks the connection when closed
func (tc *trackedConn) Close() error {
	tc.tracker.UntrackConnection(tc.Conn)
	return tc.Conn.Close()
}

// connectionTrackingHandler wraps an HTTP handler to track connections
type connectionTrackingHandler struct {
	handler http.Handler
	tracker *ConnectionTracker
	logger  *zap.Logger
}

// ServeHTTP tracks HTTP requests and responses
func (cth *connectionTrackingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get connection from request context
	if conn := cth.getConnFromRequest(r); conn != nil {
		// Update activity
		cth.tracker.UpdateActivity(conn)
		
		// Wrap response writer to track completion
		w = &trackedResponseWriter{
			ResponseWriter: w,
			conn:          conn,
			tracker:       cth.tracker,
		}
	}
	
	// Serve the request
	cth.handler.ServeHTTP(w, r)
}

// getConnFromRequest extracts the connection from the request context
func (cth *connectionTrackingHandler) getConnFromRequest(r *http.Request) net.Conn {
	// This is a simplified approach - in a real implementation,
	// you might need to store the connection in the request context
	// or use a different method to track connections per request
	
	// For now, we'll return nil and handle tracking at the connection level
	return nil
}

// trackedResponseWriter wraps http.ResponseWriter to track response completion
type trackedResponseWriter struct {
	http.ResponseWriter
	conn    net.Conn
	tracker *ConnectionTracker
}

// Write tracks write activity
func (trw *trackedResponseWriter) Write(b []byte) (int, error) {
	n, err := trw.ResponseWriter.Write(b)
	if n > 0 && trw.conn != nil {
		trw.tracker.UpdateActivity(trw.conn)
	}
	return n, err
}

// WriteHeader tracks header writing
func (trw *trackedResponseWriter) WriteHeader(statusCode int) {
	trw.ResponseWriter.WriteHeader(statusCode)
	if trw.conn != nil {
		trw.tracker.UpdateActivity(trw.conn)
	}
}

// Flush flushes the response
func (trw *trackedResponseWriter) Flush() {
	if flusher, ok := trw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack hijacks the connection
func (trw *trackedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := trw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not support hijacking")
}

// Push pushes a resource
func (trw *trackedResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := trw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// GracefulServerManager manages multiple graceful servers
type GracefulServerManager struct {
	servers []*GracefulServer
	config  *ShutdownConfig
	logger  *zap.Logger
	mu      sync.RWMutex
}

// NewGracefulServerManager creates a new server manager
func NewGracefulServerManager(config *ShutdownConfig, logger *zap.Logger) *GracefulServerManager {
	return &GracefulServerManager{
		servers: make([]*GracefulServer, 0),
		config:  config,
		logger:  logger,
	}
}

// AddServer adds a server to the manager
func (gsm *GracefulServerManager) AddServer(server *GracefulServer) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	
	gsm.servers = append(gsm.servers, server)
	gsm.logger.Info("Server added to graceful manager",
		zap.String("addr", server.server.Addr))
}

// StartAll starts all managed servers
func (gsm *GracefulServerManager) StartAll() error {
	gsm.mu.RLock()
	servers := make([]*GracefulServer, len(gsm.servers))
	copy(servers, gsm.servers)
	gsm.mu.RUnlock()
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(servers))
	
	for _, server := range servers {
		wg.Add(1)
		go func(s *GracefulServer) {
			defer wg.Done()
			if err := s.ListenAndServe(); err != nil {
				errChan <- err
			}
		}(server)
	}
	
	// Wait for all servers to start
	wg.Wait()
	
	// Check for errors
	select {
	case err := <-errChan:
		return err
	default:
		gsm.logger.Info("All servers started successfully")
		return nil
	}
}

// ShutdownAll gracefully shuts down all managed servers
func (gsm *GracefulServerManager) ShutdownAll(ctx context.Context) error {
	gsm.mu.RLock()
	servers := make([]*GracefulServer, len(gsm.servers))
	copy(servers, gsm.servers)
	gsm.mu.RUnlock()
	
	gsm.logger.Info("Shutting down all servers",
		zap.Int("server_count", len(servers)))
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(servers))
	
	for _, server := range servers {
		wg.Add(1)
		go func(s *GracefulServer) {
			defer wg.Done()
			if err := s.Shutdown(ctx); err != nil {
				errChan <- err
			}
		}(server)
	}
	
	// Wait for all servers to shutdown
	wg.Wait()
	
	// Check for errors
	select {
	case err := <-errChan:
		gsm.logger.Error("Some servers failed to shutdown gracefully", zap.Error(err))
		return err
	default:
		gsm.logger.Info("All servers shutdown successfully")
		return nil
	}
}

// GetTotalConnectionCount returns the total number of connections across all servers
func (gsm *GracefulServerManager) GetTotalConnectionCount() int {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	
	total := 0
	for _, server := range gsm.servers {
		total += server.GetActiveConnectionCount()
	}
	return total
}

// GetServerStats returns statistics for all servers
func (gsm *GracefulServerManager) GetServerStats() map[string]interface{} {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	
	stats := make(map[string]interface{})
	serverStats := make([]map[string]interface{}, 0, len(gsm.servers))
	
	for i, server := range gsm.servers {
		serverStat := map[string]interface{}{
			"index":              i,
			"addr":               server.server.Addr,
			"connection_stats":   server.GetConnectionStats(),
			"is_draining":        server.IsDraining(),
		}
		serverStats = append(serverStats, serverStat)
	}
	
	stats["servers"] = serverStats
	stats["total_servers"] = len(gsm.servers)
	stats["total_connections"] = gsm.GetTotalConnectionCount()
	
	return stats
}
