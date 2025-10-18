package autoconfig

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/mrhoseah/dolphin/internal/container"
)

// AutoConfiguration defines the interface for auto-configuration
type AutoConfiguration interface {
	Configure(ctx context.Context, container *container.Container) error
	Order() int
	Condition() Condition
}

// Condition defines when an auto-configuration should be applied
type Condition interface {
	Matches(ctx context.Context, container *container.Container) bool
}

// OnClassCondition matches when a specific class is present
type OnClassCondition struct {
	ClassName string
}

func (c *OnClassCondition) Matches(ctx context.Context, container *container.Container) bool {
	// Check if class is available in classpath
	// This is a simplified implementation - in a real scenario,
	// you would check if the class can be loaded
	return true
}

// OnPropertyCondition matches when a property has a specific value
type OnPropertyCondition struct {
	Property string
	Value    string
}

func (c *OnPropertyCondition) Matches(ctx context.Context, container *container.Container) bool {
	// Check environment variable or config property
	// This would typically check configuration values
	return true
}

// OnMissingBeanCondition matches when a specific bean is not present
type OnMissingBeanCondition struct {
	BeanType reflect.Type
}

func (c *OnMissingBeanCondition) Matches(ctx context.Context, container *container.Container) bool {
	return !container.HasService(c.BeanType)
}

// OnConditionalOnClass creates a condition that matches when a class is present
func OnConditionalOnClass(className string) Condition {
	return &OnClassCondition{ClassName: className}
}

// OnConditionalOnProperty creates a condition that matches when a property has a value
func OnConditionalOnProperty(property, value string) Condition {
	return &OnPropertyCondition{Property: property, Value: value}
}

// OnConditionalOnMissingBean creates a condition that matches when a bean is missing
func OnConditionalOnMissingBean(beanType reflect.Type) Condition {
	return &OnMissingBeanCondition{BeanType: beanType}
}

// DatabaseAutoConfiguration automatically configures database services
type DatabaseAutoConfiguration struct{}

func (c *DatabaseAutoConfiguration) Configure(ctx context.Context, container *container.Container) error {
	// Auto-configure database connection
	// This would typically create and register database services
	// based on configuration properties

	// Example: Register database service
	// container.RegisterSingleton(
	//     reflect.TypeOf((*sql.DB)(nil)),
	//     reflect.TypeOf((*sql.DB)(nil)),
	// )

	// Example: Register repositories
	// container.RegisterScoped(
	//     reflect.TypeOf((*UserRepository)(nil)).Elem(),
	//     reflect.TypeOf((*userRepositoryImpl)(nil)),
	// )

	return nil
}

func (c *DatabaseAutoConfiguration) Order() int {
	return 100 // Lower numbers have higher priority
}

func (c *DatabaseAutoConfiguration) Condition() Condition {
	return OnConditionalOnProperty("dolphin.datasource.url", "")
}

// CacheAutoConfiguration automatically configures caching services
type CacheAutoConfiguration struct{}

func (c *CacheAutoConfiguration) Configure(ctx context.Context, container *container.Container) error {
	// Auto-configure Redis cache if available
	// container.RegisterSingleton(
	//     reflect.TypeOf((*Cache)(nil)).Elem(),
	//     reflect.TypeOf((*redisCache)(nil)),
	// )

	return nil
}

func (c *CacheAutoConfiguration) Order() int {
	return 200
}

func (c *CacheAutoConfiguration) Condition() Condition {
	return OnConditionalOnProperty("dolphin.cache.type", "redis")
}

// SecurityAutoConfiguration automatically configures security services
type SecurityAutoConfiguration struct{}

func (c *SecurityAutoConfiguration) Configure(ctx context.Context, container *container.Container) error {
	// Auto-configure security services
	// container.RegisterSingleton(
	//     reflect.TypeOf((*SecurityManager)(nil)).Elem(),
	//     reflect.TypeOf((*securityManagerImpl)(nil)),
	// )

	return nil
}

func (c *SecurityAutoConfiguration) Order() int {
	return 50
}

func (c *SecurityAutoConfiguration) Condition() Condition {
	return OnConditionalOnMissingBean(reflect.TypeOf((*SecurityManager)(nil)).Elem())
}

// WebAutoConfiguration automatically configures web services
type WebAutoConfiguration struct{}

func (c *WebAutoConfiguration) Configure(ctx context.Context, container *container.Container) error {
	// Auto-configure web services
	// container.RegisterSingleton(
	//     reflect.TypeOf((*Router)(nil)).Elem(),
	//     reflect.TypeOf((*routerImpl)(nil)),
	// )

	return nil
}

func (c *WebAutoConfiguration) Order() int {
	return 300
}

func (c *WebAutoConfiguration) Condition() Condition {
	return OnConditionalOnClass("net/http")
}

// AutoConfigurationManager manages auto-configurations
type AutoConfigurationManager struct {
	configurations []AutoConfiguration
}

func NewAutoConfigurationManager() *AutoConfigurationManager {
	return &AutoConfigurationManager{
		configurations: []AutoConfiguration{
			&SecurityAutoConfiguration{},
			&DatabaseAutoConfiguration{},
			&CacheAutoConfiguration{},
			&WebAutoConfiguration{},
		},
	}
}

func (m *AutoConfigurationManager) AddConfiguration(config AutoConfiguration) {
	m.configurations = append(m.configurations, config)
}

func (m *AutoConfigurationManager) ApplyAutoConfiguration(ctx context.Context, container *container.Container) error {
	// Sort configurations by order
	sort.Slice(m.configurations, func(i, j int) bool {
		return m.configurations[i].Order() < m.configurations[j].Order()
	})

	// Apply configurations that match their conditions
	for _, config := range m.configurations {
		if config.Condition().Matches(ctx, container) {
			if err := config.Configure(ctx, container); err != nil {
				return fmt.Errorf("failed to apply auto-configuration %T: %w", config, err)
			}
		}
	}

	return nil
}

// ConfigurationProperties holds configuration properties for auto-configuration
type ConfigurationProperties struct {
	Database DatabaseProperties `yaml:"database"`
	Cache    CacheProperties    `yaml:"cache"`
	Security SecurityProperties `yaml:"security"`
	Web      WebProperties      `yaml:"web"`
}

type DatabaseProperties struct {
	URL      string `yaml:"url"`
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type CacheProperties struct {
	Type string `yaml:"type"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type SecurityProperties struct {
	JWTSecret string `yaml:"jwt_secret"`
	Enabled   bool   `yaml:"enabled"`
}

type WebProperties struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// ConfigurationCondition checks configuration properties
type ConfigurationCondition struct {
	Properties *ConfigurationProperties
	Check      func(props *ConfigurationProperties) bool
}

func (c *ConfigurationCondition) Matches(ctx context.Context, container *container.Container) bool {
	return c.Check(c.Properties)
}

// OnConditionalOnConfiguration creates a condition based on configuration
func OnConditionalOnConfiguration(props *ConfigurationProperties, check func(*ConfigurationProperties) bool) Condition {
	return &ConfigurationCondition{
		Properties: props,
		Check:      check,
	}
}

// DefaultAutoConfigurations returns the default set of auto-configurations
func DefaultAutoConfigurations() []AutoConfiguration {
	return []AutoConfiguration{
		&SecurityAutoConfiguration{},
		&DatabaseAutoConfiguration{},
		&CacheAutoConfiguration{},
		&WebAutoConfiguration{},
	}
}
