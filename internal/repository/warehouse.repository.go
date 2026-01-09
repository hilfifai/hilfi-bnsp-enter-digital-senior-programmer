// internal/repository/warehouse.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IWarehouseRepository interface {
	Create(ctx context.Context, warehouse *entity.Warehouse) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Warehouse, error)
	FindByCode(ctx context.Context, code string) (*entity.Warehouse, error)
	FindAll(ctx context.Context) ([]entity.Warehouse, error)
	FindActive(ctx context.Context) ([]entity.Warehouse, error)
	Update(ctx context.Context, id uuid.UUID, warehouse *entity.Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
	GetWarehouseStats(ctx context.Context, id uuid.UUID) (*entity.WarehouseStats, error)
}

type WarehouseRepository struct {
	db *sqlx.DB
}

func NewWarehouseRepository(db *sqlx.DB) IWarehouseRepository {
	return &WarehouseRepository{db: db}
}

func (r *WarehouseRepository) Create(ctx context.Context, warehouse *entity.Warehouse) error {
	query := `
		INSERT INTO warehouses (
			id, code, name, address, contact_person, phone, is_active,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :code, :name, :address, :contact_person, :phone, :is_active,
			:created_by, :updated_by, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, warehouse)
	return err
}

func (r *WarehouseRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Warehouse, error) {
	var warehouse entity.WarehouseWithStats
	query := `
		SELECT w.*,
			COALESCE((SELECT COUNT(*) FROM warehouse_locations wl WHERE wl.warehouse_id = w.id), 0) as location_count,
			COALESCE((SELECT SUM(wl.capacity) FROM warehouse_locations wl WHERE wl.warehouse_id = w.id), 0) as total_capacity,
			COALESCE((SELECT SUM(wl.current_usage) FROM warehouse_locations wl WHERE wl.warehouse_id = w.id), 0) as used_capacity
		FROM warehouses w
		WHERE w.id = $1
		LIMIT 1`

	err := r.db.GetContext(ctx, &warehouse, query, id)
	if err != nil {
		return nil, err
	}
	return warehouse.Warehouse, nil
}

func (r *WarehouseRepository) FindAll(ctx context.Context) ([]entity.Warehouse, error) {
	var warehouses []entity.WarehouseWithStats
	query := `
		SELECT w.*,
			COALESCE((SELECT COUNT(*) FROM warehouse_locations wl WHERE wl.warehouse_id = w.id), 0) as location_count,
			COALESCE((SELECT SUM(wl.capacity) FROM warehouse_locations wl WHERE wl.warehouse_id = w.id), 0) as total_capacity,
			COALESCE((SELECT SUM(wl.current_usage) FROM warehouse_locations wl WHERE wl.warehouse_id = w.id), 0) as used_capacity
		FROM warehouses w
		WHERE w.is_active = true
		ORDER BY w.code, w.name`

	err := r.db.SelectContext(ctx, &warehouses, query)
	if err != nil {
		return nil, err
	}

	// Convert to plain Warehouse slice
	result := make([]entity.Warehouse, len(warehouses))
	for i, w := range warehouses {
		result[i] = *w.Warehouse
	}
	return result, nil
}

func (r *WarehouseRepository) Update(ctx context.Context, id uuid.UUID, warehouse *entity.Warehouse) error {
	query := `
		UPDATE warehouses 
		SET code = :code,
			name = :name,
			address = :address,
			contact_person = :contact_person,
			phone = :phone,
			is_active = :is_active,
			updated_by = :updated_by,
			updated_at = :updated_at
		WHERE id = :id`

	warehouse.ID = id
	_, err := r.db.NamedExecContext(ctx, query, warehouse)
	return err
}

func (r *WarehouseRepository) IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM warehouses WHERE code = $1"
	args := []interface{}{code}

	if excludeID != uuid.Nil {
		query += " AND id != $2"
		args = append(args, excludeID)
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *WarehouseRepository) FindByCode(ctx context.Context, code string) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	query := `SELECT * FROM warehouses WHERE code = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &warehouse, query, code)
	return &warehouse, err
}

func (r *WarehouseRepository) FindActive(ctx context.Context) ([]entity.Warehouse, error) {
	return r.FindAll(ctx)
}

func (r *WarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE warehouses SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *WarehouseRepository) GetWarehouseStats(ctx context.Context, id uuid.UUID) (*entity.WarehouseStats, error) {
	var stats entity.WarehouseStats

	query := `
		SELECT 
			w.id,
			w.name,
			COALESCE(COUNT(DISTINCT wl.id), 0) as total_locations,
			COALESCE(SUM(wl.capacity), 0) as total_capacity,
			COALESCE(SUM(wl.current_usage), 0) as used_capacity,
			COALESCE(COUNT(DISTINCT inv.product_id), 0) as total_products,
			COALESCE(SUM(inv.quantity), 0) as total_quantity,
			COALESCE(SUM(inv.reserved_quantity), 0) as total_reserved
		FROM warehouses w
		LEFT JOIN warehouse_locations wl ON w.id = wl.warehouse_id AND wl.is_active = true
		LEFT JOIN inventory_stocks inv ON w.id = inv.warehouse_id
		WHERE w.id = $1
		GROUP BY w.id, w.name`

	err := r.db.GetContext(ctx, &stats, query, id)
	return &stats, err
}
