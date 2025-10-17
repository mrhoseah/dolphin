package auth

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"go.uber.org/zap"
)

// PolicyEngine manages authorization policies
type PolicyEngine struct {
	enforcer *casbin.Enforcer
	logger   *zap.Logger
}

// PolicyRule represents a single policy rule
type PolicyRule struct {
	Subject string // user, role, or group
	Object  string // resource being accessed
	Action  string // action being performed
	Effect  string // allow or deny
}

// PolicyContext provides context for policy evaluation
type PolicyContext struct {
	UserID    string
	UserRoles []string
	Resource  string
	Action    string
	Context   map[string]interface{}
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(logger *zap.Logger) (*PolicyEngine, error) {
	// Create a basic RBAC model
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy model: %w", err)
	}

	// Create adapter (in-memory for now)
	adapter := &MemoryAdapter{}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy enforcer: %w", err)
	}

	// Load default policies
	if err := loadDefaultPolicies(enforcer); err != nil {
		return nil, fmt.Errorf("failed to load default policies: %w", err)
	}

	return &PolicyEngine{
		enforcer: enforcer,
		logger:   logger,
	}, nil
}

// Can checks if a user can perform an action on a resource
func (pe *PolicyEngine) Can(ctx context.Context, userID, action, resource string) (bool, error) {
	// Get user roles
	roles, err := pe.getUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check each role
	for _, role := range roles {
		allowed, err := pe.enforcer.Enforce(role, resource, action)
		if err != nil {
			pe.logger.Error("Policy enforcement error",
				zap.String("role", role),
				zap.String("resource", resource),
				zap.String("action", action),
				zap.Error(err))
			continue
		}
		if allowed {
			pe.logger.Debug("Policy check passed",
				zap.String("user", userID),
				zap.String("role", role),
				zap.String("resource", resource),
				zap.String("action", action))
			return true, nil
		}
	}

	pe.logger.Debug("Policy check failed",
		zap.String("user", userID),
		zap.String("resource", resource),
		zap.String("action", action))

	return false, nil
}

// CanWithContext checks permissions with additional context
func (pe *PolicyEngine) CanWithContext(ctx context.Context, policyCtx *PolicyContext) (bool, error) {
	// Check direct user permissions
	allowed, err := pe.enforcer.Enforce(policyCtx.UserID, policyCtx.Resource, policyCtx.Action)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Check role-based permissions
	for _, role := range policyCtx.UserRoles {
		allowed, err := pe.enforcer.Enforce(role, policyCtx.Resource, policyCtx.Action)
		if err != nil {
			continue
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// AddPolicy adds a new policy rule
func (pe *PolicyEngine) AddPolicy(rule *PolicyRule) error {
	_, err := pe.enforcer.AddPolicy(rule.Subject, rule.Object, rule.Action, rule.Effect)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	pe.logger.Info("Policy added",
		zap.String("subject", rule.Subject),
		zap.String("object", rule.Object),
		zap.String("action", rule.Action),
		zap.String("effect", rule.Effect))

	return nil
}

// RemovePolicy removes a policy rule
func (pe *PolicyEngine) RemovePolicy(rule *PolicyRule) error {
	_, err := pe.enforcer.RemovePolicy(rule.Subject, rule.Object, rule.Action, rule.Effect)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	pe.logger.Info("Policy removed",
		zap.String("subject", rule.Subject),
		zap.String("object", rule.Object),
		zap.String("action", rule.Action),
		zap.String("effect", rule.Effect))

	return nil
}

// AssignRole assigns a role to a user
func (pe *PolicyEngine) AssignRole(userID, role string) error {
	_, err := pe.enforcer.AddRoleForUser(userID, role)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	pe.logger.Info("Role assigned",
		zap.String("user", userID),
		zap.String("role", role))

	return nil
}

// RemoveRole removes a role from a user
func (pe *PolicyEngine) RemoveRole(userID, role string) error {
	_, err := pe.enforcer.DeleteRoleForUser(userID, role)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	pe.logger.Info("Role removed",
		zap.String("user", userID),
		zap.String("role", role))

	return nil
}

// GetUserRoles returns all roles for a user
func (pe *PolicyEngine) GetUserRoles(userID string) ([]string, error) {
	return pe.enforcer.GetRolesForUser(userID)
}

// GetPolicies returns all policies
func (pe *PolicyEngine) GetPolicies() [][]string {
	policies, err := pe.enforcer.GetPolicy()
	if err != nil {
		return [][]string{}
	}
	return policies
}

// GetRoles returns all roles
func (pe *PolicyEngine) GetRoles() []string {
	roles, err := pe.enforcer.GetAllRoles()
	if err != nil {
		return []string{}
	}
	return roles
}

// getUserRoles gets user roles from context or database
func (pe *PolicyEngine) getUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Try to get from context first
	if roles, ok := ctx.Value("user_roles").([]string); ok {
		return roles, nil
	}

	// Fallback to enforcer
	return pe.enforcer.GetRolesForUser(userID)
}

// MemoryAdapter is a simple in-memory adapter for Casbin
type MemoryAdapter struct {
	policies [][]string
}

// LoadPolicy loads policies from storage
func (a *MemoryAdapter) LoadPolicy(model model.Model) error {
	for _, policy := range a.policies {
		model.AddPolicy("p", "p", policy)
	}
	return nil
}

// SavePolicy saves policies to storage
func (a *MemoryAdapter) SavePolicy(model model.Model) error {
	policies, err := model.GetPolicy("p", "p")
	if err != nil {
		return err
	}
	a.policies = policies
	return nil
}

// AddPolicy adds a policy rule
func (a *MemoryAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	a.policies = append(a.policies, rule)
	return nil
}

// RemovePolicy removes a policy rule
func (a *MemoryAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	for i, policy := range a.policies {
		if equalPolicy(policy, rule) {
			a.policies = append(a.policies[:i], a.policies[i+1:]...)
			break
		}
	}
	return nil
}

// RemoveFilteredPolicy removes filtered policies
func (a *MemoryAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	// Simple implementation - remove all policies matching the filter
	var filtered [][]string
	for _, policy := range a.policies {
		if !matchesFilter(policy, fieldIndex, fieldValues) {
			filtered = append(filtered, policy)
		}
	}
	a.policies = filtered
	return nil
}

// equalPolicy checks if two policies are equal
func equalPolicy(p1, p2 []string) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i, v := range p1 {
		if v != p2[i] {
			return false
		}
	}
	return true
}

// matchesFilter checks if a policy matches a filter
func matchesFilter(policy []string, fieldIndex int, fieldValues []string) bool {
	if fieldIndex >= len(policy) {
		return false
	}
	for i, value := range fieldValues {
		if fieldIndex+i >= len(policy) || policy[fieldIndex+i] != value {
			return false
		}
	}
	return true
}

// loadDefaultPolicies loads default authorization policies
func loadDefaultPolicies(enforcer *casbin.Enforcer) error {
	// Define default roles and permissions
	defaultPolicies := [][]string{
		// Admin role - can do everything
		{"admin", "posts", "create", "allow"},
		{"admin", "posts", "read", "allow"},
		{"admin", "posts", "update", "allow"},
		{"admin", "posts", "delete", "allow"},
		{"admin", "users", "create", "allow"},
		{"admin", "users", "read", "allow"},
		{"admin", "users", "update", "allow"},
		{"admin", "users", "delete", "allow"},
		{"admin", "comments", "create", "allow"},
		{"admin", "comments", "read", "allow"},
		{"admin", "comments", "update", "allow"},
		{"admin", "comments", "delete", "allow"},

		// Editor role - can manage content
		{"editor", "posts", "create", "allow"},
		{"editor", "posts", "read", "allow"},
		{"editor", "posts", "update", "allow"},
		{"editor", "posts", "delete", "allow"},
		{"editor", "comments", "create", "allow"},
		{"editor", "comments", "read", "allow"},
		{"editor", "comments", "update", "allow"},
		{"editor", "comments", "delete", "allow"},

		// User role - basic permissions
		{"user", "posts", "read", "allow"},
		{"user", "comments", "create", "allow"},
		{"user", "comments", "read", "allow"},
		{"user", "profile", "read", "allow"},
		{"user", "profile", "update", "allow"},

		// Guest role - read-only
		{"guest", "posts", "read", "allow"},
		{"guest", "comments", "read", "allow"},
	}

	// Add policies
	for _, policy := range defaultPolicies {
		_, err := enforcer.AddPolicy(policy[0], policy[1], policy[2], policy[3])
		if err != nil {
			return fmt.Errorf("failed to add default policy %v: %w", policy, err)
		}
	}

	return nil
}

// PolicyHelper provides template and controller helpers
type PolicyHelper struct {
	engine *PolicyEngine
}

// NewPolicyHelper creates a new policy helper
func NewPolicyHelper(engine *PolicyEngine) *PolicyHelper {
	return &PolicyHelper{engine: engine}
}

// Can is a template helper function
func (ph *PolicyHelper) Can(userID, action, resource string) bool {
	allowed, err := ph.engine.Can(context.Background(), userID, action, resource)
	if err != nil {
		return false
	}
	return allowed
}

// CanWithRoles checks permissions with specific roles
func (ph *PolicyHelper) CanWithRoles(roles []string, action, resource string) bool {
	policyCtx := &PolicyContext{
		UserRoles: roles,
		Resource:  resource,
		Action:    action,
	}
	allowed, err := ph.engine.CanWithContext(context.Background(), policyCtx)
	if err != nil {
		return false
	}
	return allowed
}

// HasRole checks if user has a specific role
func (ph *PolicyHelper) HasRole(userID, role string) bool {
	roles, err := ph.engine.GetUserRoles(userID)
	if err != nil {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// GetUserPermissions returns all permissions for a user
func (ph *PolicyHelper) GetUserPermissions(userID string) map[string][]string {
	permissions := make(map[string][]string)
	roles, err := ph.engine.GetUserRoles(userID)
	if err != nil {
		return permissions
	}

	// Get all policies and filter by user roles
	allPolicies := ph.engine.GetPolicies()
	for _, policy := range allPolicies {
		if len(policy) >= 3 {
			subject, resource, action := policy[0], policy[1], policy[2]
			for _, role := range roles {
				if subject == role {
					permissions[resource] = append(permissions[resource], action)
				}
			}
		}
	}

	return permissions
}
