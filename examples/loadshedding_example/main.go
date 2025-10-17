package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrhoseah/dolphin/internal/loadshedding"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic Load Shedding
	fmt.Println("=== Example 1: Basic Load Shedding ===")

	// Create load shedding configuration
	config := loadshedding.DefaultConfig()
	config.Strategy = loadshedding.StrategyCombined
	config.LightThreshold = 0.6
	config.ModerateThreshold = 0.75
	config.HeavyThreshold = 0.85
	config.CriticalThreshold = 0.95
	config.EnableLogging = true

	// Create load shedder
	shedder := loadshedding.NewLoadShedder(config, logger)

	// Example 2: Test Load Shedding
	fmt.Println("\n=== Example 2: Test Load Shedding ===")

	// Simulate different load conditions
	loadConditions := []struct {
		name        string
		cpuUsage    float64
		memoryUsage float64
		goroutines  int
		requestRate float64
	}{
		{"Normal", 0.4, 0.3, 50, 100},
		{"Light", 0.65, 0.5, 100, 200},
		{"Moderate", 0.8, 0.7, 200, 400},
		{"Heavy", 0.9, 0.85, 500, 800},
		{"Critical", 0.98, 0.95, 1000, 1500},
	}

	for _, condition := range loadConditions {
		fmt.Printf("Testing %s load: CPU=%.1f%%, Memory=%.1f%%, Goroutines=%d, Rate=%.0f req/s\n",
			condition.name, condition.cpuUsage*100, condition.memoryUsage*100, condition.goroutines, condition.requestRate)

		// Simulate requests
		for i := 0; i < 10; i++ {
			shouldShed := shedder.ShouldShed(context.Background())
			if shouldShed {
				fmt.Printf("  Request %d: SHED\n", i+1)
			} else {
				fmt.Printf("  Request %d: PROCESSED\n", i+1)
			}
		}

		// Show current stats
		stats := shedder.GetStats()
		fmt.Printf("  Current Level: %s, Shed Rate: %.1f%%\n\n",
			stats.CurrentLevel.String(), stats.CurrentShedRate*100)
	}

	// Example 3: HTTP Middleware Integration
	fmt.Println("=== Example 3: HTTP Middleware Integration ===")

	// Create middleware configuration
	middlewareConfig := loadshedding.DefaultMiddlewareConfig()
	middlewareConfig.ErrorResponse = []byte(`{"error":"Service temporarily unavailable","code":"LOAD_SHEDDING"}`)
	middlewareConfig.ErrorStatusCode = http.StatusServiceUnavailable

	// Create middleware
	middleware := loadshedding.NewMiddleware(shedder, middlewareConfig, logger)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate work
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Request processed","timestamp":"%s"}`,
			time.Now().Format(time.RFC3339))
	})

	// Create HTTP server with load shedding middleware
	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware.Handler(handler),
	}

	fmt.Println("Starting HTTP server with load shedding on :8080")

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Example 4: Load Shedding Manager
	fmt.Println("\n=== Example 4: Load Shedding Manager ===")

	// Create manager
	manager := loadshedding.NewLoadSheddingManager(logger)

	// Create multiple shedders
	apiShedder, _ := manager.CreateShedder("api-shedder", config)
	dbShedder, _ := manager.CreateShedder("db-shedder", config)
	cacheShedder, _ := manager.CreateShedder("cache-shedder", config)

	fmt.Printf("Created shedders: %v\n", manager.GetShedderNames())

	// Create middlewares for each shedder
	apiMiddleware, _ := manager.CreateMiddleware("api-middleware", "api-shedder", middlewareConfig)
	dbMiddleware, _ := manager.CreateMiddleware("db-middleware", "db-shedder", middlewareConfig)
	cacheMiddleware, _ := manager.CreateMiddleware("cache-middleware", "cache-shedder", middlewareConfig)

	fmt.Printf("Created middlewares: %v\n", manager.GetMiddlewareNames())

	// Example 5: Force Shedding Levels
	fmt.Println("\n=== Example 5: Force Shedding Levels ===")

	// Force different shedding levels
	levels := []loadshedding.SheddingLevel{
		loadshedding.LevelNone,
		loadshedding.LevelLight,
		loadshedding.LevelModerate,
		loadshedding.LevelHeavy,
		loadshedding.LevelCritical,
	}

	for _, level := range levels {
		apiShedder.ForceLevel(level)
		stats := apiShedder.GetStats()
		fmt.Printf("Forced level %s: Shed Rate = %.1f%%\n",
			level.String(), stats.CurrentShedRate*100)
	}

	// Example 6: Metrics and Monitoring
	fmt.Println("\n=== Example 6: Metrics and Monitoring ===")

	// Get individual shedder stats
	apiStats := apiShedder.GetStats()
	fmt.Printf("API Shedder Stats: Level=%s, ShedRate=%.1f%%, CPU=%.1f%%\n",
		apiStats.CurrentLevel.String(), apiStats.CurrentShedRate*100, apiStats.CPUUsage*100)

	// Get manager stats
	managerStats := manager.GetManagerStats()
	fmt.Printf("Manager Stats: Shedders=%d, Middlewares=%d\n",
		managerStats.ShedderCount, managerStats.MiddlewareCount)

	// Get aggregated stats
	aggregatedStats := manager.GetAggregatedStats()
	fmt.Printf("Aggregated Stats: TotalRequests=%d, AvgShedRate=%.1f%%\n",
		aggregatedStats.TotalRequests, aggregatedStats.AvgShedRate*100)

	// Example 7: Adaptive Adjustment
	fmt.Println("\n=== Example 7: Adaptive Adjustment ===")

	// Simulate adaptive adjustment over time
	for i := 0; i < 5; i++ {
		// Simulate some requests
		for j := 0; j < 10; j++ {
			apiShedder.ShouldShed(context.Background())
		}

		// Get current stats
		stats := apiShedder.GetStats()
		fmt.Printf("Iteration %d: Level=%s, ShedRate=%.1f%%, Adjustments=%d\n",
			i+1, stats.CurrentLevel.String(), stats.CurrentShedRate*100, stats.AdjustmentCount)

		time.Sleep(100 * time.Millisecond)
	}

	// Example 8: Different Strategies
	fmt.Println("\n=== Example 8: Different Strategies ===")

	// Create shedders with different strategies
	cpuConfig := loadshedding.DefaultConfig()
	cpuConfig.Strategy = loadshedding.StrategyCPU
	cpuShedder, _ := manager.CreateShedder("cpu-shedder", cpuConfig)

	memoryConfig := loadshedding.DefaultConfig()
	memoryConfig.Strategy = loadshedding.StrategyMemory
	memoryShedder, _ := manager.CreateShedder("memory-shedder", memoryConfig)

	goroutineConfig := loadshedding.DefaultConfig()
	goroutineConfig.Strategy = loadshedding.StrategyGoroutines
	goroutineShedder, _ := manager.CreateShedder("goroutine-shedder", goroutineConfig)

	fmt.Printf("Created strategy-specific shedders: %v\n", manager.GetShedderNames())

	// Example 9: Reset and Cleanup
	fmt.Println("\n=== Example 9: Reset and Cleanup ===")

	// Reset individual shedder
	apiShedder.Reset()
	fmt.Println("API shedder reset")

	// Reset all shedders
	manager.ResetAll()
	fmt.Println("All shedders reset")

	// Stop manager
	manager.Stop()
	fmt.Println("Manager stopped")

	// Stop main shedder
	shedder.Stop()
	fmt.Println("Main shedder stopped")

	fmt.Println("\n🎉 All load shedding examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin loadshed status' to view shedding status")
	fmt.Println("2. Use 'dolphin loadshed create <name>' to create shedders")
	fmt.Println("3. Use 'dolphin loadshed test <name>' to test shedders")
	fmt.Println("4. Use 'dolphin loadshed metrics' to view metrics")
	fmt.Println("5. Integrate load shedding in your HTTP middleware")
	fmt.Println("6. Monitor shedding levels in production")
	fmt.Println("7. Set up alerts for high shedding levels")
	fmt.Println("8. Use different strategies for different services")
}
