package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Model represents the base model interface
type Model interface {
	TableName() string
}

// BaseModel provides common functionality for all models
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Repository provides database operations for models
type Repository[T Model] struct {
	db    *gorm.DB
	model T
}

// NewRepository creates a new repository instance
func NewRepository[T Model](db *gorm.DB, model T) *Repository[T] {
	return &Repository[T]{
		db:    db,
		model: model,
	}
}

// Create creates a new record
func (r *Repository[T]) Create(ctx context.Context, model *T) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// Find finds a record by ID
func (r *Repository[T]) Find(ctx context.Context, id uint) (*T, error) {
	var model T
	err := r.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// FindBy finds a record by field and value
func (r *Repository[T]) FindBy(ctx context.Context, field string, value interface{}) (*T, error) {
	var model T
	err := r.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", field), value).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// FindAll finds all records
func (r *Repository[T]) FindAll(ctx context.Context) ([]T, error) {
	var models []T
	err := r.db.WithContext(ctx).Find(&models).Error
	return models, err
}

// FindWhere finds records with conditions
func (r *Repository[T]) FindWhere(ctx context.Context, conditions map[string]interface{}) ([]T, error) {
	var models []T
	query := r.db.WithContext(ctx)

	for field, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	err := query.Find(&models).Error
	return models, err
}

// Update updates a record
func (r *Repository[T]) Update(ctx context.Context, model *T) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// UpdateBy updates a record by ID
func (r *Repository[T]) UpdateBy(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(r.model).Where("id = ?", id).Updates(updates).Error
}

// Delete soft deletes a record
func (r *Repository[T]) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(r.model, id).Error
}

// ForceDelete permanently deletes a record
func (r *Repository[T]) ForceDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(r.model, id).Error
}

// Count counts records
func (r *Repository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(r.model).Count(&count).Error
	return count, err
}

// CountWhere counts records with conditions
func (r *Repository[T]) CountWhere(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(r.model)

	for field, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	err := query.Count(&count).Error
	return count, err
}

// Exists checks if a record exists
func (r *Repository[T]) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(r.model).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsBy checks if a record exists by field and value
func (r *Repository[T]) ExistsBy(ctx context.Context, field string, value interface{}) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(r.model).Where(fmt.Sprintf("%s = ?", field), value).Count(&count).Error
	return count > 0, err
}

// Paginate returns paginated results
func (r *Repository[T]) Paginate(ctx context.Context, page, limit int) (*PaginatedResult[T], error) {
	var models []T
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(r.model).Count(&total).Error; err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch records
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginatedResult[T]{
		Data:       models,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}, nil
}

// PaginatedResult represents paginated query results
type PaginatedResult[T Model] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// QueryBuilder provides a fluent query builder
type QueryBuilder[T Model] struct {
	db    *gorm.DB
	model T
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder[T Model](db *gorm.DB, model T) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		db:    db,
		model: model,
	}
}

// Where adds a WHERE condition
func (qb *QueryBuilder[T]) Where(field string, value interface{}) *QueryBuilder[T] {
	qb.db = qb.db.Where(fmt.Sprintf("%s = ?", field), value)
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder[T]) WhereIn(field string, values []interface{}) *QueryBuilder[T] {
	qb.db = qb.db.Where(fmt.Sprintf("%s IN ?", field), values)
	return qb
}

// WhereNot adds a WHERE NOT condition
func (qb *QueryBuilder[T]) WhereNot(field string, value interface{}) *QueryBuilder[T] {
	qb.db = qb.db.Where(fmt.Sprintf("%s != ?", field), value)
	return qb
}

// WhereLike adds a WHERE LIKE condition
func (qb *QueryBuilder[T]) WhereLike(field string, value string) *QueryBuilder[T] {
	qb.db = qb.db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder[T]) OrderBy(field string, direction string) *QueryBuilder[T] {
	if direction == "" {
		direction = "ASC"
	}
	qb.db = qb.db.Order(fmt.Sprintf("%s %s", field, strings.ToUpper(direction)))
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	qb.db = qb.db.Limit(limit)
	return qb
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	qb.db = qb.db.Offset(offset)
	return qb
}

// Get executes the query and returns results
func (qb *QueryBuilder[T]) Get(ctx context.Context) ([]T, error) {
	var models []T
	err := qb.db.WithContext(ctx).Find(&models).Error
	return models, err
}

// First executes the query and returns the first result
func (qb *QueryBuilder[T]) First(ctx context.Context) (*T, error) {
	var model T
	err := qb.db.WithContext(ctx).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// Count executes the query and returns the count
func (qb *QueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	err := qb.db.WithContext(ctx).Model(qb.model).Count(&count).Error
	return count, err
}

// Transaction executes operations within a database transaction
func Transaction[T Model](db *gorm.DB, fn func(*gorm.DB) error) error {
	return db.Transaction(fn)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// Validator provides validation functionality
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a model
func (v *Validator) Validate(model interface{}) []ValidationError {
	var errors []ValidationError

	// Use reflection to validate model fields
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check for required fields
		if fieldType.Tag.Get("validate") == "required" {
			if field.Kind() == reflect.String && field.String() == "" {
				errors = append(errors, ValidationError{
					Field:   fieldType.Name,
					Message: "is required",
				})
			}
		}

		// Check for email validation
		if fieldType.Tag.Get("validate") == "email" {
			if field.Kind() == reflect.String && field.String() != "" {
				if !isValidEmail(field.String()) {
					errors = append(errors, ValidationError{
						Field:   fieldType.Name,
						Message: "must be a valid email",
					})
				}
			}
		}
	}

	return errors
}

// isValidEmail checks if an email is valid
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
