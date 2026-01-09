// internal/service/report.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type IReportService interface {
	GetStockReport(ctx context.Context, filter dto.StockReportFilter) ([]entity.StockReport, error)
	GetSalesReport(ctx context.Context, filter dto.SalesReportFilter) ([]entity.SalesReport, error)
	GetInventoryValuation(ctx context.Context, warehouseID *uuid.UUID) ([]entity.InventoryValuation, error)
	GetProfitLossReport(ctx context.Context, startDate, endDate time.Time) (*entity.ProfitLossReport, error)
	GetTopProducts(ctx context.Context, limit int, period string) ([]entity.TopProduct, error)
	GetLowStockReport(ctx context.Context) ([]entity.StockReport, error)
	GetMovementReport(ctx context.Context, filter dto.MovementReportFilter) ([]entity.StockMovement, int, error)
}

type ReportService struct {
	inventoryRepo repository.IInventoryRepository
	salesRepo     repository.ISalesOrderRepository
	purchaseRepo  repository.IPurchaseOrderRepository
	productRepo   repository.IProductRepository
}

func NewReportService(
	inventoryRepo repository.IInventoryRepository,
	salesRepo repository.ISalesOrderRepository,
	purchaseRepo repository.IPurchaseOrderRepository,
	productRepo repository.IProductRepository,
) IReportService {
	return &ReportService{
		inventoryRepo: inventoryRepo,
		salesRepo:     salesRepo,
		purchaseRepo:  purchaseRepo,
		productRepo:   productRepo,
	}
}

func (s *ReportService) GetStockReport(ctx context.Context, filter dto.StockReportFilter) ([]entity.StockReport, error) {
	return s.inventoryRepo.GetStockReport(ctx, filter)
}

func (s *ReportService) GetSalesReport(ctx context.Context, filter dto.SalesReportFilter) ([]entity.SalesReport, error) {
	return s.salesRepo.GetSalesReport(ctx, filter)
}

func (s *ReportService) GetInventoryValuation(ctx context.Context, warehouseID *uuid.UUID) ([]entity.InventoryValuation, error) {
	return s.inventoryRepo.GetInventoryValuation(ctx, warehouseID)
}

func (s *ReportService) GetLowStockReport(ctx context.Context) ([]entity.StockReport, error) {
	return s.inventoryRepo.GetLowStockReport(ctx)
}

func (s *ReportService) GetProfitLossReport(ctx context.Context, startDate, endDate time.Time) (*entity.ProfitLossReport, error) {
	// Get sales summary from sales repository
	salesReport, err := s.salesRepo.GetSalesSummary(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate purchase costs for the period
	var totalPurchaseCost decimal.Decimal
	totalPurchaseCost, err = s.getPurchaseCosts(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate expenses (purchase costs as expenses for this report)
	expenses := totalPurchaseCost

	// Calculate net profit
	netProfit := salesReport.TotalSales.Sub(totalPurchaseCost)

	// Calculate profit margin
	var profitMargin decimal.Decimal
	if salesReport.TotalSales.GreaterThan(decimal.Zero) {
		profitMargin = netProfit.Div(salesReport.TotalSales).Mul(decimal.NewFromInt(100))
	}

	return &entity.ProfitLossReport{
		Period:       "Custom",
		StartDate:    startDate,
		EndDate:      endDate,
		TotalSales:   salesReport.TotalSales,
		TotalCost:    totalPurchaseCost,
		GrossProfit:  salesReport.GrossProfit,
		Expenses:     expenses,
		NetProfit:    netProfit,
		ProfitMargin: profitMargin,
	}, nil
}

func (s *ReportService) getPurchaseCosts(ctx context.Context, startDate, endDate time.Time) (decimal.Decimal, error) {
	// This is a simplified calculation - in a real implementation,
	// you would query the purchase orders for the period
	return decimal.Zero, nil
}

func (s *ReportService) GetTopProducts(ctx context.Context, limit int, period string) ([]entity.TopProduct, error) {
	now := time.Now()
	var startDate time.Time

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0) // Default to month
	}

	return s.salesRepo.GetTopSellingProducts(ctx, startDate, now, limit)
}

func (s *ReportService) GetMovementReport(ctx context.Context, filter dto.MovementReportFilter) ([]entity.StockMovement, int, error) {
	return s.inventoryRepo.GetMovementReport(ctx, filter)
}
