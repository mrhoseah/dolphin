package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mrhoseah/dolphin/internal/http"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic HTTP Client
	fmt.Println("=== Example 1: Basic HTTP Client ===")

	// Create HTTP client configuration
	config := http.DefaultConfig()
	config.BaseURL = "https://jsonplaceholder.typicode.com"
	config.Timeout = 30 * time.Second
	config.MaxRetries = 3
	config.RetryDelay = 1 * time.Second
	config.EnableCircuitBreaker = true
	config.EnableRateLimit = true
	config.RateLimitRPS = 10
	config.EnableLogging = true
	config.EnableMetrics = true
	config.EnableCorrelationID = true

	// Create HTTP client
	client, err := http.NewClient(config, logger)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}
	defer client.Close()

	// Example 2: Basic GET Request
	fmt.Println("\n=== Example 2: Basic GET Request ===")

	response, err := client.Get("/posts/1")
	if err != nil {
		log.Printf("GET request failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
		fmt.Printf("Response Time: %v\n", response.Duration)
		fmt.Printf("Retry Count: %d\n", response.RetryCount)
	}

	// Example 3: POST Request with JSON Body
	fmt.Println("\n=== Example 3: POST Request with JSON Body ===")

	postData := map[string]interface{}{
		"title":  "Dolphin HTTP Client Test",
		"body":   "This is a test post created by Dolphin HTTP Client",
		"userId": 1,
	}

	response, err = client.Post("/posts", postData)
	if err != nil {
		log.Printf("POST request failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
		fmt.Printf("Response Time: %v\n", response.Duration)
	}

	// Example 4: Request with Headers
	fmt.Println("\n=== Example 4: Request with Headers ===")

	response, err = client.Get("/posts/1",
		http.WithHeader("Accept", "application/json"),
		http.WithHeader("X-Custom-Header", "Dolphin-Test"),
	)
	if err != nil {
		log.Printf("GET request with headers failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 5: Request with Query Parameters
	fmt.Println("\n=== Example 5: Request with Query Parameters ===")

	response, err = client.Get("/posts",
		http.WithQueryParam("userId", 1),
		http.WithQueryParam("_limit", 5),
	)
	if err != nil {
		log.Printf("GET request with query params failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 6: Request with Custom Timeout
	fmt.Println("\n=== Example 6: Request with Custom Timeout ===")

	response, err = client.Get("/posts/1",
		http.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Printf("GET request with timeout failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Response Time: %v\n", response.Duration)
	}

	// Example 7: Request with Context
	fmt.Println("\n=== Example 7: Request with Context ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err = client.Get("/posts/1",
		http.WithContext(ctx),
	)
	if err != nil {
		log.Printf("GET request with context failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 8: Request with Retries
	fmt.Println("\n=== Example 8: Request with Retries ===")

	response, err = client.Get("/posts/999", // This will likely return 404
		http.WithRetries(5),
		http.WithRetryDelay(2*time.Second),
	)
	if err != nil {
		log.Printf("GET request with retries failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Retry Count: %d\n", response.RetryCount)
	}

	// Example 9: Request with Custom Correlation ID
	fmt.Println("\n=== Example 9: Request with Custom Correlation ID ===")

	response, err = client.Get("/posts/1",
		http.WithCorrelationID("custom-correlation-id-123"),
	)
	if err != nil {
		log.Printf("GET request with custom correlation ID failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 10: Request with Authentication
	fmt.Println("\n=== Example 10: Request with Authentication ===")

	response, err = client.Get("/posts/1",
		http.WithBearerToken("your-token-here"),
	)
	if err != nil {
		log.Printf("GET request with auth failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 11: Request with JSON Body
	fmt.Println("\n=== Example 11: Request with JSON Body ===")

	response, err = client.Post("/posts", postData,
		http.WithJSON(postData),
	)
	if err != nil {
		log.Printf("POST request with JSON failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 12: Request with Form Data
	fmt.Println("\n=== Example 12: Request with Form Data ===")

	formData := map[string]interface{}{
		"title":  "Dolphin Form Test",
		"body":   "This is a form test",
		"userId": 1,
	}

	response, err = client.Post("/posts", formData,
		http.WithFormData(formData),
	)
	if err != nil {
		log.Printf("POST request with form data failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 13: Request with Custom Headers
	fmt.Println("\n=== Example 13: Request with Custom Headers ===")

	response, err = client.Get("/posts/1",
		http.WithHeaders(map[string]string{
			"Accept":        "application/json",
			"X-API-Version": "v1",
			"X-Client":      "Dolphin-HTTP-Client",
		}),
	)
	if err != nil {
		log.Printf("GET request with custom headers failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 14: Request with Multiple Query Parameters
	fmt.Println("\n=== Example 14: Request with Multiple Query Parameters ===")

	response, err = client.Get("/posts",
		http.WithQueryParams(map[string]interface{}{
			"userId": 1,
			"_limit": 10,
			"_sort":  "id",
			"_order": "desc",
		}),
	)
	if err != nil {
		log.Printf("GET request with multiple query params failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 15: Request with Circuit Breaker
	fmt.Println("\n=== Example 15: Request with Circuit Breaker ===")

	response, err = client.Get("/posts/1",
		http.WithCircuitBreaker(true),
	)
	if err != nil {
		log.Printf("GET request with circuit breaker failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 16: Request with Rate Limiting
	fmt.Println("\n=== Example 16: Request with Rate Limiting ===")

	response, err = client.Get("/posts/1",
		http.WithRateLimit(true),
	)
	if err != nil {
		log.Printf("GET request with rate limiting failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 17: Request with Metrics
	fmt.Println("\n=== Example 17: Request with Metrics ===")

	response, err = client.Get("/posts/1",
		http.WithMetrics(true),
	)
	if err != nil {
		log.Printf("GET request with metrics failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 18: Request with Logging
	fmt.Println("\n=== Example 18: Request with Logging ===")

	response, err = client.Get("/posts/1",
		http.WithLogging(true),
		http.WithVerboseLogging(true),
	)
	if err != nil {
		log.Printf("GET request with logging failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 19: Request with Custom Options
	fmt.Println("\n=== Example 19: Request with Custom Options ===")

	response, err = client.Get("/posts/1",
		http.WithCustomOption("debug", "true"),
		http.WithCustomOption("trace", "enabled"),
	)
	if err != nil {
		log.Printf("GET request with custom options failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
	}

	// Example 20: Request with All Options
	fmt.Println("\n=== Example 20: Request with All Options ===")

	response, err = client.Get("/posts/1",
		http.WithContext(ctx),
		http.WithTimeout(5*time.Second),
		http.WithRetries(3),
		http.WithCorrelationID("all-options-test"),
		http.WithHeaders(map[string]string{
			"Accept": "application/json",
			"X-Test": "Dolphin-HTTP-Client",
		}),
		http.WithQueryParams(map[string]interface{}{
			"include": "comments",
		}),
		http.WithBearerToken("test-token"),
		http.WithCircuitBreaker(true),
		http.WithRateLimit(true),
		http.WithMetrics(true),
		http.WithLogging(true),
	)
	if err != nil {
		log.Printf("GET request with all options failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", response.StatusCode)
		fmt.Printf("Correlation ID: %s\n", response.CorrelationID)
		fmt.Printf("Response Time: %v\n", response.Duration)
		fmt.Printf("Retry Count: %d\n", response.RetryCount)
	}

	// Example 21: Get Client Metrics
	fmt.Println("\n=== Example 21: Get Client Metrics ===")

	metrics := client.GetMetrics()
	if metrics != nil {
		stats := metrics.GetStats()
		fmt.Printf("Total Requests: %v\n", stats["total_requests"])
		fmt.Printf("Successful Requests: %v\n", stats["successful_requests"])
		fmt.Printf("Failed Requests: %v\n", stats["failed_requests"])
		fmt.Printf("Success Rate: %.2f%%\n", stats["success_rate"])
		fmt.Printf("Average Response Time: %v\n", stats["avg_response_time"])
	}

	// Example 22: Get Circuit Breaker Status
	fmt.Println("\n=== Example 22: Get Circuit Breaker Status ===")

	circuitBreaker := client.GetCircuitBreaker()
	if circuitBreaker != nil {
		stats := circuitBreaker.GetStats()
		fmt.Printf("Circuit Breaker State: %v\n", stats["state"])
		fmt.Printf("Failure Count: %v\n", stats["failure_count"])
		fmt.Printf("Success Count: %v\n", stats["success_count"])
		fmt.Printf("Is Open: %v\n", stats["is_open"])
	}

	// Example 23: Get Rate Limiter Status
	fmt.Println("\n=== Example 23: Get Rate Limiter Status ===")

	rateLimiter := client.GetRateLimiter()
	if rateLimiter != nil {
		stats := rateLimiter.GetStats()
		fmt.Printf("RPS: %v\n", stats["rps"])
		fmt.Printf("Burst: %v\n", stats["burst"])
		fmt.Printf("Tokens: %v\n", stats["tokens"])
		fmt.Printf("Is Limited: %v\n", stats["is_limited"])
	}

	// Example 24: Error Handling
	fmt.Println("\n=== Example 24: Error Handling ===")

	// Test with invalid URL
	response, err = client.Get("https://invalid-url-that-does-not-exist.com/api/test")
	if err != nil {
		fmt.Printf("Expected error for invalid URL: %v\n", err)
	} else {
		fmt.Printf("Unexpected success: %d\n", response.StatusCode)
	}

	// Test with timeout
	response, err = client.Get("/posts/1",
		http.WithTimeout(1*time.Millisecond), // Very short timeout
	)
	if err != nil {
		fmt.Printf("Expected timeout error: %v\n", err)
	} else {
		fmt.Printf("Unexpected success: %d\n", response.StatusCode)
	}

	// Example 25: Configuration
	fmt.Println("\n=== Example 25: Configuration ===")

	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Timeout: %v\n", config.Timeout)
	fmt.Printf("Max Retries: %d\n", config.MaxRetries)
	fmt.Printf("Retry Delay: %v\n", config.RetryDelay)
	fmt.Printf("Circuit Breaker: %v\n", config.EnableCircuitBreaker)
	fmt.Printf("Rate Limiting: %v\n", config.EnableRateLimit)
	fmt.Printf("Rate Limit RPS: %d\n", config.RateLimitRPS)
	fmt.Printf("Logging: %v\n", config.EnableLogging)
	fmt.Printf("Metrics: %v\n", config.EnableMetrics)
	fmt.Printf("Correlation ID: %v\n", config.EnableCorrelationID)

	fmt.Println("\n🎉 All HTTP client examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin http test' to test the client")
	fmt.Println("2. Use 'dolphin http stats' to view statistics")
	fmt.Println("3. Use 'dolphin http config' to view configuration")
	fmt.Println("4. Use 'dolphin http health' to check health status")
	fmt.Println("5. Use 'dolphin http reset' to reset metrics")
	fmt.Println("6. Configure the client for your specific needs")
	fmt.Println("7. Implement error handling and retry logic")
	fmt.Println("8. Monitor performance with metrics and logging")
	fmt.Println("9. Use circuit breakers for fault tolerance")
	fmt.Println("10. Implement rate limiting for API protection")
}
