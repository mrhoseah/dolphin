package graphql

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"go.uber.org/zap"
)

// QueryAnalysisResult contains the results of query analysis
type QueryAnalysisResult struct {
	Depth      int
	Complexity int
	FieldCount int
	Valid      bool
	Errors     []string
}

// QueryAnalyzer analyzes GraphQL queries for depth and complexity
type QueryAnalyzer struct {
	maxDepth      int
	maxComplexity int
	logger        *zap.Logger
}

// NewQueryAnalyzer creates a new query analyzer
func NewQueryAnalyzer(maxDepth, maxComplexity int, logger *zap.Logger) *QueryAnalyzer {
	return &QueryAnalyzer{
		maxDepth:      maxDepth,
		maxComplexity: maxComplexity,
		logger:        logger,
	}
}

// AnalyzeQuery analyzes a GraphQL query
func (qa *QueryAnalyzer) AnalyzeQuery(query string) (*QueryAnalysisResult, error) {
	// Parse the query
	document, err := parser.Parse(parser.ParseParams{
		Source: source.Source{Body: []byte(query)},
	})
	if err != nil {
		return &QueryAnalysisResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Parse error: %v", err)},
		}, err
	}

	result := &QueryAnalysisResult{
		Valid:  true,
		Errors: []string{},
	}

	// Analyze the document
	qa.analyzeDocument(document, result)

	// Check limits
	if result.Depth > qa.maxDepth {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Query depth %d exceeds maximum %d", result.Depth, qa.maxDepth))
	}

	if result.Complexity > qa.maxComplexity {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Query complexity %d exceeds maximum %d", result.Complexity, qa.maxComplexity))
	}

	qa.logger.Info("Query analyzed",
		zap.Int("depth", result.Depth),
		zap.Int("complexity", result.Complexity),
		zap.Int("field_count", result.FieldCount),
		zap.Bool("valid", result.Valid),
	)

	return result, nil
}

// analyzeDocument analyzes a GraphQL document
func (qa *QueryAnalyzer) analyzeDocument(document *ast.Document, result *QueryAnalysisResult) {
	for _, definition := range document.Definitions {
		switch def := definition.(type) {
		case *ast.OperationDefinition:
			qa.analyzeOperation(def, result, 0)
		case *ast.FragmentDefinition:
			// Fragments are analyzed when referenced
			continue
		}
	}
}

// analyzeOperation analyzes a GraphQL operation
func (qa *QueryAnalyzer) analyzeOperation(operation *ast.OperationDefinition, result *QueryAnalysisResult, depth int) {
	if operation.SelectionSet == nil {
		return
	}

	// Update depth
	if depth > result.Depth {
		result.Depth = depth
	}

	// Analyze selection set
	qa.analyzeSelectionSet(operation.SelectionSet, result, depth)
}

// analyzeSelectionSet analyzes a selection set
func (qa *QueryAnalyzer) analyzeSelectionSet(selectionSet *ast.SelectionSet, result *QueryAnalysisResult, depth int) {
	for _, selection := range selectionSet.Selections {
		qa.analyzeSelection(selection, result, depth)
	}
}

// analyzeSelection analyzes a selection
func (qa *QueryAnalyzer) analyzeSelection(selection ast.Selection, result *QueryAnalysisResult, depth int) {
	switch sel := selection.(type) {
	case *ast.Field:
		qa.analyzeField(sel, result, depth)
	case *ast.FragmentSpread:
		// Fragment spreads are handled by the resolver
		qa.analyzeField(&ast.Field{
			Name: &ast.Name{Value: "fragment"},
		}, result, depth)
	case *ast.InlineFragment:
		if sel.SelectionSet != nil {
			qa.analyzeSelectionSet(sel.SelectionSet, result, depth)
		}
	}
}

// analyzeField analyzes a field
func (qa *QueryAnalyzer) analyzeField(field *ast.Field, result *QueryAnalysisResult, depth int) {
	// Count fields
	result.FieldCount++

	// Calculate field complexity
	fieldComplexity := qa.calculateFieldComplexity(field)
	result.Complexity += fieldComplexity

	// Analyze nested fields
	if field.SelectionSet != nil {
		qa.analyzeSelectionSet(field.SelectionSet, result, depth+1)
	}

	qa.logger.Debug("Field analyzed",
		zap.String("name", field.Name.Value),
		zap.Int("depth", depth),
		zap.Int("complexity", fieldComplexity),
	)
}

// calculateFieldComplexity calculates the complexity of a field
func (qa *QueryAnalyzer) calculateFieldComplexity(field *ast.Field) int {
	complexity := 1 // Base complexity

	// Add complexity for arguments
	if field.Arguments != nil {
		complexity += len(field.Arguments)
	}

	// Add complexity for nested fields
	if field.SelectionSet != nil {
		complexity += qa.countNestedFields(field.SelectionSet)
	}

	// Add complexity based on field name (some fields are more expensive)
	fieldName := strings.ToLower(field.Name.Value)
	switch fieldName {
	case "users", "posts", "comments":
		complexity *= 2 // List fields are more expensive
	case "search", "filter":
		complexity *= 3 // Search/filter operations are expensive
	}

	return complexity
}

// countNestedFields counts nested fields recursively
func (qa *QueryAnalyzer) countNestedFields(selectionSet *ast.SelectionSet) int {
	count := 0
	for _, selection := range selectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			count++
			if sel.SelectionSet != nil {
				count += qa.countNestedFields(sel.SelectionSet)
			}
		case *ast.FragmentSpread:
			count++
		case *ast.InlineFragment:
			if sel.SelectionSet != nil {
				count += qa.countNestedFields(sel.SelectionSet)
			}
		}
	}
	return count
}

// QueryComplexityCalculator provides advanced complexity calculation
type QueryComplexityCalculator struct {
	fieldComplexity map[string]int
	logger          *zap.Logger
}

// NewQueryComplexityCalculator creates a new complexity calculator
func NewQueryComplexityCalculator(logger *zap.Logger) *QueryComplexityCalculator {
	return &QueryComplexityCalculator{
		fieldComplexity: make(map[string]int),
		logger:          logger,
	}
}

// SetFieldComplexity sets the complexity for a specific field
func (qcc *QueryComplexityCalculator) SetFieldComplexity(fieldName string, complexity int) {
	qcc.fieldComplexity[fieldName] = complexity
	qcc.logger.Info("Field complexity set",
		zap.String("field", fieldName),
		zap.Int("complexity", complexity),
	)
}

// CalculateComplexity calculates the complexity of a field
func (qcc *QueryComplexityCalculator) CalculateComplexity(fieldName string) int {
	if complexity, exists := qcc.fieldComplexity[fieldName]; exists {
		return complexity
	}
	return 1 // Default complexity
}

// QueryDepthAnalyzer provides advanced depth analysis
type QueryDepthAnalyzer struct {
	maxDepth int
	logger   *zap.Logger
}

// NewQueryDepthAnalyzer creates a new depth analyzer
func NewQueryDepthAnalyzer(maxDepth int, logger *zap.Logger) *QueryDepthAnalyzer {
	return &QueryDepthAnalyzer{
		maxDepth: maxDepth,
		logger:   logger,
	}
}

// AnalyzeDepth analyzes the depth of a query
func (qda *QueryDepthAnalyzer) AnalyzeDepth(query string) (int, error) {
	document, err := parser.Parse(parser.ParseParams{
		Source: source.Source{Body: []byte(query)},
	})
	if err != nil {
		return 0, err
	}

	maxDepth := 0
	for _, definition := range document.Definitions {
		if operation, ok := definition.(*ast.OperationDefinition); ok {
			depth := qda.calculateDepth(operation, 0)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	qda.logger.Info("Query depth analyzed",
		zap.Int("depth", maxDepth),
		zap.Int("max_allowed", qda.maxDepth),
	)

	return maxDepth, nil
}

// calculateDepth calculates the depth of a selection
func (qda *QueryDepthAnalyzer) calculateDepth(selection ast.Selection, currentDepth int) int {
	maxDepth := currentDepth

	switch sel := selection.(type) {
	case *ast.Field:
		if sel.SelectionSet != nil {
			for _, nestedSelection := range sel.SelectionSet.Selections {
				depth := qda.calculateDepth(nestedSelection, currentDepth+1)
				if depth > maxDepth {
					maxDepth = depth
				}
			}
		}
	case *ast.FragmentSpread:
		// Fragment depth is handled by the resolver
		maxDepth = currentDepth + 1
	case *ast.InlineFragment:
		if sel.SelectionSet != nil {
			for _, nestedSelection := range sel.SelectionSet.Selections {
				depth := qda.calculateDepth(nestedSelection, currentDepth+1)
				if depth > maxDepth {
					maxDepth = depth
				}
			}
		}
	}

	return maxDepth
}

// QueryValidator validates GraphQL queries
type QueryValidator struct {
	analyzer   *QueryAnalyzer
	calculator *QueryComplexityCalculator
	logger     *zap.Logger
}

// NewQueryValidator creates a new query validator
func NewQueryValidator(maxDepth, maxComplexity int, logger *zap.Logger) *QueryValidator {
	return &QueryValidator{
		analyzer:   NewQueryAnalyzer(maxDepth, maxComplexity, logger),
		calculator: NewQueryComplexityCalculator(logger),
		logger:     logger,
	}
}

// ValidateQuery validates a GraphQL query
func (qv *QueryValidator) ValidateQuery(query string) (*QueryAnalysisResult, error) {
	result, err := qv.analyzer.AnalyzeQuery(query)
	if err != nil {
		return result, err
	}

	if !result.Valid {
		qv.logger.Warn("Query validation failed",
			zap.Strings("errors", result.Errors),
		)
	}

	return result, nil
}

// SetFieldComplexity sets the complexity for a field
func (qv *QueryValidator) SetFieldComplexity(fieldName string, complexity int) {
	qv.calculator.SetFieldComplexity(fieldName, complexity)
}
