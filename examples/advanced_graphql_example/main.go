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
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Create GraphQL configuration
	config := graphql.DefaultSchemaConfig()
	config.Enabled = true
	config.EnableIntrospection = true
	config.EnablePlayground = true
	config.MaxQueryDepth = 10
	config.MaxQueryComplexity = 1000

	// Create schema manager
	schemaManager := graphql.NewSchemaManager(config, logger)

	// Build schema with advanced features
	if err := schemaManager.BuildSchema(); err != nil {
		log.Fatal("Failed to build schema:", err)
	}

	// Create handler
	handler := graphql.NewHandler(schemaManager, logger)

	// Example 1: Global Object Identification
	fmt.Println("=== Example 1: Global Object Identification ===")

	// Create a user node
	userNode := &graphql.UserNode{
		ID:        graphql.EncodeGlobalID("User", "123"),
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}

	fmt.Printf("User Node ID: %s\n", userNode.GetID())
	fmt.Printf("User Node Type: %s\n", userNode.GetType())

	// Decode the global ID
	nodeType, id, err := graphql.DecodeGlobalID(userNode.GetID())
	if err != nil {
		log.Printf("Failed to decode global ID: %v", err)
	} else {
		fmt.Printf("Decoded - Type: %s, ID: %s\n", nodeType, id)
	}

	// Example 2: Custom Directives
	fmt.Println("\n=== Example 2: Custom Directives ===")

	directiveRegistry := schemaManager.GetDirectiveRegistry()
	directives := directiveRegistry.ListDirectives()

	for _, directive := range directives {
		fmt.Printf("Available directive: %s - %s\n", directive.Name, directive.Description)
	}

	// Example 3: Query Analysis
	fmt.Println("\n=== Example 3: Query Analysis ===")

	queryValidator := schemaManager.GetQueryValidator()

	// Test query
	testQuery := `
		query {
			user(id: 1) {
				id
				name
				email
			}
			users(first: 10) {
				edges {
					node {
						id
						name
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	analysis, err := queryValidator.ValidateQuery(testQuery)
	if err != nil {
		log.Printf("Query validation failed: %v", err)
	} else {
		fmt.Printf("Query Analysis:\n")
		fmt.Printf("  Depth: %d\n", analysis.Depth)
		fmt.Printf("  Complexity: %d\n", analysis.Complexity)
		fmt.Printf("  Field Count: %d\n", analysis.FieldCount)
		fmt.Printf("  Valid: %t\n", analysis.Valid)
		if len(analysis.Errors) > 0 {
			fmt.Printf("  Errors: %v\n", analysis.Errors)
		}
	}

	// Example 4: Persisted Queries
	fmt.Println("\n=== Example 4: Persisted Queries ===")

	persistedQueryManager := schemaManager.GetPersistedQueryManager()

	// Persist a query
	persistedQuery, err := persistedQueryManager.PersistQuery(testQuery, "GetUserAndUsers", "Example query for user data")
	if err != nil {
		log.Printf("Failed to persist query: %v", err)
	} else {
		fmt.Printf("Persisted Query ID: %s\n", persistedQuery.ID)
		fmt.Printf("Description: %s\n", persistedQuery.Description)
		fmt.Printf("Created At: %s\n", persistedQuery.CreatedAt.Format(time.RFC3339))
	}

	// Load persisted query
	loadedQuery, err := persistedQueryManager.LoadQuery(persistedQuery.ID)
	if err != nil {
		log.Printf("Failed to load persisted query: %v", err)
	} else {
		fmt.Printf("Loaded Query: %s\n", loadedQuery.Query[:50]+"...")
		fmt.Printf("Use Count: %d\n", loadedQuery.UseCount)
	}

	// Example 5: Subscriptions
	fmt.Println("\n=== Example 5: Subscriptions ===")

	subscriptionManager := schemaManager.GetSubscriptionManager()

	// Get subscription stats
	stats := subscriptionManager.GetStats()
	fmt.Printf("Subscription Stats: %+v\n", stats)

	// Example 6: Execute GraphQL Query
	fmt.Println("\n=== Example 6: Execute GraphQL Query ===")

	ctx := context.Background()

	// Simple query
	simpleQuery := `
		query {
			user(id: 1) {
				id
				name
				email
			}
		}
	`

	result := schemaManager.Execute(ctx, simpleQuery, nil)
	if len(result.Errors) > 0 {
		fmt.Printf("Query errors: %v\n", result.Errors)
	} else {
		fmt.Printf("Query result: %+v\n", result.Data)
	}

	// Example 7: Node Query
	fmt.Println("\n=== Example 7: Node Query ===")

	nodeQuery := `
		query {
			node(id: "VXNlcjox") {
				... on User {
					id
					name
					email
				}
			}
		}
	`

	nodeResult := schemaManager.Execute(ctx, nodeQuery, nil)
	if len(nodeResult.Errors) > 0 {
		fmt.Printf("Node query errors: %v\n", nodeResult.Errors)
	} else {
		fmt.Printf("Node query result: %+v\n", nodeResult.Data)
	}

	// Example 8: Connection Query
	fmt.Println("\n=== Example 8: Connection Query ===")

	connectionQuery := `
		query {
			users(first: 2) {
				edges {
					node {
						id
						name
						email
					}
					cursor
				}
				pageInfo {
					hasNextPage
					hasPreviousPage
					startCursor
					endCursor
				}
			}
		}
	`

	connectionResult := schemaManager.Execute(ctx, connectionQuery, nil)
	if len(connectionResult.Errors) > 0 {
		fmt.Printf("Connection query errors: %v\n", connectionResult.Errors)
	} else {
		fmt.Printf("Connection query result: %+v\n", connectionResult.Data)
	}

	// Example 9: Start HTTP Server
	fmt.Println("\n=== Example 9: Start HTTP Server ===")

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/graphql", handler)
	mux.Handle("/graphql/playground", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.PlaygroundHandler(w, r)
	}))

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Starting GraphQL server on :8080")
	fmt.Println("GraphQL endpoint: http://localhost:8080/graphql")
	fmt.Println("GraphQL Playground: http://localhost:8080/graphql/playground")
	fmt.Println("Press Ctrl+C to stop")

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for a moment to show the server is running
	time.Sleep(2 * time.Second)

	fmt.Println("\n✅ Advanced GraphQL features implemented successfully!")
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("  • Global Object Identification (Node Interface)")
	fmt.Println("  • Relay-style Connections (Pagination)")
	fmt.Println("  • Custom Directives (Auth, Cache, Transform, etc.)")
	fmt.Println("  • Query Depth/Complexity Analysis")
	fmt.Println("  • Persisted Queries")
	fmt.Println("  • WebSocket Subscriptions")
	fmt.Println("  • Query Validation and Security")
	fmt.Println("  • Enterprise-grade Error Handling")
}
