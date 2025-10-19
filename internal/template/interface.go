package template

import (
	"io"
)

// FinTemplateEngine defines the interface for the Fin template engine
type FinTemplateEngine interface {
	// Render renders a Fin template with data
	Render(templateName string, data interface{}) (string, error)

	// RenderToWriter renders a Fin template to a writer
	RenderToWriter(w io.Writer, templateName string, data interface{}) error

	// RegisterDirective registers a custom Fin template directive
	RegisterDirective(name string, fn DirectiveFunc)

	// RegisterComponent registers a reusable Fin component
	RegisterComponent(name string, template string)

	// RegisterLayout registers a Fin template layout
	RegisterLayout(name string, template string)
}

// Ensure FinEngine implements FinTemplateEngine
var _ FinTemplateEngine = (*FinEngine)(nil)
