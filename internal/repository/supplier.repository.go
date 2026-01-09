// internal/repository/supplier.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ISupplierRepository interface {
	Create(ctx context.Context, supplier *entity.Supplier) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error)
	FindByCode(ctx context.Context, code string) (*entity.Supplier, error)
	FindAll(ctx context.Context, activeOnly bool) ([]entity.Supplier, error)
	Update(ctx context.Context, id uuid.UUID, supplier *entity.Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
}

type SupplierRepository struct {
	db *sqlx.DB
}

func NewSupplierRepository(db *sqlx.DB) ISupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) Create(ctx context.Context, supplier *entity.Supplier) error {
	query := `
		INSERT INTO suppliers (
			id, code, name, contact_person, email, phone, address, tax_number,
			payment_terms, is_active, created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :code, :name, :contact_person, :email, :phone, :address, :tax_number,
			:payment_terms, :is_active, :created_by, :updated_by, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, query, supplier)
	return err
}

func (r *SupplierRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error) {
	var supplier entity.Supplier
	query := `SELECT * FROM suppliers WHERE id = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &supplier, query, id)
	return &supplier, err
}

func (r *SupplierRepository) FindByCode(ctx context.Context, code string) (*entity.Supplier, error) {
	var supplier entity.Supplier
	query := `SELECT * FROM suppliers WHERE code = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &supplier, query, code)
	return &supplier, err
}

func (r *SupplierRepository) FindAll(ctx context.Context, activeOnly bool) ([]entity.Supplier, error) {
	var suppliers []entity.Supplier
	query := `SELECT * FROM suppliers`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name`
	err := r.db.SelectContext(ctx, &suppliers, query)
	return suppliers, err
}

func (r *SupplierRepository) Update(ctx context.Context, id uuid.UUID, supplier *entity.Supplier) error {
	query := `
		UPDATE suppliers
		SET name = :name, contact_person = :contact_person, email = :email,
			phone = :phone, address = :address, tax_number = :tax_number,
			payment_terms = :payment_terms, is_active = :is_active,
			updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	supplier.ID = id
	_, err := r.db.NamedExecContext(ctx, query, supplier)
	return err
}

func (r *SupplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE suppliers SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SupplierRepository) IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM suppliers WHERE code = $1`
	args := []interface{}{code}
	if excludeID != uuid.Nil {
		query += ` AND id != $2`
		args = append(args, excludeID)
	}
	err := r.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}
