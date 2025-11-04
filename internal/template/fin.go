package template

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// FinEngine represents the Fin template engine
type FinEngine struct {
	viewsPath     string
	cachePath     string
	cacheEnabled  bool
	debugMode     bool
	extensions    []string
	directives    map[string]DirectiveFunc
	components    map[string]*Component
	layouts       map[string]*Layout
	mu            sync.RWMutex
	compiledCache map[string]*CompiledTemplate
}

// DirectiveFunc represents a template directive function
type DirectiveFunc func(args []string, content string, data interface{}) (string, error)

// Component represents a reusable template component
type Component struct {
	Name     string
	Template string
	Slots    map[string]string
	Props    map[string]interface{}
}

// Layout represents a template layout
type Layout struct {
	Name     string
	Template string
	Sections map[string]string
}

// CompiledTemplate represents a compiled template
type CompiledTemplate struct {
	Template *template.Template
	Compiled string
	Hash     string
	Created  time.Time
}

// Config represents template engine configuration
type Config struct {
	ViewsPath    string
	CachePath    string
	CacheEnabled bool
	DebugMode    bool
	Extensions   []string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ViewsPath:    "views",
		CachePath:    "storage/cache/views",
		CacheEnabled: true,
		DebugMode:    false,
		Extensions:   []string{".fin.html"}, // Only .fin.html is supported
	}
}

// NewFinEngine creates a new Fin template engine
func NewFinEngine(config *Config) *FinEngine {
	if config == nil {
		config = DefaultConfig()
	}

	engine := &FinEngine{
		viewsPath:     config.ViewsPath,
		cachePath:     config.CachePath,
		cacheEnabled:  config.CacheEnabled,
		debugMode:     config.DebugMode,
		extensions:    config.Extensions,
		directives:    make(map[string]DirectiveFunc),
		components:    make(map[string]*Component),
		layouts:       make(map[string]*Layout),
		compiledCache: make(map[string]*CompiledTemplate),
	}

	// Register default directives
	engine.registerDefaultDirectives()

	// Ensure cache directory exists
	if engine.cacheEnabled {
		os.MkdirAll(engine.cachePath, 0755)
	}

	return engine
}

// registerDefaultDirectives registers built-in Fin template directives
func (e *FinEngine) registerDefaultDirectives() {
	// @extends directive
	e.RegisterDirective("extends", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("@extends requires a layout name")
		}
		layoutName := args[0]
		return e.renderWithLayout(layoutName, content, data)
	})

	// @section directive
	e.RegisterDirective("section", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("@section requires a name")
		}
		sectionName := args[0]
		return fmt.Sprintf("{{define \"%s\"}}%s{{end}}", sectionName, content), nil
	})

	// @yield directive
	e.RegisterDirective("yield", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("@yield requires a section name")
		}
		sectionName := args[0]
		return fmt.Sprintf("{{template \"%s\" .}}", sectionName), nil
	})

	// @if directive
	e.RegisterDirective("if", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("@if requires a condition")
		}
		condition := args[0]
		return fmt.Sprintf("{{if %s}}%s{{end}}", condition, content), nil
	})

	// @elseif directive
	e.RegisterDirective("elseif", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("@elseif requires a condition")
		}
		condition := args[0]
		return fmt.Sprintf("{{else if %s}}%s", condition, content), nil
	})

	// @else directive
	e.RegisterDirective("else", func(args []string, content string, data interface{}) (string, error) {
		return fmt.Sprintf("{{else}}%s", content), nil
	})

	// @endif directive
	e.RegisterDirective("endif", func(args []string, content string, data interface{}) (string, error) {
		return "{{end}}", nil
	})

	// @foreach directive
	e.RegisterDirective("foreach", func(args []string, content string, data interface{}) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("@foreach requires 'item' and 'collection'")
		}
		item := args[0]
		collection := args[1]
		return fmt.Sprintf("{{range %s}}%s{{end}}", collection, strings.ReplaceAll(content, "$"+item, ".")), nil
	})

	// @endforeach directive
	e.RegisterDirective("endforeach", func(args []string, content string, data interface{}) (string, error) {
		return "{{end}}", nil
	})

	// @auth directive
	e.RegisterDirective("auth", func(args []string, content string, data interface{}) (string, error) {
		return fmt.Sprintf("{{if .User}}%s{{end}}", content), nil
	})

	// @guest directive
	e.RegisterDirective("guest", func(args []string, content string, data interface{}) (string, error) {
		return fmt.Sprintf("{{if not .User}}%s{{end}}", content), nil
	})

	// @csrf directive
	e.RegisterDirective("csrf", func(args []string, content string, data interface{}) (string, error) {
		return `<input type="hidden" name="_token" value="{{.CSRFToken}}">`, nil
	})

	// @component directive
	e.RegisterDirective("component", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("@component requires a component name")
		}
		componentName := args[0]
		return e.renderComponent(componentName, content, data), nil
	})

	// @slot directive
	e.RegisterDirective("slot", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("@slot requires a name")
		}
		slotName := args[0]
		return fmt.Sprintf("{{define \"slot_%s\"}}%s{{end}}", slotName, content), nil
	})

	// @endslot directive
	e.RegisterDirective("endslot", func(args []string, content string, data interface{}) (string, error) {
		return "", nil
	})

	// @model directive for clean model annotations (deprecated - use controller data instead)
	e.RegisterDirective("model", func(args []string, content string, data interface{}) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("@model requires model type and variable name")
		}
		modelType := args[0]
		varName := args[1]
		// Store model information for later use in template processing
		// Note: This is deprecated - controllers should provide data instead
		return fmt.Sprintf("{{/* Model: %s as %s (deprecated - use controller data) */}}", modelType, varName), nil
	})

	// @foreach directive with improved syntax
	e.RegisterDirective("foreach", func(args []string, content string, data interface{}) (string, error) {
		if len(args) < 2 {
			return "", fmt.Errorf("@foreach requires collection and item variable")
		}
		collection := args[0]
		_ = args[1] // itemVar - used for validation but not in template generation
		return fmt.Sprintf("{{range %s}}%s{{end}}", collection, content), nil
	})

	// @endforeach directive
	e.RegisterDirective("endforeach", func(args []string, content string, data interface{}) (string, error) {
		return "", nil
	})
}

// RegisterDirective registers a custom Fin template directive
func (e *FinEngine) RegisterDirective(name string, fn DirectiveFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.directives[name] = fn
}

// RegisterComponent registers a reusable Fin component
func (e *FinEngine) RegisterComponent(name string, template string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.components[name] = &Component{
		Name:     name,
		Template: template,
		Slots:    make(map[string]string),
		Props:    make(map[string]interface{}),
	}
}

// RegisterLayout registers a Fin template layout
func (e *FinEngine) RegisterLayout(name string, template string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.layouts[name] = &Layout{
		Name:     name,
		Template: template,
		Sections: make(map[string]string),
	}
}

// Render renders a Fin template with data
// Template names should use .fin.html extension or omit extension (will be enforced)
func (e *FinEngine) Render(templateName string, data interface{}) (string, error) {
	// Enforce .fin.html extension
	if !strings.HasSuffix(templateName, ".fin.html") {
		if strings.Contains(templateName, ".") {
			// Replace any extension with .fin.html
			if e.debugMode {
				fmt.Printf("Warning: Template '%s' should use .fin.html extension. Converting...\n", templateName)
			}
			templateName = strings.TrimSuffix(templateName, filepath.Ext(templateName)) + ".fin.html"
		} else {
			// If no extension provided, append .fin.html
			templateName = templateName + ".fin.html"
		}
	}

	// Get template content
	content, err := e.getTemplateContent(templateName)
	if err != nil {
		return "", err
	}

	// Check if template uses {{extend}} directive BEFORE compiling
	extendRegex := regexp.MustCompile(`\{\{\s*extend\s+['"]([^'"]+)['"]\s*\}\}`)
	if extendMatch := extendRegex.FindStringSubmatch(content); extendMatch != nil {
		layoutName := extendMatch[1]
		// Enforce .fin.html extension for layout
		if !strings.HasSuffix(layoutName, ".fin.html") {
			layoutName = strings.TrimSuffix(layoutName, filepath.Ext(layoutName)) + ".fin.html"
		}
		// Remove the extend line from content
		lines := strings.Split(content, "\n")
		var newLines []string
		for _, line := range lines {
			if !extendRegex.MatchString(line) {
				newLines = append(newLines, line)
			}
		}
		contentWithoutExtend := strings.Join(newLines, "\n")
		// Render with layout
		return e.renderWithLayout(layoutName, contentWithoutExtend, data)
	}

	// Compile template (no layout)
	compiled, err := e.compileTemplate(content, data)
	if err != nil {
		return "", err
	}

	// Execute template
	var buf bytes.Buffer
	tmpl, err := template.New(templateName).Parse(compiled)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderToWriter renders a Fin template to a writer
func (e *FinEngine) RenderToWriter(w io.Writer, templateName string, data interface{}) error {
	content, err := e.Render(templateName, data)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(content))
	return err
}

// getTemplateContent retrieves template content from file or cache
func (e *FinEngine) getTemplateContent(templateName string) (string, error) {
	// Check cache first
	if e.cacheEnabled {
		if cached := e.getCachedTemplate(templateName); cached != nil {
			return cached.Compiled, nil
		}
	}

	// Load from file
	filePath := e.findTemplateFile(templateName)
	if filePath == "" {
		return "", fmt.Errorf("template '%s' not found", templateName)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// findTemplateFile finds the template file
// Enforces .fin.html extension only
func (e *FinEngine) findTemplateFile(templateName string) string {
	// Remove any existing extension from templateName to enforce .fin.html
	templateName = strings.TrimSuffix(templateName, ".fin.html")
	templateName = strings.TrimSuffix(templateName, ".fin")
	templateName = strings.TrimSuffix(templateName, ".html")
	templateName = strings.TrimSuffix(templateName, ".go.html")

	// Only use .fin.html extension
	for _, ext := range e.extensions {
		filePath := filepath.Join(e.viewsPath, templateName+ext)
		if _, err := os.Stat(filePath); err == nil {
			return filePath
		}
	}

	// If not found with configured extensions, try .fin.html as fallback
	filePath := filepath.Join(e.viewsPath, templateName+".fin.html")
	if _, err := os.Stat(filePath); err == nil {
		return filePath
	}

	return ""
}

// compileTemplate compiles template content with directives
func (e *FinEngine) compileTemplate(content string, data interface{}) (string, error) {
	// Process directives
	compiled := content

	// Process @extends directive
	compiled = e.processExtendsDirective(compiled, data)

	// Process @section/@yield directives
	compiled = e.processSectionDirectives(compiled, data)

	// Process @component/@slot directives
	compiled = e.processComponentDirectives(compiled, data)

	// Process other directives
	compiled = e.processDirectives(compiled, data)

	// Convert Fin syntax to Go template syntax
	compiled = e.convertFinToGo(compiled)

	return compiled, nil
}

// processExtendsDirective processes @extends directive
// Enforces .fin.html extension in layout references
func (e *FinEngine) processExtendsDirective(content string, data interface{}) string {
	// Support both @extends('layout') and {{extend "layout"}} syntax
	extendsRegex := regexp.MustCompile(`@extends\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	extendRegex := regexp.MustCompile(`\{\{\s*extend\s+['"]([^'"]+)['"]\s*\}\}`)

	// Process @extends syntax
	matches := extendsRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		layoutName := match[1]
		// Enforce .fin.html extension
		if !strings.HasSuffix(layoutName, ".fin.html") {
			layoutName = strings.TrimSuffix(layoutName, filepath.Ext(layoutName)) + ".fin.html"
			// Replace in content
			newMatch := strings.Replace(match[0], match[1], layoutName, 1)
			content = strings.Replace(content, match[0], newMatch, 1)
		}
	}

	// Process {{extend}} syntax
	extendMatches := extendRegex.FindAllStringSubmatch(content, -1)
	for _, match := range extendMatches {
		layoutName := match[1]
		// Enforce .fin.html extension
		if !strings.HasSuffix(layoutName, ".fin.html") {
			layoutName = strings.TrimSuffix(layoutName, filepath.Ext(layoutName)) + ".fin.html"
		}

		// Remove the {{extend}} line from content - layout will be loaded separately
		// The extend directive will be processed by renderWithLayout
		lines := strings.Split(content, "\n")
		var newLines []string
		for _, line := range lines {
			if !extendRegex.MatchString(line) {
				newLines = append(newLines, line)
			}
		}
		content = strings.Join(newLines, "\n")

		// Store layout name for later processing
		// This will be handled by checking for extend directive in Render method
		_ = layoutName // Layout processing happens in renderWithLayout
	}

	return content
}

// processSectionDirectives processes @section and @yield directives
func (e *FinEngine) processSectionDirectives(content string, data interface{}) string {
	// Process @section directives
	sectionRegex := regexp.MustCompile(`@section\s*\(\s*['"]([^'"]+)['"]\s*\)\s*(.*?)\s*@endsection`)
	content = sectionRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := sectionRegex.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			sectionName := submatches[1]
			sectionContent := submatches[2]
			return fmt.Sprintf("{{define \"%s\"}}%s{{end}}", sectionName, sectionContent)
		}
		return match
	})

	// Process @yield directives
	yieldRegex := regexp.MustCompile(`@yield\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	content = yieldRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := yieldRegex.FindStringSubmatch(match)
		if len(submatches) >= 2 {
			sectionName := submatches[1]
			return fmt.Sprintf("{{template \"%s\" .}}", sectionName)
		}
		return match
	})

	return content
}

// processComponentDirectives processes @component and @slot directives
func (e *FinEngine) processComponentDirectives(content string, data interface{}) string {
	// Process @component directives
	componentRegex := regexp.MustCompile(`@component\s*\(\s*['"]([^'"]+)['"]\s*\)\s*(.*?)\s*@endcomponent`)
	content = componentRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := componentRegex.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			componentName := submatches[1]
			componentContent := submatches[2]
			return e.renderComponent(componentName, componentContent, data)
		}
		return match
	})

	return content
}

// processDirectives processes all other directives
func (e *FinEngine) processDirectives(content string, data interface{}) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for directiveName, directiveFunc := range e.directives {
		pattern := fmt.Sprintf(`@%s\s*\(([^)]*)\)`, directiveName)
		regex := regexp.MustCompile(pattern)

		content = regex.ReplaceAllStringFunc(content, func(match string) string {
			submatches := regex.FindStringSubmatch(match)
			if len(submatches) >= 2 {
				argsStr := submatches[1]
				args := e.parseDirectiveArgs(argsStr)

				result, err := directiveFunc(args, "", data)
				if err != nil {
					return fmt.Sprintf("<!-- Error: %v -->", err)
				}
				return result
			}
			return match
		})
	}

	return content
}

// convertFinToGo converts Fin syntax to Go template syntax
func (e *FinEngine) convertFinToGo(content string) string {
	// Go template keywords that should NOT be converted to variables
	goKeywords := map[string]bool{
		"define": true, "block": true, "template": true, "if": true, "else": true,
		"end": true, "range": true, "with": true, "index": true, "len": true,
		"eq": true, "ne": true, "lt": true, "le": true, "gt": true, "ge": true,
		"and": true, "or": true, "not": true, "print": true, "printf": true,
	}

	// Convert {{ $variable }} to {{.Variable}}
	content = regexp.MustCompile(`\{\{\s*\$(\w+)\s*\}\}`).ReplaceAllString(content, "{{.$1}}")

	// Convert {{ $variable.property }} to {{.Variable.Property}}
	content = regexp.MustCompile(`\{\{\s*\$(\w+)\.(\w+)\s*\}\}`).ReplaceAllString(content, "{{.$1.$2}}")

	// Convert {{ $variable['key'] }} to {{index .Variable "key"}}
	content = regexp.MustCompile(`\{\{\s*\$(\w+)\['([^']+)'\]\s*\}\}`).ReplaceAllString(content, "{{index .$1 \"$2\"}}")

	// Convert {{variable}} to {{.Variable}} (without $ prefix)
	// BUT skip Go template keywords and commands
	// First, protect Go template keywords like {{define}}, {{block}}, {{template}}, etc.
	// These have the pattern: {{keyword "string"}} or {{keyword .}} or {{keyword}}
	keywordPattern := regexp.MustCompile(`\{\{\s*(define|block|template|if|else|end|range|with)\s+[^}]+\}\}`)
	protectedKeywords := make(map[string]string)
	protectedCount := 0

	// Protect all Go template keywords by replacing them temporarily
	content = keywordPattern.ReplaceAllStringFunc(content, func(match string) string {
		protectedCount++
		placeholder := fmt.Sprintf("__PROTECTED_KEYWORD_%d__", protectedCount)
		protectedKeywords[placeholder] = match
		return placeholder
	})

	// Also protect standalone keywords like {{end}}
	standaloneKeywords := regexp.MustCompile(`\{\{\s*(end|else)\s*\}\}`)
	content = standaloneKeywords.ReplaceAllStringFunc(content, func(match string) string {
		protectedCount++
		placeholder := fmt.Sprintf("__PROTECTED_KEYWORD_%d__", protectedCount)
		protectedKeywords[placeholder] = match
		return placeholder
	})

	// Now convert remaining {{variable}} patterns
	content = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`).ReplaceAllStringFunc(content, func(match string) string {
		// Extract the variable name
		re := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
		matches := re.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		varName := matches[1]

		// Don't convert Go template keywords
		if goKeywords[varName] {
			return match
		}

		// Only convert if it's not already a Go template expression
		if strings.Contains(match, ".") || strings.Contains(match, "|") || strings.Contains(match, "(") {
			return match
		}
		return strings.Replace(match, "{{", "{{.", 1)
	})

	// Restore protected keywords
	for placeholder, original := range protectedKeywords {
		content = strings.Replace(content, placeholder, original, -1)
	}

	// Convert {{variable.property}} to {{.Variable.Property}}
	content = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`).ReplaceAllStringFunc(content, func(match string) string {
		// Only convert if it's not already a Go template expression
		if strings.Contains(match, "|") || strings.Contains(match, "(") {
			return match
		}
		return strings.Replace(match, "{{", "{{.", 1)
	})

	return content
}

// parseDirectiveArgs parses directive arguments
func (e *FinEngine) parseDirectiveArgs(argsStr string) []string {
	argsStr = strings.TrimSpace(argsStr)
	if argsStr == "" {
		return []string{}
	}

	var args []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	for _, char := range argsStr {
		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			} else {
				current.WriteRune(char)
			}
		} else if char == ',' && !inQuotes {
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	return args
}

// renderWithLayout renders content with a layout
// Enforces .fin.html extension for layout names
func (e *FinEngine) renderWithLayout(layoutName string, content string, data interface{}) (string, error) {
	// Enforce .fin.html extension
	if !strings.HasSuffix(layoutName, ".fin.html") {
		layoutName = strings.TrimSuffix(layoutName, filepath.Ext(layoutName)) + ".fin.html"
	}

	e.mu.RLock()
	layout, exists := e.layouts[layoutName]
	e.mu.RUnlock()

	if !exists {
		// Try loading layout from file if not in memory
		layoutContent, err := e.getTemplateContent(layoutName)
		if err != nil {
			return "", fmt.Errorf("layout '%s' not found", layoutName)
		}
		// Register layout in memory
		e.mu.Lock()
		e.layouts[layoutName] = &Layout{
			Name:     layoutName,
			Template: layoutContent,
			Sections: make(map[string]string),
		}
		layout = e.layouts[layoutName]
		e.mu.Unlock()
	}

	// Parse both templates into the same template set
	// Go templates require that {{define}} blocks be parsed into the same template set
	// as the {{block}} that references them.

	// First, compile the content template (which will have {{define "content"}})
	// Skip directive processing for extend since we're already handling it
	compiledContent := content
	// Remove any {{extend}} directive from content if it exists
	extendRegex := regexp.MustCompile(`\{\{\s*extend\s+['"]([^'"]+)['"]\s*\}\}`)
	compiledContent = extendRegex.ReplaceAllString(compiledContent, "")
	compiledContent = strings.TrimSpace(compiledContent)

	// Process other directives but not extend
	// Note: processSectionDirectives only processes @section syntax, not {{define}}
	compiledContent = e.processSectionDirectives(compiledContent, data)
	compiledContent = e.processComponentDirectives(compiledContent, data)
	compiledContent = e.processDirectives(compiledContent, data)
	compiledContent = e.convertFinToGo(compiledContent)

	// Compile layout template (which has {{block "content"}})
	// Don't process extend directive on layout - it's already loaded
	compiledLayout := layout.Template
	// Layout templates use Go template syntax directly ({{block}}, {{define}}, etc.)
	// Only process component directives if needed, but preserve Go template keywords
	compiledLayout = e.processComponentDirectives(compiledLayout, data)
	// DO NOT process directives or convert Fin syntax - layout is pure Go template syntax
	// {{block}} and {{define}} are Go template keywords and must remain unchanged

	// Create a new template set
	// Parse content first to define the {{define "content"}} block
	var buf bytes.Buffer

	// Create a template set with layout as the main template
	// The layout template will be the root template (not wrapped in {{define}})
	// The content template will define {{define "content"}} which the layout references via {{block}}
	tmpl := template.New(layoutName)

	// Parse layout first as the main template
	// This will be the root template that contains {{block "content"}}
	tmpl, err := tmpl.Parse(compiledLayout)
	if err != nil {
		return "", fmt.Errorf("failed to parse layout template: %w\nLayout preview (first 500 chars):\n%s", err, compiledLayout[:min(500, len(compiledLayout))])
	}

	// Parse content template into the same set
	// This defines {{define "content"}}...{{end}} which the layout's {{block}} will use
	tmpl, err = tmpl.Parse(compiledContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse content template: %w\nContent preview (first 500 chars):\n%s", err, compiledContent[:min(500, len(compiledContent))])
	}

	// Execute the layout template (which is the main/root template)
	// The layout's {{block "content"}} will automatically use the {{define "content"}} from the content template
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute layout template: %w", err)
	}

	return buf.String(), nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderComponent renders a component
func (e *FinEngine) renderComponent(componentName string, content string, data interface{}) string {
	e.mu.RLock()
	component, exists := e.components[componentName]
	e.mu.RUnlock()

	if !exists {
		return fmt.Sprintf("<!-- Component '%s' not found -->", componentName)
	}

	// Process slots in content
	slotRegex := regexp.MustCompile(`@slot\s*\(\s*['"]([^'"]+)['"]\s*\)\s*(.*?)\s*@endslot`)
	slots := make(map[string]string)

	slotMatches := slotRegex.FindAllStringSubmatch(content, -1)
	for _, match := range slotMatches {
		if len(match) >= 3 {
			slotName := match[1]
			slotContent := match[2]
			slots[slotName] = slotContent
		}
	}

	// Replace slots in component template
	componentTemplate := component.Template
	for slotName, slotContent := range slots {
		slotPlaceholder := fmt.Sprintf("{{slot_%s}}", slotName)
		componentTemplate = strings.ReplaceAll(componentTemplate, slotPlaceholder, slotContent)
	}

	return componentTemplate
}

// getCachedTemplate retrieves cached template
func (e *FinEngine) getCachedTemplate(templateName string) *CompiledTemplate {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if cached, exists := e.compiledCache[templateName]; exists {
		// Check if cache is still valid (5 minutes)
		if time.Since(cached.Created) < 5*time.Minute {
			return cached
		}
	}

	return nil
}

// setCachedTemplate caches compiled template
func (e *FinEngine) setCachedTemplate(templateName string, compiled *CompiledTemplate) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.compiledCache[templateName] = compiled
}

// ClearCache clears the template cache
func (e *FinEngine) ClearCache() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.compiledCache = make(map[string]*CompiledTemplate)

	// Clear file cache
	if e.cacheEnabled {
		os.RemoveAll(e.cachePath)
		os.MkdirAll(e.cachePath, 0755)
	}
}

// GetDirectives returns all registered directives
func (e *FinEngine) GetDirectives() map[string]DirectiveFunc {
	e.mu.RLock()
	defer e.mu.RUnlock()

	directives := make(map[string]DirectiveFunc)
	for name, fn := range e.directives {
		directives[name] = fn
	}

	return directives
}

// GetComponents returns all registered components
func (e *FinEngine) GetComponents() map[string]*Component {
	e.mu.RLock()
	defer e.mu.RUnlock()

	components := make(map[string]*Component)
	for name, component := range e.components {
		components[name] = component
	}

	return components
}

// GetLayouts returns all registered layouts
func (e *FinEngine) GetLayouts() map[string]*Layout {
	e.mu.RLock()
	defer e.mu.RUnlock()

	layouts := make(map[string]*Layout)
	for name, layout := range e.layouts {
		layouts[name] = layout
	}

	return layouts
}
