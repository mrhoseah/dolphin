package loadshedding

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// Metrics represents load shedding metrics
type Metrics struct {
	name string

	// Prometheus metrics
	requestsTotal     prometheus.Counter
	requestsShed      prometheus.Counter
	requestsProcessed prometheus.Counter
	sheddingLevel     prometheus.Gauge
	sheddingRate      prometheus.Gauge
	cpuUsage          prometheus.Gauge
	memoryUsage       prometheus.Gauge
	goroutines        prometheus.Gauge
	requestRate       prometheus.Gauge
	responseTime      prometheus.Gauge

	// Internal counters
	requestCount    int64
	shedCount       int64
	processedCount  int64
	lastRequestTime time.Time
	responseTimes   []time.Duration
	responseTimeMu  sync.RWMutex

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewMetrics creates new load shedding metrics
func NewMetrics(name string) *Metrics {
	labels := prometheus.Labels{"shedder": name}

	return &Metrics{
		name: name,

		requestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "load_shedder_requests_total",
			Help:        "Total number of requests to the load shedder",
			ConstLabels: labels,
		}),

		requestsShed: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "load_shedder_requests_shed_total",
			Help:        "Total number of requests shed by the load shedder",
			ConstLabels: labels,
		}),

		requestsProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "load_shedder_requests_processed_total",
			Help:        "Total number of requests processed by the load shedder",
			ConstLabels: labels,
		}),

		sheddingLevel: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "load_shedder_level",
			Help:        "Current shedding level (0=None, 1=Light, 2=Moderate, 3=Heavy, 4=Critical)",
			ConstLabels: labels,
		}),

		sheddingRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "load_shedder_rate",
			Help:        "Current shedding rate (0.0 to 1.0)",
			ConstLabels: labels,
		}),

		cpuUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "load_shedder_cpu_usage",
			Help:        "Current CPU usage percentage",
			ConstLabels: labels,
		}),

		memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "load_shedder_memory_usage",
			Help:        "Current memory usage percentage",
			ConstLabels: labels,
		}),

		goroutines: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "load_shedder_goroutines",
			Help:        "Current number of goroutines",
			ConstLabels: labels,
		}),

		requestRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "load_shedder_request_rate",
			Help:        "Current request rate (requests per second)",
			ConstLabels: labels,
		}),

		responseTime: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "load_shedder_response_time_seconds",
			Help:        "Average response time in seconds",
			ConstLabels: labels,
		}),

		responseTimes: make([]time.Duration, 0, 1000), // Keep last 1000 response times
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(shed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount++
	m.lastRequestTime = time.Now()
	m.requestsTotal.Inc()

	if shed {
		m.shedCount++
		m.requestsShed.Inc()
	} else {
		m.processedCount++
		m.requestsProcessed.Inc()
	}
}

// RecordResponseTime records a response time
func (m *Metrics) RecordResponseTime(duration time.Duration) {
	m.responseTimeMu.Lock()
	defer m.responseTimeMu.Unlock()

	// Keep only the last 1000 response times
	if len(m.responseTimes) >= 1000 {
		m.responseTimes = m.responseTimes[1:]
	}
	m.responseTimes = append(m.responseTimes, duration)

	// Update Prometheus metric
	m.responseTime.Set(duration.Seconds())
}

// UpdateSystemMetrics updates system metrics
func (m *Metrics) UpdateSystemMetrics(cpuUsage, memoryUsage, requestRate float64, goroutines int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cpuUsage.Set(cpuUsage)
	m.memoryUsage.Set(memoryUsage)
	m.requestRate.Set(requestRate)
	m.goroutines.Set(float64(goroutines))
}

// UpdateSheddingMetrics updates shedding metrics
func (m *Metrics) UpdateSheddingMetrics(level SheddingLevel, rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sheddingLevel.Set(float64(level))
	m.sheddingRate.Set(rate)
}

// GetRequestRate returns the current request rate
func (m *Metrics) GetRequestRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.requestCount == 0 {
		return 0.0
	}

	// Calculate rate based on time since first request
	now := time.Now()
	if m.lastRequestTime.IsZero() {
		return 0.0
	}

	elapsed := now.Sub(m.lastRequestTime).Seconds()
	if elapsed == 0 {
		return 0.0
	}

	return float64(m.requestCount) / elapsed
}

// GetShedRate returns the current shedding rate
func (m *Metrics) GetShedRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.requestCount == 0 {
		return 0.0
	}

	return float64(m.shedCount) / float64(m.requestCount)
}

// GetAverageResponseTime returns the average response time
func (m *Metrics) GetAverageResponseTime() time.Duration {
	m.responseTimeMu.RLock()
	defer m.responseTimeMu.RUnlock()

	if len(m.responseTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, rt := range m.responseTimes {
		total += rt
	}

	return total / time.Duration(len(m.responseTimes))
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() MetricsStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.responseTimeMu.RLock()
	avgResponseTime := m.GetAverageResponseTime()
	m.responseTimeMu.RUnlock()

	return MetricsStats{
		RequestCount:        m.requestCount,
		ShedCount:           m.shedCount,
		ProcessedCount:      m.processedCount,
		ShedRate:            m.GetShedRate(),
		RequestRate:         m.GetRequestRate(),
		AverageResponseTime: avgResponseTime,
		LastRequestTime:     m.lastRequestTime,
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount = 0
	m.shedCount = 0
	m.processedCount = 0
	m.lastRequestTime = time.Time{}

	m.responseTimeMu.Lock()
	m.responseTimes = m.responseTimes[:0]
	m.responseTimeMu.Unlock()

	// Reset Prometheus metrics
	// Note: Prometheus counters cannot be reset directly
	// They are cumulative and reset on application restart
	m.sheddingLevel.Set(0)
	m.sheddingRate.Set(0)
	m.cpuUsage.Set(0)
	m.memoryUsage.Set(0)
	m.goroutines.Set(0)
	m.requestRate.Set(0)
	m.responseTime.Set(0)
}

// MetricsStats represents metrics statistics
type MetricsStats struct {
	RequestCount        int64         `json:"request_count"`
	ShedCount           int64         `json:"shed_count"`
	ProcessedCount      int64         `json:"processed_count"`
	ShedRate            float64       `json:"shed_rate"`
	RequestRate         float64       `json:"request_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastRequestTime     time.Time     `json:"last_request_time"`
}

// MetricsCollector collects metrics from multiple load shedders
type MetricsCollector struct {
	shedders map[string]*Metrics
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		shedders: make(map[string]*Metrics),
		logger:   logger,
	}
}

// RegisterShedder registers a load shedder for metrics collection
func (mc *MetricsCollector) RegisterShedder(name string, metrics *Metrics) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.shedders[name] = metrics

	if mc.logger != nil {
		mc.logger.Info("Load shedder registered for metrics",
			zap.String("shedder", name))
	}
}

// UnregisterShedder unregisters a load shedder
func (mc *MetricsCollector) UnregisterShedder(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.shedders, name)

	if mc.logger != nil {
		mc.logger.Info("Load shedder unregistered from metrics",
			zap.String("shedder", name))
	}
}

// GetShedderMetrics returns metrics for a specific shedder
func (mc *MetricsCollector) GetShedderMetrics(name string) (*Metrics, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.shedders[name]
	return metrics, exists
}

// GetAllMetrics returns metrics for all shedders
func (mc *MetricsCollector) GetAllMetrics() map[string]*Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*Metrics)
	for name, metrics := range mc.shedders {
		result[name] = metrics
	}
	return result
}

// GetAggregatedStats returns aggregated statistics for all shedders
func (mc *MetricsCollector) GetAggregatedStats() AggregatedStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalRequests, totalShed, totalProcessed int64
	var totalShedRate, totalRequestRate float64
	shedderCount := len(mc.shedders)

	for _, metrics := range mc.shedders {
		stats := metrics.GetStats()
		totalRequests += stats.RequestCount
		totalShed += stats.ShedCount
		totalProcessed += stats.ProcessedCount
		totalShedRate += stats.ShedRate
		totalRequestRate += stats.RequestRate
	}

	avgShedRate := 0.0
	avgRequestRate := 0.0
	if shedderCount > 0 {
		avgShedRate = totalShedRate / float64(shedderCount)
		avgRequestRate = totalRequestRate / float64(shedderCount)
	}

	return AggregatedStats{
		ShedderCount:   shedderCount,
		TotalRequests:  totalRequests,
		TotalShed:      totalShed,
		TotalProcessed: totalProcessed,
		AvgShedRate:    avgShedRate,
		AvgRequestRate: avgRequestRate,
	}
}

// ResetAll resets metrics for all shedders
func (mc *MetricsCollector) ResetAll() {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for _, metrics := range mc.shedders {
		metrics.Reset()
	}

	if mc.logger != nil {
		mc.logger.Info("All load shedder metrics reset")
	}
}

// AggregatedStats represents aggregated statistics for all shedders
type AggregatedStats struct {
	ShedderCount   int     `json:"shedder_count"`
	TotalRequests  int64   `json:"total_requests"`
	TotalShed      int64   `json:"total_shed"`
	TotalProcessed int64   `json:"total_processed"`
	AvgShedRate    float64 `json:"avg_shed_rate"`
	AvgRequestRate float64 `json:"avg_request_rate"`
}
