package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// ClientGenerator generates type-safe API clients from OpenAPI specs
type ClientGenerator struct {
	outputDir string
	language  string // "go" or "typescript"
}

// NewClientGenerator creates a new client generator
func NewClientGenerator(outputDir, language string) *ClientGenerator {
	return &ClientGenerator{
		outputDir: outputDir,
		language:  language,
	}
}

// Generate generates a client from OpenAPI spec
func (g *ClientGenerator) Generate(specPath string) error {
	// Read OpenAPI spec
	spec, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}

	// Parse OpenAPI spec (simplified - in production, use openapi3 library)
	switch g.language {
	case "go":
		return g.generateGoClient(spec)
	case "typescript":
		return g.generateTypeScriptClient(spec)
	default:
		return fmt.Errorf("unsupported language: %s", g.language)
	}
}

// generateGoClient generates a Go client
func (g *ClientGenerator) generateGoClient(spec []byte) error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return err
	}

	// Generate client.go
	clientCode := g.generateGoClientCode(spec)
	clientPath := filepath.Join(g.outputDir, "client.go")
	if err := os.WriteFile(clientPath, []byte(clientCode), 0644); err != nil {
		return err
	}

	// Generate types.go
	typesCode := g.generateGoTypes(spec)
	typesPath := filepath.Join(g.outputDir, "types.go")
	if err := os.WriteFile(typesPath, []byte(typesCode), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Generated Go client in %s\n", g.outputDir)
	return nil
}

// generateTypeScriptClient generates a TypeScript client
func (g *ClientGenerator) generateTypeScriptClient(spec []byte) error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return err
	}

	// Generate client.ts
	clientCode := g.generateTypeScriptClientCode(spec)
	clientPath := filepath.Join(g.outputDir, "client.ts")
	if err := os.WriteFile(clientPath, []byte(clientCode), 0644); err != nil {
		return err
	}

	// Generate types.ts
	typesCode := g.generateTypeScriptTypes(spec)
	typesPath := filepath.Join(g.outputDir, "types.ts")
	if err := os.WriteFile(typesPath, []byte(typesCode), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Generated TypeScript client in %s\n", g.outputDir)
	return nil
}

// generateGoClientCode generates Go client code
func (g *ClientGenerator) generateGoClientCode(spec []byte) string {
	return `package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	Headers    map[string]string
}

// NewClient creates a new API client
func NewClient(baseURL string, options ...ClientOption) *Client {
	client := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Headers: make(map[string]string),
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// ClientOption represents a client option
type ClientOption func(*Client)

// WithAPIKey sets the API key
func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.APIKey = key
		c.Headers["X-API-Key"] = key
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// WithHeader sets a custom header
func WithHeader(key, value string) ClientOption {
	return func(c *Client) {
		c.Headers[key] = value
	}
}

// Request performs an HTTP request
func (c *Client) Request(method, path string, body interface{}, result interface{}) error {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Unmarshal response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request
func (c *Client) Get(path string, result interface{}) error {
	return c.Request("GET", path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.Request("POST", path, body, result)
}

// Put performs a PUT request
func (c *Client) Put(path string, body interface{}, result interface{}) error {
	return c.Request("PUT", path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string, result interface{}) error {
	return c.Request("DELETE", path, nil, result)
}
`
}

// generateGoTypes generates Go type definitions
func (g *ClientGenerator) generateGoTypes(spec []byte) string {
	return `package client

// Types generated from OpenAPI spec
// TODO: Parse OpenAPI spec and generate types automatically

type ErrorResponse struct {
	Error   string            ` + "`json:\"error\"`" + `
	Message string            ` + "`json:\"message\"`" + `
	Details map[string]interface{} ` + "`json:\"details,omitempty\"`" + `
}
`
}

// generateTypeScriptClientCode generates TypeScript client code
func (g *ClientGenerator) generateTypeScriptClientCode(spec []byte) string {
	return `/**
 * API Client generated from OpenAPI spec
 */

export class APIClient {
	private baseURL: string;
	private apiKey?: string;
	private headers: Record<string, string>;

	constructor(baseURL: string, options?: ClientOptions) {
		this.baseURL = baseURL;
		this.headers = {
			'Content-Type': 'application/json',
			...(options?.headers || {}),
		};
		if (options?.apiKey) {
			this.apiKey = options.apiKey;
			this.headers['X-API-Key'] = options.apiKey;
		}
	}

	async request<T>(
		method: string,
		path: string,
		body?: any
	): Promise<T> {
		const url = this.baseURL + path;

		const options: RequestInit = {
			method,
			headers: this.headers,
		};

		if (body) {
			options.body = JSON.stringify(body);
		}

		const response = await fetch(url, options);

		if (!response.ok) {
			const error = await response.json().catch(() => ({
				error: 'Request failed',
				message: response.statusText,
			}));
			throw new APIError(error.error || 'Request failed', response.status, error);
		}

		return response.json();
	}

	async get<T>(path: string): Promise<T> {
		return this.request<T>('GET', path);
	}

	async post<T>(path: string, body?: any): Promise<T> {
		return this.request<T>('POST', path, body);
	}

	async put<T>(path: string, body?: any): Promise<T> {
		return this.request<T>('PUT', path, body);
	}

	async delete<T>(path: string): Promise<T> {
		return this.request<T>('DELETE', path);
	}
}

export interface ClientOptions {
	apiKey?: string;
	headers?: Record<string, string>;
}

export class APIError extends Error {
	constructor(
		message: string,
		public status: number,
		public details?: any
	) {
		super(message);
		this.name = 'APIError';
	}
}
`
}

// generateTypeScriptTypes generates TypeScript type definitions
func (g *ClientGenerator) generateTypeScriptTypes(spec []byte) string {
	return `/**
 * Types generated from OpenAPI spec
 * TODO: Parse OpenAPI spec and generate types automatically
 */

export interface ErrorResponse {
	error: string;
	message: string;
	details?: Record<string, any>;
}
`
}

// GenerateClientCommand generates a CLI command for client generation
func GenerateClientCommand(specPath, outputDir, language string) error {
	generator := NewClientGenerator(outputDir, language)
	return generator.Generate(specPath)
}

