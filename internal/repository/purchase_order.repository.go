// internal/repository/purchase_order.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IPurchaseOrderRepository interface {
	Create(ctx context.Context, order *entity.PurchaseOrder) error
	CreateItem(ctx context.Context, item *entity.PurchaseOrderItem) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrder, error)
	FindByPONumber(ctx context.Context, poNumber string) (*entity.PurchaseOrder, error)
	FindAll(ctx context.Context, filter entity.PurchaseOrderFilter) ([]entity.PurchaseOrder, error)
	Update(ctx context.Context, id uuid.UUID, order *entity.PurchaseOrder) error
	GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.PurchaseOrderItem, error)
	FindItemByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrderItem, error)
	UpdateItem(ctx context.Context, id uuid.UUID, item *entity.PurchaseOrderItem) error
	GetNextPONumber(ctx context.Context) (string, error)
	GetPendingCount(ctx context.Context) (int, error)
}

type PurchaseOrderRepository struct {
	db *sqlx.DB
}

func NewPurchaseOrderRepository(db *sqlx.DB) IPurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

func (r *PurchaseOrderRepository) Create(ctx context.Context, order *entity.PurchaseOrder) error {
	query := `
		INSERT INTO purchase_orders (
			id, po_number, supplier_id, order_date, expected_delivery_date, status,
			sub_total, tax_amount, total_amount, notes, approved_by, approved_at,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :po_number, :supplier_id, :order_date, :expected_delivery_date, :status,
			:sub_total, :tax_amount, :total_amount, :notes, :approved_by, :approved_at,
			:created_by, :updated_by, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, order)
	return err
}

func (r *PurchaseOrderRepository) CreateItem(ctx context.Context, item *entity.PurchaseOrderItem) error {
	query := `
		INSERT INTO purchase_order_items (
			id, purchase_order_id, product_id, quantity, unit_price, total_price,
			received_qty, remaining_qty, notes, created_at, updated_at
		) VALUES (
			:id, :purchase_order_id, :product_id, :quantity, :unit_price, :total_price,
			:received_qty, :remaining_qty, :notes, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *PurchaseOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrder, error) {
	var order entity.PurchaseOrder
	query := `SELECT * FROM purchase_orders WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &order, query, id)
	return &order, err
}

func (r *PurchaseOrderRepository) FindByPONumber(ctx context.Context, poNumber string) (*entity.PurchaseOrder, error) {
	var order entity.PurchaseOrder
	query := `SELECT * FROM purchase_orders WHERE po_number = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &order, query, poNumber)
	return &order, err
}

func (r *PurchaseOrderRepository) FindAll(ctx context.Context, filter entity.PurchaseOrderFilter) ([]entity.PurchaseOrder, error) {
	var orders []entity.PurchaseOrder
	query := `SELECT * FROM purchase_orders WHERE 1=1`
	args := []interface{}{}

	if filter.SupplierID != nil {
		query += ` AND supplier_id = ?`
		args = append(args, *filter.SupplierID)
	}
	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, *filter.Status)
	}

	query += ` ORDER BY created_at DESC`
	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET %d`, filter.Offset)
		}
	}

	err := r.db.SelectContext(ctx, &orders, query, args...)
	return orders, err
}

func (r *PurchaseOrderRepository) Update(ctx context.Context, id uuid.UUID, order *entity.PurchaseOrder) error {
	query := `
		UPDATE purchase_orders
		SET status = :status, notes = :notes, updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	order.ID = id
	_, err := r.db.NamedExecContext(ctx, query, order)
	return err
}

func (r *PurchaseOrderRepository) GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.PurchaseOrderItem, error) {
	var items []entity.PurchaseOrderItem
	query := `SELECT * FROM purchase_order_items WHERE purchase_order_id = $1 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &items, query, orderID)
	return items, err
}

func (r *PurchaseOrderRepository) GetNextPONumber(ctx context.Context) (string, error) {
	return "PO-" + time.Now().Format("20060102") + "-0001", nil
}

func (r *PurchaseOrderRepository) GetPendingCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM purchase_orders WHERE status = 'PENDING'`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *PurchaseOrderRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrderItem, error) {
	var item entity.PurchaseOrderItem
	query := `SELECT * FROM purchase_order_items WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &item, query, id)
	return &item, err
}

func (r *PurchaseOrderRepository) UpdateItem(ctx context.Context, id uuid.UUID, item *entity.PurchaseOrderItem) error {
	query := `
		UPDATE purchase_order_items
		SET received_qty = :received_qty, remaining_qty = :remaining_qty, updated_at = :updated_at
		WHERE id = :id`
	item.ID = id
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}
