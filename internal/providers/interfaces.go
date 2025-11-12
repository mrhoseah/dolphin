package providers

import (
	"io"
	"time"

	"github.com/mrhoseah/dolphin/internal/events"
)

// ServiceProvider defines the interface for all service providers
type ServiceProvider interface {
	// Register services in the container
	Register() error

	// Boot services after registration
	Boot() error

	// Get provider name
	Name() string

	// Get provider priority (lower = higher priority)
	Priority() int
}

// EmailProvider handles email sending
type EmailProvider interface {
	// Send sends an email
	Send(to, subject, body string) error

	// SendWithTemplate sends email using a template
	SendWithTemplate(to, subject, template string, data map[string]interface{}) error

	// SendBulk sends emails to multiple recipients
	SendBulk(recipients []string, subject, body string) error

	// Queue queues an email for later sending
	Queue(to, subject, body string, delay time.Duration) error
}

// NotificationProvider handles notifications
type NotificationProvider interface {
	// Send sends a notification
	Send(userID uint, title, message string) error

	// SendToChannel sends notification to a channel
	SendToChannel(channel string, title, message string) error

	// MarkAsRead marks notification as read
	MarkAsRead(notificationID uint) error

	// GetUserNotifications gets user's notifications
	GetUserNotifications(userID uint, limit int) ([]Notification, error)
}

// StorageProvider handles file storage
type StorageProvider interface {
	// Put stores a file
	Put(path string, content io.Reader) error

	// Get retrieves a file
	Get(path string) (io.ReadCloser, error)

	// Delete removes a file
	Delete(path string) error

	// Exists checks if file exists
	Exists(path string) bool

	// URL generates public URL for file
	URL(path string) string

	// Size gets file size
	Size(path string) (int64, error)
}

// CacheProvider handles caching
type CacheProvider interface {
	// Get retrieves value from cache
	Get(key string) (interface{}, error)

	// Put stores value in cache
	Put(key string, value interface{}, ttl time.Duration) error

	// Delete removes value from cache
	Delete(key string) error

	// Exists checks if key exists
	Exists(key string) bool

	// Clear clears all cache
	Clear() error

	// Increment increments a numeric value
	Increment(key string, delta int64) (int64, error)
}

// QueueProvider handles job queues
type QueueProvider interface {
	// Push adds job to queue
	Push(queue string, job Job) error

	// Pop gets job from queue
	Pop(queue string) (Job, error)

	// Process processes jobs from queue
	Process(queue string, handler JobHandler) error

	// Size gets queue size
	Size(queue string) (int, error)

	// Clear clears queue
	Clear(queue string) error
}

// SearchProvider handles search functionality
type SearchProvider interface {
	// Index adds document to search index
	Index(index string, id string, document map[string]interface{}) error

	// Search searches documents
	Search(index string, query string, filters map[string]interface{}) ([]SearchResult, error)

	// Delete removes document from index
	Delete(index string, id string) error

	// Update updates document in index
	Update(index string, id string, document map[string]interface{}) error
}

// PaymentProvider handles payments
type PaymentProvider interface {
	// CreatePayment creates a payment
	CreatePayment(amount float64, currency string, description string) (*Payment, error)

	// ProcessPayment processes a payment
	ProcessPayment(paymentID string) error

	// Refund refunds a payment
	Refund(paymentID string, amount float64) error

	// GetPaymentStatus gets payment status
	GetPaymentStatus(paymentID string) (PaymentStatus, error)
}

// SMSProvider handles SMS sending
type SMSProvider interface {
	// Send sends SMS
	Send(to, message string) error

	// SendBulk sends SMS to multiple recipients
	SendBulk(recipients []string, message string) error

	// SendWithTemplate sends SMS using template
	SendWithTemplate(to, template string, data map[string]interface{}) error
}

// SocialProvider handles social media integration
type SocialProvider interface {
	// GetAuthURL gets OAuth authorization URL
	GetAuthURL(state string) string

	// HandleCallback handles OAuth callback
	HandleCallback(code string) (*SocialUser, error)

	// GetUserInfo gets user info from social platform
	GetUserInfo(accessToken string) (*SocialUser, error)

	// PostContent posts content to social platform
	PostContent(accessToken, content string) error
}

// AnalyticsProvider handles analytics
type AnalyticsProvider interface {
	// Track tracks an event
	Track(userID uint, event string, properties map[string]interface{}) error

	// PageView tracks page view
	PageView(userID uint, page string, properties map[string]interface{}) error

	// Identify identifies a user
	Identify(userID uint, traits map[string]interface{}) error

	// GetMetrics gets analytics metrics
	GetMetrics(startDate, endDate time.Time) (*AnalyticsMetrics, error)
}

// LogProvider handles logging
type LogProvider interface {
	// Log logs a message
	Log(level LogLevel, message string, fields map[string]interface{}) error

	// Debug logs debug message
	Debug(message string, fields map[string]interface{}) error

	// Info logs info message
	Info(message string, fields map[string]interface{}) error

	// Warn logs warning message
	Warn(message string, fields map[string]interface{}) error

	// Error logs error message
	Error(message string, fields map[string]interface{}) error

	// Fatal logs fatal message
	Fatal(message string, fields map[string]interface{}) error
}

// ConfigProvider handles configuration management
type ConfigProvider interface {
	// Get gets configuration value
	Get(key string) interface{}

	// GetString gets string configuration value
	GetString(key string) string

	// GetInt gets int configuration value
	GetInt(key string) int

	// GetBool gets bool configuration value
	GetBool(key string) bool

	// GetFloat gets float configuration value
	GetFloat(key string) float64

	// Set sets configuration value
	Set(key string, value interface{}) error

	// Watch watches for configuration changes
	Watch(key string, callback func(interface{})) error

	// Reload reloads configuration
	Reload() error
}

// DatabaseProvider handles database operations
type DatabaseProvider interface {
	// Query executes raw query
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)

	// Transaction executes function in transaction
	Transaction(fn func(tx Transaction) error) error

	// Migrate runs migrations
	Migrate() error

	// Rollback rolls back migrations
	Rollback(steps int) error

	// Seed seeds database
	Seed() error

	// Backup creates database backup
	Backup(path string) error

	// Restore restores database from backup
	Restore(path string) error
}

// SecurityProvider handles security features
type SecurityProvider interface {
	// Hash hashes a password
	Hash(password string) (string, error)

	// Verify verifies a password
	Verify(password, hash string) bool

	// GenerateToken generates a secure token
	GenerateToken(length int) (string, error)

	// Encrypt encrypts data
	Encrypt(data []byte) ([]byte, error)

	// Decrypt decrypts data
	Decrypt(data []byte) ([]byte, error)

	// ValidateCSRF validates CSRF token
	ValidateCSRF(token string) bool

	// GenerateCSRF generates CSRF token
	GenerateCSRF() string
}

// MonitoringProvider handles application monitoring
type MonitoringProvider interface {
	// StartTransaction starts a performance transaction
	StartTransaction(name string) Transaction

	// EndTransaction ends a transaction
	EndTransaction(transaction Transaction) error

	// RecordMetric records a custom metric
	RecordMetric(name string, value float64, tags map[string]string) error

	// RecordError records an error
	RecordError(err error, tags map[string]string) error

	// HealthCheck performs health check
	HealthCheck() (*HealthStatus, error)

	// GetMetrics gets monitoring metrics
	GetMetrics() (*MonitoringMetrics, error)
}

// Data structures for providers

type Notification struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type Job struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Payload  map[string]interface{} `json:"payload"`
	Attempts int                    `json:"attempts"`
	Delay    time.Duration          `json:"delay"`
}

type JobHandler func(job Job) error

type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Document map[string]interface{} `json:"document"`
}

type Payment struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type SocialUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
}

type AnalyticsMetrics struct {
	PageViews   int64   `json:"page_views"`
	UniqueUsers int64   `json:"unique_users"`
	BounceRate  float64 `json:"bounce_rate"`
	AvgSession  float64 `json:"avg_session"`
	Conversions int64   `json:"conversions"`
}

type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
	LogFatal LogLevel = "fatal"
)

type Transaction interface {
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)
	Execute(query string, args ...interface{}) error
	Commit() error
	Rollback() error
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Services  map[string]string `json:"services"`
	Timestamp time.Time         `json:"timestamp"`
}

type MonitoringMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	RequestRate float64 `json:"request_rate"`
	ErrorRate   float64 `json:"error_rate"`
}

// EventProvider handles event dispatching and listening
type EventProvider interface {
	ServiceProvider
	EventBus() events.EventBus
	EventFactory() events.EventFactory
	EventSerializer() events.EventSerializer
	EventStore() events.EventStore
}
