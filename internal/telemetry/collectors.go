package telemetry

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// SystemCollector collects system-level telemetry data
type SystemCollector struct {
	enabled bool
}

// NewSystemCollector creates a new system collector
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{
		enabled: true,
	}
}

// Collect gathers system information
func (sc *SystemCollector) Collect(ctx context.Context) (*TelemetryData, error) {
	if !sc.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	eventData := map[string]interface{}{
		"go_version":     runtime.Version(),
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
		"num_cpu":        runtime.NumCPU(),
		"num_goroutines": runtime.NumGoroutine(),
		"memory_alloc":   m.Alloc,
		"memory_total":   m.TotalAlloc,
		"gc_runs":        m.NumGC,
	}

	return &TelemetryData{
		SessionID:        sc.generateSessionID(),
		FrameworkVersion: "1.0.0",
		GoVersion:        runtime.Version(),
		OS:               runtime.GOOS,
		Architecture:     runtime.GOARCH,
		Timestamp:        time.Now(),
		EventType:        string(EventTypeStartup),
		EventData:        eventData,
	}, nil
}

// GetName returns the collector's name
func (sc *SystemCollector) GetName() string {
	return "system"
}

// IsEnabled checks if this collector is enabled
func (sc *SystemCollector) IsEnabled() bool {
	return sc.enabled
}

// SetEnabled enables or disables the collector
func (sc *SystemCollector) SetEnabled(enabled bool) {
	sc.enabled = enabled
}

// generateSessionID generates a unique session ID
func (sc *SystemCollector) generateSessionID() string {
	return fmt.Sprintf("system_%d", time.Now().UnixNano())
}

// PerformanceCollector collects performance-related telemetry data
type PerformanceCollector struct {
	enabled bool
	metrics map[string]interface{}
}

// NewPerformanceCollector creates a new performance collector
func NewPerformanceCollector() *PerformanceCollector {
	return &PerformanceCollector{
		enabled: true,
		metrics: make(map[string]interface{}),
	}
}

// Collect gathers performance metrics
func (pc *PerformanceCollector) Collect(ctx context.Context) (*TelemetryData, error) {
	if !pc.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	eventData := map[string]interface{}{
		"memory_alloc":   m.Alloc,
		"memory_sys":     m.Sys,
		"memory_heap":    m.HeapAlloc,
		"memory_stack":   m.StackInuse,
		"gc_pause_total": m.PauseTotalNs,
		"gc_runs":        m.NumGC,
		"goroutines":     runtime.NumGoroutine(),
		"timestamp":      time.Now().Unix(),
	}

	// Add custom metrics if any
	for key, value := range pc.metrics {
		eventData[key] = value
	}

	return &TelemetryData{
		SessionID:        pc.generateSessionID(),
		FrameworkVersion: "1.0.0",
		GoVersion:        runtime.Version(),
		OS:               runtime.GOOS,
		Architecture:     runtime.GOARCH,
		Timestamp:        time.Now(),
		EventType:        string(EventTypePerformance),
		EventData:        eventData,
	}, nil
}

// GetName returns the collector's name
func (pc *PerformanceCollector) GetName() string {
	return "performance"
}

// IsEnabled checks if this collector is enabled
func (pc *PerformanceCollector) IsEnabled() bool {
	return pc.enabled
}

// SetEnabled enables or disables the collector
func (pc *PerformanceCollector) SetEnabled(enabled bool) {
	pc.enabled = enabled
}

// AddMetric adds a custom performance metric
func (pc *PerformanceCollector) AddMetric(key string, value interface{}) {
	pc.metrics[key] = value
}

// generateSessionID generates a unique session ID
func (pc *PerformanceCollector) generateSessionID() string {
	return fmt.Sprintf("perf_%d", time.Now().UnixNano())
}

// ErrorCollector collects error-related telemetry data
type ErrorCollector struct {
	enabled bool
	errors  []ErrorInfo
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Error     string    `json:"error"`
	Stack     string    `json:"stack"`
	Timestamp time.Time `json:"timestamp"`
	Context   string    `json:"context"`
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		enabled: true,
		errors:  make([]ErrorInfo, 0),
	}
}

// Collect gathers error information
func (ec *ErrorCollector) Collect(ctx context.Context) (*TelemetryData, error) {
	if !ec.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	eventData := map[string]interface{}{
		"error_count": len(ec.errors),
		"errors":      ec.errors,
	}

	return &TelemetryData{
		SessionID:        ec.generateSessionID(),
		FrameworkVersion: "1.0.0",
		GoVersion:        runtime.Version(),
		OS:               runtime.GOOS,
		Architecture:     runtime.GOARCH,
		Timestamp:        time.Now(),
		EventType:        string(EventTypeError),
		EventData:        eventData,
	}, nil
}

// GetName returns the collector's name
func (ec *ErrorCollector) GetName() string {
	return "errors"
}

// IsEnabled checks if this collector is enabled
func (ec *ErrorCollector) IsEnabled() bool {
	return ec.enabled
}

// SetEnabled enables or disables the collector
func (ec *ErrorCollector) SetEnabled(enabled bool) {
	ec.enabled = enabled
}

// AddError adds an error to the collector
func (ec *ErrorCollector) AddError(err error, context string) {
	errorInfo := ErrorInfo{
		Error:     err.Error(),
		Stack:     fmt.Sprintf("%+v", err),
		Timestamp: time.Now(),
		Context:   context,
	}

	ec.errors = append(ec.errors, errorInfo)

	// Keep only last 100 errors to prevent memory issues
	if len(ec.errors) > 100 {
		ec.errors = ec.errors[1:]
	}
}

// generateSessionID generates a unique session ID
func (ec *ErrorCollector) generateSessionID() string {
	return fmt.Sprintf("error_%d", time.Now().UnixNano())
}

// FeatureCollector collects feature usage telemetry data
type FeatureCollector struct {
	enabled  bool
	features map[string]int
}

// NewFeatureCollector creates a new feature collector
func NewFeatureCollector() *FeatureCollector {
	return &FeatureCollector{
		enabled:  true,
		features: make(map[string]int),
	}
}

// Collect gathers feature usage information
func (fc *FeatureCollector) Collect(ctx context.Context) (*TelemetryData, error) {
	if !fc.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	eventData := map[string]interface{}{
		"features":       fc.features,
		"total_features": len(fc.features),
	}

	return &TelemetryData{
		SessionID:        fc.generateSessionID(),
		FrameworkVersion: "1.0.0",
		GoVersion:        runtime.Version(),
		OS:               runtime.GOOS,
		Architecture:     runtime.GOARCH,
		Timestamp:        time.Now(),
		EventType:        string(EventTypeFeature),
		EventData:        eventData,
	}, nil
}

// GetName returns the collector's name
func (fc *FeatureCollector) GetName() string {
	return "features"
}

// IsEnabled checks if this collector is enabled
func (fc *FeatureCollector) IsEnabled() bool {
	return fc.enabled
}

// SetEnabled enables or disables the collector
func (fc *FeatureCollector) SetEnabled(enabled bool) {
	fc.enabled = enabled
}

// TrackFeature tracks feature usage
func (fc *FeatureCollector) TrackFeature(feature string) {
	fc.features[feature]++
}

// generateSessionID generates a unique session ID
func (fc *FeatureCollector) generateSessionID() string {
	return fmt.Sprintf("feature_%d", time.Now().UnixNano())
}
