package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Warehouse struct {
	ID            uuid.UUID `db:"id" json:"id"`
	Code          string    `db:"code" json:"code"`
	Name          string    `db:"name" json:"name"`
	Address       *string   `db:"address" json:"address,omitempty"`
	ContactPerson *string   `db:"contact_person" json:"contact_person,omitempty"`
	Phone         *string   `db:"phone" json:"phone,omitempty"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedBy     uuid.UUID `db:"created_by" json:"-"`
	UpdatedBy     uuid.UUID `db:"updated_by" json:"-"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// WarehouseWithStats represents warehouse with location statistics
type WarehouseWithStats struct {
	*Warehouse
	LocationCount *int `db:"location_count" json:"location_count,omitempty"`
	TotalCapacity *int `db:"total_capacity" json:"total_capacity,omitempty"`
	UsedCapacity  *int `db:"used_capacity" json:"used_capacity,omitempty"`
}

// WarehouseLocationWithWarehouse represents warehouse location with warehouse info
type WarehouseLocationWithWarehouse struct {
	*WarehouseLocation
	WarehouseCode *string `db:"warehouse_code" json:"warehouse_code,omitempty"`
	WarehouseName *string `db:"warehouse_name" json:"warehouse_name,omitempty"`
}

type WarehouseLocation struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	WarehouseID  uuid.UUID  `db:"warehouse_id" json:"warehouse_id"`
	Warehouse    *Warehouse `db:"-" json:"warehouse,omitempty"`
	Code         string     `db:"code" json:"code"`
	Name         string     `db:"name" json:"name"`
	Description  *string    `db:"description" json:"description,omitempty"`
	Zone         *string    `db:"zone" json:"zone,omitempty"`
	Aisle        *string    `db:"aisle" json:"aisle,omitempty"`
	Rack         *string    `db:"rack" json:"rack,omitempty"`
	Shelf        *string    `db:"shelf" json:"shelf,omitempty"`
	Bin          *string    `db:"bin" json:"bin,omitempty"`
	Capacity     *int       `db:"capacity" json:"capacity,omitempty"`
	CurrentUsage int        `db:"current_usage" json:"current_usage"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedBy    uuid.UUID  `db:"created_by" json:"-"`
	UpdatedBy    uuid.UUID  `db:"updated_by" json:"-"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

type InventoryStock struct {
	ID                uuid.UUID          `db:"id" json:"id"`
	ProductID         uuid.UUID          `db:"product_id" json:"product_id"`
	Product           *Product           `db:"-" json:"product,omitempty"`
	ProductSKU        string             `db:"product_sku" json:"product_sku,omitempty"`
	ProductName       string             `db:"product_name" json:"product_name,omitempty"`
	WarehouseID       uuid.UUID          `db:"warehouse_id" json:"warehouse_id"`
	Warehouse         *Warehouse         `db:"-" json:"warehouse,omitempty"`
	WarehouseCode     string             `db:"warehouse_code" json:"warehouse_code,omitempty"`
	WarehouseName     string             `db:"warehouse_name" json:"warehouse_name,omitempty"`
	LocationID        *uuid.UUID         `db:"location_id" json:"location_id,omitempty"`
	Location          *WarehouseLocation `db:"-" json:"location,omitempty"`
	Quantity          int                `db:"quantity" json:"quantity"`
	ReservedQuantity  int                `db:"reserved_quantity" json:"reserved_quantity"`
	AvailableQuantity int                `db:"available_quantity" json:"available_quantity"`
	BatchNumber       *string            `db:"batch_number" json:"batch_number,omitempty"`
	ExpiryDate        *time.Time         `db:"expiry_date" json:"expiry_date,omitempty"`
	CreatedBy         uuid.UUID          `db:"created_by" json:"-"`
	UpdatedBy         uuid.UUID          `db:"updated_by" json:"-"`
	CreatedAt         time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `db:"updated_at" json:"updated_at"`
}

type StockMovement struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	ReferenceNumber string          `db:"reference_number" json:"reference_number"`
	MovementType    string          `db:"movement_type" json:"movement_type"`
	ProductID       uuid.UUID       `db:"product_id" json:"product_id"`
	Product         *Product        `db:"-" json:"product,omitempty"`
	ProductSKU      string          `db:"product_sku" json:"product_sku,omitempty"`
	ProductName     string          `db:"product_name" json:"product_name,omitempty"`
	FromWarehouseID *uuid.UUID      `db:"from_warehouse_id" json:"from_warehouse_id,omitempty"`
	FromLocationID  *uuid.UUID      `db:"from_location_id" json:"from_location_id,omitempty"`
	ToWarehouseID   *uuid.UUID      `db:"to_warehouse_id" json:"to_warehouse_id,omitempty"`
	ToLocationID    *uuid.UUID      `db:"to_location_id" json:"to_location_id,omitempty"`
	Quantity        int             `db:"quantity" json:"quantity"`
	UnitPrice       decimal.Decimal `db:"unit_price" json:"unit_price"`
	TotalValue      decimal.Decimal `db:"total_value" json:"total_value"`
	Notes           *string         `db:"notes" json:"notes,omitempty"`
	Status          string          `db:"status" json:"status"`
	MovementDate    time.Time       `db:"movement_date" json:"movement_date"`
	ApprovedBy      *uuid.UUID      `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `db:"approved_at" json:"approved_at,omitempty"`
	CreatedBy       uuid.UUID       `db:"created_by" json:"-"`
	UpdatedBy       uuid.UUID       `db:"updated_by" json:"-"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

type StockAdjustment struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	ReferenceNumber string     `db:"reference_number" json:"reference_number"`
	ProductID       uuid.UUID  `db:"product_id" json:"product_id"`
	WarehouseID     uuid.UUID  `db:"warehouse_id" json:"warehouse_id"`
	LocationID      *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	PreviousQty     int        `db:"previous_qty" json:"previous_qty"`
	NewQty          int        `db:"new_qty" json:"new_qty"`
	Difference      int        `db:"difference" json:"difference"`
	Reason          string     `db:"reason" json:"reason"`
	Notes           *string    `db:"notes" json:"notes,omitempty"`
	Status          string     `db:"status" json:"status"`
	AdjustedBy      uuid.UUID  `db:"adjusted_by" json:"-"`
	ApprovedBy      *uuid.UUID `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `db:"approved_at" json:"approved_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// WarehouseStats represents warehouse statistics
type WarehouseStats struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	TotalLocations int       `db:"total_locations" json:"total_locations"`
	TotalCapacity  int       `db:"total_capacity" json:"total_capacity"`
	UsedCapacity   int       `db:"used_capacity" json:"used_capacity"`
	TotalProducts  int       `db:"total_products" json:"total_products"`
	TotalQuantity  int       `db:"total_quantity" json:"total_quantity"`
	TotalReserved  int       `db:"total_reserved" json:"total_reserved"`
}

// ProductStock represents product stock information
type ProductStock struct {
	ID           uuid.UUID `db:"id" json:"id"`
	ProductID    uuid.UUID `db:"product_id" json:"product_id"`
	WarehouseID  uuid.UUID `db:"warehouse_id" json:"warehouse_id"`
	Quantity     int       `db:"quantity" json:"quantity"`
	Reserved     int       `db:"reserved" json:"reserved"`
	Available    int       `db:"available" json:"available"`
	ReorderLevel int       `db:"reorder_level" json:"reorder_level"`
	LastUpdated  time.Time `db:"last_updated" json:"last_updated"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// StockOpnameItem represents a stock opname item
type StockOpnameItem struct {
	ID          uuid.UUID `db:"id" json:"id"`
	OpnameID    uuid.UUID `db:"opname_id" json:"opname_id"`
	ProductID   uuid.UUID `db:"product_id" json:"product_id"`
	SystemQty   int       `db:"system_qty" json:"system_qty"`
	PhysicalQty int       `db:"physical_qty" json:"physical_qty"`
	Difference  int       `db:"difference" json:"difference"`
	Notes       *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// StockOpname represents a stock opname document
type StockOpname struct {
	ID           uuid.UUID         `db:"id" json:"id"`
	OpnameNumber string            `db:"opname_number" json:"opname_number"`
	WarehouseID  uuid.UUID         `db:"warehouse_id" json:"warehouse_id"`
	Status       string            `db:"status" json:"status"`
	OpnameDate   time.Time         `db:"opname_date" json:"opname_date"`
	CountedBy    *uuid.UUID        `db:"counted_by" json:"counted_by,omitempty"`
	VerifiedBy   *uuid.UUID        `db:"verified_by" json:"verified_by,omitempty"`
	Notes        *string           `db:"notes" json:"notes,omitempty"`
	Items        []StockOpnameItem `db:"-" json:"items,omitempty"`
	CreatedAt    time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `db:"updated_at" json:"updated_at"`
}

// StockSummary represents a summary of stock levels
type StockSummary struct {
	TotalProducts    int             `json:"total_products"`
	TotalItems       int             `json:"total_items"`
	TotalValue       decimal.Decimal `json:"total_value"`
	LowStockCount    int             `json:"low_stock_count"`
	OutOfStockCount  int             `json:"out_of_stock_count"`
	PendingMovements int             `json:"pending_movements"`
}

// StockAlert represents a stock alert
type StockAlert struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"product_id"`
	Product      *Product  `db:"-" json:"product,omitempty"`
	AlertType    string    `json:"alert_type"`
	CurrentStock int       `json:"current_stock"`
	Threshold    int       `json:"threshold"`
	Message      string    `json:"message"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
}
