package graphql

import (
	"context"
	"fmt"
	"math"

	"github.com/graphql-go/graphql"
	"go.uber.org/zap"
)

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor"`
	EndCursor       *string `json:"endCursor"`
}

// Edge represents a connection edge
type Edge struct {
	Node   interface{} `json:"node"`
	Cursor string      `json:"cursor"`
}

// Connection represents a paginated connection
type Connection struct {
	Edges    []Edge   `json:"edges"`
	PageInfo PageInfo `json:"pageInfo"`
}

// ConnectionArgs represents arguments for connection queries
type ConnectionArgs struct {
	First  *int    `json:"first"`
	Last   *int    `json:"last"`
	After  *string `json:"after"`
	Before *string `json:"before"`
}

// Cursor represents a pagination cursor
type Cursor struct {
	Value string
}

// EncodeCursor encodes a cursor value
func EncodeCursor(value string) string {
	return value // In a real implementation, this would be base64 encoded
}

// DecodeCursor decodes a cursor value
func DecodeCursor(cursor string) (string, error) {
	return cursor, nil // In a real implementation, this would be base64 decoded
}

// CreatePageInfoType creates the GraphQL PageInfo type
func CreatePageInfoType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "PageInfo",
		Fields: graphql.Fields{
			"hasNextPage": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "When paginating forwards, are there more items?",
			},
			"hasPreviousPage": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "When paginating backwards, are there more items?",
			},
			"startCursor": &graphql.Field{
				Type:        graphql.String,
				Description: "When paginating backwards, the cursor to continue.",
			},
			"endCursor": &graphql.Field{
				Type:        graphql.String,
				Description: "When paginating forwards, the cursor to continue.",
			},
		},
	})
}

// CreateEdgeType creates a GraphQL Edge type for a given node type
func CreateEdgeType(nodeType *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: fmt.Sprintf("%sEdge", nodeType.Name()),
		Fields: graphql.Fields{
			"node": &graphql.Field{
				Type:        nodeType,
				Description: "The item at the end of the edge",
			},
			"cursor": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "A cursor for use in pagination",
			},
		},
	})
}

// CreateConnectionType creates a GraphQL Connection type for a given edge type
func CreateConnectionType(edgeType *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: fmt.Sprintf("%sConnection", edgeType.Name()),
		Fields: graphql.Fields{
			"edges": &graphql.Field{
				Type:        graphql.NewList(edgeType),
				Description: "A list of edges",
			},
			"pageInfo": &graphql.Field{
				Type:        graphql.NewNonNull(CreatePageInfoType()),
				Description: "Information to aid in pagination",
			},
		},
	})
}

// ConnectionResolver resolves a connection
type ConnectionResolver func(ctx context.Context, args ConnectionArgs) (*Connection, error)

// CreateConnectionField creates a connection field
func CreateConnectionField(
	name string,
	nodeType *graphql.Object,
	resolver ConnectionResolver,
) *graphql.Field {
	edgeType := CreateEdgeType(nodeType)
	connectionType := CreateConnectionType(edgeType)

	return &graphql.Field{
		Type: connectionType,
		Args: graphql.FieldConfigArgument{
			"first": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "Returns the first n elements from the list",
			},
			"last": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "Returns the last n elements from the list",
			},
			"after": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "Returns the elements in the list that come after the specified cursor",
			},
			"before": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "Returns the elements in the list that come before the specified cursor",
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Parse arguments
			args := ConnectionArgs{}
			if first, ok := p.Args["first"].(int); ok {
				args.First = &first
			}
			if last, ok := p.Args["last"].(int); ok {
				args.Last = &last
			}
			if after, ok := p.Args["after"].(string); ok {
				args.After = &after
			}
			if before, ok := p.Args["before"].(string); ok {
				args.Before = &before
			}

			// Resolve the connection
			return resolver(p.Context, args)
		},
	}
}

// PaginationHelper provides helper functions for pagination
type PaginationHelper struct {
	logger *zap.Logger
}

// NewPaginationHelper creates a new pagination helper
func NewPaginationHelper(logger *zap.Logger) *PaginationHelper {
	return &PaginationHelper{logger: logger}
}

// ApplyPagination applies pagination to a slice of items
func (ph *PaginationHelper) ApplyPagination(
	items []interface{},
	args ConnectionArgs,
) ([]interface{}, PageInfo) {
	total := len(items)
	if total == 0 {
		return []interface{}{}, PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
		}
	}

	// Determine pagination direction and limits
	var start, end int
	var hasNext, hasPrev bool

	if args.First != nil {
		// Forward pagination
		limit := *args.First
		start = 0
		if args.After != nil {
			// Find the cursor position
			cursor, err := DecodeCursor(*args.After)
			if err == nil {
				start = ph.findCursorPosition(items, cursor) + 1
			}
		}
		end = start + limit
		if end > total {
			end = total
		}
		hasNext = end < total
		hasPrev = start > 0
	} else if args.Last != nil {
		// Backward pagination
		limit := *args.Last
		end = total
		if args.Before != nil {
			// Find the cursor position
			cursor, err := DecodeCursor(*args.Before)
			if err == nil {
				end = ph.findCursorPosition(items, cursor)
			}
		}
		start = end - limit
		if start < 0 {
			start = 0
		}
		hasNext = end < total
		hasPrev = start > 0
	} else {
		// No pagination
		start = 0
		end = total
		hasNext = false
		hasPrev = false
	}

	// Apply pagination
	paginatedItems := items[start:end]

	// Create page info
	var startCursor, endCursor *string
	if len(paginatedItems) > 0 {
		start := EncodeCursor(fmt.Sprintf("%d", start))
		end := EncodeCursor(fmt.Sprintf("%d", end-1))
		startCursor = &start
		endCursor = &end
	}

	pageInfo := PageInfo{
		HasNextPage:     hasNext,
		HasPreviousPage: hasPrev,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	return paginatedItems, pageInfo
}

// findCursorPosition finds the position of a cursor in the items
func (ph *PaginationHelper) findCursorPosition(items []interface{}, cursor string) int {
	// This is a simplified implementation
	// In a real implementation, you would decode the cursor and find the actual position
	for i, item := range items {
		if fmt.Sprintf("%v", item) == cursor {
			return i
		}
	}
	return 0
}

// CreateEdges creates edges from items
func (ph *PaginationHelper) CreateEdges(items []interface{}) []Edge {
	edges := make([]Edge, len(items))
	for i, item := range items {
		edges[i] = Edge{
			Node:   item,
			Cursor: EncodeCursor(fmt.Sprintf("%d", i)),
		}
	}
	return edges
}

// ValidateConnectionArgs validates connection arguments
func (ph *PaginationHelper) ValidateConnectionArgs(args ConnectionArgs) error {
	if args.First != nil && args.Last != nil {
		return fmt.Errorf("cannot specify both first and last")
	}
	if args.After != nil && args.Before != nil {
		return fmt.Errorf("cannot specify both after and before")
	}
	if args.First != nil && *args.First < 0 {
		return fmt.Errorf("first must be positive")
	}
	if args.Last != nil && *args.Last < 0 {
		return fmt.Errorf("last must be positive")
	}
	return nil
}

// CalculateTotalPages calculates total pages for a given limit
func (ph *PaginationHelper) CalculateTotalPages(total, limit int) int {
	if limit <= 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
