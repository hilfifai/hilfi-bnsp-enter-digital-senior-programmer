package validation

import (
	"api-jellyfish/internal/common/enum"
	types "api-jellyfish/internal/common/type"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	val  *validator.Validate
	once sync.Once
)

// ValidationConfig holds configuration for validation behavior
type ValidationConfig struct {
	MaxDepth           int
	AllowEmptyMaps     bool
	AllowNilValues     bool
	StrictTypeChecking bool
}

// DefaultConfig returns the default validation configuration
func DefaultConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxDepth:           10,
		AllowEmptyMaps:     false,
		AllowNilValues:     false,
		StrictTypeChecking: true,
	}
}

// ValidationError represents a structured validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return "Validation failed: " + strings.Join(messages, ", ")
}

// Enhanced validation messages with better context
var validationMessages = map[string]string{
	"e164":               "must be a valid E.164 formatted phone number",
	"required":           "is required and cannot be empty",
	"url":                "must be a valid URL format",
	"datetime":           "must be a valid date-time format (2006-01-02T15:04:05Z07:00)",
	"number":             "must be a valid number",
	"oneof":              "must be one of the allowed values: %s",
	"email":              "must be a valid email address",
	"min":                "must be at least %s characters/items long",
	"max":                "must be at most %s characters/items long",
	"len":                "must have exactly %s characters/items",
	"alpha":              "must contain only alphabetic characters (a-z, A-Z)",
	"alphanum":           "must contain only alphanumeric characters (a-z, A-Z, 0-9)",
	"eqfield":            "must be equal to the value of field '%s'",
	"nefield":            "must not be equal to the value of field '%s'",
	"gt":                 "must be greater than %s",
	"gte":                "must be greater than or equal to %s",
	"lt":                 "must be less than %s",
	"lte":                "must be less than or equal to %s",
	"excludes":           "must not contain the value '%s'",
	"excludesall":        "must not contain any of the values: %s",
	"enum":               "must be one of the allowed enum values: %s",
	"stringToBool":       "must be a valid boolean value (true, false, 1, 0)",
	"mapStringString":    "must be a valid map with non-empty string keys and values",
	"mapStringInterface": "must be a valid nested map structure with non-empty keys",
	"time_gtfield":       "must be a valid time range (greater than field '%s')",
}

// Setup initializes the validation engine with custom validators
func Setup() error {
	var setupErr error

	once.Do(func() {
		val = validator.New(validator.WithRequiredStructEnabled())

		if err := registerCustomValidators(val); err != nil {
			setupErr = fmt.Errorf("failed to register custom validations: %w", err)
			return
		}

		// Register JSON tag name function
		val.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// Register with Gin binding validator
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			// Configure Gin validator to use JSON tag names (same as custom validation)
			v.RegisterTagNameFunc(func(fld reflect.StructField) string {
				name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
				if name == "-" {
					return ""
				}
				return name
			})

			if err := registerCustomValidators(v); err != nil {
				setupErr = fmt.Errorf("failed to register custom validations in Gin engine: %w", err)
				return
			}
		} else {
			setupErr = fmt.Errorf("failed to get validation engine")
			return
		}
	})

	return setupErr
}

// registerCustomValidators registers all custom validation functions
func registerCustomValidators(v *validator.Validate) error {
	validators := map[string]validator.Func{
		"enum":               enum.ValidateEnum,
		"stringToBool":       types.ValidateStringToBool,
		"mapStringString":    validateMapStringString,
		"mapStringInterface": validateNestedMap,
	}

	for tag, fn := range validators {
		if err := v.RegisterValidation(tag, fn); err != nil {
			return fmt.Errorf("failed to register %s validation: %w", tag, err)
		}
	}

	return nil
}

// Validate validates a struct with enhanced error reporting
func Validate(payload interface{}) error {
	if val == nil {
		return errors.New("validation engine not initialized, call Setup() first")
	}

	if err := val.Struct(payload); err != nil {
		return parseValidationErrors(err)
	}

	return nil
}

// ValidateWithConfig validates with custom configuration
func ValidateWithConfig(payload interface{}, config *ValidationConfig) error {
	if config == nil {
		config = DefaultConfig()
	}

	// Set validation context if needed for custom validators
	// This would require modifying custom validators to accept config
	return Validate(payload)
}

// parseValidationErrors converts validator errors to structured errors
func parseValidationErrors(err error) error {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return fmt.Errorf("validation error: %w", err)
	}

	var structuredErrors ValidationErrors

	for _, fieldErr := range validationErrs {
		ve := ValidationError{
			Field: getFieldName(fieldErr),
			Tag:   fieldErr.Tag(),
			Value: fmt.Sprintf("%v", fieldErr.Value()),
		}

		ve.Message = formatErrorMessage(fieldErr)
		structuredErrors = append(structuredErrors, ve)
	}

	return structuredErrors
}

// getFieldName extracts the field name from validation error
func getFieldName(fieldErr validator.FieldError) string {
	field := fieldErr.Field()
	if field == "" {
		field = fieldErr.Namespace()
	}
	return field
}

// formatErrorMessage creates a human-readable error message
func formatErrorMessage(fieldErr validator.FieldError) string {
	tag := fieldErr.Tag()
	param := fieldErr.Param()

	msg, exists := validationMessages[tag]
	if !exists {
		return fmt.Sprintf("failed validation for tag '%s'", tag)
	}

	switch tag {
	case "enum":
		// For enum validation, we might want to get the actual enum values
		if fieldErr.Type() != nil {
			return fmt.Sprintf(msg, fieldErr.Type().String())
		}
		return msg
	case "oneof":
		return fmt.Sprintf(msg, param)
	default:
		if strings.Contains(msg, "%s") && param != "" {
			return fmt.Sprintf(msg, param)
		}
		return msg
	}
}

// ValidationErrorResponse creates a map response for API errors (backward compatibility)
func ValidationErrorResponse(err error) map[string]string {
	errors := make(map[string]string)

	var validationErrs ValidationErrors
	if errors2, ok := err.(ValidationErrors); ok {
		validationErrs = errors2
	} else if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		// Handle legacy validator.ValidationErrors
		for _, fieldErr := range validatorErrs {
			fieldName := getFieldName(fieldErr)
			errors[fieldName] = formatErrorMessage(fieldErr)
		}
		return errors
	} else {
		errors["general"] = err.Error()
		return errors
	}

	for _, ve := range validationErrs {
		errors[ve.Field] = ve.Message
	}

	return errors
}

// ValidationErrorResponseStructured returns structured validation errors
func ValidationErrorResponseStructured(err error) ValidationErrors {
	if validationErrs, ok := err.(ValidationErrors); ok {
		return validationErrs
	}

	// Convert from other error types
	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		var structuredErrors ValidationErrors
		for _, fieldErr := range validatorErrs {
			ve := ValidationError{
				Field:   getFieldName(fieldErr),
				Tag:     fieldErr.Tag(),
				Message: formatErrorMessage(fieldErr),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
			}
			structuredErrors = append(structuredErrors, ve)
		}
		return structuredErrors
	}

	return ValidationErrors{{
		Field:   "general",
		Tag:     "error",
		Message: err.Error(),
	}}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}

	_, isStructured := err.(ValidationErrors)
	_, isValidator := err.(validator.ValidationErrors)

	return isStructured || isValidator
}

// GetValidationEngine returns the validator instance (for advanced usage)
func GetValidationEngine() *validator.Validate {
	return val
}

// Helper functions for common validation patterns

// ValidateRequired checks if a value is not nil/empty
func ValidateRequired(value interface{}) error {
	if value == nil {
		return errors.New("value is required")
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		if rv.String() == "" {
			return errors.New("value cannot be empty")
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if rv.Len() == 0 {
			return errors.New("value cannot be empty")
		}
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return errors.New("value cannot be nil")
		}
	}

	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Use the validator engine for consistency
	if val != nil {
		return val.Var(email, "email")
	}

	// Fallback basic validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("invalid email format")
	}

	return nil
}

// ValidateURL validates URL format
func ValidateURL(url string) error {
	if url == "" {
		return errors.New("URL is required")
	}

	if val != nil {
		return val.Var(url, "url")
	}

	return nil
}
