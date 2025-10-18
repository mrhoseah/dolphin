package forms

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

// Form represents a form
type Form struct {
	Method     string
	Action     string
	Attributes map[string]string
	Fields     []Field
	CSRFToken  string
	Errors     map[string][]string
	OldValues  map[string]interface{}
}

// Field represents a form field
type Field struct {
	Type        string
	Name        string
	Label       string
	Value       interface{}
	Placeholder string
	Required    bool
	Disabled    bool
	Readonly    bool
	Attributes  map[string]string
	Options     []Option
	Help        string
	Error       string
}

// Option represents a select option
type Option struct {
	Value    interface{}
	Text     string
	Selected bool
	Disabled bool
}

// Builder helps build forms
type Builder struct {
	form *Form
}

// NewBuilder creates a new form builder
func NewBuilder() *Builder {
	return &Builder{
		form: &Form{
			Method:     "POST",
			Attributes: make(map[string]string),
			Fields:     make([]Field, 0),
			Errors:     make(map[string][]string),
			OldValues:  make(map[string]interface{}),
		},
	}
}

// Method sets the form method
func (fb *Builder) Method(method string) *Builder {
	fb.form.Method = method
	return fb
}

// Action sets the form action
func (fb *Builder) Action(action string) *Builder {
	fb.form.Action = action
	return fb
}

// Attribute adds an attribute
func (fb *Builder) Attribute(key, value string) *Builder {
	fb.form.Attributes[key] = value
	return fb
}

// CSRFToken sets the CSRF token
func (fb *Builder) CSRFToken(token string) *Builder {
	fb.form.CSRFToken = token
	return fb
}

// Errors sets form errors
func (fb *Builder) Errors(errors map[string][]string) *Builder {
	fb.form.Errors = errors
	return fb
}

// OldValues sets old values
func (fb *Builder) OldValues(values map[string]interface{}) *Builder {
	fb.form.OldValues = values
	return fb
}

// Text creates a text input field
func (fb *Builder) Text(name, label string, value interface{}) *Builder {
	field := Field{
		Type:       "text",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Email creates an email input field
func (fb *Builder) Email(name, label string, value interface{}) *Builder {
	field := Field{
		Type:       "email",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Password creates a password input field
func (fb *Builder) Password(name, label string) *Builder {
	field := Field{
		Type:       "password",
		Name:       name,
		Label:      label,
		Value:      "",
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Number creates a number input field
func (fb *Builder) Number(name, label string, value interface{}) *Builder {
	field := Field{
		Type:       "number",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Textarea creates a textarea field
func (fb *Builder) Textarea(name, label string, value interface{}) *Builder {
	field := Field{
		Type:       "textarea",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Select creates a select field
func (fb *Builder) Select(name, label string, options []Option, value interface{}) *Builder {
	field := Field{
		Type:       "select",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Options:    options,
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Checkbox creates a checkbox field
func (fb *Builder) Checkbox(name, label string, checked bool) *Builder {
	field := Field{
		Type:       "checkbox",
		Name:       name,
		Label:      label,
		Value:      checked,
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Radio creates a radio field
func (fb *Builder) Radio(name, label string, options []Option, value interface{}) *Builder {
	field := Field{
		Type:       "radio",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Options:    options,
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// File creates a file input field
func (fb *Builder) File(name, label string) *Builder {
	field := Field{
		Type:       "file",
		Name:       name,
		Label:      label,
		Value:      "",
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Date creates a date input field
func (fb *Builder) Date(name, label string, value interface{}) *Builder {
	field := Field{
		Type:       "date",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// DateTime creates a datetime input field
func (fb *Builder) DateTime(name, label string, value interface{}) *Builder {
	field := Field{
		Type:       "datetime-local",
		Name:       name,
		Label:      label,
		Value:      fb.getFieldValue(name, value),
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Hidden creates a hidden field
func (fb *Builder) Hidden(name string, value interface{}) *Builder {
	field := Field{
		Type:       "hidden",
		Name:       name,
		Value:      value,
		Attributes: make(map[string]string),
	}
	fb.form.Fields = append(fb.form.Fields, field)
	return fb
}

// Required marks the last field as required
func (fb *Builder) Required() *Builder {
	if len(fb.form.Fields) > 0 {
		fb.form.Fields[len(fb.form.Fields)-1].Required = true
	}
	return fb
}

// Placeholder sets placeholder for the last field
func (fb *Builder) Placeholder(placeholder string) *Builder {
	if len(fb.form.Fields) > 0 {
		fb.form.Fields[len(fb.form.Fields)-1].Placeholder = placeholder
	}
	return fb
}

// Disabled marks the last field as disabled
func (fb *Builder) Disabled() *Builder {
	if len(fb.form.Fields) > 0 {
		fb.form.Fields[len(fb.form.Fields)-1].Disabled = true
	}
	return fb
}

// Readonly marks the last field as readonly
func (fb *Builder) Readonly() *Builder {
	if len(fb.form.Fields) > 0 {
		fb.form.Fields[len(fb.form.Fields)-1].Readonly = true
	}
	return fb
}

// Help sets help text for the last field
func (fb *Builder) Help(help string) *Builder {
	if len(fb.form.Fields) > 0 {
		fb.form.Fields[len(fb.form.Fields)-1].Help = help
	}
	return fb
}

// Attribute adds an attribute to the last field
func (fb *Builder) FieldAttribute(key, value string) *Builder {
	if len(fb.form.Fields) > 0 {
		fb.form.Fields[len(fb.form.Fields)-1].Attributes[key] = value
	}
	return fb
}

// getFieldValue gets the field value from old values or provided value
func (fb *Builder) getFieldValue(name string, value interface{}) interface{} {
	if oldValue, exists := fb.form.OldValues[name]; exists {
		return oldValue
	}
	return value
}

// Build builds the form
func (fb *Builder) Build() *Form {
	return fb.form
}

// Render renders the form as HTML
func (f *Form) Render() template.HTML {
	var html strings.Builder

	// Form opening tag
	html.WriteString(fmt.Sprintf(`<form method="%s"`, f.Method))
	if f.Action != "" {
		html.WriteString(fmt.Sprintf(` action="%s"`, f.Action))
	}

	// Add attributes
	for key, value := range f.Attributes {
		html.WriteString(fmt.Sprintf(` %s="%s"`, key, value))
	}

	html.WriteString(">\n")

	// CSRF token
	if f.CSRFToken != "" {
		html.WriteString(fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, f.CSRFToken))
	}

	// Render fields
	for _, field := range f.Fields {
		html.WriteString(f.renderField(field))
	}

	// Form closing tag
	html.WriteString("</form>")

	return template.HTML(html.String())
}

// renderField renders a single field
func (f *Form) renderField(field Field) string {
	var html strings.Builder

	switch field.Type {
	case "text", "email", "password", "number", "date", "datetime-local":
		html.WriteString(f.renderInput(field))
	case "textarea":
		html.WriteString(f.renderTextarea(field))
	case "select":
		html.WriteString(f.renderSelect(field))
	case "checkbox":
		html.WriteString(f.renderCheckbox(field))
	case "radio":
		html.WriteString(f.renderRadio(field))
	case "file":
		html.WriteString(f.renderFile(field))
	case "hidden":
		html.WriteString(f.renderHidden(field))
	}

	return html.String()
}

// renderInput renders an input field
func (f *Form) renderInput(field Field) string {
	var html strings.Builder

	// Label
	if field.Label != "" {
		html.WriteString(fmt.Sprintf(`<label for="%s">%s`, field.Name, field.Label))
		if field.Required {
			html.WriteString(" *")
		}
		html.WriteString("</label>\n")
	}

	// Input
	html.WriteString(fmt.Sprintf(`<input type="%s" name="%s" id="%s"`, field.Type, field.Name, field.Name))

	// Value
	if field.Value != nil && field.Value != "" {
		html.WriteString(fmt.Sprintf(` value="%v"`, field.Value))
	}

	// Placeholder
	if field.Placeholder != "" {
		html.WriteString(fmt.Sprintf(` placeholder="%s"`, field.Placeholder))
	}

	// Required
	if field.Required {
		html.WriteString(` required`)
	}

	// Disabled
	if field.Disabled {
		html.WriteString(` disabled`)
	}

	// Readonly
	if field.Readonly {
		html.WriteString(` readonly`)
	}

	// Attributes
	for key, value := range field.Attributes {
		html.WriteString(fmt.Sprintf(` %s="%s"`, key, value))
	}

	html.WriteString(">\n")

	// Help text
	if field.Help != "" {
		html.WriteString(fmt.Sprintf(`<small class="form-text text-muted">%s</small>`, field.Help))
	}

	// Error
	if field.Error != "" {
		html.WriteString(fmt.Sprintf(`<div class="invalid-feedback">%s</div>`, field.Error))
	}

	return html.String()
}

// renderTextarea renders a textarea field
func (f *Form) renderTextarea(field Field) string {
	var html strings.Builder

	// Label
	if field.Label != "" {
		html.WriteString(fmt.Sprintf(`<label for="%s">%s`, field.Name, field.Label))
		if field.Required {
			html.WriteString(" *")
		}
		html.WriteString("</label>\n")
	}

	// Textarea
	html.WriteString(fmt.Sprintf(`<textarea name="%s" id="%s"`, field.Name, field.Name))

	// Required
	if field.Required {
		html.WriteString(` required`)
	}

	// Disabled
	if field.Disabled {
		html.WriteString(` disabled`)
	}

	// Readonly
	if field.Readonly {
		html.WriteString(` readonly`)
	}

	// Attributes
	for key, value := range field.Attributes {
		html.WriteString(fmt.Sprintf(` %s="%s"`, key, value))
	}

	html.WriteString(">")

	// Value
	if field.Value != nil {
		html.WriteString(fmt.Sprintf("%v", field.Value))
	}

	html.WriteString("</textarea>\n")

	// Help text
	if field.Help != "" {
		html.WriteString(fmt.Sprintf(`<small class="form-text text-muted">%s</small>`, field.Help))
	}

	// Error
	if field.Error != "" {
		html.WriteString(fmt.Sprintf(`<div class="invalid-feedback">%s</div>`, field.Error))
	}

	return html.String()
}

// renderSelect renders a select field
func (f *Form) renderSelect(field Field) string {
	var html strings.Builder

	// Label
	if field.Label != "" {
		html.WriteString(fmt.Sprintf(`<label for="%s">%s`, field.Name, field.Label))
		if field.Required {
			html.WriteString(" *")
		}
		html.WriteString("</label>\n")
	}

	// Select
	html.WriteString(fmt.Sprintf(`<select name="%s" id="%s"`, field.Name, field.Name))

	// Required
	if field.Required {
		html.WriteString(` required`)
	}

	// Disabled
	if field.Disabled {
		html.WriteString(` disabled`)
	}

	// Attributes
	for key, value := range field.Attributes {
		html.WriteString(fmt.Sprintf(` %s="%s"`, key, value))
	}

	html.WriteString(">\n")

	// Options
	for _, option := range field.Options {
		html.WriteString(fmt.Sprintf(`<option value="%v"`, option.Value))
		if option.Selected {
			html.WriteString(` selected`)
		}
		if option.Disabled {
			html.WriteString(` disabled`)
		}
		html.WriteString(fmt.Sprintf(`>%s</option>`, option.Text))
	}

	html.WriteString("</select>\n")

	// Help text
	if field.Help != "" {
		html.WriteString(fmt.Sprintf(`<small class="form-text text-muted">%s</small>`, field.Help))
	}

	// Error
	if field.Error != "" {
		html.WriteString(fmt.Sprintf(`<div class="invalid-feedback">%s</div>`, field.Error))
	}

	return html.String()
}

// renderCheckbox renders a checkbox field
func (f *Form) renderCheckbox(field Field) string {
	var html strings.Builder

	html.WriteString(fmt.Sprintf(`<div class="form-check">`))
	html.WriteString(fmt.Sprintf(`<input type="checkbox" name="%s" id="%s"`, field.Name, field.Name))

	if field.Value == true {
		html.WriteString(` checked`)
	}

	if field.Disabled {
		html.WriteString(` disabled`)
	}

	html.WriteString(fmt.Sprintf(` class="form-check-input">`))
	html.WriteString(fmt.Sprintf(`<label class="form-check-label" for="%s">%s</label>`, field.Name, field.Label))
	html.WriteString(`</div>`)

	return html.String()
}

// renderRadio renders radio fields
func (f *Form) renderRadio(field Field) string {
	var html strings.Builder

	// Label
	if field.Label != "" {
		html.WriteString(fmt.Sprintf(`<label>%s`, field.Label))
		if field.Required {
			html.WriteString(" *")
		}
		html.WriteString("</label>\n")
	}

	// Radio options
	for i, option := range field.Options {
		html.WriteString(fmt.Sprintf(`<div class="form-check">`))
		html.WriteString(fmt.Sprintf(`<input type="radio" name="%s" id="%s_%d" value="%v"`, field.Name, field.Name, i, option.Value))

		if option.Selected || fmt.Sprintf("%v", field.Value) == fmt.Sprintf("%v", option.Value) {
			html.WriteString(` checked`)
		}

		if field.Disabled || option.Disabled {
			html.WriteString(` disabled`)
		}

		html.WriteString(fmt.Sprintf(` class="form-check-input">`))
		html.WriteString(fmt.Sprintf(`<label class="form-check-label" for="%s_%d">%s</label>`, field.Name, i, option.Text))
		html.WriteString(`</div>`)
	}

	return html.String()
}

// renderFile renders a file input field
func (f *Form) renderFile(field Field) string {
	var html strings.Builder

	// Label
	if field.Label != "" {
		html.WriteString(fmt.Sprintf(`<label for="%s">%s`, field.Name, field.Label))
		if field.Required {
			html.WriteString(" *")
		}
		html.WriteString("</label>\n")
	}

	// File input
	html.WriteString(fmt.Sprintf(`<input type="file" name="%s" id="%s"`, field.Name, field.Name))

	// Required
	if field.Required {
		html.WriteString(` required`)
	}

	// Disabled
	if field.Disabled {
		html.WriteString(` disabled`)
	}

	// Attributes
	for key, value := range field.Attributes {
		html.WriteString(fmt.Sprintf(` %s="%s"`, key, value))
	}

	html.WriteString(">\n")

	// Help text
	if field.Help != "" {
		html.WriteString(fmt.Sprintf(`<small class="form-text text-muted">%s</small>`, field.Help))
	}

	// Error
	if field.Error != "" {
		html.WriteString(fmt.Sprintf(`<div class="invalid-feedback">%s</div>`, field.Error))
	}

	return html.String()
}

// renderHidden renders a hidden field
func (f *Form) renderHidden(field Field) string {
	return fmt.Sprintf(`<input type="hidden" name="%s" value="%v">`, field.Name, field.Value)
}

// Helper functions

// OpenForm opens a form
func OpenForm(method, action string, attributes map[string]string) *Builder {
	builder := NewBuilder()
	builder.Method(method).Action(action)

	if attributes != nil {
		for key, value := range attributes {
			builder.Attribute(key, value)
		}
	}

	return builder
}

// CloseForm closes a form
func CloseForm() string {
	return "</form>"
}

// CSRF generates a CSRF token field
func CSRF(token string) string {
	return fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, token)
}

// Label generates a label
func Label(name, text string, attributes map[string]string) string {
	html := fmt.Sprintf(`<label for="%s"`, name)

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += fmt.Sprintf(`>%s</label>`, text)
	return html
}

// Input generates an input field
func Input(inputType, name string, value interface{}, attributes map[string]string) string {
	html := fmt.Sprintf(`<input type="%s" name="%s"`, inputType, name)

	if value != nil {
		html += fmt.Sprintf(` value="%v"`, value)
	}

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += ">"
	return html
}

// Text generates a text input
func Text(name string, value interface{}, attributes map[string]string) string {
	return Input("text", name, value, attributes)
}

// Email generates an email input
func Email(name string, value interface{}, attributes map[string]string) string {
	return Input("email", name, value, attributes)
}

// Password generates a password input
func Password(name string, attributes map[string]string) string {
	return Input("password", name, nil, attributes)
}

// Number generates a number input
func Number(name string, value interface{}, attributes map[string]string) string {
	return Input("number", name, value, attributes)
}

// File generates a file input
func File(name string, attributes map[string]string) string {
	return Input("file", name, nil, attributes)
}

// Checkbox generates a checkbox
func Checkbox(name string, value interface{}, checked bool, attributes map[string]string) string {
	html := fmt.Sprintf(`<input type="checkbox" name="%s" value="%v"`, name, value)

	if checked {
		html += ` checked`
	}

	if attributes != nil {
		for key, val := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, val)
		}
	}

	html += ">"
	return html
}

// Radio generates a radio input
func Radio(name string, value interface{}, checked bool, attributes map[string]string) string {
	html := fmt.Sprintf(`<input type="radio" name="%s" value="%v"`, name, value)

	if checked {
		html += ` checked`
	}

	if attributes != nil {
		for key, val := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, val)
		}
	}

	html += ">"
	return html
}

// Select generates a select field
func Select(name string, options []Option, selectedValue interface{}, attributes map[string]string) string {
	html := fmt.Sprintf(`<select name="%s"`, name)

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += ">\n"

	for _, option := range options {
		html += fmt.Sprintf(`<option value="%v"`, option.Value)

		if option.Selected || fmt.Sprintf("%v", selectedValue) == fmt.Sprintf("%v", option.Value) {
			html += ` selected`
		}

		if option.Disabled {
			html += ` disabled`
		}

		html += fmt.Sprintf(`>%s</option>`, option.Text)
	}

	html += "</select>"
	return html
}

// Textarea generates a textarea
func Textarea(name string, value interface{}, attributes map[string]string) string {
	html := fmt.Sprintf(`<textarea name="%s"`, name)

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += ">"

	if value != nil {
		html += fmt.Sprintf("%v", value)
	}

	html += "</textarea>"
	return html
}

// Submit generates a submit button
func Submit(text string, attributes map[string]string) string {
	html := fmt.Sprintf(`<button type="submit"`)

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += fmt.Sprintf(`>%s</button>`, text)
	return html
}

// Button generates a button
func Button(text string, attributes map[string]string) string {
	html := fmt.Sprintf(`<button`)

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += fmt.Sprintf(`>%s</button>`, text)
	return html
}

// URL helpers

// URL generates a URL
func URL(path string, params map[string]interface{}) string {
	if params == nil || len(params) == 0 {
		return path
	}

	values := url.Values{}
	for key, value := range params {
		values.Add(key, fmt.Sprintf("%v", value))
	}

	return path + "?" + values.Encode()
}

// Route generates a route URL
func Route(name string, params map[string]interface{}) string {
	// In a real implementation, this would look up the route
	return URL("/"+name, params)
}

// Asset generates an asset URL
func Asset(path string) string {
	return "/assets/" + path
}

// Image generates an image URL
func Image(path string, attributes map[string]string) string {
	html := fmt.Sprintf(`<img src="%s"`, path)

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += ">"
	return html
}

// Link generates a link
func Link(text, url string, attributes map[string]string) string {
	html := fmt.Sprintf(`<a href="%s"`, url)

	if attributes != nil {
		for key, value := range attributes {
			html += fmt.Sprintf(` %s="%s"`, key, value)
		}
	}

	html += fmt.Sprintf(`>%s</a>`, text)
	return html
}
