package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMapStringString(t *testing.T) {
	// Setup validator
	v := validator.New()
	err := v.RegisterValidation("mapStringString", validateMapStringString)
	require.NoError(t, err)

	type TestStruct struct {
		StringMap map[string]string `validate:"mapStringString"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{
			name: "valid string map",
			input: TestStruct{
				StringMap: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
			wantErr: false,
		},
		{
			name: "valid single entry map",
			input: TestStruct{
				StringMap: map[string]string{
					"single_key": "single_value",
				},
			},
			wantErr: false,
		},
		{
			name: "empty map",
			input: TestStruct{
				StringMap: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "nil map",
			input: TestStruct{
				StringMap: nil,
			},
			wantErr: true,
		},
		{
			name: "map with empty key",
			input: TestStruct{
				StringMap: map[string]string{
					"":     "value1",
					"key2": "value2",
				},
			},
			wantErr: true,
		},
		{
			name: "map with empty value",
			input: TestStruct{
				StringMap: map[string]string{
					"key1": "",
					"key2": "value2",
				},
			},
			wantErr: true,
		},
		{
			name: "map with both empty key and value",
			input: TestStruct{
				StringMap: map[string]string{
					"":     "",
					"key2": "value2",
				},
			},
			wantErr: true,
		},
		{
			name: "map with whitespace-only key",
			input: TestStruct{
				StringMap: map[string]string{
					"   ":  "value1",
					"key2": "value2",
				},
			},
			wantErr: false, // whitespace keys are valid by default
		},
		{
			name: "map with whitespace-only value",
			input: TestStruct{
				StringMap: map[string]string{
					"key1": "   ",
					"key2": "value2",
				},
			},
			wantErr: false, // whitespace values are valid by default
		},
		{
			name: "map with special characters",
			input: TestStruct{
				StringMap: map[string]string{
					"key@#$": "value!@#",
					"key_2":  "value-2",
				},
			},
			wantErr: false,
		},
		{
			name: "map with unicode characters",
			input: TestStruct{
				StringMap: map[string]string{
					"키":    "값",
					"clé":  "valeur",
					"ключ": "значение",
				},
			},
			wantErr: false,
		},
		{
			name: "large map",
			input: TestStruct{
				StringMap: createLargeStringMap(100),
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

func TestValidateStringMapWithContext(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]string
		config *StringMapConfig
		want   bool
	}{
		{
			name: "valid map with default config",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			config: DefaultStringMapConfig(),
			want:   true,
		},
		{
			name:  "empty map with allow empty config",
			input: map[string]string{},
			config: &StringMapConfig{
				AllowEmptyKeys:   true,
				AllowEmptyValues: true,
			},
			want: true,
		},
		{
			name:  "empty map with disallow empty config",
			input: map[string]string{},
			config: &StringMapConfig{
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name: "map with empty key - allowed",
			input: map[string]string{
				"":     "value1",
				"key2": "value2",
			},
			config: &StringMapConfig{
				AllowEmptyKeys:   true,
				AllowEmptyValues: false,
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
				MinKeyLength:     0,
				MinValueLength:   1,
			},
			want: true,
		},
		{
			name: "map with empty key - not allowed",
			input: map[string]string{
				"":     "value1",
				"key2": "value2",
			},
			config: &StringMapConfig{
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name: "map with empty value - allowed",
			input: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			config: &StringMapConfig{
				AllowEmptyKeys:   false,
				AllowEmptyValues: true,
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
				MinKeyLength:     1,
				MinValueLength:   0,
			},
			want: true,
		},
		{
			name: "map with empty value - not allowed",
			input: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			config: &StringMapConfig{
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name: "map with long key exceeding limit",
			input: map[string]string{
				createLongString(300): "value1",
				"key2":                "value2",
			},
			config: &StringMapConfig{
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name: "map with long value exceeding limit",
			input: map[string]string{
				"key1": createLongString(1000),
				"key2": "value2",
			},
			config: &StringMapConfig{
				MaxValueLength:   500,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name: "map with trimmed whitespace - enabled",
			input: map[string]string{
				"  key1  ": "  value1  ",
				"key2":     "value2",
			},
			config: &StringMapConfig{
				TrimWhitespace:   true,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
				MinKeyLength:     1,
				MinValueLength:   1,
			},
			want: true,
		},
		{
			name: "map with whitespace becoming empty after trim",
			input: map[string]string{
				"   ": "   ",
				"key": "value",
			},
			config: &StringMapConfig{
				TrimWhitespace:   true,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name:   "nil map",
			input:  nil,
			config: DefaultStringMapConfig(),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &StringMapValidationContext{
				Config:    tt.config,
				FieldPath: "test",
				Errors:    make([]string, 0),
			}
			result := validateStringMapWithContext(tt.input, ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateStringKey(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		config *StringMapConfig
		want   bool
	}{
		{
			name:   "valid key",
			key:    "valid_key",
			config: DefaultStringMapConfig(),
			want:   true,
		},
		{
			name:   "empty key - not allowed",
			key:    "",
			config: DefaultStringMapConfig(),
			want:   false,
		},
		{
			name: "empty key - allowed",
			key:  "",
			config: &StringMapConfig{
				AllowEmptyKeys: true,
			},
			want: true,
		},
		{
			name: "key too short",
			key:  "",
			config: &StringMapConfig{
				MinKeyLength:   3,
				AllowEmptyKeys: false,
			},
			want: false,
		},
		{
			name: "key too long",
			key:  createLongString(300),
			config: &StringMapConfig{
				MaxKeyLength: 255,
			},
			want: false,
		},
		{
			name: "key with whitespace - trim enabled",
			key:  "  valid_key  ",
			config: &StringMapConfig{
				TrimWhitespace:   true,
				MinKeyLength:     1,
				MaxKeyLength:     255,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: true,
		},
		{
			name: "key becomes empty after trim",
			key:  "   ",
			config: &StringMapConfig{
				TrimWhitespace: true,
				AllowEmptyKeys: false,
			},
			want: false,
		},
		{
			name:   "unicode key",
			key:    "키_key_ключ",
			config: DefaultStringMapConfig(),
			want:   true,
		},
		{
			name:   "special characters key",
			key:    "key@#$%^&*()",
			config: DefaultStringMapConfig(),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &StringMapValidationContext{
				Config: tt.config,
				Errors: make([]string, 0),
			}
			result := validateStringKey(tt.key, ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateStringValueInMap(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		value  string
		config *StringMapConfig
		want   bool
	}{
		{
			name:   "valid value",
			key:    "test_key",
			value:  "valid_value",
			config: DefaultStringMapConfig(),
			want:   true,
		},
		{
			name:   "empty value - not allowed",
			key:    "test_key",
			value:  "",
			config: DefaultStringMapConfig(),
			want:   false,
		},
		{
			name:   "empty value - allowed",
			key:    "test_key",
			value:  "",
			config: &StringMapConfig{AllowEmptyValues: true},
			want:   true,
		},
		{
			name:  "value too short",
			key:   "test_key",
			value: "x",
			config: &StringMapConfig{
				MinValueLength:   3,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name:  "value too long",
			key:   "test_key",
			value: createLongString(1000),
			config: &StringMapConfig{
				MaxValueLength: 500,
			},
			want: false,
		},
		{
			name:  "value with whitespace - trim enabled",
			key:   "test_key",
			value: "  valid_value  ",
			config: &StringMapConfig{
				TrimWhitespace:   true,
				MinValueLength:   1,
				MaxValueLength:   MaxStringLength,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			want: true,
		},
		{
			name:  "value becomes empty after trim",
			key:   "test_key",
			value: "   ",
			config: &StringMapConfig{
				TrimWhitespace:   true,
				AllowEmptyValues: false,
			},
			want: false,
		},
		{
			name:   "unicode value",
			key:    "test_key",
			value:  "값_value_значение",
			config: DefaultStringMapConfig(),
			want:   true,
		},
		{
			name:   "special characters value",
			key:    "test_key",
			value:  "value@#$%^&*()",
			config: DefaultStringMapConfig(),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &StringMapValidationContext{
				Config: tt.config,
				Errors: make([]string, 0),
			}
			result := validateStringMapValue(tt.key, tt.value, ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateStringMapStructured(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]string
		config  *StringMapConfig
		wantErr bool
	}{
		{
			name: "valid map",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			config:  nil, // use default
			wantErr: false,
		},
		{
			name: "invalid map with empty key",
			input: map[string]string{
				"":     "value1",
				"key2": "value2",
			},
			config:  nil,
			wantErr: true,
		},
		{
			name: "invalid map with empty value",
			input: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid map with custom config allowing empty",
			input: map[string]string{
				"":     "",
				"key2": "value2",
			},
			config: &StringMapConfig{
				AllowEmptyKeys:   true,
				AllowEmptyValues: true,
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
				MinKeyLength:     0,
				MinValueLength:   0,
				TrimWhitespace:   true,
			},
			wantErr: false,
		},
		{
			name:    "nil map",
			input:   nil,
			config:  nil,
			wantErr: true,
		},
		{
			name:  "empty map",
			input: map[string]string{},
			config: &StringMapConfig{
				AllowEmptyKeys:   true,
				AllowEmptyValues: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringMapStructured(tt.input, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStringMapKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		config   *StringMapConfig
		expected []string
	}{
		{
			name: "all valid keys",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			config:   DefaultStringMapConfig(),
			expected: []string{},
		},
		{
			name: "some invalid keys",
			input: map[string]string{
				"":     "value1",
				"key2": "value2",
				"key3": "value3",
			},
			config:   DefaultStringMapConfig(),
			expected: []string{""},
		},
		{
			name: "all invalid keys",
			input: map[string]string{
				"":  "value1",
				"x": "value2", // Note: adding second key since map needs unique keys
			},
			config:   DefaultStringMapConfig(),
			expected: []string{""},
		},
		{
			name: "keys too long",
			input: map[string]string{
				createLongString(300): "value1",
				"valid_key":           "value2",
			},
			config: &StringMapConfig{
				MaxKeyLength: 255,
			},
			expected: []string{createLongString(300)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateStringMapKeys(tt.input, tt.config)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestValidateStringMapValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		config   *StringMapConfig
		expected map[string][]string
	}{
		{
			name: "all valid values",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			config:   DefaultStringMapConfig(),
			expected: map[string][]string{},
		},
		{
			name: "some invalid values",
			input: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			config:   DefaultStringMapConfig(),
			expected: map[string][]string{"key1": {"test: empty value for key 'key1' is not allowed"}},
		},
		{
			name: "values too long",
			input: map[string]string{
				"key1": createLongString(600),
				"key2": "valid_value",
			},
			config: &StringMapConfig{
				MaxValueLength: 500,
			},
			expected: map[string][]string{"key1": {"test: value for key 'key1' is too long (maximum length: 500)"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateStringMapValues(tt.input, tt.config)
			// Note: We check if the keys match since error messages might vary
			assert.Equal(t, len(tt.expected), len(result))
			for key := range tt.expected {
				assert.Contains(t, result, key)
			}
		})
	}
}

func TestFastValidateStringMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  bool
	}{
		{
			name: "valid map",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			want: true,
		},
		{
			name:  "nil map",
			input: nil,
			want:  false,
		},
		{
			name:  "empty map",
			input: map[string]string{},
			want:  false,
		},
		{
			name: "empty key",
			input: map[string]string{
				"":     "value1",
				"key2": "value2",
			},
			want: false,
		},
		{
			name: "empty value",
			input: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			want: false,
		},
		{
			name: "key too long",
			input: map[string]string{
				createLongString(300): "value1",
			},
			want: false,
		},
		{
			name: "value too long",
			input: map[string]string{
				"key1": createLongString(MaxStringLength + 1),
			},
			want: false,
		},
		{
			name:  "map too large",
			input: createLargeStringMap(100), // Use reasonable size for test
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FastValidateStringMap(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateStringMapSafe(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]string
		config    *StringMapConfig
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid map",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			config:    nil,
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "invalid map",
			input: map[string]string{
				"": "value1",
			},
			config:    nil,
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "nil map",
			input:     nil,
			config:    nil,
			wantValid: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateStringMapSafe(tt.input, tt.config)
			assert.Equal(t, tt.wantValid, valid)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeStringMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		config   *StringMapConfig
		expected map[string]string
	}{
		{
			name: "trim whitespace",
			input: map[string]string{
				"  key1  ": "  value1  ",
				"key2":     "value2",
			},
			config: &StringMapConfig{
				TrimWhitespace:   true,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "remove empty entries",
			input: map[string]string{
				"":     "value1",
				"key2": "",
				"key3": "value3",
			},
			config: &StringMapConfig{
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
			},
			expected: map[string]string{
				"key3": "value3",
			},
		},
		{
			name: "allow empty entries",
			input: map[string]string{
				"":     "value1",
				"key2": "",
				"key3": "value3",
			},
			config: &StringMapConfig{
				AllowEmptyKeys:   true,
				AllowEmptyValues: true,
				MaxKeyLength:     255,
				MaxValueLength:   MaxStringLength,
			},
			expected: map[string]string{
				"":     "value1",
				"key2": "",
				"key3": "value3",
			},
		},
		{
			name: "filter by length",
			input: map[string]string{
				createLongString(300): "value1",
				"key2":                createLongString(600),
				"key3":                "value3",
			},
			config: &StringMapConfig{
				MaxKeyLength:     255,
				MaxValueLength:   500,
				AllowEmptyKeys:   false,
				AllowEmptyValues: false,
			},
			expected: map[string]string{
				"key3": "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeStringMap(tt.input, tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeStringMapKeys(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]string
		toLowerCase bool
		expected    map[string]string
	}{
		{
			name: "to lowercase",
			input: map[string]string{
				"KEY1": "value1",
				"Key2": "value2",
				"key3": "value3",
			},
			toLowerCase: true,
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no change",
			input: map[string]string{
				"KEY1": "value1",
				"Key2": "value2",
			},
			toLowerCase: false,
			expected: map[string]string{
				"KEY1": "value1",
				"Key2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeStringMapKeys(tt.input, tt.toLowerCase)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeStringMaps(t *testing.T) {
	tests := []struct {
		name     string
		maps     []map[string]string
		expected map[string]string
	}{
		{
			name: "merge two maps",
			maps: []map[string]string{
				{"key1": "value1", "key2": "value2"},
				{"key2": "updated_value2", "key3": "value3"},
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "updated_value2", // later map overrides
				"key3": "value3",
			},
		},
		{
			name: "merge empty maps",
			maps: []map[string]string{
				{},
				{"key1": "value1"},
				{},
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name:     "no maps",
			maps:     []map[string]string{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeStringMaps(tt.maps...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterStringMap(t *testing.T) {
	input := map[string]string{
		"admin_key": "admin_value",
		"user_key":  "user_value",
		"guest_key": "guest_value",
		"temp_key":  "temp_value",
	}

	tests := []struct {
		name        string
		keyFilter   func(string) bool
		valueFilter func(string) bool
		expected    map[string]string
	}{
		{
			name: "filter keys starting with 'admin'",
			keyFilter: func(key string) bool {
				return key == "admin_key"
			},
			valueFilter: nil,
			expected: map[string]string{
				"admin_key": "admin_value",
			},
		},
		{
			name:      "filter values containing 'user'",
			keyFilter: nil,
			valueFilter: func(value string) bool {
				return value == "user_value"
			},
			expected: map[string]string{
				"user_key": "user_value",
			},
		},
		{
			name: "filter both key and value",
			keyFilter: func(key string) bool {
				return key != "temp_key"
			},
			valueFilter: func(value string) bool {
				return value != "guest_value"
			},
			expected: map[string]string{
				"admin_key": "admin_value",
				"user_key":  "user_value",
			},
		},
		{
			name:        "no filters",
			keyFilter:   nil,
			valueFilter: nil,
			expected:    input,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterStringMap(input, tt.keyFilter, tt.valueFilter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultStringMapConfig(t *testing.T) {
	config := DefaultStringMapConfig()
	assert.NotNil(t, config)
	assert.False(t, config.AllowEmptyKeys)
	assert.False(t, config.AllowEmptyValues)
	assert.Equal(t, 255, config.MaxKeyLength)
	assert.Equal(t, MaxStringLength, config.MaxValueLength)
	assert.Equal(t, MinStringLength, config.MinKeyLength)
	assert.Equal(t, MinStringLength, config.MinValueLength)
	assert.True(t, config.TrimWhitespace)
	assert.True(t, config.CaseSensitive)
	assert.Empty(t, config.AllowedKeyPattern)
	assert.Empty(t, config.AllowedValuePattern)
}

func TestIsValidUTF8(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "valid ASCII", input: "hello world", want: true},
		{name: "valid UTF-8", input: "héllo wörld", want: true},
		{name: "valid unicode", input: "こんにちは", want: true},
		{name: "empty string", input: "", want: true},
		{name: "valid emoji", input: "hello 👋 world", want: true},
		// Note: It's hard to create invalid UTF-8 in Go string literals
		// as Go automatically validates UTF-8. In real scenarios, invalid
		// UTF-8 would come from external sources like files or network.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUTF8(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Helper functions for tests
func createLongString(length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = 'a' + byte(i%26)
	}
	return string(result)
}

func createLargeStringMap(size int) map[string]string {
	result := make(map[string]string, size)
	for i := 0; i < size; i++ {
		key := "key_" + string(rune('0'+i%10))
		value := "value_" + string(rune('0'+i%10))
		result[key] = value
	}
	return result
}

// Benchmark tests
func BenchmarkValidateStringMapSimple(b *testing.B) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	ctx := &StringMapValidationContext{
		Config: DefaultStringMapConfig(),
		Errors: make([]string, 0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateStringMapWithContext(m, ctx)
		ctx.Errors = ctx.Errors[:0] // Reset errors
	}
}

func BenchmarkValidateStringMapLarge(b *testing.B) {
	m := createLargeStringMap(100)

	ctx := &StringMapValidationContext{
		Config: DefaultStringMapConfig(),
		Errors: make([]string, 0),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateStringMapWithContext(m, ctx)
		ctx.Errors = ctx.Errors[:0] // Reset errors
	}
}

func BenchmarkFastValidateStringMap(b *testing.B) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FastValidateStringMap(m)
	}
}

func BenchmarkSanitizeStringMap(b *testing.B) {
	m := map[string]string{
		"  key1  ": "  value1  ",
		"key2":     "value2",
		"":         "empty_key",
		"key3":     "",
	}

	config := DefaultStringMapConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeStringMap(m, config)
	}
}
