package container

import (
	"fmt"
	"reflect"
	"sync"
)

// ServiceDescriptor defines how a service should be created
type ServiceDescriptor struct {
	ServiceType     reflect.Type
	Implementation  reflect.Type
	Instance        interface{}
	Lifetime        ServiceLifetime
	Factory         func(container *Container) (interface{}, error)
}

type ServiceLifetime int

const (
	Singleton ServiceLifetime = iota
	Scoped
	Transient
)

// Container manages service registration and resolution
type Container struct {
	services map[reflect.Type]*ServiceDescriptor
	scoped   map[string]interface{}
	mu       sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		services: make(map[reflect.Type]*ServiceDescriptor),
		scoped:   make(map[string]interface{}),
	}
}

// RegisterSingleton registers a singleton service
func (c *Container) RegisterSingleton(serviceType, implementation reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[serviceType] = &ServiceDescriptor{
		ServiceType:    serviceType,
		Implementation: implementation,
		Lifetime:       Singleton,
	}
}

// RegisterScoped registers a scoped service
func (c *Container) RegisterScoped(serviceType, implementation reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[serviceType] = &ServiceDescriptor{
		ServiceType:    serviceType,
		Implementation: implementation,
		Lifetime:       Scoped,
	}
}

// RegisterTransient registers a transient service
func (c *Container) RegisterTransient(serviceType, implementation reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[serviceType] = &ServiceDescriptor{
		ServiceType:    serviceType,
		Implementation: implementation,
		Lifetime:       Transient,
	}
}

// RegisterFactory registers a service with a custom factory
func (c *Container) RegisterFactory(serviceType reflect.Type, factory func(*Container) (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[serviceType] = &ServiceDescriptor{
		ServiceType: serviceType,
		Lifetime:    Transient,
		Factory:     factory,
	}
}

// Resolve resolves a service from the container
func (c *Container) Resolve(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	descriptor, exists := c.services[serviceType]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service %v not registered", serviceType)
	}

	switch descriptor.Lifetime {
	case Singleton:
		return c.resolveSingleton(descriptor)
	case Scoped:
		return c.resolveScoped(descriptor)
	case Transient:
		return c.resolveTransient(descriptor)
	default:
		return nil, fmt.Errorf("unknown service lifetime")
	}
}

func (c *Container) resolveSingleton(descriptor *ServiceDescriptor) (interface{}, error) {
	if descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	instance, err := c.createInstance(descriptor)
	if err != nil {
		return nil, err
	}

	descriptor.Instance = instance
	return instance, nil
}

func (c *Container) resolveScoped(descriptor *ServiceDescriptor) (interface{}, error) {
	scopeKey := descriptor.ServiceType.String()

	if instance, exists := c.scoped[scopeKey]; exists {
		return instance, nil
	}

	instance, err := c.createInstance(descriptor)
	if err != nil {
		return nil, err
	}

	c.scoped[scopeKey] = instance
	return instance, nil
}

func (c *Container) resolveTransient(descriptor *ServiceDescriptor) (interface{}, error) {
	return c.createInstance(descriptor)
}

func (c *Container) createInstance(descriptor *ServiceDescriptor) (interface{}, error) {
	if descriptor.Factory != nil {
		return descriptor.Factory(c)
	}

	// Use reflection to create instance and inject dependencies
	return c.createInstanceWithDependencies(descriptor.Implementation)
}

func (c *Container) createInstanceWithDependencies(implType reflect.Type) (interface{}, error) {
	// Get the constructor (assuming it's the first method or a specific constructor)
	if implType.Kind() == reflect.Ptr {
		implType = implType.Elem()
	}

	// Create new instance
	instance := reflect.New(implType).Interface()

	// Try to find and call constructor method
	constructorMethod := reflect.ValueOf(instance).MethodByName("New")
	if constructorMethod.IsValid() {
		// Analyze constructor parameters and inject dependencies
		constructorType := constructorMethod.Type()
		args := make([]reflect.Value, constructorType.NumIn())

		for i := 0; i < constructorType.NumIn(); i++ {
			paramType := constructorType.In(i)
			arg, err := c.Resolve(paramType)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve dependency %v: %w", paramType, err)
			}
			args[i] = reflect.ValueOf(arg)
		}

		// Call constructor
		results := constructorMethod.Call(args)
		if len(results) > 0 {
			if err, ok := results[len(results)-1].Interface().(error); ok && err != nil {
				return nil, err
			}
		}
	}

	return instance, nil
}

// ClearScoped clears all scoped instances (useful for testing)
func (c *Container) ClearScoped() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scoped = make(map[string]interface{})
}

// GetRegisteredServices returns all registered service types
func (c *Container) GetRegisteredServices() []reflect.Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	services := make([]reflect.Type, 0, len(c.services))
	for serviceType := range c.services {
		services = append(services, serviceType)
	}
	return services
}

// HasService checks if a service is registered
func (c *Container) HasService(serviceType reflect.Type) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.services[serviceType]
	return exists
}
