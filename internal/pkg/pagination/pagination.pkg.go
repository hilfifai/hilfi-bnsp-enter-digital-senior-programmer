package pagination

import (
	"fmt"
	"strconv"
	"strings"

	"api-jellyfish/internal/pkg/errors"

	"github.com/gin-gonic/gin"
)

func NewPaginationFromQuery(c *gin.Context, config PaginationConfig) (*Pagination, error) {
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	if err != nil {
		return nil, errors.ErrInvalidPaginationParam
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))
	if err != nil {
		return nil, errors.ErrInvalidPaginationParam
	}
	sortBy := c.DefaultQuery("sort_by", "")
	order := strings.ToUpper(c.DefaultQuery("order", "DESC"))

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	filters := make(map[string]string)
	// Ambil dan validation filter dari query parameters
	for field, _ := range config.AllowedSearch {
		if val := c.Query(field); val != "" {
			if IsValidString(val) {
				filters[field] = val
			} else {
				return nil, fmt.Errorf("%w: invalid filter value for field %s", errors.ErrInvalidPaginationParam, field)
			}
		}
	}

	for field, filterConfig := range config.AllowedFilters {
		if val := c.Query(field); val != "" {
			if validateFilterValue(val, filterConfig) {
				filters[field] = val
			} else {
				return nil, fmt.Errorf("%w: invalid filter value for field %s", errors.ErrInvalidPaginationParam, field)
			}
		} else if filterConfig.Required {
			filters[field] = getDefaultValueForType(filterConfig)
		}
	}

	finalSortBy := config.DefaultSort.Field
	if sortBy != "" {
		if _, exists := config.AllowedSorts[sortBy]; exists {
			finalSortBy = sortBy
		} else {
			return nil, fmt.Errorf("%w: invalid sort field %s", errors.ErrInvalidPaginationParam, sortBy)
		}
	}

	return &Pagination{
		Page:             page,
		PageSize:         pageSize,
		Offset:           (page - 1) * pageSize,
		SortBy:           finalSortBy,
		Order:            order,
		Filters:          filters,
		PaginationConfig: config,
	}, nil
}

// NewDefaultPaginationConfig create konfigurasi pagination default
func NewDefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		AllowedSearch:  make(map[string]SearchConfig),
		AllowedFilters: make(map[string]FieldConfig),
		AllowedSorts: map[string]SortConfig{
			"id": {
				Field: "id",
			},
		},
		DefaultSort: SortConfig{
			Field: "id",
		},
		DefaultFilter: make(map[string]DefaultFilterField),
	}
}

// WithFilter menambahkan konfigurasi filter
func (pc *PaginationConfig) WithFilter(field string, opts ...FilterOption) *PaginationConfig {
	config := FieldConfig{Field: field}
	for _, opt := range opts {
		opt(&config)
	}
	pc.AllowedFilters[field] = config
	return pc
}

// WithSort menambahkan konfigurasi pengurutan
func (pc *PaginationConfig) WithSort(field string, opts ...SortOption) *PaginationConfig {
	config := SortConfig{Field: field}
	for _, opt := range opts {
		opt(&config)
	}
	pc.AllowedSorts[field] = config
	return pc
}

// SetDefaultSort mengatur pengurutan default
func (pc *PaginationConfig) SetDefaultSort(field string, opts ...SortOption) *PaginationConfig {
	config := SortConfig{Field: field}
	for _, opt := range opts {
		opt(&config)
	}
	pc.DefaultSort = config
	return pc
}

func NewPaginatedResponse[T any](data []T, total, page, pageSize int) PaginatedResponse[T] {
	totalPages := (total + pageSize - 1) / pageSize
	return PaginatedResponse[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
