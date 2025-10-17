package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"go.uber.org/zap"
)

// Node represents the Global Object Identification interface
type Node interface {
	GetID() string
	GetType() string
}

// NodeResolver resolves nodes by their global ID
type NodeResolver func(ctx context.Context, id string) (Node, error)

// NodeRegistry manages node types and their resolvers
type NodeRegistry struct {
	resolvers map[string]NodeResolver
	logger    *zap.Logger
}

// NewNodeRegistry creates a new node registry
func NewNodeRegistry(logger *zap.Logger) *NodeRegistry {
	return &NodeRegistry{
		resolvers: make(map[string]NodeResolver),
		logger:    logger,
	}
}

// RegisterNodeType registers a node type with its resolver
func (nr *NodeRegistry) RegisterNodeType(nodeType string, resolver NodeResolver) {
	nr.resolvers[nodeType] = resolver
	nr.logger.Info("Registered node type", zap.String("type", nodeType))
}

// ResolveNode resolves a node by its global ID
func (nr *NodeRegistry) ResolveNode(ctx context.Context, globalID string) (Node, error) {
	// Decode the global ID to get type and ID
	nodeType, id, err := DecodeGlobalID(globalID)
	if err != nil {
		return nil, fmt.Errorf("invalid global ID: %w", err)
	}

	// Get the resolver for this node type
	resolver, exists := nr.resolvers[nodeType]
	if !exists {
		return nil, fmt.Errorf("unknown node type: %s", nodeType)
	}

	// Resolve the node
	return resolver(ctx, id)
}

// EncodeGlobalID creates a global ID from type and ID
func EncodeGlobalID(nodeType, id string) string {
	combined := fmt.Sprintf("%s:%s", nodeType, id)
	return base64.StdEncoding.EncodeToString([]byte(combined))
}

// DecodeGlobalID decodes a global ID to get type and ID
func DecodeGlobalID(globalID string) (nodeType, id string, err error) {
	decoded, err := base64.StdEncoding.DecodeString(globalID)
	if err != nil {
		return "", "", err
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid global ID format")
	}

	return parts[0], parts[1], nil
}

// CreateNodeInterface creates the GraphQL Node interface
func CreateNodeInterface() *graphql.Interface {
	return graphql.NewInterface(graphql.InterfaceConfig{
		Name: "Node",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
				Description: "The globally unique identifier for this object",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			// This will be implemented by individual types
			return nil
		},
	})
}

// CreateNodeQuery creates the root node query field
func CreateNodeQuery(nodeRegistry *NodeRegistry) *graphql.Field {
	return &graphql.Field{
		Type: CreateNodeInterface(),
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "The global ID of the object to fetch",
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, ok := p.Args["id"].(string)
			if !ok {
				return nil, fmt.Errorf("id is required")
			}

			// Resolve the node
			node, err := nodeRegistry.ResolveNode(p.Context, id)
			if err != nil {
				return nil, err
			}

			return node, nil
		},
	}
}

// NodeType represents a concrete node type
type NodeType struct {
	Name        string
	Object      *graphql.Object
	NodeResolver NodeResolver
}

// CreateNodeType creates a node type that implements the Node interface
func CreateNodeType(name string, fields graphql.Fields, nodeResolver NodeResolver) *NodeType {
	// Add the id field to the fields
	fields["id"] = &graphql.Field{
		Type: graphql.NewNonNull(graphql.ID),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Get the node from the source
			if node, ok := p.Source.(Node); ok {
				return node.GetID(), nil
			}
			return nil, fmt.Errorf("source does not implement Node interface")
		},
	}

	object := graphql.NewObject(graphql.ObjectConfig{
		Name:       name,
		Fields:     fields,
		Interfaces: []*graphql.Interface{CreateNodeInterface()},
	})

	return &NodeType{
		Name:         name,
		Object:       object,
		NodeResolver: nodeResolver,
	}
}

// GetNodeType returns the GraphQL object type
func (nt *NodeType) GetNodeType() *graphql.Object {
	return nt.Object
}

// GetNodeResolver returns the node resolver
func (nt *NodeType) GetNodeResolver() NodeResolver {
	return nt.NodeResolver
}
