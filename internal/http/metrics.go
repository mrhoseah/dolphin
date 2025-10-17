package http

import (
	"fmt"
	"sync"
	"time"
)

// Metrics represents HTTP client metrics
type Metrics struct {
	// Request counts
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64

	// Response time metrics
	totalResponseTime time.Duration
	minResponseTime   time.Duration
	maxResponseTime   time.Duration

	// Status code counts
	statusCodes map[int]int64

	// Method counts
	methods map[HTTPMethod]int64

	// Error counts
	errors map[string]int64

	// Retry metrics
	totalRetries int64
	retryCounts  map[int]int64

	// Circuit breaker metrics
	circuitBreakerTrips  int64
	circuitBreakerResets int64

	// Rate limiter metrics
	rateLimitHits int64

	// Timing
	startTime   time.Time
	lastRequest time.Time

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		statusCodes: make(map[int]int64),
		methods:     make(map[HTTPMethod]int64),
		errors:      make(map[string]int64),
		retryCounts: make(map[int]int64),
		startTime:   time.Now(),
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(method HTTPMethod, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update counts
	m.totalRequests++
	m.lastRequest = time.Now()

	// Update method count
	m.methods[method]++

	// Update status code count
	m.statusCodes[statusCode]++

	// Update success/failure counts
	if statusCode >= 200 && statusCode < 400 {
		m.successfulRequests++
	} else {
		m.failedRequests++
	}

	// Update response time metrics
	m.totalResponseTime += duration

	if m.minResponseTime == 0 || duration < m.minResponseTime {
		m.minResponseTime = duration
	}

	if duration > m.maxResponseTime {
		m.maxResponseTime = duration
	}
}

// RecordRetry records a retry
func (m *Metrics) RecordRetry(retryCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRetries++
	m.retryCounts[retryCount]++
}

// RecordError records an error
func (m *Metrics) RecordError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors[errorType]++
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func (m *Metrics) RecordCircuitBreakerTrip() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.circuitBreakerTrips++
}

// RecordCircuitBreakerReset records a circuit breaker reset
func (m *Metrics) RecordCircuitBreakerReset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.circuitBreakerResets++
}

// RecordRateLimitHit records a rate limit hit
func (m *Metrics) RecordRateLimitHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rateLimitHits++
}

// GetStats returns current metrics
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate averages
	var avgResponseTime time.Duration
	if m.totalRequests > 0 {
		avgResponseTime = m.totalResponseTime / time.Duration(m.totalRequests)
	}

	// Calculate success rate
	var successRate float64
	if m.totalRequests > 0 {
		successRate = float64(m.successfulRequests) / float64(m.totalRequests) * 100
	}

	// Calculate failure rate
	var failureRate float64
	if m.totalRequests > 0 {
		failureRate = float64(m.failedRequests) / float64(m.totalRequests) * 100
	}

	// Calculate retry rate
	var retryRate float64
	if m.totalRequests > 0 {
		retryRate = float64(m.totalRetries) / float64(m.totalRequests) * 100
	}

	// Calculate uptime
	uptime := time.Since(m.startTime)

	// Calculate requests per second
	var rps float64
	if uptime.Seconds() > 0 {
		rps = float64(m.totalRequests) / uptime.Seconds()
	}

	return map[string]interface{}{
		"total_requests":         m.totalRequests,
		"successful_requests":    m.successfulRequests,
		"failed_requests":        m.failedRequests,
		"success_rate":           successRate,
		"failure_rate":           failureRate,
		"total_retries":          m.totalRetries,
		"retry_rate":             retryRate,
		"avg_response_time":      avgResponseTime,
		"min_response_time":      m.minResponseTime,
		"max_response_time":      m.maxResponseTime,
		"status_codes":           m.statusCodes,
		"methods":                m.methods,
		"errors":                 m.errors,
		"retry_counts":           m.retryCounts,
		"circuit_breaker_trips":  m.circuitBreakerTrips,
		"circuit_breaker_resets": m.circuitBreakerResets,
		"rate_limit_hits":        m.rateLimitHits,
		"uptime":                 uptime,
		"rps":                    rps,
		"start_time":             m.startTime,
		"last_request":           m.lastRequest,
	}
}

// GetRequestStats returns request-specific statistics
func (m *Metrics) GetRequestStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate averages
	var avgResponseTime time.Duration
	if m.totalRequests > 0 {
		avgResponseTime = m.totalResponseTime / time.Duration(m.totalRequests)
	}

	// Calculate success rate
	var successRate float64
	if m.totalRequests > 0 {
		successRate = float64(m.successfulRequests) / float64(m.totalRequests) * 100
	}

	// Calculate failure rate
	var failureRate float64
	if m.totalRequests > 0 {
		failureRate = float64(m.failedRequests) / float64(m.totalRequests) * 100
	}

	// Calculate retry rate
	var retryRate float64
	if m.totalRequests > 0 {
		retryRate = float64(m.totalRetries) / float64(m.totalRequests) * 100
	}

	// Calculate uptime
	uptime := time.Since(m.startTime)

	// Calculate requests per second
	var rps float64
	if uptime.Seconds() > 0 {
		rps = float64(m.totalRequests) / uptime.Seconds()
	}

	return map[string]interface{}{
		"total_requests":      m.totalRequests,
		"successful_requests": m.successfulRequests,
		"failed_requests":     m.failedRequests,
		"success_rate":        successRate,
		"failure_rate":        failureRate,
		"total_retries":       m.totalRetries,
		"retry_rate":          retryRate,
		"avg_response_time":   avgResponseTime,
		"min_response_time":   m.minResponseTime,
		"max_response_time":   m.maxResponseTime,
		"uptime":              uptime,
		"rps":                 rps,
	}
}

// GetStatusCodeStats returns status code statistics
func (m *Metrics) GetStatusCodeStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate status code distribution
	statusCodeDistribution := make(map[string]float64)
	for statusCode, count := range m.statusCodes {
		if m.totalRequests > 0 {
			statusCodeDistribution[fmt.Sprintf("%d", statusCode)] = float64(count) / float64(m.totalRequests) * 100
		}
	}

	return map[string]interface{}{
		"status_codes":             m.statusCodes,
		"status_code_distribution": statusCodeDistribution,
		"total_status_codes":       len(m.statusCodes),
		"most_common_status_code":  m.getMostCommonStatusCode(),
		"least_common_status_code": m.getLeastCommonStatusCode(),
	}
}

// GetMethodStats returns method statistics
func (m *Metrics) GetMethodStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate method distribution
	methodDistribution := make(map[string]float64)
	for method, count := range m.methods {
		if m.totalRequests > 0 {
			methodDistribution[method.String()] = float64(count) / float64(m.totalRequests) * 100
		}
	}

	return map[string]interface{}{
		"methods":             m.methods,
		"method_distribution": methodDistribution,
		"total_methods":       len(m.methods),
		"most_common_method":  m.getMostCommonMethod(),
		"least_common_method": m.getLeastCommonMethod(),
	}
}

// GetErrorStats returns error statistics
func (m *Metrics) GetErrorStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate error distribution
	errorDistribution := make(map[string]float64)
	for errorType, count := range m.errors {
		if m.totalRequests > 0 {
			errorDistribution[errorType] = float64(count) / float64(m.totalRequests) * 100
		}
	}

	return map[string]interface{}{
		"errors":             m.errors,
		"error_distribution": errorDistribution,
		"total_errors":       len(m.errors),
		"most_common_error":  m.getMostCommonError(),
		"least_common_error": m.getLeastCommonError(),
	}
}

// GetRetryStats returns retry statistics
func (m *Metrics) GetRetryStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate retry distribution
	retryDistribution := make(map[string]float64)
	for retryCount, count := range m.retryCounts {
		if m.totalRequests > 0 {
			retryDistribution[fmt.Sprintf("%d", retryCount)] = float64(count) / float64(m.totalRequests) * 100
		}
	}

	// Calculate average retries
	var avgRetries float64
	if m.totalRequests > 0 {
		avgRetries = float64(m.totalRetries) / float64(m.totalRequests)
	}

	return map[string]interface{}{
		"total_retries":      m.totalRetries,
		"retry_counts":       m.retryCounts,
		"retry_distribution": retryDistribution,
		"avg_retries":        avgRetries,
		"max_retries":        m.getMaxRetries(),
		"min_retries":        m.getMinRetries(),
	}
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (m *Metrics) GetCircuitBreakerStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"trips":  m.circuitBreakerTrips,
		"resets": m.circuitBreakerResets,
		"uptime": time.Since(m.startTime),
	}
}

// GetRateLimiterStats returns rate limiter statistics
func (m *Metrics) GetRateLimiterStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"hits":   m.rateLimitHits,
		"uptime": time.Since(m.startTime),
	}
}

// GetHealth returns health metrics
func (m *Metrics) GetHealth() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate health score
	var healthScore float64
	if m.totalRequests > 0 {
		healthScore = float64(m.successfulRequests) / float64(m.totalRequests) * 100
	}

	// Determine health status
	var healthStatus string
	if healthScore >= 95 {
		healthStatus = "healthy"
	} else if healthScore >= 80 {
		healthStatus = "degraded"
	} else {
		healthStatus = "unhealthy"
	}

	return map[string]interface{}{
		"health_score":   healthScore,
		"health_status":  healthStatus,
		"uptime":         time.Since(m.startTime),
		"total_requests": m.totalRequests,
		"success_rate":   healthScore,
		"failure_rate":   100 - healthScore,
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests = 0
	m.successfulRequests = 0
	m.failedRequests = 0
	m.totalResponseTime = 0
	m.minResponseTime = 0
	m.maxResponseTime = 0
	m.totalRetries = 0
	m.circuitBreakerTrips = 0
	m.circuitBreakerResets = 0
	m.rateLimitHits = 0

	m.statusCodes = make(map[int]int64)
	m.methods = make(map[HTTPMethod]int64)
	m.errors = make(map[string]int64)
	m.retryCounts = make(map[int]int64)

	m.startTime = time.Now()
	m.lastRequest = time.Time{}
}

// Helper methods
func (m *Metrics) getMostCommonStatusCode() int {
	var maxCount int64
	var mostCommon int

	for statusCode, count := range m.statusCodes {
		if count > maxCount {
			maxCount = count
			mostCommon = statusCode
		}
	}

	return mostCommon
}

func (m *Metrics) getLeastCommonStatusCode() int {
	var minCount int64 = -1
	var leastCommon int

	for statusCode, count := range m.statusCodes {
		if minCount == -1 || count < minCount {
			minCount = count
			leastCommon = statusCode
		}
	}

	return leastCommon
}

func (m *Metrics) getMostCommonMethod() HTTPMethod {
	var maxCount int64
	var mostCommon HTTPMethod

	for method, count := range m.methods {
		if count > maxCount {
			maxCount = count
			mostCommon = method
		}
	}

	return mostCommon
}

func (m *Metrics) getLeastCommonMethod() HTTPMethod {
	var minCount int64 = -1
	var leastCommon HTTPMethod

	for method, count := range m.methods {
		if minCount == -1 || count < minCount {
			minCount = count
			leastCommon = method
		}
	}

	return leastCommon
}

func (m *Metrics) getMostCommonError() string {
	var maxCount int64
	var mostCommon string

	for errorType, count := range m.errors {
		if count > maxCount {
			maxCount = count
			mostCommon = errorType
		}
	}

	return mostCommon
}

func (m *Metrics) getLeastCommonError() string {
	var minCount int64 = -1
	var leastCommon string

	for errorType, count := range m.errors {
		if minCount == -1 || count < minCount {
			minCount = count
			leastCommon = errorType
		}
	}

	return leastCommon
}

func (m *Metrics) getMaxRetries() int {
	var maxRetries int

	for retryCount := range m.retryCounts {
		if retryCount > maxRetries {
			maxRetries = retryCount
		}
	}

	return maxRetries
}

func (m *Metrics) getMinRetries() int {
	var minRetries int = -1

	for retryCount := range m.retryCounts {
		if minRetries == -1 || retryCount < minRetries {
			minRetries = retryCount
		}
	}

	return minRetries
}
