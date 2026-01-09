// internal/service/inventory.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	errpkg "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type IInventoryService interface {
	CreateStockMovement(ctx context.Context, req dto.StockMovementRequest, userID uuid.UUID) (*entity.StockMovement, error)
	TransferStock(ctx context.Context, req dto.StockTransferRequest, userID uuid.UUID) (*entity.StockMovement, error)
	AdjustStock(ctx context.Context, req dto.StockAdjustmentRequest, userID uuid.UUID) (*entity.StockMovement, error)
	PerformStockOpname(ctx context.Context, req dto.StockOpnameRequest, userID uuid.UUID) ([]entity.StockMovement, error)

	GetStockMovements(ctx context.Context, filter dto.StockReportFilter) ([]entity.StockMovement, int, error)
	GetStockByProduct(ctx context.Context, productID uuid.UUID) ([]entity.InventoryStock, error)
	GetStockByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.InventoryStock, error)
	GetStockSummary(ctx context.Context) (*entity.StockSummary, error)
	GetLowStockAlerts(ctx context.Context) ([]entity.StockAlert, error)
	GetStockHistory(ctx context.Context, productID uuid.UUID, warehouseID uuid.UUID, days int) ([]entity.StockMovement, error)
}

type InventoryService struct {
	inventoryRepo repository.IInventoryRepository
	productRepo   repository.IProductRepository
	warehouseRepo repository.IWarehouseRepository
	locationRepo  repository.IWarehouseLocationRepository
}

func NewInventoryService(
	inventoryRepo repository.IInventoryRepository,
	productRepo repository.IProductRepository,
	warehouseRepo repository.IWarehouseRepository,
	locationRepo repository.IWarehouseLocationRepository,
) IInventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
	}
}

func (s *InventoryService) validateStockMovement(ctx context.Context, req dto.StockMovementRequest) error {
	switch req.MovementType {
	case "IN":
		// Stock In: must have destination warehouse
		if req.ToWarehouseID == nil {
			return errpkg.ErrMissingDestinationWarehouse
		}
		// Validate destination warehouse
		_, err := s.warehouseRepo.FindByID(ctx, *req.ToWarehouseID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errpkg.ErrWarehouseNotFound
			}
			return fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
		}
		// Validate location if provided
		if req.ToLocationID != nil {
			_, err := s.locationRepo.FindByID(ctx, *req.ToLocationID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errpkg.ErrLocationNotFound
				}
				return fmt.Errorf("%w: %v", errpkg.ErrLocationNotFound, err)
			}
		}

	case "OUT":
		// Stock Out: must have source warehouse
		if req.FromWarehouseID == nil {
			return errpkg.ErrMissingSourceWarehouse
		}
		// Validate source warehouse
		_, err := s.warehouseRepo.FindByID(ctx, *req.FromWarehouseID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errpkg.ErrWarehouseNotFound
			}
			return fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
		}
		// Check stock availability
		stock, err := s.inventoryRepo.GetStockByProductAndWarehouse(
			ctx, req.ProductID, *req.FromWarehouseID, req.FromLocationID,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: stock record not found for product in warehouse", errpkg.ErrStockNotFound)
			}
			return fmt.Errorf("%w: %v", errpkg.ErrGetStock, err)
		}
		if stock.AvailableQuantity < req.Quantity {
			return errpkg.ErrInsufficientStock
		}

	case "TRANSFER":
		// Transfer: must have both source and destination
		if req.FromWarehouseID == nil || req.ToWarehouseID == nil {
			return errpkg.ErrMissingWarehouseForTransfer
		}
		// Source and destination cannot be same
		if *req.FromWarehouseID == *req.ToWarehouseID {
			return errpkg.ErrSameWarehouseTransfer
		}
		// Validate both warehouses
		_, err := s.warehouseRepo.FindByID(ctx, *req.FromWarehouseID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errpkg.ErrWarehouseNotFound
			}
			return fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
		}
		_, err = s.warehouseRepo.FindByID(ctx, *req.ToWarehouseID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errpkg.ErrWarehouseNotFound
			}
			return fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
		}
		// Check source stock availability
		stock, err := s.inventoryRepo.GetStockByProductAndWarehouse(
			ctx, req.ProductID, *req.FromWarehouseID, req.FromLocationID,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: stock record not found for product in warehouse", errpkg.ErrStockNotFound)
			}
			return fmt.Errorf("%w: %v", errpkg.ErrGetStock, err)
		}
		if stock.AvailableQuantity < req.Quantity {
			return errpkg.ErrInsufficientStock
		}
	}

	// Validate price
	if req.UnitPrice.LessThan(decimal.Zero) {
		return errpkg.ErrInvalidPrice
	}

	return nil
}

func (s *InventoryService) CreateStockMovement(ctx context.Context, req dto.StockMovementRequest, userID uuid.UUID) (*entity.StockMovement, error) {
	if err := s.validateStockMovement(ctx, req); err != nil {
		return nil, err
	}

	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errpkg.ErrProductNotFound
		}
		return nil, fmt.Errorf("%w: %v", errpkg.ErrProductNotFound, err)
	}

	movement := &entity.StockMovement{
		ReferenceNumber: s.generateMovementReference(req.MovementType),
		MovementType:    req.MovementType,
		ProductID:       req.ProductID,
		FromWarehouseID: req.FromWarehouseID,
		FromLocationID:  req.FromLocationID,
		ToWarehouseID:   req.ToWarehouseID,
		ToLocationID:    req.ToLocationID,
		Quantity:        req.Quantity,
		UnitPrice:       req.UnitPrice,
		TotalValue:      req.UnitPrice.Mul(decimal.NewFromInt(int64(req.Quantity))),
		Notes:           req.Notes,
		Status:          "PENDING",
		MovementDate:    req.MovementDate,
		CreatedBy:       userID,
		UpdatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if movement.MovementDate.IsZero() {
		movement.MovementDate = time.Now()
	}

	movement.ID = uuid.New()

	// Start transaction
	tx, err := s.inventoryRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrBeginTransaction, err)
	}
	defer func() {
		if tx != nil {
			sqlxTx, _ := tx.(*sqlx.Tx)
			sqlxTx.Rollback()
		}
	}()

	// Process based on movement type
	switch req.MovementType {
	case "IN":
		err = s.processStockIn(ctx, tx, movement, product, userID, req.BatchNumber, req.ExpiryDate)
	case "OUT":
		err = s.processStockOut(ctx, tx, movement, product, userID)
	case "TRANSFER":
		err = s.processStockTransfer(ctx, tx, movement, product, userID)
	case "ADJUSTMENT":
		err = s.processStockAdjustment(ctx, tx, movement, product, userID)
	default:
		err = errpkg.ErrInvalidMovementType
	}

	if err != nil {
		return nil, err
	}

	// Create stock movement record first
	if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, movement); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrCreateStockMovement, err)
	}

	// Update movement status
	movement.Status = "COMPLETED"
	if err := s.inventoryRepo.UpdateStockMovement(ctx, tx, movement); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrUpdateStockMovement, err)
	}

	// Commit transaction
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type: %w", errpkg.ErrBeginTransaction)
	}
	if err := sqlxTx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrCommitTransaction, err)
	}
	tx = nil // Prevent rollback

	// Add product info to response
	movement.Product = product

	return movement, nil
}

func (s *InventoryService) processStockIn(ctx context.Context, tx interface{}, movement *entity.StockMovement, product *entity.Product, userID uuid.UUID, batchNumber *string, expiryDate *time.Time) error {
	// Get or create inventory stock record
	stock, err := s.inventoryRepo.GetStockByProductAndWarehouseTx(
		ctx, tx, movement.ProductID, *movement.ToWarehouseID, movement.ToLocationID,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if stock == nil {
		// Create new stock record
		stock = &entity.InventoryStock{
			ID:               uuid.New(),
			ProductID:        movement.ProductID,
			WarehouseID:      *movement.ToWarehouseID,
			LocationID:       movement.ToLocationID,
			Quantity:         movement.Quantity,
			ReservedQuantity: 0,
			BatchNumber:      batchNumber,
			ExpiryDate:       expiryDate,
			CreatedBy:        userID,
			UpdatedBy:        userID,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := s.inventoryRepo.CreateStockTx(ctx, tx, stock); err != nil {
			return fmt.Errorf("%w: %v", errpkg.ErrCreateStock, err)
		}
	} else {
		// Update existing stock
		stock.Quantity += movement.Quantity
		stock.UpdatedBy = userID
		stock.UpdatedAt = time.Now()
		if err := s.inventoryRepo.UpdateStockTx(ctx, tx, stock.ID, stock); err != nil {
			return fmt.Errorf("%w: %v", errpkg.ErrUpdateStock, err)
		}
	}

	// Update product total stock
	if err := s.productRepo.UpdateStock(ctx, movement.ProductID, movement.Quantity); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrUpdateProductStock, err)
	}

	// Update location usage if applicable
	if movement.ToLocationID != nil {
		if err := s.locationRepo.UpdateUsage(ctx, *movement.ToLocationID, movement.Quantity); err != nil {
			return fmt.Errorf("%w: %v", errpkg.ErrUpdateLocationUsage, err)
		}
	}

	return nil
}

func (s *InventoryService) processStockOut(ctx context.Context, tx interface{}, movement *entity.StockMovement, product *entity.Product, userID uuid.UUID) error {
	// Get stock record
	stock, err := s.inventoryRepo.GetStockByProductAndWarehouseTx(
		ctx, tx, movement.ProductID, *movement.FromWarehouseID, movement.FromLocationID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: stock record not found for product in warehouse", errpkg.ErrStockNotFound)
		}
		return fmt.Errorf("%w: %v", errpkg.ErrStockNotFound, err)
	}

	// Check if we should use reserved quantity first
	if stock.ReservedQuantity > 0 {
		reservedToUse := stock.ReservedQuantity
		if reservedToUse > movement.Quantity {
			reservedToUse = movement.Quantity
		}
		stock.ReservedQuantity -= reservedToUse
		movement.Quantity -= reservedToUse
	}

	// Use regular quantity for remaining
	if movement.Quantity > 0 {
		if stock.Quantity < movement.Quantity {
			return errpkg.ErrInsufficientStock
		}
		stock.Quantity -= movement.Quantity
	}

	stock.UpdatedBy = userID
	stock.UpdatedAt = time.Now()

	if err := s.inventoryRepo.UpdateStockTx(ctx, tx, stock.ID, stock); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrUpdateStock, err)
	}

	// Update product total stock (negative)
	if err := s.productRepo.UpdateStock(ctx, movement.ProductID, -movement.Quantity); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrUpdateProductStock, err)
	}

	// Update location usage if applicable
	if movement.FromLocationID != nil {
		if err := s.locationRepo.UpdateUsage(ctx, *movement.FromLocationID, -movement.Quantity); err != nil {
			return fmt.Errorf("%w: %v", errpkg.ErrUpdateLocationUsage, err)
		}
	}

	return nil
}

func (s *InventoryService) processStockTransfer(ctx context.Context, tx interface{}, movement *entity.StockMovement, product *entity.Product, userID uuid.UUID) error {
	// Process as stock out from source
	outMovement := *movement
	outMovement.ReferenceNumber = s.generateMovementReference("OUT")
	outMovement.ID = uuid.New()
	outMovement.Status = "COMPLETED"

	if err := s.processStockOut(ctx, tx, &outMovement, product, userID); err != nil {
		return err
	}

	// Process as stock in to destination
	inMovement := *movement
	inMovement.ReferenceNumber = s.generateMovementReference("IN")
	inMovement.ID = uuid.New()
	inMovement.Status = "COMPLETED"
	inMovement.FromWarehouseID = nil
	inMovement.FromLocationID = nil

	if err := s.processStockIn(ctx, tx, &inMovement, product, userID, nil, nil); err != nil {
		return err
	}

	// Update main movement to link both
	movement.ReferenceNumber = s.generateMovementReference("TRANSFER")

	return nil
}

func (s *InventoryService) processStockAdjustment(ctx context.Context, tx interface{}, movement *entity.StockMovement, product *entity.Product, userID uuid.UUID) error {
	// For adjustments, we just update the stock directly
	stock, err := s.inventoryRepo.GetStockByProductAndWarehouseTx(
		ctx, tx, movement.ProductID, *movement.ToWarehouseID, movement.ToLocationID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: stock record not found for product in warehouse", errpkg.ErrStockNotFound)
		}
		return fmt.Errorf("%w: %v", errpkg.ErrStockNotFound, err)
	}

	// Calculate the difference
	difference := movement.Quantity - stock.Quantity
	stock.Quantity = movement.Quantity
	stock.UpdatedBy = userID
	stock.UpdatedAt = time.Now()

	if err := s.inventoryRepo.UpdateStockTx(ctx, tx, stock.ID, stock); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrUpdateStock, err)
	}

	// Update product total stock
	if err := s.productRepo.UpdateStock(ctx, movement.ProductID, difference); err != nil {
		return fmt.Errorf("%w: %v", errpkg.ErrUpdateProductStock, err)
	}

	return nil
}

func (s *InventoryService) AdjustStock(ctx context.Context, req dto.StockAdjustmentRequest, userID uuid.UUID) (*entity.StockMovement, error) {
	// Validate product exists
	_, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errpkg.ErrProductNotFound
		}
		return nil, fmt.Errorf("%w: %v", errpkg.ErrProductNotFound, err)
	}

	// Validate warehouse exists
	_, err = s.warehouseRepo.FindByID(ctx, req.WarehouseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errpkg.ErrWarehouseNotFound
		}
		return nil, fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
	}

	// Validate location if provided
	if req.LocationID != nil {
		_, err = s.locationRepo.FindByID(ctx, *req.LocationID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errpkg.ErrLocationNotFound
			}
			return nil, fmt.Errorf("%w: %v", errpkg.ErrLocationNotFound, err)
		}
	}

	movement := &entity.StockMovement{
		ReferenceNumber: s.generateMovementReference("ADJUSTMENT"),
		MovementType:    "ADJUSTMENT",
		ProductID:       req.ProductID,
		ToWarehouseID:   &req.WarehouseID,
		ToLocationID:    req.LocationID,
		Quantity:        req.NewQuantity,
		UnitPrice:       decimal.Zero,
		TotalValue:      decimal.Zero,
		Notes:           &req.Reason,
		Status:          "COMPLETED",
		MovementDate:    time.Now(),
		CreatedBy:       userID,
		UpdatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	movement.ID = uuid.New()

	// Start transaction
	tx, err := s.inventoryRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrBeginTransaction, err)
	}
	defer func() {
		if tx != nil {
			sqlxTx, _ := tx.(*sqlx.Tx)
			sqlxTx.Rollback()
		}
	}()

	// Process adjustment
	if err := s.processStockAdjustment(ctx, tx, movement, nil, userID); err != nil {
		return nil, err
	}

	// Create stock movement record first
	if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, movement); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrCreateStockMovement, err)
	}

	// Update movement status
	if err := s.inventoryRepo.UpdateStockMovement(ctx, tx, movement); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrUpdateStockMovement, err)
	}

	// Commit transaction
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type: %w", errpkg.ErrBeginTransaction)
	}
	if err := sqlxTx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrCommitTransaction, err)
	}
	tx = nil // Prevent rollback

	return movement, nil
}

func (s *InventoryService) TransferStock(ctx context.Context, req dto.StockTransferRequest, userID uuid.UUID) (*entity.StockMovement, error) {
	movementReq := dto.StockMovementRequest{
		MovementType:    "TRANSFER",
		ProductID:       req.ProductID,
		FromWarehouseID: &req.FromWarehouseID,
		FromLocationID:  req.FromLocationID,
		ToWarehouseID:   &req.ToWarehouseID,
		ToLocationID:    req.ToLocationID,
		Quantity:        req.Quantity,
		UnitPrice:       decimal.Zero,
		Notes:           req.Notes,
		MovementDate:    time.Now(),
	}

	return s.CreateStockMovement(ctx, movementReq, userID)
}

func (s *InventoryService) PerformStockOpname(ctx context.Context, req dto.StockOpnameRequest, userID uuid.UUID) ([]entity.StockMovement, error) {
	// Validate warehouse exists
	_, err := s.warehouseRepo.FindByID(ctx, req.WarehouseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errpkg.ErrWarehouseNotFound
		}
		return nil, fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
	}

	var movements []entity.StockMovement
	now := time.Now()

	// Start transaction
	tx, err := s.inventoryRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrBeginTransaction, err)
	}
	defer func() {
		if tx != nil {
			sqlxTx, _ := tx.(*sqlx.Tx)
			sqlxTx.Rollback()
		}
	}()

	// Process each opname item
	for _, item := range req.Items {
		// Validate product exists
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // Skip invalid products
			}
			continue
		}

		// Get current stock for the product in the warehouse
		stocks, err := s.inventoryRepo.GetStockByWarehouseAndProduct(ctx, req.WarehouseID, item.ProductID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			continue
		}

		systemQty := 0
		if len(stocks) > 0 {
			systemQty = stocks[0].Quantity
		}

		// Calculate variance
		variance := item.PhysicalQty - systemQty

		// If there's a variance, create a stock movement to adjust
		if variance != 0 {
			movement := &entity.StockMovement{
				ID:              uuid.New(),
				ReferenceNumber: s.generateMovementReference("OPNAME"),
				MovementType:    "ADJUSTMENT",
				ProductID:       item.ProductID,
				ToWarehouseID:   &req.WarehouseID,
				ToLocationID:    item.LocationID,
				Quantity:        systemQty,
				UnitPrice:       product.CostPrice,
				TotalValue:      product.CostPrice.Mul(decimal.NewFromInt(int64(systemQty))),
				Notes:           item.VarianceNotes,
				Status:          "COMPLETED",
				MovementDate:    now,
				CreatedBy:       userID,
				UpdatedBy:       userID,
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			// Process the adjustment
			if err := s.processStockAdjustment(ctx, tx, movement, product, userID); err != nil {
				continue
			}

			// Create stock movement record
			if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, movement); err != nil {
				continue
			}

			// Update movement status
			movement.Status = "COMPLETED"
			if err := s.inventoryRepo.UpdateStockMovement(ctx, tx, movement); err != nil {
				continue
			}

			movements = append(movements, *movement)
		}
	}

	// Commit transaction
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type: %w", errpkg.ErrBeginTransaction)
	}
	if err := sqlxTx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", errpkg.ErrCommitTransaction, err)
	}
	tx = nil // Prevent rollback

	return movements, nil
}

func (s *InventoryService) GetStockMovements(ctx context.Context, filter dto.StockReportFilter) ([]entity.StockMovement, int, error) {
	movementFilter := dto.MovementReportFilter{
		WarehouseID:  filter.WarehouseID,
		ProductID:    filter.ProductID,
		StartDate:    filter.StartDate,
		EndDate:      filter.EndDate,
		MovementType: filter.MovementType,
		Limit:        filter.Limit,
		Offset:       filter.Offset,
	}

	return s.inventoryRepo.GetMovementReport(ctx, movementFilter)
}

func (s *InventoryService) GetStockByProduct(ctx context.Context, productID uuid.UUID) ([]entity.InventoryStock, error) {
	return s.inventoryRepo.GetProductStockByProductID(ctx, productID)
}

func (s *InventoryService) GetStockByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.InventoryStock, error) {
	// Validate warehouse exists
	_, err := s.warehouseRepo.FindByID(ctx, warehouseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errpkg.ErrWarehouseNotFound
		}
		return nil, fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
	}

	return s.inventoryRepo.GetStockByWarehouse(ctx, warehouseID)
}

func (s *InventoryService) GetStockSummary(ctx context.Context) (*entity.StockSummary, error) {
	// Get total products
	totalProducts, err := s.inventoryRepo.GetTotalProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total products: %w", err)
	}

	// Get total items (sum of all inventory)
	valuations, err := s.inventoryRepo.GetInventoryValuation(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory valuation: %w", err)
	}

	totalItems := 0
	var totalValue decimal.Decimal
	for _, v := range valuations {
		totalItems += v.TotalQuantity
		totalValue = totalValue.Add(v.TotalValue)
	}

	// Get low stock count
	lowStockCount, err := s.inventoryRepo.GetLowStockCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock count: %w", err)
	}

	// Get out of stock count
	outOfStockCount, err := s.inventoryRepo.GetOutOfStockCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get out of stock count: %w", err)
	}

	// Get pending movements count
	pendingMovements, err := s.inventoryRepo.GetPendingMovementsCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending movements count: %w", err)
	}

	return &entity.StockSummary{
		TotalProducts:    totalProducts,
		TotalItems:       totalItems,
		TotalValue:       totalValue,
		LowStockCount:    lowStockCount,
		OutOfStockCount:  outOfStockCount,
		PendingMovements: pendingMovements,
	}, nil
}

func (s *InventoryService) GetLowStockAlerts(ctx context.Context) ([]entity.StockAlert, error) {
	// Get low stock report
	reports, err := s.inventoryRepo.GetLowStockReport(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock report: %w", err)
	}

	var alerts []entity.StockAlert
	for _, report := range reports {
		// Get product details
		product, err := s.productRepo.FindByID(ctx, report.ProductID)
		if err != nil {
			continue
		}

		alertType := "LOW_STOCK"
		if report.CurrentStock == 0 {
			alertType = "OUT_OF_STOCK"
		}

		message := fmt.Sprintf("Product %s has low stock: %d units (min: %d)",
			product.Name, report.CurrentStock, report.MinStock)

		alerts = append(alerts, entity.StockAlert{
			ID:           uuid.New(),
			ProductID:    report.ProductID,
			Product:      product,
			AlertType:    alertType,
			CurrentStock: report.CurrentStock,
			Threshold:    report.MinStock,
			Message:      message,
			IsRead:       false,
			CreatedAt:    time.Now(),
		})
	}

	return alerts, nil
}

func (s *InventoryService) GetStockHistory(ctx context.Context, productID uuid.UUID, warehouseID uuid.UUID, days int) ([]entity.StockMovement, error) {
	// Validate product exists
	_, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errpkg.ErrProductNotFound
		}
		return nil, fmt.Errorf("%w: %v", errpkg.ErrProductNotFound, err)
	}

	// Validate warehouse exists
	_, err = s.warehouseRepo.FindByID(ctx, warehouseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errpkg.ErrWarehouseNotFound
		}
		return nil, fmt.Errorf("%w: %v", errpkg.ErrWarehouseNotFound, err)
	}

	return s.inventoryRepo.GetProductStockHistory(ctx, productID, warehouseID, days)
}

func (s *InventoryService) generateMovementReference(movementType string) string {
	timestamp := time.Now().Format("20060102150405")
	prefix := ""
	switch movementType {
	case "IN":
		prefix = "STK-IN"
	case "OUT":
		prefix = "STK-OUT"
	case "TRANSFER":
		prefix = "STK-TRF"
	case "ADJUSTMENT":
		prefix = "STK-ADJ"
	case "OPNAME":
		prefix = "STK-OPN"
	default:
		prefix = "STK"
	}
	return fmt.Sprintf("%s-%s", prefix, timestamp)
}
