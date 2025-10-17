# 🧪 Dolphin Framework Testing Guide

This guide covers how to write, run, and organize tests for Dolphin applications.

## 📋 Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Structure](#test-structure)
- [Writing Tests](#writing-tests)
- [Running Tests](#running-tests)
- [Testing Utilities](#testing-utilities)
- [Test Examples](#test-examples)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)

## 🎯 Testing Philosophy

Dolphin follows Go's testing conventions with additional utilities for web application testing:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **HTTP Tests**: Test API endpoints and web routes
- **Database Tests**: Test data persistence and queries
- **Middleware Tests**: Test request/response processing

## 📁 Test Structure

```
my-app/
├── app/
│   ├── controllers/
│   │   ├── user_controller.go
│   │   └── user_controller_test.go
│   ├── models/
│   │   ├── user.go
│   │   └── user_test.go
│   └── services/
│       ├── auth_service.go
│       └── auth_service_test.go
├── internal/
│   ├── middleware/
│   │   ├── auth.go
│   │   └── auth_test.go
│   └── database/
│       ├── migrations/
│       └── testdata/
├── tests/
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
└── go.mod
```

## ✍️ Writing Tests

### Basic Test Structure

```go
package controllers

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserController_Create(t *testing.T) {
    // Arrange
    controller := NewUserController()
    payload := `{"name":"John Doe","email":"john@example.com"}`
    
    req := httptest.NewRequest("POST", "/users", strings.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    // Act
    controller.Create(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    assert.Contains(t, w.Body.String(), "John Doe")
}
```

### Database Testing

```go
package models

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/mrhoseah/dolphin/internal/testing"
)

func TestUser_Create(t *testing.T) {
    // Setup test database
    db := testing.SetupTestDB(t)
    defer testing.CleanupTestDB(t, db)
    
    // Test data
    user := &User{
        Name:      "Jane Doe",
        Email:     "jane@example.com",
        CreatedAt: time.Now(),
    }
    
    // Test creation
    err := user.Create(db)
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
    
    // Test retrieval
    foundUser, err := GetUserByEmail(db, "jane@example.com")
    assert.NoError(t, err)
    assert.Equal(t, user.Name, foundUser.Name)
}
```

### HTTP Handler Testing

```go
package controllers

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    
    "github.com/stretchr/testify/assert"
    "github.com/go-chi/chi/v5"
)

func TestUserRoutes(t *testing.T) {
    // Setup router
    r := chi.NewRouter()
    controller := NewUserController()
    
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/users", controller.Index)
        r.Post("/users", controller.Create)
        r.Get("/users/{id}", controller.Show)
        r.Put("/users/{id}", controller.Update)
        r.Delete("/users/{id}", controller.Delete)
    })
    
    tests := []struct {
        name           string
        method         string
        url            string
        body           string
        expectedStatus int
    }{
        {
            name:           "GET users list",
            method:         "GET",
            url:            "/api/v1/users",
            expectedStatus: http.StatusOK,
        },
        {
            name:           "POST create user",
            method:         "POST",
            url:            "/api/v1/users",
            body:           `{"name":"Test User","email":"test@example.com"}`,
            expectedStatus: http.StatusCreated,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var req *http.Request
            if tt.body != "" {
                req = httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
                req.Header.Set("Content-Type", "application/json")
            } else {
                req = httptest.NewRequest(tt.method, tt.url, nil)
            }
            
            w := httptest.NewRecorder()
            r.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

### Middleware Testing

```go
package middleware

import (
    "testing"
    "net/http"
    "net/http/httptest"
    
    "github.com/stretchr/testify/assert"
    "github.com/go-chi/chi/v5"
)

func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name           string
        token          string
        expectedStatus int
    }{
        {
            name:           "Valid token",
            token:          "valid-token-123",
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Invalid token",
            token:          "invalid-token",
            expectedStatus: http.StatusUnauthorized,
        },
        {
            name:           "No token",
            token:          "",
            expectedStatus: http.StatusUnauthorized,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := chi.NewRouter()
            r.Use(AuthMiddleware())
            r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            })
            
            req := httptest.NewRequest("GET", "/protected", nil)
            if tt.token != "" {
                req.Header.Set("Authorization", "Bearer "+tt.token)
            }
            
            w := httptest.NewRecorder()
            r.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

## 🏃 Running Tests

### Basic Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./app/controllers

# Run specific test
go test -run TestUserController_Create ./app/controllers

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Dolphin Test Commands

```bash
# Run all tests with Dolphin utilities
dolphin test

# Run tests with coverage
dolphin test --coverage

# Run specific test suite
dolphin test --suite=integration

# Run tests with database
dolphin test --with-db

# Run tests in watch mode
dolphin test --watch
```

### Test Configuration

Create a `test.yaml` file for test configuration:

```yaml
# test.yaml
database:
  driver: "sqlite3"
  dsn: ":memory:"
  
coverage:
  threshold: 80
  exclude:
    - "**/migrations/**"
    - "**/testdata/**"
    
suites:
  unit:
    timeout: "30s"
    parallel: true
  integration:
    timeout: "60s"
    parallel: false
  e2e:
    timeout: "300s"
    parallel: false
```

## 🛠️ Testing Utilities

### Database Testing Helpers

```go
package testing

import (
    "database/sql"
    "testing"
    "os"
    
    _ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates an in-memory test database
func SetupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    return db
}

// CleanupTestDB cleans up test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
    db.Close()
}

// SeedTestData populates database with test data
func SeedTestData(t *testing.T, db *sql.DB) {
    // Insert test users, posts, etc.
}
```

### HTTP Testing Helpers

```go
package testing

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

// MakeRequest creates an HTTP request for testing
func MakeRequest(t *testing.T, method, url string, body interface{}) *http.Request {
    var reqBody *bytes.Buffer
    
    if body != nil {
        jsonBody, err := json.Marshal(body)
        require.NoError(t, err)
        reqBody = bytes.NewBuffer(jsonBody)
    }
    
    req := httptest.NewRequest(method, url, reqBody)
    req.Header.Set("Content-Type", "application/json")
    
    return req
}

// AssertJSONResponse checks if response matches expected JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
    var actual interface{}
    err := json.Unmarshal(w.Body.Bytes(), &actual)
    require.NoError(t, err)
    
    assert.Equal(t, expected, actual)
}
```

### Mock Helpers

```go
package testing

import (
    "testing"
    "github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
    mock.Mock
}

func (m *MockUserService) CreateUser(name, email string) (*User, error) {
    args := m.Called(name, email)
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserService) GetUser(id int) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}

// Usage in tests
func TestUserController_WithMock(t *testing.T) {
    mockService := new(MockUserService)
    controller := NewUserController(mockService)
    
    // Setup mock expectations
    mockService.On("CreateUser", "John", "john@example.com").Return(&User{
        ID: 1, Name: "John", Email: "john@example.com",
    }, nil)
    
    // Test controller
    // ... test code ...
    
    // Verify mock calls
    mockService.AssertExpectations(t)
}
```

## 📝 Test Examples

### Controller Test Example

```go
package controllers

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/mrhoseah/dolphin/internal/testing"
)

func TestUserController_Integration(t *testing.T) {
    // Setup test database
    db := testing.SetupTestDB(t)
    defer testing.CleanupTestDB(t, db)
    
    // Setup controller with test database
    controller := NewUserController(db)
    
    t.Run("Create User", func(t *testing.T) {
        payload := `{"name":"Alice","email":"alice@example.com"}`
        req := httptest.NewRequest("POST", "/users", strings.NewReader(payload))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        
        controller.Create(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)
        
        assert.Equal(t, "Alice", response["name"])
        assert.NotNil(t, response["id"])
    })
    
    t.Run("Get User", func(t *testing.T) {
        // First create a user
        user := &User{Name: "Bob", Email: "bob@example.com"}
        err := user.Create(db)
        require.NoError(t, err)
        
        // Then test getting the user
        req := httptest.NewRequest("GET", "/users/"+string(rune(user.ID)), nil)
        w := httptest.NewRecorder()
        
        controller.Show(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Contains(t, w.Body.String(), "Bob")
    })
}
```

### Service Test Example

```go
package services

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/mrhoseah/dolphin/internal/testing"
)

func TestAuthService_Login(t *testing.T) {
    db := testing.SetupTestDB(t)
    defer testing.CleanupTestDB(t, db)
    
    service := NewAuthService(db)
    
    // Create test user
    user := &User{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "hashedpassword",
    }
    err := user.Create(db)
    require.NoError(t, err)
    
    t.Run("Valid Login", func(t *testing.T) {
        token, err := service.Login("test@example.com", "password")
        
        assert.NoError(t, err)
        assert.NotEmpty(t, token)
    })
    
    t.Run("Invalid Email", func(t *testing.T) {
        _, err := service.Login("wrong@example.com", "password")
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "user not found")
    })
    
    t.Run("Invalid Password", func(t *testing.T) {
        _, err := service.Login("test@example.com", "wrongpassword")
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid password")
    })
}
```

## 🎯 Best Practices

### 1. Test Organization
- Keep tests close to the code they test
- Use descriptive test names
- Group related tests with subtests
- Use table-driven tests for multiple scenarios

### 2. Test Data
- Use factories for creating test data
- Keep test data minimal and focused
- Use realistic but simple test data
- Clean up test data after each test

### 3. Assertions
- Use specific assertions (`assert.Equal`, `assert.Contains`)
- Test both success and failure cases
- Verify side effects (database changes, external calls)
- Use `require` for critical assertions that should stop the test

### 4. Test Isolation
- Each test should be independent
- Use fresh database for each test
- Mock external dependencies
- Avoid shared state between tests

### 5. Performance
- Use parallel tests where possible
- Keep tests fast (< 100ms for unit tests)
- Use in-memory databases for testing
- Mock slow operations

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### Test Scripts

```bash
#!/bin/bash
# scripts/test.sh

echo "🧪 Running Dolphin Framework Tests"
echo "=================================="

# Run unit tests
echo "📋 Running unit tests..."
go test -v -short ./...

# Run integration tests
echo "🔗 Running integration tests..."
go test -v -tags=integration ./tests/integration/...

# Run E2E tests
echo "🌐 Running E2E tests..."
go test -v -tags=e2e ./tests/e2e/...

# Generate coverage report
echo "📊 Generating coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo "✅ All tests completed!"
```

## 📚 Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://golang.org/doc/effective_go.html#testing)
- [Dolphin Testing Examples](https://github.com/mrhoseah/dolphin/tree/main/examples)

## 🆘 Troubleshooting

### Common Issues

1. **Database Connection Errors**
   ```bash
   # Use in-memory database for tests
   export DB_DSN=":memory:"
   ```

2. **Import Cycle Errors**
   ```bash
   # Move test utilities to separate package
   mkdir internal/testing
   ```

3. **Slow Tests**
   ```bash
   # Run tests in parallel
   go test -parallel 4 ./...
   ```

4. **Coverage Issues**
   ```bash
   # Exclude generated files from coverage
   go test -coverprofile=coverage.out -covermode=atomic ./...
   ```

Happy testing! 🎉
