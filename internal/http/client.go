package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HTTPMethod represents HTTP methods
type HTTPMethod string

const (
	MethodGET     HTTPMethod = "GET"
	MethodPOST    HTTPMethod = "POST"
	MethodPUT     HTTPMethod = "PUT"
	MethodPATCH   HTTPMethod = "PATCH"
	MethodDELETE  HTTPMethod = "DELETE"
	MethodHEAD    HTTPMethod = "HEAD"
	MethodOPTIONS HTTPMethod = "OPTIONS"
)

// Config represents HTTP client configuration
type Config struct {
	// Basic settings
	BaseURL   string        `yaml:"base_url" json:"base_url"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout"`
	UserAgent string        `yaml:"user_agent" json:"user_agent"`

	// Retry settings
	MaxRetries    int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay" json:"retry_delay"`
	RetryBackoff  float64       `yaml:"retry_backoff" json:"retry_backoff"`
	MaxRetryDelay time.Duration `yaml:"max_retry_delay" json:"max_retry_delay"`
	RetryOnStatus []int         `yaml:"retry_on_status" json:"retry_on_status"`

	// Connection settings
	MaxIdleConns        int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" json:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout" json:"idle_conn_timeout"`
	DisableKeepAlives   bool          `yaml:"disable_keep_alives" json:"disable_keep_alives"`

	// TLS settings
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	CertFile           string `yaml:"cert_file" json:"cert_file"`
	KeyFile            string `yaml:"key_file" json:"key_file"`
	CAFile             string `yaml:"ca_file" json:"ca_file"`

	// Authentication
	AuthType     string `yaml:"auth_type" json:"auth_type"` // basic, bearer, api_key, digest
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password" json:"password"`
	Token        string `yaml:"token" json:"token"`
	APIKey       string `yaml:"api_key" json:"api_key"`
	APIKeyHeader string `yaml:"api_key_header" json:"api_key_header"`

	// Headers
	DefaultHeaders map[string]string `yaml:"default_headers" json:"default_headers"`

	// Circuit breaker
	EnableCircuitBreaker bool          `yaml:"enable_circuit_breaker" json:"enable_circuit_breaker"`
	FailureThreshold     int           `yaml:"failure_threshold" json:"failure_threshold"`
	SuccessThreshold     int           `yaml:"success_threshold" json:"success_threshold"`
	OpenTimeout          time.Duration `yaml:"open_timeout" json:"open_timeout"`

	// Rate limiting
	EnableRateLimit bool `yaml:"enable_rate_limit" json:"enable_rate_limit"`
	RateLimitRPS    int  `yaml:"rate_limit_rps" json:"rate_limit_rps"`
	RateLimitBurst  int  `yaml:"rate_limit_burst" json:"rate_limit_burst"`

	// Logging
	EnableLogging   bool `yaml:"enable_logging" json:"enable_logging"`
	VerboseLogging  bool `yaml:"verbose_logging" json:"verbose_logging"`
	LogRequestBody  bool `yaml:"log_request_body" json:"log_request_body"`
	LogResponseBody bool `yaml:"log_response_body" json:"log_response_body"`

	// Metrics
	EnableMetrics bool `yaml:"enable_metrics" json:"enable_metrics"`

	// Correlation ID
	EnableCorrelationID bool   `yaml:"enable_correlation_id" json:"enable_correlation_id"`
	CorrelationIDHeader string `yaml:"correlation_id_header" json:"correlation_id_header"`
}

// DefaultConfig returns default HTTP client configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:              "",
		Timeout:              30 * time.Second,
		UserAgent:            "Dolphin-HTTP-Client/1.0",
		MaxRetries:           3,
		RetryDelay:           1 * time.Second,
		RetryBackoff:         2.0,
		MaxRetryDelay:        30 * time.Second,
		RetryOnStatus:        []int{500, 502, 503, 504, 429},
		MaxIdleConns:         100,
		MaxIdleConnsPerHost:  10,
		IdleConnTimeout:      90 * time.Second,
		DisableKeepAlives:    false,
		InsecureSkipVerify:   false,
		AuthType:             "",
		DefaultHeaders:       make(map[string]string),
		EnableCircuitBreaker: true,
		FailureThreshold:     5,
		SuccessThreshold:     3,
		OpenTimeout:          60 * time.Second,
		EnableRateLimit:      false,
		RateLimitRPS:         100,
		RateLimitBurst:       10,
		EnableLogging:        true,
		VerboseLogging:       false,
		LogRequestBody:       false,
		LogResponseBody:      false,
		EnableMetrics:        true,
		EnableCorrelationID:  true,
		CorrelationIDHeader:  "X-Correlation-ID",
	}
}

// Request represents an HTTP request
type Request struct {
	Method        HTTPMethod             `json:"method"`
	URL           string                 `json:"url"`
	Headers       map[string]string      `json:"headers"`
	QueryParams   map[string]interface{} `json:"query_params"`
	Body          interface{}            `json:"body"`
	Context       context.Context        `json:"-"`
	Timeout       time.Duration          `json:"timeout"`
	Retries       int                    `json:"retries"`
	CorrelationID string                 `json:"correlation_id"`
}

// Response represents an HTTP response
type Response struct {
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Body          []byte            `json:"body"`
	Request       *Request          `json:"request"`
	Duration      time.Duration     `json:"duration"`
	RetryCount    int               `json:"retry_count"`
	CorrelationID string            `json:"correlation_id"`
	Error         error             `json:"error,omitempty"`
}

// Client represents an HTTP client
type Client struct {
	config     *Config
	httpClient *http.Client
	logger     *zap.Logger

	// Circuit breaker
	circuitBreaker *CircuitBreaker

	// Rate limiter
	rateLimiter *RateLimiter

	// Metrics
	metrics *Metrics

	// Correlation ID generator
	correlationIDGen *CorrelationIDGenerator

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewClient creates a new HTTP client
func NewClient(config *Config, logger *zap.Logger) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client
	httpClient, err := createHTTPClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create circuit breaker if enabled
	var circuitBreaker *CircuitBreaker
	if config.EnableCircuitBreaker {
		circuitBreaker = NewCircuitBreaker(&CircuitBreakerConfig{
			FailureThreshold: config.FailureThreshold,
			SuccessThreshold: config.SuccessThreshold,
			OpenTimeout:      config.OpenTimeout,
		})
	}

	// Create rate limiter if enabled
	var rateLimiter *RateLimiter
	if config.EnableRateLimit {
		rateLimiter = NewRateLimiter(config.RateLimitRPS, config.RateLimitBurst)
	}

	// Create metrics if enabled
	var metrics *Metrics
	if config.EnableMetrics {
		metrics = NewMetrics()
	}

	// Create correlation ID generator
	correlationIDGen := NewCorrelationIDGenerator()

	client := &Client{
		config:           config,
		httpClient:       httpClient,
		logger:           logger,
		circuitBreaker:   circuitBreaker,
		rateLimiter:      rateLimiter,
		metrics:          metrics,
		correlationIDGen: correlationIDGen,
	}

	return client, nil
}

// createHTTPClient creates the underlying HTTP client
func createHTTPClient(config *Config) (*http.Client, error) {
	// Create transport
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
	}

	// Configure TLS
	if config.InsecureSkipVerify || config.CertFile != "" || config.KeyFile != "" || config.CAFile != "" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		}

		// Load client certificate
		if config.CertFile != "" && config.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		// Load CA certificate
		if config.CAFile != "" {
			// This would typically load the CA certificate
			// For now, we'll skip this implementation
		}

		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}, nil
}

// Do executes an HTTP request
func (c *Client) Do(req *Request) (*Response, error) {
	// Generate correlation ID if enabled
	if c.config.EnableCorrelationID && req.CorrelationID == "" {
		req.CorrelationID = c.correlationIDGen.Generate()
	}

	// Apply rate limiting if enabled
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(req.Context); err != nil {
			return nil, fmt.Errorf("rate limit exceeded: %w", err)
		}
	}

	// Execute with circuit breaker if enabled
	if c.circuitBreaker != nil {
		return c.executeWithCircuitBreaker(req)
	}

	return c.executeRequest(req)
}

// executeWithCircuitBreaker executes request with circuit breaker
func (c *Client) executeWithCircuitBreaker(req *Request) (*Response, error) {
	result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.executeRequest(req)
	})

	if err != nil {
		return nil, err
	}

	response, ok := result.(*Response)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from circuit breaker")
	}

	return response, nil
}

// executeRequest executes the actual HTTP request
func (c *Client) executeRequest(req *Request) (*Response, error) {
	start := time.Now()

	// Build HTTP request
	httpReq, err := c.buildHTTPRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// Execute request with retries
	response, err := c.executeWithRetries(httpReq, req)
	if err != nil {
		return nil, err
	}

	// Record metrics
	if c.metrics != nil {
		c.metrics.RecordRequest(req.Method, response.StatusCode, time.Since(start))
	}

	// Log request/response if enabled
	if c.config.EnableLogging && c.logger != nil {
		c.logRequestResponse(req, response, time.Since(start))
	}

	return response, nil
}

// buildHTTPRequest builds the underlying HTTP request
func (c *Client) buildHTTPRequest(req *Request) (*http.Request, error) {
	// Build URL
	url, err := c.buildURL(req)
	if err != nil {
		return nil, err
	}

	// Build body
	body, err := c.buildBody(req.Body)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(req.Context, string(req.Method), url, body)
	if err != nil {
		return nil, err
	}

	// Set headers
	c.setHeaders(httpReq, req)

	return httpReq, nil
}

// buildURL builds the complete URL
func (c *Client) buildURL(req *Request) (string, error) {
	// Start with base URL or request URL
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = req.URL
	} else {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(req.URL, "/")
	}

	// Parse URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		query := parsedURL.Query()
		for key, value := range req.QueryParams {
			query.Set(key, fmt.Sprintf("%v", value))
		}
		parsedURL.RawQuery = query.Encode()
	}

	return parsedURL.String(), nil
}

// buildBody builds the request body
func (c *Client) buildBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	switch v := body.(type) {
	case string:
		return strings.NewReader(v), nil
	case []byte:
		return bytes.NewReader(v), nil
	case io.Reader:
		return v, nil
	default:
		// JSON encode
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body to JSON: %w", err)
		}
		return bytes.NewReader(jsonData), nil
	}
}

// setHeaders sets request headers
func (c *Client) setHeaders(httpReq *http.Request, req *Request) {
	// Set default headers
	for key, value := range c.config.DefaultHeaders {
		httpReq.Header.Set(key, value)
	}

	// Set user agent
	if c.config.UserAgent != "" {
		httpReq.Header.Set("User-Agent", c.config.UserAgent)
	}

	// Set content type for JSON body
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set request headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set correlation ID
	if c.config.EnableCorrelationID && req.CorrelationID != "" {
		httpReq.Header.Set(c.config.CorrelationIDHeader, req.CorrelationID)
	}

	// Set authentication
	c.setAuthentication(httpReq)
}

// setAuthentication sets authentication headers
func (c *Client) setAuthentication(httpReq *http.Request) {
	switch c.config.AuthType {
	case "basic":
		if c.config.Username != "" && c.config.Password != "" {
			httpReq.SetBasicAuth(c.config.Username, c.config.Password)
		}
	case "bearer":
		if c.config.Token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.config.Token)
		}
	case "api_key":
		if c.config.APIKey != "" {
			header := "X-API-Key"
			if c.config.APIKeyHeader != "" {
				header = c.config.APIKeyHeader
			}
			httpReq.Header.Set(header, c.config.APIKey)
		}
	}
}

// executeWithRetries executes request with retries
func (c *Client) executeWithRetries(httpReq *http.Request, req *Request) (*Response, error) {
	var lastErr error
	retries := req.Retries
	if retries == 0 {
		retries = c.config.MaxRetries
	}

	for attempt := 0; attempt <= retries; attempt++ {
		// Execute request
		httpResp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			if attempt < retries {
				c.waitBeforeRetry(attempt)
				continue
			}
			return nil, err
		}

		// Read response body
		body, err := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if err != nil {
			lastErr = err
			if attempt < retries {
				c.waitBeforeRetry(attempt)
				continue
			}
			return nil, err
		}

		// Check if we should retry
		if c.shouldRetry(httpResp.StatusCode, attempt, retries) {
			lastErr = fmt.Errorf("status %d", httpResp.StatusCode)
			c.waitBeforeRetry(attempt)
			continue
		}

		// Build response
		response := &Response{
			StatusCode:    httpResp.StatusCode,
			Headers:       make(map[string]string),
			Body:          body,
			Request:       req,
			Duration:      time.Since(time.Now()),
			RetryCount:    attempt,
			CorrelationID: req.CorrelationID,
		}

		// Copy headers
		for key, values := range httpResp.Header {
			if len(values) > 0 {
				response.Headers[key] = values[0]
			}
		}

		return response, nil
	}

	return nil, lastErr
}

// shouldRetry determines if we should retry the request
func (c *Client) shouldRetry(statusCode, attempt, maxRetries int) bool {
	if attempt >= maxRetries {
		return false
	}

	for _, retryStatus := range c.config.RetryOnStatus {
		if statusCode == retryStatus {
			return true
		}
	}

	return false
}

// waitBeforeRetry waits before retrying
func (c *Client) waitBeforeRetry(attempt int) {
	delay := c.config.RetryDelay
	if c.config.RetryBackoff > 1.0 {
		delay = time.Duration(float64(delay) * c.config.RetryBackoff * float64(attempt))
		if delay > c.config.MaxRetryDelay {
			delay = c.config.MaxRetryDelay
		}
	}

	time.Sleep(delay)
}

// logRequestResponse logs request and response
func (c *Client) logRequestResponse(req *Request, resp *Response, duration time.Duration) {
	if c.logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("method", string(req.Method)),
		zap.String("url", req.URL),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", duration),
		zap.Int("retry_count", resp.RetryCount),
	}

	if req.CorrelationID != "" {
		fields = append(fields, zap.String("correlation_id", req.CorrelationID))
	}

	if c.config.VerboseLogging {
		fields = append(fields, zap.Any("headers", req.Headers))
		if c.config.LogRequestBody && req.Body != nil {
			fields = append(fields, zap.Any("request_body", req.Body))
		}
		if c.config.LogResponseBody && len(resp.Body) > 0 {
			fields = append(fields, zap.String("response_body", string(resp.Body)))
		}
	}

	c.logger.Info("HTTP request completed", fields...)
}

// Convenience methods
func (c *Client) Get(url string, options ...RequestOption) (*Response, error) {
	return c.Request(MethodGET, url, options...)
}

func (c *Client) Post(url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.Request(MethodPOST, url, append(options, WithBody(body))...)
}

func (c *Client) Put(url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.Request(MethodPUT, url, append(options, WithBody(body))...)
}

func (c *Client) Patch(url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.Request(MethodPATCH, url, append(options, WithBody(body))...)
}

func (c *Client) Delete(url string, options ...RequestOption) (*Response, error) {
	return c.Request(MethodDELETE, url, options...)
}

func (c *Client) Head(url string, options ...RequestOption) (*Response, error) {
	return c.Request(MethodHEAD, url, options...)
}

func (c *Client) Options(url string, options ...RequestOption) (*Response, error) {
	return c.Request(MethodOPTIONS, url, options...)
}

// Request executes a request with options
func (c *Client) Request(method HTTPMethod, url string, options ...RequestOption) (*Response, error) {
	req := &Request{
		Method:  method,
		URL:     url,
		Headers: make(map[string]string),
		Context: context.Background(),
	}

	// Apply options
	for _, option := range options {
		option(req)
	}

	return c.Do(req)
}

// Close closes the client
func (c *Client) Close() error {
	if c.httpClient != nil {
		// Close idle connections
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	return nil
}

// GetMetrics returns client metrics
func (c *Client) GetMetrics() *Metrics {
	return c.metrics
}

// GetCircuitBreaker returns circuit breaker
func (c *Client) GetCircuitBreaker() *CircuitBreaker {
	return c.circuitBreaker
}

// GetRateLimiter returns rate limiter
func (c *Client) GetRateLimiter() *RateLimiter {
	return c.rateLimiter
}

// ConfigFromEnv creates config from environment variables
func ConfigFromEnv() *Config {
	config := DefaultConfig()

	// Override with environment variables if present
	// This would typically read from environment variables
	// For now, return the default config
	return config
}
