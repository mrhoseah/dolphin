package main

import (
	"fmt"

	"dolphin/internal/validation"
)

// User represents a user with validation
type User struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Age             int    `json:"age"`
	Bio             string `json:"bio"`
	Website         string `json:"website"`
}

// Post represents a blog post with validation
type Post struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
	IsPublished bool     `json:"is_published"`
	Category    string   `json:"category"`
}

func main() {
	fmt.Println("🚀 Dolphin Framework - Validation Demo")
	fmt.Println("=====================================")
	fmt.Println("")

	// Example 1: Basic Validation
	fmt.Println("=== Example 1: Basic Validation ===")

	// Create validator
	validator := validation.NewValidator()

	// Test data
	data := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      25,
		"password": "secret123",
		"website":  "https://johndoe.com",
	}

	// Add validation rules
	validator.AddRule("name", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "name", Message: "Name is required"}})
	validator.AddRule("name", validation.NewStringRule("name", "Name must be a string", 0, 0))
	validator.AddRule("email", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "email", Message: "Email is required"}})
	validator.AddRule("email", &validation.EmailRule{BaseRule: &validation.BaseRule{Field: "email", Message: "Email must be valid"}})
	validator.AddRule("age", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "age", Message: "Age is required"}})
	validator.AddRule("age", &validation.IntegerRule{BaseRule: &validation.BaseRule{Field: "age", Message: "Age must be an integer"}})
	validator.AddRule("age", validation.NewMinRule("age", "Age must be at least 18", 18))
	validator.AddRule("age", validation.NewMaxRule("age", "Age must be at most 120", 120))

	// Validate
	if validator.Validate(data) {
		fmt.Println("✅ Validation passed")
		fmt.Printf("  Name: %s\n", data["name"])
		fmt.Printf("  Email: %s\n", data["email"])
		fmt.Printf("  Age: %v\n", data["age"])
	} else {
		fmt.Println("❌ Validation failed")
		for field, errors := range validator.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Example 2: Validation with Errors
	fmt.Println("\n=== Example 2: Validation with Errors ===")

	// Invalid data
	invalidData := map[string]interface{}{
		"name":     "",              // Empty name
		"email":    "invalid-email", // Invalid email
		"age":      15,              // Too young
		"password": "123",           // Too short
		"website":  "not-a-url",     // Invalid URL
	}

	// Create new validator for invalid data
	invalidValidator := validation.NewValidator()
	invalidValidator.AddRule("name", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "name", Message: "Name is required"}})
	invalidValidator.AddRule("name", validation.NewStringRule("name", "Name must be a string", 2, 50))
	invalidValidator.AddRule("email", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "email", Message: "Email is required"}})
	invalidValidator.AddRule("email", &validation.EmailRule{BaseRule: &validation.BaseRule{Field: "email", Message: "Email must be valid"}})
	invalidValidator.AddRule("age", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "age", Message: "Age is required"}})
	invalidValidator.AddRule("age", validation.NewMinRule("age", "Age must be at least 18", 18))
	invalidValidator.AddRule("password", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "password", Message: "Password is required"}})
	invalidValidator.AddRule("password", validation.NewStringRule("password", "Password must be a string", 8, 128))

	if !invalidValidator.Validate(invalidData) {
		fmt.Println("✅ Validation correctly failed")
		for field, errors := range invalidValidator.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Example 3: Request-based Validation
	fmt.Println("\n=== Example 3: Request-based Validation ===")

	// Create request validator
	request := validation.NewRequest(map[string]interface{}{
		"username": "jane_doe",
		"email":    "jane@example.com",
		"age":      28,
		"password": "securepassword123",
		"website":  "https://jane.com",
	})

	// Add validation rules using fluent API
	request.Required("username").
		StringWithLength("username", 3, 20).
		Required("email").
		Email("email").
		Required("age").
		IntegerRange("age", 18, 100).
		Required("password").
		StringWithLength("password", 8, 128).
		URL("website")

	// Validate
	if request.Validate() {
		fmt.Println("✅ Request validation passed")
		fmt.Printf("  Username: %s\n", request.GetString("username"))
		fmt.Printf("  Email: %s\n", request.GetString("email"))
		fmt.Printf("  Age: %d\n", request.GetInt("age"))
	} else {
		fmt.Println("❌ Request validation failed")
		for field, errors := range request.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Example 4: Custom Validation Rules
	fmt.Println("\n=== Example 4: Custom Validation Rules ===")

	// Custom rule: username cannot be "admin"
	customRequest := validation.NewRequest(map[string]interface{}{
		"username": "admin",
		"role":     "user",
	})

	// Add custom validation logic
	customRequest.Required("username").
		String("username").
		Custom("username", &validation.BaseRule{
			Field:   "username",
			Message: "Username 'admin' is reserved",
		})

	if !customRequest.Validate() {
		fmt.Println("✅ Custom validation correctly failed")
		fmt.Printf("  Username error: %s\n", customRequest.GetError("username"))
	}

	// Example 5: Struct-based Validation
	fmt.Println("\n=== Example 5: Struct-based Validation ===")

	userData := User{
		Username:        "john_doe",
		Email:           "john@example.com",
		Password:        "securepassword123",
		ConfirmPassword: "securepassword123",
		FirstName:       "John",
		LastName:        "Doe",
		Age:             30,
		Bio:             "Software developer",
		Website:         "https://johndoe.com",
	}

	// Parse rules from struct (this would be implemented)
	fmt.Println("Struct validation features:")
	fmt.Println("  - Automatic rule parsing from struct tags")
	fmt.Println("  - Type-safe validation")
	fmt.Println("  - Reusable validation rules")
	fmt.Printf("  - User data prepared: %s (%s)\n", userData.Username, userData.Email)

	// Example 6: Post Validation
	fmt.Println("\n=== Example 6: Post Validation ===")

	postData := Post{
		Title:       "My First Blog Post",
		Content:     "This is the content of my first blog post.",
		Tags:        []string{"go", "programming", "web development"},
		IsPublished: true,
		Category:    "tech",
	}

	// Create post validator
	postValidator := validation.NewValidator()
	postValidator.AddRule("title", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "title", Message: "Title is required"}})
	postValidator.AddRule("title", validation.NewStringRule("title", "Title must be a string", 5, 200))
	postValidator.AddRule("content", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "content", Message: "Content is required"}})
	postValidator.AddRule("content", validation.NewStringRule("content", "Content must be a string", 10, 1000))
	postValidator.AddRule("category", &validation.RequiredRule{BaseRule: &validation.BaseRule{Field: "category", Message: "Category is required"}})
	postValidator.AddRule("category", validation.NewInRule("category", "Category must be one of: tech, business, lifestyle", []interface{}{"tech", "business", "lifestyle"}))

	// Convert struct to map for validation
	postMap := map[string]interface{}{
		"title":    postData.Title,
		"content":  postData.Content,
		"tags":     postData.Tags,
		"category": postData.Category,
	}

	if postValidator.Validate(postMap) {
		fmt.Println("✅ Post validation passed")
		fmt.Printf("  Title: %s\n", postData.Title)
		fmt.Printf("  Category: %s\n", postData.Category)
		fmt.Printf("  Tags: %v\n", postData.Tags)
	} else {
		fmt.Println("❌ Post validation failed")
		for field, errors := range postValidator.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Example 7: Available Validation Rules
	fmt.Println("\n=== Example 7: Available Validation Rules ===")

	fmt.Println("Available validation rules:")
	fmt.Println("  - Required: Field must be present and not empty")
	fmt.Println("  - String: Field must be a string")
	fmt.Println("  - Email: Field must be a valid email address")
	fmt.Println("  - Integer: Field must be an integer")
	fmt.Println("  - MinLength: String must be at least N characters")
	fmt.Println("  - MaxLength: String must be at most N characters")
	fmt.Println("  - Min: Number must be at least N")
	fmt.Println("  - Max: Number must be at most N")
	fmt.Println("  - Regex: Field must match regular expression")
	fmt.Println("  - In: Field must be one of the specified values")
	fmt.Println("  - NotIn: Field must not be one of the specified values")
	fmt.Println("  - URL: Field must be a valid URL")
	fmt.Println("  - Custom: Custom validation logic")

	fmt.Println("\n🎉 All validation examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use validation in your web applications")
	fmt.Println("2. Create custom validation rules for your domain")
	fmt.Println("3. Integrate with form handling")
	fmt.Println("4. Use struct tags for automatic validation")
	fmt.Println("5. Implement validation middleware")
	fmt.Println("6. Add validation to API endpoints")
	fmt.Println("7. Create reusable validation components")
	fmt.Println("8. Implement client-side validation integration")
}
