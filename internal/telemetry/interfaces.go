package telemetry

import (
	"context"
	"time"
)

// TelemetryData represents the core data structure for telemetry
type TelemetryData struct {
	SessionID        string                 `json:"session_id"`
	FrameworkVersion string                 `json:"framework_version"`
	GoVersion        string                 `json:"go_version"`
	OS               string                 `json:"os"`
	Architecture     string                 `json:"arch"`
	Timestamp        time.Time              `json:"timestamp"`
	EventType        string                 `json:"event_type"`
	EventData        map[string]interface{} `json:"event_data"`
	UserAgent        string                 `json:"user_agent,omitempty"`
	IPHash           string                 `json:"ip_hash,omitempty"` // Hashed IP for privacy
}

// Collector defines the interface for different types of telemetry collectors
type Collector interface {
	// Collect gathers telemetry data
	Collect(ctx context.Context) (*TelemetryData, error)

	// GetName returns the collector's name
	GetName() string

	// IsEnabled checks if this collector is enabled
	IsEnabled() bool

	// SetEnabled enables or disables the collector
	SetEnabled(enabled bool)
}

// Sender defines the interface for sending telemetry data
type Sender interface {
	// Send transmits telemetry data to the endpoint
	Send(ctx context.Context, data *TelemetryData) error

	// GetEndpoint returns the sender's endpoint
	GetEndpoint() string

	// SetEndpoint sets the sender's endpoint
	SetEndpoint(endpoint string)
}

// Storage defines the interface for storing telemetry configuration
type Storage interface {
	// IsEnabled checks if telemetry is enabled globally
	IsEnabled() bool

	// SetEnabled enables or disables telemetry globally
	SetEnabled(enabled bool) error

	// GetConfig returns the telemetry configuration
	GetConfig() (*Config, error)

	// SetConfig sets the telemetry configuration
	SetConfig(config *Config) error
}

// Observer defines the interface for telemetry observers (Observer pattern)
type Observer interface {
	// OnTelemetryEvent is called when a telemetry event occurs
	OnTelemetryEvent(ctx context.Context, data *TelemetryData)

	// GetName returns the observer's name
	GetName() string
}

// Config represents the telemetry configuration
type Config struct {
	Enabled       bool            `json:"enabled"`
	Endpoint      string          `json:"endpoint"`
	BatchSize     int             `json:"batch_size"`
	FlushInterval time.Duration   `json:"flush_interval"`
	RetryAttempts int             `json:"retry_attempts"`
	Timeout       time.Duration   `json:"timeout"`
	Collectors    map[string]bool `json:"collectors"`
	PrivacyMode   bool            `json:"privacy_mode"`
	DataRetention time.Duration   `json:"data_retention"`
}

// EventType represents different types of telemetry events
type EventType string

const (
	EventTypeStartup     EventType = "startup"
	EventTypeShutdown    EventType = "shutdown"
	EventTypeCommand     EventType = "command"
	EventTypeError       EventType = "error"
	EventTypePerformance EventType = "performance"
	EventTypeFeature     EventType = "feature"
	EventTypeCustom      EventType = "custom"
)

// DefaultConfig returns the default telemetry configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       false, // Opt-in by default
		Endpoint:      "https://telemetry.dolphin-framework.dev/api/v1/events",
		BatchSize:     100,
		FlushInterval: 5 * time.Minute,
		RetryAttempts: 3,
		Timeout:       30 * time.Second,
		Collectors: map[string]bool{
			"system":      true,
			"performance": true,
			"errors":      true,
			"features":    true,
		},
		PrivacyMode:   true,
		DataRetention: 30 * 24 * time.Hour, // 30 days
	}
}
