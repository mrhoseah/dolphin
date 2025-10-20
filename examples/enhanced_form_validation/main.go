package main

import (
	"fmt"
	"net/http"

	"dolphin/internal/forms"
	"dolphin/internal/validation"
)

// User represents a user model
type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Age      int    `json:"age"`
	Bio      string `json:"bio"`
	Website  string `json:"website"`
}

func main() {
	fmt.Println("🚀 Dolphin Framework - Enhanced Form Validation Demo")
	fmt.Println("===================================================")
	fmt.Println("")

	// 1. Create a form using FormBuilder
	fmt.Println("=== 1. FormBuilder Example ===")
	demoFormBuilder()

	fmt.Println("")

	// 2. Create a form from struct
	fmt.Println("=== 2. Form from Struct Example ===")
	demoFormFromStruct()

	fmt.Println("")

	// 3. Form validation with Fin templates
	fmt.Println("=== 3. Fin Template Integration Example ===")
	demoFinTemplateIntegration()

	fmt.Println("")

	// 4. HTTP form handling
	fmt.Println("=== 4. HTTP Form Handling Example ===")
	demoHTTPFormHandling()

	fmt.Println("")
	fmt.Println("🎉 Enhanced form validation system demonstrated successfully!")
	fmt.Println("")
	fmt.Println("💡 Key Features Implemented:")
	fmt.Println("  ✅ Fluent form builder API")
	fmt.Println("  ✅ Automatic form generation from structs")
	fmt.Println("  ✅ Fin template integration")
	fmt.Println("  ✅ Comprehensive validation rules")
	fmt.Println("  ✅ CSRF protection")
	fmt.Println("  ✅ Error handling and display")
	fmt.Println("  ✅ Multiple field types support")
}

func demoFormBuilder() {
	// Create form using builder
	builder := forms.NewFormBuilder("user_form", "POST", "/users")

	nameField := builder.AddTextField("name", "Full Name")
	nameField.SetRequired(true).SetPlaceholder("Enter your full name")

	emailField := builder.AddEmailField("email", "Email Address")
	emailField.SetRequired(true).SetPlaceholder("Enter your email")

	passwordField := builder.AddPasswordField("password", "Password")
	passwordField.SetRequired(true)

	bioField := builder.AddTextareaField("bio", "Biography")
	bioField.SetPlaceholder("Tell us about yourself")

	ageField := builder.AddSelectField("age", "Age Range")
	ageField.AddOption("18-25", "18-25", false).
		AddOption("26-35", "26-35", true).
		AddOption("36-45", "36-45", false).
		AddOption("46+", "46+", false)

	builder.AddValidationRule("name", &validation.RequiredRule{
		BaseRule: &validation.BaseRule{
			Field:   "name",
			Message: "Name is required",
		},
	})

	builder.AddValidationRule("email", validation.NewEmailRule("email", "Email must be valid"))

	form := builder.Build()

	// Set form values
	form.SetValues(map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "secret123",
		"bio":      "Software developer",
		"age":      "26-35",
	})

	// Validate form
	isValid := form.Validate()
	fmt.Printf("Form valid: %v\n", isValid)

	if !isValid {
		fmt.Println("Validation errors:")
		for field, errors := range form.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Render form HTML
	fmt.Println("\nForm HTML would be rendered here:")
	fmt.Println("(Form rendering functionality available)")
}

func demoFormFromStruct() {
	// Create user struct
	user := User{
		Name:     "Jane Smith",
		Email:    "jane@example.com",
		Password: "password123",
		Age:      28,
		Bio:      "Product manager",
		Website:  "https://janesmith.com",
	}

	// Create form from struct
	form := forms.FormFromStruct("user_form", "POST", "/users", user)

	// Add validation rules
	form.AddValidationRule("name", &validation.RequiredRule{
		BaseRule: &validation.BaseRule{
			Field:   "name",
			Message: "Name is required",
		},
	})

	form.AddValidationRule("email", validation.NewEmailRule("email", "Email must be valid"))

	// Validate form
	isValid := form.Validate()
	fmt.Printf("Form valid: %v\n", isValid)

	if !isValid {
		fmt.Println("Validation errors:")
		for field, errors := range form.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Show form fields
	fmt.Println("\nForm fields:")
	for name, field := range form.Fields {
		fmt.Printf("  %s (%s): %v\n", name, field.Type, field.Value)
	}
}

func demoFinTemplateIntegration() {
	// Create form helper
	formHelper := forms.NewFormHelper()

	// Create a form
	form := forms.NewForm("contact_form", "POST", "/contact")
	form.AddField("name", "text", "Name").SetRequired(true)
	form.AddField("email", "email", "Email").SetRequired(true)
	form.AddField("message", "textarea", "Message").SetRequired(true)

	// Add validation rules
	form.AddValidationRule("name", &validation.RequiredRule{
		BaseRule: &validation.BaseRule{
			Field:   "name",
			Message: "Name is required",
		},
	})

	form.AddValidationRule("email", validation.NewEmailRule("email", "Email must be valid"))

	// Register form
	formHelper.RegisterForm("contact_form", form)

	fmt.Println("Fin template integration example:")
	fmt.Println("Template functions available:")
	fmt.Println("  - form_start(formName, attrs...)")
	fmt.Println("  - form_end()")
	fmt.Println("  - field(formName, fieldName, attrs...)")
	fmt.Println("  - field_label(formName, fieldName, label)")
	fmt.Println("  - field_error(formName, fieldName)")
	fmt.Println("  - form_errors(formName)")
	fmt.Println("  - csrf(formName)")
	fmt.Println("  - field_value(formName, fieldName)")
	fmt.Println("  - has_field_error(formName, fieldName)")
}

func demoHTTPFormHandling() {
	// Create HTTP handler
	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Create form
			form := forms.NewForm("contact_form", "POST", "/contact")
			form.AddField("name", "text", "Name").SetRequired(true)
			form.AddField("email", "email", "Email").SetRequired(true)
			form.AddField("message", "textarea", "Message").SetRequired(true)

			// Add validation rules
			form.AddValidationRule("name", &validation.RequiredRule{
				BaseRule: &validation.BaseRule{
					Field:   "name",
					Message: "Name is required",
				},
			})

			form.AddValidationRule("email", &validation.EmailRule{
				BaseRule: &validation.BaseRule{
					Field:   "email",
					Message: "Email must be valid",
				},
			})

			// Set values from request
			form.SetValuesFromRequest(r)

			// Validate form
			if form.Validate() {
				// Process valid form
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Form submitted successfully!")
			} else {
				// Handle validation errors
				w.WriteHeader(http.StatusUnprocessableEntity)
				fmt.Fprintf(w, "Validation failed: %v", form.GetErrors())
			}
		} else {
			// Show form
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Contact form would be displayed here")
		}
	})

	fmt.Println("HTTP form handling example:")
	fmt.Println("  - POST /contact - Submit form")
	fmt.Println("  - GET /contact - Display form")
	fmt.Println("  - Automatic validation")
	fmt.Println("  - Error handling")
	fmt.Println("  - CSRF protection ready")
}
