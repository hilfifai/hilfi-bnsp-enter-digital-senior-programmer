package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNestedMap(t *testing.T) {
	// Setup validator
	v := validator.New()
	err := v.RegisterValidation("mapStringInterface", validateNestedMap)
	require.NoError(t, err)

	type TestStruct struct {
		NestedMap map[string]interface{} `validate:"mapStringInterface"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{
			name: "valid nested map with strings",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantErr: false,
		},
		{
			name: "valid nested map with mixed types",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"string_key":  "string_value",
					"number_key":  42.5,
					"boolean_key": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid nested map with arrays",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"array_key": []interface{}{
						"item1",
						"item2",
						42,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid deeply nested map",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": "deep_value",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid nested map with nil values (allowed by config)",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"key1":   "value1",
					"nilkey": nil,
				},
			},
			wantErr: true, // nil values not allowed by default
		},
		{
			name: "empty map",
			input: TestStruct{
				NestedMap: map[string]interface{}{},
			},
			wantErr: true, // empty maps not allowed by default
		},
		{
			name: "nil map",
			input: TestStruct{
				NestedMap: nil,
			},
			wantErr: true,
		},
		{
			name: "map with empty key",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"":     "value1",
					"key2": "value2",
				},
			},
			wantErr: true,
		},
		{
			name: "map with empty string value",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"key1": "",
					"key2": "value2",
				},
			},
			wantErr: true,
		},
		{
			name: "map with negative number",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"negative": -42.5,
				},
			},
			wantErr: false, // negative numbers allowed by default
		},
		{
			name: "map with array containing empty string",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"array_key": []interface{}{
						"valid_item",
						"",
						"another_item",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "map with array containing nil",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"array_key": []interface{}{
						"valid_item",
						nil,
						"another_item",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "map with unsupported type",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"func_key": func() {},
				},
			},
			wantErr: true,
		},
		{
			name: "complex valid structure",
			input: TestStruct{
				NestedMap: map[string]interface{}{
					"user": map[string]interface{}{
						"name":   "John Doe",
						"age":    30,
						"active": true,
						"preferences": map[string]interface{}{
							"theme":    "dark",
							"language": "en",
						},
						"tags": []interface{}{
							"developer",
							"golang",
							"testing",
						},
					},
					"metadata": map[string]interface{}{
						"created_at": "2023-01-01",
						"version":    1,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNestedMapWithContext(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]interface{}
		config *ValidationConfig
		want   bool
	}{
		{
			name: "valid map with default config",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:  "empty map with allow empty config",
			input: map[string]interface{}{},
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     true,
				AllowNilValues:     false,
				StrictTypeChecking: true,
			},
			want: true,
		},
		{
			name:  "empty map with disallow empty config",
			input: map[string]interface{}{},
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     false,
				AllowNilValues:     false,
				StrictTypeChecking: true,
			},
			want: false,
		},
		{
			name: "map with nil values - allowed",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": nil,
			},
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     false,
				AllowNilValues:     true,
				StrictTypeChecking: true,
			},
			want: true,
		},
		{
			name: "map with nil values - not allowed",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": nil,
			},
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     false,
				AllowNilValues:     false,
				StrictTypeChecking: true,
			},
			want: false,
		},
		{
			name:  "deeply nested map exceeding max depth",
			input: createDeeplyNestedMap(15),
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     false,
				AllowNilValues:     false,
				StrictTypeChecking: true,
			},
			want: false,
		},
		{
			name:  "deeply nested map within max depth",
			input: createDeeplyNestedMap(5),
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     false,
				AllowNilValues:     false,
				StrictTypeChecking: true,
			},
			want: true,
		},
		{
			name: "map with unsupported type - strict mode",
			input: map[string]interface{}{
				"func_key": func() {},
			},
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     false,
				AllowNilValues:     false,
				StrictTypeChecking: true,
			},
			want: false,
		},
		{
			name: "map with unsupported type - non-strict mode",
			input: map[string]interface{}{
				"func_key": func() {},
			},
			config: &ValidationConfig{
				MaxDepth:           10,
				AllowEmptyMaps:     false,
				AllowNilValues:     false,
				StrictTypeChecking: false,
			},
			want: false, // func is nil when converted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				Depth:     0,
				FieldPath: "test",
				Config:    tt.config,
			}
			result := validateNestedMapWithContext(tt.input, ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateMapValue(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		config *ValidationConfig
		want   bool
	}{
		{
			name:   "valid string",
			value:  "test_string",
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "empty string - strict mode",
			value:  "",
			config: DefaultConfig(),
			want:   false,
		},
		{
			name:   "valid integer",
			value:  42,
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "valid float",
			value:  42.5,
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "valid boolean",
			value:  true,
			config: DefaultConfig(),
			want:   true,
		},
		{
			name: "valid array",
			value: []interface{}{
				"item1",
				"item2",
				42,
			},
			config: DefaultConfig(),
			want:   true,
		},
		{
			name: "valid nested map",
			value: map[string]interface{}{
				"nested_key": "nested_value",
			},
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "nil value - allowed",
			value:  nil,
			config: &ValidationConfig{AllowNilValues: true},
			want:   true,
		},
		{
			name:   "nil value - not allowed",
			value:  nil,
			config: &ValidationConfig{AllowNilValues: false},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				Depth:     0,
				FieldPath: "test",
				Config:    tt.config,
			}
			result := validateMapValue(tt.value, ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateArrayValue(t *testing.T) {
	tests := []struct {
		name   string
		array  []interface{}
		config *ValidationConfig
		want   bool
	}{
		{
			name:   "valid array with strings",
			array:  []interface{}{"item1", "item2", "item3"},
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "valid array with mixed types",
			array:  []interface{}{"string", 42, true},
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "empty array - allowed",
			array:  []interface{}{},
			config: &ValidationConfig{AllowEmptyMaps: true},
			want:   true,
		},
		{
			name:   "empty array - not allowed",
			array:  []interface{}{},
			config: &ValidationConfig{AllowEmptyMaps: false},
			want:   false,
		},
		{
			name:   "nil array - allowed",
			array:  nil,
			config: &ValidationConfig{AllowNilValues: true},
			want:   true,
		},
		{
			name:   "nil array - not allowed",
			array:  nil,
			config: &ValidationConfig{AllowNilValues: false},
			want:   false,
		},
		{
			name: "array with empty string",
			array: []interface{}{
				"valid_item",
				"",
				"another_item",
			},
			config: DefaultConfig(),
			want:   false,
		},
		{
			name: "array with nil element",
			array: []interface{}{
				"valid_item",
				nil,
				"another_item",
			},
			config: DefaultConfig(),
			want:   false,
		},
		{
			name: "nested array",
			array: []interface{}{
				[]interface{}{"nested1", "nested2"},
				"regular_item",
			},
			config: DefaultConfig(),
			want:   true,
		},
		{
			name: "array with nested map",
			array: []interface{}{
				map[string]interface{}{
					"nested_key": "nested_value",
				},
			},
			config: DefaultConfig(),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ValidationContext{
				Depth:     0,
				FieldPath: "test",
				Config:    tt.config,
			}
			result := validateArrayValue(tt.array, ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateNestedMapStructure(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		config  *ValidationConfig
		wantErr bool
	}{
		{
			name: "valid structure",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			config:  nil, // will use default
			wantErr: false,
		},
		{
			name: "invalid structure",
			input: map[string]interface{}{
				"": "empty_key",
			},
			config:  nil,
			wantErr: true,
		},
		{
			name:  "custom config - allow empty",
			input: map[string]interface{}{},
			config: &ValidationConfig{
				AllowEmptyMaps: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNestedMapStructure(tt.input, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWithSchema(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		schema  map[string]*FieldSchema
		wantErr bool
	}{
		{
			name: "valid structure matches schema",
			input: map[string]interface{}{
				"name": "John Doe",
				"age":  30,
			},
			schema: map[string]*FieldSchema{
				"name": {Type: "string", Required: true},
				"age":  {Type: "integer", Required: true},
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			input: map[string]interface{}{
				"name": "John Doe",
				// age is missing
			},
			schema: map[string]*FieldSchema{
				"name": {Type: "string", Required: true},
				"age":  {Type: "integer", Required: true},
			},
			wantErr: true,
		},
		{
			name: "wrong type",
			input: map[string]interface{}{
				"name": "John Doe",
				"age":  "thirty", // should be number
			},
			schema: map[string]*FieldSchema{
				"name": {Type: "string", Required: true},
				"age":  {Type: "integer", Required: true},
			},
			wantErr: true,
		},
		{
			name: "nested schema validation",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John Doe",
					"age":  30,
				},
			},
			schema: map[string]*FieldSchema{
				"user": {
					Type:     "object",
					Required: true,
					Children: map[string]*FieldSchema{
						"name": {Type: "string", Required: true},
						"age":  {Type: "integer", Required: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "optional field missing",
			input: map[string]interface{}{
				"name": "John Doe",
				// optional field missing
			},
			schema: map[string]*FieldSchema{
				"name":    {Type: "string", Required: true},
				"address": {Type: "string", Required: false},
			},
			wantErr: false,
		},
		{
			name: "any type field",
			input: map[string]interface{}{
				"data": []interface{}{1, 2, 3},
			},
			schema: map[string]*FieldSchema{
				"data": {Type: "any", Required: true},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithSchema(tt.input, tt.schema)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetValueType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{name: "nil", value: nil, expected: "null"},
		{name: "string", value: "test", expected: "string"},
		{name: "int", value: 42, expected: "integer"},
		{name: "int64", value: int64(42), expected: "integer"},
		{name: "float64", value: 42.5, expected: "number"},
		{name: "bool", value: true, expected: "boolean"},
		{name: "array", value: []interface{}{}, expected: "array"},
		{name: "object", value: map[string]interface{}{}, expected: "object"},
		{name: "unknown", value: func() {}, expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getValueType(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFastValidateType(t *testing.T) {
	ctx := &ValidationContext{
		Config: DefaultConfig(),
	}

	tests := []struct {
		name         string
		value        interface{}
		expectedType string
		want         bool
	}{
		{name: "valid string", value: "test", expectedType: "string", want: true},
		{name: "invalid string type", value: 42, expectedType: "string", want: false},
		{name: "valid number", value: 42, expectedType: "number", want: true},
		{name: "valid float number", value: 42.5, expectedType: "number", want: true},
		{name: "invalid number type", value: "test", expectedType: "number", want: false},
		{name: "valid boolean", value: true, expectedType: "boolean", want: true},
		{name: "invalid boolean type", value: 1, expectedType: "boolean", want: false},
		{name: "valid array", value: []interface{}{1, 2, 3}, expectedType: "array", want: true},
		{name: "valid object", value: map[string]interface{}{"key": "value"}, expectedType: "object", want: true},
		{name: "unknown type", value: "test", expectedType: "unknown", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FastValidateType(tt.value, tt.expectedType, ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Helper functions for tests
func createDeeplyNestedMap(depth int) map[string]interface{} {
	if depth <= 0 {
		return map[string]interface{}{
			"end": "value",
		}
	}

	return map[string]interface{}{
		"level": createDeeplyNestedMap(depth - 1),
	}
}

// Benchmark tests
func BenchmarkValidateNestedMapSimple(b *testing.B) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 42,
	}

	ctx := &ValidationContext{
		Config: DefaultConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateNestedMapWithContext(m, ctx)
	}
}

func BenchmarkValidateNestedMapComplex(b *testing.B) {
	m := map[string]interface{}{
		"user": map[string]interface{}{
			"name":   "John Doe",
			"age":    30,
			"active": true,
			"tags":   []interface{}{"developer", "golang"},
			"preferences": map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
		},
		"metadata": map[string]interface{}{
			"version":    1,
			"created_at": "2023-01-01",
		},
	}

	ctx := &ValidationContext{
		Config: DefaultConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateNestedMapWithContext(m, ctx)
	}
}

func BenchmarkValidateNestedMapDeep(b *testing.B) {
	m := createDeeplyNestedMap(8)

	ctx := &ValidationContext{
		Config: DefaultConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateNestedMapWithContext(m, ctx)
	}
}
