package webhooks

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Event represents a webhook event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
}

// Webhook represents a webhook endpoint
type Webhook struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Events      []string          `json:"events"`      // Event types to listen for
	Secret      string            `json:"secret"`      // Secret for signature verification
	Headers     map[string]string `json:"headers"`     // Additional headers
	Enabled     bool              `json:"enabled"`
	Retries     int               `json:"retries"`     // Number of retries on failure
	Timeout     time.Duration     `json:"timeout"`     // Request timeout
	LastTrigger *time.Time        `json:"last_trigger"`
	SuccessCount int              `json:"success_count"`
	FailureCount int              `json:"failure_count"`
}

// Manager manages webhooks
type Manager struct {
	webhooks map[string]*Webhook
	events   chan Event
	mu       sync.RWMutex
	logger   *zap.Logger
	client   *http.Client
	workers  int
}

// NewManager creates a new webhook manager
func NewManager(logger *zap.Logger, workers int) *Manager {
	if workers <= 0 {
		workers = 5 // Default to 5 workers
	}

	return &Manager{
		webhooks: make(map[string]*Webhook),
		events:   make(chan Event, 100), // Buffer 100 events
		logger:   logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		workers: workers,
	}
}

// Start starts the webhook manager workers
func (m *Manager) Start() {
	for i := 0; i < m.workers; i++ {
		go m.worker()
	}
	m.logger.Info("Webhook manager started", zap.Int("workers", m.workers))
}

// Stop stops the webhook manager
func (m *Manager) Stop() {
	close(m.events)
	m.logger.Info("Webhook manager stopped")
}

// Register registers a new webhook
func (m *Manager) Register(webhook *Webhook) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if webhook.ID == "" {
		webhook.ID = fmt.Sprintf("wh_%d", time.Now().UnixNano())
	}

	if _, exists := m.webhooks[webhook.ID]; exists {
		return fmt.Errorf("webhook '%s' already exists", webhook.ID)
	}

	// Set defaults
	if webhook.Timeout == 0 {
		webhook.Timeout = 30 * time.Second
	}
	if webhook.Retries == 0 {
		webhook.Retries = 3
	}

	m.webhooks[webhook.ID] = webhook
	m.logger.Info("Webhook registered", zap.String("id", webhook.ID), zap.String("url", webhook.URL))

	return nil
}

// Dispatch dispatches an event to all matching webhooks
func (m *Manager) Dispatch(eventType string, data map[string]interface{}) {
	event := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Source:    "dolphin",
	}

	select {
	case m.events <- event:
	default:
		m.logger.Warn("Event queue full, dropping event", zap.String("type", eventType))
	}
}

// worker processes webhook events
func (m *Manager) worker() {
	for event := range m.events {
		m.processEvent(event)
	}
}

// processEvent processes a single event
func (m *Manager) processEvent(event Event) {
	m.mu.RLock()
	webhooks := make([]*Webhook, 0)
	for _, webhook := range m.webhooks {
		if !webhook.Enabled {
			continue
		}

		// Check if webhook listens to this event type
		matches := false
		for _, eventType := range webhook.Events {
			if eventType == "*" || eventType == event.Type {
				matches = true
				break
			}
		}

		if matches {
			webhooks = append(webhooks, webhook)
		}
	}
	m.mu.RUnlock()

	// Dispatch to all matching webhooks
	for _, webhook := range webhooks {
		m.dispatchToWebhook(webhook, event)
	}
}

// dispatchToWebhook dispatches an event to a specific webhook
func (m *Manager) dispatchToWebhook(webhook *Webhook, event Event) {
	now := time.Now()
	webhook.LastTrigger = &now

	// Prepare payload
	payload, err := json.Marshal(event)
	if err != nil {
		m.logger.Error("Failed to marshal webhook payload", zap.Error(err))
		webhook.FailureCount++
		return
	}

	// Create request
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		m.logger.Error("Failed to create webhook request", zap.Error(err))
		webhook.FailureCount++
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Dolphin-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", event.Type)
	req.Header.Set("X-Webhook-ID", webhook.ID)

	// Add signature if secret is set
	if webhook.Secret != "" {
		signature := m.generateSignature(payload, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Add custom headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Send request with retries
	client := &http.Client{Timeout: webhook.Timeout}
	var lastErr error

	for attempt := 0; attempt <= webhook.Retries; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < webhook.Retries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			break
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			webhook.SuccessCount++
			m.logger.Info("Webhook delivered successfully",
				zap.String("webhook_id", webhook.ID),
				zap.String("event_type", event.Type),
			)
			return
		}

		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		if attempt < webhook.Retries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	webhook.FailureCount++
	m.logger.Error("Webhook delivery failed",
		zap.String("webhook_id", webhook.ID),
		zap.String("event_type", event.Type),
		zap.Error(lastErr),
		zap.Int("attempts", webhook.Retries+1),
	)
}

// generateSignature generates HMAC signature for webhook payload
func (m *Manager) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies webhook signature
func (m *Manager) VerifySignature(payload []byte, signature, secret string) bool {
	expected := m.generateSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// GetWebhooks returns all webhooks
func (m *Manager) GetWebhooks() map[string]*Webhook {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Webhook)
	for id, webhook := range m.webhooks {
		webhookCopy := *webhook
		result[id] = &webhookCopy
	}
	return result
}

// GetWebhook returns a specific webhook
func (m *Manager) GetWebhook(id string) (*Webhook, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	webhook, exists := m.webhooks[id]
	if !exists {
		return nil, fmt.Errorf("webhook '%s' not found", id)
	}

	webhookCopy := *webhook
	return &webhookCopy, nil
}

// Delete deletes a webhook
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.webhooks[id]; !exists {
		return fmt.Errorf("webhook '%s' not found", id)
	}

	delete(m.webhooks, id)
	m.logger.Info("Webhook deleted", zap.String("id", id))
	return nil
}

// Enable enables a webhook
func (m *Manager) Enable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	webhook, exists := m.webhooks[id]
	if !exists {
		return fmt.Errorf("webhook '%s' not found", id)
	}

	webhook.Enabled = true
	return nil
}

// Disable disables a webhook
func (m *Manager) Disable(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	webhook, exists := m.webhooks[id]
	if !exists {
		return fmt.Errorf("webhook '%s' not found", id)
	}

	webhook.Enabled = false
	return nil
}

