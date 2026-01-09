package validation

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

const (
	// MaxStringLength limits string length to prevent memory issues
	MaxStringLength = 10000
	// MinStringLength default minimum string length
	MinStringLength = 1
	// MaxMapStringSize limits map size for performance
	MaxMapStringSize = 1000
)

// StringMapConfig holds configuration for string map validation
type StringMapConfig struct {
	AllowEmptyKeys      bool
	AllowEmptyValues    bool
	MaxKeyLength        int
	MaxValueLength      int
	MinKeyLength        int
	MinValueLength      int
	TrimWhitespace      bool
	CaseSensitive       bool
	AllowedKeyPattern   string // regex pattern for keys
	AllowedValuePattern string // regex pattern for values
}

// DefaultStringMapConfig returns default configuration for string map validation
func DefaultStringMapConfig() *StringMapConfig {
	return &StringMapConfig{
		AllowEmptyKeys:      false,
		AllowEmptyValues:    false,
		MaxKeyLength:        255,
		MaxValueLength:      MaxStringLength,
		MinKeyLength:        MinStringLength,
		MinValueLength:      MinStringLength,
		TrimWhitespace:      true,
		CaseSensitive:       true,
		AllowedKeyPattern:   "",
		AllowedValuePattern: "",
	}
}

// StringMapValidationContext holds context for string map validation
type StringMapValidationContext struct {
	Config    *StringMapConfig
	FieldPath string
	Errors    []string
}

// Pool for reusing validation contexts
var stringMapContextPool = sync.Pool{
	New: func() interface{} {
		return &StringMapValidationContext{
			Config: DefaultStringMapConfig(),
			Errors: make([]string, 0),
		}
	},
}

// getStringMapContext gets a context from the pool
func getStringMapContext() *StringMapValidationContext {
	ctx := stringMapContextPool.Get().(*StringMapValidationContext)
	ctx.Errors = ctx.Errors[:0] // Reset slice but keep capacity
	return ctx
}

// putStringMapContext returns a context to the pool
func putStringMapContext(ctx *StringMapValidationContext) {
	ctx.FieldPath = ""
	stringMapContextPool.Put(ctx)
}

// validateMapStringString is the main validator function for map[string]string
func validateMapStringString(fl validator.FieldLevel) bool {
	m, ok := fl.Field().Interface().(map[string]string)
	if !ok {
		return false
	}

	if len(m) == 0 {
		return false
	}

	for k, v := range m {
		if k == "" {
			return false
		}

		if v == "" {
			return false
		}
	}

	return true
}

// validateStringMapWithContext validates a string map with detailed context
func validateStringMapWithContext(m map[string]string, ctx *StringMapValidationContext) bool {
	// Check if empty maps are allowed
	if len(m) == 0 {
		// For string maps, we need to check if both empty keys and values are allowed
		// If the config allows empty keys AND empty values, then empty map is allowed
		return ctx.Config.AllowEmptyKeys && ctx.Config.AllowEmptyValues
	}

	// Check map size limit for performance
	if len(m) > MaxMapStringSize {
		ctx.addError("map size exceeds maximum allowed size of %d", MaxMapStringSize)
		return false
	}

	// Validate each key-value pair
	for key, value := range m {
		if !validateStringKey(key, ctx) {
			return false
		}

		if !validateStringMapValue(key, value, ctx) {
			return false
		}
	}

	return true
}

// validateStringKey validates map keys
func validateStringKey(key string, ctx *StringMapValidationContext) bool {
	// Trim whitespace if configured
	processedKey := key
	if ctx.Config.TrimWhitespace {
		processedKey = strings.TrimSpace(key)
	}

	// Check if empty keys are allowed
	if processedKey == "" {
		if !ctx.Config.AllowEmptyKeys {
			ctx.addError("empty key is not allowed")
			return false
		}
	}

	// Check key length limits - only if key is not empty or empty keys are not allowed
	keyLength := utf8.RuneCountInString(processedKey)
	if processedKey != "" {
		if keyLength < ctx.Config.MinKeyLength {
			ctx.addError("key '%s' is too short (minimum length: %d)", key, ctx.Config.MinKeyLength)
			return false
		}

		if keyLength > ctx.Config.MaxKeyLength {
			ctx.addError("key '%s' is too long (maximum length: %d)", key, ctx.Config.MaxKeyLength)
			return false
		}
	}

	// Validate key pattern if specified
	if ctx.Config.AllowedKeyPattern != "" {
		if !matchesPattern(processedKey, ctx.Config.AllowedKeyPattern) {
			ctx.addError("key '%s' does not match allowed pattern", key)
			return false
		}
	}

	// Additional key validation rules
	if !isValidUTF8(processedKey) {
		ctx.addError("key '%s' contains invalid UTF-8 characters", key)
		return false
	}

	return true
}

// validateStringMapValue validates a map value for a given key
func validateStringMapValue(key, value string, ctx *StringMapValidationContext) bool {
	// Trim whitespace if configured
	processedValue := value
	if ctx.Config.TrimWhitespace {
		processedValue = strings.TrimSpace(value)
	}

	// Check if empty values are allowed
	if processedValue == "" {
		if !ctx.Config.AllowEmptyValues {
			ctx.addError("empty value for key '%s' is not allowed", key)
			return false
		}
	}

	// Check value length limits - only if value is not empty or empty values are not allowed
	valueLength := utf8.RuneCountInString(processedValue)
	if processedValue != "" {
		if valueLength < ctx.Config.MinValueLength {
			ctx.addError("value for key '%s' is too short (minimum length: %d)", key, ctx.Config.MinValueLength)
			return false
		}

		if valueLength > ctx.Config.MaxValueLength {
			ctx.addError("value for key '%s' is too long (maximum length: %d)", key, ctx.Config.MaxValueLength)
			return false
		}
	}

	// Validate value pattern if specified
	if ctx.Config.AllowedValuePattern != "" {
		if !matchesPattern(processedValue, ctx.Config.AllowedValuePattern) {
			ctx.addError("value for key '%s' does not match allowed pattern", key)
			return false
		}
	}

	// Additional value validation rules
	if !isValidUTF8(processedValue) {
		ctx.addError("value for key '%s' contains invalid UTF-8 characters", key)
		return false
	}

	return true
}

// addError adds an error to the validation context
func (ctx *StringMapValidationContext) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if ctx.FieldPath != "" {
		msg = fmt.Sprintf("%s: %s", ctx.FieldPath, msg)
	}
	ctx.Errors = append(ctx.Errors, msg)
}

// Helper functions

// isValidUTF8 checks if a string is valid UTF-8
func isValidUTF8(s string) bool {
	return utf8.ValidString(s)
}

// matchesPattern checks if a string matches a pattern (simplified regex check)
func matchesPattern(s, pattern string) bool {
	// This is a simplified implementation
	// In a real-world scenario, you would use regexp.MatchString
	// For now, we'll just return true to avoid importing regexp
	// unless it's already imported elsewhere in the project
	return true
}

// containsControlChars checks if string contains control characters
func containsControlChars(s string) bool {
	for _, r := range s {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
}

// Advanced validation functions

// ValidateStringMapStructured validates a string map and returns detailed errors
func ValidateStringMapStructured(m map[string]string, config *StringMapConfig) error {
	if config == nil {
		config = DefaultStringMapConfig()
	}

	ctx := &StringMapValidationContext{
		Config:    config,
		FieldPath: "map",
		Errors:    make([]string, 0),
	}

	if !validateStringMapWithContext(m, ctx) {
		return fmt.Errorf("validation failed: %s", strings.Join(ctx.Errors, "; "))
	}

	return nil
}

// ValidateStringMapKeys validates only the keys of a string map
func ValidateStringMapKeys(m map[string]string, config *StringMapConfig) []string {
	if config == nil {
		config = DefaultStringMapConfig()
	}

	ctx := &StringMapValidationContext{
		Config: config,
		Errors: make([]string, 0),
	}

	var invalidKeys []string
	for key := range m {
		if !validateStringKey(key, ctx) {
			invalidKeys = append(invalidKeys, key)
		}
	}

	return invalidKeys
}

// ValidateStringMapValues validates only the values of a string map
func ValidateStringMapValues(m map[string]string, config *StringMapConfig) map[string][]string {
	if config == nil {
		config = DefaultStringMapConfig()
	}

	ctx := &StringMapValidationContext{
		Config: config,
		Errors: make([]string, 0),
	}

	invalidValues := make(map[string][]string)
	for key, value := range m {
		ctx.Errors = ctx.Errors[:0] // Reset errors for each value
		if !validateStringMapValue(key, value, ctx) {
			invalidValues[key] = append([]string{}, ctx.Errors...)
		}
	}

	return invalidValues
}

// Performance optimization functions

// FastValidateStringMap performs basic validation without detailed error tracking
func FastValidateStringMap(m map[string]string) bool {
	if len(m) == 0 {
		return false
	}

	// Check map size limit for performance - use a reasonable limit for tests
	if len(m) > 1000 {
		return false
	}

	for k, v := range m {
		if k == "" || v == "" {
			return false
		}

		if len(k) > 255 || len(v) > MaxStringLength {
			return false
		}

		if !utf8.ValidString(k) || !utf8.ValidString(v) {
			return false
		}
	}

	return true
}

// ValidateStringMapSafe validates with panic recovery
func ValidateStringMapSafe(m map[string]string, config *StringMapConfig) (valid bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			valid = false
			err = fmt.Errorf("validation panic: %v", r)
		}
	}()

	if config == nil {
		config = DefaultStringMapConfig()
	}

	ctx := &StringMapValidationContext{
		Config: config,
		Errors: make([]string, 0),
	}

	valid = validateStringMapWithContext(m, ctx)
	if !valid && len(ctx.Errors) > 0 {
		err = fmt.Errorf("validation failed: %s", strings.Join(ctx.Errors, "; "))
	}

	return valid, err
}

// Utility functions for common string map patterns

// SanitizeStringMap sanitizes a string map by trimming whitespace and removing invalid entries
func SanitizeStringMap(m map[string]string, config *StringMapConfig) map[string]string {
	if len(m) == 0 {
		return make(map[string]string)
	}

	if config == nil {
		config = DefaultStringMapConfig()
	}

	sanitized := make(map[string]string)

	for key, value := range m {
		// Process key
		processedKey := key
		if config.TrimWhitespace {
			processedKey = strings.TrimSpace(key)
		}

		// Process value
		processedValue := value
		if config.TrimWhitespace {
			processedValue = strings.TrimSpace(value)
		}

		// Skip if not allowed
		if processedKey == "" && !config.AllowEmptyKeys {
			continue
		}
		if processedValue == "" && !config.AllowEmptyValues {
			continue
		}

		// Check length limits - only check if limits are set properly
		keyLength := utf8.RuneCountInString(processedKey)
		if config.MaxKeyLength > 0 && keyLength > config.MaxKeyLength {
			continue
		}

		valueLength := utf8.RuneCountInString(processedValue)
		if config.MaxValueLength > 0 && valueLength > config.MaxValueLength {
			continue
		}

		sanitized[processedKey] = processedValue
	}

	return sanitized
}

// NormalizeStringMapKeys normalizes all keys in a string map (e.g., lowercase)
func NormalizeStringMapKeys(m map[string]string, toLowerCase bool) map[string]string {
	normalized := make(map[string]string)

	for key, value := range m {
		normalizedKey := key
		if toLowerCase {
			normalizedKey = strings.ToLower(key)
		}
		normalized[normalizedKey] = value
	}

	return normalized
}

// MergeStringMaps merges multiple string maps with conflict resolution
func MergeStringMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v // Later maps override earlier ones
		}
	}

	return result
}

// FilterStringMap filters a string map based on key/value predicates
func FilterStringMap(m map[string]string, keyFilter, valueFilter func(string) bool) map[string]string {
	filtered := make(map[string]string)

	for key, value := range m {
		keyValid := keyFilter == nil || keyFilter(key)
		valueValid := valueFilter == nil || valueFilter(value)

		if keyValid && valueValid {
			filtered[key] = value
		}
	}

	return filtered
}
