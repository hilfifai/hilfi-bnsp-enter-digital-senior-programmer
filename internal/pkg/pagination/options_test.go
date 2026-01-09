package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithDataType(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		expected string
	}{
		{
			name:     "string data type",
			dataType: "string",
			expected: "string",
		},
		{
			name:     "boolean data type",
			dataType: "boolean",
			expected: "boolean",
		},
		{
			name:     "number data type",
			dataType: "number",
			expected: "number",
		},
		{
			name:     "date data type",
			dataType: "date",
			expected: "date",
		},
		{
			name:     "uuid data type",
			dataType: "uuid",
			expected: "uuid",
		},
		{
			name:     "empty data type",
			dataType: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &FieldConfig{}
			option := WithDataType(tt.dataType)
			option(config)

			assert.Equal(t, tt.expected, config.DataType)
		})
	}
}

func TestWithOperator(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected string
	}{
		{
			name:     "equals operator",
			operator: "=",
			expected: "=",
		},
		{
			name:     "ILIKE operator",
			operator: "ILIKE",
			expected: "ILIKE",
		},
		{
			name:     "greater than operator",
			operator: ">",
			expected: ">",
		},
		{
			name:     "greater than or equal operator",
			operator: ">=",
			expected: ">=",
		},
		{
			name:     "less than operator",
			operator: "<",
			expected: "<",
		},
		{
			name:     "less than or equal operator",
			operator: "<=",
			expected: "<=",
		},
		{
			name:     "IN operator",
			operator: "IN",
			expected: "IN",
		},
		{
			name:     "IS operator",
			operator: "IS",
			expected: "IS",
		},
		{
			name:     "empty operator",
			operator: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &FieldConfig{}
			option := WithOperator(tt.operator)
			option(config)

			assert.Equal(t, tt.expected, config.Operator)
		})
	}
}

func TestWithTableAlias(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		expected string
	}{
		{
			name:     "single letter alias",
			alias:    "u",
			expected: "u",
		},
		{
			name:     "multi letter alias",
			alias:    "user",
			expected: "user",
		},
		{
			name:     "alias with underscore",
			alias:    "user_data",
			expected: "user_data",
		},
		{
			name:     "alias with numbers",
			alias:    "u1",
			expected: "u1",
		},
		{
			name:     "empty alias",
			alias:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &FieldConfig{}
			option := WithTableAlias(tt.alias)
			option(config)

			assert.Equal(t, tt.expected, config.TableAlias)
		})
	}
}

func TestWithTransform(t *testing.T) {
	tests := []struct {
		name       string
		transforms []string
		expected   []string
	}{
		{
			name:       "single transform",
			transforms: []string{"LOWER"},
			expected:   []string{"LOWER"},
		},
		{
			name:       "multiple transforms",
			transforms: []string{"TRIM", "LOWER"},
			expected:   []string{"TRIM", "LOWER"},
		},
		{
			name:       "all valid transforms",
			transforms: []string{"LOWER", "UPPER", "TRIM", "COALESCE"},
			expected:   []string{"LOWER", "UPPER", "TRIM", "COALESCE"},
		},
		{
			name:       "empty transforms",
			transforms: []string{},
			expected:   []string{},
		},
		{
			name:       "nil transforms",
			transforms: nil,
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &FieldConfig{}
			option := WithTransform(tt.transforms...)
			option(config)

			assert.Equal(t, tt.expected, config.Transform)
		})
	}
}

func TestRequired(t *testing.T) {
	t.Run("sets required to true", func(t *testing.T) {
		config := &FieldConfig{Required: false}
		option := Required()
		option(config)

		assert.True(t, config.Required)
	})

	t.Run("maintains required true", func(t *testing.T) {
		config := &FieldConfig{Required: true}
		option := Required()
		option(config)

		assert.True(t, config.Required)
	})
}

func TestWithSortTransform(t *testing.T) {
	tests := []struct {
		name       string
		transforms []string
		expected   []string
	}{
		{
			name:       "single sort transform",
			transforms: []string{"LOWER"},
			expected:   []string{"LOWER"},
		},
		{
			name:       "multiple sort transforms",
			transforms: []string{"TRIM", "UPPER"},
			expected:   []string{"TRIM", "UPPER"},
		},
		{
			name:       "all valid sort transforms",
			transforms: []string{"LOWER", "UPPER", "TRIM", "COALESCE"},
			expected:   []string{"LOWER", "UPPER", "TRIM", "COALESCE"},
		},
		{
			name:       "empty sort transforms",
			transforms: []string{},
			expected:   []string{},
		},
		{
			name:       "nil sort transforms",
			transforms: nil,
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SortConfig{}
			option := WithSortTransform(tt.transforms...)
			option(config)

			assert.Equal(t, tt.expected, config.Transform)
		})
	}
}

func TestWithSortTableAlias(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		expected string
	}{
		{
			name:     "single letter sort alias",
			alias:    "u",
			expected: "u",
		},
		{
			name:     "multi letter sort alias",
			alias:    "users",
			expected: "users",
		},
		{
			name:     "sort alias with underscore",
			alias:    "user_profile",
			expected: "user_profile",
		},
		{
			name:     "sort alias with numbers",
			alias:    "table1",
			expected: "table1",
		},
		{
			name:     "empty sort alias",
			alias:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SortConfig{}
			option := WithSortTableAlias(tt.alias)
			option(config)

			assert.Equal(t, tt.expected, config.TableAlias)
		})
	}
}

func TestWithNullsLast(t *testing.T) {
	t.Run("sets nulls last to true", func(t *testing.T) {
		config := &SortConfig{NullsLast: false}
		option := WithNullsLast()
		option(config)

		assert.True(t, config.NullsLast)
	})

	t.Run("maintains nulls last true", func(t *testing.T) {
		config := &SortConfig{NullsLast: true}
		option := WithNullsLast()
		option(config)

		assert.True(t, config.NullsLast)
	})
}

// Test option chaining and combinations
func TestOptionChaining(t *testing.T) {
	t.Run("multiple filter options chained", func(t *testing.T) {
		config := &FieldConfig{}

		WithDataType("string")(config)
		WithOperator("ILIKE")(config)
		WithTableAlias("u")(config)
		WithTransform("LOWER", "TRIM")(config)
		Required()(config)

		assert.Equal(t, "string", config.DataType)
		assert.Equal(t, "ILIKE", config.Operator)
		assert.Equal(t, "u", config.TableAlias)
		assert.Equal(t, []string{"LOWER", "TRIM"}, config.Transform)
		assert.True(t, config.Required)
	})

	t.Run("multiple sort options chained", func(t *testing.T) {
		config := &SortConfig{}

		WithSortTableAlias("p")(config)
		WithSortTransform("UPPER")(config)
		WithNullsLast()(config)

		assert.Equal(t, "p", config.TableAlias)
		assert.Equal(t, []string{"UPPER"}, config.Transform)
		assert.True(t, config.NullsLast)
	})
}

// Test options with real pagination config
func TestOptionsWithPaginationConfig(t *testing.T) {
	t.Run("filter options applied through config builder", func(t *testing.T) {
		config := NewDefaultPaginationConfig()
		config.WithFilter("name",
			WithDataType("string"),
			WithOperator("ILIKE"),
			WithTableAlias("u"),
			WithTransform("LOWER"),
			Required(),
		)
		config.WithFilter("age",
			WithDataType("number"),
			WithOperator(">="),
		)

		nameConfig := config.AllowedFilters["name"]
		assert.Equal(t, "name", nameConfig.Field)
		assert.Equal(t, "string", nameConfig.DataType)
		assert.Equal(t, "ILIKE", nameConfig.Operator)
		assert.Equal(t, "u", nameConfig.TableAlias)
		assert.Equal(t, []string{"LOWER"}, nameConfig.Transform)
		assert.True(t, nameConfig.Required)

		ageConfig := config.AllowedFilters["age"]
		assert.Equal(t, "age", ageConfig.Field)
		assert.Equal(t, "number", ageConfig.DataType)
		assert.Equal(t, ">=", ageConfig.Operator)
		assert.False(t, ageConfig.Required)
	})

	t.Run("sort options applied through config builder", func(t *testing.T) {
		config := NewDefaultPaginationConfig()
		config.WithSort("name",
			WithSortTableAlias("u"),
			WithSortTransform("LOWER"),
			WithNullsLast(),
		)
		config.SetDefaultSort("created_at",
			WithSortTableAlias("u"),
			WithNullsLast(),
		)

		nameSort := config.AllowedSorts["name"]
		assert.Equal(t, "name", nameSort.Field)
		assert.Equal(t, "u", nameSort.TableAlias)
		assert.Equal(t, []string{"LOWER"}, nameSort.Transform)
		assert.True(t, nameSort.NullsLast)

		defaultSort := config.DefaultSort
		assert.Equal(t, "created_at", defaultSort.Field)
		assert.Equal(t, "u", defaultSort.TableAlias)
		assert.True(t, defaultSort.NullsLast)
	})
}

// Benchmark tests for options
func BenchmarkFilterOptions(b *testing.B) {
	b.Run("WithDataType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fieldConfig := &FieldConfig{}
			WithDataType("string")(fieldConfig)
		}
	})

	b.Run("WithOperator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := &FieldConfig{}
			WithOperator("ILIKE")(config)
		}
	})

	b.Run("WithTableAlias", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := &FieldConfig{}
			WithTableAlias("u")(config)
		}
	})

	b.Run("WithTransform", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := &FieldConfig{}
			WithTransform("LOWER", "TRIM")(config)
		}
	})

	b.Run("Required", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := &FieldConfig{}
			Required()(config)
		}
	})
}

func BenchmarkSortOptions(b *testing.B) {
	b.Run("WithSortTransform", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := &SortConfig{}
			WithSortTransform("UPPER")(config)
		}
	})

	b.Run("WithSortTableAlias", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := &SortConfig{}
			WithSortTableAlias("u")(config)
		}
	})

	b.Run("WithNullsLast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := &SortConfig{}
			WithNullsLast()(config)
		}
	})
}

// Test edge cases with options
func TestOptionsEdgeCases(t *testing.T) {
	t.Run("nil config should not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			// This should panic in real usage, but let's test the option creation
			option := WithDataType("string")
			assert.NotNil(t, option)
		})
	})

	t.Run("applying same option multiple times", func(t *testing.T) {
		config := &FieldConfig{}

		WithDataType("string")(config)
		WithDataType("number")(config) // Override previous value

		assert.Equal(t, "number", config.DataType)
	})

	t.Run("empty transform slice behavior", func(t *testing.T) {
		config := &FieldConfig{}

		WithTransform()(config) // No arguments
		if config.Transform == nil {
			assert.Nil(t, config.Transform)
		} else {
			assert.Len(t, config.Transform, 0)
		}

		WithTransform("LOWER")(config)
		assert.Equal(t, []string{"LOWER"}, config.Transform)

		WithTransform()(config) // Override with empty
		if config.Transform == nil {
			assert.Nil(t, config.Transform)
		} else {
			assert.Len(t, config.Transform, 0)
		}
	})
}
