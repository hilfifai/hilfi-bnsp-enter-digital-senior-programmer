package repository

import (
	"api-jellyfish/internal/entity"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	FindBySKU(ctx context.Context, sku string) (*entity.Product, error)
	FindAll(ctx context.Context, filter entity.ProductFilter) ([]entity.Product, error)
	FindActive(ctx context.Context) ([]entity.Product, error)
	Update(ctx context.Context, id uuid.UUID, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
	IsSKUExists(ctx context.Context, sku string, excludeID uuid.UUID) (bool, error)
	GetWithStockInfo(ctx context.Context, id uuid.UUID) (*entity.ProductWithStock, error)
	GetCount(ctx context.Context) (int, error)
	GetLowStockCount(ctx context.Context) (int, error)
}

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) IProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *entity.Product) error {
	query := `
		INSERT INTO products (
			id, sku, name, description, category_id, unit_price, cost_price,
			current_stock, min_stock, max_stock, weight, dimensions, is_active,
			attributes, image_urls, barcode, created_by, updated_by,
			created_at, updated_at
		) VALUES (
			:id, :sku, :name, :description, :category_id, :unit_price, :cost_price,
			:current_stock, :min_stock, :max_stock, :weight, :dimensions, :is_active,
			:attributes, :image_urls, :barcode, :created_by, :updated_by,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var product entity.Product
	query := `
		SELECT p.*, pc.name as category_name
		FROM products p
		LEFT JOIN product_categories pc ON p.category_id = pc.id
		WHERE p.id = $1 AND p.is_active = true
		LIMIT 1`

	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	var product entity.Product
	query := `
		SELECT p.*, pc.name as category_name
		FROM products p
		LEFT JOIN product_categories pc ON p.category_id = pc.id
		WHERE p.sku = $1 AND p.is_active = true
		LIMIT 1`

	err := r.db.GetContext(ctx, &product, query, sku)
	if err != nil {
		return nil, fmt.Errorf("product with SKU %s not found: %w", sku, err)
	}
	return &product, nil
}

func (r *ProductRepository) FindAll(ctx context.Context, filter entity.ProductFilter) ([]entity.Product, error) {
	var products []entity.Product
	query := `
		SELECT p.*, pc.name as category_name
		FROM products p
		LEFT JOIN product_categories pc ON p.category_id = pc.id
		WHERE 1=1
	`

	var args []interface{}
	argCounter := 1

	if filter.CategoryID != nil {
		query += fmt.Sprintf(" AND p.category_id = $%d", argCounter)
		args = append(args, *filter.CategoryID)
		argCounter++
	}

	if filter.SKU != nil && *filter.SKU != "" {
		query += fmt.Sprintf(" AND p.sku ILIKE $%d", argCounter)
		args = append(args, "%"+*filter.SKU+"%")
		argCounter++
	}

	if filter.Name != nil && *filter.Name != "" {
		query += fmt.Sprintf(" AND p.name ILIKE $%d", argCounter)
		args = append(args, "%"+*filter.Name+"%")
		argCounter++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND p.is_active = $%d", argCounter)
		args = append(args, *filter.IsActive)
		argCounter++
	}

	if filter.LowStockOnly != nil && *filter.LowStockOnly {
		query += " AND p.current_stock <= p.min_stock"
	}

	// Sorting
	if filter.SortBy != "" {
		sortField := strings.ToLower(filter.SortBy)
		sortOrder := strings.ToUpper(filter.SortOrder)
		if sortOrder != "ASC" && sortOrder != "DESC" {
			sortOrder = "ASC"
		}

		switch sortField {
		case "name", "sku", "created_at", "updated_at", "current_stock", "unit_price":
			query += fmt.Sprintf(" ORDER BY p.%s %s", sortField, sortOrder)
		default:
			query += " ORDER BY p.created_at DESC"
		}
	} else {
		query += " ORDER BY p.created_at DESC"
	}

	// Pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCounter)
		args = append(args, filter.Limit)
		argCounter++

		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argCounter)
			args = append(args, filter.Offset)
		}
	}

	err := r.db.SelectContext(ctx, &products, query, args...)
	return products, err
}

func (r *ProductRepository) Update(ctx context.Context, id uuid.UUID, product *entity.Product) error {
	query := `
		UPDATE products 
		SET name = :name,
			description = :description,
			category_id = :category_id,
			unit_price = :unit_price,
			cost_price = :cost_price,
			min_stock = :min_stock,
			max_stock = :max_stock,
			weight = :weight,
			dimensions = :dimensions,
			is_active = :is_active,
			attributes = :attributes,
			image_urls = :image_urls,
			barcode = :barcode,
			updated_by = :updated_by,
			updated_at = :updated_at
		WHERE id = :id`

	product.ID = id
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

func (r *ProductRepository) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	query := `
		UPDATE products 
		SET current_stock = current_stock + $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, quantity, productID)
	return err
}

func (r *ProductRepository) IsSKUExists(ctx context.Context, sku string, excludeID uuid.UUID) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM products WHERE sku = $1"
	args := []interface{}{sku}

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

func (r *ProductRepository) GetWithStockInfo(ctx context.Context, id uuid.UUID) (*entity.ProductWithStock, error) {
	var productWithStock entity.ProductWithStock

	// Get product info
	product, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	productWithStock.Product = *product

	// Get stock info from inventory_stocks
	query := `
		SELECT 
			is.id, is.product_id, is.warehouse_id, is.location_id,
			is.quantity, is.reserved_quantity, is.available_quantity,
			is.batch_number, is.expiry_date,
			w.code as warehouse_code, w.name as warehouse_name,
			wl.code as location_code, wl.name as location_name
		FROM inventory_stocks is
		LEFT JOIN warehouses w ON is.warehouse_id = w.id
		LEFT JOIN warehouse_locations wl ON is.location_id = wl.id
		WHERE is.product_id = $1
		ORDER BY w.name, wl.name`

	var stocks []entity.InventoryStock
	err = r.db.SelectContext(ctx, &stocks, query, id)
	if err != nil {
		return nil, err
	}
	productWithStock.WarehouseStocks = stocks

	// Calculate totals
	totalAvailable := 0
	totalReserved := 0
	for _, stock := range stocks {
		totalAvailable += stock.AvailableQuantity
		totalReserved += stock.ReservedQuantity
	}
	productWithStock.TotalAvailable = totalAvailable
	productWithStock.TotalReserved = totalReserved

	return &productWithStock, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ProductRepository) FindActive(ctx context.Context) ([]entity.Product, error) {
	var products []entity.Product
	query := `
		SELECT p.*, pc.name as category_name
		FROM products p
		LEFT JOIN product_categories pc ON p.category_id = pc.id
		WHERE p.is_active = true
		ORDER BY p.name`

	err := r.db.SelectContext(ctx, &products, query)
	return products, err
}

func (r *ProductRepository) GetCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products WHERE is_active = true`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *ProductRepository) GetLowStockCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products WHERE is_active = true AND current_stock <= min_stock`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}
