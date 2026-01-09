// internal/service/delivery_order.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type IDeliveryOrderService interface {
	CreateDeliveryOrder(ctx context.Context, req dto.CreateDeliveryOrderRequest, userID uuid.UUID) (*entity.DeliveryOrder, error)
	GetDeliveryOrders(ctx context.Context, filter dto.DeliveryOrderFilter) ([]entity.DeliveryOrder, int, error)
	GetDeliveryOrderByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrder, error)
	ProcessSalesReturn(ctx context.Context, req dto.SalesReturnRequest, userID uuid.UUID) (*entity.SalesReturn, error)
}

type DeliveryOrderService struct {
	deliveryOrderRepo repository.IDeliveryOrderRepository
	salesOrderRepo    repository.ISalesOrderRepository
	inventoryService  IInventoryService
}

func NewDeliveryOrderService(
	deliveryOrderRepo repository.IDeliveryOrderRepository,
	salesOrderRepo repository.ISalesOrderRepository,
	inventoryService IInventoryService,
) IDeliveryOrderService {
	return &DeliveryOrderService{
		deliveryOrderRepo: deliveryOrderRepo,
		salesOrderRepo:    salesOrderRepo,
		inventoryService:  inventoryService,
	}
}

func (s *DeliveryOrderService) CreateDeliveryOrder(ctx context.Context, req dto.CreateDeliveryOrderRequest, userID uuid.UUID) (*entity.DeliveryOrder, error) {
	// Get sales order
	salesOrder, err := s.salesOrderRepo.FindByID(ctx, req.SalesOrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrSalesOrderNotFound, err)
	}

	// Validate sales order status
	if salesOrder.Status != "APPROVED" && salesOrder.Status != "PARTIAL_DELIVERED" {
		return nil, errors.ErrSalesOrderNotApproved
	}

	// Generate DO number
	doNumber, err := s.deliveryOrderRepo.GetNextDONumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGenerateDONumber, err)
	}

	notes := fmt.Sprintf("Delivery Order: %s", doNumber)
	deliveryOrder := &entity.DeliveryOrder{
		ID:              uuid.New(),
		DONumber:        doNumber,
		SalesOrderID:    req.SalesOrderID,
		WarehouseID:     req.WarehouseID,
		DeliveryDate:    req.DeliveryDate,
		DeliveryAddress: req.DeliveryAddress,
		Status:          "DRAFT",
		Notes:           &notes,
		DeliveredBy:     &userID,
		CreatedBy:       userID,
		UpdatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Start transaction
	tx, err := s.deliveryOrderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrBeginTransaction, err)
	}
	sqlxTx := tx.(*sqlx.Tx)
	defer sqlxTx.Rollback()

	// Save delivery order
	if err := s.deliveryOrderRepo.CreateTx(ctx, tx, deliveryOrder); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCreateDeliveryOrder, err)
	}

	// Process each item
	for _, itemReq := range req.Items {
		// Get sales order item
		soItem, err := s.salesOrderRepo.FindItemByID(ctx, itemReq.SalesOrderItemID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrSalesOrderItemNotFound, err)
		}

		// Validate delivered quantity
		if itemReq.DeliveredQty <= 0 {
			return nil, errors.ErrInvalidDeliveredQuantity
		}
		if itemReq.DeliveredQty > soItem.RemainingQty {
			return nil, errors.ErrExceedRemainingQuantity
		}

		// Check stock availability
		unitPrice := decimal.NewFromFloat(soItem.UnitPrice)
		stockMovementReq := dto.StockMovementRequest{
			MovementType:    "OUT",
			ProductID:       soItem.ProductID,
			FromWarehouseID: &req.WarehouseID,
			FromLocationID:  itemReq.LocationID,
			Quantity:        itemReq.DeliveredQty,
			UnitPrice:       unitPrice,
			Notes:           &notes,
			MovementDate:    req.DeliveryDate,
		}

		// This will check stock availability and create stock movement
		if _, err := s.inventoryService.CreateStockMovement(ctx, stockMovementReq, userID); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrCreateStockMovement, err)
		}

		// Create delivery order item
		doItem := &entity.DeliveryOrderItem{
			ID:               uuid.New(),
			DeliveryOrderID:  deliveryOrder.ID,
			SalesOrderItemID: itemReq.SalesOrderItemID,
			DeliveredQty:     itemReq.DeliveredQty,
			LocationID:       itemReq.LocationID,
			BatchNumber:      itemReq.BatchNumber,
			Notes:            itemReq.Notes,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := s.deliveryOrderRepo.CreateItemTx(ctx, tx, doItem); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrCreateDeliveryOrderItem, err)
		}

		// Update sales order item
		soItem.DeliveredQty += itemReq.DeliveredQty
		soItem.RemainingQty = soItem.Quantity - soItem.DeliveredQty
		soItem.UpdatedAt = time.Now()

		if err := s.salesOrderRepo.UpdateItemTx(ctx, tx, soItem.ID, soItem); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrUpdateSalesOrderItem, err)
		}
	}

	// Update sales order status
	totalItems := len(salesOrder.Items)
	completedItems := 0
	for _, item := range salesOrder.Items {
		if item.RemainingQty == 0 {
			completedItems++
		}
	}

	if completedItems == totalItems {
		salesOrder.Status = "COMPLETED"
	} else if completedItems > 0 {
		salesOrder.Status = "PARTIAL_DELIVERED"
	}

	salesOrder.UpdatedAt = time.Now()
	if err := s.salesOrderRepo.UpdateTx(ctx, tx, salesOrder.ID, salesOrder); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrUpdateSalesOrder, err)
	}

	// Update delivery order status
	deliveryOrder.Status = "DELIVERED"
	if err := s.deliveryOrderRepo.UpdateTx(ctx, tx, deliveryOrder.ID, deliveryOrder); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrUpdateDeliveryOrder, err)
	}

	if err := sqlxTx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCommitTransaction, err)
	}

	return deliveryOrder, nil
}

func (s *DeliveryOrderService) GetDeliveryOrders(ctx context.Context, filter dto.DeliveryOrderFilter) ([]entity.DeliveryOrder, int, error) {
	dos, total, err := s.deliveryOrderRepo.FindWithFilter(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	// Fetch items for each delivery order
	for i := range dos {
		items, err := s.deliveryOrderRepo.GetItemsByDOID(ctx, dos[i].ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get delivery order items")
			continue
		}
		dos[i].Items = items
	}
	return dos, total, err
}

func (s *DeliveryOrderService) GetDeliveryOrderByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrder, error) {
	do, err := s.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDeliveryOrderNotFound, err)
	}
	// Fetch items
	items, err := s.deliveryOrderRepo.GetItemsByDOID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get delivery order items")
	}
	do.Items = items
	return do, nil
}

func (s *DeliveryOrderService) ProcessSalesReturn(ctx context.Context, req dto.SalesReturnRequest, userID uuid.UUID) (*entity.SalesReturn, error) {
	// Get delivery order
	deliveryOrder, err := s.deliveryOrderRepo.FindByID(ctx, req.DeliveryOrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDeliveryOrderNotFound, err)
	}

	// Validate delivery order status
	if deliveryOrder.Status != "DELIVERED" && deliveryOrder.Status != "PARTIAL_DELIVERED" {
		return nil, errors.ErrInvalidReturnedQuantity
	}

	// Get sales order
	salesOrder, err := s.salesOrderRepo.FindByID(ctx, deliveryOrder.SalesOrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrSalesOrderNotFound, err)
	}

	// Generate return number
	returnNumber, err := s.deliveryOrderRepo.GetNextReturnNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGenerateReturnNumber, err)
	}

	notes := fmt.Sprintf("Sales Return: %s", returnNumber)

	// Start transaction
	tx, err := s.deliveryOrderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrBeginTransaction, err)
	}
	sqlxTx := tx.(*sqlx.Tx)
	defer sqlxTx.Rollback()

	// Create sales return
	salesReturn := &entity.SalesReturn{
		ID:              uuid.New(),
		ReturnNumber:    returnNumber,
		DeliveryOrderID: req.DeliveryOrderID,
		CustomerID:      salesOrder.CustomerID,
		ReturnDate:      req.ReturnDate,
		Reason:          req.Reason,
		Status:          "PENDING",
		TotalRefund:     0,
		Notes:           req.Notes,
		ProcessedBy:     userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.deliveryOrderRepo.CreateSalesReturn(ctx, salesReturn); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCreateSalesReturn, err)
	}

	// Process each return item
	var totalRefund float64
	for _, itemReq := range req.Items {
		// Get delivery order item
		doItem, err := s.deliveryOrderRepo.GetItemByID(ctx, itemReq.DeliveryOrderItemID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDeliveryOrderNotFound, err)
		}

		// Validate returned quantity
		if itemReq.ReturnedQty <= 0 {
			return nil, errors.ErrInvalidReturnedQuantity
		}
		if itemReq.ReturnedQty > doItem.DeliveredQty {
			return nil, errors.ErrExceedDeliveredQuantity
		}

		// Get sales order item to get unit price
		soItem, err := s.salesOrderRepo.FindItemByID(ctx, doItem.SalesOrderItemID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrSalesOrderItemNotFound, err)
		}

		// Calculate refund
		unitRefund := soItem.UnitPrice
		totalItemRefund := unitRefund * float64(itemReq.ReturnedQty)
		totalRefund += totalItemRefund

		// Create stock movement (IN type) to return items to inventory
		stockMovementReq := dto.StockMovementRequest{
			MovementType:  "IN",
			ProductID:     soItem.ProductID,
			ToWarehouseID: &deliveryOrder.WarehouseID,
			ToLocationID:  doItem.LocationID,
			Quantity:      itemReq.ReturnedQty,
			UnitPrice:     decimal.NewFromFloat(unitRefund),
			Notes:         &notes,
			MovementDate:  req.ReturnDate,
			BatchNumber:   doItem.BatchNumber,
		}

		if _, err := s.inventoryService.CreateStockMovement(ctx, stockMovementReq, userID); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrCreateStockMovement, err)
		}

		// Create sales return item
		salesReturnItem := &entity.SalesReturnItem{
			ID:                  uuid.New(),
			SalesReturnID:       salesReturn.ID,
			DeliveryOrderItemID: itemReq.DeliveryOrderItemID,
			ProductID:           soItem.ProductID,
			ReturnedQty:         itemReq.ReturnedQty,
			UnitRefund:          unitRefund,
			TotalRefund:         totalItemRefund,
			Reason:              itemReq.Reason,
			Notes:               itemReq.Notes,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		if err := s.deliveryOrderRepo.CreateSalesReturnItem(ctx, salesReturnItem); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrCreateSalesReturnItem, err)
		}

		// Update delivery order item
		doItem.DeliveredQty -= itemReq.ReturnedQty
		doItem.UpdatedAt = time.Now()

		if err := s.deliveryOrderRepo.UpdateItemTx(ctx, tx, doItem.ID, doItem); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrUpdateDeliveryOrderItem, err)
		}

		// Update sales order item
		soItem.DeliveredQty -= itemReq.ReturnedQty
		soItem.RemainingQty = soItem.Quantity - soItem.DeliveredQty
		soItem.UpdatedAt = time.Now()

		if err := s.salesOrderRepo.UpdateItemTx(ctx, tx, soItem.ID, soItem); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrUpdateSalesOrderItem, err)
		}
	}

	// Update sales return with total refund
	salesReturn.TotalRefund = totalRefund
	salesReturn.Status = "COMPLETED"

	// Commit transaction
	if err := sqlxTx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCommitTransaction, err)
	}

	return salesReturn, nil
}
