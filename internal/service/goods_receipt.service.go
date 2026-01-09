// internal/service/goods_receipt.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type IGoodsReceiptService interface {
	CreateGoodsReceipt(ctx context.Context, req dto.GoodsReceiptRequest, userID uuid.UUID) (*entity.GoodsReceipt, error)
	GetGoodsReceiptByID(ctx context.Context, id uuid.UUID) (*entity.GoodsReceipt, error)
}

type GoodsReceiptService struct {
	goodsReceiptRepo  repository.IGoodsReceiptRepository
	purchaseOrderRepo repository.IPurchaseOrderRepository
	inventoryService  IInventoryService
}

func NewGoodsReceiptService(
	goodsReceiptRepo repository.IGoodsReceiptRepository,
	purchaseOrderRepo repository.IPurchaseOrderRepository,
	inventoryService IInventoryService,
) IGoodsReceiptService {
	return &GoodsReceiptService{
		goodsReceiptRepo:  goodsReceiptRepo,
		purchaseOrderRepo: purchaseOrderRepo,
		inventoryService:  inventoryService,
	}
}

func (s *GoodsReceiptService) CreateGoodsReceipt(ctx context.Context, req dto.GoodsReceiptRequest, userID uuid.UUID) (*entity.GoodsReceipt, error) {
	// Get purchase order
	purchaseOrder, err := s.purchaseOrderRepo.FindByID(ctx, req.PurchaseOrderID)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found: %w", err)
	}

	// Validate purchase order status
	if purchaseOrder.Status != "APPROVED" && purchaseOrder.Status != "PARTIAL_RECEIVED" {
		return nil, fmt.Errorf("purchase order must be approved before receiving goods")
	}

	// Get items for the purchase order
	items, err := s.purchaseOrderRepo.GetItemsByOrderID(ctx, req.PurchaseOrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchase order items: %w", err)
	}

	// Create goods receipt
	grNumber := fmt.Sprintf("GR-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)

	goodsReceipt := &entity.GoodsReceipt{
		ID:              uuid.New(),
		GRNumber:        grNumber,
		PurchaseOrderID: req.PurchaseOrderID,
		WarehouseID:     req.WarehouseID,
		ReceiptDate:     req.ReceiptDate,
		Status:          "COMPLETED",
		Notes:           req.Notes,
		ReceivedBy:      userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save goods receipt
	if err := s.goodsReceiptRepo.Create(ctx, goodsReceipt); err != nil {
		return nil, fmt.Errorf("failed to create goods receipt: %w", err)
	}

	// Update purchase order item remaining quantities
	for _, itemReq := range req.Items {
		// Get purchase order item
		poItem, err := s.purchaseOrderRepo.FindItemByID(ctx, itemReq.PurchaseOrderItemID)
		if err != nil {
			return nil, fmt.Errorf("purchase order item not found: %w", err)
		}

		// Validate received quantity
		if itemReq.ReceivedQty <= 0 {
			return nil, fmt.Errorf("received quantity must be greater than 0")
		}
		if itemReq.ReceivedQty > poItem.RemainingQty {
			return nil, fmt.Errorf("received quantity exceeds remaining quantity")
		}

		// Create goods receipt item
		grItem := &entity.GoodsReceiptItem{
			ID:                  uuid.New(),
			GoodsReceiptID:      goodsReceipt.ID,
			PurchaseOrderItemID: itemReq.PurchaseOrderItemID,
			ReceivedQty:         itemReq.ReceivedQty,
			LocationID:          itemReq.LocationID,
			BatchNumber:         itemReq.BatchNumber,
			ExpiryDate:          itemReq.ExpiryDate,
			Notes:               itemReq.Notes,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		if err := s.goodsReceiptRepo.CreateItem(ctx, grItem); err != nil {
			return nil, fmt.Errorf("failed to create goods receipt item: %w", err)
		}

		// Update purchase order item
		poItem.ReceivedQty += itemReq.ReceivedQty
		poItem.RemainingQty = poItem.Quantity - poItem.ReceivedQty
		poItem.UpdatedAt = time.Now()

		if err := s.purchaseOrderRepo.UpdateItem(ctx, poItem.ID, poItem); err != nil {
			return nil, fmt.Errorf("failed to update purchase order item: %w", err)
		}

		// Create stock movement
		notes := fmt.Sprintf("Goods Receipt: %s", grNumber)
		stockMovement := dto.StockMovementRequest{
			MovementType:  "IN",
			ProductID:     poItem.ProductID,
			ToWarehouseID: &req.WarehouseID,
			ToLocationID:  itemReq.LocationID,
			Quantity:      itemReq.ReceivedQty,
			UnitPrice:     poItem.UnitPrice,
			Notes:         &notes,
			BatchNumber:   itemReq.BatchNumber,
			ExpiryDate:    itemReq.ExpiryDate,
			MovementDate:  req.ReceiptDate,
		}

		if _, err := s.inventoryService.CreateStockMovement(ctx, stockMovement, userID); err != nil {
			return nil, fmt.Errorf("failed to create stock movement: %w", err)
		}
	}

	// Update purchase order status
	totalItems := len(items)
	completedItems := 0
	for _, item := range items {
		if item.RemainingQty == 0 {
			completedItems++
		}
	}

	if completedItems == totalItems {
		purchaseOrder.Status = "COMPLETED"
	} else if completedItems > 0 {
		purchaseOrder.Status = "PARTIAL_RECEIVED"
	}

	purchaseOrder.UpdatedAt = time.Now()
	if err := s.purchaseOrderRepo.Update(ctx, purchaseOrder.ID, purchaseOrder); err != nil {
		return nil, fmt.Errorf("failed to update purchase order: %w", err)
	}

	return goodsReceipt, nil
}

func (s *GoodsReceiptService) GetGoodsReceiptByID(ctx context.Context, id uuid.UUID) (*entity.GoodsReceipt, error) {
	goodsReceipt, err := s.goodsReceiptRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("goods receipt not found: %w", err)
	}

	// Fetch items
	items, err := s.goodsReceiptRepo.GetItemsByReceiptID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch goods receipt items: %w", err)
	}
	goodsReceipt.Items = items

	return goodsReceipt, nil
}
