// internal/service/purchase.service.go
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
	"github.com/shopspring/decimal"
)

type IPurchaseService interface {
	CreateSupplier(ctx context.Context, req dto.CreateSupplierRequest, userID uuid.UUID) (*entity.Supplier, error)
	GetSupplierByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error)
	GetSuppliers(ctx context.Context, activeOnly bool) ([]entity.Supplier, error)

	CreatePurchaseOrder(ctx context.Context, req dto.CreatePurchaseOrderRequest, userID uuid.UUID) (*entity.PurchaseOrder, error)
	GetPurchaseOrderByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrder, error)
	GetPurchaseOrders(ctx context.Context, filter dto.PurchaseOrderFilter) ([]entity.PurchaseOrder, int, error)
	ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error)
	CancelPurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error)

	CreateGoodsReceipt(ctx context.Context, req dto.GoodsReceiptRequest, userID uuid.UUID) (*entity.GoodsReceipt, error)
	GetGoodsReceiptByID(ctx context.Context, id uuid.UUID) (*entity.GoodsReceipt, error)
}

type PurchaseService struct {
	supplierRepo      repository.ISupplierRepository
	purchaseOrderRepo repository.IPurchaseOrderRepository
	goodsReceiptRepo  repository.IGoodsReceiptRepository
	productRepo       repository.IProductRepository
	inventoryService  IInventoryService
}

func NewPurchaseService(
	supplierRepo repository.ISupplierRepository,
	purchaseOrderRepo repository.IPurchaseOrderRepository,
	goodsReceiptRepo repository.IGoodsReceiptRepository,
	productRepo repository.IProductRepository,
	inventoryService IInventoryService,
) IPurchaseService {
	return &PurchaseService{
		supplierRepo:      supplierRepo,
		purchaseOrderRepo: purchaseOrderRepo,
		goodsReceiptRepo:  goodsReceiptRepo,
		productRepo:       productRepo,
		inventoryService:  inventoryService,
	}
}

func (s *PurchaseService) CreatePurchaseOrder(ctx context.Context, req dto.CreatePurchaseOrderRequest, userID uuid.UUID) (*entity.PurchaseOrder, error) {
	// Validate supplier exists
	supplier, err := s.supplierRepo.FindByID(ctx, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrSupplierNotFound, err)
	}

	// Validate items
	var totalAmount decimal.Decimal
	var items []entity.PurchaseOrderItem

	for _, itemReq := range req.Items {
		// Validate product exists
		_, err := s.productRepo.FindByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrProductNotFound, err)
		}

		item := entity.PurchaseOrderItem{
			ID:           uuid.New(),
			ProductID:    itemReq.ProductID,
			Quantity:     itemReq.Quantity,
			UnitPrice:    itemReq.UnitPrice,
			TotalPrice:   itemReq.UnitPrice.Mul(decimal.NewFromInt(int64(itemReq.Quantity))),
			ReceivedQty:  0,
			RemainingQty: itemReq.Quantity,
			Notes:        itemReq.Notes,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		items = append(items, item)
		totalAmount = totalAmount.Add(item.TotalPrice)
	}

	// Generate PO number
	poNumber, err := s.purchaseOrderRepo.GetNextPONumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PO number: %w", err)
	}

	purchaseOrder := &entity.PurchaseOrder{
		ID:                   uuid.New(),
		PONumber:             poNumber,
		SupplierID:           req.SupplierID,
		OrderDate:            req.OrderDate,
		ExpectedDeliveryDate: req.ExpectedDeliveryDate,
		Status:               "DRAFT",
		SubTotal:             totalAmount,
		TaxAmount:            decimal.Zero,
		TotalAmount:          totalAmount,
		Notes:                req.Notes,
		CreatedBy:            userID,
		UpdatedBy:            userID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save purchase order
	if err := s.purchaseOrderRepo.Create(ctx, purchaseOrder); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCreatePurchaseOrder, err)
	}

	// Save purchase order items
	for i := range items {
		items[i].PurchaseOrderID = purchaseOrder.ID
		if err := s.purchaseOrderRepo.CreateItem(ctx, &items[i]); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrCreatePurchaseOrderItem, err)
		}
	}

	// Add supplier and items to response
	purchaseOrder.Supplier = supplier
	purchaseOrder.Items = items

	return purchaseOrder, nil
}

func (s *PurchaseService) GetPurchaseOrderByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrder, error) {
	purchaseOrder, err := s.purchaseOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrPurchaseOrderNotFound, err)
	}

	// Fetch items
	items, err := s.purchaseOrderRepo.GetItemsByOrderID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchase order items: %w", err)
	}
	purchaseOrder.Items = items

	return purchaseOrder, nil
}

func (s *PurchaseService) ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error) {
	purchaseOrder, err := s.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if purchaseOrder.Status != "DRAFT" {
		return nil, fmt.Errorf("purchase order cannot be approved in status: %s", purchaseOrder.Status)
	}

	now := time.Now()
	purchaseOrder.Status = "APPROVED"
	purchaseOrder.ApprovedBy = &userID
	purchaseOrder.ApprovedAt = &now
	purchaseOrder.UpdatedAt = now

	if err := s.purchaseOrderRepo.Update(ctx, id, purchaseOrder); err != nil {
		return nil, fmt.Errorf("failed to approve purchase order: %w", err)
	}

	return purchaseOrder, nil
}

func (s *PurchaseService) CancelPurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error) {
	purchaseOrder, err := s.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if purchaseOrder.Status == "COMPLETED" || purchaseOrder.Status == "CANCELLED" {
		return nil, fmt.Errorf("purchase order cannot be cancelled in status: %s", purchaseOrder.Status)
	}

	purchaseOrder.Status = "CANCELLED"
	purchaseOrder.UpdatedBy = userID
	purchaseOrder.UpdatedAt = time.Now()

	if err := s.purchaseOrderRepo.Update(ctx, id, purchaseOrder); err != nil {
		return nil, fmt.Errorf("failed to cancel purchase order: %w", err)
	}

	return purchaseOrder, nil
}

func (s *PurchaseService) CreateGoodsReceipt(ctx context.Context, req dto.GoodsReceiptRequest, userID uuid.UUID) (*entity.GoodsReceipt, error) {
	// Get purchase order
	purchaseOrder, err := s.purchaseOrderRepo.FindByID(ctx, req.PurchaseOrderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrPurchaseOrderNotFound, err)
	}

	// Validate purchase order status
	if purchaseOrder.Status != "APPROVED" && purchaseOrder.Status != "PARTIAL_RECEIVED" {
		return nil, errors.ErrPurchaseOrderNotApproved
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
		return nil, fmt.Errorf("%w: %v", errors.ErrCreateGoodsReceipt, err)
	}

	// Update purchase order item remaining quantities
	for _, itemReq := range req.Items {
		// Get purchase order item
		poItem, err := s.purchaseOrderRepo.FindItemByID(ctx, itemReq.PurchaseOrderItemID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrPurchaseOrderItemNotFound, err)
		}

		// Validate received quantity
		if itemReq.ReceivedQty <= 0 {
			return nil, errors.ErrInvalidReceivedQuantity
		}
		if itemReq.ReceivedQty > poItem.RemainingQty {
			return nil, errors.ErrExceedRemainingQuantity
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
			return nil, fmt.Errorf("%w: %v", errors.ErrCreateGoodsReceiptItem, err)
		}

		// Update purchase order item
		poItem.ReceivedQty += itemReq.ReceivedQty
		poItem.RemainingQty = poItem.Quantity - poItem.ReceivedQty
		poItem.UpdatedAt = time.Now()

		if err := s.purchaseOrderRepo.UpdateItem(ctx, poItem.ID, poItem); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrUpdatePurchaseOrderItem, err)
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
			return nil, fmt.Errorf("%w: %v", errors.ErrCreateStockMovement, err)
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
		return nil, fmt.Errorf("%w: %v", errors.ErrUpdatePurchaseOrder, err)
	}

	return goodsReceipt, nil
}

func (s *PurchaseService) CreateSupplier(ctx context.Context, req dto.CreateSupplierRequest, userID uuid.UUID) (*entity.Supplier, error) {
	supplier := &entity.Supplier{
		ID:            uuid.New(),
		Code:          req.Code,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		TaxNumber:     req.TaxNumber,
		PaymentTerms:  req.PaymentTerms,
		IsActive:      true,
		CreatedBy:     userID,
		UpdatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	return supplier, nil
}

func (s *PurchaseService) GetSupplierByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}
	return supplier, nil
}

func (s *PurchaseService) GetSuppliers(ctx context.Context, activeOnly bool) ([]entity.Supplier, error) {
	return s.supplierRepo.FindAll(ctx, activeOnly)
}

func (s *PurchaseService) GetGoodsReceiptByID(ctx context.Context, id uuid.UUID) (*entity.GoodsReceipt, error) {
	return s.goodsReceiptRepo.FindByID(ctx, id)
}

func (s *PurchaseService) GetPurchaseOrders(ctx context.Context, filter dto.PurchaseOrderFilter) ([]entity.PurchaseOrder, int, error) {
	entityFilter := entity.PurchaseOrderFilter{
		SupplierID: filter.SupplierID,
		Status:     filter.Status,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	}

	orders, err := s.purchaseOrderRepo.FindAll(ctx, entityFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get purchase orders: %w", err)
	}

	return orders, len(orders), nil
}
