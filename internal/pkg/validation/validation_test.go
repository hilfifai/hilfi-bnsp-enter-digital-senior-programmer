package validation

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs
type TestStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Age      int    `json:"age" validate:"required,min=1,max=120"`
	Website  string `json:"website,omitempty" validate:"omitempty,url"`
	Password string `json:"password" validate:"required,min=8"`
}

type TestNestedStruct struct {
	User    TestStruct `json:"user" validate:"required"`
	Address string     `json:"address" validate:"required"`
}

type TestMapStruct struct {
	StringMap    map[string]string      `json:"string_map" validate:"required,mapStringString"`
	InterfaceMap map[string]interface{} `json:"interface_map" validate:"required,mapStringInterface"`
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful setup",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Setup()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, val)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Ensure setup is called
	err := Setup()
	require.NoError(t, err)

	tests := []struct {
		name    string
		payload interface{}
		wantErr bool
		errType string
	}{
		{
			name: "valid struct",
			payload: TestStruct{
				Email:    "test@example.com",
				Name:     "John Doe",
				Age:      25,
				Website:  "https://example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			payload: TestStruct{
				Email:    "invalid-email",
				Name:     "John Doe",
				Age:      25,
				Password: "password123",
			},
			wantErr: true,
			errType: "ValidationErrors",
		},
		{
			name: "missing required fields",
			payload: TestStruct{
				Email: "test@example.com",
				// Name missing
				Age:      25,
				Password: "password123",
			},
			wantErr: true,
			errType: "ValidationErrors",
		},
		{
			name: "age out of range",
			payload: TestStruct{
				Email:    "test@example.com",
				Name:     "John Doe",
				Age:      150, // too old
				Password: "password123",
			},
			wantErr: true,
			errType: "ValidationErrors",
		},
		{
			name: "password too short",
			payload: TestStruct{
				Email:    "test@example.com",
				Name:     "John Doe",
				Age:      25,
				Password: "123", // too short
			},
			wantErr: true,
			errType: "ValidationErrors",
		},
		{
			name: "invalid URL",
			payload: TestStruct{
				Email:    "test@example.com",
				Name:     "John Doe",
				Age:      25,
				Website:  "not-a-url",
				Password: "password123",
			},
			wantErr: true,
			errType: "ValidationErrors",
		},
		{
			name:    "nil payload",
			payload: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType == "ValidationErrors" {
					var validationErrs ValidationErrors
					assert.True(t, errors.As(err, &validationErrs))
					assert.Greater(t, len(validationErrs), 0)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWithConfig(t *testing.T) {
	err := Setup()
	require.NoError(t, err)

	tests := []struct {
		name    string
		payload interface{}
		config  *ValidationConfig
		wantErr bool
	}{
		{
			name: "with default config",
			payload: TestStruct{
				Email:    "test@example.com",
				Name:     "John Doe",
				Age:      25,
				Password: "password123",
			},
			config:  nil,
			wantErr: false,
		},
		{
			name: "with custom config",
			payload: TestStruct{
				Email:    "test@example.com",
				Name:     "John Doe",
				Age:      25,
				Password: "password123",
			},
			config: &ValidationConfig{
				MaxDepth:           5,
				AllowEmptyMaps:     true,
				AllowNilValues:     false,
				StrictTypeChecking: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithConfig(tt.payload, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseValidationErrors(t *testing.T) {
	err := Setup()
	require.NoError(t, err)

	// Create a validation error
	invalidStruct := TestStruct{
		Email: "invalid-email",
		Name:  "", // too short
		Age:   0,  // too small
	}

	validationErr := Validate(invalidStruct)
	require.Error(t, validationErr)

	// Test that it's parsed correctly
	var structuredErr ValidationErrors
	assert.True(t, errors.As(validationErr, &structuredErr))
	assert.Greater(t, len(structuredErr), 0)

	// Check that each error has required fields
	for _, err := range structuredErr {
		assert.NotEmpty(t, err.Field)
		assert.NotEmpty(t, err.Tag)
		assert.NotEmpty(t, err.Message)
	}
}

func TestValidationErrorResponse(t *testing.T) {
	err := Setup()
	require.NoError(t, err)

	tests := []struct {
		name     string
		err      error
		expected map[string]string
	}{
		{
			name: "validation errors",
			err: ValidationErrors{
				{Field: "email", Tag: "email", Message: "must be a valid email"},
				{Field: "name", Tag: "required", Message: "is required"},
			},
			expected: map[string]string{
				"email": "must be a valid email",
				"name":  "is required",
			},
		},
		{
			name: "general error",
			err:  errors.New("general error"),
			expected: map[string]string{
				"general": "general error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidationErrorResponse(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationErrorResponseStructured(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ValidationErrors
	}{
		{
			name: "structured validation errors",
			err: ValidationErrors{
				{Field: "email", Tag: "email", Message: "must be a valid email"},
				{Field: "name", Tag: "required", Message: "is required"},
			},
			expected: ValidationErrors{
				{Field: "email", Tag: "email", Message: "must be a valid email"},
				{Field: "name", Tag: "required", Message: "is required"},
			},
		},
		{
			name: "general error",
			err:  errors.New("general error"),
			expected: ValidationErrors{
				{Field: "general", Tag: "error", Message: "general error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidationErrorResponseStructured(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "validation errors",
			err:  ValidationErrors{{Field: "test", Tag: "required", Message: "is required"}},
			want: true,
		},
		{
			name: "general error",
			err:  errors.New("general error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidationError(tt.err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetValidationEngine(t *testing.T) {
	err := Setup()
	require.NoError(t, err)

	engine := GetValidationEngine()
	assert.NotNil(t, engine)
	assert.IsType(t, &validator.Validate{}, engine)
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "nil value",
			value:   nil,
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "valid string",
			value:   "test",
			wantErr: false,
		},
		{
			name:    "empty slice",
			value:   []string{},
			wantErr: true,
		},
		{
			name:    "valid slice",
			value:   []string{"test"},
			wantErr: false,
		},
		{
			name:    "empty map",
			value:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "valid map",
			value:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "nil pointer",
			value:   (*string)(nil),
			wantErr: true,
		},
		{
			name:    "valid value",
			value:   42,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	err := Setup()
	require.NoError(t, err)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email without @",
			email:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "invalid email without domain",
			email:   "test@",
			wantErr: true,
		},
		{
			name:    "invalid email format",
			email:   "test@example",
			wantErr: true,
		},
		{
			name:    "valid complex email",
			email:   "user.name+tag@example.co.uk",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	err := Setup()
	require.NoError(t, err)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid HTTPS URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "valid URL with path",
			url:     "https://example.com/path/to/resource",
			wantErr: false,
		},
		{
			name:    "valid URL with query params",
			url:     "https://example.com/search?q=test&limit=10",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errors:   ValidationErrors{},
			expected: "validation failed",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "email", Message: "must be a valid email"},
			},
			expected: "Validation failed: email: must be a valid email",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "email", Message: "must be a valid email"},
				{Field: "name", Message: "is required"},
			},
			expected: "Validation failed: email: must be a valid email, name: is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 10, config.MaxDepth)
	assert.False(t, config.AllowEmptyMaps)
	assert.False(t, config.AllowNilValues)
	assert.True(t, config.StrictTypeChecking)
}

func TestMapValidation(t *testing.T) {
	err := Setup()
	require.NoError(t, err)

	tests := []struct {
		name    string
		payload TestMapStruct
		wantErr bool
	}{
		{
			name: "valid maps",
			payload: TestMapStruct{
				StringMap: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				InterfaceMap: map[string]interface{}{
					"key1": "string_value",
					"key2": 123,
					"key3": true,
				},
			},
			wantErr: false,
		},
		{
			name: "empty string map",
			payload: TestMapStruct{
				StringMap: map[string]string{},
				InterfaceMap: map[string]interface{}{
					"key1": "value1",
				},
			},
			wantErr: true,
		},
		{
			name: "empty interface map",
			payload: TestMapStruct{
				StringMap: map[string]string{
					"key1": "value1",
				},
				InterfaceMap: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "string map with empty key",
			payload: TestMapStruct{
				StringMap: map[string]string{
					"":     "value1",
					"key2": "value2",
				},
				InterfaceMap: map[string]interface{}{
					"key1": "value1",
				},
			},
			wantErr: true,
		},
		{
			name: "string map with empty value",
			payload: TestMapStruct{
				StringMap: map[string]string{
					"key1": "",
					"key2": "value2",
				},
				InterfaceMap: map[string]interface{}{
					"key1": "value1",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidate(b *testing.B) {
	err := Setup()
	require.NoError(b, err)

	payload := TestStruct{
		Email:    "test@example.com",
		Name:     "John Doe",
		Age:      25,
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(payload)
	}
}

func BenchmarkValidateComplex(b *testing.B) {
	err := Setup()
	require.NoError(b, err)

	payload := TestNestedStruct{
		User: TestStruct{
			Email:    "test@example.com",
			Name:     "John Doe",
			Age:      25,
			Website:  "https://example.com",
			Password: "password123",
		},
		Address: "123 Main St, City, Country",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(payload)
	}
}

func BenchmarkValidateMap(b *testing.B) {
	err := Setup()
	require.NoError(b, err)

	payload := TestMapStruct{
		StringMap: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		InterfaceMap: map[string]interface{}{
			"key1": "string_value",
			"key2": 123,
			"key3": true,
			"key4": map[string]interface{}{
				"nested": "value",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(payload)
	}
}
