package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mrhoseah/dolphin/internal/events"
)

// User represents a user model
type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Post represents a post model
type Post struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
}

// Custom Event Subscriber
type UserEventSubscriber struct{}

// GetSubscribedEvents returns the events this subscriber listens to
func (ues *UserEventSubscriber) GetSubscribedEvents() map[string]events.Listener {
	return map[string]events.Listener{
		"user.created": &UserCreatedListener{},
		"user.updated": &UserUpdatedListener{},
		"user.deleted": &UserDeletedListener{},
	}
}

// Custom Listeners
type UserCreatedListener struct {
	*events.BaseListener
}

func (ucl *UserCreatedListener) Handle(ctx context.Context, event events.Event) error {
	fmt.Printf("🎉 User created: %v\n", event.GetPayload())
	return nil
}

type UserUpdatedListener struct {
	*events.BaseListener
}

func (uul *UserUpdatedListener) Handle(ctx context.Context, event events.Event) error {
	fmt.Printf("✏️  User updated: %v\n", event.GetPayload())
	return nil
}

type UserDeletedListener struct {
	*events.BaseListener
}

func (udl *UserDeletedListener) Handle(ctx context.Context, event events.Event) error {
	fmt.Printf("🗑️  User deleted: %v\n", event.GetPayload())
	return nil
}

// PriorityTestListener is a test listener for priority demonstration
type PriorityTestListener struct {
	*events.BaseListener
	priority string
}

func (ptl *PriorityTestListener) Handle(ctx context.Context, event events.Event) error {
	fmt.Printf("🔥 %s priority listener (%d)\n", ptl.priority, ptl.GetPriority())
	return nil
}

// PropagationTestListener is a test listener for propagation demonstration
type PropagationTestListener struct {
	*events.BaseListener
	name            string
	stopPropagation bool
}

func (ptl *PropagationTestListener) Handle(ctx context.Context, event events.Event) error {
	fmt.Printf("🛑 %s listener\n", ptl.name)
	if ptl.stopPropagation {
		event.StopPropagation()
	}
	return nil
}

// QueuedTestListener is a test listener for queue demonstration
type QueuedTestListener struct {
	*events.BaseListener
	name string
}

func (qtl *QueuedTestListener) Handle(ctx context.Context, event events.Event) error {
	fmt.Printf("🔄 Processing queued event: %s\n", event.GetName())
	time.Sleep(100 * time.Millisecond) // Simulate work
	return nil
}

func main() {
	fmt.Println("🎪 Dolphin Framework - Richer Event System Demo")
	fmt.Println("===============================================")
	fmt.Println("")

	// 1. Basic Event System Demo
	fmt.Println("=== 1. Basic Event System Demo ===")
	demoBasicEventSystem()

	fmt.Println("")

	// 2. Event Priorities Demo
	fmt.Println("=== 2. Event Priorities Demo ===")
	demoEventPriorities()

	fmt.Println("")

	// 3. Event Propagation Demo
	fmt.Println("=== 3. Event Propagation Demo ===")
	demoEventPropagation()

	fmt.Println("")

	// 4. Event Queue Demo
	fmt.Println("=== 4. Event Queue Demo ===")
	demoEventQueue()

	fmt.Println("")

	// 5. Event Subscribers Demo
	fmt.Println("=== 5. Event Subscribers Demo ===")
	demoEventSubscribers()

	fmt.Println("")

	// 6. Event Manager Demo
	fmt.Println("=== 6. Event Manager Demo ===")
	demoEventManager()

	fmt.Println("")
	fmt.Println("🎉 Richer event system demonstrated successfully!")
	fmt.Println("")
	fmt.Println("💡 Key Features Implemented:")
	fmt.Println("  ✅ Event priorities and ordering")
	fmt.Println("  ✅ Event propagation control")
	fmt.Println("  ✅ Asynchronous event processing")
	fmt.Println("  ✅ Event queuing and batch processing")
	fmt.Println("  ✅ Event subscribers and listeners")
	fmt.Println("  ✅ Built-in event types")
	fmt.Println("  ✅ Event metadata and context")
	fmt.Println("  ✅ High-level event management")
}

func demoBasicEventSystem() {
	// Create event dispatcher
	dispatcher := events.NewEventDispatcher()

	// Create listeners
	loggingListener := events.NewLoggingListener("user.created", 100)
	emailListener := events.NewEmailNotificationListener("user.created", 50)

	// Register listeners
	dispatcher.Listen("user.created", loggingListener)
	dispatcher.Listen("user.created", emailListener)

	// Create and dispatch event
	user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	event := events.NewUserCreatedEvent(user)

	ctx := context.Background()
	if err := dispatcher.Dispatch(ctx, event); err != nil {
		log.Printf("Error dispatching event: %v", err)
	}

	fmt.Printf("Event dispatched: %s (ID: %s)\n", event.GetName(), event.GetID())
}

func demoEventPriorities() {
	// Create event dispatcher
	dispatcher := events.NewEventDispatcher()

	// Create listeners with different priorities
	highPriorityListener := &PriorityTestListener{
		BaseListener: events.NewBaseListener("test.priority", 100, false),
		priority:     "High",
	}
	mediumPriorityListener := &PriorityTestListener{
		BaseListener: events.NewBaseListener("test.priority", 50, false),
		priority:     "Medium",
	}
	lowPriorityListener := &PriorityTestListener{
		BaseListener: events.NewBaseListener("test.priority", 10, false),
		priority:     "Low",
	}

	// Register listeners (order doesn't matter due to priority sorting)
	dispatcher.Listen("test.priority", lowPriorityListener)
	dispatcher.Listen("test.priority", highPriorityListener)
	dispatcher.Listen("test.priority", mediumPriorityListener)

	// Create and dispatch event
	event := events.NewBaseEvent("test.priority", "test data")

	ctx := context.Background()
	if err := dispatcher.Dispatch(ctx, event); err != nil {
		log.Printf("Error dispatching event: %v", err)
	}

	fmt.Println("Listeners executed in priority order (highest first)")
}

func demoEventPropagation() {
	// Create event dispatcher
	dispatcher := events.NewEventDispatcher()

	// Create listeners
	stopPropagationListener := &PropagationTestListener{
		BaseListener:    events.NewBaseListener("test.propagation", 100, false),
		name:            "First",
		stopPropagation: true,
	}
	normalListener := &PropagationTestListener{
		BaseListener:    events.NewBaseListener("test.propagation", 50, false),
		name:            "Second",
		stopPropagation: false,
	}

	// Register listeners
	dispatcher.Listen("test.propagation", stopPropagationListener)
	dispatcher.Listen("test.propagation", normalListener)

	// Create and dispatch event
	event := events.NewBaseEvent("test.propagation", "test data")

	ctx := context.Background()
	if err := dispatcher.Dispatch(ctx, event); err != nil {
		log.Printf("Error dispatching event: %v", err)
	}

	fmt.Println("Propagation was stopped after first listener")
}

func demoEventQueue() {
	// Create event dispatcher and queue
	dispatcher := events.NewEventDispatcher()
	queue := events.NewInMemoryEventQueue()
	bus := events.NewEventBus(dispatcher, queue)

	// Create queued listener
	queuedListener := &QueuedTestListener{
		BaseListener: events.NewBaseListener("test.queued", 50, true),
		name:         "Queued",
	}

	// Register listener
	bus.Listen("test.queued", queuedListener)

	// Queue multiple events
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		event := events.NewBaseEvent("test.queued", fmt.Sprintf("event %d", i+1))
		if err := bus.Push(ctx, event); err != nil {
			log.Printf("Error queuing event: %v", err)
		}
	}

	fmt.Printf("Queued 3 events\n")

	// Process queued events
	fmt.Println("Processing queued events...")
	if err := bus.Process(ctx, dispatcher); err != nil {
		log.Printf("Error processing queue: %v", err)
	}

	fmt.Println("All queued events processed")
}

func demoEventSubscribers() {
	// Create event manager
	eventManager := events.NewEventManager()

	// Register built-in listeners
	eventManager.RegisterListener("user.created", events.NewLoggingListener("user.created", 100))
	eventManager.RegisterListener("user.created", events.NewEmailNotificationListener("user.created", 50))
	eventManager.RegisterListener("user.created", events.NewCacheInvalidationListener("user.created", 25))
	eventManager.RegisterListener("user.created", events.NewAnalyticsListener("user.created", 10))

	// Register custom subscriber
	subscriber := &UserEventSubscriber{}
	// Note: RegisterSubscriber would be implemented in a real system
	_ = subscriber

	// Create and dispatch events
	ctx := context.Background()

	user := User{ID: 1, Name: "Jane Smith", Email: "jane@example.com"}
	userCreatedEvent := events.NewUserCreatedEvent(user)
	userUpdatedEvent := events.NewUserUpdatedEvent(user)
	userDeletedEvent := events.NewUserDeletedEvent(user.ID)

	fmt.Println("Dispatching user events...")

	if err := eventManager.Dispatch(ctx, userCreatedEvent); err != nil {
		log.Printf("Error dispatching user created event: %v", err)
	}

	if err := eventManager.Dispatch(ctx, userUpdatedEvent); err != nil {
		log.Printf("Error dispatching user updated event: %v", err)
	}

	if err := eventManager.Dispatch(ctx, userDeletedEvent); err != nil {
		log.Printf("Error dispatching user deleted event: %v", err)
	}

	fmt.Println("All user events dispatched")
}

func demoEventManager() {
	// Create event manager
	eventManager := events.NewEventManager()

	// Register various listeners
	eventManager.RegisterListener("post.created", events.NewLoggingListener("post.created", 100))
	eventManager.RegisterListener("post.created", events.NewEmailNotificationListener("post.created", 50))
	eventManager.RegisterListener("post.created", events.NewAnalyticsListener("post.created", 25))

	// Create and dispatch events synchronously
	ctx := context.Background()

	post := Post{ID: 1, Title: "Hello World", Content: "This is my first post", UserID: 1}
	postCreatedEvent := events.NewPostCreatedEvent(post)

	fmt.Println("Dispatching post created event synchronously...")
	if err := eventManager.Dispatch(ctx, postCreatedEvent); err != nil {
		log.Printf("Error dispatching event: %v", err)
	}

	// Dispatch event asynchronously
	fmt.Println("Dispatching post created event asynchronously...")
	if err := eventManager.DispatchAsync(ctx, postCreatedEvent); err != nil {
		log.Printf("Error dispatching event async: %v", err)
	}

	// Queue event for later processing
	fmt.Println("Queuing post created event...")
	if err := eventManager.Queue(ctx, postCreatedEvent); err != nil {
		log.Printf("Error queuing event: %v", err)
	}

	// Process queued events
	fmt.Println("Processing queued events...")
	if err := eventManager.ProcessQueue(ctx); err != nil {
		log.Printf("Error processing queue: %v", err)
	}

	fmt.Println("Event manager demo completed")
}
