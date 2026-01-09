// internal/service/sales_order.service.go
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

type ISalesOrderService interface {
	CreateSalesOrder(ctx context.Context, req dto.CreateSalesOrderRequest, userID uuid.UUID) (*entity.SalesOrder, error)
	GetSalesOrderByID(ctx context.Context, id uuid.UUID) (*entity.SalesOrder, error)
	GetSalesOrders(ctx context.Context, filter dto.SalesOrderFilter) ([]entity.SalesOrder, int, error)
	ApproveSalesOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.SalesOrder, error)
	CancelSalesOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.SalesOrder, error)
}

type SalesOrderService struct {
	salesOrderRepo repository.ISalesOrderRepository
	customerRepo   repository.ICustomerRepository
	productRepo    repository.IProductRepository
}

func NewSalesOrderService(
	salesOrderRepo repository.ISalesOrderRepository,
	customerRepo repository.ICustomerRepository,
	productRepo repository.IProductRepository,
) ISalesOrderService {
	return &SalesOrderService{
		salesOrderRepo: salesOrderRepo,
		customerRepo:   customerRepo,
		productRepo:    productRepo,
	}
}

func (s *SalesOrderService) CreateSalesOrder(ctx context.Context, req dto.CreateSalesOrderRequest, userID uuid.UUID) (*entity.SalesOrder, error) {
	// Validate customer exists
	customer, err := s.customerRepo.FindByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCustomerNotFound, err)
	}

	// Validate items and calculate totals
	var totalAmount decimal.Decimal
	var items []entity.SalesOrderItem

	for _, itemReq := range req.Items {
		// Validate product exists and has enough stock
		_, err := s.productRepo.FindByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrProductNotFound, err)
		}

		// Validate unit price
		if itemReq.UnitPrice.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("%w: unit price cannot be negative", errors.ErrInvalidPrice)
		}

		unitPrice := itemReq.UnitPrice
		totalPrice := unitPrice.Mul(decimal.NewFromInt(int64(itemReq.Quantity)))

		item := entity.SalesOrderItem{
			ID:           uuid.New(),
			ProductID:    itemReq.ProductID,
			Quantity:     itemReq.Quantity,
			UnitPrice:    unitPrice.InexactFloat64(),
			TotalPrice:   totalPrice.InexactFloat64(),
			DeliveredQty: 0,
			RemainingQty: itemReq.Quantity,
			Notes:        itemReq.Notes,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		items = append(items, item)
		totalAmount = totalAmount.Add(totalPrice)
	}

	// Generate SO number
	soNumber, err := s.salesOrderRepo.GetNextSONumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGenerateSONumber, err)
	}

	// Start transaction
	tx, err := s.salesOrderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrBeginTransaction, err)
	}
	sqlxTx := tx.(*sqlx.Tx)
	defer sqlxTx.Rollback()

	salesOrder := &entity.SalesOrder{
		ID:                   uuid.New(),
		SONumber:             soNumber,
		CustomerID:           req.CustomerID,
		WarehouseID:          req.WarehouseID,
		OrderDate:            req.OrderDate,
		ExpectedDeliveryDate: req.ExpectedDeliveryDate,
		Status:               "DRAFT",
		SubTotal:             totalAmount.InexactFloat64(),
		TaxAmount:            0, // Can be calculated based on tax rules
		TotalAmount:          totalAmount.InexactFloat64(),
		Notes:                req.Notes,
		CreatedBy:            userID,
		UpdatedBy:            userID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save sales order
	if err := s.salesOrderRepo.CreateTx(ctx, tx, salesOrder); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCreateSalesOrder, err)
	}

	// Save sales order items
	for i := range items {
		items[i].SalesOrderID = salesOrder.ID
		if err := s.salesOrderRepo.CreateItemTx(ctx, tx, &items[i]); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrCreateSalesOrderItem, err)
		}
	}

	if err := sqlxTx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCommitTransaction, err)
	}

	// Add customer and items to response
	salesOrder.Customer = customer
	salesOrder.Items = items

	return salesOrder, nil
}

func (s *SalesOrderService) GetSalesOrderByID(ctx context.Context, id uuid.UUID) (*entity.SalesOrder, error) {
	salesOrder, err := s.salesOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrSalesOrderNotFound, err)
	}

	// Fetch items
	items, err := s.salesOrderRepo.GetItemsByOrderID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sales order items: %w", err)
	}
	salesOrder.Items = items

	return salesOrder, nil
}

func (s *SalesOrderService) GetSalesOrders(ctx context.Context, filter dto.SalesOrderFilter) ([]entity.SalesOrder, int, error) {
	entityFilter := entity.SalesOrderFilter{
		CustomerID:  filter.CustomerID,
		WarehouseID: filter.WarehouseID,
		Status:      nil,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
	}
	if filter.Status != nil {
		status := entity.SalesOrderStatus(*filter.Status)
		entityFilter.Status = &status
	}

	orders, err := s.salesOrderRepo.FindAll(ctx, entityFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get sales orders: %w", err)
	}

	return orders, len(orders), nil
}

func (s *SalesOrderService) ApproveSalesOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.SalesOrder, error) {
	salesOrder, err := s.salesOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrSalesOrderNotFound, err)
	}

	if salesOrder.Status != "DRAFT" {
		return nil, fmt.Errorf("sales order cannot be approved in status: %s", salesOrder.Status)
	}

	salesOrder.Status = "APPROVED"
	salesOrder.UpdatedBy = userID
	salesOrder.UpdatedAt = time.Now()

	if err := s.salesOrderRepo.Update(ctx, id, salesOrder); err != nil {
		return nil, fmt.Errorf("failed to approve sales order: %w", err)
	}

	return salesOrder, nil
}

func (s *SalesOrderService) CancelSalesOrder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.SalesOrder, error) {
	salesOrder, err := s.salesOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrSalesOrderNotFound, err)
	}

	if salesOrder.Status == "COMPLETED" || salesOrder.Status == "CANCELLED" {
		return nil, fmt.Errorf("sales order cannot be cancelled in status: %s", salesOrder.Status)
	}

	salesOrder.Status = "CANCELLED"
	salesOrder.UpdatedBy = userID
	salesOrder.UpdatedAt = time.Now()

	if err := s.salesOrderRepo.Update(ctx, id, salesOrder); err != nil {
		return nil, fmt.Errorf("failed to cancel sales order: %w", err)
	}

	return salesOrder, nil
}
