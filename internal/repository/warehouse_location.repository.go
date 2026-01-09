// internal/repository/warehouse_location.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IWarehouseLocationRepository interface {
	Create(ctx context.Context, location *entity.WarehouseLocation) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.WarehouseLocation, error)
	FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error)
	FindActiveByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error)
	FindByCodeAndWarehouse(ctx context.Context, code string, warehouseID uuid.UUID) (*entity.WarehouseLocation, error)
	Update(ctx context.Context, id uuid.UUID, location *entity.WarehouseLocation) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateUsage(ctx context.Context, id uuid.UUID, quantity int) error
	IsCodeExists(ctx context.Context, code string, warehouseID uuid.UUID, excludeID uuid.UUID) (bool, error)
	GetAvailableLocations(ctx context.Context, warehouseID uuid.UUID, requiredCapacity int) ([]entity.WarehouseLocation, error)
}

type WarehouseLocationRepository struct {
	db *sqlx.DB
}

func NewWarehouseLocationRepository(db *sqlx.DB) IWarehouseLocationRepository {
	return &WarehouseLocationRepository{db: db}
}

func (r *WarehouseLocationRepository) Create(ctx context.Context, location *entity.WarehouseLocation) error {
	query := `
		INSERT INTO warehouse_locations (
			id, warehouse_id, code, name, description, zone, aisle, rack,
			shelf, bin, capacity, current_usage, is_active,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :warehouse_id, :code, :name, :description, :zone, :aisle, :rack,
			:shelf, :bin, :capacity, :current_usage, :is_active,
			:created_by, :updated_by, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, location)
	return err
}

func (r *WarehouseLocationRepository) FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error) {
	var locations []entity.WarehouseLocationWithWarehouse
	query := `
		SELECT wl.*,
			w.code as warehouse_code,
			w.name as warehouse_name
		FROM warehouse_locations wl
		JOIN warehouses w ON wl.warehouse_id = w.id
		WHERE wl.warehouse_id = $1 AND wl.is_active = true
		ORDER BY 
			COALESCE(wl.zone, ''),
			COALESCE(wl.aisle, ''),
			COALESCE(wl.rack, ''),
			COALESCE(wl.shelf, ''),
			COALESCE(wl.bin, ''),
			wl.code`

	err := r.db.SelectContext(ctx, &locations, query, warehouseID)
	if err != nil {
		return nil, err
	}

	// Convert to plain WarehouseLocation slice
	result := make([]entity.WarehouseLocation, len(locations))
	for i, loc := range locations {
		result[i] = *loc.WarehouseLocation
	}
	return result, nil
}

func (r *WarehouseLocationRepository) GetAvailableLocations(ctx context.Context, warehouseID uuid.UUID, requiredCapacity int) ([]entity.WarehouseLocation, error) {
	var locations []entity.WarehouseLocation
	query := `
		SELECT wl.*
		FROM warehouse_locations wl
		WHERE wl.warehouse_id = $1 
			AND wl.is_active = true
			AND (wl.capacity IS NULL OR (wl.capacity - wl.current_usage) >= $2)
		ORDER BY wl.current_usage ASC, wl.capacity DESC
		LIMIT 10`

	err := r.db.SelectContext(ctx, &locations, query, warehouseID, requiredCapacity)
	return locations, err
}

func (r *WarehouseLocationRepository) UpdateUsage(ctx context.Context, id uuid.UUID, quantity int) error {
	query := `
		UPDATE warehouse_locations 
		SET current_usage = current_usage + $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, quantity, id)
	return err
}

func (r *WarehouseLocationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.WarehouseLocation, error) {
	var location entity.WarehouseLocation
	query := `SELECT * FROM warehouse_locations WHERE id = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &location, query, id)
	if err != nil {
		return nil, err
	}
	return &location, nil
}

func (r *WarehouseLocationRepository) FindActiveByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error) {
	return r.FindByWarehouseID(ctx, warehouseID)
}

func (r *WarehouseLocationRepository) FindByCodeAndWarehouse(ctx context.Context, code string, warehouseID uuid.UUID) (*entity.WarehouseLocation, error) {
	var location entity.WarehouseLocation
	query := `SELECT * FROM warehouse_locations WHERE code = $1 AND warehouse_id = $2 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &location, query, code, warehouseID)
	return &location, err
}

func (r *WarehouseLocationRepository) Update(ctx context.Context, id uuid.UUID, location *entity.WarehouseLocation) error {
	query := `
		UPDATE warehouse_locations
		SET code = :code, name = :name, description = :description,
			zone = :zone, aisle = :aisle, rack = :rack, shelf = :shelf, bin = :bin,
			capacity = :capacity, is_active = :is_active,
			updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	location.ID = id
	_, err := r.db.NamedExecContext(ctx, query, location)
	return err
}

func (r *WarehouseLocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE warehouse_locations SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *WarehouseLocationRepository) IsCodeExists(ctx context.Context, code string, warehouseID uuid.UUID, excludeID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM warehouse_locations WHERE code = $1 AND warehouse_id = $2`
	args := []interface{}{code, warehouseID}
	if excludeID != uuid.Nil {
		query += ` AND id != $3`
		args = append(args, excludeID)
	}
	err := r.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}
