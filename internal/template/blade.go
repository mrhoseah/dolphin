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

// Engine represents the template engine
type Engine struct {
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
		ViewsPath:    "ui/views",
		CachePath:    "storage/cache/views",
		CacheEnabled: true,
		DebugMode:    false,
		Extensions:   []string{".blade.go", ".go.html"},
	}
}

// NewEngine creates a new template engine
func NewEngine(config *Config) *Engine {
	if config == nil {
		config = DefaultConfig()
	}

	engine := &Engine{
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

// registerDefaultDirectives registers built-in template directives
func (e *Engine) registerDefaultDirectives() {
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
		return e.renderComponent(componentName, content, data)
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
}

// RegisterDirective registers a custom template directive
func (e *Engine) RegisterDirective(name string, fn DirectiveFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.directives[name] = fn
}

// RegisterComponent registers a reusable component
func (e *Engine) RegisterComponent(name string, template string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.components[name] = &Component{
		Name:     name,
		Template: template,
		Slots:    make(map[string]string),
		Props:    make(map[string]interface{}),
	}
}

// RegisterLayout registers a template layout
func (e *Engine) RegisterLayout(name string, template string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.layouts[name] = &Layout{
		Name:     name,
		Template: template,
		Sections: make(map[string]string),
	}
}

// Render renders a template with data
func (e *Engine) Render(templateName string, data interface{}) (string, error) {
	// Get template content
	content, err := e.getTemplateContent(templateName)
	if err != nil {
		return "", err
	}

	// Compile template
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

// RenderToWriter renders a template to a writer
func (e *Engine) RenderToWriter(w io.Writer, templateName string, data interface{}) error {
	content, err := e.Render(templateName, data)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(content))
	return err
}

// getTemplateContent retrieves template content from file or cache
func (e *Engine) getTemplateContent(templateName string) (string, error) {
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
func (e *Engine) findTemplateFile(templateName string) string {
	for _, ext := range e.extensions {
		filePath := filepath.Join(e.viewsPath, templateName+ext)
		if _, err := os.Stat(filePath); err == nil {
			return filePath
		}
	}
	return ""
}

// compileTemplate compiles template content with directives
func (e *Engine) compileTemplate(content string, data interface{}) (string, error) {
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

	// Convert Blade syntax to Go template syntax
	compiled = e.convertBladeToGo(compiled)

	return compiled, nil
}

// processExtendsDirective processes @extends directive
func (e *Engine) processExtendsDirective(content string, data interface{}) string {
	extendsRegex := regexp.MustCompile(`@extends\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	matches := extendsRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		layoutName := match[1]
		// Remove @extends line
		content = strings.Replace(content, match[0], "", 1)
		// This will be handled in renderWithLayout
	}

	return content
}

// processSectionDirectives processes @section and @yield directives
func (e *Engine) processSectionDirectives(content string, data interface{}) string {
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
func (e *Engine) processComponentDirectives(content string, data interface{}) string {
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
func (e *Engine) processDirectives(content string, data interface{}) string {
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

// convertBladeToGo converts Blade syntax to Go template syntax
func (e *Engine) convertBladeToGo(content string) string {
	// Convert {{ $variable }} to {{.Variable}}
	content = regexp.MustCompile(`\{\{\s*\$(\w+)\s*\}\}`).ReplaceAllString(content, "{{.$1}}")

	// Convert {{ $variable.property }} to {{.Variable.Property}}
	content = regexp.MustCompile(`\{\{\s*\$(\w+)\.(\w+)\s*\}\}`).ReplaceAllString(content, "{{.$1.$2}}")

	// Convert {{ $variable['key'] }} to {{index .Variable "key"}}
	content = regexp.MustCompile(`\{\{\s*\$(\w+)\['([^']+)'\]\s*\}\}`).ReplaceAllString(content, "{{index .$1 \"$2\"}}")

	return content
}

// parseDirectiveArgs parses directive arguments
func (e *Engine) parseDirectiveArgs(argsStr string) []string {
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
func (e *Engine) renderWithLayout(layoutName string, content string, data interface{}) (string, error) {
	e.mu.RLock()
	layout, exists := e.layouts[layoutName]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("layout '%s' not found", layoutName)
	}

	// Compile layout template
	compiledLayout, err := e.compileTemplate(layout.Template, data)
	if err != nil {
		return "", err
	}

	// Compile content template
	compiledContent, err := e.compileTemplate(content, data)
	if err != nil {
		return "", err
	}

	// Combine layout and content
	combined := compiledLayout + "\n" + compiledContent

	// Execute combined template
	var buf bytes.Buffer
	tmpl, err := template.New(layoutName).Parse(combined)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderComponent renders a component
func (e *Engine) renderComponent(componentName string, content string, data interface{}) string {
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
func (e *Engine) getCachedTemplate(templateName string) *CompiledTemplate {
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
func (e *Engine) setCachedTemplate(templateName string, compiled *CompiledTemplate) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.compiledCache[templateName] = compiled
}

// ClearCache clears the template cache
func (e *Engine) ClearCache() {
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
func (e *Engine) GetDirectives() map[string]DirectiveFunc {
	e.mu.RLock()
	defer e.mu.RUnlock()

	directives := make(map[string]DirectiveFunc)
	for name, fn := range e.directives {
		directives[name] = fn
	}

	return directives
}

// GetComponents returns all registered components
func (e *Engine) GetComponents() map[string]*Component {
	e.mu.RLock()
	defer e.mu.RUnlock()

	components := make(map[string]*Component)
	for name, component := range e.components {
		components[name] = component
	}

	return components
}

// GetLayouts returns all registered layouts
func (e *Engine) GetLayouts() map[string]*Layout {
	e.mu.RLock()
	defer e.mu.RUnlock()

	layouts := make(map[string]*Layout)
	for name, layout := range e.layouts {
		layouts[name] = layout
	}

	return layouts
}
