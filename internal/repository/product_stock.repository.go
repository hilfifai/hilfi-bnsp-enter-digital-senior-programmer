package repository

import (
	"api-jellyfish/internal/entity"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IProductStockRepository interface {
	GetStock(ctx context.Context, productID, warehouseID uuid.UUID) (*entity.ProductStock, error)
	UpdateStock(ctx context.Context, stock *entity.ProductStock) error
	GetStockMovementHistory(ctx context.Context, filter map[string]interface{}) ([]entity.StockMovement, error)
}

type ProductStockRepository struct {
	db *sqlx.DB
}

func NewProductStockRepository(db *sqlx.DB) *ProductStockRepository {
	return &ProductStockRepository{db: db}
}

func (r *ProductStockRepository) GetStock(ctx context.Context, productID, warehouseID uuid.UUID) (*entity.ProductStock, error) {
	var stock entity.ProductStock
	query := `
		SELECT * FROM product_stocks 
		WHERE product_id = $1 AND warehouse_id = $2
	`
	err := r.db.GetContext(ctx, &stock, query, productID, warehouseID)
	return &stock, err
}

func (r *ProductStockRepository) UpdateStock(ctx context.Context, stock *entity.ProductStock) error {
	query := `
		INSERT INTO product_stocks (
			id, product_id, warehouse_id, quantity, reserved, available,
			reorder_level, last_updated
		) VALUES (
			:id, :product_id, :warehouse_id, :quantity, :reserved, :available,
			:reorder_level, :last_updated
		)
		ON CONFLICT (product_id, warehouse_id) 
		DO UPDATE SET
			quantity = EXCLUDED.quantity,
			reserved = EXCLUDED.reserved,
			available = EXCLUDED.available,
			last_updated = EXCLUDED.last_updated
	`
	_, err := r.db.NamedExecContext(ctx, query, stock)
	return err
}

func (r *ProductStockRepository) GetStockMovementHistory(ctx context.Context, filter map[string]interface{}) ([]entity.StockMovement, error) {
	var movements []entity.StockMovement
	query := `
		SELECT sm.*, p.name as product_name, w.name as warehouse_name
		FROM stock_movements sm
		JOIN products p ON sm.product_id = p.id
		JOIN warehouses w ON sm.warehouse_id = w.id
		WHERE 1=1
	`

	// Add filters dynamically
	args := []interface{}{}
	argIndex := 1

	if productID, ok := filter["product_id"]; ok {
		query += fmt.Sprintf(" AND sm.product_id = $%d", argIndex)
		args = append(args, productID)
		argIndex++
	}

	if warehouseID, ok := filter["warehouse_id"]; ok {
		query += fmt.Sprintf(" AND sm.warehouse_id = $%d", argIndex)
		args = append(args, warehouseID)
		argIndex++
	}

	if movementType, ok := filter["movement_type"]; ok {
		query += fmt.Sprintf(" AND sm.movement_type = $%d", argIndex)
		args = append(args, movementType)
		argIndex++
	}

	if startDate, ok := filter["start_date"]; ok {
		query += fmt.Sprintf(" AND sm.created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}

	if endDate, ok := filter["end_date"]; ok {
		query += fmt.Sprintf(" AND sm.created_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	query += " ORDER BY sm.created_at DESC"

	err := r.db.SelectContext(ctx, &movements, query, args...)
	return movements, err
}
