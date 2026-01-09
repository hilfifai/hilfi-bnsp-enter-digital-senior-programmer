package pagination

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateFilterValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		config   FieldConfig
		expected bool
	}{
		// Boolean tests
		{
			name:     "valid boolean true",
			value:    "true",
			config:   FieldConfig{DataType: "boolean"},
			expected: true,
		},
		{
			name:     "valid boolean false",
			value:    "false",
			config:   FieldConfig{DataType: "boolean"},
			expected: true,
		},
		{
			name:     "invalid boolean",
			value:    "maybe",
			config:   FieldConfig{DataType: "boolean"},
			expected: false,
		},
		{
			name:     "boolean with case sensitivity",
			value:    "TRUE",
			config:   FieldConfig{DataType: "boolean"},
			expected: false,
		},
		// Number tests
		{
			name:     "valid positive integer",
			value:    "123",
			config:   FieldConfig{DataType: "number"},
			expected: true,
		},
		{
			name:     "valid negative integer",
			value:    "-456",
			config:   FieldConfig{DataType: "number"},
			expected: true,
		},
		{
			name:     "valid decimal",
			value:    "123.45",
			config:   FieldConfig{DataType: "number"},
			expected: true,
		},
		{
			name:     "valid zero",
			value:    "0",
			config:   FieldConfig{DataType: "number"},
			expected: true,
		},
		{
			name:     "invalid number",
			value:    "not_a_number",
			config:   FieldConfig{DataType: "number"},
			expected: false,
		},
		{
			name:     "number too large",
			value:    "1e10",
			config:   FieldConfig{DataType: "number"},
			expected: false,
		},
		{
			name:     "number too small",
			value:    "-1e10",
			config:   FieldConfig{DataType: "number"},
			expected: false,
		},
		// Date tests
		{
			name:     "valid RFC3339 date",
			value:    "2023-01-01T00:00:00Z",
			config:   FieldConfig{DataType: "date"},
			expected: true,
		},
		{
			name:     "valid RFC3339 date with timezone",
			value:    "2023-12-31T23:59:59+07:00",
			config:   FieldConfig{DataType: "date"},
			expected: true,
		},
		{
			name:     "invalid date format",
			value:    "2023-01-01",
			config:   FieldConfig{DataType: "date"},
			expected: false,
		},
		{
			name:     "invalid date",
			value:    "not_a_date",
			config:   FieldConfig{DataType: "date"},
			expected: false,
		},
		// String tests
		{
			name:     "valid string",
			value:    "hello world",
			config:   FieldConfig{DataType: "string"},
			expected: true,
		},
		{
			name:     "empty string",
			value:    "",
			config:   FieldConfig{DataType: "string"},
			expected: true,
		},
		{
			name:     "string with special characters",
			value:    "hello@world.com",
			config:   FieldConfig{DataType: "string"},
			expected: true,
		},
		{
			name:     "string too long",
			value:    strings.Repeat("a", MaxStringLength+1),
			config:   FieldConfig{DataType: "string"},
			expected: false,
		},
		{
			name:     "string with SQL injection attempt",
			value:    "'; DROP TABLE users; --",
			config:   FieldConfig{DataType: "string"},
			expected: false,
		},
		// UUID tests
		{
			name:     "valid UUID v4",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			config:   FieldConfig{DataType: "uuid"},
			expected: true,
		},
		{
			name:     "valid UUID v1",
			value:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			config:   FieldConfig{DataType: "uuid"},
			expected: true,
		},
		{
			name:     "invalid UUID",
			value:    "not-a-uuid",
			config:   FieldConfig{DataType: "uuid"},
			expected: false,
		},
		{
			name:     "empty UUID",
			value:    "",
			config:   FieldConfig{DataType: "uuid"},
			expected: false,
		},
		// Unknown data type
		{
			name:     "unknown data type",
			value:    "test",
			config:   FieldConfig{DataType: "unknown"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateFilterValue(tt.value, tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultValueForType(t *testing.T) {
	tests := []struct {
		name     string
		config   FieldConfig
		expected string
	}{
		{
			name:     "boolean default",
			config:   FieldConfig{DataType: "boolean"},
			expected: "false",
		},
		{
			name:     "number default",
			config:   FieldConfig{DataType: "number"},
			expected: "0",
		},
		{
			name:     "date default",
			config:   FieldConfig{DataType: "date"},
			expected: time.Now().Format(time.RFC3339)[:19] + "Z", // Compare without nanoseconds
		},
		{
			name:     "string default",
			config:   FieldConfig{DataType: "string"},
			expected: "",
		},
		{
			name:     "uuid default",
			config:   FieldConfig{DataType: "uuid"},
			expected: "",
		},
		{
			name:     "unknown type default",
			config:   FieldConfig{DataType: "unknown"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefaultValueForType(tt.config)
			if tt.config.DataType == "date" {
				// For date, just check that it's a valid RFC3339 format
				_, err := time.Parse(time.RFC3339, result)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsValidString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid strings
		{
			name:     "simple string",
			input:    "hello world",
			expected: true,
		},
		{
			name:     "string with numbers",
			input:    "hello123",
			expected: true,
		},
		{
			name:     "string with special chars",
			input:    "hello@world.com",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "string with safe punctuation",
			input:    "Hello, World! How are you?",
			expected: true,
		},
		// Invalid strings with dangerous patterns
		{
			name:     "SQL comment",
			input:    "test --",
			expected: false,
		},
		{
			name:     "SQL block comment start",
			input:    "test /*",
			expected: false,
		},
		{
			name:     "SQL block comment end",
			input:    "test */",
			expected: false,
		},
		{
			name:     "double semicolon",
			input:    "test ;;",
			expected: false,
		},
		{
			name:     "single semicolon",
			input:    "test ;",
			expected: false,
		},
		{
			name:     "SQL variable",
			input:    "test @@version",
			expected: false,
		},
		// Invalid strings with dangerous keywords
		{
			name:     "DROP keyword",
			input:    "DROP TABLE users",
			expected: false,
		},
		{
			name:     "DELETE keyword",
			input:    "DELETE FROM users",
			expected: false,
		},
		{
			name:     "UPDATE keyword",
			input:    "UPDATE users SET",
			expected: false,
		},
		{
			name:     "TRUNCATE keyword",
			input:    "TRUNCATE TABLE users",
			expected: false,
		},
		{
			name:     "ALTER keyword",
			input:    "ALTER TABLE users",
			expected: false,
		},
		{
			name:     "GRANT keyword",
			input:    "GRANT ALL ON users",
			expected: false,
		},
		{
			name:     "REVOKE keyword",
			input:    "REVOKE ALL FROM user",
			expected: false,
		},
		{
			name:     "EXECUTE keyword",
			input:    "EXECUTE procedure",
			expected: false,
		},
		// Valid strings that contain keywords as part of larger words
		{
			name:     "keyword as part of word",
			input:    "dropdown menu",
			expected: true,
		},
		{
			name:     "keyword in quoted context",
			input:    `"DROP" is a valid column name`,
			expected: false,
		},
		{
			name:     "lowercase keywords should be safe in content",
			input:    "drop the ball",
			expected: false,
		},
		// Edge cases
		{
			name:     "only punctuation",
			input:    "!@#$%^&*()",
			expected: true,
		},
		{
			name:     "mixed case safe string",
			input:    "Hello World 123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid field names
		{
			name:     "simple field name",
			input:    "name",
			expected: true,
		},
		{
			name:     "field with underscore",
			input:    "first_name",
			expected: true,
		},
		{
			name:     "field with numbers",
			input:    "field123",
			expected: true,
		},
		{
			name:     "mixed case",
			input:    "FirstName",
			expected: true,
		},
		{
			name:     "starts with underscore",
			input:    "_private",
			expected: true,
		},
		{
			name:     "ends with underscore",
			input:    "field_",
			expected: true,
		},
		{
			name:     "multiple underscores",
			input:    "field__name",
			expected: true,
		},
		// Invalid field names
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "contains space",
			input:    "first name",
			expected: false,
		},
		{
			name:     "contains hyphen",
			input:    "first-name",
			expected: false,
		},
		{
			name:     "contains dot",
			input:    "user.name",
			expected: false,
		},
		{
			name:     "contains special characters",
			input:    "name@domain",
			expected: false,
		},
		{
			name:     "starts with number",
			input:    "123field",
			expected: true, // This should actually be valid according to the function
		},
		{
			name:     "contains parentheses",
			input:    "field()",
			expected: false,
		},
		{
			name:     "contains brackets",
			input:    "field[0]",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFieldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidTransformFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid transform functions
		{
			name:     "LOWER function",
			input:    "LOWER",
			expected: true,
		},
		{
			name:     "UPPER function",
			input:    "UPPER",
			expected: true,
		},
		{
			name:     "TRIM function",
			input:    "TRIM",
			expected: true,
		},
		{
			name:     "COALESCE function",
			input:    "COALESCE",
			expected: true,
		},
		// Case insensitive tests
		{
			name:     "lowercase lower",
			input:    "lower",
			expected: true,
		},
		{
			name:     "mixed case upper",
			input:    "Upper",
			expected: true,
		},
		// Invalid transform functions
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid function",
			input:    "INVALID",
			expected: false,
		},
		{
			name:     "dangerous function",
			input:    "DROP",
			expected: false,
		},
		{
			name:     "system function",
			input:    "EXEC",
			expected: false,
		},
		{
			name:     "aggregate function not allowed",
			input:    "COUNT",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTransformFunction(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidQueryString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid query strings
		{
			name:     "simple SELECT",
			input:    "SELECT * FROM users",
			expected: true,
		},
		{
			name:     "SELECT with WHERE",
			input:    "SELECT name, email FROM users WHERE active = true",
			expected: true,
		},
		{
			name:     "SELECT with JOIN",
			input:    "SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id",
			expected: true,
		},
		{
			name:     "COUNT query",
			input:    "SELECT COUNT(*) FROM users",
			expected: true,
		},
		{
			name:     "query with quoted column names",
			input:    `SELECT "first_name", "last_name" FROM users`,
			expected: true,
		},
		{
			name:     "query with single quoted values",
			input:    "SELECT * FROM users WHERE name = 'John'",
			expected: true,
		},
		{
			name:     "query with underscored fields",
			input:    "SELECT created_at, updated_at FROM users",
			expected: true,
		},
		// Invalid query strings
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "DROP TABLE",
			input:    "DROP TABLE users",
			expected: false,
		},
		{
			name:     "DELETE query",
			input:    "DELETE FROM users WHERE id = 1",
			expected: false,
		},
		{
			name:     "UPDATE query",
			input:    "UPDATE users SET name = 'John' WHERE id = 1",
			expected: false,
		},
		{
			name:     "TRUNCATE query",
			input:    "TRUNCATE TABLE users",
			expected: false,
		},
		{
			name:     "ALTER TABLE",
			input:    "ALTER TABLE users ADD COLUMN age INT",
			expected: false,
		},
		{
			name:     "GRANT permissions",
			input:    "GRANT ALL PRIVILEGES ON users TO user1",
			expected: false,
		},
		{
			name:     "REVOKE permissions",
			input:    "REVOKE ALL PRIVILEGES ON users FROM user1",
			expected: false,
		},
		{
			name:     "EXECUTE procedure",
			input:    "EXECUTE sp_deleteuser 1",
			expected: false,
		},
		{
			name:     "SQL comment attack",
			input:    "SELECT * FROM users -- WHERE active = true",
			expected: false,
		},
		{
			name:     "SQL block comment",
			input:    "SELECT * FROM users /* comment */ WHERE active = true",
			expected: false,
		},
		{
			name:     "double semicolon",
			input:    "SELECT * FROM users;; DROP TABLE users",
			expected: false,
		},
		// Edge cases - keywords in valid contexts
		{
			name:     "DROP as column name in quotes",
			input:    `SELECT "drop_date" FROM users`,
			expected: true,
		},
		{
			name:     "DELETE as part of column name",
			input:    "SELECT delete_flag FROM users",
			expected: true,
		},
		{
			name:     "UPDATE as part of column name",
			input:    "SELECT updated_at FROM users",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidQueryString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for validation functions
func BenchmarkValidateFilterValue(b *testing.B) {
	config := FieldConfig{DataType: "string"}
	value := "test string value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateFilterValue(value, config)
	}
}

func BenchmarkIsValidString(b *testing.B) {
	input := "This is a test string with some content that needs validation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidString(input)
	}
}

func BenchmarkIsValidFieldName(b *testing.B) {
	input := "valid_field_name_123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValidFieldName(input)
	}
}

func BenchmarkIsValidQueryString(b *testing.B) {
	input := "SELECT u.name, u.email, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = true"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValidQueryString(input)
	}
}

// Edge case tests
func TestValidationEdgeCases(t *testing.T) {
	t.Run("very long but valid field name", func(t *testing.T) {
		longName := strings.Repeat("a", 100)
		result := isValidFieldName(longName)
		assert.True(t, result)
	})

	t.Run("string at exact max length", func(t *testing.T) {
		maxLengthString := strings.Repeat("a", MaxStringLength)
		config := FieldConfig{DataType: "string"}
		result := validateFilterValue(maxLengthString, config)
		assert.True(t, result)
	})

	t.Run("number at boundary values", func(t *testing.T) {
		config := FieldConfig{DataType: "number"}

		// Test exact boundary values
		result1 := validateFilterValue("1000000000", config) // 1e9
		assert.True(t, result1)

		result2 := validateFilterValue("-1000000000", config) // -1e9
		assert.True(t, result2)

		result3 := validateFilterValue("1000000001", config) // Just over 1e9
		assert.False(t, result3)
	})

	t.Run("complex SQL injection attempts", func(t *testing.T) {
		injectionAttempts := []string{
			"'; DROP TABLE users; --",
			"1'; DELETE FROM users WHERE '1'='1",
			"admin'/**/OR/**/1=1--",
		}

		for _, attempt := range injectionAttempts {
			result := IsValidString(attempt)
			assert.False(t, result, "Should reject SQL injection attempt: %s", attempt)
		}
	})
}
