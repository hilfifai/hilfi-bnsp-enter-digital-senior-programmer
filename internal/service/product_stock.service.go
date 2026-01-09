// internal/service/product_stock.service.go
package service

import (
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type IProductStockService interface {
	GetStock(ctx context.Context, productID, warehouseID uuid.UUID) (*entity.ProductStock, error)
	UpdateStock(ctx context.Context, stock *entity.ProductStock) error
	GetStockMovementHistory(ctx context.Context, filter map[string]interface{}) ([]entity.StockMovement, error)
	GetLowStockProducts(ctx context.Context) ([]entity.ProductStock, error)
	GetStockByProduct(ctx context.Context, productID uuid.UUID) ([]entity.ProductStock, error)
	GetStockByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.ProductStock, error)
}

type ProductStockService struct {
	stockRepo     repository.IProductStockRepository
	productRepo   repository.IProductRepository
	warehouseRepo repository.IWarehouseRepository
	inventoryRepo repository.IInventoryRepository
}

func NewProductStockService(
	stockRepo repository.IProductStockRepository,
	productRepo repository.IProductRepository,
	warehouseRepo repository.IWarehouseRepository,
	inventoryRepo repository.IInventoryRepository,
) IProductStockService {
	return &ProductStockService{
		stockRepo:     stockRepo,
		productRepo:   productRepo,
		warehouseRepo: warehouseRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *ProductStockService) GetStock(ctx context.Context, productID, warehouseID uuid.UUID) (*entity.ProductStock, error) {
	stock, err := s.stockRepo.GetStock(ctx, productID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}
	return stock, nil
}

func (s *ProductStockService) UpdateStock(ctx context.Context, stock *entity.ProductStock) error {
	// Validate product exists
	product, err := s.productRepo.FindByID(ctx, stock.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Validate warehouse exists
	warehouse, err := s.warehouseRepo.FindByID(ctx, stock.WarehouseID)
	if err != nil {
		return fmt.Errorf("warehouse not found: %w", err)
	}

	// Set available quantity
	stock.Available = stock.Quantity - stock.Reserved
	stock.LastUpdated = time.Now()

	if err := s.stockRepo.UpdateStock(ctx, stock); err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	// Update product total stock
	_ = product
	_ = warehouse

	return nil
}

func (s *ProductStockService) GetStockMovementHistory(ctx context.Context, filter map[string]interface{}) ([]entity.StockMovement, error) {
	movements, err := s.stockRepo.GetStockMovementHistory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock movement history: %w", err)
	}
	return movements, nil
}

func (s *ProductStockService) GetLowStockProducts(ctx context.Context) ([]entity.ProductStock, error) {
	filter := map[string]interface{}{
		"low_stock": true,
	}
	movements, err := s.stockRepo.GetStockMovementHistory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	// Convert movements to stocks (simplified)
	var stocks []entity.ProductStock
	for _, m := range movements {
		stock := entity.ProductStock{
			ProductID:   m.ProductID,
			WarehouseID: *m.ToWarehouseID,
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (s *ProductStockService) GetStockByProduct(ctx context.Context, productID uuid.UUID) ([]entity.ProductStock, error) {
	filter := map[string]interface{}{
		"product_id": productID,
	}
	movements, err := s.stockRepo.GetStockMovementHistory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock by product: %w", err)
	}

	var stocks []entity.ProductStock
	for _, m := range movements {
		stock := entity.ProductStock{
			ProductID:   m.ProductID,
			WarehouseID: *m.ToWarehouseID,
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (s *ProductStockService) GetStockByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.ProductStock, error) {
	filter := map[string]interface{}{
		"warehouse_id": warehouseID,
	}
	movements, err := s.stockRepo.GetStockMovementHistory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock by warehouse: %w", err)
	}

	var stocks []entity.ProductStock
	for _, m := range movements {
		stock := entity.ProductStock{
			ProductID:   m.ProductID,
			WarehouseID: *m.ToWarehouseID,
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}
