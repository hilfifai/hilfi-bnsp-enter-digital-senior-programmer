// internal/handler/product_stock/route.go
package product_stock

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NewRoutes(e *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	stocks := e.Group("/product-stocks")
	stocks.Use(authMiddleware)

	stocks.GET("", h.GetStock)
	stocks.POST("", h.UpdateStock)
	stocks.GET("/movements", h.GetStockMovements)
	stocks.GET("/low-stock", h.GetLowStockProducts)
	stocks.GET("/product/:id", h.GetStockByProduct)
	stocks.GET("/warehouse/:id", h.GetStockByWarehouse)
}
