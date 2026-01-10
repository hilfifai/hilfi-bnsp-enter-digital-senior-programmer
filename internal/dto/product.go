package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateProductRequest struct {
	SKU             string                `json:"sku" validate:"required,min=3,max=100"`
	Name            string                `json:"name" validate:"required,min=3,max=200"`
	Description     *string               `json:"description,omitempty"`
	CategoryID      uuid.UUID             `json:"category_id" validate:"required"`
	UnitPrice       decimal.Decimal       `json:"unit_price" validate:"required"`
	CostPrice       decimal.Decimal       `json:"cost_price" validate:"required"`
	MinStock        int                   `json:"min_stock" validate:"min=0"`
	MaxStock        int                   `json:"max_stock" validate:"min=0"`
	Weight          *decimal.Decimal      `json:"weight,omitempty"`
	LongDescription *string               `db:"long_description" json:"long_description,omitempty"`
	Dimensions      map[string]string     `json:"dimensions,omitempty"`
	IsActive        bool                  `json:"is_active"`
	Attributes      []ProductAttributeDTO `json:"attributes,omitempty"`
	ImageURLs       []string              `json:"image_urls,omitempty"`
	Barcode         *string               `json:"barcode,omitempty"`
}

type UpdateProductRequest struct {
	Name        string                `json:"name" validate:"required,min=3,max=200"`
	Description *string               `json:"description,omitempty"`
	CategoryID  uuid.UUID             `json:"category_id" validate:"required"`
	UnitPrice   decimal.Decimal       `json:"unit_price" validate:"required"`
	CostPrice   decimal.Decimal       `json:"cost_price" validate:"required"`
	MinStock    int                   `json:"min_stock" validate:"min=0"`
	MaxStock    int                   `json:"max_stock" validate:"min=0"`
	Weight      *decimal.Decimal      `json:"weight,omitempty"`
	Dimensions  map[string]string     `json:"dimensions,omitempty"`
	IsActive    bool                  `json:"is_active"`
	Attributes  []ProductAttributeDTO `json:"attributes,omitempty"`
	ImageURLs   []string              `json:"image_urls,omitempty"`
	Barcode     *string               `json:"barcode,omitempty"`
}

type ProductAttributeDTO struct {
	Key   string      `json:"key" validate:"required"`
	Value interface{} `json:"value" validate:"required"`
	Label string      `json:"label,omitempty"`
}

type ProductResponseDTO struct {
	ID           uuid.UUID             `json:"id"`
	SKU          string                `json:"sku"`
	Name         string                `json:"name"`
	Description  *string               `json:"description,omitempty"`
	CategoryID   uuid.UUID             `json:"category_id"`
	CategoryName *string               `json:"category_name,omitempty"`
	UnitPrice    decimal.Decimal       `json:"unit_price"`
	CostPrice    decimal.Decimal       `json:"cost_price"`
	CurrentStock int                   `json:"current_stock"`
	MinStock     int                   `json:"min_stock"`
	MaxStock     int                   `json:"max_stock"`
	Weight       *decimal.Decimal      `json:"weight,omitempty"`
	Dimensions   map[string]string     `json:"dimensions,omitempty"`
	IsActive     bool                  `json:"is_active"`
	Attributes   []ProductAttributeDTO `json:"attributes,omitempty"`
	ImageURLs    []string              `json:"image_urls,omitempty"`
	Barcode      *string               `json:"barcode,omitempty"`
	StockInfo    ProductStockInfoDTO   `json:"stock_info"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

type ProductStockInfoDTO struct {
	TotalAvailable int  `json:"total_available"`
	TotalReserved  int  `json:"total_reserved"`
	TotalOnHand    int  `json:"total_on_hand"`
	LowStockAlert  bool `json:"low_stock_alert"`
}

type ProductFilterDTO struct {
	CategoryID   *uuid.UUID `form:"category_id,omitempty"`
	SKU          *string    `form:"sku,omitempty"`
	Name         *string    `form:"name,omitempty"`
	Barcode      *string    `form:"barcode,omitempty"`
	IsActive     *bool      `form:"is_active,omitempty"`
	LowStockOnly *bool      `form:"low_stock_only,omitempty"`
	Page         int        `form:"page" validate:"min=1"`
	Limit        int        `form:"limit" validate:"min=1,max=100"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order" validate:"oneof=asc desc"`
}

type BulkUpdateStockRequest struct {
	ProductID   uuid.UUID  `json:"product_id" validate:"required"`
	WarehouseID uuid.UUID  `json:"warehouse_id" validate:"required"`
	LocationID  *uuid.UUID `json:"location_id,omitempty"`
	Quantity    int        `json:"quantity" validate:"required,min=0"`
	Type        string     `json:"type" validate:"required,oneof=INCREMENT DECREMENT SET"`
	Reason      string     `json:"reason" validate:"required"`
	Notes       *string    `json:"notes,omitempty"`
}
