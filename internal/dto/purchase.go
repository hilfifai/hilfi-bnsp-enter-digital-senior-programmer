// internal/dto/purchase.go
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateSupplierRequest struct {
	Code          string  `json:"code" validate:"required,min=3,max=50"`
	Name          string  `json:"name" validate:"required,min=3,max=200"`
	ContactPerson *string `json:"contact_person,omitempty"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone         *string `json:"phone,omitempty"`
	Address       *string `json:"address,omitempty"`
	TaxNumber     *string `json:"tax_number,omitempty"`
	PaymentTerms  *string `json:"payment_terms,omitempty"`
	IsActive      bool    `json:"is_active"`
}

type UpdateSupplierRequest struct {
	Code          string  `json:"code" validate:"required,min=3,max=50"`
	Name          string  `json:"name" validate:"required,min=3,max=200"`
	ContactPerson *string `json:"contact_person,omitempty"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone         *string `json:"phone,omitempty"`
	Address       *string `json:"address,omitempty"`
	TaxNumber     *string `json:"tax_number,omitempty"`
	PaymentTerms  *string `json:"payment_terms,omitempty"`
	IsActive      bool    `json:"is_active"`
}

type CreatePurchaseOrderRequest struct {
	SupplierID           uuid.UUID                  `json:"supplier_id" validate:"required"`
	OrderDate            time.Time                  `json:"order_date" validate:"required"`
	ExpectedDeliveryDate *time.Time                 `json:"expected_delivery_date,omitempty"`
	Items                []PurchaseOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	Notes                *string                    `json:"notes,omitempty"`
}

type PurchaseOrderItemRequest struct {
	ProductID uuid.UUID       `json:"product_id" validate:"required"`
	Quantity  int             `json:"quantity" validate:"required,min=1"`
	UnitPrice decimal.Decimal `json:"unit_price" validate:"required"`
	Notes     *string         `json:"notes,omitempty"`
}

type GoodsReceiptRequest struct {
	PurchaseOrderID uuid.UUID                 `json:"purchase_order_id" validate:"required"`
	WarehouseID     uuid.UUID                 `json:"warehouse_id" validate:"required"`
	ReceiptDate     time.Time                 `json:"receipt_date" validate:"required"`
	Items           []GoodsReceiptItemRequest `json:"items" validate:"required,min=1,dive"`
	Notes           *string                   `json:"notes,omitempty"`
}

type GoodsReceiptItemRequest struct {
	PurchaseOrderItemID uuid.UUID  `json:"purchase_order_item_id" validate:"required"`
	ReceivedQty         int        `json:"received_qty" validate:"required,min=0"`
	LocationID          *uuid.UUID `json:"location_id,omitempty"`
	BatchNumber         *string    `json:"batch_number,omitempty"`
	ExpiryDate          *time.Time `json:"expiry_date,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
}

type PurchaseOrderFilter struct {
	SupplierID *uuid.UUID `json:"supplier_id,omitempty"`
	Status     *string    `json:"status,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}
