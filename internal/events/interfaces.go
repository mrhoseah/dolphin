package events

import (
	"context"
	"time"
)

// Event represents a domain event
type Event interface {
	GetName() string
	GetPayload() interface{}
	GetTimestamp() time.Time
	GetID() string
}

// Listener defines the interface for event listeners
type Listener interface {
	Handle(ctx context.Context, event Event) error
	GetPriority() int
	ShouldQueue() bool
}

// EventDispatcher manages event dispatching and listener registration
type EventDispatcher interface {
	// Register a listener for a specific event
	Listen(eventName string, listener Listener)

	// Dispatch an event to all registered listeners
	Dispatch(ctx context.Context, event Event) error

	// Dispatch an event asynchronously
	DispatchAsync(ctx context.Context, event Event) error

	// Remove a listener for a specific event
	RemoveListener(eventName string, listener Listener)

	// Get all listeners for an event
	GetListeners(eventName string) []Listener

	// Check if an event has listeners
	HasListeners(eventName string) bool

	// Clear all listeners for an event
	ClearListeners(eventName string)

	// Clear all listeners
	ClearAllListeners()
}

// EventQueue manages queued events for asynchronous processing
type EventQueue interface {
	// Push an event to the queue
	Push(ctx context.Context, event Event) error

	// Pop an event from the queue
	Pop(ctx context.Context) (Event, error)

	// Get queue size
	Size(ctx context.Context) (int, error)

	// Clear the queue
	Clear(ctx context.Context) error

	// Process events from the queue
	Process(ctx context.Context, dispatcher EventDispatcher) error
}

// EventBus combines dispatcher and queue functionality
type EventBus interface {
	EventDispatcher
	EventQueue

	// Subscribe to an event with a listener
	Subscribe(eventName string, listener Listener)

	// Publish an event (dispatch immediately)
	Publish(ctx context.Context, event Event) error

	// Publish an event to queue
	PublishAsync(ctx context.Context, event Event) error

	// Start processing queued events
	StartWorker(ctx context.Context) error

	// Stop processing queued events
	StopWorker() error
}

// EventStore persists events for audit/replay
type EventStore interface {
	// Store an event
	Store(ctx context.Context, event Event) error

	// Retrieve events by name
	GetEvents(ctx context.Context, eventName string, limit int) ([]Event, error)

	// Retrieve events by time range
	GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]Event, error)

	// Retrieve events by ID
	GetEventByID(ctx context.Context, id string) (Event, error)

	// Get event count
	GetEventCount(ctx context.Context, eventName string) (int, error)
}

// EventMiddleware defines middleware for event processing
type EventMiddleware interface {
	// Process event before dispatching
	BeforeDispatch(ctx context.Context, event Event) (Event, error)

	// Process event after dispatching
	AfterDispatch(ctx context.Context, event Event, err error) error

	// Process event on error
	OnError(ctx context.Context, event Event, err error) error
}

// EventFactory creates events
type EventFactory interface {
	// Create a new event
	Create(name string, payload interface{}) Event

	// Create a new event with custom ID
	CreateWithID(id, name string, payload interface{}) Event
}

// EventSerializer serializes/deserializes events
type EventSerializer interface {
	// Serialize an event to bytes
	Serialize(event Event) ([]byte, error)

	// Deserialize bytes to an event
	Deserialize(data []byte) (Event, error)
}
