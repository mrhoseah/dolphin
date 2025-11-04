# Dolphin Framework Public API

This directory contains the public API for Dolphin Framework, allowing external applications to use Dolphin features.

## Available Packages

### Core
- **`app`** - Application instance and lifecycle
- **`config`** - Configuration management
- **`database`** - Database manager and connection
- **`logger`** - Logging utilities
- **`router`** - HTTP router and middleware

### Authentication & Security
- **`auth`** - Authentication (MFA, OAuth2, Guards, Providers)
- **`apikey`** - API key management
- **`middleware`** - HTTP middleware (API key auth, etc.)

### Real-time & Communication
- **`realtime`** - WebSocket and Server-Sent Events (SSE)
- **`grpc`** - gRPC server and gateway

### Observability
- **`observability`** - OpenTelemetry integration (OTLP, Prometheus)

### Storage
- **`storage`** - File storage and symlink management

## Usage Example

```go
package main

import (
    "dolphin/pkg/app"
    "dolphin/pkg/config"
    "dolphin/pkg/database"
    "dolphin/pkg/logger"
    "dolphin/pkg/router"
    "dolphin/pkg/auth"
    "dolphin/pkg/apikey"
    "dolphin/pkg/realtime"
    "dolphin/pkg/grpc"
    "dolphin/pkg/observability"
)

func main() {
    // Load config
    cfg, _ := config.Load()
    
    // Create logger
    log := logger.New(cfg.Log.Level, cfg.Log.Format)
    
    // Connect database
    db, _ := database.New(&cfg.Database)
    
    // Create application
    app := app.New(cfg, log, db)
    
    // Setup MFA
    mfaManager := auth.NewMFAManager(db.GetDB(), "My App")
    
    // Setup OAuth
    oauthManager := auth.NewOAuthManager(db.GetDB())
    googleProvider := auth.NewGoogleProvider(
        "client-id",
        "client-secret",
        "http://localhost:8080/auth/google/callback",
        []string{"openid", "profile", "email"},
    )
    oauthManager.RegisterProvider("google", googleProvider)
    
    // Setup API Keys
    apiKeyManager := apikey.NewAPIKeyManager(db.GetDB())
    
    // Setup Real-time
    hub := realtime.NewHub(log)
    go hub.Run(context.Background())
    
    // Setup gRPC
    grpcConfig := grpc.DefaultServerConfig()
    grpcServer, _ := grpc.NewServer(grpcConfig, log)
    
    // Setup OpenTelemetry
    otlpConfig := observability.DefaultOTLPConfig()
    otlpConfig.Enabled = true
    tp, _ := observability.SetupOTLPTracing(
        context.Background(),
        otlpConfig,
        "my-service",
        "1.0.0",
        "production",
        log,
    )
    
    // Create router
    r := router.New(app)
    
    // Use features...
}
```

## Features Available

All features implemented in Dolphin Framework are available through these packages:

✅ **Authentication**
- Multi-guard authentication
- JWT tokens
- Session management
- Two-Factor Authentication (MFA)
- OAuth2 Social Login

✅ **API Management**
- API key generation and validation
- Scope-based permissions
- Rate limiting per key

✅ **Real-time Communication**
- WebSocket server
- Server-Sent Events (SSE)
- Channel broadcasting

✅ **gRPC**
- gRPC server
- gRPC-Gateway for REST compatibility

✅ **Observability**
- OpenTelemetry tracing
- Prometheus metrics
- OTLP export

✅ **Storage**
- Multi-driver storage
- Symlink management

