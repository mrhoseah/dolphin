package router

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// WebSocketConfig configures WebSocket connection
type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool
	Logger          *zap.Logger
}

// DefaultWebSocketConfig returns default WebSocket configuration
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler func(*websocket.Conn, *http.Request) error

// HandleWebSocket upgrades HTTP connection to WebSocket and handles it
func HandleWebSocket(w http.ResponseWriter, r *http.Request, handler WebSocketHandler, config *WebSocketConfig) error {
	if config == nil {
		config = DefaultWebSocketConfig()
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     config.CheckOrigin,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	if config.Logger != nil {
		config.Logger.Info("WebSocket connection established",
			zap.String("remote_addr", r.RemoteAddr),
		)
	}

	return handler(conn, r)
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Error   string      `json:"error,omitempty"`
}

// SendMessage sends a JSON message through WebSocket
func SendMessage(conn *websocket.Conn, messageType string, payload interface{}) error {
	msg := WebSocketMessage{
		Type:    messageType,
		Payload: payload,
	}
	return conn.WriteJSON(msg)
}

// SendError sends an error message through WebSocket
func SendError(conn *websocket.Conn, err error) error {
	msg := WebSocketMessage{
		Type:  "error",
		Error: err.Error(),
	}
	return conn.WriteJSON(msg)
}

// ReadMessage reads a JSON message from WebSocket
func ReadMessage(conn *websocket.Conn) (*WebSocketMessage, error) {
	var msg WebSocketMessage
	if err := conn.ReadJSON(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

