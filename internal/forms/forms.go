package forms

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"

	"github.com/mrhoseah/dolphin/internal/validation"
)

// Form represents a form with validation and rendering capabilities
type Form struct {
	Name       string
	Method     string
	Action     string
	Fields     map[string]*Field
	Errors     map[string][]string
	Data       map[string]interface{}
	Attributes map[string]string
	validator  *validation.Validator
	csrfToken  string
}

// Field represents a form field
type Field struct {
	Name        string
	Type        string
	Label       string
	Value       interface{}
	Placeholder string
	Required    bool
	Attributes  map[string]string
	Options     []Option
	Errors      []string
}

// Option represents a select option
type Option struct {
	Value    string
	Label    string
	Text     string // Alias for Label for compatibility
	Selected bool
}

// NewForm creates a new form
func NewForm(name, method, action string) *Form {
	return &Form{
		Name:      name,
		Method:    method,
		Action:    action,
		Fields:    make(map[string]*Field),
		Errors:    make(map[string][]string),
		Data:      make(map[string]interface{}),
		validator: validation.NewValidator(),
	}
}

// AddField adds a field to the form
func (f *Form) AddField(name, fieldType, label string) *Field {
	field := &Field{
		Name:       name,
		Type:       fieldType,
		Label:      label,
		Attributes: make(map[string]string),
		Options:    make([]Option, 0),
	}
	f.Fields[name] = field
	return field
}

// SetValue sets the value for a field
func (f *Form) SetValue(name string, value interface{}) {
	if field, exists := f.Fields[name]; exists {
		field.Value = value
	}
	f.Data[name] = value
}

// SetValues sets multiple values from a map
func (f *Form) SetValues(values map[string]interface{}) {
	for name, value := range values {
		f.SetValue(name, value)
	}
}

// SetValuesFromRequest sets values from HTTP request
func (f *Form) SetValuesFromRequest(r *http.Request) {
	if err := r.ParseForm(); err != nil {
		return
	}

	for name, values := range r.Form {
		if len(values) > 0 {
			f.SetValue(name, values[0])
		}
	}
}

// AddValidationRule adds a validation rule to a field
func (f *Form) AddValidationRule(fieldName string, rule validation.Rule) {
	f.validator.AddRule(fieldName, rule)
}

// Validate validates the form data
func (f *Form) Validate() bool {
	isValid := f.validator.Validate(f.Data)
	f.Errors = f.validator.GetErrors()

	// Set field errors
	for fieldName, errors := range f.Errors {
		if field, exists := f.Fields[fieldName]; exists {
			field.Errors = errors
		}
	}

	return isValid
}

// HasErrors checks if the form has any errors
func (f *Form) HasErrors() bool {
	return len(f.Errors) > 0
}

// GetError returns the first error for a field
func (f *Form) GetError(fieldName string) string {
	if errors, exists := f.Errors[fieldName]; exists && len(errors) > 0 {
		return errors[0]
	}
	return ""
}

// GetErrors returns all validation errors
func (f *Form) GetErrors() map[string][]string {
	return f.Errors
}

// GetField returns a field by name
func (f *Form) GetField(name string) *Field {
	return f.Fields[name]
}

// SetCSRFToken sets the CSRF token
func (f *Form) SetCSRFToken(token string) {
	f.csrfToken = token
}

// GetCSRFToken returns the CSRF token
func (f *Form) GetCSRFToken() string {
	return f.csrfToken
}

// Field methods for fluent API
func (field *Field) SetValue(value interface{}) *Field {
	field.Value = value
	return field
}

func (field *Field) SetPlaceholder(placeholder string) *Field {
	field.Placeholder = placeholder
	return field
}

func (field *Field) SetRequired(required bool) *Field {
	field.Required = required
	return field
}

func (field *Field) AddAttribute(name, value string) *Field {
	field.Attributes[name] = value
	return field
}

func (field *Field) AddOption(value, label string, selected bool) *Field {
	field.Options = append(field.Options, Option{
		Value:    value,
		Label:    label,
		Selected: selected,
	})
	return field
}

// FormBuilder provides a fluent API for building forms
type FormBuilder struct {
	form *Form
}

// NewFormBuilder creates a new form builder
func NewFormBuilder(name, method, action string) *FormBuilder {
	return &FormBuilder{
		form: NewForm(name, method, action),
	}
}

// AddTextField adds a text field
func (fb *FormBuilder) AddTextField(name, label string) *Field {
	return fb.form.AddField(name, "text", label)
}

// AddEmailField adds an email field
func (fb *FormBuilder) AddEmailField(name, label string) *Field {
	return fb.form.AddField(name, "email", label)
}

// AddPasswordField adds a password field
func (fb *FormBuilder) AddPasswordField(name, label string) *Field {
	return fb.form.AddField(name, "password", label)
}

// AddTextareaField adds a textarea field
func (fb *FormBuilder) AddTextareaField(name, label string) *Field {
	return fb.form.AddField(name, "textarea", label)
}

// AddSelectField adds a select field
func (fb *FormBuilder) AddSelectField(name, label string) *Field {
	return fb.form.AddField(name, "select", label)
}

// AddCheckboxField adds a checkbox field
func (fb *FormBuilder) AddCheckboxField(name, label string) *Field {
	return fb.form.AddField(name, "checkbox", label)
}

// AddValidationRule adds a validation rule
func (fb *FormBuilder) AddValidationRule(fieldName string, rule validation.Rule) *FormBuilder {
	fb.form.AddValidationRule(fieldName, rule)
	return fb
}

// Build returns the built form
func (fb *FormBuilder) Build() *Form {
	return fb.form
}

// Method sets the form method
func (fb *FormBuilder) Method(method string) *FormBuilder {
	fb.form.Method = method
	return fb
}

// Action sets the form action
func (fb *FormBuilder) Action(action string) *FormBuilder {
	fb.form.Action = action
	return fb
}

// Attribute adds an attribute to the form
func (fb *FormBuilder) Attribute(key, value string) *FormBuilder {
	if fb.form.Attributes == nil {
		fb.form.Attributes = make(map[string]string)
	}
	fb.form.Attributes[key] = value
	return fb
}

// CSRFToken sets the CSRF token
func (fb *FormBuilder) CSRFToken(token string) *FormBuilder {
	fb.form.csrfToken = token
	return fb
}

// Text adds a text field
func (fb *FormBuilder) Text(name, label string, value interface{}) *Field {
	field := &Field{
		Name:  name,
		Type:  "text",
		Label: label,
		Value: value,
	}
	fb.form.Fields[name] = field
	return field
}

// Email adds an email field
func (fb *FormBuilder) Email(name, label string, value interface{}) *Field {
	field := &Field{
		Name:  name,
		Type:  "email",
		Label: label,
		Value: value,
	}
	fb.form.Fields[name] = field
	return field
}

// Password adds a password field
func (fb *FormBuilder) Password(name, label string) *Field {
	field := &Field{
		Name:  name,
		Type:  "password",
		Label: label,
	}
	fb.form.Fields[name] = field
	return field
}

// Number adds a number field
func (fb *FormBuilder) Number(name, label string, value interface{}) *Field {
	field := &Field{
		Name:  name,
		Type:  "number",
		Label: label,
		Value: value,
	}
	fb.form.Fields[name] = field
	return field
}

// Textarea adds a textarea field
func (fb *FormBuilder) Textarea(name, label string, value interface{}) *Field {
	field := &Field{
		Name:  name,
		Type:  "textarea",
		Label: label,
		Value: value,
	}
	fb.form.Fields[name] = field
	return field
}

// Select adds a select field
func (fb *FormBuilder) Select(name, label string, options []Option, value interface{}) *Field {
	field := &Field{
		Name:    name,
		Type:    "select",
		Label:   label,
		Value:   value,
		Options: options,
	}
	fb.form.Fields[name] = field
	return field
}

// Checkbox adds a checkbox field
func (fb *FormBuilder) Checkbox(name, label string, checked bool) *Field {
	field := &Field{
		Name:  name,
		Type:  "checkbox",
		Label: label,
		Value: checked,
	}
	fb.form.Fields[name] = field
	return field
}

// Date adds a date field
func (fb *FormBuilder) Date(name, label string, value interface{}) *Field {
	field := &Field{
		Name:  name,
		Type:  "date",
		Label: label,
		Value: value,
	}
	fb.form.Fields[name] = field
	return field
}

// File adds a file field
func (fb *FormBuilder) File(name, label string) *Field {
	field := &Field{
		Name:  name,
		Type:  "file",
		Label: label,
	}
	fb.form.Fields[name] = field
	return field
}

// Hidden adds a hidden field
func (fb *FormBuilder) Hidden(name string, value interface{}) *Field {
	field := &Field{
		Name:  name,
		Type:  "hidden",
		Value: value,
	}
	fb.form.Fields[name] = field
	return field
}

// FormHelper provides template helper functions
type FormHelper struct {
	forms map[string]*Form
}

// NewFormHelper creates a new form helper
func NewFormHelper() *FormHelper {
	return &FormHelper{
		forms: make(map[string]*Form),
	}
}

// RegisterForm registers a form
func (fh *FormHelper) RegisterForm(name string, form *Form) {
	fh.forms[name] = form
}

// GetForm returns a form by name
func (fh *FormHelper) GetForm(name string) *Form {
	return fh.forms[name]
}

// FinTemplateFunctions returns template functions for Fin templates
func FinTemplateFunctions(formHelper *FormHelper) map[string]interface{} {
	return map[string]interface{}{
		"form_start": func(formName string, attrs ...string) template.HTML {
			form := formHelper.GetForm(formName)
			if form == nil {
				return template.HTML("")
			}

			var html strings.Builder
			html.WriteString(fmt.Sprintf(`<form name="%s" method="%s" action="%s"`, form.Name, form.Method, form.Action))

			for _, attr := range attrs {
				html.WriteString(fmt.Sprintf(` %s`, attr))
			}

			html.WriteString(`>`)

			if form.csrfToken != "" {
				html.WriteString(fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, form.csrfToken))
			}

			return template.HTML(html.String())
		},
		"form_end": func() template.HTML {
			return template.HTML("</form>")
		},
		"field": func(formName, fieldName string, attrs ...string) template.HTML {
			form := formHelper.GetForm(formName)
			if form == nil {
				return template.HTML("")
			}

			field := form.GetField(fieldName)
			if field == nil {
				return template.HTML("")
			}

			var html strings.Builder
			html.WriteString(fmt.Sprintf(`<input type="%s" name="%s" id="%s"`, field.Type, field.Name, field.Name))

			if field.Value != nil {
				html.WriteString(fmt.Sprintf(` value="%v"`, field.Value))
			}

			if field.Placeholder != "" {
				html.WriteString(fmt.Sprintf(` placeholder="%s"`, field.Placeholder))
			}

			if field.Required {
				html.WriteString(` required`)
			}

			for _, attr := range attrs {
				html.WriteString(fmt.Sprintf(` %s`, attr))
			}

			if len(field.Errors) > 0 {
				html.WriteString(` class="error"`)
			}

			html.WriteString(`>`)

			return template.HTML(html.String())
		},
		"field_error": func(formName, fieldName string) template.HTML {
			form := formHelper.GetForm(formName)
			if form == nil {
				return template.HTML("")
			}

			field := form.GetField(fieldName)
			if field == nil || len(field.Errors) == 0 {
				return template.HTML("")
			}

			var html strings.Builder
			html.WriteString(`<div class="field-errors">`)
			for _, error := range field.Errors {
				html.WriteString(fmt.Sprintf(`<span class="error">%s</span>`, error))
			}
			html.WriteString(`</div>`)

			return template.HTML(html.String())
		},
		"csrf": func(formName string) template.HTML {
			form := formHelper.GetForm(formName)
			if form == nil || form.csrfToken == "" {
				return template.HTML("")
			}

			return template.HTML(fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, form.csrfToken))
		},
	}
}

// FormFromStruct creates a form from a struct
func FormFromStruct(name, method, action string, data interface{}) *Form {
	form := NewForm(name, method, action)

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := strings.ToLower(field.Name)
		fieldType := "text"

		// Determine field type based on struct field type
		switch value.Kind() {
		case reflect.String:
			fieldType = "text"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldType = "number"
		case reflect.Bool:
			fieldType = "checkbox"
		case reflect.Slice:
			fieldType = "select"
		}

		// Check for email type
		if strings.Contains(strings.ToLower(fieldName), "email") {
			fieldType = "email"
		}

		// Check for password type
		if strings.Contains(strings.ToLower(fieldName), "password") {
			fieldType = "password"
		}

		formField := form.AddField(fieldName, fieldType, field.Name)
		formField.SetValue(value.Interface())
	}

	return form
}

// NewBuilder creates a new form builder for fluent API
func NewBuilder() *FormBuilder {
	return &FormBuilder{
		form: NewForm("", "POST", ""),
	}
}

// Helper functions for individual form elements
func Text(name, value string, attrs map[string]string) string {
	attrsStr := ""
	if attrs != nil {
		for k, v := range attrs {
			attrsStr += fmt.Sprintf(" %s=\"%s\"", k, v)
		}
	}
	return fmt.Sprintf("<input type=\"text\" name=\"%s\" value=\"%s\"%s>", name, value, attrsStr)
}

func Email(name, value string, attrs map[string]string) string {
	attrsStr := ""
	if attrs != nil {
		for k, v := range attrs {
			attrsStr += fmt.Sprintf(" %s=\"%s\"", k, v)
		}
	}
	return fmt.Sprintf("<input type=\"email\" name=\"%s\" value=\"%s\"%s>", name, value, attrsStr)
}

func Submit(text string, attrs map[string]string) string {
	attrsStr := ""
	if attrs != nil {
		for k, v := range attrs {
			attrsStr += fmt.Sprintf(" %s=\"%s\"", k, v)
		}
	}
	return fmt.Sprintf("<button type=\"submit\"%s>%s</button>", attrsStr, text)
}

func Link(text, url string, attrs map[string]string) string {
	attrsStr := ""
	if attrs != nil {
		for k, v := range attrs {
			attrsStr += fmt.Sprintf(" %s=\"%s\"", k, v)
		}
	}
	return fmt.Sprintf("<a href=\"%s\"%s>%s</a>", url, attrsStr, text)
}

func URL(path string, params map[string]string) string {
	if len(params) == 0 {
		return path
	}

	query := ""
	for k, v := range params {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("%s=%s", k, v)
	}
	return fmt.Sprintf("%s?%s", path, query)
}

func Asset(path string) string {
	return fmt.Sprintf("/assets/%s", path)
}

func Image(src, alt string) string {
	return fmt.Sprintf("<img src=\"%s\" alt=\"%s\">", src, alt)
}
