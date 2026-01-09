// internal/repository/customer.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ICustomerRepository interface {
	Create(ctx context.Context, customer *entity.Customer) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
	FindByCode(ctx context.Context, code string) (*entity.Customer, error)
	FindAll(ctx context.Context, activeOnly bool) ([]entity.Customer, error)
	Update(ctx context.Context, id uuid.UUID, customer *entity.Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
}

type CustomerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) ICustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	query := `
		INSERT INTO customers (
			id, code, name, contact_person, email, phone, address, tax_number,
			credit_limit, payment_terms, is_active, created_by, updated_by,
			created_at, updated_at
		) VALUES (
			:id, :code, :name, :contact_person, :email, :phone, :address, :tax_number,
			:credit_limit, :payment_terms, :is_active, :created_by, :updated_by,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, customer)
	return err
}

func (r *CustomerRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error) {
	var customer entity.Customer
	query := `SELECT * FROM customers WHERE id = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &customer, query, id)
	return &customer, err
}

func (r *CustomerRepository) FindByCode(ctx context.Context, code string) (*entity.Customer, error) {
	var customer entity.Customer
	query := `SELECT * FROM customers WHERE code = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &customer, query, code)
	return &customer, err
}

func (r *CustomerRepository) FindAll(ctx context.Context, activeOnly bool) ([]entity.Customer, error) {
	var customers []entity.Customer
	query := `SELECT * FROM customers`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name`
	err := r.db.SelectContext(ctx, &customers, query)
	return customers, err
}

func (r *CustomerRepository) Update(ctx context.Context, id uuid.UUID, customer *entity.Customer) error {
	query := `
		UPDATE customers
		SET name = :name, contact_person = :contact_person, email = :email,
			phone = :phone, address = :address, tax_number = :tax_number,
			credit_limit = :credit_limit, payment_terms = :payment_terms,
			is_active = :is_active, updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	customer.ID = id
	_, err := r.db.NamedExecContext(ctx, query, customer)
	return err
}

func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE customers SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *CustomerRepository) IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM customers WHERE code = $1`
	args := []interface{}{code}
	if excludeID != uuid.Nil {
		query += ` AND id != $2`
		args = append(args, excludeID)
	}
	err := r.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}
