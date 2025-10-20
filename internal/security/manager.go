package security

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Permission represents a permission
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// Role represents a role with permissions
type Role struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// UserRole represents a user-role assignment
type UserRole struct {
	UserID     uint       `json:"user_id"`
	RoleID     uint       `json:"role_id"`
	AssignedAt time.Time  `json:"assigned_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	UserID    *uint                  `json:"user_id,omitempty"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Voter represents a security voter
type Voter interface {
	Supports(attribute string, subject interface{}) bool
	Vote(context.Context, *User, string, interface{}) VoterResult
}

// VoterResult represents the result of a voter decision
type VoterResult int

const (
	VoterResultAbstain VoterResult = iota
	VoterResultGrant
	VoterResultDeny
)

// User represents a user with security context
type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Roles     []Role    `json:"roles"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SecurityManager manages security operations
type SecurityManager struct {
	voters            []Voter
	eventLogger       SecurityEventLogger
	roleManager       RoleManager
	permissionManager PermissionManager
}

// SecurityEventLogger logs security events
type SecurityEventLogger interface {
	LogEvent(event *SecurityEvent) error
	GetEvents(userID *uint, eventType string, limit int) ([]*SecurityEvent, error)
}

// RoleManager manages roles and permissions
type RoleManager interface {
	CreateRole(role *Role) error
	GetRole(id uint) (*Role, error)
	GetRoleByName(name string) (*Role, error)
	UpdateRole(role *Role) error
	DeleteRole(id uint) error
	ListRoles() ([]*Role, error)

	AssignRoleToUser(userID, roleID uint, expiresAt *time.Time) error
	RemoveRoleFromUser(userID, roleID uint) error
	GetUserRoles(userID uint) ([]*Role, error)
	HasRole(userID uint, roleName string) bool
}

// PermissionManager manages permissions
type PermissionManager interface {
	CreatePermission(permission *Permission) error
	GetPermission(name string) (*Permission, error)
	UpdatePermission(permission *Permission) error
	DeletePermission(name string) error
	ListPermissions() ([]*Permission, error)

	AddPermissionToRole(roleID uint, permissionName string) error
	RemovePermissionFromRole(roleID uint, permissionName string) error
	GetRolePermissions(roleID uint) ([]*Permission, error)
	HasPermission(userID uint, permissionName string) bool
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(
	eventLogger SecurityEventLogger,
	roleManager RoleManager,
	permissionManager PermissionManager,
) *SecurityManager {
	return &SecurityManager{
		voters:            make([]Voter, 0),
		eventLogger:       eventLogger,
		roleManager:       roleManager,
		permissionManager: permissionManager,
	}
}

// AddVoter adds a voter to the security manager
func (sm *SecurityManager) AddVoter(voter Voter) {
	sm.voters = append(sm.voters, voter)
}

// IsGranted checks if a user is granted access to a resource/action
func (sm *SecurityManager) IsGranted(ctx context.Context, user *User, attribute string, subject interface{}) bool {
	// Check if user is active
	if !user.IsActive {
		sm.logSecurityEvent(&SecurityEvent{
			Type:      "access_denied",
			UserID:    &user.ID,
			Resource:  fmt.Sprintf("%v", subject),
			Action:    attribute,
			Success:   false,
			Message:   "User account is inactive",
			Timestamp: time.Now(),
		})
		return false
	}

	// Check role-based permissions first
	if sm.hasPermission(user.ID, attribute) {
		sm.logSecurityEvent(&SecurityEvent{
			Type:      "access_granted",
			UserID:    &user.ID,
			Resource:  fmt.Sprintf("%v", subject),
			Action:    attribute,
			Success:   true,
			Message:   "Access granted via role permission",
			Timestamp: time.Now(),
		})
		return true
	}

	// Use voters for more complex authorization logic
	voterResults := make([]VoterResult, 0)

	for _, voter := range sm.voters {
		if voter.Supports(attribute, subject) {
			result := voter.Vote(ctx, user, attribute, subject)
			voterResults = append(voterResults, result)
		}
	}

	// Determine final result based on voter results
	granted := sm.evaluateVoterResults(voterResults)

	sm.logSecurityEvent(&SecurityEvent{
		Type:      "access_decision",
		UserID:    &user.ID,
		Resource:  fmt.Sprintf("%v", subject),
		Action:    attribute,
		Success:   granted,
		Message:   fmt.Sprintf("Access %s via voters", map[bool]string{true: "granted", false: "denied"}[granted]),
		Timestamp: time.Now(),
	})

	return granted
}

// hasPermission checks if user has a specific permission
func (sm *SecurityManager) hasPermission(userID uint, permissionName string) bool {
	return sm.permissionManager.HasPermission(userID, permissionName)
}

// evaluateVoterResults evaluates voter results using consensus strategy
func (sm *SecurityManager) evaluateVoterResults(results []VoterResult) bool {
	if len(results) == 0 {
		return false // No voters = deny
	}

	grantCount := 0
	denyCount := 0

	for _, result := range results {
		switch result {
		case VoterResultGrant:
			grantCount++
		case VoterResultDeny:
			denyCount++
		}
	}

	// If any voter denies, deny access
	if denyCount > 0 {
		return false
	}

	// If at least one voter grants, grant access
	return grantCount > 0
}

// logSecurityEvent logs a security event
func (sm *SecurityManager) logSecurityEvent(event *SecurityEvent) {
	if sm.eventLogger != nil {
		sm.eventLogger.LogEvent(event)
	}
}

// GetSecurityEvents retrieves security events
func (sm *SecurityManager) GetSecurityEvents(userID *uint, eventType string, limit int) ([]*SecurityEvent, error) {
	return sm.eventLogger.GetEvents(userID, eventType, limit)
}

// CreateRole creates a new role
func (sm *SecurityManager) CreateRole(role *Role) error {
	return sm.roleManager.CreateRole(role)
}

// AssignRoleToUser assigns a role to a user
func (sm *SecurityManager) AssignRoleToUser(userID, roleID uint, expiresAt *time.Time) error {
	err := sm.roleManager.AssignRoleToUser(userID, roleID, expiresAt)
	if err != nil {
		return err
	}

	sm.logSecurityEvent(&SecurityEvent{
		Type:      "role_assigned",
		UserID:    &userID,
		Resource:  fmt.Sprintf("role_%d", roleID),
		Action:    "assign",
		Success:   true,
		Message:   "Role assigned to user",
		Timestamp: time.Now(),
	})

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (sm *SecurityManager) RemoveRoleFromUser(userID, roleID uint) error {
	err := sm.roleManager.RemoveRoleFromUser(userID, roleID)
	if err != nil {
		return err
	}

	sm.logSecurityEvent(&SecurityEvent{
		Type:      "role_removed",
		UserID:    &userID,
		Resource:  fmt.Sprintf("role_%d", roleID),
		Action:    "remove",
		Success:   true,
		Message:   "Role removed from user",
		Timestamp: time.Now(),
	})

	return nil
}

// HasRole checks if user has a specific role
func (sm *SecurityManager) HasRole(userID uint, roleName string) bool {
	return sm.roleManager.HasRole(userID, roleName)
}

// HasPermission checks if user has a specific permission
func (sm *SecurityManager) HasPermission(userID uint, permissionName string) bool {
	return sm.permissionManager.HasPermission(userID, permissionName)
}

// Built-in Voters

// RoleVoter votes based on user roles
type RoleVoter struct {
	roleManager RoleManager
}

// NewRoleVoter creates a new role voter
func NewRoleVoter(roleManager RoleManager) *RoleVoter {
	return &RoleVoter{
		roleManager: roleManager,
	}
}

// Supports checks if this voter supports the attribute
func (rv *RoleVoter) Supports(attribute string, subject interface{}) bool {
	return strings.HasPrefix(attribute, "ROLE_")
}

// Vote votes on the attribute
func (rv *RoleVoter) Vote(ctx context.Context, user *User, attribute string, subject interface{}) VoterResult {
	roleName := strings.TrimPrefix(attribute, "ROLE_")

	if rv.roleManager.HasRole(user.ID, roleName) {
		return VoterResultGrant
	}

	return VoterResultAbstain
}

// PermissionVoter votes based on user permissions
type PermissionVoter struct {
	permissionManager PermissionManager
}

// NewPermissionVoter creates a new permission voter
func NewPermissionVoter(permissionManager PermissionManager) *PermissionVoter {
	return &PermissionVoter{
		permissionManager: permissionManager,
	}
}

// Supports checks if this voter supports the attribute
func (pv *PermissionVoter) Supports(attribute string, subject interface{}) bool {
	return !strings.HasPrefix(attribute, "ROLE_")
}

// Vote votes on the attribute
func (pv *PermissionVoter) Vote(ctx context.Context, user *User, attribute string, subject interface{}) VoterResult {
	if pv.permissionManager.HasPermission(user.ID, attribute) {
		return VoterResultGrant
	}

	return VoterResultAbstain
}

// ResourceVoter votes based on resource ownership
type ResourceVoter struct{}

// NewResourceVoter creates a new resource voter
func NewResourceVoter() *ResourceVoter {
	return &ResourceVoter{}
}

// Supports checks if this voter supports the attribute
func (rv *ResourceVoter) Supports(attribute string, subject interface{}) bool {
	return strings.HasSuffix(attribute, "_OWNER")
}

// Vote votes on the attribute
func (rv *ResourceVoter) Vote(ctx context.Context, user *User, attribute string, subject interface{}) VoterResult {
	// Check if subject has an OwnerID field
	subjectValue := reflect.ValueOf(subject)
	if subjectValue.Kind() == reflect.Ptr {
		subjectValue = subjectValue.Elem()
	}

	ownerIDField := subjectValue.FieldByName("OwnerID")
	if !ownerIDField.IsValid() {
		return VoterResultAbstain
	}

	if ownerIDField.Uint() == uint64(user.ID) {
		return VoterResultGrant
	}

	return VoterResultDeny
}

// ExpressionVoter votes based on expression evaluation
type ExpressionVoter struct {
	expressions map[string]string
}

// NewExpressionVoter creates a new expression voter
func NewExpressionVoter() *ExpressionVoter {
	return &ExpressionVoter{
		expressions: make(map[string]string),
	}
}

// AddExpression adds an expression for an attribute
func (ev *ExpressionVoter) AddExpression(attribute, expression string) {
	ev.expressions[attribute] = expression
}

// Supports checks if this voter supports the attribute
func (ev *ExpressionVoter) Supports(attribute string, subject interface{}) bool {
	_, exists := ev.expressions[attribute]
	return exists
}

// Vote votes on the attribute using expression evaluation
func (ev *ExpressionVoter) Vote(ctx context.Context, user *User, attribute string, subject interface{}) VoterResult {
	expression, exists := ev.expressions[attribute]
	if !exists {
		return VoterResultAbstain
	}

	// Simple expression evaluation (in a real implementation, you'd use a proper expression engine)
	// This is a placeholder for demonstration
	if ev.evaluateExpression(expression, user, subject) {
		return VoterResultGrant
	}

	return VoterResultDeny
}

// evaluateExpression evaluates a simple expression
func (ev *ExpressionVoter) evaluateExpression(expression string, user *User, subject interface{}) bool {
	// This is a simplified implementation
	// In a real system, you'd use a proper expression engine like expr or govaluate

	switch expression {
	case "user.isActive":
		return user.IsActive
	case "user.id > 0":
		return user.ID > 0
	case "subject != nil":
		return subject != nil
	default:
		return false
	}
}

// InMemorySecurityEventLogger logs security events in memory
type InMemorySecurityEventLogger struct {
	events []*SecurityEvent
}

// NewInMemorySecurityEventLogger creates a new in-memory security event logger
func NewInMemorySecurityEventLogger() *InMemorySecurityEventLogger {
	return &InMemorySecurityEventLogger{
		events: make([]*SecurityEvent, 0),
	}
}

// LogEvent logs a security event
func (imsel *InMemorySecurityEventLogger) LogEvent(event *SecurityEvent) error {
	imsel.events = append(imsel.events, event)
	return nil
}

// GetEvents retrieves security events
func (imsel *InMemorySecurityEventLogger) GetEvents(userID *uint, eventType string, limit int) ([]*SecurityEvent, error) {
	var filtered []*SecurityEvent

	for _, event := range imsel.events {
		if userID != nil && event.UserID != nil && *event.UserID != *userID {
			continue
		}

		if eventType != "" && event.Type != eventType {
			continue
		}

		filtered = append(filtered, event)

		if limit > 0 && len(filtered) >= limit {
			break
		}
	}

	return filtered, nil
}

// SecurityMiddleware provides security middleware for HTTP requests
type SecurityMiddleware struct {
	securityManager *SecurityManager
	userProvider    UserProvider
}

// UserProvider provides user information from requests
type UserProvider interface {
	GetUserFromRequest(r *http.Request) (*User, error)
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(securityManager *SecurityManager, userProvider UserProvider) *SecurityMiddleware {
	return &SecurityMiddleware{
		securityManager: securityManager,
		userProvider:    userProvider,
	}
}

// RequireAuth middleware requires authentication
func (sm *SecurityMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := sm.userProvider.GetUserFromRequest(r)
		if err != nil || user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware requires specific roles
func (sm *SecurityMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := sm.userProvider.GetUserFromRequest(r)
			if err != nil || user == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				if sm.securityManager.HasRole(user.ID, role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders middleware adds security headers
func (sm *SecurityMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}

// GetAvailablePresets returns available security header presets
func GetAvailablePresets() []string {
	return []string{"strict", "moderate", "permissive"}
}

// GetPresetInfo returns basic information about a preset
// Only fields required by examples are provided
func GetPresetInfo(preset string) map[string]string {
	switch preset {
	case "strict":
		return map[string]string{
			"name":        "Strict Security",
			"description": "Maximum security with all headers enabled",
		}
	case "moderate":
		return map[string]string{
			"name":        "Moderate Security",
			"description": "Balanced security with essential headers",
		}
	case "permissive":
		return map[string]string{
			"name":        "Permissive Security",
			"description": "Minimal security headers",
		}
	default:
		return map[string]string{
			"name":        preset,
			"description": "Custom preset",
		}
	}
}

// SecurityHeaderManager provides preset-based security headers
type SecurityHeaderManager struct {
	preset string
}

// NewSecurityHeaderManager creates a header manager with a preset
func NewSecurityHeaderManager(preset string) *SecurityHeaderManager {
	return &SecurityHeaderManager{preset: preset}
}

// GetHeaders returns standard security headers based on preset
func (shm *SecurityHeaderManager) GetHeaders() map[string]string {
	headers := make(map[string]string)

	switch shm.preset {
	case "strict":
		headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
		headers["X-Content-Type-Options"] = "nosniff"
		headers["X-Frame-Options"] = "DENY"
		headers["X-XSS-Protection"] = "1; mode=block"
		headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
		headers["Content-Security-Policy"] = "default-src 'self'"
	case "moderate":
		headers["X-Content-Type-Options"] = "nosniff"
		headers["X-Frame-Options"] = "SAMEORIGIN"
		headers["X-XSS-Protection"] = "1; mode=block"
	case "permissive":
		headers["X-Content-Type-Options"] = "nosniff"
	default:
		headers["X-Content-Type-Options"] = "nosniff"
	}

	return headers
}

// CSPBuilder builds Content Security Policy headers
type CSPBuilder struct {
	directives map[string][]string
}

// NewCSPBuilder creates a new CSP builder
func NewCSPBuilder() *CSPBuilder {
	return &CSPBuilder{directives: make(map[string][]string)}
}

// AddDirective adds a directive and returns the builder for chaining
func (c *CSPBuilder) AddDirective(name string, values ...string) *CSPBuilder {
	c.directives[name] = values
	return c
}

// Build generates the CSP header value
func (c *CSPBuilder) Build() string {
	parts := make([]string, 0, len(c.directives))
	for name, values := range c.directives {
		if len(values) > 0 {
			parts = append(parts, name+" "+strings.Join(values, " "))
		} else {
			parts = append(parts, name)
		}
	}
	return strings.Join(parts, "; ")
}
