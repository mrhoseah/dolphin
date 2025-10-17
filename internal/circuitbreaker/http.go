package circuitbreaker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HTTPClient represents an HTTP client with circuit breaker protection
type HTTPClient struct {
	client         *http.Client
	circuitBreaker *CircuitBreaker
	logger         *zap.Logger
}

// HTTPClientConfig represents HTTP client configuration
type HTTPClientConfig struct {
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout" json:"idle_conn_timeout"`
	MaxConnsPerHost int           `yaml:"max_conns_per_host" json:"max_conns_per_host"`
}

// DefaultHTTPClientConfig returns default HTTP client configuration
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:         30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
		MaxConnsPerHost: 10,
	}
}

// NewHTTPClient creates a new HTTP client with circuit breaker protection
func NewHTTPClient(circuitName string, config *Config, httpConfig *HTTPClientConfig, logger *zap.Logger) *HTTPClient {
	if config == nil {
		config = DefaultConfig()
	}

	if httpConfig == nil {
		httpConfig = DefaultHTTPClientConfig()
	}

	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker(circuitName, config, logger)

	// Create HTTP client
	client := &http.Client{
		Timeout: httpConfig.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:    httpConfig.MaxIdleConns,
			IdleConnTimeout: httpConfig.IdleConnTimeout,
			MaxConnsPerHost: httpConfig.MaxConnsPerHost,
		},
	}

	return &HTTPClient{
		client:         client,
		circuitBreaker: circuitBreaker,
		logger:         logger,
	}
}

// Get performs a GET request with circuit breaker protection
func (hc *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	return hc.Do(ctx, "GET", url, nil, nil)
}

// Post performs a POST request with circuit breaker protection
func (hc *HTTPClient) Post(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return hc.Do(ctx, "POST", url, body, headers)
}

// Put performs a PUT request with circuit breaker protection
func (hc *HTTPClient) Put(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return hc.Do(ctx, "PUT", url, body, headers)
}

// Delete performs a DELETE request with circuit breaker protection
func (hc *HTTPClient) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return hc.Do(ctx, "DELETE", url, nil, headers)
}

// Do performs an HTTP request with circuit breaker protection
func (hc *HTTPClient) Do(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	var result *http.Response
	var err error

	// Execute with circuit breaker
	_, err = hc.circuitBreaker.Execute(ctx, func() (interface{}, error) {
		// Create request
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}

		// Add headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// Perform request
		resp, err := hc.client.Do(req)
		if err != nil {
			return nil, err
		}

		// Check for HTTP error status codes
		if resp.StatusCode >= 400 {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
		}

		result = resp
		return resp, nil
	})

	return result, err
}

// DoAsync performs an HTTP request asynchronously with circuit breaker protection
func (hc *HTTPClient) DoAsync(ctx context.Context, method, url string, body io.Reader, headers map[string]string) <-chan HTTPResult {
	resultChan := make(chan HTTPResult, 1)

	go func() {
		defer close(resultChan)

		resp, err := hc.Do(ctx, method, url, body, headers)
		resultChan <- HTTPResult{
			Response: resp,
			Error:    err,
		}
	}()

	return resultChan
}

// GetCircuitBreaker returns the underlying circuit breaker
func (hc *HTTPClient) GetCircuitBreaker() *CircuitBreaker {
	return hc.circuitBreaker
}

// GetStats returns circuit breaker statistics
func (hc *HTTPClient) GetStats() Stats {
	return hc.circuitBreaker.GetStats()
}

// Reset resets the circuit breaker
func (hc *HTTPClient) Reset() {
	hc.circuitBreaker.Reset()
}

// ForceOpen forces the circuit breaker to open state
func (hc *HTTPClient) ForceOpen() {
	hc.circuitBreaker.ForceOpen()
}

// ForceClose forces the circuit breaker to closed state
func (hc *HTTPClient) ForceClose() {
	hc.circuitBreaker.ForceClose()
}

// HTTPResult represents the result of an async HTTP request
type HTTPResult struct {
	Response *http.Response
	Error    error
}

// HTTPClientManager manages multiple HTTP clients with circuit breakers
type HTTPClientManager struct {
	clients map[string]*HTTPClient
	manager *Manager
	logger  *zap.Logger
	mu      sync.RWMutex
}

// NewHTTPClientManager creates a new HTTP client manager
func NewHTTPClientManager(manager *Manager, logger *zap.Logger) *HTTPClientManager {
	return &HTTPClientManager{
		clients: make(map[string]*HTTPClient),
		manager: manager,
		logger:  logger,
	}
}

// CreateClient creates a new HTTP client with circuit breaker
func (hcm *HTTPClientManager) CreateClient(name string, config *Config, httpConfig *HTTPClientConfig) (*HTTPClient, error) {
	if name == "" {
		return nil, fmt.Errorf("client name cannot be empty")
	}

	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	// Check if client already exists
	if _, exists := hcm.clients[name]; exists {
		return nil, fmt.Errorf("HTTP client %s already exists", name)
	}

	// Create HTTP client
	client := NewHTTPClient(name, config, httpConfig, hcm.logger)
	hcm.clients[name] = client

	// Register circuit breaker with manager
	if hcm.manager != nil {
		hcm.manager.circuits[name] = client.GetCircuitBreaker()
	}

	if hcm.logger != nil {
		hcm.logger.Info("HTTP client created",
			zap.String("client", name),
			zap.Duration("timeout", httpConfig.Timeout))
	}

	return client, nil
}

// GetClient returns an HTTP client by name
func (hcm *HTTPClientManager) GetClient(name string) (*HTTPClient, bool) {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()

	client, exists := hcm.clients[name]
	return client, exists
}

// RemoveClient removes an HTTP client
func (hcm *HTTPClientManager) RemoveClient(name string) error {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	client, exists := hcm.clients[name]
	if !exists {
		return fmt.Errorf("HTTP client %s not found", name)
	}

	// Unregister circuit breaker from manager
	if hcm.manager != nil {
		delete(hcm.manager.circuits, name)
	}

	// Remove client
	delete(hcm.clients, name)

	if hcm.logger != nil {
		hcm.logger.Info("HTTP client removed",
			zap.String("client", name))
	}

	return nil
}

// GetClientNames returns all client names
func (hcm *HTTPClientManager) GetClientNames() []string {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()

	names := make([]string, 0, len(hcm.clients))
	for name := range hcm.clients {
		names = append(names, name)
	}
	return names
}

// GetAllStats returns statistics for all clients
func (hcm *HTTPClientManager) GetAllStats() map[string]Stats {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()

	stats := make(map[string]Stats)
	for name, client := range hcm.clients {
		stats[name] = client.GetStats()
	}
	return stats
}

// ResetAll resets all circuit breakers
func (hcm *HTTPClientManager) ResetAll() {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()

	for name, client := range hcm.clients {
		client.Reset()

		if hcm.logger != nil {
			hcm.logger.Info("HTTP client circuit breaker reset",
				zap.String("client", name))
		}
	}
}

// GetManagerStats returns manager statistics
func (hcm *HTTPClientManager) GetManagerStats() HTTPClientManagerStats {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()

	clientCount := len(hcm.clients)
	openClients := 0
	closedClients := 0
	halfOpenClients := 0

	for _, client := range hcm.clients {
		switch client.GetCircuitBreaker().GetState() {
		case StateOpen:
			openClients++
		case StateClosed:
			closedClients++
		case StateHalfOpen:
			halfOpenClients++
		}
	}

	return HTTPClientManagerStats{
		ClientCount:     clientCount,
		OpenClients:     openClients,
		ClosedClients:   closedClients,
		HalfOpenClients: halfOpenClients,
	}
}

// HTTPClientManagerStats represents HTTP client manager statistics
type HTTPClientManagerStats struct {
	ClientCount     int `json:"client_count"`
	OpenClients     int `json:"open_clients"`
	ClosedClients   int `json:"closed_clients"`
	HalfOpenClients int `json:"half_open_clients"`
}
