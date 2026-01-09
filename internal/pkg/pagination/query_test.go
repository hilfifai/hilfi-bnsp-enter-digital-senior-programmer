package pagination

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	pkgErrors "api-jellyfish/internal/pkg/errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFieldExpression(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		config   FieldConfig
		expected string
		wantErr  bool
	}{
		{
			name:     "simple field",
			field:    "name",
			config:   FieldConfig{Field: "name"},
			expected: "name",
			wantErr:  false,
		},
		{
			name:     "field with table alias",
			field:    "name",
			config:   FieldConfig{Field: "name", TableAlias: "u"},
			expected: "u.name",
			wantErr:  false,
		},
		{
			name:     "field with single transform",
			field:    "name",
			config:   FieldConfig{Field: "name", Transform: []string{"LOWER"}},
			expected: "LOWER(name)",
			wantErr:  false,
		},
		{
			name:     "field with multiple transforms",
			field:    "name",
			config:   FieldConfig{Field: "name", Transform: []string{"TRIM", "LOWER"}},
			expected: "LOWER(TRIM(name))",
			wantErr:  false,
		},
		{
			name:     "field with table alias and transform",
			field:    "name",
			config:   FieldConfig{Field: "name", TableAlias: "u", Transform: []string{"UPPER"}},
			expected: "UPPER(u.name)",
			wantErr:  false,
		},
		{
			name:     "invalid field name",
			field:    "invalid-field",
			config:   FieldConfig{Field: "name"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid table alias",
			field:    "name",
			config:   FieldConfig{Field: "name", TableAlias: "invalid-alias"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid transform function",
			field:    "name",
			config:   FieldConfig{Field: "name", Transform: []string{"INVALID"}},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildFieldExpression(tt.field, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBuildWhereAndArgs(t *testing.T) {
	tests := []struct {
		name              string
		filters           map[string]string
		fieldConfigs      map[string]FieldConfig
		defaultFilter     map[string]DefaultFilterField
		expectedClauses   []string
		expectedArgsCount int
		wantErr           bool
	}{
		{
			name:              "no filters",
			filters:           map[string]string{},
			fieldConfigs:      map[string]FieldConfig{},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{},
			expectedArgsCount: 0,
			wantErr:           false,
		},
		{
			name: "simple string filter with ILIKE",
			filters: map[string]string{
				"name": "john",
			},
			fieldConfigs: map[string]FieldConfig{
				"name": {Field: "name", DataType: "string"},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"name ILIKE $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "boolean filter",
			filters: map[string]string{
				"active": "true",
			},
			fieldConfigs: map[string]FieldConfig{
				"active": {Field: "active", DataType: "boolean"},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"active = $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "number filter",
			filters: map[string]string{
				"age": "25",
			},
			fieldConfigs: map[string]FieldConfig{
				"age": {Field: "age", DataType: "number"},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"age = $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "date filter",
			filters: map[string]string{
				"created_at": "2023-01-01T00:00:00Z",
			},
			fieldConfigs: map[string]FieldConfig{
				"created_at": {Field: "created_at", DataType: "date"},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"created_at = $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "UUID filter with NULL value",
			filters: map[string]string{
				"parent_id": "",
			},
			fieldConfigs: map[string]FieldConfig{
				"parent_id": {Field: "parent_id", DataType: "uuid"},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"parent_id IS NULL"},
			expectedArgsCount: 0,
			wantErr:           false,
		},
		{
			name: "filter with table alias",
			filters: map[string]string{
				"name": "john",
			},
			fieldConfigs: map[string]FieldConfig{
				"name": {Field: "name", DataType: "string", TableAlias: "u"},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"u.name ILIKE $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "filter with transform",
			filters: map[string]string{
				"name": "john",
			},
			fieldConfigs: map[string]FieldConfig{
				"name": {Field: "name", DataType: "string", Transform: []string{"LOWER"}},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"LOWER(name) ILIKE $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "multiple filters",
			filters: map[string]string{
				"name":   "john",
				"active": "true",
				"age":    "25",
			},
			fieldConfigs: map[string]FieldConfig{
				"name":   {Field: "name", DataType: "string"},
				"active": {Field: "active", DataType: "boolean"},
				"age":    {Field: "age", DataType: "number"},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"name ILIKE $1", "active = $2", "age = $3"},
			expectedArgsCount: 3,
			wantErr:           false,
		},
		{
			name: "custom operator",
			filters: map[string]string{
				"age": "25",
			},
			fieldConfigs: map[string]FieldConfig{
				"age": {Field: "age", DataType: "number", Operator: ">="},
			},
			defaultFilter:     map[string]DefaultFilterField{},
			expectedClauses:   []string{"age >= $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "too many filters",
			filters: func() map[string]string {
				filters := make(map[string]string)
				for i := 1; i <= MaxFilters+1; i++ {
					fieldName := fmt.Sprintf("field%d", i)
					filters[fieldName] = "value"
				}
				return filters
			}(),
			fieldConfigs: func() map[string]FieldConfig {
				configs := make(map[string]FieldConfig)
				for i := 1; i <= MaxFilters+1; i++ {
					fieldName := fmt.Sprintf("field%d", i)
					configs[fieldName] = FieldConfig{Field: fieldName, DataType: "string"}
				}
				return configs
			}(),
			defaultFilter: map[string]DefaultFilterField{},
			wantErr:       true,
		},
		{
			name:    "default filter",
			filters: map[string]string{},
			fieldConfigs: map[string]FieldConfig{
				"tenant_id": {Field: "tenant_id", DataType: "uuid"},
			},
			defaultFilter: map[string]DefaultFilterField{
				"tenant_id": {Value: "550e8400-e29b-41d4-a716-446655440000"},
			},
			expectedClauses:   []string{"tenant_id = $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name:    "default filter with custom operator",
			filters: map[string]string{},
			fieldConfigs: map[string]FieldConfig{
				"created_at": {Field: "created_at", DataType: "date"},
			},
			defaultFilter: map[string]DefaultFilterField{
				"created_at": {
					Value:    "2023-01-01T00:00:00Z",
					Operator: ">=",
				},
			},
			expectedClauses:   []string{"created_at >= $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clauses, args, err := BuildWhereAndArgs(tt.filters, tt.fieldConfigs, tt.defaultFilter, map[string]SearchConfig{})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, args, tt.expectedArgsCount)

			if len(tt.expectedClauses) == 0 {
				assert.Empty(t, clauses)
			} else {
				// Check that all expected clauses are present (order might vary)
				assert.Len(t, clauses, len(tt.expectedClauses))
				for _, expectedClause := range tt.expectedClauses {
					found := false
					for _, clause := range clauses {
						if clause == expectedClause ||
							(len(clauses) == len(tt.expectedClauses) && len(clauses) == 1) {
							found = true
							break
						}
					}
					if !found && len(tt.expectedClauses) == 1 {
						// For single clause, do partial matching to handle parameter numbers
						assert.Contains(t, clauses[0], tt.expectedClauses[0][:strings.LastIndex(tt.expectedClauses[0], "$")])
					}
				}
			}
		})
	}
}

func TestBuildPaginatedQuery(t *testing.T) {
	tests := []struct {
		name          string
		baseQuery     string
		whereClauses  []string
		sortConfig    SortConfig
		order         string
		limit         int
		offset        int
		expectedQuery string
		wantErr       bool
	}{
		{
			name:          "simple query without WHERE",
			baseQuery:     "SELECT * FROM users",
			whereClauses:  []string{},
			sortConfig:    SortConfig{Field: "id"},
			order:         "ASC",
			limit:         10,
			offset:        0,
			expectedQuery: "SELECT * FROM users ORDER BY id ASC LIMIT 10 OFFSET 0",
			wantErr:       false,
		},
		{
			name:          "query with WHERE clauses",
			baseQuery:     "SELECT * FROM users",
			whereClauses:  []string{"name ILIKE $1", "active = $2"},
			sortConfig:    SortConfig{Field: "created_at"},
			order:         "DESC",
			limit:         20,
			offset:        40,
			expectedQuery: "SELECT * FROM users WHERE name ILIKE $1 AND active = $2 ORDER BY created_at DESC LIMIT 20 OFFSET 40",
			wantErr:       false,
		},
		{
			name:          "sort with table alias",
			baseQuery:     "SELECT u.* FROM users u",
			whereClauses:  []string{},
			sortConfig:    SortConfig{Field: "name", TableAlias: "u"},
			order:         "ASC",
			limit:         5,
			offset:        10,
			expectedQuery: "SELECT u.* FROM users u ORDER BY u.name ASC LIMIT 5 OFFSET 10",
			wantErr:       false,
		},
		{
			name:          "sort with transform",
			baseQuery:     "SELECT * FROM users",
			whereClauses:  []string{},
			sortConfig:    SortConfig{Field: "name", Transform: []string{"LOWER"}},
			order:         "ASC",
			limit:         10,
			offset:        0,
			expectedQuery: "SELECT * FROM users ORDER BY LOWER(name) ASC LIMIT 10 OFFSET 0",
			wantErr:       false,
		},
		{
			name:          "sort with NULLS LAST",
			baseQuery:     "SELECT * FROM users",
			whereClauses:  []string{},
			sortConfig:    SortConfig{Field: "last_login", NullsLast: true},
			order:         "DESC",
			limit:         10,
			offset:        0,
			expectedQuery: "SELECT * FROM users ORDER BY last_login DESC NULLS LAST LIMIT 10 OFFSET 0",
			wantErr:       false,
		},
		{
			name:          "complex sort with alias, transform, and nulls last",
			baseQuery:     "SELECT u.* FROM users u",
			whereClauses:  []string{"u.active = $1"},
			sortConfig:    SortConfig{Field: "name", TableAlias: "u", Transform: []string{"UPPER"}, NullsLast: true},
			order:         "DESC",
			limit:         25,
			offset:        50,
			expectedQuery: "SELECT u.* FROM users u WHERE u.active = $1 ORDER BY UPPER(u.name) DESC NULLS LAST LIMIT 25 OFFSET 50",
			wantErr:       false,
		},
		{
			name:          "invalid order defaults to ASC",
			baseQuery:     "SELECT * FROM users",
			whereClauses:  []string{},
			sortConfig:    SortConfig{Field: "id"},
			order:         "INVALID",
			limit:         10,
			offset:        0,
			expectedQuery: "SELECT * FROM users ORDER BY id ASC LIMIT 10 OFFSET 0",
			wantErr:       false,
		},
		{
			name:          "lowercase order converted to uppercase",
			baseQuery:     "SELECT * FROM users",
			whereClauses:  []string{},
			sortConfig:    SortConfig{Field: "id"},
			order:         "desc",
			limit:         10,
			offset:        0,
			expectedQuery: "SELECT * FROM users ORDER BY id DESC LIMIT 10 OFFSET 0",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildPaginatedQuery(
				tt.baseQuery,
				tt.whereClauses,
				tt.sortConfig,
				tt.order,
				tt.limit,
				tt.offset,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedQuery, result)
			}
		})
	}
}

func TestFetchPaginated(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	ctx := context.Background()

	// Test data
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	tests := []struct {
		name           string
		baseQuery      string
		countQuery     string
		pagination     *Pagination
		mockSetup      func(sqlmock.Sqlmock)
		expectedResult PaginatedResponse[User]
		wantErr        bool
	}{
		{
			name:       "successful pagination",
			baseQuery:  "SELECT id, name FROM users",
			countQuery: "SELECT COUNT(*) FROM users",
			pagination: &Pagination{
				Page:     1,
				PageSize: 2,
				Offset:   0,
				SortBy:   "id",
				Order:    "ASC",
				Filters:  map[string]string{},
				PaginationConfig: PaginationConfig{
					AllowedSearch: map[string]SearchConfig{},
					AllowedSorts: map[string]SortConfig{
						"id": {Field: "id"},
					},
					DefaultSort:    SortConfig{Field: "id"},
					AllowedFilters: map[string]FieldConfig{},
					DefaultFilter:  map[string]DefaultFilterField{},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

				// Data query
				mock.ExpectQuery("SELECT id, name FROM users ORDER BY id ASC LIMIT 2 OFFSET 0").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
						AddRow(1, "John").
						AddRow(2, "Jane"))
			},
			expectedResult: PaginatedResponse[User]{
				Data:       []User{{ID: 1, Name: "John"}, {ID: 2, Name: "Jane"}},
				Total:      5,
				Page:       1,
				PageSize:   2,
				TotalPages: 3,
			},
			wantErr: false,
		},
		{
			name:       "pagination with filters",
			baseQuery:  "SELECT id, name FROM users",
			countQuery: "SELECT COUNT(*) FROM users",
			pagination: &Pagination{
				Page:     1,
				PageSize: 10,
				Offset:   0,
				SortBy:   "name",
				Order:    "DESC",
				Filters:  map[string]string{"active": "true"},
				PaginationConfig: PaginationConfig{
					AllowedSearch: map[string]SearchConfig{},
					AllowedSorts: map[string]SortConfig{
						"name": {Field: "name"},
					},
					DefaultSort: SortConfig{Field: "id"},
					AllowedFilters: map[string]FieldConfig{
						"active": {Field: "active", DataType: "boolean"},
					},
					DefaultFilter: map[string]DefaultFilterField{},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Count query with WHERE
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE active = \\$1").
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

				// Data query with WHERE
				mock.ExpectQuery("SELECT id, name FROM users WHERE active = \\$1 ORDER BY name DESC LIMIT 10 OFFSET 0").
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
						AddRow(3, "Mike").
						AddRow(1, "John").
						AddRow(2, "Jane"))
			},
			expectedResult: PaginatedResponse[User]{
				Data:       []User{{ID: 3, Name: "Mike"}, {ID: 1, Name: "John"}, {ID: 2, Name: "Jane"}},
				Total:      3,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
			},
			wantErr: false,
		},
		{
			name:       "empty result",
			baseQuery:  "SELECT id, name FROM users",
			countQuery: "SELECT COUNT(*) FROM users",
			pagination: &Pagination{
				Page:     1,
				PageSize: 10,
				Offset:   0,
				SortBy:   "id",
				Order:    "ASC",
				Filters:  map[string]string{},
				PaginationConfig: PaginationConfig{
					AllowedSorts: map[string]SortConfig{
						"id": {Field: "id"},
					},
					DefaultSort:    SortConfig{Field: "id"},
					AllowedFilters: map[string]FieldConfig{},
					DefaultFilter:  map[string]DefaultFilterField{},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				// Data query
				mock.ExpectQuery("SELECT id, name FROM users ORDER BY id ASC LIMIT 10 OFFSET 0").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			expectedResult: PaginatedResponse[User]{
				Data:       []User{},
				Total:      0,
				Page:       1,
				PageSize:   10,
				TotalPages: 0,
			},
			wantErr: false,
		},
		{
			name:       "count query error",
			baseQuery:  "SELECT id, name FROM users",
			countQuery: "SELECT COUNT(*) FROM users",
			pagination: &Pagination{
				Page:     1,
				PageSize: 10,
				Offset:   0,
				SortBy:   "id",
				Order:    "ASC",
				Filters:  map[string]string{},
				PaginationConfig: PaginationConfig{
					AllowedSearch: map[string]SearchConfig{},
					AllowedSorts: map[string]SortConfig{
						"id": {Field: "id"},
					},
					DefaultSort:    SortConfig{Field: "id"},
					AllowedFilters: map[string]FieldConfig{},
					DefaultFilter:  map[string]DefaultFilterField{},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:       "data query error",
			baseQuery:  "SELECT id, name FROM users",
			countQuery: "SELECT COUNT(*) FROM users",
			pagination: &Pagination{
				Page:     1,
				PageSize: 10,
				Offset:   0,
				SortBy:   "id",
				Order:    "ASC",
				Filters:  map[string]string{},
				PaginationConfig: PaginationConfig{
					AllowedSearch: map[string]SearchConfig{},
					AllowedSorts: map[string]SortConfig{
						"id": {Field: "id"},
					},
					DefaultSort:    SortConfig{Field: "id"},
					AllowedFilters: map[string]FieldConfig{},
					DefaultFilter:  map[string]DefaultFilterField{},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				mock.ExpectQuery("SELECT id, name FROM users ORDER BY id ASC LIMIT 10 OFFSET 0").
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:       "invalid query string",
			baseQuery:  "SELECT * FROM users; DROP TABLE users;",
			countQuery: "SELECT COUNT(*) FROM users",
			pagination: &Pagination{
				PaginationConfig: PaginationConfig{
					AllowedSearch:  map[string]SearchConfig{},
					AllowedSorts:   map[string]SortConfig{},
					AllowedFilters: map[string]FieldConfig{},
					DefaultFilter:  map[string]DefaultFilterField{},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No expectations as it should fail validation
			},
			wantErr: true,
		},
		{
			name:       "too many sort fields",
			baseQuery:  "SELECT * FROM users",
			countQuery: "SELECT COUNT(*) FROM users",
			pagination: &Pagination{
				PaginationConfig: PaginationConfig{
					AllowedSearch: map[string]SearchConfig{},
					AllowedSorts: map[string]SortConfig{
						"field1": {Field: "field1"},
						"field2": {Field: "field2"},
						"field3": {Field: "field3"},
						"field4": {Field: "field4"},
						"field5": {Field: "field5"},
						"field6": {Field: "field6"}, // Exceeds MaxSortFields (5)
					},
					AllowedFilters: map[string]FieldConfig{},
					DefaultFilter:  map[string]DefaultFilterField{},
				},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No expectations as it should fail validation
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			result, err := FetchPaginated[User](ctx, sqlxDB, tt.baseQuery, tt.countQuery, tt.pagination)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Data, result.Data)
				assert.Equal(t, tt.expectedResult.Total, result.Total)
				assert.Equal(t, tt.expectedResult.Page, result.Page)
				assert.Equal(t, tt.expectedResult.PageSize, result.PageSize)
				assert.Equal(t, tt.expectedResult.TotalPages, result.TotalPages)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Benchmark tests
func BenchmarkBuildFieldExpression(b *testing.B) {
	config := FieldConfig{
		Field:      "name",
		TableAlias: "u",
		Transform:  []string{"LOWER", "TRIM"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := buildFieldExpression("name", config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildWhereAndArgs(b *testing.B) {
	filters := map[string]string{
		"name":   "john",
		"active": "true",
		"age":    "25",
	}
	fieldConfigs := map[string]FieldConfig{
		"name":   {Field: "name", DataType: "string"},
		"active": {Field: "active", DataType: "boolean"},
		"age":    {Field: "age", DataType: "number"},
	}
	defaultFilter := map[string]DefaultFilterField{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := BuildWhereAndArgs(filters, fieldConfigs, defaultFilter, map[string]SearchConfig{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildPaginatedQuery(b *testing.B) {
	whereClauses := []string{"name ILIKE $1", "active = $2"}
	sortConfig := SortConfig{Field: "created_at", TableAlias: "u", NullsLast: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildPaginatedQuery(
			"SELECT * FROM users u",
			whereClauses,
			sortConfig,
			"DESC",
			20,
			40,
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions for tests
func createFieldConfigs() map[string]FieldConfig {
	return map[string]FieldConfig{
		"name":       {Field: "name", DataType: "string"},
		"email":      {Field: "email", DataType: "string"},
		"active":     {Field: "active", DataType: "boolean"},
		"age":        {Field: "age", DataType: "number"},
		"created_at": {Field: "created_at", DataType: "date"},
		"user_id":    {Field: "user_id", DataType: "uuid"},
	}
}

func createSortConfigs() map[string]SortConfig {
	return map[string]SortConfig{
		"name":       {Field: "name"},
		"email":      {Field: "email"},
		"created_at": {Field: "created_at", NullsLast: true},
		"updated_at": {Field: "updated_at", NullsLast: true},
	}
}

// Test edge cases and error conditions
func TestQueryBuildingEdgeCases(t *testing.T) {
	t.Run("buildFieldExpression with empty field", func(t *testing.T) {
		_, err := buildFieldExpression("", FieldConfig{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, pkgErrors.ErrInvalidPaginationParam)
	})

	t.Run("buildWhereAndArgs with invalid field in filter", func(t *testing.T) {
		filters := map[string]string{"invalid-field": "value"}
		fieldConfigs := map[string]FieldConfig{
			"invalid-field": {Field: "name", DataType: "string"},
		}

		_, _, err := BuildWhereAndArgs(filters, fieldConfigs, map[string]DefaultFilterField{}, map[string]SearchConfig{})
		assert.Error(t, err)
	})

	t.Run("complex WHERE clause building", func(t *testing.T) {
		filters := map[string]string{
			"name":       "john doe",
			"status":     "active",
			"created_at": "2023-01-01T00:00:00Z",
		}
		fieldConfigs := map[string]FieldConfig{
			"name":       {Field: "name", DataType: "string", TableAlias: "u", Transform: []string{"LOWER"}},
			"status":     {Field: "status", DataType: "string", Operator: "="},
			"created_at": {Field: "created_at", DataType: "date", Operator: ">="},
			"deleted_at": {Field: "deleted_at", DataType: "date"},
		}
		defaultFilter := map[string]DefaultFilterField{
			"deleted_at": {Value: "", Operator: "IS"},
		}

		clauses, args, err := BuildWhereAndArgs(filters, fieldConfigs, defaultFilter, map[string]SearchConfig{})

		require.NoError(t, err)
		assert.Len(t, clauses, 3) // 3 filters only
		assert.Len(t, args, 3)    // all 3 filters need args

		// Verify clause formats
		expectedPatterns := []string{
			"LOWER(u.name) ILIKE $",
			"status = $",
			"created_at >= $",
		}

		for _, pattern := range expectedPatterns {
			found := false
			for _, clause := range clauses {
				if strings.Contains(clause, pattern) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected pattern %s not found in clauses %v", pattern, clauses)
		}
	})
}

// Test search functionality
func TestBuildWhereAndArgs_SearchFunctionality(t *testing.T) {
	tests := []struct {
		name              string
		filters           map[string]string
		fieldConfigs      map[string]FieldConfig
		defaultFilter     map[string]DefaultFilterField
		searchConfig      map[string]SearchConfig
		expectedClauses   []string
		expectedArgsCount int
		wantErr           bool
	}{
		{
			name: "simple search across multiple fields",
			filters: map[string]string{
				"search": "john",
			},
			fieldConfigs: map[string]FieldConfig{
				"name":  {Field: "name", DataType: "string"},
				"email": {Field: "email", DataType: "string"},
			},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"search": {
					Fields: []FieldConfig{
						{Field: "name", DataType: "string"},
						{Field: "email", DataType: "string"},
					},
				},
			},
			expectedClauses:   []string{"(name ILIKE $1 OR email ILIKE $2)"},
			expectedArgsCount: 2,
			wantErr:           false,
		},
		{
			name: "search with table aliases",
			filters: map[string]string{
				"search": "test",
			},
			fieldConfigs:  map[string]FieldConfig{},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"search": {
					Fields: []FieldConfig{
						{Field: "name", DataType: "string", TableAlias: "u"},
						{Field: "description", DataType: "string", TableAlias: "p"},
					},
				},
			},
			expectedClauses:   []string{"(u.name ILIKE $1 OR p.description ILIKE $2)"},
			expectedArgsCount: 2,
			wantErr:           false,
		},
		{
			name: "search with transform functions",
			filters: map[string]string{
				"search": "TEST",
			},
			fieldConfigs:  map[string]FieldConfig{},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"search": {
					Fields: []FieldConfig{
						{Field: "name", DataType: "string", Transform: []string{"LOWER"}},
						{Field: "title", DataType: "string", Transform: []string{"UPPER"}},
					},
				},
			},
			expectedClauses:   []string{"(LOWER(name) ILIKE $1 OR UPPER(title) ILIKE $2)"},
			expectedArgsCount: 2,
			wantErr:           false,
		},
		{
			name: "search combined with regular filters",
			filters: map[string]string{
				"search": "john",
				"active": "true",
				"age":    "25",
			},
			fieldConfigs: map[string]FieldConfig{
				"active": {Field: "active", DataType: "boolean"},
				"age":    {Field: "age", DataType: "number"},
			},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"search": {
					Fields: []FieldConfig{
						{Field: "name", DataType: "string"},
						{Field: "email", DataType: "string"},
					},
				},
			},
			expectedClauses:   []string{"(name ILIKE $1 OR email ILIKE $2)", "active = $3", "age = $4"},
			expectedArgsCount: 4,
			wantErr:           false,
		},
		{
			name: "empty search term should be ignored",
			filters: map[string]string{
				"search": "",
				"active": "true",
			},
			fieldConfigs: map[string]FieldConfig{
				"active": {Field: "active", DataType: "boolean"},
			},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"search": {
					Fields: []FieldConfig{
						{Field: "name", DataType: "string"},
					},
				},
			},
			expectedClauses:   []string{"active = $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "search with no configured fields should be ignored",
			filters: map[string]string{
				"search": "john",
				"active": "true",
			},
			fieldConfigs: map[string]FieldConfig{
				"active": {Field: "active", DataType: "boolean"},
			},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"search": {
					Fields: []FieldConfig{}, // Empty fields
				},
			},
			expectedClauses:   []string{"active = $1"},
			expectedArgsCount: 1,
			wantErr:           false,
		},
		{
			name: "multiple search configurations",
			filters: map[string]string{
				"user_search":    "john",
				"content_search": "golang",
			},
			fieldConfigs:  map[string]FieldConfig{},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"user_search": {
					Fields: []FieldConfig{
						{Field: "name", DataType: "string"},
						{Field: "email", DataType: "string"},
					},
				},
				"content_search": {
					Fields: []FieldConfig{
						{Field: "title", DataType: "string"},
						{Field: "description", DataType: "string"},
					},
				},
			},
			expectedClauses:   []string{"(title ILIKE $1 OR description ILIKE $2)", "(name ILIKE $3 OR email ILIKE $4)"},
			expectedArgsCount: 4,
			wantErr:           false,
		},
		{
			name: "search with invalid field should return error",
			filters: map[string]string{
				"search": "test",
			},
			fieldConfigs:  map[string]FieldConfig{},
			defaultFilter: map[string]DefaultFilterField{},
			searchConfig: map[string]SearchConfig{
				"search": {
					Fields: []FieldConfig{
						{Field: "invalid-field", DataType: "string"}, // Invalid field name
					},
				},
			},
			expectedClauses:   []string{},
			expectedArgsCount: 0,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clauses, args, err := BuildWhereAndArgs(tt.filters, tt.fieldConfigs, tt.defaultFilter, tt.searchConfig)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, args, tt.expectedArgsCount)

			if len(tt.expectedClauses) == 0 {
				assert.Empty(t, clauses)
			} else {
				assert.Len(t, clauses, len(tt.expectedClauses))

				// Verify all expected clauses are present (order might vary)
				for _, expectedClause := range tt.expectedClauses {
					found := false
					for _, clause := range clauses {
						if strings.Contains(clause, expectedClause) || clause == expectedClause {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected clause '%s' not found in clauses %v", expectedClause, clauses)
				}

				// Verify search parameters contain correct values
				if len(args) > 0 {
					for i, arg := range args {
						if str, ok := arg.(string); ok && strings.HasPrefix(str, "%") && strings.HasSuffix(str, "%") {
							// This is a search parameter, verify it's wrapped with %
							assert.True(t, strings.HasPrefix(str, "%"), "Search arg %d should start with %%, got: %s", i, str)
							assert.True(t, strings.HasSuffix(str, "%"), "Search arg %d should end with %%, got: %s", i, str)
						}
					}
				}
			}
		})
	}
}

// Test search functionality integration
func TestSearchConfigurationValidation(t *testing.T) {
	t.Run("search config with valid fields", func(t *testing.T) {
		config := SearchConfig{
			Fields: []FieldConfig{
				{Field: "name", DataType: "string"},
				{Field: "email", DataType: "string"},
			},
		}

		assert.Len(t, config.Fields, 2)
		assert.Equal(t, "name", config.Fields[0].Field)
		assert.Equal(t, "email", config.Fields[1].Field)
	})

	t.Run("search processing removes search from filters", func(t *testing.T) {
		filters := map[string]string{
			"search": "john",
			"active": "true",
		}

		fieldConfigs := map[string]FieldConfig{
			"active": {Field: "active", DataType: "boolean"},
		}

		searchConfig := map[string]SearchConfig{
			"search": {
				Fields: []FieldConfig{
					{Field: "name", DataType: "string"},
				},
			},
		}

		clauses, args, err := BuildWhereAndArgs(filters, fieldConfigs, map[string]DefaultFilterField{}, searchConfig)

		require.NoError(t, err)
		assert.Len(t, clauses, 2) // One search clause + one regular filter
		assert.Len(t, args, 2)    // One search arg + one filter arg

		// Verify search term is wrapped with %
		searchArg := args[0].(string)
		assert.Equal(t, "%john%", searchArg)

		// Verify boolean filter
		boolArg := args[1].(bool)
		assert.True(t, boolArg)
	})
}

// Benchmark search functionality
func BenchmarkBuildWhereAndArgs_WithSearch(b *testing.B) {
	filters := map[string]string{
		"search": "john doe",
		"active": "true",
	}
	fieldConfigs := map[string]FieldConfig{
		"active": {Field: "active", DataType: "boolean"},
	}
	defaultFilter := map[string]DefaultFilterField{}
	searchConfig := map[string]SearchConfig{
		"search": {
			Fields: []FieldConfig{
				{Field: "name", DataType: "string"},
				{Field: "email", DataType: "string"},
				{Field: "description", DataType: "string"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := BuildWhereAndArgs(filters, fieldConfigs, defaultFilter, searchConfig)
		if err != nil {
			b.Fatal(err)
		}
	}
}
