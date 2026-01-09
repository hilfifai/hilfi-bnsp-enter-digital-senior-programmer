package repository

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IInventoryRepository interface {
	// Stock management
	GetStockByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID, locationID *uuid.UUID) (*entity.InventoryStock, error)
	GetStockByProductAndWarehouseTx(ctx context.Context, tx interface{}, productID, warehouseID uuid.UUID, locationID *uuid.UUID) (*entity.InventoryStock, error)
	CreateStock(ctx context.Context, stock *entity.InventoryStock) error
	CreateStockTx(ctx context.Context, tx interface{}, stock *entity.InventoryStock) error
	UpdateStock(ctx context.Context, id uuid.UUID, stock *entity.InventoryStock) error
	UpdateStockTx(ctx context.Context, tx interface{}, id uuid.UUID, stock *entity.InventoryStock) error

	// Stock movement
	CreateStockMovement(ctx context.Context, movement *entity.StockMovement) error
	CreateStockMovementTx(ctx context.Context, tx interface{}, movement *entity.StockMovement) error
	UpdateStockMovement(ctx context.Context, tx interface{}, movement *entity.StockMovement) error

	// Product stock
	GetProductStockByProductID(ctx context.Context, productID uuid.UUID) ([]entity.InventoryStock, error)
	GetStockByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.InventoryStock, error)
	GetStockByWarehouseAndProduct(ctx context.Context, warehouseID, productID uuid.UUID) ([]entity.InventoryStock, error)

	// Warehouse
	GetWarehouseByID(ctx context.Context, warehouseID uuid.UUID) (*entity.Warehouse, error)

	// Transaction
	BeginTx(ctx context.Context) (interface{}, error)

	// Additional methods for dashboard
	GetLowStockCount(ctx context.Context) (int, error)
	GetOutOfStockCount(ctx context.Context) (int, error)
	GetTotalProducts(ctx context.Context) (int, error)
	GetPendingMovementsCount(ctx context.Context) (int, error)

	// Report methods
	GetStockReport(ctx context.Context, filter dto.StockReportFilter) ([]entity.StockReport, error)
	GetLowStockReport(ctx context.Context) ([]entity.StockReport, error)
	GetMovementReport(ctx context.Context, filter dto.MovementReportFilter) ([]entity.StockMovement, int, error)
	GetInventoryValuation(ctx context.Context, warehouseID *uuid.UUID) ([]entity.InventoryValuation, error)
	GetAllWarehouses(ctx context.Context) ([]entity.Warehouse, error)

	// Stock opname
	CreateStockOpname(ctx context.Context, opname *entity.StockOpname) error
	GetProductStockHistory(ctx context.Context, productID, warehouseID uuid.UUID, days int) ([]entity.StockMovement, error)
}

type InventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) RecordStockMovement(ctx context.Context, movement *entity.StockMovement) error {
	query := `
		INSERT INTO stock_movements (
			id, product_id, warehouse_id, movement_type, quantity,
			reference_id, reference_type, notes, created_by, created_at
		) VALUES (
			:id, :product_id, :warehouse_id, :movement_type, :quantity,
			:reference_id, :reference_type, :notes, :created_by, :created_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, movement)
	return err
}

func (r *InventoryRepository) GetStockSummary(ctx context.Context, warehouseID uuid.UUID) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			p.id as product_id,
			p.sku,
			p.name,
			ps.quantity,
			ps.reserved,
			ps.available,
			ps.reorder_level,
			pc.name as category_name
		FROM products p
		LEFT JOIN product_stocks ps ON p.id = ps.product_id AND ps.warehouse_id = $1
		LEFT JOIN product_categories pc ON p.category_id = pc.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.name
	`

	var results []map[string]interface{}
	err := r.db.SelectContext(ctx, &results, query, warehouseID)
	return results, err
}

func (r *InventoryRepository) CreateStockOpname(ctx context.Context, opname *entity.StockOpname) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert opname header
	query := `
		INSERT INTO stock_opnames (
			id, opname_number, warehouse_id, status, opname_date,
			counted_by, verified_by, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, query,
		opname.ID, opname.OpnameNumber, opname.WarehouseID, opname.Status,
		opname.OpnameDate, opname.CountedBy, opname.VerifiedBy,
		opname.Notes, opname.CreatedAt, opname.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert opname items
	for _, item := range opname.Items {
		itemQuery := `
			INSERT INTO stock_opname_items (
				id, opname_id, product_id, system_qty, physical_qty,
				difference, notes, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID, opname.ID, item.ProductID, item.SystemQty,
			item.PhysicalQty, item.Difference, item.Notes, time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *InventoryRepository) GetWarehouseStockValue(ctx context.Context, warehouseID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(ps.available * p.cost), 0) as total_value
		FROM product_stocks ps
		JOIN products p ON ps.product_id = p.id
		WHERE ps.warehouse_id = $1 AND p.deleted_at IS NULL
	`

	var totalValue float64
	err := r.db.GetContext(ctx, &totalValue, query, warehouseID)
	return totalValue, err
}

func (r *InventoryRepository) GetStockByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID, locationID *uuid.UUID) (*entity.InventoryStock, error) {
	var stock entity.InventoryStock
	query := `SELECT * FROM inventory_stocks WHERE product_id = $1 AND warehouse_id = $2`
	args := []interface{}{productID, warehouseID}

	if locationID != nil {
		query += ` AND location_id = $3`
		args = append(args, *locationID)
	} else {
		query += ` AND location_id IS NULL`
	}

	query += ` LIMIT 1`
	err := r.db.GetContext(ctx, &stock, query, args...)
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (r *InventoryRepository) GetStockByProductAndWarehouseTx(ctx context.Context, tx interface{}, productID, warehouseID uuid.UUID, locationID *uuid.UUID) (*entity.InventoryStock, error) {
	// For now, just call the non-tx version
	return r.GetStockByProductAndWarehouse(ctx, productID, warehouseID, locationID)
}

func (r *InventoryRepository) CreateStock(ctx context.Context, stock *entity.InventoryStock) error {
	query := `
		INSERT INTO inventory_stocks (
			id, product_id, warehouse_id, location_id, quantity, reserved_quantity,
			batch_number, expiry_date, created_by, updated_by,
			created_at, updated_at
		) VALUES (
			:id, :product_id, :warehouse_id, :location_id, :quantity, :reserved_quantity,
			:batch_number, :expiry_date, :created_by, :updated_by,
			:created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, stock)
	return err
}

func (r *InventoryRepository) CreateStockTx(ctx context.Context, tx interface{}, stock *entity.InventoryStock) error {
	return r.CreateStock(ctx, stock)
}

func (r *InventoryRepository) UpdateStock(ctx context.Context, id uuid.UUID, stock *entity.InventoryStock) error {
	query := `
		UPDATE inventory_stocks
		SET quantity = :quantity, reserved_quantity = :reserved_quantity, updated_by = :updated_by,
			updated_at = :updated_at
		WHERE id = :id
	`
	stock.ID = id
	_, err := r.db.NamedExecContext(ctx, query, stock)
	return err
}

func (r *InventoryRepository) UpdateStockTx(ctx context.Context, tx interface{}, id uuid.UUID, stock *entity.InventoryStock) error {
	return r.UpdateStock(ctx, id, stock)
}

func (r *InventoryRepository) CreateStockMovement(ctx context.Context, movement *entity.StockMovement) error {
	query := `
		INSERT INTO stock_movements (
			id, reference_number, movement_type, product_id, from_warehouse_id,
			from_location_id, to_warehouse_id, to_location_id, quantity, unit_price,
			total_value, notes, status, movement_date, approved_by, approved_at,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :reference_number, :movement_type, :product_id, :from_warehouse_id,
			:from_location_id, :to_warehouse_id, :to_location_id, :quantity, :unit_price,
			:total_value, :notes, :status, :movement_date, :approved_by, :approved_at,
			:created_by, :updated_by, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, movement)
	return err
}

func (r *InventoryRepository) CreateStockMovementTx(ctx context.Context, tx interface{}, movement *entity.StockMovement) error {
	query := `
		INSERT INTO stock_movements (
			id, reference_number, movement_type, product_id, from_warehouse_id,
			from_location_id, to_warehouse_id, to_location_id, quantity, unit_price,
			total_value, notes, status, movement_date, approved_by, approved_at,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :reference_number, :movement_type, :product_id, :from_warehouse_id,
			:from_location_id, :to_warehouse_id, :to_location_id, :quantity, :unit_price,
			:total_value, :notes, :status, :movement_date, :approved_by, :approved_at,
			:created_by, :updated_by, :created_at, :updated_at
		)
	`
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	_, err := sqlxTx.NamedExecContext(ctx, query, movement)
	return err
}

func (r *InventoryRepository) UpdateStockMovement(ctx context.Context, tx interface{}, movement *entity.StockMovement) error {
	query := `
		UPDATE stock_movements
		SET status = :status, updated_at = :updated_at
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, movement)
	return err
}

func (r *InventoryRepository) GetProductStockByProductID(ctx context.Context, productID uuid.UUID) ([]entity.InventoryStock, error) {
	var stocks []entity.InventoryStock
	query := `
		SELECT istok.*, w.code as warehouse_code, w.name as warehouse_name
		FROM inventory_stocks istok
		JOIN warehouses w ON istok.warehouse_id = w.id
		WHERE istok.product_id = $1
		ORDER BY w.name
	`
	err := r.db.SelectContext(ctx, &stocks, query, productID)
	return stocks, err
}

func (r *InventoryRepository) GetWarehouseByID(ctx context.Context, warehouseID uuid.UUID) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	query := `SELECT * FROM warehouses WHERE id = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &warehouse, query, warehouseID)
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

func (r *InventoryRepository) BeginTx(ctx context.Context) (interface{}, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *InventoryRepository) GetLowStockCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products WHERE current_stock <= min_stock AND is_active = true`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *InventoryRepository) GetTotalProducts(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products WHERE is_active = true`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *InventoryRepository) GetStockReport(ctx context.Context, filter dto.StockReportFilter) ([]entity.StockReport, error) {
	var reports []entity.StockReport
	query := `
		SELECT 
			p.id as product_id,
			p.sku as product_sku,
			p.name as product_name,
			pc.name as category_name,
			w.name as warehouse_name,
			wl.code as location_code,
			COALESCE(istok.quantity, 0) as current_stock,
			COALESCE(istok.reserved_quantity, 0) as reserved_stock,
			COALESCE(istok.available_quantity, 0) as available_stock,
			p.min_stock,
			p.max_stock,
			p.unit_price,
			COALESCE(istok.available_quantity, 0) * p.unit_price as total_value,
			(SELECT MAX(created_at) FROM stock_movements WHERE product_id = p.id) as last_movement
		FROM products p
		LEFT JOIN product_categories pc ON p.category_id = pc.id
		LEFT JOIN inventory_stocks istok ON p.id = istok.product_id
		LEFT JOIN warehouses w ON istok.warehouse_id = w.id
		LEFT JOIN warehouse_locations wl ON istok.location_id = wl.id
		WHERE p.deleted_at IS NULL
	`
	args := []interface{}{}
	argIdx := 1

	if filter.WarehouseID != nil {
		query += fmt.Sprintf(` AND istok.warehouse_id = $%d`, argIdx)
		args = append(args, *filter.WarehouseID)
		argIdx++
	}
	if filter.ProductID != nil {
		query += fmt.Sprintf(` AND p.id = $%d`, argIdx)
		args = append(args, *filter.ProductID)
		argIdx++
	}
	if filter.CategoryID != nil {
		query += fmt.Sprintf(` AND p.category_id = $%d`, argIdx)
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	query += ` ORDER BY p.name`
	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIdx)
		args = append(args, filter.Limit)
		argIdx++
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET $%d`, argIdx)
			args = append(args, filter.Offset)
		}
	}

	err := r.db.SelectContext(ctx, &reports, query, args...)
	return reports, err
}

func (r *InventoryRepository) GetLowStockReport(ctx context.Context) ([]entity.StockReport, error) {
	var reports []entity.StockReport
	query := `
		SELECT 
			p.id as product_id,
			p.sku as product_sku,
			p.name as product_name,
			pc.name as category_name,
			w.name as warehouse_name,
			wl.code as location_code,
			COALESCE(SUM(istok.quantity), 0) as current_stock,
			COALESCE(SUM(istok.reserved_quantity), 0) as reserved_stock,
			COALESCE(SUM(istok.available_quantity), 0) as available_stock,
			p.min_stock,
			p.max_stock,
			p.unit_price,
			COALESCE(SUM(istok.available_quantity), 0) * p.unit_price as total_value
		FROM products p
		LEFT JOIN product_categories pc ON p.category_id = pc.id
		LEFT JOIN inventory_stocks istok ON p.id = istok.product_id
		LEFT JOIN warehouses w ON istok.warehouse_id = w.id
		LEFT JOIN warehouse_locations wl ON istok.location_id = wl.id
		WHERE p.is_active = true AND p.current_stock <= p.min_stock
		GROUP BY p.id, p.sku, p.name, pc.name, w.name, wl.code, p.min_stock, p.max_stock, p.unit_price
		ORDER BY p.current_stock ASC
	`
	err := r.db.SelectContext(ctx, &reports, query)
	return reports, err
}

func (r *InventoryRepository) GetMovementReport(ctx context.Context, filter dto.MovementReportFilter) ([]entity.StockMovement, int, error) {
	var movements []entity.StockMovement
	var total int

	countQuery := `SELECT COUNT(*) FROM stock_movements WHERE 1=1`
	query := `SELECT sm.*, p.sku as product_sku, p.name as product_name FROM stock_movements sm LEFT JOIN products p ON sm.product_id = p.id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.WarehouseID != nil {
		countQuery += fmt.Sprintf(` AND (from_warehouse_id = $%d OR to_warehouse_id = $%d)`, argIdx, argIdx)
		query += fmt.Sprintf(` AND (sm.from_warehouse_id = $%d OR sm.to_warehouse_id = $%d)`, argIdx, argIdx)
		args = append(args, *filter.WarehouseID)
		argIdx++
	}
	if filter.ProductID != nil {
		countQuery += fmt.Sprintf(` AND product_id = $%d`, argIdx)
		query += fmt.Sprintf(` AND sm.product_id = $%d`, argIdx)
		args = append(args, *filter.ProductID)
		argIdx++
	}
	if filter.LocationID != nil {
		countQuery += fmt.Sprintf(` AND (from_location_id = $%d OR to_location_id = $%d)`, argIdx, argIdx)
		query += fmt.Sprintf(` AND (sm.from_location_id = $%d OR sm.to_location_id = $%d)`, argIdx, argIdx)
		args = append(args, *filter.LocationID)
		argIdx++
	}
	if filter.MovementType != nil {
		countQuery += fmt.Sprintf(` AND movement_type = $%d`, argIdx)
		query += fmt.Sprintf(` AND sm.movement_type = $%d`, argIdx)
		args = append(args, *filter.MovementType)
		argIdx++
	}
	if filter.StartDate != nil {
		countQuery += fmt.Sprintf(` AND movement_date >= $%d`, argIdx)
		query += fmt.Sprintf(` AND sm.movement_date >= $%d`, argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		countQuery += fmt.Sprintf(` AND movement_date <= $%d`, argIdx)
		query += fmt.Sprintf(` AND sm.movement_date <= $%d`, argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return movements, 0, err
	}

	query += ` ORDER BY sm.movement_date DESC`
	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIdx)
		args = append(args, filter.Limit)
		argIdx++
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET $%d`, argIdx)
			args = append(args, filter.Offset)
		}
	}

	err = r.db.SelectContext(ctx, &movements, query, args...)
	return movements, total, err
}

func (r *InventoryRepository) GetInventoryValuation(ctx context.Context, warehouseID *uuid.UUID) ([]entity.InventoryValuation, error) {
	var valuations []entity.InventoryValuation
	var query string
	var args []interface{}

	if warehouseID != nil {
		query = `
			SELECT 
				w.id as warehouse_id,
				w.name as warehouse_name,
				COUNT(DISTINCT istok.product_id) as total_products,
				COALESCE(SUM(istok.quantity), 0) as total_quantity,
				COALESCE(SUM(CAST(istok.quantity AS numeric) * COALESCE(CAST(p.cost_price AS numeric), 0)), 0) as total_value,
				COALESCE(AVG(CAST(istok.quantity AS numeric) * COALESCE(CAST(p.cost_price AS numeric), 0)), 0) as average_value
			FROM warehouses w
			LEFT JOIN inventory_stocks istok ON w.id = istok.warehouse_id
			LEFT JOIN products p ON istok.product_id = p.id
			WHERE w.is_active = true AND w.id = $1
			GROUP BY w.id, w.name ORDER BY w.name
		`
		args = append(args, *warehouseID)
	} else {
		query = `
			SELECT 
				w.id as warehouse_id,
				w.name as warehouse_name,
				COUNT(DISTINCT istok.product_id) as total_products,
				COALESCE(SUM(istok.quantity), 0) as total_quantity,
				COALESCE(SUM(CAST(istok.quantity AS numeric) * COALESCE(CAST(p.cost_price AS numeric), 0)), 0) as total_value,
				COALESCE(AVG(CAST(istok.quantity AS numeric) * COALESCE(CAST(p.cost_price AS numeric), 0)), 0) as average_value
			FROM warehouses w
			LEFT JOIN inventory_stocks istok ON w.id = istok.warehouse_id
			LEFT JOIN products p ON istok.product_id = p.id
			WHERE w.is_active = true
			GROUP BY w.id, w.name ORDER BY w.name
		`
	}

	err := r.db.SelectContext(ctx, &valuations, query, args...)
	return valuations, err
}

func (r *InventoryRepository) GetAllWarehouses(ctx context.Context) ([]entity.Warehouse, error) {
	var warehouses []entity.Warehouse
	query := `SELECT * FROM warehouses WHERE is_active = true ORDER BY name`
	err := r.db.SelectContext(ctx, &warehouses, query)
	return warehouses, err
}

func (r *InventoryRepository) GetStockByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.InventoryStock, error) {
	var stocks []entity.InventoryStock
	query := `
		SELECT istok.*, p.sku as product_sku, p.name as product_name, w.code as warehouse_code, w.name as warehouse_name
		FROM inventory_stocks istok
		JOIN products p ON istok.product_id = p.id
		JOIN warehouses w ON istok.warehouse_id = w.id
		WHERE istok.warehouse_id = $1
		ORDER BY p.name
	`
	err := r.db.SelectContext(ctx, &stocks, query, warehouseID)
	return stocks, err
}

func (r *InventoryRepository) GetStockByWarehouseAndProduct(ctx context.Context, warehouseID, productID uuid.UUID) ([]entity.InventoryStock, error) {
	var stocks []entity.InventoryStock
	query := `
		SELECT istok.*, w.code as warehouse_code, w.name as warehouse_name
		FROM inventory_stocks istok
		JOIN warehouses w ON istok.warehouse_id = w.id
		WHERE istok.warehouse_id = $1 AND istok.product_id = $2
		ORDER BY istok.created_at
	`
	err := r.db.SelectContext(ctx, &stocks, query, warehouseID, productID)
	return stocks, err
}

func (r *InventoryRepository) GetOutOfStockCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products WHERE is_active = true AND current_stock = 0`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *InventoryRepository) GetPendingMovementsCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM stock_movements WHERE status = 'PENDING'`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *InventoryRepository) GetProductStockHistory(ctx context.Context, productID, warehouseID uuid.UUID, days int) ([]entity.StockMovement, error) {
	var movements []entity.StockMovement
	query := `
		SELECT sm.*, p.sku as product_sku, p.name as product_name
		FROM stock_movements sm
		LEFT JOIN products p ON sm.product_id = p.id
		WHERE sm.product_id = $1 
		AND (sm.from_warehouse_id = $2 OR sm.to_warehouse_id = $2)
		AND sm.movement_date >= NOW() - INTERVAL '` + fmt.Sprintf("%d days", days) + `'
		ORDER BY sm.movement_date DESC
	`
	err := r.db.SelectContext(ctx, &movements, query, productID, warehouseID)
	return movements, err
}
