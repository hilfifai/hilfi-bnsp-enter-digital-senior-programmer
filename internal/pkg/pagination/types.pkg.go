package pagination

// Constants for pagination
const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
	MaxFilters      = 10  // Maksimum jumlah filter yang diizinkan
	MaxSortFields   = 5   // Maksimum jumlah field pengurutan
	MaxStringLength = 255 // Maksimum panjang string untuk input
)

// FieldConfig menentukan konfigurasi untuk field yang bisa difilter
type FieldConfig struct {
	Field      string
	TableAlias string   // Opsional, untuk query dengan JOIN
	DataType   string   // "string", "boolean", "number", "date", "uuid"
	Operator   string   // Opsional, default: "=" untuk boolean/number/date, "ILIKE" untuk string
	Transform  []string // Opsional, fungsi SQL yang akan diapply ke field
	Required   bool     // Opsional, untuk required fields
}

// SortConfig menentukan konfigurasi untuk field yang bisa diurutkan
type SortConfig struct {
	Field      string
	TableAlias string   // Opsional, untuk query dengan JOIN
	Transform  []string // Opsional, fungsi SQL yang akan diapply ke field
	NullsLast  bool     // Opsional, menambahkan NULLS LAST ke sort clause
}
type SearchConfig struct {
	Fields []FieldConfig // Untuk multi-column search
}

// PaginationConfig menyimpan konfigurasi untuk pagination
type PaginationConfig struct {
	// Map of filterable fields, key is the query parameter name
	// Example for multi-column search:
	// "search": {Fields: []FieldConfig{
	//     {Field: "name", DataType: "string", Operator: "like"},
	//     {Field: "email", DataType: "string", Operator: "like"},
	// }}
	AllowedSearch map[string]SearchConfig
	// Map of filterable fields, key is the query parameter name
	// Example:
	// "status": {Field: "status", DataType: "string", Operator: "="}
	// "created_at": {Field: "created_at", DataType: "date", Operator: ">="}
	// "is_active": {Field: "is_active", DataType: "boolean"}
	// "age": {Field: "age", DataType: "number", Operator: "<"}
	// "username": {Field: "username", DataType: "string", Operator: "like", Transform: []string{"LOWER"}}
	// "email": {Field: "email", DataType: "string", Operator: "like", Transform: []string{"LOWER"}}
	AllowedFilters map[string]FieldConfig
	// Map of sortable fields, key is the query parameter name
	// Example:
	// "id": {Field: "id"}
	// "created_at": {Field: "created_at"}
	// "name": {Field: "name", Transform: []string{"LOWER"}}
	// "email": {Field: "email", Transform: []string{"LOWER"}}
	// "age": {Field: "age"}
	AllowedSorts map[string]SortConfig
	// Default sort if none provided in query
	DefaultSort SortConfig
	// Default filters to always apply, key is the field name in DB
	// Example:
	// "is_active": "true"
	// "status": "active"
	// "created_at": ">=2023-01-01"
	DefaultFilter map[string]DefaultFilterField // Changed from map[string]string
}

// DefaultFilterField provides granular control for default filters
type DefaultFilterField struct {
	Value     string
	Operator  string   // Optional, override default operator from FieldConfig
	Transform []string // Optional, override transform from FieldConfig
}

// Pagination menyimpan informasi paging dari request
type Pagination struct {
	Page             int
	PageSize         int
	Offset           int
	SortBy           string
	Order            string
	Filters          map[string]string
	PaginationConfig PaginationConfig
}

// PaginatedResponse untuk response pagination
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// Options untuk konfigurasi
type FilterOption func(*FieldConfig)
type SortOption func(*SortConfig)
