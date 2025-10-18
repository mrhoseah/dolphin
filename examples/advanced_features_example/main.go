package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/mrhoseah/dolphin/internal/aop"
	"github.com/mrhoseah/dolphin/internal/autoconfig"
	"github.com/mrhoseah/dolphin/internal/container"
	"github.com/mrhoseah/dolphin/internal/lifecycle"
	"github.com/mrhoseah/dolphin/internal/transaction"
)

// Example interfaces and implementations
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(user *User) error
}

type EmailService interface {
	SendEmail(to, subject, body string) error
}

type UserRepository interface {
	FindByID(id int) (*User, error)
	Save(user *User) error
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Implementations
type userServiceImpl struct {
	userRepo     UserRepository
	emailService EmailService
}

func (s *userServiceImpl) New(userRepo UserRepository, emailService EmailService) {
	s.userRepo = userRepo
	s.emailService = emailService
}

func (s *userServiceImpl) GetUser(id int) (*User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Send welcome email
	s.emailService.SendEmail(user.Email, "Welcome!", "Welcome to our platform!")

	return user, nil
}

func (s *userServiceImpl) CreateUser(user *User) error {
	return s.userRepo.Save(user)
}

type userRepositoryImpl struct {
	db *sql.DB
}

func (r *userRepositoryImpl) New(db *sql.DB) {
	r.db = db
}

func (r *userRepositoryImpl) FindByID(id int) (*User, error) {
	// Simulate database query
	return &User{ID: id, Name: "John Doe", Email: "john@example.com"}, nil
}

func (r *userRepositoryImpl) Save(user *User) error {
	// Simulate database save
	return nil
}

type emailServiceImpl struct {
	smtpClient SMTPClient
}

func (s *emailServiceImpl) New(smtpClient SMTPClient) {
	s.smtpClient = smtpClient
}

func (s *emailServiceImpl) SendEmail(to, subject, body string) error {
	return s.smtpClient.Send(to, subject, body)
}

// Mock implementations
type SMTPClient interface {
	Send(to, subject, body string) error
}

type smtpClientImpl struct{}

func (c *smtpClientImpl) Send(to, subject, body string) error {
	fmt.Printf("Sending email to %s: %s - %s\n", to, subject, body)
	return nil
}

// Mock database
type MockDB struct{}

func (db *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// Return a mock transaction
	return &sql.Tx{}, nil
}

// Mock logger
type MockLogger struct{}

func (l *MockLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *MockLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

// Mock cache
type MockCache struct {
	data map[string]interface{}
}

func NewMockCache() *MockCache {
	return &MockCache{data: make(map[string]interface{})}
}

func (c *MockCache) Get(key string) (interface{}, bool) {
	value, exists := c.data[key]
	return value, exists
}

func (c *MockCache) Set(key string, value interface{}, ttl time.Duration) {
	c.data[key] = value
}

// Mock metrics collector
type MockMetricsCollector struct{}

func (m *MockMetricsCollector) RecordMethodDuration(methodName string, duration time.Duration) {
	fmt.Printf("[METRICS] Method %s took %v\n", methodName, duration)
}

func (m *MockMetricsCollector) RecordMethodError(methodName string, duration time.Duration, err error) {
	fmt.Printf("[METRICS] Method %s failed after %v: %v\n", methodName, duration, err)
}

func (m *MockMetricsCollector) IncrementTransactionCount() {
	fmt.Printf("[METRICS] Transaction started\n")
}

func (m *MockMetricsCollector) IncrementTransactionSuccessCount() {
	fmt.Printf("[METRICS] Transaction succeeded\n")
}

func (m *MockMetricsCollector) IncrementTransactionErrorCount() {
	fmt.Printf("[METRICS] Transaction failed\n")
}

// Lifecycle callbacks
type DatabaseLifecycleCallback struct {
	db *MockDB
}

func (c *DatabaseLifecycleCallback) OnPhase(ctx context.Context, phase lifecycle.Phase) error {
	switch phase {
	case lifecycle.PhaseStartup:
		fmt.Println("Database initialized")
	case lifecycle.PhaseShutdown:
		fmt.Println("Database connections closed")
	}
	return nil
}

type CacheLifecycleCallback struct {
	cache *MockCache
}

func (c *CacheLifecycleCallback) OnPhase(ctx context.Context, phase lifecycle.Phase) error {
	switch phase {
	case lifecycle.PhaseStartup:
		fmt.Println("Cache initialized")
	case lifecycle.PhaseShutdown:
		fmt.Println("Cache connections closed")
	}
	return nil
}

func main() {
	// Create application context
	ctx := context.Background()

	// Create container
	container := container.NewContainer()

	// Register services
	container.RegisterSingleton(
		reflect.TypeOf((*SMTPClient)(nil)).Elem(),
		reflect.TypeOf((*smtpClientImpl)(nil)),
	)

	container.RegisterScoped(
		reflect.TypeOf((*UserRepository)(nil)).Elem(),
		reflect.TypeOf((*userRepositoryImpl)(nil)),
	)

	container.RegisterScoped(
		reflect.TypeOf((*EmailService)(nil)).Elem(),
		reflect.TypeOf((*emailServiceImpl)(nil)),
	)

	container.RegisterScoped(
		reflect.TypeOf((*UserService)(nil)).Elem(),
		reflect.TypeOf((*userServiceImpl)(nil)),
	)

	// Apply auto-configuration
	autoConfigManager := autoconfig.NewAutoConfigurationManager()
	if err := autoConfigManager.ApplyAutoConfiguration(ctx, container); err != nil {
		log.Fatalf("Failed to apply auto-configuration: %v", err)
	}

	// Setup lifecycle management
	lifecycleManager := lifecycle.NewApplicationLifecycleManager(container)
	lifecycleManager.RegisterCallback(lifecycle.PhaseStartup, &DatabaseLifecycleCallback{})
	lifecycleManager.RegisterCallback(lifecycle.PhaseStartup, &CacheLifecycleCallback{})

	// Setup AOP
	logger := &MockLogger{}
	cache := NewMockCache()
	metrics := &MockMetricsCollector{}

	loggingAspect := aop.NewLoggingAspect(logger)
	cachingAspect := aop.NewCachingAspect(cache)
	performanceAspect := aop.NewPerformanceAspect(metrics)

	aspectRegistry := aop.NewAspectRegistry()
	aspectRegistry.RegisterAspect(reflect.TypeOf((*UserService)(nil)).Elem(), loggingAspect)
	aspectRegistry.RegisterAspect(reflect.TypeOf((*UserService)(nil)).Elem(), cachingAspect)
	aspectRegistry.RegisterAspect(reflect.TypeOf((*UserService)(nil)).Elem(), performanceAspect)

	// Setup transaction management
	mockDB := &MockDB{}
	transactionManager := transaction.NewTransactionManager(mockDB)
	transactionTemplate := transaction.NewTransactionTemplate(transactionManager)

	// Add transaction callbacks
	transactionCallbackManager := transaction.NewTransactionCallbackManager()
	transactionCallbackManager.AddCallback(transaction.NewLoggingTransactionCallback(logger))
	transactionCallbackManager.AddCallback(transaction.NewMetricsTransactionCallback(metrics))

	// Start application
	if err := lifecycleManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Resolve and use services
	userService, err := container.Resolve(reflect.TypeOf((*UserService)(nil)).Elem())
	if err != nil {
		log.Fatalf("Failed to resolve UserService: %v", err)
	}

	// Apply aspects to the service
	userServiceWithAspects := aspectRegistry.ApplyAspects(userService)

	// Use the service
	user, err := userServiceWithAspects.(UserService).GetUser(1)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	fmt.Printf("Retrieved user: %+v\n", user)

	// Example of transaction usage
	err = transactionCallbackManager.ExecuteWithCallbacks(ctx, transactionManager, func(tx *sql.Tx) error {
		// Simulate transaction work
		fmt.Println("Executing transaction...")
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if err != nil {
		log.Fatalf("Transaction failed: %v", err)
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := lifecycleManager.Stop(shutdownCtx); err != nil {
		log.Fatalf("Failed to stop application: %v", err)
	}

	fmt.Println("Application stopped gracefully")
}
