package resources

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Resource represents an API resource
type Resource interface {
	ToArray() map[string]interface{}
	ToJSON() ([]byte, error)
}

// BaseResource provides common resource functionality
type BaseResource struct {
	data map[string]interface{}
}

// NewResource creates a new resource from data
func NewResource(data interface{}) *BaseResource {
	return &BaseResource{
		data: toMap(data),
	}
}

// ToArray returns resource as map
func (r *BaseResource) ToArray() map[string]interface{} {
	return r.data
}

// ToJSON returns resource as JSON
func (r *BaseResource) ToJSON() ([]byte, error) {
	return json.Marshal(r.data)
}

// Transform transforms data using a transformer function
func (r *BaseResource) Transform(transformer func(map[string]interface{}) map[string]interface{}) *BaseResource {
	r.data = transformer(r.data)
	return r
}

// Only returns only specified fields
func (r *BaseResource) Only(fields ...string) *BaseResource {
	result := make(map[string]interface{})
	for _, field := range fields {
		if value, exists := r.data[field]; exists {
			result[field] = value
		}
	}
	return &BaseResource{data: result}
}

// Except returns all fields except specified ones
func (r *BaseResource) Except(fields ...string) *BaseResource {
	result := make(map[string]interface{})
	excluded := make(map[string]bool)
	for _, field := range fields {
		excluded[field] = true
	}
	for key, value := range r.data {
		if !excluded[key] {
			result[key] = value
		}
	}
	return &BaseResource{data: result}
}

// Merge merges additional data
func (r *BaseResource) Merge(data map[string]interface{}) *BaseResource {
	for key, value := range data {
		r.data[key] = value
	}
	return r
}

// Collection represents a collection of resources
type Collection struct {
	data []map[string]interface{}
}

// NewCollection creates a new collection
func NewCollection(items interface{}) *Collection {
	itemsValue := reflect.ValueOf(items)
	if itemsValue.Kind() == reflect.Ptr {
		itemsValue = itemsValue.Elem()
	}

	var data []map[string]interface{}

	switch itemsValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < itemsValue.Len(); i++ {
			item := itemsValue.Index(i).Interface()
			data = append(data, toMap(item))
		}
	default:
		data = []map[string]interface{}{toMap(items)}
	}

	return &Collection{data: data}
}

// ToArray returns collection as array of maps
func (c *Collection) ToArray() []map[string]interface{} {
	return c.data
}

// ToJSON returns collection as JSON
func (c *Collection) ToJSON() ([]byte, error) {
	return json.Marshal(c.data)
}

// Transform transforms each item in collection
func (c *Collection) Transform(transformer func(map[string]interface{}) map[string]interface{}) *Collection {
	for i, item := range c.data {
		c.data[i] = transformer(item)
	}
	return c
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       []map[string]interface{} `json:"data"`
	Pagination PaginationMeta           `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
	From        int   `json:"from"`
	To          int   `json:"to"`
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(
	items interface{},
	page, perPage int,
	total int64,
) *PaginatedResponse {
	collection := NewCollection(items)
	data := collection.ToArray()

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	from := (page-1)*perPage + 1
	to := from + len(data) - 1

	return &PaginatedResponse{
		Data: data,
		Pagination: PaginationMeta{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       total,
			LastPage:    lastPage,
			From:        from,
			To:          to,
		},
	}
}

// ToJSON returns paginated response as JSON
func (p *PaginatedResponse) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// toMap converts any value to map[string]interface{}
func toMap(value interface{}) map[string]interface{} {
	if value == nil {
		return make(map[string]interface{})
	}

	// If already a map, return it
	if m, ok := value.(map[string]interface{}); ok {
		return m
	}

	// Use reflection to convert struct to map
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return map[string]interface{}{"value": value}
	}

	result := make(map[string]interface{})
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get JSON tag or field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			jsonTag = field.Name
		} else {
			// Remove omitempty and other options
			if idx := strings.Index(jsonTag, ","); idx != -1 {
				jsonTag = jsonTag[:idx]
			}
		}

		// Handle pointers
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				result[jsonTag] = nil
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		result[jsonTag] = fieldValue.Interface()
	}

	return result
}

