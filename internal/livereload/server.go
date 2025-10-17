package livereload

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// HotReloadServer provides hot reload functionality for browsers
type HotReloadServer struct {
	port   int
	paths  []string
	logger *zap.Logger

	// WebSocket connections
	connections map[*websocket.Conn]bool
	connMu      sync.RWMutex

	// HTTP server
	server *http.Server

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// NewHotReloadServer creates a new hot reload server
func NewHotReloadServer(port int, paths []string, logger *zap.Logger) (*HotReloadServer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &HotReloadServer{
		port:        port,
		paths:       paths,
		logger:      logger,
		connections: make(map[*websocket.Conn]bool),
		ctx:         ctx,
		cancel:      cancel,
		done:        make(chan struct{}),
	}, nil
}

// Start starts the hot reload server
func (hrs *HotReloadServer) Start() error {
	// Create HTTP server
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/livereload", hrs.handleWebSocket)

	// Script injection endpoint
	mux.HandleFunc("/livereload.js", hrs.handleScript)

	// Health check endpoint
	mux.HandleFunc("/health", hrs.handleHealth)

	hrs.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", hrs.port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		defer close(hrs.done)

		if err := hrs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if hrs.logger != nil {
				hrs.logger.Error("Hot reload server error", zap.Error(err))
			}
		}
	}()

	if hrs.logger != nil {
		hrs.logger.Info("Hot reload server started",
			zap.Int("port", hrs.port),
			zap.Strings("paths", hrs.paths))
	}

	return nil
}

// Stop stops the hot reload server
func (hrs *HotReloadServer) Stop() error {
	// Cancel context
	hrs.cancel()

	// Close all WebSocket connections
	hrs.connMu.Lock()
	for conn := range hrs.connections {
		conn.Close()
	}
	hrs.connections = make(map[*websocket.Conn]bool)
	hrs.connMu.Unlock()

	// Shutdown HTTP server
	if hrs.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := hrs.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown hot reload server: %w", err)
		}
	}

	// Wait for server to stop
	<-hrs.done

	if hrs.logger != nil {
		hrs.logger.Info("Hot reload server stopped")
	}

	return nil
}

// handleWebSocket handles WebSocket connections
func (hrs *HotReloadServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if hrs.logger != nil {
			hrs.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		}
		return
	}

	// Add connection
	hrs.connMu.Lock()
	hrs.connections[conn] = true
	hrs.connMu.Unlock()

	if hrs.logger != nil {
		hrs.logger.Debug("WebSocket connection established")
	}

	// Handle connection
	go hrs.handleConnection(conn)
}

// handleConnection handles a WebSocket connection
func (hrs *HotReloadServer) handleConnection(conn *websocket.Conn) {
	defer func() {
		// Remove connection
		hrs.connMu.Lock()
		delete(hrs.connections, conn)
		hrs.connMu.Unlock()

		conn.Close()
	}()

	// Send initial message
	conn.WriteMessage(websocket.TextMessage, []byte(`{"command":"hello"}`))

	// Keep connection alive
	for {
		select {
		case <-hrs.ctx.Done():
			return
		default:
			// Read message (we don't need to process it, just keep connection alive)
			_, _, err := conn.ReadMessage()
			if err != nil {
				if hrs.logger != nil {
					hrs.logger.Debug("WebSocket connection closed", zap.Error(err))
				}
				return
			}
		}
	}
}

// handleScript handles the livereload script request
func (hrs *HotReloadServer) handleScript(w http.ResponseWriter, r *http.Request) {
	script := hrs.generateScript()

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(script)
}

// generateScript generates the livereload script
func (hrs *HotReloadServer) generateScript() []byte {
	script := fmt.Sprintf(`
(function() {
  var script = document.createElement('script');
  script.src = 'http://localhost:%d/livereload.js';
  script.async = true;
  document.head.appendChild(script);
  
  var ws = new WebSocket('ws://localhost:%d/livereload');
  
  ws.onopen = function() {
    console.log('Live reload connected');
  };
  
  ws.onmessage = function(event) {
    var data = JSON.parse(event.data);
    if (data.command === 'reload') {
      console.log('Live reload triggered');
      window.location.reload();
    }
  };
  
  ws.onclose = function() {
    console.log('Live reload disconnected');
  };
  
  ws.onerror = function(error) {
    console.error('Live reload error:', error);
  };
})();
`, hrs.port, hrs.port)

	return []byte(script)
}

// handleHealth handles health check requests
func (hrs *HotReloadServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	hrs.connMu.RLock()
	connectionCount := len(hrs.connections)
	hrs.connMu.RUnlock()

	response := fmt.Sprintf(`{
  "status": "healthy",
  "port": %d,
  "connections": %d,
  "paths": %q
}`, hrs.port, connectionCount, hrs.paths)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

// NotifyReload notifies all connected clients to reload
func (hrs *HotReloadServer) NotifyReload() {
	hrs.connMu.RLock()
	defer hrs.connMu.RUnlock()

	message := []byte(`{"command":"reload"}`)

	for conn := range hrs.connections {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			if hrs.logger != nil {
				hrs.logger.Debug("Failed to send reload message", zap.Error(err))
			}
			// Remove failed connection
			delete(hrs.connections, conn)
			conn.Close()
		}
	}

	if hrs.logger != nil {
		hrs.logger.Debug("Reload notification sent",
			zap.Int("connections", len(hrs.connections)))
	}
}

// GetConnectionCount returns the number of active connections
func (hrs *HotReloadServer) GetConnectionCount() int {
	hrs.connMu.RLock()
	defer hrs.connMu.RUnlock()
	return len(hrs.connections)
}

// GetStats returns server statistics
func (hrs *HotReloadServer) GetStats() map[string]interface{} {
	hrs.connMu.RLock()
	defer hrs.connMu.RUnlock()

	return map[string]interface{}{
		"port":        hrs.port,
		"paths":       hrs.paths,
		"connections": len(hrs.connections),
		"is_running":  hrs.server != nil,
	}
}
