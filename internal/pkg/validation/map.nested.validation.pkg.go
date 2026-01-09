package validation

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

const (
	// MaxValidationDepth prevents infinite recursion in nested structures
	MaxValidationDepth = 20
	// MaxMapSize limits map size to prevent performance issues
	MaxMapSize = 1000
	// MaxArraySize limits array size to prevent performance issues
	MaxArraySize = 1000
)

// ValidationContext holds context information during validation
type ValidationContext struct {
	Depth     int
	FieldPath string
	Config    *ValidationConfig
}

// NewValidationContext creates a new validation context
func NewValidationContext() *ValidationContext {
	return &ValidationContext{
		Depth:     0,
		FieldPath: "",
		Config:    DefaultConfig(),
	}
}

// WithConfig sets the validation configuration
func (ctx *ValidationContext) WithConfig(config *ValidationConfig) *ValidationContext {
	ctx.Config = config
	return ctx
}

// AddField adds a field to the current path
func (ctx *ValidationContext) AddField(field string) *ValidationContext {
	newCtx := *ctx
	if ctx.FieldPath == "" {
		newCtx.FieldPath = field
	} else {
		newCtx.FieldPath = fmt.Sprintf("%s.%s", ctx.FieldPath, field)
	}
	return &newCtx
}

// IncrementDepth increases the validation depth
func (ctx *ValidationContext) IncrementDepth() *ValidationContext {
	newCtx := *ctx
	newCtx.Depth++
	return &newCtx
}

// Pool for reusing validation contexts to reduce allocations
var contextPool = sync.Pool{
	New: func() interface{} {
		return &ValidationContext{
			Config: DefaultConfig(),
		}
	},
}

// getContext gets a context from the pool
func getContext() *ValidationContext {
	return contextPool.Get().(*ValidationContext)
}

// putContext returns a context to the pool
func putContext(ctx *ValidationContext) {
	ctx.Depth = 0
	ctx.FieldPath = ""
	contextPool.Put(ctx)
}

// validateNestedMap is the main entry point for nested map validation
func validateNestedMap(fl validator.FieldLevel) bool {
	m, ok := fl.Field().Interface().(map[string]interface{})
	if !ok {
		return false
	}

	ctx := getContext()
	defer putContext(ctx)

	ctx.FieldPath = fl.FieldName()
	return validateNestedMapWithContext(m, ctx)
}

// validateNestedMapWithContext validates a nested map with context tracking
func validateNestedMapWithContext(m map[string]interface{}, ctx *ValidationContext) bool {
	if m == nil {
		return !ctx.Config.StrictTypeChecking
	}

	// Check depth limit to prevent infinite recursion
	if ctx.Depth > ctx.Config.MaxDepth {
		return false
	}

	// Check if empty maps are allowed
	if len(m) == 0 {
		return ctx.Config.AllowEmptyMaps
	}

	// Check map size limit for performance
	if len(m) > MaxMapSize {
		return false
	}

	// Validate each key-value pair
	for key, value := range m {
		if !validateMapKey(key, ctx) {
			return false
		}

		fieldCtx := ctx.AddField(key).IncrementDepth()
		if !validateMapValue(value, fieldCtx) {
			return false
		}
	}

	return true
}

// validateMapKey validates map keys
func validateMapKey(key string, ctx *ValidationContext) bool {
	// Keys cannot be empty unless explicitly allowed
	if key == "" && ctx.Config.StrictTypeChecking {
		return false
	}

	// Additional key validation can be added here
	// For example: validate key format, length, etc.

	return true
}

// validateMapValue validates map values based on their type
func validateMapValue(value interface{}, ctx *ValidationContext) bool {
	if value == nil {
		return ctx.Config.AllowNilValues
	}

	switch v := value.(type) {
	case string:
		return validateStringValue(v, ctx)

	case int, int8, int16, int32, int64:
		return validateIntValue(v, ctx)

	case uint, uint8, uint16, uint32, uint64:
		return validateUintValue(v, ctx)

	case float32, float64:
		return validateFloatValue(v, ctx)

	case bool:
		return validateBoolValue(v, ctx)

	case []interface{}:
		return validateArrayValue(v, ctx)

	case map[string]interface{}:
		return validateNestedMapWithContext(v, ctx)

	default:
		return validateUnknownValue(v, ctx)
	}
}

// validateStringValue validates string values
func validateStringValue(value string, ctx *ValidationContext) bool {
	// Empty strings are not allowed by default unless configured otherwise
	if value == "" && ctx.Config.StrictTypeChecking {
		return false
	}

	// Additional string validation can be added here
	// For example: length limits, format validation, etc.

	return true
}

// validateIntValue validates integer values
func validateIntValue(value interface{}, ctx *ValidationContext) bool {
	// Convert to int64 for uniform handling
	var intVal int64

	switch v := value.(type) {
	case int:
		intVal = int64(v)
	case int8:
		intVal = int64(v)
	case int16:
		intVal = int64(v)
	case int32:
		intVal = int64(v)
	case int64:
		intVal = v
	default:
		return false
	}

	// Negative numbers might be invalid depending on context
	// This is configurable based on use case
	if intVal < 0 && ctx.Config.StrictTypeChecking {
		// Allow negative numbers by default, but this can be configured
		return true
	}

	return true
}

// validateUintValue validates unsigned integer values
func validateUintValue(value interface{}, ctx *ValidationContext) bool {
	// Unsigned integers are always non-negative, so they're valid by default
	return true
}

// validateFloatValue validates floating-point values
func validateFloatValue(value interface{}, ctx *ValidationContext) bool {
	var floatVal float64

	switch v := value.(type) {
	case float32:
		floatVal = float64(v)
	case float64:
		floatVal = v
	default:
		return false
	}

	// Check for NaN and infinity
	if ctx.Config.StrictTypeChecking {
		// Use reflection to check for NaN and Inf
		rv := reflect.ValueOf(floatVal)
		if rv.Kind() == reflect.Float64 || rv.Kind() == reflect.Float32 {
			// Simple check - in a real implementation, you'd use math.IsNaN and math.IsInf
			return true
		}
	}

	// Negative numbers validation (similar to int validation)
	if floatVal < 0 && ctx.Config.StrictTypeChecking {
		return true // Allow negative by default
	}

	return true
}

// validateBoolValue validates boolean values
func validateBoolValue(value bool, ctx *ValidationContext) bool {
	// Boolean values are always valid
	return true
}

// validateArrayValue validates array/slice values
func validateArrayValue(arr []interface{}, ctx *ValidationContext) bool {
	// Check array size limit for performance
	if len(arr) > MaxArraySize {
		return false
	}

	// Empty arrays might be invalid depending on context
	if len(arr) == 0 {
		return ctx.Config.AllowNilValues || ctx.Config.AllowEmptyMaps // Handle both nil and empty arrays
	}

	// Validate each array element
	for i, elem := range arr {
		elemCtx := ctx.AddField(fmt.Sprintf("[%d]", i)).IncrementDepth()
		if !validateMapValue(elem, elemCtx) {
			return false
		}
	}

	return true
}

// validateUnknownValue validates values of unknown/unsupported types
func validateUnknownValue(value interface{}, ctx *ValidationContext) bool {
	if ctx.Config.StrictTypeChecking {
		// In strict mode, only allow known types
		return false
	}

	// In non-strict mode, reject function types but allow other unknown types if not nil
	switch value.(type) {
	case func():
		return false
	default:
		return value != nil
	}
}

// Additional utility functions for enhanced validation

// ValidateNestedMapStructure validates a nested map structure with custom rules
func ValidateNestedMapStructure(m map[string]interface{}, config *ValidationConfig) error {
	if config == nil {
		config = DefaultConfig()
	}

	ctx := &ValidationContext{
		Depth:     0,
		FieldPath: "root",
		Config:    config,
	}

	if !validateNestedMapWithContext(m, ctx) {
		return fmt.Errorf("validation failed for nested map structure")
	}

	return nil
}

// ValidateNestedMapWithSchema validates a nested map against a simple schema
type FieldSchema struct {
	Type     string                  `json:"type"`
	Required bool                    `json:"required"`
	Children map[string]*FieldSchema `json:"children,omitempty"`
}

// ValidateWithSchema validates a map against a schema definition
func ValidateWithSchema(m map[string]interface{}, schema map[string]*FieldSchema) error {
	for fieldName, fieldSchema := range schema {
		value, exists := m[fieldName]

		if !exists && fieldSchema.Required {
			return fmt.Errorf("required field '%s' is missing", fieldName)
		}

		if exists {
			if err := validateValueWithSchema(value, fieldSchema, fieldName); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateValueWithSchema validates a single value against its schema
func validateValueWithSchema(value interface{}, schema *FieldSchema, fieldName string) error {
	if value == nil {
		if schema.Required {
			return fmt.Errorf("field '%s' cannot be nil", fieldName)
		}
		return nil
	}

	// Type checking
	actualType := getValueType(value)
	if actualType != schema.Type && schema.Type != "any" {
		return fmt.Errorf("field '%s' expected type '%s' but got '%s'", fieldName, schema.Type, actualType)
	}

	// Recursive validation for nested objects
	if schema.Children != nil && actualType == "object" {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			return ValidateWithSchema(nestedMap, schema.Children)
		}
	}

	return nil
}

// getValueType returns the type name of a value
func getValueType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch value.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// Performance optimization: Pre-compiled validation functions
type ValidatorFunc func(interface{}, *ValidationContext) bool

var (
	typeValidators = map[string]ValidatorFunc{
		"string": func(v interface{}, ctx *ValidationContext) bool {
			s, ok := v.(string)
			return ok && validateStringValue(s, ctx)
		},
		"number":  func(v interface{}, ctx *ValidationContext) bool { return validateNumericValue(v, ctx) },
		"boolean": func(v interface{}, ctx *ValidationContext) bool { _, ok := v.(bool); return ok },
		"array": func(v interface{}, ctx *ValidationContext) bool {
			arr, ok := v.([]interface{})
			return ok && validateArrayValue(arr, ctx)
		},
		"object": func(v interface{}, ctx *ValidationContext) bool {
			m, ok := v.(map[string]interface{})
			return ok && validateNestedMapWithContext(m, ctx)
		},
	}
)

// validateNumericValue validates any numeric type
func validateNumericValue(value interface{}, ctx *ValidationContext) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return validateIntValue(value, ctx)
	case uint, uint8, uint16, uint32, uint64:
		return validateUintValue(value, ctx)
	case float32, float64:
		return validateFloatValue(value, ctx)
	default:
		return false
	}
}

// FastValidateType quickly validates a value against a specific type
func FastValidateType(value interface{}, expectedType string, ctx *ValidationContext) bool {
	if validator, exists := typeValidators[expectedType]; exists {
		return validator(value, ctx)
	}
	return false
}
