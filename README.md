# 🐬 Dolphin Framework

**Enterprise-grade Go web framework for rapid development**

Dolphin Framework is a modern, enterprise-grade web framework written in Go, inspired by the elegant syntax and developer experience of Laravel, CodeIgniter, and CakePHP. It combines Go's performance and concurrency capabilities with the productivity features of modern PHP frameworks.

## ✨ Key Features

- **🚀 Rapid Development**: Built-in scaffolding and code generation
- **🗄️ Database Migrations**: Integrated with [Raptor](https://github.com/mrhoseah/raptor) for Laravel-style migrations
- **🔄 Active Record ORM**: GORM-based ORM with repository pattern
- **🛡️ Middleware System**: Comprehensive middleware for auth, CORS, logging, and more
- **📱 Frontend Integration**: Built-in support for Vue.js, React.js, and Tailwind CSS
- **📚 API Documentation**: Automatic Swagger/OpenAPI documentation with SwagGo
- **⚡ High Performance**: Built on Go's concurrency and performance
- **🔧 CLI Tools**: Powerful command-line interface for development
- **📦 Dependency Injection**: Service container for clean architecture
- **🔐 Authentication**: JWT-based authentication system
- **💾 Caching**: Redis and memory-based caching
- **📊 Session Management**: Cookie and database session storage

## 🚀 Quick Start

### Installation

#### Option 1: Install CLI Tool (Recommended)
```bash
# Install the Dolphin CLI globally
go install github.com/mrhoseah/dolphin/cmd/cli@latest

# Create a new project
dolphin new my-app
cd my-app

# Install dependencies
go mod tidy

# Start the development server
dolphin serve
```

#### Option 2: Clone Repository
```bash
# Clone the repository
git clone https://github.com/mrhoseah/dolphin.git
cd dolphin

# Install dependencies
go mod tidy

# Copy environment configuration
cp env.example .env

# Run the application
go run main.go serve
```

### Configuration

Configure your application using `config/config.yaml` or environment variables:

```yaml
# Database Configuration
database:
  driver: "postgres"  # postgres, mysql, sqlite
  host: "localhost"
  port: 5432
  database: "dolphin"
  username: "postgres"
  password: "password"

# Server Configuration
server:
  host: "localhost"
  port: 8080
```

## 🔧 CLI Commands (Dolphin CLI - Like Laravel Artisan)

Dolphin provides a powerful CLI tool similar to Laravel's Artisan. Install it globally for the best experience:

```bash
# Install CLI globally
go install github.com/mrhoseah/dolphin/cmd/cli@latest

# Create new project
dolphin new my-awesome-app
```

### 🚀 Development Commands

```bash
# Start development server
dolphin serve
dolphin serve --port 3000 --host 0.0.0.0

# List all available commands
dolphin list

# Show version information
dolphin version
```

### 🗄️ Database Commands

```bash
# Run migrations
dolphin migrate
dolphin migrate --force

# Rollback migrations
dolphin rollback
dolphin rollback --steps 3

# Check migration status
dolphin status

# Fresh start (DESTRUCTIVE)
dolphin fresh

# Database operations
dolphin db:seed
dolphin db:wipe
```

### 🔨 Code Generation (Make Commands)

```bash
# Controllers
dolphin make:controller UserController
dolphin make:controller UserController --resource --api

# Models
dolphin make:model User
dolphin make:model User --migration --factory

# Migrations
dolphin make:migration create_users_table
dolphin make:migration add_email_to_users_table

# Middleware
dolphin make:middleware AuthMiddleware

# Seeders
dolphin make:seeder UserSeeder

# Form Requests
dolphin make:request UserRequest
```

### 📚 Documentation & Utilities

```bash
# Generate Swagger documentation
dolphin swagger

# Cache management
dolphin cache:clear
dolphin cache:warm

# Route listing
dolphin route:list

# Security
dolphin key:generate
```

### 🎯 Laravel Artisan Comparison

| Laravel Artisan | Dolphin CLI | Description |
|----------------|-------------|-------------|
| `php artisan serve` | `dolphin serve` | Start development server |
| `php artisan migrate` | `dolphin migrate` | Run database migrations |
| `php artisan make:controller` | `dolphin make:controller` | Create controller |
| `php artisan make:model` | `dolphin make:model` | Create model |
| `php artisan make:migration` | `dolphin make:migration` | Create migration |
| `php artisan make:middleware` | `dolphin make:middleware` | Create middleware |
| `php artisan route:list` | `dolphin route:list` | List routes |
| `php artisan cache:clear` | `dolphin cache:clear` | Clear cache |
| `php artisan key:generate` | `dolphin key:generate` | Generate app key |

## 🏗️ Architecture

### Core Components

- **Application**: Main application container and lifecycle management
- **Router**: HTTP routing with middleware support
- **Database**: GORM integration with migration support
- **ORM**: Active Record pattern with repository design
- **Validation**: Comprehensive validation system
- **Cache**: Redis and memory caching
- **Session**: Session management with multiple backends
- **Frontend**: Vue.js, React.js, and Tailwind CSS integration

### Project Structure

```
dolphin/
├── cmd/dolphin/          # CLI application
├── internal/             # Internal packages
│   ├── app/              # Application core
│   ├── config/           # Configuration management
│   ├── database/         # Database and migrations
│   ├── orm/              # ORM and repositories
│   ├── router/           # HTTP routing
│   ├── middleware/       # Middleware components
│   ├── validation/       # Validation system
│   ├── cache/            # Caching system
│   ├── session/          # Session management
│   ├── logger/           # Logging system
│   └── frontend/         # Frontend integration
├── app/                  # Application code
│   ├── http/controllers/ # HTTP controllers
│   ├── models/           # Data models
│   └── middleware/        # Custom middleware
├── migrations/           # Database migrations
├── config/               # Configuration files
└── public/               # Static assets
```

## 📚 Usage Examples

### Creating a Controller

```bash
go run main.go make:controller UserController
```

This generates:

```go
package controllers

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/render"
)

type UserController struct{}

func (c *UserController) Index(w http.ResponseWriter, r *http.Request) {
    render.JSON(w, r, map[string]interface{}{
        "message": "List of users",
        "data":    []interface{}{},
    })
}
```

### Creating a Model

```bash
go run main.go make:model User
```

This generates:

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
    
    // Add your fields here
    Name  string `gorm:"not null" json:"name"`
    Email string `gorm:"uniqueIndex" json:"email"`
}
```

### Database Migrations

```bash
go run main.go make:migration create_users_table
```

This generates:

```go
package migrations

import (
    raptor "github.com/mrhoseah/raptor/core"
)

type CreateUsersTable struct{}

func (m *CreateUsersTable) Name() string {
    return "create_users_table"
}

func (m *CreateUsersTable) Up(s raptor.Schema) error {
    return s.CreateTable("users", []string{"id", "name", "email", "created_at"})
}

func (m *CreateUsersTable) Down(s raptor.Schema) error {
    return s.DropTable("users")
}
```

### Using the ORM

```go
// Repository pattern
type UserRepository struct {
    *orm.Repository[User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        Repository: orm.NewRepository(db, User{}),
    }
}

// Usage
userRepo := NewUserRepository(db)
user, err := userRepo.Find(ctx, 1)
users, err := userRepo.FindAll(ctx)
```

### API Documentation with Swagger

Dolphin includes automatic API documentation generation using SwagGo:

#### Generate Documentation
```bash
# Install SwagGo
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g main.go

# Start server and visit documentation
go run main.go serve
# Visit: http://localhost:8080/swagger/index.html
```

#### API Endpoints

**Authentication:**
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh JWT token

**Users:**
- `GET /api/v1/users` - List all users
- `POST /api/v1/users` - Create new user
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

**Protected Routes:**
- `GET /api/v1/protected/user` - Get current user (requires JWT)
- `PUT /api/v1/protected/user` - Update current user
- `DELETE /api/v1/protected/user` - Delete current user

#### Example API Usage

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Get users (with JWT token)
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Frontend Integration

Dolphin includes built-in support for modern frontend frameworks:

#### Vue.js Integration
```go
vueApp := frontend.NewVueJSIntegration("My App", "3.3.4")
html := vueApp.GenerateVueApp()
```

#### React.js Integration
```go
reactApp := frontend.NewReactJSIntegration("My App", "18.2.0")
html := reactApp.GenerateReactApp()
```

#### Tailwind CSS Integration
```go
tailwind := frontend.NewTailwindCSSIntegration("3.3.0")
config := tailwind.GenerateTailwindConfig()
```

## 🔧 Configuration

### Environment Variables

```bash
# Application
APP_NAME=Dolphin Framework
APP_ENV=development
APP_DEBUG=true
APP_URL=http://localhost:8080

# Database
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=dolphin
DB_USERNAME=postgres
DB_PASSWORD=password

# JWT
JWT_SECRET=your-jwt-secret-key-here
JWT_EXPIRATION=24h
```

### YAML Configuration

```yaml
app:
  name: "Dolphin Framework"
  environment: "development"
  debug: true

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  database: "dolphin"

server:
  host: "localhost"
  port: 8080
```

## 🚀 Deployment

### Docker Support

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o dolphin main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dolphin .
CMD ["./dolphin", "serve"]
```

### Production Considerations

- Set `APP_ENV=production`
- Use HTTPS in production
- Configure proper database credentials
- Set up Redis for caching
- Use environment variables for secrets

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Inspired by Laravel's elegant architecture
- Built with Go's performance and concurrency
- Integrates with [Raptor](https://github.com/mrhoseah/raptor) for migrations
- Uses modern Go libraries and best practices

## 📞 Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Join our community discussions

---

**Dolphin Framework** - Making Go web development as elegant as Laravel! 🐬