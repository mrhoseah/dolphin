package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mrhoseah/dolphin/internal/circuitbreaker"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic Circuit Breaker
	fmt.Println("=== Example 1: Basic Circuit Breaker ===")

	// Create circuit breaker configuration
	config := circuitbreaker.DefaultConfig()
	config.FailureThreshold = 3
	config.SuccessThreshold = 2
	config.OpenTimeout = 10 * time.Second
	config.HalfOpenTimeout = 5 * time.Second
	config.RequestTimeout = 2 * time.Second
	config.EnableLogging = true

	// Create circuit breaker
	circuit := circuitbreaker.NewCircuitBreaker("user-service", config, logger)

	// Example 2: Execute with Circuit Breaker
	fmt.Println("\n=== Example 2: Execute with Circuit Breaker ===")

	// Simulate successful calls
	for i := 0; i < 3; i++ {
		result, err := circuit.Execute(context.Background(), func() (interface{}, error) {
			// Simulate successful service call
			time.Sleep(100 * time.Millisecond)
			return fmt.Sprintf("Success %d", i+1), nil
		})

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Result: %v\n", result)
		}
	}

	// Example 3: Simulate Failures
	fmt.Println("\n=== Example 3: Simulate Failures ===")

	// Simulate failures to trigger circuit opening
	for i := 0; i < 5; i++ {
		result, err := circuit.Execute(context.Background(), func() (interface{}, error) {
			// Simulate service failure
			time.Sleep(50 * time.Millisecond)
			return nil, fmt.Errorf("service unavailable")
		})

		if err != nil {
			fmt.Printf("Error: %v (State: %s)\n", err, circuit.GetState().String())
		} else {
			fmt.Printf("Result: %v\n", result)
		}
	}

	// Example 4: Circuit Manager
	fmt.Println("\n=== Example 4: Circuit Manager ===")

	// Create circuit breaker manager
	managerConfig := circuitbreaker.DefaultManagerConfig()
	manager := circuitbreaker.NewManager(managerConfig, logger)

	// Create multiple circuit breakers
	userServiceCircuit, _ := manager.CreateCircuit("user-service", config)
	orderServiceCircuit, _ := manager.CreateCircuit("order-service", config)
	paymentServiceCircuit, _ := manager.CreateCircuit("payment-service", config)

	fmt.Printf("Created circuits: %v\n", manager.GetCircuitNames())

	// Example 5: HTTP Client with Circuit Breaker
	fmt.Println("\n=== Example 5: HTTP Client with Circuit Breaker ===")

	// Create HTTP client with circuit breaker
	httpConfig := circuitbreaker.DefaultHTTPClientConfig()
	httpClient := circuitbreaker.NewHTTPClient("api-client", config, httpConfig, logger)

	// Simulate HTTP requests
	ctx := context.Background()

	// This would normally make actual HTTP requests
	fmt.Println("Simulating HTTP requests with circuit breaker...")

	// Example 6: Async Execution
	fmt.Println("\n=== Example 6: Async Execution ===")

	// Execute async operations
	resultChan := circuit.ExecuteAsync(ctx, func() (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "Async result", nil
	})

	// Wait for result
	select {
	case result := <-resultChan:
		if result.Error != nil {
			fmt.Printf("Async Error: %v\n", result.Error)
		} else {
			fmt.Printf("Async Result: %v\n", result.Value)
		}
	case <-time.After(5 * time.Second):
		fmt.Println("Async operation timed out")
	}

	// Example 7: Circuit Breaker States
	fmt.Println("\n=== Example 7: Circuit Breaker States ===")

	// Show circuit breaker states
	fmt.Printf("User Service State: %s\n", userServiceCircuit.GetState())
	fmt.Printf("Order Service State: %s\n", orderServiceCircuit.GetState())
	fmt.Printf("Payment Service State: %s\n", paymentServiceCircuit.GetState())

	// Get statistics
	stats := circuit.GetStats()
	fmt.Printf("Circuit Stats: %+v\n", stats)

	// Example 8: Force Operations
	fmt.Println("\n=== Example 8: Force Operations ===")

	// Force open circuit
	circuit.ForceOpen()
	fmt.Printf("Circuit state after force open: %s\n", circuit.GetState())

	// Try to execute (should be rejected)
	_, err := circuit.Execute(ctx, func() (interface{}, error) {
		return "This should be rejected", nil
	})
	fmt.Printf("Execution result after force open: %v\n", err)

	// Force close circuit
	circuit.ForceClose()
	fmt.Printf("Circuit state after force close: %s\n", circuit.GetState())

	// Example 9: Metrics
	fmt.Println("\n=== Example 9: Metrics ===")

	// Get metrics
	metrics := circuit.GetMetrics()
	if metrics != nil {
		metricsStats := metrics.GetStats()
		fmt.Printf("Metrics: %+v\n", metricsStats)
	}

	// Get manager statistics
	managerStats := manager.GetManagerStats()
	fmt.Printf("Manager Stats: %+v\n", managerStats)

	// Example 10: HTTP Client Manager
	fmt.Println("\n=== Example 10: HTTP Client Manager ===")

	// Create HTTP client manager
	httpManager := circuitbreaker.NewHTTPClientManager(manager, logger)

	// Create HTTP clients
	userClient, _ := httpManager.CreateClient("user-api", config, httpConfig)
	orderClient, _ := httpManager.CreateClient("order-api", config, httpConfig)

	fmt.Printf("Created HTTP clients: %v\n", httpManager.GetClientNames())

	// Get HTTP client statistics
	httpStats := httpManager.GetManagerStats()
	fmt.Printf("HTTP Client Manager Stats: %+v\n", httpStats)

	// Example 11: Custom Error Handling
	fmt.Println("\n=== Example 11: Custom Error Handling ===")

	// Create circuit with custom error handling
	customConfig := circuitbreaker.DefaultConfig()
	customConfig.IsFailure = func(err error) bool {
		// Only treat specific errors as failures
		return err != nil && err.Error() == "service unavailable"
	}
	customConfig.IsSuccess = func(err error) bool {
		// Only treat nil errors as success
		return err == nil
	}

	customCircuit := circuitbreaker.NewCircuitBreaker("custom-service", customConfig, logger)

	// Test with different error types
	testErrors := []error{
		nil,                               // Success
		fmt.Errorf("service unavailable"), // Failure
		fmt.Errorf("network timeout"),     // Not a failure
		nil,                               // Success
		fmt.Errorf("service unavailable"), // Failure
	}

	for i, testErr := range testErrors {
		_, err := customCircuit.Execute(ctx, func() (interface{}, error) {
			return fmt.Sprintf("Test %d", i+1), testErr
		})
		fmt.Printf("Test %d: Error=%v, Circuit State=%s\n", i+1, err, customCircuit.GetState())
	}

	// Example 12: Reset and Cleanup
	fmt.Println("\n=== Example 12: Reset and Cleanup ===")

	// Reset circuit breaker
	circuit.Reset()
	fmt.Printf("Circuit state after reset: %s\n", circuit.GetState())

	// Reset all circuits in manager
	manager.ResetAll()
	fmt.Println("All circuits reset")

	// Stop manager
	manager.Stop()
	fmt.Println("Manager stopped")

	fmt.Println("\n🎉 All circuit breaker examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin circuit status' to view circuit status")
	fmt.Println("2. Use 'dolphin circuit create <name>' to create circuits")
	fmt.Println("3. Use 'dolphin circuit test <name>' to test circuits")
	fmt.Println("4. Use 'dolphin circuit metrics' to view metrics")
	fmt.Println("5. Integrate circuit breakers in your microservices")
	fmt.Println("6. Monitor circuit breaker states in production")
	fmt.Println("7. Set up alerts for open circuits")
	fmt.Println("8. Use HTTP client integration for external APIs")
}
