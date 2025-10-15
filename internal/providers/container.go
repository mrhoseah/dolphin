package providers

import (
	"fmt"
	"sort"
	"sync"
)

// ServiceContainer manages service providers and their instances
type ServiceContainer struct {
	providers map[string]ServiceProvider
	services  map[string]interface{}
	mutex     sync.RWMutex
}

// NewServiceContainer creates a new service container
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		providers: make(map[string]ServiceProvider),
		services:  make(map[string]interface{}),
	}
}

// RegisterProvider registers a service provider
func (c *ServiceContainer) RegisterProvider(provider ServiceProvider) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	name := provider.Name()
	if _, exists := c.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	c.providers[name] = provider
	return nil
}

// BootProviders boots all registered providers
func (c *ServiceContainer) BootProviders() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Sort providers by priority
	var sortedProviders []ServiceProvider
	for _, provider := range c.providers {
		sortedProviders = append(sortedProviders, provider)
	}

	sort.Slice(sortedProviders, func(i, j int) bool {
		return sortedProviders[i].Priority() < sortedProviders[j].Priority()
	})

	// Register services
	for _, provider := range sortedProviders {
		if err := provider.Register(); err != nil {
			return fmt.Errorf("failed to register provider %s: %w", provider.Name(), err)
		}
	}

	// Boot services
	for _, provider := range sortedProviders {
		if err := provider.Boot(); err != nil {
			return fmt.Errorf("failed to boot provider %s: %w", provider.Name(), err)
		}
	}

	return nil
}

// Bind binds a service to the container
func (c *ServiceContainer) Bind(name string, service interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.services[name] = service
}

// Get retrieves a service from the container
func (c *ServiceContainer) Get(name string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// MustGet retrieves a service or panics
func (c *ServiceContainer) MustGet(name string) interface{} {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

// Has checks if a service exists
func (c *ServiceContainer) Has(name string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, exists := c.services[name]
	return exists
}

// GetEmailProvider gets the email provider
func (c *ServiceContainer) GetEmailProvider() EmailProvider {
	return c.MustGet("email").(EmailProvider)
}

// GetNotificationProvider gets the notification provider
func (c *ServiceContainer) GetNotificationProvider() NotificationProvider {
	return c.MustGet("notification").(NotificationProvider)
}

// GetStorageProvider gets the storage provider
func (c *ServiceContainer) GetStorageProvider() StorageProvider {
	return c.MustGet("storage").(StorageProvider)
}

// GetCacheProvider gets the cache provider
func (c *ServiceContainer) GetCacheProvider() CacheProvider {
	return c.MustGet("cache").(CacheProvider)
}

// GetQueueProvider gets the queue provider
func (c *ServiceContainer) GetQueueProvider() QueueProvider {
	return c.MustGet("queue").(QueueProvider)
}

// GetSearchProvider gets the search provider
func (c *ServiceContainer) GetSearchProvider() SearchProvider {
	return c.MustGet("search").(SearchProvider)
}

// GetPaymentProvider gets the payment provider
func (c *ServiceContainer) GetPaymentProvider() PaymentProvider {
	return c.MustGet("payment").(PaymentProvider)
}

// GetSMSProvider gets the SMS provider
func (c *ServiceContainer) GetSMSProvider() SMSProvider {
	return c.MustGet("sms").(SMSProvider)
}

// GetSocialProvider gets the social provider
func (c *ServiceContainer) GetSocialProvider() SocialProvider {
	return c.MustGet("social").(SocialProvider)
}

// GetAnalyticsProvider gets the analytics provider
func (c *ServiceContainer) GetAnalyticsProvider() AnalyticsProvider {
	return c.MustGet("analytics").(AnalyticsProvider)
}

// GetLogProvider gets the log provider
func (c *ServiceContainer) GetLogProvider() LogProvider {
	return c.MustGet("log").(LogProvider)
}

// GetConfigProvider gets the config provider
func (c *ServiceContainer) GetConfigProvider() ConfigProvider {
	return c.MustGet("config").(ConfigProvider)
}

// GetDatabaseProvider gets the database provider
func (c *ServiceContainer) GetDatabaseProvider() DatabaseProvider {
	return c.MustGet("database").(DatabaseProvider)
}

// GetSecurityProvider gets the security provider
func (c *ServiceContainer) GetSecurityProvider() SecurityProvider {
	return c.MustGet("security").(SecurityProvider)
}

// GetMonitoringProvider gets the monitoring provider
func (c *ServiceContainer) GetMonitoringProvider() MonitoringProvider {
	return c.MustGet("monitoring").(MonitoringProvider)
}

// ProviderManager manages service providers
type ProviderManager struct {
	container *ServiceContainer
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		container: NewServiceContainer(),
	}
}

// Register registers a provider
func (m *ProviderManager) Register(provider ServiceProvider) error {
	return m.container.RegisterProvider(provider)
}

// Boot boots all providers
func (m *ProviderManager) Boot() error {
	return m.container.BootProviders()
}

// Container returns the service container
func (m *ProviderManager) Container() *ServiceContainer {
	return m.container
}

// DefaultProviders returns a list of default providers
func DefaultProviders() []ServiceProvider {
	return []ServiceProvider{
		NewConfigProvider(),
		NewLogProvider(),
		NewSecurityProvider(),
		NewEmailProvider(),
		NewStorageProvider(),
		NewCacheProvider(),
		NewNotificationProvider(),
		NewQueueProvider(),
		NewSearchProvider(),
		NewPaymentProvider(),
		NewSMSProvider(),
		NewSocialProvider(),
		NewAnalyticsProvider(),
		NewDatabaseProvider(),
		NewMonitoringProvider(),
	}
}
