package graphql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"go.uber.org/zap"
)

// SchemaManager manages GraphQL schemas and provides schema-first development
type SchemaManager struct {
	schema                *graphql.Schema
	resolvers             map[string]Resolver
	logger                *zap.Logger
	config                *SchemaConfig
	nodeRegistry          *NodeRegistry
	directiveRegistry     *DirectiveRegistry
	subscriptionManager   *SubscriptionManager
	queryValidator        *QueryValidator
	persistedQueryManager *PersistedQueryManager
}

// SchemaConfig holds configuration for the GraphQL schema
type SchemaConfig struct {
	Enabled             bool // Master enable/disable switch
	EnableIntrospection bool
	EnablePlayground    bool
	PlaygroundPath      string
	IntrospectionPath   string
	QueryPath           string
	MutationPath        string
	SubscriptionPath    string
	MaxQueryDepth       int
	MaxQueryComplexity  int
	QueryTimeout        time.Duration
	EnableTracing       bool
	EnableMetrics       bool
	AutoEnable          bool // Auto-enable when schema is built
}

// DefaultSchemaConfig returns default configuration
func DefaultSchemaConfig() *SchemaConfig {
	return &SchemaConfig{
		Enabled:             false, // Disabled by default for pluggability
		EnableIntrospection: true,
		EnablePlayground:    true,
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
}

// Resolver defines a GraphQL resolver function
type Resolver func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// FieldResolver defines a field resolver
type FieldResolver struct {
	Type    *graphql.Object
	Resolve Resolver
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(config *SchemaConfig, logger *zap.Logger) *SchemaManager {
	if config == nil {
		config = DefaultSchemaConfig()
	}

	// Initialize advanced features
	nodeRegistry := NewNodeRegistry(logger)
	directiveRegistry := NewDirectiveRegistry(logger)
	directiveRegistry.InitializeDefaultDirectives()

	subscriptionManager := NewSubscriptionManager(logger)
	queryValidator := NewQueryValidator(config.MaxQueryDepth, config.MaxQueryComplexity, logger)
	persistedQueryManager := NewPersistedQueryManager(logger)

	return &SchemaManager{
		resolvers:             make(map[string]Resolver),
		logger:                logger,
		config:                config,
		nodeRegistry:          nodeRegistry,
		directiveRegistry:     directiveRegistry,
		subscriptionManager:   subscriptionManager,
		queryValidator:        queryValidator,
		persistedQueryManager: persistedQueryManager,
	}
}

// AddType adds a GraphQL type to the schema
func (sm *SchemaManager) AddType(name string, objectConfig graphql.ObjectConfig) {
	// This would be implemented to register types
	sm.logger.Info("Adding GraphQL type", zap.String("name", name))
}

// AddResolver adds a resolver function
func (sm *SchemaManager) AddResolver(name string, resolver Resolver) {
	sm.resolvers[name] = resolver
	sm.logger.Info("Added GraphQL resolver", zap.String("name", name))
}

// AddQuery adds a query field
func (sm *SchemaManager) AddQuery(name string, fieldConfig graphql.Field) {
	// This would be implemented to add query fields
	sm.logger.Info("Adding GraphQL query", zap.String("name", name))
}

// AddMutation adds a mutation field
func (sm *SchemaManager) AddMutation(name string, fieldConfig graphql.Field) {
	// This would be implemented to add mutation fields
	sm.logger.Info("Adding GraphQL mutation", zap.String("name", name))
}

// AddSubscription adds a subscription field
func (sm *SchemaManager) AddSubscription(name string, fieldConfig graphql.Field) {
	// This would be implemented to add subscription fields
	sm.logger.Info("Adding GraphQL subscription", zap.String("name", name))
}

// Enable enables the GraphQL endpoint
func (sm *SchemaManager) Enable() {
	sm.config.Enabled = true
	sm.logger.Info("GraphQL endpoint enabled")
}

// Disable disables the GraphQL endpoint
func (sm *SchemaManager) Disable() {
	sm.config.Enabled = false
	sm.logger.Info("GraphQL endpoint disabled")
}

// IsEnabled returns whether GraphQL is enabled
func (sm *SchemaManager) IsEnabled() bool {
	return sm.config.Enabled
}

// Toggle toggles the GraphQL endpoint state
func (sm *SchemaManager) Toggle() {
	sm.config.Enabled = !sm.config.Enabled
	state := "disabled"
	if sm.config.Enabled {
		state = "enabled"
	}
	sm.logger.Info("GraphQL endpoint toggled", zap.String("state", state))
}

// BuildSchema builds the final GraphQL schema
func (sm *SchemaManager) BuildSchema() error {
	// Create Node interface
	nodeInterface := CreateNodeInterface()

	// Create a User type that implements Node interface
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if user, ok := p.Source.(Node); ok {
						return user.GetID(), nil
					}
					return nil, fmt.Errorf("source does not implement Node interface")
				},
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
		Interfaces: []*graphql.Interface{nodeInterface},
	})

	// Register User node type
	sm.nodeRegistry.RegisterNodeType("User", func(ctx context.Context, id string) (Node, error) {
		// This would normally fetch from database
		return &UserNode{
			ID:        EncodeGlobalID("User", id),
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		}, nil
	})

	// Create query type with advanced features
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// Node query for Global Object Identification
			"node": CreateNodeQuery(sm.nodeRegistry),

			// User query with directive support
			"user": sm.directiveRegistry.CreateDirectiveField(
				"user",
				userType,
				[]string{"auth", "cache"},
				func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if !ok {
						return nil, fmt.Errorf("id is required")
					}
					return &UserNode{
						ID:        EncodeGlobalID("User", fmt.Sprintf("%d", id)),
						Name:      "John Doe",
						Email:     "john@example.com",
						CreatedAt: time.Now(),
					}, nil
				},
			),

			// Users connection for pagination
			"users": CreateConnectionField(
				"users",
				userType,
				func(ctx context.Context, args ConnectionArgs) (*Connection, error) {
					// Mock data
					users := []interface{}{
						&UserNode{
							ID:        EncodeGlobalID("User", "1"),
							Name:      "John Doe",
							Email:     "john@example.com",
							CreatedAt: time.Now(),
						},
						&UserNode{
							ID:        EncodeGlobalID("User", "2"),
							Name:      "Jane Smith",
							Email:     "jane@example.com",
							CreatedAt: time.Now(),
						},
					}

					// Apply pagination
					helper := NewPaginationHelper(sm.logger)
					paginatedUsers, pageInfo := helper.ApplyPagination(users, args)
					edges := helper.CreateEdges(paginatedUsers)

					return &Connection{
						Edges:    edges,
						PageInfo: pageInfo,
					}, nil
				},
			),
		},
	})

	// Create mutation type
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createUser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name, _ := p.Args["name"].(string)
					email, _ := p.Args["email"].(string)
					return map[string]interface{}{
						"id":        3,
						"name":      name,
						"email":     email,
						"createdAt": time.Now(),
					}, nil
				},
			},
		},
	})

	// Create subscription type
	subscriptionType := CreateSubscriptionType(sm.subscriptionManager)

	// Create schema with all types
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:        queryType,
		Mutation:     mutationType,
		Subscription: subscriptionType,
		Types:        []graphql.Type{nodeInterface},
	})
	if err != nil {
		return fmt.Errorf("failed to create GraphQL schema: %w", err)
	}

	sm.schema = &schema

	// Auto-enable if configured
	if sm.config.AutoEnable {
		sm.Enable()
	}

	sm.logger.Info("GraphQL schema built successfully", zap.Bool("enabled", sm.config.Enabled))
	return nil
}

// Execute executes a GraphQL query
func (sm *SchemaManager) Execute(ctx context.Context, query string, variables map[string]interface{}) *graphql.Result {
	// Check if GraphQL is enabled
	if !sm.config.Enabled {
		return &graphql.Result{
			Errors: []gqlerrors.FormattedError{
				{Message: "GraphQL endpoint is disabled"},
			},
		}
	}

	if sm.schema == nil {
		return &graphql.Result{
			Errors: []gqlerrors.FormattedError{
				{Message: "Schema not built"},
			},
		}
	}

	// Validate query depth and complexity
	if sm.queryValidator != nil {
		analysis, err := sm.queryValidator.ValidateQuery(query)
		if err != nil {
			return &graphql.Result{
				Errors: []gqlerrors.FormattedError{
					{Message: fmt.Sprintf("Query validation failed: %v", err)},
				},
			}
		}

		if !analysis.Valid {
			return &graphql.Result{
				Errors: []gqlerrors.FormattedError{
					{Message: fmt.Sprintf("Query validation failed: %v", analysis.Errors)},
				},
			}
		}
	}

	// Add timeout context
	if sm.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sm.config.QueryTimeout)
		defer cancel()
	}

	result := graphql.Do(graphql.Params{
		Schema:         *sm.schema,
		RequestString:  query,
		VariableValues: variables,
		Context:        ctx,
	})

	return result
}

// GetSchema returns the current schema
func (sm *SchemaManager) GetSchema() *graphql.Schema {
	return sm.schema
}

// GetNodeRegistry returns the node registry
func (sm *SchemaManager) GetNodeRegistry() *NodeRegistry {
	return sm.nodeRegistry
}

// GetDirectiveRegistry returns the directive registry
func (sm *SchemaManager) GetDirectiveRegistry() *DirectiveRegistry {
	return sm.directiveRegistry
}

// GetSubscriptionManager returns the subscription manager
func (sm *SchemaManager) GetSubscriptionManager() *SubscriptionManager {
	return sm.subscriptionManager
}

// GetQueryValidator returns the query validator
func (sm *SchemaManager) GetQueryValidator() *QueryValidator {
	return sm.queryValidator
}

// GetPersistedQueryManager returns the persisted query manager
func (sm *SchemaManager) GetPersistedQueryManager() *PersistedQueryManager {
	return sm.persistedQueryManager
}

// UserNode represents a user node that implements the Node interface
type UserNode struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}

// GetID returns the global ID
func (u *UserNode) GetID() string {
	return u.ID
}

// GetType returns the node type
func (u *UserNode) GetType() string {
	return "User"
}

// GetConfig returns the schema configuration
func (sm *SchemaManager) GetConfig() *SchemaConfig {
	return sm.config
}

// LoadFromSDL loads schema from GraphQL SDL
func (sm *SchemaManager) LoadFromSDL(sdl string) error {
	// This would parse SDL and build schema
	sm.logger.Info("Loading schema from SDL", zap.String("sdl", sdl[:min(100, len(sdl))]))
	return nil
}

// GenerateTypes generates Go types from schema
func (sm *SchemaManager) GenerateTypes(outputDir string) error {
	sm.logger.Info("Generating types from schema", zap.String("output", outputDir))
	// This would generate Go types
	return nil
}

// ValidateQuery validates a GraphQL query
func (sm *SchemaManager) ValidateQuery(query string) error {
	if sm.schema == nil {
		return fmt.Errorf("schema not built")
	}

	// Parse query
	ast, err := parser.Parse(parser.ParseParams{
		Source: &source.Source{Body: []byte(query)},
	})
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	// Validate query
	validationResult := graphql.ValidateDocument(sm.schema, ast, nil)
	if !validationResult.IsValid {
		var errors []string
		for _, err := range validationResult.Errors {
			errors = append(errors, err.Message)
		}
		return fmt.Errorf("query validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// GetIntrospectionQuery returns the introspection query
func (sm *SchemaManager) GetIntrospectionQuery() string {
	return `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
				directives {
					name
					description
					locations
					args {
						...InputValue
					}
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes {
				...TypeRef
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`
}

// GetPlaygroundHTML returns the GraphiQL playground HTML
func (sm *SchemaManager) GetPlaygroundHTML() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<title>GraphQL Playground</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
	<link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
	<script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
	<div id="root">
		<style>
			body {
				background-color: rgb(23, 42, 58);
				font-family: Open Sans, sans-serif;
				height: 90vh;
				margin: 0;
				overflow: hidden;
			}
			#root {
				height: 100vh;
				width: 100vw;
			}
		</style>
		<script>
			window.addEventListener('load', function (event) {
				const root = document.getElementById('root');
				root.innerHTML = GraphQLPlayground.init({
					endpoint: '%s',
					subscriptionEndpoint: '%s',
					settings: {
						'request.credentials': 'include',
						'editor.theme': 'dark',
						'editor.fontSize': 14,
						'editor.fontFamily': "'Source Code Pro', 'Consolas', 'Inconsolata', 'Droid Sans Mono', 'Monaco', monospace",
						'editor.reuseHeaders': true,
						'tracing.hideTracingResponse': true,
						'queryPlan.hideQueryPlanResponse': true,
						'editor.cursorShape': 'line',
						'editor.autoComplete': true,
						'editor.tabSize': 2,
					},
				});
			});
		</script>
	</div>
</body>
</html>
	`, sm.config.QueryPath, sm.config.SubscriptionPath)
}

// GetSchemaSDL returns the schema in SDL format
func (sm *SchemaManager) GetSchemaSDL() (string, error) {
	if sm.schema == nil {
		return "", fmt.Errorf("schema not built")
	}

	// This would convert the schema to SDL format
	// For now, return a basic SDL
	return `
		type User {
			id: Int!
			name: String!
			email: String!
			createdAt: DateTime!
		}

		type Query {
			user(id: Int!): User
			users: [User!]!
		}

		type Mutation {
			createUser(name: String!, email: String!): User!
		}
	`, nil
}

// AddMiddleware adds middleware to the schema
func (sm *SchemaManager) AddMiddleware(middleware func(next graphql.FieldResolveFn) graphql.FieldResolveFn) {
	// This would add middleware to field resolvers
	sm.logger.Info("Adding GraphQL middleware")
}

// GetMetrics returns schema metrics
func (sm *SchemaManager) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"types_count":     len(sm.resolvers),
		"introspection":   sm.config.EnableIntrospection,
		"playground":      sm.config.EnablePlayground,
		"max_query_depth": sm.config.MaxQueryDepth,
		"query_timeout":   sm.config.QueryTimeout.String(),
		"tracing_enabled": sm.config.EnableTracing,
		"metrics_enabled": sm.config.EnableMetrics,
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
