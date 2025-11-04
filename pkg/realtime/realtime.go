package realtime

import (
	"dolphin/internal/realtime"
	"go.uber.org/zap"
)

// Hub represents a real-time hub
type Hub = realtime.Hub

// NewHub creates a new real-time hub
func NewHub(logger *zap.Logger) *Hub {
	return realtime.NewHub(logger)
}

// SSEServer represents an SSE server
type SSEServer = realtime.SSEServer

// NewSSEServer creates a new SSE server
func NewSSEServer(logger *zap.Logger) *SSEServer {
	return realtime.NewSSEServer(logger)
}

// Message represents a real-time message
type Message = realtime.Message

// Client represents a WebSocket client
type Client = realtime.Client

