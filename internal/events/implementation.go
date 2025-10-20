package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BaseEvent provides a base implementation for events
type BaseEvent struct {
	name               string
	payload            interface{}
	timestamp          time.Time
	id                 string
	propagationStopped bool
}

// NewBaseEvent creates a new base event
func NewBaseEvent(name string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		name:      name,
		payload:   payload,
		timestamp: time.Now(),
		id:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
	}
}

// GetName returns the event name
func (be *BaseEvent) GetName() string {
	return be.name
}

// GetPayload returns the event payload
func (be *BaseEvent) GetPayload() interface{} {
	return be.payload
}

// GetTimestamp returns the event timestamp
func (be *BaseEvent) GetTimestamp() time.Time {
	return be.timestamp
}

// GetID returns the event ID
func (be *BaseEvent) GetID() string {
	return be.id
}

// IsPropagationStopped returns whether propagation is stopped
func (be *BaseEvent) IsPropagationStopped() bool {
	return be.propagationStopped
}

// StopPropagation stops event propagation
func (be *BaseEvent) StopPropagation() {
	be.propagationStopped = true
}

// BaseListener provides a base implementation for listeners
type BaseListener struct {
	priority    int
	shouldQueue bool
	eventName   string
}

// NewBaseListener creates a new base listener
func NewBaseListener(eventName string, priority int, shouldQueue bool) *BaseListener {
	return &BaseListener{
		priority:    priority,
		shouldQueue: shouldQueue,
		eventName:   eventName,
	}
}

// GetPriority returns the listener priority
func (bl *BaseListener) GetPriority() int {
	return bl.priority
}

// ShouldQueue returns whether the listener should be queued
func (bl *BaseListener) ShouldQueue() bool {
	return bl.shouldQueue
}

// EventDispatcherImpl implements EventDispatcher
type EventDispatcherImpl struct {
	listeners map[string][]Listener
	mu        sync.RWMutex
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() EventDispatcher {
	return &EventDispatcherImpl{
		listeners: make(map[string][]Listener),
	}
}

// Listen registers a listener for a specific event
func (ed *EventDispatcherImpl) Listen(eventName string, listener Listener) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if ed.listeners[eventName] == nil {
		ed.listeners[eventName] = make([]Listener, 0)
	}

	ed.listeners[eventName] = append(ed.listeners[eventName], listener)
}

// Dispatch dispatches an event to all registered listeners
func (ed *EventDispatcherImpl) Dispatch(ctx context.Context, event Event) error {
	ed.mu.RLock()
	listeners := ed.listeners[event.GetName()]
	ed.mu.RUnlock()

	if len(listeners) == 0 {
		return nil
	}

	// Handle listeners
	for _, listener := range listeners {
		if err := listener.Handle(ctx, event); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	return nil
}

// DispatchAsync dispatches an event asynchronously
func (ed *EventDispatcherImpl) DispatchAsync(ctx context.Context, event Event) error {
	go func() {
		if err := ed.Dispatch(ctx, event); err != nil {
			fmt.Printf("Async dispatch error: %v\n", err)
		}
	}()
	return nil
}

// RemoveListener removes a listener for a specific event
func (ed *EventDispatcherImpl) RemoveListener(eventName string, listener Listener) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	listeners := ed.listeners[eventName]
	for i, l := range listeners {
		if l == listener {
			ed.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

// GetListeners returns all listeners for an event
func (ed *EventDispatcherImpl) GetListeners(eventName string) []Listener {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	listeners := make([]Listener, len(ed.listeners[eventName]))
	copy(listeners, ed.listeners[eventName])
	return listeners
}

// HasListeners checks if an event has listeners
func (ed *EventDispatcherImpl) HasListeners(eventName string) bool {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	return len(ed.listeners[eventName]) > 0
}

// ClearListeners clears all listeners for an event
func (ed *EventDispatcherImpl) ClearListeners(eventName string) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	ed.listeners[eventName] = nil
}

// ClearAllListeners clears all listeners
func (ed *EventDispatcherImpl) ClearAllListeners() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	ed.listeners = make(map[string][]Listener)
}

// Built-in Event Types

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	*BaseEvent
	User interface{}
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(user interface{}) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent: NewBaseEvent("user.created", user),
		User:      user,
	}
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	*BaseEvent
	User interface{}
}

// NewUserUpdatedEvent creates a new user updated event
func NewUserUpdatedEvent(user interface{}) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseEvent: NewBaseEvent("user.updated", user),
		User:      user,
	}
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	*BaseEvent
	UserID interface{}
}

// NewUserDeletedEvent creates a new user deleted event
func NewUserDeletedEvent(userID interface{}) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseEvent: NewBaseEvent("user.deleted", userID),
		UserID:    userID,
	}
}

// PostCreatedEvent represents a post creation event
type PostCreatedEvent struct {
	*BaseEvent
	Post interface{}
}

// NewPostCreatedEvent creates a new post created event
func NewPostCreatedEvent(post interface{}) *PostCreatedEvent {
	return &PostCreatedEvent{
		BaseEvent: NewBaseEvent("post.created", post),
		Post:      post,
	}
}

// Built-in Listeners

// EmailNotificationListener sends email notifications
type EmailNotificationListener struct {
	*BaseListener
}

// NewEmailNotificationListener creates a new email notification listener
func NewEmailNotificationListener(eventName string, priority int) *EmailNotificationListener {
	return &EmailNotificationListener{
		BaseListener: NewBaseListener(eventName, priority, true), // Queue this listener
	}
}

// Handle handles the event
func (enl *EmailNotificationListener) Handle(ctx context.Context, event Event) error {
	fmt.Printf("📧 Sending email notification for event: %s\n", event.GetName())
	// In a real implementation, you'd send actual emails
	return nil
}

// LoggingListener logs events
type LoggingListener struct {
	*BaseListener
}

// NewLoggingListener creates a new logging listener
func NewLoggingListener(eventName string, priority int) *LoggingListener {
	return &LoggingListener{
		BaseListener: NewBaseListener(eventName, priority, false), // Don't queue this listener
	}
}

// Handle handles the event
func (ll *LoggingListener) Handle(ctx context.Context, event Event) error {
	fmt.Printf("📝 Logging event: %s (ID: %s)\n", event.GetName(), event.GetID())
	return nil
}

// CacheInvalidationListener invalidates cache
type CacheInvalidationListener struct {
	*BaseListener
}

// NewCacheInvalidationListener creates a new cache invalidation listener
func NewCacheInvalidationListener(eventName string, priority int) *CacheInvalidationListener {
	return &CacheInvalidationListener{
		BaseListener: NewBaseListener(eventName, priority, false), // Don't queue this listener
	}
}

// Handle handles the event
func (cil *CacheInvalidationListener) Handle(ctx context.Context, event Event) error {
	fmt.Printf("🗑️  Invalidating cache for event: %s\n", event.GetName())
	// In a real implementation, you'd invalidate actual cache entries
	return nil
}

// AnalyticsListener tracks analytics
type AnalyticsListener struct {
	*BaseListener
}

// NewAnalyticsListener creates a new analytics listener
func NewAnalyticsListener(eventName string, priority int) *AnalyticsListener {
	return &AnalyticsListener{
		BaseListener: NewBaseListener(eventName, priority, true), // Queue this listener
	}
}

// Handle handles the event
func (al *AnalyticsListener) Handle(ctx context.Context, event Event) error {
	fmt.Printf("📊 Tracking analytics for event: %s\n", event.GetName())
	// In a real implementation, you'd send analytics data
	return nil
}

// InMemoryEventQueue provides an in-memory event queue
type InMemoryEventQueue struct {
	events []Event
	mu     sync.RWMutex
}

// NewInMemoryEventQueue creates a new in-memory event queue
func NewInMemoryEventQueue() *InMemoryEventQueue {
	return &InMemoryEventQueue{
		events: make([]Event, 0),
	}
}

// Push pushes an event to the queue
func (imeq *InMemoryEventQueue) Push(ctx context.Context, event Event) error {
	imeq.mu.Lock()
	defer imeq.mu.Unlock()

	imeq.events = append(imeq.events, event)
	return nil
}

// Pop pops an event from the queue
func (imeq *InMemoryEventQueue) Pop(ctx context.Context) (Event, error) {
	imeq.mu.Lock()
	defer imeq.mu.Unlock()

	if len(imeq.events) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	event := imeq.events[0]
	imeq.events = imeq.events[1:]
	return event, nil
}

// Size returns the queue size
func (imeq *InMemoryEventQueue) Size(ctx context.Context) (int, error) {
	imeq.mu.RLock()
	defer imeq.mu.RUnlock()

	return len(imeq.events), nil
}

// Clear clears the queue
func (imeq *InMemoryEventQueue) Clear(ctx context.Context) error {
	imeq.mu.Lock()
	defer imeq.mu.Unlock()

	imeq.events = make([]Event, 0)
	return nil
}

// Process processes events from the queue
func (imeq *InMemoryEventQueue) Process(ctx context.Context, dispatcher EventDispatcher) error {
	for {
		event, err := imeq.Pop(ctx)
		if err != nil {
			break // Queue is empty
		}

		if err := dispatcher.Dispatch(ctx, event); err != nil {
			return fmt.Errorf("failed to dispatch event: %w", err)
		}
	}

	return nil
}

// EventBusImpl implements EventBus
type EventBusImpl struct {
	dispatcher EventDispatcher
	queue      EventQueue
}

// NewEventBus creates a new event bus
func NewEventBus(dispatcher EventDispatcher, queue EventQueue) EventBus {
	return &EventBusImpl{
		dispatcher: dispatcher,
		queue:      queue,
	}
}

// Listen registers a listener for a specific event
func (eb *EventBusImpl) Listen(eventName string, listener Listener) {
	eb.dispatcher.Listen(eventName, listener)
}

// Dispatch dispatches an event to all registered listeners
func (eb *EventBusImpl) Dispatch(ctx context.Context, event Event) error {
	return eb.dispatcher.Dispatch(ctx, event)
}

// DispatchAsync dispatches an event asynchronously
func (eb *EventBusImpl) DispatchAsync(ctx context.Context, event Event) error {
	return eb.dispatcher.DispatchAsync(ctx, event)
}

// Push pushes an event to the queue
func (eb *EventBusImpl) Push(ctx context.Context, event Event) error {
	return eb.queue.Push(ctx, event)
}

// Pop pops an event from the queue
func (eb *EventBusImpl) Pop(ctx context.Context) (Event, error) {
	return eb.queue.Pop(ctx)
}

// Process processes events from the queue
func (eb *EventBusImpl) Process(ctx context.Context, dispatcher EventDispatcher) error {
	return eb.queue.Process(ctx, dispatcher)
}

// Subscribe subscribes to an event with a listener
func (eb *EventBusImpl) Subscribe(eventName string, listener Listener) {
	eb.dispatcher.Listen(eventName, listener)
}

// Publish publishes an event (dispatch immediately)
func (eb *EventBusImpl) Publish(ctx context.Context, event Event) error {
	return eb.dispatcher.Dispatch(ctx, event)
}

// PublishAsync publishes an event to queue
func (eb *EventBusImpl) PublishAsync(ctx context.Context, event Event) error {
	return eb.queue.Push(ctx, event)
}

// StartWorker starts processing queued events
func (eb *EventBusImpl) StartWorker(ctx context.Context) error {
	return eb.queue.Process(ctx, eb.dispatcher)
}

// StopWorker stops processing queued events
func (eb *EventBusImpl) StopWorker() error {
	return nil // Simplified implementation
}

// Size returns the queue size
func (eb *EventBusImpl) Size(ctx context.Context) (int, error) {
	return eb.queue.Size(ctx)
}

// Clear clears the queue
func (eb *EventBusImpl) Clear(ctx context.Context) error {
	return eb.queue.Clear(ctx)
}

// RemoveListener removes a listener for a specific event
func (eb *EventBusImpl) RemoveListener(eventName string, listener Listener) {
	eb.dispatcher.RemoveListener(eventName, listener)
}

// GetListeners returns all listeners for an event
func (eb *EventBusImpl) GetListeners(eventName string) []Listener {
	return eb.dispatcher.GetListeners(eventName)
}

// HasListeners checks if an event has listeners
func (eb *EventBusImpl) HasListeners(eventName string) bool {
	return eb.dispatcher.HasListeners(eventName)
}

// ClearListeners clears all listeners for an event
func (eb *EventBusImpl) ClearListeners(eventName string) {
	eb.dispatcher.ClearListeners(eventName)
}

// ClearAllListeners clears all listeners
func (eb *EventBusImpl) ClearAllListeners() {
	eb.dispatcher.ClearAllListeners()
}

// EventManager provides a high-level event management interface
type EventManager struct {
	bus        EventBus
	dispatcher EventDispatcher
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	dispatcher := NewEventDispatcher()
	queue := NewInMemoryEventQueue()
	bus := NewEventBus(dispatcher, queue)

	return &EventManager{
		bus:        bus,
		dispatcher: dispatcher,
	}
}

// RegisterListener registers a listener
func (em *EventManager) RegisterListener(eventName string, listener Listener) {
	em.bus.Listen(eventName, listener)
}

// Dispatch dispatches an event
func (em *EventManager) Dispatch(ctx context.Context, event Event) error {
	return em.bus.Dispatch(ctx, event)
}

// DispatchAsync dispatches an event asynchronously
func (em *EventManager) DispatchAsync(ctx context.Context, event Event) error {
	return em.bus.DispatchAsync(ctx, event)
}

// Queue queues an event for later processing
func (em *EventManager) Queue(ctx context.Context, event Event) error {
	return em.bus.Push(ctx, event)
}

// ProcessQueue processes queued events
func (em *EventManager) ProcessQueue(ctx context.Context) error {
	return em.bus.Process(ctx, em.dispatcher)
}
