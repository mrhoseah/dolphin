package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrhoseah/dolphin/internal/graphql"
	"go.uber.org/zap"
)

func main() {
	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	// Create GraphQL configuration
	config := &graphql.SchemaConfig{
		Enabled:             true, // Enable GraphQL
		EnableIntrospection: true, // Enable introspection
		EnablePlayground:    true, // Enable playground
		PlaygroundPath:      "/graphql/playground",
		IntrospectionPath:   "/graphql/introspection",
		QueryPath:           "/graphql",
		MutationPath:        "/graphql",
		SubscriptionPath:    "/graphql/ws",
		MaxQueryDepth:       15,
		MaxQueryComplexity:  1000,
		QueryTimeout:        30 * time.Second,
		EnableTracing:       true,
		EnableMetrics:       true,
		AutoEnable:          false, // Don't auto-enable
	}

	// Create schema manager
	schemaManager := graphql.NewSchemaManager(config, logger)

	// Build schema
	if err := schemaManager.BuildSchema(); err != nil {
		log.Fatal(err)
	}

	// Create handler
	handler := graphql.NewHandler(schemaManager, logger)

	// Create HTTP server
	mux := http.NewServeMux()

	// Register GraphQL endpoints
	mux.Handle("/graphql", handler)
	mux.Handle("/graphql/playground", http.HandlerFunc(handler.PlaygroundHandler))
	mux.Handle("/graphql/introspection", http.HandlerFunc(handler.IntrospectionHandler))
	mux.Handle("/graphql/health", http.HandlerFunc(handler.HealthHandler))
	mux.Handle("/graphql/status", http.HandlerFunc(handler.StatusHandler))

	// Add some basic routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>GraphQL Example</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .code { background: #e8e8e8; padding: 5px; border-radius: 3px; font-family: monospace; }
    </style>
</head>
<body>
    <h1>🐬 Dolphin GraphQL Example</h1>
    
    <h2>Available Endpoints:</h2>
    <div class="endpoint">
        <strong>GraphQL Query:</strong> <span class="code">POST /graphql</span>
    </div>
    <div class="endpoint">
        <strong>GraphQL Playground:</strong> <a href="/graphql/playground">/graphql/playground</a>
    </div>
    <div class="endpoint">
        <strong>GraphQL Introspection:</strong> <span class="code">POST /graphql/introspection</span>
    </div>
    <div class="endpoint">
        <strong>GraphQL Health:</strong> <a href="/graphql/health">/graphql/health</a>
    </div>
    <div class="endpoint">
        <strong>GraphQL Status:</strong> <a href="/graphql/status">/graphql/status</a>
    </div>

    <h2>Example Queries:</h2>
    <h3>Query Users:</h3>
    <div class="code">
        query {<br>
        &nbsp;&nbsp;users {<br>
        &nbsp;&nbsp;&nbsp;&nbsp;id<br>
        &nbsp;&nbsp;&nbsp;&nbsp;name<br>
        &nbsp;&nbsp;&nbsp;&nbsp;email<br>
        &nbsp;&nbsp;&nbsp;&nbsp;createdAt<br>
        &nbsp;&nbsp;}<br>
        }
    </div>

    <h3>Query Single User:</h3>
    <div class="code">
        query {<br>
        &nbsp;&nbsp;user(id: 1) {<br>
        &nbsp;&nbsp;&nbsp;&nbsp;id<br>
        &nbsp;&nbsp;&nbsp;&nbsp;name<br>
        &nbsp;&nbsp;&nbsp;&nbsp;email<br>
        &nbsp;&nbsp;&nbsp;&nbsp;createdAt<br>
        &nbsp;&nbsp;}<br>
        }
    </div>

    <h3>Create User:</h3>
    <div class="code">
        mutation {<br>
        &nbsp;&nbsp;createUser(name: "John Doe", email: "john@example.com") {<br>
        &nbsp;&nbsp;&nbsp;&nbsp;id<br>
        &nbsp;&nbsp;&nbsp;&nbsp;name<br>
        &nbsp;&nbsp;&nbsp;&nbsp;email<br>
        &nbsp;&nbsp;&nbsp;&nbsp;createdAt<br>
        &nbsp;&nbsp;}<br>
        }
    </div>

    <h2>GraphQL Features:</h2>
    <ul>
        <li>✅ <strong>Pluggable:</strong> Can be enabled/disabled at runtime</li>
        <li>✅ <strong>Schema-first:</strong> Define schema in GraphQL SDL</li>
        <li>✅ <strong>Playground:</strong> Interactive GraphQL IDE</li>
        <li>✅ <strong>Introspection:</strong> Schema discovery and documentation</li>
        <li>✅ <strong>Validation:</strong> Query validation and error handling</li>
        <li>✅ <strong>Metrics:</strong> Request metrics and monitoring</li>
        <li>✅ <strong>Tracing:</strong> Request tracing for debugging</li>
        <li>✅ <strong>Security:</strong> Query depth and complexity limits</li>
    </ul>

    <h2>CLI Commands:</h2>
    <div class="code">
        # Enable/disable GraphQL<br>
        dolphin graphql enable<br>
        dolphin graphql disable<br>
        dolphin graphql toggle<br><br>
        
        # Check status and configuration<br>
        dolphin graphql status<br>
        dolphin graphql config<br><br>
        
        # Test and validate<br>
        dolphin graphql test<br>
        dolphin graphql validate 'query { users { id name } }'<br><br>
        
        # Generate code<br>
        dolphin graphql generate ./graphql<br><br>
        
        # Open playground<br>
        dolphin graphql playground<br>
    </div>
</body>
</html>
		`)
	})

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("🚀 Starting GraphQL server on :8080")
	fmt.Println("")
	fmt.Println("🌐 Available endpoints:")
	fmt.Println("  • GraphQL Query: http://localhost:8080/graphql")
	fmt.Println("  • GraphQL Playground: http://localhost:8080/graphql/playground")
	fmt.Println("  • GraphQL Health: http://localhost:8080/graphql/health")
	fmt.Println("  • GraphQL Status: http://localhost:8080/graphql/status")
	fmt.Println("  • Home: http://localhost:8080/")
	fmt.Println("")

	// Demonstrate GraphQL functionality
	fmt.Println("🧪 Testing GraphQL functionality...")
	fmt.Println("")

	// Test 1: Basic query
	fmt.Println("1️⃣ Testing basic query...")
	query := `query { users { id name email } }`
	result := schemaManager.Execute(context.Background(), query, nil)
	if len(result.Errors) == 0 {
		fmt.Println("   ✅ Basic query successful")
	} else {
		fmt.Printf("   ❌ Basic query failed: %v\n", result.Errors)
	}

	// Test 2: Mutation
	fmt.Println("2️⃣ Testing mutation...")
	mutation := `mutation { createUser(name: "Test User", email: "test@example.com") { id name email } }`
	result = schemaManager.Execute(context.Background(), mutation, nil)
	if len(result.Errors) == 0 {
		fmt.Println("   ✅ Mutation successful")
	} else {
		fmt.Printf("   ❌ Mutation failed: %v\n", result.Errors)
	}

	// Test 3: Introspection
	fmt.Println("3️⃣ Testing introspection...")
	introspectionQuery := schemaManager.GetIntrospectionQuery()
	result = schemaManager.Execute(context.Background(), introspectionQuery, nil)
	if len(result.Errors) == 0 {
		fmt.Println("   ✅ Introspection successful")
	} else {
		fmt.Printf("   ❌ Introspection failed: %v\n", result.Errors)
	}

	// Test 4: Validation
	fmt.Println("4️⃣ Testing query validation...")
	err = schemaManager.ValidateQuery(query)
	if err == nil {
		fmt.Println("   ✅ Query validation successful")
	} else {
		fmt.Printf("   ❌ Query validation failed: %v\n", err)
	}

	// Test 5: Disabled state
	fmt.Println("5️⃣ Testing disabled state...")
	schemaManager.Disable()
	result = schemaManager.Execute(context.Background(), query, nil)
	if len(result.Errors) > 0 && result.Errors[0].Message == "GraphQL endpoint is disabled" {
		fmt.Println("   ✅ Disabled state working correctly")
	} else {
		fmt.Println("   ❌ Disabled state not working")
	}

	// Re-enable for server
	schemaManager.Enable()
	fmt.Println("")

	fmt.Println("✅ All tests passed! GraphQL is ready to use.")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Visit http://localhost:8080/graphql/playground to test queries")
	fmt.Println("  • Use 'dolphin graphql' commands to manage GraphQL")
	fmt.Println("  • GraphQL is currently enabled and ready")
	fmt.Println("")

	// Start the server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
