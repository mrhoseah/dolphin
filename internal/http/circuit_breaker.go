package http

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           `yaml:"failure_threshold" json:"failure_threshold"`
	SuccessThreshold int           `yaml:"success_threshold" json:"success_threshold"`
	OpenTimeout      time.Duration `yaml:"open_timeout" json:"open_timeout"`
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		OpenTimeout:      60 * time.Second,
	}
}

// CircuitBreaker represents a circuit breaker
type CircuitBreaker struct {
	config *CircuitBreakerConfig

	// State
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailTime    time.Time
	lastSuccessTime time.Time

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit is open
	if cb.state == StateOpen {
		// Check if we should transition to half-open
		if time.Since(cb.lastFailTime) >= cb.config.OpenTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
		} else {
			return nil, ErrCircuitBreakerOpen
		}
	}

	// Execute function
	result, err := fn()

	// Update state based on result
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}

	return result, err
}

// onFailure handles a failure
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()

	// Check if we should open the circuit
	if cb.failureCount >= cb.config.FailureThreshold {
		cb.state = StateOpen
		cb.failureCount = 0
	}
}

// onSuccess handles a success
func (cb *CircuitBreaker) onSuccess() {
	cb.successCount++
	cb.lastSuccessTime = time.Now()

	// Reset failure count
	cb.failureCount = 0

	// Check if we should close the circuit
	if cb.state == StateHalfOpen && cb.successCount >= cb.config.SuccessThreshold {
		cb.state = StateClosed
		cb.successCount = 0
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_fail_time":    cb.lastFailTime,
		"last_success_time": cb.lastSuccessTime,
		"is_open":           cb.state == StateOpen,
		"is_closed":         cb.state == StateClosed,
		"is_half_open":      cb.state == StateHalfOpen,
	}
}

// Reset resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailTime = time.Time{}
	cb.lastSuccessTime = time.Time{}
}

// ForceOpen forces the circuit breaker to open
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateOpen
	cb.failureCount = 0
	cb.successCount = 0
}

// ForceClose forces the circuit breaker to close
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == StateOpen
}

// IsClosed returns true if the circuit breaker is closed
func (cb *CircuitBreaker) IsClosed() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == StateClosed
}

// IsHalfOpen returns true if the circuit breaker is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == StateHalfOpen
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount
}

// GetSuccessCount returns the current success count
func (cb *CircuitBreaker) GetSuccessCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.successCount
}

// GetLastFailTime returns the last failure time
func (cb *CircuitBreaker) GetLastFailTime() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastFailTime
}

// GetLastSuccessTime returns the last success time
func (cb *CircuitBreaker) GetLastSuccessTime() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.lastSuccessTime
}

// GetConfig returns the circuit breaker configuration
func (cb *CircuitBreaker) GetConfig() *CircuitBreakerConfig {
	return cb.config
}

// SetConfig sets the circuit breaker configuration
func (cb *CircuitBreaker) SetConfig(config *CircuitBreakerConfig) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config = config
}

// UpdateFailureThreshold updates the failure threshold
func (cb *CircuitBreaker) UpdateFailureThreshold(threshold int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config.FailureThreshold = threshold
}

// UpdateSuccessThreshold updates the success threshold
func (cb *CircuitBreaker) UpdateSuccessThreshold(threshold int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config.SuccessThreshold = threshold
}

// UpdateOpenTimeout updates the open timeout
func (cb *CircuitBreaker) UpdateOpenTimeout(timeout time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config.OpenTimeout = timeout
}

// GetHealth returns the health status of the circuit breaker
func (cb *CircuitBreaker) GetHealth() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	health := map[string]interface{}{
		"state":         cb.state.String(),
		"failure_count": cb.failureCount,
		"success_count": cb.successCount,
		"is_healthy":    cb.state == StateClosed,
		"is_degraded":   cb.state == StateHalfOpen,
		"is_unhealthy":  cb.state == StateOpen,
	}

	// Add timing information
	if !cb.lastFailTime.IsZero() {
		health["last_fail_time"] = cb.lastFailTime
		health["time_since_last_fail"] = time.Since(cb.lastFailTime)
	}

	if !cb.lastSuccessTime.IsZero() {
		health["last_success_time"] = cb.lastSuccessTime
		health["time_since_last_success"] = time.Since(cb.lastSuccessTime)
	}

	return health
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	metrics := map[string]interface{}{
		"state":          cb.state.String(),
		"failure_count":  cb.failureCount,
		"success_count":  cb.successCount,
		"total_requests": cb.failureCount + cb.successCount,
		"failure_rate":   0.0,
		"success_rate":   0.0,
	}

	// Calculate rates
	total := cb.failureCount + cb.successCount
	if total > 0 {
		metrics["failure_rate"] = float64(cb.failureCount) / float64(total)
		metrics["success_rate"] = float64(cb.successCount) / float64(total)
	}

	return metrics
}

// GetStatus returns a human-readable status
func (cb *CircuitBreaker) GetStatus() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return "Circuit breaker is closed (healthy)"
	case StateOpen:
		return "Circuit breaker is open (unhealthy)"
	case StateHalfOpen:
		return "Circuit breaker is half-open (testing)"
	default:
		return "Circuit breaker is in unknown state"
	}
}

// GetSummary returns a summary of the circuit breaker state
func (cb *CircuitBreaker) GetSummary() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return fmt.Sprintf("Circuit Breaker: %s (Failures: %d, Successes: %d)",
		cb.state.String(), cb.failureCount, cb.successCount)
}
