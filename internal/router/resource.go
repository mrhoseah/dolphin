package router

// Resource represents a transformable API resource
type Resource interface {
	ToMap() map[string]interface{}
}

// ResourceTransformer transforms a resource to API format
type ResourceTransformer struct {
	resource interface{}
}

// NewResource creates a new resource transformer
func NewResource(resource interface{}) *ResourceTransformer {
	return &ResourceTransformer{resource: resource}
}

// Transform transforms a single resource
func (rt *ResourceTransformer) Transform() map[string]interface{} {
	if resource, ok := rt.resource.(Resource); ok {
		return resource.ToMap()
	}
	return nil
}

// Collection transforms a collection of resources
func Collection(resources []interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, resource := range resources {
		if r, ok := resource.(Resource); ok {
			result = append(result, r.ToMap())
		}
	}
	return result
}

// TransformCollection transforms a collection of resources using a transformer function
func TransformCollection[T any](items []T, transformer func(T) map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, item := range items {
		result = append(result, transformer(item))
	}
	return result
}

