package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PurchaseOrderFilter represents filter options for purchase order queries
type PurchaseOrderFilter struct {
	SupplierID *uuid.UUID
	Status     *string
	Limit      int
	Offset     int
}

type Supplier struct {
	ID            uuid.UUID `db:"id" json:"id"`
	Code          string    `db:"code" json:"code"`
	Name          string    `db:"name" json:"name"`
	ContactPerson *string   `db:"contact_person" json:"contact_person,omitempty"`
	Email         *string   `db:"email" json:"email,omitempty"`
	Phone         *string   `db:"phone" json:"phone,omitempty"`
	Address       *string   `db:"address" json:"address,omitempty"`
	TaxNumber     *string   `db:"tax_number" json:"tax_number,omitempty"`
	PaymentTerms  *string   `db:"payment_terms" json:"payment_terms,omitempty"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedBy     uuid.UUID `db:"created_by" json:"-"`
	UpdatedBy     uuid.UUID `db:"updated_by" json:"-"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type PurchaseOrder struct {
	ID                   uuid.UUID           `db:"id" json:"id"`
	PONumber             string              `db:"po_number" json:"po_number"`
	SupplierID           uuid.UUID           `db:"supplier_id" json:"supplier_id"`
	Supplier             *Supplier           `db:"-" json:"supplier,omitempty"`
	OrderDate            time.Time           `db:"order_date" json:"order_date"`
	ExpectedDeliveryDate *time.Time          `db:"expected_delivery_date" json:"expected_delivery_date,omitempty"`
	Status               string              `db:"status" json:"status"`
	SubTotal             decimal.Decimal     `db:"sub_total" json:"sub_total"`
	TaxAmount            decimal.Decimal     `db:"tax_amount" json:"tax_amount"`
	TotalAmount          decimal.Decimal     `db:"total_amount" json:"total_amount"`
	Notes                *string             `db:"notes" json:"notes,omitempty"`
	ApprovedBy           *uuid.UUID          `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt           *time.Time          `db:"approved_at" json:"approved_at,omitempty"`
	CreatedBy            uuid.UUID           `db:"created_by" json:"-"`
	UpdatedBy            uuid.UUID           `db:"updated_by" json:"-"`
	CreatedAt            time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time           `db:"updated_at" json:"updated_at"`
	Items                []PurchaseOrderItem `db:"-" json:"items,omitempty"`
}

type PurchaseOrderItem struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	PurchaseOrderID uuid.UUID       `db:"purchase_order_id" json:"purchase_order_id"`
	ProductID       uuid.UUID       `db:"product_id" json:"product_id"`
	Product         *Product        `db:"-" json:"product,omitempty"`
	Quantity        int             `db:"quantity" json:"quantity"`
	UnitPrice       decimal.Decimal `db:"unit_price" json:"unit_price"`
	TotalPrice      decimal.Decimal `db:"total_price" json:"total_price"`
	ReceivedQty     int             `db:"received_qty" json:"received_qty"`
	RemainingQty    int             `db:"remaining_qty" json:"remaining_qty"`
	Notes           *string         `db:"notes" json:"notes,omitempty"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

type GoodsReceipt struct {
	ID              uuid.UUID          `db:"id" json:"id"`
	GRNumber        string             `db:"gr_number" json:"gr_number"`
	PurchaseOrderID uuid.UUID          `db:"purchase_order_id" json:"purchase_order_id"`
	WarehouseID     uuid.UUID          `db:"warehouse_id" json:"warehouse_id"`
	ReceiptDate     time.Time          `db:"receipt_date" json:"receipt_date"`
	Status          string             `db:"status" json:"status"`
	Notes           *string            `db:"notes" json:"notes,omitempty"`
	ReceivedBy      uuid.UUID          `db:"received_by" json:"-"`
	ApprovedBy      *uuid.UUID         `db:"approved_by" json:"approved_by,omitempty"`
	CreatedAt       time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `db:"updated_at" json:"updated_at"`
	Items           []GoodsReceiptItem `db:"-" json:"items,omitempty"`
}

type GoodsReceiptItem struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	GoodsReceiptID      uuid.UUID  `db:"goods_receipt_id" json:"goods_receipt_id"`
	PurchaseOrderItemID uuid.UUID  `db:"purchase_order_item_id" json:"purchase_order_item_id"`
	ReceivedQty         int        `db:"received_qty" json:"received_qty"`
	LocationID          *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	BatchNumber         *string    `db:"batch_number" json:"batch_number,omitempty"`
	ExpiryDate          *time.Time `db:"expiry_date" json:"expiry_date,omitempty"`
	Notes               *string    `db:"notes" json:"notes,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
