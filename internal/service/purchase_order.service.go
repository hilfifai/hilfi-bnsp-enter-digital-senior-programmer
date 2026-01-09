// internal/service/purchase_order.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	errpkg "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type IPurchaseOrderService interface {
	CreatePurchaseOrder(ctx context.Context, req dto.CreatePurchaseOrderRequest, userID uuid.UUID) (*entity.PurchaseOrder, error)
	GetPurchaseOrderByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrder, error)
	GetPurchaseOrders(ctx context.Context, filter dto.PurchaseOrderFilter) ([]entity.PurchaseOrder, int, error)
	ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error)
	CancelPurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error)
}

type PurchaseOrderService struct {
	purchaseOrderRepo repository.IPurchaseOrderRepository
	supplierRepo      repository.ISupplierRepository
	productRepo       repository.IProductRepository
}

func NewPurchaseOrderService(
	purchaseOrderRepo repository.IPurchaseOrderRepository,
	supplierRepo repository.ISupplierRepository,
	productRepo repository.IProductRepository,
) IPurchaseOrderService {
	return &PurchaseOrderService{
		purchaseOrderRepo: purchaseOrderRepo,
		supplierRepo:      supplierRepo,
		productRepo:       productRepo,
	}
}

func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, req dto.CreatePurchaseOrderRequest, userID uuid.UUID) (*entity.PurchaseOrder, error) {
	// Validate supplier exists
	supplier, err := s.supplierRepo.FindByID(ctx, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	// Validate items
	var totalAmount decimal.Decimal
	var items []entity.PurchaseOrderItem

	for _, itemReq := range req.Items {
		// Validate product exists
		_, err := s.productRepo.FindByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %w", err)
		}

		// Validate unit price
		if itemReq.UnitPrice.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("invalid unit price: %w", errpkg.ErrInvalidPrice)
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
		return nil, fmt.Errorf("failed to create purchase order: %w", err)
	}

	// Save purchase order items
	for i := range items {
		items[i].PurchaseOrderID = purchaseOrder.ID
		if err := s.purchaseOrderRepo.CreateItem(ctx, &items[i]); err != nil {
			return nil, fmt.Errorf("failed to create purchase order item: %w", err)
		}
	}

	// Add supplier and items to response
	purchaseOrder.Supplier = supplier
	purchaseOrder.Items = items

	return purchaseOrder, nil
}

func (s *PurchaseOrderService) GetPurchaseOrderByID(ctx context.Context, id uuid.UUID) (*entity.PurchaseOrder, error) {
	purchaseOrder, err := s.purchaseOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found: %w", err)
	}

	// Fetch items
	items, err := s.purchaseOrderRepo.GetItemsByOrderID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchase order items: %w", err)
	}
	purchaseOrder.Items = items

	return purchaseOrder, nil
}

func (s *PurchaseOrderService) GetPurchaseOrders(ctx context.Context, filter dto.PurchaseOrderFilter) ([]entity.PurchaseOrder, int, error) {
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

func (s *PurchaseOrderService) ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error) {
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

func (s *PurchaseOrderService) CancelPurchaseOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PurchaseOrder, error) {
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
