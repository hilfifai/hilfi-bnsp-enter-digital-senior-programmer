package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-jellyfish/internal/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPaginationFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    map[string]string
		config         PaginationConfig
		expectedResult *Pagination
		expectedError  error
	}{
		{
			name:        "default values when no params provided",
			queryParams: map[string]string{},
			config:      NewDefaultPaginationConfig(),
			expectedResult: &Pagination{
				Page:     DefaultPage,
				PageSize: DefaultPageSize,
				Offset:   0,
				SortBy:   "id",
				Order:    "DESC",
				Filters:  map[string]string{},
			},
		},
		{
			name: "valid pagination parameters",
			queryParams: map[string]string{
				"page":      "2",
				"page_size": "20",
				"sort_by":   "name",
				"order":     "DESC",
			},
			config: func() PaginationConfig {
				c := NewDefaultPaginationConfig()
				c.WithSort("name")
				return c
			}(),
			expectedResult: &Pagination{
				Page:     2,
				PageSize: 20,
				Offset:   20,
				SortBy:   "name",
				Order:    "DESC",
				Filters:  map[string]string{},
			},
		},
		{
			name: "with filters",
			queryParams: map[string]string{
				"name":   "john",
				"active": "true",
				"age":    "25",
			},
			config: func() PaginationConfig {
				c := NewDefaultPaginationConfig()
				c.WithFilter("name", WithDataType("string"))
				c.WithFilter("active", WithDataType("boolean"))
				c.WithFilter("age", WithDataType("number"))
				return c
			}(),
			expectedResult: &Pagination{
				Page:     DefaultPage,
				PageSize: DefaultPageSize,
				Offset:   0,
				SortBy:   "id",
				Order:    "DESC",
				Filters: map[string]string{
					"name":   "john",
					"active": "true",
					"age":    "25",
				},
			},
		},
		{
			name: "required filter with default value",
			queryParams: map[string]string{
				"name": "john",
			},
			config: func() PaginationConfig {
				c := NewDefaultPaginationConfig()
				c.WithFilter("name", WithDataType("string"))
				c.WithFilter("status", WithDataType("string"), Required())
				return c
			}(),
			expectedResult: &Pagination{
				Page:     DefaultPage,
				PageSize: DefaultPageSize,
				Offset:   0,
				SortBy:   "id",
				Order:    "DESC",
				Filters: map[string]string{
					"name":   "john",
					"status": "",
				},
			},
		},
		{
			name: "invalid page parameter",
			queryParams: map[string]string{
				"page": "invalid",
			},
			config:        NewDefaultPaginationConfig(),
			expectedError: errors.ErrInvalidPaginationParam,
		},
		{
			name: "invalid page_size parameter",
			queryParams: map[string]string{
				"page_size": "invalid",
			},
			config:        NewDefaultPaginationConfig(),
			expectedError: errors.ErrInvalidPaginationParam,
		},
		{
			name: "invalid filter value",
			queryParams: map[string]string{
				"age": "not_a_number",
			},
			config: func() PaginationConfig {
				c := NewDefaultPaginationConfig()
				c.WithFilter("age", WithDataType("number"))
				return c
			}(),
			expectedError: errors.ErrInvalidPaginationParam,
		},
		{
			name: "invalid sort field",
			queryParams: map[string]string{
				"sort_by": "invalid_field",
			},
			config:        NewDefaultPaginationConfig(),
			expectedError: errors.ErrInvalidPaginationParam,
		},
		{
			name: "page size exceeds maximum",
			queryParams: map[string]string{
				"page_size": "1000",
			},
			config: NewDefaultPaginationConfig(),
			expectedResult: &Pagination{
				Page:     DefaultPage,
				PageSize: MaxPageSize,
				Offset:   0,
				SortBy:   "id",
				Order:    "DESC",
				Filters:  map[string]string{},
			},
		},
		{
			name: "negative page defaults to 1",
			queryParams: map[string]string{
				"page": "-1",
			},
			config: NewDefaultPaginationConfig(),
			expectedResult: &Pagination{
				Page:     DefaultPage,
				PageSize: DefaultPageSize,
				Offset:   0,
				SortBy:   "id",
				Order:    "DESC",
				Filters:  map[string]string{},
			},
		},
		{
			name: "invalid order defaults to ASC",
			queryParams: map[string]string{
				"order": "INVALID",
			},
			config: NewDefaultPaginationConfig(),
			expectedResult: &Pagination{
				Page:     DefaultPage,
				PageSize: DefaultPageSize,
				Offset:   0,
				SortBy:   "id",
				Order:    "DESC",
				Filters:  map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test HTTP request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute function
			result, err := NewPaginationFromQuery(c, tt.config)

			// Assertions
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Page, result.Page)
				assert.Equal(t, tt.expectedResult.PageSize, result.PageSize)
				assert.Equal(t, tt.expectedResult.Offset, result.Offset)
				assert.Equal(t, tt.expectedResult.SortBy, result.SortBy)
				assert.Equal(t, tt.expectedResult.Order, result.Order)
				assert.Equal(t, tt.expectedResult.Filters, result.Filters)
			}
		})
	}
}

func TestNewDefaultPaginationConfig(t *testing.T) {
	config := NewDefaultPaginationConfig()

	assert.NotNil(t, config.AllowedFilters)
	assert.NotNil(t, config.AllowedSorts)
	assert.NotNil(t, config.DefaultFilter)
	assert.Equal(t, "id", config.DefaultSort.Field)
	assert.Contains(t, config.AllowedSorts, "id")
}

func TestPaginationConfig_WithFilter(t *testing.T) {
	config := NewDefaultPaginationConfig()

	result := config.WithFilter("name", WithDataType("string"), WithOperator("ILIKE"))

	assert.NotNil(t, result) // Should return pointer for chaining
	assert.Contains(t, config.AllowedFilters, "name")

	filterConfig := config.AllowedFilters["name"]
	assert.Equal(t, "name", filterConfig.Field)
	assert.Equal(t, "string", filterConfig.DataType)
	assert.Equal(t, "ILIKE", filterConfig.Operator)
}

func TestPaginationConfig_WithSort(t *testing.T) {
	config := NewDefaultPaginationConfig()

	result := config.WithSort("created_at", WithSortTableAlias("u"), WithNullsLast())

	assert.NotNil(t, result) // Should return pointer for chaining
	assert.Contains(t, config.AllowedSorts, "created_at")

	sortConfig := config.AllowedSorts["created_at"]
	assert.Equal(t, "created_at", sortConfig.Field)
	assert.Equal(t, "u", sortConfig.TableAlias)
	assert.True(t, sortConfig.NullsLast)
}

func TestPaginationConfig_SetDefaultSort(t *testing.T) {
	config := NewDefaultPaginationConfig()

	result := config.SetDefaultSort("name", WithSortTableAlias("u"))

	assert.NotNil(t, result) // Should return pointer for chaining
	assert.Equal(t, "name", config.DefaultSort.Field)
	assert.Equal(t, "u", config.DefaultSort.TableAlias)
}

func TestNewPaginatedResponse(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	total := 23
	page := 2
	pageSize := 10

	response := NewPaginatedResponse(data, total, page, pageSize)

	assert.Equal(t, data, response.Data)
	assert.Equal(t, total, response.Total)
	assert.Equal(t, page, response.Page)
	assert.Equal(t, pageSize, response.PageSize)
	assert.Equal(t, 3, response.TotalPages) // ceil(23/10) = 3
}

func TestNewPaginatedResponse_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		total         int
		pageSize      int
		expectedPages int
	}{
		{"exact division", 20, 10, 2},
		{"zero total", 0, 10, 0},
		{"single item", 1, 10, 1},
		{"less than page size", 5, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewPaginatedResponse([]string{}, tt.total, 1, tt.pageSize)
			assert.Equal(t, tt.expectedPages, response.TotalPages)
		})
	}
}

// Benchmark tests
func BenchmarkNewPaginationFromQuery(b *testing.B) {
	gin.SetMode(gin.TestMode)

	config := NewDefaultPaginationConfig()
	config.WithFilter("name", WithDataType("string"))
	config.WithFilter("active", WithDataType("boolean"))
	config.WithSort("created_at")

	req := httptest.NewRequest(http.MethodGet, "/test?page=2&page_size=20&name=john&active=true&sort_by=created_at&order=DESC", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		_, err := NewPaginationFromQuery(c, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewPaginatedResponse(b *testing.B) {
	data := make([]int, 100)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPaginatedResponse(data, 1000, 5, 20)
	}
}

// Helper function for tests
func createGinContextWithQuery(queryParams map[string]string) *gin.Context {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("createGinContextWithQuery", func(t *testing.T) {
		params := map[string]string{
			"page":   "2",
			"name":   "john",
			"active": "true",
		}

		c := createGinContextWithQuery(params)

		assert.Equal(t, "2", c.Query("page"))
		assert.Equal(t, "john", c.Query("name"))
		assert.Equal(t, "true", c.Query("active"))
	})
}

// Test complex scenarios
func TestComplexPaginationScenarios(t *testing.T) {
	t.Run("complex filter configuration", func(t *testing.T) {
		config := NewDefaultPaginationConfig()
		config.WithFilter("name", WithDataType("string"), WithOperator("ILIKE"), WithTableAlias("u"))
		config.WithFilter("created_at", WithDataType("date"), WithTransform("DATE"))
		config.WithFilter("status_id", WithDataType("uuid"), Required())
		config.WithSort("name", WithSortTableAlias("u"), WithSortTransform("LOWER"))
		config.WithSort("created_at", WithNullsLast())
		config.SetDefaultSort("created_at", WithNullsLast())

		params := map[string]string{
			"page":       "3",
			"page_size":  "15",
			"name":       "john",
			"created_at": "2023-01-01T00:00:00Z",
			"sort_by":    "name",
			"order":      "DESC",
		}

		c := createGinContextWithQuery(params)
		result, err := NewPaginationFromQuery(c, config)

		require.NoError(t, err)
		assert.Equal(t, 3, result.Page)
		assert.Equal(t, 15, result.PageSize)
		assert.Equal(t, 30, result.Offset) // (3-1) * 15
		assert.Equal(t, "name", result.SortBy)
		assert.Equal(t, "DESC", result.Order)
		assert.Equal(t, "john", result.Filters["name"])
		assert.Equal(t, "2023-01-01T00:00:00Z", result.Filters["created_at"])
		assert.Equal(t, "", result.Filters["status_id"]) // Required field with default
	})

	t.Run("edge case with empty filter values", func(t *testing.T) {
		config := NewDefaultPaginationConfig()
		config.WithFilter("name", WithDataType("string"))
		config.WithFilter("description", WithDataType("string"))

		params := map[string]string{
			"name":        "john",
			"description": "", // Empty value should be ignored
		}

		c := createGinContextWithQuery(params)
		result, err := NewPaginationFromQuery(c, config)

		require.NoError(t, err)
		assert.Equal(t, "john", result.Filters["name"])
		assert.NotContains(t, result.Filters, "description")
	})
}

// Error handling tests
func TestPaginationErrorHandling(t *testing.T) {
	t.Run("multiple validation errors", func(t *testing.T) {
		config := NewDefaultPaginationConfig()
		config.WithFilter("age", WithDataType("number"))
		config.WithFilter("active", WithDataType("boolean"))

		params := map[string]string{
			"age":    "not_a_number",
			"active": "maybe", // Invalid boolean
		}

		c := createGinContextWithQuery(params)
		_, err := NewPaginationFromQuery(c, config)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrInvalidPaginationParam)
	})

	t.Run("SQL injection attempt in sort field", func(t *testing.T) {
		config := NewDefaultPaginationConfig()

		params := map[string]string{
			"sort_by": "name; DROP TABLE users;",
		}

		c := createGinContextWithQuery(params)
		_, err := NewPaginationFromQuery(c, config)

		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrInvalidPaginationParam)
	})
}
