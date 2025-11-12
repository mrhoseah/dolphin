package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrhoseah/dolphin/internal/graceful"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic Graceful Shutdown
	fmt.Println("=== Example 1: Basic Graceful Shutdown ===")

	// Create shutdown configuration
	config := graceful.DefaultShutdownConfig()
	config.ShutdownTimeout = 30 * time.Second
	config.DrainTimeout = 5 * time.Second
	config.LogShutdownEvents = true

	// Create shutdown manager
	shutdownManager := graceful.NewShutdownManager(config, logger)

	// Start shutdown manager
	if err := shutdownManager.Start(); err != nil {
		log.Fatalf("Failed to start shutdown manager: %v", err)
	}

	// Example 2: Graceful HTTP Server
	fmt.Println("\n=== Example 2: Graceful HTTP Server ===")

	// Create HTTP server
	httpServer := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate work
			time.Sleep(100 * time.Millisecond)

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"message":"Hello, World!","timestamp":"%s"}`,
				time.Now().Format(time.RFC3339))
		}),
	}

	// Create graceful server
	gracefulServer := graceful.NewGracefulServer(httpServer, config, logger)

	// Register with shutdown manager
	shutdownManager.RegisterHTTPServer(httpServer)

	// Start server in goroutine
	go func() {
		fmt.Println("Starting graceful HTTP server on :8080")
		if err := gracefulServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Example 3: Custom Shutdownable Service
	fmt.Println("\n=== Example 3: Custom Shutdownable Service ===")

	// Create a custom service
	databaseService := &DatabaseService{
		name:   "database",
		logger: logger,
	}

	// Register service
	shutdownManager.RegisterService(databaseService)

	// Example 4: Connection Tracking
	fmt.Println("\n=== Example 4: Connection Tracking ===")

	// Monitor connections
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := gracefulServer.GetConnectionStats()
			fmt.Printf("Connection Stats: %+v\n", stats)
		}
	}()

	// Example 5: Health Check Integration
	fmt.Println("\n=== Example 5: Health Check Integration ===")

	// Set health status
	shutdownManager.SetHealthStatus(true)

	// Simulate health check endpoint
	healthServer := &http.Server{
		Addr: ":8081",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			healthy := shutdownManager.GetHealthStatus()
			if healthy {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`,
					time.Now().Format(time.RFC3339))
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, `{"status":"unhealthy","timestamp":"%s"}`,
					time.Now().Format(time.RFC3339))
			}
		}),
	}

	// Register health server
	shutdownManager.RegisterHTTPServer(healthServer)

	// Start health server
	go func() {
		fmt.Println("Starting health check server on :8081")
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health server error: %v", err)
		}
	}()

	// Example 6: Multiple Servers Management
	fmt.Println("\n=== Example 6: Multiple Servers Management ===")

	// Create server manager
	serverManager := graceful.NewGracefulServerManager(config, logger)

	// Add servers to manager
	serverManager.AddServer(gracefulServer)

	// Create additional server
	apiServer := &http.Server{
		Addr: ":8082",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"api":"v1","endpoint":"%s"}`, r.URL.Path)
		}),
	}

	gracefulAPIServer := graceful.NewGracefulServer(apiServer, config, logger)
	serverManager.AddServer(gracefulAPIServer)

	// Start additional server
	go func() {
		fmt.Println("Starting API server on :8082")
		if err := gracefulAPIServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	// Example 7: Signal Handling
	fmt.Println("\n=== Example 7: Signal Handling ===")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Application running. Press Ctrl+C to trigger graceful shutdown...")
	fmt.Println("")
	fmt.Println("Available endpoints:")
	fmt.Println("  • Main server: http://localhost:8080/")
	fmt.Println("  • Health check: http://localhost:8081/health")
	fmt.Println("  • API server: http://localhost:8082/")
	fmt.Println("")

	// Wait for signal
	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v\n", sig)

	// Set health status to unhealthy
	shutdownManager.SetHealthStatus(false)

	// Initiate graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Initiating graceful shutdown...")

	if err := shutdownManager.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	} else {
		fmt.Println("Graceful shutdown completed successfully!")
	}

	// Example 8: Connection Draining Demonstration
	fmt.Println("\n=== Example 8: Connection Draining Demonstration ===")

	// Create drain config
	drainConfig := graceful.DefaultDrainConfig()
	drainConfig.DrainTimeout = 5 * time.Second
	drainConfig.MaxDrainWait = 10 * time.Second
	drainConfig.LogDrainEvents = true

	// Create connection tracker
	tracker := graceful.NewConnectionTracker(drainConfig, logger)

	// Simulate some connections
	fmt.Println("Simulating connections...")
	for i := 0; i < 5; i++ {
		// In a real scenario, these would be actual network connections
		// For demo purposes, we'll just track them
		fmt.Printf("Connection %d tracked\n", i+1)
	}

	// Start draining
	fmt.Println("Starting connection draining...")
	if err := tracker.StartDraining(context.Background()); err != nil {
		log.Printf("Failed to start draining: %v", err)
	}

	// Wait for draining to complete
	if err := tracker.WaitForDraining(context.Background()); err != nil {
		log.Printf("Draining error: %v", err)
	} else {
		fmt.Println("Connection draining completed!")
	}

	fmt.Println("\n🎉 All graceful shutdown examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin graceful status' to view shutdown status")
	fmt.Println("2. Use 'dolphin graceful test' to test shutdown process")
	fmt.Println("3. Use 'dolphin graceful config' to view configuration")
	fmt.Println("4. Integrate GracefulServer in your application")
	fmt.Println("5. Implement Shutdownable interface for your services")
	fmt.Println("6. Configure appropriate timeouts for your use case")
}

// DatabaseService implements the Shutdownable interface
type DatabaseService struct {
	name   string
	logger *zap.Logger
}

func (ds *DatabaseService) Shutdown(ctx context.Context) error {
	ds.logger.Info("Shutting down database service",
		zap.String("service", ds.name))

	// Simulate database shutdown
	time.Sleep(2 * time.Second)

	ds.logger.Info("Database service shutdown completed",
		zap.String("service", ds.name))

	return nil
}

func (ds *DatabaseService) Name() string {
	return ds.name
}
