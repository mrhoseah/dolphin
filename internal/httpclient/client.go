package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents the Dolphin HTTP client
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	retries    int
	timeout    time.Duration
}

// Config represents client configuration
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	Retries    int
	Headers    map[string]string
	Transport  http.RoundTripper
}

// DefaultConfig returns the default client configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout: 30 * time.Second,
		Retries: 3,
		Headers: map[string]string{
			"User-Agent": "Dolphin-HTTP-Client/1.0.0",
			"Accept":     "application/json",
		},
	}
}

// NewClient creates a new HTTP client
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}
	
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: config.Transport,
	}
	
	return &Client{
		baseURL:    config.BaseURL,
		httpClient: httpClient,
		headers:    config.Headers,
		retries:    config.Retries,
		timeout:    config.Timeout,
	}
}

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
	Timeout time.Duration
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Request    *Request
}

// Get performs a GET request
func (c *Client) Get(url string, options ...RequestOption) (*Response, error) {
	return c.Request("GET", url, options...)
}

// Post performs a POST request
func (c *Client) Post(url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.Request("POST", url, append(options, WithBody(body))...)
}

// Put performs a PUT request
func (c *Client) Put(url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.Request("PUT", url, append(options, WithBody(body))...)
}

// Patch performs a PATCH request
func (c *Client) Patch(url string, body interface{}, options ...RequestOption) (*Response, error) {
	return c.Request("PATCH", url, append(options, WithBody(body))...)
}

// Delete performs a DELETE request
func (c *Client) Delete(url string, options ...RequestOption) (*Response, error) {
	return c.Request("DELETE", url, options...)
}

// Request performs an HTTP request with retries
func (c *Client) Request(method, url string, options ...RequestOption) (*Response, error) {
	req := &Request{
		Method: method,
		URL:    url,
		Headers: make(map[string]string),
		Timeout: c.timeout,
	}
	
	// Apply options
	for _, option := range options {
		option(req)
	}
	
	// Build full URL
	fullURL := c.buildURL(url)
	
	// Add query parameters
	if len(req.Query) > 0 {
		fullURL = c.addQueryParams(fullURL, req.Query)
	}
	
	// Create HTTP request
	httpReq, err := c.createHTTPRequest(method, fullURL, req.Body)
	if err != nil {
		return nil, err
	}
	
	// Set headers
	c.setHeaders(httpReq, req.Headers)
	
	// Perform request with retries
	var resp *http.Response
	var lastErr error
	
	for attempt := 0; attempt <= c.retries; attempt++ {
		ctx := context.Background()
		if req.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, req.Timeout)
			defer cancel()
		}
		
		resp, lastErr = c.httpClient.Do(httpReq.WithContext(ctx))
		if lastErr == nil {
			break
		}
		
		if attempt < c.retries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.retries, lastErr)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Request:    req,
	}, nil
}

// buildURL builds the full URL
func (c *Client) buildURL(url string) string {
	if c.baseURL == "" {
		return url
	}
	
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	
	baseURL := strings.TrimSuffix(c.baseURL, "/")
	url = strings.TrimPrefix(url, "/")
	
	return fmt.Sprintf("%s/%s", baseURL, url)
}

// addQueryParams adds query parameters to URL
func (c *Client) addQueryParams(url string, params map[string]string) string {
	u, err := url.Parse(url)
	if err != nil {
		return url
	}
	
	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	
	u.RawQuery = q.Encode()
	return u.String()
}

// createHTTPRequest creates an HTTP request
func (c *Client) createHTTPRequest(method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		case io.Reader:
			bodyReader = v
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
		}
	}
	
	return http.NewRequest(method, url, bodyReader)
}

// setHeaders sets request headers
func (c *Client) setHeaders(req *http.Request, requestHeaders map[string]string) {
	// Set default headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	
	// Set request-specific headers
	for key, value := range requestHeaders {
		req.Header.Set(key, value)
	}
	
	// Set Content-Type for JSON if body is present and not already set
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
}

// RequestOption represents a request option
type RequestOption func(*Request)

// WithHeaders sets request headers
func WithHeaders(headers map[string]string) RequestOption {
	return func(req *Request) {
		for key, value := range headers {
			req.Headers[key] = value
		}
	}
}

// WithHeader sets a single request header
func WithHeader(key, value string) RequestOption {
	return func(req *Request) {
		req.Headers[key] = value
	}
}

// WithBody sets request body
func WithBody(body interface{}) RequestOption {
	return func(req *Request) {
		req.Body = body
	}
}

// WithQuery sets query parameters
func WithQuery(params map[string]string) RequestOption {
	return func(req *Request) {
		req.Query = params
	}
}

// WithQueryParam sets a single query parameter
func WithQueryParam(key, value string) RequestOption {
	return func(req *Request) {
		if req.Query == nil {
			req.Query = make(map[string]string)
		}
		req.Query[key] = value
	}
}

// WithTimeout sets request timeout
func WithTimeout(timeout time.Duration) RequestOption {
	return func(req *Request) {
		req.Timeout = timeout
	}
}

// Response methods

// JSON unmarshals response body as JSON
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns response body as string
func (r *Response) String() string {
	return string(r.Body)
}

// Bytes returns response body as bytes
func (r *Response) Bytes() []byte {
	return r.Body
}

// IsSuccess checks if response is successful (2xx status code)
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError checks if response is a client error (4xx status code)
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError checks if response is a server error (5xx status code)
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// GetHeader returns a response header value
func (r *Response) GetHeader(key string) string {
	return r.Headers.Get(key)
}

// Error represents an HTTP client error
type Error struct {
	StatusCode int
	Message    string
	Response   *Response
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// CheckResponse checks if response is successful and returns error if not
func (r *Response) CheckResponse() error {
	if r.IsSuccess() {
		return nil
	}
	
	return &Error{
		StatusCode: r.StatusCode,
		Message:    fmt.Sprintf("Request failed with status %d", r.StatusCode),
		Response:   r,
	}
}

// Builder provides a fluent interface for building requests
type Builder struct {
	client *Client
	req    *Request
}

// NewBuilder creates a new request builder
func (c *Client) NewBuilder() *Builder {
	return &Builder{
		client: c,
		req: &Request{
			Headers: make(map[string]string),
			Query:   make(map[string]string),
			Timeout: c.timeout,
		},
	}
}

// Method sets the HTTP method
func (b *Builder) Method(method string) *Builder {
	b.req.Method = method
	return b
}

// URL sets the request URL
func (b *Builder) URL(url string) *Builder {
	b.req.URL = url
	return b
}

// Header sets a request header
func (b *Builder) Header(key, value string) *Builder {
	b.req.Headers[key] = value
	return b
}

// Headers sets multiple request headers
func (b *Builder) Headers(headers map[string]string) *Builder {
	for key, value := range headers {
		b.req.Headers[key] = value
	}
	return b
}

// Body sets the request body
func (b *Builder) Body(body interface{}) *Builder {
	b.req.Body = body
	return b
}

// Query sets query parameters
func (b *Builder) Query(params map[string]string) *Builder {
	for key, value := range params {
		b.req.Query[key] = value
	}
	return b
}

// QueryParam sets a single query parameter
func (b *Builder) QueryParam(key, value string) *Builder {
	b.req.Query[key] = value
	return b
}

// Timeout sets the request timeout
func (b *Builder) Timeout(timeout time.Duration) *Builder {
	b.req.Timeout = timeout
	return b
}

// Send sends the request
func (b *Builder) Send() (*Response, error) {
	return b.client.Request(b.req.Method, b.req.URL, 
		WithHeaders(b.req.Headers),
		WithBody(b.req.Body),
		WithQuery(b.req.Query),
		WithTimeout(b.req.Timeout),
	)
}

// Convenience methods for common HTTP methods

// GET creates a GET request builder
func (c *Client) GET(url string) *Builder {
	return c.NewBuilder().Method("GET").URL(url)
}

// POST creates a POST request builder
func (c *Client) POST(url string) *Builder {
	return c.NewBuilder().Method("POST").URL(url)
}

// PUT creates a PUT request builder
func (c *Client) PUT(url string) *Builder {
	return c.NewBuilder().Method("PUT").URL(url)
}

// PATCH creates a PATCH request builder
func (c *Client) PATCH(url string) *Builder {
	return c.NewBuilder().Method("PATCH").URL(url)
}

// DELETE creates a DELETE request builder
func (c *Client) DELETE(url string) *Builder {
	return c.NewBuilder().Method("DELETE").URL(url)
}
