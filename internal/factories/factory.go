package factories

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"gorm.io/gorm"
)

// Factory defines the interface for model factories
type Factory interface {
	// Create creates a single instance of the model
	Create(attributes ...map[string]interface{}) interface{}
	
	// CreateMany creates multiple instances of the model
	CreateMany(count int, attributes ...map[string]interface{}) []interface{}
	
	// Make creates an instance without persisting to database
	Make(attributes ...map[string]interface{}) interface{}
	
	// MakeMany creates multiple instances without persisting
	MakeMany(count int, attributes ...map[string]interface{}) []interface{}
	
	// GetModel returns the model type
	GetModel() interface{}
}

// BaseFactory provides common factory functionality
type BaseFactory struct {
	model     interface{}
	db        *gorm.DB
	callbacks []func(interface{})
	states    map[string][]func(interface{})
}

// NewFactory creates a new factory instance
func NewFactory(model interface{}, db *gorm.DB) *BaseFactory {
	return &BaseFactory{
		model:     model,
		db:        db,
		callbacks: make([]func(interface{}), 0),
		states:    make(map[string][]func(interface{})),
	}
}

// Create creates a single instance and persists to database
func (f *BaseFactory) Create(attributes ...map[string]interface{}) interface{} {
	instance := f.Make(attributes...)
	
	// Apply callbacks
	for _, callback := range f.callbacks {
		callback(instance)
	}
	
	// Persist to database
	if err := f.db.Create(instance).Error; err != nil {
		panic(fmt.Sprintf("Failed to create model: %v", err))
	}
	
	return instance
}

// CreateMany creates multiple instances and persists to database
func (f *BaseFactory) CreateMany(count int, attributes ...map[string]interface{}) []interface{} {
	instances := make([]interface{}, count)
	
	for i := 0; i < count; i++ {
		instances[i] = f.Create(attributes...)
	}
	
	return instances
}

// Make creates an instance without persisting
func (f *BaseFactory) Make(attributes ...map[string]interface{}) interface{} {
	instance := f.createInstance()
	
	// Apply default attributes
	f.applyAttributes(instance, f.getDefaultAttributes())
	
	// Apply provided attributes
	if len(attributes) > 0 {
		f.applyAttributes(instance, attributes[0])
	}
	
	return instance
}

// MakeMany creates multiple instances without persisting
func (f *BaseFactory) MakeMany(count int, attributes ...map[string]interface{}) []interface{} {
	instances := make([]interface{}, count)
	
	for i := 0; i < count; i++ {
		instances[i] = f.Make(attributes...)
	}
	
	return instances
}

// GetModel returns the model type
func (f *BaseFactory) GetModel() interface{} {
	return f.model
}

// State defines a factory state
func (f *BaseFactory) State(name string, callback func(interface{})) *BaseFactory {
	if f.states[name] == nil {
		f.states[name] = make([]func(interface{}), 0)
	}
	f.states[name] = append(f.states[name], callback)
	return f
}

// AfterCreating adds a callback to run after creation
func (f *BaseFactory) AfterCreating(callback func(interface{})) *BaseFactory {
	f.callbacks = append(f.callbacks, callback)
	return f
}

// createInstance creates a new instance of the model
func (f *BaseFactory) createInstance() interface{} {
	modelType := reflect.TypeOf(f.model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	return reflect.New(modelType).Interface()
}

// applyAttributes applies attributes to an instance
func (f *BaseFactory) applyAttributes(instance interface{}, attributes map[string]interface{}) {
	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}
	
	for key, value := range attributes {
		field := instanceValue.FieldByName(key)
		if field.IsValid() && field.CanSet() {
			fieldValue := reflect.ValueOf(value)
			if fieldValue.Type().AssignableTo(field.Type()) {
				field.Set(fieldValue)
			}
		}
	}
}

// getDefaultAttributes returns default attributes for the model
func (f *BaseFactory) getDefaultAttributes() map[string]interface{} {
	// This will be overridden by specific factory implementations
	return make(map[string]interface{})
}

// FactoryManager manages all factories
type FactoryManager struct {
	factories map[string]Factory
	db        *gorm.DB
}

// NewFactoryManager creates a new factory manager
func NewFactoryManager(db *gorm.DB) *FactoryManager {
	return &FactoryManager{
		factories: make(map[string]Factory),
		db:        db,
	}
}

// Register registers a factory
func (fm *FactoryManager) Register(name string, factory Factory) {
	fm.factories[name] = factory
}

// Get retrieves a factory by name
func (fm *FactoryManager) Get(name string) Factory {
	return fm.factories[name]
}

// Create creates an instance using the specified factory
func (fm *FactoryManager) Create(factoryName string, attributes ...map[string]interface{}) interface{} {
	factory := fm.Get(factoryName)
	if factory == nil {
		panic(fmt.Sprintf("Factory '%s' not found", factoryName))
	}
	return factory.Create(attributes...)
}

// CreateMany creates multiple instances using the specified factory
func (fm *FactoryManager) CreateMany(factoryName string, count int, attributes ...map[string]interface{}) []interface{} {
	factory := fm.Get(factoryName)
	if factory == nil {
		panic(fmt.Sprintf("Factory '%s' not found", factoryName))
	}
	return factory.CreateMany(count, attributes...)
}

// Make creates an instance without persisting using the specified factory
func (fm *FactoryManager) Make(factoryName string, attributes ...map[string]interface{}) interface{} {
	factory := fm.Get(factoryName)
	if factory == nil {
		panic(fmt.Sprintf("Factory '%s' not found", factoryName))
	}
	return factory.Make(attributes...)
}

// MakeMany creates multiple instances without persisting using the specified factory
func (fm *FactoryManager) MakeMany(factoryName string, count int, attributes ...map[string]interface{}) []interface{} {
	factory := fm.Get(factoryName)
	if factory == nil {
		panic(fmt.Sprintf("Factory '%s' not found", factoryName))
	}
	return factory.MakeMany(count, attributes...)
}

// FakeData provides fake data generation utilities
type FakeData struct {
	faker *gofakeit.Faker
}

// NewFakeData creates a new fake data generator
func NewFakeData() *FakeData {
	return &FakeData{
		faker: gofakeit.New(0), // Use time-based seed
	}
}

// Name generates a fake name
func (fd *FakeData) Name() string {
	return fd.faker.Name()
}

// Email generates a fake email
func (fd *FakeData) Email() string {
	return fd.faker.Email()
}

// Username generates a fake username
func (fd *FakeData) Username() string {
	return fd.faker.Username()
}

// Password generates a fake password
func (fd *FakeData) Password() string {
	return fd.faker.Password(true, true, true, true, false, 12)
}

// Phone generates a fake phone number
func (fd *FakeData) Phone() string {
	return fd.faker.Phone()
}

// Address generates a fake address
func (fd *FakeData) Address() string {
	return fd.faker.Address().Address
}

// City generates a fake city
func (fd *FakeData) City() string {
	return fd.faker.Address().City
}

// Country generates a fake country
func (fd *FakeData) Country() string {
	return fd.faker.Address().Country
}

// Company generates a fake company name
func (fd *FakeData) Company() string {
	return fd.faker.Company()
}

// JobTitle generates a fake job title
func (fd *FakeData) JobTitle() string {
	return fd.faker.JobTitle()
}

// Sentence generates a fake sentence
func (fd *FakeData) Sentence() string {
	return fd.faker.Sentence(10)
}

// Paragraph generates a fake paragraph
func (fd *FakeData) Paragraph() string {
	return fd.faker.Paragraph(3, 5, 10, " ")
}

// URL generates a fake URL
func (fd *FakeData) URL() string {
	return fd.faker.URL()
}

// ImageURL generates a fake image URL
func (fd *FakeData) ImageURL() string {
	return fd.faker.ImageURL(800, 600)
}

// Price generates a fake price
func (fd *FakeData) Price(min, max float64) float64 {
	return fd.faker.Price(min, max)
}

// Number generates a fake number within range
func (fd *FakeData) Number(min, max int) int {
	return fd.faker.Number(min, max)
}

// Float generates a fake float within range
func (fd *FakeData) Float(min, max float64) float64 {
	return fd.faker.Float64Range(min, max)
}

// Bool generates a fake boolean
func (fd *FakeData) Bool() bool {
	return fd.faker.Bool()
}

// Date generates a fake date
func (fd *FakeData) Date() time.Time {
	return fd.faker.Date()
}

// DateRange generates a fake date within range
func (fd *FakeData) DateRange(start, end time.Time) time.Time {
	return fd.faker.DateRange(start, end)
}

// UUID generates a fake UUID
func (fd *FakeData) UUID() string {
	return fd.faker.UUID()
}

// Lorem generates lorem ipsum text
func (fd *FakeData) Lorem(wordCount int) string {
	return fd.faker.LoremIpsumWord(wordCount)
}

// RandomElement selects a random element from a slice
func (fd *FakeData) RandomElement(slice interface{}) interface{} {
	return fd.faker.RandomString([]string{"a", "b", "c"}) // Simplified for now
}

// Sequence generates sequential data
type Sequence struct {
	counter int
	prefix  string
}

// NewSequence creates a new sequence generator
func NewSequence(prefix string) *Sequence {
	return &Sequence{
		counter: 0,
		prefix:  prefix,
	}
}

// Next generates the next sequence value
func (s *Sequence) Next() string {
	s.counter++
	return fmt.Sprintf("%s%d", s.prefix, s.counter)
}

// Reset resets the sequence counter
func (s *Sequence) Reset() {
	s.counter = 0
}

// Set sets the sequence counter to a specific value
func (s *Sequence) Set(value int) {
	s.counter = value
}
