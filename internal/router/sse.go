package router

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SSEServer handles Server-Sent Events
type SSEServer struct {
	logger *zap.Logger
}

// NewSSEServer creates a new SSE server
func NewSSEServer(logger *zap.Logger) *SSEServer {
	return &SSEServer{logger: logger}
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	ID    string
	Event string
	Data  interface{}
	Retry int
}

// HandleSSE handles Server-Sent Events connection
func (s *SSEServer) HandleSSE(w http.ResponseWriter, r *http.Request, eventChan <-chan SSEEvent) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable buffering in nginx

	// Create a flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"message\":\"connected\"}\n\n")
	flusher.Flush()

	// Keep connection alive with ping
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Handle client disconnect
	ctx := r.Context()
	done := ctx.Done()

	for {
		select {
		case <-done:
			if s.logger != nil {
				s.logger.Info("SSE client disconnected", zap.String("remote_addr", r.RemoteAddr))
			}
			return
		case event := <-eventChan:
			if err := s.writeEvent(w, event); err != nil {
				if s.logger != nil {
					s.logger.Error("Failed to write SSE event", zap.Error(err))
				}
				return
			}
			flusher.Flush()
		case <-ticker.C:
			// Send keepalive ping
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// writeEvent writes an SSE event to the response
func (s *SSEServer) writeEvent(w http.ResponseWriter, event SSEEvent) error {
	if event.ID != "" {
		fmt.Fprintf(w, "id: %s\n", event.ID)
	}
	if event.Event != "" {
		fmt.Fprintf(w, "event: %s\n", event.Event)
	}
	if event.Retry > 0 {
		fmt.Fprintf(w, "retry: %d\n", event.Retry)
	}
	
	// Convert data to string
	var dataStr string
	switch v := event.Data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		// Try to format as JSON
		dataStr = fmt.Sprintf("%v", v)
	}
	
	fmt.Fprintf(w, "data: %s\n\n", dataStr)
	return nil
}

// SendEvent sends an SSE event
func (s *SSEServer) SendEvent(w http.ResponseWriter, event SSEEvent) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	
	if err := s.writeEvent(w, event); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

