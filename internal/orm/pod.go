package orm

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Model represents a base model with common functionality
type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// SoftDeletes adds soft delete functionality
type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Timestamps adds timestamp functionality
type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Relationship represents a model relationship
type Relationship struct {
	Type         string
	Model        interface{}
	ForeignKey   string
	LocalKey     string
	PivotTable   string
	ForeignValue interface{}
	LocalValue   interface{}
}

// HasMany represents a has-many relationship
type HasMany struct {
	Model      interface{}
	ForeignKey string
	LocalKey   string
}

// BelongsTo represents a belongs-to relationship
type BelongsTo struct {
	Model      interface{}
	ForeignKey string
	LocalKey   string
}

// BelongsToMany represents a many-to-many relationship
type BelongsToMany struct {
	Model      interface{}
	PivotTable string
	ForeignKey string
	LocalKey   string
}

// HasOne represents a has-one relationship
type HasOne struct {
	Model      interface{}
	ForeignKey string
	LocalKey   string
}

// Scope represents a query scope
type Scope func(*gorm.DB) *gorm.DB

// Repository represents a model repository
type Repository struct {
	db     *gorm.DB
	model  interface{}
	scopes map[string]Scope
}

// NewRepository creates a new repository
func NewRepository(db *gorm.DB, model interface{}) *Repository {
	return &Repository{
		db:     db,
		model:  model,
		scopes: make(map[string]Scope),
	}
}

// All retrieves all records
func (r *Repository) All() (interface{}, error) {
	var results interface{}
	err := r.db.Find(&results).Error
	return results, err
}

// Find finds a record by ID
func (r *Repository) Find(id interface{}) (interface{}, error) {
	var result interface{}
	err := r.db.First(&result, id).Error
	return result, err
}

// Where adds a where condition
func (r *Repository) Where(query interface{}, args ...interface{}) *Repository {
	return &Repository{
		db:     r.db.Where(query, args...),
		model:  r.model,
		scopes: r.scopes,
	}
}

// With applies a scope
func (r *Repository) With(scopeName string) *Repository {
	if scope, exists := r.scopes[scopeName]; exists {
		return &Repository{
			db:     scope(r.db),
			model:  r.model,
			scopes: r.scopes,
		}
	}
	return r
}

// WithRelationships loads relationships
func (r *Repository) WithRelationships(relationships ...string) *Repository {
	return &Repository{
		db:     r.db.Preload(strings.Join(relationships, " ")),
		model:  r.model,
		scopes: r.scopes,
	}
}

// Order adds ordering
func (r *Repository) Order(value interface{}) *Repository {
	return &Repository{
		db:     r.db.Order(value),
		model:  r.model,
		scopes: r.scopes,
	}
}

// Limit adds a limit
func (r *Repository) Limit(limit int) *Repository {
	return &Repository{
		db:     r.db.Limit(limit),
		model:  r.model,
		scopes: r.scopes,
	}
}

// Offset adds an offset
func (r *Repository) Offset(offset int) *Repository {
	return &Repository{
		db:     r.db.Offset(offset),
		model:  r.model,
		scopes: r.scopes,
	}
}

// Create creates a new record
func (r *Repository) Create(data interface{}) error {
	return r.db.Create(data).Error
}

// Update updates a record
func (r *Repository) Update(id interface{}, data interface{}) error {
	return r.db.Model(r.model).Where("id = ?", id).Updates(data).Error
}

// Delete deletes a record
func (r *Repository) Delete(id interface{}) error {
	return r.db.Delete(r.model, id).Error
}

// SoftDelete soft deletes a record
func (r *Repository) SoftDelete(id interface{}) error {
	return r.db.Delete(r.model, id).Error
}

// Restore restores a soft deleted record
func (r *Repository) Restore(id interface{}) error {
	return r.db.Unscoped().Model(r.model).Where("id = ?", id).Update("deleted_at", nil).Error
}

// Count counts records
func (r *Repository) Count() (int64, error) {
	var count int64
	err := r.db.Model(r.model).Count(&count).Error
	return count, err
}

// Exists checks if a record exists
func (r *Repository) Exists(id interface{}) bool {
	var count int64
	r.db.Model(r.model).Where("id = ?", id).Count(&count)
	return count > 0
}

// AddScope adds a query scope
func (r *Repository) AddScope(name string, scope Scope) {
	r.scopes[name] = scope
}

// GetDB returns the underlying database connection
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// PodModel represents a Pod-style model
type PodModel struct {
	Model
	relationships map[string]Relationship
	scopes        map[string]Scope
	accessors     map[string]func(interface{}) interface{}
	mutators      map[string]func(interface{}) interface{}
	events        map[string][]func(interface{})
	db            *gorm.DB
}

// NewPodModel creates a new Pod model
func NewPodModel(db *gorm.DB) *PodModel {
	return &PodModel{
		relationships: make(map[string]Relationship),
		scopes:        make(map[string]Scope),
		accessors:     make(map[string]func(interface{}) interface{}),
		mutators:      make(map[string]func(interface{}) interface{}),
		events:        make(map[string][]func(interface{})),
		db:            db,
	}
}

// HasMany defines a has-many relationship
func (pm *PodModel) HasMany(related interface{}, foreignKey, localKey string) Relationship {
	return Relationship{
		Type:       "hasMany",
		Model:      related,
		ForeignKey: foreignKey,
		LocalKey:   localKey,
	}
}

// BelongsTo defines a belongs-to relationship
func (pm *PodModel) BelongsTo(related interface{}, foreignKey, localKey string) Relationship {
	return Relationship{
		Type:       "belongsTo",
		Model:      related,
		ForeignKey: foreignKey,
		LocalKey:   localKey,
	}
}

// BelongsToMany defines a many-to-many relationship
func (pm *PodModel) BelongsToMany(related interface{}, pivotTable, foreignKey, localKey string) Relationship {
	return Relationship{
		Type:       "belongsToMany",
		Model:      related,
		PivotTable: pivotTable,
		ForeignKey: foreignKey,
		LocalKey:   localKey,
	}
}

// HasOne defines a has-one relationship
func (pm *PodModel) HasOne(related interface{}, foreignKey, localKey string) Relationship {
	return Relationship{
		Type:       "hasOne",
		Model:      related,
		ForeignKey: foreignKey,
		LocalKey:   localKey,
	}
}

// AddScope adds a query scope
func (pm *PodModel) AddScope(name string, scope Scope) {
	pm.scopes[name] = scope
}

// AddAccessor adds an accessor
func (pm *PodModel) AddAccessor(attribute string, accessor func(interface{}) interface{}) {
	pm.accessors[attribute] = accessor
}

// AddMutator adds a mutator
func (pm *PodModel) AddMutator(attribute string, mutator func(interface{}) interface{}) {
	pm.mutators[attribute] = mutator
}

// AddEvent adds an event listener
func (pm *PodModel) AddEvent(event string, listener func(interface{})) {
	pm.events[event] = append(pm.events[event], listener)
}

// FireEvent fires an event
func (pm *PodModel) FireEvent(event string, model interface{}) {
	if listeners, exists := pm.events[event]; exists {
		for _, listener := range listeners {
			listener(model)
		}
	}
}

// QueryBuilder represents a fluent query builder
type QueryBuilder struct {
	db            *gorm.DB
	model         interface{}
	relationships []string
	scopes        []string
	conditions    []Condition
	orders        []string
	limit         int
	offset        int
}

// Condition represents a query condition
type Condition struct {
	Column   string
	Operator string
	Value    interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB, model interface{}) *QueryBuilder {
	return &QueryBuilder{
		db:            db,
		model:         model,
		relationships: make([]string, 0),
		scopes:        make([]string, 0),
		conditions:    make([]Condition, 0),
		orders:        make([]string, 0),
	}
}

// Where adds a where condition
func (qb *QueryBuilder) Where(column string, operator string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// WhereEqual adds an equality condition
func (qb *QueryBuilder) WhereEqual(column string, value interface{}) *QueryBuilder {
	return qb.Where(column, "=", value)
}

// WhereIn adds an IN condition
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	return qb.Where(column, "IN", values)
}

// WhereBetween adds a BETWEEN condition
func (qb *QueryBuilder) WhereBetween(column string, start, end interface{}) *QueryBuilder {
	return qb.Where(column, "BETWEEN", []interface{}{start, end})
}

// WhereNull adds a NULL condition
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	return qb.Where(column, "IS NULL", nil)
}

// WhereNotNull adds a NOT NULL condition
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	return qb.Where(column, "IS NOT NULL", nil)
}

// With adds a relationship to load
func (qb *QueryBuilder) With(relationship string) *QueryBuilder {
	qb.relationships = append(qb.relationships, relationship)
	return qb
}

// Scope applies a scope
func (qb *QueryBuilder) Scope(scopeName string) *QueryBuilder {
	qb.scopes = append(qb.scopes, scopeName)
	return qb
}

// OrderBy adds ordering
func (qb *QueryBuilder) OrderBy(column string, direction string) *QueryBuilder {
	if direction == "" {
		direction = "ASC"
	}
	qb.orders = append(qb.orders, fmt.Sprintf("%s %s", column, strings.ToUpper(direction)))
	return qb
}

// Limit sets the limit
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the offset
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Get executes the query and returns results
func (qb *QueryBuilder) Get() (interface{}, error) {
	query := qb.db.Model(qb.model)

	// Apply conditions
	for _, condition := range qb.conditions {
		if condition.Operator == "IN" {
			query = query.Where(fmt.Sprintf("%s IN ?", condition.Column), condition.Value)
		} else if condition.Operator == "BETWEEN" {
			values := condition.Value.([]interface{})
			query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", condition.Column), values[0], values[1])
		} else if condition.Operator == "IS NULL" {
			query = query.Where(fmt.Sprintf("%s IS NULL", condition.Column))
		} else if condition.Operator == "IS NOT NULL" {
			query = query.Where(fmt.Sprintf("%s IS NOT NULL", condition.Column))
		} else {
			query = query.Where(fmt.Sprintf("%s %s ?", condition.Column, condition.Operator), condition.Value)
		}
	}

	// Apply relationships
	if len(qb.relationships) > 0 {
		query = query.Preload(strings.Join(qb.relationships, " "))
	}

	// Apply ordering
	if len(qb.orders) > 0 {
		query = query.Order(strings.Join(qb.orders, ", "))
	}

	// Apply limit and offset
	if qb.limit > 0 {
		query = query.Limit(qb.limit)
	}
	if qb.offset > 0 {
		query = query.Offset(qb.offset)
	}

	var results interface{}
	err := query.Find(&results).Error
	return results, err
}

// First executes the query and returns the first result
func (qb *QueryBuilder) First() (interface{}, error) {
	query := qb.db.Model(qb.model)

	// Apply conditions (same as Get)
	for _, condition := range qb.conditions {
		if condition.Operator == "IN" {
			query = query.Where(fmt.Sprintf("%s IN ?", condition.Column), condition.Value)
		} else if condition.Operator == "BETWEEN" {
			values := condition.Value.([]interface{})
			query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", condition.Column), values[0], values[1])
		} else if condition.Operator == "IS NULL" {
			query = query.Where(fmt.Sprintf("%s IS NULL", condition.Column))
		} else if condition.Operator == "IS NOT NULL" {
			query = query.Where(fmt.Sprintf("%s IS NOT NULL", condition.Column))
		} else {
			query = query.Where(fmt.Sprintf("%s %s ?", condition.Column, condition.Operator), condition.Value)
		}
	}

	// Apply relationships
	if len(qb.relationships) > 0 {
		query = query.Preload(strings.Join(qb.relationships, " "))
	}

	// Apply ordering
	if len(qb.orders) > 0 {
		query = query.Order(strings.Join(qb.orders, ", "))
	}

	var result interface{}
	err := query.First(&result).Error
	return result, err
}

// Count counts the results
func (qb *QueryBuilder) Count() (int64, error) {
	query := qb.db.Model(qb.model)

	// Apply conditions
	for _, condition := range qb.conditions {
		if condition.Operator == "IN" {
			query = query.Where(fmt.Sprintf("%s IN ?", condition.Column), condition.Value)
		} else if condition.Operator == "BETWEEN" {
			values := condition.Value.([]interface{})
			query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", condition.Column), values[0], values[1])
		} else if condition.Operator == "IS NULL" {
			query = query.Where(fmt.Sprintf("%s IS NULL", condition.Column))
		} else if condition.Operator == "IS NOT NULL" {
			query = query.Where(fmt.Sprintf("%s IS NOT NULL", condition.Column))
		} else {
			query = query.Where(fmt.Sprintf("%s %s ?", condition.Column, condition.Operator), condition.Value)
		}
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// Exists checks if any records exist
func (qb *QueryBuilder) Exists() bool {
	count, err := qb.Count()
	return err == nil && count > 0
}

// Paginate paginates the results
func (qb *QueryBuilder) Paginate(page, perPage int) (interface{}, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 15
	}

	offset := (page - 1) * perPage
	qb.Offset(offset).Limit(perPage)

	return qb.Get()
}

// ModelManager manages models and their relationships
type ModelManager struct {
	db           *gorm.DB
	models       map[string]interface{}
	repositories map[string]*Repository
}

// NewModelManager creates a new model manager
func NewModelManager(db *gorm.DB) *ModelManager {
	return &ModelManager{
		db:           db,
		models:       make(map[string]interface{}),
		repositories: make(map[string]*Repository),
	}
}

// RegisterModel registers a model
func (mm *ModelManager) RegisterModel(name string, model interface{}) {
	mm.models[name] = model
	mm.repositories[name] = NewRepository(mm.db, model)
}

// GetRepository returns a repository for a model
func (mm *ModelManager) GetRepository(name string) *Repository {
	if repo, exists := mm.repositories[name]; exists {
		return repo
	}
	return nil
}

// GetModel returns a model by name
func (mm *ModelManager) GetModel(name string) interface{} {
	if model, exists := mm.models[name]; exists {
		return model
	}
	return nil
}

// AutoMigrate runs auto migration for all registered models
func (mm *ModelManager) AutoMigrate() error {
	var models []interface{}
	for _, model := range mm.models {
		models = append(models, model)
	}
	return mm.db.AutoMigrate(models...)
}
