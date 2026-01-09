// internal/entity/report.go
package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StockReport struct {
	ProductID      uuid.UUID       `db:"product_id" json:"product_id"`
	ProductSKU     string          `db:"product_sku" json:"product_sku"`
	ProductName    string          `db:"product_name" json:"product_name"`
	CategoryName   string          `db:"category_name" json:"category_name"`
	WarehouseName  string          `db:"warehouse_name" json:"warehouse_name"`
	LocationCode   *string         `db:"location_code" json:"location_code,omitempty"`
	CurrentStock   int             `db:"current_stock" json:"current_stock"`
	ReservedStock  int             `db:"reserved_stock" json:"reserved_stock"`
	AvailableStock int             `db:"available_stock" json:"available_stock"`
	MinStock       int             `db:"min_stock" json:"min_stock"`
	MaxStock       int             `db:"max_stock" json:"max_stock"`
	UnitPrice      decimal.Decimal `db:"unit_price" json:"unit_price"`
	TotalValue     decimal.Decimal `db:"total_value" json:"total_value"`
	LastMovement   *time.Time      `db:"last_movement" json:"last_movement,omitempty"`
}

type SalesReport struct {
	Period        string          `json:"period"` // Daily, Weekly, Monthly
	Date          time.Time       `db:"date" json:"date"`
	TotalOrders   int             `db:"total_orders" json:"total_orders"`
	TotalQuantity int             `db:"total_quantity" json:"total_quantity"`
	TotalSales    decimal.Decimal `db:"total_sales" json:"total_sales"`
	TotalCost     decimal.Decimal `db:"total_cost" json:"total_cost"`
	TotalProfit   decimal.Decimal `db:"total_profit" json:"total_profit"`
	AverageOrder  decimal.Decimal `db:"average_order" json:"average_order"`
	TopProduct    *string         `json:"top_product,omitempty"`
	TopCustomer   *string         `json:"top_customer,omitempty"`
}

type InventoryValuation struct {
	WarehouseID   uuid.UUID       `db:"warehouse_id" json:"warehouse_id"`
	WarehouseName string          `db:"warehouse_name" json:"warehouse_name"`
	TotalProducts int             `db:"total_products" json:"total_products"`
	TotalQuantity int             `db:"total_quantity" json:"total_quantity"`
	TotalValue    decimal.Decimal `db:"total_value" json:"total_value"`
	AverageValue  decimal.Decimal `db:"average_value" json:"average_value"`
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
	Period       string          `db:"period" json:"period"`
	StartDate    time.Time       `db:"start_date" json:"start_date"`
	EndDate      time.Time       `db:"end_date" json:"end_date"`
	TotalSales   decimal.Decimal `db:"total_sales" json:"total_sales"`
	TotalCost    decimal.Decimal `db:"total_cost" json:"total_cost"`
	GrossProfit  decimal.Decimal `db:"gross_profit" json:"gross_profit"`
	Expenses     decimal.Decimal `db:"expenses" json:"expenses"`
	NetProfit    decimal.Decimal `db:"net_profit" json:"net_profit"`
	ProfitMargin decimal.Decimal `db:"profit_margin" json:"profit_margin"`
}

// TopProduct represents top selling products
type TopProduct struct {
	ProductID     uuid.UUID       `db:"product_id" json:"product_id"`
	ProductSKU    string          `db:"product_sku" json:"product_sku"`
	ProductName   string          `db:"product_name" json:"product_name"`
	TotalQuantity int             `db:"total_quantity" json:"total_quantity"`
	TotalRevenue  decimal.Decimal `db:"total_revenue" json:"total_revenue"`
	TotalProfit   decimal.Decimal `db:"total_profit" json:"total_profit"`
	OrderCount    int             `db:"order_count" json:"order_count"`
	AveragePrice  decimal.Decimal `db:"average_price" json:"average_price"`
}
