package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/mrhoseah/dolphin/internal/auth"
	"github.com/mrhoseah/dolphin/internal/security"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Declarative Authorization with Policy Engine
	fmt.Println("=== Example 1: Declarative Authorization ===")

	// Create policy engine
	policyEngine, err := auth.NewPolicyEngine(logger)
	if err != nil {
		log.Fatalf("Failed to create policy engine: %v", err)
	}

	// Add some policies
	policyEngine.AddPolicy(&auth.PolicyRule{
		Subject: "admin",
		Object:  "posts",
		Action:  "delete",
		Effect:  "allow",
	})

	policyEngine.AddPolicy(&auth.PolicyRule{
		Subject: "user",
		Object:  "posts",
		Action:  "read",
		Effect:  "allow",
	})

	// Assign roles to users
	policyEngine.AssignRole("alice", "admin")
	policyEngine.AssignRole("bob", "user")

	// Test permissions
	canDelete, _ := policyEngine.Can(context.Background(), "alice", "delete", "posts")
	fmt.Printf("Alice can delete posts: %v\n", canDelete)

	canRead, _ := policyEngine.Can(context.Background(), "bob", "read", "posts")
	fmt.Printf("Bob can read posts: %v\n", canRead)

	canDeleteBob, _ := policyEngine.Can(context.Background(), "bob", "delete", "posts")
	fmt.Printf("Bob can delete posts: %v\n", canDeleteBob)

	// Example 2: Secure Headers
	fmt.Println("\n=== Example 2: Secure Headers ===")

	// Create security header manager
	headerManager := security.NewSecurityHeaderManager("strict")

	// Get headers
	headers := headerManager.GetHeaders()
	fmt.Println("Security Headers:")
	for name, value := range headers {
		fmt.Printf("  %s: %s\n", name, value)
	}

	// Example 3: Credential Management
	fmt.Println("\n=== Example 3: Credential Management ===")

	// Create credential manager
	credManager, err := security.NewCredentialManager(".dolphin/credentials.key")
	if err != nil {
		log.Fatalf("Failed to create credential manager: %v", err)
	}

	// Set some credentials
	credManager.SetCredential("DB_PASSWORD", "super-secret-password")
	credManager.SetCredential("API_KEY", "sk-1234567890abcdef")

	// Retrieve credentials
	dbPassword, _ := credManager.GetCredential("DB_PASSWORD")
	fmt.Printf("DB Password: %s\n", dbPassword)

	apiKey, _ := credManager.GetCredential("API_KEY")
	fmt.Printf("API Key: %s\n", apiKey)

	// Example 4: CSRF Protection
	fmt.Println("\n=== Example 4: CSRF Protection ===")

	// Create session store
	store := sessions.NewCookieStore([]byte("super-secret-session-key"))

	// Create CSRF manager
	csrfManager, err := security.NewCSRFManager(security.DefaultCSRFConfig(), store, logger)
	if err != nil {
		log.Fatalf("Failed to create CSRF manager: %v", err)
	}

	// Generate CSRF token
	sessionID := "session-12345"
	token, err := csrfManager.GenerateToken(sessionID)
	if err != nil {
		log.Fatalf("Failed to generate CSRF token: %v", err)
	}

	fmt.Printf("Generated CSRF token: %s\n", token)

	// Validate token
	valid, err := csrfManager.ValidateToken(sessionID, token)
	if err != nil {
		log.Fatalf("Failed to validate CSRF token: %v", err)
	}

	fmt.Printf("Token is valid: %v\n", valid)

	// Example 5: Template Helpers
	fmt.Println("\n=== Example 5: Template Helpers ===")

	// Create policy helper
	policyHelper := auth.NewPolicyHelper(policyEngine)

	// Test template helpers
	canEdit := policyHelper.Can("alice", "edit", "posts")
	fmt.Printf("Alice can edit posts (template helper): %v\n", canEdit)

	hasRole := policyHelper.HasRole("alice", "admin")
	fmt.Printf("Alice has admin role: %v\n", hasRole)

	// Create CSRF template helper
	csrfHelper := security.NewCSRFTemplateHelper(csrfManager)

	// Generate template helpers
	tokenField := csrfHelper.TokenField(sessionID)
	fmt.Printf("CSRF Token Field: %s\n", tokenField)

	metaTag := csrfHelper.MetaTag(sessionID)
	fmt.Printf("CSRF Meta Tag: %s\n", metaTag)

	// Example 6: Security Headers Middleware
	fmt.Println("\n=== Example 6: Security Headers Middleware ===")

	// Create HTTP server with security middleware
	mux := http.NewServeMux()

	// Add security headers middleware
	securityMiddleware := security.SecurityHeadersMiddleware(headerManager)

	// Add a test handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World! This response has security headers.")
	})

	// Wrap with security middleware
	handler := securityMiddleware(mux)

	// Start server in goroutine
	go func() {
		fmt.Println("Starting server on :8080 with security headers...")
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Test the server
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		log.Printf("Failed to test server: %v", err)
	} else {
		fmt.Println("Server Response Headers:")
		for name, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", name, value)
			}
		}
		resp.Body.Close()
	}

	// Example 7: CSP Builder
	fmt.Println("\n=== Example 7: CSP Builder ===")

	cspBuilder := security.NewCSPBuilder()
	csp := cspBuilder.
		SetDefaultSrc("'self'").
		SetScriptSrc("'self'", "'unsafe-inline'").
		SetStyleSrc("'self'", "'unsafe-inline'").
		SetImgSrc("'self'", "data:", "https:").
		SetFontSrc("'self'", "data:").
		SetConnectSrc("'self'").
		SetFrameAncestors("'none'").
		SetBaseURI("'self'").
		SetFormAction("'self'").
		Build()

	fmt.Printf("Generated CSP: %s\n", csp)

	// Example 8: Security Presets
	fmt.Println("\n=== Example 8: Security Presets ===")

	presets := security.GetAvailablePresets()
	fmt.Println("Available Security Presets:")
	for _, preset := range presets {
		info, _ := security.GetPresetInfo(preset)
		fmt.Printf("  %s: %s\n", preset, info.Description)
	}

	// Example 9: Environment Credential Manager
	fmt.Println("\n=== Example 9: Environment Credential Manager ===")

	envManager := security.NewEnvironmentCredentialManager("DOLPHIN_")

	// Set some environment variables for testing
	os.Setenv("DOLPHIN_DB_HOST", "localhost")
	os.Setenv("DOLPHIN_DB_PORT", "5432")

	// Get credentials
	dbHost, _ := envManager.GetCredential("db_host")
	fmt.Printf("DB Host from environment: %s\n", dbHost)

	dbPort, _ := envManager.GetCredential("db_port")
	fmt.Printf("DB Port from environment: %s\n", dbPort)

	fmt.Println("\n🎉 All security examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin security policy create' to create custom policies")
	fmt.Println("2. Use 'dolphin security credentials encrypt' to encrypt your .env file")
	fmt.Println("3. Use 'dolphin security csrf generate' to test CSRF tokens")
	fmt.Println("4. Integrate these components into your application")
}
