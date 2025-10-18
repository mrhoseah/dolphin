package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FileStorage implements Storage interface using file system
type FileStorage struct {
	configPath string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(configPath string) *FileStorage {
	return &FileStorage{
		configPath: configPath,
	}
}

// IsEnabled checks if telemetry is enabled
func (fs *FileStorage) IsEnabled() bool {
	config, err := fs.GetConfig()
	if err != nil {
		return false
	}
	return config.Enabled
}

// SetEnabled enables or disables telemetry
func (fs *FileStorage) SetEnabled(enabled bool) error {
	config, err := fs.GetConfig()
	if err != nil {
		config = DefaultConfig()
	}

	config.Enabled = enabled
	return fs.SetConfig(config)
}

// GetConfig returns the telemetry configuration
func (fs *FileStorage) GetConfig() (*Config, error) {
	data, err := os.ReadFile(fs.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SetConfig sets the telemetry configuration
func (fs *FileStorage) SetConfig(config *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(fs.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.configPath, data, 0644)
}

// HTTPSender implements Sender interface using HTTP
type HTTPSender struct {
	endpoint string
	client   *http.Client
}

// NewHTTPSender creates a new HTTP sender
func NewHTTPSender(endpoint string) *HTTPSender {
	return &HTTPSender{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send transmits telemetry data to the endpoint
func (hs *HTTPSender) Send(ctx context.Context, data *TelemetryData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", hs.endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Dolphin-Framework-Telemetry/1.0.0")

	resp, err := hs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telemetry send failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetEndpoint returns the sender's endpoint
func (hs *HTTPSender) GetEndpoint() string {
	return hs.endpoint
}

// SetEndpoint sets the sender's endpoint
func (hs *HTTPSender) SetEndpoint(endpoint string) {
	hs.endpoint = endpoint
}

// NoOpSender implements Sender interface as a no-op (for testing or disabled mode)
type NoOpSender struct {
	endpoint string
}

// NewNoOpSender creates a new no-op sender
func NewNoOpSender(endpoint string) *NoOpSender {
	return &NoOpSender{
		endpoint: endpoint,
	}
}

// Send does nothing (no-op)
func (ns *NoOpSender) Send(ctx context.Context, data *TelemetryData) error {
	// No-op implementation
	return nil
}

// GetEndpoint returns the sender's endpoint
func (ns *NoOpSender) GetEndpoint() string {
	return ns.endpoint
}

// SetEndpoint sets the sender's endpoint
func (ns *NoOpSender) SetEndpoint(endpoint string) {
	ns.endpoint = endpoint
}

// ConsoleObserver implements Observer interface for console output
type ConsoleObserver struct {
	name string
}

// NewConsoleObserver creates a new console observer
func NewConsoleObserver(name string) *ConsoleObserver {
	return &ConsoleObserver{
		name: name,
	}
}

// OnTelemetryEvent outputs telemetry events to console
func (co *ConsoleObserver) OnTelemetryEvent(ctx context.Context, data *TelemetryData) {
	fmt.Printf("[TELEMETRY] %s: %s - %s\n",
		co.name,
		data.EventType,
		data.Timestamp.Format(time.RFC3339))
}

// GetName returns the observer's name
func (co *ConsoleObserver) GetName() string {
	return co.name
}

// LogObserver implements Observer interface for logging
type LogObserver struct {
	name string
}

// NewLogObserver creates a new log observer
func NewLogObserver(name string) *LogObserver {
	return &LogObserver{
		name: name,
	}
}

// OnTelemetryEvent logs telemetry events
func (lo *LogObserver) OnTelemetryEvent(ctx context.Context, data *TelemetryData) {
	// This would integrate with your logging system
	// For now, we'll just print to console
	fmt.Printf("[LOG] Telemetry event: %s\n", data.EventType)
}

// GetName returns the observer's name
func (lo *LogObserver) GetName() string {
	return lo.name
}
