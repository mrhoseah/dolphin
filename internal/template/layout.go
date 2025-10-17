package template

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// Layout represents a template layout
type Layout struct {
	Name         string                 `json:"name"`
	Content      string                 `json:"content"`
	Compiled     *template.Template     `json:"-"`
	Blocks       map[string]string      `json:"blocks"`
	Extends      string                 `json:"extends"`
	Includes     []string               `json:"includes"`
	Variables    map[string]interface{} `json:"variables"`
	LastModified time.Time              `json:"last_modified"`
}

// LayoutManager manages template layouts
type LayoutManager struct {
	layouts map[string]*Layout
	engine  *Engine
}

// NewLayoutManager creates a new layout manager
func NewLayoutManager(engine *Engine) *LayoutManager {
	return &LayoutManager{
		layouts: make(map[string]*Layout),
		engine:  engine,
	}
}

// ParseLayout parses a layout template
func (lm *LayoutManager) ParseLayout(name, content string) (*Layout, error) {
	layout := &Layout{
		Name:      name,
		Content:   content,
		Blocks:    make(map[string]string),
		Includes:  make([]string, 0),
		Variables: make(map[string]interface{}),
	}

	// Parse layout content
	if err := lm.parseLayoutContent(layout); err != nil {
		return nil, err
	}

	// Compile layout
	if err := lm.compileLayout(layout); err != nil {
		return nil, err
	}

	return layout, nil
}

// parseLayoutContent parses layout content for blocks, extends, and includes
func (lm *LayoutManager) parseLayoutContent(layout *Layout) error {
	lines := strings.Split(layout.Content, "\n")
	var currentBlock string
	var blockContent []string

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Parse extends directive
		if strings.HasPrefix(line, "{{extends") {
			extends := lm.extractDirective(line, "extends")
			layout.Extends = extends
			continue
		}

		// Parse include directive
		if strings.HasPrefix(line, "{{include") {
			include := lm.extractDirective(line, "include")
			layout.Includes = append(layout.Includes, include)
			continue
		}

		// Parse block start
		if strings.HasPrefix(line, "{{block") {
			// Save previous block
			if currentBlock != "" {
				layout.Blocks[currentBlock] = strings.Join(blockContent, "\n")
			}

			// Start new block
			blockName := lm.extractDirective(line, "block")
			currentBlock = blockName
			blockContent = []string{}
			continue
		}

		// Parse block end
		if strings.HasPrefix(line, "{{endblock") {
			if currentBlock != "" {
				layout.Blocks[currentBlock] = strings.Join(blockContent, "\n")
				currentBlock = ""
				blockContent = []string{}
			}
			continue
		}

		// Add line to current block or main content
		if currentBlock != "" {
			blockContent = append(blockContent, lines[i])
		}
	}

	// Save last block
	if currentBlock != "" {
		layout.Blocks[currentBlock] = strings.Join(blockContent, "\n")
	}

	return nil
}

// extractDirective extracts a directive value from a template line
func (lm *LayoutManager) extractDirective(line, directive string) string {
	// Find directive start
	start := strings.Index(line, directive+" ")
	if start == -1 {
		return ""
	}

	// Find opening quote
	quoteStart := strings.Index(line[start:], "\"")
	if quoteStart == -1 {
		return ""
	}
	quoteStart += start + 1

	// Find closing quote
	quoteEnd := strings.Index(line[quoteStart:], "\"")
	if quoteEnd == -1 {
		return ""
	}
	quoteEnd += quoteStart

	return line[quoteStart:quoteEnd]
}

// compileLayout compiles a layout template
func (lm *LayoutManager) compileLayout(layout *Layout) error {
	// Create template with helpers
	funcMap := template.FuncMap{}

	// Add helper functions
	if lm.engine.config.EnableHelpers {
		for name, helper := range lm.engine.helpers {
			funcMap[name] = helper
		}
	}

	// Add layout-specific functions
	funcMap["block"] = lm.blockHelper(layout)
	funcMap["include"] = lm.includeHelper(layout)
	funcMap["extends"] = lm.extendsHelper(layout)

	// Compile template
	compiled, err := template.New(layout.Name).Funcs(funcMap).Parse(layout.Content)
	if err != nil {
		return err
	}

	layout.Compiled = compiled
	return nil
}

// blockHelper returns a helper function for rendering blocks
func (lm *LayoutManager) blockHelper(layout *Layout) func(string) string {
	return func(blockName string) string {
		if block, exists := layout.Blocks[blockName]; exists {
			return block
		}
		return ""
	}
}

// includeHelper returns a helper function for including partials
func (lm *LayoutManager) includeHelper(layout *Layout) func(string, ...interface{}) (string, error) {
	return func(partialName string, data ...interface{}) (string, error) {
		// Get partial template
		partial, exists := lm.engine.partials[partialName]
		if !exists {
			return "", fmt.Errorf("partial %s not found", partialName)
		}

		// Prepare data
		templateData := make(TemplateData)
		if len(data) > 0 {
			if dataMap, ok := data[0].(map[string]interface{}); ok {
				templateData = TemplateData(dataMap)
			}
		}

		// Render partial
		return lm.engine.RenderPartial(partialName, templateData)
	}
}

// extendsHelper returns a helper function for extending layouts
func (lm *LayoutManager) extendsHelper(layout *Layout) func(string, map[string]interface{}) (string, error) {
	return func(parentLayout string, data map[string]interface{}) (string, error) {
		// Get parent layout
		parent, exists := lm.engine.layouts[parentLayout]
		if !exists {
			return "", fmt.Errorf("parent layout %s not found", parentLayout)
		}

		// Merge blocks from current layout into parent
		mergedBlocks := make(map[string]string)
		for name, content := range parent.Blocks {
			mergedBlocks[name] = content
		}
		for name, content := range layout.Blocks {
			mergedBlocks[name] = content
		}

		// Create merged layout
		mergedLayout := &Layout{
			Name:     parent.Name,
			Content:  parent.Content,
			Blocks:   mergedBlocks,
			Extends:  parent.Extends,
			Includes: parent.Includes,
		}

		// Compile merged layout
		if err := lm.compileLayout(mergedLayout); err != nil {
			return "", err
		}

		// Render merged layout
		var buf strings.Builder
		if err := mergedLayout.Compiled.Execute(&buf, data); err != nil {
			return "", err
		}

		return buf.String(), nil
	}
}

// RenderLayout renders a layout with data
func (lm *LayoutManager) RenderLayout(layoutName string, data TemplateData) (string, error) {
	layout, exists := lm.layouts[layoutName]
	if !exists {
		return "", fmt.Errorf("layout %s not found", layoutName)
	}

	// Render layout
	var buf strings.Builder
	if err := layout.Compiled.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RegisterLayout registers a layout
func (lm *LayoutManager) RegisterLayout(layout *Layout) {
	lm.layouts[layout.Name] = layout
}

// GetLayout returns a layout by name
func (lm *LayoutManager) GetLayout(name string) (*Layout, bool) {
	layout, exists := lm.layouts[name]
	return layout, exists
}

// GetAllLayouts returns all layouts
func (lm *LayoutManager) GetAllLayouts() map[string]*Layout {
	return lm.layouts
}

// LayoutBuilder provides a fluent interface for building layouts
type LayoutBuilder struct {
	layout *Layout
	lm     *LayoutManager
}

// NewLayoutBuilder creates a new layout builder
func NewLayoutBuilder(name string, lm *LayoutManager) *LayoutBuilder {
	return &LayoutBuilder{
		layout: &Layout{
			Name:      name,
			Blocks:    make(map[string]string),
			Includes:  make([]string, 0),
			Variables: make(map[string]interface{}),
		},
		lm: lm,
	}
}

// Content sets the layout content
func (lb *LayoutBuilder) Content(content string) *LayoutBuilder {
	lb.layout.Content = content
	return lb
}

// Block adds a block to the layout
func (lb *LayoutBuilder) Block(name, content string) *LayoutBuilder {
	lb.layout.Blocks[name] = content
	return lb
}

// Extends sets the parent layout
func (lb *LayoutBuilder) Extends(parentLayout string) *LayoutBuilder {
	lb.layout.Extends = parentLayout
	return lb
}

// Include adds an include to the layout
func (lb *LayoutBuilder) Include(partialName string) *LayoutBuilder {
	lb.layout.Includes = append(lb.layout.Includes, partialName)
	return lb
}

// Variable sets a layout variable
func (lb *LayoutBuilder) Variable(name string, value interface{}) *LayoutBuilder {
	lb.layout.Variables[name] = value
	return lb
}

// Build builds the layout
func (lb *LayoutBuilder) Build() (*Layout, error) {
	// Compile layout
	if err := lb.lm.compileLayout(lb.layout); err != nil {
		return nil, err
	}

	// Register layout
	lb.lm.RegisterLayout(lb.layout)

	return lb.layout, nil
}

// LayoutHelper provides helper functions for layouts
type LayoutHelper struct {
	engine *Engine
}

// NewLayoutHelper creates a new layout helper
func NewLayoutHelper(engine *Engine) *LayoutHelper {
	return &LayoutHelper{
		engine: engine,
	}
}

// RenderBlock renders a block with fallback
func (lh *LayoutHelper) RenderBlock(blockName, fallback string) string {
	// This would typically be called from within a template
	// For now, return the fallback
	return fallback
}

// RenderPartial renders a partial within a layout
func (lh *LayoutHelper) RenderPartial(partialName string, data TemplateData) (string, error) {
	return lh.engine.RenderPartial(partialName, data)
}

// RenderComponent renders a component within a layout
func (lh *LayoutHelper) RenderComponent(componentName string, data TemplateData) (string, error) {
	return lh.engine.RenderComponent(componentName, data)
}

// LayoutVariable gets a layout variable
func (lh *LayoutHelper) LayoutVariable(name string, defaultValue interface{}) interface{} {
	// This would typically be called from within a template
	// For now, return the default value
	return defaultValue
}

// LayoutExists checks if a layout exists
func (lh *LayoutHelper) LayoutExists(name string) bool {
	_, exists := lh.engine.layouts[name]
	return exists
}

// BlockExists checks if a block exists
func (lh *LayoutHelper) BlockExists(blockName string) bool {
	// This would typically be called from within a template
	// For now, return false
	return false
}
