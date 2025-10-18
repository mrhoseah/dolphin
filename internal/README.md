# Advanced Features Implementation

This directory contains the implementation of advanced framework features that make Dolphin Framework enterprise-grade.

## 🚀 Implemented Features

### 1. **Advanced Dependency Injection Container** (`internal/container/`)
- **Type-based resolution** using Go reflection
- **Service lifetimes**: Singleton, Scoped, Transient
- **Constructor injection** with automatic dependency resolution
- **Factory pattern** support for complex object creation
- **Thread-safe** container operations

**Key Features:**
- Automatic dependency resolution
- Circular dependency detection
- Service lifetime management
- Factory-based service creation

### 2. **Aspect-Oriented Programming (AOP)** (`internal/aop/`)
- **Method interception** and weaving
- **Cross-cutting concerns** separation
- **Built-in aspects**: Logging, Caching, Transactions, Performance
- **Aspect registry** for managing aspects
- **Dynamic proxy** generation

**Available Aspects:**
- `LoggingAspect` - Automatic method logging
- `CachingAspect` - Method result caching
- `TransactionAspect` - Automatic transaction management
- `PerformanceAspect` - Performance monitoring

### 3. **Auto-Configuration System** (`internal/autoconfig/`)
- **Conditional configuration** based on environment
- **Service auto-discovery** and registration
- **Priority-based** configuration application
- **Zero-configuration** startup

**Built-in Auto-Configurations:**
- `DatabaseAutoConfiguration` - Database services
- `CacheAutoConfiguration` - Caching services
- `SecurityAutoConfiguration` - Security services
- `WebAutoConfiguration` - Web services

### 4. **Application Lifecycle Management** (`internal/lifecycle/`)
- **Phase-based** application startup/shutdown
- **Lifecycle callbacks** and hooks
- **Graceful shutdown** handling
- **Resource management** automation

**Lifecycle Phases:**
- `PhaseInitialization` - Service registration
- `PhaseConfiguration` - Configuration loading
- `PhaseStartup` - Service initialization
- `PhaseRunning` - Application running
- `PhaseShutdown` - Graceful shutdown

### 5. **Advanced Transaction Management** (`internal/transaction/`)
- **Nested transactions** with savepoints
- **Retry logic** for transient failures
- **Transaction templates** for common patterns
- **Isolation levels** support
- **Distributed transactions** (basic implementation)

**Transaction Features:**
- Automatic retry on deadlocks
- Configurable isolation levels
- Transaction callbacks and monitoring
- Distributed transaction support

## 📁 Directory Structure

```
internal/
├── container/          # Advanced DI container
│   └── container.go
├── aop/               # Aspect-oriented programming
│   └── aspect.go
├── autoconfig/        # Auto-configuration system
│   └── autoconfig.go
├── lifecycle/         # Application lifecycle management
│   └── lifecycle.go
└── transaction/       # Advanced transaction management
    └── manager.go

examples/
└── advanced_features_example/
    └── main.go        # Complete example showing all features
```

## 🎯 Usage Examples

### Dependency Injection

```go
// Register services
container.RegisterSingleton(
    reflect.TypeOf((*EmailService)(nil)).Elem(),
    reflect.TypeOf((*emailServiceImpl)(nil)),
)

// Resolve services
emailService, err := container.Resolve(reflect.TypeOf((*EmailService)(nil)).Elem())
```

### Aspect-Oriented Programming

```go
// Create aspects
loggingAspect := aop.NewLoggingAspect(logger)
cachingAspect := aop.NewCachingAspect(cache)

// Register aspects
aspectRegistry := aop.NewAspectRegistry()
aspectRegistry.RegisterAspect(reflect.TypeOf((*UserService)(nil)).Elem(), loggingAspect)

// Apply aspects
userServiceWithAspects := aspectRegistry.ApplyAspects(userService)
```

### Auto-Configuration

```go
// Apply auto-configuration
autoConfigManager := autoconfig.NewAutoConfigurationManager()
err := autoConfigManager.ApplyAutoConfiguration(ctx, container)
```

### Lifecycle Management

```go
// Setup lifecycle
lifecycleManager := lifecycle.NewApplicationLifecycleManager(container)
lifecycleManager.RegisterCallback(lifecycle.PhaseStartup, &DatabaseLifecycleCallback{})

// Start application
err := lifecycleManager.Start(ctx)
```

### Advanced Transactions

```go
// Execute with retry logic
err := transactionTemplate.ExecuteWithRetry(ctx, 3, func(tx *sql.Tx) error {
    // Transaction logic
    return nil
})

// Execute with isolation level
err := transactionManager.ExecuteWithIsolation(ctx, transaction.ReadCommitted, func(tx *sql.Tx) error {
    // Transaction logic
    return nil
})
```

## 🔧 Integration with Existing Features

These advanced features integrate seamlessly with existing Dolphin Framework features:

- **Service Providers** - Enhanced with advanced DI
- **Middleware** - Can be applied as aspects
- **Configuration** - Enhanced with auto-configuration
- **Database** - Advanced transaction management
- **Logging** - Integrated with AOP logging
- **Metrics** - Performance aspects provide metrics

## 🚀 Running the Example

```bash
# Run the advanced features example
go run examples/advanced_features_example/main.go
```

This example demonstrates:
- Service registration and resolution
- Aspect application and execution
- Auto-configuration application
- Lifecycle management
- Transaction management with callbacks

## 📚 Documentation

For detailed documentation on each feature, see:
- [Advanced Framework Features Guide](../website/docs/advanced/advanced-framework-features.md)

## 🎯 Benefits

These advanced features provide:

- **Developer Productivity** - Auto-configuration and DI reduce boilerplate
- **Maintainability** - AOP separates cross-cutting concerns
- **Reliability** - Advanced transaction management ensures data integrity
- **Flexibility** - Lifecycle management provides fine-grained control
- **Testability** - DI and aspect weaving enable easy testing
- **Enterprise Readiness** - Production-grade features for serious applications

The implementation follows Go best practices and integrates seamlessly with the existing Dolphin Framework architecture.
