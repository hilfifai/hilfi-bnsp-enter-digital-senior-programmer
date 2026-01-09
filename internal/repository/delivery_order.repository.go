// internal/repository/delivery_order.repository.go
package repository

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IDeliveryOrderRepository interface {
	Create(ctx context.Context, do *entity.DeliveryOrder) error
	CreateItem(ctx context.Context, item *entity.DeliveryOrderItem) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrder, error)
	FindByDONumber(ctx context.Context, doNumber string) (*entity.DeliveryOrder, error)
	FindBySalesOrderID(ctx context.Context, salesOrderID uuid.UUID) ([]entity.DeliveryOrder, error)
	FindWithFilter(ctx context.Context, filter dto.DeliveryOrderFilter) ([]entity.DeliveryOrder, int, error)
	Update(ctx context.Context, id uuid.UUID, do *entity.DeliveryOrder) error
	GetItemsByDOID(ctx context.Context, doID uuid.UUID) ([]entity.DeliveryOrderItem, error)
	GetNextDONumber(ctx context.Context) (string, error)
	BeginTx(ctx context.Context) (interface{}, error)
	CreateTx(ctx context.Context, tx interface{}, do *entity.DeliveryOrder) error
	CreateItemTx(ctx context.Context, tx interface{}, item *entity.DeliveryOrderItem) error
	UpdateTx(ctx context.Context, tx interface{}, id uuid.UUID, do *entity.DeliveryOrder) error
	UpdateItemTx(ctx context.Context, tx interface{}, id uuid.UUID, item *entity.DeliveryOrderItem) error

	// Sales return methods
	CreateSalesReturn(ctx context.Context, sr *entity.SalesReturn) error
	CreateSalesReturnItem(ctx context.Context, item *entity.SalesReturnItem) error
	GetNextReturnNumber(ctx context.Context) (string, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrderItem, error)
}

type DeliveryOrderRepository struct {
	db *sqlx.DB
}

func NewDeliveryOrderRepository(db *sqlx.DB) IDeliveryOrderRepository {
	return &DeliveryOrderRepository{db: db}
}

func (r *DeliveryOrderRepository) Create(ctx context.Context, do *entity.DeliveryOrder) error {
	query := `
		INSERT INTO delivery_orders (
			id, do_number, sales_order_id, warehouse_id, delivery_date, status,
			driver_name, vehicle_number, notes, delivered_by,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :do_number, :sales_order_id, :warehouse_id, :delivery_date, :status,
			:driver_name, :vehicle_number, :notes, :delivered_by,
			:created_by, :updated_by, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, do)
	return err
}

func (r *DeliveryOrderRepository) CreateItem(ctx context.Context, item *entity.DeliveryOrderItem) error {
	query := `
		INSERT INTO delivery_order_items (
			id, delivery_order_id, sales_order_item_id, delivered_qty, location_id, batch_number, notes
		) VALUES (
			:id, :delivery_order_id, :sales_order_item_id, :delivered_qty, :location_id, :batch_number, :notes
		)`

	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *DeliveryOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrder, error) {
	var do entity.DeliveryOrder
	query := `SELECT * FROM delivery_orders WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &do, query, id)
	return &do, err
}

func (r *DeliveryOrderRepository) FindByDONumber(ctx context.Context, doNumber string) (*entity.DeliveryOrder, error) {
	var do entity.DeliveryOrder
	query := `SELECT * FROM delivery_orders WHERE do_number = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &do, query, doNumber)
	return &do, err
}

func (r *DeliveryOrderRepository) FindBySalesOrderID(ctx context.Context, salesOrderID uuid.UUID) ([]entity.DeliveryOrder, error) {
	var dos []entity.DeliveryOrder
	query := `SELECT * FROM delivery_orders WHERE sales_order_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &dos, query, salesOrderID)
	return dos, err
}

func (r *DeliveryOrderRepository) FindWithFilter(ctx context.Context, filter dto.DeliveryOrderFilter) ([]entity.DeliveryOrder, int, error) {
	var dos []entity.DeliveryOrder
	var total int

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.SalesOrderID != nil {
		whereClause += fmt.Sprintf(" AND sales_order_id = $%d", argIndex)
		args = append(args, *filter.SalesOrderID)
		argIndex++
	}
	if filter.WarehouseID != nil {
		whereClause += fmt.Sprintf(" AND warehouse_id = $%d", argIndex)
		args = append(args, *filter.WarehouseID)
		argIndex++
	}
	if filter.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}
	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND delivery_date >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}
	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND delivery_date <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM delivery_orders %s", whereClause)
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := filter.Offset
	whereClause += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Fetch data
	query := fmt.Sprintf("SELECT * FROM delivery_orders %s", whereClause)
	err = r.db.SelectContext(ctx, &dos, query, args...)
	return dos, total, err
}

func (r *DeliveryOrderRepository) Update(ctx context.Context, id uuid.UUID, do *entity.DeliveryOrder) error {
	query := `
		UPDATE delivery_orders
		SET status = :status, driver_name = :driver_name, vehicle_number = :vehicle_number,
			delivery_address = :delivery_address, notes = :notes,
			updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	do.ID = id
	_, err := r.db.NamedExecContext(ctx, query, do)
	return err
}

func (r *DeliveryOrderRepository) GetItemsByDOID(ctx context.Context, doID uuid.UUID) ([]entity.DeliveryOrderItem, error) {
	var items []entity.DeliveryOrderItem
	query := `SELECT * FROM delivery_order_items WHERE delivery_order_id = $1 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &items, query, doID)
	return items, err
}

func (r *DeliveryOrderRepository) GetNextDONumber(ctx context.Context) (string, error) {
	var lastNumber string
	query := `SELECT do_number FROM delivery_orders 
	          WHERE do_number LIKE 'DO-' || to_char(CURRENT_DATE, 'YYYYMMDD') || '-%' 
	          ORDER BY do_number DESC LIMIT 1`

	err := r.db.GetContext(ctx, &lastNumber, query)
	if err != nil {
		return "DO-" + time.Now().Format("20060102") + "-0001", nil
	}

	parts := strings.Split(lastNumber, "-")
	if len(parts) == 3 {
		seq, _ := strconv.Atoi(parts[2])
		return fmt.Sprintf("DO-%s-%04d", time.Now().Format("20060102"), seq+1), nil
	}

	return "DO-" + time.Now().Format("20060102") + "-0001", nil
}

func (r *DeliveryOrderRepository) BeginTx(ctx context.Context) (interface{}, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *DeliveryOrderRepository) CreateTx(ctx context.Context, tx interface{}, do *entity.DeliveryOrder) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		INSERT INTO delivery_orders (
			id, do_number, sales_order_id, warehouse_id, delivery_date, status,
			driver_name, vehicle_number, notes, delivered_by,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :do_number, :sales_order_id, :warehouse_id, :delivery_date, :status,
			:driver_name, :vehicle_number, :notes, :delivered_by,
			:created_by, :updated_by, :created_at, :updated_at
		)`
	_, err := sqlxTx.NamedExecContext(ctx, query, do)
	return err
}

func (r *DeliveryOrderRepository) CreateItemTx(ctx context.Context, tx interface{}, item *entity.DeliveryOrderItem) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		INSERT INTO delivery_order_items (
			id, delivery_order_id, sales_order_item_id, delivered_qty, location_id, batch_number, notes
		) VALUES (
			:id, :delivery_order_id, :sales_order_item_id, :delivered_qty, :location_id, :batch_number, :notes
		)`
	_, err := sqlxTx.NamedExecContext(ctx, query, item)
	return err
}

func (r *DeliveryOrderRepository) UpdateTx(ctx context.Context, tx interface{}, id uuid.UUID, do *entity.DeliveryOrder) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		UPDATE delivery_orders
		SET status = :status, driver_name = :driver_name, vehicle_number = :vehicle_number,
			delivery_address = :delivery_address, notes = :notes,
			updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	do.ID = id
	_, err := sqlxTx.NamedExecContext(ctx, query, do)
	return err
}

// Sales return methods

func (r *DeliveryOrderRepository) UpdateItemTx(ctx context.Context, tx interface{}, id uuid.UUID, item *entity.DeliveryOrderItem) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		UPDATE delivery_order_items
		SET delivered_qty = :delivered_qty, notes = :notes, updated_at = :updated_at
		WHERE id = :id`
	item.ID = id
	_, err := sqlxTx.NamedExecContext(ctx, query, item)
	return err
}

func (r *DeliveryOrderRepository) CreateSalesReturn(ctx context.Context, sr *entity.SalesReturn) error {
	query := `
		INSERT INTO sales_returns (
			id, return_number, delivery_order_id, customer_id, return_date, reason, status,
			total_refund, notes, processed_by, created_at, updated_at
		) VALUES (
			:id, :return_number, :delivery_order_id, :customer_id, :return_date, :reason, :status,
			:total_refund, :notes, :processed_by, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, query, sr)
	return err
}

func (r *DeliveryOrderRepository) CreateSalesReturnItem(ctx context.Context, item *entity.SalesReturnItem) error {
	query := `
		INSERT INTO sales_return_items (
			id, sales_return_id, delivery_order_item_id, product_id, returned_qty,
			unit_refund, total_refund, reason, notes, created_at, updated_at
		) VALUES (
			:id, :sales_return_id, :delivery_order_item_id, :product_id, :returned_qty,
			:unit_refund, :total_refund, :reason, :notes, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *DeliveryOrderRepository) GetNextReturnNumber(ctx context.Context) (string, error) {
	var lastNumber string
	query := `SELECT return_number FROM sales_returns
	          WHERE return_number LIKE 'RET-' || to_char(CURRENT_DATE, 'YYYYMMDD') || '-%'
	          ORDER BY return_number DESC LIMIT 1`

	err := r.db.GetContext(ctx, &lastNumber, query)
	if err != nil {
		return "RET-" + time.Now().Format("20060102") + "-0001", nil
	}

	parts := strings.Split(lastNumber, "-")
	if len(parts) == 3 {
		seq, _ := strconv.Atoi(parts[2])
		return fmt.Sprintf("RET-%s-%04d", time.Now().Format("20060102"), seq+1), nil
	}

	return "RET-" + time.Now().Format("20060102") + "-0001", nil
}

func (r *DeliveryOrderRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrderItem, error) {
	var item entity.DeliveryOrderItem
	query := `SELECT * FROM delivery_order_items WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &item, query, id)
	return &item, err
}
