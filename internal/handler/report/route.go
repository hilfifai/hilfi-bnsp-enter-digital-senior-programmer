// internal/handler/report/route.go
package report

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	reports := e.Group("/reports")
	reports.Use(authMiddleware)

	reports.GET("/stock", h.GetStockReport)
	reports.GET("/sales", h.GetSalesReport)
	reports.GET("/inventory-valuation", h.GetInventoryValuation)
	reports.GET("/profit-loss", h.GetProfitLossReport)
	reports.GET("/top-products", h.GetTopProducts)
	reports.GET("/low-stock", h.GetLowStockReport)
	reports.GET("/movements", h.GetMovementReport)
}
