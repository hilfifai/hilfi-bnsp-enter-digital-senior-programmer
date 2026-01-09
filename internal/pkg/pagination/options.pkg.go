package pagination

// WithDataType menentukan tipe data field
func WithDataType(dataType string) FilterOption {
	return func(c *FieldConfig) {
		c.DataType = dataType
	}
}

// WithOperator menentukan operator SQL untuk filter
func WithOperator(operator string) FilterOption {
	return func(c *FieldConfig) {
		c.Operator = operator
	}
}

// WithTableAlias menentukan alias tabel untuk field
func WithTableAlias(alias string) FilterOption {
	return func(c *FieldConfig) {
		c.TableAlias = alias
	}
}

// WithTransform menentukan fungsi SQL yang akan diapply ke field
func WithTransform(transforms ...string) FilterOption {
	return func(c *FieldConfig) {
		c.Transform = transforms
	}
}

// Required menandai field sebagai required
func Required() FilterOption {
	return func(c *FieldConfig) {
		c.Required = true
	}
}

// WithSortTransform menentukan fungsi SQL yang akan diapply saat sorting
func WithSortTransform(transforms ...string) SortOption {
	return func(c *SortConfig) {
		c.Transform = transforms
	}
}

// WithSortTableAlias menentukan alias tabel untuk sorting
func WithSortTableAlias(alias string) SortOption {
	return func(c *SortConfig) {
		c.TableAlias = alias
	}
}

// WithNullsLast menambahkan NULLS LAST ke sort clause
func WithNullsLast() SortOption {
	return func(c *SortConfig) {
		c.NullsLast = true
	}
}
