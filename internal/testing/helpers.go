package testing

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig holds configuration for tests
type TestConfig struct {
	Database TestDatabaseConfig
	HTTP     TestHTTPConfig
	Coverage TestCoverageConfig
}

type TestDatabaseConfig struct {
	Driver string
	DSN    string
}

type TestHTTPConfig struct {
	BaseURL string
	Timeout time.Duration
}

type TestCoverageConfig struct {
	Threshold float64
	Exclude   []string
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Database: TestDatabaseConfig{
			Driver: "sqlite3",
			DSN:    ":memory:",
		},
		HTTP: TestHTTPConfig{
			BaseURL: "http://localhost:8080",
			Timeout: 30 * time.Second,
		},
		Coverage: TestCoverageConfig{
			Threshold: 80.0,
			Exclude: []string{
				"**/migrations/**",
				"**/testdata/**",
				"**/examples/**",
			},
		},
	}
}

// TestDatabase manages test database operations
type TestDatabase struct {
	db     *sql.DB
	config TestDatabaseConfig
}

// NewTestDatabase creates a new test database
func NewTestDatabase(t *testing.T, config TestDatabaseConfig) *TestDatabase {
	db, err := sql.Open(config.Driver, config.DSN)
	require.NoError(t, err, "Failed to open test database")

	// Set connection pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	return &TestDatabase{
		db:     db,
		config: config,
	}
}

// Close closes the test database
func (td *TestDatabase) Close() error {
	return td.db.Close()
}

// DB returns the underlying database connection
func (td *TestDatabase) DB() *sql.DB {
	return td.db
}

// RunMigrations runs database migrations for testing
func (td *TestDatabase) RunMigrations(t *testing.T, migrations []string) {
	for _, migration := range migrations {
		_, err := td.db.Exec(migration)
		require.NoError(t, err, "Failed to run migration: %s", migration)
	}
}

// SeedData seeds the database with test data
func (td *TestDatabase) SeedData(t *testing.T, data map[string][]map[string]interface{}) {
	for table, rows := range data {
		for _, row := range rows {
			columns := make([]string, 0, len(row))
			values := make([]interface{}, 0, len(row))
			placeholders := make([]string, 0, len(row))

			for col, val := range row {
				columns = append(columns, col)
				values = append(values, val)
				placeholders = append(placeholders, "?")
			}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				table,
				fmt.Sprintf("%v", columns),
				fmt.Sprintf("%v", placeholders))

			_, err := td.db.Exec(query, values...)
			require.NoError(t, err, "Failed to seed data in table %s", table)
		}
	}
}

// Cleanup removes all data from test tables
func (td *TestDatabase) Cleanup(t *testing.T, tables []string) {
	for _, table := range tables {
		_, err := td.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, "Failed to cleanup table %s", table)
	}
}

// HTTPTestHelper provides utilities for HTTP testing
type HTTPTestHelper struct {
	baseURL string
	timeout time.Duration
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(config TestHTTPConfig) *HTTPTestHelper {
	return &HTTPTestHelper{
		baseURL: config.BaseURL,
		timeout: config.Timeout,
	}
}

// MakeRequest creates an HTTP request for testing
func (h *HTTPTestHelper) MakeRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	var reqBody *bytes.Buffer

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req
}

// MakeRequestWithHeaders creates an HTTP request with custom headers
func (h *HTTPTestHelper) MakeRequestWithHeaders(t *testing.T, method, url string, body interface{}, headers map[string]string) *http.Request {
	req := h.MakeRequest(t, method, url, body)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req
}

// AssertJSONResponse checks if response matches expected JSON
func (h *HTTPTestHelper) AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
	var actual interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(t, err, "Failed to unmarshal response body")

	assert.Equal(t, expected, actual)
}

// AssertJSONContains checks if response contains expected fields
func (h *HTTPTestHelper) AssertJSONContains(t *testing.T, w *httptest.ResponseRecorder, expected map[string]interface{}) {
	var actual map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(t, err, "Failed to unmarshal response body")

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		assert.True(t, exists, "Response should contain key: %s", key)
		assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
	}
}

// AssertStatus checks if response has expected status code
func (h *HTTPTestHelper) AssertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	assert.Equal(t, expected, w.Code, "Expected status %d, got %d", expected, w.Code)
}

// AssertHeader checks if response has expected header
func (h *HTTPTestHelper) AssertHeader(t *testing.T, w *httptest.ResponseRecorder, key, expected string) {
	actual := w.Header().Get(key)
	assert.Equal(t, expected, actual, "Expected header %s: %s, got: %s", key, expected, actual)
}

// TestFileHelper provides utilities for file testing
type TestFileHelper struct {
	t        *testing.T
	tempDir  string
	testData map[string][]byte
}

// NewTestFileHelper creates a new test file helper
func NewTestFileHelper(t *testing.T) *TestFileHelper {
	tempDir := t.TempDir()
	return &TestFileHelper{
		t:        t,
		tempDir:  tempDir,
		testData: make(map[string][]byte),
	}
}

// CreateFile creates a test file with given content
func (tfh *TestFileHelper) CreateFile(filename string, content []byte) string {
	filepath := filepath.Join(tfh.tempDir, filename)
	err := os.WriteFile(filepath, content, 0644)
	require.NoError(tfh.t, err, "Failed to create test file: %s", filename)

	tfh.testData[filename] = content
	return filepath
}

// CreateJSONFile creates a test JSON file
func (tfh *TestFileHelper) CreateJSONFile(filename string, data interface{}) string {
	content, err := json.MarshalIndent(data, "", "  ")
	require.NoError(tfh.t, err, "Failed to marshal JSON data")

	return tfh.CreateFile(filename, content)
}

// ReadFile reads content from a test file
func (tfh *TestFileHelper) ReadFile(filename string) []byte {
	filepath := filepath.Join(tfh.tempDir, filename)
	content, err := os.ReadFile(filepath)
	require.NoError(tfh.t, err, "Failed to read test file: %s", filename)

	return content
}

// TempDir returns the temporary directory path
func (tfh *TestFileHelper) TempDir() string {
	return tfh.tempDir
}

// MockService provides a base mock service implementation
type MockService struct {
	t *testing.T
}

// NewMockService creates a new mock service
func NewMockService(t *testing.T) *MockService {
	return &MockService{t: t}
}

// AssertCalled checks if a method was called with expected arguments
func (ms *MockService) AssertCalled(t *testing.T, method string, args ...interface{}) {
	// This is a placeholder for actual mock implementation
	// In real usage, you would use testify/mock or similar
	t.Helper()
	// Implementation would depend on your mock framework
}

// AssertNotCalled checks if a method was not called
func (ms *MockService) AssertNotCalled(t *testing.T, method string) {
	t.Helper()
	// Implementation would depend on your mock framework
}

// TestSuite provides a base test suite with common setup/teardown
type TestSuite struct {
	t      *testing.T
	db     *TestDatabase
	http   *HTTPTestHelper
	files  *TestFileHelper
	config *TestConfig
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T) *TestSuite {
	config := DefaultTestConfig()
	db := NewTestDatabase(t, config.Database)
	http := NewHTTPTestHelper(config.HTTP)
	files := NewTestFileHelper(t)

	return &TestSuite{
		t:      t,
		db:     db,
		http:   http,
		files:  files,
		config: config,
	}
}

// Setup runs before each test
func (ts *TestSuite) Setup() {
	// Override in your test suite
}

// Teardown runs after each test
func (ts *TestSuite) Teardown() {
	// Override in your test suite
}

// Run runs the test suite
func (ts *TestSuite) Run(name string, testFunc func(*TestSuite)) {
	ts.t.Run(name, func(t *testing.T) {
		ts.t = t
		ts.Setup()
		defer ts.Teardown()
		testFunc(ts)
	})
}

// DB returns the test database
func (ts *TestSuite) DB() *TestDatabase {
	return ts.db
}

// HTTP returns the HTTP test helper
func (ts *TestSuite) HTTP() *HTTPTestHelper {
	return ts.http
}

// Files returns the file test helper
func (ts *TestSuite) Files() *TestFileHelper {
	return ts.files
}

// Config returns the test configuration
func (ts *TestSuite) Config() *TestConfig {
	return ts.config
}

// TestData provides common test data
var TestData = struct {
	Users []map[string]interface{}
	Posts []map[string]interface{}
}{
	Users: []map[string]interface{}{
		{
			"id":         1,
			"name":       "John Doe",
			"email":      "john@example.com",
			"created_at": time.Now().Format(time.RFC3339),
		},
		{
			"id":         2,
			"name":       "Jane Smith",
			"email":      "jane@example.com",
			"created_at": time.Now().Format(time.RFC3339),
		},
	},
	Posts: []map[string]interface{}{
		{
			"id":         1,
			"title":      "Test Post 1",
			"content":    "This is test content",
			"user_id":    1,
			"created_at": time.Now().Format(time.RFC3339),
		},
		{
			"id":         2,
			"title":      "Test Post 2",
			"content":    "This is more test content",
			"user_id":    2,
			"created_at": time.Now().Format(time.RFC3339),
		},
	},
}

// CommonMigrations provides common database migrations for testing
var CommonMigrations = []string{
	`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title VARCHAR(255) NOT NULL,
		content TEXT,
		user_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`,
}
