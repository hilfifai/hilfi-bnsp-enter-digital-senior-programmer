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
	"github.com/shopspring/decimal"
)

type IDeliveryOrderService interface {
	CreateDeliveryOrder(ctx context.Context, req dto.CreateDeliveryOrderRequest, userID uuid.UUID) (*entity.DeliveryOrder, error)
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
		ID:           uuid.New(),
		DONumber:     doNumber,
		SalesOrderID: req.SalesOrderID,
		WarehouseID:  req.WarehouseID,
		DeliveryDate: req.DeliveryDate,
		Status:       "DRAFT",
		Notes:        &notes,
		DeliveredBy:  &userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
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

func (s *DeliveryOrderService) ProcessSalesReturn(ctx context.Context, req dto.SalesReturnRequest, userID uuid.UUID) (*entity.SalesReturn, error) {
	// TODO: Implement ProcessSalesReturn
	return nil, fmt.Errorf("not implemented")
}
