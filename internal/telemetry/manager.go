package telemetry

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TelemetryManager manages telemetry collection and sending
type TelemetryManager struct {
	config     *Config
	storage    Storage
	sender     Sender
	collectors map[string]Collector
	observers  []Observer
	buffer     []*TelemetryData
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewTelemetryManager creates a new telemetry manager
func NewTelemetryManager(storage Storage, sender Sender) *TelemetryManager {
	ctx, cancel := context.WithCancel(context.Background())

	tm := &TelemetryManager{
		config:     DefaultConfig(),
		storage:    storage,
		sender:     sender,
		collectors: make(map[string]Collector),
		observers:  make([]Observer, 0),
		buffer:     make([]*TelemetryData, 0),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Load configuration from storage
	if config, err := storage.GetConfig(); err == nil {
		tm.config = config
	}

	return tm
}

// Start initializes the telemetry manager
func (tm *TelemetryManager) Start() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.config.Enabled {
		return nil // Telemetry is disabled
	}

	// Start the flush routine
	tm.wg.Add(1)
	go tm.flushRoutine()

	return nil
}

// Stop stops the telemetry manager
func (tm *TelemetryManager) Stop() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Cancel context to stop routines
	tm.cancel()

	// Flush remaining data
	if err := tm.flush(); err != nil {
		return fmt.Errorf("failed to flush telemetry data: %w", err)
	}

	// Wait for routines to finish
	tm.wg.Wait()

	return nil
}

// Enable enables telemetry collection
func (tm *TelemetryManager) Enable() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.config.Enabled = true

	if err := tm.storage.SetEnabled(true); err != nil {
		return fmt.Errorf("failed to enable telemetry: %w", err)
	}

	// Start if not already started
	go tm.Start()

	return nil
}

// Disable disables telemetry collection
func (tm *TelemetryManager) Disable() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.config.Enabled = false

	if err := tm.storage.SetEnabled(false); err != nil {
		return fmt.Errorf("failed to disable telemetry: %w", err)
	}

	// Stop the manager
	tm.cancel()

	return nil
}

// IsEnabled returns whether telemetry is enabled
func (tm *TelemetryManager) IsEnabled() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.config.Enabled
}

// AddCollector adds a telemetry collector
func (tm *TelemetryManager) AddCollector(name string, collector Collector) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.collectors[name] = collector
}

// RemoveCollector removes a telemetry collector
func (tm *TelemetryManager) RemoveCollector(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.collectors, name)
}

// AddObserver adds a telemetry observer
func (tm *TelemetryManager) AddObserver(observer Observer) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.observers = append(tm.observers, observer)
}

// RemoveObserver removes a telemetry observer
func (tm *TelemetryManager) RemoveObserver(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, observer := range tm.observers {
		if observer.GetName() == name {
			tm.observers = append(tm.observers[:i], tm.observers[i+1:]...)
			break
		}
	}
}

// CollectEvent collects a telemetry event
func (tm *TelemetryManager) CollectEvent(ctx context.Context, eventType EventType, eventData map[string]interface{}) error {
	tm.mu.RLock()
	enabled := tm.config.Enabled
	tm.mu.RUnlock()

	if !enabled {
		return nil // Telemetry is disabled
	}

	// Create telemetry data
	data := &TelemetryData{
		SessionID:        tm.generateSessionID(),
		FrameworkVersion: "1.0.0",  // This should come from version package
		GoVersion:        "1.21.0", // This should come from runtime
		OS:               "linux",  // This should come from runtime
		Architecture:     "amd64",  // This should come from runtime
		Timestamp:        time.Now(),
		EventType:        string(eventType),
		EventData:        eventData,
	}

	// Add to buffer
	tm.mu.Lock()
	tm.buffer = append(tm.buffer, data)
	tm.mu.Unlock()

	// Notify observers
	tm.notifyObservers(ctx, data)

	return nil
}

// CollectFromCollectors collects data from all enabled collectors
func (tm *TelemetryManager) CollectFromCollectors(ctx context.Context) error {
	tm.mu.RLock()
	enabled := tm.config.Enabled
	collectors := make(map[string]Collector)
	for name, collector := range tm.collectors {
		collectors[name] = collector
	}
	tm.mu.RUnlock()

	if !enabled {
		return nil
	}

	for _, collector := range collectors {
		if !collector.IsEnabled() {
			continue
		}

		data, err := collector.Collect(ctx)
		if err != nil {
			continue // Skip failed collectors
		}

		tm.mu.Lock()
		tm.buffer = append(tm.buffer, data)
		tm.mu.Unlock()

		// Notify observers
		tm.notifyObservers(ctx, data)
	}

	return nil
}

// GetConfig returns the current configuration
func (tm *TelemetryManager) GetConfig() *Config {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Return a copy to prevent external modification
	config := *tm.config
	return &config
}

// SetConfig updates the configuration
func (tm *TelemetryManager) SetConfig(config *Config) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.config = config

	if err := tm.storage.SetConfig(config); err != nil {
		return fmt.Errorf("failed to save telemetry config: %w", err)
	}

	return nil
}

// flushRoutine periodically flushes telemetry data
func (tm *TelemetryManager) flushRoutine() {
	defer tm.wg.Done()

	ticker := time.NewTicker(tm.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			if err := tm.flush(); err != nil {
				// Log error but don't fail
				continue
			}
		}
	}
}

// flush sends buffered telemetry data
func (tm *TelemetryManager) flush() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if len(tm.buffer) == 0 {
		return nil
	}

	// Send data in batches
	batchSize := tm.config.BatchSize
	for i := 0; i < len(tm.buffer); i += batchSize {
		end := i + batchSize
		if end > len(tm.buffer) {
			end = len(tm.buffer)
		}

		batch := tm.buffer[i:end]

		// Send batch
		for _, data := range batch {
			ctx, cancel := context.WithTimeout(tm.ctx, tm.config.Timeout)
			if err := tm.sender.Send(ctx, data); err != nil {
				cancel()
				continue // Skip failed sends
			}
			cancel()
		}
	}

	// Clear buffer
	tm.buffer = tm.buffer[:0]

	return nil
}

// notifyObservers notifies all observers of a telemetry event
func (tm *TelemetryManager) notifyObservers(ctx context.Context, data *TelemetryData) {
	for _, observer := range tm.observers {
		go observer.OnTelemetryEvent(ctx, data)
	}
}

// generateSessionID generates a unique session ID
func (tm *TelemetryManager) generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
