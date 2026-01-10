package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductAttribute struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Label string      `json:"label,omitempty"`
}

type ProductAttributes []ProductAttribute

func (pa ProductAttributes) Value() (driver.Value, error) {
	return json.Marshal(pa)
}

func (pa *ProductAttributes) Scan(value interface{}) error {
	if value == nil {
		*pa = []ProductAttribute{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, pa)
}

// JSONMap is a custom type for handling JSON maps in database
type JSONMap map[string]string

func (jm JSONMap) Value() (driver.Value, error) {
	if jm == nil {
		return nil, nil
	}
	return json.Marshal(jm)
}

func (jm *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*jm = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, jm)
}

// StringSlice is a custom type for handling string slices in database
type StringSlice []string

func (ss StringSlice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return json.Marshal(ss)
}

func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*ss = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, ss)
}

type Product struct {
	ID              uuid.UUID         `db:"id" json:"id"`
	SKU             string            `db:"sku" json:"sku"`
	Name            string            `db:"name" json:"name"`
	Description     *string           `db:"description" json:"description,omitempty"`
	CategoryID      uuid.UUID         `db:"category_id" json:"category_id"`
	CategoryName    string            `db:"category_name" json:"category_name,omitempty"`
	Category        *ProductCategory  `db:"-" json:"category,omitempty"`
	UnitPrice       decimal.Decimal   `db:"unit_price" json:"unit_price"`
	CostPrice       decimal.Decimal   `db:"cost_price" json:"cost_price"`
	CurrentStock    int               `db:"current_stock" json:"current_stock"`
	MinStock        int               `db:"min_stock" json:"min_stock"`
	MaxStock        int               `db:"max_stock" json:"max_stock"`
	Weight          *decimal.Decimal  `db:"weight" json:"weight,omitempty"`
	Dimensions      JSONMap           `db:"dimensions" json:"dimensions,omitempty"`
	IsActive        bool              `db:"is_active" json:"is_active"`
	Attributes      ProductAttributes `db:"attributes" json:"attributes,omitempty"`
	ImageURLs       StringSlice       `db:"image_urls" json:"image_urls,omitempty"`
	Barcode         *string           `db:"barcode" json:"barcode,omitempty"`
	LongDescription *string           `db:"long_description" json:"long_description,omitempty"`
	CreatedBy       uuid.UUID         `db:"created_by" json:"-"`
	UpdatedBy       uuid.UUID         `db:"updated_by" json:"-"`
	CreatedAt       time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at" json:"updated_at"`
}

type ProductCategory struct {
	ID               uuid.UUID        `db:"id" json:"id"`
	Code             string           `db:"code" json:"code"`
	Name             string           `db:"name" json:"name"`
	Description      *string          `db:"description" json:"description,omitempty"`
	ParentCategoryID *uuid.UUID       `db:"parent_category_id" json:"parent_category_id,omitempty"`
	ParentCategory   *ProductCategory `db:"-" json:"parent_category,omitempty"`
	IsActive         bool             `db:"is_active" json:"is_active"`
	CreatedBy        uuid.UUID        `db:"created_by" json:"-"`
	UpdatedBy        uuid.UUID        `db:"updated_by" json:"-"`
	CreatedAt        time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time        `db:"updated_at" json:"updated_at"`
}

type ProductWithStock struct {
	Product
	WarehouseStocks []InventoryStock `json:"warehouse_stocks,omitempty"`
	TotalAvailable  int              `json:"total_available"`
	TotalReserved   int              `json:"total_reserved"`
}

type ProductResponse struct {
	Product
	CategoryName *string `json:"category_name,omitempty"`
	StockInfo    struct {
		TotalAvailable int `json:"total_available"`
		TotalReserved  int `json:"total_reserved"`
		TotalOnHand    int `json:"total_on_hand"`
	} `json:"stock_info"`
}

// ProductFilter represents filter options for product queries
type ProductFilter struct {
	CategoryID   *uuid.UUID
	SKU          *string
	Name         *string
	Barcode      *string
	IsActive     *bool
	LowStockOnly *bool
	Page         int
	Limit        int
	Offset       int
	SortBy       string
	SortOrder    string
}
