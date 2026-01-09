// internal/service/dashboard.service.go
package service

import (
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/repository"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type IDashboardService interface {
	GetDashboardSummary(ctx context.Context) (*entity.DashboardSummary, error)
	GetRecentActivities(ctx context.Context, limit int) ([]entity.RecentActivity, error)
	GetSalesTrend(ctx context.Context, days int) ([]entity.SalesTrend, error)
	GetInventoryOverview(ctx context.Context) (*entity.InventoryOverview, error)
}

type DashboardService struct {
	salesRepo     repository.ISalesOrderRepository
	purchaseRepo  repository.IPurchaseOrderRepository
	inventoryRepo repository.IInventoryRepository
	productRepo   repository.IProductRepository
	customerRepo  repository.ICustomerRepository
	supplierRepo  repository.ISupplierRepository
}

func (s *DashboardService) GetDashboardSummary(ctx context.Context) (*entity.DashboardSummary, error) {
	summary := &entity.DashboardSummary{}

	// Get today's date
	today := time.Now().Truncate(24 * time.Hour)

	// Today's sales
	todaySales, err := s.salesRepo.GetTotalSales(ctx, today, today.Add(24*time.Hour))
	if err == nil {
		summary.TodaySales = decimal.NewFromFloat(todaySales)
	}

	// Total products
	totalProducts, err := s.productRepo.GetCount(ctx)
	if err == nil {
		summary.TotalProducts = totalProducts
	}

	// Low stock products
	lowStock, err := s.productRepo.GetLowStockCount(ctx)
	if err == nil {
		summary.LowStockProducts = lowStock
	}

	// Pending purchase orders
	pendingPO, err := s.purchaseRepo.GetPendingCount(ctx)
	if err == nil {
		summary.PendingPurchaseOrders = pendingPO
	}

	// Pending sales orders
	pendingSO, err := s.salesRepo.GetPendingCount(ctx)
	if err == nil {
		summary.PendingSalesOrders = pendingSO
	}

	return summary, nil
}
