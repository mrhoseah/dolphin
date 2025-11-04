package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Message represents a real-time message
type Message struct {
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel,omitempty"`
	Event     string                 `json:"event,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	Conn     *websocket.Conn
	Channels map[string]bool
	Send     chan Message
	Hub      *Hub
	mu       sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[string]*Client
	channels   map[string]map[string]*Client // channel -> clientID -> client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *zap.Logger
	upgrader   websocket.Upgrader
}

// NewHub creates a new Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		channels:   make(map[string]map[string]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure based on your needs
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
	h.logger.Info("Client registered", zap.String("client_id", client.ID))
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)

		// Remove from all channels
		for channel := range client.Channels {
			if channelClients, exists := h.channels[channel]; exists {
				delete(channelClients, client.ID)
				if len(channelClients) == 0 {
					delete(h.channels, channel)
				}
			}
		}

		h.logger.Info("Client unregistered", zap.String("client_id", client.ID))
	}
}

// broadcastMessage broadcasts a message to clients
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if message.Channel != "" {
		// Send to specific channel
		if channelClients, exists := h.channels[message.Channel]; exists {
			for _, client := range channelClients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
		}
	} else {
		// Broadcast to all clients
		for _, client := range h.clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

// Subscribe subscribes a client to a channel
func (h *Hub) Subscribe(clientID, channel string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	client.mu.Lock()
	client.Channels[channel] = true
	client.mu.Unlock()

	if h.channels[channel] == nil {
		h.channels[channel] = make(map[string]*Client)
	}
	h.channels[channel][clientID] = client

	h.logger.Info("Client subscribed to channel",
		zap.String("client_id", clientID),
		zap.String("channel", channel))

	return nil
}

// Unsubscribe unsubscribes a client from a channel
func (h *Hub) Unsubscribe(clientID, channel string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, exists := h.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	client.mu.Lock()
	delete(client.Channels, channel)
	client.mu.Unlock()

	if channelClients, exists := h.channels[channel]; exists {
		delete(channelClients, clientID)
		if len(channelClients) == 0 {
			delete(h.channels, channel)
		}
	}

	return nil
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(message Message) {
	message.Timestamp = time.Now()
	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("Broadcast channel full, dropping message")
	}
}

// BroadcastToChannel sends a message to a specific channel
func (h *Hub) BroadcastToChannel(channel string, event string, data map[string]interface{}) {
	h.Broadcast(Message{
		Type:    "event",
		Channel: channel,
		Event:   event,
		Data:    data,
	})
}

// HandleWebSocket handles WebSocket connections
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) (*Client, error) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	clientID := generateClientID()
	client := &Client{
		ID:       clientID,
		Conn:     conn,
		Channels: make(map[string]bool),
		Send:     make(chan Message, 256),
		Hub:      h,
	}

	h.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	return client, nil
}

// readPump pumps messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			c.Hub.logger.Warn("Invalid message format", zap.Error(err))
			continue
		}

		c.handleMessage(message)
	}
}

// writePump pumps messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			messageBytes, _ := json.Marshal(message)
			w.Write(messageBytes)

			// Write queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				msg := <-c.Send
				msgBytes, _ := json.Marshal(msg)
				w.Write(msgBytes)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages from client
func (c *Client) handleMessage(message Message) {
	switch message.Type {
	case "subscribe":
		if channel, ok := message.Data["channel"].(string); ok {
			c.Hub.Subscribe(c.ID, channel)
			c.Send <- Message{
				Type:    "subscribed",
				Channel: channel,
			}
		}
	case "unsubscribe":
		if channel, ok := message.Data["channel"].(string); ok {
			c.Hub.Unsubscribe(c.ID, channel)
			c.Send <- Message{
				Type:    "unsubscribed",
				Channel: channel,
			}
		}
	case "ping":
		c.Send <- Message{
			Type: "pong",
		}
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

// SSEServer handles Server-Sent Events
type SSEServer struct {
	clients map[string]chan []byte
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewSSEServer creates a new SSE server
func NewSSEServer(logger *zap.Logger) *SSEServer {
	return &SSEServer{
		clients: make(map[string]chan []byte),
		logger:  logger,
	}
}

// HandleSSE handles SSE connections
func (s *SSEServer) HandleSSE(w http.ResponseWriter, r *http.Request) (string, error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientID := generateClientID()
	clientChan := make(chan []byte, 256)

	s.mu.Lock()
	s.clients[clientID] = clientChan
	s.mu.Unlock()

	// Send initial connection message
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"client_id\":\"%s\"}\n\n", clientID)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Handle client disconnect
	ctx := r.Context()
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.clients, clientID)
		close(clientChan)
		s.mu.Unlock()
	}()

	// Stream messages
	for {
		select {
		case <-ctx.Done():
			return clientID, nil
		case message := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", message)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// Broadcast sends a message to all SSE clients
func (s *SSEServer) Broadcast(message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, clientChan := range s.clients {
		select {
		case clientChan <- message:
		default:
			// Channel full, skip
		}
	}
}

// SendToClient sends a message to a specific client
func (s *SSEServer) SendToClient(clientID string, message []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clientChan, exists := s.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	select {
	case clientChan <- message:
		return nil
	default:
		return fmt.Errorf("client channel full")
	}
}

