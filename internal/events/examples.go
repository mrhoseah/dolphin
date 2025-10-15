package events

import (
	"context"
	"fmt"
	"time"
)

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	*BaseEvent
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(userID uint, email, username string) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent: NewBaseEvent("user.created", map[string]interface{}{
			"user_id":    userID,
			"email":      email,
			"username":   username,
			"created_at": time.Now(),
		}),
		UserID:    userID,
		Email:     email,
		Username:  username,
		CreatedAt: time.Now(),
	}
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	*BaseEvent
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUserUpdatedEvent creates a new user updated event
func NewUserUpdatedEvent(userID uint, email, username string) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseEvent: NewBaseEvent("user.updated", map[string]interface{}{
			"user_id":    userID,
			"email":      email,
			"username":   username,
			"updated_at": time.Now(),
		}),
		UserID:    userID,
		Email:     email,
		Username:  username,
		UpdatedAt: time.Now(),
	}
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	*BaseEvent
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
}

// NewUserDeletedEvent creates a new user deleted event
func NewUserDeletedEvent(userID uint, email string) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseEvent: NewBaseEvent("user.deleted", map[string]interface{}{
			"user_id":    userID,
			"email":      email,
			"deleted_at": time.Now(),
		}),
		UserID:    userID,
		Email:     email,
		DeletedAt: time.Now(),
	}
}

// OrderCreatedEvent represents an order creation event
type OrderCreatedEvent struct {
	*BaseEvent
	OrderID   uint      `json:"order_id"`
	UserID    uint      `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

// NewOrderCreatedEvent creates a new order created event
func NewOrderCreatedEvent(orderID, userID uint, amount float64, currency string) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		BaseEvent: NewBaseEvent("order.created", map[string]interface{}{
			"order_id":   orderID,
			"user_id":    userID,
			"amount":     amount,
			"currency":   currency,
			"created_at": time.Now(),
		}),
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		CreatedAt: time.Now(),
	}
}

// PaymentProcessedEvent represents a payment processing event
type PaymentProcessedEvent struct {
	*BaseEvent
	PaymentID   string    `json:"payment_id"`
	OrderID     uint      `json:"order_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

// NewPaymentProcessedEvent creates a new payment processed event
func NewPaymentProcessedEvent(paymentID string, orderID uint, amount float64, status string) *PaymentProcessedEvent {
	return &PaymentProcessedEvent{
		BaseEvent: NewBaseEvent("payment.processed", map[string]interface{}{
			"payment_id":   paymentID,
			"order_id":     orderID,
			"amount":       amount,
			"status":       status,
			"processed_at": time.Now(),
		}),
		PaymentID:   paymentID,
		OrderID:     orderID,
		Amount:      amount,
		Status:      status,
		ProcessedAt: time.Now(),
	}
}

// EmailNotificationListener handles email notifications
type EmailNotificationListener struct {
	priority int
}

// NewEmailNotificationListener creates a new email notification listener
func NewEmailNotificationListener() *EmailNotificationListener {
	return &EmailNotificationListener{
		priority: 100,
	}
}

func (l *EmailNotificationListener) Handle(ctx context.Context, event Event) error {
	switch e := event.(type) {
	case *UserCreatedEvent:
		return l.handleUserCreated(ctx, e)
	case *UserUpdatedEvent:
		return l.handleUserUpdated(ctx, e)
	case *OrderCreatedEvent:
		return l.handleOrderCreated(ctx, e)
	case *PaymentProcessedEvent:
		return l.handlePaymentProcessed(ctx, e)
	default:
		return fmt.Errorf("unsupported event type: %T", event)
	}
}

func (l *EmailNotificationListener) GetPriority() int {
	return l.priority
}

func (l *EmailNotificationListener) ShouldQueue() bool {
	return true // Email notifications should be queued
}

func (l *EmailNotificationListener) handleUserCreated(ctx context.Context, event *UserCreatedEvent) error {
	fmt.Printf("📧 Sending welcome email to: %s (%s)\n", event.Email, event.Username)
	// In a real implementation, this would send an actual email
	return nil
}

func (l *EmailNotificationListener) handleUserUpdated(ctx context.Context, event *UserUpdatedEvent) error {
	fmt.Printf("📧 Sending profile update notification to: %s\n", event.Email)
	return nil
}

func (l *EmailNotificationListener) handleOrderCreated(ctx context.Context, event *OrderCreatedEvent) error {
	fmt.Printf("📧 Sending order confirmation email for order #%d (%.2f %s)\n",
		event.OrderID, event.Amount, event.Currency)
	return nil
}

func (l *EmailNotificationListener) handlePaymentProcessed(ctx context.Context, event *PaymentProcessedEvent) error {
	fmt.Printf("📧 Sending payment confirmation email for payment %s (%.2f)\n",
		event.PaymentID, event.Amount)
	return nil
}

// AuditLogListener handles audit logging
type AuditLogListener struct {
	priority int
}

// NewAuditLogListener creates a new audit log listener
func NewAuditLogListener() *AuditLogListener {
	return &AuditLogListener{
		priority: 50,
	}
}

func (l *AuditLogListener) Handle(ctx context.Context, event Event) error {
	fmt.Printf("📝 Audit Log: %s - %s at %s\n",
		event.GetName(), event.GetID(), event.GetTimestamp().Format(time.RFC3339))

	// In a real implementation, this would write to an audit log database
	return nil
}

func (l *AuditLogListener) GetPriority() int {
	return l.priority
}

func (l *AuditLogListener) ShouldQueue() bool {
	return false // Audit logs should be processed immediately
}

// AnalyticsListener handles analytics tracking
type AnalyticsListener struct {
	priority int
}

// NewAnalyticsListener creates a new analytics listener
func NewAnalyticsListener() *AnalyticsListener {
	return &AnalyticsListener{
		priority: 25,
	}
}

func (l *AnalyticsListener) Handle(ctx context.Context, event Event) error {
	fmt.Printf("📊 Analytics: Tracking event %s for user behavior analysis\n", event.GetName())

	// In a real implementation, this would send data to analytics service
	return nil
}

func (l *AnalyticsListener) GetPriority() int {
	return l.priority
}

func (l *AnalyticsListener) ShouldQueue() bool {
	return true // Analytics can be processed asynchronously
}

// CacheInvalidationListener handles cache invalidation
type CacheInvalidationListener struct {
	priority int
}

// NewCacheInvalidationListener creates a new cache invalidation listener
func NewCacheInvalidationListener() *CacheInvalidationListener {
	return &CacheInvalidationListener{
		priority: 10,
	}
}

func (l *CacheInvalidationListener) Handle(ctx context.Context, event Event) error {
	switch e := event.(type) {
	case *UserUpdatedEvent:
		fmt.Printf("🗑️ Invalidating cache for user: %d\n", e.UserID)
	case *UserDeletedEvent:
		fmt.Printf("🗑️ Clearing cache for deleted user: %d\n", e.UserID)
	default:
		fmt.Printf("🗑️ Cache invalidation not needed for event: %s\n", event.GetName())
	}

	return nil
}

func (l *CacheInvalidationListener) GetPriority() int {
	return l.priority
}

func (l *CacheInvalidationListener) ShouldQueue() bool {
	return false // Cache invalidation should be immediate
}
