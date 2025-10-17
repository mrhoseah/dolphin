package circuitbreaker

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// Metrics represents circuit breaker metrics
type Metrics struct {
	name string

	// Prometheus metrics
	requestsTotal    prometheus.Counter
	requestsSuccess  prometheus.Counter
	requestsFailure  prometheus.Counter
	requestsRejected prometheus.Counter
	stateChanges     prometheus.Counter
	stateGauge       prometheus.Gauge
	failureRate      prometheus.Gauge
	successRate      prometheus.Gauge

	// Internal counters
	requestCount     int64
	successCount     int64
	failureCount     int64
	rejectedCount    int64
	stateChangeCount int64

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewMetrics creates new circuit breaker metrics
func NewMetrics(name string) *Metrics {
	labels := prometheus.Labels{"circuit": name}

	return &Metrics{
		name: name,

		requestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "circuit_breaker_requests_total",
			Help:        "Total number of requests through the circuit breaker",
			ConstLabels: labels,
		}),

		requestsSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "circuit_breaker_requests_success_total",
			Help:        "Total number of successful requests through the circuit breaker",
			ConstLabels: labels,
		}),

		requestsFailure: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "circuit_breaker_requests_failure_total",
			Help:        "Total number of failed requests through the circuit breaker",
			ConstLabels: labels,
		}),

		requestsRejected: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "circuit_breaker_requests_rejected_total",
			Help:        "Total number of rejected requests by the circuit breaker",
			ConstLabels: labels,
		}),

		stateChanges: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "circuit_breaker_state_changes_total",
			Help:        "Total number of state changes in the circuit breaker",
			ConstLabels: labels,
		}),

		stateGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "circuit_breaker_state",
			Help:        "Current state of the circuit breaker (0=Closed, 1=Open, 2=HalfOpen)",
			ConstLabels: labels,
		}),

		failureRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "circuit_breaker_failure_rate",
			Help:        "Current failure rate of the circuit breaker (percentage)",
			ConstLabels: labels,
		}),

		successRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "circuit_breaker_success_rate",
			Help:        "Current success rate of the circuit breaker (percentage)",
			ConstLabels: labels,
		}),
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(isFailure bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount++
	m.requestsTotal.Inc()

	if isFailure {
		m.failureCount++
		m.requestsFailure.Inc()
	} else {
		m.successCount++
		m.requestsSuccess.Inc()
	}

	// Update rates
	m.updateRates()
}

// RecordRejected records a rejected request
func (m *Metrics) RecordRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rejectedCount++
	m.requestsRejected.Inc()
}

// RecordStateChange records a state change
func (m *Metrics) RecordStateChange(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stateChangeCount++
	m.stateChanges.Inc()
	m.stateGauge.Set(float64(state))
}

// UpdateRates updates the success and failure rates
func (m *Metrics) UpdateRates() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateRates()
}

// updateRates updates the rates (internal method)
func (m *Metrics) updateRates() {
	if m.requestCount > 0 {
		failureRate := float64(m.failureCount) / float64(m.requestCount) * 100.0
		successRate := float64(m.successCount) / float64(m.requestCount) * 100.0

		m.failureRate.Set(failureRate)
		m.successRate.Set(successRate)
	}
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() MetricsStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsStats{
		RequestCount:     m.requestCount,
		SuccessCount:     m.successCount,
		FailureCount:     m.failureCount,
		RejectedCount:    m.rejectedCount,
		StateChangeCount: m.stateChangeCount,
		FailureRate:      m.calculateFailureRate(),
		SuccessRate:      m.calculateSuccessRate(),
	}
}

// calculateFailureRate calculates the failure rate
func (m *Metrics) calculateFailureRate() float64 {
	if m.requestCount == 0 {
		return 0.0
	}
	return float64(m.failureCount) / float64(m.requestCount) * 100.0
}

// calculateSuccessRate calculates the success rate
func (m *Metrics) calculateSuccessRate() float64 {
	if m.requestCount == 0 {
		return 0.0
	}
	return float64(m.successCount) / float64(m.requestCount) * 100.0
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount = 0
	m.successCount = 0
	m.failureCount = 0
	m.rejectedCount = 0
	m.stateChangeCount = 0

	// Reset Prometheus metrics
	m.requestsTotal.Add(-m.requestsTotal.Get())
	m.requestsSuccess.Add(-m.requestsSuccess.Get())
	m.requestsFailure.Add(-m.requestsFailure.Get())
	m.requestsRejected.Add(-m.requestsRejected.Get())
	m.stateChanges.Add(-m.stateChanges.Get())
	m.stateGauge.Set(0)
	m.failureRate.Set(0)
	m.successRate.Set(0)
}

// MetricsStats represents metrics statistics
type MetricsStats struct {
	RequestCount     int64   `json:"request_count"`
	SuccessCount     int64   `json:"success_count"`
	FailureCount     int64   `json:"failure_count"`
	RejectedCount    int64   `json:"rejected_count"`
	StateChangeCount int64   `json:"state_change_count"`
	FailureRate      float64 `json:"failure_rate"`
	SuccessRate      float64 `json:"success_rate"`
}

// MetricsCollector collects metrics from multiple circuit breakers
type MetricsCollector struct {
	circuits map[string]*Metrics
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		circuits: make(map[string]*Metrics),
		logger:   logger,
	}
}

// RegisterCircuit registers a circuit breaker for metrics collection
func (mc *MetricsCollector) RegisterCircuit(name string, metrics *Metrics) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.circuits[name] = metrics

	if mc.logger != nil {
		mc.logger.Info("Circuit breaker registered for metrics",
			zap.String("circuit", name))
	}
}

// UnregisterCircuit unregisters a circuit breaker
func (mc *MetricsCollector) UnregisterCircuit(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.circuits, name)

	if mc.logger != nil {
		mc.logger.Info("Circuit breaker unregistered from metrics",
			zap.String("circuit", name))
	}
}

// GetCircuitMetrics returns metrics for a specific circuit
func (mc *MetricsCollector) GetCircuitMetrics(name string) (*Metrics, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.circuits[name]
	return metrics, exists
}

// GetAllMetrics returns metrics for all circuits
func (mc *MetricsCollector) GetAllMetrics() map[string]*Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*Metrics)
	for name, metrics := range mc.circuits {
		result[name] = metrics
	}
	return result
}

// GetAggregatedStats returns aggregated statistics for all circuits
func (mc *MetricsCollector) GetAggregatedStats() AggregatedStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalRequests, totalSuccess, totalFailure, totalRejected, totalStateChanges int64
	var totalFailureRate, totalSuccessRate float64
	circuitCount := len(mc.circuits)

	for _, metrics := range mc.circuits {
		stats := metrics.GetStats()
		totalRequests += stats.RequestCount
		totalSuccess += stats.SuccessCount
		totalFailure += stats.FailureCount
		totalRejected += stats.RejectedCount
		totalStateChanges += stats.StateChangeCount
		totalFailureRate += stats.FailureRate
		totalSuccessRate += stats.SuccessRate
	}

	avgFailureRate := 0.0
	avgSuccessRate := 0.0
	if circuitCount > 0 {
		avgFailureRate = totalFailureRate / float64(circuitCount)
		avgSuccessRate = totalSuccessRate / float64(circuitCount)
	}

	return AggregatedStats{
		CircuitCount:      circuitCount,
		TotalRequests:     totalRequests,
		TotalSuccess:      totalSuccess,
		TotalFailure:      totalFailure,
		TotalRejected:     totalRejected,
		TotalStateChanges: totalStateChanges,
		AvgFailureRate:    avgFailureRate,
		AvgSuccessRate:    avgSuccessRate,
	}
}

// ResetAll resets metrics for all circuits
func (mc *MetricsCollector) ResetAll() {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, metrics := range mc.circuits {
		metrics.Reset()
	}

	if mc.logger != nil {
		mc.logger.Info("All circuit breaker metrics reset")
	}
}

// AggregatedStats represents aggregated statistics for all circuits
type AggregatedStats struct {
	CircuitCount      int     `json:"circuit_count"`
	TotalRequests     int64   `json:"total_requests"`
	TotalSuccess      int64   `json:"total_success"`
	TotalFailure      int64   `json:"total_failure"`
	TotalRejected     int64   `json:"total_rejected"`
	TotalStateChanges int64   `json:"total_state_changes"`
	AvgFailureRate    float64 `json:"avg_failure_rate"`
	AvgSuccessRate    float64 `json:"avg_success_rate"`
}
