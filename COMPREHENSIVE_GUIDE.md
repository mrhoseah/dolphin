# 🐬 Dolphin Framework - Comprehensive Guide

**Complete documentation for building web applications with Dolphin Framework**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Core Features](#core-features)
5. [ORM & Database](#orm--database)
6. [Templates](#templates)
7. [Authentication](#authentication)
8. [Advanced Features](#advanced-features)
9. [Deployment](#deployment)
10. [API Reference](#api-reference)

---

## Introduction

Dolphin Framework is a modern, enterprise-grade web framework written in Go, taking inspiration from productive web frameworks like Laravel. It combines Go's performance and concurrency capabilities with a batteries-included developer workflow.

### Why Dolphin?

Building web applications shouldn't feel like navigating through endless documentation, configuring complex build systems, or wrestling with boilerplate code. Dolphin brings a polished developer experience to Go—so rapid development feels natural and delightful.

### Key Features

- **🚀 Rapid Development**: Built-in scaffolding and code generation
- **🗄️ Database Migrations**: Built-in migration system with GORM integration
- **🔄 Active Record ORM**: GORM-based ORM with repository pattern
- **🛡️ Middleware System**: Comprehensive middleware for auth, CORS, logging, and more
- **📱 Frontend Integration**: Built-in support for Vue.js, React.js, and Tailwind CSS
- **📚 API Documentation**: Automatic Swagger/OpenAPI documentation
- **🎨 HTMX Support**: Modern web interactions without heavy JavaScript
- **🔐 Authentication**: Multi-guard authentication system
- **💾 Caching**: Redis and memory-based caching
- **📊 Session Management**: Cookie and database session storage

---

## Installation

### Windows

**PowerShell (Recommended):**
```powershell
irm https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.ps1 | iex
```

**Manual:**
```powershell
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest
```

### macOS

**Automated Installer:**
```bash
curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install-mac.sh | bash
```

**Homebrew:**
```bash
brew install go
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest
```

### Linux

**Automated Installer:**
```bash
curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh | bash
```

**Manual:**
```bash
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest
```

### Verify Installation

```bash
dolphin --version
dolphin --help
```

---

## Quick Start

### Create Your First Project

```bash
# Create a new project
dolphin new my-app
cd my-app

# Create project with authentication
dolphin new my-app --auth

# Start development server
dolphin serve
```

Visit `http://localhost:8080` to see your application running!

### Generate a Complete Module

```bash
# Generate a complete module (model, controller, views, migration)
dolphin make:module Product

# Generate an API resource
dolphin make:resource User
```

---

## Core Features

### Routing

Dolphin uses the Chi router for fast, flexible routing:

```go
// In bootstrap/routes.go
func SetupRoutes(r *router.Router, cfg *config.Config, logger *zap.Logger, db *gorm.DB) {
    chiRouter := r.GetRouter()
    
    // Simple route
    chiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, Dolphin!"))
    })
    
    // Route groups
    chiRouter.Route("/api/v1", func(api chi.Router) {
        api.Get("/users", handleUsers)
        api.Post("/users", createUser)
    })
    
    // Protected routes
    chiRouter.Group(func(protected chi.Router) {
        protected.Use(authMiddleware.Authenticate)
        protected.Get("/dashboard", handleDashboard)
    })
}
```

### Controllers

Controllers handle request logic:

```go
type UserController struct {
    db *gorm.DB
}

func (c *UserController) Index(w http.ResponseWriter, r *http.Request) {
    var users []models.User
    c.db.Find(&users)
    json.NewEncoder(w).Encode(users)
}
```

### Middleware

Built-in middleware for common tasks:

```go
// Authentication middleware
authMiddleware := dolphinMiddleware.NewAuthMiddleware(authManager, logger)

// CORS middleware (already included)
// Logging middleware (already included)
// Recovery middleware (already included)
```

### Request & Response

Working with requests:

```go
// Parse form data
r.ParseForm()
name := r.FormValue("name")

// Parse JSON
var user models.User
json.NewDecoder(r.Body).Decode(&user)

// Send JSON response
json.NewEncoder(w).Encode(map[string]interface{}{
    "success": true,
    "data": user,
})
```

---

## ORM & Database

### Models

Define models with GORM:

```go
type User struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"not null" json:"name"`
    Email     string    `gorm:"unique;not null" json:"email"`
    Password  string    `gorm:"not null" json:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Relationships

Define relationships:

```go
type Post struct {
    ID        uint      `gorm:"primaryKey"`
    Title     string
    UserID    uint
    User      User      `gorm:"foreignKey:UserID"`
    Comments  []Comment `gorm:"foreignKey:PostID"`
}

type Comment struct {
    ID     uint
    PostID uint
    Post   Post `gorm:"foreignKey:PostID"`
    UserID uint
    User   User `gorm:"foreignKey:UserID"`
}
```

### Migrations

Run migrations:

```bash
# Run migrations
dolphin migrate

# Rollback
dolphin rollback

# Check status
dolphin status
```

Or programmatically:

```go
db.AutoMigrate(&models.User{}, &models.Post{})
```

### Query Builder

Build complex queries:

```go
// Where clauses
db.Where("age > ?", 18).Find(&users)

// Joins
db.Joins("Company").Find(&users)

// Preload relationships
db.Preload("Posts").Preload("Posts.Comments").Find(&users)

// Scopes
db.Where("active = ?", true).Find(&users)
```

---

## Templates

### Fin Templates

Dolphin uses Fin templates (similar to Blade/Laravel):

```html
<!-- views/pages/home.fin.html -->
{{extend "layouts/base.fin.html"}}

{{define "title"}}Home - My App{{end}}

{{define "content"}}
<div class="container">
    <h1>Welcome, {{.User.Name}}!</h1>
    
    {{if .User}}
        <p>You are logged in.</p>
    {{else}}
        <a href="/auth/login">Login</a>
    {{end}}
</div>
{{end}}
```

### Template Directives

```html
<!-- Extend layout -->
{{extend "layouts/base.fin.html"}}

<!-- Define sections -->
{{define "title"}}Page Title{{end}}
{{define "content"}}Page Content{{end}}

<!-- Include partials -->
{{include "partials/header.fin.html"}}

<!-- Components -->
{{component "Button" "text" "Click Me" "color" "blue"}}
```

### HTMX Integration

Build reactive UIs without heavy JavaScript:

```html
<!-- HTMX form submission -->
<form hx-post="/api/posts" hx-swap="outerHTML">
    <input type="text" name="title" required>
    <button type="submit">Create Post</button>
</form>

<!-- HTMX button -->
<button hx-get="/api/posts" hx-target="#posts-list">
    Load Posts
</button>
```

---

## Authentication

### Setup Authentication

```go
// In bootstrap/routes.go
authManager := auth.NewAuthManager()

// Create session store
sessionStore := auth.NewMemorySessionStore()

// Create user provider
userProvider := auth.NewDatabaseProvider(db, &models.User{})

// Create session guard
sessionGuard := auth.NewSessionGuard("web", userProvider, sessionStore)
authManager.RegisterGuard("web", sessionGuard)
authManager.SetDefaultGuard("web")
```

### Authentication Routes

```go
// Login
authManager.Attempt(map[string]string{
    "email": "user@example.com",
    "password": "password",
})

// Check authentication
if authManager.Check() {
    user := authManager.User()
    // User is authenticated
}

// Logout
authManager.Logout()
```

### Protected Routes

```go
authMiddleware := dolphinMiddleware.NewAuthMiddleware(authManager, logger)

chiRouter.Group(func(protected chi.Router) {
    protected.Use(authMiddleware.Authenticate)
    protected.Get("/dashboard", handleDashboard)
})
```

---

## Advanced Features

### Dependency Injection

Dolphin includes a service container:

```go
// Register services
container.Bind("userService", func() interface{} {
    return NewUserService()
})

// Resolve services
userService := container.Make("userService").(*UserService)
```

### Event System

Dispatch and listen to events:

```go
// Dispatch event
eventBus.Dispatch("user.created", user)

// Listen to event
eventBus.Listen("user.created", func(user interface{}) {
    // Handle event
})
```

### Queue System

Process background jobs:

```go
// Dispatch job
queue.Dispatch(NewSendEmailJob(user))

// Process queue
queue.Process()
```

### Caching

Cache data for performance:

```go
// Cache data
cache.Put("users", users, 3600) // 1 hour

// Retrieve from cache
users, found := cache.Get("users")

// Cache with tags
cache.Tags("users", "active").Put("active_users", users)
```

---

## Deployment

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["./app", "serve"]
```

### Environment Configuration

```yaml
# config/config.yaml
app:
  environment: production
  debug: false

database:
  driver: postgres
  host: localhost
  port: 5432
  database: myapp
  username: user
  password: password
```

### Health Checks

```bash
# Health check endpoint (automatically available)
curl http://localhost:8080/health
```

---

## API Reference

### CLI Commands

```bash
# Create new project
dolphin new my-app

# Generate code
dolphin make:controller UserController
dolphin make:model User
dolphin make:migration create_users_table
dolphin make:middleware AdminMiddleware

# Database
dolphin migrate
dolphin rollback
dolphin status

# Server
dolphin serve

# Storage
dolphin storage:link
```

### Configuration

```yaml
app:
  name: "My App"
  environment: development
  debug: true
  url: "http://localhost:8080"

server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: sqlite
  database: app.db
```

---

## Best Practices

### Code Organization

```
app/
├── models/          # Database models
├── http/
│   └── controllers/ # Request handlers
├── repositories/    # Data access layer
└── services/        # Business logic

bootstrap/
└── routes.go        # Route configuration

views/
├── layouts/         # Base layouts
├── pages/           # Page templates
└── partials/        # Reusable components

config/
└── config.yaml      # Configuration
```

### Security

- Always hash passwords using `bcrypt`
- Use HTTPS in production
- Validate all user input
- Use CSRF protection
- Implement rate limiting
- Keep dependencies updated

### Performance

- Use database indexes
- Implement caching
- Use connection pooling
- Optimize queries (avoid N+1)
- Enable compression
- Use CDN for static assets

---

## Troubleshooting

### Common Issues

**"dolphin: command not found"**
- Add Go bin to PATH
- Restart terminal

**Database connection errors**
- Check database credentials
- Ensure database server is running
- Verify network connectivity

**Template not found**
- Check view path in configuration
- Verify file extension is `.fin.html`
- Check file permissions

---

## Resources

- **Documentation**: https://dolphin-docs.netlify.app/
- **GitHub**: https://github.com/mrhoseah/dolphin
- **Issues**: https://github.com/mrhoseah/dolphin/issues

---

## License

Dolphin Framework is open-source software. See LICENSE file for details.

---

**Built with ❤️ by the Dolphin Framework team**

*Last updated: 2025*

