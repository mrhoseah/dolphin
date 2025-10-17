package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	"go.uber.org/zap"
)

// SubscriptionManager manages GraphQL subscriptions
type SubscriptionManager struct {
	clients  map[string]*SubscriptionClient
	topics   map[string][]string // topic -> client IDs
	mu       sync.RWMutex
	logger   *zap.Logger
	upgrader websocket.Upgrader
}

// SubscriptionClient represents a WebSocket client
type SubscriptionClient struct {
	ID      string
	Conn    *websocket.Conn
	Topics  map[string]bool
	Send    chan []byte
	Done    chan bool
	Context context.Context
	Cancel  context.CancelFunc
}

// SubscriptionMessage represents a subscription message
type SubscriptionMessage struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// SubscriptionEvent represents a subscription event
type SubscriptionEvent struct {
	Topic   string
	Payload interface{}
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(logger *zap.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		clients: make(map[string]*SubscriptionClient),
		topics:  make(map[string][]string),
		logger:  logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, implement proper origin checking
			},
		},
	}
}

// AddClient adds a new subscription client
func (sm *SubscriptionManager) AddClient(conn *websocket.Conn) *SubscriptionClient {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())

	client := &SubscriptionClient{
		ID:      clientID,
		Conn:    conn,
		Topics:  make(map[string]bool),
		Send:    make(chan []byte, 256),
		Done:    make(chan bool),
		Context: ctx,
		Cancel:  cancel,
	}

	sm.clients[clientID] = client

	// Start goroutines for the client
	go sm.writePump(client)
	go sm.readPump(client)

	sm.logger.Info("Subscription client added", zap.String("client_id", clientID))
	return client
}

// RemoveClient removes a subscription client
func (sm *SubscriptionManager) RemoveClient(clientID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	client, exists := sm.clients[clientID]
	if !exists {
		return
	}

	// Unsubscribe from all topics
	for topic := range client.Topics {
		sm.unsubscribeFromTopic(clientID, topic)
	}

	// Close the client
	client.Cancel()
	close(client.Send)
	client.Conn.Close()

	delete(sm.clients, clientID)
	sm.logger.Info("Subscription client removed", zap.String("client_id", clientID))
}

// Subscribe subscribes a client to a topic
func (sm *SubscriptionManager) Subscribe(clientID, topic string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	client, exists := sm.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Add to client topics
	client.Topics[topic] = true

	// Add to topic clients
	sm.topics[topic] = append(sm.topics[topic], clientID)

	sm.logger.Info("Client subscribed to topic",
		zap.String("client_id", clientID),
		zap.String("topic", topic),
	)

	return nil
}

// Unsubscribe unsubscribes a client from a topic
func (sm *SubscriptionManager) Unsubscribe(clientID, topic string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.unsubscribeFromTopic(clientID, topic)
}

// unsubscribeFromTopic unsubscribes a client from a topic (internal method)
func (sm *SubscriptionManager) unsubscribeFromTopic(clientID, topic string) error {
	client, exists := sm.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Remove from client topics
	delete(client.Topics, topic)

	// Remove from topic clients
	if clients, exists := sm.topics[topic]; exists {
		for i, id := range clients {
			if id == clientID {
				sm.topics[topic] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}

	sm.logger.Info("Client unsubscribed from topic",
		zap.String("client_id", clientID),
		zap.String("topic", topic),
	)

	return nil
}

// Publish publishes an event to all subscribers of a topic
func (sm *SubscriptionManager) Publish(topic string, payload interface{}) error {
	sm.mu.RLock()
	clients, exists := sm.topics[topic]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("topic not found: %s", topic)
	}

	event := SubscriptionEvent{
		Topic:   topic,
		Payload: payload,
	}

	message := SubscriptionMessage{
		Type: "data",
		Payload: map[string]interface{}{
			"data": event,
		},
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send to all subscribers
	sm.mu.RLock()
	for _, clientID := range clients {
		if client, exists := sm.clients[clientID]; exists {
			select {
			case client.Send <- messageData:
			default:
				sm.logger.Warn("Failed to send message to client",
					zap.String("client_id", clientID),
				)
			}
		}
	}
	sm.mu.RUnlock()

	sm.logger.Info("Event published",
		zap.String("topic", topic),
		zap.Int("subscribers", len(clients)),
	)

	return nil
}

// writePump pumps messages from the send channel to the websocket connection
func (sm *SubscriptionManager) writePump(client *SubscriptionClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-client.Done:
			return
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (sm *SubscriptionManager) readPump(client *SubscriptionClient) {
	defer func() {
		sm.RemoveClient(client.ID)
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				sm.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		sm.handleMessage(client, message)
	}
}

// handleMessage handles incoming WebSocket messages
func (sm *SubscriptionManager) handleMessage(client *SubscriptionClient, message []byte) {
	var msg SubscriptionMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		sm.logger.Error("Failed to unmarshal message", zap.Error(err))
		return
	}

	switch msg.Type {
	case "start":
		sm.handleStart(client, msg)
	case "stop":
		sm.handleStop(client, msg)
	case "ping":
		sm.handlePing(client, msg)
	default:
		sm.logger.Warn("Unknown message type", zap.String("type", msg.Type))
	}
}

// handleStart handles subscription start messages
func (sm *SubscriptionManager) handleStart(client *SubscriptionClient, msg SubscriptionMessage) {
	if topic, ok := msg.Payload["topic"].(string); ok {
		if err := sm.Subscribe(client.ID, topic); err != nil {
			sm.logger.Error("Failed to subscribe", zap.Error(err))
		}
	}
}

// handleStop handles subscription stop messages
func (sm *SubscriptionManager) handleStop(client *SubscriptionClient, msg SubscriptionMessage) {
	if topic, ok := msg.Payload["topic"].(string); ok {
		if err := sm.Unsubscribe(client.ID, topic); err != nil {
			sm.logger.Error("Failed to unsubscribe", zap.Error(err))
		}
	}
}

// handlePing handles ping messages
func (sm *SubscriptionManager) handlePing(client *SubscriptionClient, msg SubscriptionMessage) {
	pong := SubscriptionMessage{
		Type: "pong",
		ID:   msg.ID,
	}

	data, err := json.Marshal(pong)
	if err != nil {
		sm.logger.Error("Failed to marshal pong", zap.Error(err))
		return
	}

	select {
	case client.Send <- data:
	default:
		sm.logger.Warn("Failed to send pong", zap.String("client_id", client.ID))
	}
}

// GetStats returns subscription statistics
func (sm *SubscriptionManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"clients": len(sm.clients),
		"topics":  len(sm.topics),
	}
}

// CreateSubscriptionType creates a GraphQL subscription type
func CreateSubscriptionType(subscriptionManager *SubscriptionManager) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Subscription",
		Fields: graphql.Fields{
			"userUpdated": &graphql.Field{
				Type: graphql.String, // In a real implementation, this would be a User type
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// This would be implemented to handle real-time updates
					return "User updated", nil
				},
			},
		},
	})
}
