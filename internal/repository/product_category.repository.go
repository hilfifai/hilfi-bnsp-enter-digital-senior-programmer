// internal/repository/product_category.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IProductCategoryRepository interface {
	Create(ctx context.Context, category *entity.ProductCategory) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ProductCategory, error)
	FindByCode(ctx context.Context, code string) (*entity.ProductCategory, error)
	FindAll(ctx context.Context, activeOnly bool) ([]entity.ProductCategory, error)
	Update(ctx context.Context, id uuid.UUID, category *entity.ProductCategory) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
}

type ProductCategoryRepository struct {
	db *sqlx.DB
}

func NewProductCategoryRepository(db *sqlx.DB) IProductCategoryRepository {
	return &ProductCategoryRepository{db: db}
}

func (r *ProductCategoryRepository) Create(ctx context.Context, category *entity.ProductCategory) error {
	query := `
		INSERT INTO product_categories (
			id, code, name, description, parent_category_id, is_active,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :code, :name, :description, :parent_category_id, :is_active,
			:created_by, :updated_by, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, query, category)
	return err
}

func (r *ProductCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ProductCategory, error) {
	var category entity.ProductCategory
	query := `SELECT * FROM product_categories WHERE id = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) FindByCode(ctx context.Context, code string) (*entity.ProductCategory, error) {
	var category entity.ProductCategory
	query := `SELECT * FROM product_categories WHERE code = $1 AND is_active = true LIMIT 1`
	err := r.db.GetContext(ctx, &category, query, code)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) FindAll(ctx context.Context, activeOnly bool) ([]entity.ProductCategory, error) {
	var categories []entity.ProductCategory
	query := `SELECT * FROM product_categories`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name`
	err := r.db.SelectContext(ctx, &categories, query)
	return categories, err
}

func (r *ProductCategoryRepository) Update(ctx context.Context, id uuid.UUID, category *entity.ProductCategory) error {
	query := `
		UPDATE product_categories
		SET name = :name, description = :description, parent_category_id = :parent_category_id,
			is_active = :is_active, updated_by = :updated_by, updated_at = :updated_at
		WHERE id = :id`
	category.ID = id
	_, err := r.db.NamedExecContext(ctx, query, category)
	return err
}

func (r *ProductCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE product_categories SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ProductCategoryRepository) IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM product_categories WHERE code = $1`
	args := []interface{}{code}
	if excludeID != uuid.Nil {
		query += ` AND id != $2`
		args = append(args, excludeID)
	}
	err := r.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}
