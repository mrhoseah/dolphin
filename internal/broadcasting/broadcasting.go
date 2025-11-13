package broadcasting

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Channel represents a broadcasting channel
type Channel interface {
	Name() string
	Subscribe(ctx context.Context, subscriber Subscriber) error
	Unsubscribe(ctx context.Context, subscriberID string) error
	Publish(ctx context.Context, event Event) error
	GetSubscribers() []Subscriber
}

// Subscriber represents a channel subscriber
type Subscriber interface {
	ID() string
	Send(ctx context.Context, event Event) error
	Close() error
}

// Event represents a broadcast event
type Event struct {
	Channel string                 `json:"channel"`
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
}

// PrivateChannel represents a private channel (requires authentication)
type PrivateChannel interface {
	Channel
	Authorize(ctx context.Context, subscriberID string, channelName string) (bool, error)
}

// PresenceChannel represents a presence channel (tracks who's present)
type PresenceChannel interface {
	Channel
	GetPresence(ctx context.Context) (map[string]PresenceInfo, error)
}

// PresenceInfo represents presence information
type PresenceInfo struct {
	UserID   string                 `json:"user_id"`
	UserInfo map[string]interface{} `json:"user_info"`
}

// Broadcaster manages broadcasting channels
type Broadcaster struct {
	channels    map[string]Channel
	subscribers map[string]map[string]Subscriber // channel -> subscriberID -> subscriber
	mu          sync.RWMutex
	logger      *zap.Logger
}

// NewBroadcaster creates a new broadcaster
func NewBroadcaster(logger *zap.Logger) *Broadcaster {
	return &Broadcaster{
		channels:    make(map[string]Channel),
		subscribers: make(map[string]map[string]Subscriber),
		logger:      logger,
	}
}

// RegisterChannel registers a channel
func (b *Broadcaster) RegisterChannel(channel Channel) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.channels[channel.Name()] = channel
	b.subscribers[channel.Name()] = make(map[string]Subscriber)
}

// GetChannel returns a channel by name
func (b *Broadcaster) GetChannel(name string) (Channel, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	channel, exists := b.channels[name]
	return channel, exists
}

// Subscribe subscribes to a channel
func (b *Broadcaster) Subscribe(ctx context.Context, channelName string, subscriber Subscriber) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	channel, exists := b.channels[channelName]
	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	// Check authorization for private channels
	if privateChannel, ok := channel.(PrivateChannel); ok {
		authorized, err := privateChannel.Authorize(ctx, subscriber.ID(), channelName)
		if err != nil {
			return err
		}
		if !authorized {
			return fmt.Errorf("unauthorized to subscribe to channel %s", channelName)
		}
	}

	if err := channel.Subscribe(ctx, subscriber); err != nil {
		return err
	}

	b.subscribers[channelName][subscriber.ID()] = subscriber

	if b.logger != nil {
		b.logger.Info("Subscriber joined channel",
			zap.String("channel", channelName),
			zap.String("subscriber_id", subscriber.ID()),
		)
	}

	return nil
}

// Unsubscribe unsubscribes from a channel
func (b *Broadcaster) Unsubscribe(ctx context.Context, channelName, subscriberID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	channel, exists := b.channels[channelName]
	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	subscriber, exists := b.subscribers[channelName][subscriberID]
	if !exists {
		return fmt.Errorf("subscriber %s not found in channel %s", subscriberID, channelName)
	}

	if err := channel.Unsubscribe(ctx, subscriberID); err != nil {
		return err
	}

	delete(b.subscribers[channelName], subscriberID)
	subscriber.Close()

	if b.logger != nil {
		b.logger.Info("Subscriber left channel",
			zap.String("channel", channelName),
			zap.String("subscriber_id", subscriberID),
		)
	}

	return nil
}

// Publish publishes an event to a channel
func (b *Broadcaster) Publish(ctx context.Context, channelName string, event Event) error {
	b.mu.RLock()
	channel, exists := b.channels[channelName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	event.Channel = channelName
	return channel.Publish(ctx, event)
}

// Broadcast broadcasts an event to all subscribers of a channel
func (b *Broadcaster) Broadcast(ctx context.Context, channelName, eventName string, data map[string]interface{}) error {
	event := Event{
		Channel: channelName,
		Event:   eventName,
		Data:    data,
	}

	return b.Publish(ctx, channelName, event)
}

// MemoryChannel implements Channel using in-memory storage
type MemoryChannel struct {
	name        string
	subscribers map[string]Subscriber
	mu          sync.RWMutex
}

// NewMemoryChannel creates a new memory channel
func NewMemoryChannel(name string) *MemoryChannel {
	return &MemoryChannel{
		name:        name,
		subscribers: make(map[string]Subscriber),
	}
}

// Name returns the channel name
func (mc *MemoryChannel) Name() string {
	return mc.name
}

// Subscribe subscribes to the channel
func (mc *MemoryChannel) Subscribe(ctx context.Context, subscriber Subscriber) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.subscribers[subscriber.ID()] = subscriber
	return nil
}

// Unsubscribe unsubscribes from the channel
func (mc *MemoryChannel) Unsubscribe(ctx context.Context, subscriberID string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	delete(mc.subscribers, subscriberID)
	return nil
}

// Publish publishes an event to all subscribers
func (mc *MemoryChannel) Publish(ctx context.Context, event Event) error {
	mc.mu.RLock()
	subscribers := make([]Subscriber, 0, len(mc.subscribers))
	for _, sub := range mc.subscribers {
		subscribers = append(subscribers, sub)
	}
	mc.mu.RUnlock()

	// Send event to all subscribers
	for _, subscriber := range subscribers {
		go func(sub Subscriber) {
			if err := sub.Send(ctx, event); err != nil {
				// Log error but don't fail the publish
			}
		}(subscriber)
	}

	return nil
}

// GetSubscribers returns all subscribers
func (mc *MemoryChannel) GetSubscribers() []Subscriber {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	subscribers := make([]Subscriber, 0, len(mc.subscribers))
	for _, sub := range mc.subscribers {
		subscribers = append(subscribers, sub)
	}
	return subscribers
}

// WebSocketSubscriber implements Subscriber using WebSocket
type WebSocketSubscriber struct {
	id       string
	sendChan chan Event
	closeChan chan struct{}
}

// NewWebSocketSubscriber creates a new WebSocket subscriber
func NewWebSocketSubscriber(id string) *WebSocketSubscriber {
	return &WebSocketSubscriber{
		id:        id,
		sendChan:  make(chan Event, 100),
		closeChan: make(chan struct{}),
	}
}

// ID returns the subscriber ID
func (ws *WebSocketSubscriber) ID() string {
	return ws.id
}

// Send sends an event to the subscriber
func (ws *WebSocketSubscriber) Send(ctx context.Context, event Event) error {
	select {
	case ws.sendChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-ws.closeChan:
		return fmt.Errorf("subscriber closed")
	}
}

// Close closes the subscriber
func (ws *WebSocketSubscriber) Close() error {
	close(ws.closeChan)
	close(ws.sendChan)
	return nil
}

// GetSendChan returns the send channel
func (ws *WebSocketSubscriber) GetSendChan() <-chan Event {
	return ws.sendChan
}

// EventSerializer serializes events
type EventSerializer interface {
	Serialize(event Event) ([]byte, error)
	Deserialize(data []byte) (Event, error)
}

// JSONEventSerializer implements EventSerializer using JSON
type JSONEventSerializer struct{}

// Serialize serializes an event to JSON
func (j *JSONEventSerializer) Serialize(event Event) ([]byte, error) {
	return json.Marshal(event)
}

// Deserialize deserializes JSON to an event
func (j *JSONEventSerializer) Deserialize(data []byte) (Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	return event, err
}

