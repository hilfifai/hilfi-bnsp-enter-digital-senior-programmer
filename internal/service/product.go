package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	errpkg "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type IProductService interface {
	CreateProduct(ctx context.Context, req dto.CreateProductRequest, userID uuid.UUID) (*entity.Product, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*entity.ProductResponse, error)
	GetProducts(ctx context.Context, filter dto.ProductFilterDTO) ([]entity.ProductResponse, int, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest, userID uuid.UUID) (*entity.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	UpdateStock(ctx context.Context, req dto.BulkUpdateStockRequest, userID uuid.UUID) error
	GetLowStockProducts(ctx context.Context) ([]entity.ProductResponse, error)
	SearchProducts(ctx context.Context, query string, limit int) ([]entity.ProductResponse, error)
}

type ProductService struct {
	productRepo   repository.IProductRepository
	categoryRepo  repository.IProductCategoryRepository
	inventoryRepo repository.IInventoryRepository
}

func NewProductService(
	productRepo repository.IProductRepository,
	categoryRepo repository.IProductCategoryRepository,
	inventoryRepo repository.IInventoryRepository,
) IProductService {
	return &ProductService{
		productRepo:   productRepo,
		categoryRepo:  categoryRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *ProductService) validateProductData(ctx context.Context, req dto.CreateProductRequest, excludeID uuid.UUID) error {
	// Validate SKU uniqueness
	exists, err := s.productRepo.IsSKUExists(ctx, req.SKU, excludeID)
	if err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrProductValidation, err)
	}
	if exists {
		return errpkg.ErrProductSKUExists
	}

	// Validate category exists
	category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil || category == nil {
		return errpkg.ErrProductCategoryNotFound
	}

	return s.validatePriceAndStock(req.UnitPrice, req.CostPrice, req.MinStock, req.MaxStock)
}

func (s *ProductService) validateProductDataUpdate(ctx context.Context, req dto.UpdateProductRequest, productID uuid.UUID) error {
	return s.validatePriceAndStock(req.UnitPrice, req.CostPrice, req.MinStock, req.MaxStock)
}

func (s *ProductService) validatePriceAndStock(unitPrice, costPrice decimal.Decimal, minStock, maxStock int) error {
	// Validate price
	if unitPrice.LessThan(decimal.Zero) {
		return errpkg.ErrInvalidPrice
	}
	if costPrice.LessThan(decimal.Zero) {
		return errpkg.ErrInvalidPrice
	}

	// Validate stock levels
	if minStock < 0 || maxStock < 0 {
		return errpkg.ErrInvalidStock
	}
	if minStock > maxStock {
		return errpkg.ErrInvalidStockRange
	}

	return nil
}

func (s *ProductService) CreateProduct(ctx context.Context, req dto.CreateProductRequest, userID uuid.UUID) (*entity.Product, error) {
	// Validate input
	if err := s.validateProductData(ctx, req, uuid.Nil); err != nil {
		return nil, err
	}

	// Convert attributes
	attributes := make(entity.ProductAttributes, len(req.Attributes))
	for i, attr := range req.Attributes {
		attributes[i] = entity.ProductAttribute{
			Key:   attr.Key,
			Value: attr.Value,
			Label: attr.Label,
		}
	}

	// Create product entity
	product := &entity.Product{
		SKU:          req.SKU,
		Name:         req.Name,
		Description:  req.Description,
		CategoryID:   req.CategoryID,
		UnitPrice:    req.UnitPrice,
		CostPrice:    req.CostPrice,
		CurrentStock: 0,
		MinStock:     req.MinStock,
		MaxStock:     req.MaxStock,
		Weight:       req.Weight,
		Dimensions:   req.Dimensions,
		IsActive:     req.IsActive,
		Attributes:   attributes,
		ImageURLs:    req.ImageURLs,
		Barcode:      req.Barcode,
		CreatedBy:    userID,
		UpdatedBy:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Generate UUID
	product.ID = uuid.New()

	// Save to database
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrCreateProduct, err)
	}

	// Create audit log
	s.auditProductCreation(ctx, product, userID)

	return product, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*entity.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrProductNotFound, err)
	}

	// Get stock information
	stockInfo, err := s.getProductStockInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &entity.ProductResponse{
		Product: *product,
		StockInfo: struct {
			TotalAvailable int `json:"total_available"`
			TotalReserved  int `json:"total_reserved"`
			TotalOnHand    int `json:"total_on_hand"`
		}{
			TotalAvailable: stockInfo.TotalAvailable,
			TotalReserved:  stockInfo.TotalReserved,
			TotalOnHand:    product.CurrentStock,
		},
	}

	return response, nil
}

func (s *ProductService) getProductStockInfo(ctx context.Context, productID uuid.UUID) (*dto.ProductStockInfoDTO, error) {
	stockInfo := &dto.ProductStockInfoDTO{}

	// Get warehouse stock details
	stocks, err := s.inventoryRepo.GetProductStockByProductID(ctx, productID)
	if err != nil {
		return stockInfo, nil // Return empty if error
	}

	// Calculate totals
	for _, stock := range stocks {
		stockInfo.TotalAvailable += stock.AvailableQuantity
		stockInfo.TotalReserved += stock.ReservedQuantity
	}

	// Get product for current stock
	product, err := s.productRepo.FindByID(ctx, productID)
	if err == nil {
		stockInfo.TotalOnHand = product.CurrentStock
		stockInfo.LowStockAlert = product.CurrentStock <= product.MinStock
	}

	return stockInfo, nil
}

func (s *ProductService) GetProducts(ctx context.Context, filter dto.ProductFilterDTO) ([]entity.ProductResponse, int, error) {
	// Convert DTO filter to entity filter
	entityFilter := entity.ProductFilter{
		CategoryID:   filter.CategoryID,
		SKU:          filter.SKU,
		Name:         filter.Name,
		Barcode:      filter.Barcode,
		IsActive:     filter.IsActive,
		LowStockOnly: filter.LowStockOnly,
		Page:         filter.Page,
		Limit:        filter.Limit,
		Offset:       (filter.Page - 1) * filter.Limit,
		SortBy:       filter.SortBy,
		SortOrder:    filter.SortOrder,
	}

	// Get products from repository
	products, err := s.productRepo.FindAll(ctx, entityFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", errpkg.ErrGetProducts, err)
	}

	// Get total count for pagination
	count, err := s.productRepo.GetCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", errpkg.ErrGetProducts, err)
	}

	// Convert to response DTO
	responses := make([]entity.ProductResponse, len(products))
	for i, product := range products {
		stockInfo, _ := s.getProductStockInfo(ctx, product.ID)

		responses[i] = entity.ProductResponse{
			Product: product,
			StockInfo: struct {
				TotalAvailable int `json:"total_available"`
				TotalReserved  int `json:"total_reserved"`
				TotalOnHand    int `json:"total_on_hand"`
			}{
				TotalAvailable: stockInfo.TotalAvailable,
				TotalReserved:  stockInfo.TotalReserved,
				TotalOnHand:    product.CurrentStock,
			},
		}
	}

	return responses, count, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest, userID uuid.UUID) (*entity.Product, error) {
	// Check if product exists
	existing, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrProductNotFound, err)
	}

	// Validate input
	if err := s.validateProductDataUpdate(ctx, req, id); err != nil {
		return nil, err
	}

	// Validate category exists if changed
	if existing.CategoryID != req.CategoryID {
		category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
		if err != nil || category == nil {
			return nil, errpkg.ErrProductCategoryNotFound
		}
	}

	// Convert attributes
	attributes := make(entity.ProductAttributes, len(req.Attributes))
	for i, attr := range req.Attributes {
		attributes[i] = entity.ProductAttribute{
			Key:   attr.Key,
			Value: attr.Value,
			Label: attr.Label,
		}
	}

	// Update product entity
	product := &entity.Product{
		ID:           id,
		Name:         req.Name,
		Description:  req.Description,
		CategoryID:   req.CategoryID,
		UnitPrice:    req.UnitPrice,
		CostPrice:    req.CostPrice,
		MinStock:     req.MinStock,
		MaxStock:     req.MaxStock,
		Weight:       req.Weight,
		Dimensions:   req.Dimensions,
		IsActive:     req.IsActive,
		Attributes:   attributes,
		ImageURLs:    req.ImageURLs,
		Barcode:      req.Barcode,
		UpdatedBy:    userID,
		UpdatedAt:    time.Now(),
		CreatedBy:    existing.CreatedBy,
		CreatedAt:    existing.CreatedAt,
		CurrentStock: existing.CurrentStock,
		SKU:          existing.SKU, // SKU cannot be changed
	}

	// Save to database
	if err := s.productRepo.Update(ctx, id, product); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrUpdateProduct, err)
	}

	// Create audit log
	s.auditProductUpdate(ctx, existing, product, userID)

	return product, nil
}

func (s *ProductService) UpdateStock(ctx context.Context, req dto.BulkUpdateStockRequest, userID uuid.UUID) error {
	// Validate product exists
	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrProductNotFound, err)
	}

	// Validate warehouse exists
	_, err = s.inventoryRepo.GetWarehouseByID(ctx, req.WarehouseID)
	if err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
	}

	// Get current stock in warehouse
	currentStock, err := s.inventoryRepo.GetStockByProductAndWarehouse(ctx, req.ProductID, req.WarehouseID, req.LocationID)
	if err != nil && !errors.Is(err, errpkg.ErrStockNotFound) {
		return fmt.Errorf("%w: %v", errpkg.ErrGetStock, err)
	}

	var newQuantity int
	var movementType string

	switch req.Type {
	case "INCREMENT":
		newQuantity = currentStock.Quantity + req.Quantity
		movementType = "STOCK_IN"
	case "DECREMENT":
		if currentStock.Quantity < req.Quantity {
			return errpkg.ErrInsufficientStock
		}
		newQuantity = currentStock.Quantity - req.Quantity
		movementType = "STOCK_OUT"
	case "SET":
		newQuantity = req.Quantity
		movementType = "STOCK_ADJUSTMENT"
	default:
		return errpkg.ErrInvalidStockOperation
	}

	// Create stock movement record
	movement := &entity.StockMovement{
		ReferenceNumber: s.generateReferenceNumber("STK"),
		MovementType:    movementType,
		ProductID:       req.ProductID,
		ToWarehouseID:   &req.WarehouseID,
		ToLocationID:    req.LocationID,
		Quantity:        req.Quantity,
		UnitPrice:       product.CostPrice,
		TotalValue:      product.CostPrice.Mul(decimal.NewFromInt(int64(req.Quantity))),
		Notes:           &req.Reason,
		Status:          "COMPLETED",
		MovementDate:    time.Now(),
		CreatedBy:       userID,
		UpdatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Update or create inventory stock record
	inventoryStock := &entity.InventoryStock{
		ProductID:        req.ProductID,
		WarehouseID:      req.WarehouseID,
		LocationID:       req.LocationID,
		Quantity:         newQuantity,
		ReservedQuantity: currentStock.ReservedQuantity,
		UpdatedBy:        userID,
		UpdatedAt:        time.Now(),
	}

	if currentStock.ID == uuid.Nil {
		// Create new stock record
		inventoryStock.ID = uuid.New()
		inventoryStock.CreatedBy = userID
		inventoryStock.CreatedAt = time.Now()
		err = s.inventoryRepo.CreateStock(ctx, inventoryStock)
	} else {
		// Update existing stock record
		inventoryStock.ID = currentStock.ID
		inventoryStock.CreatedBy = currentStock.CreatedBy
		inventoryStock.CreatedAt = currentStock.CreatedAt
		err = s.inventoryRepo.UpdateStock(ctx, currentStock.ID, inventoryStock)
	}

	if err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrUpdateStock, err)
	}

	// Update product total stock
	productUpdate := req.Quantity
	if req.Type == "DECREMENT" {
		productUpdate = -req.Quantity
	} else if req.Type == "SET" {
		productUpdate = newQuantity - product.CurrentStock
	}

	if err := s.productRepo.UpdateStock(ctx, req.ProductID, productUpdate); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrUpdateProductStock, err)
	}

	// Save stock movement
	if err := s.inventoryRepo.CreateStockMovement(ctx, movement); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrCreateStockMovement, err)
	}

	return nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Check if product exists
	_, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrProductNotFound, err)
	}

	// Soft delete the product
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrDeleteProduct, err)
	}

	return nil
}

func (s *ProductService) GetLowStockProducts(ctx context.Context) ([]entity.ProductResponse, error) {
	// Get low stock products from repository
	products, err := s.productRepo.FindAll(ctx, entity.ProductFilter{LowStockOnly: boolPtr(true), Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrGetProducts, err)
	}

	// Convert to response DTO
	responses := make([]entity.ProductResponse, len(products))
	for i, product := range products {
		stockInfo, _ := s.getProductStockInfo(ctx, product.ID)

		responses[i] = entity.ProductResponse{
			Product: product,
			StockInfo: struct {
				TotalAvailable int `json:"total_available"`
				TotalReserved  int `json:"total_reserved"`
				TotalOnHand    int `json:"total_on_hand"`
			}{
				TotalAvailable: stockInfo.TotalAvailable,
				TotalReserved:  stockInfo.TotalReserved,
				TotalOnHand:    product.CurrentStock,
			},
		}
	}

	return responses, nil
}

func (s *ProductService) SearchProducts(ctx context.Context, query string, limit int) ([]entity.ProductResponse, error) {
	// Search products by name or SKU
	namePtr := &query
	filter := entity.ProductFilter{
		Name:  namePtr,
		Limit: limit,
	}

	products, err := s.productRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrGetProducts, err)
	}

	// Convert to response DTO
	responses := make([]entity.ProductResponse, len(products))
	for i, product := range products {
		stockInfo, _ := s.getProductStockInfo(ctx, product.ID)

		responses[i] = entity.ProductResponse{
			Product: product,
			StockInfo: struct {
				TotalAvailable int `json:"total_available"`
				TotalReserved  int `json:"total_reserved"`
				TotalOnHand    int `json:"total_on_hand"`
			}{
				TotalAvailable: stockInfo.TotalAvailable,
				TotalReserved:  stockInfo.TotalReserved,
				TotalOnHand:    product.CurrentStock,
			},
		}
	}

	return responses, nil
}

func boolPtr(b bool) *bool {
	return &b
}

func (s *ProductService) generateReferenceNumber(prefix string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s-%s", prefix, timestamp)
}

func (s *ProductService) auditProductCreation(ctx context.Context, product *entity.Product, userID uuid.UUID) {
	// Implementation for audit logging
}

func (s *ProductService) auditProductUpdate(ctx context.Context, oldProduct, newProduct *entity.Product, userID uuid.UUID) {
	// Implementation for audit logging
}
