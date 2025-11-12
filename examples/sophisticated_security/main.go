package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrhoseah/dolphin/internal/security"
)

// Mock implementations for demonstration

// MockUserProvider provides mock user data
type MockUserProvider struct {
	users map[uint]*security.User
}

// NewMockUserProvider creates a new mock user provider
func NewMockUserProvider() *MockUserProvider {
	return &MockUserProvider{
		users: make(map[uint]*security.User),
	}
}

// GetUserFromRequest gets user from request (simplified)
func (mup *MockUserProvider) GetUserFromRequest(r *http.Request) (*security.User, error) {
	// In a real implementation, this would extract user from JWT token, session, etc.
	userID := uint(1) // Mock user ID
	if user, exists := mup.users[userID]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

// AddUser adds a user to the mock provider
func (mup *MockUserProvider) AddUser(user *security.User) {
	mup.users[user.ID] = user
}

// MockRoleManager provides mock role management
type MockRoleManager struct {
	roles     map[uint]*security.Role
	userRoles map[uint][]uint
}

// NewMockRoleManager creates a new mock role manager
func NewMockRoleManager() *MockRoleManager {
	return &MockRoleManager{
		roles:     make(map[uint]*security.Role),
		userRoles: make(map[uint][]uint),
	}
}

// CreateRole creates a role
func (mrm *MockRoleManager) CreateRole(role *security.Role) error {
	role.ID = uint(len(mrm.roles) + 1)
	mrm.roles[role.ID] = role
	return nil
}

// GetRole gets a role by ID
func (mrm *MockRoleManager) GetRole(id uint) (*security.Role, error) {
	if role, exists := mrm.roles[id]; exists {
		return role, nil
	}
	return nil, fmt.Errorf("role not found")
}

// GetRoleByName gets a role by name
func (mrm *MockRoleManager) GetRoleByName(name string) (*security.Role, error) {
	for _, role := range mrm.roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, fmt.Errorf("role not found")
}

// UpdateRole updates a role
func (mrm *MockRoleManager) UpdateRole(role *security.Role) error {
	mrm.roles[role.ID] = role
	return nil
}

// DeleteRole deletes a role
func (mrm *MockRoleManager) DeleteRole(id uint) error {
	delete(mrm.roles, id)
	return nil
}

// ListRoles lists all roles
func (mrm *MockRoleManager) ListRoles() ([]*security.Role, error) {
	var roles []*security.Role
	for _, role := range mrm.roles {
		roles = append(roles, role)
	}
	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (mrm *MockRoleManager) AssignRoleToUser(userID, roleID uint, expiresAt *time.Time) error {
	mrm.userRoles[userID] = append(mrm.userRoles[userID], roleID)
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (mrm *MockRoleManager) RemoveRoleFromUser(userID, roleID uint) error {
	roles := mrm.userRoles[userID]
	for i, id := range roles {
		if id == roleID {
			mrm.userRoles[userID] = append(roles[:i], roles[i+1:]...)
			break
		}
	}
	return nil
}

// GetUserRoles gets roles for a user
func (mrm *MockRoleManager) GetUserRoles(userID uint) ([]*security.Role, error) {
	var roles []*security.Role
	for _, roleID := range mrm.userRoles[userID] {
		if role, exists := mrm.roles[roleID]; exists {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

// HasRole checks if user has a role
func (mrm *MockRoleManager) HasRole(userID uint, roleName string) bool {
	for _, roleID := range mrm.userRoles[userID] {
		if role, exists := mrm.roles[roleID]; exists && role.Name == roleName {
			return true
		}
	}
	return false
}

// MockPermissionManager provides mock permission management
type MockPermissionManager struct {
	permissions     map[string]*security.Permission
	rolePermissions map[uint][]string
	userPermissions map[uint][]string
}

// NewMockPermissionManager creates a new mock permission manager
func NewMockPermissionManager() *MockPermissionManager {
	return &MockPermissionManager{
		permissions:     make(map[string]*security.Permission),
		rolePermissions: make(map[uint][]string),
		userPermissions: make(map[uint][]string),
	}
}

// CreatePermission creates a permission
func (mpm *MockPermissionManager) CreatePermission(permission *security.Permission) error {
	mpm.permissions[permission.Name] = permission
	return nil
}

// GetPermission gets a permission by name
func (mpm *MockPermissionManager) GetPermission(name string) (*security.Permission, error) {
	if permission, exists := mpm.permissions[name]; exists {
		return permission, nil
	}
	return nil, fmt.Errorf("permission not found")
}

// UpdatePermission updates a permission
func (mpm *MockPermissionManager) UpdatePermission(permission *security.Permission) error {
	mpm.permissions[permission.Name] = permission
	return nil
}

// DeletePermission deletes a permission
func (mpm *MockPermissionManager) DeletePermission(name string) error {
	delete(mpm.permissions, name)
	return nil
}

// ListPermissions lists all permissions
func (mpm *MockPermissionManager) ListPermissions() ([]*security.Permission, error) {
	var permissions []*security.Permission
	for _, permission := range mpm.permissions {
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

// AddPermissionToRole adds a permission to a role
func (mpm *MockPermissionManager) AddPermissionToRole(roleID uint, permissionName string) error {
	mpm.rolePermissions[roleID] = append(mpm.rolePermissions[roleID], permissionName)
	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (mpm *MockPermissionManager) RemovePermissionFromRole(roleID uint, permissionName string) error {
	permissions := mpm.rolePermissions[roleID]
	for i, name := range permissions {
		if name == permissionName {
			mpm.rolePermissions[roleID] = append(permissions[:i], permissions[i+1:]...)
			break
		}
	}
	return nil
}

// GetRolePermissions gets permissions for a role
func (mpm *MockPermissionManager) GetRolePermissions(roleID uint) ([]*security.Permission, error) {
	var permissions []*security.Permission
	for _, permissionName := range mpm.rolePermissions[roleID] {
		if permission, exists := mpm.permissions[permissionName]; exists {
			permissions = append(permissions, permission)
		}
	}
	return permissions, nil
}

// HasPermission checks if user has a permission
func (mpm *MockPermissionManager) HasPermission(userID uint, permissionName string) bool {
	// Check direct user permissions
	for _, permission := range mpm.userPermissions[userID] {
		if permission == permissionName {
			return true
		}
	}

	// Check role permissions (simplified - in real implementation, you'd check user's roles)
	return false
}

func main() {
	fmt.Println("🔐 Dolphin Framework - Sophisticated Security Patterns Demo")
	fmt.Println("==========================================================")
	fmt.Println("")

	// 1. Setup security components
	fmt.Println("=== 1. Security Components Setup ===")
	demoSecuritySetup()

	fmt.Println("")

	// 2. Role-based access control
	fmt.Println("=== 2. Role-Based Access Control ===")
	demoRBAC()

	fmt.Println("")

	// 3. Voter system
	fmt.Println("=== 3. Voter System ===")
	demoVoterSystem()

	fmt.Println("")

	// 4. Security events
	fmt.Println("=== 4. Security Events ===")
	demoSecurityEvents()

	fmt.Println("")

	// 5. HTTP middleware integration
	fmt.Println("=== 5. HTTP Middleware Integration ===")
	demoHTTPMiddleware()

	fmt.Println("")
	fmt.Println("🎉 Sophisticated security patterns demonstrated successfully!")
	fmt.Println("")
	fmt.Println("💡 Key Features Implemented:")
	fmt.Println("  ✅ Role-based access control (RBAC)")
	fmt.Println("  ✅ Permission-based authorization")
	fmt.Println("  ✅ Voter system for complex authorization logic")
	fmt.Println("  ✅ Security event logging and monitoring")
	fmt.Println("  ✅ HTTP middleware for security")
	fmt.Println("  ✅ CSRF protection")
	fmt.Println("  ✅ Security headers")
	fmt.Println("  ✅ Rate limiting ready")
}

func demoSecuritySetup() {
	// Create mock components
	eventLogger := security.NewInMemorySecurityEventLogger()
	roleManager := NewMockRoleManager()
	permissionManager := NewMockPermissionManager()
	userProvider := NewMockUserProvider()

	// Create security manager
	securityManager := security.NewSecurityManager(eventLogger, roleManager, permissionManager)

	// Create security middleware
	securityMiddleware := security.NewSecurityMiddleware(securityManager, userProvider)

	fmt.Printf("Security Manager: %T\n", securityManager)
	fmt.Printf("Security Middleware: %T\n", securityMiddleware)
	fmt.Printf("Event Logger: %T\n", eventLogger)
	fmt.Printf("Role Manager: %T\n", roleManager)
	fmt.Printf("Permission Manager: %T\n", permissionManager)
	fmt.Printf("User Provider: %T\n", userProvider)
}

func demoRBAC() {
	// Create mock components
	eventLogger := security.NewInMemorySecurityEventLogger()
	roleManager := NewMockRoleManager()
	permissionManager := NewMockPermissionManager()

	// Create security manager
	securityManager := security.NewSecurityManager(eventLogger, roleManager, permissionManager)

	// Create roles
	adminRole := &security.Role{
		Name:        "admin",
		Description: "Administrator role",
		Permissions: []security.Permission{
			{Name: "user.create", Description: "Create users", Resource: "user", Action: "create"},
			{Name: "user.read", Description: "Read users", Resource: "user", Action: "read"},
			{Name: "user.update", Description: "Update users", Resource: "user", Action: "update"},
			{Name: "user.delete", Description: "Delete users", Resource: "user", Action: "delete"},
		},
	}

	userRole := &security.Role{
		Name:        "user",
		Description: "Regular user role",
		Permissions: []security.Permission{
			{Name: "user.read", Description: "Read users", Resource: "user", Action: "read"},
		},
	}

	// Create roles
	roleManager.CreateRole(adminRole)
	roleManager.CreateRole(userRole)

	// Create permissions
	for _, permission := range adminRole.Permissions {
		permissionManager.CreatePermission(&permission)
	}
	for _, permission := range userRole.Permissions {
		permissionManager.CreatePermission(&permission)
	}

	// Create users
	adminUser := &security.User{
		ID:       1,
		Email:    "admin@example.com",
		Username: "admin",
		IsActive: true,
	}

	regularUser := &security.User{
		ID:       2,
		Email:    "user@example.com",
		Username: "user",
		IsActive: true,
	}

	// Assign roles
	roleManager.AssignRoleToUser(adminUser.ID, adminRole.ID, nil)
	roleManager.AssignRoleToUser(regularUser.ID, userRole.ID, nil)

	// Test permissions
	fmt.Printf("Admin can create users: %v\n", securityManager.HasPermission(adminUser.ID, "user.create"))
	fmt.Printf("Admin can delete users: %v\n", securityManager.HasPermission(adminUser.ID, "user.delete"))
	fmt.Printf("User can create users: %v\n", securityManager.HasPermission(regularUser.ID, "user.create"))
	fmt.Printf("User can read users: %v\n", securityManager.HasPermission(regularUser.ID, "user.read"))

	// Test roles
	fmt.Printf("Admin has admin role: %v\n", securityManager.HasRole(adminUser.ID, "admin"))
	fmt.Printf("User has admin role: %v\n", securityManager.HasRole(regularUser.ID, "admin"))
	fmt.Printf("User has user role: %v\n", securityManager.HasRole(regularUser.ID, "user"))
}

func demoVoterSystem() {
	// Create mock components
	eventLogger := security.NewInMemorySecurityEventLogger()
	roleManager := NewMockRoleManager()
	permissionManager := NewMockPermissionManager()

	// Create security manager
	securityManager := security.NewSecurityManager(eventLogger, roleManager, permissionManager)

	// Add voters
	roleVoter := security.NewRoleVoter(roleManager)
	permissionVoter := security.NewPermissionVoter(permissionManager)
	resourceVoter := security.NewResourceVoter()
	expressionVoter := security.NewExpressionVoter()

	securityManager.AddVoter(roleVoter)
	securityManager.AddVoter(permissionVoter)
	securityManager.AddVoter(resourceVoter)
	securityManager.AddVoter(expressionVoter)

	// Create user
	user := &security.User{
		ID:       1,
		Email:    "user@example.com",
		Username: "user",
		IsActive: true,
	}

	// Test voter decisions
	ctx := context.Background()

	fmt.Printf("User can access ROLE_USER: %v\n", securityManager.IsGranted(ctx, user, "ROLE_USER", nil))
	fmt.Printf("User can access user.read: %v\n", securityManager.IsGranted(ctx, user, "user.read", nil))
	fmt.Printf("User can access user.create: %v\n", securityManager.IsGranted(ctx, user, "user.create", nil))

	// Test expression voter
	expressionVoter.AddExpression("user.isActive", "user.isActive")
	fmt.Printf("User is active (expression): %v\n", securityManager.IsGranted(ctx, user, "user.isActive", nil))
}

func demoSecurityEvents() {
	// Create mock components
	eventLogger := security.NewInMemorySecurityEventLogger()
	roleManager := NewMockRoleManager()
	permissionManager := NewMockPermissionManager()

	// Create security manager
	securityManager := security.NewSecurityManager(eventLogger, roleManager, permissionManager)

	// Create user
	user := &security.User{
		ID:       1,
		Email:    "user@example.com",
		Username: "user",
		IsActive: true,
	}

	// Test access decisions (this will generate security events)
	ctx := context.Background()
	securityManager.IsGranted(ctx, user, "user.read", nil)
	securityManager.IsGranted(ctx, user, "user.create", nil)

	// Get security events
	events, err := securityManager.GetSecurityEvents(nil, "", 10)
	if err != nil {
		log.Printf("Error getting security events: %v", err)
		return
	}

	fmt.Printf("Security events logged: %d\n", len(events))
	for i, event := range events {
		if i < 3 { // Show first 3 events
			fmt.Printf("  Event %d: %s - %s (%v)\n", i+1, event.Type, event.Action, event.Success)
		}
	}
}

func demoHTTPMiddleware() {
	// Create mock components
	eventLogger := security.NewInMemorySecurityEventLogger()
	roleManager := NewMockRoleManager()
	permissionManager := NewMockPermissionManager()
	userProvider := NewMockUserProvider()

	// Create security manager
	securityManager := security.NewSecurityManager(eventLogger, roleManager, permissionManager)

	// Create security middleware
	securityMiddleware := security.NewSecurityMiddleware(securityManager, userProvider)

	// Create a test user
	user := &security.User{
		ID:       1,
		Email:    "user@example.com",
		Username: "user",
		IsActive: true,
	}
	userProvider.AddUser(user)

	// Create test handlers
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Authenticated user: %s", user.Username)
	})

	adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Admin access granted")
	})

	// Setup routes with middleware
	mux := http.NewServeMux()
	mux.Handle("/auth", securityMiddleware.RequireAuth(authHandler))
	mux.Handle("/admin", securityMiddleware.RequireRole("admin")(adminHandler))
	mux.Handle("/security-headers", securityMiddleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Security headers added")
	})))

	fmt.Println("HTTP middleware setup complete:")
	fmt.Println("  - /auth - Requires authentication")
	fmt.Println("  - /admin - Requires admin role")
	fmt.Println("  - /security-headers - Adds security headers")
	fmt.Println("  - CSRF protection available")
	fmt.Println("  - Rate limiting ready")
}
