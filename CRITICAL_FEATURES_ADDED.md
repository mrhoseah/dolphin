# 🚀 Critical Features Added for Rapid Development

## Overview

Three critical features have been added to make Dolphin Framework even more powerful for rapid development in 2025:

1. **Task Scheduler (Cron-like)** - Scheduled task execution
2. **API Resources/Transformers** - Clean, standardized API responses
3. **Webhook System** - Event-driven integrations

---

## 1. ⏰ Task Scheduler (`internal/scheduler/`)

A powerful cron-like task scheduler for running scheduled jobs without external dependencies.

### Features
- ✅ **Cron Expression Support** - Standard cron syntax with seconds support
- ✅ **Task Management** - Register, enable, disable, and run tasks manually
- ✅ **Error Handling** - Automatic error tracking and retry logic
- ✅ **Runtime Control** - Start/stop scheduler, run tasks on-demand
- ✅ **Task Monitoring** - Track last run, next run, error counts

### Usage

```go
import "dolphin/internal/scheduler"

// Initialize
scheduler := scheduler.NewScheduler(logger)

// Register a task
scheduler.Register(&scheduler.Task{
    Name:        "cleanup_temp_files",
    Schedule:    "0 2 * * *", // Run daily at 2 AM
    Description: "Clean up temporary files",
    Enabled:     true,
    Handler: func() error {
        // Cleanup logic
        return cleanupTempFiles()
    },
})

// Register with seconds (every 30 seconds)
scheduler.Register(&scheduler.Task{
    Name:     "heartbeat",
    Schedule: "*/30 * * * * *", // Every 30 seconds
    Handler:  func() error { return sendHeartbeat() },
})

// Start scheduler
scheduler.Start()

// Run task immediately
scheduler.RunNow("cleanup_temp_files")

// Get task status
task, _ := scheduler.GetTask("cleanup_temp_files")
fmt.Printf("Last run: %v\n", task.LastRun)
fmt.Printf("Next run: %v\n", task.NextRun)
fmt.Printf("Error count: %d\n", task.ErrorCount)

// Stop scheduler
scheduler.Stop()
```

### Cron Expression Examples

```
"0 0 * * *"        // Daily at midnight
"0 */6 * * *"      // Every 6 hours
"*/30 * * * * *"   // Every 30 seconds
"0 9 * * 1-5"      // Weekdays at 9 AM
"0 0 1 * *"        // First day of month at midnight
```

### Integration with Queue System

```go
// Schedule a job to run periodically
scheduler.Register(&scheduler.Task{
    Name:     "generate_reports",
    Schedule: "0 0 * * *", // Daily
    Handler: func() error {
        // Dispatch to queue
        return queueManager.Dispatch(generateReportJob, "reports")
    },
})
```

---

## 2. 📦 API Resources (`internal/resources/`)

Transform your models into clean, standardized API responses with ease.

### Features
- ✅ **Resource Transformation** - Transform models to API responses
- ✅ **Collection Support** - Handle arrays of resources
- ✅ **Pagination** - Built-in paginated responses
- ✅ **Field Filtering** - Only/Except fields
- ✅ **Data Merging** - Add computed fields
- ✅ **JSON Serialization** - Automatic JSON conversion

### Usage

```go
import "dolphin/internal/resources"

// Single Resource
user := &models.User{
    ID:    1,
    Name:  "John Doe",
    Email: "john@example.com",
    Password: "hashed...", // Should not be exposed
}

// Create resource and exclude sensitive fields
resource := resources.NewResource(user).
    Except("password", "remember_token").
    Merge(map[string]interface{}{
        "avatar_url": fmt.Sprintf("/avatars/%d.jpg", user.ID),
    })

// Return as JSON
w.Header().Set("Content-Type", "application/json")
jsonData, _ := resource.ToJSON()
w.Write(jsonData)

// Collection
users := []models.User{user1, user2, user3}
collection := resources.NewCollection(users).
    Transform(func(data map[string]interface{}) map[string]interface{} {
        // Remove sensitive fields
        delete(data, "password")
        delete(data, "remember_token")
        // Add computed field
        data["profile_url"] = fmt.Sprintf("/users/%v", data["id"])
        return data
    })

// Paginated Response
page := 1
perPage := 20
total := int64(100)

paginated := resources.NewPaginatedResponse(
    users,
    page,
    perPage,
    total,
)

// Response:
// {
//   "data": [...],
//   "pagination": {
//     "current_page": 1,
//     "per_page": 20,
//     "total": 100,
//     "last_page": 5,
//     "from": 1,
//     "to": 20
//   }
// }
```

### Controller Example

```go
func (c *UserController) Index(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
    
    users, total, _ := c.repo.Paginate(page, perPage)
    
    response := resources.NewPaginatedResponse(users, page, perPage, total)
    jsonData, _ := response.ToJSON()
    
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}

func (c *UserController) Show(w http.ResponseWriter, r *http.Request) {
    user, _ := c.repo.FindByID(id)
    
    resource := resources.NewResource(user).
        Except("password", "remember_token")
    
    jsonData, _ := resource.ToJSON()
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}
```

---

## 3. 🔔 Webhook System (`internal/webhooks/`)

Event-driven webhook system for integrations and real-time notifications.

### Features
- ✅ **Event Dispatching** - Fire events to registered webhooks
- ✅ **Event Filtering** - Webhooks can listen to specific events
- ✅ **Retry Logic** - Automatic retries on failure
- ✅ **Signature Verification** - HMAC-SHA256 signatures
- ✅ **Worker Pool** - Concurrent webhook delivery
- ✅ **Statistics** - Track success/failure counts

### Usage

```go
import "dolphin/internal/webhooks"

// Initialize
webhookManager := webhooks.NewManager(logger, 5) // 5 workers
webhookManager.Start()
defer webhookManager.Stop()

// Register a webhook
webhookManager.Register(&webhooks.Webhook{
    URL:    "https://example.com/webhooks/user-created",
    Events: []string{"user.created", "user.updated"},
    Secret: "your-secret-key",
    Enabled: true,
    Retries: 3,
    Timeout: 30 * time.Second,
    Headers: map[string]string{
        "X-Custom-Header": "value",
    },
})

// Dispatch events
webhookManager.Dispatch("user.created", map[string]interface{}{
    "user_id": 123,
    "email": "user@example.com",
    "name": "John Doe",
})

webhookManager.Dispatch("order.completed", map[string]interface{}{
    "order_id": 456,
    "amount": 99.99,
    "currency": "USD",
})

// Listen to all events
webhookManager.Register(&webhooks.Webhook{
    URL:     "https://example.com/webhooks/all",
    Events:  []string{"*"}, // Wildcard
    Enabled: true,
})

// Get webhook statistics
webhook, _ := webhookManager.GetWebhook("webhook_id")
fmt.Printf("Success: %d, Failures: %d\n", 
    webhook.SuccessCount, webhook.FailureCount)
```

### Webhook Payload Format

```json
{
  "id": "evt_1234567890",
  "type": "user.created",
  "data": {
    "user_id": 123,
    "email": "user@example.com"
  },
  "timestamp": "2025-01-15T10:30:00Z",
  "source": "dolphin"
}
```

### Signature Verification (Receiver Side)

```go
// In your webhook receiver
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    signature := r.Header.Get("X-Webhook-Signature")
    body, _ := io.ReadAll(r.Body)
    secret := "your-secret-key"
    
    manager := webhooks.NewManager(logger, 1)
    if !manager.VerifySignature(body, signature, secret) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Process webhook...
}
```

### Integration with Events

```go
// In your event listeners
eventDispatcher.On("user.created", func(event events.Event) {
    // Dispatch to webhooks
    webhookManager.Dispatch("user.created", event.Data)
})

eventDispatcher.On("order.completed", func(event events.Event) {
    webhookManager.Dispatch("order.completed", event.Data)
})
```

---

## 🎯 Why These Features Are Critical

### Task Scheduler
- **No External Dependencies**: No need for cron daemon or external schedulers
- **Application-Level Control**: Manage scheduled tasks within your application
- **Easy Testing**: Run tasks on-demand for testing
- **Perfect for**: Cleanup jobs, report generation, data synchronization, periodic API calls

### API Resources
- **Consistent API Responses**: Standardized response format across all endpoints
- **Security**: Easy to exclude sensitive fields
- **Flexibility**: Transform data before sending to clients
- **Developer Experience**: Clean, readable code
- **Perfect for**: REST APIs, microservices, mobile app backends

### Webhook System
- **Real-time Integrations**: Notify external services instantly
- **Event-Driven Architecture**: Decouple your application from integrations
- **Reliability**: Built-in retry logic and error handling
- **Security**: Signature verification for authenticity
- **Perfect for**: Payment processing, notifications, third-party integrations, SaaS platforms

---

## 📝 Integration Examples

### Complete Example: User Registration with Webhooks

```go
// User registration
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
    user := createUser(...)
    
    // Dispatch webhook
    webhookManager.Dispatch("user.created", map[string]interface{}{
        "user_id": user.ID,
        "email": user.Email,
    })
    
    // Return resource
    resource := resources.NewResource(user).
        Except("password")
    
    jsonData, _ := resource.ToJSON()
    w.Write(jsonData)
}
```

### Scheduled Cleanup Task

```go
// Register cleanup task
scheduler.Register(&scheduler.Task{
    Name:     "cleanup_expired_sessions",
    Schedule: "0 */6 * * *", // Every 6 hours
    Handler: func() error {
        return cleanupExpiredSessions()
    },
})

// Start scheduler in main.go
scheduler.Start()
defer scheduler.Stop()
```

### API Endpoint with Resources

```go
func (c *ProductController) Index(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    perPage := 20
    
    products, total, _ := c.repo.Paginate(page, perPage)
    
    // Transform products
    collection := resources.NewCollection(products).
        Transform(func(data map[string]interface{}) map[string]interface{} {
            // Add image URL
            data["image_url"] = fmt.Sprintf("/images/products/%v.jpg", data["id"])
            // Format price
            data["formatted_price"] = fmt.Sprintf("$%.2f", data["price"])
            return data
        })
    
    response := resources.NewPaginatedResponse(
        collection.ToArray(),
        page,
        perPage,
        total,
    )
    
    jsonData, _ := response.ToJSON()
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}
```

---

## 🚀 Next Steps

1. **CLI Commands**:
   - `dolphin schedule:list` - List all scheduled tasks
   - `dolphin schedule:run <task>` - Run task immediately
   - `dolphin webhook:list` - List webhooks
   - `dolphin webhook:test <id>` - Test webhook delivery

2. **Enhanced Features**:
   - Task scheduling UI dashboard
   - Webhook delivery logs
   - Resource transformers with custom functions
   - Scheduled task dependencies

3. **Production Ready**:
   - All features are production-ready
   - Thread-safe implementations
   - Comprehensive error handling
   - Built-in monitoring and statistics

---

**All three features are now available and ready to use! 🐬**

