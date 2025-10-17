package template

import (
	"fmt"
	"html/template"
	"strings"
)

// Component represents a template component
type Component struct {
	Name        string                 `json:"name"`
	Content     string                 `json:"content"`
	Compiled    *template.Template     `json:"-"`
	Props       map[string]interface{} `json:"props"`
	Slots       map[string]string      `json:"slots"`
	Events      map[string]string      `json:"events"`
	Styles      string                 `json:"styles"`
	Scripts     string                 `json:"scripts"`
	LastModified time.Time             `json:"last_modified"`
}

// ComponentManager manages template components
type ComponentManager struct {
	components map[string]*Component
	engine     *Engine
}

// NewComponentManager creates a new component manager
func NewComponentManager(engine *Engine) *ComponentManager {
	return &ComponentManager{
		components: make(map[string]*Component),
		engine:     engine,
	}
}

// ParseComponent parses a component template
func (cm *ComponentManager) ParseComponent(name, content string) (*Component, error) {
	component := &Component{
		Name:    name,
		Content: content,
		Props:   make(map[string]interface{}),
		Slots:   make(map[string]string),
		Events:  make(map[string]string),
	}
	
	// Parse component content
	if err := cm.parseComponentContent(component); err != nil {
		return nil, err
	}
	
	// Compile component
	if err := cm.compileComponent(component); err != nil {
		return nil, err
	}
	
	return component, nil
}

// parseComponentContent parses component content for props, slots, and events
func (cm *ComponentManager) parseComponentContent(component *Component) error {
	lines := strings.Split(component.Content, "\n")
	var currentSlot string
	var slotContent []string
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Parse props directive
		if strings.HasPrefix(line, "{{props") {
			props := cm.extractDirective(line, "props")
			component.Props[props] = nil
			continue
		}
		
		// Parse slot start
		if strings.HasPrefix(line, "{{slot") {
			// Save previous slot
			if currentSlot != "" {
				component.Slots[currentSlot] = strings.Join(slotContent, "\n")
			}
			
			// Start new slot
			slotName := cm.extractDirective(line, "slot")
			currentSlot = slotName
			slotContent = []string{}
			continue
		}
		
		// Parse slot end
		if strings.HasPrefix(line, "{{endslot") {
			if currentSlot != "" {
				component.Slots[currentSlot] = strings.Join(slotContent, "\n")
				currentSlot = ""
				slotContent = []string{}
			}
			continue
		}
		
		// Parse event directive
		if strings.HasPrefix(line, "{{event") {
			event := cm.extractDirective(line, "event")
			component.Events[event] = ""
			continue
		}
		
		// Parse style directive
		if strings.HasPrefix(line, "{{style") {
			style := cm.extractDirective(line, "style")
			component.Styles = style
			continue
		}
		
		// Parse script directive
		if strings.HasPrefix(line, "{{script") {
			script := cm.extractDirective(line, "script")
			component.Scripts = script
			continue
		}
		
		// Add line to current slot or main content
		if currentSlot != "" {
			slotContent = append(slotContent, lines[i])
		}
	}
	
	// Save last slot
	if currentSlot != "" {
		component.Slots[currentSlot] = strings.Join(slotContent, "\n")
	}
	
	return nil
}

// extractDirective extracts a directive value from a template line
func (cm *ComponentManager) extractDirective(line, directive string) string {
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

// compileComponent compiles a component template
func (cm *ComponentManager) compileComponent(component *Component) error {
	// Create template with helpers
	funcMap := template.FuncMap{}
	
	// Add helper functions
	if cm.engine.config.EnableHelpers {
		for name, helper := range cm.engine.helpers {
			funcMap[name] = helper
		}
	}
	
	// Add component-specific functions
	funcMap["slot"] = cm.slotHelper(component)
	funcMap["prop"] = cm.propHelper(component)
	funcMap["event"] = cm.eventHelper(component)
	funcMap["style"] = cm.styleHelper(component)
	funcMap["script"] = cm.scriptHelper(component)
	
	// Compile template
	compiled, err := template.New(component.Name).Funcs(funcMap).Parse(component.Content)
	if err != nil {
		return err
	}
	
	component.Compiled = compiled
	return nil
}

// slotHelper returns a helper function for rendering slots
func (cm *ComponentManager) slotHelper(component *Component) func(string) string {
	return func(slotName string) string {
		if slot, exists := component.Slots[slotName]; exists {
			return slot
		}
		return ""
	}
}

// propHelper returns a helper function for accessing props
func (cm *ComponentManager) propHelper(component *Component) func(string, ...interface{}) interface{} {
	return func(propName string, defaultValue ...interface{}) interface{} {
		if prop, exists := component.Props[propName]; exists {
			return prop
		}
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
}

// eventHelper returns a helper function for handling events
func (cm *ComponentManager) eventHelper(component *Component) func(string, string) string {
	return func(eventName, handler string) string {
		if _, exists := component.Events[eventName]; exists {
			return fmt.Sprintf(`on%s="%s"`, eventName, handler)
		}
		return ""
	}
}

// styleHelper returns a helper function for rendering styles
func (cm *ComponentManager) styleHelper(component *Component) func() string {
	return func() string {
		if component.Styles != "" {
			return fmt.Sprintf(`<style>%s</style>`, component.Styles)
		}
		return ""
	}
}

// scriptHelper returns a helper function for rendering scripts
func (cm *ComponentManager) scriptHelper(component *Component) func() string {
	return func() string {
		if component.Scripts != "" {
			return fmt.Sprintf(`<script>%s</script>`, component.Scripts)
		}
		return ""
	}
}

// RenderComponent renders a component with data
func (cm *ComponentManager) RenderComponent(componentName string, data TemplateData) (string, error) {
	component, exists := cm.components[componentName]
	if !exists {
		return "", fmt.Errorf("component %s not found", componentName)
	}
	
	// Render component
	var buf strings.Builder
	if err := component.Compiled.Execute(&buf, data); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// RegisterComponent registers a component
func (cm *ComponentManager) RegisterComponent(component *Component) {
	cm.components[component.Name] = component
}

// GetComponent returns a component by name
func (cm *ComponentManager) GetComponent(name string) (*Component, bool) {
	component, exists := cm.components[name]
	return component, exists
}

// GetAllComponents returns all components
func (cm *ComponentManager) GetAllComponents() map[string]*Component {
	return cm.components
}

// ComponentBuilder provides a fluent interface for building components
type ComponentBuilder struct {
	component *Component
	cm        *ComponentManager
}

// NewComponentBuilder creates a new component builder
func NewComponentBuilder(name string, cm *ComponentManager) *ComponentBuilder {
	return &ComponentBuilder{
		component: &Component{
			Name:   name,
			Props:  make(map[string]interface{}),
			Slots:  make(map[string]string),
			Events: make(map[string]string),
		},
		cm: cm,
	}
}

// Content sets the component content
func (cb *ComponentBuilder) Content(content string) *ComponentBuilder {
	cb.component.Content = content
	return cb
}

// Prop adds a prop to the component
func (cb *ComponentBuilder) Prop(name string, defaultValue interface{}) *ComponentBuilder {
	cb.component.Props[name] = defaultValue
	return cb
}

// Slot adds a slot to the component
func (cb *ComponentBuilder) Slot(name, content string) *ComponentBuilder {
	cb.component.Slots[name] = content
	return cb
}

// Event adds an event to the component
func (cb *ComponentBuilder) Event(name, handler string) *ComponentBuilder {
	cb.component.Events[name] = handler
	return cb
}

// Style sets the component styles
func (cb *ComponentBuilder) Style(styles string) *ComponentBuilder {
	cb.component.Styles = styles
	return cb
}

// Script sets the component scripts
func (cb *ComponentBuilder) Script(scripts string) *ComponentBuilder {
	cb.component.Scripts = scripts
	return cb
}

// Build builds the component
func (cb *ComponentBuilder) Build() (*Component, error) {
	// Compile component
	if err := cb.cm.compileComponent(cb.component); err != nil {
		return nil, err
	}
	
	// Register component
	cb.cm.RegisterComponent(cb.component)
	
	return cb.component, nil
}

// ComponentHelper provides helper functions for components
type ComponentHelper struct {
	engine *Engine
}

// NewComponentHelper creates a new component helper
func NewComponentHelper(engine *Engine) *ComponentHelper {
	return &ComponentHelper{
		engine: engine,
	}
}

// RenderSlot renders a slot with fallback
func (ch *ComponentHelper) RenderSlot(slotName, fallback string) string {
	// This would typically be called from within a template
	// For now, return the fallback
	return fallback
}

// RenderProp renders a prop with fallback
func (ch *ComponentHelper) RenderProp(propName string, fallback interface{}) interface{} {
	// This would typically be called from within a template
	// For now, return the fallback
	return fallback
}

// RenderEvent renders an event handler
func (ch *ComponentHelper) RenderEvent(eventName, handler string) string {
	// This would typically be called from within a template
	// For now, return empty string
	return ""
}

// ComponentExists checks if a component exists
func (ch *ComponentHelper) ComponentExists(name string) bool {
	_, exists := ch.engine.components[name]
	return exists
}

// SlotExists checks if a slot exists
func (ch *ComponentHelper) SlotExists(slotName string) bool {
	// This would typically be called from within a template
	// For now, return false
	return false
}

// PropExists checks if a prop exists
func (ch *ComponentHelper) PropExists(propName string) bool {
	// This would typically be called from within a template
	// For now, return false
	return false
}
