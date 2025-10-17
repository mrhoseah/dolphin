package main

import (
	"fmt"
	"log"

	"github.com/mrhoseah/dolphin/internal/validation"
)

// User represents a user with validation and sanitization tags
type User struct {
	Username        string `json:"username" validate:"required,min_length:3,max_length:20,alpha_numeric" sanitize:"trim,lowercase"`
	Email           string `json:"email" validate:"required,email" sanitize:"trim,lowercase"`
	Password        string `json:"password" validate:"required,min_length:8" sanitize:"trim"`
	ConfirmPassword string `json:"confirm_password" validate:"required" sanitize:"trim"`
	FirstName       string `json:"first_name" validate:"required,alpha" sanitize:"trim"`
	LastName        string `json:"last_name" validate:"required,alpha" sanitize:"trim"`
	Age             int    `json:"age" validate:"required,min:18,max:120"`
	Bio             string `json:"bio" validate:"max_length:500" sanitize:"trim,strip_html,normalize_whitespace"`
	Website         string `json:"website" validate:"url" sanitize:"trim,lowercase"`
}

// Post represents a blog post with validation
type Post struct {
	Title       string   `json:"title" validate:"required,min_length:5,max_length:200" sanitize:"trim,strip_html"`
	Content     string   `json:"content" validate:"required,min_length:10" sanitize:"trim,strip_html,normalize_whitespace"`
	Tags        []string `json:"tags" validate:"max_length:10" sanitize:"trim,lowercase"`
	IsPublished bool     `json:"is_published"`
	Category    string   `json:"category" validate:"required,in:tech,business,lifestyle" sanitize:"trim,lowercase"`
}

func main() {
	// Create validation manager
	validator := validation.NewFieldValidator()
	sanitizer := validation.NewFieldSanitizer()
	manager := validation.NewValidationManager(nil)

	// Example 1: Valid user data
	fmt.Println("=== Example 1: Valid User Data ===")
	user := User{
		Username:        "  john_doe123  ",
		Email:           "  JOHN@EXAMPLE.COM  ",
		Password:        "securepassword123",
		ConfirmPassword: "securepassword123",
		FirstName:       "  John  ",
		LastName:        "  Doe  ",
		Age:             25,
		Bio:             "  <p>Software developer with 5 years experience</p>  ",
		Website:         "  HTTPS://JOHNDOE.COM  ",
	}

	fmt.Println("Before sanitization:")
	fmt.Printf("Username: '%s'\n", user.Username)
	fmt.Printf("Email: '%s'\n", user.Email)
	fmt.Printf("FirstName: '%s'\n", user.FirstName)
	fmt.Printf("Bio: '%s'\n", user.Bio)
	fmt.Printf("Website: '%s'\n", user.Website)

	// Sanitize the data
	if err := sanitizer.Sanitize(&user); err != nil {
		log.Fatalf("Sanitization failed: %v", err)
	}

	fmt.Println("\nAfter sanitization:")
	fmt.Printf("Username: '%s'\n", user.Username)
	fmt.Printf("Email: '%s'\n", user.Email)
	fmt.Printf("FirstName: '%s'\n", user.FirstName)
	fmt.Printf("Bio: '%s'\n", user.Bio)
	fmt.Printf("Website: '%s'\n", user.Website)

	// Validate the data
	if err := validator.Validate(&user); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("✅ User data is valid!")

	// Example 2: Invalid user data
	fmt.Println("\n=== Example 2: Invalid User Data ===")
	invalidUser := User{
		Username:        "ab",            // Too short
		Email:           "invalid-email", // Invalid email
		Password:        "123",           // Too short
		ConfirmPassword: "456",           // Doesn't match
		FirstName:       "John123",       // Contains numbers
		LastName:        "Doe",
		Age:             15, // Too young
		Bio: "A very long bio that exceeds the maximum length limit of 500 characters. " +
			"This bio is intentionally made very long to demonstrate the max_length validation rule. " +
			"It contains multiple sentences and should trigger the validation error when the length " +
			"exceeds the specified limit. The validation system should catch this and return an " +
			"appropriate error message indicating that the bio field is too long and needs to be " +
			"shortened to meet the requirements. This is a comprehensive test of the validation " +
			"system's ability to handle string length constraints effectively.",
		Website: "not-a-url", // Invalid URL
	}

	fmt.Println("Testing invalid user data...")
	if err := validator.Validate(&invalidUser); err != nil {
		fmt.Printf("❌ Validation failed as expected:\n%v\n", err)
	}

	// Example 3: Post validation
	fmt.Println("\n=== Example 3: Post Validation ===")
	post := Post{
		Title:       "  <h1>My First Blog Post</h1>  ",
		Content:     "  <p>This is the content of my first blog post. It contains some HTML tags that should be stripped during sanitization.</p>  ",
		Tags:        []string{"  GO  ", "  PROGRAMMING  ", "  WEB DEVELOPMENT  "},
		IsPublished: true,
		Category:    "  TECH  ",
	}

	fmt.Println("Before sanitization:")
	fmt.Printf("Title: '%s'\n", post.Title)
	fmt.Printf("Content: '%s'\n", post.Content)
	fmt.Printf("Tags: %v\n", post.Tags)
	fmt.Printf("Category: '%s'\n", post.Category)

	// Sanitize the data
	if err := sanitizer.Sanitize(&post); err != nil {
		log.Fatalf("Sanitization failed: %v", err)
	}

	fmt.Println("\nAfter sanitization:")
	fmt.Printf("Title: '%s'\n", post.Title)
	fmt.Printf("Content: '%s'\n", post.Content)
	fmt.Printf("Tags: %v\n", post.Tags)
	fmt.Printf("Category: '%s'\n", post.Category)

	// Validate the data
	if err := validator.Validate(&post); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("✅ Post data is valid!")

	// Example 4: Field-level validation
	fmt.Println("\n=== Example 4: Field-level Validation ===")

	// Test email validation
	email := "test@example.com"
	if err := validator.ValidateField(email, []string{"required", "email"}); err != nil {
		fmt.Printf("Email validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Email is valid")
	}

	// Test invalid email
	invalidEmail := "not-an-email"
	if err := validator.ValidateField(invalidEmail, []string{"required", "email"}); err != nil {
		fmt.Printf("❌ Invalid email validation failed as expected: %v\n", err)
	}

	// Test field sanitization
	username := "  USERNAME123  "
	sanitized, err := sanitizer.SanitizeField(username, []string{"trim", "lowercase"})
	if err != nil {
		log.Fatalf("Sanitization failed: %v", err)
	}
	fmt.Printf("Username sanitized: '%s' -> '%s'\n", username, sanitized)

	// Example 5: Using ValidationManager
	fmt.Println("\n=== Example 5: Using ValidationManager ===")

	user2 := User{
		Username:        "jane_doe",
		Email:           "jane@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
		FirstName:       "Jane",
		LastName:        "Doe",
		Age:             30,
		Bio:             "Software engineer",
		Website:         "https://jane.com",
	}

	// Validate and sanitize in one step
	if err := manager.ValidateAndSanitize(&user2); err != nil {
		log.Fatalf("Validation and sanitization failed: %v", err)
	}

	fmt.Println("✅ User data validated and sanitized successfully!")

	fmt.Println("\n🎉 All validation examples completed successfully!")
}
