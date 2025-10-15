package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BaseEvent provides a default implementation of Event interface
type BaseEvent struct {
	ID        string
	Name      string
	Payload   interface{}
	Timestamp time.Time
}

// NewBaseEvent creates a new base event
func NewBaseEvent(name string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		ID:        uuid.New().String(),
		Name:      name,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// NewBaseEventWithID creates a new base event with custom ID
func NewBaseEventWithID(id, name string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		ID:        id,
		Name:      name,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

func (e *BaseEvent) GetName() string {
	return e.Name
}

func (e *BaseEvent) GetPayload() interface{} {
	return e.Payload
}

func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *BaseEvent) GetID() string {
	return e.ID
}

// EventDispatcher implementation
type eventDispatcher struct {
	listeners map[string][]Listener
	mutex     sync.RWMutex
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() EventDispatcher {
	return &eventDispatcher{
		listeners: make(map[string][]Listener),
	}
}

func (d *eventDispatcher) Listen(eventName string, listener Listener) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.listeners[eventName] = append(d.listeners[eventName], listener)

	// Sort listeners by priority (higher priority first)
	d.sortListeners(eventName)
}

func (d *eventDispatcher) Dispatch(ctx context.Context, event Event) error {
	d.mutex.RLock()
	listeners := d.listeners[event.GetName()]
	d.mutex.RUnlock()

	if len(listeners) == 0 {
		return nil
	}

	var errors []error

	for _, listener := range listeners {
		if err := listener.Handle(ctx, event); err != nil {
			errors = append(errors, fmt.Errorf("listener error for event %s: %w", event.GetName(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple listener errors: %v", errors)
	}

	return nil
}

func (d *eventDispatcher) DispatchAsync(ctx context.Context, event Event) error {
	go func() {
		if err := d.Dispatch(ctx, event); err != nil {
			// Log error or handle as needed
			fmt.Printf("Async dispatch error: %v\n", err)
		}
	}()
	return nil
}

func (d *eventDispatcher) RemoveListener(eventName string, listener Listener) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	listeners := d.listeners[eventName]
	for i, l := range listeners {
		if l == listener {
			d.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

func (d *eventDispatcher) GetListeners(eventName string) []Listener {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	listeners := d.listeners[eventName]
	result := make([]Listener, len(listeners))
	copy(result, listeners)
	return result
}

func (d *eventDispatcher) HasListeners(eventName string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return len(d.listeners[eventName]) > 0
}

func (d *eventDispatcher) ClearListeners(eventName string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.listeners, eventName)
}

func (d *eventDispatcher) ClearAllListeners() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.listeners = make(map[string][]Listener)
}

// sortListeners sorts listeners by priority (higher priority first)
func (d *eventDispatcher) sortListeners(eventName string) {
	listeners := d.listeners[eventName]

	// Simple bubble sort by priority
	for i := 0; i < len(listeners)-1; i++ {
		for j := 0; j < len(listeners)-i-1; j++ {
			if listeners[j].GetPriority() < listeners[j+1].GetPriority() {
				listeners[j], listeners[j+1] = listeners[j+1], listeners[j]
			}
		}
	}
}

// EventQueue implementation using in-memory queue
type eventQueue struct {
	queue []Event
	mutex sync.Mutex
}

// NewEventQueue creates a new event queue
func NewEventQueue() EventQueue {
	return &eventQueue{
		queue: make([]Event, 0),
	}
}

func (q *eventQueue) Push(ctx context.Context, event Event) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.queue = append(q.queue, event)
	return nil
}

func (q *eventQueue) Pop(ctx context.Context) (Event, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.queue) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	event := q.queue[0]
	q.queue = q.queue[1:]
	return event, nil
}

func (q *eventQueue) Size(ctx context.Context) (int, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.queue), nil
}

func (q *eventQueue) Clear(ctx context.Context) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.queue = make([]Event, 0)
	return nil
}

func (q *eventQueue) Process(ctx context.Context, dispatcher EventDispatcher) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			event, err := q.Pop(ctx)
			if err != nil {
				// Queue is empty, wait a bit
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if err := dispatcher.Dispatch(ctx, event); err != nil {
				// Log error or handle as needed
				fmt.Printf("Event processing error: %v\n", err)
			}
		}
	}
}

// EventBus implementation
type eventBus struct {
	EventDispatcher
	EventQueue
	workerCtx    context.Context
	workerCancel context.CancelFunc
	workerWg     sync.WaitGroup
}

// NewEventBus creates a new event bus
func NewEventBus() EventBus {
	return &eventBus{
		EventDispatcher: NewEventDispatcher(),
		EventQueue:      NewEventQueue(),
	}
}

func (b *eventBus) Subscribe(eventName string, listener Listener) {
	b.Listen(eventName, listener)
}

func (b *eventBus) Publish(ctx context.Context, event Event) error {
	return b.Dispatch(ctx, event)
}

func (b *eventBus) PublishAsync(ctx context.Context, event Event) error {
	return b.Push(ctx, event)
}

func (b *eventBus) StartWorker(ctx context.Context) error {
	b.workerCtx, b.workerCancel = context.WithCancel(ctx)

	b.workerWg.Add(1)
	go func() {
		defer b.workerWg.Done()
		b.Process(b.workerCtx, b.EventDispatcher)
	}()

	return nil
}

func (b *eventBus) StopWorker() error {
	if b.workerCancel != nil {
		b.workerCancel()
		b.workerWg.Wait()
	}
	return nil
}

// EventFactory implementation
type eventFactory struct{}

// NewEventFactory creates a new event factory
func NewEventFactory() EventFactory {
	return &eventFactory{}
}

func (f *eventFactory) Create(name string, payload interface{}) Event {
	return NewBaseEvent(name, payload)
}

func (f *eventFactory) CreateWithID(id, name string, payload interface{}) Event {
	return NewBaseEventWithID(id, name, payload)
}

// EventSerializer implementation using JSON
type eventSerializer struct{}

// NewEventSerializer creates a new event serializer
func NewEventSerializer() EventSerializer {
	return &eventSerializer{}
}

func (s *eventSerializer) Serialize(event Event) ([]byte, error) {
	// This would use JSON marshaling in a real implementation
	return []byte(fmt.Sprintf("Event: %s, ID: %s", event.GetName(), event.GetID())), nil
}

func (s *eventSerializer) Deserialize(data []byte) (Event, error) {
	// This would use JSON unmarshaling in a real implementation
	return NewBaseEvent("deserialized", string(data)), nil
}
