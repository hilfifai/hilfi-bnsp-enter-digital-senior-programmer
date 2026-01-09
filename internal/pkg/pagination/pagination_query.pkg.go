package pagination

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"api-jellyfish/internal/pkg/errors"
)

// DBInterface defines the database interface needed for pagination
// This mirrors the main DBInterface to avoid import cycles
type DBInterface interface {
	// Query methods without context (for backward compatibility)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	// Additional methods from the main DBInterface that might be needed
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// buildFieldExpression create ekspresi SQL untuk field dengan transform functions
func buildFieldExpression(field string, config FieldConfig) (string, error) {
	// Validasi nama field
	if !isValidFieldName(field) {
		return "", errors.ErrInvalidPaginationParam
	}

	expression := field
	if config.TableAlias != "" {
		// Validasi table alias
		if !isValidFieldName(config.TableAlias) {
			return "", errors.ErrInvalidPaginationParam
		}
		expression = fmt.Sprintf("%s.%s", config.TableAlias, config.Field)
	}

	// Validasi dan apply transforms
	for _, transform := range config.Transform {
		if !isValidTransformFunction(transform) {
			return "", errors.ErrInvalidPaginationParam
		}
		expression = fmt.Sprintf("%s(%s)", transform, expression)
	}

	return expression, nil
}

// BuildWhereAndArgs membangun WHERE clause dan parameter untuk query
func BuildWhereAndArgs(filters map[string]string, fieldConfigs map[string]FieldConfig, defaultFilter map[string]DefaultFilterField, searchFilter map[string]SearchConfig) ([]string, []interface{}, error) {
	// Batasi jumlah filter untuk keamanan
	if len(filters) > MaxFilters {
		return nil, nil, errors.ErrTooManyFilters
	}

	clauses := []string{}
	args := []interface{}{}
	i := 1

	// Sort search filter keys for consistent ordering
	var searchKeys []string
	for key := range searchFilter {
		searchKeys = append(searchKeys, key)
	}
	sort.Strings(searchKeys)

	for _, field := range searchKeys {
		val := searchFilter[field]
		// Handle search across multiple fields
		if searchTerm, exists := filters[field]; exists && searchTerm != "" && len(val.Fields) > 0 {
			searchClauses := []string{}
			for _, fieldConfig := range val.Fields {
				fullFieldName, err := buildFieldExpression(fieldConfig.Field, fieldConfig)
				if err != nil {
					return nil, nil, err
				}
				searchClauses = append(searchClauses, fmt.Sprintf("%s ILIKE $%d", fullFieldName, i))
				args = append(args, "%"+searchTerm+"%")
				i++
			}
			if len(searchClauses) > 0 {
				clauses = append(clauses, "("+strings.Join(searchClauses, " OR ")+")")
			}
			// Remove search from filters to avoid double processing
			delete(filters, field)
		}
	}

	// Sort filter keys for consistent ordering
	var filterKeys []string
	for key := range filters {
		filterKeys = append(filterKeys, key)
	}
	sort.Strings(filterKeys)

	// Process each filter
	for _, field := range filterKeys {
		val := filters[field]
		if config, exists := fieldConfigs[field]; exists {
			fullFieldName, err := buildFieldExpression(field, config)
			if err != nil {
				return nil, nil, err
			}

			// Tentukan operator default berdasarkan tipe data
			operator := config.Operator
			if operator == "" {
				switch config.DataType {
				case "string":
					operator = "ILIKE"
				default:
					operator = "="
				}
			}

			// Convert dan validation nilai berdasarkan tipe data
			var processedVal interface{}
			var valid bool
			switch config.DataType {
			case "uuid":
				if val == "" {
					operator = "IS"
					processedVal = "NULL"
					valid = true
				} else {
					processedVal = val
					valid = true
				}
			case "boolean":
				if val == "true" || val == "false" {
					processedVal = val == "true"
					valid = true
				}
			case "number":
				if num, err := strconv.ParseFloat(val, 64); err == nil {
					processedVal = num
					valid = true
				}
			case "date":
				if _, err := time.Parse(time.RFC3339, val); err == nil {
					processedVal = val
					valid = true
				}
			case "in_year":
				processedVal = val
				clauses = append(clauses, fmt.Sprintf("EXTRACT(YEAR FROM %s) = %s", fullFieldName, processedVal))
			default: // string
				if operator == "ILIKE" {
					// Escape LIKE special characters dan tambah wildcard
					val = strings.Replace(val, "%", "\\%", -1)
					val = strings.Replace(val, "_", "\\_", -1)
					processedVal = "%" + val + "%"
				} else {
					processedVal = val
				}
				valid = true
			}

			if valid {
				if processedVal == "NULL" {
					clauses = append(clauses, fmt.Sprintf("%s %s NULL", fullFieldName, operator))
				} else {
					clauses = append(clauses, fmt.Sprintf("%s %s $%d", fullFieldName, operator, i))
					args = append(args, processedVal)
					i++
				}

			}
		}
	}

	// Add default filters
	for field, defaultValConfig := range defaultFilter {
		if config, exists := fieldConfigs[field]; exists {
			// Use defaultValConfig's transform if provided, otherwise use fieldConfig's
			transforms := config.Transform
			if len(defaultValConfig.Transform) > 0 {
				transforms = defaultValConfig.Transform
			}

			fieldConfigWithTransforms := FieldConfig{
				Field:      config.Field,
				TableAlias: config.TableAlias,
				DataType:   config.DataType,
				Operator:   config.Operator,
				Transform:  transforms,
				Required:   config.Required,
			}

			fullFieldName, err := buildFieldExpression(field, fieldConfigWithTransforms)
			if err != nil {
				return nil, nil, err
			}

			// Use defaultValConfig's operator if provided, otherwise use fieldConfig's or default
			operator := defaultValConfig.Operator
			if operator == "" {
				operator = config.Operator
				if operator == "" {
					switch config.DataType {
					case "string":
						operator = "ILIKE"
					default:
						operator = "="
					}
				}
			}

			// Convert dan validation nilai berdasarkan tipe data
			var processedVal interface{}
			var valid bool
			val := defaultValConfig.Value
			switch config.DataType {
			case "uuid":
				if val == "" {
					operator = "IS"
					processedVal = "NULL"
					valid = true
				} else {
					processedVal = val
					valid = true
				}
			case "boolean":
				if val == "true" || val == "false" {
					processedVal = val == "true"
					valid = true
				}
			case "number":
				if num, err := strconv.ParseFloat(val, 64); err == nil {
					processedVal = num
					valid = true
				}
			case "date":
				if _, err := time.Parse(time.RFC3339, val); err == nil {
					processedVal = val
					valid = true
				}
			default: // string
				if operator == "ILIKE" {
					val = strings.Replace(val, "%", "\\%", -1)
					val = strings.Replace(val, "_", "\\_", -1)
					processedVal = "%" + val + "%"
				} else {
					processedVal = val
				}
				valid = true
			}

			if valid {
				if processedVal == "NULL" {
					clauses = append(clauses, fmt.Sprintf("%s %s NULL", fullFieldName, operator))
				} else {
					clauses = append(clauses, fmt.Sprintf("%s %s $%d", fullFieldName, operator, i))
					args = append(args, processedVal)
					i++
				}
			}
		}
	}
	return clauses, args, nil
}

func BuildPaginatedQuery(baseQuery string, whereClauses []string, sortConfig SortConfig, order string, limit int, offset int) (string, error) {
	query := baseQuery

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Build sort expression dengan transform functions
	sortField, err := buildFieldExpression(sortConfig.Field, FieldConfig{
		Field:      sortConfig.Field,
		TableAlias: sortConfig.TableAlias,
		Transform:  sortConfig.Transform,
	})
	if err != nil {
		return "", err
	}

	// Sanitize order direction
	order = strings.ToUpper(order)
	if order != "ASC" && order != "DESC" {
		order = "ASC"
	}

	// Tambahkan NULLS LAST jika diperlukan
	nullsClause := ""
	if sortConfig.NullsLast {
		nullsClause = " NULLS LAST"
	}

	query += fmt.Sprintf(" ORDER BY %s %s%s", sortField, order, nullsClause)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	return query, nil
}

// FetchPaginated get data dengan pagination dari database
func FetchPaginated[T any](
	ctx context.Context,
	db DBInterface,
	baseQuery string,
	countQuery string,
	pagination *Pagination,
) (PaginatedResponse[T], error) {
	config := pagination.PaginationConfig
	// Validate query strings
	if !isValidQueryString(baseQuery) || !isValidQueryString(countQuery) {
		return PaginatedResponse[T]{}, errors.ErrInvalidQueryString
	}

	// Validate sort fields
	if len(config.AllowedSorts) > MaxSortFields { // Changed from queryConfig.SortConfigs
		return PaginatedResponse[T]{}, errors.ErrTooManySortFields
	}

	// Build where clauses with field configurations
	whereClauses, args, err := BuildWhereAndArgs(
		pagination.Filters,
		config.AllowedFilters,
		config.DefaultFilter,
		config.AllowedSearch,
	) // Changed from queryConfig.FieldConfigs and queryConfig.DefaultFilter
	if err != nil {
		return PaginatedResponse[T]{}, fmt.Errorf("error building WHERE clauses: %w", err)
	}

	// Get total count
	countSQL := countQuery
	if len(whereClauses) > 0 {
		countSQL += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	if err := db.Get(&total, countSQL, args...); err != nil {
		return PaginatedResponse[T]{}, fmt.Errorf("error getting total count: %w", err)
	}

	// Get sort configuration
	sortConfig, exists := config.AllowedSorts[pagination.SortBy] // Changed from queryConfig.SortConfigs
	if !exists {
		sortConfig = config.DefaultSort // Changed from queryConfig.DefaultSort
	}

	// Build and execute main query
	query, err := BuildPaginatedQuery(
		baseQuery,
		whereClauses,
		sortConfig,
		pagination.Order,
		pagination.PageSize,
		pagination.Offset,
	)
	if err != nil {
		return PaginatedResponse[T]{}, fmt.Errorf("error building paginated query: %w", err)
	}

	var data []T
	if err := db.Select(&data, query, args...); err != nil {
		return PaginatedResponse[T]{}, fmt.Errorf("error fetching data: %w", err)
	}

	// Initialize empty slice if null
	if data == nil {
		data = make([]T, 0)
	}

	return NewPaginatedResponse(data, total, pagination.Page, pagination.PageSize), nil
}
