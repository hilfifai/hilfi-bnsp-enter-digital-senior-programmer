// internal/dto/sales.go
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateCustomerRequest struct {
	Code          string           `json:"code" validate:"required,min=3,max=50"`
	Name          string           `json:"name" validate:"required,min=3,max=200"`
	ContactPerson *string          `json:"contact_person,omitempty"`
	Email         *string          `json:"email,omitempty" validate:"omitempty,email"`
	Phone         *string          `json:"phone,omitempty"`
	Address       *string          `json:"address,omitempty"`
	TaxNumber     *string          `json:"tax_number,omitempty"`
	CreditLimit   *decimal.Decimal `json:"credit_limit,omitempty"`
	PaymentTerms  *string          `json:"payment_terms,omitempty"`
	IsActive      bool             `json:"is_active"`
}

type CreateSalesOrderRequest struct {
	CustomerID           uuid.UUID               `json:"customer_id" validate:"required"`
	WarehouseID          uuid.UUID               `json:"warehouse_id" validate:"required"`
	OrderDate            time.Time               `json:"order_date" validate:"required"`
	ExpectedDeliveryDate *time.Time              `json:"expected_delivery_date,omitempty"`
	Items                []SalesOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	Notes                *string                 `json:"notes,omitempty"`
}

type SalesOrderItemRequest struct {
	ProductID uuid.UUID       `json:"product_id" validate:"required"`
	Quantity  int             `json:"quantity" validate:"required,min=1"`
	UnitPrice decimal.Decimal `json:"unit_price" validate:"required"`
	Notes     *string         `json:"notes,omitempty"`
}

type CreateDeliveryOrderRequest struct {
	SalesOrderID    uuid.UUID                  `json:"sales_order_id" validate:"required"`
	WarehouseID     uuid.UUID                  `json:"warehouse_id" validate:"required"`
	DeliveryDate    time.Time                  `json:"delivery_date" validate:"required"`
	DeliveryAddress *string                    `json:"delivery_address,omitempty"`
	Items           []DeliveryOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	Notes           *string                    `json:"notes,omitempty"`
}

type DeliveryOrderItemRequest struct {
	SalesOrderItemID uuid.UUID  `json:"sales_order_item_id" validate:"required"`
	DeliveredQty     int        `json:"delivered_qty" validate:"required,min=1"`
	LocationID       *uuid.UUID `json:"location_id,omitempty"`
	BatchNumber      *string    `json:"batch_number,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
}

type SalesReturnRequest struct {
	DeliveryOrderID uuid.UUID                `json:"delivery_order_id" validate:"required"`
	ReturnDate      time.Time                `json:"return_date" validate:"required"`
	Items           []SalesReturnItemRequest `json:"items" validate:"required,min=1,dive"`
	Reason          string                   `json:"reason" validate:"required"`
	Notes           *string                  `json:"notes,omitempty"`
}

type SalesReturnItemRequest struct {
	DeliveryOrderItemID uuid.UUID `json:"delivery_order_item_id" validate:"required"`
	ReturnedQty         int       `json:"returned_qty" validate:"required,min=1"`
	Reason              string    `json:"reason" validate:"required"`
	Notes               *string   `json:"notes,omitempty"`
}

type UpdateCustomerRequest struct {
	Code          string           `json:"code" validate:"required,min=3,max=50"`
	Name          string           `json:"name" validate:"required,min=3,max=200"`
	ContactPerson *string          `json:"contact_person,omitempty"`
	Email         *string          `json:"email,omitempty" validate:"omitempty,email"`
	Phone         *string          `json:"phone,omitempty"`
	Address       *string          `json:"address,omitempty"`
	TaxNumber     *string          `json:"tax_number,omitempty"`
	CreditLimit   *decimal.Decimal `json:"credit_limit,omitempty"`
	PaymentTerms  *string          `json:"payment_terms,omitempty"`
	IsActive      bool             `json:"is_active"`
}

type SalesOrderFilter struct {
	CustomerID  *uuid.UUID `json:"customer_id,omitempty"`
	WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
	Status      *string    `json:"status,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

type DeliveryOrderFilter struct {
	SalesOrderID *uuid.UUID `json:"sales_order_id,omitempty"`
	WarehouseID  *uuid.UUID `json:"warehouse_id,omitempty"`
	Status       *string    `json:"status,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

type DeliveryOrderResponse struct {
	ID              uuid.UUID `json:"id"`
	DONumber        string    `json:"do_number"`
	SalesOrderID    uuid.UUID `json:"sales_order_id"`
	WarehouseID     uuid.UUID `json:"warehouse_id"`
	Status          string    `json:"status"`
	DeliveryDate    time.Time `json:"delivery_date"`
	DriverName      string    `json:"driver_name"`
	VehicleNumber   string    `json:"vehicle_number"`
	DeliveryAddress *string   `json:"delivery_address,omitempty"`
	Notes           *string   `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
