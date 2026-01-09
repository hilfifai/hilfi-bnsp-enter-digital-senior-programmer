// internal/dto/inventory.go
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StockMovementRequest struct {
	MovementType    string          `json:"movement_type" validate:"required,oneof=IN OUT TRANSFER ADJUSTMENT OPNAME"`
	ProductID       uuid.UUID       `json:"product_id" validate:"required"`
	FromWarehouseID *uuid.UUID      `json:"from_warehouse_id,omitempty"`
	FromLocationID  *uuid.UUID      `json:"from_location_id,omitempty"`
	ToWarehouseID   *uuid.UUID      `json:"to_warehouse_id,omitempty"`
	ToLocationID    *uuid.UUID      `json:"to_location_id,omitempty"`
	Quantity        int             `json:"quantity" validate:"required,min=1"`
	UnitPrice       decimal.Decimal `json:"unit_price" validate:"required"`
	Notes           *string         `json:"notes,omitempty"`
	MovementDate    time.Time       `json:"movement_date"`
	BatchNumber     *string         `json:"batch_number,omitempty"`
	ExpiryDate      *time.Time      `json:"expiry_date,omitempty"`
}

type StockTransferRequest struct {
	FromWarehouseID uuid.UUID  `json:"from_warehouse_id" validate:"required"`
	FromLocationID  *uuid.UUID `json:"from_location_id,omitempty"`
	ToWarehouseID   uuid.UUID  `json:"to_warehouse_id" validate:"required"`
	ToLocationID    *uuid.UUID `json:"to_location_id,omitempty"`
	ProductID       uuid.UUID  `json:"product_id" validate:"required"`
	Quantity        int        `json:"quantity" validate:"required,min=1"`
	Notes           *string    `json:"notes,omitempty"`
}

type StockAdjustmentRequest struct {
	ProductID   uuid.UUID  `json:"product_id" validate:"required"`
	WarehouseID uuid.UUID  `json:"warehouse_id" validate:"required"`
	LocationID  *uuid.UUID `json:"location_id,omitempty"`
	NewQuantity int        `json:"new_quantity" validate:"required,min=0"`
	Reason      string     `json:"reason" validate:"required"`
	Notes       *string    `json:"notes,omitempty"`
}

type StockOpnameRequest struct {
	WarehouseID uuid.UUID                `json:"warehouse_id" validate:"required"`
	OpnameDate  time.Time                `json:"opname_date"`
	Items       []StockOpnameItemRequest `json:"items" validate:"required,min=1,dive"`
}

type StockOpnameItemRequest struct {
	ProductID     uuid.UUID  `json:"product_id" validate:"required"`
	LocationID    *uuid.UUID `json:"location_id,omitempty"`
	SystemQty     int        `json:"system_qty"`
	PhysicalQty   int        `json:"physical_qty" validate:"required,min=0"`
	Variance      int        `json:"variance"`
	VarianceNotes *string    `json:"variance_notes,omitempty"`
}

type StockReportFilter struct {
	WarehouseID  *uuid.UUID `form:"warehouse_id,omitempty"`
	ProductID    *uuid.UUID `form:"product_id,omitempty"`
	CategoryID   *uuid.UUID `form:"category_id,omitempty"`
	StartDate    *time.Time `form:"start_date,omitempty"`
	EndDate      *time.Time `form:"end_date,omitempty"`
	MovementType *string    `form:"movement_type,omitempty"`
	Page         int        `form:"page" validate:"min=1"`
	Limit        int        `form:"limit" validate:"min=1,max=100"`
	Offset       int        `form:"offset"`
}

type SalesReportFilter struct {
	StartDate  *time.Time `form:"start_date,omitempty"`
	EndDate    *time.Time `form:"end_date,omitempty"`
	CustomerID *uuid.UUID `form:"customer_id,omitempty"`
	ProductID  *uuid.UUID `form:"product_id,omitempty"`
	Period     string     `form:"period"` // daily, weekly, monthly
	Page       int        `form:"page" validate:"min=1"`
	Limit      int        `form:"limit" validate:"min=1,max=100"`
	Offset     int        `form:"offset"`
}

type MovementReportFilter struct {
	WarehouseID  *uuid.UUID `form:"warehouse_id,omitempty"`
	ProductID    *uuid.UUID `form:"product_id,omitempty"`
	LocationID   *uuid.UUID `form:"location_id,omitempty"`
	MovementType *string    `form:"movement_type,omitempty"`
	StartDate    *time.Time `form:"start_date,omitempty"`
	EndDate      *time.Time `form:"end_date,omitempty"`
	Page         int        `form:"page" validate:"min=1"`
	Limit        int        `form:"limit" validate:"min=1,max=100"`
	Offset       int        `form:"offset"`
}
