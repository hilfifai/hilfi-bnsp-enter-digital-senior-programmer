// internal/handler/inventory/route.go
package inventory

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	inventory := e.Group("/inventory")
	inventory.Use(authMiddleware)

	inventory.POST("/movements", h.CreateStockMovement)
	inventory.POST("/transfer", h.TransferStock)
	inventory.POST("/adjustment", h.AdjustStock)
	inventory.POST("/opname", h.PerformStockOpname)

	inventory.GET("/movements", h.GetStockMovements)
	inventory.GET("/summary", h.GetStockSummary)
	inventory.GET("/alerts/low-stock", h.GetLowStockAlerts)
	inventory.GET("/product/:id", h.GetStockByProduct)
	inventory.GET("/warehouse/:id", h.GetStockByWarehouse)
	inventory.GET("/history/:product_id", h.GetStockHistory)
}
