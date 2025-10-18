package lifecycle

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"dolphin/internal/container"
)

// ApplicationContext manages the application lifecycle
type ApplicationContext struct {
	container     *container.Container
	phases        map[Phase][]LifecycleCallback
	mu            sync.RWMutex
	currentPhase  Phase
	started       bool
	shutdownHooks []ShutdownHook
}

type Phase int

const (
	PhaseInitialization Phase = iota
	PhaseConfiguration
	PhaseStartup
	PhaseRunning
	PhaseShutdown
)

// LifecycleCallback defines a callback for lifecycle events
type LifecycleCallback interface {
	OnPhase(ctx context.Context, phase Phase) error
}

// ShutdownHook defines a hook for shutdown events
type ShutdownHook interface {
	OnShutdown(ctx context.Context) error
	Timeout() time.Duration
}

// NewApplicationContext creates a new application context
func NewApplicationContext(container *container.Container) *ApplicationContext {
	return &ApplicationContext{
		container:     container,
		phases:        make(map[Phase][]LifecycleCallback),
		shutdownHooks: make([]ShutdownHook, 0),
	}
}

// RegisterCallback registers a callback for a specific phase
func (ac *ApplicationContext) RegisterCallback(phase Phase, callback LifecycleCallback) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.phases[phase] = append(ac.phases[phase], callback)
}

// RegisterShutdownHook registers a shutdown hook
func (ac *ApplicationContext) RegisterShutdownHook(hook ShutdownHook) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.shutdownHooks = append(ac.shutdownHooks, hook)
}

// Start starts the application lifecycle
func (ac *ApplicationContext) Start(ctx context.Context) error {
	ac.mu.Lock()
	if ac.started {
		ac.mu.Unlock()
		return fmt.Errorf("application context already started")
	}
	ac.started = true
	ac.mu.Unlock()

	phases := []Phase{
		PhaseInitialization,
		PhaseConfiguration,
		PhaseStartup,
		PhaseRunning,
	}

	for _, phase := range phases {
		ac.currentPhase = phase

		if err := ac.executePhase(ctx, phase); err != nil {
			return fmt.Errorf("failed in phase %v: %w", phase, err)
		}
	}

	return nil
}

// Stop stops the application lifecycle
func (ac *ApplicationContext) Stop(ctx context.Context) error {
	ac.mu.Lock()
	if !ac.started {
		ac.mu.Unlock()
		return fmt.Errorf("application context not started")
	}
	ac.started = false
	ac.mu.Unlock()

	ac.currentPhase = PhaseShutdown

	// Execute shutdown hooks
	for _, hook := range ac.shutdownHooks {
		hookCtx, cancel := context.WithTimeout(ctx, hook.Timeout())
		defer cancel()

		if err := hook.OnShutdown(hookCtx); err != nil {
			return fmt.Errorf("shutdown hook failed: %w", err)
		}
	}

	// Execute shutdown phase callbacks
	if err := ac.executePhase(ctx, PhaseShutdown); err != nil {
		return fmt.Errorf("failed during shutdown: %w", err)
	}

	return nil
}

// GetCurrentPhase returns the current phase
func (ac *ApplicationContext) GetCurrentPhase() Phase {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.currentPhase
}

// IsStarted returns whether the application is started
func (ac *ApplicationContext) IsStarted() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.started
}

func (ac *ApplicationContext) executePhase(ctx context.Context, phase Phase) error {
	ac.mu.RLock()
	callbacks := ac.phases[phase]
	ac.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback.OnPhase(ctx, phase); err != nil {
			return err
		}
	}

	return nil
}

// DatabaseLifecycleCallback handles database lifecycle events
type DatabaseLifecycleCallback struct {
	db *sql.DB
}

func NewDatabaseLifecycleCallback(db *sql.DB) *DatabaseLifecycleCallback {
	return &DatabaseLifecycleCallback{db: db}
}

func (c *DatabaseLifecycleCallback) OnPhase(ctx context.Context, phase Phase) error {
	switch phase {
	case PhaseStartup:
		// Run migrations
		return c.runMigrations(ctx)
	case PhaseShutdown:
		// Close database connections
		return c.db.Close()
	}
	return nil
}

func (c *DatabaseLifecycleCallback) runMigrations(ctx context.Context) error {
	// Implementation for running database migrations
	// This would typically run database migrations
	return nil
}

// CacheLifecycleCallback handles cache lifecycle events
type CacheLifecycleCallback struct {
	cache Cache
}

func NewCacheLifecycleCallback(cache Cache) *CacheLifecycleCallback {
	return &CacheLifecycleCallback{cache: cache}
}

func (c *CacheLifecycleCallback) OnPhase(ctx context.Context, phase Phase) error {
	switch phase {
	case PhaseStartup:
		// Initialize cache
		return c.cache.Initialize()
	case PhaseShutdown:
		// Close cache connections
		return c.cache.Close()
	}
	return nil
}

// ServerLifecycleCallback handles server lifecycle events
type ServerLifecycleCallback struct {
	server Server
}

func NewServerLifecycleCallback(server Server) *ServerLifecycleCallback {
	return &ServerLifecycleCallback{server: server}
}

func (c *ServerLifecycleCallback) OnPhase(ctx context.Context, phase Phase) error {
	switch phase {
	case PhaseStartup:
		// Start server
		return c.server.Start()
	case PhaseShutdown:
		// Stop server
		return c.server.Stop(ctx)
	}
	return nil
}

// GracefulShutdownHook provides graceful shutdown functionality
type GracefulShutdownHook struct {
	timeout time.Duration
	handler func(ctx context.Context) error
}

func NewGracefulShutdownHook(timeout time.Duration, handler func(ctx context.Context) error) *GracefulShutdownHook {
	return &GracefulShutdownHook{
		timeout: timeout,
		handler: handler,
	}
}

func (h *GracefulShutdownHook) OnShutdown(ctx context.Context) error {
	return h.handler(ctx)
}

func (h *GracefulShutdownHook) Timeout() time.Duration {
	return h.timeout
}

// ApplicationLifecycleManager manages the overall application lifecycle
type ApplicationLifecycleManager struct {
	context *ApplicationContext
	mu      sync.RWMutex
}

func NewApplicationLifecycleManager(container *container.Container) *ApplicationLifecycleManager {
	return &ApplicationLifecycleManager{
		context: NewApplicationContext(container),
	}
}

func (m *ApplicationLifecycleManager) Start(ctx context.Context) error {
	return m.context.Start(ctx)
}

func (m *ApplicationLifecycleManager) Stop(ctx context.Context) error {
	return m.context.Stop(ctx)
}

func (m *ApplicationLifecycleManager) RegisterCallback(phase Phase, callback LifecycleCallback) {
	m.context.RegisterCallback(phase, callback)
}

func (m *ApplicationLifecycleManager) RegisterShutdownHook(hook ShutdownHook) {
	m.context.RegisterShutdownHook(hook)
}

func (m *ApplicationLifecycleManager) GetContext() *ApplicationContext {
	return m.context
}

// Interfaces for dependencies
type Cache interface {
	Initialize() error
	Close() error
}

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

// LifecycleEvent represents a lifecycle event
type LifecycleEvent struct {
	Phase     Phase
	Timestamp time.Time
	Data      map[string]interface{}
}

// LifecycleEventPublisher publishes lifecycle events
type LifecycleEventPublisher struct {
	subscribers []LifecycleEventSubscriber
	mu          sync.RWMutex
}

type LifecycleEventSubscriber interface {
	OnLifecycleEvent(event LifecycleEvent)
}

func NewLifecycleEventPublisher() *LifecycleEventPublisher {
	return &LifecycleEventPublisher{
		subscribers: make([]LifecycleEventSubscriber, 0),
	}
}

func (p *LifecycleEventPublisher) Subscribe(subscriber LifecycleEventSubscriber) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *LifecycleEventPublisher) PublishEvent(event LifecycleEvent) {
	p.mu.RLock()
	subscribers := p.subscribers
	p.mu.RUnlock()

	for _, subscriber := range subscribers {
		subscriber.OnLifecycleEvent(event)
	}
}

// DefaultLifecycleCallbacks returns default lifecycle callbacks
func DefaultLifecycleCallbacks(container *container.Container) []LifecycleCallback {
	callbacks := make([]LifecycleCallback, 0)

	// Add database callback if database service is available
	// Add cache callback if cache service is available
	// Add server callback if server service is available

	return callbacks
}
