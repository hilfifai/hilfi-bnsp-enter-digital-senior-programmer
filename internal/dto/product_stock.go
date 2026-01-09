package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateProductStockRequest struct {
	ProductID    uuid.UUID `json:"product_id" validate:"required"`
	WarehouseID  uuid.UUID `json:"warehouse_id" validate:"required"`
	Quantity     int       `json:"quantity" validate:"min=0"`
	ReorderLevel int       `json:"reorder_level" validate:"min=0"`
}

type UpdateProductStockRequest struct {
	Quantity     int `json:"quantity" validate:"min=0"`
	Reserved     int `json:"reserved" validate:"min=0"`
	ReorderLevel int `json:"reorder_level" validate:"min=0"`
}

type ProductStockResponse struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name"`
	WarehouseID   uuid.UUID `json:"warehouse_id"`
	WarehouseName string    `json:"warehouse_name"`
	Quantity      int       `json:"quantity"`
	Reserved      int       `json:"reserved"`
	Available     int       `json:"available"`
	ReorderLevel  int       `json:"reorder_level"`
	LastUpdated   time.Time `json:"last_updated"`
}

type StockMovementResponse struct {
	ID              uuid.UUID       `json:"id"`
	ReferenceNumber string          `json:"reference_number"`
	MovementType    string          `json:"movement_type"`
	ProductID       uuid.UUID       `json:"product_id"`
	ProductName     string          `json:"product_name"`
	WarehouseID     uuid.UUID       `json:"warehouse_id"`
	WarehouseName   string          `json:"warehouse_name"`
	Quantity        int             `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	TotalValue      decimal.Decimal `json:"total_value"`
	Status          string          `json:"status"`
	MovementDate    time.Time       `json:"movement_date"`
	Notes           *string         `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

type ProductStockFilterDTO struct {
	ProductID   *uuid.UUID `form:"product_id,omitempty"`
	WarehouseID *uuid.UUID `form:"warehouse_id,omitempty"`
	LowStock    *bool      `form:"low_stock,omitempty"`
	Page        int        `form:"page" validate:"min=1"`
	Limit       int        `form:"limit" validate:"min=1,max=100"`
}

type StockMovementFilterDTO struct {
	ProductID    *uuid.UUID `form:"product_id,omitempty"`
	WarehouseID  *uuid.UUID `form:"warehouse_id,omitempty"`
	MovementType *string    `form:"movement_type,omitempty"`
	StartDate    *time.Time `form:"start_date,omitempty"`
	EndDate      *time.Time `form:"end_date,omitempty"`
	Page         int        `form:"page" validate:"min=1"`
	Limit        int        `form:"limit" validate:"min=1,max=100"`
}
