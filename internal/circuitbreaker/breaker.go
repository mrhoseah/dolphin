package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed   State = iota // Circuit is closed, requests pass through
	StateOpen                  // Circuit is open, requests are blocked
	StateHalfOpen              // Circuit is half-open, testing if service is back
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config represents circuit breaker configuration
type Config struct {
	// Failure thresholds
	FailureThreshold int           `yaml:"failure_threshold" json:"failure_threshold"`
	SuccessThreshold int           `yaml:"success_threshold" json:"success_threshold"`
	TimeoutThreshold time.Duration `yaml:"timeout_threshold" json:"timeout_threshold"`

	// Timeouts
	OpenTimeout     time.Duration `yaml:"open_timeout" json:"open_timeout"`
	HalfOpenTimeout time.Duration `yaml:"half_open_timeout" json:"half_open_timeout"`
	RequestTimeout  time.Duration `yaml:"request_timeout" json:"request_timeout"`

	// Retry configuration
	MaxRetries        int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay        time.Duration `yaml:"retry_delay" json:"retry_delay"`
	BackoffMultiplier float64       `yaml:"backoff_multiplier" json:"backoff_multiplier"`
	MaxBackoffDelay   time.Duration `yaml:"max_backoff_delay" json:"max_backoff_delay"`

	// Monitoring
	EnableMetrics bool `yaml:"enable_metrics" json:"enable_metrics"`
	EnableLogging bool `yaml:"enable_logging" json:"enable_logging"`

	// Custom error handling
	IsFailure func(error) bool `yaml:"-" json:"-"`
	IsSuccess func(error) bool `yaml:"-" json:"-"`
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() *Config {
	return &Config{
		FailureThreshold:  5,
		SuccessThreshold:  3,
		TimeoutThreshold:  1 * time.Second,
		OpenTimeout:       30 * time.Second,
		HalfOpenTimeout:   10 * time.Second,
		RequestTimeout:    5 * time.Second,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		BackoffMultiplier: 2.0,
		MaxBackoffDelay:   30 * time.Second,
		EnableMetrics:     true,
		EnableLogging:     true,
		IsFailure:         defaultIsFailure,
		IsSuccess:         defaultIsSuccess,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name   string
	config *Config
	logger *zap.Logger

	// State management
	state   State
	stateMu sync.RWMutex

	// Counters
	failureCount int
	successCount int
	requestCount int

	// Timestamps
	lastFailureTime time.Time
	lastRequestTime time.Time
	stateChangeTime time.Time

	// Mutex for thread safety
	mu sync.RWMutex

	// Metrics
	metrics *Metrics
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config *Config, logger *zap.Logger) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig()
	}

	cb := &CircuitBreaker{
		name:            name,
		config:          config,
		logger:          logger,
		state:           StateClosed,
		stateChangeTime: time.Now(),
	}

	if config.EnableMetrics {
		cb.metrics = NewMetrics(name)
	}

	return cb
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Check if circuit is open
	if !cb.allowRequest() {
		cb.recordRejectedRequest()
		return nil, fmt.Errorf("circuit breaker %s is %s", cb.name, cb.getState())
	}

	// Execute the function
	result, err := cb.executeWithTimeout(ctx, fn)

	// Record the result
	cb.recordResult(err)

	return result, err
}

// ExecuteAsync executes a function asynchronously with circuit breaker protection
func (cb *CircuitBreaker) ExecuteAsync(ctx context.Context, fn func() (interface{}, error)) <-chan Result {
	resultChan := make(chan Result, 1)

	go func() {
		defer close(resultChan)

		result, err := cb.Execute(ctx, fn)
		resultChan <- Result{
			Value: result,
			Error: err,
		}
	}()

	return resultChan
}

// allowRequest checks if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.stateMu.RLock()
	defer cb.stateMu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if enough time has passed to try half-open
		return time.Since(cb.lastFailureTime) >= cb.config.OpenTimeout
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// executeWithTimeout executes a function with timeout
func (cb *CircuitBreaker) executeWithTimeout(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	if cb.config.RequestTimeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, cb.config.RequestTimeout)
		defer cancel()
		ctx = timeoutCtx
	}

	// Create a channel to receive the result
	resultChan := make(chan struct {
		value interface{}
		err   error
	}, 1)

	// Execute the function in a goroutine
	go func() {
		value, err := fn()
		resultChan <- struct {
			value interface{}
			err   error
		}{value, err}
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result.value, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// recordResult records the result of a request
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.requestCount++
	cb.lastRequestTime = time.Now()

	isFailure := cb.config.IsFailure(err)
	isSuccess := cb.config.IsSuccess(err)

	if isFailure {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.config.EnableLogging {
			cb.logger.Warn("Circuit breaker request failed",
				zap.String("circuit", cb.name),
				zap.Int("failure_count", cb.failureCount),
				zap.Error(err))
		}

		cb.updateState()
	} else if isSuccess {
		cb.successCount++

		if cb.config.EnableLogging {
			cb.logger.Debug("Circuit breaker request succeeded",
				zap.String("circuit", cb.name),
				zap.Int("success_count", cb.successCount))
		}

		cb.updateState()
	}

	// Update metrics
	if cb.metrics != nil {
		cb.metrics.RecordRequest(isFailure)
	}
}

// recordRejectedRequest records a rejected request
func (cb *CircuitBreaker) recordRejectedRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.requestCount++

	if cb.config.EnableLogging {
		cb.logger.Warn("Circuit breaker request rejected",
			zap.String("circuit", cb.name),
			zap.String("state", cb.getState().String()))
	}

	// Update metrics
	if cb.metrics != nil {
		cb.metrics.RecordRejected()
	}
}

// updateState updates the circuit breaker state based on current conditions
func (cb *CircuitBreaker) updateState() {
	cb.stateMu.Lock()
	defer cb.stateMu.Unlock()

	oldState := cb.state

	switch cb.state {
	case StateClosed:
		// Check if we should open the circuit
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = StateOpen
			cb.stateChangeTime = time.Now()

			if cb.config.EnableLogging {
				cb.logger.Warn("Circuit breaker opened",
					zap.String("circuit", cb.name),
					zap.Int("failure_count", cb.failureCount),
					zap.Duration("open_timeout", cb.config.OpenTimeout))
			}
		}

	case StateOpen:
		// Check if we should try half-open
		if time.Since(cb.lastFailureTime) >= cb.config.OpenTimeout {
			cb.state = StateHalfOpen
			cb.stateChangeTime = time.Now()
			cb.successCount = 0

			if cb.config.EnableLogging {
				cb.logger.Info("Circuit breaker half-opened",
					zap.String("circuit", cb.name),
					zap.Duration("half_open_timeout", cb.config.HalfOpenTimeout))
			}
		}

	case StateHalfOpen:
		// Check if we should close or open the circuit
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.stateChangeTime = time.Now()
			cb.failureCount = 0
			cb.successCount = 0

			if cb.config.EnableLogging {
				cb.logger.Info("Circuit breaker closed",
					zap.String("circuit", cb.name),
					zap.Int("success_count", cb.successCount))
			}
		} else if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = StateOpen
			cb.stateChangeTime = time.Now()

			if cb.config.EnableLogging {
				cb.logger.Warn("Circuit breaker re-opened",
					zap.String("circuit", cb.name),
					zap.Int("failure_count", cb.failureCount))
			}
		}
	}

	// Log state change
	if oldState != cb.state && cb.config.EnableLogging {
		cb.logger.Info("Circuit breaker state changed",
			zap.String("circuit", cb.name),
			zap.String("old_state", oldState.String()),
			zap.String("new_state", cb.state.String()))
	}
}

// getState returns the current state
func (cb *CircuitBreaker) getState() State {
	cb.stateMu.RLock()
	defer cb.stateMu.RUnlock()
	return cb.state
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() State {
	return cb.getState()
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	cb.stateMu.RLock()
	state := cb.state
	stateChangeTime := cb.stateChangeTime
	cb.stateMu.RUnlock()

	return Stats{
		Name:            cb.name,
		State:           state,
		RequestCount:    cb.requestCount,
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		LastFailureTime: cb.lastFailureTime,
		LastRequestTime: cb.lastRequestTime,
		StateChangeTime: stateChangeTime,
		FailureRate:     cb.calculateFailureRate(),
		SuccessRate:     cb.calculateSuccessRate(),
	}
}

// calculateFailureRate calculates the failure rate
func (cb *CircuitBreaker) calculateFailureRate() float64 {
	if cb.requestCount == 0 {
		return 0.0
	}
	return float64(cb.failureCount) / float64(cb.requestCount) * 100.0
}

// calculateSuccessRate calculates the success rate
func (cb *CircuitBreaker) calculateSuccessRate() float64 {
	if cb.requestCount == 0 {
		return 0.0
	}
	return float64(cb.successCount) / float64(cb.requestCount) * 100.0
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.stateMu.Lock()
	cb.state = StateClosed
	cb.stateChangeTime = time.Now()
	cb.stateMu.Unlock()

	cb.failureCount = 0
	cb.successCount = 0
	cb.requestCount = 0
	cb.lastFailureTime = time.Time{}
	cb.lastRequestTime = time.Time{}

	if cb.config.EnableLogging {
		cb.logger.Info("Circuit breaker reset",
			zap.String("circuit", cb.name))
	}
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.stateMu.Lock()
	defer cb.stateMu.Unlock()

	cb.state = StateOpen
	cb.stateChangeTime = time.Now()

	if cb.config.EnableLogging {
		cb.logger.Info("Circuit breaker forced open",
			zap.String("circuit", cb.name))
	}
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.stateMu.Lock()
	defer cb.stateMu.Unlock()

	cb.state = StateClosed
	cb.stateChangeTime = time.Now()

	if cb.config.EnableLogging {
		cb.logger.Info("Circuit breaker forced closed",
			zap.String("circuit", cb.name))
	}
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() *Metrics {
	return cb.metrics
}

// Result represents the result of an async operation
type Result struct {
	Value interface{}
	Error error
}

// Stats represents circuit breaker statistics
type Stats struct {
	Name            string    `json:"name"`
	State           State     `json:"state"`
	RequestCount    int       `json:"request_count"`
	FailureCount    int       `json:"failure_count"`
	SuccessCount    int       `json:"success_count"`
	LastFailureTime time.Time `json:"last_failure_time"`
	LastRequestTime time.Time `json:"last_request_time"`
	StateChangeTime time.Time `json:"state_change_time"`
	FailureRate     float64   `json:"failure_rate"`
	SuccessRate     float64   `json:"success_rate"`
}

// Default error handling functions
func defaultIsFailure(err error) bool {
	return err != nil
}

func defaultIsSuccess(err error) bool {
	return err == nil
}

// ConfigFromEnv creates config from environment variables
func ConfigFromEnv(name string) *Config {
	config := DefaultConfig()

	// Override with environment variables if present
	// This would typically read from environment variables
	// For now, return the default config
	return config
}
