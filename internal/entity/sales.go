package entity

import (
	"time"

	"github.com/google/uuid"
)

type SalesOrderStatus string

const (
	SOStatusDraft     SalesOrderStatus = "DRAFT"
	SOStatusPending   SalesOrderStatus = "PENDING"
	SOStatusConfirmed SalesOrderStatus = "CONFIRMED"
	SOStatusShipped   SalesOrderStatus = "SHIPPED"
	SOStatusDelivered SalesOrderStatus = "DELIVERED"
	SOStatusCancelled SalesOrderStatus = "CANCELLED"
)

type SalesOrder struct {
	ID                   uuid.UUID        `db:"id" json:"id"`
	SONumber             string           `db:"so_number" json:"so_number"`
	CustomerID           uuid.UUID        `db:"customer_id" json:"customer_id"`
	Customer             *Customer        `db:"-" json:"customer,omitempty"`
	WarehouseID          uuid.UUID        `db:"warehouse_id" json:"warehouse_id"`
	Status               SalesOrderStatus `db:"status" json:"status"`
	SubTotal             float64          `db:"sub_total" json:"sub_total"`
	TaxAmount            float64          `db:"tax_amount" json:"tax_amount"`
	TotalAmount          float64          `db:"total_amount" json:"total_amount"`
	OrderDate            time.Time        `db:"order_date" json:"order_date"`
	ExpectedDeliveryDate *time.Time       `db:"expected_delivery_date" json:"expected_delivery_date,omitempty"`
	DeliveryDate         *time.Time       `db:"delivery_date" json:"delivery_date,omitempty"`
	ApprovedBy           *uuid.UUID       `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt           *time.Time       `db:"approved_at" json:"approved_at,omitempty"`
	CreatedBy            uuid.UUID        `db:"created_by" json:"-"`
	UpdatedBy            uuid.UUID        `db:"updated_by" json:"-"`
	Notes                *string          `db:"notes" json:"notes,omitempty"`
	CreatedAt            time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time        `db:"updated_at" json:"updated_at"`
	Items                []SalesOrderItem `db:"-" json:"items,omitempty"`
}

type SalesOrderItem struct {
	ID           uuid.UUID `db:"id" json:"id"`
	SalesOrderID uuid.UUID `db:"sales_order_id" json:"sales_order_id"`
	ProductID    uuid.UUID `db:"product_id" json:"product_id"`
	Quantity     int       `db:"quantity" json:"quantity"`
	UnitPrice    float64   `db:"unit_price" json:"unit_price"`
	TotalPrice   float64   `db:"total_price" json:"total_price"`
	ShippedQty   int       `db:"shipped_qty" json:"shipped_qty"`
	DeliveredQty int       `db:"delivered_qty" json:"delivered_qty"`
	RemainingQty int       `db:"remaining_qty" json:"remaining_qty"`
	Notes        *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Customer struct {
	ID            uuid.UUID `db:"id" json:"id"`
	Code          string    `db:"code" json:"code"`
	Name          string    `db:"name" json:"name"`
	ContactPerson string    `db:"contact_person" json:"contact_person"`
	Phone         string    `db:"phone" json:"phone"`
	Email         string    `db:"email" json:"email"`
	Address       string    `db:"address" json:"address"`
	TaxNumber     string    `db:"tax_number" json:"tax_number"`
	CustomerType  string    `db:"customer_type" json:"customer_type"`
	CreditLimit   float64   `db:"credit_limit" json:"credit_limit"`
	PaymentTerms  int       `db:"payment_terms" json:"payment_terms"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	Status        string    `db:"status" json:"status"`
	CreatedBy     uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedBy     uuid.UUID `db:"updated_by" json:"updated_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type DeliveryOrder struct {
	ID              uuid.UUID           `db:"id" json:"id"`
	DONumber        string              `db:"do_number" json:"do_number"`
	SalesOrderID    uuid.UUID           `db:"sales_order_id" json:"sales_order_id"`
	WarehouseID     uuid.UUID           `db:"warehouse_id" json:"warehouse_id"`
	Status          string              `db:"status" json:"status"`
	DeliveryDate    time.Time           `db:"delivery_date" json:"delivery_date"`
	DriverName      string              `db:"driver_name" json:"driver_name"`
	VehicleNumber   string              `db:"vehicle_number" json:"vehicle_number"`
	DeliveryAddress *string             `db:"delivery_address" json:"delivery_address,omitempty"`
	Notes           *string             `db:"notes" json:"notes,omitempty"`
	DeliveredBy     *uuid.UUID          `db:"delivered_by" json:"delivered_by,omitempty"`
	CreatedBy       uuid.UUID           `db:"created_by" json:"-"`
	UpdatedBy       uuid.UUID           `db:"updated_by" json:"-"`
	CreatedAt       time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `db:"updated_at" json:"updated_at"`
	Items           []DeliveryOrderItem `db:"-" json:"items,omitempty"`
}

type DeliveryOrderItem struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	DeliveryOrderID  uuid.UUID  `db:"delivery_order_id" json:"delivery_order_id"`
	SalesOrderItemID uuid.UUID  `db:"sales_order_item_id" json:"sales_order_item_id"`
	DeliveredQty     int        `db:"delivered_qty" json:"delivered_qty"`
	LocationID       *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	BatchNumber      *string    `db:"batch_number" json:"batch_number,omitempty"`
	Notes            *string    `db:"notes" json:"notes,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// SalesOrderFilter represents filter options for sales order queries
type SalesOrderFilter struct {
	CustomerID  *uuid.UUID
	WarehouseID *uuid.UUID
	Status      *SalesOrderStatus
	StartDate   *time.Time
	EndDate     *time.Time
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
}

// SalesReturn represents a sales return record
type SalesReturn struct {
	ID              uuid.UUID `db:"id" json:"id"`
	ReturnNumber    string    `db:"return_number" json:"return_number"`
	DeliveryOrderID uuid.UUID `db:"delivery_order_id" json:"delivery_order_id"`
	CustomerID      uuid.UUID `db:"customer_id" json:"customer_id"`
	ReturnDate      time.Time `db:"return_date" json:"return_date"`
	Reason          string    `db:"reason" json:"reason"`
	Status          string    `db:"status" json:"status"`
	TotalRefund     float64   `db:"total_refund" json:"total_refund"`
	Notes           *string   `db:"notes" json:"notes,omitempty"`
	ProcessedBy     uuid.UUID `db:"processed_by" json:"processed_by"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// SalesReturnItem represents a sales return item
type SalesReturnItem struct {
	ID                  uuid.UUID `db:"id" json:"id"`
	SalesReturnID       uuid.UUID `db:"sales_return_id" json:"sales_return_id"`
	DeliveryOrderItemID uuid.UUID `db:"delivery_order_item_id" json:"delivery_order_item_id"`
	ProductID           uuid.UUID `db:"product_id" json:"product_id"`
	ReturnedQty         int       `db:"returned_qty" json:"returned_qty"`
	UnitRefund          float64   `db:"unit_refund" json:"unit_refund"`
	TotalRefund         float64   `db:"total_refund" json:"total_refund"`
	Reason              string    `db:"reason" json:"reason"`
	Notes               *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}
