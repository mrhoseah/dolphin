# Using Dolphin Framework for OddMetrics

This guide shows you how to use Dolphin framework to build your OddMetrics project.

## Quick Start

### 1. Initialize Your Project

```bash
# Create your oddmetrics project
mkdir oddmetrics && cd oddmetrics
go mod init github.com/yourusername/oddmetrics

# Add Dolphin as a dependency
go get github.com/mrhoseah/dolphin
```

### 2. Project Structure (CodeIgniter-like)

```
oddmetrics/
├── app/
│   ├── http/
│   │   └── controllers/
│   │       ├── HomeController.go
│   │       └── MetricsController.go
│   ├── models/
│   │   └── Metric.go
│   └── views/
│       ├── layouts/
│       │   └── app.fin.html
│       └── pages/
│           ├── home.fin.html
│           └── dashboard.fin.html
├── config/
│   └── config.yaml
├── routes/
│   └── web.go
├── main.go
└── go.mod
```

### 3. Main Application File

```go
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrhoseah/dolphin/internal/app"
	"github.com/mrhoseah/dolphin/internal/config"
	"github.com/mrhoseah/dolphin/internal/database"
	"github.com/mrhoseah/dolphin/internal/logger"
	"github.com/mrhoseah/dolphin/internal/router"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize application
	app := app.New(cfg, logger, db)

	// Initialize router
	r := router.New(app)

	// Setup your routes (CodeIgniter-like simple API)
	setupRoutes(r)

	// Setup default routes
	r.SetupRoutes()

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("🚀 OddMetrics server running", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	app.Close()
}
```

### 4. Routes Setup (Simple CodeIgniter-style)

Create `routes/web.go`:

```go
package routes

import (
	"net/http"

	"github.com/mrhoseah/dolphin/internal/router"
	"oddmetrics/app/http/controllers"
)

func SetupRoutes(r *router.Router) {
	// Get simple routes helper
	routes := r.SimpleRoutes()

	// Public routes
	routes.Get("/", controllers.HomeController{}.Index)
	routes.View("/about", "pages/about")

	// Template routes (automatically protected with @auth if directive exists)
	routes.View("/dashboard", "pages/dashboard")
	routes.View("/metrics", "pages/metrics")

	// Authenticated routes
	routes.Auth("/profile", controllers.ProfileController{}.Show)
	routes.AuthPost("/profile/update", controllers.ProfileController{}.Update)

	// Guest-only routes
	routes.Guest("/login", controllers.AuthController{}.ShowLogin)
	routes.GuestPost("/login", controllers.AuthController{}.HandleLogin)

	// Route groups
	routes.AuthGroup("/api", func(group *router.RouteGroup) {
		group.Get("/metrics", controllers.MetricsController{}.Index)
		group.Post("/metrics", controllers.MetricsController{}.Create)
		group.Get("/metrics/{id}", controllers.MetricsController{}.Show)
	})

	// RESTful resources
	routes.ResourceWithAuth("/metrics", router.ResourceController{
		Index:   controllers.MetricsController{}.Index,
		Create:  controllers.MetricsController{}.Create,
		Store:   controllers.MetricsController{}.Store,
		Show:    controllers.MetricsController{}.Show,
		Edit:    controllers.MetricsController{}.Edit,
		Update:  controllers.MetricsController{}.Update,
		Destroy: controllers.MetricsController{}.Destroy,
	})

	// Role-based routes
	routes.RoleGroup("/admin", "admin", func(group *router.RouteGroup) {
		group.Get("/users", controllers.AdminController{}.Users)
		group.Get("/settings", controllers.AdminController{}.Settings)
	})
}
```

### 5. Controller Example

```go
package controllers

import (
	"net/http"
)

type HomeController struct{}

func (c HomeController) Index(w http.ResponseWriter, r *http.Request) {
	// Simple handler
	w.Write([]byte("Welcome to OddMetrics!"))
}
```

### 6. Fin Template with @auth Protection

Create `app/views/pages/dashboard.fin.html`:

```html
@auth
@extends('layouts/app')

@section('title', 'Dashboard')

@section('content')
<div class="container">
    <h1>Welcome, {{.User.Name}}!</h1>
    
    @if .Metrics
        <div class="metrics">
            @foreach .Metrics as metric
                <div class="metric-card">
                    <h3>{{metric.Name}}</h3>
                    <p>{{metric.Value}}</p>
                </div>
            @endforeach
        </div>
    @else
        <p>No metrics yet.</p>
    @endif
</div>
@endsection
```

The `@auth` directive at the top automatically protects this route - unauthenticated users will be redirected to `/auth/login`.

### 7. Layout Template

Create `app/views/layouts/app.fin.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>@yield('title') - OddMetrics</title>
    <link rel="stylesheet" href="{{asset('css/app.css')}}">
    @stack('styles')
</head>
<body>
    <nav>
        <a href="/">Home</a>
        @auth
            <a href="/dashboard">Dashboard</a>
            <a href="/profile">Profile</a>
            <a href="/logout">Logout</a>
        @else
            <a href="/login">Login</a>
            <a href="/register">Register</a>
        @endif
    </nav>

    <main>
        @yield('content')
    </main>

    @stack('scripts')
</body>
</html>
```

### 8. Using Template Helpers

Fin templates have Laravel-like helpers:

```html
<!-- URL helpers -->
<a href="{{url('/dashboard')}}">Dashboard</a>
<img src="{{asset('images/logo.png')}}">

<!-- String helpers -->
<p>{{str_limit .Description 100}}</p>
<p>{{str_upper .Title}}</p>

<!-- Date helpers -->
<p>Created {{time_ago .CreatedAt}}</p>
<p>{{date_format "2006-01-02" .CreatedAt}}</p>

<!-- Auth helpers -->
@auth
    <p>Welcome, {{.User.Name}}!</p>
@else
    <p>Please login</p>
@endif

<!-- CSRF protection -->
<form method="POST" action="/metrics">
    {{csrf_field}}
    <!-- form fields -->
</form>
```

### 9. Model Example

```go
package models

import (
	"time"
	"gorm.io/gorm"
)

type Metric struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Value     float64        `json:"value"`
	UserID    uint           `json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

### 10. Available Route Methods

```go
routes := r.SimpleRoutes()

// Basic routes
routes.Get(path, handler)
routes.Post(path, handler)
routes.Put(path, handler)
routes.Patch(path, handler)
routes.Delete(path, handler)
routes.Any(path, handler)  // All methods

// Protected routes
routes.Auth(path, handler)           // Requires authentication
routes.AuthPost(path, handler)       // POST with auth
routes.Guest(path, handler)          // Guest only
routes.Role(path, role, handler)     // Requires role
routes.Permission(path, perm, handler) // Requires permission

// Template routes
routes.View(path, template)          // Render template
routes.ViewWithAuth(path, template)  // Render template with auth

// Route groups
routes.AuthGroup(prefix, func(group) { ... })
routes.GuestGroup(prefix, func(group) { ... })
routes.RoleGroup(prefix, role, func(group) { ... })
routes.PermissionGroup(prefix, perm, func(group) { ... })

// RESTful resources
routes.Resource(path, controller)
routes.ResourceWithAuth(path, controller)
```

### 11. Template Directives (Laravel-like)

```html
<!-- Layouts -->
@extends('layouts/app')
@section('content') ... @endsection
@yield('content')

<!-- Conditionals -->
@if .Condition
    ...
@elseif .OtherCondition
    ...
@else
    ...
@endif

<!-- Loops -->
@foreach .Items as item
    {{item.Name}}
@endforeach

<!-- Components -->
@component('components/alert')
    @slot('title') Error @endslot
    Something went wrong!
@endcomponent

<!-- Includes -->
@include('partials/header')

<!-- Stacks -->
@push('scripts')
    <script src="custom.js"></script>
@endpush
@stack('scripts')

<!-- Auth -->
@auth
    Authenticated content
@else
    Guest content
@endif

<!-- Permissions -->
@can('edit', .Post)
    <a href="/posts/{{.Post.ID}}/edit">Edit</a>
@endcan

<!-- Roles -->
@hasrole('admin')
    Admin panel
@endhasrole

<!-- Translations -->
@lang('messages.welcome')
@choice('messages.items', .Count)
```

### 12. Configuration

Create `config/config.yaml`:

```yaml
app:
  name: "OddMetrics"
  environment: "development"
  debug: true
  key: "your-secret-key-here-change-in-production"

server:
  port: 8080
  read_timeout: 15
  write_timeout: 15
  idle_timeout: 60

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  database: "oddmetrics"
  username: "postgres"
  password: "password"
  ssl_mode: "disable"

log:
  level: "info"
  format: "json"
```

## Key Features for OddMetrics

1. **Simple Routes**: CodeIgniter-like straightforward route registration
2. **Template Protection**: `@auth` directive automatically protects templates
3. **Laravel-like Templates**: All the template features you need
4. **Full-Stack Framework**: Database, auth, queues, events, mail, etc.
5. **Easy to Use**: Straightforward app structure

## Next Steps

1. Set up your database models
2. Create your controllers
3. Design your Fin templates
4. Configure authentication
5. Add your business logic

Happy coding with Dolphin! 🐬

