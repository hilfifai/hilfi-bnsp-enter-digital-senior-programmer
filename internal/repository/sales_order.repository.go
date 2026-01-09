// internal/repository/sales_order.repository.go
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

type ISalesOrderRepository interface {
	Create(ctx context.Context, order *entity.SalesOrder) error
	CreateItem(ctx context.Context, item *entity.SalesOrderItem) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.SalesOrder, error)
	FindBySONumber(ctx context.Context, soNumber string) (*entity.SalesOrder, error)
	FindItemByID(ctx context.Context, id uuid.UUID) (*entity.SalesOrderItem, error)
	FindAll(ctx context.Context, filter entity.SalesOrderFilter) ([]entity.SalesOrder, error)
	Update(ctx context.Context, id uuid.UUID, order *entity.SalesOrder) error
	UpdateItem(ctx context.Context, id uuid.UUID, item *entity.SalesOrderItem) error
	GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.SalesOrderItem, error)
	GetNextSONumber(ctx context.Context) (string, error)
	GetPendingCount(ctx context.Context) (int, error)
	GetTotalSales(ctx context.Context, startDate, endDate time.Time) (float64, error)
	GetCount(ctx context.Context) (int, error)
	BeginTx(ctx context.Context) (interface{}, error)
	CreateTx(ctx context.Context, tx interface{}, order *entity.SalesOrder) error
	CreateItemTx(ctx context.Context, tx interface{}, item *entity.SalesOrderItem) error
	UpdateTx(ctx context.Context, tx interface{}, id uuid.UUID, order *entity.SalesOrder) error
	UpdateItemTx(ctx context.Context, tx interface{}, id uuid.UUID, item *entity.SalesOrderItem) error

	// Report methods
	GetSalesReport(ctx context.Context, filter dto.SalesReportFilter) ([]entity.SalesReport, error)
	GetSalesSummary(ctx context.Context, startDate, endDate time.Time) (*entity.ProfitLossReport, error)
	GetTopSellingProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.TopProduct, error)
}

type SalesOrderRepository struct {
	db *sqlx.DB
}

func NewSalesOrderRepository(db *sqlx.DB) ISalesOrderRepository {
	return &SalesOrderRepository{db: db}
}

func (r *SalesOrderRepository) Create(ctx context.Context, order *entity.SalesOrder) error {
	query := `
		INSERT INTO sales_orders (
			id, so_number, customer_id, warehouse_id, order_date, expected_delivery_date, status,
			sub_total, tax_amount, total_amount, notes, approved_by, approved_at,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :so_number, :customer_id, :warehouse_id, :order_date, :expected_delivery_date, :status,
			:sub_total, :tax_amount, :total_amount, :notes, :approved_by, :approved_at,
			:created_by, :updated_by, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, order)
	return err
}

func (r *SalesOrderRepository) GetNextSONumber(ctx context.Context) (string, error) {
	var lastNumber string
	query := `SELECT so_number FROM sales_orders 
	          WHERE so_number LIKE 'SO-' || to_char(CURRENT_DATE, 'YYYYMMDD') || '-%' 
	          ORDER BY so_number DESC LIMIT 1`

	err := r.db.GetContext(ctx, &lastNumber, query)
	if err != nil {
		return "SO-" + time.Now().Format("20060102") + "-0001", nil
	}

	parts := strings.Split(lastNumber, "-")
	if len(parts) == 3 {
		seq, _ := strconv.Atoi(parts[2])
		return fmt.Sprintf("SO-%s-%04d", time.Now().Format("20060102"), seq+1), nil
	}

	return "SO-" + time.Now().Format("20060102") + "-0001", nil
}

func (r *SalesOrderRepository) CreateItem(ctx context.Context, item *entity.SalesOrderItem) error {
	query := `
		INSERT INTO sales_order_items (
			id, sales_order_id, product_id, quantity, unit_price, total_price,
			shipped_qty, delivered_qty
		) VALUES (
			:id, :sales_order_id, :product_id, :quantity, :unit_price, :total_price,
			:shipped_qty, :delivered_qty
		)`
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *SalesOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.SalesOrder, error) {
	var order entity.SalesOrder
	query := `SELECT * FROM sales_orders WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &order, query, id)
	return &order, err
}

func (r *SalesOrderRepository) FindBySONumber(ctx context.Context, soNumber string) (*entity.SalesOrder, error) {
	var order entity.SalesOrder
	query := `SELECT * FROM sales_orders WHERE so_number = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &order, query, soNumber)
	return &order, err
}

func (r *SalesOrderRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*entity.SalesOrderItem, error) {
	var item entity.SalesOrderItem
	query := `SELECT * FROM sales_order_items WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &item, query, id)
	return &item, err
}

func (r *SalesOrderRepository) FindAll(ctx context.Context, filter entity.SalesOrderFilter) ([]entity.SalesOrder, error) {
	var orders []entity.SalesOrder
	query := `SELECT * FROM sales_orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.CustomerID != nil {
		query += fmt.Sprintf(` AND customer_id = $%d`, argIdx)
		args = append(args, *filter.CustomerID)
		argIdx++
	}
	if filter.WarehouseID != nil {
		query += fmt.Sprintf(` AND warehouse_id = $%d`, argIdx)
		args = append(args, *filter.WarehouseID)
		argIdx++
	}
	if filter.Status != nil {
		query += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	query += ` ORDER BY created_at DESC`
	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIdx)
		args = append(args, filter.Limit)
		argIdx++
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET $%d`, argIdx)
			args = append(args, filter.Offset)
		}
	}

	err := r.db.SelectContext(ctx, &orders, query, args...)
	return orders, err
}

func (r *SalesOrderRepository) Update(ctx context.Context, id uuid.UUID, order *entity.SalesOrder) error {
	query := `
		UPDATE sales_orders
		SET status = :status, notes = :notes, updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	order.ID = id
	_, err := r.db.NamedExecContext(ctx, query, order)
	return err
}

func (r *SalesOrderRepository) UpdateItem(ctx context.Context, id uuid.UUID, item *entity.SalesOrderItem) error {
	query := `
		UPDATE sales_order_items
		SET shipped_qty = :shipped_qty, delivered_qty = :delivered_qty
		WHERE id = :id`
	item.ID = id
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *SalesOrderRepository) GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.SalesOrderItem, error) {
	var items []entity.SalesOrderItem
	query := `SELECT * FROM sales_order_items WHERE sales_order_id = $1 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &items, query, orderID)
	return items, err
}

func (r *SalesOrderRepository) GetPendingCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sales_orders WHERE status = 'PENDING'`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *SalesOrderRepository) GetTotalSales(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(final_amount), 0) FROM sales_orders WHERE order_date >= $1 AND order_date < $2 AND status != 'CANCELLED'`
	err := r.db.GetContext(ctx, &total, query, startDate, endDate)
	return total, err
}

func (r *SalesOrderRepository) GetCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sales_orders`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *SalesOrderRepository) BeginTx(ctx context.Context) (interface{}, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *SalesOrderRepository) CreateTx(ctx context.Context, tx interface{}, order *entity.SalesOrder) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		INSERT INTO sales_orders (
			id, so_number, customer_id, warehouse_id, order_date, expected_delivery_date, status,
			sub_total, tax_amount, total_amount, notes, approved_by, approved_at,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :so_number, :customer_id, :warehouse_id, :order_date, :expected_delivery_date, :status,
			:sub_total, :tax_amount, :total_amount, :notes, :approved_by, :approved_at,
			:created_by, :updated_by, :created_at, :updated_at
		)`
	_, err := sqlxTx.NamedExecContext(ctx, query, order)
	return err
}

func (r *SalesOrderRepository) CreateItemTx(ctx context.Context, tx interface{}, item *entity.SalesOrderItem) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		INSERT INTO sales_order_items (
			id, sales_order_id, product_id, quantity, unit_price, total_price,
			shipped_qty, delivered_qty
		) VALUES (
			:id, :sales_order_id, :product_id, :quantity, :unit_price, :total_price,
			:shipped_qty, :delivered_qty
		)`
	_, err := sqlxTx.NamedExecContext(ctx, query, item)
	return err
}

func (r *SalesOrderRepository) UpdateTx(ctx context.Context, tx interface{}, id uuid.UUID, order *entity.SalesOrder) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		UPDATE sales_orders
		SET status = :status, notes = :notes, updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	order.ID = id
	_, err := sqlxTx.NamedExecContext(ctx, query, order)
	return err
}

func (r *SalesOrderRepository) UpdateItemTx(ctx context.Context, tx interface{}, id uuid.UUID, item *entity.SalesOrderItem) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		UPDATE sales_order_items
		SET shipped_qty = :shipped_qty, delivered_qty = :delivered_qty
		WHERE id = :id`
	item.ID = id
	_, err := sqlxTx.NamedExecContext(ctx, query, item)
	return err
}

func (r *SalesOrderRepository) GetSalesReport(ctx context.Context, filter dto.SalesReportFilter) ([]entity.SalesReport, error) {
	var reports []entity.SalesReport
	query := `
		SELECT 
			DATE_TRUNC('day', so.order_date) as date,
			COUNT(*) as total_orders,
			COALESCE(SUM(soi.quantity), 0) as total_quantity,
			COALESCE(SUM(so.total_amount), 0) as total_sales,
			COALESCE(SUM(soi.quantity * p.cost_price), 0) as total_cost,
			COALESCE(SUM(so.total_amount) - SUM(soi.quantity * p.cost_price), 0) as total_profit,
			COALESCE(AVG(so.total_amount), 0) as average_order
		FROM sales_orders so
		LEFT JOIN sales_order_items soi ON so.id = soi.sales_order_id
		LEFT JOIN products p ON soi.product_id = p.id
		WHERE so.status NOT IN ('DRAFT', 'CANCELLED')
	`
	args := []interface{}{}
	argIdx := 1

	if filter.StartDate != nil {
		query += fmt.Sprintf(` AND so.order_date >= $%d`, argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		query += fmt.Sprintf(` AND so.order_date <= $%d`, argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}
	if filter.CustomerID != nil {
		query += fmt.Sprintf(` AND so.customer_id = $%d`, argIdx)
		args = append(args, *filter.CustomerID)
		argIdx++
	}
	if filter.ProductID != nil {
		query += fmt.Sprintf(` AND soi.product_id = $%d`, argIdx)
		args = append(args, *filter.ProductID)
		argIdx++
	}

	query += ` GROUP BY DATE_TRUNC('day', so.order_date) ORDER BY date DESC`
	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIdx)
		args = append(args, filter.Limit)
	}

	err := r.db.SelectContext(ctx, &reports, query, args...)
	return reports, err
}

func (r *SalesOrderRepository) GetSalesSummary(ctx context.Context, startDate, endDate time.Time) (*entity.ProfitLossReport, error) {
	var report entity.ProfitLossReport
	// Convert dates to string format for PostgreSQL to avoid type inference issues
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")
	query := `
		SELECT 
			$1::date as start_date,
			$2::date as end_date,
			COALESCE(SUM(so.total_amount), 0) as total_sales,
			COALESCE(SUM(soi.quantity * p.cost_price), 0) as total_cost,
			COALESCE(SUM(so.total_amount) - SUM(soi.quantity * p.cost_price), 0) as gross_profit,
			0 as expenses,
			COALESCE(SUM(so.total_amount) - SUM(soi.quantity * p.cost_price), 0) as net_profit,
			CASE 
				WHEN SUM(so.total_amount) > 0 
				THEN ((SUM(so.total_amount) - SUM(soi.quantity * p.cost_price)) / SUM(so.total_amount)) * 100
				ELSE 0 
			END as profit_margin
		FROM sales_orders so
		LEFT JOIN sales_order_items soi ON so.id = soi.sales_order_id
		LEFT JOIN products p ON soi.product_id = p.id
		WHERE so.order_date >= $1::date AND so.order_date <= $2::date AND so.status NOT IN ('DRAFT', 'CANCELLED')
	`
	err := r.db.GetContext(ctx, &report, query, startDateStr, endDateStr)
	return &report, err
}

func (r *SalesOrderRepository) GetTopSellingProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.TopProduct, error) {
	var products []entity.TopProduct
	query := `
		SELECT 
			p.id as product_id,
			p.sku as product_sku,
			p.name as product_name,
			SUM(soi.quantity) as total_quantity,
			COALESCE(SUM(soi.quantity * soi.unit_price), 0) as total_revenue,
			COALESCE(SUM(soi.quantity * (soi.unit_price - p.cost_price)), 0) as total_profit,
			COUNT(DISTINCT so.id) as order_count,
			COALESCE(AVG(soi.unit_price), 0) as average_price
		FROM sales_orders so
		JOIN sales_order_items soi ON so.id = soi.sales_order_id
		JOIN products p ON soi.product_id = p.id
		WHERE so.order_date >= $1 AND so.order_date <= $2 AND so.status NOT IN ('DRAFT', 'CANCELLED')
		GROUP BY p.id, p.sku, p.name
		ORDER BY total_revenue DESC
		LIMIT $3
	`
	err := r.db.SelectContext(ctx, &products, query, startDate, endDate, limit)
	return products, err
}
