package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestHelper provides testing utilities for Dolphin applications
type TestHelper struct {
	DB      *gorm.DB
	Router  chi.Router
	Server  *httptest.Server
	Cleanup func()
	t       *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Create router
	router := chi.NewRouter()

	// Create test server
	server := httptest.NewServer(router)

	// Setup cleanup
	cleanup := func() {
		server.Close()
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return &TestHelper{
		DB:      db,
		Router:  router,
		Server:  server,
		Cleanup: cleanup,
		t:       t,
	}
}

// MockDatabase creates a mock database for testing
func (h *TestHelper) MockDatabase() *gorm.DB {
	return h.DB
}

// MockCache creates a mock cache for testing
func (h *TestHelper) MockCache() *MockCache {
	return NewMockCache()
}

// MockEventDispatcher creates a mock event dispatcher for testing
func (h *TestHelper) MockEventDispatcher() *MockEventDispatcher {
	return NewMockEventDispatcher()
}

// MockMailManager creates a mock mail manager for testing
func (h *TestHelper) MockMailManager() *MockMailManager {
	return NewMockMailManager()
}

// SeedDatabase seeds the test database with data
func (h *TestHelper) SeedDatabase(seedFunc func(db *gorm.DB) error) {
	err := seedFunc(h.DB)
	require.NoError(h.t, err)
}

// MakeRequest makes an HTTP request to the test server
func (h *TestHelper) MakeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(h.t, err)
		reqBody = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, h.Server.URL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, req)

	return recorder
}

// MakeRequestWithHeaders makes an HTTP request with custom headers
func (h *TestHelper) MakeRequestWithHeaders(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(h.t, err)
		reqBody = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, h.Server.URL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, req)

	return recorder
}

// AssertResponseStatus asserts the response status code
func (h *TestHelper) AssertResponseStatus(recorder *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(h.t, expectedStatus, recorder.Code)
}

// AssertResponseJSON asserts the response JSON
func (h *TestHelper) AssertResponseJSON(recorder *httptest.ResponseRecorder, expected interface{}) {
	var actual interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(h.t, err)

	assert.Equal(h.t, expected, actual)
}

// AssertResponseContains asserts the response contains a string
func (h *TestHelper) AssertResponseContains(recorder *httptest.ResponseRecorder, expected string) {
	assert.Contains(h.t, recorder.Body.String(), expected)
}

// AssertResponseHeader asserts the response header
func (h *TestHelper) AssertResponseHeader(recorder *httptest.ResponseRecorder, header, expected string) {
	assert.Equal(h.t, expected, recorder.Header().Get(header))
}

// MockCache implements a mock cache for testing
type MockCache struct {
	data map[string]interface{}
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	if value, exists := m.data[key]; exists {
		return value.(string), nil
	}
	return "", fmt.Errorf("key not found")
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockCache) Flush(ctx context.Context) error {
	m.data = make(map[string]interface{})
	return nil
}

// MockEventDispatcher implements a mock event dispatcher for testing
type MockEventDispatcher struct {
	events    []Event
	listeners map[string][]Listener
}

type Event struct {
	Name    string
	Payload interface{}
}

type Listener interface {
	Handle(ctx context.Context, event Event) error
}

// NewMockEventDispatcher creates a new mock event dispatcher
func NewMockEventDispatcher() *MockEventDispatcher {
	return &MockEventDispatcher{
		events:    make([]Event, 0),
		listeners: make(map[string][]Listener),
	}
}

func (m *MockEventDispatcher) Dispatch(ctx context.Context, event Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventDispatcher) GetEvents() []Event {
	return m.events
}

func (m *MockEventDispatcher) ClearEvents() {
	m.events = make([]Event, 0)
}

// MockMailManager implements a mock mail manager for testing
type MockMailManager struct {
	sentEmails []SentEmail
}

type SentEmail struct {
	To      []string
	Subject string
	Body    string
	SentAt  time.Time
}

// NewMockMailManager creates a new mock mail manager
func NewMockMailManager() *MockMailManager {
	return &MockMailManager{
		sentEmails: make([]SentEmail, 0),
	}
}

func (m *MockMailManager) Send(ctx context.Context, message *Message) error {
	m.sentEmails = append(m.sentEmails, SentEmail{
		To:      message.To,
		Subject: message.Subject,
		Body:    message.HTML,
		SentAt:  time.Now(),
	})
	return nil
}

func (m *MockMailManager) GetSentEmails() []SentEmail {
	return m.sentEmails
}

func (m *MockMailManager) ClearSentEmails() {
	m.sentEmails = make([]SentEmail, 0)
}

// Message represents an email message (simplified for testing)
type Message struct {
	To      []string
	Subject string
	HTML    string
}

// TestDatabaseSeeder provides database seeding utilities
type TestDatabaseSeeder struct {
	db *gorm.DB
}

// NewTestDatabaseSeeder creates a new test database seeder
func NewTestDatabaseSeeder(db *gorm.DB) *TestDatabaseSeeder {
	return &TestDatabaseSeeder{db: db}
}

// SeedUsers seeds users in the test database
func (s *TestDatabaseSeeder) SeedUsers(count int) error {
	users := make([]User, count)
	for i := 0; i < count; i++ {
		users[i] = User{
			Name:  fmt.Sprintf("User %d", i+1),
			Email: fmt.Sprintf("user%d@example.com", i+1),
		}
	}

	return s.db.Create(&users).Error
}

// SeedPosts seeds posts in the test database
func (s *TestDatabaseSeeder) SeedPosts(count int, userID uint) error {
	posts := make([]Post, count)
	for i := 0; i < count; i++ {
		posts[i] = Post{
			Title:   fmt.Sprintf("Post %d", i+1),
			Content: fmt.Sprintf("Content for post %d", i+1),
			UserID:  userID,
		}
	}

	return s.db.Create(&posts).Error
}

// User represents a user model for testing
type User struct {
	ID    uint `gorm:"primarykey"`
	Name  string
	Email string
}

// Post represents a post model for testing
type Post struct {
	ID      uint `gorm:"primarykey"`
	Title   string
	Content string
	UserID  uint
}

// TestFileHelper provides file testing utilities
type TestFileHelper struct {
	tempDir string
	t       *testing.T
}

// NewTestFileHelper creates a new test file helper
func NewTestFileHelper(t *testing.T) *TestFileHelper {
	tempDir, err := os.MkdirTemp("", "dolphin_test_*")
	require.NoError(t, err)

	return &TestFileHelper{tempDir: tempDir, t: t}
}

// CreateTestFile creates a test file
func (h *TestFileHelper) CreateTestFile(filename, content string) string {
	filepath := filepath.Join(h.tempDir, filename)
	err := os.WriteFile(filepath, []byte(content), 0644)
	require.NoError(h.t, err)
	return filepath
}

// Cleanup cleans up test files
func (h *TestFileHelper) Cleanup() {
	os.RemoveAll(h.tempDir)
}

// GetTempDir returns the temporary directory
func (h *TestFileHelper) GetTempDir() string {
	return h.tempDir
}

// TestConfigHelper provides configuration testing utilities
type TestConfigHelper struct {
	config map[string]interface{}
}

// NewTestConfigHelper creates a new test config helper
func NewTestConfigHelper() *TestConfigHelper {
	return &TestConfigHelper{
		config: make(map[string]interface{}),
	}
}

// Set sets a configuration value
func (h *TestConfigHelper) Set(key string, value interface{}) {
	h.config[key] = value
}

// Get gets a configuration value
func (h *TestConfigHelper) Get(key string) interface{} {
	return h.config[key]
}

// GetString gets a string configuration value
func (h *TestConfigHelper) GetString(key string) string {
	if value, exists := h.config[key]; exists {
		return value.(string)
	}
	return ""
}

// GetInt gets an int configuration value
func (h *TestConfigHelper) GetInt(key string) int {
	if value, exists := h.config[key]; exists {
		return value.(int)
	}
	return 0
}

// GetBool gets a bool configuration value
func (h *TestConfigHelper) GetBool(key string) bool {
	if value, exists := h.config[key]; exists {
		return value.(bool)
	}
	return false
}

// TestAssertionHelper provides assertion utilities
type TestAssertionHelper struct {
	t *testing.T
}

// NewTestAssertionHelper creates a new test assertion helper
func NewTestAssertionHelper(t *testing.T) *TestAssertionHelper {
	return &TestAssertionHelper{t: t}
}

// AssertDatabaseHas asserts that the database has a record
func (h *TestAssertionHelper) AssertDatabaseHas(db *gorm.DB, table string, conditions map[string]interface{}) {
	var count int64
	query := db.Table(table)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	query.Count(&count)
	assert.Greater(h.t, count, int64(0), "Database should have record in table %s with conditions %v", table, conditions)
}

// AssertDatabaseMissing asserts that the database doesn't have a record
func (h *TestAssertionHelper) AssertDatabaseMissing(db *gorm.DB, table string, conditions map[string]interface{}) {
	var count int64
	query := db.Table(table)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	query.Count(&count)
	assert.Equal(h.t, int64(0), count, "Database should not have record in table %s with conditions %v", table, conditions)
}

// AssertEmailSent asserts that an email was sent
func (h *TestAssertionHelper) AssertEmailSent(mailManager *MockMailManager, subject string) {
	emails := mailManager.GetSentEmails()
	found := false
	for _, email := range emails {
		if email.Subject == subject {
			found = true
			break
		}
	}
	assert.True(h.t, found, "Email with subject '%s' should have been sent", subject)
}

// AssertEventDispatched asserts that an event was dispatched
func (h *TestAssertionHelper) AssertEventDispatched(eventDispatcher *MockEventDispatcher, eventName string) {
	events := eventDispatcher.GetEvents()
	found := false
	for _, event := range events {
		if event.Name == eventName {
			found = true
			break
		}
	}
	assert.True(h.t, found, "Event '%s' should have been dispatched", eventName)
}
