package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appTesting "github.com/mrhoseah/dolphin/internal/testing"
)

// Example User model
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Example UserController
type UserController struct {
	db *appTesting.TestDatabase
}

func NewUserController(db *appTesting.TestDatabase) *UserController {
	return &UserController{db: db}
}

func (uc *UserController) Index(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com", CreatedAt: time.Now()},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com", CreatedAt: time.Now()},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (uc *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simulate user creation
	user.ID = 3
	user.CreatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (uc *UserController) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Simulate user retrieval
	user := User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Example tests using Dolphin testing utilities

func TestUserController_Index(t *testing.T) {
	// Setup test database
	db := appTesting.NewTestDatabase(t, appTesting.TestDatabaseConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	defer db.Close()

	// Setup controller
	controller := NewUserController(db)

	// Create request
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	// Execute
	controller.Index(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var users []User
	err := json.Unmarshal(w.Body.Bytes(), &users)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "John Doe", users[0].Name)
}

func TestUserController_Create(t *testing.T) {
	// Setup test database
	db := appTesting.NewTestDatabase(t, appTesting.TestDatabaseConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	defer db.Close()

	// Setup controller
	controller := NewUserController(db)

	// Test data
	userData := User{
		Name:  "Alice Johnson",
		Email: "alice@example.com",
	}

	// Create request
	req := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"Alice Johnson","email":"alice@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	controller.Create(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var createdUser User
	err := json.Unmarshal(w.Body.Bytes(), &createdUser)
	require.NoError(t, err)
	assert.Equal(t, "Alice Johnson", createdUser.Name)
	assert.Equal(t, "alice@example.com", createdUser.Email)
	assert.NotZero(t, createdUser.ID)
}

func TestUserController_Show(t *testing.T) {
	// Setup test database
	db := appTesting.NewTestDatabase(t, appTesting.TestDatabaseConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	defer db.Close()

	// Setup controller
	controller := NewUserController(db)

	// Create request with URL parameter
	req := httptest.NewRequest("GET", "/users/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Execute
	controller.Show(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var user User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "John Doe", user.Name)
}

// Example of using HTTP test helper
func TestUserController_WithHTTPHelper(t *testing.T) {
	// Setup test database
	db := appTesting.NewTestDatabase(t, appTesting.TestDatabaseConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	defer db.Close()

	// Setup HTTP helper
	httpHelper := appTesting.NewHTTPTestHelper(appTesting.TestHTTPConfig{
		BaseURL: "http://localhost:8080",
		Timeout: 30 * time.Second,
	})

	// Setup controller
	controller := NewUserController(db)

	t.Run("Create User with Helper", func(t *testing.T) {
		// Test data
		userData := map[string]interface{}{
			"name":  "Bob Wilson",
			"email": "bob@example.com",
		}

		// Create request using helper
		req := httpHelper.MakeRequest(t, "POST", "/users", userData)
		w := httptest.NewRecorder()

		// Execute
		controller.Create(w, req)

		// Assertions using helper
		httpHelper.AssertStatus(t, w, http.StatusCreated)
		httpHelper.AssertHeader(t, w, "Content-Type", "application/json")

		// Check response contains expected data
		expected := map[string]interface{}{
			"name":  "Bob Wilson",
			"email": "bob@example.com",
		}
		httpHelper.AssertJSONContains(t, w, expected)
	})
}

// Example of using test suite
type UserTestSuite struct {
	*appTesting.TestSuite
	controller *UserController
}

func NewUserTestSuite(t *testing.T) *UserTestSuite {
	suite := appTesting.NewTestSuite(t)
	return &UserTestSuite{
		TestSuite:  suite,
		controller: NewUserController(suite.DB()),
	}
}

func (uts *UserTestSuite) Setup() {
	// Run migrations
	uts.DB().RunMigrations(uts.TestSuite.T(), appTesting.CommonMigrations)

	// Seed test data
	uts.DB().SeedData(uts.TestSuite.T(), map[string][]map[string]interface{}{
		"users": appTesting.TestData.Users,
		"posts": appTesting.TestData.Posts,
	})
}

func (uts *UserTestSuite) Teardown() {
	// Cleanup test data
	uts.DB().Cleanup(uts.TestSuite.T(), []string{"users", "posts"})
}

func TestUserController_WithTestSuite(t *testing.T) {
	suite := NewUserTestSuite(t)

	suite.Run("Index with seeded data", func(ts *appTesting.TestSuite) {
		userSuite := ts.(*UserTestSuite)

		// Create request
		req := httptest.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()

		// Execute
		userSuite.controller.Index(w, req)

		// Assertions
		userSuite.HTTP().AssertStatus(userSuite.t, w, http.StatusOK)
		userSuite.HTTP().AssertHeader(userSuite.t, w, "Content-Type", "application/json")

		// Check response contains expected data
		expected := map[string]interface{}{
			"users": appTesting.TestData.Users,
		}
		userSuite.HTTP().AssertJSONContains(userSuite.t, w, expected)
	})

	suite.Run("Create User with Helper", func(ts *appTesting.TestSuite) {
		userSuite := ts.(*UserTestSuite)

		// Test data
		userData := map[string]interface{}{
			"name":  "Charlie Brown",
			"email": "charlie@example.com",
		}

		// Create request using helper
		req := userSuite.HTTP().MakeRequest(userSuite.TestSuite.T(), "POST", "/users", userData)

		// Execute
		userSuite.controller.Create(w, req)

		// Assertions using helper
		userSuite.HTTP().AssertStatus(userSuite.TestSuite.T(), w, http.StatusCreated)
		userSuite.HTTP().AssertHeader(userSuite.TestSuite.T(), w, "Content-Type", "application/json")

		// Check response contains expected data
		expected := map[string]interface{}{
			"name":  "Charlie Brown",
			"email": "charlie@example.com",
		}
		userSuite.HTTP().AssertJSONContains(userSuite.TestSuite.T(), w, expected)
	})
}

// Example of table-driven tests
func TestUserController_Create_TableDriven(t *testing.T) {
	db := appTesting.NewTestDatabase(t, appTesting.TestDatabaseConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	defer db.Close()

	controller := NewUserController(db)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedName   string
		expectedEmail  string
	}{
		{
			name:           "Valid user creation",
			requestBody:    `{"name":"Alice","email":"alice@example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedName:   "Alice",
			expectedEmail:  "alice@example.com",
		},
		{
			name:           "User with long name",
			requestBody:    `{"name":"Very Long Name That Exceeds Normal Limits","email":"long@example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedName:   "Very Long Name That Exceeds Normal Limits",
			expectedEmail:  "long@example.com",
		},
		{
			name:           "User with special characters",
			requestBody:    `{"name":"José María","email":"jose@example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedName:   "José María",
			expectedEmail:  "jose@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/users", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var user User
			err := json.Unmarshal(w.Body.Bytes(), &user)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedName, user.Name)
			assert.Equal(t, tt.expectedEmail, user.Email)
		})
	}
}

// Example of testing with file helpers
func TestUserController_ExportUsers(t *testing.T) {
	db := appTesting.NewTestDatabase(t, appTesting.TestDatabaseConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	defer db.Close()

	// Setup file helper
	fileHelper := appTesting.NewTestFileHelper(t)

	// Create test JSON file
	expectedUsers := []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
	}

	fileHelper.CreateJSONFile("expected_users.json", expectedUsers)

	// Read and verify the file was created correctly
	content := fileHelper.ReadFile("expected_users.json")
	var actualUsers []User
	err := json.Unmarshal(content, &actualUsers)
	require.NoError(t, err)

	assert.Equal(t, expectedUsers, actualUsers)
	assert.Equal(t, 2, len(actualUsers))
}

// Example of testing error cases
func TestUserController_Create_ErrorCases(t *testing.T) {
	db := appTesting.NewTestDatabase(t, appTesting.TestDatabaseConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
	defer db.Close()

	controller := NewUserController(db)

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"John","email":}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid JSON")
	})

	t.Run("Empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/users", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"John"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Create(w, req)

		// In a real implementation, you might return 422 Unprocessable Entity
		// For this example, we'll assume it still creates the user
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func main() {
	fmt.Println("🧪 Dolphin Framework Testing Examples")
	fmt.Println("====================================")
	fmt.Println("")
	fmt.Println("This file contains comprehensive examples of testing with Dolphin Framework.")
	fmt.Println("Run the tests with: go test -v")
	fmt.Println("")
	fmt.Println("Examples included:")
	fmt.Println("  ✅ Basic controller testing")
	fmt.Println("  ✅ HTTP helper utilities")
	fmt.Println("  ✅ Test suite with setup/teardown")
	fmt.Println("  ✅ Table-driven tests")
	fmt.Println("  ✅ File testing utilities")
	fmt.Println("  ✅ Error case testing")
	fmt.Println("  ✅ Database testing with migrations")
	fmt.Println("")
	fmt.Println("💡 Key features demonstrated:")
	fmt.Println("  • TestDatabase for in-memory SQLite testing")
	fmt.Println("  • HTTPTestHelper for request/response testing")
	fmt.Println("  • TestFileHelper for file operations")
	fmt.Println("  • TestSuite for organized test structure")
	fmt.Println("  • Common test data and migrations")
	fmt.Println("  • Comprehensive assertions and error handling")
}
