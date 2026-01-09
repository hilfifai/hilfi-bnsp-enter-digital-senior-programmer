// internal/entity/report.go
package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StockReport struct {
	ProductID      uuid.UUID       `json:"product_id"`
	ProductSKU     string          `json:"product_sku"`
	ProductName    string          `json:"product_name"`
	CategoryName   string          `json:"category_name"`
	WarehouseName  string          `json:"warehouse_name"`
	LocationCode   *string         `json:"location_code,omitempty"`
	CurrentStock   int             `json:"current_stock"`
	ReservedStock  int             `json:"reserved_stock"`
	AvailableStock int             `json:"available_stock"`
	MinStock       int             `json:"min_stock"`
	MaxStock       int             `json:"max_stock"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	TotalValue     decimal.Decimal `json:"total_value"`
	LastMovement   *time.Time      `json:"last_movement,omitempty"`
}

type SalesReport struct {
	Period        string          `json:"period"` // Daily, Weekly, Monthly
	Date          time.Time       `json:"date"`
	TotalOrders   int             `json:"total_orders"`
	TotalQuantity int             `json:"total_quantity"`
	TotalSales    decimal.Decimal `json:"total_sales"`
	TotalCost     decimal.Decimal `json:"total_cost"`
	TotalProfit   decimal.Decimal `json:"total_profit"`
	AverageOrder  decimal.Decimal `json:"average_order"`
	TopProduct    *string         `json:"top_product,omitempty"`
	TopCustomer   *string         `json:"top_customer,omitempty"`
}

type InventoryValuation struct {
	WarehouseID   uuid.UUID       `json:"warehouse_id"`
	WarehouseName string          `json:"warehouse_name"`
	TotalProducts int             `json:"total_products"`
	TotalQuantity int             `json:"total_quantity"`
	TotalValue    decimal.Decimal `json:"total_value"`
	AverageValue  decimal.Decimal `json:"average_value"`
}

// DashboardSummary represents the dashboard summary data
type DashboardSummary struct {
	TodaySales            decimal.Decimal `json:"today_sales"`
	TotalProducts         int             `json:"total_products"`
	LowStockProducts      int             `json:"low_stock_products"`
	PendingPurchaseOrders int             `json:"pending_purchase_orders"`
	PendingSalesOrders    int             `json:"pending_sales_orders"`
}

// RecentActivity represents a recent activity in the system
type RecentActivity struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// SalesTrend represents sales trend data
type SalesTrend struct {
	Date       time.Time       `json:"date"`
	TotalSales decimal.Decimal `json:"total_sales"`
	OrderCount int             `json:"order_count"`
}

// InventoryOverview represents the inventory overview
type InventoryOverview struct {
	TotalProducts   int             `json:"total_products"`
	TotalQuantity   int             `json:"total_quantity"`
	TotalValue      decimal.Decimal `json:"total_value"`
	LowStockCount   int             `json:"low_stock_count"`
	OutOfStockCount int             `json:"out_of_stock_count"`
}

// ProfitLossReport represents profit and loss report
type ProfitLossReport struct {
	Period       string          `json:"period"`
	StartDate    time.Time       `json:"start_date"`
	EndDate      time.Time       `json:"end_date"`
	TotalSales   decimal.Decimal `json:"total_sales"`
	TotalCost    decimal.Decimal `json:"total_cost"`
	GrossProfit  decimal.Decimal `json:"gross_profit"`
	Expenses     decimal.Decimal `json:"expenses"`
	NetProfit    decimal.Decimal `json:"net_profit"`
	ProfitMargin decimal.Decimal `json:"profit_margin"`
}

// TopProduct represents top selling products
type TopProduct struct {
	ProductID     uuid.UUID       `json:"product_id"`
	ProductSKU    string          `json:"product_sku"`
	ProductName   string          `json:"product_name"`
	TotalQuantity int             `json:"total_quantity"`
	TotalRevenue  decimal.Decimal `json:"total_revenue"`
	TotalProfit   decimal.Decimal `json:"total_profit"`
	OrderCount    int             `json:"order_count"`
	AveragePrice  decimal.Decimal `json:"average_price"`
}
