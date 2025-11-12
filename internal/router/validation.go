package router

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Validator interface for request validation
type Validator interface {
	Validate() error
}

// ValidationErrors represents validation errors
type ValidationErrors map[string][]string

// Error implements error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for field, errors := range ve {
		for _, err := range errors {
			messages = append(messages, field+": "+err)
		}
	}
	return strings.Join(messages, "; ")
}

// Add adds an error for a field
func (ve ValidationErrors) Add(field, message string) {
	ve[field] = append(ve[field], message)
}

// HasErrors checks if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// ValidateRequest validates a request using a validator
func ValidateRequest(r *http.Request, v Validator) error {
	// Bind JSON body if Content-Type is application/json
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(v); err != nil {
			return err
		}
	} else {
		// For form data, use reflection to populate struct
		if err := r.ParseForm(); err != nil {
			return err
		}
		if err := bindFormToStruct(r, v); err != nil {
			return err
		}
	}
	
	// Validate
	return v.Validate()
}

// ValidationMiddleware creates middleware for request validation
func ValidationMiddleware(validator Validator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new instance of the validator type
			v := reflect.New(reflect.TypeOf(validator).Elem()).Interface().(Validator)
			
			if err := ValidateRequest(r, v); err != nil {
				if ve, ok := err.(ValidationErrors); ok {
					ValidationError(w, r, ve)
					return
				}
				BadRequest(w, r, err.Error())
				return
			}
			
			// Store validated request in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "validated_request", v)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// bindFormToStruct binds form values to struct fields
func bindFormToStruct(r *http.Request, v interface{}) error {
	val := reflect.ValueOf(v).Elem()
	typ := val.Type()
	
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		
		// Get form value
		formKey := field.Tag.Get("form")
		if formKey == "" {
			formKey = strings.ToLower(field.Name)
		}
		
		formValue := r.FormValue(formKey)
		if formValue == "" {
			continue
		}
		
		// Set field value based on type
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(formValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// Parse int - simplified, in production use proper parsing
			if intVal, err := strconv.ParseInt(formValue, 10, 64); err == nil {
				fieldVal.SetInt(intVal)
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(formValue); err == nil {
				fieldVal.SetBool(boolVal)
			}
		}
	}
	
	return nil
}

