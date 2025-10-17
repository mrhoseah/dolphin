package http

import (
	"context"
	"sync"
	"time"
)

// RateLimiter represents a rate limiter
type RateLimiter struct {
	// Configuration
	rps   int
	burst int
	
	// Token bucket
	tokens     int
	lastUpdate time.Time
	
	// Mutex for thread safety
	mu sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps, burst int) *RateLimiter {
	if rps <= 0 {
		rps = 100
	}
	if burst <= 0 {
		burst = 10
	}
	
	return &RateLimiter{
		rps:        rps,
		burst:      burst,
		tokens:     burst,
		lastUpdate: time.Now(),
	}
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Update tokens based on elapsed time
	rl.updateTokens()
	
	// If we have tokens available, consume one
	if rl.tokens > 0 {
		rl.tokens--
		return nil
	}
	
	// Calculate wait time
	waitTime := time.Duration(1000/rl.rps) * time.Millisecond
	
	// Wait for token with context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		// Update tokens again after waiting
		rl.updateTokens()
		if rl.tokens > 0 {
			rl.tokens--
			return nil
		}
		return ErrRateLimitExceeded
	}
}

// updateTokens updates the token count based on elapsed time
func (rl *RateLimiter) updateTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate)
	
	// Add tokens based on elapsed time
	tokensToAdd := int(elapsed.Seconds() * float64(rl.rps))
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.burst {
			rl.tokens = rl.burst
		}
		rl.lastUpdate = now
	}
}

// GetTokens returns the current token count
func (rl *RateLimiter) GetTokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	return rl.tokens
}

// GetRPS returns the requests per second limit
func (rl *RateLimiter) GetRPS() int {
	return rl.rps
}

// GetBurst returns the burst limit
func (rl *RateLimiter) GetBurst() int {
	return rl.burst
}

// SetRPS sets the requests per second limit
func (rl *RateLimiter) SetRPS(rps int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.rps = rps
}

// SetBurst sets the burst limit
func (rl *RateLimiter) SetBurst(burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.burst = burst
	if rl.tokens > rl.burst {
		rl.tokens = rl.burst
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	return map[string]interface{}{
		"rps":           rl.rps,
		"burst":         rl.burst,
		"tokens":        rl.tokens,
		"last_update":   rl.lastUpdate,
		"utilization":   float64(rl.burst-rl.tokens) / float64(rl.burst),
		"is_limited":    rl.tokens == 0,
		"is_available":  rl.tokens > 0,
	}
}

// GetHealth returns the health status of the rate limiter
func (rl *RateLimiter) GetHealth() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	health := map[string]interface{}{
		"rps":           rl.rps,
		"burst":         rl.burst,
		"tokens":        rl.tokens,
		"is_healthy":    rl.tokens > 0,
		"is_limited":    rl.tokens == 0,
		"utilization":   float64(rl.burst-rl.tokens) / float64(rl.burst),
	}
	
	// Add timing information
	health["last_update"] = rl.lastUpdate
	health["time_since_last_update"] = time.Since(rl.lastUpdate)
	
	return health
}

// GetMetrics returns rate limiter metrics
func (rl *RateLimiter) GetMetrics() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	metrics := map[string]interface{}{
		"rps":           rl.rps,
		"burst":         rl.burst,
		"tokens":        rl.tokens,
		"utilization":   float64(rl.burst-rl.tokens) / float64(rl.burst),
		"is_limited":    rl.tokens == 0,
		"is_available":  rl.tokens > 0,
	}
	
	// Add timing information
	metrics["last_update"] = rl.lastUpdate
	metrics["time_since_last_update"] = time.Since(rl.lastUpdate)
	
	return metrics
}

// GetStatus returns a human-readable status
func (rl *RateLimiter) GetStatus() string {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	if rl.tokens > 0 {
		return fmt.Sprintf("Rate limiter is available (%d/%d tokens)", rl.tokens, rl.burst)
	}
	return fmt.Sprintf("Rate limiter is limited (0/%d tokens)", rl.burst)
}

// GetSummary returns a summary of the rate limiter state
func (rl *RateLimiter) GetSummary() string {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	return fmt.Sprintf("Rate Limiter: %d RPS, %d burst, %d tokens available", 
		rl.rps, rl.burst, rl.tokens)
}

// Reset resets the rate limiter
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.tokens = rl.burst
	rl.lastUpdate = time.Now()
}

// IsLimited returns true if the rate limiter is currently limiting requests
func (rl *RateLimiter) IsLimited() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	return rl.tokens == 0
}

// IsAvailable returns true if the rate limiter has tokens available
func (rl *RateLimiter) IsAvailable() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	return rl.tokens > 0
}

// GetUtilization returns the current utilization percentage
func (rl *RateLimiter) GetUtilization() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	return float64(rl.burst-rl.tokens) / float64(rl.burst)
}

// GetWaitTime returns the estimated wait time for the next token
func (rl *RateLimiter) GetWaitTime() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	if rl.tokens > 0 {
		return 0
	}
	
	return time.Duration(1000/rl.rps) * time.Millisecond
}

// GetNextTokenTime returns the time when the next token will be available
func (rl *RateLimiter) GetNextTokenTime() time.Time {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	if rl.tokens > 0 {
		return time.Now()
	}
	
	waitTime := time.Duration(1000/rl.rps) * time.Millisecond
	return time.Now().Add(waitTime)
}

// GetConfig returns the rate limiter configuration
func (rl *RateLimiter) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"rps":   rl.rps,
		"burst": rl.burst,
	}
}

// UpdateConfig updates the rate limiter configuration
func (rl *RateLimiter) UpdateConfig(rps, burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.rps = rps
	rl.burst = burst
	
	// Adjust tokens if necessary
	if rl.tokens > rl.burst {
		rl.tokens = rl.burst
	}
}

// GetInfo returns detailed information about the rate limiter
func (rl *RateLimiter) GetInfo() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.updateTokens()
	
	return map[string]interface{}{
		"rps":                    rl.rps,
		"burst":                  rl.burst,
		"tokens":                 rl.tokens,
		"last_update":            rl.lastUpdate,
		"time_since_last_update": time.Since(rl.lastUpdate),
		"utilization":            float64(rl.burst-rl.tokens) / float64(rl.burst),
		"is_limited":             rl.tokens == 0,
		"is_available":           rl.tokens > 0,
		"wait_time":              rl.GetWaitTime(),
		"next_token_time":        rl.GetNextTokenTime(),
	}
}
